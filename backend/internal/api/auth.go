package api

import (
	"fmt"
	"net/url"
	"strings"

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
		if config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
			utils.Error(c, 400, "请先配置微信AppID和AppSecret")
			return
		}
		h.wechatClient.AppID = config.AppConfig.WeChat.AppID
		h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
			// 如果数据库中没有AppSecret，尝试使用配置文件中的配置
			if config.AppConfig.WeChat.AppSecret == "" {
				utils.Error(c, 400, "请先配置微信AppSecret")
				return
			}
			h.wechatClient.AppID = wechatAppIDConfig.Value
			h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
		} else {
			// 从数据库读取配置，去除首尾空格
			h.wechatClient.AppID = strings.TrimSpace(wechatAppIDConfig.Value)
			h.wechatClient.AppSecret = strings.TrimSpace(wechatAppSecretConfig.Value)
		}
		// 验证配置是否为空
		if h.wechatClient.AppID == "" || h.wechatClient.AppSecret == "" {
			utils.Error(c, 400, "微信AppID或AppSecret配置为空，请检查配置")
			return
		}
	}

	// 设置AccountType和Scope（优先从数据库读取，其次从配置文件，最后使用默认值）
	var accountTypeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error; err == nil {
		h.wechatClient.AccountType = strings.TrimSpace(accountTypeConfig.Value)
	} else {
		h.wechatClient.AccountType = config.AppConfig.WeChat.AccountType
	}
	if h.wechatClient.AccountType == "" {
		h.wechatClient.AccountType = "open_platform" // 默认使用开放平台
	}

	var scopeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error; err == nil {
		h.wechatClient.Scope = strings.TrimSpace(scopeConfig.Value)
	} else {
		h.wechatClient.Scope = config.AppConfig.WeChat.Scope
	}
	if h.wechatClient.Scope == "" {
		h.wechatClient.Scope = "snsapi_userinfo" // 默认需要用户确认
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
		// 生成唯一的用户名（如果昵称已存在，自动添加数字后缀）
		username := GenerateUniqueUsername(ctx.DB, ctx.UserInfo.Nickname, ctx.UserInfo.OpenID)

		// 确保昵称不为空（如果微信昵称为空，使用用户名作为默认昵称）
		nickname := ctx.UserInfo.Nickname
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
		if err := ctx.DB.Create(&user).Error; err != nil {
			return nil, &CallbackError{Message: "创建用户失败", Err: err}
		}
	} else if result.Error != nil {
		return nil, &CallbackError{Message: "查询用户失败", Err: result.Error}
	} else {
		// 更新用户信息（如果用户名变化，需要检查是否重复）
		// 注意：这里不更新用户名，因为用户名可能已被用户修改过
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

	// 更新登录次数
	if err := ctx.DB.Model(&user).Update("login_count", gorm.Expr("login_count + 1")).Error; err != nil {
		return nil, &CallbackError{Message: "更新登录次数失败", Err: err}
	}

	// 重新查询用户获取更新后的登录次数
	if err := ctx.DB.First(&user, user.ID).Error; err != nil {
		return nil, &CallbackError{Message: "查询用户失败", Err: err}
	}

	// 微信登录不需要密码，首次登录也不需要强制修改密码
	// 只有用户名密码登录的首次登录才需要修改密码
	isFirstLogin := false

	// 生成JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, roleNames)
	if err != nil {
		return nil, &CallbackError{Message: "生成Token失败", Err: err}
	}

	// 通过WebSocket通知PC前端登录成功
	if ctx.Ticket != "" && ctx.Hub != nil {
		ctx.Hub.SendMessage(ctx.Ticket, "info", nil, "正在登录...")
		ctx.Hub.SendMessage(ctx.Ticket, "success", gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"avatar":   user.Avatar,
				"roles":    roleNames,
			},
			"is_first_login": isFirstLogin,
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
		"is_first_login": isFirstLogin,
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
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, websocket.GetHub(), code, state, handler)

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
		if config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
			utils.Error(c, 400, "请先配置微信AppID和AppSecret")
			return
		}
		h.wechatClient.AppID = config.AppConfig.WeChat.AppID
		h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
			// 如果数据库中没有AppSecret，尝试使用配置文件中的配置
			if config.AppConfig.WeChat.AppSecret == "" {
				utils.Error(c, 400, "请先配置微信AppSecret")
				return
			}
			h.wechatClient.AppID = wechatAppIDConfig.Value
			h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
		} else {
			// 从数据库读取配置，去除首尾空格
			h.wechatClient.AppID = strings.TrimSpace(wechatAppIDConfig.Value)
			h.wechatClient.AppSecret = strings.TrimSpace(wechatAppSecretConfig.Value)
		}
		// 验证配置是否为空
		if h.wechatClient.AppID == "" || h.wechatClient.AppSecret == "" {
			utils.Error(c, 400, "微信AppID或AppSecret配置为空，请检查配置")
			return
		}
	}

	// 设置AccountType和Scope（优先从数据库读取，其次从配置文件，最后使用默认值）
	var accountTypeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error; err == nil {
		h.wechatClient.AccountType = strings.TrimSpace(accountTypeConfig.Value)
	} else {
		h.wechatClient.AccountType = config.AppConfig.WeChat.AccountType
	}
	if h.wechatClient.AccountType == "" {
		h.wechatClient.AccountType = "open_platform" // 默认使用开放平台
	}

	var scopeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error; err == nil {
		h.wechatClient.Scope = strings.TrimSpace(scopeConfig.Value)
	} else {
		h.wechatClient.Scope = config.AppConfig.WeChat.Scope
	}
	if h.wechatClient.Scope == "" {
		h.wechatClient.Scope = "snsapi_userinfo" // 默认需要用户确认
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
		errorMsg := "获取access_token失败: " + err.Error()
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, errorMsg)
		}
		utils.Error(c, utils.CodeError, errorMsg)
		return
	}

	// 如果存在ticket，通知正在获取用户信息
	if ticket != "" {
		websocket.GetHub().SendMessage(ticket, "info", nil, "正在获取用户信息...")
	}

	// 获取用户信息
	userInfo, err := h.wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		errorMsg := "获取用户信息失败: " + err.Error()
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, errorMsg)
		}
		utils.Error(c, utils.CodeError, errorMsg)
		return
	}

	// 查找或创建用户
	var user model.User
	result := h.db.Where("wechat_open_id = ?", userInfo.OpenID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		// 生成唯一的用户名（如果昵称已存在，自动添加数字后缀）
		username := GenerateUniqueUsername(h.db, userInfo.Nickname, userInfo.OpenID)

		// 确保昵称不为空（如果微信昵称为空，使用用户名作为默认昵称）
		nickname := userInfo.Nickname
		if nickname == "" {
			nickname = username
		}

		// 创建新用户
		wechatOpenID := userInfo.OpenID
		user = model.User{
			WeChatOpenID: &wechatOpenID,
			Username:     username,
			Nickname:     nickname, // 设置昵称（从微信昵称获取，如果为空则使用用户名）
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
		// 更新用户信息（不更新用户名，因为用户名可能已被用户修改过）
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

	// 更新登录次数
	if err := h.db.Model(&user).Update("login_count", gorm.Expr("login_count + 1")).Error; err != nil {
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "更新登录次数失败")
		}
		utils.Error(c, utils.CodeError, "更新登录次数失败")
		return
	}

	// 重新查询用户获取更新后的登录次数
	if err := h.db.First(&user, user.ID).Error; err != nil {
		// 如果存在ticket，通知错误
		if ticket != "" {
			websocket.GetHub().SendMessage(ticket, "error", nil, "查询用户失败")
		}
		utils.Error(c, utils.CodeError, "查询用户失败")
		return
	}

	// 微信登录不需要密码，首次登录也不需要强制修改密码
	// 只有用户名密码登录的首次登录才需要修改密码
	isFirstLogin := false

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
			"is_first_login": isFirstLogin,
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
		"is_first_login": isFirstLogin,
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

	// 判断是否是首次登录：LoginCount == 1 且 用户有密码（只有用户名密码登录的首次登录才需要修改密码）
	// 微信用户首次登录时LoginCount==1，但没有密码，不应该强制修改密码
	isFirstLogin := user.LoginCount == 1 && user.Password != ""

	utils.Success(c, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"nickname":       user.Nickname,
		"email":          user.Email,
		"avatar":         user.Avatar,
		"phone":          user.Phone,
		"wechat_open_id": user.WeChatOpenID,
		"department":     user.Department,
		"roles":          roleNames,
		"is_first_login": isFirstLogin,
	})
}

