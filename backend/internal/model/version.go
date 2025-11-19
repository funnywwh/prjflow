package model

import (
	"time"

	"gorm.io/gorm"
)

// Version 版本表
type Version struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	VersionNumber string `gorm:"size:50;not null" json:"version_number"` // 版本号
	ReleaseNotes  string `gorm:"type:text" json:"release_notes"`         // 发布说明（Markdown）
	Status        string `gorm:"size:20;default:'draft'" json:"status"`   // 状态：draft, released, archived

	ProjectID uint    `gorm:"index;not null" json:"project_id"` // 关联项目
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	ReleaseDate *time.Time `json:"release_date"` // 发布日期

	// 关联需求和Bug
	Requirements []Requirement `gorm:"many2many:version_requirements;" json:"requirements,omitempty"`
	Bugs         []Bug         `gorm:"many2many:version_bugs;" json:"bugs,omitempty"`
}

