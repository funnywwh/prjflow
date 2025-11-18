package api

import (
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/websocket"

	"github.com/gin-gonic/gin"
)

// AddUserCallbackHandler 添加用户场景的微信回调处理
type AddUserCallbackHandler struct {
	db *gorm.DB
}

func (h *AddUserCallbackHandler) Validate(ctx *WeChatCallbackContext) error {
	// 添加用户场景无需特殊验证
	return nil
}

func (h *AddUserCallbackHandler) Process(ctx *WeChatCallbackContext) (interface{}, error) {
	// 检查用户是否已存在
	var existingUser model.User
	result := ctx.DB.Where("wechat_open_id = ?", ctx.UserInfo.OpenID).First(&existingUser)
	if result.Error == nil {
		// 用户已存在
		return nil, &CallbackError{Message: "该微信用户已存在"}
	} else if result.Error != gorm.ErrRecordNotFound {
		// 查询出错
		return nil, &CallbackError{Message: "查询用户失败", Err: result.Error}
	}

	// 创建新用户
	user := model.User{
		WeChatOpenID: ctx.UserInfo.OpenID,
		Username:     ctx.UserInfo.Nickname,
		Avatar:       ctx.UserInfo.HeadImgURL,
		Status:       1,
	}
	if err := ctx.DB.Create(&user).Error; err != nil {
		return nil, &CallbackError{Message: "创建用户失败", Err: err}
	}

	// 加载用户信息（包含关联数据）
	ctx.DB.Preload("Department").Preload("Roles").First(&user, user.ID)

	// 通过WebSocket通知成功
	if ctx.Ticket != "" {
		websocket.GetHub().SendMessage(ctx.Ticket, "info", nil, "用户添加成功")
		websocket.GetHub().SendMessage(ctx.Ticket, "success", gin.H{
			"user": gin.H{
				"id":            user.ID,
				"username":      user.Username,
				"email":         user.Email,
				"avatar":        user.Avatar,
				"wechat_open_id": user.WeChatOpenID,
			},
		}, "用户添加成功")
	}

	return gin.H{
		"user": gin.H{
			"id":            user.ID,
			"username":      user.Username,
			"email":         user.Email,
			"avatar":        user.Avatar,
			"wechat_open_id": user.WeChatOpenID,
		},
	}, nil
}

func (h *AddUserCallbackHandler) GetSuccessHTML(ctx *WeChatCallbackContext, data interface{}) string {
	return GetDefaultSuccessHTML("用户添加成功", "请返回 PC 端查看")
}

func (h *AddUserCallbackHandler) GetErrorHTML(ctx *WeChatCallbackContext, err error) string {
	return GetDefaultErrorHTML("添加用户失败", err.Error())
}

