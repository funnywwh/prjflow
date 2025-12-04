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
	"project-management/internal/config"
	"project-management/internal/model"
)

func TestAttachmentHandler_GetAttachments(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 初始化配置
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Upload: config.UploadConfig{
				StoragePath:  "./test_uploads",
				MaxFileSize:  10 * 1024 * 1024, // 10MB
				AllowedTypes: []string{"image/", "application/pdf"},
			},
		}
	}

	project := CreateTestProject(t, db, "附件测试项目")
	user := CreateTestUser(t, db, "attachmentuser", "附件用户")
	_ = CreateTestAdminUser(t, db, "adminattachment", "管理员附件用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	// 创建测试附件
	attachment := &model.Attachment{
		FileName: "test.pdf",
		FilePath: "2024/01/01/test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
		CreatorID: user.ID,
	}
	db.Create(attachment)
	// 关联到项目
	db.Model(attachment).Association("Projects").Append(project)

	handler := api.NewAttachmentHandler(db)

	t.Run("获取附件列表-项目筛选", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/attachments?project_id=1", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetAttachments(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].([]interface{})
		assert.GreaterOrEqual(t, len(data), 1)
	})

	t.Run("获取附件列表-无权限", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/attachments?project_id=999", nil)
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetAttachments(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAttachmentHandler_GetAttachment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 初始化配置
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Upload: config.UploadConfig{
				StoragePath:  "./test_uploads",
				MaxFileSize:  10 * 1024 * 1024,
				AllowedTypes: []string{"image/", "application/pdf"},
			},
		}
	}

	project := CreateTestProject(t, db, "附件详情项目")
	user := CreateTestUser(t, db, "attachmentdetail", "附件详情用户")
	_ = CreateTestAdminUser(t, db, "adminattachmentdetail", "管理员附件详情用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	attachment := &model.Attachment{
		FileName: "test.pdf",
		FilePath: "2024/01/01/test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
		CreatorID: user.ID,
	}
	db.Create(attachment)
	db.Model(attachment).Association("Projects").Append(project)

	handler := api.NewAttachmentHandler(db)

	t.Run("获取存在的附件", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/attachments/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetAttachment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test.pdf", data["file_name"])
	})

	t.Run("获取不存在的附件", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/attachments/999", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetAttachment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusNotFound || (response["code"] != nil && response["code"] != float64(200)))
	})

	t.Run("获取附件-无权限", func(t *testing.T) {
		// 创建另一个项目
		project2 := CreateTestProject(t, db, "其他项目")
		attachment2 := &model.Attachment{
			FileName: "test2.pdf",
			FilePath: "2024/01/01/test2.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			CreatorID: user.ID,
		}
		db.Create(attachment2)
		db.Model(attachment2).Association("Projects").Append(project2)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/attachments/2", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.GetAttachment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAttachmentHandler_DeleteAttachment(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 初始化配置
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Upload: config.UploadConfig{
				StoragePath:  "./test_uploads",
				MaxFileSize:  10 * 1024 * 1024,
				AllowedTypes: []string{"image/", "application/pdf"},
			},
		}
	}

	project := CreateTestProject(t, db, "删除附件项目")
	user := CreateTestUser(t, db, "deleteattachment", "删除附件用户")
	adminUser := CreateTestAdminUser(t, db, "admindeleteattachment", "管理员删除附件用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	attachment := &model.Attachment{
		FileName: "test.pdf",
		FilePath: "2024/01/01/test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
		CreatorID: user.ID,
	}
	db.Create(attachment)
	db.Model(attachment).Association("Projects").Append(project)

	handler := api.NewAttachmentHandler(db)

	t.Run("删除附件成功-管理员", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/attachments/1", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", adminUser.ID)
		c.Set("roles", []string{"admin"})
		c.Set("permissions", []string{"attachment:delete"})

		handler.DeleteAttachment(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证附件已软删除
		var deletedAttachment model.Attachment
		err := db.First(&deletedAttachment, attachment.ID).Error
		assert.Error(t, err) // 应该找不到（软删除）
	})

	t.Run("删除附件-无权限", func(t *testing.T) {
		// 重新创建附件
		attachment2 := &model.Attachment{
			FileName: "test2.pdf",
			FilePath: "2024/01/01/test2.pdf",
			FileSize: 1024,
			MimeType: "application/pdf",
			CreatorID: user.ID,
		}
		db.Create(attachment2)
		db.Model(attachment2).Association("Projects").Append(project)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/attachments/2", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})
		c.Set("permissions", []string{}) // 没有删除权限

		handler.DeleteAttachment(c)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, w.Code == http.StatusForbidden || (response["code"] != nil && response["code"] != float64(200)))
	})
}

func TestAttachmentHandler_AttachToEntity(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	// 初始化配置
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Upload: config.UploadConfig{
				StoragePath:  "./test_uploads",
				MaxFileSize:  10 * 1024 * 1024,
				AllowedTypes: []string{"image/", "application/pdf"},
			},
		}
	}

	project := CreateTestProject(t, db, "关联附件项目")
	user := CreateTestUser(t, db, "attachtoentity", "关联附件用户")

	// 添加用户到项目
	AddUserToProject(t, db, user.ID, project.ID, "member")

	attachment := &model.Attachment{
		FileName: "test.pdf",
		FilePath: "2024/01/01/test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
		CreatorID: user.ID,
	}
	db.Create(attachment)
	db.Model(attachment).Association("Projects").Append(project)

	handler := api.NewAttachmentHandler(db)

	t.Run("关联附件到项目", func(t *testing.T) {
		project2 := CreateTestProject(t, db, "关联附件项目2")
		AddUserToProject(t, db, user.ID, project2.ID, "member")

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]interface{}{
			"project_id": project2.ID,
		}
		jsonData, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/attachments/1/attach", bytes.NewBuffer(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", user.ID)
		c.Set("roles", []string{"user"})

		handler.AttachToEntity(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(200), response["code"])

		// 验证关联已创建
		var updatedAttachment model.Attachment
		db.Preload("Projects").First(&updatedAttachment, attachment.ID)
		assert.GreaterOrEqual(t, len(updatedAttachment.Projects), 1)
	})
}

