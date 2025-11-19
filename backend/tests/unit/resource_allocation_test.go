package unit

import (
	"bytes"
	"encoding/json"
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

func TestResourceAllocationHandler_GetResourceAllocations(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源分配测试项目")
	user := CreateTestUser(t, db, "allocuser", "资源分配用户")

	// 创建资源
	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	// 创建资源分配
	projectID := project.ID
	allocation1 := &model.ResourceAllocation{
		ResourceID: resource.ID,
		ProjectID:  &projectID,
		Date:       time.Now(),
		Hours:      8.0,
	}
	db.Create(allocation1)

	handler := api.NewResourceAllocationHandler(db)

	t.Run("获取所有资源分配", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resource-allocations", nil)

		handler.GetResourceAllocations(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("按资源筛选", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resource-allocations?resource_id=1", nil)

		handler.GetResourceAllocations(c)

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

func TestResourceAllocationHandler_GetResourceAllocation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "资源分配详情项目")
	user := CreateTestUser(t, db, "allocdetail", "资源分配详情用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	projectID := project.ID
	allocation := &model.ResourceAllocation{
		ResourceID: resource.ID,
		ProjectID:  &projectID,
		Date:       time.Now(),
		Hours:      8.0,
	}
	db.Create(&allocation)

	handler := api.NewResourceAllocationHandler(db)

	t.Run("获取存在的资源分配", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resource-allocations/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetResourceAllocation(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, 8.0, data["hours"])
	})

	t.Run("获取不存在的资源分配", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/resource-allocations/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetResourceAllocation(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceAllocationHandler_CreateResourceAllocation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建资源分配项目")
	user := CreateTestUser(t, db, "createalloc", "创建资源分配用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	handler := api.NewResourceAllocationHandler(db)

	t.Run("创建资源分配成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"resource_id": resource.ID,
			"project_id":  project.ID,
			"date":        time.Now().Format("2006-01-02"),
			"hours":       8.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resource-allocations", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResourceAllocation(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证资源分配已创建
		var allocation model.ResourceAllocation
		err = db.Where("resource_id = ?", resource.ID).First(&allocation).Error
		assert.NoError(t, err)
		assert.Equal(t, 8.0, allocation.Hours)
	})

	t.Run("创建资源分配失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"hours": 8.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/resource-allocations", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateResourceAllocation(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceAllocationHandler_UpdateResourceAllocation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新资源分配项目")
	user := CreateTestUser(t, db, "updatealloc", "更新资源分配用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	projectID := project.ID
	allocation := &model.ResourceAllocation{
		ResourceID: resource.ID,
		ProjectID:  &projectID,
		Date:       time.Now(),
		Hours:      8.0,
	}
	db.Create(&allocation)

	handler := api.NewResourceAllocationHandler(db)

	t.Run("更新资源分配成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"hours": 6.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/resource-allocations/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateResourceAllocation(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证资源分配已更新
		var updatedAllocation model.ResourceAllocation
		err := db.First(&updatedAllocation, allocation.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, 6.0, updatedAllocation.Hours)
	})

	t.Run("更新不存在的资源分配", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"hours": 6.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/resource-allocations/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateResourceAllocation(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestResourceAllocationHandler_DeleteResourceAllocation(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除资源分配项目")
	user := CreateTestUser(t, db, "deletealloc", "删除资源分配用户")

	resource := &model.Resource{
		UserID:    user.ID,
		ProjectID: project.ID,
		Role:      "developer",
	}
	db.Create(&resource)

	projectID := project.ID
	allocation := &model.ResourceAllocation{
		ResourceID: resource.ID,
		ProjectID:  &projectID,
		Date:       time.Now(),
		Hours:      8.0,
	}
	db.Create(&allocation)

	handler := api.NewResourceAllocationHandler(db)

	t.Run("删除资源分配成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/resource-allocations/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteResourceAllocation(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证资源分配已软删除
		var deletedAllocation model.ResourceAllocation
		err := db.First(&deletedAllocation, allocation.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedAllocation, allocation.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedAllocation.DeletedAt)
	})
}

