import request from '../utils/request'

export interface WeChatConfig {
  wechat_app_id: string
  wechat_app_secret: string
}

export interface WeChatConfigRequest {
  wechat_app_id: string
  wechat_app_secret: string
}

// 获取微信配置
export const getWeChatConfig = async (): Promise<WeChatConfig> => {
  return request.get('/system/wechat-config')
}

// 保存微信配置
export const saveWeChatConfig = async (data: WeChatConfigRequest): Promise<{ message: string }> => {
  return request.post('/system/wechat-config', data)
}

