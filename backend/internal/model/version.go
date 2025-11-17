package model

import (
	"time"

	"gorm.io/gorm"
)

// Build 构建表
type Build struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BuildNumber string `gorm:"size:50;not null" json:"build_number"` // 构建号
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // 状态：pending, building, success, failed
	Branch      string `gorm:"size:100" json:"branch"`                // 分支
	Commit      string `gorm:"size:100" json:"commit"`                // 提交
	BuildTime   *time.Time `json:"build_time"`                         // 构建时间

	ProjectID uint    `gorm:"index;not null" json:"project_id"`
	Project   Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`

	CreatorID uint `gorm:"index" json:"creator_id"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// Version 版本表
type Version struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	VersionNumber string `gorm:"size:50;not null" json:"version_number"` // 版本号
	ReleaseNotes  string `gorm:"type:text" json:"release_notes"`         // 发布说明（Markdown）
	Status        string `gorm:"size:20;default:'draft'" json:"status"`   // 状态：draft, released, archived

	BuildID uint   `gorm:"index;not null;unique" json:"build_id"` // 一个构建产生一个版本
	Build   Build  `gorm:"foreignKey:BuildID" json:"build,omitempty"`

	ReleaseDate *time.Time `json:"release_date"` // 发布日期
}

