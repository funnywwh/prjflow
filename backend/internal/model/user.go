package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	WeChatOpenID *string `gorm:"column:wechat_open_id;uniqueIndex;size:100" json:"wechat_open_id"` // 微信OpenID（指针类型，允许NULL）
	Username     string `gorm:"size:50;not null;uniqueIndex" json:"username"`            // 用户名（唯一，用于登录）
	Nickname     string `gorm:"size:50" json:"nickname"`                       // 昵称（用于前端显示，不能为空，应用层保证）
	Password     string `gorm:"size:255" json:"-"`                       // 密码（不返回给前端）
	Email        string `gorm:"size:100" json:"email"`                       // 邮箱
	Avatar       string `gorm:"size:255" json:"avatar"`                     // 头像URL
	Phone        string `gorm:"size:20" json:"phone"`                       // 手机号
	Status       int    `gorm:"default:1" json:"status"`                      // 状态：1-正常，0-禁用
	LoginCount   int    `gorm:"default:0" json:"login_count"`                 // 登录次数

	DepartmentID *uint      `gorm:"index" json:"department_id"` // 部门ID
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Department 部门表
type Department struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name     string `gorm:"size:100;not null" json:"name"`        // 部门名称
	Code     string `gorm:"size:50;uniqueIndex" json:"code"`      // 部门编码
	ParentID *uint  `gorm:"index" json:"parent_id"`               // 父部门ID
	Parent   *Department `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Level    int    `gorm:"default:1" json:"level"`               // 层级
	Sort     int    `gorm:"default:0" json:"sort"`                // 排序
	Status   int    `gorm:"default:1" json:"status"`             // 状态：1-正常，0-禁用
}

// Role 角色表
type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:50;not null;uniqueIndex" json:"name"` // 角色名称
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"` // 角色代码
	Description string `gorm:"size:255" json:"description"`               // 描述
	Status      int    `gorm:"default:1" json:"status"`                   // 状态：1-正常，0-禁用

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// Permission 权限表
type Permission struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Code        string `gorm:"size:100;not null;uniqueIndex" json:"code"` // 权限代码
	Name        string `gorm:"size:100;not null" json:"name"`               // 权限名称
	Resource    string `gorm:"size:50" json:"resource"`                    // 资源类型
	Action      string `gorm:"size:50" json:"action"`                      // 操作类型
	Description string `gorm:"size:255" json:"description"`                // 描述
	Status      int    `gorm:"default:1" json:"status"`                    // 状态：1-正常，0-禁用

	// 菜单相关字段
	MenuPath      string `gorm:"size:200" json:"menu_path"`       // 菜单路径（路由路径）
	MenuIcon      string `gorm:"size:50" json:"menu_icon"`        // 菜单图标
	MenuTitle     string `gorm:"size:100" json:"menu_title"`      // 菜单标题（如果为空则使用Name）
	ParentMenuID  *uint  `gorm:"index" json:"parent_menu_id"`    // 父菜单ID（用于构建多级菜单）
	MenuOrder     int    `gorm:"default:0" json:"menu_order"`     // 菜单排序
	IsMenu        bool   `gorm:"default:false" json:"is_menu"`   // 是否显示在菜单中

	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

