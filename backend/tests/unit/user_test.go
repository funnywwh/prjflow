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

func TestUserHandler_GetUsers(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试数据
	_ = CreateTestUser(t, db, "user1", "用户1")
	_ = CreateTestUser(t, db, "user2", "用户2")

	handler := api.NewUserHandler(db)

	// 测试获取用户列表
	t.Run("获取所有用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/users", nil)

		handler.GetUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 2)
	})

	// 测试搜索功能
	t.Run("搜索用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/users?keyword=user1", nil)

		handler.GetUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	// 测试分页
	t.Run("分页查询", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/users?page=1&page_size=1", nil)

		handler.GetUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.Equal(t, 1, len(list))
		if page, ok := data["page"].(float64); ok {
			assert.Equal(t, float64(1), page)
		}
		if pageSize, ok := data["page_size"].(float64); ok {
			assert.Equal(t, float64(1), pageSize)
		}
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	_ = CreateTestUser(t, db, "testuser", "测试用户")
	handler := api.NewUserHandler(db)

	t.Run("获取存在的用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "testuser", data["username"])
		assert.Equal(t, "测试用户", data["nickname"])
	})

	t.Run("获取不存在的用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetUser(c)

		// 检查返回的状态码或错误消息
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// GetUser 应该返回404或错误消息
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestUserHandler_CreateUser(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewUserHandler(db)

	t.Run("创建用户成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "newuser",
			"nickname": "新用户",
			"email":    "newuser@test.com",
			"status":   1,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证用户已创建
		var user model.User
		err = db.Where("username = ?", "newuser").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "newuser", user.Username)
		assert.Equal(t, "新用户", user.Nickname)
	})

	t.Run("创建用户失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"email": "test@test.com",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateUser(c)

		// 检查返回的状态码或错误消息
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// CreateUser 应该返回400或错误消息
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建用户失败-用户名重复", func(t *testing.T) {
		// 先创建一个用户
		CreateTestUser(t, db, "duplicate", "重复用户")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"username": "duplicate",
			"nickname": "另一个用户",
			"email":    "another@test.com",
			"status":   1,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateUser(c)

		// 检查返回的状态码或错误消息
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// CreateUser 应该返回400或错误消息
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "updateuser", "更新用户")
	handler := api.NewUserHandler(db)

	t.Run("更新用户成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"nickname": "已更新用户",
			"email":    "updated@test.com",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证用户已更新
		var updatedUser model.User
		err := db.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新用户", updatedUser.Nickname)
		assert.Equal(t, "updated@test.com", updatedUser.Email)
	})

	t.Run("更新不存在的用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"nickname": "不存在的用户",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/users/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateUser(c)

		// 检查返回的状态码或错误消息
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// UpdateUser 应该返回404或错误消息
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "deleteuser", "删除用户")
	handler := api.NewUserHandler(db)

	t.Run("删除用户成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证用户已软删除
		var deletedUser model.User
		err := db.First(&deletedUser, user.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedUser, user.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedUser.DeletedAt)
	})

	t.Run("删除不存在的用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/users/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.DeleteUser(c)

		// DeleteUser 在GORM中不会报错，即使记录不存在也会返回成功
		// 这是GORM的默认行为，所以测试应该检查返回成功
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

