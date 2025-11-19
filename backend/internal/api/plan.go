package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type PlanHandler struct {
	db *gorm.DB
}

func NewPlanHandler(db *gorm.DB) *PlanHandler {
	return &PlanHandler{db: db}
}

// GetPlans 获取计划列表
func (h *PlanHandler) GetPlans(c *gin.Context) {
	var plans []model.Plan
	query := h.db.Preload("Project").Preload("Creator").Preload("Executions")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 类型筛选
	if planType := c.Query("type"); planType != "" {
		query = query.Where("type = ?", planType)
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 产品筛选

	// 项目筛选
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
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
	query.Model(&model.Plan{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&plans).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	// 计算每个计划的进度（基于执行进度）
	// 注意：进度会在响应中动态计算，不修改plan对象

	// 为每个计划添加进度字段
	planList := make([]gin.H, 0, len(plans))
	for _, plan := range plans {
		progress := h.calculatePlanProgress(plan)
		planList = append(planList, gin.H{
			"id":          plan.ID,
			"name":        plan.Name,
			"description": plan.Description,
			"type":        plan.Type,
			"status":      plan.Status,
			"start_date":  plan.StartDate,
			"end_date":    plan.EndDate,
			"project_id":  plan.ProjectID,
			"project":     plan.Project,
			"creator_id":  plan.CreatorID,
			"creator":     plan.Creator,
			"executions":  plan.Executions,
			"progress":    progress,
			"created_at":  plan.CreatedAt,
			"updated_at":  plan.UpdatedAt,
		})
	}

	utils.Success(c, gin.H{
		"list":      planList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPlan 获取计划详情
func (h *PlanHandler) GetPlan(c *gin.Context) {
	id := c.Param("id")
	var plan model.Plan
	if err := h.db.Preload("Project").Preload("Creator").
		Preload("Executions", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Assignee").Order("created_at ASC")
		}).
		First(&plan, id).Error; err != nil {
		utils.Error(c, 404, "计划不存在")
		return
	}

	// 计算计划进度
	progress := h.calculatePlanProgress(plan)
	planResponse := gin.H{
		"id":          plan.ID,
		"name":        plan.Name,
		"description": plan.Description,
		"type":        plan.Type,
		"status":      plan.Status,
		"start_date":  plan.StartDate,
		"end_date":    plan.EndDate,
		"project_id":  plan.ProjectID,
		"project":     plan.Project,
		"creator_id":  plan.CreatorID,
		"creator":     plan.Creator,
		"executions":  plan.Executions,
		"progress":    progress,
		"created_at":  plan.CreatedAt,
		"updated_at":  plan.UpdatedAt,
	}

	utils.Success(c, planResponse)
}

// CreatePlan 创建计划
func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Type        string  `json:"type" binding:"required"`
		Status      string  `json:"status"`
		ProjectID   *uint   `json:"project_id"`
		StartDate   *string `json:"start_date"` // 接收字符串格式的日期
		EndDate     *string `json:"end_date"`   // 接收字符串格式的日期
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 解析日期
	var startDate, endDate *time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			endDate = &t
		}
	}

	// 验证计划类型（只保留项目计划）
	if req.Type != "project_plan" {
		utils.Error(c, 400, "无效的计划类型，只支持项目计划")
		return
	}

	// 验证状态
	if req.Status == "" {
		req.Status = "draft"
	}
	if !isValidPlanStatus(req.Status) {
		utils.Error(c, 400, "无效的计划状态")
		return
	}

	// 验证关联关系
	if req.ProjectID == nil {
		utils.Error(c, 400, "项目计划必须关联项目")
		return
	}

	// 验证项目是否存在
	if req.ProjectID != nil {
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 404, "项目不存在")
			return
		}
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未登录")
		return
	}
	uid := userID.(uint)

	plan := model.Plan{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		ProjectID:   req.ProjectID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatorID:   uid,
	}

	if err := h.db.Create(&plan).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Executions").First(&plan, plan.ID)

	utils.Success(c, plan)
}

// UpdatePlan 更新计划
func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var plan model.Plan
	if err := h.db.First(&plan, id).Error; err != nil {
		utils.Error(c, 404, "计划不存在")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		StartDate   *string `json:"start_date"` // 接收字符串格式的日期
		EndDate     *string `json:"end_date"`    // 接收字符串格式的日期
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Status != nil {
		if !isValidPlanStatus(*req.Status) {
			utils.Error(c, 400, "无效的计划状态")
			return
		}
		plan.Status = *req.Status
	}
	// 解析日期
	if req.StartDate != nil && *req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			plan.StartDate = &t
		}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			plan.EndDate = &t
		}
	}

	if err := h.db.Save(&plan).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Executions").First(&plan, plan.ID)

	utils.Success(c, plan)
}

// DeletePlan 删除计划
func (h *PlanHandler) DeletePlan(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有执行
	var count int64
	h.db.Model(&model.PlanExecution{}).Where("plan_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "计划下有执行，无法删除")
		return
	}

	if err := h.db.Delete(&model.Plan{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// UpdatePlanStatus 更新计划状态
func (h *PlanHandler) UpdatePlanStatus(c *gin.Context) {
	id := c.Param("id")
	var plan model.Plan
	if err := h.db.First(&plan, id).Error; err != nil {
		utils.Error(c, 404, "计划不存在")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if !isValidPlanStatus(req.Status) {
		utils.Error(c, 400, "无效的计划状态")
		return
	}

	plan.Status = req.Status

	if err := h.db.Save(&plan).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新状态失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Executions").First(&plan, plan.ID)

	utils.Success(c, plan)
}

// calculatePlanProgress 计算计划进度（基于执行进度）
func (h *PlanHandler) calculatePlanProgress(plan model.Plan) int {
	if len(plan.Executions) == 0 {
		return 0
	}

	totalProgress := 0
	for _, execution := range plan.Executions {
		totalProgress += execution.Progress
	}
	progress := totalProgress / len(plan.Executions)

	// 如果所有执行都完成，自动更新计划状态为已完成
	if plan.Status != "completed" && plan.Status != "cancelled" {
		allCompleted := true
		for _, execution := range plan.Executions {
			if execution.Status != "completed" && execution.Status != "cancelled" {
				allCompleted = false
				break
			}
		}
		if allCompleted && len(plan.Executions) > 0 {
			h.db.Model(&plan).Update("status", "completed")
		}
	}

	return progress
}

// isValidPlanStatus 检查计划状态是否合法
func isValidPlanStatus(status string) bool {
	switch status {
	case "draft", "active", "completed", "cancelled":
		return true
	}
	return false
}

