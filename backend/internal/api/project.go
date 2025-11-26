package api

import (
	"time"

	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// GetProjects 获取项目列表
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	var projects []model.Project
	query := h.db.Preload("Tags")

	// 权限过滤：普通用户只能看到自己参与的项目
	query = utils.FilterProjectsByUser(h.db, c, query)

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 标签筛选（支持多个标签，OR逻辑：包含任意一个标签即可）
	// 支持两种方式：单个tag参数（向后兼容）和多个tags参数
	if tagIDs := c.QueryArray("tags"); len(tagIDs) > 0 {
		// 多个标签ID，使用OR逻辑：通过JOIN查询包含任意一个标签的项目
		query = query.Joins("JOIN project_tags ON project_tags.project_id = projects.id").
			Where("project_tags.tag_id IN ?", tagIDs).
			Group("projects.id")
	} else if tagID := c.Query("tag"); tagID != "" {
		// 单个标签ID（向后兼容）
		query = query.Joins("JOIN project_tags ON project_tags.project_id = projects.id").
			Where("project_tags.tag_id = ?", tagID).
			Group("projects.id")
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
	// 计算总数时需要去掉JOIN和GROUP BY的影响
	countQuery := h.db.Model(&model.Project{})
	// 权限过滤：普通用户只能看到自己参与的项目
	countQuery = utils.FilterProjectsByUser(h.db, c, countQuery)
	if keyword := c.Query("keyword"); keyword != "" {
		countQuery = countQuery.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if tagIDs := c.QueryArray("tags"); len(tagIDs) > 0 {
		countQuery = countQuery.Joins("JOIN project_tags ON project_tags.project_id = projects.id").
			Where("project_tags.tag_id IN ?", tagIDs).
			Group("projects.id")
	} else if tagID := c.Query("tag"); tagID != "" {
		countQuery = countQuery.Joins("JOIN project_tags ON project_tags.project_id = projects.id").
			Where("project_tags.tag_id = ?", tagID).
			Group("projects.id")
	}
	if status := c.Query("status"); status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

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
		Preload("Members.User").
		Preload("Tags").
		First(&project, id).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能查看自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	// 获取项目统计信息
	statistics := h.getProjectStatistics(project.ID)

	utils.Success(c, gin.H{
		"project":    project,
		"statistics": statistics,
	})
}

// GetProjectStatistics 获取项目统计信息
func (h *ProjectHandler) GetProjectStatistics(c *gin.Context) {
	id := c.Param("id")

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, id).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能查看自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	statistics := h.getProjectStatistics(project.ID)
	utils.Success(c, statistics)
}

// getProjectStatistics 获取项目统计信息（内部方法）
func (h *ProjectHandler) getProjectStatistics(projectID uint) gin.H {
	var taskCount, bugCount, requirementCount, memberCount int64
	var todoTaskCount, inProgressTaskCount, doneTaskCount int64
	var openBugCount, inProgressBugCount, resolvedBugCount int64
	var inProgressRequirementCount, completedRequirementCount int64

	// 任务统计
	h.db.Model(&model.Task{}).Where("project_id = ?", projectID).Count(&taskCount)
	h.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, "wait").Count(&todoTaskCount)
	h.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, "doing").Count(&inProgressTaskCount)
	h.db.Model(&model.Task{}).Where("project_id = ? AND status = ?", projectID, "done").Count(&doneTaskCount)

	// Bug统计
	h.db.Model(&model.Bug{}).Where("project_id = ?", projectID).Count(&bugCount)
	h.db.Model(&model.Bug{}).Where("project_id = ? AND status = ?", projectID, "active").Count(&openBugCount)
	h.db.Model(&model.Bug{}).Where("project_id = ? AND status = ?", projectID, "resolved").Count(&inProgressBugCount)
	h.db.Model(&model.Bug{}).Where("project_id = ? AND status = ?", projectID, "resolved").Count(&resolvedBugCount)

	// 需求统计
	h.db.Model(&model.Requirement{}).Where("project_id = ?", projectID).Count(&requirementCount)
	h.db.Model(&model.Requirement{}).Where("project_id = ? AND status = ?", projectID, "active").Count(&inProgressRequirementCount)
	h.db.Model(&model.Requirement{}).Where("project_id = ? AND status = ?", projectID, "closed").Count(&completedRequirementCount)

	// 成员统计
	h.db.Model(&model.ProjectMember{}).Where("project_id = ?", projectID).Count(&memberCount)

	return gin.H{
		"total_tasks":              int(taskCount),
		"todo_tasks":               int(todoTaskCount),
		"in_progress_tasks":        int(inProgressTaskCount),
		"done_tasks":               int(doneTaskCount),
		"total_bugs":               int(bugCount),
		"open_bugs":                int(openBugCount),
		"in_progress_bugs":         int(inProgressBugCount),
		"resolved_bugs":            int(resolvedBugCount),
		"total_requirements":       int(requirementCount),
		"in_progress_requirements": int(inProgressRequirementCount),
		"completed_requirements":   int(completedRequirementCount),
		"total_members":            int(memberCount),
	}
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Code        string  `json:"code"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		TagIDs      []uint  `json:"tag_ids"`    // 标签ID数组
		StartDate   *string `json:"start_date"` // 接收字符串格式的日期
		EndDate     *string `json:"end_date"`   // 接收字符串格式的日期
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证状态
	if req.Status == "" {
		req.Status = "wait"
	}
	if !isValidProjectStatus(req.Status) {
		utils.Error(c, 400, "状态值无效，有效值：wait, doing, suspended, closed, done")
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

	project := model.Project{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	// 关联标签
	if len(req.TagIDs) > 0 {
		var tags []model.Tag
		if err := h.db.Where("id IN ?", req.TagIDs).Find(&tags).Error; err != nil {
			utils.Error(c, utils.CodeError, "标签查询失败")
			return
		}
		if len(tags) != len(req.TagIDs) {
			utils.Error(c, 400, "部分标签不存在")
			return
		}
		project.Tags = tags
	}

	// 创建项目（GORM会自动处理多对多关联）
	if err := h.db.Create(&project).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
		return
	}

	// 自动将创建者添加为项目成员（角色：项目经理）
	userID := utils.GetUserID(c)
	if userID > 0 {
		// 检查是否已经是成员（防止重复）
		var existingMember model.ProjectMember
		if err := h.db.Where("project_id = ? AND user_id = ?", project.ID, userID).First(&existingMember).Error; err != nil {
			// 不存在，创建新成员
			creatorMember := model.ProjectMember{
				ProjectID: project.ID,
				UserID:    userID,
				Role:      "项目经理", // 创建者默认为项目经理
			}
			if err := h.db.Create(&creatorMember).Error; err != nil {
				// 如果添加成员失败，记录错误但不影响项目创建
				// 因为项目已经创建成功
			}
		}
	}

	// 重新加载关联数据
	if err := h.db.Preload("Members.User").Preload("Tags").First(&project, project.ID).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

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

	// 权限检查：普通用户只能更新自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限更新该项目")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Code        *string `json:"code"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		TagIDs      *[]uint `json:"tag_ids"`    // 标签ID数组
		StartDate   *string `json:"start_date"` // 接收字符串格式的日期
		EndDate     *string `json:"end_date"`   // 接收字符串格式的日期
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 更新字段
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Code != nil {
		project.Code = *req.Code
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		if !isValidProjectStatus(*req.Status) {
			utils.Error(c, 400, "状态值无效，有效值：wait, doing, suspended, closed, done")
			return
		}
		project.Status = *req.Status
	}

	// 更新标签关联
	if req.TagIDs != nil {
		var tags []model.Tag
		if len(*req.TagIDs) > 0 {
			if err := h.db.Where("id IN ?", *req.TagIDs).Find(&tags).Error; err != nil {
				utils.Error(c, utils.CodeError, "标签查询失败")
				return
			}
			if len(tags) != len(*req.TagIDs) {
				utils.Error(c, 400, "部分标签不存在")
				return
			}
		}
		// 使用Association替换标签
		h.db.Model(&project).Association("Tags").Replace(tags)
	}

	// 解析日期
	if req.StartDate != nil {
		if *req.StartDate != "" {
			if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
				project.StartDate = &t
			}
		} else {
			project.StartDate = nil
		}
	}
	if req.EndDate != nil {
		if *req.EndDate != "" {
			if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
				project.EndDate = &t
			}
		} else {
			project.EndDate = nil
		}
	}

	if err := h.db.Save(&project).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Members.User").First(&project, project.ID)

	utils.Success(c, project)
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, id).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能删除自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限删除该项目")
		return
	}

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

	// 验证项目是否存在并检查权限
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能查看自己参与的项目的成员
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	var members []model.ProjectMember
	if err := h.db.Where("project_id = ?", projectID).Preload("User").Preload("User.Department").Find(&members).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, members)
}

