package unit

import (
	"os"
	"testing"

	_ "modernc.org/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	// 初始化测试配置（如果未初始化）
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Database: config.DatabaseConfig{
				Type: "sqlite",
				DSN:  ":memory:",
			},
		}
	} else if config.AppConfig.Database.Type == "" {
		config.AppConfig.Database.Type = "sqlite"
		config.AppConfig.Database.DSN = ":memory:"
	}

	// 使用内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试时关闭日志
	})
	if err != nil {
		t.Fatalf("Failed to connect test database: %v", err)
	}

	// 自动迁移所有模型
	if err := utils.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// SetupTestDBWithFile 创建基于文件的测试数据库（用于需要持久化的测试）
func SetupTestDBWithFile(t *testing.T) (*gorm.DB, string) {
	// 初始化测试配置（如果未初始化）
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Database: config.DatabaseConfig{
				Type: "sqlite",
			},
		}
	} else if config.AppConfig.Database.Type == "" {
		config.AppConfig.Database.Type = "sqlite"
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	tmpPath := tmpFile.Name()

	// 设置DSN
	config.AppConfig.Database.DSN = tmpPath

	db, err := gorm.Open(sqlite.Open(tmpPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		os.Remove(tmpPath)
		t.Fatalf("Failed to connect test database: %v", err)
	}

	// 自动迁移所有模型
	if err := utils.AutoMigrate(db); err != nil {
		os.Remove(tmpPath)
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db, tmpPath
}

// TeardownTestDBWithFile 清理基于文件的测试数据库
func TeardownTestDBWithFile(t *testing.T, db *gorm.DB, filePath string) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	if filePath != "" {
		os.Remove(filePath)
	}
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB, username, nickname string) *model.User {
	user := &model.User{
		Username:     username,
		Nickname:     nickname,
		Email:        username + "@test.com",
		Status:       1,
		WeChatOpenID: "test_openid_" + username, // 设置唯一的wechat_open_id
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

// CreateTestProject 创建测试项目
func CreateTestProject(t *testing.T, db *gorm.DB, name string) *model.Project {
	project := &model.Project{
		Name:        name,
		Description: "Test project",
		Status:      1, // 1=active
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	return project
}

