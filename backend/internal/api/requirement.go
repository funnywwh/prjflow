package api

import (
	"fmt"
	"time"

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
	query := h.db.Preload("Project").Preload("Creator").Preload("Assignee")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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
	if err := h.db.Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	utils.Success(c, requirement)
}

// CreateRequirement 创建需求
func (h *RequirementHandler) CreateRequirement(c *gin.Context) {
	var req struct {
		Title          string   `json:"title" binding:"required"`
		Description    string   `json:"description"`
		Status         string   `json:"status"`
		Priority       string   `json:"priority"`
		ProjectID      uint     `json:"project_id" binding:"required"`
		AssigneeID     *uint    `json:"assignee_id"`
		EstimatedHours *float64 `json:"estimated_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
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

	// 验证项目是否存在（项目ID是必填的）
	var project model.Project
	if err := h.db.First(&project, req.ProjectID).Error; err != nil {
		utils.Error(c, 400, "项目不存在")
		return
	}

	// 如果指定了负责人，验证用户是否存在
	if req.AssigneeID != nil {
		var user model.User
		if err := h.db.First(&user, *req.AssigneeID).Error; err != nil {
			utils.Error(c, 400, "负责人不存在")
			return
		}
	}

	// 验证预估工时
	if req.EstimatedHours != nil && *req.EstimatedHours < 0 {
		utils.Error(c, 400, "预估工时不能为负数")
		return
	}

	requirement := model.Requirement{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		ProjectID:      req.ProjectID, // 必填
		CreatorID:      userID.(uint),
		AssigneeID:     req.AssigneeID,
		EstimatedHours: req.EstimatedHours,
	}

	if err := h.db.Create(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

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
		Title          *string  `json:"title"`
		Description    *string  `json:"description"`
		Status         *string  `json:"status"`
		Priority       *string  `json:"priority"`
		ProjectID      *uint    `json:"project_id"` // 更新时可选，但如果提供则必须有效
		AssigneeID     *uint    `json:"assignee_id"`
		EstimatedHours *float64 `json:"estimated_hours"`
		ActualHours    *float64 `json:"actual_hours"` // 实际工时，会自动创建资源分配
		WorkDate       *string  `json:"work_date"`    // 工作日期（YYYY-MM-DD），用于资源分配
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
	if req.ProjectID != nil {
		// 验证项目是否存在（项目ID不能为0，且必须存在）
		if *req.ProjectID == 0 {
			utils.Error(c, 400, "项目ID不能为空")
			return
		}
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 400, "项目不存在")
			return
		}
		requirement.ProjectID = *req.ProjectID
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
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			utils.Error(c, 400, "预估工时不能为负数")
			return
		}
		requirement.EstimatedHours = req.EstimatedHours
	}

	// 如果更新了实际工时，自动创建或更新资源分配
	if req.ActualHours != nil {
		if *req.ActualHours < 0 {
			utils.Error(c, 400, "实际工时不能为负数")
			return
		}
		// 需求必须有项目ID和负责人才能创建资源分配（ProjectID现在是必填的，但为了安全还是检查一下）
		if requirement.ProjectID == 0 {
			utils.Error(c, 400, "需求必须关联项目才能记录工时")
			return
		}
		if requirement.AssigneeID == nil {
			utils.Error(c, 400, "需求必须有负责人才能记录工时")
			return
		}
		// 确定工作日期
		var workDate time.Time
		if req.WorkDate != nil && *req.WorkDate != "" {
			if t, err := time.Parse("2006-01-02", *req.WorkDate); err == nil {
				workDate = t
			} else {
				utils.Error(c, 400, "工作日期格式错误，应为 YYYY-MM-DD")
				return
			}
		} else {
			// 默认使用今天
			workDate = time.Now()
		}
		workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())
		
		// 同步到资源分配
		if err := h.syncRequirementActualHours(&requirement, *req.ActualHours, workDate); err != nil {
			utils.Error(c, utils.CodeError, "同步资源分配失败: "+err.Error())
			return
		}
		// 从资源分配中汇总实际工时（确保actual_hours正确）
		h.calculateAndUpdateActualHours(&requirement)
	}

	if err := h.db.Save(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

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
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

	utils.Success(c, requirement)
}

// syncRequirementActualHours 同步需求实际工时到资源分配
func (h *RequirementHandler) syncRequirementActualHours(requirement *model.Requirement, actualHours float64, workDate time.Time) error {
	// ProjectID现在是必填的，但AssigneeID仍然是可选的
	if requirement.AssigneeID == nil {
		return nil // 没有负责人时，不创建资源分配，但不报错
	}

	// 查找或创建资源
	var resource model.Resource
	err := h.db.Where("user_id = ? AND project_id = ?", *requirement.AssigneeID, requirement.ProjectID).First(&resource).Error
	if err != nil {
		// 资源不存在，创建资源
		resource = model.Resource{
			UserID:    *requirement.AssigneeID,
			ProjectID: requirement.ProjectID,
		}
		if err := h.db.Create(&resource).Error; err != nil {
			return err
		}
	}

	// 查找是否已存在该需求和日期的资源分配
	// 先删除可能存在的重复记录（确保同一天只有一条记录）
	h.db.Where("resource_id = ? AND requirement_id = ? AND date = ?", resource.ID, requirement.ID, workDate).
		Delete(&model.ResourceAllocation{})
	
	// 创建新的资源分配记录
	allocation := model.ResourceAllocation{
		ResourceID:    resource.ID,
		RequirementID: &requirement.ID,
		ProjectID:     &requirement.ProjectID, // ProjectID现在是uint，需要取地址
		Date:          workDate,
		Hours:         actualHours,
		Description:   fmt.Sprintf("需求: %s", requirement.Title),
	}
	if err := h.db.Create(&allocation).Error; err != nil {
		return err
	}

	return nil
}

// calculateAndUpdateActualHours 计算并更新需求的实际工时（从资源分配中汇总）
func (h *RequirementHandler) calculateAndUpdateActualHours(requirement *model.Requirement) {
	// 先清理重复记录：对于同一个需求、同一个资源、同一天，只保留一条记录（保留最新的）
	var duplicateAllocations []model.ResourceAllocation
	h.db.Model(&model.ResourceAllocation{}).
		Where("requirement_id = ?", requirement.ID).
		Order("created_at DESC").
		Find(&duplicateAllocations)
	
	// 使用 map 记录已处理的 (resource_id, date) 组合
	seen := make(map[string]bool)
	var toDelete []uint
	for _, alloc := range duplicateAllocations {
		if alloc.RequirementID == nil {
			continue
		}
		key := fmt.Sprintf("%d_%s", alloc.ResourceID, alloc.Date.Format("2006-01-02"))
		if seen[key] {
			// 重复记录，标记为删除
			toDelete = append(toDelete, alloc.ID)
		} else {
			seen[key] = true
		}
	}
	
	// 删除重复记录
	if len(toDelete) > 0 {
		h.db.Where("id IN ?", toDelete).Delete(&model.ResourceAllocation{})
	}
	
	// 重新计算总工时
	var totalHours float64
	h.db.Model(&model.ResourceAllocation{}).
		Where("requirement_id = ?", requirement.ID).
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalHours)

	requirement.ActualHours = &totalHours
	h.db.Model(requirement).Update("actual_hours", totalHours)
}
