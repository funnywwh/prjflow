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

