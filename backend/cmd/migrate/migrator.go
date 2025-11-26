package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite" // 纯Go SQLite驱动

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"project-management/internal/model"
	"project-management/internal/utils"
)

// Migrator 迁移器
type Migrator struct {
	zenTaoDB    *gorm.DB
	goProjectDB *gorm.DB

	// ID映射表
	deptIDMap        map[int]uint
	roleIDMap        map[int]uint
	userIDMap        map[int]uint
	projectIDMap     map[int]uint
	requirementIDMap map[int]uint
	moduleIDMap      map[int]uint

	// 统计信息
	stats struct {
		taskCount          int
		bugCount           int
		projectMemberCount int
		moduleCount        int
	}
}

// NewMigrator 创建迁移器
func NewMigrator(config *MigrateConfig) (*Migrator, error) {
	m := &Migrator{
		deptIDMap:        make(map[int]uint),
		roleIDMap:        make(map[int]uint),
		userIDMap:        make(map[int]uint),
		projectIDMap:     make(map[int]uint),
		requirementIDMap: make(map[int]uint),
		moduleIDMap:      make(map[int]uint),
	}

	// 连接zentao数据库
	zenTaoDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.ZenTao.User,
		config.ZenTao.Password,
		config.ZenTao.Host,
		config.ZenTao.Port,
		config.ZenTao.DBName,
	)

	var err error
	m.zenTaoDB, err = gorm.Open(mysql.Open(zenTaoDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接zentao数据库失败: %w", err)
	}

	// 连接goproject数据库
	var dialector gorm.Dialector
	if config.GoProject.Type == "sqlite" {
		dialector = sqlite.Open(config.GoProject.DSN)
	} else {
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.GoProject.Type)
	}

	m.goProjectDB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接goproject数据库失败: %w", err)
	}

	return m, nil
}

// MigrateAll 执行所有迁移
func (m *Migrator) MigrateAll() error {
	log.Println("==========================================")
	log.Println("开始数据迁移...")
	log.Println("==========================================")

	// 1. 迁移部门
	if err := m.MigrateDepartments(); err != nil {
		return fmt.Errorf("迁移部门失败: %w", err)
	}

	// 2. 迁移角色和权限
	if err := m.MigrateRoles(); err != nil {
		return fmt.Errorf("迁移角色失败: %w", err)
	}

	// 3. 迁移用户
	if err := m.MigrateUsers(); err != nil {
		return fmt.Errorf("迁移用户失败: %w", err)
	}

	// 4. 迁移项目
	if err := m.MigrateProjects(); err != nil {
		return fmt.Errorf("迁移项目失败: %w", err)
	}

	// 5. 迁移项目模块
	if err := m.MigrateModules(); err != nil {
		return fmt.Errorf("迁移项目模块失败: %w", err)
	}

	// 6. 迁移需求
	if err := m.MigrateRequirements(); err != nil {
		return fmt.Errorf("迁移需求失败: %w", err)
	}

	// 7. 迁移任务
	if err := m.MigrateTasks(); err != nil {
		return fmt.Errorf("迁移任务失败: %w", err)
	}

	// 8. 迁移Bug
	if err := m.MigrateBugs(); err != nil {
		return fmt.Errorf("迁移Bug失败: %w", err)
	}

	// 9. 迁移项目成员
	if err := m.MigrateProjectMembers(); err != nil {
		return fmt.Errorf("迁移项目成员失败: %w", err)
	}

	log.Println("==========================================")
	log.Println("数据迁移完成！")
	log.Println("==========================================")
	log.Printf("迁移统计:")
	log.Printf("  - 部门: %d 个", len(m.deptIDMap))
	log.Printf("  - 角色: %d 个", len(m.roleIDMap))
	log.Printf("  - 用户: %d 个", len(m.userIDMap))
	log.Printf("  - 项目: %d 个", len(m.projectIDMap))
	log.Printf("  - 需求: %d 个", len(m.requirementIDMap))
	log.Printf("  - 任务: %d 个", m.stats.taskCount)
	log.Printf("  - Bug: %d 个", m.stats.bugCount)
	log.Printf("  - 项目成员: %d 个", m.stats.projectMemberCount)
	log.Printf("  - 项目模块: %d 个", m.stats.moduleCount)
	log.Println("==========================================")

	// 更新初始化状态为已初始化
	log.Println("更新初始化状态...")
	initConfig := model.SystemConfig{
		Key:   "initialized",
		Value: "true",
		Type:  "boolean",
	}
	if err := m.goProjectDB.Where("key = ?", "initialized").Assign(model.SystemConfig{Value: "true", Type: "boolean"}).FirstOrCreate(&initConfig).Error; err != nil {
		return fmt.Errorf("更新初始化状态失败: %w", err)
	}
	log.Println("初始化状态已更新为已完成")

	return nil
}

