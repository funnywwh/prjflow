package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/auth"
	"project-management/pkg/wechat"
)

type AuthHandler struct {
	db          *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db:          db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// GetQRCode 获取微信登录二维码
func (h *AuthHandler) GetQRCode(c *gin.Context) {
	qrCode, err := h.wechatClient.GetQRCode()
	if err != nil {
		utils.Error(c, utils.CodeError, "获取二维码失败")
		return
	}

	qrCodeURL := h.wechatClient.GetQRCodeURL(qrCode.Ticket)

	utils.Success(c, gin.H{
		"ticket":        qrCode.Ticket,
		"qr_code_url":   qrCodeURL,
		"expire_seconds": qrCode.ExpireSeconds,
	})
}

// WeChatLogin 微信登录
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 获取access_token
	accessTokenResp, err := h.wechatClient.GetAccessToken(req.Code)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取access_token失败")
		return
	}

	// 获取用户信息
	userInfo, err := h.wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取用户信息失败")
		return
	}

	// 查找或创建用户
	var user model.User
	result := h.db.Where("wechat_open_id = ?", userInfo.OpenID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		user = model.User{
			WeChatOpenID: userInfo.OpenID,
			Username:     userInfo.Nickname,
			Avatar:       userInfo.HeadImgURL,
			Status:       1,
		}
		if err := h.db.Create(&user).Error; err != nil {
			utils.Error(c, utils.CodeError, "创建用户失败")
			return
		}
	} else if result.Error != nil {
		utils.Error(c, utils.CodeError, "查询用户失败")
		return
	} else {
		// 更新用户信息
		user.Username = userInfo.Nickname
		user.Avatar = userInfo.HeadImgURL
		h.db.Save(&user)
	}

	// 获取用户角色
	var roles []model.Role
	h.db.Model(&user).Association("Roles").Find(&roles)
	
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Code)
	}

	// 生成JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, roleNames)
	if err != nil {
		utils.Error(c, utils.CodeError, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"roles":    roleNames,
		},
	})
}

// GetUserInfo 获取当前用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	var user model.User
	if err := h.db.Preload("Department").Preload("Roles").First(&user, userID).Error; err != nil {
		utils.Error(c, utils.CodeError, "用户不存在")
		return
	}

	roleNames := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Code)
	}

	utils.Success(c, gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"avatar":      user.Avatar,
		"phone":       user.Phone,
		"department":  user.Department,
		"roles":       roleNames,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	utils.Success(c, gin.H{
		"message": "登出成功",
	})
}

