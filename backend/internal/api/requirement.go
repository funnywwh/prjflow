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

	// 权限过滤：普通用户只能看到自己创建或参与的需求
	query = utils.FilterRequirementsByUser(h.db, c, query)

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
	// 计算总数时需要应用与查询相同的筛选条件
	countQuery := utils.FilterRequirementsByUser(h.db, c, h.db.Model(&model.Requirement{}))

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		countQuery = countQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 项目筛选
	if projectID := c.Query("project_id"); projectID != "" {
		countQuery = countQuery.Where("project_id = ?", projectID)
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}

	// 优先级筛选
	if priority := c.Query("priority"); priority != "" {
		countQuery = countQuery.Where("priority = ?", priority)
	}

	// 负责人筛选
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		countQuery = countQuery.Where("assignee_id = ?", assigneeID)
	}

	// 创建人筛选
	if creatorID := c.Query("creator_id"); creatorID != "" {
		countQuery = countQuery.Where("creator_id = ?", creatorID)
	}

	countQuery.Count(&total)

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
	if err := h.db.Preload("Project").Preload("Creator").Preload("Assignee").
		Preload("Attachments").Preload("Attachments.Creator").
		First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	// 权限检查：普通用户只能查看自己创建或参与的需求
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限访问该需求")
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
		req.Status = "draft"
	}
	validStatuses := map[string]bool{
		"draft":     true,
		"reviewing": true,
		"active":    true,
		"changing":  true,
		"closed":    true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效，有效值：draft, reviewing, active, changing, closed")
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

	// 权限检查：普通用户只能在自己参与的项目中创建需求
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限在该项目中创建需求")
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

	// 记录创建操作（userID已在上面定义）
	if userID, exists := c.Get("user_id"); exists {
		dbValue, _ := c.Get("db")
		if db, ok := dbValue.(*gorm.DB); ok {
			utils.RecordAction(db, "requirement", requirement.ID, "created", userID.(uint), "", nil)
		}
	}

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

	// 权限检查：普通用户只能更新自己创建或参与的需求
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限更新该需求")
		return
	}

	// 保存旧对象用于比较
	oldRequirement := requirement

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
		AttachmentIDs  *[]uint  `json:"attachment_ids"` // 附件ID列表
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
			"draft":     true,
			"reviewing": true,
			"active":    true,
			"changing":  true,
			"closed":    true,
		}
		if !validStatuses[*req.Status] {
			utils.Error(c, 400, "状态值无效，有效值：draft, reviewing, active, changing, closed")
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

	// 更新附件关联
	if req.AttachmentIDs != nil {
		projectID := requirement.ProjectID
		if req.ProjectID != nil {
			projectID = *req.ProjectID
		}
		var attachments []model.Attachment
		if len(*req.AttachmentIDs) > 0 {
			// 验证附件是否存在且属于同一项目
			if err := h.db.Where("id IN ?", *req.AttachmentIDs).Find(&attachments).Error; err != nil {
				utils.Error(c, 400, "附件查询失败: "+err.Error())
				return
			}
			if len(attachments) != len(*req.AttachmentIDs) {
				// 检查是否有附件被软删除
				var deletedAttachments []model.Attachment
				h.db.Unscoped().Where("id IN ? AND deleted_at IS NOT NULL", *req.AttachmentIDs).Find(&deletedAttachments)
				if len(deletedAttachments) > 0 {
					utils.Error(c, 400, fmt.Sprintf("部分附件已被删除：期望 %d 个，实际找到 %d 个，已删除 %d 个", len(*req.AttachmentIDs), len(attachments), len(deletedAttachments)))
					return
				}
				utils.Error(c, 400, fmt.Sprintf("附件不存在：期望 %d 个，实际找到 %d 个", len(*req.AttachmentIDs), len(attachments)))
				return
			}
			// 验证附件是否属于同一项目（通过检查附件是否关联到项目）
			for _, attachment := range attachments {
				var count int64
				if err := h.db.Table("project_attachments").
					Where("attachment_id = ? AND project_id = ?", attachment.ID, projectID).
					Count(&count).Error; err != nil {
					utils.Error(c, 400, "验证附件项目关联失败: "+err.Error())
					return
				}
				if count == 0 {
					utils.Error(c, 400, fmt.Sprintf("附件 %d 不属于项目 %d", attachment.ID, projectID))
					return
				}
			}
		}
		// 使用Replace方法同步附件关联（空数组表示移除所有附件）
		if err := h.db.Model(&requirement).Association("Attachments").Replace(attachments); err != nil {
			utils.Error(c, utils.CodeError, "更新附件关联失败: "+err.Error())
			return
		}
	}

	// 重新加载关联数据（包含附件）
	h.db.Session(&gorm.Session{}).Preload("Project").Preload("Creator").Preload("Assignee").
		Preload("Attachments").Preload("Attachments.Creator").
		First(&requirement, requirement.ID)

	// 记录编辑操作和字段变更
	userID, exists := c.Get("user_id")
	if exists {
		dbValue, _ := c.Get("db")
		if db, ok := dbValue.(*gorm.DB); ok {
			// 比较新旧对象并记录变更
			utils.CompareAndRecord(db, oldRequirement, requirement, "requirement", requirement.ID, userID.(uint), "edited")
		}
	}

	utils.Success(c, requirement)
}

