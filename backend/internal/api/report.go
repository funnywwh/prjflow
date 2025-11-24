package api

import (
	"fmt"
	"time"

	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// summarizeWorkContent 汇总用户的工作内容
// userID: 用户ID
// startDate: 开始日期
// endDate: 结束日期（对于日报，startDate和endDate相同）
// 返回：Markdown格式的工作内容摘要和总工时
func (h *ReportHandler) summarizeWorkContent(userID uint, startDate, endDate time.Time) (string, float64) {
	// 使用 JOIN 查询，直接通过 user_id 查询资源分配记录
	// 这样可以确保查询到所有相关的资源分配记录，即使没有Resource记录也能查询到
	var allocations []model.ResourceAllocation

	// 将日期转换为只包含日期的格式（去掉时间部分）
	// date字段是date类型，直接比较日期部分即可
	// 注意：使用与工作台相同的日期范围逻辑：date >= startDate AND date < endDate+1
	startDateOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDateOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())
	// 结束日期加1天，使用 < 比较（与工作台逻辑一致）
	endDateExclusive := endDateOnly.AddDate(0, 0, 1)

	// 使用 JOIN 查询，通过 resources 表关联 user_id
	// GORM 会自动处理软删除（deleted_at IS NULL）
	// 注意：如果用户没有 Resource 记录，JOIN 查询可能返回空结果，这是正常的
	err := h.db.Model(&model.ResourceAllocation{}).
		Joins("JOIN resources ON resource_allocations.resource_id = resources.id").
		Where("resources.user_id = ?", userID).
		Where("resource_allocations.date >= ? AND resource_allocations.date < ?", startDateOnly, endDateExclusive).
		Preload("Task").Preload("Bug").Preload("Requirement").Preload("Project").
		Find(&allocations).Error

	// 如果查询出错，返回空内容（不返回错误信息，避免前端显示错误）
	if err != nil {
		// 记录错误但不返回给前端，避免影响用户体验
		// log.Printf("汇总工作内容查询失败: userID=%d, startDate=%s, endDate=%s, error=%v",
		// 	userID, startDateOnly.Format("2006-01-02"), endDateOnly.Format("2006-01-02"), err)
		// 即使资源分配查询失败，也继续查询创建的bug
		allocations = []model.ResourceAllocation{}
	}

	// 查询用户在指定日期范围内创建的Bug（即使没有资源分配记录）
	// 使用时间范围比较：created_at >= startDateOnly AND created_at < endDateExclusive
	// 与资源分配的日期范围逻辑保持一致
	var createdBugs []model.Bug
	bugErr := h.db.Where("creator_id = ?", userID).
		Where("created_at >= ? AND created_at < ?", startDateOnly, endDateExclusive).
		Preload("Project").
		Find(&createdBugs).Error
	if bugErr != nil {
		// 查询失败不影响其他汇总，继续处理
		createdBugs = []model.Bug{}
	}

	// 如果既没有资源分配记录，也没有创建的bug，返回空内容
	if len(allocations) == 0 && len(createdBugs) == 0 {
		return "暂无工作记录", 0
	}

	// 按工作类型分组汇总
	type WorkItem struct {
		ID          uint
		Title       string
		ProjectName string
		Hours       float64
	}

	var requirements []WorkItem
	var tasks []WorkItem
	var bugs []WorkItem
	var totalHours float64

	for _, alloc := range allocations {
		totalHours += alloc.Hours

		projectName := "未知项目"
		if alloc.Project != nil {
			projectName = alloc.Project.Name
		}

		if alloc.RequirementID != nil && alloc.Requirement != nil {
			requirements = append(requirements, WorkItem{
				ID:          *alloc.RequirementID,
				Title:       alloc.Requirement.Title,
				ProjectName: projectName,
				Hours:       alloc.Hours,
			})
		} else if alloc.TaskID != nil && alloc.Task != nil {
			tasks = append(tasks, WorkItem{
				ID:          *alloc.TaskID,
				Title:       alloc.Task.Title,
				ProjectName: projectName,
				Hours:       alloc.Hours,
			})
		} else if alloc.BugID != nil && alloc.Bug != nil {
			bugs = append(bugs, WorkItem{
				ID:          *alloc.BugID,
				Title:       alloc.Bug.Title,
				ProjectName: projectName,
				Hours:       alloc.Hours,
			})
		}
	}

	// 添加用户创建的Bug（去重，避免与资源分配中的Bug重复）
	bugIDMap := make(map[uint]bool)
	for _, bug := range bugs {
		bugIDMap[bug.ID] = true
	}

	for _, bug := range createdBugs {
		// 如果这个bug已经在资源分配中，跳过（避免重复）
		if bugIDMap[bug.ID] {
			continue
		}

		projectName := "未知项目"
		if bug.Project.ID > 0 {
			projectName = bug.Project.Name
		}

		bugs = append(bugs, WorkItem{
			ID:          bug.ID,
			Title:       bug.Title,
			ProjectName: projectName,
			Hours:       0, // 创建的bug如果没有资源分配，工时为0
		})
	}

	// 生成Markdown格式的工作内容摘要
	content := ""

	if len(requirements) > 0 {
		content += "## 需求\n\n"
		var reqHours float64
		for _, req := range requirements {
			content += fmt.Sprintf("- **%s** (项目: %s) - %.2f小时\n", req.Title, req.ProjectName, req.Hours)
			reqHours += req.Hours
		}
		content += fmt.Sprintf("\n**需求总工时**: %.2f小时\n\n", reqHours)
	}

	if len(tasks) > 0 {
		content += "## 任务\n\n"
		var taskHours float64
		for _, task := range tasks {
			content += fmt.Sprintf("- **%s** (项目: %s) - %.2f小时\n", task.Title, task.ProjectName, task.Hours)
			taskHours += task.Hours
		}
		content += fmt.Sprintf("\n**任务总工时**: %.2f小时\n\n", taskHours)
	}

	if len(bugs) > 0 {
		content += "## Bug\n\n"
		var bugHours float64
		for _, bug := range bugs {
			content += fmt.Sprintf("- **%s** (项目: %s) - %.2f小时\n", bug.Title, bug.ProjectName, bug.Hours)
			bugHours += bug.Hours
		}
		content += fmt.Sprintf("\n**Bug总工时**: %.2f小时\n\n", bugHours)
	}

	content += fmt.Sprintf("**总工时**: %.2f小时", totalHours)

	return content, totalHours
}

