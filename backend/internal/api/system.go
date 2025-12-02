package api

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemHandler struct {
	db *gorm.DB
}

func NewSystemHandler(db *gorm.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

// GetBackupConfig 获取备份配置
func (h *SystemHandler) GetBackupConfig(c *gin.Context) {
	var enabledConfig model.SystemConfig
	var timeConfig model.SystemConfig
	var lastDateConfig model.SystemConfig

	enabled := false
	backupTime := "02:00"
	lastBackupDate := ""

	// 读取备份启用状态
	if err := h.db.Where("key = ?", "backup_enabled").First(&enabledConfig).Error; err == nil {
		enabled = enabledConfig.Value == "true"
	}

	// 读取备份时间
	if err := h.db.Where("key = ?", "backup_time").First(&timeConfig).Error; err == nil {
		backupTime = timeConfig.Value
	}

	// 读取上次备份日期
	if err := h.db.Where("key = ?", "backup_last_date").First(&lastDateConfig).Error; err == nil {
		lastBackupDate = lastDateConfig.Value
	}

	utils.Success(c, gin.H{
		"enabled":         enabled,
		"backup_time":     backupTime,
		"last_backup_date": lastBackupDate,
	})
}

// SaveBackupConfig 保存备份配置
func (h *SystemHandler) SaveBackupConfig(c *gin.Context) {
	var req struct {
		Enabled    bool   `json:"enabled" binding:"required"`
		BackupTime string `json:"backup_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证备份时间格式 (HH:MM)
	if _, err := time.Parse("15:04", req.BackupTime); err != nil {
		utils.Error(c, 400, "备份时间格式错误，应为 HH:MM (24小时制)")
		return
	}

	// 保存备份启用状态
	enabledValue := "false"
	if req.Enabled {
		enabledValue = "true"
	}
	enabledConfig := model.SystemConfig{
		Key:   "backup_enabled",
		Value: enabledValue,
		Type:  "boolean",
	}
	if err := h.db.Where("key = ?", "backup_enabled").
		Assign(model.SystemConfig{Value: enabledValue, Type: "boolean"}).
		FirstOrCreate(&enabledConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存备份启用状态失败: "+err.Error())
		return
	}

	// 保存备份时间
	timeConfig := model.SystemConfig{
		Key:   "backup_time",
		Value: req.BackupTime,
		Type:  "string",
	}
	if err := h.db.Where("key = ?", "backup_time").
		Assign(model.SystemConfig{Value: req.BackupTime, Type: "string"}).
		FirstOrCreate(&timeConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存备份时间失败: "+err.Error())
		return
	}

	// 重新加载定时任务配置
	scheduler := utils.GetBackupScheduler(h.db)
	scheduler.Reload()

	utils.Success(c, gin.H{
		"message": "备份配置已保存",
	})
}

// TriggerBackup 手动触发备份
func (h *SystemHandler) TriggerBackup(c *gin.Context) {
	// 执行备份
	if err := utils.TriggerBackup(h.db); err != nil {
		if err.Error() == "backup is already in progress" {
			utils.Error(c, 400, "备份正在进行中，请稍后再试")
			return
		}
		utils.Error(c, utils.CodeError, "备份失败: "+err.Error())
		return
	}

	// 更新上次备份日期
	today := time.Now().Format("2006-01-02")
	lastDateConfig := model.SystemConfig{
		Key:   "backup_last_date",
		Value: today,
		Type:  "string",
	}
	h.db.Where("key = ?", "backup_last_date").
		Assign(model.SystemConfig{Value: today, Type: "string"}).
		FirstOrCreate(&lastDateConfig)

	utils.Success(c, gin.H{
		"message": "备份已触发",
	})
}

// GetLogLevel 获取当前日志级别
func (h *SystemHandler) GetLogLevel(c *gin.Context) {
	level := utils.GetLogLevel()
	utils.Success(c, gin.H{
		"level": level,
	})
}

// SetLogLevel 设置日志级别
func (h *SystemHandler) SetLogLevel(c *gin.Context) {
	var req struct {
		Level string `json:"level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证日志级别
	validLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, v := range validLevels {
		if req.Level == v {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.Error(c, 400, fmt.Sprintf("无效的日志级别: %s，支持的值: %v", req.Level, validLevels))
		return
	}

	// 设置日志级别
	if err := utils.SetLogLevel(req.Level, h.db); err != nil {
		utils.Error(c, utils.CodeError, "设置日志级别失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "日志级别已更新",
		"level":   req.Level,
	})
}

// LogFileInfo 日志文件信息
type LogFileInfo struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	SizeFormatted string `json:"size_formatted"`
	ModTime     string `json:"mod_time"`
}

// GetLogFiles 获取日志文件列表
func (h *SystemHandler) GetLogFiles(c *gin.Context) {
	logDir := "logs"
	
	// 检查目录是否存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		utils.Success(c, gin.H{
			"files": []LogFileInfo{},
		})
		return
	}

	// 读取目录
	entries, err := os.ReadDir(logDir)
	if err != nil {
		utils.Error(c, utils.CodeError, "读取日志目录失败: "+err.Error())
		return
	}

	var files []LogFileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只返回日志文件（.log 和 .zip 文件）
		name := entry.Name()
		if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".zip") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 格式化文件大小
		size := info.Size()
		sizeFormatted := formatFileSize(size)

		// 格式化修改时间
		modTime := info.ModTime().Format("2006-01-02 15:04:05")

		files = append(files, LogFileInfo{
			Filename:      name,
			Size:          size,
			SizeFormatted: sizeFormatted,
			ModTime:       modTime,
		})
	}

	// 按修改时间倒序排列
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].ModTime < files[j].ModTime {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	utils.Success(c, gin.H{
		"files": files,
	})
}

