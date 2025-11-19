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

func TestVersionHandler_GetVersions(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "版本测试项目")

	// 创建测试版本
	version1 := &model.Version{
		VersionNumber: "v1.0.0",
		ProjectID:    project.ID,
		Status:       "draft",
	}
	db.Create(version1)

	handler := api.NewVersionHandler(db)

	t.Run("获取所有版本", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/versions", nil)

		handler.GetVersions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("搜索版本", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/versions?keyword=v1.0.0", nil)

		handler.GetVersions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	_ = project
}

func TestVersionHandler_GetVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "版本详情项目")

	version := &model.Version{
		VersionNumber: "v1.0.0",
		ProjectID:    project.ID,
		Status:       "draft",
	}
	db.Create(&version)

	handler := api.NewVersionHandler(db)

	t.Run("获取存在的版本", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/versions/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetVersion(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "v1.0.0", data["version_number"])
	})

	t.Run("获取不存在的版本", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/versions/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetVersion(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestVersionHandler_CreateVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建版本项目")
	handler := api.NewVersionHandler(db)

	t.Run("创建版本成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"version_number": "v2.0.0",
			"release_notes": "这是一个新版本",
			"status":         "draft",
			"project_id":     project.ID,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/versions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateVersion(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证版本已创建
		var version model.Version
		err = db.Where("version_number = ?", "v2.0.0").First(&version).Error
		assert.NoError(t, err)
		assert.Equal(t, "v2.0.0", version.VersionNumber)
		assert.Equal(t, project.ID, version.ProjectID)
	})

	t.Run("创建版本失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"release_notes": "只有发布说明",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/versions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateVersion(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建版本失败-项目不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"version_number": "v3.0.0",
			"project_id":     999,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/versions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateVersion(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestVersionHandler_UpdateVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新版本项目")

	version := &model.Version{
		VersionNumber: "v1.0.0",
		ProjectID:    project.ID,
		Status:       "draft",
	}
	db.Create(&version)

	handler := api.NewVersionHandler(db)

	t.Run("更新版本成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"version_number": "v1.1.0",
			"release_notes": "更新后的发布说明",
			"status":         "released",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/versions/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateVersion(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证版本已更新
		var updatedVersion model.Version
		err := db.First(&updatedVersion, version.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "v1.1.0", updatedVersion.VersionNumber)
		assert.Equal(t, "released", updatedVersion.Status)
	})

	t.Run("更新不存在的版本", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"version_number": "v999.0.0",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/versions/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateVersion(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestVersionHandler_DeleteVersion(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除版本项目")

	version := &model.Version{
		VersionNumber: "v1.0.0",
		ProjectID:    project.ID,
		Status:       "draft",
	}
	db.Create(&version)

	handler := api.NewVersionHandler(db)

	t.Run("删除版本成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/versions/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteVersion(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证版本已软删除
		var deletedVersion model.Version
		err := db.First(&deletedVersion, version.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedVersion, version.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedVersion.DeletedAt)
	})
}

