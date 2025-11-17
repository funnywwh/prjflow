package model

import (
	"time"

	"gorm.io/gorm"
)

// Plugin 插件表
type Plugin struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:100;not null" json:"name"`        // 插件名称
	Code        string `gorm:"size:50;uniqueIndex" json:"code"`      // 插件代码
	Version     string `gorm:"size:20" json:"version"`               // 版本
	Description string `gorm:"type:text" json:"description"`          // 描述
	Path        string `gorm:"size:255" json:"path"`                 // 插件路径
	Status      string `gorm:"size:20;default:'disabled'" json:"status"` // 状态：enabled, disabled
	Author      string `gorm:"size:100" json:"author"`               // 作者
	Homepage    string `gorm:"size:255" json:"homepage"`             // 主页

	Configs []PluginConfig `gorm:"foreignKey:PluginID" json:"configs,omitempty"`
	Hooks   []PluginHook   `gorm:"foreignKey:PluginID" json:"hooks,omitempty"`
	Routes  []PluginRoute  `gorm:"foreignKey:PluginID" json:"routes,omitempty"`
}

// PluginConfig 插件配置表
type PluginConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PluginID uint   `gorm:"index;not null" json:"plugin_id"`
	Plugin   Plugin `gorm:"foreignKey:PluginID" json:"plugin,omitempty"`

	Key   string `gorm:"size:100;not null" json:"key"`   // 配置键
	Value string `gorm:"type:text" json:"value"`         // 配置值
	Type  string `gorm:"size:20" json:"type"`            // 配置类型：string, number, boolean, json
}

// PluginHook 插件钩子表
type PluginHook struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PluginID uint   `gorm:"index;not null" json:"plugin_id"`
	Plugin   Plugin `gorm:"foreignKey:PluginID" json:"plugin,omitempty"`

	HookName string `gorm:"size:100;not null" json:"hook_name"` // 钩子名称
	Handler  string `gorm:"size:255" json:"handler"`            // 处理器
	Priority int    `gorm:"default:10" json:"priority"`          // 优先级
}

// PluginRoute 插件路由表
type PluginRoute struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PluginID uint   `gorm:"index;not null" json:"plugin_id"`
	Plugin   Plugin `gorm:"foreignKey:PluginID" json:"plugin,omitempty"`

	Path      string `gorm:"size:255;not null" json:"path"`       // 路由路径
	Name      string `gorm:"size:100" json:"name"`                // 路由名称
	Component  string `gorm:"size:255" json:"component"`         // 组件路径
	MenuTitle string `gorm:"size:100" json:"menu_title"`          // 菜单标题
	MenuIcon  string `gorm:"size:50" json:"menu_icon"`            // 菜单图标
	MenuOrder int    `gorm:"default:0" json:"menu_order"`        // 菜单排序
	Hidden    bool   `gorm:"default:false" json:"hidden"`         // 是否隐藏
}

