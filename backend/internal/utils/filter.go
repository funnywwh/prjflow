package utils

import (
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const AdminRoleCode = "admin"

// IsAdmin 判断用户是否是管理员
func IsAdmin(c *gin.Context) bool {
	roles, exists := c.Get("roles")
	if !exists {
		return false
	}

	roleList, ok := roles.([]string)
	if !ok {
		return false
	}

	for _, role := range roleList {
		if role == AdminRoleCode {
			return true
		}
	}

	return false
}

// GetUserID 获取当前用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	id, ok := userID.(uint)
	if !ok {
		return 0
	}

	return id
}

// GetUserProjectIDs 获取用户参与的项目ID列表（优化：一次查询，可缓存）
func GetUserProjectIDs(db *gorm.DB, userID uint) []uint {
	var projectIDs []uint
	db.Model(&model.ProjectMember{}).
		Where("user_id = ?", userID).
		Pluck("project_id", &projectIDs)
	return projectIDs
}

// FilterProjectsByUser 过滤项目查询：普通用户只能看到自己参与的项目
func FilterProjectsByUser(db *gorm.DB, c *gin.Context, query *gorm.DB) *gorm.DB {
	// 管理员可以看到所有项目
	if IsAdmin(c) {
		return query
	}

	userID := GetUserID(c)
	if userID == 0 {
		// 如果没有用户ID，返回空结果
		return query.Where("1 = 0")
	}

	// 普通用户只能看到自己参与的项目
	// 使用 EXISTS 子查询优化性能
	return query.Where(
		"EXISTS (SELECT 1 FROM project_members WHERE project_members.project_id = projects.id AND project_members.user_id = ? AND project_members.deleted_at IS NULL)",
		userID,
	)
}

// FilterRequirementsByUser 过滤需求查询：普通用户只能看到自己创建或参与的需求
func FilterRequirementsByUser(db *gorm.DB, c *gin.Context, query *gorm.DB) *gorm.DB {
	// 管理员可以看到所有需求
	if IsAdmin(c) {
		return query
	}

	userID := GetUserID(c)
	if userID == 0 {
		// 如果没有用户ID，返回空结果
		return query.Where("1 = 0")
	}

	// 先获取用户参与的项目ID列表（优化：一次查询）
	projectIDs := GetUserProjectIDs(db, userID)

	// 普通用户只能看到：
	// 1. 自己创建的需求（creator_id = userID）
	// 2. 自己负责的需求（assignee_id = userID）
	// 3. 自己参与的项目中的需求（project_id IN projectIDs）
	// 使用 OR 条件组合，注意：如果 projectIDs 为空，只检查前两个条件
	if len(projectIDs) > 0 {
		return query.Where(
			"creator_id = ? OR assignee_id = ? OR project_id IN ?",
			userID, userID, projectIDs,
		)
	} else {
		// 如果用户没有参与任何项目，只检查创建者和负责人
		return query.Where(
			"creator_id = ? OR assignee_id = ?",
			userID, userID,
		)
	}
}

// FilterTasksByUser 过滤任务查询：普通用户只能看到自己创建或参与的任务
func FilterTasksByUser(db *gorm.DB, c *gin.Context, query *gorm.DB) *gorm.DB {
	// 管理员可以看到所有任务
	if IsAdmin(c) {
		return query
	}

	userID := GetUserID(c)
	if userID == 0 {
		// 如果没有用户ID，返回空结果
		return query.Where("1 = 0")
	}

	// 先获取用户参与的项目ID列表（优化：一次查询）
	projectIDs := GetUserProjectIDs(db, userID)

	// 普通用户只能看到：
	// 1. 自己创建的任务（creator_id = userID）
	// 2. 自己负责的任务（assignee_id = userID）
	// 3. 自己参与的项目中的任务（project_id IN projectIDs）
	if len(projectIDs) > 0 {
		return query.Where(
			"creator_id = ? OR assignee_id = ? OR project_id IN ?",
			userID, userID, projectIDs,
		)
	} else {
		// 如果用户没有参与任何项目，只检查创建者和负责人
		return query.Where(
			"creator_id = ? OR assignee_id = ?",
			userID, userID,
		)
	}
}

