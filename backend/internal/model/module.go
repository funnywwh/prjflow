package model

import (
	"time"

	"gorm.io/gorm"
)

// Module 功能模块表（系统资源，不属于项目）
type Module struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"` // 模块名称（唯一）
	Code        string `gorm:"size:50;uniqueIndex" json:"code"`            // 模块编码（唯一）
	Description string `gorm:"type:text" json:"description"`              // 模块描述
	Status      int    `gorm:"default:1" json:"status"`                    // 状态：1-正常，0-禁用
	Sort        int    `gorm:"default:0" json:"sort"`                     // 排序

	Bugs []Bug `gorm:"foreignKey:ModuleID" json:"bugs,omitempty"`
}

