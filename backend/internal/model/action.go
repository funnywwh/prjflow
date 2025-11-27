package model

import (
	"time"
)

// Action 操作记录表（参考禅道的 zt_action 表）
type Action struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ObjectType string `gorm:"size:30;not null;index:idx_object" json:"object_type"` // 对象类型：bug, task, requirement等
	ObjectID   uint    `gorm:"not null;index:idx_object" json:"object_id"`         // 对象ID
	ProjectID  uint    `gorm:"index" json:"project_id"`                            // 项目ID（用于快速查询项目相关操作）
	ActorID    uint    `gorm:"index" json:"actor_id"`                                // 操作人ID
	Actor      User    `gorm:"foreignKey:ActorID" json:"actor,omitempty"`           // 操作人关联

	Action  string `gorm:"size:80;not null;index" json:"action"` // 操作类型：created, edited, assigned, resolved, closed, confirmed, commented
	Date    time.Time `gorm:"index" json:"date"`                  // 操作时间
	Comment string    `gorm:"type:text" json:"comment"`          // 备注/评论内容
	Extra   string    `gorm:"type:text" json:"extra"`             // 额外信息（JSON格式，存储操作相关的额外数据）

	// 关联历史记录
	Histories []History `gorm:"foreignKey:ActionID" json:"histories,omitempty"`
}

// History 字段变更记录表（参考禅道的 zt_history 表）
type History struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ActionID uint   `gorm:"not null;index" json:"action_id"` // 关联Action表的ID
	Action   Action `gorm:"foreignKey:ActionID" json:"action,omitempty"`

	Field    string `gorm:"size:30;not null" json:"field"`     // 字段名（如 'status', 'priority', 'assignee_ids'）
	Old      string `gorm:"type:text" json:"old"`              // 旧值（原始值，可能是ID或代码值，如 'active', 用户ID等）
	OldValue string `gorm:"type:text" json:"old_value"`       // 旧值显示文本（转换后的可读值，如 '激活', 用户名等）
	New      string `gorm:"type:text" json:"new"`              // 新值（原始值）
	NewValue string `gorm:"type:text" json:"new_value"`       // 新值显示文本（转换后的可读值）
	Diff     string `gorm:"type:text" json:"diff"`             // 差异对比（用于文本字段的diff显示，可选）
}

