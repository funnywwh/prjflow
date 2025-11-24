package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
)

type BugHandler struct {
	db *gorm.DB
}

func NewBugHandler(db *gorm.DB) *BugHandler {
	return &BugHandler{db: db}
}

// GetBugs 获取Bug列表
func (h *BugHandler) GetBugs(c *gin.Context) {
	var bugs []model.Bug
	query := h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module")

	// 权限过滤：普通用户只能看到自己创建或参与的Bug
	query = utils.FilterBugsByUser(h.db, c, query)

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

	// 严重程度筛选
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// 需求筛选
	if requirementID := c.Query("requirement_id"); requirementID != "" {
		query = query.Where("requirement_id = ?", requirementID)
	}

	// 功能模块筛选
	if moduleID := c.Query("module_id"); moduleID != "" {
		query = query.Where("module_id = ?", moduleID)
	}

	// 创建人筛选
	if creatorID := c.Query("creator_id"); creatorID != "" {
		query = query.Where("creator_id = ?", creatorID)
	}

	// 分配人筛选（通过关联表）
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		query = query.Joins("JOIN bug_assignees ON bug_assignees.bug_id = bugs.id").
			Where("bug_assignees.user_id = ?", assigneeID).
			Group("bugs.id")
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	// 权限过滤：普通用户只能看到自己创建或参与的Bug
	countQuery := utils.FilterBugsByUser(h.db, c, h.db.Model(&model.Bug{}))
	countQuery.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&bugs).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      bugs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetBug 获取Bug详情
func (h *BugHandler) GetBug(c *gin.Context) {
	id := c.Param("id")
	var bug model.Bug
	if err := h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module").First(&bug, id).Error; err != nil {
		utils.Error(c, 404, "Bug不存在")
		return
	}

	// 权限检查：普通用户只能查看自己创建或参与的Bug
	if !utils.CheckBugAccess(h.db, c, bug.ID) {
		utils.Error(c, 403, "没有权限访问该Bug")
		return
	}

	utils.Success(c, bug)
}

