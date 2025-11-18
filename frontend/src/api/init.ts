import request from '../utils/request'

export interface InitStatus {
  initialized: boolean
}

export interface WeChatConfigRequest {
  wechat_app_id: string
  wechat_app_secret: string
}

export interface InitRequest {
  code: string
  state?: string
}

export interface InitResponse {
  message: string
  token: string
  user: {
    id: number
    username: string
    avatar?: string
    roles: string[]
  }
}

export interface QRCodeResponse {
  ticket: string
  qrCodeUrl: string
  authUrl?: string  // 授权URL
  expireSeconds: number
}

export const checkInitStatus = async (): Promise<InitStatus> => {
  return request.get('/init/status')
}

export const saveWeChatConfig = async (data: WeChatConfigRequest): Promise<{ message: string }> => {
  return request.post('/init/wechat-config', data)
}

export const getInitQRCode = async (): Promise<QRCodeResponse> => {
  // 后端会优先使用配置文件中的 callback_domain
  // 如果后端未配置 callback_domain，才需要传递 redirect_uri
  // 这里不传递 redirect_uri，让后端使用配置文件中的值
  const data: any = await request.get('/init/qrcode')
  return {
    ticket: data.ticket,
    qrCodeUrl: data.qr_code_url || data.auth_url || '',
    authUrl: data.auth_url || data.qr_code_url || '',
    expireSeconds: data.expire_seconds || 600
  }
}

export const initSystem = async (data: InitRequest): Promise<InitResponse> => {
  return request.post('/init', data)
}

