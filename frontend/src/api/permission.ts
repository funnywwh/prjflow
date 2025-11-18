import request from '../utils/request'
import type { Role, Permission } from './user'

// 角色相关API
export const getRoles = async (): Promise<Role[]> => {
  return request.get('/permissions/roles')
}

export const createRole = async (data: {
  name: string
  code: string
  description?: string
  status?: number
}): Promise<Role> => {
  return request.post('/permissions/roles', data)
}

export const updateRole = async (id: number, data: Partial<{
  name: string
  code: string
  description?: string
  status?: number
}>): Promise<Role> => {
  return request.put(`/permissions/roles/${id}`, data)
}

export const deleteRole = async (id: number): Promise<void> => {
  return request.delete(`/permissions/roles/${id}`)
}

// 权限相关API
export const getPermissions = async (): Promise<Permission[]> => {
  return request.get('/permissions/permissions')
}

export const createPermission = async (data: {
  code: string
  name: string
  resource?: string
  action?: string
  description?: string
  status?: number
}): Promise<Permission> => {
  return request.post('/permissions/permissions', data)
}

// 分配角色权限
export const assignRolePermissions = async (roleId: number, permissionIds: number[]): Promise<void> => {
  return request.post(`/permissions/roles/${roleId}/permissions`, {
    permission_ids: permissionIds
  })
}

// 分配用户角色
export const assignUserRoles = async (userId: number, roleIds: number[]): Promise<void> => {
  return request.post(`/permissions/users/${userId}/roles`, {
    role_ids: roleIds
  })
}

