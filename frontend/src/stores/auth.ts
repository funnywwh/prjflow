import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types/user'
import { getUserInfo, logout as logoutAPI } from '../api/auth'

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
      setUser(userData)
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

