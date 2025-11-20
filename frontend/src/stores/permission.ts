import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserPermissions } from '@/api/permission'

export interface Permission {
  code: string
  name: string
  resource: string
}

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([]) // 存储权限代码列表
  const roles = ref<string[]>([])

  const setPermissions = (perms: string[]) => {
    permissions.value = perms
  }

  const setRoles = (roleList: string[]) => {
    roles.value = roleList
  }

  // 加载用户权限
  const loadPermissions = async () => {
    try {
      const permCodes = await getUserPermissions()
      console.log('从API获取的权限代码:', permCodes)
      setPermissions(permCodes)
      console.log('设置后的权限列表:', permissions.value)
    } catch (error) {
      console.error('加载权限失败:', error)
      permissions.value = []
    }
  }

  const hasPermission = (code: string): boolean => {
    // 管理员拥有所有权限
    if (roles.value.includes('admin')) {
      return true
    }
    return permissions.value.includes(code)
  }

  const hasRole = (role: string): boolean => {
    return roles.value.includes(role)
  }

  const hasAnyPermission = (codes: string[]): boolean => {
    return codes.some(code => hasPermission(code))
  }

  const hasAnyRole = (roleList: string[]): boolean => {
    return roleList.some(role => hasRole(role))
  }

  // 清空权限（退出登录时调用）
  const clearPermissions = () => {
    permissions.value = []
    roles.value = []
  }

  return {
    permissions,
    roles,
    setPermissions,
    setRoles,
    loadPermissions,
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAnyRole,
    clearPermissions
  }
})

