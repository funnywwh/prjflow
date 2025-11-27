import request from '../utils/request'

export interface Bug {
  id: number
  title: string
  description?: string
  status: 'active' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  severity: 'low' | 'medium' | 'high' | 'critical'
  confirmed?: boolean  // 是否确认
  project_id: number
  project?: any
  creator_id: number
  creator?: any
  assignees?: any[]
  requirement_id?: number
  requirement?: any
  module_id?: number
  module?: any
  estimated_hours?: number
  actual_hours?: number
  solution?: string
  solution_note?: string
  resolved_version_id?: number
  resolved_version?: any
  created_at?: string
  updated_at?: string
}

export interface BugListResponse {
  list: Bug[]
  total: number
  page: number
  page_size: number
}

export interface CreateBugRequest {
  title: string
  description?: string
  status?: 'active' | 'resolved' | 'closed'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  severity?: 'low' | 'medium' | 'high' | 'critical'
  project_id: number
  requirement_id?: number
  module_id?: number
  assignee_ids?: number[]
  estimated_hours?: number
  actual_hours?: number
  work_date?: string
}

export interface UpdateBugStatusRequest {
  status: 'active' | 'resolved' | 'closed'
  solution?: string
  solution_note?: string
  estimated_hours?: number
  actual_hours?: number
  work_date?: string
  resolved_version_id?: number
  version_number?: string
  create_version?: boolean
}

export interface AssignBugRequest {
  assignee_ids: number[]
}

// Bug相关API
export const getBugs = async (params?: {
  keyword?: string
  project_id?: number
  status?: string
  priority?: string
  severity?: string
  requirement_id?: number
  module_id?: number
  creator_id?: number
  assignee_id?: number
  page?: number
  size?: number
}): Promise<BugListResponse> => {
  return request.get('/bugs', { params })
}

export const getBug = async (id: number): Promise<Bug> => {
  return request.get(`/bugs/${id}`)
}

export const createBug = async (data: CreateBugRequest): Promise<Bug> => {
  return request.post('/bugs', data)
}

export const updateBug = async (id: number, data: Partial<CreateBugRequest>): Promise<Bug> => {
  return request.put(`/bugs/${id}`, data)
}

export const deleteBug = async (id: number): Promise<void> => {
  return request.delete(`/bugs/${id}`)
}

export const updateBugStatus = async (id: number, data: UpdateBugStatusRequest): Promise<Bug> => {
  return request.patch(`/bugs/${id}/status`, data)
}

export const assignBug = async (id: number, data: AssignBugRequest): Promise<Bug> => {
  return request.post(`/bugs/${id}/assign`, data)
}

export const confirmBug = async (id: number): Promise<Bug> => {
  return request.post(`/bugs/${id}/confirm`)
}

export interface BugStatistics {
  total: number
  active: number  // 激活状态
  resolved: number
  closed: number
  low_priority: number
  medium_priority: number
  high_priority: number
  urgent_priority: number
  low_severity: number
  medium_severity: number
  high_severity: number
  critical_severity: number
}

export const getBugStatistics = async (params?: {
  keyword?: string
  project_id?: number
  requirement_id?: number
  creator_id?: number
}): Promise<BugStatistics> => {
  return request.get('/bugs/statistics', { params })
}

// 历史记录相关接口
export interface History {
  id: number
  action_id: number
  field: string
  old: string
  old_value: string
  new: string
  new_value: string
  diff?: string
  created_at?: string
}

export interface Action {
  id: number
  object_type: string
  object_id: number
  project_id: number
  actor_id: number
  actor?: {
    id: number
    username: string
    nickname?: string
  }
  action: 'created' | 'edited' | 'assigned' | 'resolved' | 'closed' | 'confirmed' | 'commented'
  date: string
  comment?: string
  extra?: string
  histories?: History[]
  created_at?: string
}

export interface BugHistoryResponse {
  list: Action[]
}

export interface AddBugHistoryNoteRequest {
  comment: string
}

export const getBugHistory = async (id: number): Promise<BugHistoryResponse> => {
  return request.get(`/bugs/${id}/history`)
}

export const addBugHistoryNote = async (id: number, data: AddBugHistoryNoteRequest): Promise<{ message: string }> => {
  return request.post(`/bugs/${id}/history/note`, data)
}

