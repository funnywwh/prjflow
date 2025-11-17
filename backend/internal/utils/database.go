package utils

import (
	"fmt"

	_ "modernc.org/sqlite" // 纯Go SQLite驱动，支持静态编译，必须在 gorm.io/driver/sqlite 之前导入
	
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"project-management/internal/config"
)

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

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}
