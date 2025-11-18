package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type RequirementHandler struct {
	db *gorm.DB
}

func NewRequirementHandler(db *gorm.DB) *RequirementHandler {
	return &RequirementHandler{db: db}
}

// GetRequirements 获取需求列表
func (h *RequirementHandler) GetRequirements(c *gin.Context) {
	var requirements []model.Requirement
	query := h.db.Preload("Product").Preload("Project").Preload("Creator").Preload("Assignee")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 产品筛选
	if productID := c.Query("product_id"); productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	// 项目筛选
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 优先级筛选
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// 负责人筛选
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		query = query.Where("assignee_id = ?", assigneeID)
	}

	// 创建人筛选
	if creatorID := c.Query("creator_id"); creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.Requirement{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&requirements).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      requirements,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRequirement 获取需求详情
func (h *RequirementHandler) GetRequirement(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.Preload("Product").Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	utils.Success(c, requirement)
}

// CreateRequirement 创建需求
func (h *RequirementHandler) CreateRequirement(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		ProductID   *uint  `json:"product_id"`
		ProjectID  *uint  `json:"project_id"`
		AssigneeID *uint  `json:"assignee_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未登录")
		return
	}

	// 验证状态
	if req.Status == "" {
		req.Status = "pending"
	}
	validStatuses := map[string]bool{
		"pending":    true,
		"in_progress": true,
		"completed":  true,
		"cancelled":  true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	// 验证优先级
	if req.Priority == "" {
		req.Priority = "medium"
	}
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	if !validPriorities[req.Priority] {
		utils.Error(c, 400, "优先级值无效")
		return
	}

	// 如果指定了产品，验证产品是否存在
	if req.ProductID != nil {
		var product model.Product
		if err := h.db.First(&product, *req.ProductID).Error; err != nil {
			utils.Error(c, 400, "产品不存在")
			return
		}
	}

	// 如果指定了项目，验证项目是否存在
	if req.ProjectID != nil {
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 400, "项目不存在")
			return
		}
	}

	// 如果指定了负责人，验证用户是否存在
	if req.AssigneeID != nil {
		var user model.User
		if err := h.db.First(&user, *req.AssigneeID).Error; err != nil {
			utils.Error(c, 400, "负责人不存在")
			return
		}
	}

	requirement := model.Requirement{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		ProductID:   req.ProductID,
		ProjectID:  req.ProjectID,
		CreatorID:  userID.(uint),
		AssigneeID: req.AssigneeID,
	}

	if err := h.db.Create(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Product").Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

	utils.Success(c, requirement)
}

// UpdateRequirement 更新需求
func (h *RequirementHandler) UpdateRequirement(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *string `json:"priority"`
		ProductID   *uint   `json:"product_id"`
		ProjectID  *uint   `json:"project_id"`
		AssigneeID *uint   `json:"assignee_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 更新字段
	if req.Title != nil {
		requirement.Title = *req.Title
	}
	if req.Description != nil {
		requirement.Description = *req.Description
	}
	if req.Status != nil {
		// 验证状态
		validStatuses := map[string]bool{
			"pending":    true,
			"in_progress": true,
			"completed":  true,
			"cancelled":  true,
		}
		if !validStatuses[*req.Status] {
			utils.Error(c, 400, "状态值无效")
			return
		}
		requirement.Status = *req.Status
	}
	if req.Priority != nil {
		// 验证优先级
		validPriorities := map[string]bool{
			"low":    true,
			"medium": true,
			"high":   true,
			"urgent": true,
		}
		if !validPriorities[*req.Priority] {
			utils.Error(c, 400, "优先级值无效")
			return
		}
		requirement.Priority = *req.Priority
	}
	if req.ProductID != nil {
		// 验证产品是否存在
		if *req.ProductID != 0 {
			var product model.Product
			if err := h.db.First(&product, *req.ProductID).Error; err != nil {
				utils.Error(c, 400, "产品不存在")
				return
			}
			requirement.ProductID = req.ProductID
		} else {
			requirement.ProductID = nil
		}
	}
	if req.ProjectID != nil {
		// 验证项目是否存在
		if *req.ProjectID != 0 {
			var project model.Project
			if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
				utils.Error(c, 400, "项目不存在")
				return
			}
			requirement.ProjectID = req.ProjectID
		} else {
			requirement.ProjectID = nil
		}
	}
	if req.AssigneeID != nil {
		// 验证负责人是否存在
		if *req.AssigneeID != 0 {
			var user model.User
			if err := h.db.First(&user, *req.AssigneeID).Error; err != nil {
				utils.Error(c, 400, "负责人不存在")
				return
			}
			requirement.AssigneeID = req.AssigneeID
		} else {
			requirement.AssigneeID = nil
		}
	}

	if err := h.db.Save(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Product").Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

	utils.Success(c, requirement)
}

// DeleteRequirement 删除需求
func (h *RequirementHandler) DeleteRequirement(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有Bug关联
	var count int64
	h.db.Model(&model.Bug{}).Where("requirement_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "需求下存在关联的Bug，无法删除")
		return
	}

	if err := h.db.Delete(&model.Requirement{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetRequirementStatistics 获取需求统计
func (h *RequirementHandler) GetRequirementStatistics(c *gin.Context) {
	var stats struct {
		Total           int64 `json:"total"`
		Pending         int64 `json:"pending"`
		InProgress      int64 `json:"in_progress"`
		Completed       int64 `json:"completed"`
		Cancelled       int64 `json:"cancelled"`
		LowPriority     int64 `json:"low_priority"`
		MediumPriority  int64 `json:"medium_priority"`
		HighPriority    int64 `json:"high_priority"`
		UrgentPriority  int64 `json:"urgent_priority"`
	}

	baseQuery := h.db.Model(&model.Requirement{})

	// 应用筛选条件（与列表查询保持一致）
	if keyword := c.Query("keyword"); keyword != "" {
		baseQuery = baseQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if productID := c.Query("product_id"); productID != "" {
		baseQuery = baseQuery.Where("product_id = ?", productID)
	}
	if projectID := c.Query("project_id"); projectID != "" {
		baseQuery = baseQuery.Where("project_id = ?", projectID)
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		baseQuery = baseQuery.Where("assignee_id = ?", assigneeID)
	}
	if creatorID := c.Query("creator_id"); creatorID != "" {
		baseQuery = baseQuery.Where("creator_id = ?", creatorID)
	}

	// 统计总数
	baseQuery.Session(&gorm.Session{}).Count(&stats.Total)

	// 按状态统计
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "pending").Count(&stats.Pending)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "in_progress").Count(&stats.InProgress)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "completed").Count(&stats.Completed)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "cancelled").Count(&stats.Cancelled)

	// 按优先级统计
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "low").Count(&stats.LowPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "medium").Count(&stats.MediumPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "high").Count(&stats.HighPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "urgent").Count(&stats.UrgentPriority)

	utils.Success(c, stats)
}

// UpdateRequirementStatus 更新需求状态
func (h *RequirementHandler) UpdateRequirementStatus(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证状态
	validStatuses := map[string]bool{
		"pending":    true,
		"in_progress": true,
		"completed":  true,
		"cancelled":  true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	requirement.Status = req.Status
	if err := h.db.Save(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Product").Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

	utils.Success(c, requirement)
}

