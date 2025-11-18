package api

import (
	"gorm.io/gorm"
	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/websocket"
	"project-management/pkg/wechat"
)

// WeChatCallbackContext 微信回调上下文
type WeChatCallbackContext struct {
	Code          string
	State         string
	Ticket        string
	WeChatClient  *wechat.WeChatClient
	DB            *gorm.DB
	AccessToken   *wechat.AccessTokenResponse
	UserInfo      *wechat.UserInfoResponse
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
	wechatClient *wechat.WeChatClient,
	code string,
	state string,
	handler WeChatCallbackHandler,
) (*WeChatCallbackContext, interface{}, error) {
	ctx := &WeChatCallbackContext{
		Code:         code,
		State:        state,
		WeChatClient: wechatClient,
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
		if ctx.Ticket != "" {
			websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, "未获取到授权码")
		}
		return ctx, nil, &CallbackError{Message: "未获取到授权码"}
	}
	
	// 3. 读取微信配置
	var wechatAppIDConfig model.SystemConfig
	if err := db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		// 如果数据库中没有配置，尝试使用配置文件中的配置
		if config.AppConfig.WeChat.AppID == "" {
			if ctx.Ticket != "" {
				websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, "请先配置微信AppID和AppSecret")
			}
			return ctx, nil, &CallbackError{Message: "请先配置微信AppID和AppSecret"}
		}
		wechatClient.AppID = config.AppConfig.WeChat.AppID
		wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig)
		wechatClient.AppID = wechatAppIDConfig.Value
		wechatClient.AppSecret = wechatAppSecretConfig.Value
	}
	
	// 4. 验证前置条件
	if err := handler.Validate(ctx); err != nil {
		if ctx.Ticket != "" {
			websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, err.Error())
		}
		return ctx, nil, err
	}
	
	// 5. 通知已扫码
	if ctx.Ticket != "" {
		websocket.GetHub().SendMessage(ctx.Ticket, "info", nil, "已扫码，正在获取授权...")
	}
	
	// 6. 获取access_token
	accessTokenResp, err := wechatClient.GetAccessToken(code)
	if err != nil {
		if ctx.Ticket != "" {
			websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, "获取access_token失败")
		}
		return ctx, nil, &CallbackError{Message: "获取access_token失败", Err: err}
	}
	ctx.AccessToken = accessTokenResp
	
	// 7. 通知正在获取用户信息
	if ctx.Ticket != "" {
		websocket.GetHub().SendMessage(ctx.Ticket, "info", nil, "正在获取用户信息...")
	}
	
	// 8. 获取用户信息
	userInfo, err := wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		if ctx.Ticket != "" {
			websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, "获取用户信息失败")
		}
		return ctx, nil, &CallbackError{Message: "获取用户信息失败", Err: err}
	}
	ctx.UserInfo = userInfo
	
	// 9. 处理业务逻辑
	result, err := handler.Process(ctx)
	if err != nil {
		if ctx.Ticket != "" {
			websocket.GetHub().SendMessage(ctx.Ticket, "error", nil, err.Error())
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

