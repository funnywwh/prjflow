package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type ResourceHandler struct {
	db *gorm.DB
}

func NewResourceHandler(db *gorm.DB) *ResourceHandler {
	return &ResourceHandler{db: db}
}

// GetResources 获取人员资源列表
func (h *ResourceHandler) GetResources(c *gin.Context) {
	var resources []model.Resource
	query := h.db.Preload("User").Preload("Project")

	// 用户筛选
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 项目筛选
	if projectID := c.Query("project_id"); projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	// 角色筛选
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&model.Resource{}).Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&resources).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      resources,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetResource 获取人员资源详情
func (h *ResourceHandler) GetResource(c *gin.Context) {
	id := c.Param("id")
	var resource model.Resource
	if err := h.db.Preload("User").Preload("Project").Preload("Allocations").First(&resource, id).Error; err != nil {
		utils.Error(c, 404, "资源不存在")
		return
	}

	utils.Success(c, resource)
}

// CreateResource 创建人员资源
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	var req struct {
		UserID    uint   `json:"user_id" binding:"required"`
		ProjectID uint   `json:"project_id" binding:"required"`
		Role      string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证用户是否存在
	var user model.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, req.ProjectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	// 检查是否已存在相同的资源分配
	var existingResource model.Resource
	if err := h.db.Where("user_id = ? AND project_id = ?", req.UserID, req.ProjectID).First(&existingResource).Error; err == nil {
		utils.Error(c, 400, "该用户已在此项目中分配资源")
		return
	}

	resource := model.Resource{
		UserID:    req.UserID,
		ProjectID: req.ProjectID,
		Role:      req.Role,
	}

	if err := h.db.Create(&resource).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("User").Preload("Project").Preload("Allocations").First(&resource, resource.ID)

	utils.Success(c, resource)
}

// UpdateResource 更新人员资源
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	id := c.Param("id")
	var resource model.Resource
	if err := h.db.First(&resource, id).Error; err != nil {
		utils.Error(c, 404, "资源不存在")
		return
	}

	var req struct {
		Role *string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if req.Role != nil {
		resource.Role = *req.Role
	}

	if err := h.db.Save(&resource).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("User").Preload("Project").Preload("Allocations").First(&resource, resource.ID)

	utils.Success(c, resource)
}

