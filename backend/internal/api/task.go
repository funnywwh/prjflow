package api

import (
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
	query := h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies")

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
	if err := h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, id).Error; err != nil {
		utils.Error(c, 404, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Title         string    `json:"title" binding:"required"`
		Description   string    `json:"description"`
		Status        string    `json:"status"`
		Priority      string    `json:"priority"`
		ProjectID     uint      `json:"project_id" binding:"required"`
		AssigneeID    *uint     `json:"assignee_id"`
		StartDate     *string   `json:"start_date"`
		EndDate       *string   `json:"end_date"`
		DueDate       *string   `json:"due_date"`
		Progress      int       `json:"progress"`
		DependencyIDs []uint    `json:"dependency_ids"`
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
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		ProjectID:   req.ProjectID,
		CreatorID:   userID.(uint),
		AssigneeID:  req.AssigneeID,
		StartDate:   startDate,
		EndDate:     endDate,
		DueDate:     dueDate,
		Progress:    req.Progress,
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
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

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
		Title         *string  `json:"title"`
		Description   *string  `json:"description"`
		Status        *string  `json:"status"`
		Priority      *string  `json:"priority"`
		ProjectID     *uint    `json:"project_id"`
		AssigneeID    *uint    `json:"assignee_id"`
		StartDate     *string  `json:"start_date"`
		EndDate       *string  `json:"end_date"`
		DueDate       *string  `json:"due_date"`
		Progress      *int     `json:"progress"`
		DependencyIDs *[]uint  `json:"dependency_ids"`
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

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

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
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

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
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

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
		Progress int `json:"progress" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证进度
	if req.Progress < 0 || req.Progress > 100 {
		utils.Error(c, 400, "进度值必须在0-100之间")
		return
	}

	task.Progress = req.Progress
	// 如果进度为100，自动设置状态为done
	if req.Progress == 100 {
		task.Status = "done"
	}
	// 如果进度大于0且状态为todo，自动设置为in_progress
	if req.Progress > 0 && task.Status == "todo" {
		task.Status = "in_progress"
	}

	if err := h.db.Save(&task).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignee").Preload("Dependencies").First(&task, task.ID)

	utils.Success(c, task)
}

