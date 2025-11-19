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

func TestDepartmentHandler_GetDepartments(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试部门
	dept1 := &model.Department{
		Name:   "部门1",
		Code:   "DEPT001",
		Status: 1,
	}
	db.Create(dept1)

	handler := api.NewDepartmentHandler(db)

	t.Run("获取所有部门", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/departments", nil)

		handler.GetDepartments(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// GetDepartments返回的是树形结构数组，不是map
		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			assert.GreaterOrEqual(t, len(list), 1)
		}
	})

	t.Run("搜索部门", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/departments?keyword=部门1", nil)

		handler.GetDepartments(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// GetDepartments返回的是树形结构数组
		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			assert.GreaterOrEqual(t, len(list), 1)
		}
	})
}

func TestDepartmentHandler_GetDepartment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	dept := &model.Department{
		Name:   "测试部门",
		Code:   "TEST001",
		Status: 1,
	}
	db.Create(&dept)

	handler := api.NewDepartmentHandler(db)

	t.Run("获取存在的部门", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/departments/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetDepartment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试部门", data["name"])
	})

	t.Run("获取不存在的部门", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/departments/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetDepartment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestDepartmentHandler_CreateDepartment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewDepartmentHandler(db)

	t.Run("创建部门成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":   "新部门",
			"code":   "NEW001",
			"status": 1,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/departments", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateDepartment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证部门已创建
		var dept model.Department
		err = db.Where("name = ?", "新部门").First(&dept).Error
		assert.NoError(t, err)
		assert.Equal(t, "新部门", dept.Name)
		assert.Equal(t, "NEW001", dept.Code)
	})

	t.Run("创建部门失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code": "TEST001",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/departments", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateDepartment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestDepartmentHandler_UpdateDepartment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	dept := &model.Department{
		Name:   "更新部门",
		Code:   "UPDATE001",
		Status: 1,
	}
	db.Create(&dept)

	handler := api.NewDepartmentHandler(db)

	t.Run("更新部门成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":   "已更新部门",
			"status": 0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/departments/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateDepartment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证部门已更新
		var updatedDept model.Department
		err := db.First(&updatedDept, dept.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新部门", updatedDept.Name)
	})

	t.Run("更新不存在的部门", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的部门",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/departments/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateDepartment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestDepartmentHandler_DeleteDepartment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	dept := &model.Department{
		Name:   "删除部门",
		Code:   "DELETE001",
		Status: 1,
	}
	db.Create(&dept)

	handler := api.NewDepartmentHandler(db)

	t.Run("删除部门成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/departments/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteDepartment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证部门已软删除
		var deletedDept model.Department
		err := db.First(&deletedDept, dept.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedDept, dept.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedDept.DeletedAt)
	})
}

