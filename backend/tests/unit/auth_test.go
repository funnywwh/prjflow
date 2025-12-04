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
	"project-management/internal/utils"
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
			"new_password": "newpassword123",
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
		assert.True(t, utils.CheckPassword("newpassword123", updatedUser.Password))
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

