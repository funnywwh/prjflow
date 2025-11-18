import request from '../utils/request'

export interface BoardColumn {
  id: number
  name: string
  color?: string
  sort: number
  board_id: number
  status: string
  created_at?: string
  updated_at?: string
}

export interface Board {
  id: number
  name: string
  description?: string
  project_id: number
  project?: any
  columns?: BoardColumn[]
  created_at?: string
  updated_at?: string
}

export interface BoardTasksResponse {
  board: Board
  tasks_by_column: Record<number, any[]>
}

export interface CreateBoardRequest {
  name: string
  description?: string
  columns?: Array<{
    name: string
    color?: string
    status: string
    sort?: number
  }>
}

export interface CreateBoardColumnRequest {
  name: string
  color?: string
  status: string
  sort?: number
}

export interface MoveTaskRequest {
  column_id: string
  position?: number
}

// 看板相关API
export const getProjectBoards = async (projectId: number): Promise<Board[]> => {
  return request.get(`/projects/${projectId}/boards`)
}

export const getBoard = async (id: number): Promise<Board> => {
  return request.get(`/boards/${id}`)
}

export const getBoardTasks = async (id: number): Promise<BoardTasksResponse> => {
  return request.get(`/boards/${id}/tasks`)
}

export const createBoard = async (projectId: number, data: CreateBoardRequest): Promise<Board> => {
  return request.post(`/projects/${projectId}/boards`, data)
}

export const updateBoard = async (id: number, data: Partial<CreateBoardRequest>): Promise<Board> => {
  return request.put(`/boards/${id}`, data)
}

export const deleteBoard = async (id: number): Promise<void> => {
  return request.delete(`/boards/${id}`)
}

export const createBoardColumn = async (boardId: number, data: CreateBoardColumnRequest): Promise<BoardColumn> => {
  return request.post(`/boards/${boardId}/columns`, data)
}

export const updateBoardColumn = async (boardId: number, columnId: number, data: Partial<CreateBoardColumnRequest>): Promise<BoardColumn> => {
  return request.put(`/boards/${boardId}/columns/${columnId}`, data)
}

export const deleteBoardColumn = async (boardId: number, columnId: number): Promise<void> => {
  return request.delete(`/boards/${boardId}/columns/${columnId}`)
}

export const moveTask = async (boardId: number, taskId: number, data: MoveTaskRequest): Promise<any> => {
  return request.patch(`/boards/${boardId}/tasks/${taskId}/move`, data)
}

