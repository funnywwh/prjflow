package model

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务表
type Task struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:200;not null" json:"title"`        // 任务标题
	Description string `gorm:"type:text" json:"description"`         // 任务描述（Markdown）
	Status      string `gorm:"size:20;default:'todo'" json:"status"`  // 状态：todo, in_progress, done, cancelled
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // 优先级：low, medium, high, urgent

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	RequirementID *uint       `gorm:"index" json:"requirement_id"`
	Requirement   *Requirement `gorm:"foreignKey:RequirementID" json:"requirement,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	AssigneeID *uint `gorm:"index" json:"assignee_id"`
	Assignee   *User `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`

	StartDate *time.Time `json:"start_date"` // 开始日期
	EndDate   *time.Time `json:"end_date"`   // 结束日期
	DueDate   *time.Time `json:"due_date"`   // 截止日期

	Progress int `gorm:"default:0" json:"progress"` // 进度：0-100

	EstimatedHours *float64 `gorm:"default:0" json:"estimated_hours"` // 预估工时（小时）
	ActualHours    *float64 `gorm:"default:0" json:"actual_hours"`    // 实际工时（小时），从资源分配自动计算

	// 任务依赖关系（多对多）
	Dependencies []Task `gorm:"many2many:task_dependencies;joinForeignKey:task_id;joinReferences:dependency_id" json:"dependencies,omitempty"`
}

// TaskDependency 任务依赖关系表
type TaskDependency struct {
	TaskID       uint `gorm:"primaryKey" json:"task_id"`
	DependencyID uint `gorm:"primaryKey" json:"dependency_id"`
	Type         string `gorm:"size:20;default:'finish_to_start'" json:"type"` // 依赖类型：finish_to_start, start_to_start, finish_to_finish, start_to_finish
}

// Board 看板表
type Board struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:100;not null" json:"name"`        // 看板名称
	Description string `gorm:"type:text" json:"description"`            // 描述

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	Columns []BoardColumn `gorm:"foreignKey:BoardID" json:"columns,omitempty"`
}

// BoardColumn 看板列表
type BoardColumn struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name  string `gorm:"size:50;not null" json:"name"`  // 列名称
	Color string `gorm:"size:20" json:"color"`           // 列颜色
	Sort  int    `gorm:"default:0" json:"sort"`          // 排序

	BoardID uint   `gorm:"index;not null" json:"board_id"`
	Board   Board  `gorm:"foreignKey:BoardID" json:"board,omitempty"`

	Status string `gorm:"size:20" json:"status"` // 关联的任务状态
}

