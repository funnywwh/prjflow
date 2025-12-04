package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestWeChatVerifyHandler_HandleVerifyFile(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewWeChatVerifyHandler(db)

	t.Run("从数据库读取验证文件成功", func(t *testing.T) {
		// 设置验证文件内容到数据库
		verifyConfig := model.SystemConfig{
			Key:   "wechat_verify_file_MP_verify_test123.txt",
			Value: "test123",
			Type:  "string",
		}
		db.Create(&verifyConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/wechat/verify/test123.txt", nil)
		c.Params = gin.Params{gin.Param{Key: "code", Value: "test123.txt"}}

		handler.HandleVerifyFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test123", w.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("验证文件不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/wechat/verify/nonexistent.txt", nil)
		c.Params = gin.Params{gin.Param{Key: "code", Value: "nonexistent.txt"}}

		handler.HandleVerifyFile(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Verification file not found")
	})

	t.Run("code参数为空", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/wechat/verify/", nil)
		c.Params = gin.Params{gin.Param{Key: "code", Value: ""}}

		handler.HandleVerifyFile(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "File not found")
	})

	t.Run("code参数不包含.txt后缀", func(t *testing.T) {
		// 设置验证文件内容到数据库
		verifyConfig := model.SystemConfig{
			Key:   "wechat_verify_file_MP_verify_test456.txt",
			Value: "test456",
			Type:  "string",
		}
		db.Create(&verifyConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/wechat/verify/test456", nil)
		c.Params = gin.Params{gin.Param{Key: "code", Value: "test456"}}

		handler.HandleVerifyFile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test456", w.Body.String())
	})
}

func TestWeChatVerifyHandler_SaveVerifyFile(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewWeChatVerifyHandler(db)

	t.Run("保存验证文件成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"filename": "MP_verify_test789.txt",
			"content":  "test789",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/verify", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveVerifyFile(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证文件已保存到数据库
		var verifyConfig model.SystemConfig
		err := db.Where("key = ?", "wechat_verify_file_MP_verify_test789.txt").First(&verifyConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "test789", verifyConfig.Value)
	})

	t.Run("更新已存在的验证文件", func(t *testing.T) {
		// 先创建一个验证文件
		verifyConfig := model.SystemConfig{
			Key:   "wechat_verify_file_MP_verify_test999.txt",
			Value: "old_value",
			Type:  "string",
		}
		db.Create(&verifyConfig)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"filename": "MP_verify_test999.txt",
			"content":  "new_value",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/verify", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveVerifyFile(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证文件已更新
		var updatedConfig model.SystemConfig
		err := db.Where("key = ?", "wechat_verify_file_MP_verify_test999.txt").First(&updatedConfig).Error
		assert.NoError(t, err)
		assert.Equal(t, "new_value", updatedConfig.Value)
	})

	t.Run("保存验证文件失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"content": "test789",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/wechat/verify", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveVerifyFile(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

