import request from '../utils/request'

export interface Task {
  id: number
  title: string
  description?: string
  status: 'todo' | 'in_progress' | 'done' | 'cancelled'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  project_id: number
  project?: any
  creator_id: number
  creator?: any
  assignee_id?: number
  assignee?: any
  start_date?: string
  end_date?: string
  due_date?: string
  progress: number
  estimated_hours?: number
  actual_hours?: number
  dependencies?: Task[]
  created_at?: string
  updated_at?: string
}

export interface TaskListResponse {
  list: Task[]
  total: number
  page: number
  page_size: number
}

export interface CreateTaskRequest {
  title: string
  description?: string
  status?: 'todo' | 'in_progress' | 'done' | 'cancelled'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  project_id: number
  assignee_id?: number
  start_date?: string
  end_date?: string
  due_date?: string
  progress?: number
  estimated_hours?: number
  actual_hours?: number
  work_date?: string
  dependency_ids?: number[]
}

export interface UpdateTaskStatusRequest {
  status: 'todo' | 'in_progress' | 'done' | 'cancelled'
}

export interface UpdateTaskProgressRequest {
  progress?: number
  estimated_hours?: number
  actual_hours?: number
  work_date?: string
}

// 任务相关API
export const getTasks = async (params?: {
  keyword?: string
  project_id?: number
  status?: string
  priority?: string
  assignee_id?: number
  creator_id?: number
  page?: number
  page_size?: number
}): Promise<TaskListResponse> => {
  return request.get('/tasks', { params })
}

export const getTask = async (id: number): Promise<Task> => {
  return request.get(`/tasks/${id}`)
}

export const createTask = async (data: CreateTaskRequest): Promise<Task> => {
  return request.post('/tasks', data)
}

export const updateTask = async (id: number, data: Partial<CreateTaskRequest>): Promise<Task> => {
  return request.put(`/tasks/${id}`, data)
}

export const deleteTask = async (id: number): Promise<void> => {
  return request.delete(`/tasks/${id}`)
}

export const updateTaskStatus = async (id: number, data: UpdateTaskStatusRequest): Promise<Task> => {
  return request.patch(`/tasks/${id}/status`, data)
}

export const updateTaskProgress = async (id: number, data: UpdateTaskProgressRequest): Promise<Task> => {
  return request.patch(`/tasks/${id}/progress`, data)
}

