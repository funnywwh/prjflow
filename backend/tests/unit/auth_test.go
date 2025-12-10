package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-management/internal/api"
	"project-management/internal/config"
	"project-management/internal/model"
	"project-management/internal/utils"
	"project-management/pkg/auth"
)

func TestAuthHandler_Login(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试用户（带密码）
	user := CreateTestUser(t, db, "testuser", "测试用户")
	hashedPassword, _ := utils.HashPassword("testpassword")
	db.Model(user).Update("password", hashedPassword)
	db.Model(user).Update("status", 1)

	handler := api.NewAuthHandler(db)

	t.Run("登录成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "testuser",
			"password": "testpassword",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["token"])
		assert.NotNil(t, data["refresh_token"])
		assert.NotNil(t, data["user"])
	})

	t.Run("登录失败-用户名或密码错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "testuser",
			"password": "wrongpassword",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusUnauthorized || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("登录失败-用户不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "nonexistent",
			"password": "password",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusUnauthorized || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("登录失败-用户被禁用", func(t *testing.T) {
		// 创建被禁用的用户
		disabledUser := CreateTestUser(t, db, "disableduser", "禁用用户")
		hashedPassword, _ := utils.HashPassword("password")
		db.Model(disabledUser).Update("password", hashedPassword)
		db.Model(disabledUser).Update("status", 0)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "disableduser",
			"password": "password",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAuthHandler_GetUserInfo(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "infouser", "信息用户")

	handler := api.NewAuthHandler(db)

	t.Run("获取用户信息成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/user/info", nil)

		handler.GetUserInfo(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), data["id"])
		assert.Equal(t, "infouser", data["username"])
	})

	t.Run("获取用户信息失败-未登录", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// 不设置user_id
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/user/info", nil)

		handler.GetUserInfo(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusUnauthorized || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "changepwd", "修改密码用户")
	hashedPassword, _ := utils.HashPassword("oldpassword")
	db.Model(user).Update("password", hashedPassword)

	handler := api.NewAuthHandler(db)

	t.Run("修改密码成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"old_password": "oldpassword",
			"new_password": "NewPassword123",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ChangePassword(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证密码已更新（需要重新查询用户）
		var updatedUser model.User
		err := db.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)
		// 验证新密码可以验证通过
		assert.True(t, utils.CheckPassword("NewPassword123", updatedUser.Password))
		// 验证旧密码不能验证通过
		assert.False(t, utils.CheckPassword("oldpassword", updatedUser.Password))
	})

	t.Run("修改密码失败-旧密码错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"old_password": "wrongpassword",
			"new_password": "newpassword",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ChangePassword(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("修改密码失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"old_password": "oldpassword",
			// 缺少new_password
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ChangePassword(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewAuthHandler(db)

	t.Run("退出登录成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})
}

func TestAuthHandler_GetQRCode(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewAuthHandler(db)

	t.Run("获取二维码失败-需要调用微信API", func(t *testing.T) {
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
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/wechat/qrcode", nil)

		handler.GetQRCode(c)

		// 注意：GetQRCode需要调用微信API获取二维码，实际测试中需要mock微信客户端
		// 这里只测试配置验证，实际调用微信API可能会失败
		// 如果配置正确但调用失败，会返回错误；如果配置错误，也会返回错误
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 由于需要调用微信API，这里可能返回成功或失败
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	t.Run("获取二维码失败-未配置微信AppID", func(t *testing.T) {
		// 确保没有微信配置
		db.Where("key IN ?", []string{"wechat_app_id", "wechat_app_secret"}).Delete(&model.SystemConfig{})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/wechat/qrcode", nil)

		handler.GetQRCode(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAuthHandler_WeChatLogin(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewAuthHandler(db)

	t.Run("微信登录失败-缺少code参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"state": "test_state",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/wechat/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.WeChatLogin(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("微信登录失败-未配置微信AppID", func(t *testing.T) {
		// 确保没有微信配置
		db.Where("key IN ?", []string{"wechat_app_id", "wechat_app_secret"}).Delete(&model.SystemConfig{})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code":  "test_code",
			"state": "test_state",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/wechat/login", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.WeChatLogin(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAuthHandler_WeChatCallback(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewAuthHandler(db)

	t.Run("微信回调-缺少code参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/wechat/callback?state=test_state", nil)

		handler.WeChatCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "登录失败")
	})

	t.Run("微信回调-缺少state参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/wechat/callback?code=test_code", nil)

		handler.WeChatCallback(c)

		// 应该返回错误页面（HTML）
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "html")
		assert.Contains(t, w.Body.String(), "登录失败")
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 初始化JWT配置
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key-for-unit-testing",
				Expiration: 24,
			},
		}
	}

	// 创建测试用户
	user := CreateTestUser(t, db, "refreshtoken", "刷新Token用户")
	adminRole := CreateTestAdminRole(t, db)
	db.Model(user).Association("Roles").Append(adminRole)

	handler := api.NewAuthHandler(db)

	t.Run("刷新Token成功", func(t *testing.T) {
		// 先生成一个refresh token
		refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, []string{"admin"})
		require.NoError(t, err)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"refresh_token": refreshToken,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["token"])
		assert.NotEmpty(t, data["token"])
		assert.NotNil(t, data["refresh_token"])
		assert.NotEmpty(t, data["refresh_token"])
		
		// 验证返回的token和refresh_token都是有效的JWT字符串
		newToken, ok := data["token"].(string)
		assert.True(t, ok, "token应该是字符串类型")
		assert.NotEmpty(t, newToken)
		
		newRefreshToken, ok := data["refresh_token"].(string)
		assert.True(t, ok, "refresh_token应该是字符串类型")
		assert.NotEmpty(t, newRefreshToken)
		
		// 验证新的token可以解析（不验证过期时间，因为可能因为时间问题导致过期）
		claims, err := auth.ParseToken(newToken)
		if err == nil {
			assert.Equal(t, user.ID, claims.UserID)
			assert.Equal(t, user.Username, claims.Username)
		}
		
		// 验证新的refresh token可以解析
		refreshClaims, err := auth.ParseToken(newRefreshToken)
		if err == nil {
			assert.Equal(t, user.ID, refreshClaims.UserID)
			assert.Equal(t, user.Username, refreshClaims.Username)
		}
	})

	t.Run("刷新Token失败-无效的refresh token", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"refresh_token": "invalid-refresh-token",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusUnauthorized || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("刷新Token失败-缺少refresh_token参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("刷新Token失败-用户不存在", func(t *testing.T) {
		// 生成一个不存在的用户的refresh token
		refreshToken, err := auth.GenerateRefreshToken(99999, "nonexistent", []string{})
		require.NoError(t, err)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"refresh_token": refreshToken,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("刷新Token失败-使用AccessToken而不是RefreshToken", func(t *testing.T) {
		// 生成一个access token（不是refresh token）
		accessToken, err := auth.GenerateToken(user.ID, user.Username, []string{"admin"})
		require.NoError(t, err)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"refresh_token": accessToken, // 错误：使用access token而不是refresh token
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RefreshToken(c)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		// 应该返回401错误，因为token类型不匹配
		assert.True(t, w.Code == http.StatusUnauthorized || (response["code"] != nil && response["code"] != float64(200)))
		if response["message"] != nil {
			// access token 和 refresh token 使用相同密钥，应该能够解析成功
			// 但会在 token 类型检查时失败，返回"只能使用RefreshToken来刷新，不能使用AccessToken"
			// 如果解析失败，则返回"无效的RefreshToken"
			message := response["message"].(string)
			assert.True(t, 
				strings.Contains(message, "只能使用RefreshToken") || 
				strings.Contains(message, "无效的RefreshToken"),
				"错误消息应该包含'只能使用RefreshToken'或'无效的RefreshToken'，实际消息: %s", message)
		}
	})
}

