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

func TestBugHandler_GetBugs(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "Bug测试项目")
	user := CreateTestUser(t, db, "buguser", "Bug用户")
	adminUser := CreateTestAdminUser(t, db, "adminbug", "管理员Bug用户")
	otherUser := CreateTestUser(t, db, "otherbug", "其他Bug用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	// 创建测试Bug
	bug1 := &model.Bug{
		Title:     "Bug1",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "open",
		Priority:  "high",
		Severity:  "critical",
	}
	db.Create(bug1)

	// 创建另一个项目的Bug
	project2 := CreateTestProject(t, db, "Bug测试项目2")
	bug2 := &model.Bug{
		Title:     "Bug2",
		ProjectID: project2.ID,
		CreatorID: otherUser.ID,
		Status:    "open",
		Priority:  "high",
		Severity:  "critical",
	}
	db.Create(bug2)

	handler := api.NewBugHandler(db)

	t.Run("管理员可以获取所有Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/bugs", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetBugs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 管理员应该能看到所有Bug
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("普通用户只能看到自己创建或参与的Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/bugs", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetBugs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 用户创建了Bug1且是项目成员，应该能看到Bug1
		assert.Equal(t, 1, len(list))
	})

	t.Run("搜索Bug-管理员", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/bugs?keyword=Bug1", nil)
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetBugs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("搜索Bug-普通用户", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/bugs?keyword=Bug1", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetBugs(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		// 用户应该能看到自己创建的Bug1
		assert.Equal(t, 1, len(list))
	})

	_ = project
	_ = user
}

func TestBugHandler_GetBug(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "Bug详情项目")
	user := CreateTestUser(t, db, "bugdetail", "Bug详情用户")
	adminUser := CreateTestAdminUser(t, db, "adminbug2", "管理员Bug用户2")
	otherUser := CreateTestUser(t, db, "otherbug2", "其他Bug用户2")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	bug := &model.Bug{
		Title:     "测试Bug",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "open",
		Priority:  "high",
		Severity:  "medium",
	}
	db.Create(&bug)

	handler := api.NewBugHandler(db)

	t.Run("管理员可以获取任何Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/bugs/%d", bug.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", bug.ID)}}
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})

		handler.GetBug(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "测试Bug", data["title"])
	})

	t.Run("创建者可以获取Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/bugs/%d", bug.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", bug.ID)}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		handler.GetBug(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])
	})

	t.Run("非项目成员不能获取Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/bugs/%d", bug.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", bug.ID)}}
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"developer"})

		handler.GetBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 应该返回403或code不为200
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("获取不存在的Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/bugs/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.GetBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBugHandler_CreateBug(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "创建Bug项目")
	user := CreateTestUser(t, db, "createbug", "创建Bug用户")
	handler := api.NewBugHandler(db)

	t.Run("创建Bug成功-项目成员", func(t *testing.T) {
		// 添加用户到项目
		AddUserToProject(t, db, user.ID, project.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id（CreateBug需要）
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"title":          "新Bug",
			"description":    "这是一个新Bug",
			"status":         "active",
			"priority":       "high",
			"severity":       "critical",
			"project_id":     project.ID,
			"estimated_hours": 8.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/bugs", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateBug(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证Bug已创建
		var bug model.Bug
		err = db.Where("title = ?", "新Bug").First(&bug).Error
		assert.NoError(t, err)
		assert.Equal(t, "新Bug", bug.Title)
		assert.Equal(t, project.ID, bug.ProjectID)
	})

	t.Run("创建Bug失败-非项目成员", func(t *testing.T) {
		otherUser := CreateTestUser(t, db, "othercreatebug", "其他创建Bug用户")
		otherProject := CreateTestProject(t, db, "其他Bug项目")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", otherUser.ID)
		c.Set("roles", []string{"developer"})

		reqBody := map[string]interface{}{
			"title":          "新Bug",
			"description":    "这是一个新Bug",
			"status":         "active",
			"priority":       "high",
			"severity":       "critical",
			"project_id":     otherProject.ID,
			"estimated_hours": 8.0,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/bugs", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		// 应该返回403或code不为200
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建Bug失败-缺少必填字段", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"description": "只有描述",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/bugs", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("创建Bug失败-项目不存在", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 设置user_id
		c.Set("user_id", user.ID)

		reqBody := map[string]interface{}{
			"title":      "新Bug",
			"project_id": 999,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/bugs", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreateBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusBadRequest || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBugHandler_UpdateBug(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "更新Bug项目")
	user := CreateTestUser(t, db, "updatebug", "更新Bug用户")

	bug := &model.Bug{
		Title:     "更新Bug",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "open",
		Priority:  "high",
		Severity:  "medium",
	}
	db.Create(&bug)

	handler := api.NewBugHandler(db)

	t.Run("更新Bug成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})

		// 先将Bug状态改为active（因为初始状态是open，需要先改为active）
		bug.Status = "active"
		db.Save(&bug)

		reqBody := map[string]interface{}{
			"title":     "已更新Bug",
			"status":    "resolved", // 使用有效的状态值：active -> resolved
			"priority":  "medium",
			"severity":  "high",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/bugs/%d", bug.ID), bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", bug.ID)}}

		handler.UpdateBug(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证Bug已更新
		var updatedBug model.Bug
		err := db.First(&updatedBug, bug.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "已更新Bug", updatedBug.Title)
		assert.Equal(t, "resolved", updatedBug.Status)
	})

	t.Run("更新不存在的Bug", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"title": "不存在的Bug",
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/bugs/999", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		handler.UpdateBug(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestBugHandler_DeleteBug(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	project := CreateTestProject(t, db, "删除Bug项目")
	user := CreateTestUser(t, db, "deletebug", "删除Bug用户")

	bug := &model.Bug{
		Title:     "删除Bug",
		ProjectID: project.ID,
		CreatorID: user.ID,
		Status:    "open",
	}
	db.Create(&bug)

	handler := api.NewBugHandler(db)

	t.Run("删除Bug成功", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"developer"})
		c.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/bugs/%d", bug.ID), nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", bug.ID)}}

		handler.DeleteBug(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证Bug已软删除
		var deletedBug model.Bug
		err := db.First(&deletedBug, bug.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）

		// 验证软删除后仍可通过Unscoped查询
		err = db.Unscoped().First(&deletedBug, bug.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedBug.DeletedAt)
	})
}

