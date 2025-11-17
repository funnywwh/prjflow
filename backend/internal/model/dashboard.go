package model

import (
	"time"

	"gorm.io/gorm"
)

// UserDashboard 用户工作台配置表（可选，用于保存用户个性化配置）
type UserDashboard struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"index;not null;unique" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Config string `gorm:"type:text" json:"config"` // JSON配置：卡片排序、显示/隐藏等
}