// GetWorkSummary 获取工作内容汇总（不创建报告，只返回汇总内容）
// 用于前端在新增界面时自动填充工作内容
func (h *ReportHandler) GetWorkSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.Error(c, 400, "请提供开始日期和结束日期")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.Error(c, 400, "开始日期格式错误")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.Error(c, 400, "结束日期格式错误")
		return
	}

	if startDate.After(endDate) {
		utils.Error(c, 400, "开始日期不能晚于结束日期")
		return
	}

	content, hours := h.summarizeWorkContent(userID.(uint), startDate, endDate)

	// 调试信息：记录查询参数和结果
	// fmt.Printf("汇总查询 - 用户ID: %d, 开始日期: %s, 结束日期: %s, 结果: 内容长度=%d, 工时=%.2f\n",
	// 	userID.(uint), startDateStr, endDateStr, len(content), hours)

	utils.Success(c, gin.H{
		"content": content,
		"hours":   hours,
	})
}

// GetDailyReports 获取日报列表
func (h *ReportHandler) GetDailyReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	var reports []model.DailyReport
	query := h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver")

	// 检查是否是获取审批列表
	forApproval := c.Query("for_approval") == "true"
	uid := userID.(uint)
	if !utils.IsAdmin(c) {
		if forApproval {
			// 获取需要当前用户审批的报告
			query = query.Where("id IN (SELECT daily_report_id FROM daily_report_approvers WHERE user_id = ?)", uid)
		} else {
			// 只查询自己创建的
			query = query.Where("user_id = ?", uid)
		}
	} else {
		// 管理员可以筛选用户
		if filterUserID := c.Query("user_id"); filterUserID != "" {
			query = query.Where("user_id = ?", filterUserID)
		}
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 日期范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	countQuery := h.db.Model(&model.DailyReport{})
	if !utils.IsAdmin(c) {
		if forApproval {
			// 获取需要当前用户审批的报告
			countQuery = countQuery.Where("id IN (SELECT daily_report_id FROM daily_report_approvers WHERE user_id = ?)", uid)
		} else {
			// 只查询自己创建的
			countQuery = countQuery.Where("user_id = ?", uid)
		}
	} else {
		// 管理员可以筛选用户
		if filterUserID := c.Query("user_id"); filterUserID != "" {
			countQuery = countQuery.Where("user_id = ?", filterUserID)
		}
	}
	if status := c.Query("status"); status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("date DESC").Find(&reports).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDailyReport 获取日报详情
func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.DailyReport
	if err := h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, id).Error; err != nil {
		utils.Error(c, 404, "日报不存在")
		return
	}

	// 权限检查：普通用户可以查看自己的报告或需要审批的报告
	userID, _ := c.Get("user_id")
	uid := userID.(uint)
	if !utils.IsAdmin(c) {
		// 检查是否是报告创建者
		if report.UserID != uid {
			// 检查是否是审批人
			isApprover := false
			for _, approver := range report.Approvers {
				if approver.ID == uid {
					isApprover = true
					break
				}
			}
			if !isApprover {
				utils.Error(c, 403, "没有权限访问该报告")
				return
			}
		}
	}

	utils.Success(c, report)
}