// MigrateDepartments 迁移部门
func (m *Migrator) MigrateDepartments() error {
	log.Println("开始迁移部门...")

	type ZenTaoDept struct {
		ID     int    `gorm:"column:id"`
		Name   string `gorm:"column:name"`
		Parent int    `gorm:"column:parent"`
		Grade  int    `gorm:"column:grade"`
		Order  int    `gorm:"column:order"`
	}

	var zentaoDepts []ZenTaoDept
	if err := m.zenTaoDB.Table("zt_dept").Order("grade ASC, `order` ASC").Find(&zentaoDepts).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个部门", len(zentaoDepts))

	// 构建部门映射表，用于快速查找
	deptMap := make(map[int]*ZenTaoDept)
	for i := range zentaoDepts {
		deptMap[zentaoDepts[i].ID] = &zentaoDepts[i]
	}

	// 按层级排序，确保父部门先于子部门创建
	// 使用拓扑排序：先处理没有父部门的，再处理有父部门的
	sortedDepts := make([]*ZenTaoDept, 0, len(zentaoDepts))
	processed := make(map[int]bool)

	// 第一轮：添加所有根部门（parent=0或parent不在列表中的）
	for i := range zentaoDepts {
		zd := &zentaoDepts[i]
		if zd.Parent == 0 || deptMap[zd.Parent] == nil {
			sortedDepts = append(sortedDepts, zd)
			processed[zd.ID] = true
		}
	}

	// 后续轮次：添加父部门已处理的部门
	maxIterations := len(zentaoDepts) // 防止无限循环
	for iteration := 0; iteration < maxIterations && len(processed) < len(zentaoDepts); iteration++ {
		for i := range zentaoDepts {
			zd := &zentaoDepts[i]
			if processed[zd.ID] {
				continue
			}
			// 如果父部门已处理，则添加当前部门
			if zd.Parent == 0 || processed[zd.Parent] {
				sortedDepts = append(sortedDepts, zd)
				processed[zd.ID] = true
			}
		}
	}

	// 如果还有未处理的部门，按原顺序添加（可能是数据问题）
	for i := range zentaoDepts {
		zd := &zentaoDepts[i]
		if !processed[zd.ID] {
			sortedDepts = append(sortedDepts, zd)
			log.Printf("警告: 部门 %s (ID: %d) 的父部门可能不存在，将按原顺序处理", zd.Name, zd.ID)
		}
	}

	log.Printf("部门排序完成，共 %d 个部门，将按层级顺序迁移", len(sortedDepts))

	// 按排序后的顺序迁移部门
	for _, zd := range sortedDepts {
		var parentID *uint
		if zd.Parent > 0 {
			if newID, ok := m.deptIDMap[zd.Parent]; ok {
				parentID = &newID
			} else {
				log.Printf("警告: 部门 %s (ID: %d) 的父部门 (ID: %d) 未找到，将作为根部门处理", zd.Name, zd.ID, zd.Parent)
			}
		}

		dept := model.Department{
			Name:     zd.Name,
			Code:     GenerateDeptCode(zd.Name, zd.ID),
			ParentID: parentID,
			Level:    zd.Grade,
			Sort:     zd.Order,
			Status:   1,
		}

		// 检查是否已存在（基于code）
		var existing model.Department
		if err := m.goProjectDB.Where("code = ?", dept.Code).First(&existing).Error; err == nil {
			m.deptIDMap[zd.ID] = existing.ID
			log.Printf("部门已存在: %s (ID: %d -> %d, 父部门: %v)", dept.Name, zd.ID, existing.ID, parentID)
			continue
		}

		if err := m.goProjectDB.Create(&dept).Error; err != nil {
			log.Printf("创建部门失败: %s, 错误: %v", dept.Name, err)
			continue
		}

		m.deptIDMap[zd.ID] = dept.ID
		parentInfo := "根部门"
		if parentID != nil {
			parentInfo = fmt.Sprintf("父部门ID: %d", *parentID)
		}
		log.Printf("迁移部门: %s (ID: %d -> %d, 层级: %d, %s)", dept.Name, zd.ID, dept.ID, zd.Grade, parentInfo)
	}

	log.Printf("部门迁移完成，共迁移 %d 个部门", len(m.deptIDMap))
	return nil
}

// MigrateRoles 迁移角色
func (m *Migrator) MigrateRoles() error {
	log.Println("开始迁移角色...")

	type ZenTaoGroup struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Desc string `gorm:"column:desc"`
	}

	type ZenTaoGroupPriv struct {
		Group  int    `gorm:"column:group"`
		Module string `gorm:"column:module"`
		Method string `gorm:"column:method"`
	}

	var zentaoGroups []ZenTaoGroup
	if err := m.zenTaoDB.Table("zt_group").Find(&zentaoGroups).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个角色组", len(zentaoGroups))

	// 先获取goproject的默认admin角色
	var adminRole model.Role
	if err := m.goProjectDB.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		log.Println("警告: 未找到admin角色，将创建")
		adminRole = model.Role{
			Name:        "管理员",
			Code:        "admin",
			Description: "系统管理员，拥有所有权限",
			Status:      1,
		}
		if err := m.goProjectDB.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("创建admin角色失败: %w", err)
		}
	}

	// 获取所有权限
	var allPermissions []model.Permission
	if err := m.goProjectDB.Find(&allPermissions).Error; err != nil {
		return fmt.Errorf("获取权限列表失败: %w", err)
	}

	// 创建权限代码到权限的映射
	permMap := make(map[string]*model.Permission)
	for i := range allPermissions {
		permMap[allPermissions[i].Code] = &allPermissions[i]
	}

	// 迁移角色
	for _, zg := range zentaoGroups {
		// 获取该角色的权限
		var groupPrivs []ZenTaoGroupPriv
		m.zenTaoDB.Table("zt_grouppriv").Where("`group` = ?", zg.ID).Find(&groupPrivs)

		// 判断是否是管理员角色（根据名称或权限数量）
		isAdmin := false
		roleName := strings.ToLower(zg.Name)
		if strings.Contains(roleName, "admin") || strings.Contains(roleName, "管理员") || strings.Contains(roleName, "管理") {
			isAdmin = true
		}

		if isAdmin {
			// 使用默认的admin角色
			m.roleIDMap[zg.ID] = adminRole.ID
			log.Printf("角色映射到admin: %s (ID: %d -> %d)", zg.Name, zg.ID, adminRole.ID)
			continue
		}

		// 创建新角色
		roleCode := GenerateRoleCode(zg.Name)
		role := model.Role{
			Name:        zg.Name,
			Code:        roleCode,
			Description: zg.Desc,
			Status:      1,
		}

		// 检查是否已存在
		var existing model.Role
		if err := m.goProjectDB.Where("code = ?", roleCode).First(&existing).Error; err == nil {
			m.roleIDMap[zg.ID] = existing.ID
			log.Printf("角色已存在: %s (ID: %d -> %d)", role.Name, zg.ID, existing.ID)
			continue
		}

		if err := m.goProjectDB.Create(&role).Error; err != nil {
			log.Printf("创建角色失败: %s, 错误: %v", role.Name, err)
			continue
		}

		// 映射权限
		var rolePerms []*model.Permission
		for _, gp := range groupPrivs {
			permCode := MapZenTaoPermissionToGoProject(gp.Module, gp.Method)
			if permCode != "" {
				if perm, ok := permMap[permCode]; ok {
					rolePerms = append(rolePerms, perm)
				}
			}
		}

		if len(rolePerms) > 0 {
			if err := m.goProjectDB.Model(&role).Association("Permissions").Replace(rolePerms); err != nil {
				log.Printf("分配权限失败: %s, 错误: %v", role.Name, err)
			}
		}

		m.roleIDMap[zg.ID] = role.ID
		log.Printf("迁移角色: %s (ID: %d -> %d, 权限数: %d)", role.Name, zg.ID, role.ID, len(rolePerms))
	}

	log.Printf("角色迁移完成，共迁移 %d 个角色", len(m.roleIDMap))
	return nil
}

