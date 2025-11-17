package model

import (
	"time"

	"gorm.io/gorm"
)

// TestCase 测试单表
type TestCase struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:200;not null" json:"name"`        // 测试单名称
	Description string `gorm:"type:text" json:"description"`         // 测试描述
	TestSteps   string `gorm:"type:text" json:"test_steps"`          // 测试步骤（Markdown）
	Type        string `gorm:"size:20" json:"type"`                  // 测试类型：functional, performance, security, etc.
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // 状态：pending, running, passed, failed

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	Bugs    []Bug        `gorm:"many2many:test_case_bugs;" json:"bugs,omitempty"`
	Reports []TestReport `gorm:"many2many:test_case_reports;" json:"reports,omitempty"`
}

// TestReport 测试报告表
type TestReport struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:200;not null" json:"title"`        // 报告标题
	Content     string `gorm:"type:text" json:"content"`             // 报告内容（Markdown）
	Result      string `gorm:"size:20" json:"result"`                 // 测试结果：passed, failed, blocked
	Summary     string `gorm:"type:text" json:"summary"`              // 测试摘要

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	TestCases []TestCase `gorm:"many2many:test_case_reports;" json:"test_cases,omitempty"`
}

// TestCaseBug 测试单-Bug关联表
type TestCaseBug struct {
	TestCaseID uint `gorm:"primaryKey" json:"test_case_id"`
	BugID      uint `gorm:"primaryKey" json:"bug_id"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TestCaseReport 测试单-测试报告关联表
type TestCaseReport struct {
	TestCaseID  uint `gorm:"primaryKey" json:"test_case_id"`
	TestReportID uint `gorm:"primaryKey" json:"test_report_id"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

