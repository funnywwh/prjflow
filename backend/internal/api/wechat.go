package api

import (
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/wechat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WeChatHandler struct {
	db           *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewWeChatHandler(db *gorm.DB) *WeChatHandler {
	return &WeChatHandler{
		db:           db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// GetWeChatConfig 获取微信配置
func (h *WeChatHandler) GetWeChatConfig(c *gin.Context) {
	var wechatAppIDConfig model.SystemConfig
	var wechatAppSecretConfig model.SystemConfig

	wechatAppID := ""
	wechatAppSecret := ""

	// 从数据库读取微信配置
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err == nil {
		wechatAppID = wechatAppIDConfig.Value
	}

	if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err == nil {
		wechatAppSecret = wechatAppSecretConfig.Value
	}

	utils.Success(c, gin.H{
		"wechat_app_id":     wechatAppID,
		"wechat_app_secret": wechatAppSecret,
	})
}

// SaveWeChatConfig 保存微信配置
func (h *WeChatHandler) SaveWeChatConfig(c *gin.Context) {
	var req struct {
		WeChatAppID     string `json:"wechat_app_id" binding:"required"`
		WeChatAppSecret string `json:"wechat_app_secret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 保存微信配置
	wechatAppIDConfig := model.SystemConfig{
		Key:   "wechat_app_id",
		Value: req.WeChatAppID,
		Type:  "string",
	}
	if err := h.db.Where("key = ?", "wechat_app_id").Assign(model.SystemConfig{Value: req.WeChatAppID, Type: "string"}).FirstOrCreate(&wechatAppIDConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存微信AppID失败")
		return
	}

	wechatAppSecretConfig := model.SystemConfig{
		Key:   "wechat_app_secret",
		Value: req.WeChatAppSecret,
		Type:  "string",
	}
	if err := h.db.Where("key = ?", "wechat_app_secret").Assign(model.SystemConfig{Value: req.WeChatAppSecret, Type: "string"}).FirstOrCreate(&wechatAppSecretConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存微信AppSecret失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "微信配置保存成功",
	})
}