// MigrateUsers 迁移用户
func (m *Migrator) MigrateUsers() error {
	log.Println("开始迁移用户...")

	type ZenTaoUser struct {
		ID       int    `gorm:"column:id"`
		Account  string `gorm:"column:account"`
		Realname string `gorm:"column:realname"`
		Email    string `gorm:"column:email"`
		Mobile   string `gorm:"column:mobile"`
		Avatar   string `gorm:"column:avatar"`
		Dept     int    `gorm:"column:dept"`
		Role     string `gorm:"column:role"`
		Deleted  string `gorm:"column:deleted"`
	}

	var zentaoUsers []ZenTaoUser
	if err := m.zenTaoDB.Table("zt_user").Find(&zentaoUsers).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个用户", len(zentaoUsers))

	// 生成默认密码哈希
	defaultPassword, err := utils.HashPassword("123")
	if err != nil {
		return fmt.Errorf("生成默认密码失败: %w", err)
	}

	// 获取admin角色
	var adminRole model.Role
	if err := m.goProjectDB.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return fmt.Errorf("未找到admin角色: %w", err)
	}

	for _, zu := range zentaoUsers {
		// 跳过已删除的用户（如果需要）
		// if zu.Deleted == "1" {
		// 	continue
		// }

		var deptID *uint
		if zu.Dept > 0 {
			if newID, ok := m.deptIDMap[zu.Dept]; ok {
				deptID = &newID
			}
		}

		nickname := zu.Realname
		if nickname == "" {
			nickname = zu.Account
		}

		user := model.User{
			Username:     zu.Account,
			Nickname:     nickname,
			Email:        zu.Email,
			Phone:        zu.Mobile,
			Avatar:       zu.Avatar,
			Password:     defaultPassword,
			DepartmentID: deptID,
			Status:       ConvertUserStatus(zu.Deleted),
		}

		// 检查是否已存在
		var existing model.User
		if err := m.goProjectDB.Where("username = ?", user.Username).First(&existing).Error; err == nil {
			m.userIDMap[zu.ID] = existing.ID
			log.Printf("用户已存在: %s (ID: %d -> %d)", user.Username, zu.ID, existing.ID)

			// 如果用户已存在，检查admin账号是否有管理员角色
			if strings.ToLower(zu.Account) == "admin" {
				var existingRoles []model.Role
				m.goProjectDB.Model(&existing).Association("Roles").Find(&existingRoles)
				hasAdminRole := false
				for _, r := range existingRoles {
					if r.Code == "admin" {
						hasAdminRole = true
						break
					}
				}
				if !hasAdminRole {
					// 为已存在的admin账号添加管理员角色
					if err := m.goProjectDB.Model(&existing).Association("Roles").Append(&adminRole); err != nil {
						log.Printf("为已存在的admin账号添加管理员角色失败: %v", err)
					} else {
						log.Printf("为已存在的admin账号添加了管理员角色")
					}
				}
			}
			continue
		}

		if err := m.goProjectDB.Create(&user).Error; err != nil {
			log.Printf("创建用户失败: %s, 错误: %v", user.Username, err)
			continue
		}

		// 分配角色
		// 根据zentao的role字段判断，如果是admin则分配admin角色
		// 否则查找对应的角色组
		var roles []model.Role

		// 特殊处理：admin账号必须赋予管理员角色
		if strings.ToLower(zu.Account) == "admin" || strings.ToLower(zu.Role) == "admin" || strings.Contains(strings.ToLower(zu.Role), "admin") {
			roles = append(roles, adminRole)
			log.Printf("用户 %s 被赋予管理员角色", user.Username)
		}

		// 查找用户所属的角色组（通过zt_usergroup表）
		type ZenTaoUserGroup struct {
			Group int `gorm:"column:group"`
		}
		var userGroups []ZenTaoUserGroup
		m.zenTaoDB.Table("zt_usergroup").Where("account = ?", zu.Account).Find(&userGroups)

		for _, ug := range userGroups {
			if roleID, ok := m.roleIDMap[ug.Group]; ok {
				var role model.Role
				if err := m.goProjectDB.First(&role, roleID).Error; err == nil {
					// 检查是否已添加
					found := false
					for _, r := range roles {
						if r.ID == role.ID {
							found = true
							break
						}
					}
					if !found {
						roles = append(roles, role)
					}
				}
			}
		}

		// 如果用户没有任何角色，且是admin账号，确保赋予管理员角色
		if len(roles) == 0 && strings.ToLower(zu.Account) == "admin" {
			roles = append(roles, adminRole)
			log.Printf("admin账号未找到角色组，强制赋予管理员角色")
		}

		if len(roles) > 0 {
			if err := m.goProjectDB.Model(&user).Association("Roles").Replace(roles); err != nil {
				log.Printf("分配角色失败: %s, 错误: %v", user.Username, err)
			}
		} else {
			log.Printf("警告: 用户 %s 未分配任何角色", user.Username)
		}

		m.userIDMap[zu.ID] = user.ID
		log.Printf("迁移用户: %s (ID: %d -> %d, 角色数: %d)", user.Username, zu.ID, user.ID, len(roles))
	}

	log.Printf("用户迁移完成，共迁移 %d 个用户", len(m.userIDMap))
	return nil
}

