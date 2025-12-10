package api

import (
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

// GetRoles 获取所有角色
func (h *PermissionHandler) GetRoles(c *gin.Context) {
	var roles []model.Role
	if err := h.db.Find(&roles).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, roles)
}

// GetRole 获取角色详情（包含权限）
func (h *PermissionHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	var role model.Role
	if err := h.db.Preload("Permissions").First(&role, id).Error; err != nil {
		utils.Error(c, 404, "角色不存在")
		return
	}

	utils.Success(c, role)
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查角色代码是否已存在
	var existingRole model.Role
	if err := h.db.Where("code = ?", req.Code).First(&existingRole).Error; err == nil {
		utils.Error(c, 400, "角色代码已存在")
		return
	}

	// 设置默认状态
	if req.Status == 0 {
		req.Status = 1
	}

	role := model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.db.Create(&role).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.Error(c, 400, "角色代码或名称已存在")
			return
		}
		utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	utils.RecordAuditLog(h.db, userID.(uint), username.(string), "create", "role", role.ID, c, true, "", "")

	utils.Success(c, role)
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var role model.Role
	if err := h.db.First(&role, id).Error; err != nil {
		utils.Error(c, 404, "角色不存在")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Code        *string `json:"code"`
		Description *string `json:"description"`
		Status      *int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Code != nil {
		// 检查角色代码是否已被其他角色使用
		var existingRole model.Role
		if err := h.db.Where("code = ? AND id != ?", *req.Code, id).First(&existingRole).Error; err == nil {
			utils.Error(c, 400, "角色代码已存在")
			return
		}
		role.Code = *req.Code
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Status != nil {
		role.Status = *req.Status
	}

	if err := h.db.Save(&role).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.Error(c, 400, "角色代码或名称已存在")
			return
		}
		utils.Error(c, utils.CodeError, "更新失败: "+err.Error())
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	utils.RecordAuditLog(h.db, userID.(uint), username.(string), "update", "role", role.ID, c, true, "", "")

	utils.Success(c, role)
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	
	// 检查是否有用户使用此角色
	var count int64
	h.db.Table("user_roles").Where("role_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "该角色正在被用户使用，无法删除")
		return
	}

	var role model.Role
	if err := h.db.First(&role, id).Error; err == nil {
		// 记录审计日志（在删除前记录）
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		utils.RecordAuditLog(h.db, userID.(uint), username.(string), "delete", "role", role.ID, c, true, "", "")
	}

	if err := h.db.Delete(&model.Role{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetRolePermissions 获取角色权限
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	var role model.Role
	if err := h.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		utils.Error(c, 404, "角色不存在")
		return
	}

	utils.Success(c, role.Permissions)
}

// GetPermissions 获取所有权限
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var permissions []model.Permission
	if err := h.db.Find(&permissions).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, permissions)
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	var permission model.Permission
	if err := h.db.First(&permission, id).Error; err != nil {
		utils.Error(c, 404, "权限不存在")
		return
	}

	utils.Success(c, permission)
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Code         string `json:"code" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Resource     string `json:"resource"`
		Action       string `json:"action"`
		Description  string `json:"description"`
		Status       int    `json:"status"`
		MenuPath     string `json:"menu_path"`
		MenuIcon     string `json:"menu_icon"`
		MenuTitle    string `json:"menu_title"`
		ParentMenuID *uint  `json:"parent_menu_id"`
		MenuOrder    int    `json:"menu_order"`
		IsMenu       bool   `json:"is_menu"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查权限代码是否已存在
	var existingPerm model.Permission
	if err := h.db.Where("code = ?", req.Code).First(&existingPerm).Error; err == nil {
		utils.Error(c, 400, "权限代码已存在")
		return
	}

	// 设置默认状态
	if req.Status == 0 {
		req.Status = 1
	}

	permission := model.Permission{
		Code:         req.Code,
		Name:         req.Name,
		Resource:     req.Resource,
		Action:       req.Action,
		Description:  req.Description,
		Status:       req.Status,
		MenuPath:     req.MenuPath,
		MenuIcon:     req.MenuIcon,
		MenuTitle:    req.MenuTitle,
		ParentMenuID: req.ParentMenuID,
		MenuOrder:    req.MenuOrder,
		IsMenu:       req.IsMenu,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.Error(c, 400, "权限代码已存在")
			return
		}
		utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	utils.RecordAuditLog(h.db, userID.(uint), username.(string), "create", "permission", permission.ID, c, true, "", "")

	utils.Success(c, permission)
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var permission model.Permission
	if err := h.db.First(&permission, id).Error; err != nil {
		utils.Error(c, 404, "权限不存在")
		return
	}

	var req struct {
		Name         *string `json:"name"`
		Resource     *string `json:"resource"`
		Action       *string `json:"action"`
		Description  *string `json:"description"`
		Status       *int    `json:"status"`
		MenuPath     *string `json:"menu_path"`
		MenuIcon     *string `json:"menu_icon"`
		MenuTitle    *string `json:"menu_title"`
		ParentMenuID *uint   `json:"parent_menu_id"`
		MenuOrder    *int    `json:"menu_order"`
		IsMenu       *bool   `json:"is_menu"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.Name != nil {
		permission.Name = *req.Name
	}
	if req.Resource != nil {
		permission.Resource = *req.Resource
	}
	if req.Action != nil {
		permission.Action = *req.Action
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.Status != nil {
		permission.Status = *req.Status
	}
	if req.MenuPath != nil {
		permission.MenuPath = *req.MenuPath
	}
	if req.MenuIcon != nil {
		permission.MenuIcon = *req.MenuIcon
	}
	if req.MenuTitle != nil {
		permission.MenuTitle = *req.MenuTitle
	}
	if req.ParentMenuID != nil {
		permission.ParentMenuID = req.ParentMenuID
	}
	if req.MenuOrder != nil {
		permission.MenuOrder = *req.MenuOrder
	}
	if req.IsMenu != nil {
		permission.IsMenu = *req.IsMenu
	}

	if err := h.db.Save(&permission).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	utils.RecordAuditLog(h.db, userID.(uint), username.(string), "update", "permission", permission.ID, c, true, "", "")

	utils.Success(c, permission)
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	
	// 检查是否有角色使用此权限
	var count int64
	h.db.Table("role_permissions").Where("permission_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "该权限正在被角色使用，无法删除")
		return
	}

	var permission model.Permission
	if err := h.db.First(&permission, id).Error; err == nil {
		// 记录审计日志（在删除前记录）
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		utils.RecordAuditLog(h.db, userID.(uint), username.(string), "delete", "permission", permission.ID, c, true, "", "")
	}

	if err := h.db.Delete(&model.Permission{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// AssignRolePermissions 分配角色权限
func (h *PermissionHandler) AssignRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	var role model.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		utils.Error(c, 404, "角色不存在")
		return
	}

	var permissions []model.Permission
	if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询权限失败")
		return
	}

	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		utils.Error(c, utils.CodeError, "分配权限失败")
		return
	}

	utils.Success(c, gin.H{"message": "分配成功"})
}

// GetUserRoles 获取用户角色
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	var user model.User
	if err := h.db.Preload("Roles").First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	utils.Success(c, user.Roles)
}

// AssignUserRoles 分配用户角色
func (h *PermissionHandler) AssignUserRoles(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		RoleIDs []uint `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	var roles []model.Role
	if err := h.db.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询角色失败")
		return
	}

	if err := h.db.Model(&user).Association("Roles").Replace(roles); err != nil {
		utils.Error(c, utils.CodeError, "分配角色失败")
		return
	}

	utils.Success(c, gin.H{"message": "分配成功"})
}

