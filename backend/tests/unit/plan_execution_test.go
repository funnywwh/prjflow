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

func TestPlanExecutionHandler_GetPlanExecutions(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "计划执行测试项目")
	user := CreateTestUser(t, db, "planexecuser", "计划执行用户")

	projectID := project.ID
	plan := &model.Plan{
		Name:      "计划执行测试计划",
		Type:      "project_plan",
		ProjectID: &projectID,
		CreatorID: user.ID,
		Status:    "active",
	}
	db.Create(&plan)

	// 创建测试执行
	execution1 := &model.PlanExecution{
		Name:   "执行1",
		PlanID: plan.ID,
		Status: "pending",
	}
	db.Create(execution1)

	handler := api.NewPlanExecutionHandler(db)

	t.Run("获取计划执行列表", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/plans/1/executions", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.GetPlanExecutions(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// GetPlanExecutions返回的是数组
		data := response["data"]
		if list, ok := data.([]interface{}); ok {
			assert.GreaterOrEqual(t, len(list), 1)
		}
	})
}

func TestPlanExecutionHandler_GetPlanExecution(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "计划执行详情项目")
	user := CreateTestUser(t, db, "planexecdetail", "计划执行详情用户")

	projectID := project.ID
	plan := &model.Plan{
		Name:      "计划执行详情计划",
		Type:      "project_plan",
		ProjectID: &projectID,
		CreatorID: user.ID,
		Status:    "active",
	}
	db.Create(&plan)

	execution := &model.PlanExecution{
		Name:   "测试执行",
		PlanID: plan.ID,
		Status: "pending",
	}
	db.Create(&execution)

	handler := api.NewPlanExecutionHandler(db)

	t.Run("获取存在的执行", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/plans/1/executions/1", nil)
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "execution_id", Value: "1"},
		}

		handler.GetPlanExecution(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试执行", data["name"])
	})

	t.Run("获取不存在的执行", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/plans/1/executions/999", nil)
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "execution_id", Value: "999"},
		}

		handler.GetPlanExecution(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestPlanExecutionHandler_CreatePlanExecution(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建计划执行项目")
	user := CreateTestUser(t, db, "createplanexec", "创建计划执行用户")

	projectID := project.ID
	plan := &model.Plan{
		Name:      "创建计划执行计划",
		Type:      "project_plan",
		ProjectID: &projectID,
		CreatorID: user.ID,
		Status:    "active",
	}
	db.Create(&plan)

	handler := api.NewPlanExecutionHandler(db)

	t.Run("创建计划执行成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":        "新执行",
			"description": "这是一个新执行",
			"status":      "pending",
			"progress":    0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/plans/1/executions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.CreatePlanExecution(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证执行已创建
		var execution model.PlanExecution
		err = db.Where("name = ?", "新执行").First(&execution).Error
		assert.NoError(t, err)
		assert.Equal(t, "新执行", execution.Name)
		assert.Equal(t, plan.ID, execution.PlanID)
	})

	t.Run("创建计划执行失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"description": "只有描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/plans/1/executions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		handler.CreatePlanExecution(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建计划执行失败-计划不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "新执行",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/plans/999/executions", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.CreatePlanExecution(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestPlanExecutionHandler_UpdatePlanExecution(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新计划执行项目")
	user := CreateTestUser(t, db, "updateplanexec", "更新计划执行用户")

	projectID := project.ID
	plan := &model.Plan{
		Name:      "更新计划执行计划",
		Type:      "project_plan",
		ProjectID: &projectID,
		CreatorID: user.ID,
		Status:    "active",
	}
	db.Create(&plan)

	execution := &model.PlanExecution{
		Name:   "更新执行",
		PlanID: plan.ID,
		Status: "pending",
	}
	db.Create(&execution)

	handler := api.NewPlanExecutionHandler(db)

	t.Run("更新计划执行成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name":     "已更新执行",
			"status":   "in_progress",
			"progress": 50,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/plans/1/executions/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "execution_id", Value: "1"},
		}

		handler.UpdatePlanExecution(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证执行已更新
		var updatedExecution model.PlanExecution
		err := db.First(&updatedExecution, execution.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新执行", updatedExecution.Name)
		assert.Equal(t, "in_progress", updatedExecution.Status)
		assert.Equal(t, 50, updatedExecution.Progress)
	})

	t.Run("更新不存在的执行", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"name": "不存在的执行",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/plans/1/executions/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "execution_id", Value: "999"},
		}

		handler.UpdatePlanExecution(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestPlanExecutionHandler_DeletePlanExecution(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除计划执行项目")
	user := CreateTestUser(t, db, "deleteplanexec", "删除计划执行用户")

	projectID := project.ID
	plan := &model.Plan{
		Name:      "删除计划执行计划",
		Type:      "project_plan",
		ProjectID: &projectID,
		CreatorID: user.ID,
		Status:    "active",
	}
	db.Create(&plan)

	execution := &model.PlanExecution{
		Name:   "删除执行",
		PlanID: plan.ID,
		Status: "pending",
	}
	db.Create(&execution)

	handler := api.NewPlanExecutionHandler(db)

	t.Run("删除计划执行成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/plans/1/executions/1", nil)
		c.Params = gin.Params{
			gin.Param{Key: "id", Value: "1"},
			gin.Param{Key: "execution_id", Value: "1"},
		}

		handler.DeletePlanExecution(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证执行已软删除
		var deletedExecution model.PlanExecution
		err := db.First(&deletedExecution, execution.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedExecution, execution.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedExecution.DeletedAt)
	})
}

