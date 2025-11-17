import request from '../utils/request'
import type { User } from '../types/user'

export interface LoginRequest {
  code: string
  state?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface QRCodeResponse {
  ticket: string
  qrCodeUrl: string
  authUrl?: string  // 授权URL
  expireSeconds: number
}

export const getWeChatQRCode = async (): Promise<QRCodeResponse> => {
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

