import request from '../utils/request'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

export interface Attachment {
  id: number
  file_name: string
  file_path: string
  file_size: number
  mime_type: string
  creator_id: number
  creator?: {
    id: number
    username: string
    nickname: string
  }
  created_at?: string
  updated_at?: string
}

export interface AttachToEntityRequest {
  project_id?: number
  requirement_id?: number
  task_id?: number
  bug_id?: number
}

// 上传文件
export const uploadFile = async (file: File, projectId: number, onProgress?: (progress: number) => void): Promise<Attachment> => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('project_id', projectId.toString())

  const authStore = useAuthStore()
  const token = authStore.token

  return new Promise((resolve, reject) => {
    axios.post('/api/attachments/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      }
    }).then((response) => {
      const res = response.data
      if (res.code !== undefined) {
        if (res.code === 200) {
          resolve(res.data)
        } else {
          reject(new Error(res.message || '上传失败'))
        }
      } else {
        resolve(res)
      }
    }).catch((error) => {
      reject(error)
    })
  })
}

// 获取附件信息
export const getAttachment = async (id: number): Promise<Attachment> => {
  return request.get(`/attachments/${id}`)
}

// 下载文件
export const downloadFile = async (id: number, fileName: string): Promise<void> => {
  const authStore = useAuthStore()
  const token = authStore.token

  const response = await axios.get(`/api/attachments/${id}/download`, {
    responseType: 'blob',
    headers: {
      'Authorization': token ? `Bearer ${token}` : ''
    }
  })

  // 创建下载链接
  const url = window.URL.createObjectURL(new Blob([response.data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', fileName)
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}

// 删除附件
export const deleteAttachment = async (id: number): Promise<void> => {
  return request.delete(`/attachments/${id}`)
}

// 获取附件列表
export const getAttachments = async (params?: {
  project_id?: number
  requirement_id?: number
  task_id?: number
  bug_id?: number
}): Promise<Attachment[]> => {
  return request.get('/attachments', { params })
}

// 关联附件到实体
export const attachToEntity = async (id: number, data: AttachToEntityRequest): Promise<Attachment> => {
  return request.post(`/attachments/${id}/attach`, data)
}

