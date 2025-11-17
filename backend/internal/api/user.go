package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []model.User
	query := h.db.Preload("Department").Preload("Roles")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 部门筛选
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.User{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":  users,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := h.db.Preload("Department").Preload("Roles").First(&user, id).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Create(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, user)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&model.User{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

