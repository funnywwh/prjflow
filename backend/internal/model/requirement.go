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
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // 状态：pending, in_progress, completed, cancelled
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // 优先级：low, medium, high, urgent

	ProductID *uint   `gorm:"index" json:"product_id"` // 可选关联产品
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	ProjectID *uint   `gorm:"index" json:"project_id"` // 可选关联项目
	Project   *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	AssigneeID *uint `gorm:"index" json:"assignee_id"`
	Assignee   *User `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}

// Bug Bug表
type Bug struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:200;not null" json:"title"`        // Bug标题
	Description string `gorm:"type:text" json:"description"`         // Bug描述（Markdown）
	Status      string `gorm:"size:20;default:'open'" json:"status"` // 状态：open, assigned, in_progress, resolved, closed
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // 优先级：low, medium, high, urgent
	Severity    string `gorm:"size:20;default:'medium'" json:"severity"` // 严重程度：low, medium, high, critical

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	Assignees []User `gorm:"many2many:bug_assignees;" json:"assignees,omitempty"`

	RequirementID *uint       `gorm:"index" json:"requirement_id"` // 关联需求
	Requirement   *Requirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`

	EstimatedHours *float64 `gorm:"default:0" json:"estimated_hours"` // 预估工时（小时）
	ActualHours    *float64 `gorm:"default:0" json:"actual_hours"`    // 实际工时（小时），从资源分配自动计算
}

// BugAssignee Bug分配表
type BugAssignee struct {
	BugID    uint `gorm:"primaryKey" json:"bug_id"`
	UserID   uint `gorm:"primaryKey" json:"user_id"`
	AssignedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"assigned_at"`
}

