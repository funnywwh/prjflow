package utils

import (
	"project-management/internal/config"
	"project-management/internal/model"

	"gorm.io/gorm"
)

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB) error {
	// 先清理可能存在的临时表（GORM 重建表失败时留下的）
	// 这对于 SQLite 特别重要，因为 GORM 在重建表时可能只复制部分字段，导致失败
	if config.AppConfig.Database.Type == "sqlite" {
		cleanupTemporaryTables(db)
		// 迁移 modules 表：移除 project_id 字段
		migrateModuleTable(db)
	}

	// 执行 AutoMigrate
	err := db.AutoMigrate(
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
		// 功能模块
		&model.Module{},

		// 需求与Bug
		&model.Requirement{},
		&model.Bug{},
		&model.BugAssignee{},

		// 任务与看板
		&model.Task{},
		&model.TaskDependency{},
		&model.Board{},
		&model.BoardColumn{},

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
	
	// AutoMigrate 之后，再次清理可能产生的临时表
	if config.AppConfig.Database.Type == "sqlite" {
		cleanupTemporaryTables(db)
	}
	
	return err
}

// cleanupTemporaryTables 清理所有 GORM 重建表失败时留下的临时表
// GORM 在重建表时会创建 `表名__temp` 格式的临时表
// 如果重建失败，这些临时表可能会残留，导致后续迁移失败
func cleanupTemporaryTables(db *gorm.DB) {
	if config.AppConfig.Database.Type != "sqlite" {
		return
	}

	// 查询所有以 __temp 结尾的表（GORM 创建的临时表）
	var tempTables []string
	db.Raw(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name LIKE '%__temp'
	`).Scan(&tempTables)

	// 删除所有临时表
	for _, tableName := range tempTables {
		db.Exec("DROP TABLE IF EXISTS `" + tableName + "`")
	}
}

// migrateModuleTable 迁移 modules 表：移除 project_id 字段
// 功能模块改为系统资源，不再属于项目
func migrateModuleTable(db *gorm.DB) {
	if config.AppConfig.Database.Type != "sqlite" {
		return
	}

	// 检查 modules 表是否存在 project_id 字段
	var columns []struct {
		Name string
	}
	db.Raw("PRAGMA table_info(modules)").Scan(&columns)
	
	hasProjectID := false
	for _, col := range columns {
		if col.Name == "project_id" {
			hasProjectID = true
			break
		}
	}

	// 如果存在 project_id 字段，需要迁移
	if hasProjectID {
		// 1. 创建新表（不包含 project_id）
		db.Exec(`
			CREATE TABLE IF NOT EXISTS modules_new (
				id integer PRIMARY KEY AUTOINCREMENT,
				created_at datetime,
				updated_at datetime,
				deleted_at datetime,
				name text NOT NULL,
				code text,
				description text,
				status integer DEFAULT 1,
				sort integer DEFAULT 0
			)
		`)

		// 2. 复制数据（忽略 project_id）
		db.Exec(`
			INSERT INTO modules_new (id, created_at, updated_at, deleted_at, name, code, description, status, sort)
			SELECT id, created_at, updated_at, deleted_at, name, code, description, status, sort
			FROM modules
		`)

		// 3. 删除旧表
		db.Exec("DROP TABLE modules")

		// 4. 重命名新表
		db.Exec("ALTER TABLE modules_new RENAME TO modules")

		// 5. 创建索引（如果不存在）
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_modules_name ON modules(name)")
		db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_modules_code ON modules(code)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_modules_deleted_at ON modules(deleted_at)")
	}
}

