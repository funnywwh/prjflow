package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// GetProjectGroups 获取项目集列表
func (h *ProjectHandler) GetProjectGroups(c *gin.Context) {
	var projectGroups []model.ProjectGroup
	query := h.db

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&projectGroups).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, projectGroups)
}

// GetProjectGroup 获取项目集详情
func (h *ProjectHandler) GetProjectGroup(c *gin.Context) {
	id := c.Param("id")
	var projectGroup model.ProjectGroup
	if err := h.db.Preload("Projects").First(&projectGroup, id).Error; err != nil {
		utils.Error(c, 404, "项目集不存在")
		return
	}

	utils.Success(c, projectGroup)
}

// CreateProjectGroup 创建项目集
func (h *ProjectHandler) CreateProjectGroup(c *gin.Context) {
	var projectGroup model.ProjectGroup
	if err := c.ShouldBindJSON(&projectGroup); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Create(&projectGroup).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	utils.Success(c, projectGroup)
}

// UpdateProjectGroup 更新项目集
func (h *ProjectHandler) UpdateProjectGroup(c *gin.Context) {
	id := c.Param("id")
	var projectGroup model.ProjectGroup
	if err := h.db.First(&projectGroup, id).Error; err != nil {
		utils.Error(c, 404, "项目集不存在")
		return
	}

	if err := c.ShouldBindJSON(&projectGroup); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.db.Save(&projectGroup).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, projectGroup)
}

// DeleteProjectGroup 删除项目集
func (h *ProjectHandler) DeleteProjectGroup(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有项目
	var count int64
	h.db.Model(&model.Project{}).Where("project_group_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "项目集下存在项目，无法删除")
		return
	}

	if err := h.db.Delete(&model.ProjectGroup{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetProjects 获取项目列表
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var projects []model.Project
	query := h.db.Preload("ProjectGroup").Preload("Product")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 项目集筛选
	if projectGroupID := c.Query("project_group_id"); projectGroupID != "" {
		query = query.Where("project_group_id = ?", projectGroupID)
	}

	// 产品筛选
	if productID := c.Query("product_id"); productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.Project{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&projects).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":  projects,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetProject 获取项目详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	var project model.Project
	if err := h.db.
		Preload("ProjectGroup").
		Preload("Product").
		Preload("Members.User").
		First(&project, id).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	utils.Success(c, project)
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project model.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证项目集是否存在
	if project.ProjectGroupID > 0 {
		var projectGroup model.ProjectGroup
		if err := h.db.First(&projectGroup, project.ProjectGroupID).Error; err != nil {
			utils.Error(c, 404, "项目集不存在")
			return
		}
	}

	// 验证产品是否存在（如果提供了产品ID）
	if project.ProductID != nil && *project.ProductID > 0 {
		var product model.Product
		if err := h.db.First(&product, *project.ProductID).Error; err != nil {
			utils.Error(c, 404, "产品不存在")
			return
		}
	}

	if err := h.db.Create(&project).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("ProjectGroup").Preload("Product").First(&project, project.ID)

	utils.Success(c, project)
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var project model.Project
	if err := h.db.First(&project, id).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	if err := c.ShouldBindJSON(&project); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证项目集是否存在
	if project.ProjectGroupID > 0 {
		var projectGroup model.ProjectGroup
		if err := h.db.First(&projectGroup, project.ProjectGroupID).Error; err != nil {
			utils.Error(c, 404, "项目集不存在")
			return
		}
	}

	// 验证产品是否存在（如果提供了产品ID）
	if project.ProductID != nil && *project.ProductID > 0 {
		var product model.Product
		if err := h.db.First(&product, *project.ProductID).Error; err != nil {
			utils.Error(c, 404, "产品不存在")
			return
		}
	}

	if err := h.db.Save(&project).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("ProjectGroup").Preload("Product").First(&project, project.ID)

	utils.Success(c, project)
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有任务、Bug、需求等关联数据
	var count int64
	h.db.Model(&model.Task{}).Where("project_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "项目下存在任务，无法删除")
		return
	}

	h.db.Model(&model.Bug{}).Where("project_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "项目下存在Bug，无法删除")
		return
	}

	h.db.Model(&model.Requirement{}).Where("project_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "项目下存在需求，无法删除")
		return
	}

	if err := h.db.Delete(&model.Project{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetProjectMembers 获取项目成员列表
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	projectID := c.Param("id")
	var members []model.ProjectMember
	if err := h.db.Where("project_id = ?", projectID).Preload("User").Preload("User.Department").Find(&members).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, members)
}

// AddProjectMembers 添加项目成员
func (h *ProjectHandler) AddProjectMembers(c *gin.Context) {
	projectID := c.Param("id")

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
		Role    string `json:"role" binding:"required"`
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

	// 创建项目成员
	members := make([]model.ProjectMember, 0, len(req.UserIDs))
	for _, userID := range req.UserIDs {
		// 检查是否已经是成员
		var existingMember model.ProjectMember
		if err := h.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&existingMember).Error; err == nil {
			// 如果已存在，更新角色
			existingMember.Role = req.Role
			h.db.Save(&existingMember)
			continue
		}

		members = append(members, model.ProjectMember{
			ProjectID: project.ID,
			UserID:    userID,
			Role:      req.Role,
		})
	}

	if len(members) > 0 {
		if err := h.db.Create(&members).Error; err != nil {
			utils.Error(c, utils.CodeError, "添加成员失败")
			return
		}
	}

	utils.Success(c, gin.H{"message": "添加成功"})
}

// UpdateProjectMember 更新项目成员角色
func (h *ProjectHandler) UpdateProjectMember(c *gin.Context) {
	projectID := c.Param("id")
	memberID := c.Param("member_id")

	var member model.ProjectMember
	if err := h.db.Where("project_id = ? AND id = ?", projectID, memberID).First(&member).Error; err != nil {
		utils.Error(c, 404, "项目成员不存在")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	member.Role = req.Role
	if err := h.db.Save(&member).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	utils.Success(c, member)
}

// RemoveProjectMember 移除项目成员
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	projectID := c.Param("id")
	memberID := c.Param("member_id")

	if err := h.db.Where("project_id = ? AND id = ?", projectID, memberID).Delete(&model.ProjectMember{}).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// ParseTime 解析时间字符串
func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

