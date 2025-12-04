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

func TestTaskHandler_GetTasks(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "任务测试项目")
	user := CreateTestUser(t, db, "taskuser", "任务用户")
	adminUser := CreateTestAdminUser(t, db, "admintask", "管理员任务用户")
	otherUser := CreateTestUser(t, db, "othertask", "其他任务用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	// 创建测试任务
	task1 := &model.Task{
		Title:     "任务1",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait", // 使用有效的状态值
		Priority:  "high",
	}
	db.Create(task1)

	// 创建另一个项目的任务
	project2 := CreateTestProject(t, db, "任务测试项目2")
	task2 := &model.Task{
		Title:     "任务2",
		ProjectID: project2.ID,
		CreatorID: otherUser.ID,
		Status:    "wait", // 使用有效的状态值
		Priority:  "high",
	}
	db.Create(task2)

	handler := api.NewTaskHandler(db)

	t.Run("管理员可以获取所有任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetTasks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 管理员应该能看到所有任务
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("普通用户只能看到自己创建或参与的任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetTasks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 用户创建了任务1且是项目成员，应该能看到任务1
		assert.Equal(t, 1, len(list))
	})

	t.Run("搜索任务-管理员", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks?keyword=任务1", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetTasks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("搜索任务-普通用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks?keyword=任务1", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetTasks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 用户应该能看到自己创建的任务1
		assert.Equal(t, 1, len(list))
	})

	_ = project
	_ = user
}

func TestTaskHandler_GetTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "任务详情项目")
	user := CreateTestUser(t, db, "taskdetail", "任务详情用户")
	adminUser := CreateTestAdminUser(t, db, "admintask2", "管理员任务用户2")
	otherUser := CreateTestUser(t, db, "othertask2", "其他任务用户2")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	task := &model.Task{
		Title:     "测试任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait", // 使用有效的状态值
		Priority:  "high",
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("管理员可以获取任何任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d", task.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试任务", data["title"])
	})

	t.Run("创建者可以获取任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d", task.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})

	t.Run("非项目成员不能获取任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d", task.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"developer"})

		handler.GetTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 应该返回403或code不为200
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("获取不存在的任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/tasks/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTaskHandler_CreateTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建任务项目")
	user := CreateTestUser(t, db, "createtask", "创建任务用户")
	handler := api.NewTaskHandler(db)

	t.Run("创建任务成功-项目成员", func(t *testing.T) {
		// 添加用户到项目
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id（CreateTask需要）
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"title":          "新任务",
			"description":    "这是一个新任务",
			"status":         "wait", // 使用有效的状态值
			"priority":       "high",
			"project_id":     project.ID,
			"estimated_hours": 4.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证任务已创建
		var task model.Task
		err = db.Where("title = ?", "新任务").First(&task).Error
		assert.NoError(t, err)
		assert.Equal(t, "新任务", task.Title)
		assert.Equal(t, project.ID, task.ProjectID)
	})

	t.Run("创建任务失败-非项目成员", func(t *testing.T) {
		otherUser := CreateTestUser(t, db, "othercreatetask", "其他创建任务用户")
		otherProject := CreateTestProject(t, db, "其他任务项目")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"title":          "新任务",
			"description":    "这是一个新任务",
			"status":         "wait", // 使用有效的状态值
			"priority":       "high",
			"project_id":     otherProject.ID,
			"estimated_hours": 4.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 应该返回403或code不为200
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建任务失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"description": "只有描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建任务失败-项目不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"title":      "新任务",
			"project_id": 999,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新任务项目")
	user := CreateTestUser(t, db, "updatetask", "更新任务用户")

	task := &model.Task{
		Title:     "更新任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait", // 使用有效的状态值
		Priority:  "high",
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("更新任务成功", func(t *testing.T) {
		// 添加用户到项目（作为项目成员）
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"title":     "已更新任务",
			"status":    "doing", // 使用有效的状态值
			"priority":  "medium",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/tasks/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证任务已更新
		var updatedTask model.Task
		err := db.First(&updatedTask, task.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新任务", updatedTask.Title)
		assert.Equal(t, "in_progress", updatedTask.Status)
	})

	t.Run("更新不存在的任务", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"title": "不存在的任务",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/tasks/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateTask(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除任务项目")
	user := CreateTestUser(t, db, "deletetask", "删除任务用户")

	task := &model.Task{
		Title:     "删除任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait", // 使用有效的状态值
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("删除任务成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/tasks/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证任务已软删除
		var deletedTask model.Task
		err := db.First(&deletedTask, task.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedTask, task.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedTask.DeletedAt)
	})
}

func TestTaskHandler_UpdateTaskStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新任务状态项目")
	user := CreateTestUser(t, db, "updatestatus", "更新状态用户")

	task := &model.Task{
		Title:     "更新状态任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait",
	}
	db.Create(&task)

	// 添加用户到项目，以便有权限更新任务
	AddUserToProject(t, db, user.ID, project.ID, "member")

	handler := api.NewTaskHandler(db)

	t.Run("更新任务状态成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"status": "doing",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tasks/%d/status", task.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.UpdateTaskStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证任务状态已更新
		var updatedTask model.Task
		err := db.First(&updatedTask, task.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "doing", updatedTask.Status)
	})

	t.Run("更新任务状态失败-无效状态", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "invalid_status",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tasks/%d/status", task.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.UpdateTaskStatus(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTaskHandler_UpdateTaskProgress(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新任务进度项目")
	user := CreateTestUser(t, db, "updateprogress", "更新进度用户")

	estimatedHours := 10.0
	task := &model.Task{
		Title:          "更新进度任务",
		ProjectID:      project.ID,
		CreatorID:      user.ID,
		Status:         "doing",
		EstimatedHours: &estimatedHours,
		Progress:       0,
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("更新任务进度成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"progress": 50,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tasks/%d/progress", task.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.UpdateTaskProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证任务进度已更新
		var updatedTask model.Task
		err := db.First(&updatedTask, task.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, 50, updatedTask.Progress)
	})

	t.Run("更新任务进度失败-无效进度值", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"progress": 150, // 超过100
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tasks/%d/progress", task.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.UpdateTaskProgress(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestTaskHandler_GetTaskHistory(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "任务历史项目")
	user := CreateTestUser(t, db, "taskhistory", "任务历史用户")

	task := &model.Task{
		Title:     "历史任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "doing",
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("获取任务历史", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tasks/%d/history", task.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.GetTaskHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			// 历史记录可能为空，所以只验证返回的是数组
			assert.NotNil(t, list)
		}
	})
}

func TestTaskHandler_AddTaskHistoryNote(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "任务历史备注项目")
	user := CreateTestUser(t, db, "tasknote", "任务备注用户")

	task := &model.Task{
		Title:     "备注任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "doing",
	}
	db.Create(&task)

	handler := api.NewTaskHandler(db)

	t.Run("添加任务历史备注成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"note": "这是一个历史备注",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/history", task.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", task.ID)}}

		handler.AddTaskHistoryNote(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

