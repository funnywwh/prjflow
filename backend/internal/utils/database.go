package utils

import (
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // 纯Go SQLite驱动，支持静态编译，必须在 gorm.io/driver/sqlite 之前导入
	
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"project-management/internal/config"
)

// gormLogrusWriter 实现GORM的logger.Writer接口，将GORM日志输出到logrus
type gormLogrusWriter struct {
	logger *logrus.Logger
}

// Printf 实现logger.Writer接口
func (w *gormLogrusWriter) Printf(format string, args ...interface{}) {
	if w.logger != nil {
		// 将GORM的日志格式转换为logrus日志
		message := fmt.Sprintf(format, args...)
		w.logger.Info(message)
	}
}

func InitDB() (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.AppConfig.Database.Type {
	case "sqlite":
		// 使用纯Go实现的SQLite驱动（modernc.org/sqlite）
		// 支持静态编译（CGO_ENABLED=0），无需CGO和系统库
		// modernc.org/sqlite 注册为 "sqlite" 驱动（不是 "sqlite3"）
		// 使用 sqlite.New() 并指定 DriverName 为 "sqlite"
		dialector = sqlite.New(sqlite.Config{
			DriverName: "sqlite", // 使用 modernc.org/sqlite 注册的驱动名
			DSN:        config.AppConfig.Database.DSN,
		})
	case "mysql":
		dsn := config.AppConfig.Database.DSN
		if dsn == "" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				config.AppConfig.Database.User,
				config.AppConfig.Database.Password,
				config.AppConfig.Database.Host,
				config.AppConfig.Database.Port,
				config.AppConfig.Database.DBName,
			)
		}
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.AppConfig.Database.Type)
	}

	// 配置GORM logger
	var gormLogger logger.Interface
	if Logger != nil {
		// 创建logrus适配器，实现GORM的logger.Writer接口
		gormWriter := &gormLogrusWriter{logger: Logger}
		gormLogger = logger.New(
			gormWriter,
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info,
				Colorful:      false,
			},
		)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}

// IsUniqueConstraintError 检查是否是唯一约束错误
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint")
}

// IsUniqueConstraintOnField 检查是否是特定字段的唯一约束错误
func IsUniqueConstraintOnField(err error, fieldName string) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// SQLite: UNIQUE constraint failed: modules.name
	// MySQL: Duplicate entry 'xxx' for key 'modules.name'
	return strings.Contains(errStr, fieldName) && IsUniqueConstraintError(err)
}

// IsRecordNotFound 检查是否是记录不存在错误
func IsRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err == gorm.ErrRecordNotFound
}
