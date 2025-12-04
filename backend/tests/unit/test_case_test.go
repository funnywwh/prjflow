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
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id（CreateTestCase需要）
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"name":        "新测试单",
			"description": "这是一个新测试单",
			"test_steps":  "1. 步骤1\n2. 步骤2",
			"types":       []string{"functional", "performance"},
			"status":      "wait", // 使用有效的状态值
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
		Status:    "wait", // 使用有效的状态值
		Types:     model.StringArray{"functional"},
	}
	db.Create(&testCase)

	handler := api.NewTestCaseHandler(db)

	t.Run("更新测试单成功", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"name":   "已更新测试单",
			"status": "normal", // 使用有效的状态值
			"types":  []string{"functional", "security"},
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/test-cases/%d", testCase.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", testCase.ID)}}

		handler.UpdateTestCase(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证测试单已更新
		var updatedTestCase model.TestCase
		err := db.First(&updatedTestCase, testCase.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新测试单", updatedTestCase.Name)
		assert.Equal(t, "normal", updatedTestCase.Status) // 使用有效的状态值
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

func TestTestCaseHandler_UpdateTestCaseStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新测试单状态项目")
	user := CreateTestUser(t, db, "updatestatus", "更新状态用户")

	testCase := &model.TestCase{
		Name:      "更新状态测试单",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "pending",
	}
	db.Create(&testCase)

	handler := api.NewTestCaseHandler(db)

	t.Run("更新测试单状态成功", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"status": "normal", // 使用有效的状态值
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/test-cases/%d/status", testCase.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", testCase.ID)}}

		handler.UpdateTestCaseStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证状态已更新
		var updatedTestCase model.TestCase
		err := db.First(&updatedTestCase, testCase.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "normal", updatedTestCase.Status) // 使用有效的状态值
	})

	t.Run("更新测试单状态失败-无效状态", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "invalid_status",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/test-cases/1/status", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateTestCaseStatus(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["code"] != nil && response["code"] != float64(200))
	})
}

func TestTestCaseHandler_GetTestCaseStatistics(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "测试统计项目")
	user := CreateTestUser(t, db, "teststat", "测试统计用户")

	// 创建测试数据
	testCase1 := &model.TestCase{
		Name:      "测试1",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "passed",
		Types:     model.StringArray{"功能测试"},
	}
	db.Create(testCase1)

	testCase2 := &model.TestCase{
		Name:      "测试2",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "failed",
		Types:     model.StringArray{"性能测试"},
	}
	db.Create(testCase2)

	handler := api.NewTestCaseHandler(db)

	t.Run("获取测试单统计", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test-cases/statistics", nil)

		handler.GetTestCaseStatistics(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["total"])
		assert.NotNil(t, data["pass_rate"])
		assert.NotNil(t, data["fail_rate"])
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

