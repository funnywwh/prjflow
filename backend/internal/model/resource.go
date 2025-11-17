package model

import (
	"time"

	"gorm.io/gorm"
)

// Resource 人员资源表
type Resource struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	Role string `gorm:"size:50" json:"role"` // 资源角色

	Allocations []ResourceAllocation `gorm:"foreignKey:ResourceID" json:"allocations,omitempty"`
}

// ResourceAllocation 资源分配表（按小时记录）
type ResourceAllocation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ResourceID uint     `gorm:"index;not null" json:"resource_id"`
	Resource   Resource `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`

	Date     time.Time `gorm:"type:date;not null" json:"date"`     // 日期
	Hours    float64   `gorm:"not null" json:"hours"`              // 小时数
	TaskID   *uint     `gorm:"index" json:"task_id"`               // 可选关联任务
	Task     *Task     `gorm:"foreignKey:TaskID" json:"task,omitempty"`
	ProjectID *uint    `gorm:"index" json:"project_id"`           // 可选关联项目
	Project   *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	Description string `gorm:"type:text" json:"description"` // 工作描述
}

