package utils

import (
	"project-management/internal/config"
	"project-management/internal/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB) error {
	// 先处理 User 表的 nickname 字段迁移（特殊处理）
	if err := migrateUserNickname(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		// 用户与权限
		&model.User{},
		&model.Department{},
		&model.Role{},
		&model.Permission{},

		// 产品与项目
		&model.ProductLine{},
		&model.Product{},
		&model.ProjectGroup{},
		&model.Project{},
		&model.ProjectMember{},

		// 需求与Bug
		&model.Requirement{},
		&model.Bug{},
		&model.BugAssignee{},

		// 任务与看板
		&model.Task{},
		&model.TaskDependency{},
		&model.Board{},
		&model.BoardColumn{},

		// 计划与执行
		&model.Plan{},
		&model.PlanExecution{},

		// 版本与构建
		&model.Build{},
		&model.Version{},

		// 测试
		&model.TestCase{},
		&model.TestReport{},
		&model.TestCaseBug{},
		&model.TestCaseReport{},

		// 资源管理
		&model.Resource{},
		&model.ResourceAllocation{},

		// 工作报告
		&model.DailyReport{},
		&model.WeeklyReport{},

		// 插件管理
		&model.Plugin{},
		&model.PluginConfig{},
		&model.PluginHook{},
		&model.PluginRoute{},

		// 关系图
		&model.EntityRelation{},

		// 工作台
		&model.UserDashboard{},

		// 系统配置
		&model.SystemConfig{},
	)
}

// migrateUserNickname 迁移 User 表的 nickname 字段
// SQLite 不支持直接添加 NOT NULL 列，需要先添加可空列，更新数据
// 注意：由于 SQLite 的限制，nickname 列在数据库中可能是可空的，
// 但我们在应用层保证它不为空（通过验证和默认值）
func migrateUserNickname(db *gorm.DB) error {
	// 检查数据库类型
	if config.AppConfig.Database.Type != "sqlite" {
		// 不是 SQLite，让 AutoMigrate 处理
		return nil
	}

	// 检查 nickname 列是否已存在
	var count int64
	err := db.Raw(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = 'nickname'
	`).Scan(&count).Error
	
	if err != nil {
		// 查询失败，让 AutoMigrate 处理
		return nil
	}

	// 如果列已存在，更新现有记录（确保所有记录都有 nickname）
	if count > 0 {
		// 更新现有记录：将 nickname 设置为 username（如果为空）
		if err := db.Exec("UPDATE `users` SET `nickname` = `username` WHERE `nickname` IS NULL OR `nickname` = ''").Error; err != nil {
			return err
		}
		return nil
	}

	// 列不存在，先添加可空列
	if err := db.Exec("ALTER TABLE `users` ADD COLUMN `nickname` TEXT").Error; err != nil {
		// 如果添加失败（可能列已存在），继续
		return nil
	}

	// 更新现有记录：将 nickname 设置为 username
	if err := db.Exec("UPDATE `users` SET `nickname` = `username` WHERE `nickname` IS NULL OR `nickname` = ''").Error; err != nil {
		return err
	}

	return nil
}