// Login 用户名密码登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "用户名和密码不能为空")
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, 401, "用户名或密码错误")
		} else {
			utils.Error(c, utils.CodeError, "查询用户失败")
		}
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		utils.Error(c, 403, "用户已被禁用")
		return
	}

	// 验证密码
	if user.Password == "" || !utils.CheckPassword(req.Password, user.Password) {
		utils.Error(c, 401, "用户名或密码错误")
		return
	}

	// 获取用户角色
	var roles []model.Role
	h.db.Model(&user).Association("Roles").Find(&roles)

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Code)
	}

	// 更新登录次数
	if err := h.db.Model(&user).Update("login_count", gorm.Expr("login_count + 1")).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新登录次数失败")
		return
	}

	// 重新查询用户获取更新后的登录次数
	if err := h.db.First(&user, user.ID).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询用户失败")
		return
	}

	// 判断是否是首次登录（更新后LoginCount == 1）
	isFirstLogin := user.LoginCount == 1

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
		"is_first_login": isFirstLogin,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"` // 可选，如果用户没有密码则不需要
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	// 检查用户是否已有密码
	hasPassword := user.Password != ""

	// 如果用户已有密码，需要验证旧密码
	// 如果用户没有密码（微信登录用户），则可以直接设置新密码
	if hasPassword {
		if req.OldPassword == "" {
			utils.Error(c, 400, "请输入旧密码")
			return
		}
		if !utils.CheckPassword(req.OldPassword, user.Password) {
			utils.Error(c, 400, "旧密码错误")
			return
		}
	}

	// 验证新密码强度：必须包含大小写字母和数字
	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.Error(c, utils.CodeError, "加密密码失败")
		return
	}

	// 更新密码
	user.Password = hashedPassword
	// 如果用户是首次登录（LoginCount == 1），修改密码后将LoginCount更新为2
	// 这样下次调用GetUserInfo时就不会返回is_first_login=true了
	if user.LoginCount == 1 {
		user.LoginCount = 2
	}
	if err := h.db.Save(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新密码失败")
		return
	}

	// 根据是否有旧密码返回不同的消息
	message := "密码修改成功"
	if !hasPassword {
		message = "密码设置成功，现在可以使用用户名密码登录了"
	}

	utils.Success(c, gin.H{
		"message": message,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除token即可
	utils.Success(c, gin.H{
		"message": "登出成功",
	})
}

// GetWeChatBindQRCode 获取微信绑定二维码
func (h *AuthHandler) GetWeChatBindQRCode(c *gin.Context) {
	// 检查用户是否已登录
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权，请先登录")
		return
	}

	// 检查用户是否已绑定微信
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	if user.WeChatOpenID != nil && *user.WeChatOpenID != "" {
		utils.Error(c, 400, "您已绑定微信，请先解绑后再绑定新的微信")
		return
	}

	// 从数据库读取微信配置
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		// 如果数据库中没有配置，尝试使用配置文件中的配置
		if config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
			utils.Error(c, 400, "请先配置微信AppID和AppSecret")
			return
		}
		h.wechatClient.AppID = config.AppConfig.WeChat.AppID
		h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
	} else {
		// 从数据库读取配置
		var wechatAppSecretConfig model.SystemConfig
		if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
			// 如果数据库中没有AppSecret，尝试使用配置文件中的配置
			if config.AppConfig.WeChat.AppSecret == "" {
				utils.Error(c, 400, "请先配置微信AppSecret")
				return
			}
			h.wechatClient.AppID = wechatAppIDConfig.Value
			h.wechatClient.AppSecret = config.AppConfig.WeChat.AppSecret
		} else {
			// 从数据库读取配置，去除首尾空格
			h.wechatClient.AppID = strings.TrimSpace(wechatAppIDConfig.Value)
			h.wechatClient.AppSecret = strings.TrimSpace(wechatAppSecretConfig.Value)
		}
		// 验证配置是否为空
		if h.wechatClient.AppID == "" || h.wechatClient.AppSecret == "" {
			utils.Error(c, 400, "微信AppID或AppSecret配置为空，请检查配置")
			return
		}
	}

	// 设置AccountType和Scope（优先从数据库读取，其次从配置文件，最后使用默认值）
	var accountTypeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error; err == nil {
		h.wechatClient.AccountType = strings.TrimSpace(accountTypeConfig.Value)
	} else {
		h.wechatClient.AccountType = config.AppConfig.WeChat.AccountType
	}
	if h.wechatClient.AccountType == "" {
		h.wechatClient.AccountType = "open_platform" // 默认使用开放平台
	}

	var scopeConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error; err == nil {
		h.wechatClient.Scope = strings.TrimSpace(scopeConfig.Value)
	} else {
		h.wechatClient.Scope = config.AppConfig.WeChat.Scope
	}
	if h.wechatClient.Scope == "" {
		h.wechatClient.Scope = "snsapi_userinfo" // 默认需要用户确认
	}

	// 获取回调地址（指向绑定回调接口）
	var redirectURI string
	if config.AppConfig.WeChat.CallbackDomain != "" {
		domain := config.AppConfig.WeChat.CallbackDomain
		if len(domain) > 0 && domain[len(domain)-1] != '/' {
			domain += "/"
		}
		redirectURI = domain + "api/auth/wechat/bind/callback"
	} else {
		// 从 Referer 头获取
		referer := c.GetHeader("Referer")
		if referer != "" {
			redirectURI = referer + "/api/auth/wechat/bind/callback"
		} else {
			redirectURI = "http://localhost:8080/api/auth/wechat/bind/callback"
		}
	}

	// 生成二维码获取ticket
	qrCode, err := h.wechatClient.GetQRCode(redirectURI)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取二维码失败: "+err.Error())
		return
	}

	// 生成唯一的ticket
	ticket := qrCode.Ticket
	// state格式：bind:{ticket}:{user_id}
	var userIDStr string
	if userIDInt, ok := userID.(int); ok {
		userIDStr = fmt.Sprintf("%d", userIDInt)
	} else if userIDUint, ok := userID.(uint); ok {
		userIDStr = fmt.Sprintf("%d", userIDUint)
	} else {
		utils.Error(c, utils.CodeError, "无效的用户ID")
		return
	}
	stateWithTicket := "bind:" + ticket + ":" + userIDStr

	// 重新生成二维码，将ticket和user_id包含在state中
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

