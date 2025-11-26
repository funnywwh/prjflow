package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// StringArray 字符串数组类型，用于JSON序列化
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// TestCase 测试单表
type TestCase struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string      `gorm:"size:200;not null" json:"name"`        // 测试单名称
	Description string      `gorm:"type:text" json:"description"`         // 测试描述
	TestSteps   string      `gorm:"type:text" json:"test_steps"`          // 测试步骤（Markdown）
	Types       StringArray  `gorm:"type:text" json:"types"`                // 测试类型（多选）：functional, performance, security, etc. (JSON数组)
	Status      string       `gorm:"size:20;default:'wait'" json:"status"` // 状态：wait(待评审), normal(正常), blocked(被阻塞), investigate(研究中)
	Result      string       `gorm:"size:20" json:"result"`                 // 测试结果：passed, failed, blocked（合并自TestReport）
	Summary     string       `gorm:"type:text" json:"summary"`              // 测试摘要（合并自TestReport）

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	Bugs    []Bug        `gorm:"many2many:test_case_bugs;" json:"bugs,omitempty"`
}

// TestCaseBug 测试单-Bug关联表
type TestCaseBug struct {
	TestCaseID uint `gorm:"primaryKey" json:"test_case_id"`
	BugID      uint `gorm:"primaryKey" json:"bug_id"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

