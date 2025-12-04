package unit

import (
	"os"
	"testing"
	"time"

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
	wechatOpenID := "test_openid_" + username
	user := &model.User{
		Username:     username,
		Nickname:     nickname,
		Email:        username + "@test.com",
		Status:       1,
		WeChatOpenID: &wechatOpenID, // 设置唯一的wechat_open_id
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

// CreateTestTag 创建测试标签
func CreateTestTag(t *testing.T, db *gorm.DB, name string) *model.Tag {
	// 确保标签名称唯一（添加时间戳）
	uniqueName := name + "_" + time.Now().Format("20060102150405")
	
	// 检查是否已存在，如果存在则添加纳秒时间戳
	var existingTag model.Tag
	if err := db.Where("name = ?", uniqueName).First(&existingTag).Error; err == nil {
		uniqueName = name + "_" + time.Now().Format("20060102150405000000")
	}

	tag := &model.Tag{
		Name:  uniqueName,
		Color: "blue",
	}
	if err := db.Create(tag).Error; err != nil {
		t.Fatalf("Failed to create test tag: %v", err)
	}
	return tag
}

// CreateTestProject 创建测试项目
func CreateTestProject(t *testing.T, db *gorm.DB, name string) *model.Project {
	// 生成唯一的项目代码
	code := "TEST_" + name
	// 如果代码已存在，添加时间戳确保唯一性
	var existingProject model.Project
	if err := db.Where("code = ?", code).First(&existingProject).Error; err == nil {
		code = code + "_" + time.Now().Format("20060102150405")
	}

	project := &model.Project{
		Name:        name,
		Code:        code,
		Description: "Test project",
		Status:      "doing", // doing=进行中
	}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	return project
}

// CreateTestAdminRole 创建管理员角色
func CreateTestAdminRole(t *testing.T, db *gorm.DB) *model.Role {
	var adminRole model.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = model.Role{
			Name:        "管理员",
			Code:        "admin",
			Description: "系统管理员，拥有所有权限",
			Status:      1,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			t.Fatalf("Failed to create admin role: %v", err)
		}
	}
	return &adminRole
}

// CreateTestAdminUser 创建管理员用户
func CreateTestAdminUser(t *testing.T, db *gorm.DB, username, nickname string) *model.User {
	user := CreateTestUser(t, db, username, nickname)
	adminRole := CreateTestAdminRole(t, db)
	if err := db.Model(user).Association("Roles").Append(adminRole); err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}
	return user
}

// AddUserToProject 添加用户到项目
func AddUserToProject(t *testing.T, db *gorm.DB, userID, projectID uint, role string) *model.ProjectMember {
	member := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	if err := db.Create(member).Error; err != nil {
		t.Fatalf("Failed to add user to project: %v", err)
	}
	return member
}

