export interface User {
  id: number
  username: string
  nickname?: string
  email?: string
  avatar?: string
  roles?: string[]
  wechat_open_id?: string
  is_first_login?: boolean
}

