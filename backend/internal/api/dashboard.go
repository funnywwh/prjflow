package api

import (
	"encoding/json"
	"time"

	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetDashboard 获取个人工作台数据
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	uid := userID.(uint)

	// 获取我的任务统计
	taskStats := h.getTaskStats(uid)

	// 获取我的Bug统计
	bugStats := h.getBugStats(uid)

	// 获取我的需求统计
	requirementStats := h.getRequirementStats(uid)

	// 获取我的项目列表
	projects := h.getMyProjects(uid)

	// 获取我的工作报告统计
	reportStats := h.getReportStats(uid)

	// 获取资源分配统计（本周、本月工时）
	resourceStats := h.getResourceStats(uid)

	// 汇总统计
	statistics := gin.H{
		"total_tasks":        taskStats["todo"].(int) + taskStats["in_progress"].(int) + taskStats["done"].(int),
		"total_bugs":         bugStats["active"].(int) + bugStats["resolved"].(int) + bugStats["closed"].(int),
		"total_requirements": requirementStats["in_progress"].(int) + requirementStats["completed"].(int),
		"total_projects":     len(projects),
		"week_hours":         resourceStats["week_hours"],
		"month_hours":        resourceStats["month_hours"],
	}

	utils.Success(c, gin.H{
		"tasks":        taskStats,
		"bugs":         bugStats,
		"requirements": requirementStats,
		"projects":     projects,
		"reports":      reportStats,
		"statistics":   statistics,
	})
}

// getTaskStats 获取任务统计
func (h *DashboardHandler) getTaskStats(userID uint) gin.H {
	var todoCount, inProgressCount, doneCount int64

	h.db.Model(&model.Task{}).
		Where("assignee_id = ? AND status = ?", userID, "wait").
		Count(&todoCount)

	h.db.Model(&model.Task{}).
		Where("assignee_id = ? AND status = ?", userID, "doing").
		Count(&inProgressCount)

	h.db.Model(&model.Task{}).
		Where("assignee_id = ? AND status = ?", userID, "done").
		Count(&doneCount)

	return gin.H{
		"todo":        int(todoCount),
		"in_progress": int(inProgressCount),
		"done":        int(doneCount),
	}
}

// getBugStats 获取Bug统计
func (h *DashboardHandler) getBugStats(userID uint) gin.H {
	var activeCount, resolvedCount, closedCount int64

	// 查询分配给当前用户的Bug（禅道状态：active, resolved, closed）
	// 使用 Model(&model.Bug{}) 确保自动过滤软删除的记录（deleted_at IS NULL）
	h.db.Model(&model.Bug{}).
		Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
		Where("bug_assignees.user_id = ? AND bugs.status = ?", userID, "active").
		Count(&activeCount)

	h.db.Model(&model.Bug{}).
		Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
		Where("bug_assignees.user_id = ? AND bugs.status = ?", userID, "resolved").
		Count(&resolvedCount)

	h.db.Model(&model.Bug{}).
		Joins("JOIN bug_assignees ON bugs.id = bug_assignees.bug_id").
		Where("bug_assignees.user_id = ? AND bugs.status = ?", userID, "closed").
		Count(&closedCount)

	return gin.H{
		"active":   int(activeCount),
		"resolved": int(resolvedCount),
		"closed":   int(closedCount),
	}
}

// getRequirementStats 获取需求统计
func (h *DashboardHandler) getRequirementStats(userID uint) gin.H {
	var inProgressCount, completedCount int64

	h.db.Model(&model.Requirement{}).
		Where("assignee_id = ? AND status = ?", userID, "active").
		Count(&inProgressCount)

	h.db.Model(&model.Requirement{}).
		Where("assignee_id = ? AND status = ?", userID, "completed").
		Count(&completedCount)

	return gin.H{
		"in_progress": int(inProgressCount),
		"completed":   int(completedCount),
	}
}

// getMyProjects 获取我的项目列表
func (h *DashboardHandler) getMyProjects(userID uint) []gin.H {
	var projectMembers []model.ProjectMember
	h.db.Where("user_id = ?", userID).
		Preload("Project").
		Find(&projectMembers)

	projects := make([]gin.H, 0, len(projectMembers))
	for _, pm := range projectMembers {
		projects = append(projects, gin.H{
			"id":   pm.Project.ID,
			"name": pm.Project.Name,
			"code": pm.Project.Code,
			"role": pm.Role,
		})
	}

	return projects
}