// GetProjectGantt 获取项目甘特图数据
func (h *ProjectHandler) GetProjectGantt(c *gin.Context) {
	projectID := c.Param("id")

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能查看自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	// 获取项目的所有任务（包含依赖关系）
	var tasks []model.Task
	if err := h.db.Where("project_id = ?", projectID).
		Preload("Dependencies").
		Preload("Assignee").
		Order("created_at ASC").
		Find(&tasks).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询任务失败")
		return
	}

	// 转换为甘特图数据格式
	type GanttTask struct {
		ID           uint   `json:"id"`
		Title        string `json:"title"`
		StartDate    string `json:"start_date,omitempty"`
		EndDate      string `json:"end_date,omitempty"`
		DueDate      string `json:"due_date,omitempty"`
		Progress     int    `json:"progress"`
		Status       string `json:"status"`
		Priority     string `json:"priority"`
		Assignee     string `json:"assignee,omitempty"`
		Dependencies []uint `json:"dependencies,omitempty"`
	}

	ganttTasks := make([]GanttTask, 0, len(tasks))
	for _, task := range tasks {
		ganttTask := GanttTask{
			ID:       task.ID,
			Title:    task.Title,
			Progress: task.Progress,
			Status:   task.Status,
			Priority: task.Priority,
		}

		// 格式化日期
		if task.StartDate != nil {
			ganttTask.StartDate = task.StartDate.Format("2006-01-02")
		}
		if task.EndDate != nil {
			ganttTask.EndDate = task.EndDate.Format("2006-01-02")
		}
		if task.DueDate != nil {
			ganttTask.DueDate = task.DueDate.Format("2006-01-02")
		}

		// 负责人信息
		if task.Assignee != nil {
			if task.Assignee.Nickname != "" {
				ganttTask.Assignee = task.Assignee.Username + "(" + task.Assignee.Nickname + ")"
			} else {
				ganttTask.Assignee = task.Assignee.Username
			}
		}

		// 依赖关系（转换为依赖任务ID列表）
		if len(task.Dependencies) > 0 {
			dependencyIDs := make([]uint, 0, len(task.Dependencies))
			for _, dep := range task.Dependencies {
				dependencyIDs = append(dependencyIDs, dep.ID)
			}
			ganttTask.Dependencies = dependencyIDs
		}

		ganttTasks = append(ganttTasks, ganttTask)
	}

	utils.Success(c, gin.H{
		"tasks": ganttTasks,
	})
}

