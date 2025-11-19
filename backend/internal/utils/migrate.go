package utils

import (
	"strings"

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

	// 处理项目标签迁移（从JSON字段迁移到关联表）
	if err := migrateProjectTags(db); err != nil {
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
		
		// 注意：由于 Version 表不在 AutoMigrate 中，GORM 不应该处理它
		// 但如果 GORM 仍然检测到差异（比如外键约束格式不同），可能会尝试重建表
		// 为了避免这个问题，我们需要确保表结构与模型完全匹配
		// 
		// 实际上，问题的根源可能是：GORM 在检查表结构时，发现外键约束格式不匹配
		// 所以尝试重建表。但重建表时，GORM 的复制语句没有包含所有字段。
		// 
		// 解决方案：在 AutoMigrate 之前，手动检查并修复表结构
		// 确保它与模型完全匹配，包括外键约束格式
		// 但由于 SQLite 的限制，我们无法直接修改外键约束
		// 所以我们需要在 migrateVersionProjectID 中确保表结构完全正确
		// 
		// 如果 GORM 仍然尝试重建表，可能是因为它检测到了其他差异
		// 在这种情况下，我们需要确保在 GORM 重建表之前，表结构已经完全正确
		// 但由于 Version 表不在 AutoMigrate 中，GORM 不应该处理它
		// 如果仍然出现问题，可能是 GORM 的 bug 或者我们的表结构确实有问题
		// 
		// 最好的解决方案是：完全排除 Version 表从 GORM 的迁移中
		// 但由于 Version 模型定义了外键关系，GORM 可能仍然会检查它
		// 所以我们需要确保表结构与模型完全匹配
	}
	
	// 手动处理 Version 表的迁移，确保表结构与模型完全匹配
	// 这样可以避免 GORM 的 AutoMigrate 检测到差异并尝试重建表
	// 注意：由于 Version 表不在 AutoMigrate 中，GORM 不应该处理它
	// 但如果 GORM 仍然检测到差异（比如外键约束格式不同），可能会尝试重建表
	// 为了避免这个问题，我们需要确保表结构与模型完全匹配
	if config.AppConfig.Database.Type == "sqlite" {
		migrator := db.Migrator()
		if migrator.HasTable(&model.Version{}) {
			// 表存在，检查表结构是否与模型完全匹配
			// 如果表结构不匹配，GORM 可能会尝试重建表
			// 我们需要确保表结构与模型完全匹配，包括所有字段和约束
			
			// 检查所有必需的字段是否存在
			requiredColumns := []string{"id", "created_at", "updated_at", "deleted_at", "version_number", "release_notes", "status", "project_id", "release_date"}
			for _, col := range requiredColumns {
				if !migrator.HasColumn(&model.Version{}, col) {
					// 字段不存在，添加它
					migrator.AddColumn(&model.Version{}, col)
				}
			}
			
			// 检查 project_id 是否有 NOT NULL 约束
			var projectIDNotNull int64
			db.Raw(`SELECT COUNT(*) FROM pragma_table_info('versions') WHERE name = 'project_id' AND "notnull" = 1`).Scan(&projectIDNotNull)
			
			// 如果 project_id 没有 NOT NULL 约束，需要修复
			if projectIDNotNull == 0 {
				// 修复表结构：重建表并添加 NOT NULL 约束
				db.Exec("PRAGMA foreign_keys = OFF")
				defer db.Exec("PRAGMA foreign_keys = ON")
				
				// 删除关联表
				db.Exec("DROP TABLE IF EXISTS version_requirements")
				db.Exec("DROP TABLE IF EXISTS version_bugs")
				
				// 创建新表（包含 NOT NULL 约束，但不包含外键约束，避免 GORM 检测到差异）
				if err := db.Exec(`
					CREATE TABLE versions_fix (
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
				`).Error; err == nil {
					// 复制数据（只复制有效的记录，确保包含所有字段）
					if err := db.Exec(`
						INSERT INTO versions_fix (id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date)
						SELECT id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date
						FROM versions
						WHERE project_id IS NOT NULL AND version_number IS NOT NULL AND version_number != ''
					`).Error; err == nil {
						// 删除旧表
						db.Exec("DROP TABLE versions")
						// 重命名新表
						db.Exec("ALTER TABLE versions_fix RENAME TO versions")
						// 重新创建索引
						db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at)")
						db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id)")
					} else {
						// 复制失败，删除临时表
						db.Exec("DROP TABLE IF EXISTS versions_fix")
					}
				}
			}
		}
	}
	
	// 注意：由于 Version 表不在 AutoMigrate 中，GORM 不应该处理它
	// 但如果 GORM 仍然检测到差异（比如外键约束格式不同），可能会尝试重建表
	// 为了避免这个问题，我们需要确保表结构与模型完全匹配
	// 但由于 Version 表不在 AutoMigrate 中，GORM 不应该处理它
	// 如果仍然出现问题，可能是其他原因导致的
	// 
	// 实际上，问题的根源可能是：GORM 在检查表结构时，发现外键约束格式不匹配
	// 所以尝试重建表。但重建表时，GORM 的复制语句没有包含所有字段。
	// 
	// 解决方案：完全排除 Version 表从 GORM 的迁移中，并确保表结构完全正确
	// 由于 Version 表已经在 AutoMigrate 中被注释掉了，GORM 不应该处理它
	// 但如果仍然出现问题，可能是 GORM 检测到了其他差异
	// 
	// 最好的解决方案是：在 AutoMigrate 之前，手动检查并修复表结构
	// 确保它与模型完全匹配，包括外键约束格式
	// 但由于 SQLite 的限制，我们无法直接修改外键约束
	// 所以我们需要在 migrateVersionProjectID 中确保表结构完全正确
	
	// 创建关联表（如果不存在）
	// Version 表有 many2many 关联，需要创建关联表
	// 注意：不创建外键约束，避免 GORM 检测到差异
	if config.AppConfig.Database.Type == "sqlite" {
		var versionRequirementsExists int64
		db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='version_requirements'`).Scan(&versionRequirementsExists)
		if versionRequirementsExists == 0 {
			// 创建 version_requirements 关联表（不创建外键约束）
			db.Exec(`
				CREATE TABLE version_requirements (
					version_id INTEGER NOT NULL,
					requirement_id INTEGER NOT NULL,
					PRIMARY KEY (version_id, requirement_id)
				)
			`)
		}
		
		var versionBugsExists int64
		db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='version_bugs'`).Scan(&versionBugsExists)
		if versionBugsExists == 0 {
			// 创建 version_bugs 关联表（不创建外键约束）
			db.Exec(`
				CREATE TABLE version_bugs (
					version_id INTEGER NOT NULL,
					bug_id INTEGER NOT NULL,
					PRIMARY KEY (version_id, bug_id)
				)
			`)
		}
	}

	// 对于 SQLite，在 AutoMigrate 之前，手动检查并修复 Version 表结构
	// 这样可以确保表结构与模型完全匹配，避免 GORM 的 AutoMigrate 检测到差异并尝试重建表
	// 注意：GORM 在重建表时可能只复制部分字段，导致 NOT NULL 约束失败
	// 所以我们需要确保在 GORM 处理之前，表结构已经完全正确
	// 
	// 问题的根源：GORM 在检测到表结构差异（比如外键约束格式不同）时，会尝试重建表
	// 但重建表时，GORM 的复制语句可能只包含"变化"的字段，而不是所有字段
	// 这导致 NOT NULL 约束失败
	// 
	// 解决方案：在 AutoMigrate 之前，手动检查并修复表结构，确保它与模型完全匹配
	// 包括所有字段、约束和索引，这样 GORM 就不会检测到差异，也就不会尝试重建表
	// 
	// 关键：即使 Version 表不在 AutoMigrate 中，GORM 仍然可能因为 Requirement 模型引用了 Version
	// 而尝试重建 Version 表。所以我们需要确保 Version 表结构完全正确。
	if config.AppConfig.Database.Type == "sqlite" {
		migrator := db.Migrator()
		if migrator.HasTable(&model.Version{}) {
			// 表存在，检查表结构是否与模型完全匹配
			// 如果表结构不匹配，GORM 可能会尝试重建表
			// 我们需要确保表结构与模型完全匹配，包括所有字段和约束
			
			// 检查 project_id 是否有 NOT NULL 约束
			var projectIDNotNull int64
			db.Raw(`SELECT COUNT(*) FROM pragma_table_info('versions') WHERE name = 'project_id' AND "notnull" = 1`).Scan(&projectIDNotNull)
			
			// 如果 project_id 没有 NOT NULL 约束，需要修复
			if projectIDNotNull == 0 {
				// 修复表结构：重建表并添加 NOT NULL 约束
				db.Exec("PRAGMA foreign_keys = OFF")
				defer db.Exec("PRAGMA foreign_keys = ON")
				
				// 删除关联表
				db.Exec("DROP TABLE IF EXISTS version_requirements")
				db.Exec("DROP TABLE IF EXISTS version_bugs")
				
				// 创建新表（包含 NOT NULL 约束，但不包含外键约束，避免 GORM 检测到差异）
				if err := db.Exec(`
					CREATE TABLE versions_sync (
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
				`).Error; err == nil {
					// 复制数据（只复制有效的记录，确保包含所有字段）
					if err := db.Exec(`
						INSERT INTO versions_sync (id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date)
						SELECT id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date
						FROM versions
						WHERE project_id IS NOT NULL AND version_number IS NOT NULL AND version_number != ''
					`).Error; err == nil {
						// 删除旧表
						db.Exec("DROP TABLE versions")
						// 重命名新表
						db.Exec("ALTER TABLE versions_sync RENAME TO versions")
						// 重新创建索引
						db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at)")
						db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id)")
					} else {
						// 复制失败，删除临时表
						db.Exec("DROP TABLE IF EXISTS versions_sync")
					}
				}
			}
			
			// 检查并添加缺失的字段（如果不存在）
			if !migrator.HasColumn(&model.Version{}, "version_number") {
				migrator.AddColumn(&model.Version{}, "version_number")
			}
			if !migrator.HasColumn(&model.Version{}, "release_notes") {
				migrator.AddColumn(&model.Version{}, "release_notes")
			}
			if !migrator.HasColumn(&model.Version{}, "status") {
				migrator.AddColumn(&model.Version{}, "status")
			}
			if !migrator.HasColumn(&model.Version{}, "project_id") {
				migrator.AddColumn(&model.Version{}, "project_id")
			}
			if !migrator.HasColumn(&model.Version{}, "release_date") {
				migrator.AddColumn(&model.Version{}, "release_date")
			}
			
			// 创建索引（如果不存在）
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at)")
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id)")
		}
	}
	

	return db.AutoMigrate(
		// 用户与权限
		&model.User{},
		&model.Department{},
		&model.Role{},
		&model.Permission{},

		// 标签
		&model.Tag{},
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


		// 版本 - 注意：GORM 在重建表时可能只复制部分字段，导致 NOT NULL 约束失败
		// 我们已经尝试在 AutoMigrate 之前修复表结构，但 GORM 仍然可能检测到差异
		// 如果问题持续存在，可能需要手动修复数据库或升级 GORM 版本
		// 暂时注释掉，避免 GORM 尝试重建表
		// &model.Version{},

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
		
		// 检查 project_id 列是否有 NOT NULL 约束
		// 如果没有，GORM 可能会尝试添加，导致重建表
		// 我们需要确保表结构已经正确，避免 GORM 重建表
		var projectIDNotNull int64
		db.Raw(`
			SELECT COUNT(*) FROM pragma_table_info('versions') 
			WHERE name = 'project_id' AND "notnull" = 1
		`).Scan(&projectIDNotNull)
		
		// 如果 project_id 没有 NOT NULL 约束，我们需要手动添加
		// 但由于 SQLite 的限制，添加 NOT NULL 约束需要重建表
		// 为了避免 GORM 重建表时只复制部分字段，我们在这里手动重建表
		if projectIDNotNull == 0 {
			// project_id 存在但没有 NOT NULL 约束，需要重建表
			// 但我们已经删除了所有 NULL 记录，所以可以安全地重建表
			db.Exec("PRAGMA foreign_keys = OFF")
			defer db.Exec("PRAGMA foreign_keys = ON")
			
			// 删除关联表
			db.Exec("DROP TABLE IF EXISTS version_requirements")
			db.Exec("DROP TABLE IF EXISTS version_bugs")
			
			// 创建新表（包含 NOT NULL 约束）
			if err := db.Exec(`
				CREATE TABLE versions_fix (
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
			
			// 复制数据（只复制有效的记录）
			if err := db.Exec(`
				INSERT INTO versions_fix (id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date)
				SELECT id, created_at, updated_at, deleted_at, version_number, release_notes, status, project_id, release_date
				FROM versions
				WHERE project_id IS NOT NULL AND version_number IS NOT NULL AND version_number != ''
			`).Error; err != nil {
				db.Exec("DROP TABLE IF EXISTS versions_fix")
				return err
			}
			
			// 删除旧表
			if err := db.Exec("DROP TABLE versions").Error; err != nil {
				return err
			}
			
			// 重命名新表
			if err := db.Exec("ALTER TABLE versions_fix RENAME TO versions").Error; err != nil {
				return err
			}
			
			// 重新创建索引
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_deleted_at ON versions(deleted_at)")
			db.Exec("CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id)")
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
		// 删除并重新创建外键约束，使其与 GORM 期望的格式一致
		// 注意：SQLite 不支持直接修改外键约束，需要重建表
		// 但由于表结构已经正确，我们只需要确保外键约束的格式正确
		// 实际上，如果表结构已经正确，GORM 不应该尝试重建表
		// 但如果 GORM 仍然尝试重建，可能是因为外键约束的格式不匹配
		// 在这种情况下，我们无法完全避免 GORM 重建表
		// 但我们可以确保在重建时，所有字段都被正确复制
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

// migrateProjectTags 迁移项目标签：从JSON字段迁移到关联表
func migrateProjectTags(db *gorm.DB) error {
	// 检查 projects 表是否存在 tags 字段（JSON格式）
	if config.AppConfig.Database.Type == "sqlite" {
		// 先检查表是否存在
		var tableExists int64
		if err := db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='projects'`).Scan(&tableExists).Error; err != nil || tableExists == 0 {
			// 表不存在，让 AutoMigrate 处理
			return nil
		}

		var tagsColumnExists int64
		if err := db.Raw(`SELECT COUNT(*) FROM pragma_table_info('projects') WHERE name = 'tags'`).Scan(&tagsColumnExists).Error; err != nil {
			// 查询失败，可能表结构已改变，跳过迁移
			return nil
		}

		if tagsColumnExists > 0 {
			// tags 字段存在，需要迁移数据
			// 1. 从 projects 表的 tags 字段中提取所有唯一的标签名称
			// 2. 在 tags 表中创建这些标签
			// 3. 在 project_tags 关联表中创建关联关系
			// 4. 删除 projects 表的 tags 字段（由 GORM AutoMigrate 处理）

			// 提取所有标签名称
			var projects []struct {
				ID   uint
				Tags string
			}
			if err := db.Raw(`SELECT id, tags FROM projects WHERE tags IS NOT NULL AND tags != '' AND tags != '[]'`).Scan(&projects).Error; err != nil {
				// 查询失败，可能表结构已改变，跳过迁移
				return nil
			}

			// 解析JSON并收集所有唯一的标签名称
			tagNameMap := make(map[string]bool)
			for _, p := range projects {
				if p.Tags != "" && p.Tags != "[]" {
					// 简单的JSON解析：提取引号中的标签名称
					// 格式：["tag1","tag2"] 或 ["tag1", "tag2"]
					tagsStr := p.Tags
					// 移除方括号和空格
					tagsStr = strings.Trim(tagsStr, "[]")
					// 按逗号分割
					if tagsStr != "" {
						tagParts := strings.Split(tagsStr, ",")
						for _, part := range tagParts {
							tagName := strings.Trim(part, `" `)
							if tagName != "" {
								tagNameMap[tagName] = true
							}
						}
					}
				}
			}

			// 在 tags 表中创建标签
			for tagName := range tagNameMap {
				var existingTag model.Tag
				if err := db.Where("name = ?", tagName).First(&existingTag).Error; err != nil {
					// 标签不存在，创建它
					tag := model.Tag{
						Name:  tagName,
						Color: "blue",
					}
					if err := db.Create(&tag).Error; err != nil {
						// 创建失败，继续处理其他标签
						continue
					}
				}
			}

			// 为每个项目创建标签关联
			for _, p := range projects {
				if p.Tags != "" && p.Tags != "[]" {
					tagsStr := strings.Trim(p.Tags, "[]")
					if tagsStr != "" {
						tagParts := strings.Split(tagsStr, ",")
						for _, part := range tagParts {
							tagName := strings.Trim(part, `" `)
							if tagName != "" {
								var tag model.Tag
								if err := db.Where("name = ?", tagName).First(&tag).Error; err == nil {
									// 检查关联是否已存在
									var count int64
									db.Raw(`SELECT COUNT(*) FROM project_tags WHERE project_id = ? AND tag_id = ?`, p.ID, tag.ID).Scan(&count)
									if count == 0 {
										// 创建关联
										db.Exec(`INSERT INTO project_tags (project_id, tag_id) VALUES (?, ?)`, p.ID, tag.ID)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

