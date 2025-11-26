package main

import (
	"fmt"
	"log"
	"strings"

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
	
	// 统计信息
	stats struct {
		taskCount int
		bugCount  int
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
	
	// 5. 迁移需求
	if err := m.MigrateRequirements(); err != nil {
		return fmt.Errorf("迁移需求失败: %w", err)
	}
	
	// 6. 迁移任务
	if err := m.MigrateTasks(); err != nil {
		return fmt.Errorf("迁移任务失败: %w", err)
	}
	
	// 7. 迁移Bug
	if err := m.MigrateBugs(); err != nil {
		return fmt.Errorf("迁移Bug失败: %w", err)
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
		project := model.Project{
			Name:        zp.Name,
			Code:        zp.Code,
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
			log.Printf("项目已存在: %s (ID: %d -> %d)", project.Name, zp.ID, existing.ID)
			continue
		}
		
		if err := m.goProjectDB.Create(&project).Error; err != nil {
			log.Printf("创建项目失败: %s, 错误: %v", project.Name, err)
			continue
		}
		
		m.projectIDMap[zp.ID] = project.ID
		log.Printf("迁移项目: %s (ID: %d -> %d)", project.Name, zp.ID, project.ID)
	}
	
	log.Printf("项目迁移完成，共迁移 %d 个项目", len(m.projectIDMap))
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
			Description:   description,
			Status:         ConvertRequirementStatus(zs.Status),
			Priority:      ConvertPriority(zs.Pri),
			ProjectID:      projectID,
			CreatorID:     creatorID,
			AssigneeID:    assigneeID,
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
		
		task := model.Task{
			Title:         zt.Name,
			Description:   zt.Desc,
			Status:        ConvertTaskStatus(zt.Status),
			Priority:      ConvertPriority(zt.Pri),
			ProjectID:     projectID,
			RequirementID: requirementID,
			CreatorID:     creatorID,
			AssigneeID:    assigneeID,
			StartDate:     ParseDate(zt.EstStarted),
			DueDate:       ParseDate(zt.Deadline),
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
		ID           int     `gorm:"column:id"`
		Title        string  `gorm:"column:title"`
		Steps        string  `gorm:"column:steps"`
		Status       string  `gorm:"column:status"`
		Severity     int     `gorm:"column:severity"`
		Pri          int     `gorm:"column:pri"`
		Project      int     `gorm:"column:project"`
		Story        int     `gorm:"column:story"`
		OpenedBy     string  `gorm:"column:openedBy"`
		AssignedTo   string  `gorm:"column:assignedTo"`
		Resolution   string  `gorm:"column:resolution"`
		ResolvedBuild string `gorm:"column:resolvedBuild"`
		Deleted      string  `gorm:"column:deleted"`
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

