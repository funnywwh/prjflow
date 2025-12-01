package utils

import (
	"fmt"
	"log"
	"sync"
	"time"

	"project-management/internal/model"
	"gorm.io/gorm"
)

var (
	backupScheduler     *BackupScheduler
	backupSchedulerOnce sync.Once
	backupMutex         sync.Mutex
	isBackingUp         bool
)

// BackupScheduler 备份定时任务调度器
type BackupScheduler struct {
	db        *gorm.DB
	timer     *time.Timer
	stopChan  chan struct{}
	mu        sync.Mutex
}

// GetBackupScheduler 获取备份调度器单例
func GetBackupScheduler(db *gorm.DB) *BackupScheduler {
	backupSchedulerOnce.Do(func() {
		backupScheduler = &BackupScheduler{
			db:       db,
			stopChan: make(chan struct{}),
		}
	})
	return backupScheduler
}

// Start 启动定时任务
func (s *BackupScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经有定时器在运行，先停止
	if s.timer != nil {
		s.timer.Stop()
	}

	// 读取配置并计算下次执行时间
	nextTime := s.calculateNextBackupTime()
	if nextTime.IsZero() {
		log.Printf("[Scheduler] Auto backup is disabled or not configured")
		return
	}

	duration := time.Until(nextTime)
	log.Printf("[Scheduler] Next backup scheduled at: %s (in %v)", nextTime.Format("2006-01-02 15:04:05"), duration)

	// 设置定时器
	s.timer = time.AfterFunc(duration, func() {
		s.executeBackup()
		// 递归调度下次备份
		s.Start()
	})
}

// Stop 停止定时任务
func (s *BackupScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	close(s.stopChan)
}

// Reload 重新加载配置并重启定时任务
func (s *BackupScheduler) Reload() {
	log.Printf("[Scheduler] Reloading backup scheduler configuration")
	s.Stop()
	// 重新创建 stopChan
	s.mu.Lock()
	s.stopChan = make(chan struct{})
	s.mu.Unlock()
	s.Start()
}

// calculateNextBackupTime 计算下次备份时间
func (s *BackupScheduler) calculateNextBackupTime() time.Time {
	// 读取配置
	var enabledConfig model.SystemConfig
	var timeConfig model.SystemConfig

	if err := s.db.Where("key = ?", "backup_enabled").First(&enabledConfig).Error; err != nil {
		// 配置不存在，返回零时间
		return time.Time{}
	}

	if enabledConfig.Value != "true" {
		// 自动备份未启用
		return time.Time{}
	}

	if err := s.db.Where("key = ?", "backup_time").First(&timeConfig).Error; err != nil {
		// 备份时间未配置
		return time.Time{}
	}

	// 解析备份时间 (格式: HH:MM)
	backupTime, err := time.Parse("15:04", timeConfig.Value)
	if err != nil {
		log.Printf("[Scheduler] Invalid backup time format: %s", timeConfig.Value)
		return time.Time{}
	}

	// 获取当前时间
	now := time.Now()
	
	// 构造今天的备份时间
	todayBackup := time.Date(now.Year(), now.Month(), now.Day(),
		backupTime.Hour(), backupTime.Minute(), 0, 0, now.Location())

	// 检查今天是否已经备份过
	var lastDateConfig model.SystemConfig
	if err := s.db.Where("key = ?", "backup_last_date").First(&lastDateConfig).Error; err == nil {
		lastDate, err := time.Parse("2006-01-02", lastDateConfig.Value)
		if err == nil && lastDate.Format("2006-01-02") == now.Format("2006-01-02") {
			// 今天已经备份过，返回明天的备份时间
			return todayBackup.AddDate(0, 0, 1)
		}
	}

	// 如果今天的备份时间已过，返回明天的备份时间
	if now.After(todayBackup) {
		return todayBackup.AddDate(0, 0, 1)
	}

	// 返回今天的备份时间
	return todayBackup
}

// executeBackup 执行备份
func (s *BackupScheduler) executeBackup() {
	// 检查是否正在备份
	backupMutex.Lock()
	if isBackingUp {
		backupMutex.Unlock()
		log.Printf("[Scheduler] Backup is already in progress, skipping")
		return
	}
	isBackingUp = true
	backupMutex.Unlock()

	defer func() {
		backupMutex.Lock()
		isBackingUp = false
		backupMutex.Unlock()
	}()

	// 再次检查配置（可能在执行过程中被禁用）
	var enabledConfig model.SystemConfig
	if err := s.db.Where("key = ?", "backup_enabled").First(&enabledConfig).Error; err != nil {
		log.Printf("[Scheduler] Failed to read backup_enabled config: %v", err)
		return
	}

	if enabledConfig.Value != "true" {
		log.Printf("[Scheduler] Auto backup is disabled, skipping")
		return
	}

	log.Printf("[Scheduler] Starting scheduled backup...")
	
	// 执行备份
	if err := BackupDatabase(s.db); err != nil {
		log.Printf("[Scheduler] Scheduled backup failed: %v", err)
		return
	}

	// 更新上次备份日期
	today := time.Now().Format("2006-01-02")
	lastDateConfig := model.SystemConfig{
		Key:   "backup_last_date",
		Value: today,
		Type:  "string",
	}
	
	if err := s.db.Where("key = ?", "backup_last_date").
		Assign(model.SystemConfig{Value: today, Type: "string"}).
		FirstOrCreate(&lastDateConfig).Error; err != nil {
		log.Printf("[Scheduler] Failed to update backup_last_date: %v", err)
	}

	log.Printf("[Scheduler] Scheduled backup completed successfully")
}

// TriggerBackup 手动触发备份（带并发控制）
func TriggerBackup(db *gorm.DB) error {
	backupMutex.Lock()
	if isBackingUp {
		backupMutex.Unlock()
		return fmt.Errorf("backup is already in progress")
	}
	isBackingUp = true
	backupMutex.Unlock()

	defer func() {
		backupMutex.Lock()
		isBackingUp = false
		backupMutex.Unlock()
	}()

	return BackupDatabase(db)
}

