package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type TaskHandler struct {
	db *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

// GetTasks 获取任务列表
func (h *TaskHandler) GetTasks(c *gin.Context) {
	var tasks []model.Task
	query := h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies")

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
	query.Model(&model.Task{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, id).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title          string    `json:"title" binding:"required"`
		Description    string    `json:"description"`
		Status         string    `json:"status"`
		Priority       string    `json:"priority"`
		ProjectID      uint      `json:"project_id" binding:"required"`
		RequirementID  *uint     `json:"requirement_id"`
		AssigneeID     *uint     `json:"assignee_id"`
		StartDate      *string   `json:"start_date"`
		EndDate        *string   `json:"end_date"`
		DueDate        *string   `json:"due_date"`
		Progress       int       `json:"progress"`
		EstimatedHours *float64  `json:"estimated_hours"`
		DependencyIDs  []uint    `json:"dependency_ids"`
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
		req.Status = "todo"
	}
	validStatuses := map[string]bool{
		"todo":       true,
		"in_progress": true,
		"done":       true,
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

	// 验证进度
	if req.Progress < 0 || req.Progress > 100 {
		utils.Error(c, 400, "进度值必须在0-100之间")
		return
	}

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, req.ProjectID).Error; err != nil {
		utils.Error(c, 400, "项目不存在")
		return
	}

	// 如果指定了需求，验证需求是否存在且属于同一项目
	if req.RequirementID != nil {
		var requirement model.Requirement
		if err := h.db.First(&requirement, *req.RequirementID).Error; err != nil {
			utils.Error(c, 400, "需求不存在")
			return
		}
		if requirement.ProjectID != req.ProjectID {
			utils.Error(c, 400, "需求必须属于同一项目")
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

	// 解析日期
	var startDate, endDate, dueDate *time.Time
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
	if req.DueDate != nil && *req.DueDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			dueDate = &t
		}
	}

	task := model.Task{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		ProjectID:      req.ProjectID,
		RequirementID:  req.RequirementID,
		CreatorID:      userID.(uint),
		AssigneeID:     req.AssigneeID,
		StartDate:      startDate,
		EndDate:        endDate,
		DueDate:        dueDate,
		Progress:       req.Progress,
		EstimatedHours: req.EstimatedHours,
	}

	if err := h.db.Create(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 设置任务依赖关系
	if len(req.DependencyIDs) > 0 {
		var dependencies []model.Task
		if err := h.db.Where("id IN ?", req.DependencyIDs).Find(&dependencies).Error; err != nil {
			utils.Error(c, 400, "依赖任务不存在")
			return
		}
		// 检查循环依赖
		for _, depID := range req.DependencyIDs {
			if depID == task.ID {
				utils.Error(c, 400, "任务不能依赖自己")
				return
			}
		}
		if err := h.db.Model(&task).Association("Dependencies").Replace(dependencies); err != nil {
			utils.Error(c, utils.CodeError, "设置依赖失败")
			return
		}
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := h.db.First(&task, id).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
		return
	}

	var req struct {
		Title          *string  `json:"title"`
		Description    *string  `json:"description"`
		Status         *string  `json:"status"`
		Priority       *string  `json:"priority"`
		ProjectID      *uint    `json:"project_id"`
		RequirementID  *uint    `json:"requirement_id"`
		AssigneeID     *uint    `json:"assignee_id"`
		StartDate      *string  `json:"start_date"`
		EndDate        *string  `json:"end_date"`
		DueDate        *string  `json:"due_date"`
		Progress       *int     `json:"progress"`
		EstimatedHours *float64 `json:"estimated_hours"`
		ActualHours    *float64 `json:"actual_hours"` // 实际工时，会自动创建资源分配
		WorkDate       *string  `json:"work_date"`     // 工作日期（YYYY-MM-DD），用于资源分配
		DependencyIDs  *[]uint  `json:"dependency_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 更新字段
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		// 验证状态
		validStatuses := map[string]bool{
			"todo":       true,
			"in_progress": true,
			"done":       true,
			"cancelled":  true,
		}
		if !validStatuses[*req.Status] {
			utils.Error(c, 400, "状态值无效")
			return
		}
		task.Status = *req.Status
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
		task.Priority = *req.Priority
	}
	if req.ProjectID != nil {
		// 验证项目是否存在
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 400, "项目不存在")
			return
		}
		task.ProjectID = *req.ProjectID
	}
	if req.RequirementID != nil {
		// 验证需求是否存在且属于同一项目
		if *req.RequirementID != 0 {
			var requirement model.Requirement
			if err := h.db.First(&requirement, *req.RequirementID).Error; err != nil {
				utils.Error(c, 400, "需求不存在")
				return
			}
			if requirement.ProjectID != task.ProjectID {
				utils.Error(c, 400, "需求必须属于同一项目")
				return
			}
			task.RequirementID = req.RequirementID
		} else {
			task.RequirementID = nil
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
			task.AssigneeID = req.AssigneeID
		} else {
			task.AssigneeID = nil
		}
	}
	if req.StartDate != nil {
		if *req.StartDate != "" {
			if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
				task.StartDate = &t
			}
		} else {
			task.StartDate = nil
		}
	}
	if req.EndDate != nil {
		if *req.EndDate != "" {
			if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
				task.EndDate = &t
			}
		} else {
			task.EndDate = nil
		}
	}
	if req.DueDate != nil {
		if *req.DueDate != "" {
			if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
				task.DueDate = &t
			}
		} else {
			task.DueDate = nil
		}
	}
	if req.Progress != nil {
		if *req.Progress < 0 || *req.Progress > 100 {
			utils.Error(c, 400, "进度值必须在0-100之间")
			return
		}
		task.Progress = *req.Progress
	}
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			utils.Error(c, 400, "预估工时不能为负数")
			return
		}
		task.EstimatedHours = req.EstimatedHours
	}
	// 如果更新了实际工时，自动创建或更新资源分配
	if req.ActualHours != nil {
		if *req.ActualHours < 0 {
			utils.Error(c, 400, "实际工时不能为负数")
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
			// 默认使用任务的开始日期或结束日期，如果都没有则使用今天
			if task.StartDate != nil {
				workDate = *task.StartDate
			} else if task.EndDate != nil {
				workDate = *task.EndDate
			} else {
				workDate = time.Now()
			}
		}
		workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())
		
		// 同步到资源分配
		if err := h.syncTaskActualHours(&task, *req.ActualHours, workDate); err != nil {
			utils.Error(c, utils.CodeError, "同步资源分配失败: "+err.Error())
			return
		}
	}

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 计算并更新实际工时（从资源分配中汇总）
	h.calculateAndUpdateActualHours(&task)
	
	// 根据实际工时和预估工时自动计算进度
	h.calculateProgressFromHours(&task)

	// 更新任务依赖关系
	if req.DependencyIDs != nil {
		var dependencies []model.Task
		if len(*req.DependencyIDs) > 0 {
			// 检查循环依赖
			for _, depID := range *req.DependencyIDs {
				if depID == task.ID {
					utils.Error(c, 400, "任务不能依赖自己")
					return
				}
			}
			if err := h.db.Where("id IN ?", *req.DependencyIDs).Find(&dependencies).Error; err != nil {
				utils.Error(c, 400, "依赖任务不存在")
				return
			}
		}
		if err := h.db.Model(&task).Association("Dependencies").Replace(dependencies); err != nil {
			utils.Error(c, utils.CodeError, "更新依赖失败")
			return
		}
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有其他任务依赖此任务
	var count int64
	h.db.Model(&model.TaskDependency{}).Where("dependency_id = ?", id).Count(&count)
	if count > 0 {
		utils.Error(c, 400, "有其他任务依赖此任务，无法删除")
		return
	}

	if err := h.db.Delete(&model.Task{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// UpdateTaskStatus 更新任务状态
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := h.db.First(&task, id).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
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
		"todo":       true,
		"in_progress": true,
		"done":       true,
		"cancelled":  true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	task.Status = req.Status
	// 如果状态为done，自动设置进度为100
	if req.Status == "done" {
		task.Progress = 100
	}
	// 如果状态为cancelled，进度保持原值
	// 如果状态为in_progress且进度为0，可以设置一个默认值（可选）

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}

// UpdateTaskProgress 更新任务进度
func (h *TaskHandler) UpdateTaskProgress(c *gin.Context) {
	id := c.Param("id")
	var task model.Task
	if err := h.db.First(&task, id).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
		return
	}

	var req struct {
		Progress       *int     `json:"progress"`
		EstimatedHours *float64 `json:"estimated_hours"`
		ActualHours    *float64 `json:"actual_hours"` // 消耗工时
		WorkDate       *string  `json:"work_date"`    // 工作日期（YYYY-MM-DD）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 更新进度
	if req.Progress != nil {
		// 验证进度
		if *req.Progress < 0 || *req.Progress > 100 {
			utils.Error(c, 400, "进度值必须在0-100之间")
			return
		}
		task.Progress = *req.Progress
		// 如果进度为100，自动设置状态为done
		if *req.Progress == 100 {
			task.Status = "done"
		}
		// 如果进度大于0且状态为todo，自动设置为in_progress
		if *req.Progress > 0 && task.Status == "todo" {
			task.Status = "in_progress"
		}
	}

	// 更新预估工时
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			utils.Error(c, 400, "预估工时不能为负数")
			return
		}
		task.EstimatedHours = req.EstimatedHours
	}

	// 更新实际工时
	if req.ActualHours != nil {
		if *req.ActualHours < 0 {
			utils.Error(c, 400, "实际工时不能为负数")
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
			// 默认使用任务的开始日期或结束日期，如果都没有则使用今天
			if task.StartDate != nil {
				workDate = *task.StartDate
			} else if task.EndDate != nil {
				workDate = *task.EndDate
			} else {
				workDate = time.Now()
			}
		}
		workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())
		
		// 同步到资源分配
		if err := h.syncTaskActualHours(&task, *req.ActualHours, workDate); err != nil {
			utils.Error(c, utils.CodeError, "同步资源分配失败: "+err.Error())
			return
		}
	}

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 计算并更新实际工时（从资源分配中汇总）
	h.calculateAndUpdateActualHours(&task)
	
	// 如果更新了实际工时或预估工时，自动根据工时计算进度
	// 进度 = 实际工时 / 预估工时 * 100，范围0-100%
	if req.ActualHours != nil || req.EstimatedHours != nil {
		h.calculateProgressFromHours(&task)
	} else if req.Progress == nil {
		// 如果没有更新工时，且没有手动设置进度，则根据当前工时计算进度
		h.calculateProgressFromHours(&task)
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Requirement").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}


// syncTaskActualHours 同步任务实际工时到资源分配
func (h *TaskHandler) syncTaskActualHours(task *model.Task, actualHours float64, workDate time.Time) error {
	// 如果任务没有负责人，无法创建资源分配
	if task.AssigneeID == nil {
		return nil // 没有负责人时，不创建资源分配，但不报错
	}

	// 查找或创建资源
	var resource model.Resource
	err := h.db.Where("user_id = ? AND project_id = ?", *task.AssigneeID, task.ProjectID).First(&resource).Error
	if err != nil {
		// 资源不存在，创建资源
		resource = model.Resource{
			UserID:    *task.AssigneeID,
			ProjectID: task.ProjectID,
		}
		if err := h.db.Create(&resource).Error; err != nil {
			return err
		}
	}

	// 查找是否已存在该任务和日期的资源分配
	var allocation model.ResourceAllocation
	err = h.db.Where("resource_id = ? AND task_id = ? AND date = ?", resource.ID, task.ID, workDate).First(&allocation).Error
	if err != nil {
		// 不存在，创建新的资源分配
		allocation = model.ResourceAllocation{
			ResourceID:  resource.ID,
			TaskID:      &task.ID,
			ProjectID:   &task.ProjectID,
			Date:        workDate,
			Hours:       actualHours,
			Description: fmt.Sprintf("任务: %s", task.Title),
		}
		if err := h.db.Create(&allocation).Error; err != nil {
			return err
		}
	} else {
		// 存在，更新工时
		allocation.Hours = actualHours
		if err := h.db.Save(&allocation).Error; err != nil {
			return err
		}
	}

	return nil
}

// calculateAndUpdateActualHours 计算并更新任务的实际工时（从资源分配中汇总）
func (h *TaskHandler) calculateAndUpdateActualHours(task *model.Task) {
	var totalHours float64
	h.db.Model(&model.ResourceAllocation{}).
		Where("task_id = ?", task.ID).
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalHours)

	task.ActualHours = &totalHours
	h.db.Model(task).Update("actual_hours", totalHours)
}

// calculateProgressFromHours 根据实际工时和预估工时自动计算进度
func (h *TaskHandler) calculateProgressFromHours(task *model.Task) {
	// 如果预估工时未设置或为0，无法自动计算进度
	if task.EstimatedHours == nil || *task.EstimatedHours <= 0 {
		return
	}

	// 如果实际工时未设置，使用0
	actualHours := 0.0
	if task.ActualHours != nil {
		actualHours = *task.ActualHours
	}

	// 计算进度：实际工时 / 预估工时 * 100
	progress := int((actualHours / *task.EstimatedHours) * 100)
	
	// 进度不能超过100%
	if progress > 100 {
		progress = 100
	}
	
	// 如果进度小于0，设为0
	if progress < 0 {
		progress = 0
	}

	// 更新进度
	task.Progress = progress
	
	// 如果进度为100，自动设置状态为done
	if progress == 100 && task.Status != "done" {
		task.Status = "done"
	}
	// 如果进度大于0且状态为todo，自动设置为in_progress
	if progress > 0 && task.Status == "todo" {
		task.Status = "in_progress"
	}

	// 保存更新
	h.db.Model(task).Updates(map[string]interface{}{
		"progress": progress,
		"status":   task.Status,
	})
}