// MigrateProjects 迁移项目
func (m *Migrator) MigrateProjects() error {
	log.Println("开始迁移项目...")

	type ZenTaoProject struct {
		ID      int    `gorm:"column:id"`
		Name    string `gorm:"column:name"`
		Code    string `gorm:"column:code"`
		Desc    string `gorm:"column:desc"`
		Begin   string `gorm:"column:begin"`
		End     string `gorm:"column:end"`
		Status  string `gorm:"column:status"`
		Type    string `gorm:"column:type"`
		Deleted string `gorm:"column:deleted"`
	}

	var zentaoProjects []ZenTaoProject
	query := m.zenTaoDB.Table("zt_project").Where("deleted = '0'")
	query = query.Where("type = 'sprint' OR type = 'project'")
	if err := query.Find(&zentaoProjects).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个项目", len(zentaoProjects))

	for _, zp := range zentaoProjects {
		// 生成项目code：如果为空，则基于名称和ID生成唯一code
		projectCode := zp.Code
		if projectCode == "" {
			projectCode = GenerateProjectCode(zp.Name, zp.ID)
		}

		project := model.Project{
			Name:        zp.Name,
			Code:        projectCode,
			Description: zp.Desc,
			Status:      ConvertProjectStatus(zp.Status),
		}

		if zp.Begin != "" {
			project.StartDate = ParseDate(zp.Begin)
		}
		if zp.End != "" {
			project.EndDate = ParseDate(zp.End)
		}

		// 检查是否已存在（基于code）
		var existing model.Project
		if err := m.goProjectDB.Where("code = ?", project.Code).First(&existing).Error; err == nil {
			m.projectIDMap[zp.ID] = existing.ID
			log.Printf("项目已存在: %s (ID: %d -> %d, code: %s)", project.Name, zp.ID, existing.ID, project.Code)
			continue
		}

		if err := m.goProjectDB.Create(&project).Error; err != nil {
			log.Printf("创建项目失败: %s, 错误: %v", project.Name, err)
			continue
		}

		m.projectIDMap[zp.ID] = project.ID
		log.Printf("迁移项目: %s (ID: %d -> %d, code: %s)", project.Name, zp.ID, project.ID, project.Code)
	}

	log.Printf("项目迁移完成，共迁移 %d 个项目", len(m.projectIDMap))
	return nil
}

// MigrateModules 迁移项目模块
func (m *Migrator) MigrateModules() error {
	log.Println("开始迁移项目模块...")

	type ZenTaoModule struct {
		ID      int    `gorm:"column:id"`
		Name    string `gorm:"column:name"`
		Root    int    `gorm:"column:root"`    // 所属项目/产品ID
		Type    string `gorm:"column:type"`    // 类型：project/product
		Parent  int    `gorm:"column:parent"`  // 父模块ID
		Path    string `gorm:"column:path"`    // 路径
		Grade   int    `gorm:"column:grade"`   // 层级
		Order   int    `gorm:"column:order"`   // 排序
		Deleted string `gorm:"column:deleted"` // 删除标记
	}

	var zentaoModules []ZenTaoModule
	query := m.zenTaoDB.Table("zt_module").Where("deleted = '0'")
	if err := query.Find(&zentaoModules).Error; err != nil {
		// 如果表不存在，记录警告并返回
		log.Printf("警告: 未找到zt_module表或表为空: %v", err)
		return nil
	}

	log.Printf("找到 %d 个项目模块", len(zentaoModules))

	// 用于去重的映射：模块名称 -> 模块信息
	// 由于目标系统的Module是系统资源，名称必须唯一，需要处理重名情况
	moduleNameMap := make(map[string]*ZenTaoModule)

	// 第一遍：收集所有模块，处理重名（保留第一个）
	for i := range zentaoModules {
		zm := &zentaoModules[i]
		if zm.Name == "" {
			continue
		}

		// 如果名称已存在，记录警告但继续处理（使用第一个）
		if existing, exists := moduleNameMap[zm.Name]; exists {
			log.Printf("警告: 发现重名模块 '%s' (原ID: %d, 新ID: %d)，将使用第一个", zm.Name, existing.ID, zm.ID)
			continue
		}

		moduleNameMap[zm.Name] = zm
	}

	log.Printf("去重后共 %d 个唯一模块", len(moduleNameMap))

	// 第二遍：迁移模块
	for name, zm := range moduleNameMap {
		// 生成模块编码
		moduleCode := GenerateModuleCode(name, zm.ID)

		module := model.Module{
			Name:        name,
			Code:        moduleCode,
			Description: fmt.Sprintf("从禅道迁移的模块 (原ID: %d, 类型: %s)", zm.ID, zm.Type),
			Status:      1, // 正常
			Sort:        zm.Order,
		}

		// 检查是否已存在（基于名称）
		var existing model.Module
		if err := m.goProjectDB.Where("name = ?", module.Name).First(&existing).Error; err == nil {
			m.moduleIDMap[zm.ID] = existing.ID
			log.Printf("模块已存在: %s (ID: %d -> %d)", module.Name, zm.ID, existing.ID)
			continue
		}

		// 检查编码是否已存在，如果存在则重新生成
		if module.Code != "" {
			var existingByCode model.Module
			if err := m.goProjectDB.Where("code = ?", module.Code).First(&existingByCode).Error; err == nil {
				// 编码已存在，重新生成
				module.Code = GenerateModuleCode(name, zm.ID)
				log.Printf("模块编码冲突，重新生成: %s -> %s", name, module.Code)
			}
		}

		if err := m.goProjectDB.Create(&module).Error; err != nil {
			log.Printf("创建模块失败: %s, 错误: %v", module.Name, err)
			continue
		}

		m.moduleIDMap[zm.ID] = module.ID
		m.stats.moduleCount++
		log.Printf("迁移模块: %s (ID: %d -> %d, code: %s)", module.Name, zm.ID, module.ID, module.Code)
	}

	log.Printf("项目模块迁移完成，共迁移 %d 个模块", m.stats.moduleCount)
	return nil
}

