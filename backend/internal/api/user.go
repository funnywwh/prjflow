package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/wechat"
)

type UserHandler struct {
	db          *gorm.DB
	wechatClient *wechat.WeChatClient
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db:          db,
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
	query.Model(&model.User{}).Count(&total)

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
		Password     string `json:"password"` // 可选，如果提供则加密存储
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Avatar       string `json:"avatar"`
		Status       int    `json:"status"`
		DepartmentID *uint  `json:"department_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.Error(c, 400, "用户名已存在")
		return
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

	// 如果提供了密码，则加密存储
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Error(c, utils.CodeError, "加密密码失败")
			return
		}
		user.Password = hashedPassword
	}

	if err := h.db.Create(&user).Error; err != nil {
		utils.Error(c, utils.CodeError, "创建失败")
		return
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
		Password     string `json:"password"` // 可选，如果提供则更新密码
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

	// 如果提供了新密码，则加密更新
	if req.Password != "" {
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

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&model.User{}, id).Error; err != nil {
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
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, code, state, handler)

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
	_, result, err := ProcessWeChatCallback(h.db, h.wechatClient, req.Code, req.State, handler)
	
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	// 返回JSON响应（前端回调页面使用）
	utils.Success(c, result)
}

