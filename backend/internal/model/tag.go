package model

import (
	"time"

	"gorm.io/gorm"
)

// Tag 标签表
type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:50;not null;uniqueIndex" json:"name"` // 标签名称
	Description string `gorm:"size:200" json:"description"`                // 标签描述
	Color       string `gorm:"size:20;default:'blue'" json:"color"`      // 标签颜色（用于前端显示）

	// 关联项目（多对多）
	Projects []Project `gorm:"many2many:project_tags;" json:"projects,omitempty"`
}

// ProjectTag 项目标签关联表（GORM自动创建，这里定义用于查询）
type ProjectTag struct {
	ProjectID uint `gorm:"primaryKey"`
	TagID     uint `gorm:"primaryKey"`
}