// CreateBug 创建Bug
func (h *BugHandler) CreateBug(c *gin.Context) {
	var req struct {
		Title         string   `json:"title" binding:"required"`
		Description   string   `json:"description"`
		Status        string   `json:"status"`
		Priority      string   `json:"priority"`
		Severity      string   `json:"severity"`
		ProjectID     uint     `json:"project_id" binding:"required"`
		RequirementID *uint    `json:"requirement_id"`
		ModuleID      *uint    `json:"module_id"`
		AssigneeIDs   []uint   `json:"assignee_ids"`
		EstimatedHours *float64 `json:"estimated_hours"`
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
		req.Status = "open"
	}
	validStatuses := map[string]bool{
		"open":      true,
		"assigned":  true,
		"in_progress": true,
		"resolved":  true,
		"closed":    true,
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

	// 验证严重程度
	if req.Severity == "" {
		req.Severity = "medium"
	}
	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}
	if !validSeverities[req.Severity] {
		utils.Error(c, 400, "严重程度值无效")
		return
	}

	// 验证项目是否存在
	var project model.Project
	if err := h.db.First(&project, req.ProjectID).Error; err != nil {
		utils.Error(c, 400, "项目不存在")
		return
	}

	// 权限检查：普通用户只能在自己参与的项目中创建Bug
	if !utils.CheckProjectAccess(h.db, c, project.ID) {
		utils.Error(c, 403, "没有权限在该项目中创建Bug")
		return
	}

	// 如果指定了需求，验证需求是否存在
	if req.RequirementID != nil {
		var requirement model.Requirement
		if err := h.db.First(&requirement, *req.RequirementID).Error; err != nil {
			utils.Error(c, 400, "需求不存在")
			return
		}
	}

	// 如果指定了功能模块，验证模块是否存在
	if req.ModuleID != nil {
		var module model.Module
		if err := h.db.First(&module, *req.ModuleID).Error; err != nil {
			utils.Error(c, 400, "功能模块不存在")
			return
		}
	}

	// 验证分配人是否存在
	if len(req.AssigneeIDs) > 0 {
		var users []model.User
		if err := h.db.Where("id IN ?", req.AssigneeIDs).Find(&users).Error; err != nil || len(users) != len(req.AssigneeIDs) {
			utils.Error(c, 400, "分配人不存在")
			return
		}
	}

	bug := model.Bug{
		Title:          req.Title,
		Description:    req.Description,
		Status:         req.Status,
		Priority:       req.Priority,
		Severity:       req.Severity,
		ProjectID:      req.ProjectID,
		RequirementID:  req.RequirementID,
		ModuleID:       req.ModuleID,
		CreatorID:      userID.(uint),
		EstimatedHours: req.EstimatedHours,
	}

	if err := h.db.Create(&bug).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 分配Bug给用户
	if len(req.AssigneeIDs) > 0 {
		var assignees []model.User
		h.db.Where("id IN ?", req.AssigneeIDs).Find(&assignees)
		if err := h.db.Model(&bug).Association("Assignees").Replace(assignees); err != nil {
			utils.Error(c, utils.CodeError, "分配失败")
			return
		}
		// 如果有分配人，状态自动变为assigned
		if bug.Status == "open" {
			bug.Status = "assigned"
			h.db.Save(&bug)
		}
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module").Preload("ResolvedVersion").First(&bug, bug.ID)

	utils.Success(c, bug)
}

// UpdateBug 更新Bug
func (h *BugHandler) UpdateBug(c *gin.Context) {
	id := c.Param("id")
	var bug model.Bug
	if err := h.db.First(&bug, id).Error; err != nil {
		utils.Error(c, 404, "Bug不存在")
		return
	}

	// 权限检查：普通用户只能更新自己创建或参与的Bug
	if !utils.CheckBugAccess(h.db, c, bug.ID) {
		utils.Error(c, 403, "没有权限更新该Bug")
		return
	}

	var req struct {
		Title          *string  `json:"title"`
		Description    *string  `json:"description"`
		Status         *string  `json:"status"`
		Priority       *string  `json:"priority"`
		Severity       *string  `json:"severity"`
		ProjectID      *uint    `json:"project_id"`
		RequirementID  *uint    `json:"requirement_id"`
		ModuleID       *uint    `json:"module_id"`
		AssigneeIDs    *[]uint  `json:"assignee_ids"`
		EstimatedHours *float64 `json:"estimated_hours"`
		ActualHours    *float64 `json:"actual_hours"` // 实际工时，会自动创建资源分配
		WorkDate       *string  `json:"work_date"`     // 工作日期（YYYY-MM-DD），用于资源分配
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 调试：打印接收到的请求数据
	fmt.Printf("UpdateBug: 接收到请求 - Title: %v, Description: %v\n", req.Title, req.Description)
	if req.Description != nil {
		fmt.Printf("UpdateBug: Description 值 = %q\n", *req.Description)
	}

	// 更新字段
	if req.Title != nil {
		bug.Title = *req.Title
	}
	// 注意：即使 description 是空字符串，也要更新（使用指针判断，空字符串指针不为 nil）
	if req.Description != nil {
		fmt.Printf("UpdateBug: 更新前 bug.Description = %q\n", bug.Description)
		bug.Description = *req.Description
		fmt.Printf("UpdateBug: 更新后 bug.Description = %q\n", bug.Description)
	} else {
		fmt.Printf("UpdateBug: req.Description 为 nil，不更新\n")
	}
	if req.Status != nil {
		// 验证状态
		validStatuses := map[string]bool{
			"open":      true,
			"assigned":  true,
			"in_progress": true,
			"resolved":  true,
			"closed":    true,
		}
		if !validStatuses[*req.Status] {
			utils.Error(c, 400, "状态值无效")
			return
		}
		bug.Status = *req.Status
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
		bug.Priority = *req.Priority
	}
	if req.Severity != nil {
		// 验证严重程度
		validSeverities := map[string]bool{
			"low":      true,
			"medium":   true,
			"high":     true,
			"critical": true,
		}
		if !validSeverities[*req.Severity] {
			utils.Error(c, 400, "严重程度值无效")
			return
		}
		bug.Severity = *req.Severity
	}
	if req.ProjectID != nil {
		// 验证项目是否存在
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 400, "项目不存在")
			return
		}
		bug.ProjectID = *req.ProjectID
	}
	if req.RequirementID != nil {
		// 验证需求是否存在
		if *req.RequirementID != 0 {
			var requirement model.Requirement
			if err := h.db.First(&requirement, *req.RequirementID).Error; err != nil {
				utils.Error(c, 400, "需求不存在")
				return
			}
			bug.RequirementID = req.RequirementID
		} else {
			bug.RequirementID = nil
		}
	}
	if req.ModuleID != nil {
		// 验证功能模块是否存在
		if *req.ModuleID != 0 {
			var module model.Module
			if err := h.db.First(&module, *req.ModuleID).Error; err != nil {
				utils.Error(c, 400, "功能模块不存在")
				return
			}
			bug.ModuleID = req.ModuleID
		} else {
			bug.ModuleID = nil
		}
	}
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			utils.Error(c, 400, "预估工时不能为负数")
			return
		}
		bug.EstimatedHours = req.EstimatedHours
	}
	
	// 如果更新了实际工时，自动创建或更新资源分配
	if req.ActualHours != nil {
		if *req.ActualHours < 0 {
			utils.Error(c, 400, "实际工时不能为负数")
			return
		}
		// 先加载分配人信息
		h.db.Preload("Assignees").First(&bug, bug.ID)
		
		// 注意：重新查询后，需要恢复已更新的字段（如 description）
		if req.Description != nil {
			bug.Description = *req.Description
		}
		if req.Title != nil {
			bug.Title = *req.Title
		}
		// 恢复其他可能已更新的字段
		if req.Status != nil {
			bug.Status = *req.Status
		}
		if req.Priority != nil {
			bug.Priority = *req.Priority
		}
		if req.Severity != nil {
			bug.Severity = *req.Severity
		}
		if req.RequirementID != nil {
			bug.RequirementID = req.RequirementID
		}
		if req.ModuleID != nil {
			bug.ModuleID = req.ModuleID
		}
		if req.EstimatedHours != nil {
			bug.EstimatedHours = req.EstimatedHours
		}
		
		// Bug可能有多个分配人，需要为每个分配人创建资源分配
		// 这里先处理第一个分配人，或者需要前端指定分配人
		// 暂时使用第一个分配人，如果没有分配人则直接设置actual_hours
		if len(bug.Assignees) > 0 {
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
				workDate = time.Now()
			}
			workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())
			
			// 为第一个分配人同步到资源分配
			if err := h.syncBugActualHours(&bug, *req.ActualHours, workDate, bug.Assignees[0].ID); err != nil {
				utils.Error(c, utils.CodeError, "同步资源分配失败: "+err.Error())
				return
			}
			// 从资源分配中汇总实际工时（确保actual_hours正确）
			h.calculateAndUpdateActualHours(&bug)
		} else {
			// 如果没有分配人，直接设置actual_hours，但不创建资源分配
			bug.ActualHours = req.ActualHours
		}
	}
	
	fmt.Printf("UpdateBug: 保存前 bug.Description = %q\n", bug.Description)
	if err := h.db.Save(&bug).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}
	fmt.Printf("UpdateBug: 保存后，重新查询前\n")

	// 更新分配人
	if req.AssigneeIDs != nil {
		var assignees []model.User
		if len(*req.AssigneeIDs) > 0 {
			if err := h.db.Where("id IN ?", *req.AssigneeIDs).Find(&assignees).Error; err != nil || len(assignees) != len(*req.AssigneeIDs) {
				utils.Error(c, 400, "分配人不存在")
				return
			}
		}
		if err := h.db.Model(&bug).Association("Assignees").Replace(assignees); err != nil {
			utils.Error(c, utils.CodeError, "更新分配失败")
			return
		}
		// 如果有分配人且状态为open，自动变为assigned
		// 注意：只更新状态字段，不要覆盖其他已更新的字段（如 description）
		if len(assignees) > 0 && bug.Status == "open" {
			if err := h.db.Model(&bug).Update("status", "assigned").Error; err != nil {
				utils.Error(c, utils.CodeError, "更新状态失败")
				return
			}
			bug.Status = "assigned"
		}
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module").Preload("ResolvedVersion").First(&bug, bug.ID)
	fmt.Printf("UpdateBug: 重新加载后 bug.Description = %q\n", bug.Description)

	utils.Success(c, bug)
}