// DeleteRequirement 删除需求
func (h *RequirementHandler) DeleteRequirement(c *gin.Context) {
	id := c.Param("id")

	// 验证需求是否存在
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	// 权限检查：普通用户只能删除自己创建或参与的需求
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限删除该需求")
		return
	}

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

	// 权限过滤：普通用户只能看到自己创建或参与的需求
	baseQuery = utils.FilterRequirementsByUser(h.db, c, baseQuery)

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

	// 按状态统计（禅道状态值）
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "draft").Count(&stats.Pending)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "reviewing").Count(&stats.Pending)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "active").Count(&stats.InProgress)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "changing").Count(&stats.InProgress)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "closed").Count(&stats.Completed)
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

	// 权限检查：普通用户只能更新自己创建或参与的需求
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限更新该需求")
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
		"draft":     true,
		"reviewing": true,
		"active":    true,
		"changing":  true,
		"closed":    true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效，有效值：draft, reviewing, active, changing, closed")
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
// 使用事务和 FirstOrCreate 防止并发死锁，替代先删除再创建的模式
func (h *RequirementHandler) syncRequirementActualHours(requirement *model.Requirement, actualHours float64, workDate time.Time) error {
	// ProjectID现在是必填的，但AssigneeID仍然是可选的
	if requirement.AssigneeID == nil {
		return nil // 没有负责人时，不创建资源分配，但不报错
	}

	// 使用事务包裹所有操作，防止死锁
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 使用 FirstOrCreate 查找或创建资源，避免并发创建冲突
	var resource model.Resource
	err := tx.Where("user_id = ? AND project_id = ?", *requirement.AssigneeID, requirement.ProjectID).
		FirstOrCreate(&resource, model.Resource{
			UserID:    *requirement.AssigneeID,
			ProjectID: requirement.ProjectID,
		}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查找或创建资源失败: %w", err)
	}

	// 使用 FirstOrCreate 查找或创建资源分配，避免并发创建冲突
	// 替代先删除再创建的模式，防止死锁
	var allocation model.ResourceAllocation
	err = tx.Where("resource_id = ? AND requirement_id = ? AND date = ?", resource.ID, requirement.ID, workDate).
		FirstOrCreate(&allocation, model.ResourceAllocation{
			ResourceID:    resource.ID,
			RequirementID: &requirement.ID,
			ProjectID:     &requirement.ProjectID,
			Date:          workDate,
			Hours:         actualHours,
			Description:   fmt.Sprintf("需求: %s", requirement.Title),
		}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查找或创建资源分配失败: %w", err)
	}

	// 无论记录是新创建还是已存在，都更新工时和描述（确保数据同步）
	allocation.Hours = actualHours
	allocation.Description = fmt.Sprintf("需求: %s", requirement.Title)
	if err := tx.Save(&allocation).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新资源分配失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// calculateAndUpdateActualHours 计算并更新需求的实际工时（从资源分配中汇总）
// 使用事务包裹所有操作，防止并发死锁
func (h *RequirementHandler) calculateAndUpdateActualHours(requirement *model.Requirement) {
	// 使用事务包裹所有操作
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 先清理重复记录：对于同一个需求、同一个资源、同一天，只保留一条记录（保留最新的）
	var duplicateAllocations []model.ResourceAllocation
	if err := tx.Model(&model.ResourceAllocation{}).
		Where("requirement_id = ?", requirement.ID).
		Order("created_at DESC").
		Find(&duplicateAllocations).Error; err != nil {
		tx.Rollback()
		return // 查询失败时静默返回，避免影响主流程
	}
	
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
		if err := tx.Where("id IN ?", toDelete).Delete(&model.ResourceAllocation{}).Error; err != nil {
			tx.Rollback()
			return // 删除失败时静默返回，避免影响主流程
		}
	}
	
	// 重新计算总工时
	var totalHours float64
	if err := tx.Model(&model.ResourceAllocation{}).
		Where("requirement_id = ?", requirement.ID).
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalHours).Error; err != nil {
		tx.Rollback()
		return // 查询失败时静默返回，避免影响主流程
	}

	requirement.ActualHours = &totalHours
	if err := tx.Model(requirement).Update("actual_hours", totalHours).Error; err != nil {
		tx.Rollback()
		return // 更新失败时静默返回，避免影响主流程
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		// 提交失败时静默返回，避免影响主流程
		return
	}
}