// GetProjectProgress 获取项目进度跟踪数据
func (h *ProjectHandler) GetProjectProgress(c *gin.Context) {
	projectID := c.Param("id")

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 权限检查：普通用户只能查看自己参与的项目
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	// 获取基础统计
	statistics := h.getProjectStatistics(project.ID)

	// 获取任务进度趋势（最近30天）
	taskProgressTrend := h.getTaskProgressTrend(project.ID, 30)

	// 获取任务状态分布
	taskStatusDistribution := h.getTaskStatusDistribution(project.ID)

	// 获取任务优先级分布
	taskPriorityDistribution := h.getTaskPriorityDistribution(project.ID)

	// 获取任务完成率趋势（按周）
	taskCompletionTrend := h.getTaskCompletionTrend(project.ID, 12)

	// 获取成员工作量统计
	memberWorkload := h.getMemberWorkload(project.ID)

	// 获取Bug趋势（最近30天）
	bugTrend := h.getBugTrend(project.ID, 30)

	// 获取需求完成趋势
	requirementTrend := h.getRequirementTrend(project.ID, 30)

	utils.Success(c, gin.H{
		"statistics":                 statistics,
		"task_progress_trend":        taskProgressTrend,
		"task_status_distribution":   taskStatusDistribution,
		"task_priority_distribution": taskPriorityDistribution,
		"task_completion_trend":      taskCompletionTrend,
		"member_workload":            memberWorkload,
		"bug_trend":                  bugTrend,
		"requirement_trend":          requirementTrend,
	})
}

