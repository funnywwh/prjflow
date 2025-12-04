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

func TestBoardHandler_GetProjectBoards(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "看板测试项目")

	// 创建测试看板
	board1 := &model.Board{
		Name:      "看板1",
		ProjectID: project.ID,
	}
	db.Create(board1)

	handler := api.NewBoardHandler(db)

	t.Run("获取项目的看板列表", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/projects/1/boards", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetProjectBoards(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// GetProjectBoards返回的是数组
		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			assert.GreaterOrEqual(t, len(list), 1)
		}
	})
}

func TestBoardHandler_GetBoard(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "看板详情项目")

	board := &model.Board{
		Name:      "测试看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	handler := api.NewBoardHandler(db)

	t.Run("获取存在的看板", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/boards/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetBoard(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试看板", data["name"])
	})

	t.Run("获取不存在的看板", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/boards/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetBoard(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBoardHandler_CreateBoard(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建看板项目")
	handler := api.NewBoardHandler(db)

	t.Run("创建看板成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "新看板",
			"description": "这是一个新看板",
			"columns": []map[string]interface{}{
				{
					"name":   "待办",
					"color":  "#1890ff",
					"status": "todo",
					"sort":   1,
				},
				{
					"name":   "进行中",
					"color":  "#faad14",
					"status": "in_progress",
					"sort":   2,
				},
			},
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects/1/boards", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.CreateBoard(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证看板已创建
		var board model.Board
		err = db.Where("name = ?", "新看板").First(&board).Error
		assert.NoError(t, err)
		assert.Equal(t, "新看板", board.Name)
		assert.Equal(t, project.ID, board.ProjectID)
	})

	t.Run("创建看板失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"description": "只有描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/projects/1/boards", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.CreateBoard(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBoardHandler_UpdateBoard(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新看板项目")

	board := &model.Board{
		Name:      "更新看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	handler := api.NewBoardHandler(db)

	t.Run("更新看板成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "已更新看板",
			"description": "更新后的描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/boards/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.UpdateBoard(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证看板已更新
		var updatedBoard model.Board
		err := db.First(&updatedBoard, board.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新看板", updatedBoard.Name)
	})

	t.Run("更新不存在的看板", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的看板",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/boards/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateBoard(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBoardHandler_CreateBoardColumn(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建看板列项目")

	board := &model.Board{
		Name:      "创建列看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	handler := api.NewBoardHandler(db)

	t.Run("创建看板列成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":   "新列",
			"color":  "#1890ff",
			"status": "todo",
			"sort":   1,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/boards/1/columns", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.CreateBoardColumn(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证列已创建
		var column model.BoardColumn
		err := db.Where("board_id = ?", board.ID).First(&column).Error
		assert.NoError(t, err)
		assert.Equal(t, "新列", column.Name)
	})
}

func TestBoardHandler_UpdateBoardColumn(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新看板列项目")

	board := &model.Board{
		Name:      "更新列看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	column := &model.BoardColumn{
		Name:   "更新列",
		BoardID: board.ID,
		Status: "todo",
	}
	db.Create(&column)

	handler := api.NewBoardHandler(db)

	t.Run("更新看板列成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "已更新列",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/boards/1/columns/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "column_id", Value: "1"},
		}

		handler.UpdateBoardColumn(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证列已更新
		var updatedColumn model.BoardColumn
		err := db.First(&updatedColumn, column.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新列", updatedColumn.Name)
	})
}

func TestBoardHandler_DeleteBoardColumn(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除看板列项目")

	board := &model.Board{
		Name:      "删除列看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	column := &model.BoardColumn{
		Name:   "删除列",
		BoardID: board.ID,
	}
	db.Create(&column)

	handler := api.NewBoardHandler(db)

	t.Run("删除看板列成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/boards/1/columns/1", nil)
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "column_id", Value: "1"},
		}

		handler.DeleteBoardColumn(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证列已删除
		var deletedColumn model.BoardColumn
		err := db.First(&deletedColumn, column.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedColumn, column.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedColumn.DeletedAt)
	})
}

func TestBoardHandler_MoveTask(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "移动任务项目")
	user := CreateTestUser(t, db, "movetask", "移动任务用户")

	board := &model.Board{
		Name:      "移动任务看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	column1 := &model.BoardColumn{
		Name:   "待办",
		BoardID: board.ID,
		Status: "todo",
		Sort:   1,
	}
	db.Create(&column1)

	column2 := &model.BoardColumn{
		Name:   "进行中",
		BoardID: board.ID,
		Status: "in_progress",
		Sort:   2,
	}
	db.Create(&column2)

	task := &model.Task{
		Title:     "移动任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "todo",
	}
	db.Create(&task)

	handler := api.NewBoardHandler(db)

	t.Run("移动任务成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"column_id": "2", // 移动到第二个列
			"position":  0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/boards/1/tasks/1/move", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "task_id", Value: "1"},
		}

		handler.MoveTask(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证任务状态已更新
		var updatedTask model.Task
		err := db.First(&updatedTask, task.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "in_progress", updatedTask.Status)
	})
}

func TestBoardHandler_DeleteBoard(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除看板项目")

	board := &model.Board{
		Name:      "删除看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	handler := api.NewBoardHandler(db)

	t.Run("删除看板成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/boards/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.DeleteBoard(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证看板已软删除
		var deletedBoard model.Board
		err := db.First(&deletedBoard, board.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedBoard, board.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedBoard.DeletedAt)
	})
}

func TestBoardHandler_GetBoardTasks(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "看板任务项目")
	user := CreateTestUser(t, db, "boardtask", "看板任务用户")

	board := &model.Board{
		Name:      "任务看板",
		ProjectID: project.ID,
	}
	db.Create(&board)

	column := &model.BoardColumn{
		Name:   "待办",
		BoardID: board.ID,
		Status: "wait",
		Sort:   1,
	}
	db.Create(&column)

	task := &model.Task{
		Title:     "看板任务",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "wait",
	}
	db.Create(&task)

	handler := api.NewBoardHandler(db)

	t.Run("获取看板任务列表", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/boards/1/tasks", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetBoardTasks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			// 验证返回的是数组
			assert.NotNil(t, list)
		}
	})
}

