import request from '../utils/request'
import type { Role, Permission } from './user'

// 重新导出类型
export type { Role, Permission }

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

export const getPermission = async (id: number): Promise<Permission> => {
  return request.get(`/permissions/permissions/${id}`)
}

export const createPermission = async (data: {
  code: string
  name: string
  resource?: string
  action?: string
  description?: string
  status?: number
  is_menu?: boolean
  menu_path?: string
  menu_icon?: string
  menu_title?: string
  parent_menu_id?: number
  menu_order?: number
}): Promise<Permission> => {
  return request.post('/permissions/permissions', data)
}

export const updatePermission = async (id: number, data: Partial<{
  code: string
  name: string
  resource?: string
  action?: string
  description?: string
  status?: number
  is_menu?: boolean
  menu_path?: string
  menu_icon?: string
  menu_title?: string
  parent_menu_id?: number
  menu_order?: number
}>): Promise<Permission> => {
  return request.put(`/permissions/permissions/${id}`, data)
}

export const deletePermission = async (id: number): Promise<void> => {
  return request.delete(`/permissions/permissions/${id}`)
}

// 获取角色权限
export const getRolePermissions = async (roleId: number): Promise<Permission[]> => {
  return request.get(`/permissions/roles/${roleId}/permissions`)
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

// 获取当前用户的权限列表
export const getUserPermissions = async (): Promise<string[]> => {
  return request.get('/permissions/me')
}

// 获取菜单树
export interface MenuItem {
  id?: number
  key: string
  title: string
  icon?: string
  path?: string
  permission?: string
  order?: number
  children?: MenuItem[]
}

export const getMenus = async (): Promise<MenuItem[]> => {
  return request.get('/permissions/menus')
}

