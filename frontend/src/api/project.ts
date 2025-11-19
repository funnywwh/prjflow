import request from '../utils/request'

export interface Project {
  id: number
  name: string
  code: string
  description?: string
  status: number
  tags?: Array<{ id: number; name: string; color?: string }>  // 标签对象数组
  start_date?: string
  end_date?: string
  created_at?: string
  updated_at?: string
  members?: ProjectMember[]
}

export interface ProjectStatistics {
  total_tasks: number
  todo_tasks: number
  in_progress_tasks: number
  done_tasks: number
  total_bugs: number
  open_bugs: number
  in_progress_bugs: number
  resolved_bugs: number
  total_requirements: number
  in_progress_requirements: number
  completed_requirements: number
  total_members: number
}

export interface ProjectDetailResponse {
  project: Project
  statistics: ProjectStatistics
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

export interface CreateProjectRequest {
  name: string
  code: string
  description?: string
  status?: number
  tag_ids?: number[]  // 标签ID数组
  start_date?: string
  end_date?: string
}

export interface AddProjectMembersRequest {
  user_ids: number[]
  role: string
}

// 项目相关API
export const getProjects = async (params?: {
  keyword?: string
  tag?: string  // 向后兼容单个标签
  tags?: string[]  // 多个标签（优先使用）
  status?: number
  page?: number
  size?: number
}): Promise<ProjectListResponse> => {
  return request.get('/projects', { params })
}

export const getProject = async (id: number): Promise<ProjectDetailResponse> => {
  return request.get(`/projects/${id}`)
}

export const getProjectStatistics = async (id: number): Promise<ProjectStatistics> => {
  return request.get(`/projects/${id}/statistics`)
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

// 甘特图相关API
export interface GanttTask {
  id: number
  title: string
  start_date?: string
  end_date?: string
  due_date?: string
  progress: number
  status: string
  priority: string
  assignee?: string
  dependencies?: number[]
}

export interface GanttData {
  tasks: GanttTask[]
}

export const getProjectGantt = async (projectId: number): Promise<GanttData> => {
  return request.get(`/projects/${projectId}/gantt`)
}

// 项目进度跟踪相关API
export interface ProjectProgressData {
  statistics: ProjectStatistics
  task_progress_trend: Array<{ date: string; average: number; count: number }>
  task_status_distribution: Array<{ status: string; count: number }>
  task_priority_distribution: Array<{ priority: string; count: number }>
  task_completion_trend: Array<{ week: string; total: number; completed: number; completion_rate: number }>
  member_workload: Array<{ user_id: number; username: string; nickname?: string; total: number; completed: number; in_progress: number }>
  bug_trend: Array<{ date: string; count: number }>
  requirement_trend: Array<{ date: string; count: number }>
}

export const getProjectProgress = async (projectId: number): Promise<ProjectProgressData> => {
  return request.get(`/projects/${projectId}/progress`)
}

