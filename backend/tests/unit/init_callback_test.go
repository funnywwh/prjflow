package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestInitCallbackHandler_HandleCallback(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitCallbackHandler(db)

	t.Run("缺少code参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/callback?state=ticket:test123", nil)

		handler.HandleCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "初始化失败")
	})

	t.Run("缺少state参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/callback?code=testcode", nil)

		handler.HandleCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "初始化失败")
	})

	t.Run("微信配置未设置", func(t *testing.T) {
		// 确保没有微信配置
		db.Where("key IN ?", []string{"wechat_app_id", "wechat_app_secret"}).Delete(&model.SystemConfig{})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/callback?code=testcode&state=ticket:test123", nil)

		handler.HandleCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "初始化失败")
	})

	t.Run("系统已初始化", func(t *testing.T) {
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

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/callback?code=testcode&state=ticket:test123", nil)

		handler.HandleCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "初始化失败")
	})

	// 注意：完整的成功场景测试需要mock微信API，比较复杂
	// 这里只测试基本的参数验证和错误处理
}

