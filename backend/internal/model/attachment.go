package model

import (
	"time"

	"gorm.io/gorm"
)

// Attachment 附件表
type Attachment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	FileName string `gorm:"size:255;not null" json:"file_name"` // 原始文件名
	FilePath string `gorm:"size:500;not null" json:"file_path"`  // 文件存储路径（相对路径）
	FileSize int64  `gorm:"not null" json:"file_size"`          // 文件大小（字节）
	MimeType string `gorm:"size:100" json:"mime_type"`           // MIME类型

	CreatorID uint `gorm:"index;not null" json:"creator_id"` // 创建人ID
	Creator   User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`

	// 多对多关联：项目、需求、任务、Bug
	Projects     []Project     `gorm:"many2many:project_attachments;" json:"projects,omitempty"`
	Requirements []Requirement `gorm:"many2many:requirement_attachments;" json:"requirements,omitempty"`
	Tasks        []Task        `gorm:"many2many:task_attachments;" json:"tasks,omitempty"`
	Bugs         []Bug         `gorm:"many2many:bug_attachments;" json:"bugs,omitempty"`
}

// ProjectAttachment 项目附件关联表
type ProjectAttachment struct {
	ProjectID    uint `gorm:"primaryKey" json:"project_id"`
	AttachmentID uint `gorm:"primaryKey" json:"attachment_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// RequirementAttachment 需求附件关联表
type RequirementAttachment struct {
	RequirementID uint `gorm:"primaryKey" json:"requirement_id"`
	AttachmentID  uint `gorm:"primaryKey" json:"attachment_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// TaskAttachment 任务附件关联表
type TaskAttachment struct {
	TaskID       uint `gorm:"primaryKey" json:"task_id"`
	AttachmentID uint `gorm:"primaryKey" json:"attachment_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// BugAttachment Bug附件关联表
type BugAttachment struct {
	BugID        uint `gorm:"primaryKey" json:"bug_id"`
	AttachmentID uint `gorm:"primaryKey" json:"attachment_id"`
	CreatedAt    time.Time `json:"created_at"`
}