// MigrateRequirements 迁移需求
func (m *Migrator) MigrateRequirements() error {
	log.Println("开始迁移需求...")

	type ZenTaoStory struct {
		ID         int     `gorm:"column:id"`
		Title      string  `gorm:"column:title"`
		Status     string  `gorm:"column:status"`
		Pri        int     `gorm:"column:pri"`
		Product    int     `gorm:"column:product"`
		OpenedBy   string  `gorm:"column:openedBy"`
		AssignedTo string  `gorm:"column:assignedTo"`
		Estimate   float64 `gorm:"column:estimate"`
		Deleted    string  `gorm:"column:deleted"`
	}

	type ZenTaoStorySpec struct {
		Story int    `gorm:"column:story"`
		Spec  string `gorm:"column:spec"`
	}

	type ZenTaoProjectStory struct {
		Project int `gorm:"column:project"`
		Story   int `gorm:"column:story"`
	}

	var zentaoStories []ZenTaoStory
	query := m.zenTaoDB.Table("zt_story").Where("deleted = '0'")
	if err := query.Find(&zentaoStories).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个需求", len(zentaoStories))

	for _, zs := range zentaoStories {
		// 通过zt_projectstory表获取项目ID（一个需求可能属于多个项目，取第一个）
		var projectStory ZenTaoProjectStory
		var projectID uint

		if err := m.zenTaoDB.Table("zt_projectstory").Where("story = ?", zs.ID).First(&projectStory).Error; err == nil {
			if newID, ok := m.projectIDMap[projectStory.Project]; ok {
				projectID = newID
			} else {
				log.Printf("需求 %s 的项目ID %d 不存在，跳过", zs.Title, projectStory.Project)
				continue
			}
		} else {
			// 如果没有在zt_projectstory中找到，尝试通过产品查找项目
			// 查找该产品关联的项目
			type ZenTaoProjectProduct struct {
				Project int `gorm:"column:project"`
			}
			var projectProduct ZenTaoProjectProduct
			if err := m.zenTaoDB.Table("zt_projectproduct").Where("product = ?", zs.Product).First(&projectProduct).Error; err == nil {
				if newID, ok := m.projectIDMap[projectProduct.Project]; ok {
					projectID = newID
				} else {
					log.Printf("需求 %s 通过产品 %d 找到的项目ID %d 不存在，跳过", zs.Title, zs.Product, projectProduct.Project)
					continue
				}
			} else {
				log.Printf("需求 %s 没有关联的项目，跳过", zs.Title)
				continue
			}
		}

		// 获取创建者ID
		var creatorID uint
		if zs.OpenedBy != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zs.OpenedBy).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					creatorID = newID
				}
			}
		}

		// 获取分配者ID
		var assigneeID *uint
		if zs.AssignedTo != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zs.AssignedTo).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					assigneeID = &newID
				}
			}
		}

		// 获取需求描述（从zt_storyspec表）
		var spec ZenTaoStorySpec
		description := ""
		if err := m.zenTaoDB.Table("zt_storyspec").Where("story = ?", zs.ID).First(&spec).Error; err == nil {
			description = spec.Spec
		}

		requirement := model.Requirement{
			Title:          zs.Title,
			Description:    description,
			Status:         ConvertRequirementStatus(zs.Status),
			Priority:       ConvertPriority(zs.Pri),
			ProjectID:      projectID,
			CreatorID:      creatorID,
			AssigneeID:     assigneeID,
			EstimatedHours: DaysToHours(zs.Estimate),
		}

		if err := m.goProjectDB.Create(&requirement).Error; err != nil {
			log.Printf("创建需求失败: %s, 错误: %v", requirement.Title, err)
			continue
		}

		m.requirementIDMap[zs.ID] = requirement.ID
		log.Printf("迁移需求: %s (ID: %d -> %d)", requirement.Title, zs.ID, requirement.ID)
	}

	log.Printf("需求迁移完成，共迁移 %d 个需求", len(m.requirementIDMap))
	return nil
}

