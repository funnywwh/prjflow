import request from '../utils/request'

export interface ProjectGroup {
  id: number
  name: string
  description?: string
  status: number
  created_at?: string
  updated_at?: string
  projects?: Project[]
}

export interface Project {
  id: number
  name: string
  code: string
  description?: string
  status: number
  project_group_id: number
  project_group?: ProjectGroup
  product_id?: number
  product?: any
  start_date?: string
  end_date?: string
  created_at?: string
  updated_at?: string
  members?: ProjectMember[]
}

export interface ProjectMember {
  id: number
  project_id: number
  user_id: number
  user?: any
  role: string
  created_at?: string
  updated_at?: string
}

export interface ProjectListResponse {
  list: Project[]
  total: number
  page: number
  size: number
}

export interface CreateProjectGroupRequest {
  name: string
  description?: string
  status?: number
}

export interface CreateProjectRequest {
  name: string
  code: string
  description?: string
  status?: number
  project_group_id: number
  product_id?: number
  start_date?: string
  end_date?: string
}

export interface AddProjectMembersRequest {
  user_ids: number[]
  role: string
}

// 项目集相关API
export const getProjectGroups = async (params?: {
  keyword?: string
  status?: number
}): Promise<ProjectGroup[]> => {
  return request.get('/project-groups', { params })
}

export const getProjectGroup = async (id: number): Promise<ProjectGroup> => {
  return request.get(`/project-groups/${id}`)
}

export const createProjectGroup = async (data: CreateProjectGroupRequest): Promise<ProjectGroup> => {
  return request.post('/project-groups', data)
}

export const updateProjectGroup = async (id: number, data: Partial<CreateProjectGroupRequest>): Promise<ProjectGroup> => {
  return request.put(`/project-groups/${id}`, data)
}

export const deleteProjectGroup = async (id: number): Promise<void> => {
  return request.delete(`/project-groups/${id}`)
}

// 项目相关API
export const getProjects = async (params?: {
  keyword?: string
  project_group_id?: number
  product_id?: number
  status?: number
  page?: number
  size?: number
}): Promise<ProjectListResponse> => {
  return request.get('/projects', { params })
}

export const getProject = async (id: number): Promise<Project> => {
  return request.get(`/projects/${id}`)
}

export const createProject = async (data: CreateProjectRequest): Promise<Project> => {
  return request.post('/projects', data)
}

export const updateProject = async (id: number, data: Partial<CreateProjectRequest>): Promise<Project> => {
  return request.put(`/projects/${id}`, data)
}

export const deleteProject = async (id: number): Promise<void> => {
  return request.delete(`/projects/${id}`)
}

// 项目成员相关API
export const getProjectMembers = async (projectId: number): Promise<ProjectMember[]> => {
  return request.get(`/projects/${projectId}/members`)
}

export const addProjectMembers = async (projectId: number, data: AddProjectMembersRequest): Promise<void> => {
  return request.post(`/projects/${projectId}/members`, data)
}

export const updateProjectMember = async (projectId: number, memberId: number, role: string): Promise<ProjectMember> => {
  return request.put(`/projects/${projectId}/members/${memberId}`, { role })
}

export const removeProjectMember = async (projectId: number, memberId: number): Promise<void> => {
  return request.delete(`/projects/${projectId}/members/${memberId}`)
}

