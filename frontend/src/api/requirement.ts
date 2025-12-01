import request from '../utils/request'

export interface Requirement {
  id: number
  title: string
  description?: string
  status: 'draft' | 'reviewing' | 'active' | 'changing' | 'closed'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  project_id: number // 必填
  project?: any
  creator_id: number
  creator?: any
  assignee_id?: number
  assignee?: any
  estimated_hours?: number
  actual_hours?: number
  created_at?: string
  updated_at?: string
}

export interface RequirementListResponse {
  list: Requirement[]
  total: number
  page: number
  page_size: number
}

export interface CreateRequirementRequest {
  title: string
  description?: string
  status?: 'draft' | 'reviewing' | 'active' | 'changing' | 'closed'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  project_id: number // 必填
  assignee_id?: number
  estimated_hours?: number
}

export interface UpdateRequirementStatusRequest {
  status: 'draft' | 'reviewing' | 'active' | 'changing' | 'closed'
}

// 需求相关API
export const getRequirements = async (params?: {
  keyword?: string
  project_id?: number
  status?: string
  priority?: string
  assignee_id?: number
  creator_id?: number
  page?: number
  size?: number
}): Promise<RequirementListResponse> => {
  return request.get('/requirements', { params })
}

export const getRequirement = async (id: number): Promise<Requirement> => {
  return request.get(`/requirements/${id}`)
}

export const createRequirement = async (data: CreateRequirementRequest): Promise<Requirement> => {
  return request.post('/requirements', data)
}

export const updateRequirement = async (id: number, data: Partial<CreateRequirementRequest>): Promise<Requirement> => {
  return request.put(`/requirements/${id}`, data)
}

export const deleteRequirement = async (id: number): Promise<void> => {
  return request.delete(`/requirements/${id}`)
}

export const updateRequirementStatus = async (id: number, data: UpdateRequirementStatusRequest): Promise<Requirement> => {
  return request.patch(`/requirements/${id}/status`, data)
}

export interface RequirementStatistics {
  total: number
  pending: number
  in_progress: number
  completed: number
  cancelled: number
  low_priority: number
  medium_priority: number
  high_priority: number
  urgent_priority: number
}

export const getRequirementStatistics = async (params?: {
  keyword?: string
  project_id?: number
  assignee_id?: number
  creator_id?: number
}): Promise<RequirementStatistics> => {
  return request.get('/requirements/statistics', { params })
}

// 历史记录相关接口（参考bug.ts和project.ts）
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
  action: 'created' | 'edited' | 'assigned' | 'resolved' | 'closed' | 'confirmed' | 'commented' | 'status_changed'
  date: string
  comment?: string
  extra?: string
  histories?: History[]
  created_at?: string
}

export interface RequirementHistoryResponse {
  list: Action[]
}

export interface AddRequirementHistoryNoteRequest {
  comment: string
}

export const getRequirementHistory = async (id: number): Promise<RequirementHistoryResponse> => {
  return request.get(`/requirements/${id}/history`)
}

export const addRequirementHistoryNote = async (id: number, data: AddRequirementHistoryNoteRequest): Promise<{ message: string }> => {
  return request.post(`/requirements/${id}/history/note`, data)
}

