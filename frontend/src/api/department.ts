import request from '../utils/request'
import type { Department } from './user'

// 获取部门列表（树形结构）
export const getDepartments = async (): Promise<Department[]> => {
  return request.get('/departments')
}

// 获取部门详情
export const getDepartment = async (id: number): Promise<Department> => {
  return request.get(`/departments/${id}`)
}

// 创建部门
export const createDepartment = async (data: {
  name: string
  code: string
  parent_id?: number
  level?: number
  sort?: number
  status?: number
}): Promise<Department> => {
  return request.post('/departments', data)
}

// 更新部门
export const updateDepartment = async (id: number, data: Partial<{
  name: string
  code: string
  parent_id?: number
  level?: number
  sort?: number
  status?: number
}>): Promise<Department> => {
  return request.put(`/departments/${id}`, data)
}

// 删除部门
export const deleteDepartment = async (id: number): Promise<void> => {
  return request.delete(`/departments/${id}`)
}

// 获取部门成员列表
export const getDepartmentMembers = async (departmentId: number): Promise<any[]> => {
  return request.get(`/departments/${departmentId}/members`)
}

// 添加部门成员
export const addDepartmentMembers = async (departmentId: number, userIds: number[]): Promise<void> => {
  return request.post(`/departments/${departmentId}/members`, {
    user_ids: userIds
  })
}

// 移除部门成员
export const removeDepartmentMember = async (departmentId: number, userId: number): Promise<void> => {
  return request.delete(`/departments/${departmentId}/members/${userId}`)
}

