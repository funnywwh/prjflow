package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttachmentHandler struct {
	db *gorm.DB
}

func NewAttachmentHandler(db *gorm.DB) *AttachmentHandler {
	return &AttachmentHandler{db: db}
}

// UploadFile 上传文件
// 权限要求：登录 + 项目成员验证 + attachment:upload 权限
func (h *AttachmentHandler) UploadFile(c *gin.Context) {
	// 检查权限码
	perms, exists := c.Get("permissions")
	if !exists {
		utils.Error(c, 403, "没有权限")
		c.Abort()
		return
	}

	hasUploadPerm := false
	if permList, ok := perms.([]string); ok {
		for _, perm := range permList {
			if perm == "attachment:upload" {
				hasUploadPerm = true
				break
			}
		}
	}

	if !hasUploadPerm {
		utils.Error(c, 403, "没有上传权限")
		return
	}

	// 获取项目ID（必须）
	projectIDStr := c.PostForm("project_id")
	if projectIDStr == "" {
		utils.Error(c, 400, "项目ID不能为空")
		return
	}

	var projectID uint
	if _, err := fmt.Sscanf(projectIDStr, "%d", &projectID); err != nil {
		utils.Error(c, 400, "无效的项目ID")
		return
	}

	// 验证项目是否存在并检查项目成员权限
	var project model.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		utils.Error(c, 404, "项目不存在")
		return
	}

	if !utils.CheckProjectAccess(h.db, c, projectID) {
		utils.Error(c, 403, "没有权限访问该项目")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "文件上传失败: "+err.Error())
		return
	}

	// 验证文件大小
	if file.Size > config.AppConfig.Upload.MaxFileSize {
		utils.Error(c, 400, fmt.Sprintf("文件大小超过限制（最大 %d MB）", config.AppConfig.Upload.MaxFileSize/(1024*1024)))
		return
	}

	// 验证文件类型（如果配置了允许的类型）
	if len(config.AppConfig.Upload.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range config.AppConfig.Upload.AllowedTypes {
			if strings.HasPrefix(file.Header.Get("Content-Type"), allowedType) {
				allowed = true
				break
			}
		}
		if !allowed {
			utils.Error(c, 400, "不支持的文件类型")
			return
		}
	}

	// 生成文件存储路径（按日期组织：YYYY/MM/DD/）
	now := time.Now()
	datePath := now.Format("2006/01/02")
	
	// 生成唯一文件名（UUID + 原始扩展名）
	ext := filepath.Ext(file.Filename)
	fileName := uuid.New().String() + ext
	
	// 构建完整存储路径
	storagePath := config.AppConfig.Upload.StoragePath
	if !filepath.IsAbs(storagePath) {
		// 相对路径：相对于程序运行目录
		storagePath = filepath.Join(".", storagePath)
	}
	
	fullDir := filepath.Join(storagePath, datePath)
	fullPath := filepath.Join(fullDir, fileName)

	// 创建目录
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		utils.Error(c, utils.CodeError, "创建存储目录失败: "+err.Error())
		return
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		utils.Error(c, utils.CodeError, "保存文件失败: "+err.Error())
		return
	}

	// 获取MIME类型
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 创建附件记录
	userID := utils.GetUserID(c)
	attachment := model.Attachment{
		FileName: file.Filename,
		FilePath: filepath.Join(datePath, fileName), // 存储相对路径
		FileSize: file.Size,
		MimeType: mimeType,
		CreatorID: userID,
	}

	if err := h.db.Create(&attachment).Error; err != nil {
		// 如果创建失败，删除已上传的文件
		os.Remove(fullPath)
		utils.Error(c, utils.CodeError, "创建附件记录失败: "+err.Error())
		return
	}

	// 关联到项目
	if err := h.db.Model(&attachment).Association("Projects").Append(&project); err != nil {
		// 如果关联失败，删除附件记录和文件
		h.db.Delete(&attachment)
		os.Remove(fullPath)
		utils.Error(c, utils.CodeError, "关联附件到项目失败: "+err.Error())
		return
	}

	// 预加载创建人信息
	h.db.Preload("Creator").First(&attachment, attachment.ID)

	utils.Success(c, attachment)
}

