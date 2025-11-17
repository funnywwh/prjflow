import { defineStore } from 'pinia'
import { ref } from 'vue'
import { login, getUserInfo } from '../api/auth'

export interface User {
  id: number
  username: string
  email?: string
  avatar?: string
  roles?: string[]
}

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

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    isAuthenticated.value = false
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