// CreateDailyReport 创建日报
func (h *ReportHandler) CreateDailyReport(c *gin.Context) {
	var req struct {
		Date        string `json:"date" binding:"required"`
		Content     string `json:"content"`
		Status      string `json:"status"`
		ApproverIDs []uint `json:"approver_ids"` // 审批人ID数组（多选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.Error(c, 400, "日期格式错误")
		return
	}

	// 检查是否已存在该日期的日报
	var existingReport model.DailyReport
	if err := h.db.Where("user_id = ? AND date = ?", userID.(uint), date).First(&existingReport).Error; err == nil {
		utils.Error(c, 400, "该日期已存在日报")
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "draft"
	}

	// 自动汇总工作内容（如果用户未提供Content）
	var content string
	if req.Content == "" {
		content, _ = h.summarizeWorkContent(userID.(uint), date, date)
	} else {
		content = req.Content
	}

	report := model.DailyReport{
		Date:    date,
		Content: content,
		Status:  req.Status,
		UserID:  userID.(uint),
	}

	// 创建日报
	if err := h.db.Create(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 关联审批人（多对多）
	if len(req.ApproverIDs) > 0 {
		var approvers []model.User
		if err := h.db.Where("id IN ?", req.ApproverIDs).Find(&approvers).Error; err == nil {
			h.db.Model(&report).Association("Approvers").Replace(approvers)
			// 为每个审批人创建待审批记录
			for _, approver := range approvers {
				approval := model.DailyReportApproval{
					DailyReportID: report.ID,
					ApproverID:    approver.ID,
					Status:        "pending",
				}
				h.db.Create(&approval)
			}
		}
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// UpdateDailyReport 更新日报
func (h *ReportHandler) UpdateDailyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.DailyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "日报不存在")
		return
	}

	// 权限检查：普通用户只能更新自己的报告
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限更新该报告")
		return
	}

	var req struct {
		Date        *string `json:"date"`
		Content     *string `json:"content"`
		Status      *string `json:"status"`
		ApproverIDs []uint  `json:"approver_ids"` // 审批人ID数组（多选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 如果更新日期，检查是否与其他日报冲突
	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			utils.Error(c, 400, "日期格式错误")
			return
		}
		if date != report.Date {
			var existingReport model.DailyReport
			if err := h.db.Where("user_id = ? AND date = ? AND id != ?", report.UserID, date, report.ID).First(&existingReport).Error; err == nil {
				utils.Error(c, 400, "该日期已存在日报")
				return
			}
			report.Date = date
		}
	}

	if req.Content != nil {
		report.Content = *req.Content
	}
	if req.Status != nil {
		report.Status = *req.Status
	}

	// 更新审批人关联（多对多）
	if req.ApproverIDs != nil {
		var approvers []model.User
		if len(req.ApproverIDs) > 0 {
			if err := h.db.Where("id IN ?", req.ApproverIDs).Find(&approvers).Error; err == nil {
				h.db.Model(&report).Association("Approvers").Replace(approvers)
				// 删除旧的审批记录
				h.db.Where("daily_report_id = ?", report.ID).Delete(&model.DailyReportApproval{})
				// 为每个审批人创建待审批记录
				for _, approver := range approvers {
					approval := model.DailyReportApproval{
						DailyReportID: report.ID,
						ApproverID:    approver.ID,
						Status:        "pending",
					}
					h.db.Create(&approval)
				}
			}
		} else {
			// 如果传入空数组，清空所有审批人关联和审批记录
			h.db.Model(&report).Association("Approvers").Clear()
			h.db.Where("daily_report_id = ?", report.ID).Delete(&model.DailyReportApproval{})
		}
	}

	if err := h.db.Save(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// DeleteDailyReport 删除日报
func (h *ReportHandler) DeleteDailyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.DailyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "日报不存在")
		return
	}

	// 权限检查：普通用户只能删除自己的报告
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限删除该报告")
		return
	}

	// 删除关联的审批记录
	h.db.Where("daily_report_id = ?", report.ID).Delete(&model.DailyReportApproval{})
	// 删除关联的审批人关联
	h.db.Model(&report).Association("Approvers").Clear()

	if err := h.db.Delete(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, nil)
}

