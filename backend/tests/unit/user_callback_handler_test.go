package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"project-management/internal/api"
)

func TestAddUserCallbackHandler_Validate(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 注意：AddUserCallbackHandler的db字段是私有的，但Validate方法使用ctx.DB
	// 所以我们可以直接创建handler实例，只测试Validate方法
	handler := &api.AddUserCallbackHandler{}

	t.Run("验证成功-添加用户场景无需特殊验证", func(t *testing.T) {
		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := handler.Validate(ctx)
		assert.NoError(t, err)
	})
}

func TestAddUserCallbackHandler_GetSuccessHTML(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := &api.AddUserCallbackHandler{}

	t.Run("获取成功HTML", func(t *testing.T) {
		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		html := handler.GetSuccessHTML(ctx, nil)
		assert.Contains(t, html, "用户添加成功")
		assert.Contains(t, html, "请返回 PC 端查看")
	})
}

func TestAddUserCallbackHandler_GetErrorHTML(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := &api.AddUserCallbackHandler{}

	t.Run("获取错误HTML", func(t *testing.T) {
		ctx := &api.WeChatCallbackContext{
			DB: db,
		}

		err := &api.CallbackError{Message: "测试错误"}
		html := handler.GetErrorHTML(ctx, err)
		assert.Contains(t, html, "添加用户失败")
		assert.Contains(t, html, "测试错误")
	})
}

// 注意：Process方法的测试需要mock微信API和WebSocket，比较复杂
// 这里只测试基本的Validate逻辑和HTML生成

