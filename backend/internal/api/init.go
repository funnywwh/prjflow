package api

import (
	"strings"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/auth"
	"project-management/pkg/wechat"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InitHandler struct {
	db           *gorm.DB
	wechatClient wechat.WeChatClientInterface // 使用接口类型，支持依赖注入
}

func NewInitHandler(db *gorm.DB) *InitHandler {
	return &InitHandler{
		db:           db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// SetWeChatClient 设置WeChatClient（用于测试）
func (h *InitHandler) SetWeChatClient(client wechat.WeChatClientInterface) {
	h.wechatClient = client
}

// CheckInitStatus 检查初始化状态
func (h *InitHandler) CheckInitStatus(c *gin.Context) {
	var config model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&config)

	if result.Error == gorm.ErrRecordNotFound {
		utils.Success(c, gin.H{
			"initialized": false,
		})
		return
	}

	utils.Success(c, gin.H{
		"initialized": config.Value == "true",
	})
}

// SaveWeChatConfig 保存微信配置（第一步）
func (h *InitHandler) SaveWeChatConfig(c *gin.Context) {
	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		utils.Error(c, 400, "系统已经初始化，无法重复初始化")
		return
	}

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

	// 更新WeChatClient的配置（临时，用于后续获取二维码）
	h.wechatClient.SetAppID(req.WeChatAppID)
	h.wechatClient.SetAppSecret(req.WeChatAppSecret)

	utils.Success(c, gin.H{
		"message": "微信配置保存成功",
	})
}

// InitSystem 完成初始化（第二步：通过微信登录创建管理员）
func (h *InitHandler) InitSystem(c *gin.Context) {
	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		utils.Error(c, 400, "系统已经初始化，无法重复初始化")
		return
	}

	// 检查微信配置是否已保存
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		utils.Error(c, 400, "请先配置微信AppID和AppSecret")
		return
	}

	var req struct {
		Code  string `json:"code" binding:"required"` // 微信登录返回的code
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 获取微信配置
	var wechatAppSecretConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
		utils.Error(c, 400, "微信AppSecret配置不存在，请先保存微信配置")
		return
	}

	// 临时设置WeChatClient配置，去除首尾空格
	appID := strings.TrimSpace(wechatAppIDConfig.Value)
	appSecret := strings.TrimSpace(wechatAppSecretConfig.Value)
	h.wechatClient.SetAppID(appID)
	h.wechatClient.SetAppSecret(appSecret)
	
	// 验证配置是否为空
	if h.wechatClient.GetAppID() == "" || h.wechatClient.GetAppSecret() == "" {
		utils.Error(c, 400, "微信AppID或AppSecret配置为空，请检查配置")
		return
	}
	
	// 设置AccountType和Scope（优先从数据库读取，其次从配置文件，最后使用默认值）
	var accountTypeConfig model.SystemConfig
	var accountType string
	if err := h.db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error; err == nil {
		accountType = strings.TrimSpace(accountTypeConfig.Value)
	} else {
		accountType = config.AppConfig.WeChat.AccountType
	}
	if accountType == "" {
		accountType = "open_platform" // 默认使用开放平台
	}
	h.wechatClient.SetAccountType(accountType)
	
	var scopeConfig model.SystemConfig
	var scope string
	if err := h.db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error; err == nil {
		scope = strings.TrimSpace(scopeConfig.Value)
	} else {
		scope = config.AppConfig.WeChat.Scope
	}
	if scope == "" {
		scope = "snsapi_userinfo" // 默认需要用户确认
	}
	h.wechatClient.SetScope(scope)

	// 获取access_token
	accessTokenResp, err := h.wechatClient.GetAccessToken(req.Code)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取access_token失败: "+err.Error())
		return
	}

	// 获取用户信息
	userInfo, err := h.wechatClient.GetUserInfo(accessTokenResp.AccessToken, accessTokenResp.OpenID)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取用户信息失败: "+err.Error())
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
			utils.Error(c, utils.CodeError, "创建管理员角色失败")
			return
		}
	}

	// 2. 创建管理员用户（使用微信登录获取的用户信息）
	// 使用微信昵称作为用户名，如果为空则使用默认值
	adminUsername := userInfo.Nickname
	if adminUsername == "" {
		adminUsername = "管理员"
	}

	// 确保昵称不为空（如果微信昵称为空，使用用户名作为默认昵称）
	adminNickname := userInfo.Nickname
	if adminNickname == "" {
		adminNickname = adminUsername
	}

	wechatOpenID := userInfo.OpenID
	adminUser := model.User{
		WeChatOpenID: &wechatOpenID,
		Username:     adminUsername,
		Nickname:     adminNickname, // 设置昵称（从微信昵称获取，如果为空则使用用户名）
		Avatar:       userInfo.HeadImgURL,
		Status:       1,
	}
	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "创建管理员用户失败")
		return
	}

	// 3. 分配管理员角色
	if err := tx.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "分配管理员角色失败")
		return
	}

	// 4. 标记系统已初始化
	initConfig := model.SystemConfig{
		Key:   "initialized",
		Value: "true",
		Type:  "boolean",
	}
	if err := tx.Where("key = ?", "initialized").Assign(model.SystemConfig{Value: "true", Type: "boolean"}).FirstOrCreate(&initConfig).Error; err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "保存初始化状态失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.Error(c, utils.CodeError, "初始化失败")
		return
	}

	// 生成管理员Token（可选，用于自动登录）
	roleNames := []string{"admin"}
	token, _ := auth.GenerateToken(adminUser.ID, adminUser.Username, roleNames)

	utils.Success(c, gin.H{
		"message": "系统初始化成功",
		"token":   token,
		"user": gin.H{
			"id":       adminUser.ID,
			"username": adminUser.Username,
			"avatar":   adminUser.Avatar,
			"roles":    roleNames,
		},
	})
}

