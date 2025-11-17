package model

import (
	"time"

	"gorm.io/gorm"
)

// DailyReport 日报表
type DailyReport struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Date        time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_date" json:"date"` // 日期
	Content     string    `gorm:"type:text" json:"content"`                                  // 工作内容（Markdown）
	Hours       float64   `json:"hours"`                                                    // 工时
	Status      string    `gorm:"size:20;default:'draft'" json:"status"`                    // 状态：draft, submitted, approved

	UserID uint `gorm:"index;not null;uniqueIndex:idx_user_date" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ProjectID *uint   `gorm:"index" json:"project_id"`
	Project   *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	TaskID *uint `gorm:"index" json:"task_id"`
	Task   *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// WeeklyReport 周报表
type WeeklyReport struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	WeekStart   time.Time `gorm:"type:date;not null" json:"week_start"`   // 周开始日期
	WeekEnd     time.Time `gorm:"type:date;not null" json:"week_end"`      // 周结束日期
	Summary     string    `gorm:"type:text" json:"summary"`                // 工作总结（Markdown）
	NextWeekPlan string   `gorm:"type:text" json:"next_week_plan"`         // 下周计划（Markdown）
	Status      string    `gorm:"size:20;default:'draft'" json:"status"`    // 状态：draft, submitted, approved

	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ProjectID *uint   `gorm:"index" json:"project_id"`
	Project   *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	TaskID *uint `gorm:"index" json:"task_id"`
	Task   *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

