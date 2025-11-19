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

	// 处理 Version 表的 project_id 字段迁移（从 build_id 迁移）
	if err := migrateVersionProjectID(db); err != nil {
		return err
	}

	// 删除不再使用的 builds 表（如果存在）
	// 注意：必须在 Version 迁移之后，但在 AutoMigrate 之前
	if config.AppConfig.Database.Type == "sqlite" {
		var buildsTableExists int64
		if err := db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='builds'`).Scan(&buildsTableExists).Error; err == nil && buildsTableExists > 0 {
			// 先删除可能引用 builds 表的外键约束
			db.Exec("PRAGMA foreign_keys = OFF")
			defer db.Exec("PRAGMA foreign_keys = ON")
			// 删除关联表（如果存在）
			db.Exec("DROP TABLE IF EXISTS version_requirements")
			db.Exec("DROP TABLE IF EXISTS version_bugs")
			// 删除 builds 表
			db.Exec("DROP TABLE IF EXISTS builds")
		}
	}

	return db.AutoMigrate(
		// 用户与权限
		&model.User{},
		&model.Department{},
		&model.Role{},
		&model.Permission{},

		// 项目
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

		// 版本
		&model.Version{},

		// 测试
		&model.TestCase{},
		&model.TestCaseBug{},

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

// migrateVersionProjectID 迁移 Version 表的 project_id 字段
// 从 build_id 迁移到 project_id，SQLite 不支持直接添加 NOT NULL 列
func migrateVersionProjectID(db *gorm.DB) error {
	// 检查数据库类型
	if config.AppConfig.Database.Type != "sqlite" {
		// 不是 SQLite，让 AutoMigrate 处理
		return nil
	}

	// 检查 versions 表是否存在
	var tableExists int64
	err := db.Raw(`
		SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='versions'
	`).Scan(&tableExists).Error
	if err != nil || tableExists == 0 {
		// 表不存在，让 AutoMigrate 创建
		return nil
	}

	// 检查 project_id 列是否已存在
	var projectIDExists int64
	err = db.Raw(`
		SELECT COUNT(*) FROM pragma_table_info('versions') WHERE name = 'project_id'
	`).Scan(&projectIDExists).Error
	if err != nil {
		return nil
	}

	// 检查 build_id 列是否存在
	var buildIDExists int64
	err = db.Raw(`
		SELECT COUNT(*) FROM pragma_table_info('versions') WHERE name = 'build_id'
	`).Scan(&buildIDExists).Error
	if err != nil {
		return nil
	}

	// 如果 project_id 已存在，检查是否需要清理数据
	if projectIDExists > 0 {
		// 删除所有没有 project_id 的记录（确保数据完整性）
		if err := db.Exec("DELETE FROM `versions` WHERE `project_id` IS NULL").Error; err != nil {
			return nil // 忽略错误，可能表结构已改变
		}
		// 删除所有没有 version_number 的记录（确保数据完整性）
		if err := db.Exec("DELETE FROM `versions` WHERE `version_number` IS NULL OR `version_number` = ''").Error; err != nil {
			return nil // 忽略错误，可能表结构已改变
		}
		// 如果 build_id 还存在，需要删除它（通过重建表）
		if buildIDExists > 0 {
			// 先删除 builds 表（如果存在），因为它已经不再使用
			db.Exec("DROP TABLE IF EXISTS builds")
			// SQLite 不支持直接删除列，需要重建表
			// 注意：需要先删除外键约束和关联表，否则重建表会失败
			
			// 1. 禁用外键约束（SQLite 需要）
			db.Exec("PRAGMA foreign_keys = OFF")
			defer db.Exec("PRAGMA foreign_keys = ON")
			
			// 2. 删除关联表（如果存在）
			db.Exec("DROP TABLE IF EXISTS version_requirements")
			db.Exec("DROP TABLE IF EXISTS version_bugs")
			
			// 3. 创建新表（不包含 build_id，但包含所有其他字段）
			// 注意：不创建外键约束，让 GORM 的 AutoMigrate 来处理
			if err := db.Exec(`
				CREATE TABLE versions_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					created_at DATETIME,
					updated_at DATETIME,
					deleted_at DATETIME,
					version_number TEXT NOT NULL,
					release_notes TEXT,
					status TEXT DEFAULT 'draft',
					project_id INTEGER NOT NULL,
					release_date DATETIME
				)
			`).Error; err != nil {
				return err
			}
			// 4. 复制数据（只复制有 project_id 的记录，确保所有字段都有值）
			if err := db.Exec(`
				INSERT INTO versions_new (id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date)
				SELECT id, created_at, updated_at, deleted_at, 
				       COALESCE(version_number, '') as version_number,
				       COALESCE(release_notes, '') as release_notes,
				       COALESCE(status, 'draft') as status,
				       project_id,
				       release_date
				FROM versions
				WHERE project_id IS NOT NULL AND version_number IS NOT NULL AND version_number != ''
			`).Error; err != nil {
				return err
			}
			// 5. 删除旧表
			if err := db.Exec("DROP TABLE versions").Error; err != nil {
				return err
			}
			// 6. 重命名新表
			if err := db.Exec("ALTER TABLE versions_new RENAME TO versions").Error; err != nil {
				return err
			}
			// 7. 重新创建索引（GORM 会在 AutoMigrate 时处理外键和关联表）
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at)")
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id)")
		}
		// 确保表结构与模型完全匹配，避免 GORM 再次重建表
		// 检查并修复可能的字段类型不匹配
		return nil
	}

	// project_id 不存在，需要添加
	// 先添加可空的 project_id 列
	if err := db.Exec("ALTER TABLE `versions` ADD COLUMN `project_id` integer").Error; err != nil {
		// 如果添加失败（可能列已存在），继续
		return nil
	}

	// 由于 build_id 关联的 builds 表已不存在，现有版本记录无法自动迁移到 project_id
	// 删除所有没有 project_id 的记录（因为无法确定它们属于哪个项目）
	// 这些记录无法在新系统中使用，因为版本必须关联项目
	if err := db.Exec("DELETE FROM `versions` WHERE `project_id` IS NULL").Error; err != nil {
		return err
	}

	// 注意：GORM 的 AutoMigrate 会尝试添加 NOT NULL 约束
	// 在 SQLite 中，这需要重建表。由于我们已经删除了所有 NULL 记录，
	// GORM 应该能够成功重建表并添加约束
	// 如果仍然失败，可能需要手动重建表（但通常不需要）

	return nil
}