// GetInitQRCode 获取初始化用的微信二维码
func (h *InitHandler) GetInitQRCode(c *gin.Context) {
	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		utils.Error(c, 400, "系统已经初始化")
		return
	}

	// 检查微信配置是否已保存
	var wechatAppIDConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_id").First(&wechatAppIDConfig).Error; err != nil {
		utils.Error(c, 400, "请先配置微信AppID和AppSecret")
		return
	}

	// 获取微信配置
	var wechatAppSecretConfig model.SystemConfig
	if err := h.db.Where("key = ?", "wechat_app_secret").First(&wechatAppSecretConfig).Error; err != nil {
		utils.Error(c, 400, "微信AppSecret配置不存在，请先保存微信配置")
		return
	}

	// 临时设置WeChatClient配置，去除首尾空格
	appID := strings.TrimSpace(wechatAppIDConfig.Value)
	appSecret := strings.TrimSpace(wechatAppSecretConfig.Value)
	h.wechatClient.SetAppID(appID)
	h.wechatClient.SetAppSecret(appSecret)
	
	// 验证配置是否为空
	if h.wechatClient.GetAppID() == "" || h.wechatClient.GetAppSecret() == "" {
		utils.Error(c, 400, "微信AppID或AppSecret配置为空，请检查配置")
		return
	}
	
	// 确保使用配置文件中的 account_type（如果未配置，默认使用 open_platform）
	accountType := config.AppConfig.WeChat.AccountType
	if accountType == "" {
		accountType = "open_platform" // 默认使用开放平台
	}
	h.wechatClient.SetAccountType(accountType)
	
	scope := config.AppConfig.WeChat.Scope
	if scope == "" {
		scope = "snsapi_userinfo" // 默认需要用户确认
	}
	h.wechatClient.SetScope(scope)

	// 获取回调地址（初始化回调地址）
	// 优先级：1. 配置文件中的 callback_domain 2. 查询参数 3. Referer 头 4. 默认值
	// 如果配置了 callback_domain，强制使用配置的域名，确保与微信后台配置一致
	var redirectURI string
	if config.AppConfig.WeChat.CallbackDomain != "" {
		// 优先使用配置文件中的回调域名
		domain := config.AppConfig.WeChat.CallbackDomain
		if len(domain) > 0 && domain[len(domain)-1] != '/' {
			domain += "/"
		}
		// 注意：回调路径需要包含 /api 前缀，因为后端路由都加了 /api 前缀
		redirectURI = domain + "api/init/callback"
	} else {
		// 如果未配置 callback_domain，使用查询参数或 Referer
		redirectURI = c.Query("redirect_uri")
		if redirectURI == "" {
			// 从 Referer 头获取
			referer := c.GetHeader("Referer")
			if referer != "" {
				redirectURI = referer + "/api/init/callback"
			} else {
				redirectURI = "http://localhost:8080/api/init/callback"
			}
		}
	}

	// 生成唯一的ticket（使用UUID）
	ticket := uuid.New().String()

	// 将ticket作为state参数的一部分，这样回调时可以获取到ticket
	// 注意：微信的state参数会原样返回，格式：ticket:uuid
	stateWithTicket := "ticket:" + ticket

	qrCode, err := h.wechatClient.GetQRCode(redirectURI, stateWithTicket)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取二维码失败: "+err.Error())
		return
	}

	// 返回授权URL和ticket，前端需要将其转换为二维码图片
	utils.Success(c, gin.H{
		"ticket":         ticket,     // 使用我们生成的UUID作为ticket
		"qr_code_url":    qrCode.URL, // 这是授权URL，需要转换为二维码
		"auth_url":       qrCode.URL, // 授权URL
		"expire_seconds": qrCode.ExpireSeconds,
	})
}

