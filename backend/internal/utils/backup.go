package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"project-management/internal/config"
	"gorm.io/gorm"
)

// BackupDatabase 备份数据库（改进版，使用 VACUUM INTO 确保数据一致性）
func BackupDatabase(db *gorm.DB) error {
	// 加载配置
	if err := config.LoadConfig(""); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 检查数据库类型，只支持 SQLite 备份
	if config.AppConfig.Database.Type != "sqlite" {
		return fmt.Errorf("backup only supports SQLite database, current type: %s", config.AppConfig.Database.Type)
	}

	// 获取数据库文件路径
	dbPath := config.AppConfig.Database.DSN
	if !filepath.IsAbs(dbPath) {
		// 相对路径，从当前工作目录或可执行文件目录查找
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			// 尝试从可执行文件目录查找
			exePath, err := os.Executable()
			if err == nil {
				exeDir := filepath.Dir(exePath)
				absPath := filepath.Join(exeDir, dbPath)
				if _, err := os.Stat(absPath); err == nil {
					dbPath = absPath
				}
			}
		}
	}

	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file not found: %s", dbPath)
	}

	// 创建备份目录（在数据库文件同目录下）
	dbDir := filepath.Dir(dbPath)
	backupDir := filepath.Join(dbDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名（带时间戳）
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("data_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFileName)

	// 获取绝对路径（VACUUM INTO 需要绝对路径）
	absBackupPath, err := filepath.Abs(backupPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 转义路径中的单引号
	escapedPath := strings.ReplaceAll(absBackupPath, "'", "''")

	log.Printf("[Backup] Starting database backup: %s -> %s", dbPath, absBackupPath)

	// 尝试使用 VACUUM INTO 命令创建备份（SQLite 3.27.0+）
	// 这会创建一个一致的数据库快照，支持在线备份
	sql := fmt.Sprintf("VACUUM INTO '%s'", escapedPath)
	if err := db.Exec(sql).Error; err != nil {
		// 如果 VACUUM INTO 失败，回退到文件复制方式（带警告）
		log.Printf("[Backup] Warning: VACUUM INTO failed: %v, falling back to file copy", err)
		if err := copyDatabaseFile(dbPath, absBackupPath); err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}
	}

	// 获取文件大小
	fileInfo, err := os.Stat(absBackupPath)
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}

	fileSize := float64(fileInfo.Size())
	unit := "B"
	if fileSize >= 1024*1024 {
		fileSize /= 1024 * 1024
		unit = "MB"
	} else if fileSize >= 1024 {
		fileSize /= 1024
		unit = "KB"
	}

	log.Printf("[Backup] ✓ Backup created: %s (%.2f %s)", absBackupPath, fileSize, unit)

	// 尝试压缩备份文件
	compressedPath := absBackupPath + ".gz"
	if err := compressFile(absBackupPath, compressedPath); err == nil {
		// 压缩成功，删除原文件
		os.Remove(absBackupPath)
		log.Printf("[Backup] ✓ Backup compressed: %s", compressedPath)
	} else {
		log.Printf("[Backup] Warning: Failed to compress backup: %v", err)
	}

	// 清理旧备份（保留最近7天）
	if err := cleanupOldBackups(backupDir, 7); err != nil {
		log.Printf("[Backup] Warning: Failed to cleanup old backups: %v", err)
	}

	// 显示备份统计
	backupCount, totalSize, err := getBackupStats(backupDir)
	if err == nil {
		log.Printf("[Backup] Statistics: %d backups, total size: %s", backupCount, formatSize(totalSize))
	}

	return nil
}

// copyDatabaseFile 复制数据库文件（备选方案）
func copyDatabaseFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dstFile.Close()

	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		os.Remove(dstPath)
		return fmt.Errorf("failed to copy database file: %w", err)
	}

	if written == 0 {
		return fmt.Errorf("backup file is empty")
	}

	return nil
}

// compressFile 压缩文件
func compressFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, srcFile)
	return err
}

// cleanupOldBackups 清理旧备份
func cleanupOldBackups(backupDir string, keepDays int) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	removedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理备份文件
		if !strings.HasPrefix(entry.Name(), "data_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 如果文件修改时间早于保留期限，删除
		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(backupDir, entry.Name())
			if err := os.Remove(filePath); err == nil {
				removedCount++
			}
		}
	}

	if removedCount > 0 {
		log.Printf("[Backup] Cleaned up %d old backup(s) (older than %d days)", removedCount, keepDays)
	}

	return nil
}

// getBackupStats 获取备份统计信息
func getBackupStats(backupDir string) (int, int64, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0, 0, err
	}

	count := 0
	totalSize := int64(0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasPrefix(entry.Name(), "data_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		count++
		totalSize += info.Size()
	}

	return count, totalSize, nil
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	unit := "B"
	fsize := float64(size)
	if fsize >= 1024*1024*1024 {
		fsize /= 1024 * 1024 * 1024
		unit = "GB"
	} else if fsize >= 1024*1024 {
		fsize /= 1024 * 1024
		unit = "MB"
	} else if fsize >= 1024 {
		fsize /= 1024
		unit = "KB"
	}
	return fmt.Sprintf("%.2f %s", fsize, unit)
}