// DeleteResource 删除人员资源
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&model.Resource{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetResourceStatistics 获取资源统计（从任务和Bug的工时统计）
func (h *ResourceHandler) GetResourceStatistics(c *gin.Context) {
	// 从任务和Bug的actual_hours统计，而不是从ResourceAllocation表
	// 这样可以反映实际的工作情况

	// 用户筛选
	userID := c.Query("user_id")
	
	// 项目筛选
	projectID := c.Query("project_id")

	// 日期范围筛选（这里暂时不使用，因为任务和Bug没有日期字段用于筛选工时）
	// 如果需要按日期筛选，需要从ResourceAllocation表查询

	// 统计任务工时
	taskQuery := h.db.Model(&model.Task{}).Where("actual_hours > 0")
	if userID != "" {
		taskQuery = taskQuery.Where("assignee_id = ?", userID)
	}
	if projectID != "" {
		taskQuery = taskQuery.Where("project_id = ?", projectID)
	}

	var taskHours float64
	taskQuery.Select("COALESCE(SUM(actual_hours), 0)").Scan(&taskHours)

	// 统计Bug工时
	bugQuery := h.db.Model(&model.Bug{}).Where("actual_hours > 0")
	if projectID != "" {
		bugQuery = bugQuery.Where("project_id = ?", projectID)
	}
	// Bug的分配人是多对多关系，需要特殊处理
	if userID != "" {
		bugQuery = bugQuery.Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
			Where("bug_assignees.user_id = ?", userID)
	}

	var bugHours float64
	bugQuery.Select("COALESCE(SUM(actual_hours), 0)").Scan(&bugHours)

	// 总工时
	totalHours := taskHours + bugHours

	// 按项目统计（从任务和Bug）
	var projectStats []struct {
		ProjectID   uint    `json:"project_id"`
		ProjectName string  `json:"project_name"`
		TotalHours  float64 `json:"total_hours"`
	}
	
	// 任务按项目统计
	var taskProjectStats []struct {
		ProjectID   uint    `json:"project_id"`
		ProjectName string  `json:"project_name"`
		TotalHours  float64 `json:"total_hours"`
	}
	taskProjectQuery := h.db.Model(&model.Task{}).
		Select("tasks.project_id, projects.name as project_name, COALESCE(SUM(tasks.actual_hours), 0) as total_hours").
		Joins("LEFT JOIN projects ON tasks.project_id = projects.id").
		Where("tasks.actual_hours > 0")
	if userID != "" {
		taskProjectQuery = taskProjectQuery.Where("tasks.assignee_id = ?", userID)
	}
	if projectID != "" {
		taskProjectQuery = taskProjectQuery.Where("tasks.project_id = ?", projectID)
	}
	taskProjectQuery.Group("tasks.project_id, projects.name").Scan(&taskProjectStats)
	
	// Bug按项目统计
	var bugProjectStats []struct {
		ProjectID   uint    `json:"project_id"`
		ProjectName string  `json:"project_name"`
		TotalHours  float64 `json:"total_hours"`
	}
	bugProjectQuery := h.db.Model(&model.Bug{}).
		Select("bugs.project_id, projects.name as project_name, COALESCE(SUM(bugs.actual_hours), 0) as total_hours").
		Joins("LEFT JOIN projects ON bugs.project_id = projects.id").
		Where("bugs.actual_hours > 0")
	if userID != "" {
		bugProjectQuery = bugProjectQuery.Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
			Where("bug_assignees.user_id = ?", userID)
	}
	if projectID != "" {
		bugProjectQuery = bugProjectQuery.Where("bugs.project_id = ?", projectID)
	}
	bugProjectQuery.Group("bugs.project_id, projects.name").Scan(&bugProjectStats)
	
	// 合并任务和Bug的项目统计
	projectMap := make(map[uint]struct {
		ProjectID   uint
		ProjectName string
		TotalHours  float64
	})
	for _, stat := range taskProjectStats {
		projectMap[stat.ProjectID] = struct {
			ProjectID   uint
			ProjectName string
			TotalHours  float64
		}{stat.ProjectID, stat.ProjectName, stat.TotalHours}
	}
	for _, stat := range bugProjectStats {
		if existing, ok := projectMap[stat.ProjectID]; ok {
			existing.TotalHours += stat.TotalHours
			projectMap[stat.ProjectID] = existing
		} else {
			projectMap[stat.ProjectID] = struct {
				ProjectID   uint
				ProjectName string
				TotalHours  float64
			}{stat.ProjectID, stat.ProjectName, stat.TotalHours}
		}
	}
	for _, stat := range projectMap {
		projectStats = append(projectStats, struct {
			ProjectID   uint    `json:"project_id"`
			ProjectName string  `json:"project_name"`
			TotalHours  float64 `json:"total_hours"`
		}{stat.ProjectID, stat.ProjectName, stat.TotalHours})
	}

	// 按人员统计（从任务和Bug）
	var userStats []struct {
		UserID     uint    `json:"user_id"`
		Username   string  `json:"username"`
		Nickname   string  `json:"nickname"`
		TotalHours float64 `json:"total_hours"`
	}
	
	// 任务按人员统计
	var taskUserStats []struct {
		UserID     uint    `json:"user_id"`
		Username   string  `json:"username"`
		Nickname   string  `json:"nickname"`
		TotalHours float64 `json:"total_hours"`
	}
	taskUserQuery := h.db.Model(&model.Task{}).
		Select("tasks.assignee_id as user_id, users.username, users.nickname, COALESCE(SUM(tasks.actual_hours), 0) as total_hours").
		Joins("LEFT JOIN users ON tasks.assignee_id = users.id").
		Where("tasks.actual_hours > 0 AND tasks.assignee_id IS NOT NULL")
	if userID != "" {
		taskUserQuery = taskUserQuery.Where("tasks.assignee_id = ?", userID)
	}
	if projectID != "" {
		taskUserQuery = taskUserQuery.Where("tasks.project_id = ?", projectID)
	}
	taskUserQuery.Group("tasks.assignee_id, users.username, users.nickname").Scan(&taskUserStats)
	
	// Bug按人员统计
	var bugUserStats []struct {
		UserID     uint    `json:"user_id"`
		Username   string  `json:"username"`
		Nickname   string  `json:"nickname"`
		TotalHours float64 `json:"total_hours"`
	}
	bugUserQuery := h.db.Model(&model.Bug{}).
		Select("bug_assignees.user_id, users.username, users.nickname, COALESCE(SUM(bugs.actual_hours), 0) as total_hours").
		Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
		Joins("LEFT JOIN users ON bug_assignees.user_id = users.id").
		Where("bugs.actual_hours > 0")
	if userID != "" {
		bugUserQuery = bugUserQuery.Where("bug_assignees.user_id = ?", userID)
	}
	if projectID != "" {
		bugUserQuery = bugUserQuery.Where("bugs.project_id = ?", projectID)
	}
	bugUserQuery.Group("bug_assignees.user_id, users.username, users.nickname").Scan(&bugUserStats)
	
	// 合并任务和Bug的人员统计
	userMap := make(map[uint]struct {
		UserID     uint
		Username   string
		Nickname   string
		TotalHours float64
	})
	for _, stat := range taskUserStats {
		userMap[stat.UserID] = struct {
			UserID     uint
			Username   string
			Nickname   string
			TotalHours float64
		}{stat.UserID, stat.Username, stat.Nickname, stat.TotalHours}
	}
	for _, stat := range bugUserStats {
		if existing, ok := userMap[stat.UserID]; ok {
			existing.TotalHours += stat.TotalHours
			userMap[stat.UserID] = existing
		} else {
			userMap[stat.UserID] = struct {
				UserID     uint
				Username   string
				Nickname   string
				TotalHours float64
			}{stat.UserID, stat.Username, stat.Nickname, stat.TotalHours}
		}
	}
	for _, stat := range userMap {
		userStats = append(userStats, struct {
			UserID     uint    `json:"user_id"`
			Username   string  `json:"username"`
			Nickname   string  `json:"nickname"`
			TotalHours float64 `json:"total_hours"`
		}{stat.UserID, stat.Username, stat.Nickname, stat.TotalHours})
	}

	utils.Success(c, gin.H{
		"total_hours":   totalHours,
		"project_stats": projectStats,
		"user_stats":    userStats,
	})
}

// GetResourceUtilization 获取资源利用率分析
func (h *ResourceHandler) GetResourceUtilization(c *gin.Context) {
	// 获取查询参数
	userID := c.Query("user_id")
	projectID := c.Query("project_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 默认查询当前月份
	var startDate, endDate time.Time
	now := time.Now()
	if startDateStr == "" {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		} else {
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
	}
	if endDateStr == "" {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0).AddDate(0, 0, -1)
	} else {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t
		} else {
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, 1, 0).AddDate(0, 0, -1)
		}
	}

	// 计算日期范围内的天数
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	baseQuery := h.db.Model(&model.ResourceAllocation{}).
		Where("date >= ? AND date <= ?", startDate, endDate)

	// 用户筛选
	if userID != "" {
		baseQuery = baseQuery.Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
			Where("resources.user_id = ?", userID)
	}

	// 项目筛选
	if projectID != "" {
		baseQuery = baseQuery.Where("resource_allocations.project_id = ?", projectID)
	}

	// 按资源统计利用率
	var utilizationStats []struct {
		ResourceID  uint    `json:"resource_id"`
		UserID      uint    `json:"user_id"`
		Username    string  `json:"username"`
		Nickname    string  `json:"nickname"`
		ProjectID   uint    `json:"project_id"`
		ProjectName string  `json:"project_name"`
		TotalHours  float64 `json:"total_hours"`
		MaxHours    float64 `json:"max_hours"`    // 最大可能工时（天数 * 24）
		Utilization float64 `json:"utilization"`  // 利用率（总工时 / 最大工时 * 100）
	}

	utilizationQuery := baseQuery.Session(&gorm.Session{}).
		Select(`
			resources.id as resource_id,
			resources.user_id,
			users.username,
			users.nickname,
			resources.project_id,
			projects.name as project_name,
			COALESCE(SUM(resource_allocations.hours), 0) as total_hours,
			? * 24 as max_hours,
			COALESCE(SUM(resource_allocations.hours), 0) / (? * 24) * 100 as utilization
		`, days, days).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Joins("JOIN users ON resources.user_id = users.id").
		Joins("JOIN projects ON resources.project_id = projects.id").
		Group("resources.id, resources.user_id, users.username, users.nickname, resources.project_id, projects.name").
		Order("utilization DESC")

	utilizationQuery.Scan(&utilizationStats)

	// 计算平均利用率
	var avgUtilization float64
	if len(utilizationStats) > 0 {
		totalUtilization := 0.0
		for _, stat := range utilizationStats {
			totalUtilization += stat.Utilization
		}
		avgUtilization = totalUtilization / float64(len(utilizationStats))
	}

	utils.Success(c, gin.H{
		"start_date":      startDate.Format("2006-01-02"),
		"end_date":        endDate.Format("2006-01-02"),
		"days":            days,
		"utilization_stats": utilizationStats,
		"avg_utilization": avgUtilization,
	})
}

