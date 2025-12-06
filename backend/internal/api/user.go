package api

import (
	"strings"

	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/internal/websocket"
	"project-management/pkg/wechat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db           *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db:           db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []model.User
	query := h.db.Preload("Department").Preload("Roles")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 部门筛选
	if deptID := c.Query("department_id"); deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}

	// 分页
	page := utils.GetPage(c)
	pageSize := utils.GetPageSize(c)
	offset := (page - 1) * pageSize

	var total int64
	// 计算总数时需要应用与查询相同的筛选条件
	countQuery := h.db.Model(&model.User{})

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		countQuery = countQuery.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 部门筛选
	if deptID := c.Query("department_id"); deptID != "" {
		countQuery = countQuery.Where("department_id = ?", deptID)
	}

	countQuery.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		utils.Error(c, utils.CodeError, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":  users,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := h.db.Preload("Department").Preload("Roles").First(&user, id).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username" binding:"required"`
		Nickname     string `json:"nickname" binding:"required"` // 昵称（必填）
		Password     string `json:"password"`                    // 可选，如果提供则加密存储
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Avatar       string `json:"avatar"`
		Status       int    `json:"status"`
		DepartmentID *uint  `json:"department_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在（包括软删除的用户）
	var existingUser model.User
	// 使用 Unscoped() 查询，包括软删除的用户
	if err := h.db.Unscoped().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		// 如果用户存在且是软删除的，直接硬删除它（因为现在系统已改为硬删除）
		if existingUser.DeletedAt.Valid {
			// 先删除关联关系
			h.db.Model(&existingUser).Association("Roles").Clear()
			// 硬删除软删除的用户
			h.db.Unscoped().Delete(&existingUser, existingUser.ID)
		} else {
			// 用户存在且未删除
			utils.Error(c, 400, "用户名已存在")
			return
		}
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		Nickname:     req.Nickname,
		Email:        req.Email,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Status:       req.Status,
		DepartmentID: req.DepartmentID,
	}

	// 如果提供了密码，则验证密码强度并加密存储
	if req.Password != "" {
		// 验证密码强度：必须包含大小写字母和数字
		if err := utils.ValidatePasswordStrength(req.Password); err != nil {
			utils.Error(c, 400, err.Error())
			return
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Error(c, utils.CodeError, "加密密码失败")
			return
		}
		user.Password = hashedPassword
	}

	if err := h.db.Create(&user).Error; err != nil {
		// 检查是否是唯一约束错误
		if utils.IsUniqueConstraintError(err) {
			// 检查是哪个字段的唯一约束
			if utils.IsUniqueConstraintOnField(err, "username") {
				// 可能是软删除用户导致的冲突，尝试硬删除软删除的用户
				var softDeletedUser model.User
				if err := h.db.Unscoped().Where("username = ? AND deleted_at IS NOT NULL", req.Username).First(&softDeletedUser).Error; err == nil {
					// 找到软删除的用户，硬删除它
					h.db.Model(&softDeletedUser).Association("Roles").Clear()
					h.db.Unscoped().Delete(&softDeletedUser, softDeletedUser.ID)
					// 重试创建
					if err := h.db.Create(&user).Error; err != nil {
						if utils.IsUniqueConstraintError(err) {
							utils.Error(c, 400, "用户名已存在")
							return
						}
						utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
						return
					}
					// 创建成功，继续后续处理
				} else {
					// 不是软删除用户导致的冲突，用户名确实已存在
					utils.Error(c, 400, "用户名已存在")
					return
				}
			} else if utils.IsUniqueConstraintOnField(err, "wechat_open_id") {
				utils.Error(c, 400, "微信OpenID已存在")
				return
			} else {
				utils.Error(c, 400, "数据已存在，请检查唯一性约束")
				return
			}
		} else {
			// 其他数据库错误
			utils.Error(c, utils.CodeError, "创建失败: "+err.Error())
			return
		}
	}

	// 重新加载用户（包含关联数据）
	h.db.Preload("Department").Preload("Roles").First(&user, user.ID)

	utils.Success(c, user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	var req struct {
		Username     string `json:"username"`
		Nickname     string `json:"nickname" binding:"required"` // 昵称（必填，不能为空）
		Password     string `json:"password"`                    // 可选，如果提供则更新密码
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Avatar       string `json:"avatar"`
		Status       *int   `json:"status"`
		DepartmentID *uint  `json:"department_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 验证昵称不能为空
	if req.Nickname == "" {
		utils.Error(c, 400, "昵称不能为空")
		return
	}

	// 更新字段
	if req.Username != "" {
		// 检查新用户名是否已被其他用户使用
		var existingUser model.User
		if err := h.db.Where("username = ? AND id != ?", req.Username, id).First(&existingUser).Error; err == nil {
			utils.Error(c, 400, "用户名已存在")
			return
		}
		user.Username = req.Username
	}
	// 更新昵称（必填，不能为空）
	user.Nickname = req.Nickname
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.DepartmentID != nil {
		user.DepartmentID = req.DepartmentID
	}

	// 如果提供了新密码，则验证密码强度并加密更新
	if req.Password != "" {
		// 验证密码强度：必须包含大小写字母和数字
		if err := utils.ValidatePasswordStrength(req.Password); err != nil {
			utils.Error(c, 400, err.Error())
			return
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Error(c, utils.CodeError, "加密密码失败")
			return
		}
		user.Password = hashedPassword
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "更新失败")
		return
	}

	// 重新加载用户（包含关联数据）
	h.db.Preload("Department").Preload("Roles").First(&user, user.ID)

	utils.Success(c, user)
}

// DeleteUser 删除用户（硬删除）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	
	// 先检查用户是否存在
	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, 404, "用户不存在")
			return
		}
		utils.Error(c, utils.CodeError, "查询用户失败")
		return
	}
	
	// 硬删除：先删除关联关系，再删除用户
	// 删除用户角色关联
	if err := h.db.Model(&user).Association("Roles").Clear(); err != nil {
		utils.Error(c, utils.CodeError, "删除用户角色关联失败")
		return
	}
	
	// 硬删除用户
	if err := h.db.Unscoped().Delete(&model.User{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// AddUserByWeChatCallback 处理微信授权回调（GET请求，微信直接重定向到这里）
// 这个接口在微信内打开，处理完添加用户后通过WebSocket通知PC前端
func (h *UserHandler) AddUserByWeChatCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	handler := &AddUserCallbackHandler{db: h.db}
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, websocket.GetHub(), code, state, handler)

	if err != nil {
		c.Data(200, "text/html; charset=utf-8", []byte(handler.GetErrorHTML(ctx, err)))
		return
	}

	// 返回成功页面（在微信内显示）
	c.Data(200, "text/html; charset=utf-8", []byte(handler.GetSuccessHTML(ctx, result)))
}

// AddUserByWeChat 通过微信扫码添加用户（POST请求，保留用于前端回调页面调用）
// 注意：这个方法保留用于前端回调页面调用，但主要流程已改为使用 GET 回调接口
func (h *UserHandler) AddUserByWeChat(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 使用通用处理函数
	handler := &AddUserCallbackHandler{db: h.db}
	_, result, err := ProcessWeChatCallback(h.db, h.wechatClient, websocket.GetHub(), req.Code, req.State, handler)

	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	// 返回JSON响应（前端回调页面使用）
	utils.Success(c, result)
}

// GetUserWeChatBindQRCode 获取指定用户的微信绑定二维码（管理员操作）
func (h *UserHandler) GetUserWeChatBindQRCode(c *gin.Context) {
	// 检查用户是否已登录（管理员）
	_, exists := c.Get("user_id")
	if !exists {
		utils.Error(c, 401, "未授权，请先登录")
		return
	}

	// 获取要绑定的用户ID
	userID := c.Param("id")
	if userID == "" {
		utils.Error(c, 400, "用户ID不能为空")
		return
	}

	// 查找目标用户
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	// 检查用户是否已绑定微信
	if user.WeChatOpenID != nil && *user.WeChatOpenID != "" {
		utils.Error(c, 400, "该用户已绑定微信，请先解绑后再绑定新的微信")
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
	stateWithTicket := "bind:" + ticket + ":" + userID

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
