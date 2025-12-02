package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"project-management/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	Logger *logrus.Logger
	mu     sync.RWMutex
)

// ZipRotateWriter 支持ZIP压缩的日志轮转写入器
type ZipRotateWriter struct {
	filename   string
	maxSize    int64 // 最大文件大小（字节）
	maxAge     int   // 保留天数
	file       *os.File
	size       int64
	mu         sync.Mutex
}

// NewZipRotateWriter 创建新的ZIP轮转写入器
func NewZipRotateWriter(filename string, maxSizeMB int, maxAgeDays int) (*ZipRotateWriter, error) {
	w := &ZipRotateWriter{
		filename: filename,
		maxSize:  int64(maxSizeMB) * 1024 * 1024, // 转换为字节
		maxAge:   maxAgeDays,
	}
	
	// 打开或创建日志文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	
	// 获取当前文件大小
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	
	w.file = file
	w.size = info.Size()
	
	// 启动清理goroutine
	go w.cleanupOldFiles()
	
	return w, nil
}

// Write 实现io.Writer接口
func (w *ZipRotateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// 检查是否需要轮转
	if w.size+w.size > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	
	// 写入文件
	n, err = w.file.Write(p)
	if err == nil {
		w.size += int64(n)
	}
	
	return n, err
}

// rotate 轮转日志文件
func (w *ZipRotateWriter) rotate() error {
	// 关闭当前文件
	if w.file != nil {
		w.file.Close()
	}
	
	// 如果文件不存在或为空，不需要轮转
	info, err := os.Stat(w.filename)
	if err != nil || info.Size() == 0 {
		// 重新打开文件
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = 0
		return nil
	}
	
	// 生成压缩文件名（带时间戳）
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(w.filename)
	base := filepath.Base(w.filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	zipFilename := filepath.Join(dir, fmt.Sprintf("%s-%s%s.zip", name, timestamp, ext))
	
	// 创建ZIP文件
	zipFile, err := os.Create(zipFilename)
	if err != nil {
		// 如果创建ZIP失败，重新打开原文件继续写入
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = info.Size()
		return nil
	}
	defer zipFile.Close()
	
	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	// 打开原日志文件
	logFile, err := os.Open(w.filename)
	if err != nil {
		// 如果打开失败，重新打开原文件继续写入
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = info.Size()
		return nil
	}
	defer logFile.Close()
	
	// 在ZIP中创建文件
	zipEntry, err := zipWriter.Create(base)
	if err != nil {
		logFile.Close()
		// 如果创建失败，重新打开原文件继续写入
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = info.Size()
		return nil
	}
	
	// 复制文件内容到ZIP
	_, err = io.Copy(zipEntry, logFile)
	if err != nil {
		zipFile.Close()
		os.Remove(zipFilename) // 删除失败的ZIP文件
		// 重新打开原文件继续写入
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = info.Size()
		return nil
	}
	
	// 关闭ZIP写入器
	if err := zipWriter.Close(); err != nil {
		os.Remove(zipFilename) // 删除失败的ZIP文件
		// 重新打开原文件继续写入
		file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		w.file = file
		w.size = info.Size()
		return nil
	}
	
	// 删除原日志文件
	os.Remove(w.filename)
	
	// 创建新的日志文件
	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.size = 0
	
	return nil
}

// cleanupOldFiles 清理旧文件
func (w *ZipRotateWriter) cleanupOldFiles() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()
	
	for range ticker.C {
		w.mu.Lock()
		dir := filepath.Dir(w.filename)
		cutoffTime := time.Now().AddDate(0, 0, -w.maxAge)
		
		// 读取目录
		entries, err := os.ReadDir(dir)
		if err != nil {
			w.mu.Unlock()
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			// 只处理ZIP文件
			if !strings.HasSuffix(entry.Name(), ".zip") {
				continue
			}
			
			// 检查文件修改时间
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			if info.ModTime().Before(cutoffTime) {
				filePath := filepath.Join(dir, entry.Name())
				os.Remove(filePath) // 删除过期文件
			}
		}
		w.mu.Unlock()
	}
}

// Close 关闭文件
func (w *ZipRotateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// MicrosecondFormatter 自定义Formatter，时间格式精确到微秒
type MicrosecondFormatter struct {
	logrus.TextFormatter
}

// Format 格式化日志条目
func (f *MicrosecondFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000000")
	
	// 获取调用位置信息
	caller := ""
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
		caller = fmt.Sprintf("%s %s", funcVal, fileVal)
	}
	
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message
	
	// 构建日志行
	var parts []string
	parts = append(parts, timestamp)
	parts = append(parts, fmt.Sprintf("[%s]", level))
	if caller != "" {
		parts = append(parts, caller)
	}
	parts = append(parts, message)
	
	// 添加字段
	if len(entry.Data) > 0 {
		for k, v := range entry.Data {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	
	result := strings.Join(parts, " ") + "\n"
	return []byte(result), nil
}

// InitLogger 初始化日志系统
func InitLogger(db *gorm.DB) error {
	mu.Lock()
	defer mu.Unlock()
	
	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// 如果创建目录失败，回退到标准库log
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v, falling back to standard log\n", err)
		return err
	}
	
	// 从数据库读取日志级别配置
	logLevel := "error" // 默认级别
	if db != nil {
		var config model.SystemConfig
		if err := db.Where("key = ?", "log_level").First(&config).Error; err == nil {
			logLevel = config.Value
		}
	}
	
	// 解析日志级别
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.ErrorLevel
	}
	
	// 创建ZIP轮转写入器
	logFile := filepath.Join(logDir, "app.log")
	zipWriter, err := NewZipRotateWriter(logFile, 100, 7) // 100MB, 7天
	if err != nil {
		return fmt.Errorf("failed to create zip rotate writer: %w", err)
	}
	
	// 创建logrus logger
	Logger = logrus.New()
	Logger.SetLevel(level)
	
	// 设置自定义Formatter
	Logger.SetFormatter(&MicrosecondFormatter{
		TextFormatter: logrus.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   false, // 我们自定义时间格式
			TimestampFormat: "2006-01-02 15:04:05.000000",
		},
	})
	
	// 启用调用者信息
	Logger.SetReportCaller(true)
	
	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, zipWriter)
	Logger.SetOutput(multiWriter)
	
	return nil
}

// SetLogLevel 设置日志级别
func SetLogLevel(level string, db *gorm.DB) error {
	mu.Lock()
	defer mu.Unlock()
	
	// 验证日志级别
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", level)
	}
	
	// 更新logrus日志级别
	if Logger != nil {
		Logger.SetLevel(parsedLevel)
	}
	
	// 保存到数据库
	if db != nil {
		config := model.SystemConfig{
			Key:   "log_level",
			Value: level,
			Type:  "string",
		}
		if err := db.Where("key = ?", "log_level").
			Assign(model.SystemConfig{Value: level, Type: "string"}).
			FirstOrCreate(&config).Error; err != nil {
			return fmt.Errorf("failed to save log level to database: %w", err)
		}
	}
	
	return nil
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() string {
	mu.RLock()
	defer mu.RUnlock()
	
	if Logger == nil {
		return "error"
	}
	
	return Logger.GetLevel().String()
}

