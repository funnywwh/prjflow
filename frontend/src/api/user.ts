import request from '../utils/request'

export interface User {
  id: number
  wechat_open_id?: string
  username: string
  email?: string
  avatar?: string
  phone?: string
  status: number
  department_id?: number
  department?: Department
  roles?: Role[]
  created_at?: string
  updated_at?: string
}

export interface Department {
  id: number
  name: string
  code: string
  parent_id?: number
  parent?: Department
  children?: Department[]
  level: number
  sort: number
  status: number
  created_at?: string
  updated_at?: string
}

export interface Role {
  id: number
  name: string
  code: string
  description?: string
  status: number
  permissions?: Permission[]
  created_at?: string
  updated_at?: string
}

export interface Permission {
  id: number
  code: string
  name: string
  resource?: string
  action?: string
  description?: string
  status: number
  created_at?: string
  updated_at?: string
}

export interface UserListResponse {
  list: User[]
  total: number
  page: number
  size: number
}

export interface CreateUserRequest {
  username: string
  email?: string
  phone?: string
  avatar?: string
  status?: number
  department_id?: number
}

export interface UpdateUserRequest extends CreateUserRequest {
  id: number
}

// 获取用户列表
export const getUsers = async (params?: {
  keyword?: string
  department_id?: number
  page?: number
  size?: number
}): Promise<UserListResponse> => {
  return request.get('/users', { params })
}

// 获取用户详情
export const getUser = async (id: number): Promise<User> => {
  return request.get(`/users/${id}`)
}

// 创建用户
export const createUser = async (data: CreateUserRequest): Promise<User> => {
  return request.post('/users', data)
}

// 更新用户
export const updateUser = async (id: number, data: Partial<CreateUserRequest>): Promise<User> => {
  return request.put(`/users/${id}`, data)
}

// 删除用户
export const deleteUser = async (id: number): Promise<void> => {
  return request.delete(`/users/${id}`)
}

// 扫码添加用户
export interface AddUserByWeChatRequest {
  code: string
  state?: string
}

export interface AddUserByWeChatResponse {
  user: {
    id: number
    username: string
    email?: string
    avatar?: string
    wechat_open_id?: string
  }
}

export const addUserByWeChat = async (data: AddUserByWeChatRequest): Promise<AddUserByWeChatResponse> => {
  return request.post('/users/wechat/add', data)
}