// getTaskProgressTrend 获取任务进度趋势
func (h *ProjectHandler) getTaskProgressTrend(projectID uint, days int) []gin.H {
	var tasks []model.Task
	h.db.Where("project_id = ? AND created_at >= ?", projectID, time.Now().AddDate(0, 0, -days)).
		Order("created_at ASC").
		Find(&tasks)

	// 按日期分组统计平均进度
	progressByDate := make(map[string][]int)
	for _, task := range tasks {
		date := task.CreatedAt.Format("2006-01-02")
		progressByDate[date] = append(progressByDate[date], task.Progress)
	}

	result := make([]gin.H, 0)
	for date, progresses := range progressByDate {
		sum := 0
		for _, p := range progresses {
			sum += p
		}
		avg := float64(sum) / float64(len(progresses))
		result = append(result, gin.H{
			"date":    date,
			"average": avg,
			"count":   len(progresses),
		})
	}

	return result
}

// getTaskStatusDistribution 获取任务状态分布
func (h *ProjectHandler) getTaskStatusDistribution(projectID uint) []gin.H {
	var tasks []model.Task
	h.db.Where("project_id = ?", projectID).Find(&tasks)

	statusCount := make(map[string]int)
	for _, task := range tasks {
		statusCount[task.Status]++
	}

	result := make([]gin.H, 0)
	for status, count := range statusCount {
		result = append(result, gin.H{
			"status": status,
			"count":  count,
		})
	}

	return result
}

// getTaskPriorityDistribution 获取任务优先级分布
func (h *ProjectHandler) getTaskPriorityDistribution(projectID uint) []gin.H {
	var tasks []model.Task
	h.db.Where("project_id = ?", projectID).Find(&tasks)

	priorityCount := make(map[string]int)
	for _, task := range tasks {
		priorityCount[task.Priority]++
	}

	result := make([]gin.H, 0)
	for priority, count := range priorityCount {
		result = append(result, gin.H{
			"priority": priority,
			"count":    count,
		})
	}

	return result
}

// getTaskCompletionTrend 获取任务完成率趋势（按周）
func (h *ProjectHandler) getTaskCompletionTrend(projectID uint, weeks int) []gin.H {
	result := make([]gin.H, 0)
	now := time.Now()

	for i := weeks - 1; i >= 0; i-- {
		weekStart := now.AddDate(0, 0, -i*7).Truncate(24 * time.Hour)
		weekEnd := weekStart.AddDate(0, 0, 7)

		var totalTasks, completedTasks int64
		h.db.Model(&model.Task{}).
			Where("project_id = ? AND created_at < ?", projectID, weekEnd).
			Count(&totalTasks)
		h.db.Model(&model.Task{}).
			Where("project_id = ? AND status = ? AND updated_at >= ? AND updated_at < ?", projectID, "done", weekStart, weekEnd).
			Count(&completedTasks)

		completionRate := 0.0
		if totalTasks > 0 {
			completionRate = float64(completedTasks) / float64(totalTasks) * 100
		}

		result = append(result, gin.H{
			"week":            weekStart.Format("2006-01-02"),
			"total":           totalTasks,
			"completed":       completedTasks,
			"completion_rate": completionRate,
		})
	}

	return result
}

