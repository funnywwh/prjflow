import request from '../utils/request'

export interface Bug {
  id: number
  title: string
  description?: string
  status: 'open' | 'assigned' | 'in_progress' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  severity: 'low' | 'medium' | 'high' | 'critical'
  project_id: number
  project?: any
  creator_id: number
  creator?: any
  assignees?: any[]
  requirement_id?: number
  requirement?: any
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
  status?: 'open' | 'assigned' | 'in_progress' | 'resolved' | 'closed'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  severity?: 'low' | 'medium' | 'high' | 'critical'
  project_id: number
  requirement_id?: number
  assignee_ids?: number[]
}

export interface UpdateBugStatusRequest {
  status: 'open' | 'assigned' | 'in_progress' | 'resolved' | 'closed'
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
  creator_id?: number
  page?: number
  page_size?: number
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

export interface BugStatistics {
  total: number
  open: number
  assigned: number
  in_progress: number
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

