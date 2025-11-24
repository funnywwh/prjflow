package main

import (
	"fmt"
	"strings"
	"time"
)

// 状态转换函数

// ConvertProjectStatus 转换项目状态
func ConvertProjectStatus(status string) int {
	switch strings.ToLower(status) {
	case "doing":
		return 1 // 正常
	case "done", "closed":
		return 0 // 禁用
	case "wait":
		return 1 // 正常
	default:
		return 1 // 默认正常
	}
}

// ConvertRequirementStatus 转换需求状态
func ConvertRequirementStatus(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "in_progress"
	case "closed":
		return "completed"
	case "draft":
		return "pending"
	case "changing":
		return "pending"
	case "reviewing":
		return "pending"
	default:
		return "pending"
	}
}

// ConvertTaskStatus 转换任务状态
func ConvertTaskStatus(status string) string {
	switch strings.ToLower(status) {
	case "wait":
		return "todo"
	case "doing":
		return "in_progress"
	case "done":
		return "done"
	case "pause", "cancel":
		return "cancelled"
	case "closed":
		return "done"
	default:
		return "todo"
	}
}

// ConvertBugStatus 转换Bug状态
func ConvertBugStatus(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "open"
	case "resolved":
		return "resolved"
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

// ConvertPriority 转换优先级 (zentao: 1-4, goproject: urgent/high/medium/low)
func ConvertPriority(pri int) string {
	switch pri {
	case 1:
		return "urgent"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	default:
		return "medium"
	}
}

// ConvertSeverity 转换严重程度 (zentao: 1-4, goproject: critical/high/medium/low)
func ConvertSeverity(severity int) string {
	switch severity {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	default:
		return "medium"
	}
}

// ConvertUserStatus 转换用户状态
func ConvertUserStatus(deleted string) int {
	if deleted == "1" {
		return 0 // 禁用
	}
	return 1 // 正常
}

// GenerateDeptCode 生成部门编码
func GenerateDeptCode(name string, id int) string {
	// 简单实现：使用ID作为编码，如果名称是中文则使用拼音首字母
	// 这里简化处理，使用 dept_ + ID
	return fmt.Sprintf("dept_%d", id)
}

// GenerateRoleCode 生成角色编码
func GenerateRoleCode(name string) string {
	// 转换为小写，替换空格和特殊字符
	code := strings.ToLower(name)
	code = strings.ReplaceAll(code, " ", "_")
	code = strings.ReplaceAll(code, "-", "_")
	code = strings.ReplaceAll(code, ".", "_")
	// 移除其他特殊字符
	var result strings.Builder
	for _, r := range code {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	code = result.String()
	if code == "" {
		code = "role"
	}
	return code
}

// MapZenTaoPermissionToGoProject 将zentao权限映射到goproject权限代码
func MapZenTaoPermissionToGoProject(module, method string) string {
	// 根据zentao的module和method，映射到goproject的权限代码
	// 这里需要根据实际的权限体系进行映射
	
	// 示例映射规则
	module = strings.ToLower(module)
	method = strings.ToLower(method)
	
	// 项目相关
	if module == "project" {
		switch method {
		case "create":
			return "project:create"
		case "edit", "update":
			return "project:update"
		case "view", "index", "browse":
			return "project:read"
		case "delete":
			return "project:delete"
		}
	}
	
	// 需求相关
	if module == "story" {
		switch method {
		case "create":
			return "requirement:create"
		case "edit", "change":
			return "requirement:update"
		case "view", "index", "browse":
			return "requirement:read"
		case "delete":
			return "requirement:delete"
		}
	}
	
	// 任务相关
	if module == "task" {
		switch method {
		case "create":
			return "task:create"
		case "edit", "update":
			return "task:update"
		case "view", "index", "browse":
			return "task:read"
		case "delete":
			return "task:delete"
		}
	}
	
	// Bug相关
	if module == "bug" {
		switch method {
		case "create":
			return "bug:create"
		case "edit", "update":
			return "bug:update"
		case "view", "index", "browse":
			return "bug:read"
		case "delete":
			return "bug:delete"
		case "assign":
			return "bug:assign"
		}
	}
	
	// 用户相关
	if module == "user" {
		switch method {
		case "create":
			return "user:create"
		case "edit", "update":
			return "user:update"
		case "view", "index", "browse":
			return "user:read"
		case "delete":
			return "user:delete"
		}
	}
	
	// 部门相关
	if module == "dept" || module == "department" {
		switch method {
		case "create":
			return "department:create"
		case "edit", "update":
			return "department:update"
		case "view", "index", "browse":
			return "department:read"
		case "delete":
			return "department:delete"
		}
	}
	
	// 默认返回空，表示无法映射
	return ""
}

// ParseDateTime 解析日期时间字符串
func ParseDateTime(dateStr string) *time.Time {
	if dateStr == "" || dateStr == "0000-00-00 00:00:00" || dateStr == "0000-00-00" {
		return nil
	}
	
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}
	
	return nil
}

// ParseDate 解析日期字符串
func ParseDate(dateStr string) *time.Time {
	if dateStr == "" || dateStr == "0000-00-00" {
		return nil
	}
	
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return &t
	}
	
	return nil
}

// DaysToHours 将天数转换为小时（假设1天=8小时）
func DaysToHours(days float64) *float64 {
	if days <= 0 {
		return nil
	}
	hours := days * 8
	return &hours
}

