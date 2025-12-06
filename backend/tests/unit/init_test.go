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
	"project-management/pkg/wechat"
	"project-management/tests/unit/mocks"
)

func TestInitHandler_CheckInitStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitHandler(db)

	t.Run("检查未初始化状态", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/status", nil)

		handler.CheckInitStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, false, data["initialized"])
	})

	t.Run("检查已初始化状态", func(t *testing.T) {
		// 设置初始化状态
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/status", nil)

		handler.CheckInitStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, true, data["initialized"])
	})
}

func TestInitHandler_SaveWeChatConfig(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitHandler(db)

	t.Run("保存微信配置成功", func(t *testing.T) {
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
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/wechat-config", bytes.NewBuffer(jsonData))
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

	t.Run("保存微信配置失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"wechat_app_id": "only_app_id",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/wechat-config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("保存微信配置失败-系统已初始化", func(t *testing.T) {
		// 设置初始化状态
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"wechat_app_id":     "test_app_id",
			"wechat_app_secret": "test_app_secret",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/wechat-config", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveWeChatConfig(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestInitHandler_InitSystemWithPassword(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitHandler(db)

	t.Run("通过密码初始化系统成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "admin",
			"password": "Admin123!@#",
			"nickname": "管理员",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystemWithPassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证管理员用户已创建
		var adminUser model.User
		err = db.Where("username = ?", "admin").First(&adminUser).Error
		assert.NoError(t, err)
		assert.Equal(t, "admin", adminUser.Username)
		assert.Equal(t, "管理员", adminUser.Nickname)

		// 验证管理员角色已分配
		var roles []model.Role
		db.Model(&adminUser).Association("Roles").Find(&roles)
		assert.GreaterOrEqual(t, len(roles), 1)

		// 验证系统已初始化
		var initConfig model.SystemConfig
		err = db.Where("key = ?", "initialized").First(&initConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "true", initConfig.Value)
	})

	t.Run("初始化系统失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "admin",
			// 缺少password和nickname
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystemWithPassword(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("初始化系统失败-用户名已存在", func(t *testing.T) {
		// 先创建一个用户
		existingUser := CreateTestUser(t, db, "existing_admin", "已存在用户")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": existingUser.Username, // 使用已存在的用户名
			"password": "Admin123!@#",
			"nickname": "管理员",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystemWithPassword(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("初始化系统失败-系统已初始化", func(t *testing.T) {
		// 设置初始化状态
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "admin2",
			"password": "Admin123!@#",
			"nickname": "管理员2",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystemWithPassword(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestInitHandler_GetInitQRCode(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitHandler(db)

	t.Run("获取初始化二维码失败-系统已初始化", func(t *testing.T) {
		// 设置初始化状态
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/qrcode", nil)

		handler.GetInitQRCode(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("获取初始化二维码失败-未配置微信AppID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/qrcode", nil)

		handler.GetInitQRCode(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("获取初始化二维码失败-未配置微信AppSecret", func(t *testing.T) {
		// 只配置AppID，不配置AppSecret
		appIDConfig := model.SystemConfig{
			Key:   "wechat_app_id",
			Value: "test_app_id",
			Type:  "string",
		}
		db.Create(&appIDConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/init/qrcode", nil)

		handler.GetInitQRCode(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	// 注意：GetInitQRCode需要调用微信API获取二维码，实际测试中需要mock微信客户端
	// 这里只测试参数验证和错误处理场景
}

func TestInitHandler_InitSystem(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewInitHandler(db)
	mockWeChatClient := mocks.NewMockWeChatClient()

	t.Run("通过微信扫码初始化系统成功", func(t *testing.T) {
		// 清理之前的数据
		db.Exec("DELETE FROM user_roles")
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM roles")
		db.Exec("DELETE FROM system_configs WHERE key = 'initialized'")
		db.Exec("DELETE FROM system_configs WHERE key IN ('wechat_app_id', 'wechat_app_secret')")

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
			Nickname:   "测试管理员",
			Sex:        1,
			Province:   "广东",
			City:       "深圳",
			Country:    "中国",
			HeadImgURL: "http://example.com/admin.jpg",
			Privilege:  []string{},
			UnionID:    "test_union_id",
		}

		// 替换WeChatClient
		handler.SetWeChatClient(mockWeChatClient)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code":  "test_code",
			"state": "test_state",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/system", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystem(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证管理员用户已创建
		var adminUser model.User
		err = db.Where("wechat_open_id = ?", "test_open_id").First(&adminUser).Error
		assert.NoError(t, err)
		assert.Equal(t, "测试管理员", adminUser.Nickname)

		// 验证系统已标记为初始化
		var initConfig model.SystemConfig
		err = db.Where("key = ?", "initialized").First(&initConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "true", initConfig.Value)

		// 验证WeChatClient方法被调用
		assert.Equal(t, 1, mockWeChatClient.GetAccessTokenCallCount)
		assert.Equal(t, 1, mockWeChatClient.GetUserInfoCallCount)
	})

	t.Run("初始化系统失败-系统已初始化", func(t *testing.T) {
		// 设置初始化状态
		initConfig := model.SystemConfig{
			Key:   "initialized",
			Value: "true",
			Type:  "boolean",
		}
		db.Create(&initConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code":  "test_code",
			"state": "test_state",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/system", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystem(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("初始化系统失败-未配置微信AppID", func(t *testing.T) {
		// 清理微信配置
		db.Exec("DELETE FROM system_configs WHERE key IN ('wechat_app_id', 'wechat_app_secret')")
		db.Exec("DELETE FROM system_configs WHERE key = 'initialized'")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code":  "test_code",
			"state": "test_state",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/init/system", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.InitSystem(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

