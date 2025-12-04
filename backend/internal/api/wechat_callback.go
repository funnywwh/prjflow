package api

import (
	"fmt"
	"strings"
	"time"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/websocket"
	"project-management/pkg/wechat"

	"gorm.io/gorm"
)

// WeChatCallbackContext 微信回调上下文
type WeChatCallbackContext struct {
	Code         string
	State        string
	Ticket       string
	WeChatClient wechat.WeChatClientInterface // 使用接口类型
	Hub          websocket.HubInterface        // 使用接口类型，用于发送WebSocket消息
	DB           *gorm.DB
	AccessToken  *wechat.AccessTokenResponse
	UserInfo     *wechat.UserInfoResponse
}

// WeChatCallbackHandler 微信回调业务处理接口
type WeChatCallbackHandler interface {
	// Validate 验证前置条件（如检查系统是否已初始化等）
	Validate(ctx *WeChatCallbackContext) error

	// Process 处理业务逻辑（获取用户信息后）
	Process(ctx *WeChatCallbackContext) (interface{}, error)

	// GetSuccessHTML 获取成功页面的HTML
	GetSuccessHTML(ctx *WeChatCallbackContext, data interface{}) string

	// GetErrorHTML 获取错误页面的HTML
	GetErrorHTML(ctx *WeChatCallbackContext, err error) string
}

// ProcessWeChatCallback 处理微信回调的通用流程
func ProcessWeChatCallback(
	db *gorm.DB,
	wechatClient wechat.WeChatClientInterface, // 使用接口类型
	hub websocket.HubInterface,                // 使用接口类型，用于发送WebSocket消息
	code string,
	state string,
	handler WeChatCallbackHandler,
) (*WeChatCallbackContext, interface{}, error) {
	ctx := &WeChatCallbackContext{
		Code:         code,
		State:        state,
		WeChatClient: wechatClient,
		Hub:          hub,
		DB:           db,
	}

	// 1. 从state中提取ticket
	if state != "" && len(state) > 7 && state[:7] == "ticket:" {
		ctx.Ticket = state[7:]
	} else if state != "" {
		ctx.Ticket = state
	}

	// 2. 检查code是否存在
	if code == "" {
		if ctx.Ticket != "" && ctx.Hub != nil {
			ctx.Hub.SendMessage(ctx.Ticket, "error", nil, "未获取到授权码")
		}
		return ctx, nil, &CallbackError{Message: "未获取到授权码"}
	}

	// 3. 读取微信配置
	var wechatAppIDConfig model.SystemConfig
	if err := db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		// 如果数据库中没有配置，尝试使用配置文件中的配置
		if config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
			if ctx.Ticket != "" && ctx.Hub != nil {
				ctx.Hub.SendMessage(ctx.Ticket, "error", nil, "请先配置微信AppID和AppSecret")
			}
			return ctx, nil, &CallbackError{Message: "请先配置微信AppID和AppSecret"}
		}
		wechatClient.SetAppID(config.AppConfig.WeChat.AppID)
		wechatClient.SetAppSecret(config.AppConfig.WeChat.AppSecret)
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		if err := db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
			// 如果数据库中没有AppSecret，尝试使用配置文件中的配置
			if config.AppConfig.WeChat.AppSecret == "" {
				if ctx.Ticket != "" && ctx.Hub != nil {
					ctx.Hub.SendMessage(ctx.Ticket, "error", nil, "请先配置微信AppSecret")
				}
				return ctx, nil, &CallbackError{Message: "请先配置微信AppSecret"}
			}
			wechatClient.SetAppID(wechatAppIDConfig.Value)
			wechatClient.SetAppSecret(config.AppConfig.WeChat.AppSecret)
		} else {
			// 从数据库读取配置，去除首尾空格
			wechatClient.SetAppID(strings.TrimSpace(wechatAppIDConfig.Value))
			wechatClient.SetAppSecret(strings.TrimSpace(wechatAppSecretConfig.Value))
		}
		// 验证配置是否为空
		if wechatClient.GetAppID() == "" || wechatClient.GetAppSecret() == "" {
			if ctx.Ticket != "" && ctx.Hub != nil {
				ctx.Hub.SendMessage(ctx.Ticket, "error", nil, "微信AppID或AppSecret配置为空，请检查配置")
			}
			return ctx, nil, &CallbackError{Message: "微信AppID或AppSecret配置为空，请检查配置"}
		}
	}

	// 设置AccountType和Scope（优先从数据库读取，其次从配置文件，最后使用默认值）
	var accountTypeConfig model.SystemConfig
	if err := db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error; err == nil {
		wechatClient.SetAccountType(strings.TrimSpace(accountTypeConfig.Value))
	} else {
		wechatClient.SetAccountType(config.AppConfig.WeChat.AccountType)
	}
	if wechatClient.GetAccountType() == "" {
		wechatClient.SetAccountType("open_platform") // 默认使用开放平台
	}

	var scopeConfig model.SystemConfig
	if err := db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error; err == nil {
		wechatClient.SetScope(strings.TrimSpace(scopeConfig.Value))
	} else {
		wechatClient.SetScope(config.AppConfig.WeChat.Scope)
	}
	if wechatClient.GetScope() == "" {
		wechatClient.SetScope("snsapi_userinfo") // 默认需要用户确认
	}

	// 4. 验证前置条件
	if err := handler.Validate(ctx); err != nil {
		if ctx.Ticket != "" && ctx.Hub != nil {
			ctx.Hub.SendMessage(ctx.Ticket, "error", nil, err.Error())
		}
		return ctx, nil, err
	}

	// 5. 通知已扫码
	if ctx.Ticket != "" && ctx.Hub != nil {
		ctx.Hub.SendMessage(ctx.Ticket, "info", nil, "已扫码，正在获取授权...")
	}

	// 6. 获取access_token
	// 添加调试信息：显示实际使用的配置（不显示完整的AppSecret，只显示前4位和后4位）
	appSecret := wechatClient.GetAppSecret()
	appSecretMasked := ""
	if len(appSecret) > 8 {
		appSecretMasked = appSecret[:4] + "****" + appSecret[len(appSecret)-4:]
	} else {
		appSecretMasked = "****"
	}
	debugInfo := fmt.Sprintf("使用配置: AppID=%s, AppSecret=%s, AccountType=%s, Scope=%s",
		wechatClient.GetAppID(), appSecretMasked, wechatClient.GetAccountType(), wechatClient.GetScope())

	accessTokenResp, err := wechatClient.GetAccessToken(code)
	if err != nil {
		errorMsg := fmt.Sprintf("获取access_token失败: %s。%s", err.Error(), debugInfo)
		if ctx.Ticket != "" && ctx.Hub != nil {
			ctx.Hub.SendMessage(ctx.Ticket, "error", nil, errorMsg)
		}
		return ctx, nil, &CallbackError{Message: errorMsg, Err: err}
	}
	ctx.AccessToken = accessTokenResp

	// 7. 通知正在获取用户信息
	if ctx.Ticket != "" && ctx.Hub != nil {
		ctx.Hub.SendMessage(ctx.Ticket, "info", nil, "正在获取用户信息...")
	}

	// 8. 获取用户信息
	userInfo, err := wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		if ctx.Ticket != "" && ctx.Hub != nil {
			ctx.Hub.SendMessage(ctx.Ticket, "error", nil, "获取用户信息失败")
		}
		return ctx, nil, &CallbackError{Message: "获取用户信息失败", Err: err}
	}
	ctx.UserInfo = userInfo

	// 9. 处理业务逻辑
	result, err := handler.Process(ctx)
	if err != nil {
		if ctx.Ticket != "" && ctx.Hub != nil {
			ctx.Hub.SendMessage(ctx.Ticket, "error", nil, err.Error())
		}
		return ctx, nil, err
	}

	return ctx, result, nil
}