// MigrateTasks 迁移任务
func (m *Migrator) MigrateTasks() error {
	log.Println("开始迁移任务...")

	type ZenTaoTask struct {
		ID         int     `gorm:"column:id"`
		Name       string  `gorm:"column:name"`
		Desc       string  `gorm:"column:desc"`
		Status     string  `gorm:"column:status"`
		Pri        int     `gorm:"column:pri"`
		Project    int     `gorm:"column:project"`
		Execution  int     `gorm:"column:execution"`
		Story      int     `gorm:"column:story"`
		OpenedBy   string  `gorm:"column:openedBy"`
		AssignedTo string  `gorm:"column:assignedTo"`
		EstStarted string  `gorm:"column:estStarted"`
		Deadline   string  `gorm:"column:deadline"`
		Estimate   float64 `gorm:"column:estimate"`
		Consumed   float64 `gorm:"column:consumed"`
		Deleted    string  `gorm:"column:deleted"`
	}

	var zentaoTasks []ZenTaoTask
	query := m.zenTaoDB.Table("zt_task").Where("deleted = '0'")
	if err := query.Find(&zentaoTasks).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个任务", len(zentaoTasks))

	for _, zt := range zentaoTasks {
		// 获取项目ID
		var projectID uint
		if zt.Execution > 0 {
			if newID, ok := m.projectIDMap[zt.Execution]; ok {
				projectID = newID
			}
		} else if zt.Project > 0 {
			if newID, ok := m.projectIDMap[zt.Project]; ok {
				projectID = newID
			}
		}

		if projectID == 0 {
			log.Printf("任务 %s 没有项目ID，跳过", zt.Name)
			continue
		}

		// 获取需求ID
		var requirementID *uint
		if zt.Story > 0 {
			if newID, ok := m.requirementIDMap[zt.Story]; ok {
				requirementID = &newID
			}
		}

		// 获取创建者ID
		var creatorID uint
		if zt.OpenedBy != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zt.OpenedBy).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					creatorID = newID
				}
			}
		}

		// 获取分配者ID
		var assigneeID *uint
		if zt.AssignedTo != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zt.AssignedTo).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					assigneeID = &newID
				}
			}
		}

		// 解析日期
		startDate := ParseDate(zt.EstStarted)
		dueDate := ParseDate(zt.Deadline)

		// 计算结束日期：优先使用截止日期，如果没有则根据开始日期和预估工时计算
		var endDate *time.Time
		if dueDate != nil {
			// 如果有截止日期，使用截止日期作为结束日期
			endDate = dueDate
		} else if startDate != nil && zt.Estimate > 0 {
			// 如果有开始日期和预估工时，计算结束日期（假设1天=8小时）
			days := int(zt.Estimate)
			end := startDate.AddDate(0, 0, days)
			endDate = &end
		}

		task := model.Task{
			Title:          zt.Name,
			Description:    zt.Desc,
			Status:         ConvertTaskStatus(zt.Status),
			Priority:       ConvertPriority(zt.Pri),
			ProjectID:      projectID,
			RequirementID:  requirementID,
			CreatorID:      creatorID,
			AssigneeID:     assigneeID,
			StartDate:      startDate,
			EndDate:        endDate,
			DueDate:        dueDate,
			EstimatedHours: DaysToHours(zt.Estimate),
			ActualHours:    DaysToHours(zt.Consumed),
		}

		if err := m.goProjectDB.Create(&task).Error; err != nil {
			log.Printf("创建任务失败: %s, 错误: %v", task.Title, err)
			continue
		}

		m.stats.taskCount++
		log.Printf("迁移任务: %s (ID: %d -> %d)", task.Title, zt.ID, task.ID)
	}

	log.Printf("任务迁移完成，共迁移 %d 个任务", m.stats.taskCount)
	return nil
}

// MigrateBugs 迁移Bug
func (m *Migrator) MigrateBugs() error {
	log.Println("开始迁移Bug...")

	type ZenTaoBug struct {
		ID            int    `gorm:"column:id"`
		Title         string `gorm:"column:title"`
		Steps         string `gorm:"column:steps"`
		Status        string `gorm:"column:status"`
		Severity      int    `gorm:"column:severity"`
		Pri           int    `gorm:"column:pri"`
		Project       int    `gorm:"column:project"`
		Story         int    `gorm:"column:story"`
		OpenedBy      string `gorm:"column:openedBy"`
		AssignedTo    string `gorm:"column:assignedTo"`
		Resolution    string `gorm:"column:resolution"`
		ResolvedBuild string `gorm:"column:resolvedBuild"`
		Deleted       string `gorm:"column:deleted"`
	}

	var zentaoBugs []ZenTaoBug
	query := m.zenTaoDB.Table("zt_bug").Where("deleted = '0'")
	if err := query.Find(&zentaoBugs).Error; err != nil {
		return err
	}

	log.Printf("找到 %d 个Bug", len(zentaoBugs))

	for _, zb := range zentaoBugs {
		// 获取项目ID
		var projectID uint
		if zb.Project > 0 {
			if newID, ok := m.projectIDMap[zb.Project]; ok {
				projectID = newID
			}
		}

		if projectID == 0 {
			log.Printf("Bug %s 没有项目ID，跳过", zb.Title)
			continue
		}

		// 获取需求ID
		var requirementID *uint
		if zb.Story > 0 {
			if newID, ok := m.requirementIDMap[zb.Story]; ok {
				requirementID = &newID
			}
		}

		// 获取创建者ID
		var creatorID uint
		if zb.OpenedBy != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zb.OpenedBy).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					creatorID = newID
				}
			}
		}

		bug := model.Bug{
			Title:         zb.Title,
			Description:   zb.Steps,
			Status:        ConvertBugStatus(zb.Status),
			Severity:      ConvertSeverity(zb.Severity),
			Priority:      ConvertPriority(zb.Pri),
			ProjectID:     projectID,
			RequirementID: requirementID,
			CreatorID:     creatorID,
			Solution:      zb.Resolution,
			SolutionNote:  zb.ResolvedBuild,
		}

		if err := m.goProjectDB.Create(&bug).Error; err != nil {
			log.Printf("创建Bug失败: %s, 错误: %v", bug.Title, err)
			continue
		}

		// 处理分配者（多对多关系）
		if zb.AssignedTo != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userID ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zb.AssignedTo).First(&userID).Error; err == nil {
				if newID, ok := m.userIDMap[userID.ID]; ok {
					var assignee model.User
					if err := m.goProjectDB.First(&assignee, newID).Error; err == nil {
						if err := m.goProjectDB.Model(&bug).Association("Assignees").Append(&assignee); err != nil {
							log.Printf("分配Bug失败: %s, 错误: %v", bug.Title, err)
						}
					}
				}
			}
		}

		m.stats.bugCount++
		log.Printf("迁移Bug: %s (ID: %d -> %d)", bug.Title, zb.ID, bug.ID)
	}

	log.Printf("Bug迁移完成，共迁移 %d 个Bug", m.stats.bugCount)
	return nil
}

