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

func TestProjectHandler_GetProjects(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试数据
	project1 := CreateTestProject(t, db, "项目1")
	project2 := CreateTestProject(t, db, "项目2")

	handler := api.NewProjectHandler(db)

	t.Run("获取所有项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects", nil)

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("搜索项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects?keyword=项目1", nil)

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("分页查询", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects?page=1&page_size=1", nil)

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.Equal(t, 1, len(list))
	})

	// 使用测试数据（避免未使用变量警告）
	_ = project1
	_ = project2
}

func TestProjectHandler_GetProject(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	_ = CreateTestProject(t, db, "测试项目")
	handler := api.NewProjectHandler(db)

	t.Run("获取存在的项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		projectData := data["project"].(map[string]interface{})
		assert.Equal(t, "测试项目", projectData["name"])
		assert.NotNil(t, data["statistics"])
	})

	t.Run("获取不存在的项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetProject(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestProjectHandler_CreateProject(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	handler := api.NewProjectHandler(db)

	t.Run("创建项目成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "新项目",
			"code":        "NEW001",
			"description": "这是一个新项目",
			"status":      1,
			"tags":        []string{"重要", "紧急"},
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证项目已创建
		var project model.Project
		err = db.Where("name = ?", "新项目").First(&project).Error
		assert.NoError(t, err)
		assert.Equal(t, "新项目", project.Name)
		assert.Equal(t, "NEW001", project.Code)
		assert.Equal(t, 2, len(project.Tags))
	})

	t.Run("创建项目失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"code": "TEST001",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateProject(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新项目")
	handler := api.NewProjectHandler(db)

	t.Run("更新项目成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "已更新项目",
			"description": "更新后的描述",
			"tags":        []string{"已更新"},
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/projects/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证项目已更新
		var updatedProject model.Project
		err := db.First(&updatedProject, project.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新项目", updatedProject.Name)
		assert.Equal(t, "更新后的描述", updatedProject.Description)
		assert.Equal(t, 1, len(updatedProject.Tags))
	})

	t.Run("更新不存在的项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的项目",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/projects/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateProject(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除项目")
	handler := api.NewProjectHandler(db)

	t.Run("删除项目成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/projects/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证项目已软删除
		var deletedProject model.Project
		err := db.First(&deletedProject, project.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedProject, project.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedProject.DeletedAt)
	})
}

func TestProjectHandler_GetProjectGantt(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "甘特图项目")
	user := CreateTestUser(t, db, "ganttuser", "甘特图用户")

	// 创建测试任务
	task := &model.Task{
		Title:     "甘特图任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "todo",
	}
	db.Create(&task)

	handler := api.NewProjectHandler(db)

	t.Run("获取项目甘特图数据", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/1/gantt", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetProjectGantt(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["tasks"])
	})

	t.Run("获取不存在的项目甘特图", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/999/gantt", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetProjectGantt(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProjectHandler_GetProjectProgress(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	_ = CreateTestProject(t, db, "进度跟踪项目")
	handler := api.NewProjectHandler(db)

	t.Run("获取项目进度跟踪数据", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/1/progress", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetProjectProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["statistics"])
		assert.NotNil(t, data["task_progress_trend"])
		assert.NotNil(t, data["task_status_distribution"])
	})

	t.Run("获取不存在的项目进度", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/999/progress", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetProjectProgress(c)

		// GetProjectProgress在项目不存在时返回错误消息（utils.Error返回200状态码但code不为200）
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// utils.Error返回200状态码，但code字段不为200
		assert.True(t, response["code"] != nil && response["code"] != float64(200))
	})
}

func TestProjectHandler_GetProjectStatistics(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	_ = CreateTestProject(t, db, "统计项目")
	handler := api.NewProjectHandler(db)

	t.Run("获取项目统计", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/1/statistics", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetProjectStatistics(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["total_tasks"])
		assert.NotNil(t, data["total_bugs"])
		assert.NotNil(t, data["total_requirements"])
		assert.NotNil(t, data["total_members"])
	})
}

