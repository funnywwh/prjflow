package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
)

// WeChatVerifyHandler 处理微信验证文件
type WeChatVerifyHandler struct {
	db *gorm.DB
}

func NewWeChatVerifyHandler(db *gorm.DB) *WeChatVerifyHandler {
	return &WeChatVerifyHandler{
		db: db,
	}
}

// HandleVerifyFile 处理微信域名验证文件请求
// 支持两种方式：
// 1. 从数据库读取验证文件内容（推荐，可以动态配置）
// 2. 从文件系统读取验证文件（如果文件已上传到服务器）
func (h *WeChatVerifyHandler) HandleVerifyFile(c *gin.Context) {
	// 从路由参数获取验证码，例如：xJWzw9UUmpBJxEok.txt
	// 注意：Gin 路由参数会包含 .txt 后缀（如果 URL 中包含）
	code := c.Param("code")
	
	if code == "" {
		c.String(404, "File not found")
		return
	}

	// 构建完整文件名
	// 如果 code 已经包含 .txt，直接使用；否则添加 .txt
	filename := "MP_verify_" + code
	if !strings.HasSuffix(filename, ".txt") {
		filename += ".txt"
	}

	// 方式1：从数据库读取验证文件内容（推荐）
	var verifyConfig model.SystemConfig
	result := h.db.Where("key = ?", "wechat_verify_file_"+filename).First(&verifyConfig)
	if result.Error == nil {
		// 从数据库读取成功，返回文件内容
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(200, verifyConfig.Value)
		return
	}

	// 方式2：从文件系统读取（如果文件已上传到服务器）
	// 这里可以添加从文件系统读取的逻辑
	// 例如：读取 static/wechat/ 目录下的文件
	
	c.String(404, "Verification file not found. Please upload the file content via API.")
}

// SaveVerifyFile 保存微信验证文件内容到数据库
// 可以通过API接口上传验证文件内容，避免手动上传文件
func (h *WeChatVerifyHandler) SaveVerifyFile(c *gin.Context) {
	var req struct {
		Filename string `json:"filename" binding:"required"` // 例如：MP_verify_xJWzw9UUmpBJxEok.txt
		Content  string `json:"content" binding:"required"`  // 文件内容
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 保存到数据库
	config := model.SystemConfig{
		Key:   "wechat_verify_file_" + req.Filename,
		Value: req.Content,
		Type:  "string",
	}

	// 如果已存在，更新；否则创建
	var existing model.SystemConfig
	result := h.db.Where("key = ?", config.Key).First(&existing)
	if result.Error == nil {
		// 更新
		existing.Value = req.Content
		if err := h.db.Save(&existing).Error; err != nil {
			c.JSON(500, gin.H{"error": "保存失败: " + err.Error()})
			return
		}
	} else {
		// 创建
		if err := h.db.Create(&config).Error; err != nil {
			c.JSON(500, gin.H{"error": "保存失败: " + err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{
		"message": "验证文件保存成功",
		"filename": req.Filename,
		"url": "/" + req.Filename, // 返回访问URL
	})
}

