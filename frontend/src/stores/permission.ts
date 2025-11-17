import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Permission {
  code: string
  name: string
  resource: string
}

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<Permission[]>([])
  const roles = ref<string[]>([])

  const setPermissions = (perms: Permission[]) => {
    permissions.value = perms
  }

  const setRoles = (roleList: string[]) => {
    roles.value = roleList
  }

  const hasPermission = (code: string): boolean => {
    return permissions.value.some(p => p.code === code)
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

  return {
    permissions,
    roles,
    setPermissions,
    setRoles,
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAnyRole
  }
})

