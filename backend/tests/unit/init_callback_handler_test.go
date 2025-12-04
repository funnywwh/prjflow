package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestInitCallbackHandlerImpl_Validate(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 注意：InitCallbackHandlerImpl的db字段是私有的，但Validate方法使用ctx.DB
	// 所以我们可以直接创建handler实例，只测试Validate方法
	handler := &api.InitCallbackHandlerImpl{}

	t.Run("验证成功-系统未初始化且微信配置已保存", func(t *testing.T) {
		// 设置微信配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := handler.Validate(ctx)
		assert.NoError(t, err)
	})

	t.Run("验证失败-系统已初始化", func(t *testing.T) {
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

		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := handler.Validate(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "系统已经初始化")
	})

	t.Run("验证失败-微信配置未保存", func(t *testing.T) {
		// 确保没有初始化配置和微信配置
		db.Where("key IN ?", []string{"initialized", "wechat_app_id", "wechat_app_secret"}).Delete(&model.SystemConfig{})

		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := handler.Validate(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "请先配置微信AppID和AppSecret")
	})
}

func TestInitCallbackHandlerImpl_GetSuccessHTML(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := &api.InitCallbackHandlerImpl{}

	t.Run("获取成功HTML", func(t *testing.T) {
		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		html := handler.GetSuccessHTML(ctx, nil)
		assert.Contains(t, html, "系统初始化成功")
		assert.Contains(t, html, "请返回 PC 端查看")
	})
}

func TestInitCallbackHandlerImpl_GetErrorHTML(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := &api.InitCallbackHandlerImpl{}

	t.Run("获取错误HTML", func(t *testing.T) {
		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := &api.CallbackError{Message: "测试错误"}
		html := handler.GetErrorHTML(ctx, err)
		assert.Contains(t, html, "初始化失败")
		assert.Contains(t, html, "测试错误")
	})
}

// 注意：Process方法的测试需要mock微信API，比较复杂
// 这里只测试基本的Validate逻辑和HTML生成

