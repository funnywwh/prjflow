package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestWeChatHandler_GetWeChatConfig(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewWeChatHandler(db)

	t.Run("获取微信配置-默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/wechat/config", nil)

		handler.GetWeChatConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "", data["wechat_app_id"])
		assert.Equal(t, "", data["wechat_app_secret"])
		assert.Equal(t, "", data["account_type"])
		assert.Equal(t, "", data["scope"])
	})

	t.Run("获取微信配置-已配置", func(t *testing.T) {
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

		accountTypeConfig := model.SystemConfig{
			Key:   "wechat_account_type",
			Value: "open_platform",
			Type:  "string",
		}
		db.Create(&accountTypeConfig)

		scopeConfig := model.SystemConfig{
			Key:   "wechat_scope",
			Value: "snsapi_userinfo",
			Type:  "string",
		}
		db.Create(&scopeConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/wechat/config", nil)

		handler.GetWeChatConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test_app_id", data["wechat_app_id"])
		assert.Equal(t, "test_app_secret", data["wechat_app_secret"])
		assert.Equal(t, "open_platform", data["account_type"])
		assert.Equal(t, "snsapi_userinfo", data["scope"])
	})
}

func TestWeChatHandler_SaveWeChatConfig(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewWeChatHandler(db)

	t.Run("保存微信配置成功-仅必填字段", func(t *testing.T) {
		// 创建测试用户（用于审计日志）
		user := CreateTestUser(t, db, "testuser", "测试用户")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置用户信息（用于审计日志）
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)

		reqBody := map[string]interface{}{
			"wechat_app_id":     "test_app_id",
			"wechat_app_secret": "test_app_secret",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证配置已保存
		var appIDConfig model.SystemConfig
		err = db.Where("key = ?", "wechat_app_id").First(&appIDConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "test_app_id", appIDConfig.Value)

		var appSecretConfig model.SystemConfig
		err = db.Where("key = ?", "wechat_app_secret").First(&appSecretConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "test_app_secret", appSecretConfig.Value)
	})

	t.Run("保存微信配置成功-包含可选字段", func(t *testing.T) {
		// 创建测试用户（用于审计日志）
		user := CreateTestUser(t, db, "testuser2", "测试用户2")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置用户信息（用于审计日志）
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)

		reqBody := map[string]interface{}{
			"wechat_app_id":     "test_app_id_2",
			"wechat_app_secret": "test_app_secret_2",
			"account_type":      "official_account",
			"scope":             "snsapi_base",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证配置已保存
		var appIDConfig model.SystemConfig
		err = db.Where("key = ?", "wechat_app_id").First(&appIDConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "test_app_id_2", appIDConfig.Value)

		var accountTypeConfig model.SystemConfig
		err = db.Where("key = ?", "wechat_account_type").First(&accountTypeConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "official_account", accountTypeConfig.Value)

		var scopeConfig model.SystemConfig
		err = db.Where("key = ?", "wechat_scope").First(&scopeConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "snsapi_base", scopeConfig.Value)
	})

	t.Run("保存微信配置失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"wechat_app_id": "only_app_id",
			// 缺少wechat_app_secret
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("保存微信配置失败-参数错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 发送无效的JSON
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/config", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("更新微信配置", func(t *testing.T) {
		// 创建测试用户（用于审计日志）
		user := CreateTestUser(t, db, "testuser3", "测试用户3")

		// 先创建配置
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "old_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置用户信息（用于审计日志）
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)

		reqBody := map[string]interface{}{
			"wechat_app_id":     "new_app_id",
			"wechat_app_secret": "new_app_secret",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证配置已更新
		var updatedConfig model.SystemConfig
		err := db.Where("key = ?", "wechat_app_id").First(&updatedConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "new_app_id", updatedConfig.Value)
	})
}