// MigrateProjectMembers 迁移项目成员
func (m *Migrator) MigrateProjectMembers() error {
	log.Println("开始迁移项目成员...")

	type ZenTaoTeam struct {
		Root    int    `gorm:"column:root"`    // 指向项目/执行/任务的ID
		Type    string `gorm:"column:type"`    // 类型：project, execution, task
		Account string `gorm:"column:account"` // 用户账号
		Role    string `gorm:"column:role"`    // 角色
		Join    string `gorm:"column:join"`    // 加入日期
		Days    int    `gorm:"column:days"`    // 可用天数
		Hours   string `gorm:"column:hours"`   // 可用工时
	}

	var zentaoTeams []ZenTaoTeam
	teamCount := 0

	// 尝试从 zt_team 表获取项目成员
	if err := m.zenTaoDB.Table("zt_team").Find(&zentaoTeams).Error; err != nil {
		log.Printf("警告: 未找到zt_team表或表为空，尝试从其他来源获取项目成员: %v", err)
	} else {
		teamCount = len(zentaoTeams)
		log.Printf("从zt_team表找到 %d 条项目成员记录", teamCount)
	}

	// 如果 zt_team 表为空，尝试从项目创建者、任务分配者等推断项目成员
	if teamCount == 0 {
		log.Println("zt_team表为空，尝试从项目相关数据推断项目成员...")
		if err := m.migrateProjectMembersFromInference(); err != nil {
			log.Printf("从推断方式迁移项目成员失败: %v", err)
		}
		return nil
	}

	for _, zt := range zentaoTeams {
		// 获取项目ID
		// zt_team表的root字段指向项目/执行/任务的ID，type字段表示类型
		var projectID uint
		if zt.Root == 0 {
			log.Printf("项目成员的root字段为0，跳过")
			continue
		}

		// 根据type字段处理不同的情况
		if zt.Type == "project" {
			// 直接是项目ID
			if newID, ok := m.projectIDMap[zt.Root]; ok {
				projectID = newID
			} else {
				log.Printf("项目成员的项目ID %d 不存在，跳过", zt.Root)
				continue
			}
		} else if zt.Type == "execution" {
			// 是执行（execution）ID，需要通过执行找到项目
			type ZenTaoExecution struct {
				ID      int    `gorm:"column:id"`
				Project int    `gorm:"column:project"`
				Name    string `gorm:"column:name"`
			}
			var execution ZenTaoExecution
			if err := m.zenTaoDB.Table("zt_project").Where("id = ? AND type = 'execution'", zt.Root).First(&execution).Error; err == nil {
				// 找到了执行，通过执行的项目ID查找
				if newID, ok := m.projectIDMap[execution.Project]; ok {
					projectID = newID
					log.Printf("项目成员的执行ID %d (%s) 对应项目ID %d", zt.Root, execution.Name, execution.Project)
				} else {
					log.Printf("项目成员的执行ID %d (%s) 对应的项目ID %d 不存在，跳过", zt.Root, execution.Name, execution.Project)
					continue
				}
			} else {
				log.Printf("项目成员的执行ID %d 在zt_project表中不存在，跳过", zt.Root)
				continue
			}
		} else if zt.Type == "task" {
			// 是任务ID，需要通过任务找到项目
			type ZenTaoTask struct {
				ID        int `gorm:"column:id"`
				Project   int `gorm:"column:project"`
				Execution int `gorm:"column:execution"`
			}
			var task ZenTaoTask
			if err := m.zenTaoDB.Table("zt_task").Where("id = ?", zt.Root).First(&task).Error; err == nil {
				// 优先使用execution，如果没有则使用project
				var targetProjectID int
				if task.Execution > 0 {
					targetProjectID = task.Execution
				} else if task.Project > 0 {
					targetProjectID = task.Project
				}

				if targetProjectID > 0 {
					if newID, ok := m.projectIDMap[targetProjectID]; ok {
						projectID = newID
						log.Printf("项目成员的任务ID %d 对应项目ID %d", zt.Root, targetProjectID)
					} else {
						log.Printf("项目成员的任务ID %d 对应的项目ID %d 不存在，跳过", zt.Root, targetProjectID)
						continue
					}
				} else {
					log.Printf("项目成员的任务ID %d 没有关联的项目，跳过", zt.Root)
					continue
				}
			} else {
				log.Printf("项目成员的任务ID %d 在zt_task表中不存在，跳过", zt.Root)
				continue
			}
		} else {
			log.Printf("项目成员的类型 %s 不支持，跳过", zt.Type)
			continue
		}

		// 获取用户ID
		var userID uint
		if zt.Account != "" {
			type ZenTaoUserID struct {
				ID int `gorm:"column:id"`
			}
			var userIDRecord ZenTaoUserID
			if err := m.zenTaoDB.Table("zt_user").Where("account = ?", zt.Account).First(&userIDRecord).Error; err == nil {
				if newID, ok := m.userIDMap[userIDRecord.ID]; ok {
					userID = newID
				} else {
					log.Printf("项目成员的用户账号 %s 不存在，跳过", zt.Account)
					continue
				}
			} else {
				log.Printf("项目成员的用户账号 %s 未找到，跳过", zt.Account)
				continue
			}
		} else {
			log.Printf("项目成员的用户账号为空，跳过")
			continue
		}

		// 转换角色：将禅道角色映射到goproject角色
		// 禅道角色可能是：项目经理、开发、测试、产品、设计等
		// goproject角色：owner, member, viewer
		role := ConvertProjectRole(zt.Role)

		// 检查是否已存在
		var existingMember model.ProjectMember
		if err := m.goProjectDB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&existingMember).Error; err == nil {
			// 如果已存在，更新角色
			existingMember.Role = role
			if err := m.goProjectDB.Save(&existingMember).Error; err != nil {
				log.Printf("更新项目成员失败: 项目ID %d, 用户ID %d, 错误: %v", projectID, userID, err)
			} else {
				m.stats.projectMemberCount++
				log.Printf("更新项目成员: 项目ID %d, 用户ID %d, 角色: %s", projectID, userID, role)
			}
			continue
		}

		// 创建项目成员
		member := model.ProjectMember{
			ProjectID: projectID,
			UserID:    userID,
			Role:      role,
		}

		if err := m.goProjectDB.Create(&member).Error; err != nil {
			log.Printf("创建项目成员失败: 项目ID %d, 用户ID %d, 错误: %v", projectID, userID, err)
			continue
		}

		m.stats.projectMemberCount++
		log.Printf("迁移项目成员: 项目ID %d, 用户ID %d, 角色: %s", projectID, userID, role)
	}

	log.Printf("项目成员迁移完成，共迁移 %d 个成员", m.stats.projectMemberCount)
	return nil
}

