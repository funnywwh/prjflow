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

	// 初始化默认权限和角色
	if err := initDefaultPermissionsAndRoles(db); err != nil {
		return err
	}

	return err
}

// initDefaultPermissionsAndRoles 初始化默认权限和角色
func initDefaultPermissionsAndRoles(db *gorm.DB) error {
	// 定义默认权限（包括菜单权限）
	defaultPermissions := []model.Permission{
		// 工作台（菜单，无权限要求）
		{Code: "dashboard", Name: "工作台", Resource: "dashboard", Action: "read", Description: "工作台", Status: 1, IsMenu: true, MenuPath: "/dashboard", MenuIcon: "DashboardOutlined", MenuTitle: "工作台", MenuOrder: 0},

		// 项目管理权限（操作权限）
		{Code: "project:create", Name: "创建项目", Resource: "project", Action: "create", Description: "创建新项目", Status: 1},
		{Code: "project:read", Name: "查看项目", Resource: "project", Action: "read", Description: "查看项目信息", Status: 1},
		{Code: "project:update", Name: "更新项目", Resource: "project", Action: "update", Description: "更新项目信息", Status: 1},
		{Code: "project:delete", Name: "删除项目", Resource: "project", Action: "delete", Description: "删除项目", Status: 1},
		{Code: "project:manage", Name: "管理项目", Resource: "project", Action: "manage", Description: "管理项目成员和设置", Status: 1},

		// 项目管理菜单（父菜单）
		{Code: "project-management", Name: "项目管理", Resource: "project", Action: "read", Description: "项目管理", Status: 1, IsMenu: true, MenuIcon: "ProjectOutlined", MenuTitle: "项目管理", MenuOrder: 1},
		// 项目列表（子菜单）
		{Code: "project:list", Name: "项目列表", Resource: "project", Action: "read", Description: "项目列表", Status: 1, IsMenu: true, MenuPath: "/project", MenuTitle: "项目列表", MenuOrder: 0},
		// 需求管理（子菜单）
		{Code: "requirement:menu", Name: "需求管理", Resource: "requirement", Action: "menu", Description: "需求管理菜单", Status: 1, IsMenu: true, MenuPath: "/requirement", MenuTitle: "需求管理", MenuOrder: 1},
		// 任务管理（子菜单）
		{Code: "task:read", Name: "查看任务", Resource: "task", Action: "read", Description: "查看任务信息", Status: 1, IsMenu: true, MenuPath: "/task", MenuTitle: "任务管理", MenuOrder: 2},

		// 测试管理 (父菜单，新建)
		{Code: "test-management", Name: "测试管理", Resource: "test", Action: "read", Description: "测试管理", Status: 1, IsMenu: true, MenuIcon: "ExperimentOutlined", MenuTitle: "测试管理", MenuOrder: 2},
		// 测试单管理（子菜单，将移动到测试管理下）
		{Code: "test-case:read", Name: "查看测试用例", Resource: "testcase", Action: "read", Description: "查看测试用例", Status: 1, IsMenu: true, MenuPath: "/test-case", MenuTitle: "测试单管理", MenuOrder: 0},
		// Bug管理（子菜单，将移动到测试管理下）
		{Code: "bug:read", Name: "查看Bug", Resource: "bug", Action: "read", Description: "查看Bug信息", Status: 1, IsMenu: true, MenuPath: "/bug", MenuTitle: "Bug管理", MenuOrder: 1},
		// 版本管理（子菜单，将移动到测试管理下）
		{Code: "version:read", Name: "查看版本", Resource: "version", Action: "read", Description: "查看版本信息", Status: 1, IsMenu: true, MenuPath: "/version", MenuTitle: "版本管理", MenuOrder: 2},

		// 需求管理权限（操作权限）
		{Code: "requirement:read", Name: "查看需求", Resource: "requirement", Action: "read", Description: "查看需求信息", Status: 1},
		{Code: "requirement:create", Name: "创建需求", Resource: "requirement", Action: "create", Description: "创建新需求", Status: 1},
		{Code: "requirement:update", Name: "更新需求", Resource: "requirement", Action: "update", Description: "更新需求信息", Status: 1},
		{Code: "requirement:delete", Name: "删除需求", Resource: "requirement", Action: "delete", Description: "删除需求", Status: 1},

		// Bug管理权限（操作权限）
		{Code: "bug:create", Name: "创建Bug", Resource: "bug", Action: "create", Description: "创建新Bug", Status: 1},
		{Code: "bug:update", Name: "更新Bug", Resource: "bug", Action: "update", Description: "更新Bug信息", Status: 1},
		{Code: "bug:delete", Name: "删除Bug", Resource: "bug", Action: "delete", Description: "删除Bug", Status: 1},
		{Code: "bug:assign", Name: "分配Bug", Resource: "bug", Action: "assign", Description: "分配Bug给处理人", Status: 1},

		// 任务管理权限（操作权限）
		{Code: "task:create", Name: "创建任务", Resource: "task", Action: "create", Description: "创建新任务", Status: 1},
		{Code: "task:update", Name: "更新任务", Resource: "task", Action: "update", Description: "更新任务信息", Status: 1},
		{Code: "task:delete", Name: "删除任务", Resource: "task", Action: "delete", Description: "删除任务", Status: 1},

		// 资源管理菜单（父菜单）
		{Code: "resource-management", Name: "资源管理", Resource: "resource", Action: "read", Description: "资源管理", Status: 1, IsMenu: true, MenuIcon: "TeamOutlined", MenuTitle: "资源管理", MenuOrder: 3},
		// 资源统计（子菜单）
		{Code: "resource:read", Name: "查看资源", Resource: "resource", Action: "read", Description: "查看资源统计", Status: 1, IsMenu: true, MenuPath: "/resource/statistics", MenuTitle: "资源统计", MenuOrder: 0},
		// 资源管理权限（操作权限）
		{Code: "resource:manage", Name: "管理资源", Resource: "resource", Action: "manage", Description: "管理资源分配", Status: 1},

		// 系统管理菜单（父菜单）
		{Code: "system-management", Name: "系统管理", Resource: "system", Action: "read", Description: "系统管理", Status: 1, IsMenu: true, MenuIcon: "SettingOutlined", MenuTitle: "系统管理", MenuOrder: 4},
		// 用户管理（子菜单）
		{Code: "user:menu", Name: "用户管理", Resource: "user", Action: "menu", Description: "用户管理菜单", Status: 1, IsMenu: true, MenuPath: "/user", MenuTitle: "用户管理", MenuOrder: 0},
		// 部门管理（子菜单）
		{Code: "department:read", Name: "查看部门", Resource: "department", Action: "read", Description: "查看部门信息", Status: 1, IsMenu: true, MenuPath: "/department", MenuTitle: "部门管理", MenuOrder: 1},
		// 权限管理（子菜单）
		{Code: "permission:manage", Name: "管理权限", Resource: "permission", Action: "manage", Description: "管理角色和权限", Status: 1, IsMenu: true, MenuPath: "/permission", MenuTitle: "权限管理", MenuOrder: 2},
		// 微信设置（子菜单）
		{Code: "wechat:settings", Name: "微信设置", Resource: "wechat", Action: "settings", Description: "微信配置设置", Status: 1, IsMenu: true, MenuPath: "/system/wechat-settings", MenuTitle: "微信设置", MenuOrder: 3},

		// 用户管理权限（操作权限）
		{Code: "user:read", Name: "查看用户", Resource: "user", Action: "read", Description: "查看用户信息", Status: 1},
		{Code: "user:create", Name: "创建用户", Resource: "user", Action: "create", Description: "创建新用户", Status: 1},
		{Code: "user:update", Name: "更新用户", Resource: "user", Action: "update", Description: "更新用户信息", Status: 1},
		{Code: "user:delete", Name: "删除用户", Resource: "user", Action: "delete", Description: "删除用户", Status: 1},

		// 部门管理权限（操作权限）
		{Code: "department:create", Name: "创建部门", Resource: "department", Action: "create", Description: "创建新部门", Status: 1},
		{Code: "department:update", Name: "更新部门", Resource: "department", Action: "update", Description: "更新部门信息", Status: 1},
		{Code: "department:delete", Name: "删除部门", Resource: "department", Action: "delete", Description: "删除部门", Status: 1},
	}

	// 创建或更新权限
	permMap := make(map[string]*model.Permission) // 用于存储权限代码到权限的映射
	for i := range defaultPermissions {
		perm := &defaultPermissions[i]
		var existingPerm model.Permission
		if err := db.Where("code = ?", perm.Code).First(&existingPerm).Error; err != nil {
			// 权限不存在，创建
			if err := db.Create(perm).Error; err != nil {
				return err
			}
			permMap[perm.Code] = perm
		} else {
			// 权限已存在，更新（保留现有ID）
			perm.ID = existingPerm.ID
			// 更新菜单相关字段
			updates := map[string]interface{}{
				"is_menu":    perm.IsMenu,
				"menu_path":  perm.MenuPath,
				"menu_icon":  perm.MenuIcon,
				"menu_title": perm.MenuTitle,
				"menu_order": perm.MenuOrder,
			}
			if err := db.Model(&existingPerm).Updates(updates).Error; err != nil {
				return err
			}
			// 重新加载以获取最新数据
			db.First(&existingPerm, existingPerm.ID)
			permMap[perm.Code] = &existingPerm
		}
	}

	// 设置父子菜单关系
	// 项目管理菜单的子菜单
	if projectManagement, ok := permMap["project-management"]; ok {
		parentID := projectManagement.ID
		// 项目列表
		if projectList, ok := permMap["project:list"]; ok {
			projectList.ParentMenuID = &parentID
			db.Model(projectList).Select("parent_menu_id").Updates(projectList)
		}
		// 需求管理
		if requirementMenu, ok := permMap["requirement:menu"]; ok {
			requirementMenu.ParentMenuID = &parentID
			db.Model(requirementMenu).Select("parent_menu_id").Updates(requirementMenu)
		}
		// 任务管理
		if taskRead, ok := permMap["task:read"]; ok {
			taskRead.ParentMenuID = &parentID
			db.Model(taskRead).Select("parent_menu_id").Updates(taskRead)
		}
	}

	// 测试管理菜单的子菜单
	if testManagement, ok := permMap["test-management"]; ok {
		parentID := testManagement.ID
		// 测试管理（子菜单）
		if testCaseRead, ok := permMap["test-case:read"]; ok {
			db.Model(testCaseRead).Select("parent_menu_id").Updates(map[string]interface{}{"parent_menu_id": &parentID})
		}
		// Bug管理
		if bugRead, ok := permMap["bug:read"]; ok {
			db.Model(bugRead).Select("parent_menu_id").Updates(map[string]interface{}{"parent_menu_id": &parentID})
		}
		// 版本管理
		if versionRead, ok := permMap["version:read"]; ok {
			db.Model(versionRead).Select("parent_menu_id").Updates(map[string]interface{}{"parent_menu_id": &parentID})
		}
	}

	// 资源管理菜单的子菜单
	if resourceManagement, ok := permMap["resource-management"]; ok {
		parentID := resourceManagement.ID
		if resourceRead, ok := permMap["resource:read"]; ok {
			resourceRead.ParentMenuID = &parentID
			db.Model(resourceRead).Select("parent_menu_id").Updates(resourceRead)
		}
	}

	// 系统管理菜单的子菜单
	if systemManagement, ok := permMap["system-management"]; ok {
		parentID := systemManagement.ID
		if userMenu, ok := permMap["user:menu"]; ok {
			userMenu.ParentMenuID = &parentID
			db.Model(userMenu).Select("parent_menu_id").Updates(userMenu)
		}
		if departmentRead, ok := permMap["department:read"]; ok {
			departmentRead.ParentMenuID = &parentID
			db.Model(departmentRead).Select("parent_menu_id").Updates(departmentRead)
		}
		if permissionManage, ok := permMap["permission:manage"]; ok {
			permissionManage.ParentMenuID = &parentID
			db.Model(permissionManage).Select("parent_menu_id").Updates(permissionManage)
		}
		if wechatSettings, ok := permMap["wechat:settings"]; ok {
			wechatSettings.ParentMenuID = &parentID
			db.Model(wechatSettings).Select("parent_menu_id").Updates(wechatSettings)
		}
	}

	// 创建管理员角色（如果不存在）
	var adminRole model.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = model.Role{
			Name:        "管理员",
			Code:        "admin",
			Description: "系统管理员，拥有所有权限",
			Status:      1,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return err
		}
	}

	// 为管理员角色分配所有权限
	var allPermissions []model.Permission
	if err := db.Find(&allPermissions).Error; err != nil {
		return err
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		return err
	}

	return nil
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
