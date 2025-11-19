package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestTagHandler_GetTags(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试标签
	_ = CreateTestTag(t, db, "标签1")
	_ = CreateTestTag(t, db, "标签2")

	handler := api.NewTagHandler(db)

	t.Run("获取所有标签", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tags", nil)

		handler.GetTags(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 2)
	})
}

func TestTagHandler_GetTag(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tag := CreateTestTag(t, db, "测试标签")
	handler := api.NewTagHandler(db)

	t.Run("获取存在的标签", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tags/%d", tag.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", tag.ID)}}

		handler.GetTag(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试标签", data["name"])
	})

	t.Run("获取不存在的标签", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tags/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetTag(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTagHandler_CreateTag(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewTagHandler(db)

	t.Run("创建标签成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "新标签",
			"description": "这是一个新标签",
			"color":       "red",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTag(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证标签已创建
		var tag model.Tag
		err = db.Where("name = ?", "新标签").First(&tag).Error
		assert.NoError(t, err)
		assert.Equal(t, "新标签", tag.Name)
		assert.Equal(t, "这是一个新标签", tag.Description)
		assert.Equal(t, "red", tag.Color)
	})

	t.Run("创建标签失败-标签名称已存在", func(t *testing.T) {
		// 先创建一个标签
		CreateTestTag(t, db, "已存在标签")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "已存在标签",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTag(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建标签失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTag(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTagHandler_UpdateTag(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tag := CreateTestTag(t, db, "待更新标签")
	handler := api.NewTagHandler(db)

	t.Run("更新标签成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "已更新标签",
			"description": "更新后的描述",
			"color":       "green",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tags/%d", tag.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", tag.ID)}}

		handler.UpdateTag(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证标签已更新
		var updatedTag model.Tag
		err := db.First(&updatedTag, tag.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新标签", updatedTag.Name)
		assert.Equal(t, "更新后的描述", updatedTag.Description)
		assert.Equal(t, "green", updatedTag.Color)
	})

	t.Run("更新不存在的标签", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的标签",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/tags/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateTag(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTagHandler_DeleteTag(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	tag := CreateTestTag(t, db, "待删除标签")
	handler := api.NewTagHandler(db)

	t.Run("删除标签成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tags/%d", tag.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", tag.ID)}}

		handler.DeleteTag(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证标签已删除（软删除）
		var deletedTag model.Tag
		err := db.First(&deletedTag, tag.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后，Unscoped查询能找到
		err = db.Unscoped().First(&deletedTag, tag.ID).Error
		assert.NoError(t, err)
	})

	t.Run("删除不存在的标签", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/tags/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.DeleteTag(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

