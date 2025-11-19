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

func TestTestCaseHandler_GetTestCases(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "测试单测试项目")
	user := CreateTestUser(t, db, "testcaseuser", "测试单用户")

	// 创建测试单
	testCase1 := &model.TestCase{
		Name:      "测试单1",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "pending",
		Types:     model.StringArray{"functional"},
	}
	db.Create(testCase1)

	handler := api.NewTestCaseHandler(db)

	t.Run("获取所有测试单", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test-cases", nil)

		handler.GetTestCases(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("搜索测试单", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test-cases?keyword=测试单1", nil)

		handler.GetTestCases(c)

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
	_ = user
}

func TestTestCaseHandler_GetTestCase(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "测试单详情项目")
	user := CreateTestUser(t, db, "testcasedetail", "测试单详情用户")

	testCase := &model.TestCase{
		Name:      "测试测试单",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "pending",
		Types:     model.StringArray{"functional"},
	}
	db.Create(&testCase)

	handler := api.NewTestCaseHandler(db)

	t.Run("获取存在的测试单", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test-cases/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetTestCase(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试测试单", data["name"])
	})

	t.Run("获取不存在的测试单", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test-cases/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetTestCase(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTestCaseHandler_CreateTestCase(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建测试单项目")
	user := CreateTestUser(t, db, "createtestcase", "创建测试单用户")
	handler := api.NewTestCaseHandler(db)

	t.Run("创建测试单成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id（CreateTestCase需要）
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"name":        "新测试单",
			"description": "这是一个新测试单",
			"test_steps":  "1. 步骤1\n2. 步骤2",
			"types":       []string{"functional", "performance"},
			"status":      "pending",
			"project_id":  project.ID,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/test-cases", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTestCase(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证测试单已创建
		var testCase model.TestCase
		err = db.Where("name = ?", "新测试单").First(&testCase).Error
		assert.NoError(t, err)
		assert.Equal(t, "新测试单", testCase.Name)
		assert.Equal(t, project.ID, testCase.ProjectID)
		assert.Equal(t, 2, len(testCase.Types))
	})

	t.Run("创建测试单失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"description": "只有描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/test-cases", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTestCase(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建测试单失败-项目不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"name":       "新测试单",
			"project_id": 999,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/test-cases", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTestCase(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTestCaseHandler_UpdateTestCase(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新测试单项目")
	user := CreateTestUser(t, db, "updatetestcase", "更新测试单用户")

	testCase := &model.TestCase{
		Name:      "更新测试单",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "pending",
		Types:     model.StringArray{"functional"},
	}
	db.Create(&testCase)

	handler := api.NewTestCaseHandler(db)

	t.Run("更新测试单成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":   "已更新测试单",
			"status": "running",
			"types":  []string{"functional", "security"},
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/test-cases/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateTestCase(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证测试单已更新
		var updatedTestCase model.TestCase
		err := db.First(&updatedTestCase, testCase.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新测试单", updatedTestCase.Name)
		assert.Equal(t, "running", updatedTestCase.Status)
		assert.Equal(t, 2, len(updatedTestCase.Types))
	})

	t.Run("更新不存在的测试单", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的测试单",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/test-cases/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateTestCase(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTestCaseHandler_DeleteTestCase(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除测试单项目")
	user := CreateTestUser(t, db, "deletetestcase", "删除测试单用户")

	testCase := &model.TestCase{
		Name:      "删除测试单",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "pending",
	}
	db.Create(&testCase)

	handler := api.NewTestCaseHandler(db)

	t.Run("删除测试单成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/test-cases/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteTestCase(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证测试单已软删除
		var deletedTestCase model.TestCase
		err := db.First(&deletedTestCase, testCase.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedTestCase, testCase.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedTestCase.DeletedAt)
	})
}

