package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type DepartmentHandler struct {
	db *gorm.DB
}

func NewDepartmentHandler(db *gorm.DB) *DepartmentHandler {
	return &DepartmentHandler{db: db}
}

// GetDepartments 获取部门列表（树形结构）
func (h *DepartmentHandler) GetDepartments(c *gin.Context) {
	var departments []model.Department
	query := h.db

	// 获取所有部门
	if err := query.Order("level, sort").Find(&departments).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	// 构建树形结构
	tree := h.buildDepartmentTree(departments, nil)

	utils.Success(c, tree)
}

// buildDepartmentTree 构建部门树
func (h *DepartmentHandler) buildDepartmentTree(departments []model.Department, parentID *uint) []model.Department {
	var result []model.Department
	for _, dept := range departments {
		if (parentID == nil && dept.ParentID == nil) || (parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {
			children := h.buildDepartmentTree(departments, &dept.ID)
			dept.Children = children
			result = append(result, dept)
		}
	}
	return result
}

// GetDepartment 获取部门详情
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	id := c.Param("id")
	var department model.Department
	if err := h.db.Preload("Parent").Preload("Children").First(&department, id).Error; err != nil {
		utils.Error(c, 404, "部门不存在")
		return
	}

	utils.Success(c, department)
}

// CreateDepartment 创建部门
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var department model.Department
	if err := c.ShouldBindJSON(&department); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 计算层级
	if department.ParentID != nil {
		var parent model.Department
		if err := h.db.First(&parent, *department.ParentID).Error; err != nil {
			utils.Error(c, 404, "父部门不存在")
			return
		}
		department.Level = parent.Level + 1
	} else {
		department.Level = 1
	}

	if err := h.db.Create(&department).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, department)
}

// UpdateDepartment 更新部门
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	var department model.Department
	if err := h.db.First(&department, id).Error; err != nil {
		utils.Error(c, 404, "部门不存在")
		return
	}

	if err := c.ShouldBindJSON(&department); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 重新计算层级
	if department.ParentID != nil {
		var parent model.Department
		if err := h.db.First(&parent, *department.ParentID).Error; err != nil {
			utils.Error(c, 404, "父部门不存在")
			return
		}
		department.Level = parent.Level + 1
	} else {
		department.Level = 1
	}

	if err := h.db.Save(&department).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, department)
}

// DeleteDepartment 删除部门
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	
	// 检查是否有子部门
	var count int64
	h.db.Model(&model.Department{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "存在子部门，无法删除")
		return
	}

	// 检查是否有用户
	h.db.Model(&model.User{}).Where("department_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "部门下存在用户，无法删除")
		return
	}

	if err := h.db.Delete(&model.Department{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetDepartmentMembers 获取部门成员列表
func (h *DepartmentHandler) GetDepartmentMembers(c *gin.Context) {
	deptID := c.Param("id")
	var users []model.User
	if err := h.db.Where("department_id = ?", deptID).Preload("Roles").Find(&users).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, users)
}

// AddDepartmentMembers 添加部门成员
func (h *DepartmentHandler) AddDepartmentMembers(c *gin.Context) {
	deptID := c.Param("id")

	// 验证部门是否存在
	var department model.Department
	if err := h.db.First(&department, deptID).Error; err != nil {
		utils.Error(c, 404, "部门不存在")
		return
	}

	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证用户是否存在
	var users []model.User
	if err := h.db.Where("id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询用户失败")
		return
	}

	if len(users) != len(req.UserIDs) {
		utils.Error(c, 400, "部分用户不存在")
		return
	}

	// 更新用户的部门ID
	var deptIDUint uint
	if _, err := fmt.Sscanf(deptID, "%d", &deptIDUint); err != nil {
		utils.Error(c, 400, "部门ID格式错误")
		return
	}
	if err := h.db.Model(&model.User{}).Where("id IN ?", req.UserIDs).Update("department_id", deptIDUint).Error; err != nil {
		utils.Error(c, utils.CodeError, "添加成员失败")
		return
	}

	utils.Success(c, gin.H{"message": "添加成功"})
}

// RemoveDepartmentMember 移除部门成员
func (h *DepartmentHandler) RemoveDepartmentMember(c *gin.Context) {
	deptID := c.Param("id")
	userID := c.Param("user_id")

	// 验证用户是否属于该部门
	var user model.User
	if err := h.db.Where("id = ? AND department_id = ?", userID, deptID).First(&user).Error; err != nil {
		utils.Error(c, 404, "用户不属于该部门")
		return
	}

	// 将用户的部门ID设置为NULL
	if err := h.db.Model(&user).Update("department_id", nil).Error; err != nil {
		utils.Error(c, utils.CodeError, "移除成员失败")
		return
	}

	utils.Success(c, gin.H{"message": "移除成功"})
}