// DownloadLogFile 下载日志文件
func (h *SystemHandler) DownloadLogFile(c *gin.Context) {
	filename := c.Param("filename")
	
	// 验证文件名，防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		utils.Error(c, 400, "无效的文件名")
		return
	}

	logDir := "logs"
	filePath := filepath.Join(logDir, filename)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			utils.Error(c, 404, "文件不存在")
		} else {
			utils.Error(c, utils.CodeError, "读取文件失败: "+err.Error())
		}
		return
	}

	// 如果是当前日志文件（app.log），需要压缩为ZIP后下载
	if filename == "app.log" {
		// 生成压缩文件名
		timestamp := time.Now().Format("20060102-150405")
		compressedFilename := fmt.Sprintf("app-%s.log.zip", timestamp)
		
		// 设置响应头
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", compressedFilename))
		c.Header("Content-Length", "") // 流式传输，不设置长度

		// 打开源文件
		srcFile, err := os.Open(filePath)
		if err != nil {
			utils.Error(c, utils.CodeError, "打开文件失败: "+err.Error())
			return
		}
		defer srcFile.Close()

		// 创建ZIP writer
		zipWriter := zip.NewWriter(c.Writer)
		defer zipWriter.Close()

		// 在ZIP中创建文件，设置文件头信息
		zipHeader := &zip.FileHeader{
			Name:     filename,
			Method:   zip.Deflate,
			Modified: info.ModTime(),
		}
		zipEntry, err := zipWriter.CreateHeader(zipHeader)
		if err != nil {
			utils.Error(c, utils.CodeError, "创建ZIP条目失败: "+err.Error())
			return
		}

		// 流式复制并压缩
		_, err = io.Copy(zipEntry, srcFile)
		if err != nil {
			// 如果写入失败，已经部分写入，无法返回错误响应
			return
		}
	} else {
		// 历史压缩文件（ZIP）直接下载
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))

		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			utils.Error(c, utils.CodeError, "打开文件失败: "+err.Error())
			return
		}
		defer file.Close()

		// 流式复制
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			return
		}
	}
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

