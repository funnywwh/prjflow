package model

import (
	"time"

	"gorm.io/gorm"
)

// EntityRelation 实体关系配置表（可选，用于自定义关系图展示）
type EntityRelation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:100;not null" json:"name"`        // 关系名称
	SourceType  string `gorm:"size:50;not null" json:"source_type"`  // 源实体类型
	SourceID    uint   `gorm:"not null" json:"source_id"`            // 源实体ID
	TargetType  string `gorm:"size:50;not null" json:"target_type"`  // 目标实体类型
	TargetID    uint   `gorm:"not null" json:"target_id"`             // 目标实体ID
	RelationType string `gorm:"size:50" json:"relation_type"`         // 关系类型
	Description string `gorm:"type:text" json:"description"`            // 描述
}

