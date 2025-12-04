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

func TestProjectHandler_GetProjects(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 创建测试数据
	project1 := CreateTestProject(t, db, "项目1")
	project2 := CreateTestProject(t, db, "项目2")
	user := CreateTestUser(t, db, "projectuser", "项目用户")
	adminUser := CreateTestAdminUser(t, db, "adminuser", "管理员用户")

	// 添加用户到项目1
	AddUserToProject(t, db, user.ID, project1.ID, "member")

	handler := api.NewProjectHandler(db)

	t.Run("管理员可以获取所有项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 管理员应该能看到所有项目
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("普通用户只能看到自己参与的项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 普通用户只能看到自己参与的项目1
		assert.Equal(t, 1, len(list))
	})

	t.Run("未登录用户看不到任何项目", func(t *testing.T) {
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
		// 未登录用户应该看不到任何项目
		assert.Equal(t, 0, len(list))
	})

	t.Run("搜索项目-管理员", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects?keyword=项目1", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

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

	t.Run("搜索项目-普通用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects?keyword=项目1", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 用户参与项目1，应该能看到
		assert.Equal(t, 1, len(list))
	})

	t.Run("分页查询-管理员", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects?page=1&page_size=1", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

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

	t.Run("按单个标签搜索项目", func(t *testing.T) {
		// 创建管理员用户
		adminUser := CreateTestAdminUser(t, db, "admintag", "管理员标签用户")

		// 创建标签
		tag1 := CreateTestTag(t, db, "前端")
		tag2 := CreateTestTag(t, db, "重要")

		// 创建带标签的项目
		projectWithTag := CreateTestProject(t, db, "标签项目")
		db.Model(&projectWithTag).Association("Tags").Append([]*model.Tag{tag1, tag2})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// 使用标签ID进行搜索
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects?tag=%d", tag1.ID), nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

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

	t.Run("按多个标签搜索项目（OR逻辑）", func(t *testing.T) {
		// 创建管理员用户
		adminUser := CreateTestAdminUser(t, db, "admintag2", "管理员标签用户2")

		// 创建标签
		tag1 := CreateTestTag(t, db, "前端")
		tag2 := CreateTestTag(t, db, "重要")
		tag3 := CreateTestTag(t, db, "后端")
		tag4 := CreateTestTag(t, db, "紧急")

		// 创建带不同标签的项目
		project1 := CreateTestProject(t, db, "标签项目1")
		db.Model(&project1).Association("Tags").Append([]*model.Tag{tag1, tag2})

		project2 := CreateTestProject(t, db, "标签项目2")
		db.Model(&project2).Association("Tags").Append([]*model.Tag{tag3, tag4})

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// 使用QueryArray方式传递多个标签ID
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects?tags=%d&tags=%d", tag1.ID, tag3.ID), nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetProjects(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 应该返回包含"前端"或"后端"标签的项目，至少2个
		assert.GreaterOrEqual(t, len(list), 2)
	})

	// 使用测试数据（避免未使用变量警告）
	_ = project1
	_ = project2
}

func TestProjectHandler_GetProject(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "测试项目")
	user := CreateTestUser(t, db, "getprojectuser", "获取项目用户")
	adminUser := CreateTestAdminUser(t, db, "adminuser2", "管理员用户2")
	otherUser := CreateTestUser(t, db, "otheruser", "其他用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	handler := api.NewProjectHandler(db)

	t.Run("管理员可以获取任何项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

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

	t.Run("项目成员可以获取项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})

	t.Run("非项目成员不能获取项目", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"developer"})

		handler.GetProject(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 应该返回403或code不为200
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
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
		// 创建用户（作为项目创建者）
		user := CreateTestUser(t, db, "createproject", "创建项目用户")

		// 先创建标签
		tag1 := CreateTestTag(t, db, "重要")
		tag2 := CreateTestTag(t, db, "紧急")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"name":        "新项目",
			"code":        "NEW001",
			"description": "这是一个新项目",
			"status":      "doing", // 使用字符串状态值
			"tag_ids":     []uint{tag1.ID, tag2.ID}, // 使用标签ID数组
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
		err = db.Preload("Tags").Where("name = ?", "新项目").First(&project).Error
		assert.NoError(t, err)
		assert.Equal(t, "新项目", project.Name)
		assert.Equal(t, "NEW001", project.Code)
		assert.Equal(t, 2, len(project.Tags))
		// 验证标签关联正确
		tagIDs := make([]uint, len(project.Tags))
		for i, tag := range project.Tags {
			tagIDs[i] = tag.ID
		}
		assert.Contains(t, tagIDs, tag1.ID)
		assert.Contains(t, tagIDs, tag2.ID)
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
		// 创建用户并添加到项目（作为项目成员）
		user := CreateTestUser(t, db, "updateproject", "更新项目用户")
		AddUserToProject(t, db, user.ID, project.ID, "member")

		// 先创建标签
		tag := CreateTestTag(t, db, "已更新")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"name":        "已更新项目",
			"description": "更新后的描述",
			"tag_ids":     []uint{tag.ID}, // 使用标签ID数组
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d", project.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}

		handler.UpdateProject(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证项目已更新
		var updatedProject model.Project
		err := db.Preload("Tags").First(&updatedProject, project.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新项目", updatedProject.Name)
		assert.Equal(t, "更新后的描述", updatedProject.Description)
		assert.Equal(t, 1, len(updatedProject.Tags))
		if len(updatedProject.Tags) > 0 {
			assert.Equal(t, tag.ID, updatedProject.Tags[0].ID)
		}
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
		// 创建用户并添加到项目（作为项目成员）
		user := CreateTestUser(t, db, "deleteproject", "删除项目用户")
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%d", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}

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
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/gantt", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}

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

		// GetProjectGantt在项目不存在时返回错误消息（utils.Error返回200状态码但code不为200）
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// utils.Error返回200状态码，但code字段不为200
		assert.True(t, response["code"] != nil && response["code"] != float64(200))
	})
}

func TestProjectHandler_GetProjectProgress(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "进度跟踪项目")
	user := CreateTestUser(t, db, "progressuser", "进度跟踪用户")
	handler := api.NewProjectHandler(db)

	t.Run("获取项目进度跟踪数据", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/progress", project.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)}}

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

func TestProjectHandler_AddProjectMembers(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "添加成员项目")
	user1 := CreateTestUser(t, db, "member1", "成员1")
	user2 := CreateTestUser(t, db, "member2", "成员2")

	handler := api.NewProjectHandler(db)

	t.Run("添加项目成员成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"user_ids": []uint{user1.ID, user2.ID},
			"role":     "developer",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects/1/members", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.AddProjectMembers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证成员已添加
		var members []model.ProjectMember
		err := db.Where("project_id = ?", project.ID).Find(&members).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(members), 2)
	})
}

func TestProjectHandler_UpdateProjectMember(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新成员项目")
	user := CreateTestUser(t, db, "updatemember", "更新成员用户")

	// 创建项目成员
	member := &model.ProjectMember{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      "developer",
	}
	db.Create(&member)

	handler := api.NewProjectHandler(db)

	t.Run("更新项目成员成功", func(t *testing.T) {
		// 创建另一个用户作为项目成员（用于更新操作）
		updaterUser := CreateTestUser(t, db, "updatemember2", "更新成员用户2")
		AddUserToProject(t, db, updaterUser.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", updaterUser.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"role": "manager",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d/members/%d", project.ID, member.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: fmt.Sprintf("%d", project.ID)},
			gin.Param{Key: "member_id", Value: fmt.Sprintf("%d", member.ID)},
		}

		handler.UpdateProjectMember(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证成员角色已更新
		var updatedMember model.ProjectMember
		err := db.First(&updatedMember, member.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "manager", updatedMember.Role)
	})
}

func TestProjectHandler_RemoveProjectMember(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "移除成员项目")
	user := CreateTestUser(t, db, "removemember", "移除成员用户")

	// 创建项目成员
	member := &model.ProjectMember{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      "developer",
	}
	db.Create(&member)

	handler := api.NewProjectHandler(db)

	t.Run("移除项目成员成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/projects/1/members/1", nil)
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "member_id", Value: "1"},
		}

		handler.RemoveProjectMember(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证成员已删除
		var deletedMember model.ProjectMember
		err := db.First(&deletedMember, member.ID).Error
		assert.Error(t, err) // 应该找不到
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