// GetRequirementHistory 获取需求历史记录列表
func (h *RequirementHandler) GetRequirementHistory(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	// 权限检查：普通用户只能查看自己创建或参与的需求的历史记录
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限查看该需求的历史记录")
		return
	}

	// 查询操作记录
	var actions []model.Action
	if err := h.db.Where("object_type = ? AND object_id = ?", "requirement", id).
		Preload("Actor").
		Preload("Histories").
		Order("date DESC").
		Find(&actions).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询历史记录失败")
		return
	}

	// 处理历史记录，转换字段值显示
	for i := range actions {
		for j := range actions[i].Histories {
			processedHistory := utils.ProcessHistory(h.db, &actions[i].Histories[j])
			actions[i].Histories[j] = *processedHistory
		}
	}

	utils.Success(c, gin.H{
		"list": actions,
	})
}

// AddRequirementHistoryNote 添加备注
func (h *RequirementHandler) AddRequirementHistoryNote(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	// 权限检查：普通用户只能为自己创建或参与的需求添加备注
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限为该需求添加备注")
		return
	}

	var req struct {
		Comment string `json:"comment" binding:"required"`
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

	// 记录备注操作
	dbValue, _ := c.Get("db")
	if db, ok := dbValue.(*gorm.DB); ok {
		_, err := utils.RecordAction(db, "requirement", requirement.ID, "commented", userID.(uint), req.Comment, nil)
		if err != nil {
			utils.Error(c, utils.CodeError, "添加备注失败")
			return
		}
	}

	utils.Success(c, gin.H{"message": "添加备注成功"})
}

// AssignRequirement 分配需求给用户
func (h *RequirementHandler) AssignRequirement(c *gin.Context) {
	id := c.Param("id")
	var requirement model.Requirement
	if err := h.db.First(&requirement, id).Error; err != nil {
		utils.Error(c, 404, "需求不存在")
		return
	}

	// 权限检查：普通用户只能分配自己创建或参与的需求
	if !utils.CheckRequirementAccess(h.db, c, requirement.ID) {
		utils.Error(c, 403, "没有权限分配该需求")
		return
	}

	// 获取旧的指派人ID
	oldAssigneeID := requirement.AssigneeID

	var req struct {
		AssigneeID uint    `json:"assignee_id" binding:"required"`
		Status     *string `json:"status"`
		Comment    *string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证用户是否存在
	var user model.User
	if err := h.db.First(&user, req.AssigneeID).Error; err != nil {
		utils.Error(c, 400, "用户不存在")
		return
	}

	// 更新指派人
	requirement.AssigneeID = &req.AssigneeID

	// 状态处理逻辑
	oldStatus := requirement.Status
	if req.Status != nil {
		// 如果提供了状态，使用提供的状态
		validStatuses := map[string]bool{
			"draft":     true,
			"reviewing": true,
			"active":    true,
			"changing":  true,
			"closed":    true,
		}
		if !validStatuses[*req.Status] {
			utils.Error(c, 400, "无效的状态值，有效值：draft, reviewing, active, changing, closed")
			return
		}
		requirement.Status = *req.Status
	} else {
		// 如果没有提供状态，自动修改：如果当前状态是 "draft" 或 "reviewing"，自动改为 "active"
		if requirement.Status == "draft" || requirement.Status == "reviewing" {
			requirement.Status = "active"
		}
	}

	// 保存更新
	if err := h.db.Save(&requirement).Error; err != nil {
		utils.Error(c, utils.CodeError, "分配失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").First(&requirement, requirement.ID)

	// 记录分配操作
	userID, exists := c.Get("user_id")
	if exists {
		dbValue, _ := c.Get("db")
		if db, ok := dbValue.(*gorm.DB); ok {
			// 记录分配操作
			actionID, _ := utils.RecordAction(db, "requirement", requirement.ID, "assigned", userID.(uint), "", nil)
			
			// 记录字段变更
			var changes []utils.HistoryChange
			
			// 记录指派人变更
			oldAssigneeIDStr := ""
			if oldAssigneeID != nil {
				oldAssigneeIDStr = fmt.Sprintf("%d", *oldAssigneeID)
			}
			newAssigneeIDStr := fmt.Sprintf("%d", req.AssigneeID)
			if oldAssigneeIDStr != newAssigneeIDStr {
				changes = append(changes, utils.HistoryChange{
					Field: "assignee_id",
					Old:   oldAssigneeIDStr,
					New:   newAssigneeIDStr,
				})
			}
			
			// 记录状态变更
			if oldStatus != requirement.Status {
				changes = append(changes, utils.HistoryChange{
					Field: "status",
					Old:   oldStatus,
					New:   requirement.Status,
				})
			}
			
			if len(changes) > 0 {
				utils.RecordHistory(db, actionID, changes)
			}

			// 如果提供了备注，记录备注操作
			if req.Comment != nil && *req.Comment != "" {
				_, err := utils.RecordAction(db, "requirement", requirement.ID, "commented", userID.(uint), *req.Comment, nil)
				if err != nil {
					utils.Error(c, utils.CodeError, "添加备注失败")
					return
				}
			}
		}
	}

	utils.Success(c, requirement)
}