// BindCallbackHandler 绑定场景的微信回调处理
type BindCallbackHandler struct {
	db *gorm.DB
}

func (h *BindCallbackHandler) Validate(ctx *WeChatCallbackContext) error {
	// 从state中提取user_id
	// state格式：bind:{ticket}:{user_id}
	if ctx.State == "" {
		return &CallbackError{Message: "缺少state参数"}
	}

	// 解析state：bind:{ticket}:{user_id}
	var userIDStr string
	if len(ctx.State) > 5 && ctx.State[:5] == "bind:" {
		// 提取ticket和user_id
		parts := ctx.State[5:] // 去掉"bind:"前缀
		// 找到最后一个冒号，后面是user_id
		lastColonIndex := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == ':' {
				lastColonIndex = i
				break
			}
		}
		if lastColonIndex > 0 {
			userIDStr = parts[lastColonIndex+1:]
			ctx.Ticket = parts[:lastColonIndex]
		} else {
			return &CallbackError{Message: "state参数格式错误"}
		}
	} else {
		return &CallbackError{Message: "state参数格式错误"}
	}

	// 验证用户是否存在
	var user model.User
	if err := ctx.DB.First(&user, userIDStr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &CallbackError{Message: "用户不存在"}
		}
		return &CallbackError{Message: "查询用户失败", Err: err}
	}

	// 检查用户是否已绑定其他微信
	if user.WeChatOpenID != nil && *user.WeChatOpenID != "" {
		return &CallbackError{Message: "您已绑定微信，请先解绑后再绑定新的微信"}
	}

	return nil
}

