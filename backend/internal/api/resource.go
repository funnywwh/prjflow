package api

import (
	"fmt"
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

	// 权限过滤：普通用户只能看到自己参与的项目相关的资源
	if !utils.IsAdmin(c) {
		userID := utils.GetUserID(c)
		if userID == 0 {
			query = query.Where("1 = 0")
		} else {
			// 获取用户参与的项目ID列表
			projectIDs := utils.GetUserProjectIDs(h.db, userID)
			if len(projectIDs) > 0 {
				query = query.Where("project_id IN ?", projectIDs)
			} else {
				query = query.Where("1 = 0")
			}
		}
	}

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
	countQuery := h.db.Model(&model.Resource{})
	// 权限过滤：普通用户只能看到自己参与的项目相关的资源
	if !utils.IsAdmin(c) {
		userID := utils.GetUserID(c)
		if userID == 0 {
			countQuery = countQuery.Where("1 = 0")
		} else {
			projectIDs := utils.GetUserProjectIDs(h.db, userID)
			if len(projectIDs) > 0 {
				countQuery = countQuery.Where("project_id IN ?", projectIDs)
			} else {
				countQuery = countQuery.Where("1 = 0")
			}
		}
	}

	// 用户筛选
	if userID := c.Query("user_id"); userID != "" {
		countQuery = countQuery.Where("user_id = ?", userID)
	}

	// 项目筛选
	if projectID := c.Query("project_id"); projectID != "" {
		countQuery = countQuery.Where("project_id = ?", projectID)
	}

	// 角色筛选
	if role := c.Query("role"); role != "" {
		countQuery = countQuery.Where("role = ?", role)
	}

	countQuery.Count(&total)

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

	// 权限检查：普通用户只能查看自己参与的项目相关的资源
	if !utils.IsAdmin(c) {
		if !utils.CheckProjectAccess(h.db, c, resource.ProjectID) {
			utils.Error(c, 403, "没有权限访问该资源")
			return
		}
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

// GetResourceStatistics 获取资源统计（从ResourceAllocation表统计实际工时）
func (h *ResourceHandler) GetResourceStatistics(c *gin.Context) {
	// 从ResourceAllocation表统计，这样可以准确反映实际工时情况

	// 权限过滤：普通用户只能统计自己参与的项目的数据
	var allowedProjectIDs []uint
	if !utils.IsAdmin(c) {
		userID := utils.GetUserID(c)
		if userID == 0 {
			utils.Error(c, 401, "未登录")
			return
		}
		allowedProjectIDs = utils.GetUserProjectIDs(h.db, userID)
		if len(allowedProjectIDs) == 0 {
			// 用户没有参与任何项目，返回空统计
			utils.Success(c, gin.H{
				"total_hours":   0.0,
				"project_stats": []gin.H{},
				"user_stats":     []gin.H{},
			})
			return
		}
	}

	// 用户筛选
	userID := c.Query("user_id")
	
	// 项目筛选
	projectID := c.Query("project_id")
	
	// 如果普通用户指定了项目ID，需要验证是否有权限
	if !utils.IsAdmin(c) && projectID != "" {
		var pid uint
		if _, err := fmt.Sscanf(projectID, "%d", &pid); err == nil {
			hasAccess := false
			for _, id := range allowedProjectIDs {
				if id == pid {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				utils.Error(c, 403, "没有权限访问该项目")
				return
			}
		}
	}

	// 统计任务工时（从ResourceAllocation表）
	taskAllocationQuery := h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resource_allocations.task_id IS NOT NULL")
	// 权限过滤：普通用户只能统计自己参与的项目
	if !utils.IsAdmin(c) && len(allowedProjectIDs) > 0 {
		taskAllocationQuery = taskAllocationQuery.Where("resource_allocations.project_id IN ?", allowedProjectIDs)
	}
	if userID != "" {
		taskAllocationQuery = taskAllocationQuery.Where("resources.user_id = ?", userID)
	}
	if projectID != "" {
		taskAllocationQuery = taskAllocationQuery.Where("resource_allocations.project_id = ?", projectID)
	}

	var taskHours float64
	taskAllocationQuery.Select("COALESCE(SUM(resource_allocations.hours), 0)").Scan(&taskHours)

	// 统计Bug工时（从ResourceAllocation表）
	bugAllocationQuery := h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resource_allocations.bug_id IS NOT NULL")
	// 权限过滤：普通用户只能统计自己参与的项目
	if !utils.IsAdmin(c) && len(allowedProjectIDs) > 0 {
		bugAllocationQuery = bugAllocationQuery.Where("resource_allocations.project_id IN ?", allowedProjectIDs)
	}
	if userID != "" {
		bugAllocationQuery = bugAllocationQuery.Where("resources.user_id = ?", userID)
	}
	if projectID != "" {
		bugAllocationQuery = bugAllocationQuery.Where("resource_allocations.project_id = ?", projectID)
	}

	var bugHours float64
	bugAllocationQuery.Select("COALESCE(SUM(resource_allocations.hours), 0)").Scan(&bugHours)

	// 统计需求工时（从ResourceAllocation表）
	requirementAllocationQuery := h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resource_allocations.requirement_id IS NOT NULL")
	// 权限过滤：普通用户只能统计自己参与的项目
	if !utils.IsAdmin(c) && len(allowedProjectIDs) > 0 {
		requirementAllocationQuery = requirementAllocationQuery.Where("resource_allocations.project_id IN ?", allowedProjectIDs)
	}
	if userID != "" {
		requirementAllocationQuery = requirementAllocationQuery.Where("resources.user_id = ?", userID)
	}
	if projectID != "" {
		requirementAllocationQuery = requirementAllocationQuery.Where("resource_allocations.project_id = ?", projectID)
	}

	var requirementHours float64
	requirementAllocationQuery.Select("COALESCE(SUM(resource_allocations.hours), 0)").Scan(&requirementHours)

	// 总工时
	totalHours := taskHours + bugHours + requirementHours

	// 按项目统计（从ResourceAllocation表）
	var projectStats []struct {
		ProjectID   uint    `json:"project_id"`
		ProjectName string  `json:"project_name"`
		TotalHours  float64 `json:"total_hours"`
	}
	
	projectStatsQuery := h.db.Model(&model.ResourceAllocation{}).
		Select("resource_allocations.project_id, projects.name as project_name, COALESCE(SUM(resource_allocations.hours), 0) as total_hours").
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Joins("LEFT JOIN projects ON resource_allocations.project_id = projects.id").
		Where("resource_allocations.project_id IS NOT NULL")
	// 权限过滤：普通用户只能统计自己参与的项目
	if !utils.IsAdmin(c) && len(allowedProjectIDs) > 0 {
		projectStatsQuery = projectStatsQuery.Where("resource_allocations.project_id IN ?", allowedProjectIDs)
	}
	if userID != "" {
		projectStatsQuery = projectStatsQuery.Where("resources.user_id = ?", userID)
	}
	if projectID != "" {
		projectStatsQuery = projectStatsQuery.Where("resource_allocations.project_id = ?", projectID)
	}
	projectStatsQuery.Group("resource_allocations.project_id, projects.name").Scan(&projectStats)
	

	// 按人员统计（从ResourceAllocation表）
	var userStats []struct {
		UserID     uint    `json:"user_id"`
		Username   string  `json:"username"`
		Nickname   string  `json:"nickname"`
		TotalHours float64 `json:"total_hours"`
	}
	
	userStatsQuery := h.db.Model(&model.ResourceAllocation{}).
		Select("resources.user_id, users.username, users.nickname, COALESCE(SUM(resource_allocations.hours), 0) as total_hours").
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Joins("LEFT JOIN users ON resources.user_id = users.id")
	// 权限过滤：普通用户只能统计自己参与的项目
	if !utils.IsAdmin(c) && len(allowedProjectIDs) > 0 {
		userStatsQuery = userStatsQuery.Where("resource_allocations.project_id IN ?", allowedProjectIDs)
	}
	if userID != "" {
		userStatsQuery = userStatsQuery.Where("resources.user_id = ?", userID)
	}
	if projectID != "" {
		userStatsQuery = userStatsQuery.Where("resource_allocations.project_id = ?", projectID)
	}
	userStatsQuery.Group("resources.user_id, users.username, users.nickname").Scan(&userStats)

	utils.Success(c, gin.H{
		"total_hours":   totalHours,
		"project_stats": projectStats,
		"user_stats":    userStats,
	})
}

// CheckResourceConflict 检查资源冲突（检查同一人员在同一天的工时是否超过限制）
func (h *ResourceHandler) CheckResourceConflict(c *gin.Context) {
	userID := c.Query("user_id")
	dateStr := c.Query("date")
	
	if userID == "" || dateStr == "" {
		utils.Error(c, 400, "需要提供user_id和date参数")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		utils.Error(c, 400, "日期格式错误，应为 YYYY-MM-DD")
		return
	}

	// 查询该用户在该日期的所有资源分配（从ResourceAllocation表）
	var totalHours float64
	h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resources.user_id = ? AND resource_allocations.date = ?", userID, date).
		Select("COALESCE(SUM(resource_allocations.hours), 0)").
		Scan(&totalHours)

	conflicts := []string{}
	if totalHours > 24 {
		conflicts = append(conflicts, "总工时超过24小时")
	}
	if totalHours > 12 {
		conflicts = append(conflicts, "总工时超过12小时（建议检查）")
	}

	// 获取该用户在该日期的详细分配情况
	var allocations []struct {
		ProjectName string  `json:"project_name"`
		TaskTitle    *string `json:"task_title"`
		BugTitle     *string `json:"bug_title"`
		Hours        float64 `json:"hours"`
		Description  string  `json:"description"`
	}
	h.db.Model(&model.ResourceAllocation{}).
		Select(`
			projects.name as project_name,
			tasks.title as task_title,
			bugs.title as bug_title,
			resource_allocations.hours,
			resource_allocations.description
		`).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Joins("LEFT JOIN projects ON resource_allocations.project_id = projects.id").
		Joins("LEFT JOIN tasks ON resource_allocations.task_id = tasks.id").
		Joins("LEFT JOIN bugs ON resource_allocations.bug_id = bugs.id").
		Where("resources.user_id = ? AND resource_allocations.date = ?", userID, date).
		Scan(&allocations)

	utils.Success(c, gin.H{
		"user_id":     userID,
		"date":        dateStr,
		"total_hours": totalHours,
		"conflicts":   conflicts,
		"has_conflict": totalHours > 24,
		"has_warning": totalHours > 12,
		"allocations": allocations,
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