// InitSystemWithPassword 通过密码登录完成初始化（第二步：创建管理员）
func (h *InitHandler) InitSystemWithPassword(c *gin.Context) {
	// 检查是否已经初始化
	var existingConfig model.SystemConfig
	result := h.db.Where("key = ?", "initialized").First(&existingConfig)
	if result.Error == nil && existingConfig.Value == "true" {
		utils.Error(c, 400, "系统已经初始化，无法重复初始化")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.Error(c, 400, "用户名已存在")
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
			utils.Error(c, utils.CodeError, "创建管理员角色失败")
			return
		}
	}

	// 2. 验证密码强度并加密密码
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		tx.Rollback()
		utils.Error(c, 400, err.Error())
		return
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "加密密码失败")
		return
	}

	// 3. 创建管理员用户
	adminUser := model.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: hashedPassword,
		Status:   1,
	}
	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "创建管理员用户失败")
		return
	}

	// 4. 分配管理员角色
	if err := tx.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "分配管理员角色失败")
		return
	}

	// 5. 标记系统已初始化
	initConfig := model.SystemConfig{
		Key:   "initialized",
		Value: "true",
		Type:  "boolean",
	}
	if err := tx.Where("key = ?", "initialized").Assign(model.SystemConfig{Value: "true", Type: "boolean"}).FirstOrCreate(&initConfig).Error; err != nil {
		tx.Rollback()
		utils.Error(c, utils.CodeError, "保存初始化状态失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.Error(c, utils.CodeError, "初始化失败")
		return
	}

	// 生成管理员Token
	roleNames := []string{"admin"}
	token, err := auth.GenerateToken(adminUser.ID, adminUser.Username, roleNames)
	if err != nil {
		utils.Error(c, utils.CodeError, "生成Token失败")
		return
	}

	utils.Success(c, gin.H{
		"message": "系统初始化成功",
		"token":   token,
		"user": gin.H{
			"id":       adminUser.ID,
			"username": adminUser.Username,
			"nickname": adminUser.Nickname,
			"avatar":   adminUser.Avatar,
			"roles":    roleNames,
		},
	})
}
