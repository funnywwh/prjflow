package model

import (
	"time"

	"gorm.io/gorm"
)

// UserTableColumnSetting 用户表格列设置表
type UserTableColumnSetting struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	User      User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Page      string `gorm:"size:50;not null;index" json:"page"`      // 页面标识，如 "bug", "task" 等
	ColumnKey string `gorm:"size:50;not null" json:"column_key"`     // 列标识
	Visible   bool   `gorm:"type:boolean;default:true" json:"visible"`             // 是否显示
	Order     int    `gorm:"default:0" json:"order"`                  // 排序
	Width     *int   `gorm:"" json:"width,omitempty"`                 // 列宽（可选）

	// 唯一约束：同一用户同一页面的同一列只能有一条记录
	// 使用复合索引实现
}

// TableName 指定表名
func (UserTableColumnSetting) TableName() string {
	return "user_table_column_settings"
}

