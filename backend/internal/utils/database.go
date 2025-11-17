package utils

import (
	"fmt"

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
		dialector = sqlite.Open(config.AppConfig.Database.DSN)
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

