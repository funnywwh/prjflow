package api

import (
	"strings"

	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/websocket"

	"github.com/gin-gonic/gin"
)

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// AddUserCallbackHandler 添加用户场景的微信回调处理
type AddUserCallbackHandler struct {
	db *gorm.DB
}

func (h *AddUserCallbackHandler) Validate(ctx *WeChatCallbackContext) error {
	// 添加用户场景无需特殊验证
	return nil
}

func (h *AddUserCallbackHandler) Process(ctx *WeChatCallbackContext) (interface{}, error) {
	// 检查用户是否已存在（包括软删除的记录）
	var existingUser model.User
	result := ctx.DB.Unscoped().Where("wechat_open_id = ?", ctx.UserInfo.OpenID).First(&existingUser)
	if result.Error == nil {
		// 用户已存在（可能是软删除的）
		if existingUser.DeletedAt.Valid {
			// 如果是软删除的用户，恢复它并更新信息
			// 使用 Unscoped().Update 清除软删除标记并更新字段
			updates := map[string]interface{}{
				"deleted_at": nil,
				"avatar":     ctx.UserInfo.HeadImgURL,
			}
			// 更新昵称（如果微信昵称不为空）
			if ctx.UserInfo.Nickname != "" {
				updates["nickname"] = ctx.UserInfo.Nickname
			} else if existingUser.Nickname == "" {
				// 如果昵称为空，使用用户名
				updates["nickname"] = existingUser.Username
			}
			if err := ctx.DB.Unscoped().Model(&existingUser).Updates(updates).Error; err != nil {
				return nil, &CallbackError{Message: "恢复用户失败", Err: err}
			}
			// 重新加载用户信息
			ctx.DB.Preload("Department").Preload("Roles").First(&existingUser, existingUser.ID)
			
			// 返回恢复的用户信息
			if ctx.Ticket != "" {
				websocket.GetHub().SendMessage(ctx.Ticket, "success", gin.H{
					"user": gin.H{
						"id":            existingUser.ID,
						"username":      existingUser.Username,
						"nickname":      existingUser.Nickname,
						"email":         existingUser.Email,
						"avatar":        existingUser.Avatar,
						"wechat_open_id": existingUser.WeChatOpenID,
					},
				}, "用户已恢复")
			}
			return gin.H{
				"user": gin.H{
					"id":            existingUser.ID,
					"username":      existingUser.Username,
					"nickname":      existingUser.Nickname,
					"email":         existingUser.Email,
					"avatar":        existingUser.Avatar,
					"wechat_open_id": existingUser.WeChatOpenID,
				},
			}, nil
		} else {
			// 用户已存在且未删除
		return nil, &CallbackError{Message: "该微信用户已存在"}
		}
	} else if result.Error != gorm.ErrRecordNotFound {
		// 查询出错
		return nil, &CallbackError{Message: "查询用户失败", Err: result.Error}
	}

	// 确保昵称不为空（如果微信昵称为空，使用默认值）
	nickname := ctx.UserInfo.Nickname
	if nickname == "" {
		nickname = "用户"
	}

	// 生成唯一的用户名并创建用户（带重试机制处理并发冲突）
	var user model.User
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
	// 生成唯一的用户名（如果昵称已存在，自动添加数字后缀）
		username := GenerateUniqueUsername(ctx.DB, nickname, ctx.UserInfo.OpenID)
		
		// 如果昵称为空，使用用户名
		if nickname == "" {
			nickname = username
		}

	// 创建新用户
		wechatOpenID := ctx.UserInfo.OpenID
		user = model.User{
		WeChatOpenID: &wechatOpenID,
		Username:     username,
			Nickname:     nickname, // 设置昵称（从微信昵称获取，如果为空则使用用户名）
		Avatar:       ctx.UserInfo.HeadImgURL,
		Status:       1,
	}
		
		err := ctx.DB.Create(&user).Error
		if err == nil {
			// 创建成功
			break
		}
		
		// 如果是唯一约束错误，重试
		errStr := err.Error()
		isUniqueError := errStr == "UNIQUE constraint failed: users.username" ||
			contains(errStr, "UNIQUE constraint failed") ||
			contains(errStr, "Duplicate entry") ||
			contains(errStr, "duplicate key")
		
		if i < maxRetries-1 && isUniqueError {
			// 等待一小段时间后重试（避免立即重试导致相同结果）
			continue
		}
		
		// 其他错误或达到最大重试次数
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
				"nickname":      user.Nickname,
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
			"nickname":      user.Nickname,
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