// getReportStats 获取工作报告统计
func (h *DashboardHandler) getReportStats(userID uint) gin.H {
	var pendingCount, submittedCount int64

	// 日报统计
	h.db.Model(&model.DailyReport{}).
		Where("user_id = ? AND status = ?", userID, "draft").
		Count(&pendingCount)

	h.db.Model(&model.DailyReport{}).
		Where("user_id = ? AND status = ?", userID, "submitted").
		Count(&submittedCount)

	// 周报统计
	var weeklyPending, weeklySubmitted int64
	h.db.Model(&model.WeeklyReport{}).
		Where("user_id = ? AND status = ?", userID, "draft").
		Count(&weeklyPending)

	h.db.Model(&model.WeeklyReport{}).
		Where("user_id = ? AND status = ?", userID, "submitted").
		Count(&weeklySubmitted)

	// 待审批数量：需要当前用户审批的报告（状态为submitted且当前用户是审批人，且审批状态为pending）
	var pendingApprovalCount int64
	// 日报待审批：状态为submitted，且当前用户在审批人列表中，且当前用户的审批记录状态为pending
	// 使用子查询避免重复计数
	h.db.Raw(`
		SELECT COUNT(DISTINCT daily_reports.id)
		FROM daily_reports
		INNER JOIN daily_report_approvers ON daily_reports.id = daily_report_approvers.daily_report_id
		LEFT JOIN daily_report_approvals ON daily_reports.id = daily_report_approvals.daily_report_id AND daily_report_approvals.approver_id = ?
		WHERE daily_reports.status = ? 
		  AND daily_report_approvers.user_id = ? 
		  AND (daily_report_approvals.status = 'pending' OR daily_report_approvals.id IS NULL)
	`, userID, "submitted", userID).Scan(&pendingApprovalCount)

	// 周报待审批
	var weeklyPendingApproval int64
	// 使用子查询避免重复计数
	h.db.Raw(`
		SELECT COUNT(DISTINCT weekly_reports.id)
		FROM weekly_reports
		INNER JOIN weekly_report_approvers ON weekly_reports.id = weekly_report_approvers.weekly_report_id
		LEFT JOIN weekly_report_approvals ON weekly_reports.id = weekly_report_approvals.weekly_report_id AND weekly_report_approvals.approver_id = ?
		WHERE weekly_reports.status = ? 
		  AND weekly_report_approvers.user_id = ? 
		  AND (weekly_report_approvals.status = 'pending' OR weekly_report_approvals.id IS NULL)
	`, userID, "submitted", userID).Scan(&weeklyPendingApproval)

	return gin.H{
		"pending":          int(pendingCount + weeklyPending),
		"submitted":        int(submittedCount + weeklySubmitted),
		"pending_approval": int(pendingApprovalCount + weeklyPendingApproval),
	}
}

// getResourceStats 获取资源分配统计
func (h *DashboardHandler) getResourceStats(userID uint) gin.H {
	now := time.Now()

	// 本周开始和结束
	weekStart := now
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	// 本月开始和结束
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var weekHours, monthHours float64

	// 查询本周工时
	h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resources.user_id = ? AND resource_allocations.date >= ? AND resource_allocations.date < ?",
			userID, weekStart, weekEnd).
		Select("COALESCE(SUM(resource_allocations.hours), 0)").
		Scan(&weekHours)

	// 查询本月工时
	h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resources.user_id = ? AND resource_allocations.date >= ? AND resource_allocations.date < ?",
			userID, monthStart, monthEnd).
		Select("COALESCE(SUM(resource_allocations.hours), 0)").
		Scan(&monthHours)

	return gin.H{
		"week_hours":  weekHours,
		"month_hours": monthHours,
	}
}

// GetDashboardConfig 获取工作台配置
func (h *DashboardHandler) GetDashboardConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	uid := userID.(uint)
	var dashboard model.UserDashboard
	if err := h.db.Where("user_id = ?", uid).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 返回默认配置
			defaultConfig := gin.H{
				"cards": []gin.H{
					{"key": "tasks", "visible": true, "order": 1},
					{"key": "bugs", "visible": true, "order": 2},
					{"key": "requirements", "visible": true, "order": 3},
					{"key": "projects", "visible": true, "order": 4},
					{"key": "resources", "visible": true, "order": 5},
					{"key": "reports", "visible": true, "order": 6},
				},
				"tabs": []gin.H{
					{"key": "projects", "visible": true, "order": 1},
					{"key": "tasks", "visible": true, "order": 2},
					{"key": "bugs", "visible": true, "order": 3},
					{"key": "resources", "visible": true, "order": 4},
					{"key": "reports", "visible": true, "order": 5},
				},
			}
			utils.Success(c, defaultConfig)
			return
		}
		utils.Error(c, utils.CodeError, "查询配置失败")
		return
	}

	// 解析JSON配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(dashboard.Config), &config); err != nil {
		utils.Error(c, utils.CodeError, "解析配置失败")
		return
	}

	utils.Success(c, config)
}

// SaveDashboardConfig 保存工作台配置
func (h *DashboardHandler) SaveDashboardConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	uid := userID.(uint)

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 将配置转换为JSON字符串
	configJSON, err := json.Marshal(req)
	if err != nil {
		utils.Error(c, 400, "配置格式错误")
		return
	}

	// 保存或更新配置
	var dashboard model.UserDashboard
	if err := h.db.Where("user_id = ?", uid).First(&dashboard).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新配置
			dashboard = model.UserDashboard{
				UserID: uid,
				Config: string(configJSON),
			}
			if err := h.db.Create(&dashboard).Error; err != nil {
				utils.Error(c, utils.CodeError, "保存配置失败")
				return
			}
		} else {
			utils.Error(c, utils.CodeError, "查询配置失败")
			return
		}
	} else {
		// 更新现有配置
		dashboard.Config = string(configJSON)
		if err := h.db.Save(&dashboard).Error; err != nil {
			utils.Error(c, utils.CodeError, "保存配置失败")
			return
		}
	}

	utils.Success(c, gin.H{"message": "配置已保存"})
}
