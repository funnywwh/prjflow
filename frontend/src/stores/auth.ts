import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types/user'
import { getUserInfo, logout as logoutAPI } from '../api/auth'
import { usePermissionStore } from './permission'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)
  const isAuthenticated = ref<boolean>(!!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
    isAuthenticated.value = true
  }

  const setUser = (userData: User) => {
    user.value = userData
    // 设置用户角色
    const permissionStore = usePermissionStore()
    console.log('设置用户角色，roles:', userData.roles)
    if (userData.roles) {
      permissionStore.setRoles(userData.roles)
      console.log('设置后的角色列表:', permissionStore.roles.value)
    } else {
      console.warn('用户数据中没有角色信息')
    }
  }

  const logout = async () => {
    try {
      // 调用后端退出登录API
      await logoutAPI()
    } catch (error) {
      // 即使API调用失败，也清除本地状态
      console.error('退出登录API调用失败:', error)
    } finally {
      // 清除本地状态
      token.value = null
      user.value = null
      localStorage.removeItem('token')
      isAuthenticated.value = false
    }
  }

  const loadUserInfo = async () => {
    if (!token.value) return
    
    try {
      const userData = await getUserInfo()
      console.log('获取到的用户信息:', userData)
      setUser(userData)
      // 加载用户权限
      const permissionStore = usePermissionStore()
      await permissionStore.loadPermissions()
      console.log('加载权限后的权限列表:', permissionStore.permissions.value)
      console.log('加载权限后的角色列表:', permissionStore.roles.value)
    } catch (error) {
      console.error('Failed to load user info:', error)
      logout()
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    setToken,
    setUser,
    logout,
    loadUserInfo
  }
})

