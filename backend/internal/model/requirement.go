package model

import (
	"time"

	"gorm.io/gorm"
)

// Requirement 需求表
type Requirement struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:200;not null" json:"title"`        // 需求标题
	Description string `gorm:"type:text" json:"description"`        // 需求描述（Markdown）
	Status      string `gorm:"size:20;default:'draft'" json:"status"` // 状态：draft(草稿), reviewing(评审中), active(激活), changing(变更中), closed(已关闭)
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // 优先级：low, medium, high, urgent

	ProjectID uint    `gorm:"index;not null" json:"project_id"` // 必填关联项目
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	AssigneeID *uint `gorm:"index" json:"assignee_id"`
	Assignee   *User `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`

	EstimatedHours *float64 `gorm:"default:0" json:"estimated_hours"` // 预估工时（小时）
	ActualHours    *float64 `gorm:"default:0" json:"actual_hours"`    // 实际工时（小时），从资源分配自动计算
}

// Bug Bug表
type Bug struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:200;not null" json:"title"`        // Bug标题
	Description string `gorm:"type:text" json:"description"`         // Bug描述（Markdown）
	Status      string `gorm:"size:20;default:'active'" json:"status"` // 状态：active(激活), resolved(已解决), closed(已关闭)
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // 优先级：low, medium, high, urgent
	Severity    string `gorm:"size:20;default:'medium'" json:"severity"` // 严重程度：low, medium, high, critical

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	Assignees []User `gorm:"many2many:bug_assignees;" json:"assignees,omitempty"`

	RequirementID *uint       `gorm:"index" json:"requirement_id"` // 关联需求
	Requirement   *Requirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`

	ModuleID *uint  `gorm:"index" json:"module_id"` // 关联功能模块
	Module   *Module `gorm:"foreignKey:ModuleID" json:"module,omitempty"`

	EstimatedHours *float64 `gorm:"default:0" json:"estimated_hours"` // 预估工时（小时）
	ActualHours    *float64 `gorm:"default:0" json:"actual_hours"`    // 实际工时（小时），从资源分配自动计算

	// 解决方案相关
	Solution       string `gorm:"size:50" json:"solution"`        // 解决方案：设计如此、重复Bug、外部原因、已解决、无法重现、延期处理、不予解决、转为研发需求
	SolutionNote   string `gorm:"type:text" json:"solution_note"` // 解决方案备注
	ResolvedVersionID *uint   `gorm:"index" json:"resolved_version_id"` // 解决版本ID
	ResolvedVersion   *Version `gorm:"foreignKey:ResolvedVersionID" json:"resolved_version,omitempty"`
}

// BugAssignee Bug分配表
type BugAssignee struct {
	BugID    uint `gorm:"primaryKey" json:"bug_id"`
	UserID   uint `gorm:"primaryKey" json:"user_id"`
	AssignedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"assigned_at"`
}

