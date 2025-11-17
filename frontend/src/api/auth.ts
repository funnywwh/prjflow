import request from '../utils/request'
import { User } from '../stores/auth'

export interface LoginRequest {
  code: string
  state?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export const getWeChatQRCode = async () => {
  return request.get('/auth/wechat/qrcode')
}

export const login = async (data: LoginRequest): Promise<LoginResponse> => {
  return request.post('/auth/wechat/login', data)
}

export const getUserInfo = async (): Promise<User> => {
  return request.get('/auth/user/info')
}

export const logout = async () => {
  return request.post('/auth/logout')
}

