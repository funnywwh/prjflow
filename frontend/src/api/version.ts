import request from '../utils/request'

// 版本管理相关接口
export interface Version {
  id: number
  version_number: string
  release_notes?: string
  status: 'wait' | 'normal' | 'fail' | 'terminate'
  project_id: number
  project?: any
  release_date?: string
  requirements?: any[]
  bugs?: any[]
  created_at?: string
  updated_at?: string
}

export interface VersionListResponse {
  list: Version[]
  total: number
  page: number
  page_size: number
}

export interface CreateVersionRequest {
  version_number: string
  release_notes?: string
  status?: 'wait' | 'normal' | 'fail' | 'terminate'
  project_id: number
  release_date?: string
  requirement_ids?: number[]
  bug_ids?: number[]
}

export interface UpdateVersionRequest {
  version_number?: string
  release_notes?: string
  status?: 'wait' | 'normal' | 'fail' | 'terminate'
  release_date?: string
  requirement_ids?: number[]
  bug_ids?: number[]
}

// 获取版本列表
export const getVersions = (params?: {
  keyword?: string
  project_id?: number
  status?: string
  page?: number
  size?: number
}) => {
  return request.get<VersionListResponse>('/versions', { params })
}

// 获取版本详情
export const getVersion = (id: number) => {
  return request.get<Version>(`/versions/${id}`)
}

// 创建版本
export const createVersion = (data: CreateVersionRequest) => {
  return request.post<Version>('/versions', data)
}

// 更新版本
export const updateVersion = (id: number, data: UpdateVersionRequest) => {
  return request.put<Version>(`/versions/${id}`, data)
}

// 删除版本
export const deleteVersion = (id: number) => {
  return request.delete(`/versions/${id}`)
}

// 更新版本状态
export const updateVersionStatus = (id: number, status: string) => {
  return request.patch<Version>(`/versions/${id}/status`, { status })
}

// 发布版本
export const releaseVersion = (id: number) => {
  return request.post<Version>(`/versions/${id}/release`)
}

// 系统版本信息相关接口
export interface VersionInfo {
  version: string
  build_time?: string
  git_commit?: string
  go_version: string
}

// 获取系统版本信息
export const getVersionInfo = async (): Promise<VersionInfo> => {
  return request.get('/version')
}
