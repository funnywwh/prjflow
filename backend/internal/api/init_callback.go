package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"project-management/internal/websocket"
	"project-management/pkg/wechat"
)

// InitCallbackHandler 处理微信回调
type InitCallbackHandler struct {
	db          *gorm.DB
	wechatClient wechat.WeChatClientInterface // 使用接口类型，支持依赖注入
}

func NewInitCallbackHandler(db *gorm.DB) *InitCallbackHandler {
	return &InitCallbackHandler{
		db:          db,
		wechatClient: wechat.NewWeChatClient(),
	}
}

// SetWeChatClient 设置WeChatClient（用于测试）
func (h *InitCallbackHandler) SetWeChatClient(client wechat.WeChatClientInterface) {
	h.wechatClient = client
}

// HandleCallback 处理微信授权回调
func (h *InitCallbackHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	handler := &InitCallbackHandlerImpl{db: h.db}
	ctx, result, err := ProcessWeChatCallback(h.db, h.wechatClient, websocket.GetHub(), code, state, handler)
	
	if err != nil {
		c.Data(200, "text/html; charset=utf-8", []byte(handler.GetErrorHTML(ctx, err)))
		return
	}

	// 返回成功页面（在微信内显示）
	c.Data(200, "text/html; charset=utf-8", []byte(handler.GetSuccessHTML(ctx, result)))
}

