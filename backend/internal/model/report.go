package model

import (
	"time"
)

// DailyReport 日报表
type DailyReport struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Date        time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_date" json:"date"` // 日期
	Content     string    `gorm:"type:text" json:"content"`                                  // 工作内容（Markdown）
	Status      string    `gorm:"size:20;default:'draft'" json:"status"`                    // 状态：draft, submitted, approved

	UserID uint `gorm:"index;not null;uniqueIndex:idx_user_date" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 审批人多对多关联
	Approvers []User `gorm:"many2many:daily_report_approvers;" json:"approvers,omitempty"`

	// 审批记录
	ApprovalRecords []DailyReportApproval `gorm:"foreignKey:DailyReportID" json:"approval_records,omitempty"`
}

// DailyReportApprover 日报审批人关联表
type DailyReportApprover struct {
	DailyReportID uint `gorm:"primaryKey" json:"daily_report_id"`
	UserID        uint `gorm:"primaryKey" json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// DailyReportApproval 日报审批记录表
type DailyReportApproval struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	DailyReportID uint         `gorm:"index;not null" json:"daily_report_id"`
	DailyReport   DailyReport  `gorm:"foreignKey:DailyReportID" json:"daily_report,omitempty"`

	ApproverID uint `gorm:"index;not null" json:"approver_id"`
	Approver   User `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`

	Status  string `gorm:"size:20;default:'pending'" json:"status"` // 审批状态：pending, approved, rejected
	Comment string `gorm:"type:text" json:"comment"`                // 批注
}

// WeeklyReport 周报表
type WeeklyReport struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	WeekStart   time.Time `gorm:"type:date;not null" json:"week_start"`   // 周开始日期
	WeekEnd     time.Time `gorm:"type:date;not null" json:"week_end"`      // 周结束日期
	Summary     string    `gorm:"type:text" json:"summary"`                // 工作总结（Markdown）
	NextWeekPlan string   `gorm:"type:text" json:"next_week_plan"`         // 下周计划（Markdown）
	Status      string    `gorm:"size:20;default:'draft'" json:"status"`    // 状态：draft, submitted, approved

	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 审批人多对多关联
	Approvers []User `gorm:"many2many:weekly_report_approvers;" json:"approvers,omitempty"`

	// 审批记录
	ApprovalRecords []WeeklyReportApproval `gorm:"foreignKey:WeeklyReportID" json:"approval_records,omitempty"`
}

// WeeklyReportApprover 周报审批人关联表
type WeeklyReportApprover struct {
	WeeklyReportID uint `gorm:"primaryKey" json:"weekly_report_id"`
	UserID         uint `gorm:"primaryKey" json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// WeeklyReportApproval 周报审批记录表
type WeeklyReportApproval struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	WeeklyReportID uint         `gorm:"index;not null" json:"weekly_report_id"`
	WeeklyReport   WeeklyReport `gorm:"foreignKey:WeeklyReportID" json:"weekly_report,omitempty"`

	ApproverID uint `gorm:"index;not null" json:"approver_id"`
	Approver   User `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`

	Status  string `gorm:"size:20;default:'pending'" json:"status"` // 审批状态：pending, approved, rejected
	Comment string `gorm:"type:text" json:"comment"`                  // 批注
}

