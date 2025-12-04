package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"project-management/internal/api"
	"project-management/internal/model"
	"project-management/pkg/wechat"
	"project-management/tests/unit/mocks"
)

func TestProcessWeChatCallback_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试Handler
	handler := &api.InitCallbackHandlerImpl{}

	t.Run("成功处理微信回调", func(t *testing.T) {
		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		appSecretConfig := model.SystemConfig{
			Key:   "wechat_app_secret",
			Value: "test_app_secret",
			Type:  "string",
		}
		db.Create(&appSecretConfig)

		// 创建Mock对象
		mockWeChatClient := mocks.NewMockWeChatClient()
		mockHub := mocks.NewMockWebSocketHub()

		// 配置Mock返回值
		mockWeChatClient.AccessTokenResponse = &wechat.AccessTokenResponse{
			AccessToken:  "test_access_token",
			ExpiresIn:    7200,
			RefreshToken: "test_refresh_token",
			OpenID:       "test_open_id",
			Scope:        "snsapi_userinfo",
			UnionID:      "test_union_id",
		}

		mockWeChatClient.UserInfoResponse = &wechat.UserInfoResponse{
			OpenID:     "test_open_id",
			Nickname:   "测试用户",
			Sex:        1,
			Province:   "广东",
			City:       "深圳",
			Country:    "中国",
			HeadImgURL: "http://example.com/avatar.jpg",
			Privilege:  []string{},
			UnionID:    "test_union_id",
		}
		code := "test_code"
		state := "ticket:test_ticket_123"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证没有错误
		assert.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.NotNil(t, result)

		// 验证WeChatClient方法被调用
		assert.Equal(t, 1, mockWeChatClient.GetAccessTokenCallCount)
		assert.Equal(t, 1, mockWeChatClient.GetUserInfoCallCount)
		assert.Equal(t, code, mockWeChatClient.GetAccessTokenCodes[0])

		// 验证WebSocket消息被发送
		messages := mockHub.GetMessagesByTicket("test_ticket_123")
		assert.Greater(t, len(messages), 0)

		// 验证上下文数据
		assert.Equal(t, "test_ticket_123", ctx.Ticket)
		assert.NotNil(t, ctx.AccessToken)
		assert.NotNil(t, ctx.UserInfo)
		assert.Equal(t, "test_open_id", ctx.UserInfo.OpenID)
	})

	t.Run("成功处理微信回调-无ticket", func(t *testing.T) {
		// 使用AddUserCallbackHandler，它不需要系统初始化
		userHandler := &api.AddUserCallbackHandler{}

		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id_2",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		appSecretConfig := model.SystemConfig{
			Key:   "wechat_app_secret",
			Value: "test_app_secret_2",
			Type:  "string",
		}
		db.Create(&appSecretConfig)

		// 创建新的Mock对象（避免状态污染）
		mockWeChatClient := mocks.NewMockWeChatClient()
		mockHub := mocks.NewMockWebSocketHub()

		// 配置Mock返回值
		mockWeChatClient.AccessTokenResponse = &wechat.AccessTokenResponse{
			AccessToken: "test_access_token_2",
			ExpiresIn:    7200,
			RefreshToken: "test_refresh_token_2",
			OpenID:       "test_open_id_2",
			Scope:        "snsapi_userinfo",
			UnionID:      "test_union_id_2",
		}

		mockWeChatClient.UserInfoResponse = &wechat.UserInfoResponse{
			OpenID:     "test_open_id_2",
			Nickname:   "测试用户2",
			Sex:        1,
			Province:   "广东",
			City:       "深圳",
			Country:    "中国",
			HeadImgURL: "http://example.com/avatar2.jpg",
			Privilege:  []string{},
			UnionID:    "test_union_id_2",
		}

		code := "test_code_2"
		state := "" // 空state，没有ticket

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, userHandler)

		// 验证没有错误
		assert.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.NotNil(t, result)

		// 验证ticket为空
		assert.Equal(t, "", ctx.Ticket)

		// 验证没有WebSocket消息（因为没有ticket）
		assert.Equal(t, 0, mockHub.MessageCount())
	})
}