// GetAttachment 获取附件信息
// 权限要求：登录 + 项目成员验证
func (h *AttachmentHandler) GetAttachment(c *gin.Context) {
	id := c.Param("id")
	
	var attachment model.Attachment
	if err := h.db.Preload("Creator").Preload("Projects").First(&attachment, id).Error; err != nil {
		utils.Error(c, 404, "附件不存在")
		return
	}

	// 检查权限：用户必须是附件关联的任意一个项目的成员
	hasAccess := false
	
	// 管理员可以访问所有附件
	if utils.IsAdmin(c) {
		hasAccess = true
	} else {
		// 检查用户是否是附件关联的任意一个项目的成员
		for _, project := range attachment.Projects {
			if utils.CheckProjectAccess(h.db, c, project.ID) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		utils.Error(c, 403, "没有权限访问该附件")
		return
	}

	utils.Success(c, attachment)
}

// DownloadFile 下载文件
// 权限要求：登录 + 项目成员验证
func (h *AttachmentHandler) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	
	var attachment model.Attachment
	if err := h.db.Preload("Projects").First(&attachment, id).Error; err != nil {
		utils.Error(c, 404, "附件不存在")
		return
	}

	// 检查权限：用户必须是附件关联的任意一个项目的成员
	hasAccess := false
	
	// 管理员可以访问所有附件
	if utils.IsAdmin(c) {
		hasAccess = true
	} else {
		// 检查用户是否是附件关联的任意一个项目的成员
		for _, project := range attachment.Projects {
			if utils.CheckProjectAccess(h.db, c, project.ID) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		utils.Error(c, 403, "没有权限访问该附件")
		return
	}

	// 构建文件完整路径
	storagePath := config.AppConfig.Upload.StoragePath
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(".", storagePath)
	}
	fullPath := filepath.Join(storagePath, attachment.FilePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.Error(c, 404, "文件不存在")
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.FileName))
	c.Header("Content-Type", attachment.MimeType)
	c.File(fullPath)
}

// DeleteAttachment 删除附件
// 权限要求：登录 + 项目成员验证 + attachment:delete 权限
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	// 检查权限码
	perms, exists := c.Get("permissions")
	if !exists {
		utils.Error(c, 403, "没有权限")
		c.Abort()
		return
	}

	hasDeletePerm := false
	if permList, ok := perms.([]string); ok {
		for _, perm := range permList {
			if perm == "attachment:delete" {
				hasDeletePerm = true
				break
			}
		}
	}

	if !hasDeletePerm {
		utils.Error(c, 403, "没有删除权限")
		return
	}

	id := c.Param("id")
	
	var attachment model.Attachment
	if err := h.db.Preload("Projects").First(&attachment, id).Error; err != nil {
		utils.Error(c, 404, "附件不存在")
		return
	}

	// 检查权限：用户必须是附件关联的任意一个项目的成员
	hasAccess := false
	
	// 管理员可以删除所有附件
	if utils.IsAdmin(c) {
		hasAccess = true
	} else {
		// 检查用户是否是附件关联的任意一个项目的成员
		for _, project := range attachment.Projects {
			if utils.CheckProjectAccess(h.db, c, project.ID) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		utils.Error(c, 403, "没有权限删除该附件")
		return
	}

	// 构建文件完整路径
	storagePath := config.AppConfig.Upload.StoragePath
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(".", storagePath)
	}
	fullPath := filepath.Join(storagePath, attachment.FilePath)

	// 删除文件（如果存在）
	if _, err := os.Stat(fullPath); err == nil {
		os.Remove(fullPath)
	}

	// 删除附件记录（软删除）
	if err := h.db.Delete(&attachment).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除附件失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// GetAttachments 获取附件列表
// 权限要求：登录 + 项目成员验证
func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	query := h.db.Model(&model.Attachment{}).Preload("Creator")

	// 支持按项目、需求、任务、Bug过滤
	if projectID := c.Query("project_id"); projectID != "" {
		var pid uint
		if _, err := fmt.Sscanf(projectID, "%d", &pid); err == nil {
			// 验证项目访问权限
			if !utils.CheckProjectAccess(h.db, c, pid) {
				utils.Error(c, 403, "没有权限访问该项目")
				return
			}
			query = query.Joins("JOIN project_attachments ON project_attachments.attachment_id = attachments.id").
				Where("project_attachments.project_id = ?", pid)
		}
	}

	if requirementID := c.Query("requirement_id"); requirementID != "" {
		var rid uint
		if _, err := fmt.Sscanf(requirementID, "%d", &rid); err == nil {
			// 验证需求访问权限
			if !utils.CheckRequirementAccess(h.db, c, rid) {
				utils.Error(c, 403, "没有权限访问该需求")
				return
			}
			query = query.Joins("JOIN requirement_attachments ON requirement_attachments.attachment_id = attachments.id").
				Where("requirement_attachments.requirement_id = ?", rid)
		}
	}

	if taskID := c.Query("task_id"); taskID != "" {
		var tid uint
		if _, err := fmt.Sscanf(taskID, "%d", &tid); err == nil {
			// 验证任务访问权限
			if !utils.CheckTaskAccess(h.db, c, tid) {
				utils.Error(c, 403, "没有权限访问该任务")
				return
			}
			query = query.Joins("JOIN task_attachments ON task_attachments.attachment_id = attachments.id").
				Where("task_attachments.task_id = ?", tid)
		}
	}

	if bugID := c.Query("bug_id"); bugID != "" {
		var bid uint
		if _, err := fmt.Sscanf(bugID, "%d", &bid); err == nil {
			// 验证Bug访问权限
			if !utils.CheckBugAccess(h.db, c, bid) {
				utils.Error(c, 403, "没有权限访问该Bug")
				return
			}
			query = query.Joins("JOIN bug_attachments ON bug_attachments.attachment_id = attachments.id").
				Where("bug_attachments.bug_id = ?", bid)
		}
	}

	var attachments []model.Attachment
	if err := query.Find(&attachments).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败: "+err.Error())
		return
	}

	utils.Success(c, attachments)
}

