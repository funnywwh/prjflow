package model

import (
	"time"

	"gorm.io/gorm"
)

// Plan 计划表
type Plan struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:200;not null" json:"name"`        // 计划名称
	Description string `gorm:"type:text" json:"description"`         // 计划描述（Markdown）
	Type        string `gorm:"size:20;not null" json:"type"`         // 类型：product_plan, project_plan
	Status      string `gorm:"size:20;default:'draft'" json:"status"` // 状态：draft, active, completed, cancelled

	StartDate *time.Time `json:"start_date"` // 开始日期
	EndDate   *time.Time `json:"end_date"`   // 结束日期

	ProductID *uint   `gorm:"index" json:"product_id"` // 产品计划关联产品
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	ProjectID *uint   `gorm:"index" json:"project_id"` // 项目计划关联项目
	Project   *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	Executions []PlanExecution `gorm:"foreignKey:PlanID" json:"executions,omitempty"`
}

// PlanExecution 计划执行表
type PlanExecution struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:200;not null" json:"name"`        // 执行名称
	Description string `gorm:"type:text" json:"description"`         // 执行描述
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // 状态：pending, in_progress, completed, cancelled
	Progress    int    `gorm:"default:0" json:"progress"`             // 进度：0-100

	PlanID uint `gorm:"index;not null" json:"plan_id"`
	Plan   Plan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`

	StartDate *time.Time `json:"start_date"` // 开始日期
	EndDate   *time.Time `json:"end_date"`   // 结束日期

	AssigneeID *uint `gorm:"index" json:"assignee_id"`
	Assignee   *User `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
}

