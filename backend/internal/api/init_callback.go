package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/websocket"
	"project-management/pkg/auth"
	"project-management/pkg/wechat"
)

// InitCallbackHandler 处理微信回调
type InitCallbackHandler struct {
	db          *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewInitCallbackHandler(db *gorm.DB) *InitCallbackHandler {
	return &InitCallbackHandler{
		db:          db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// HandleCallback 处理微信授权回调
func (h *InitCallbackHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	// 从state中解析ticket，格式：ticket:uuid
	var ticket string
	if state != "" && len(state) > 7 && state[:7] == "ticket:" {
		ticket = state[7:]
	}

	if code == "" {
		// 通过WebSocket通知PC端错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "未获取到授权码")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<title>初始化失败</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
						color: white;
					}
					.container {
						text-align: center;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>✗ 初始化失败</h1>
					<p>未获取到授权码</p>
				</div>
			</body>
			</html>
		`))
		return
	}

	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "系统已经初始化，无法重复初始化")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<title>初始化失败</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
						color: white;
					}
					.container {
						text-align: center;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>✗ 初始化失败</h1>
					<p>系统已经初始化</p>
				</div>
			</body>
			</html>
		`))
		return
	}

	// 检查微信配置是否已保存
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "请先配置微信AppID和AppSecret")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<title>初始化失败</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						display: flex;
						justify-content: center;
						align-items: center;
						height: 100vh;
						margin: 0;
						background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
						color: white;
					}
					.container {
						text-align: center;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>✗ 初始化失败</h1>
					<p>请先配置微信AppID和AppSecret</p>
				</div>
			</body>
			</html>
		`))
		return
	}

	// 获取微信配置
	var wechatAppSecretConfig model.SystemConfig
	h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig)
	
	// 临时设置WeChatClient配置
	h.wechatClient.AppID = wechatAppIDConfig.Value
	h.wechatClient.AppSecret = wechatAppSecretConfig.Value

	// 获取access_token
	accessTokenResp, err := h.wechatClient.GetAccessToken(code)
	if err != nil {
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "获取access_token失败: "+err.Error())
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>获取access_token失败</p></div></body></html>`))
		return
	}

	// 获取用户信息
	userInfo, err := h.wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "获取用户信息失败: "+err.Error())
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>获取用户信息失败</p></div></body></html>`))
		return
	}

	// 开始事务
	tx := h.db.Begin()
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
			if ticket != "" {
				websocket.GetHub().SendMessage(ticket, "error", nil, "创建管理员角色失败")
			}
			c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>创建管理员角色失败</p></div></body></html>`))
			return
		}
	}

	// 2. 创建管理员用户
	adminUser := model.User{
		WeChatOpenID: userInfo.OpenID,
		Username:     userInfo.Nickname,
		Avatar:       userInfo.HeadImgURL,
		Status:       1,
	}
	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "创建管理员用户失败")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>创建管理员用户失败</p></div></body></html>`))
		return
	}

	// 3. 分配管理员角色
	if err := tx.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		tx.Rollback()
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "分配管理员角色失败")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>分配管理员角色失败</p></div></body></html>`))
		return
	}

	// 4. 标记系统已初始化
	initConfig := model.SystemConfig{
		Key:   "initialized",
		Value: "true",
		Type:  "boolean",
	}
	if err := tx.Create(&initConfig).Error; err != nil {
		tx.Rollback()
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "标记系统初始化失败")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>标记系统初始化失败</p></div></body></html>`))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "初始化失败")
		}
		c.Data(200, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>初始化失败</title><style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white}.container{text-align:center}</style></head><body><div class="container"><h1>✗ 初始化失败</h1><p>初始化失败</p></div></body></html>`))
		return
	}

	// 生成管理员Token
	roleNames := []string{"admin"}
	token, _ := auth.GenerateToken(adminUser.ID, adminUser.Username, roleNames)

	// 通过WebSocket通知PC端成功
	if ticket != "" {
		websocket.GetHub().SendMessage(ticket, "success", gin.H{
			"token": token,
			"user": gin.H{
				"id":       adminUser.ID,
				"username": adminUser.Username,
				"avatar":   adminUser.Avatar,
				"roles":    roleNames,
			},
		}, "系统初始化成功")
	}

	// 返回成功页面（简单的HTML）
	c.Data(200, "text/html; charset=utf-8", []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>初始化成功</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
					color: white;
				}
				.container {
					text-align: center;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>✓ 系统初始化成功！</h1>
				<p>请返回PC端查看结果</p>
			</div>
		</body>
		</html>
	`))
}

