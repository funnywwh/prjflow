package api

import (
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

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Create(&role).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

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

	if err := c.ShouldBindJSON(&role); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Save(&role).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, role)
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&model.Role{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
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

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var permission model.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Create(&permission).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, permission)
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