// GetUserPermissions 获取当前用户的权限列表
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未登录")
		return
	}

	// 类型转换
	var userIDUint uint
	switch v := userID.(type) {
	case uint:
		userIDUint = v
	case int:
		userIDUint = uint(v)
	case float64:
		userIDUint = uint(v)
	default:
		utils.Error(c, 401, "无效的用户ID")
		return
	}

	var user model.User
	// 使用 Unscoped() 查询，包括软删除的用户（虽然不应该有软删除，但为了安全）
	if err := h.db.Unscoped().Preload("Roles.Permissions").First(&user, userIDUint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, 404, "用户不存在")
		} else {
			utils.Error(c, utils.CodeError, "查询用户失败: "+err.Error())
		}
		return
	}
	
	// 如果用户被软删除，返回错误
	if user.DeletedAt.Valid {
		utils.Error(c, 404, "用户已被删除")
		return
	}

	// 收集所有权限（去重）
	permMap := make(map[string]model.Permission)
	for _, role := range user.Roles {
		if role.Status == 0 {
			continue // 跳过禁用的角色
		}
		for _, perm := range role.Permissions {
			if perm.Status == 1 {
				permMap[perm.Code] = perm
			}
		}
	}

	permissions := make([]model.Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	// 只返回权限代码列表
	permCodes := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permCodes = append(permCodes, perm.Code)
	}

	utils.Success(c, permCodes)
}

