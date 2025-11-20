import request from '../utils/request'

export interface Module {
  id: number
  name: string
  code?: string
  description?: string
  status: number
  sort: number
  created_at?: string
  updated_at?: string
}

export const getModules = async (params?: {
  keyword?: string
  status?: number
}): Promise<Module[]> => {
  return request.get('/modules', { params })
}

export const getModule = async (id: number): Promise<Module> => {
  return request.get(`/modules/${id}`)
}

export const createModule = async (data: {
  name: string
  code?: string
  description?: string
  status?: number
  sort?: number
}): Promise<Module> => {
  return request.post('/modules', data)
}

export const updateModule = async (id: number, data: {
  name?: string
  code?: string
  description?: string
  status?: number
  sort?: number
}): Promise<Module> => {
  return request.put(`/modules/${id}`, data)
}

export const deleteModule = async (id: number): Promise<void> => {
  return request.delete(`/modules/${id}`)
}

