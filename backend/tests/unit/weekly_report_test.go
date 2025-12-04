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

func TestReportHandler_GetWeeklyReports(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "weeklyreportuser", "周报用户")
	_ = CreateTestAdminUser(t, db, "adminweeklyreport", "管理员周报用户")

	// 创建测试周报
	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 6)
	report1 := &model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      "本周工作总结",
		NextWeekPlan: "下周工作计划",
		Status:       "draft",
		UserID:       user.ID,
	}
	db.Create(report1)

	handler := api.NewReportHandler(db)

	t.Run("获取周报列表-普通用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/weekly-reports", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWeeklyReports(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("获取周报列表-管理员", func(t *testing.T) {
		adminUser := CreateTestAdminUser(t, db, "adminweeklyreport2", "管理员周报用户2")
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/weekly-reports", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetWeeklyReports(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})

	t.Run("按状态筛选周报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/weekly-reports?status=draft", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWeeklyReports(c)

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

func TestReportHandler_GetWeeklyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "weeklyreportdetail", "周报详情用户")
	_ = CreateTestAdminUser(t, db, "adminweeklyreportdetail", "管理员周报详情用户")

	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 6)
	report := &model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      "周报详情内容",
		NextWeekPlan: "下周计划",
		Status:       "draft",
		UserID:       user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("获取存在的周报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/weekly-reports/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWeeklyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "周报详情内容", data["summary"])
	})

	t.Run("获取不存在的周报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/weekly-reports/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWeeklyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_CreateWeeklyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "createweeklyreport", "创建周报用户")
	handler := api.NewReportHandler(db)

	t.Run("创建周报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		weekStart := time.Now()
		weekEnd := weekStart.AddDate(0, 0, 6)
		reqBody := map[string]interface{}{
			"week_start":    weekStart.Format("2006-01-02"),
			"week_end":      weekEnd.Format("2006-01-02"),
			"summary":       "本周工作总结",
			"next_week_plan": "下周工作计划",
			"status":        "draft",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/weekly-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateWeeklyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证周报已创建
		var report model.WeeklyReport
		err = db.Where("user_id = ? AND week_start = ?", user.ID, weekStart.Format("2006-01-02")).First(&report).Error
		assert.NoError(t, err)
		assert.Equal(t, "本周工作总结", report.Summary)
	})

	t.Run("创建周报失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"summary": "只有总结",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/weekly-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateWeeklyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建周报失败-日期格式错误", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"week_start": "invalid-date",
			"week_end":   "invalid-date",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/weekly-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateWeeklyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建周报失败-开始日期晚于结束日期", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		weekStart := time.Now()
		weekEnd := weekStart.AddDate(0, 0, -1) // 结束日期早于开始日期
		reqBody := map[string]interface{}{
			"week_start": weekStart.Format("2006-01-02"),
			"week_end":   weekEnd.Format("2006-01-02"),
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/weekly-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateWeeklyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_UpdateWeeklyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "updateweeklyreport", "更新周报用户")
	otherUser := CreateTestUser(t, db, "otherweeklyreport", "其他周报用户")

	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 6)
	report := &model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      "原始总结",
		NextWeekPlan: "原始计划",
		Status:       "draft",
		UserID:       user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("更新周报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"summary":       "更新后的总结",
			"next_week_plan": "更新后的计划",
			"status":        "submitted",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/weekly-reports/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateWeeklyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证周报已更新
		var updatedReport model.WeeklyReport
		err := db.First(&updatedReport, report.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "更新后的总结", updatedReport.Summary)
		assert.Equal(t, "更新后的计划", updatedReport.NextWeekPlan)
		assert.Equal(t, "submitted", updatedReport.Status)
	})

	t.Run("更新周报失败-无权限", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"summary": "尝试更新他人周报",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/weekly-reports/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateWeeklyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_DeleteWeeklyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "deleteweeklyreport", "删除周报用户")
	_ = CreateTestAdminUser(t, db, "admindeleteweeklyreport", "管理员删除周报用户")

	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 6)
	report := &model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      "要删除的周报",
		NextWeekPlan: "计划",
		Status:       "draft",
		UserID:       user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("删除周报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/weekly-reports/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.DeleteWeeklyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证周报已软删除
		var deletedReport model.WeeklyReport
		err := db.First(&deletedReport, report.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）
	})
}

func TestReportHandler_UpdateWeeklyReportStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "updatestatusweeklyreport", "更新状态周报用户")

	weekStart := time.Now()
	weekEnd := weekStart.AddDate(0, 0, 6)
	report := &model.WeeklyReport{
		WeekStart:    weekStart,
		WeekEnd:      weekEnd,
		Summary:      "更新状态的周报",
		NextWeekPlan: "计划",
		Status:       "draft",
		UserID:       user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("更新周报状态成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "submitted",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/weekly-reports/1/status", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateWeeklyReportStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证状态已更新
		var updatedReport model.WeeklyReport
		err := db.First(&updatedReport, report.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "submitted", updatedReport.Status)
	})

	t.Run("更新周报状态失败-无效状态", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "invalid_status",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/weekly-reports/1/status", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateWeeklyReportStatus(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["code"] != nil && response["code"] != float64(200))
	})
}

