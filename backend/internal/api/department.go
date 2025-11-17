package api

import (
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