// getMemberWorkload 获取成员工作量统计
func (h *ProjectHandler) getMemberWorkload(projectID uint) []gin.H {
	var tasks []model.Task
	h.db.Where("project_id = ? AND assignee_id IS NOT NULL", projectID).
		Preload("Assignee").
		Find(&tasks)

	memberWorkload := make(map[uint]gin.H)
	for _, task := range tasks {
		if task.AssigneeID == nil {
			continue
		}
		assigneeID := *task.AssigneeID
		if _, exists := memberWorkload[assigneeID]; !exists {
			memberWorkload[assigneeID] = gin.H{
				"user_id":     assigneeID,
				"username":    task.Assignee.Username,
				"nickname":    task.Assignee.Nickname,
				"total":       0,
				"completed":   0,
				"in_progress": 0,
			}
		}
		workload := memberWorkload[assigneeID]
		workload["total"] = workload["total"].(int) + 1
		if task.Status == "done" {
			workload["completed"] = workload["completed"].(int) + 1
		} else if task.Status == "doing" {
			workload["in_progress"] = workload["in_progress"].(int) + 1
		}
	}

	result := make([]gin.H, 0, len(memberWorkload))
	for _, workload := range memberWorkload {
		result = append(result, workload)
	}

	return result
}

// getBugTrend 获取Bug趋势
func (h *ProjectHandler) getBugTrend(projectID uint, days int) []gin.H {
	var bugs []model.Bug
	h.db.Where("project_id = ? AND created_at >= ?", projectID, time.Now().AddDate(0, 0, -days)).
		Order("created_at ASC").
		Find(&bugs)

	bugByDate := make(map[string]int)
	for _, bug := range bugs {
		date := bug.CreatedAt.Format("2006-01-02")
		bugByDate[date]++
	}

	result := make([]gin.H, 0)
	for date, count := range bugByDate {
		result = append(result, gin.H{
			"date":  date,
			"count": count,
		})
	}

	return result
}

// getRequirementTrend 获取需求趋势
func (h *ProjectHandler) getRequirementTrend(projectID uint, days int) []gin.H {
	var requirements []model.Requirement
	h.db.Where("project_id = ? AND created_at >= ?", projectID, time.Now().AddDate(0, 0, -days)).
		Order("created_at ASC").
		Find(&requirements)

	reqByDate := make(map[string]int)
	for _, req := range requirements {
		date := req.CreatedAt.Format("2006-01-02")
		reqByDate[date]++
	}

	result := make([]gin.H, 0)
	for date, count := range reqByDate {
		result = append(result, gin.H{
			"date":  date,
			"count": count,
		})
	}

	return result
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

	// 权限检查：普通用户只能为自己参与的项目添加成员
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限管理该项目成员")
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

	// 权限检查：普通用户只能为自己参与的项目更新成员
	if !utils.CheckProjectAccess(h.db, c, member.ProjectID) {
		utils.Error(c, 403, "没有权限管理该项目成员")
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

	// 先查询成员信息以验证项目
	var member model.ProjectMember
	if err := h.db.Where("project_id = ? AND id = ?", projectID, memberID).First(&member).Error; err != nil {
		utils.Error(c, 404, "项目成员不存在")
		return
	}

	// 权限检查：普通用户只能为自己参与的项目移除成员
	if !utils.CheckProjectAccess(h.db, c, member.ProjectID) {
		utils.Error(c, 403, "没有权限管理该项目成员")
		return
	}

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

// isValidProjectStatus 检查项目状态是否合法（禅道状态值）
func isValidProjectStatus(status string) bool {
	switch status {
	case "wait", "doing", "suspended", "closed", "done":
		return true
	}
	return false
}