// DeleteBug 删除Bug
func (h *BugHandler) DeleteBug(c *gin.Context) {
	id := c.Param("id")

	// 验证Bug是否存在
	var bug model.Bug
	if err := h.db.First(&bug, id).Error; err != nil {
		utils.Error(c, 404, "Bug不存在")
		return
	}

	// 权限检查：普通用户只能删除自己创建或参与的Bug
	if !utils.CheckBugAccess(h.db, c, bug.ID) {
		utils.Error(c, 403, "没有权限删除该Bug")
		return
	}

	if err := h.db.Delete(&model.Bug{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// UpdateBugStatus 更新Bug状态
func (h *BugHandler) UpdateBugStatus(c *gin.Context) {
	id := c.Param("id")
	var bug model.Bug
	if err := h.db.First(&bug, id).Error; err != nil {
		utils.Error(c, 404, "Bug不存在")
		return
	}

	// 权限检查：普通用户只能更新自己创建或参与的Bug
	if !utils.CheckBugAccess(h.db, c, bug.ID) {
		utils.Error(c, 403, "没有权限更新该Bug")
		return
	}

	var req struct {
		Status            string   `json:"status" binding:"required"`
		Solution          *string  `json:"solution"`           // 解决方案
		SolutionNote      *string  `json:"solution_note"`      // 解决方案备注
		EstimatedHours   *float64 `json:"estimated_hours"`    // 预估工时
		ActualHours       *float64 `json:"actual_hours"`      // 实际工时
		WorkDate          *string  `json:"work_date"`          // 工作日期（YYYY-MM-DD），用于资源分配
		ResolvedVersionID *uint    `json:"resolved_version_id"` // 解决版本ID
		VersionNumber     *string  `json:"version_number"`     // 版本号（如果创建新版本）
		CreateVersion     *bool    `json:"create_version"`     // 是否创建新版本
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证状态
	validStatuses := map[string]bool{
		"open":      true,
		"assigned":  true,
		"in_progress": true,
		"resolved":  true,
		"closed":    true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	// 验证解决方案（如果提供了）
	if req.Solution != nil {
		validSolutions := map[string]bool{
			"设计如此":     true,
			"重复Bug":    true,
			"外部原因":    true,
			"已解决":      true,
			"无法重现":    true,
			"延期处理":    true,
			"不予解决":    true,
			"转为研发需求": true,
		}
		if !validSolutions[*req.Solution] {
			utils.Error(c, 400, "解决方案值无效")
			return
		}
		bug.Solution = *req.Solution
	}

	// 更新解决方案备注
	if req.SolutionNote != nil {
		bug.SolutionNote = *req.SolutionNote
	}

	// 处理版本号
	var resolvedVersionID *uint
	if req.CreateVersion != nil && *req.CreateVersion && req.VersionNumber != nil && *req.VersionNumber != "" {
		// 创建新版本
		version := model.Version{
			VersionNumber: *req.VersionNumber,
			ReleaseNotes:  "Bug修复版本",
			Status:        "draft",
			ProjectID:     bug.ProjectID,
		}
		if err := h.db.Create(&version).Error; err != nil {
			utils.Error(c, utils.CodeError, "创建版本失败")
			return
		}
		// 关联当前Bug到新版本
		h.db.Model(&version).Association("Bugs").Append(&bug)
		resolvedVersionID = &version.ID
	} else if req.ResolvedVersionID != nil {
		// 使用已有版本
		// 验证版本是否存在且属于同一项目
		var version model.Version
		if err := h.db.First(&version, *req.ResolvedVersionID).Error; err != nil {
			utils.Error(c, 400, "版本不存在")
			return
		}
		if version.ProjectID != bug.ProjectID {
			utils.Error(c, 400, "版本必须属于同一项目")
			return
		}
		resolvedVersionID = req.ResolvedVersionID
		// 关联当前Bug到版本
		h.db.Model(&version).Association("Bugs").Append(&bug)
	}

	if resolvedVersionID != nil {
		bug.ResolvedVersionID = resolvedVersionID
	}

	// 更新预估工时
	if req.EstimatedHours != nil {
		if *req.EstimatedHours < 0 {
			utils.Error(c, 400, "预估工时不能为负数")
			return
		}
		bug.EstimatedHours = req.EstimatedHours
	}

	// 更新实际工时（如果提供了）
	if req.ActualHours != nil {
		if *req.ActualHours < 0 {
			utils.Error(c, 400, "实际工时不能为负数")
			return
		}
		// 先加载分配人信息
		h.db.Preload("Assignees").First(&bug, bug.ID)
		
		// 如果有分配人，创建或更新资源分配
		if len(bug.Assignees) > 0 {
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
				workDate = time.Now()
			}
			workDate = time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, workDate.Location())
			
			// 为第一个分配人同步到资源分配
			if err := h.syncBugActualHours(&bug, *req.ActualHours, workDate, bug.Assignees[0].ID); err != nil {
				utils.Error(c, utils.CodeError, "同步资源分配失败: "+err.Error())
				return
			}
			// 从资源分配中汇总实际工时（确保actual_hours正确）
			h.calculateAndUpdateActualHours(&bug)
		} else {
			// 如果没有分配人，直接设置actual_hours，但不创建资源分配
			bug.ActualHours = req.ActualHours
		}
	}

	bug.Status = req.Status
	if err := h.db.Save(&bug).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module").Preload("ResolvedVersion").First(&bug, bug.ID)

	utils.Success(c, bug)
}

// GetBugStatistics 获取Bug统计
func (h *BugHandler) GetBugStatistics(c *gin.Context) {
	var stats struct {
		Total           int64 `json:"total"`
		Open            int64 `json:"open"`
		Assigned        int64 `json:"assigned"`
		InProgress      int64 `json:"in_progress"`
		Resolved        int64 `json:"resolved"`
		Closed          int64 `json:"closed"`
		LowPriority     int64 `json:"low_priority"`
		MediumPriority  int64 `json:"medium_priority"`
		HighPriority    int64 `json:"high_priority"`
		UrgentPriority  int64 `json:"urgent_priority"`
		LowSeverity     int64 `json:"low_severity"`
		MediumSeverity  int64 `json:"medium_severity"`
		HighSeverity    int64 `json:"high_severity"`
		CriticalSeverity int64 `json:"critical_severity"`
	}

	baseQuery := h.db.Model(&model.Bug{})

	// 权限过滤：普通用户只能看到自己创建或参与的Bug
	baseQuery = utils.FilterBugsByUser(h.db, c, baseQuery)

	// 应用筛选条件（与列表查询保持一致）
	if keyword := c.Query("keyword"); keyword != "" {
		baseQuery = baseQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if projectID := c.Query("project_id"); projectID != "" {
		baseQuery = baseQuery.Where("project_id = ?", projectID)
	}
	if requirementID := c.Query("requirement_id"); requirementID != "" {
		baseQuery = baseQuery.Where("requirement_id = ?", requirementID)
	}
	if creatorID := c.Query("creator_id"); creatorID != "" {
		baseQuery = baseQuery.Where("creator_id = ?", creatorID)
	}

	// 统计总数
	baseQuery.Session(&gorm.Session{}).Count(&stats.Total)

	// 按状态统计
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "open").Count(&stats.Open)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "assigned").Count(&stats.Assigned)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "in_progress").Count(&stats.InProgress)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "resolved").Count(&stats.Resolved)
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "closed").Count(&stats.Closed)

	// 按优先级统计
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "low").Count(&stats.LowPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "medium").Count(&stats.MediumPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "high").Count(&stats.HighPriority)
	baseQuery.Session(&gorm.Session{}).Where("priority = ?", "urgent").Count(&stats.UrgentPriority)

	// 按严重程度统计
	baseQuery.Session(&gorm.Session{}).Where("severity = ?", "low").Count(&stats.LowSeverity)
	baseQuery.Session(&gorm.Session{}).Where("severity = ?", "medium").Count(&stats.MediumSeverity)
	baseQuery.Session(&gorm.Session{}).Where("severity = ?", "high").Count(&stats.HighSeverity)
	baseQuery.Session(&gorm.Session{}).Where("severity = ?", "critical").Count(&stats.CriticalSeverity)

	utils.Success(c, stats)
}

