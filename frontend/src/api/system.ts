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

// 日志级别管理
export interface LogLevel {
  level: string
}

export interface LogLevelRequest {
  level: string
}

// 获取日志级别
export const getLogLevel = async (): Promise<LogLevel> => {
  return request.get('/system/log-level')
}

// 设置日志级别
export const setLogLevel = async (data: LogLevelRequest): Promise<{ message: string; level: string }> => {
  return request.post('/system/log-level', data)
}

// 日志文件信息
export interface LogFileInfo {
  filename: string
  size: number
  size_formatted: string
  mod_time: string
}

export interface LogFilesResponse {
  files: LogFileInfo[]
}

// 获取日志文件列表
export const getLogFiles = async (): Promise<LogFilesResponse> => {
  return request.get('/system/log-files')
}

// 下载日志文件
export const downloadLogFile = async (filename: string): Promise<void> => {
  // 使用axios直接下载，绕过request拦截器（因为需要blob响应）
  const axios = (await import('axios')).default
  const { useAuthStore } = await import('../stores/auth')
  const authStore = useAuthStore()
  
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
  const response = await axios.get(`${baseURL}/system/log-files/${filename}`, {
    responseType: 'blob',
    headers: {
      'Authorization': authStore.token ? `Bearer ${authStore.token}` : ''
    }
  })
  
  // 创建临时链接并触发下载
  const url = window.URL.createObjectURL(response.data)
  const link = document.createElement('a')
  link.href = url
  
  // 从响应头获取文件名，如果没有则使用原始文件名
  const contentDisposition = response.headers['content-disposition']
  let downloadFilename = filename
  if (contentDisposition) {
    const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/)
    if (filenameMatch && filenameMatch[1]) {
      downloadFilename = filenameMatch[1].replace(/['"]/g, '')
    }
  }
  
  link.setAttribute('download', downloadFilename)
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}

