import request from '../utils/request'

export interface BackupConfig {
  enabled: boolean
  backup_time: string
  last_backup_date?: string
}

export interface BackupConfigRequest {
  enabled: boolean
  backup_time: string
}

// 获取备份配置
export const getBackupConfig = async (): Promise<BackupConfig> => {
  return request.get('/system/backup-config')
}

// 保存备份配置
export const saveBackupConfig = async (data: BackupConfigRequest): Promise<{ message: string }> => {
  return request.post('/system/backup-config', data)
}

// 手动触发备份
export const triggerBackup = async (): Promise<{ message: string }> => {
  return request.post('/system/backup/trigger')
}

