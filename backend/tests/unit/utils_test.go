package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"project-management/internal/utils"
)

func TestHashPassword(t *testing.T) {
	t.Run("加密密码成功", func(t *testing.T) {
		password := "testpassword123"
		hash, err := utils.HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash) // 加密后的密码应该和原密码不同
	})

	t.Run("相同密码加密结果不同", func(t *testing.T) {
		password := "testpassword123"
		hash1, err1 := utils.HashPassword(password)
		hash2, err2 := utils.HashPassword(password)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		// bcrypt每次加密结果都不同（因为有salt）
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("验证密码成功", func(t *testing.T) {
		password := "testpassword123"
		hash, err := utils.HashPassword(password)
		assert.NoError(t, err)
		
		result := utils.CheckPassword(password, hash)
		assert.True(t, result)
	})

	t.Run("验证错误密码失败", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hash, err := utils.HashPassword(password)
		assert.NoError(t, err)
		
		result := utils.CheckPassword(wrongPassword, hash)
		assert.False(t, result)
	})

	t.Run("验证空密码", func(t *testing.T) {
		password := "testpassword123"
		hash, err := utils.HashPassword(password)
		assert.NoError(t, err)
		
		result := utils.CheckPassword("", hash)
		assert.False(t, result)
	})
}

func TestGetPage(t *testing.T) {
	t.Run("获取默认页码", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		
		page := utils.GetPage(c)
		assert.Equal(t, 1, page)
	})

	t.Run("获取指定页码", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page=5", nil)
		
		page := utils.GetPage(c)
		assert.Equal(t, 5, page)
	})

	t.Run("无效页码返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page=abc", nil)
		
		page := utils.GetPage(c)
		assert.Equal(t, 1, page)
	})

	t.Run("负数页码返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page=-1", nil)
		
		page := utils.GetPage(c)
		assert.Equal(t, 1, page)
	})

	t.Run("零页码返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page=0", nil)
		
		page := utils.GetPage(c)
		assert.Equal(t, 1, page)
	})
}

func TestGetPageSize(t *testing.T) {
	t.Run("获取默认每页数量", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 20, pageSize)
	})

	t.Run("获取指定每页数量", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page_size=50", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 50, pageSize)
	})

	t.Run("超过最大值的每页数量返回最大值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page_size=200", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 100, pageSize) // MaxPageSize = 100
	})

	t.Run("无效每页数量返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page_size=abc", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 20, pageSize)
	})

	t.Run("负数每页数量返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page_size=-1", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 20, pageSize)
	})

	t.Run("零每页数量返回默认值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test?page_size=0", nil)
		
		pageSize := utils.GetPageSize(c)
		assert.Equal(t, 20, pageSize)
	})
}