func (h *BindCallbackHandler) Process(ctx *WeChatCallbackContext) (interface{}, error) {
	// 从state中提取user_id
	var userIDStr string
	if len(ctx.State) > 5 && ctx.State[:5] == "bind:" {
		parts := ctx.State[5:]
		lastColonIndex := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == ':' {
				lastColonIndex = i
				break
			}
		}
		if lastColonIndex > 0 {
			userIDStr = parts[lastColonIndex+1:]
		}
	}

	// 查找用户
	var user model.User
	if err := ctx.DB.First(&user, userIDStr).Error; err != nil {
		return nil, &CallbackError{Message: "用户不存在", Err: err}
	}

	// 检查该微信OpenID是否已被其他用户绑定（包括软删除的用户）
	var existingUser model.User
	if err := ctx.DB.Unscoped().Where("wechat_open_id = ? AND id != ?", ctx.UserInfo.OpenID, user.ID).First(&existingUser).Error; err == nil {
		// 如果找到的用户是软删除的，先清理它的 wechat_open_id
		if existingUser.DeletedAt.Valid {
			// 软删除的用户，清理其 wechat_open_id
			if err := ctx.DB.Unscoped().Model(&existingUser).Update("wechat_open_id", nil).Error; err != nil {
				return nil, &CallbackError{Message: "清理已删除用户的微信绑定失败", Err: err}
			}
		} else {
			// 正常用户，不能重复绑定
			return nil, &CallbackError{Message: fmt.Sprintf("该微信已被用户 %s 绑定，无法重复绑定", existingUser.Username)}
		}
	} else if err != gorm.ErrRecordNotFound {
		return nil, &CallbackError{Message: "查询用户失败", Err: err}
	}

	// 再次检查该微信OpenID是否已被其他用户绑定（防止并发问题，包括软删除的用户）
	var checkUser model.User
	if err := ctx.DB.Unscoped().Where("wechat_open_id = ? AND id != ?", ctx.UserInfo.OpenID, user.ID).First(&checkUser).Error; err == nil {
		// 如果找到的用户是软删除的，先清理它的 wechat_open_id
		if checkUser.DeletedAt.Valid {
			// 软删除的用户，清理其 wechat_open_id
			if err := ctx.DB.Unscoped().Model(&checkUser).Update("wechat_open_id", nil).Error; err != nil {
				return nil, &CallbackError{Message: "清理已删除用户的微信绑定失败", Err: err}
			}
		} else {
			// 正常用户，不能重复绑定
			return nil, &CallbackError{Message: fmt.Sprintf("该微信已被用户 %s 绑定，无法重复绑定", checkUser.Username)}
		}
	} else if err != gorm.ErrRecordNotFound {
		return nil, &CallbackError{Message: "查询用户失败", Err: err}
	}

	// 更新用户的wechat_open_id
	wechatOpenID := ctx.UserInfo.OpenID
	user.WeChatOpenID = &wechatOpenID
	// 可选：更新头像
	if ctx.UserInfo.HeadImgURL != "" {
		user.Avatar = ctx.UserInfo.HeadImgURL
	}
	if err := ctx.DB.Save(&user).Error; err != nil {
		// 检查是否是唯一约束错误
		if utils.IsUniqueConstraintOnField(err, "wechat_open_id") {
			// 再次查询，找出是哪个用户绑定了这个微信（包括软删除的用户）
			var conflictUser model.User
			if queryErr := ctx.DB.Unscoped().Where("wechat_open_id = ?", ctx.UserInfo.OpenID).First(&conflictUser).Error; queryErr == nil {
				// 如果是软删除的用户，清理它的 wechat_open_id 并重试
				if conflictUser.DeletedAt.Valid {
					// 清理软删除用户的 wechat_open_id
					if updateErr := ctx.DB.Unscoped().Model(&conflictUser).Update("wechat_open_id", nil).Error; updateErr != nil {
						return nil, &CallbackError{Message: "清理已删除用户的微信绑定失败", Err: updateErr}
					}
					// 重试保存
					if retryErr := ctx.DB.Save(&user).Error; retryErr != nil {
						return nil, &CallbackError{Message: "绑定失败", Err: retryErr}
					}
					// 保存成功，继续后续流程
				} else {
					// 正常用户，不能重复绑定
					return nil, &CallbackError{Message: fmt.Sprintf("该微信已被用户 %s 绑定，无法重复绑定", conflictUser.Username)}
				}
			} else {
				return nil, &CallbackError{Message: "该微信已被其他用户绑定，无法重复绑定"}
			}
		} else {
			// 不是唯一约束错误，返回原始错误
			return nil, &CallbackError{Message: "绑定失败", Err: err}
		}
	}

	// 通过WebSocket通知PC前端绑定成功
	if ctx.Ticket != "" && ctx.Hub != nil {
		ctx.Hub.SendMessage(ctx.Ticket, "info", nil, "正在绑定...")
		ctx.Hub.SendMessage(ctx.Ticket, "success", gin.H{
			"message": "绑定成功",
		}, "微信绑定成功")
	}

	return gin.H{
		"message": "绑定成功",
		"user": gin.H{
			"id":             user.ID,
			"username":       user.Username,
			"wechat_open_id": user.WeChatOpenID,
		},
	}, nil
}