// GetMenus 获取菜单树（根据当前用户的权限）
func (h *PermissionHandler) GetMenus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未登录")
		return
	}

	var user model.User
	if err := h.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	// 收集用户的所有权限代码
	permCodes := make(map[string]bool)
	isAdmin := false
	for _, role := range user.Roles {
		if role.Status == 0 {
			continue // 跳过禁用的角色
		}
		// 管理员拥有所有权限
		if role.Code == "admin" {
			isAdmin = true
			break
		}
		for _, perm := range role.Permissions {
			if perm.Status == 1 {
				permCodes[perm.Code] = true
			}
		}
	}

	// 获取所有菜单权限（is_menu = true）
	var menuPermissions []model.Permission
	query := h.db.Where("is_menu = ? AND status = ?", true, 1)
	if !isAdmin {
		// 非管理员只获取有权限的菜单
		var codes []string
		for code := range permCodes {
			codes = append(codes, code)
		}
		if len(codes) > 0 {
			query = query.Where("code IN ?", codes)
		} else {
			// 没有任何权限，返回空菜单（前端会使用静态配置作为后备）
			// 注意：这里返回空数组，让前端使用静态配置
			utils.Success(c, []interface{}{})
			return
		}
	}
	// 管理员获取所有 is_menu = true 的权限（不进行权限代码过滤）
	
	if err := query.Order("menu_order ASC, id ASC").Find(&menuPermissions).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询菜单失败")
		return
	}

	// 构建菜单树
	type MenuItem struct {
		ID          uint       `json:"id"`
		Key         string     `json:"key"`
		Title       string     `json:"title"`
		Icon        string     `json:"icon"`
		Path        string     `json:"path"`
		Permission  string     `json:"permission"`
		Order       int        `json:"order"`
		Children    []MenuItem `json:"children,omitempty"`
	}

	// 创建菜单项映射（第一遍：创建所有菜单项）
	menuMap := make(map[uint]*MenuItem)
	for _, perm := range menuPermissions {
		menuTitle := perm.MenuTitle
		if menuTitle == "" {
			menuTitle = perm.Name
		}

		menuItem := &MenuItem{
			ID:         perm.ID,
			Key:        perm.Code,
			Title:      menuTitle,
			Icon:       perm.MenuIcon,
			Path:       perm.MenuPath,
			Permission: perm.Code,
			Order:      perm.MenuOrder,
			Children:   []MenuItem{},
		}
		menuMap[perm.ID] = menuItem
	}

	// 第二遍：建立父子关系
	// 先收集所有根菜单（parent_menu_id 为 nil 的菜单）
	rootMenuSet := make(map[uint]bool)
	for _, perm := range menuPermissions {
		if perm.ParentMenuID == nil {
			rootMenuSet[perm.ID] = true
		}
	}
	
	// 建立父子关系
	for _, perm := range menuPermissions {
		menuItem := menuMap[perm.ID]
		if perm.ParentMenuID != nil {
			if parent, exists := menuMap[*perm.ParentMenuID]; exists {
				parent.Children = append(parent.Children, *menuItem)
			} else {
				// 父菜单不存在或不在权限范围内，作为根菜单
				rootMenuSet[perm.ID] = true
			}
		}
	}
	
	// 从 menuMap 中提取根菜单（确保使用最新的 Children 数据）
	var rootMenus []MenuItem
	for rootID := range rootMenuSet {
		if rootMenu, exists := menuMap[rootID]; exists {
			rootMenus = append(rootMenus, *rootMenu)
		}
	}

	// 排序根菜单和子菜单（当 Order 相同时，使用 ID 作为次要排序条件，确保排序稳定）
	sort.Slice(rootMenus, func(i, j int) bool {
		if rootMenus[i].Order != rootMenus[j].Order {
			return rootMenus[i].Order < rootMenus[j].Order
		}
		return rootMenus[i].ID < rootMenus[j].ID
	})
	for i := range rootMenus {
		sort.Slice(rootMenus[i].Children, func(a, b int) bool {
			if rootMenus[i].Children[a].Order != rootMenus[i].Children[b].Order {
				return rootMenus[i].Children[a].Order < rootMenus[i].Children[b].Order
			}
			return rootMenus[i].Children[a].ID < rootMenus[i].Children[b].ID
		})
	}

	utils.Success(c, rootMenus)
}