// AttachToEntity 关联附件到实体（项目/需求/任务/Bug）
// 权限要求：登录 + 项目成员验证
func (h *AttachmentHandler) AttachToEntity(c *gin.Context) {
	attachmentID := c.Param("id")
	
	var attachment model.Attachment
	if err := h.db.Preload("Projects").First(&attachment, attachmentID).Error; err != nil {
		utils.Error(c, 404, "附件不存在")
		return
	}

	var req struct {
		ProjectID     *uint `json:"project_id"`
		RequirementID *uint `json:"requirement_id"`
		TaskID        *uint `json:"task_id"`
		BugID         *uint `json:"bug_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证权限：用户必须是附件关联的任意一个项目的成员
	hasAccess := false
	if utils.IsAdmin(c) {
		hasAccess = true
	} else {
		for _, project := range attachment.Projects {
			if utils.CheckProjectAccess(h.db, c, project.ID) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		utils.Error(c, 403, "没有权限操作该附件")
		return
	}

	// 关联到项目
	if req.ProjectID != nil {
		if !utils.CheckProjectAccess(h.db, c, *req.ProjectID) {
			utils.Error(c, 403, "没有权限访问该项目")
			return
		}
		var project model.Project
		if err := h.db.First(&project, *req.ProjectID).Error; err != nil {
			utils.Error(c, 404, "项目不存在")
			return
		}
		if err := h.db.Model(&attachment).Association("Projects").Append(&project); err != nil {
			utils.Error(c, utils.CodeError, "关联失败: "+err.Error())
			return
		}
	}

	// 关联到需求
	if req.RequirementID != nil {
		if !utils.CheckRequirementAccess(h.db, c, *req.RequirementID) {
			utils.Error(c, 403, "没有权限访问该需求")
			return
		}
		var requirement model.Requirement
		if err := h.db.First(&requirement, *req.RequirementID).Error; err != nil {
			utils.Error(c, 404, "需求不存在")
			return
		}
		if err := h.db.Model(&attachment).Association("Requirements").Append(&requirement); err != nil {
			utils.Error(c, utils.CodeError, "关联失败: "+err.Error())
			return
		}
	}

	// 关联到任务
	if req.TaskID != nil {
		if !utils.CheckTaskAccess(h.db, c, *req.TaskID) {
			utils.Error(c, 403, "没有权限访问该任务")
			return
		}
		var task model.Task
		if err := h.db.First(&task, *req.TaskID).Error; err != nil {
			utils.Error(c, 404, "任务不存在")
			return
		}
		if err := h.db.Model(&attachment).Association("Tasks").Append(&task); err != nil {
			utils.Error(c, utils.CodeError, "关联失败: "+err.Error())
			return
		}
	}

	// 关联到Bug
	if req.BugID != nil {
		if !utils.CheckBugAccess(h.db, c, *req.BugID) {
			utils.Error(c, 403, "没有权限访问该Bug")
			return
		}
		var bug model.Bug
		if err := h.db.First(&bug, *req.BugID).Error; err != nil {
			utils.Error(c, 404, "Bug不存在")
			return
		}
		if err := h.db.Model(&attachment).Association("Bugs").Append(&bug); err != nil {
			utils.Error(c, utils.CodeError, "关联失败: "+err.Error())
			return
		}
	}

	// 重新加载附件信息
	h.db.Preload("Projects").Preload("Requirements").Preload("Tasks").Preload("Bugs").First(&attachment, attachment.ID)

	utils.Success(c, attachment)
}

