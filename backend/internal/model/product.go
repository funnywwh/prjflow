package model

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目表
type Project struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string     `gorm:"size:100;not null" json:"name"`        // 项目名称
	Code        string     `gorm:"size:50;uniqueIndex" json:"code"`     // 项目编码
	Description string     `gorm:"type:text" json:"description"`        // 描述
	Status      int        `gorm:"default:1" json:"status"`             // 状态：1-正常，0-禁用
	StartDate   *time.Time `json:"start_date"`                        // 开始日期
	EndDate     *time.Time `json:"end_date"`                          // 结束日期

	Members      []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	Tasks        []Task          `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	Bugs         []Bug           `gorm:"foreignKey:ProjectID" json:"bugs,omitempty"`
	Requirements []Requirement   `gorm:"foreignKey:ProjectID" json:"requirements,omitempty"`
	TestCases    []TestCase      `gorm:"foreignKey:ProjectID" json:"test_cases,omitempty"`
	Plans        []Plan          `gorm:"foreignKey:ProjectID" json:"plans,omitempty"`
	Boards       []Board         `gorm:"foreignKey:ProjectID" json:"boards,omitempty"`
	Tags         []Tag           `gorm:"many2many:project_tags;" json:"tags,omitempty"` // 标签（多对多关联）
}

// ProjectMember 项目成员表
type ProjectMember struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectID uint    `gorm:"index" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	UserID uint `gorm:"index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Role string `gorm:"size:50" json:"role"` // 项目角色：owner, member, viewer
}