// UpdateDailyReportStatus 更新日报状态
func (h *ReportHandler) UpdateDailyReportStatus(c *gin.Context) {
	id := c.Param("id")
	var report model.DailyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "日报不存在")
		return
	}

	// 权限检查：普通用户只能更新自己的报告状态（提交），管理员可以审批
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限更新该报告")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	validStatuses := map[string]bool{
		"draft":     true,
		"submitted": true,
		"approved":  true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	report.Status = req.Status
	if err := h.db.Save(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// ApproveDailyReport 审批日报
func (h *ReportHandler) ApproveDailyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.DailyReport
	if err := h.db.Preload("Approvers").First(&report, id).Error; err != nil {
		utils.Error(c, 404, "日报不存在")
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	// 检查当前用户是否是审批人
	isApprover := false
	for _, approver := range report.Approvers {
		if approver.ID == uid {
			isApprover = true
			break
		}
	}

	if !isApprover && !utils.IsAdmin(c) {
		utils.Error(c, 403, "您不是该报告的审批人")
		return
	}

	var req struct {
		Status  string `json:"status" binding:"required"` // approved 或 rejected
		Comment string `json:"comment"`                   // 批注
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		utils.Error(c, 400, "状态必须是 approved 或 rejected")
		return
	}

	// 查找或创建审批记录
	var approval model.DailyReportApproval
	if err := h.db.Where("daily_report_id = ? AND approver_id = ?", report.ID, uid).First(&approval).Error; err != nil {
		// 如果不存在，创建新的审批记录
		approval = model.DailyReportApproval{
			DailyReportID: report.ID,
			ApproverID:    uid,
			Status:        req.Status,
			Comment:       req.Comment,
		}
		if err := h.db.Create(&approval).Error; err != nil {
			utils.Error(c, utils.CodeError, "创建审批记录失败")
			return
		}
	} else {
		// 更新现有审批记录
		approval.Status = req.Status
		approval.Comment = req.Comment
		if err := h.db.Save(&approval).Error; err != nil {
			utils.Error(c, utils.CodeError, "更新审批记录失败")
			return
		}
	}

	// 检查是否所有审批人都已审批
	var pendingCount int64
	h.db.Model(&model.DailyReportApproval{}).Where("daily_report_id = ? AND status = ?", report.ID, "pending").Count(&pendingCount)
	if pendingCount == 0 {
		// 所有审批人都已审批，检查是否有拒绝的
		var rejectedCount int64
		h.db.Model(&model.DailyReportApproval{}).Where("daily_report_id = ? AND status = ?", report.ID, "rejected").Count(&rejectedCount)
		if rejectedCount > 0 {
			report.Status = "rejected"
		} else {
			report.Status = "approved"
		}
		h.db.Save(&report)
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// GetWeeklyReports 获取周报列表
func (h *ReportHandler) GetWeeklyReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	var reports []model.WeeklyReport
	query := h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver")

	// 检查是否是获取审批列表
	forApproval := c.Query("for_approval") == "true"
	uid := userID.(uint)
	if !utils.IsAdmin(c) {
		if forApproval {
			// 获取需要当前用户审批的报告
			query = query.Where("id IN (SELECT weekly_report_id FROM weekly_report_approvers WHERE user_id = ?)", uid)
		} else {
			// 只查询自己创建的
			query = query.Where("user_id = ?", uid)
		}
	} else {
		// 管理员可以筛选用户
		if filterUserID := c.Query("user_id"); filterUserID != "" {
			query = query.Where("user_id = ?", filterUserID)
		}
	}

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 日期范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("week_start >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("week_end <= ?", endDate)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	countQuery := h.db.Model(&model.WeeklyReport{})
	if !utils.IsAdmin(c) {
		if forApproval {
			// 获取需要当前用户审批的报告
			countQuery = countQuery.Where("id IN (SELECT weekly_report_id FROM weekly_report_approvers WHERE user_id = ?)", uid)
		} else {
			// 只查询自己创建的
			countQuery = countQuery.Where("user_id = ?", uid)
		}
	} else {
		// 管理员可以筛选用户
		if filterUserID := c.Query("user_id"); filterUserID != "" {
			countQuery = countQuery.Where("user_id = ?", filterUserID)
		}
	}
	if status := c.Query("status"); status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("week_start DESC").Find(&reports).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetWeeklyReport 获取周报详情
func (h *ReportHandler) GetWeeklyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.WeeklyReport
	if err := h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, id).Error; err != nil {
		utils.Error(c, 404, "周报不存在")
		return
	}

	// 权限检查：普通用户可以查看自己的报告或需要审批的报告
	userID, _ := c.Get("user_id")
	uid := userID.(uint)
	if !utils.IsAdmin(c) {
		// 检查是否是报告创建者
		if report.UserID != uid {
			// 检查是否是审批人
			isApprover := false
			for _, approver := range report.Approvers {
				if approver.ID == uid {
					isApprover = true
					break
				}
			}
			if !isApprover {
				utils.Error(c, 403, "没有权限访问该报告")
				return
			}
		}
	}

	utils.Success(c, report)
}

// CreateWeeklyReport 创建周报
func (h *ReportHandler) CreateWeeklyReport(c *gin.Context) {
	var req struct {
		WeekStart    string `json:"week_start" binding:"required"`
		WeekEnd      string `json:"week_end" binding:"required"`
		Summary      string `json:"summary"`
		NextWeekPlan string `json:"next_week_plan"`
		Status       string `json:"status"`
		ApproverIDs  []uint `json:"approver_ids"` // 审批人ID数组（多选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	// 解析日期
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		utils.Error(c, 400, "周开始日期格式错误")
		return
	}
	weekEnd, err := time.Parse("2006-01-02", req.WeekEnd)
	if err != nil {
		utils.Error(c, 400, "周结束日期格式错误")
		return
	}

	if weekStart.After(weekEnd) {
		utils.Error(c, 400, "周开始日期不能晚于周结束日期")
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "draft"
	}

	// 自动汇总工作内容（如果用户未提供Summary）
	var summary string
	if req.Summary == "" {
		summary, _ = h.summarizeWorkContent(userID.(uint), weekStart, weekEnd)
	} else {
		summary = req.Summary
	}

	report := model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      summary,
		NextWeekPlan: req.NextWeekPlan,
		Status:       req.Status,
		UserID:       userID.(uint),
	}

	// 创建周报
	if err := h.db.Create(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
	}

	// 关联审批人（多对多）
	if len(req.ApproverIDs) > 0 {
		var approvers []model.User
		if err := h.db.Where("id IN ?", req.ApproverIDs).Find(&approvers).Error; err == nil {
			h.db.Model(&report).Association("Approvers").Replace(approvers)
			// 为每个审批人创建待审批记录
			for _, approver := range approvers {
				approval := model.WeeklyReportApproval{
					WeeklyReportID: report.ID,
					ApproverID:     approver.ID,
					Status:         "pending",
				}
				h.db.Create(&approval)
			}
		}
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// UpdateWeeklyReport 更新周报
func (h *ReportHandler) UpdateWeeklyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.WeeklyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "周报不存在")
		return
	}

	// 权限检查：普通用户只能更新自己的报告
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限更新该报告")
		return
	}

	var req struct {
		WeekStart    *string `json:"week_start"`
		WeekEnd      *string `json:"week_end"`
		Summary      *string `json:"summary"`
		NextWeekPlan *string `json:"next_week_plan"`
		Status       *string `json:"status"`
		ApproverIDs  []uint  `json:"approver_ids"` // 审批人ID数组（多选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.WeekStart != nil {
		weekStart, err := time.Parse("2006-01-02", *req.WeekStart)
		if err != nil {
			utils.Error(c, 400, "周开始日期格式错误")
			return
		}
		report.WeekStart = weekStart
	}

	if req.WeekEnd != nil {
		weekEnd, err := time.Parse("2006-01-02", *req.WeekEnd)
		if err != nil {
			utils.Error(c, 400, "周结束日期格式错误")
			return
		}
		report.WeekEnd = weekEnd
	}

	if req.WeekStart != nil && req.WeekEnd != nil {
		if report.WeekStart.After(report.WeekEnd) {
			utils.Error(c, 400, "周开始日期不能晚于周结束日期")
			return
		}
	}

	if req.Summary != nil {
		report.Summary = *req.Summary
	}
	if req.NextWeekPlan != nil {
		report.NextWeekPlan = *req.NextWeekPlan
	}
	if req.Status != nil {
		report.Status = *req.Status
	}

	// 更新审批人关联（多对多）
	if req.ApproverIDs != nil {
		var approvers []model.User
		if len(req.ApproverIDs) > 0 {
			if err := h.db.Where("id IN ?", req.ApproverIDs).Find(&approvers).Error; err == nil {
				h.db.Model(&report).Association("Approvers").Replace(approvers)
				// 删除旧的审批记录
				h.db.Where("weekly_report_id = ?", report.ID).Delete(&model.WeeklyReportApproval{})
				// 为每个审批人创建待审批记录
				for _, approver := range approvers {
					approval := model.WeeklyReportApproval{
						WeeklyReportID: report.ID,
						ApproverID:     approver.ID,
						Status:         "pending",
					}
					h.db.Create(&approval)
				}
			}
		} else {
			// 如果传入空数组，清空所有审批人关联和审批记录
			h.db.Model(&report).Association("Approvers").Clear()
			h.db.Where("weekly_report_id = ?", report.ID).Delete(&model.WeeklyReportApproval{})
		}
	}

	if err := h.db.Save(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// ApproveWeeklyReport 审批周报
func (h *ReportHandler) ApproveWeeklyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.WeeklyReport
	if err := h.db.Preload("Approvers").First(&report, id).Error; err != nil {
		utils.Error(c, 404, "周报不存在")
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	// 检查当前用户是否是审批人
	isApprover := false
	for _, approver := range report.Approvers {
		if approver.ID == uid {
			isApprover = true
			break
		}
	}

	if !isApprover && !utils.IsAdmin(c) {
		utils.Error(c, 403, "您不是该报告的审批人")
		return
	}

	var req struct {
		Status  string `json:"status" binding:"required"` // approved 或 rejected
		Comment string `json:"comment"`                   // 批注
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		utils.Error(c, 400, "状态必须是 approved 或 rejected")
		return
	}

	// 查找或创建审批记录
	var approval model.WeeklyReportApproval
	if err := h.db.Where("weekly_report_id = ? AND approver_id = ?", report.ID, uid).First(&approval).Error; err != nil {
		// 如果不存在，创建新的审批记录
		approval = model.WeeklyReportApproval{
			WeeklyReportID: report.ID,
			ApproverID:     uid,
			Status:         req.Status,
			Comment:        req.Comment,
		}
		if err := h.db.Create(&approval).Error; err != nil {
			utils.Error(c, utils.CodeError, "创建审批记录失败")
			return
		}
	} else {
		// 更新现有审批记录
		approval.Status = req.Status
		approval.Comment = req.Comment
		if err := h.db.Save(&approval).Error; err != nil {
			utils.Error(c, utils.CodeError, "更新审批记录失败")
			return
		}
	}

	// 检查是否所有审批人都已审批
	var pendingCount int64
	h.db.Model(&model.WeeklyReportApproval{}).Where("weekly_report_id = ? AND status = ?", report.ID, "pending").Count(&pendingCount)
	if pendingCount == 0 {
		// 所有审批人都已审批，检查是否有拒绝的
		var rejectedCount int64
		h.db.Model(&model.WeeklyReportApproval{}).Where("weekly_report_id = ? AND status = ?", report.ID, "rejected").Count(&rejectedCount)
		if rejectedCount > 0 {
			report.Status = "rejected"
		} else {
			report.Status = "approved"
		}
		h.db.Save(&report)
	}

	h.db.Preload("User").Preload("Approvers").Preload("ApprovalRecords.Approver").First(&report, report.ID)
	utils.Success(c, report)
}

// DeleteWeeklyReport 删除周报
func (h *ReportHandler) DeleteWeeklyReport(c *gin.Context) {
	id := c.Param("id")
	var report model.WeeklyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "周报不存在")
		return
	}

	// 权限检查：普通用户只能删除自己的报告
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限删除该报告")
		return
	}

	// 删除关联的审批记录
	h.db.Where("weekly_report_id = ?", report.ID).Delete(&model.WeeklyReportApproval{})
	// 删除关联的审批人关联
	h.db.Model(&report).Association("Approvers").Clear()

	if err := h.db.Delete(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, nil)
}

// UpdateWeeklyReportStatus 更新周报状态
func (h *ReportHandler) UpdateWeeklyReportStatus(c *gin.Context) {
	id := c.Param("id")
	var report model.WeeklyReport
	if err := h.db.First(&report, id).Error; err != nil {
		utils.Error(c, 404, "周报不存在")
		return
	}

	// 权限检查：普通用户只能更新自己的报告状态（提交），管理员可以审批
	userID, _ := c.Get("user_id")
	if !utils.IsAdmin(c) && report.UserID != userID.(uint) {
		utils.Error(c, 403, "没有权限更新该报告")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	validStatuses := map[string]bool{
		"draft":     true,
		"submitted": true,
		"approved":  true,
	}
	if !validStatuses[req.Status] {
		utils.Error(c, 400, "状态值无效")
		return
	}

	report.Status = req.Status
	if err := h.db.Save(&report).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	h.db.Preload("User").First(&report, report.ID)
	utils.Success(c, report)
}
