import request from '../utils/request'
import type { User } from '../types/user'

export interface LoginRequest {
  code: string
  state?: string
}

export interface LoginResponse {
  token: string
  user: User
  is_first_login?: boolean
}

export interface QRCodeResponse {
  ticket: string
  qrCodeUrl: string
  authUrl?: string  // 授权URL
  expireSeconds: number
}

export const getWeChatQRCode = async (): Promise<QRCodeResponse> => {
  // 后端会优先使用配置文件中的 callback_domain
  // 如果后端未配置 callback_domain，才需要传递 redirect_uri
  // 这里不传递 redirect_uri，让后端使用配置文件中的值
  const data: any = await request.get('/auth/wechat/qrcode')
  return {
    ticket: data.ticket || '',
    qrCodeUrl: data.qr_code_url || data.auth_url || '',
    authUrl: data.auth_url || data.qr_code_url || '',
    expireSeconds: data.expire_seconds || 600
  }
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

// 用户名密码登录
export interface PasswordLoginRequest {
  username: string
  password: string
}

export const passwordLogin = async (data: PasswordLoginRequest): Promise<LoginResponse> => {
  return request.post('/auth/login', data)
}

// 修改密码
export interface ChangePasswordRequest {
  old_password?: string // 可选，如果用户没有密码则不需要
  new_password: string
}

export const changePassword = async (data: ChangePasswordRequest): Promise<{ message: string }> => {
  return request.post('/auth/change-password', data)
}

// 微信绑定相关
export const getWeChatBindQRCode = async (): Promise<QRCodeResponse> => {
  const data: any = await request.get('/auth/wechat/bind/qrcode')
  return {
    ticket: data.ticket || '',
    qrCodeUrl: data.qr_code_url || data.auth_url || '',
    authUrl: data.auth_url || data.qr_code_url || '',
    expireSeconds: data.expire_seconds || 600
  }
}

export const unbindWeChat = async (): Promise<{ message: string }> => {
  return request.post('/auth/wechat/unbind')
}

