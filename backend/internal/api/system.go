package api

import (
	"time"

	"project-management/internal/model"
	"project-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemHandler struct {
	db *gorm.DB
}

func NewSystemHandler(db *gorm.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

// GetBackupConfig 获取备份配置
func (h *SystemHandler) GetBackupConfig(c *gin.Context) {
	var enabledConfig model.SystemConfig
	var timeConfig model.SystemConfig
	var lastDateConfig model.SystemConfig

	enabled := false
	backupTime := "02:00"
	lastBackupDate := ""

	// 读取备份启用状态
	if err := h.db.Where("key = ?", "backup_enabled").First(&enabledConfig).Error; err == nil {
		enabled = enabledConfig.Value == "true"
	}

	// 读取备份时间
	if err := h.db.Where("key = ?", "backup_time").First(&timeConfig).Error; err == nil {
		backupTime = timeConfig.Value
	}

	// 读取上次备份日期
	if err := h.db.Where("key = ?", "backup_last_date").First(&lastDateConfig).Error; err == nil {
		lastBackupDate = lastDateConfig.Value
	}

	utils.Success(c, gin.H{
		"enabled":         enabled,
		"backup_time":     backupTime,
		"last_backup_date": lastBackupDate,
	})
}

// SaveBackupConfig 保存备份配置
func (h *SystemHandler) SaveBackupConfig(c *gin.Context) {
	var req struct {
		Enabled    bool   `json:"enabled" binding:"required"`
		BackupTime string `json:"backup_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证备份时间格式 (HH:MM)
	if _, err := time.Parse("15:04", req.BackupTime); err != nil {
		utils.Error(c, 400, "备份时间格式错误，应为 HH:MM (24小时制)")
		return
	}

	// 保存备份启用状态
	enabledValue := "false"
	if req.Enabled {
		enabledValue = "true"
	}
	enabledConfig := model.SystemConfig{
		Key:   "backup_enabled",
		Value: enabledValue,
		Type:  "boolean",
	}
	if err := h.db.Where("key = ?", "backup_enabled").
		Assign(model.SystemConfig{Value: enabledValue, Type: "boolean"}).
		FirstOrCreate(&enabledConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存备份启用状态失败: "+err.Error())
		return
	}

	// 保存备份时间
	timeConfig := model.SystemConfig{
		Key:   "backup_time",
		Value: req.BackupTime,
		Type:  "string",
	}
	if err := h.db.Where("key = ?", "backup_time").
		Assign(model.SystemConfig{Value: req.BackupTime, Type: "string"}).
		FirstOrCreate(&timeConfig).Error; err != nil {
		utils.Error(c, utils.CodeError, "保存备份时间失败: "+err.Error())
		return
	}

	// 重新加载定时任务配置
	scheduler := utils.GetBackupScheduler(h.db)
	scheduler.Reload()

	utils.Success(c, gin.H{
		"message": "备份配置已保存",
	})
}

// TriggerBackup 手动触发备份
func (h *SystemHandler) TriggerBackup(c *gin.Context) {
	// 执行备份
	if err := utils.TriggerBackup(h.db); err != nil {
		if err.Error() == "backup is already in progress" {
			utils.Error(c, 400, "备份正在进行中，请稍后再试")
			return
		}
		utils.Error(c, utils.CodeError, "备份失败: "+err.Error())
		return
	}

	// 更新上次备份日期
	today := time.Now().Format("2006-01-02")
	lastDateConfig := model.SystemConfig{
		Key:   "backup_last_date",
		Value: today,
		Type:  "string",
	}
	h.db.Where("key = ?", "backup_last_date").
		Assign(model.SystemConfig{Value: today, Type: "string"}).
		FirstOrCreate(&lastDateConfig)

	utils.Success(c, gin.H{
		"message": "备份已触发",
	})
}