// CallbackError 回调错误
type CallbackError struct {
	Message string
	Err     error
}

func (e *CallbackError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// GetDefaultErrorHTML 获取默认错误页面HTML
func GetDefaultErrorHTML(title, message string) string {
	return `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title>
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
			padding: 20px;
		}
		h1 { font-size: 48px; margin: 0 0 20px 0; }
		p { font-size: 18px; margin: 0; }
	</style>
</head>
<body>
	<div class="container">
		<h1>✗</h1>
		<p>` + title + `</p>
		<p style="font-size: 14px; margin-top: 10px;">` + message + `</p>
	</div>
</body>
</html>`
}

// GenerateUniqueUsername 生成唯一的用户名
// 如果基础用户名已存在，自动添加数字后缀（如：用户名_1, 用户名_2）
// 如果基础用户名为空，使用OpenID的一部分生成
// 为了提高唯一性，默认使用时间戳作为后缀的一部分
func GenerateUniqueUsername(db *gorm.DB, baseUsername string, openID string) string {
	username := baseUsername
	if username == "" {
		// 如果用户名为空，使用OpenID的一部分 + 时间戳确保唯一性
		openIDSuffix := openID
		if len(openID) > 8 {
			openIDSuffix = openID[len(openID)-8:]
		}
		// 使用时间戳的后6位 + OpenID后缀，提高唯一性
		timestampSuffix := time.Now().Unix() % 1000000 // 取后6位
		username = fmt.Sprintf("用户_%d_%s", timestampSuffix, openIDSuffix)
	}

	// 检查用户名是否已存在，如果存在则添加数字后缀（包括软删除的记录）
	originalUsername := username
	suffix := 1
	maxAttempts := 100
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var checkUser model.User
		// 使用 Unscoped() 检查包括软删除的记录，因为唯一索引仍然存在
		if err := db.Unscoped().Where("username = ?", username).First(&checkUser).Error; err == gorm.ErrRecordNotFound {
			// 用户名可用
			return username
		} else if err != nil && err != gorm.ErrRecordNotFound {
			// 查询出错，使用时间戳作为后缀
			username = fmt.Sprintf("%s_%d", originalUsername, time.Now().Unix())
			return username
		}
		// 用户名已存在（包括软删除的），添加数字后缀
		suffix++
		username = fmt.Sprintf("%s_%d", originalUsername, suffix)
	}

	// 如果尝试100次后仍然冲突，使用时间戳 + 随机数
	username = fmt.Sprintf("%s_%d_%d", originalUsername, time.Now().Unix(), suffix)
	return username
}

// GetDefaultSuccessHTML 获取默认成功页面HTML
func GetDefaultSuccessHTML(title, message string) string {
	return `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title>
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
			padding: 20px;
		}
		h1 { font-size: 48px; margin: 0 0 20px 0; }
		p { font-size: 18px; margin: 0; }
	</style>
</head>
<body>
	<div class="container">
		<h1>✓</h1>
		<p>` + title + `</p>
		<p style="font-size: 14px; margin-top: 10px;">` + message + `</p>
	</div>
</body>
</html>`
}