func TestProcessWeChatCallback_Errors(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	mockWeChatClient := mocks.NewMockWeChatClient()
	mockHub := mocks.NewMockWebSocketHub()
	handler := &api.InitCallbackHandlerImpl{}

	t.Run("code为空", func(t *testing.T) {
		mockWeChatClient.Reset()
		mockHub.Reset()

		code := ""
		state := "ticket:test_ticket"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证返回错误
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, ctx)
		assert.Contains(t, err.Error(), "未获取到授权码")

		// 验证WebSocket错误消息被发送
		messages := mockHub.GetMessagesByTicket("test_ticket")
		assert.Greater(t, len(messages), 0)
		assert.True(t, mockHub.HasMessage("test_ticket", "error", "未获取到授权码"))
	})

	t.Run("微信配置未设置", func(t *testing.T) {
		mockWeChatClient.Reset()
		mockHub.Reset()

		// 确保没有微信配置
		db.Where("key IN ?", []string{"wechat_app_id", "wechat_app_secret"}).Delete(&model.SystemConfig{})

		code := "test_code"
		state := "ticket:test_ticket"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证返回错误
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, ctx)
		assert.Contains(t, err.Error(), "请先配置微信AppID和AppSecret")
	})

	t.Run("GetAccessToken失败", func(t *testing.T) {
		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		appSecretConfig := model.SystemConfig{
			Key:   "wechat_app_secret",
			Value: "test_app_secret",
			Type:  "string",
		}
		db.Create(&appSecretConfig)

		mockWeChatClient.Reset()
		mockHub.Reset()

		// 配置Mock返回错误
		mockWeChatClient.AccessTokenError = &api.CallbackError{Message: "获取access_token失败"}

		code := "test_code"
		state := "ticket:test_ticket"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证返回错误
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, ctx)
		assert.Contains(t, err.Error(), "获取access_token失败")

		// 验证GetAccessToken被调用
		assert.Equal(t, 1, mockWeChatClient.GetAccessTokenCallCount)
	})

	t.Run("GetUserInfo失败", func(t *testing.T) {
		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		appSecretConfig := model.SystemConfig{
			Key:   "wechat_app_secret",
			Value: "test_app_secret",
			Type:  "string",
		}
		db.Create(&appSecretConfig)

		mockWeChatClient.Reset()
		mockHub.Reset()

		// 配置Mock返回值
		mockWeChatClient.AccessTokenResponse = &wechat.AccessTokenResponse{
			AccessToken: "test_access_token",
			OpenID:      "test_open_id",
		}
		mockWeChatClient.UserInfoError = &api.CallbackError{Message: "获取用户信息失败", Err: nil}

		code := "test_code"
		state := "ticket:test_ticket"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证返回错误
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, ctx)
		assert.Contains(t, err.Error(), "获取用户信息失败")

		// 验证GetUserInfo被调用
		assert.Equal(t, 1, mockWeChatClient.GetUserInfoCallCount)
	})

	t.Run("Handler.Validate失败", func(t *testing.T) {
		// 设置系统已初始化
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		appSecretConfig := model.SystemConfig{
			Key:   "wechat_app_secret",
			Value: "test_app_secret",
			Type:  "string",
		}
		db.Create(&appSecretConfig)

		mockWeChatClient.Reset()
		mockHub.Reset()

		code := "test_code"
		state := "ticket:test_ticket"

		ctx, result, err := api.ProcessWeChatCallback(db, mockWeChatClient, mockHub, code, state, handler)

		// 验证返回错误（系统已初始化）
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, ctx)
		assert.Contains(t, err.Error(), "系统已经初始化")
	})
}

