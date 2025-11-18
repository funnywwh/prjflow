package api

import (
	"net/url"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/internal/websocket"
	"project-management/pkg/auth"
	"project-management/pkg/wechat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db           *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db:           db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// GetQRCode 获取微信登录二维码
func (h *AuthHandler) GetQRCode(c *gin.Context) {
	// 从数据库读取微信配置
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		// 如果数据库中没有配置，尝试使用配置文件中的配置
		if config.AppConfig.WeChat.AppID == "" {
			utils.Error(c, 400, "请先配置微信AppID和AppSecret")
			return
		}
		h.wechatClient.AppID = config.AppConfig.WeChat.AppID
		h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig)
		h.wechatClient.AppID = wechatAppIDConfig.Value
		h.wechatClient.AppSecret = wechatAppSecretConfig.Value
	}

	// 获取回调地址（前端地址）
	// 优先级：1. 配置文件中的 callback_domain 2. 查询参数 3. Referer 头 4. 默认值
	// 如果配置了 callback_domain，强制使用配置的域名，确保与微信后台配置一致
	var redirectURI string
	queryRedirectURI := c.Query("redirect_uri")

	if config.AppConfig.WeChat.CallbackDomain != "" {
		// 优先使用配置文件中的回调域名
		domain := config.AppConfig.WeChat.CallbackDomain
		if len(domain) > 0 && domain[len(domain)-1] != '/' {
			domain += "/"
		}

		// 如果查询参数中传递了 redirect_uri，提取路径部分，使用配置的域名
		// 这样可以支持不同的回调路径（如 /auth/wechat/add-user/callback）
		if queryRedirectURI != "" {
			// 解析查询参数中的 redirect_uri，提取路径部分
			var path string
			if len(queryRedirectURI) > 0 && queryRedirectURI[0] == '/' {
				// 已经是路径格式
				path = queryRedirectURI
			} else {
				// 是完整URL，解析提取路径部分
				parsedURL, err := url.Parse(queryRedirectURI)
				if err == nil && parsedURL.Path != "" {
					path = parsedURL.Path
				} else {
					// 解析失败，使用默认路径（指向后端接口）
					path = "/api/auth/wechat/callback"
				}
			}
			// 确保路径以 / 开头
			if len(path) > 0 && path[0] != '/' {
				path = "/" + path
			}
			// 去掉开头的 /，因为 domain 已经以 / 结尾
			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}
			redirectURI = domain + path
		} else {
			// 默认使用登录回调路径（指向后端接口，通过 /api 前缀）
			redirectURI = domain + "api/auth/wechat/callback"
		}
	} else {
		// 如果未配置 callback_domain，使用查询参数或 Referer
		redirectURI = queryRedirectURI
		if redirectURI == "" {
			// 从 Referer 头获取
			referer := c.GetHeader("Referer")
			if referer != "" {
				redirectURI = referer + "/api/auth/wechat/callback"
			} else {
				redirectURI = "http://localhost:8080/auth/wechat/callback"
			}
		}
	}

	// 先生成二维码获取ticket
	qrCode, err := h.wechatClient.GetQRCode(redirectURI)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取二维码失败: "+err.Error())
		return
	}

	// 生成唯一的ticket（使用时间戳，前端会通过WebSocket连接）
	ticket := qrCode.Ticket
	// 将ticket作为state参数的一部分，这样回调时可以获取到ticket
	// 注意：微信的state参数会原样返回，格式：ticket:timestamp
	stateWithTicket := "ticket:" + ticket

	// 重新生成二维码，将ticket包含在state中
	qrCode, err = h.wechatClient.GetQRCode(redirectURI, stateWithTicket)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取二维码失败: "+err.Error())
		return
	}

	// 返回授权URL，前端需要将其转换为二维码图片
	utils.Success(c, gin.H{
		"ticket":         ticket,
		"qr_code_url":    qrCode.URL, // 这是授权URL，需要转换为二维码
		"auth_url":       qrCode.URL, // 授权URL
		"expire_seconds": qrCode.ExpireSeconds,
	})
}

// LoginCallbackHandler 登录场景的微信回调处理
type LoginCallbackHandler struct {
	db *gorm.DB
}

func (h *LoginCallbackHandler) Validate(ctx *WeChatCallbackContext) error {
	// 登录场景无需特殊验证
	return nil
}