func (h *BindCallbackHandler) GetSuccessHTML(ctx *WeChatCallbackContext, data interface{}) string {
	return GetDefaultSuccessHTML("绑定成功", "请返回 PC 端查看")
}

func (h *BindCallbackHandler) GetErrorHTML(ctx *WeChatCallbackContext, err error) string {
	return GetDefaultErrorHTML("绑定失败", err.Error())
}

// WeChatBindCallback 处理微信绑定回调（GET请求，微信直接重定向到这里）
func (h *AuthHandler) WeChatBindCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	handler := &BindCallbackHandler{db: h.db}
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, websocket.GetHub(), code, state, handler)

	if err != nil {
		c.Data(200, "text/html; charset=utf-8", []byte(handler.GetErrorHTML(ctx, err)))
		return
	}

	// 返回成功页面（在微信内显示）
	c.Data(200, "text/html; charset=utf-8", []byte(handler.GetSuccessHTML(ctx, result)))
}

// UnbindWeChat 解绑微信
func (h *AuthHandler) UnbindWeChat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权，请先登录")
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	// 检查用户是否已绑定微信
	if user.WeChatOpenID == nil || *user.WeChatOpenID == "" {
		utils.Error(c, 400, "您尚未绑定微信")
		return
	}

	// 解绑：将wechat_open_id设置为NULL
	user.WeChatOpenID = nil
	if err := h.db.Save(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "解绑失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "解绑成功",
	})
}
