import request from '../utils/request'

export interface Resource {
  id?: number
  user_id: number
  user?: {
    id: number
    username: string
    nickname?: string
  }
  project_id: number
  project?: {
    id: number
    name: string
  }
  role?: string
  allocations?: ResourceAllocation[]
  created_at?: string
  updated_at?: string
}

export interface ResourceAllocation {
  id?: number
  resource_id: number
  resource?: Resource
  date: string
  hours: number
  task_id?: number
  task?: {
    id: number
    name: string
  }
  bug_id?: number
  bug?: {
    id: number
    title: string
  }
  project_id?: number
  project?: {
    id: number
    name: string
  }
  description?: string
  created_at?: string
  updated_at?: string
}

export interface CreateResourceRequest {
  user_id: number
  project_id: number
  role?: string
}

export interface UpdateResourceRequest {
  role?: string
}

export interface CreateResourceAllocationRequest {
  resource_id: number
  date: string
  hours: number
  task_id?: number
  bug_id?: number
  project_id?: number
  description?: string
}

export interface UpdateResourceAllocationRequest {
  date?: string
  hours?: number
  task_id?: number
  bug_id?: number
  project_id?: number
  description?: string
}

export interface ResourceListResponse {
  list: Resource[]
  total: number
  page: number
  page_size: number
}

export interface ResourceAllocationListResponse {
  list: ResourceAllocation[]
  total: number
  page: number
  page_size: number
}

export interface ResourceStatistics {
  total_hours: number
  project_stats: Array<{
    project_id: number
    project_name: string
    total_hours: number
  }>
  user_stats: Array<{
    user_id: number
    username: string
    nickname?: string
    total_hours: number
  }>
}

export interface ResourceUtilization {
  start_date: string
  end_date: string
  days: number
  utilization_stats: Array<{
    resource_id: number
    user_id: number
    username: string
    nickname?: string
    project_id: number
    project_name: string
    total_hours: number
    max_hours: number
    utilization: number
  }>
  avg_utilization: number
}

export interface ResourceCalendar {
  start_date: string
  end_date: string
  data: Record<string, ResourceAllocation[]>
}

export interface ResourceConflict {
  resource_id: string
  date: string
  total_hours: number
  conflicts: string[]
  has_conflict: boolean
}

// 获取资源列表
export const getResources = (params?: {
  user_id?: number
  project_id?: number
  role?: string
  page?: number
  page_size?: number
}) => {
  return request.get<ResourceListResponse>('/resources', { params })
}

// 获取资源详情
export const getResource = (id: number) => {
  return request.get<Resource>(`/resources/${id}`)
}

// 创建资源
export const createResource = (data: CreateResourceRequest) => {
  return request.post<Resource>('/resources', data)
}

// 更新资源
export const updateResource = (id: number, data: UpdateResourceRequest) => {
  return request.put<Resource>(`/resources/${id}`, data)
}

// 删除资源
export const deleteResource = (id: number) => {
  return request.delete(`/resources/${id}`)
}

// 获取资源统计
export const getResourceStatistics = (params?: {
  user_id?: number
  project_id?: number
  start_date?: string
  end_date?: string
}) => {
  return request.get<ResourceStatistics>('/resources/statistics', { params })
}

// 获取资源利用率
export const getResourceUtilization = (params?: {
  user_id?: number
  project_id?: number
  start_date?: string
  end_date?: string
}) => {
  return request.get<ResourceUtilization>('/resources/utilization', { params })
}

// 获取资源分配列表
export const getResourceAllocations = (params?: {
  resource_id?: number
  user_id?: number
  project_id?: number
  task_id?: number
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
}) => {
  return request.get<ResourceAllocationListResponse>('/resource-allocations', { params })
}

// 获取资源分配详情
export const getResourceAllocation = (id: number) => {
  return request.get<ResourceAllocation>(`/resource-allocations/${id}`)
}

// 创建资源分配
export const createResourceAllocation = (data: CreateResourceAllocationRequest) => {
  return request.post<ResourceAllocation>('/resource-allocations', data)
}

// 更新资源分配
export const updateResourceAllocation = (id: number, data: UpdateResourceAllocationRequest) => {
  return request.put<ResourceAllocation>(`/resource-allocations/${id}`, data)
}

// 删除资源分配
export const deleteResourceAllocation = (id: number) => {
  return request.delete(`/resource-allocations/${id}`)
}

// 获取资源日历
export const getResourceCalendar = (params?: {
  user_id?: number
  project_id?: number
  start_date?: string
  end_date?: string
}) => {
  return request.get<ResourceCalendar>('/resource-allocations/calendar', { params })
}

// 检查资源冲突
export const checkResourceConflict = (params: {
  resource_id: number
  date: string
}) => {
  return request.get<ResourceConflict>('/resource-allocations/conflict', { params })
}