// FilterBugsByUser 过滤Bug查询：普通用户只能看到自己创建或参与的Bug
func FilterBugsByUser(db *gorm.DB, c *gin.Context, query *gorm.DB) *gorm.DB {
	// 管理员可以看到所有Bug
	if IsAdmin(c) {
		return query
	}

	userID := GetUserID(c)
	if userID == 0 {
		// 如果没有用户ID，返回空结果
		return query.Where("1 = 0")
	}

	// 先获取用户参与的项目ID列表（优化：一次查询）
	projectIDs := GetUserProjectIDs(db, userID)

	// 普通用户只能看到：
	// 1. 自己创建的Bug（creator_id = userID）
	// 2. 自己分配的Bug（通过 bug_assignees 表）
	// 3. 自己参与的项目中的Bug（project_id IN projectIDs）
	// 使用 EXISTS 子查询检查分配人关系
	hasProjectIDs := len(projectIDs) > 0

	if hasProjectIDs {
		return query.Where(
			"creator_id = ? OR project_id IN ? OR EXISTS (SELECT 1 FROM bug_assignees WHERE bug_assignees.bug_id = bugs.id AND bug_assignees.user_id = ?)",
			userID, projectIDs, userID,
		)
	} else {
		// 如果用户没有参与任何项目，只检查创建者和分配人
		return query.Where(
			"creator_id = ? OR EXISTS (SELECT 1 FROM bug_assignees WHERE bug_assignees.bug_id = bugs.id AND bug_assignees.user_id = ?)",
			userID, userID,
		)
	}
}

// CheckProjectAccess 检查用户是否有权限访问项目
func CheckProjectAccess(db *gorm.DB, c *gin.Context, projectID uint) bool {
	// 管理员可以访问所有项目
	if IsAdmin(c) {
		return true
	}

	userID := GetUserID(c)
	if userID == 0 {
		return false
	}

	// 检查用户是否是项目成员
	var count int64
	db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count)

	return count > 0
}

// CheckRequirementAccess 检查用户是否有权限访问需求
func CheckRequirementAccess(db *gorm.DB, c *gin.Context, requirementID uint) bool {
	// 管理员可以访问所有需求
	if IsAdmin(c) {
		return true
	}

	userID := GetUserID(c)
	if userID == 0 {
		return false
	}

	// 查询需求信息
	var requirement model.Requirement
	if err := db.First(&requirement, requirementID).Error; err != nil {
		return false
	}

	// 检查是否是创建者或负责人
	if requirement.CreatorID == userID || (requirement.AssigneeID != nil && *requirement.AssigneeID == userID) {
		return true
	}

	// 检查是否是项目成员
	return CheckProjectAccess(db, c, requirement.ProjectID)
}

// CheckTaskAccess 检查用户是否有权限访问任务
func CheckTaskAccess(db *gorm.DB, c *gin.Context, taskID uint) bool {
	// 管理员可以访问所有任务
	if IsAdmin(c) {
		return true
	}

	userID := GetUserID(c)
	if userID == 0 {
		return false
	}

	// 查询任务信息
	var task model.Task
	if err := db.First(&task, taskID).Error; err != nil {
		return false
	}

	// 检查是否是创建者或负责人
	if task.CreatorID == userID || (task.AssigneeID != nil && *task.AssigneeID == userID) {
		return true
	}

	// 检查是否是项目成员
	return CheckProjectAccess(db, c, task.ProjectID)
}

// CheckBugAccess 检查用户是否有权限访问Bug
func CheckBugAccess(db *gorm.DB, c *gin.Context, bugID uint) bool {
	// 管理员可以访问所有Bug
	if IsAdmin(c) {
		return true
	}

	userID := GetUserID(c)
	if userID == 0 {
		return false
	}

	// 查询Bug信息
	var bug model.Bug
	if err := db.Preload("Assignees").First(&bug, bugID).Error; err != nil {
		return false
	}

	// 检查是否是创建者
	if bug.CreatorID == userID {
		return true
	}

	// 检查是否是分配人
	for _, assignee := range bug.Assignees {
		if assignee.ID == userID {
			return true
		}
	}

	// 检查是否是项目成员
	return CheckProjectAccess(db, c, bug.ProjectID)
}





