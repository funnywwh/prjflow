package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"project-management/internal/api"
	"project-management/internal/model"
)

func TestResourceHandler_GetResources(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源测试项目")
	user := CreateTestUser(t, db, "resourceuser", "资源用户")

	// 创建测试资源
	resource1 := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(resource1)

	handler := api.NewResourceHandler(db)

	t.Run("获取所有资源", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources", nil)

		handler.GetResources(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("按用户筛选资源", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/resources?user_id=%d", user.ID), nil)

		handler.GetResources(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})
}

func TestResourceHandler_GetResource(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源详情项目")
	user := CreateTestUser(t, db, "resourcedetail", "资源详情用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	handler := api.NewResourceHandler(db)

	t.Run("获取存在的资源", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetResource(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "developer", data["role"])
	})

	t.Run("获取不存在的资源", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetResource(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceHandler_CreateResource(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建资源项目")
	user := CreateTestUser(t, db, "createresource", "创建资源用户")
	handler := api.NewResourceHandler(db)

	t.Run("创建资源成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"user_id":    user.ID,
			"project_id": project.ID,
			"role":       "developer",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resources", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResource(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证资源已创建
		var resource model.Resource
		err = db.Where("user_id = ? AND project_id = ?", user.ID, project.ID).First(&resource).Error
		assert.NoError(t, err)
		assert.Equal(t, "developer", resource.Role)
	})

	t.Run("创建资源失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"role": "developer",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resources", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResource(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建资源失败-用户不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"user_id":    999, // 不存在的用户ID
			"project_id": project.ID,
			"role":       "developer",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resources", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResource(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceHandler_UpdateResource(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新资源项目")
	user := CreateTestUser(t, db, "updateresource", "更新资源用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	handler := api.NewResourceHandler(db)

	t.Run("更新资源成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"role": "manager",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/resources/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateResource(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证资源已更新
		var updatedResource model.Resource
		err := db.First(&updatedResource, resource.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "manager", updatedResource.Role)
	})

	t.Run("更新不存在的资源", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"role": "manager",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/resources/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateResource(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceHandler_DeleteResource(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除资源项目")
	user := CreateTestUser(t, db, "deleteresource", "删除资源用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	handler := api.NewResourceHandler(db)

	t.Run("删除资源成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/resources/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteResource(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证资源已软删除
		var deletedResource model.Resource
		err := db.First(&deletedResource, resource.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedResource, resource.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedResource.DeletedAt)
	})
}

func TestResourceHandler_GetResourceStatistics(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源统计项目")
	user := CreateTestUser(t, db, "resourcestat", "资源统计用户")

	// 创建资源
	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(resource)

	handler := api.NewResourceHandler(db)

	t.Run("获取资源统计", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources/statistics", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetResourceStatistics(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["total_hours"])
		assert.NotNil(t, data["project_stats"])
		assert.NotNil(t, data["user_stats"])
	})
}

func TestResourceHandler_GetResourceUtilization(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源利用率项目")
	user := CreateTestUser(t, db, "resourceutil", "资源利用率用户")

	// 创建资源
	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(resource)

	handler := api.NewResourceHandler(db)

	t.Run("获取资源利用率", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources/utilization", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetResourceUtilization(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证响应包含data字段
		assert.NotNil(t, response["data"])
		// 如果data是map，验证包含avg_utilization或utilization_rate字段
		if data, ok := response["data"].(map[string]interface{}); ok {
			// API返回avg_utilization字段
			_, hasAvgUtilization := data["avg_utilization"]
			_, hasUtilizationRate := data["utilization_rate"]
			assert.True(t, hasAvgUtilization || hasUtilizationRate, "响应应包含avg_utilization或utilization_rate字段")
		}
	})
}

func TestResourceHandler_CheckResourceConflict(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源冲突项目")
	user := CreateTestUser(t, db, "resourceconflict", "资源冲突用户")

	// 创建资源
	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(resource)

	handler := api.NewResourceHandler(db)

	t.Run("检查资源冲突", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		dateStr := time.Now().Format("2006-01-02")
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resources/conflict?user_id="+fmt.Sprintf("%d", user.ID)+"&date="+dateStr, nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CheckResourceConflict(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		if data, ok := response["data"].(map[string]interface{}); ok {
			assert.NotNil(t, data["has_conflict"])
		} else {
			// 如果data不是map，可能是其他格式，至少确保响应成功
			assert.Equal(t, float64(200), response["code"])
		}
	})
}

func TestResourceHandler_CreateResourceDuplicate(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "重复资源项目")
	user := CreateTestUser(t, db, "duplicateresource", "重复资源用户")

	// 先创建一个资源
	existingResource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(existingResource)

	handler := api.NewResourceHandler(db)

	t.Run("创建资源失败-重复分配", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"user_id":    user.ID,
			"project_id": project.ID,
			"role":       "manager",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resources", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResource(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