func (h *LoginCallbackHandler) Process(ctx *WeChatCallbackContext) (interface{}, error) {
	// 查找或创建用户
	var user model.User
	result := ctx.DB.Where("wechat_open_id = ?", ctx.UserInfo.OpenID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		user = model.User{
			WeChatOpenID: ctx.UserInfo.OpenID,
			Username:     ctx.UserInfo.Nickname,
			Avatar:       ctx.UserInfo.HeadImgURL,
			Status:       1,
		}
		if err := ctx.DB.Create(&user).Error; err != nil {
			return nil, &CallbackError{Message: "创建用户失败", Err: err}
		}
	} else if result.Error != nil {
		return nil, &CallbackError{Message: "查询用户失败", Err: result.Error}
	} else {
		// 更新用户信息
		user.Username = ctx.UserInfo.Nickname
		user.Avatar = ctx.UserInfo.HeadImgURL
		ctx.DB.Save(&user)
	}

	// 获取用户角色
	var roles []model.Role
	ctx.DB.Model(&user).Association("Roles").Find(&roles)

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Code)
	}

	// 生成JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, roleNames)
	if err != nil {
		return nil, &CallbackError{Message: "生成Token失败", Err: err}
	}

	// 通过WebSocket通知PC前端登录成功
	if ctx.Ticket != "" {
		websocket.GetHub().SendMessage(ctx.Ticket, "info", nil, "正在登录...")
		websocket.GetHub().SendMessage(ctx.Ticket, "success", gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"avatar":   user.Avatar,
				"roles":    roleNames,
			},
		}, "登录成功")
	}

	return gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"roles":    roleNames,
		},
	}, nil
}

func (h *LoginCallbackHandler) GetSuccessHTML(ctx *WeChatCallbackContext, data interface{}) string {
	return GetDefaultSuccessHTML("登录成功", "请返回 PC 端查看")
}

func (h *LoginCallbackHandler) GetErrorHTML(ctx *WeChatCallbackContext, err error) string {
	return GetDefaultErrorHTML("登录失败", err.Error())
}

// WeChatCallback 处理微信授权回调（GET请求，微信直接重定向到这里）
// 这个接口在微信内打开，处理完登录后通过WebSocket通知PC前端
func (h *AuthHandler) WeChatCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	handler := &LoginCallbackHandler{db: h.db}
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, code, state, handler)

	if err != nil {
		c.Data(200, "text/html; charset=utf-8", []byte(handler.GetErrorHTML(ctx, err)))
		return
	}

	// 返回成功页面（在微信内显示）
	c.Data(200, "text/html; charset=utf-8", []byte(handler.GetSuccessHTML(ctx, result)))
}

// WeChatLogin 微信登录（保留用于其他场景，如前端直接调用）
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 从数据库读取微信配置
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		// 如果数据库中没有配置，尝试使用配置文件中的配置
		if config.AppConfig.WeChat.AppID == "" {
			utils.Error(c, 400, "请先配置微信AppID和AppSecret")
			return
		}
		h.wechatClient.AppID = config.AppConfig.WeChat.AppID
		h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig)
		h.wechatClient.AppID = wechatAppIDConfig.Value
		h.wechatClient.AppSecret = wechatAppSecretConfig.Value
	}

	// 从state中提取ticket（如果存在）
	var ticket string
	if req.State != "" {
		if len(req.State) > 7 && req.State[:7] == "ticket:" {
			ticket = req.State[7:]
		} else {
			ticket = req.State
		}
	}

	// 如果存在ticket，通知已扫码
	if ticket != "" {
		websocket.GetHub().SendMessage(ticket, "info", nil, "已扫码，正在获取授权...")
	}

	// 获取access_token
	accessTokenResp, err := h.wechatClient.GetAccessToken(req.Code)
	if err != nil {
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "获取access_token失败")
		}
		utils.Error(c, utils.CodeError, "获取access_token失败")
		return
	}

	// 如果存在ticket，通知正在获取用户信息
	if ticket != "" {
		websocket.GetHub().SendMessage(ticket, "info", nil, "正在获取用户信息...")
	}

	// 获取用户信息
	userInfo, err := h.wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "获取用户信息失败")
		}
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
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "生成Token失败")
		}
		utils.Error(c, utils.CodeError, "生成Token失败")
		return
	}

	// 如果存在ticket，通过WebSocket通知登录页面
	if ticket != "" {
		websocket.GetHub().SendMessage(ticket, "info", nil, "正在登录...")
		websocket.GetHub().SendMessage(ticket, "success", gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"avatar":   user.Avatar,
				"roles":    roleNames,
			},
		}, "登录成功")
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
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"avatar":     user.Avatar,
		"phone":      user.Phone,
		"department": user.Department,
		"roles":      roleNames,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	utils.Success(c, gin.H{
		"message": "登出成功",
	})
}
