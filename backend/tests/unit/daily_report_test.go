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

func TestReportHandler_GetDailyReports(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "dailyreportuser", "日报用户")
	_ = CreateTestAdminUser(t, db, "admindailyreport", "管理员日报用户")

	// 创建测试日报
	report1 := &model.DailyReport{
		Date:    time.Now(),
		Content: "今日工作内容",
		Status:  "draft",
		UserID:  user.ID,
	}
	db.Create(report1)

	handler := api.NewReportHandler(db)

	t.Run("获取日报列表-普通用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/daily-reports", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetDailyReports(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("获取日报列表-管理员", func(t *testing.T) {
		adminUser := CreateTestAdminUser(t, db, "admindailyreport2", "管理员日报用户2")
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/daily-reports", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetDailyReports(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})

	t.Run("按状态筛选日报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/daily-reports?status=draft", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetDailyReports(c)

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

func TestReportHandler_GetDailyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "dailyreportdetail", "日报详情用户")
	_ = CreateTestAdminUser(t, db, "admindailyreportdetail", "管理员日报详情用户")

	report := &model.DailyReport{
		Date:    time.Now(),
		Content: "日报详情内容",
		Status:  "draft",
		UserID:  user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("获取存在的日报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/daily-reports/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetDailyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "日报详情内容", data["content"])
	})

	t.Run("获取不存在的日报", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/daily-reports/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetDailyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_CreateDailyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "createdailyreport", "创建日报用户")
	handler := api.NewReportHandler(db)

	t.Run("创建日报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"date":    time.Now().Format("2006-01-02"),
			"content": "今日工作内容",
			"status":  "draft",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/daily-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateDailyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证日报已创建
		var report model.DailyReport
		err = db.Where("user_id = ? AND date = ?", user.ID, reqBody["date"]).First(&report).Error
		assert.NoError(t, err)
		assert.Equal(t, "今日工作内容", report.Content)
	})

	t.Run("创建日报失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"content": "只有内容",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/daily-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateDailyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建日报失败-日期已存在", func(t *testing.T) {
		// 先创建一个日报
		existingReport := &model.DailyReport{
			Date:    time.Now(),
			Content: "已存在的日报",
			Status:  "draft",
			UserID:  user.ID,
		}
		db.Create(existingReport)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"date":    existingReport.Date.Format("2006-01-02"),
			"content": "重复日期的日报",
			"status":  "draft",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/daily-reports", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.CreateDailyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_UpdateDailyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "updatedailyreport", "更新日报用户")
	otherUser := CreateTestUser(t, db, "otherdailyreport", "其他日报用户")

	report := &model.DailyReport{
		Date:    time.Now(),
		Content: "原始内容",
		Status:  "draft",
		UserID:  user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("更新日报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"content": "更新后的内容",
			"status":  "submitted",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/daily-reports/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateDailyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证日报已更新
		var updatedReport model.DailyReport
		err := db.First(&updatedReport, report.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "更新后的内容", updatedReport.Content)
		assert.Equal(t, "submitted", updatedReport.Status)
	})

	t.Run("更新日报失败-无权限", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"content": "尝试更新他人日报",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/daily-reports/1", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateDailyReport(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestReportHandler_DeleteDailyReport(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "deletedailyreport", "删除日报用户")
	_ = CreateTestAdminUser(t, db, "admindeletedailyreport", "管理员删除日报用户")

	report := &model.DailyReport{
		Date:    time.Now(),
		Content: "要删除的日报",
		Status:  "draft",
		UserID:  user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("删除日报成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/daily-reports/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.DeleteDailyReport(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证日报已软删除
		var deletedReport model.DailyReport
		err := db.First(&deletedReport, report.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）
	})
}

func TestReportHandler_UpdateDailyReportStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	user := CreateTestUser(t, db, "updatestatusdailyreport", "更新状态日报用户")

	report := &model.DailyReport{
		Date:    time.Now(),
		Content: "更新状态的日报",
		Status:  "draft",
		UserID:  user.ID,
	}
	db.Create(&report)

	handler := api.NewReportHandler(db)

	t.Run("更新日报状态成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "submitted",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/daily-reports/1/status", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateDailyReportStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证状态已更新
		var updatedReport model.DailyReport
		err := db.First(&updatedReport, report.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "submitted", updatedReport.Status)
	})

	t.Run("更新日报状态失败-无效状态", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"status": "invalid_status",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/daily-reports/1/status", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.UpdateDailyReportStatus(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["code"] != nil && response["code"] != float64(200))
	})
}

func TestReportHandler_GetWorkSummary(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "工作汇总项目")
	user := CreateTestUser(t, db, "worksummaryuser", "工作汇总用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	handler := api.NewReportHandler(db)

	t.Run("获取工作汇总", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		startDate := time.Now().Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")
		c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/work-summary?start_date="+startDate+"&end_date="+endDate, nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWorkSummary(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["content"])
		assert.NotNil(t, data["hours"])
	})

	t.Run("获取工作汇总失败-缺少日期参数", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/work-summary", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetWorkSummary(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

