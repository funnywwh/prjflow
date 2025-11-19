import request from '../utils/request'

export interface Tag {
  id: number
  name: string
  description?: string
  color?: string
  created_at?: string
  updated_at?: string
}

export interface CreateTagRequest {
  name: string
  description?: string
  color?: string
}

export interface UpdateTagRequest {
  name?: string
  description?: string
  color?: string
}

// 标签相关API
export const getTags = async (): Promise<Tag[]> => {
  return request.get('/tags')
}

export const getTag = async (id: number): Promise<Tag> => {
  return request.get(`/tags/${id}`)
}

export const createTag = async (data: CreateTagRequest): Promise<Tag> => {
  return request.post('/tags', data)
}

export const updateTag = async (id: number, data: UpdateTagRequest): Promise<Tag> => {
  return request.put(`/tags/${id}`, data)
}

export const deleteTag = async (id: number): Promise<void> => {
  return request.delete(`/tags/${id}`)
}