// AssignBug 分配Bug给用户
func (h *BugHandler) AssignBug(c *gin.Context) {
	id := c.Param("id")
	var bug model.Bug
	if err := h.db.First(&bug, id).Error; err != nil {
		utils.Error(c, 404, "Bug不存在")
		return
	}

	// 权限检查：普通用户只能分配自己创建或参与的Bug
	if !utils.CheckBugAccess(h.db, c, bug.ID) {
		utils.Error(c, 403, "没有权限分配该Bug")
		return
	}

	var req struct {
		AssigneeIDs []uint `json:"assignee_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证用户是否存在
	var users []model.User
	if err := h.db.Where("id IN ?", req.AssigneeIDs).Find(&users).Error; err != nil || len(users) != len(req.AssigneeIDs) {
		utils.Error(c, 400, "用户不存在")
		return
	}

	// 分配Bug
	if err := h.db.Model(&bug).Association("Assignees").Replace(users); err != nil {
		utils.Error(c, utils.CodeError, "分配失败")
		return
	}

	// 如果有分配人且状态为open，自动变为assigned
	if bug.Status == "open" {
		bug.Status = "assigned"
		h.db.Save(&bug)
	}

	// 重新加载关联数据
	h.db.Preload("Project").Preload("Creator").Preload("Assignees").Preload("Requirement").Preload("Module").Preload("ResolvedVersion").First(&bug, bug.ID)

	utils.Success(c, bug)
}

// syncBugActualHours 同步Bug实际工时到资源分配
func (h *BugHandler) syncBugActualHours(bug *model.Bug, actualHours float64, workDate time.Time, assigneeID uint) error {
	// 查找或创建资源
	var resource model.Resource
	err := h.db.Where("user_id = ? AND project_id = ?", assigneeID, bug.ProjectID).First(&resource).Error
	if err != nil {
		// 资源不存在，创建资源
		resource = model.Resource{
			UserID:    assigneeID,
			ProjectID: bug.ProjectID,
		}
		if err := h.db.Create(&resource).Error; err != nil {
			return err
		}
	}

	// 查找是否已存在该Bug和日期的资源分配
	var allocation model.ResourceAllocation
	err = h.db.Where("resource_id = ? AND bug_id = ? AND date = ?", resource.ID, bug.ID, workDate).First(&allocation).Error
	if err != nil {
		// 不存在，创建新的资源分配
		allocation = model.ResourceAllocation{
			ResourceID:  resource.ID,
			BugID:       &bug.ID,
			ProjectID:   &bug.ProjectID,
			Date:        workDate,
			Hours:       actualHours,
			Description: fmt.Sprintf("Bug: %s", bug.Title),
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

// calculateAndUpdateActualHours 计算并更新Bug的实际工时（从资源分配中汇总）
func (h *BugHandler) calculateAndUpdateActualHours(bug *model.Bug) {
	var totalHours float64
	h.db.Model(&model.ResourceAllocation{}).
		Where("bug_id = ?", bug.ID).
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalHours)

	bug.ActualHours = &totalHours
	h.db.Model(bug).Update("actual_hours", totalHours)
}