// GetAllMenus 获取所有菜单树（用于管理，不根据用户权限过滤）
func (h *PermissionHandler) GetAllMenus(c *gin.Context) {
	// 获取所有菜单权限（is_menu = true）
	var menuPermissions []model.Permission
	if err := h.db.Where("is_menu = ? AND status = ?", true, 1).Order("menu_order ASC, id ASC").Find(&menuPermissions).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询菜单失败")
		return
	}

	// 构建菜单树
	type MenuItem struct {
		ID          uint       `json:"id"`
		Key         string     `json:"key"`
		Title       string     `json:"title"`
		Icon        string     `json:"icon"`
		Path        string     `json:"path"`
		Permission  string     `json:"permission"`
		Order       int        `json:"order"`
		Children    []MenuItem `json:"children,omitempty"`
	}

	// 创建菜单项映射（第一遍：创建所有菜单项）
	menuMap := make(map[uint]*MenuItem)
	for _, perm := range menuPermissions {
		menuTitle := perm.MenuTitle
		if menuTitle == "" {
			menuTitle = perm.Name
		}

		menuItem := &MenuItem{
			ID:         perm.ID,
			Key:        perm.Code,
			Title:      menuTitle,
			Icon:       perm.MenuIcon,
			Path:       perm.MenuPath,
			Permission: perm.Code,
			Order:      perm.MenuOrder,
			Children:   []MenuItem{},
		}
		menuMap[perm.ID] = menuItem
	}

	// 第二遍：建立父子关系
	// 先收集所有根菜单（parent_menu_id 为 nil 的菜单）
	rootMenuSet := make(map[uint]bool)
	for _, perm := range menuPermissions {
		if perm.ParentMenuID == nil {
			rootMenuSet[perm.ID] = true
		}
	}
	
	// 建立父子关系
	for _, perm := range menuPermissions {
		menuItem := menuMap[perm.ID]
		if perm.ParentMenuID != nil {
			if parent, exists := menuMap[*perm.ParentMenuID]; exists {
				parent.Children = append(parent.Children, *menuItem)
			} else {
				// 父菜单不存在或不在权限范围内，作为根菜单
				rootMenuSet[perm.ID] = true
			}
		}
	}
	
	// 从 menuMap 中提取根菜单（确保使用最新的 Children 数据）
	var rootMenus []MenuItem
	for rootID := range rootMenuSet {
		if rootMenu, exists := menuMap[rootID]; exists {
			rootMenus = append(rootMenus, *rootMenu)
		}
	}

	// 排序根菜单和子菜单（当 Order 相同时，使用 ID 作为次要排序条件，确保排序稳定）
	sort.Slice(rootMenus, func(i, j int) bool {
		if rootMenus[i].Order != rootMenus[j].Order {
			return rootMenus[i].Order < rootMenus[j].Order
		}
		return rootMenus[i].ID < rootMenus[j].ID
	})
	for i := range rootMenus {
		sort.Slice(rootMenus[i].Children, func(a, b int) bool {
			if rootMenus[i].Children[a].Order != rootMenus[i].Children[b].Order {
				return rootMenus[i].Children[a].Order < rootMenus[i].Children[b].Order
			}
			return rootMenus[i].Children[a].ID < rootMenus[i].Children[b].ID
		})
	}

	utils.Success(c, rootMenus)
}
