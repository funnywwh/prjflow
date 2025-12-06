import request from '../utils/request'

export interface AuditLog {
  id: number
  user_id: number
  username: string
  action_type: string
  resource_type?: string
  resource_id?: number
  ip_address?: string
  path?: string
  method?: string
  params?: string
  success: boolean
  error_msg?: string
  comment?: string
  created_at: string
  updated_at: string
}

export interface AuditLogListParams {
  user_id?: number
  action_type?: string
  resource_type?: string
  success?: boolean
  start_date?: string
  end_date?: string
  keyword?: string
  page?: number
  size?: number
}

export interface AuditLogListResponse {
  list: AuditLog[]
  total: number
}

// 获取审计日志列表
export const getAuditLogs = async (params?: AuditLogListParams): Promise<AuditLogListResponse> => {
  return request.get('/audit-logs', { params })
}

// 获取审计日志详情
export const getAuditLog = async (id: number): Promise<AuditLog> => {
  return request.get(`/audit-logs/${id}`)
}



