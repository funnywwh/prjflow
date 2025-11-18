package api

import (
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/websocket"
	"project-management/pkg/auth"

	"github.com/gin-gonic/gin"
)

// InitCallbackHandlerImpl 初始化场景的微信回调处理
type InitCallbackHandlerImpl struct {
	db *gorm.DB
}

func (h *InitCallbackHandlerImpl) Validate(ctx *WeChatCallbackContext) error {
	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := ctx.DB.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		return &CallbackError{Message: "系统已经初始化，无法重复初始化"}
	}

	// 检查微信配置是否已保存
	// 注意：微信配置的读取和设置已经在 ProcessWeChatCallback 中完成，这里只需要验证配置是否存在
	var wechatAppIDConfig model.SystemConfig
	if err := ctx.DB.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		return &CallbackError{Message: "请先配置微信AppID和AppSecret"}
	}

	return nil
}

func (h *InitCallbackHandlerImpl) Process(ctx *WeChatCallbackContext) (interface{}, error) {
	// 开始事务
	tx := ctx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建管理员角色
	var adminRole model.Role
	if err := tx.Where("code = ?", "admin").First(&adminRole).Error; err == gorm.ErrRecordNotFound {
		adminRole = model.Role{
			Name:        "管理员",
			Code:        "admin",
			Description: "系统管理员，拥有所有权限",
			Status:      1,
		}
		if err := tx.Create(&adminRole).Error; err != nil {
			tx.Rollback()
			return nil, &CallbackError{Message: "创建管理员角色失败", Err: err}
		}
	}

	// 2. 创建管理员用户
	// 生成唯一的用户名（如果昵称已存在，自动添加数字后缀）
	adminUsername := GenerateUniqueUsername(tx, ctx.UserInfo.Nickname, ctx.UserInfo.OpenID)
	
	// 确保昵称不为空（如果微信昵称为空，使用用户名作为默认昵称）
	adminNickname := ctx.UserInfo.Nickname
	if adminNickname == "" {
		adminNickname = adminUsername
	}
	
	adminUser := model.User{
		WeChatOpenID: ctx.UserInfo.OpenID,
		Username:     adminUsername,
		Nickname:     adminNickname, // 设置昵称（从微信昵称获取，如果为空则使用用户名）
		Avatar:       ctx.UserInfo.HeadImgURL,
		Status:       1,
	}
	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		return nil, &CallbackError{Message: "创建管理员用户失败", Err: err}
	}

	// 3. 分配管理员角色
	if err := tx.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		tx.Rollback()
		return nil, &CallbackError{Message: "分配管理员角色失败", Err: err}
	}

	// 4. 标记系统已初始化
	initConfig := model.SystemConfig{
		Key:   "initialized",
		Value: "true",
		Type:  "boolean",
	}
	if err := tx.Create(&initConfig).Error; err != nil {
		tx.Rollback()
		return nil, &CallbackError{Message: "标记系统初始化失败", Err: err}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, &CallbackError{Message: "初始化失败", Err: err}
	}

	// 生成管理员Token
	roleNames := []string{"admin"}
	token, err := auth.GenerateToken(adminUser.ID, adminUser.Username, roleNames)
	if err != nil {
		return nil, &CallbackError{Message: "生成Token失败", Err: err}
	}

	// 通过WebSocket通知PC端成功
	if ctx.Ticket != "" {
		websocket.GetHub().SendMessage(ctx.Ticket, "success", gin.H{
			"token": token,
			"user": gin.H{
				"id":       adminUser.ID,
				"username": adminUser.Username,
				"avatar":   adminUser.Avatar,
				"roles":    roleNames,
			},
		}, "系统初始化成功")
	}

	return gin.H{
		"token": token,
		"user": gin.H{
			"id":       adminUser.ID,
			"username": adminUser.Username,
			"avatar":   adminUser.Avatar,
			"roles":    roleNames,
		},
	}, nil
}

func (h *InitCallbackHandlerImpl) GetSuccessHTML(ctx *WeChatCallbackContext, data interface{}) string {
	return GetDefaultSuccessHTML("系统初始化成功", "请返回 PC 端查看")
}

func (h *InitCallbackHandlerImpl) GetErrorHTML(ctx *WeChatCallbackContext, err error) string {
	return GetDefaultErrorHTML("初始化失败", err.Error())
}

