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

func TestModuleHandler_GetModules(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试模块
	module1 := &model.Module{
		Name:        "测试模块1",
		Code:        "test_module_1",
		Description: "测试模块描述1",
		Status:      1,
		Sort:        0,
	}
	db.Create(module1)

	module2 := &model.Module{
		Name:        "测试模块2",
		Code:        "test_module_2",
		Description: "测试模块描述2",
		Status:      1,
		Sort:        1,
	}
	db.Create(module2)

	handler := api.NewModuleHandler(db)

	t.Run("获取所有模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/modules", nil)

		handler.GetModules(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 2)
	})

	t.Run("搜索模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/modules?keyword=测试模块1", nil)

		handler.GetModules(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 1)
	})

	t.Run("按状态筛选模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/modules?status=1", nil)

		handler.GetModules(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 2)
	})
}

func TestModuleHandler_GetModule(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	module := &model.Module{
		Name:        "测试模块详情",
		Code:        "test_module_detail",
		Description: "测试模块详情描述",
		Status:      1,
		Sort:        0,
	}
	db.Create(&module)

	handler := api.NewModuleHandler(db)

	t.Run("获取存在的模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/modules/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetModule(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试模块详情", data["name"])
	})

	t.Run("获取不存在的模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/modules/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetModule(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestModuleHandler_CreateModule(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewModuleHandler(db)

	t.Run("创建模块成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "新模块",
			"code":        "new_module",
			"description": "新模块描述",
			"status":      1,
			"sort":        0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateModule(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证模块已创建
		var module model.Module
		err = db.Where("name = ?", "新模块").First(&module).Error
		assert.NoError(t, err)
		assert.Equal(t, "新模块", module.Name)
		assert.Equal(t, "new_module", module.Code)
	})

	t.Run("创建模块失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code": "only_code",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateModule(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建模块失败-名称已存在", func(t *testing.T) {
		// 先创建一个模块
		existingModule := &model.Module{
			Name:  "已存在模块",
			Code:  "existing_module",
			Status: 1,
		}
		db.Create(existingModule)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":  "已存在模块", // 重复的名称
			"code":  "another_module",
			"status": 1,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateModule(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestModuleHandler_UpdateModule(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	module := &model.Module{
		Name:        "更新模块",
		Code:        "update_module",
		Description: "更新模块描述",
		Status:      1,
		Sort:        0,
	}
	db.Create(&module)

	handler := api.NewModuleHandler(db)

	t.Run("更新模块成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "已更新模块",
			"description": "已更新描述",
			"status":      0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/modules/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateModule(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证模块已更新
		var updatedModule model.Module
		err := db.First(&updatedModule, module.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新模块", updatedModule.Name)
		assert.Equal(t, "已更新描述", updatedModule.Description)
		assert.Equal(t, 0, updatedModule.Status)
	})

	t.Run("更新不存在的模块", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的模块",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/modules/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateModule(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestModuleHandler_DeleteModule(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	module := &model.Module{
		Name:        "删除模块",
		Code:        "delete_module",
		Description: "删除模块描述",
		Status:      1,
		Sort:        0,
	}
	db.Create(&module)

	handler := api.NewModuleHandler(db)

	t.Run("删除模块成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/modules/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteModule(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证模块已软删除
		var deletedModule model.Module
		err := db.First(&deletedModule, module.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedModule, module.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedModule.DeletedAt)
	})

	t.Run("删除模块失败-有关联的Bug", func(t *testing.T) {
		// 创建模块和关联的Bug
		module2 := &model.Module{
			Name:  "有关联Bug的模块",
			Code:  "module_with_bug",
			Status: 1,
		}
		db.Create(module2)

		project := CreateTestProject(t, db, "Bug项目")
		user := CreateTestUser(t, db, "buguser", "Bug用户")
		bug := &model.Bug{
			Title:     "测试Bug",
			ProjectID: project.ID,
			CreatorID: user.ID,
			ModuleID:  &module2.ID,
			Status:    "open",
		}
		db.Create(bug)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/modules/2", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}

		handler.DeleteModule(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