// migrateProjectMembersFromInference 从任务、需求、Bug等数据推断项目成员
func (m *Migrator) migrateProjectMembersFromInference() error {
	log.Println("开始从推断方式迁移项目成员...")

	// 用于存储项目成员映射：projectID -> map[userID]role
	projectMembersMap := make(map[uint]map[uint]string)

	// 1. 从任务中获取项目成员（分配者）
	type ZenTaoTaskMember struct {
		Project    int    `gorm:"column:project"`
		Execution  int    `gorm:"column:execution"`
		AssignedTo string `gorm:"column:assignedTo"`
	}
	var taskMembers []ZenTaoTaskMember
	if err := m.zenTaoDB.Table("zt_task").Where("deleted = '0' AND assignedTo != ''").
		Select("project, execution, assignedTo").Find(&taskMembers).Error; err == nil {
		log.Printf("从任务中找到 %d 条记录", len(taskMembers))
		for _, tm := range taskMembers {
			var projectID uint
			if tm.Execution > 0 {
				if newID, ok := m.projectIDMap[tm.Execution]; ok {
					projectID = newID
				}
			} else if tm.Project > 0 {
				if newID, ok := m.projectIDMap[tm.Project]; ok {
					projectID = newID
				}
			}

			if projectID > 0 && tm.AssignedTo != "" {
				type ZenTaoUserID struct {
					ID int `gorm:"column:id"`
				}
				var userIDRecord ZenTaoUserID
				if err := m.zenTaoDB.Table("zt_user").Where("account = ?", tm.AssignedTo).First(&userIDRecord).Error; err == nil {
					if newID, ok := m.userIDMap[userIDRecord.ID]; ok {
						if projectMembersMap[projectID] == nil {
							projectMembersMap[projectID] = make(map[uint]string)
						}
						projectMembersMap[projectID][newID] = "member"
					}
				}
			}
		}
		log.Printf("从任务中推断出 %d 个项目的成员", len(projectMembersMap))
	}

	// 2. 从需求中获取项目成员（分配者）
	type ZenTaoStoryMember struct {
		AssignedTo string `gorm:"column:assignedTo"`
	}
	type ZenTaoProjectStory struct {
		Project int `gorm:"column:project"`
		Story   int `gorm:"column:story"`
	}
	var stories []struct {
		ID         int    `gorm:"column:id"`
		AssignedTo string `gorm:"column:assignedTo"`
	}
	if err := m.zenTaoDB.Table("zt_story").Where("deleted = '0' AND assignedTo != ''").
		Select("id, assignedTo").Find(&stories).Error; err == nil {
		log.Printf("从需求中找到 %d 条记录", len(stories))
		for _, story := range stories {
			var projectStory ZenTaoProjectStory
			if err := m.zenTaoDB.Table("zt_projectstory").Where("story = ?", story.ID).First(&projectStory).Error; err == nil {
				if projectID, ok := m.projectIDMap[projectStory.Project]; ok {
					type ZenTaoUserID struct {
						ID int `gorm:"column:id"`
					}
					var userIDRecord ZenTaoUserID
					if err := m.zenTaoDB.Table("zt_user").Where("account = ?", story.AssignedTo).First(&userIDRecord).Error; err == nil {
						if userID, ok := m.userIDMap[userIDRecord.ID]; ok {
							if projectMembersMap[projectID] == nil {
								projectMembersMap[projectID] = make(map[uint]string)
							}
							projectMembersMap[projectID][userID] = "member"
						}
					}
				}
			}
		}
		log.Printf("从需求中推断出 %d 个项目的成员", len(projectMembersMap))
	}

	// 3. 从Bug中获取项目成员（分配者）
	type ZenTaoBugMember struct {
		Project    int    `gorm:"column:project"`
		AssignedTo string `gorm:"column:assignedTo"`
	}
	var bugMembers []ZenTaoBugMember
	if err := m.zenTaoDB.Table("zt_bug").Where("deleted = '0' AND assignedTo != ''").
		Select("project, assignedTo").Find(&bugMembers).Error; err == nil {
		log.Printf("从Bug中找到 %d 条记录", len(bugMembers))
		for _, bm := range bugMembers {
			if projectID, ok := m.projectIDMap[bm.Project]; ok {
				type ZenTaoUserID struct {
					ID int `gorm:"column:id"`
				}
				var userIDRecord ZenTaoUserID
				if err := m.zenTaoDB.Table("zt_user").Where("account = ?", bm.AssignedTo).First(&userIDRecord).Error; err == nil {
					if userID, ok := m.userIDMap[userIDRecord.ID]; ok {
						if projectMembersMap[projectID] == nil {
							projectMembersMap[projectID] = make(map[uint]string)
						}
						projectMembersMap[projectID][userID] = "member"
					}
				}
			}
		}
	}

	// 创建项目成员记录
	log.Printf("准备创建项目成员，共 %d 个项目有成员", len(projectMembersMap))
	totalMembers := 0
	for projectID, members := range projectMembersMap {
		totalMembers += len(members)
		for userID, role := range members {
			// 检查是否已存在
			var existingMember model.ProjectMember
			if err := m.goProjectDB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&existingMember).Error; err == nil {
				log.Printf("项目成员已存在: 项目ID %d, 用户ID %d，跳过", projectID, userID)
				continue
			}

			member := model.ProjectMember{
				ProjectID: projectID,
				UserID:    userID,
				Role:      role,
			}

			if err := m.goProjectDB.Create(&member).Error; err != nil {
				log.Printf("创建项目成员失败: 项目ID %d, 用户ID %d, 错误: %v", projectID, userID, err)
				continue
			}

			m.stats.projectMemberCount++
			log.Printf("从推断方式迁移项目成员: 项目ID %d, 用户ID %d, 角色: %s", projectID, userID, role)
		}
	}

	log.Printf("从推断方式迁移项目成员完成，共找到 %d 个成员，成功迁移 %d 个成员", totalMembers, m.stats.projectMemberCount)
	return nil
}
