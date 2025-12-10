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

	// 获取所有部门（确保 ParentID 字段正确加载）
	if err := query.Select("id, name, code, parent_id, level, sort, status, created_at, updated_at").Order("level, sort").Find(&departments).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	// 构建树形结构
	tree := h.buildDepartmentTree(departments, nil)

	utils.Success(c, tree)
}

// buildDepartmentTree 构建部门树
// 确保每个部门只出现一次
func (h *DepartmentHandler) buildDepartmentTree(departments []model.Department, parentID *uint) []model.Department {
	// 使用 processed map 确保每个部门只处理一次
	processed := make(map[uint]bool)
	
	// 创建部门ID到部门的映射
	deptMap := make(map[uint]*model.Department)
	for i := range departments {
		// 初始化 Children 为空切片
		departments[i].Children = []model.Department{}
		deptMap[departments[i].ID] = &departments[i]
	}

	// 收集所有根节点和子节点
	var rootNodeIDs []uint
	childMap := make(map[uint][]*model.Department) // 父ID -> 子节点列表
	
	for i := range departments {
		dept := &departments[i]
		if dept.ParentID == nil {
			// 根节点
			rootNodeIDs = append(rootNodeIDs, dept.ID)
		} else {
			// 子节点，按父ID分组
			parentID := *dept.ParentID
			childMap[parentID] = append(childMap[parentID], dept)
		}
	}

	// 递归构建树形结构
	var buildNode func(uint) *model.Department
	buildNode = func(deptID uint) *model.Department {
		if processed[deptID] {
			// 部门已被处理，返回 nil 避免重复
			return nil
		}
		processed[deptID] = true
		
		dept := *deptMap[deptID]
		// 构建子节点
		if children, exists := childMap[deptID]; exists {
			dept.Children = make([]model.Department, 0, len(children))
			for _, child := range children {
				if !processed[child.ID] {
					if childNode := buildNode(child.ID); childNode != nil {
						dept.Children = append(dept.Children, *childNode)
					}
				}
			}
		}
		return &dept
	}

	// 构建根节点
	result := make([]model.Department, 0, len(rootNodeIDs))
	for _, rootID := range rootNodeIDs {
		if !processed[rootID] {
			if rootNode := buildNode(rootID); rootNode != nil {
				result = append(result, *rootNode)
			}
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

	// 检查部门代码是否已存在
	var existingDept model.Department
	if err := h.db.Where("code = ?", department.Code).First(&existingDept).Error; err == nil {
		utils.Error(c, 400, "部门代码已存在")
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
		if utils.IsUniqueConstraintError(err) {
			utils.Error(c, 400, "部门代码已存在")
			return
		}
		utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	if userID != nil && username != nil {
		utils.RecordAuditLog(h.db, userID.(uint), username.(string), "create", "department", department.ID, c, true, "", "")
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

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	if userID != nil && username != nil {
		utils.RecordAuditLog(h.db, userID.(uint), username.(string), "update", "department", department.ID, c, true, "", "")
	}

	utils.Success(c, department)
}

// DeleteDepartment 删除部门（硬删除）
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	
	// 检查部门是否存在
	var department model.Department
	if err := h.db.First(&department, id).Error; err != nil {
		utils.Error(c, 404, "部门不存在")
		return
	}
	
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

	// 硬删除（物理删除）
	if err := h.db.Unscoped().Delete(&department).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败: "+err.Error())
		return
	}

	// 记录审计日志
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	if userID != nil && username != nil {
		utils.RecordAuditLog(h.db, userID.(uint), username.(string), "delete", "department", department.ID, c, true, "", "")
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

