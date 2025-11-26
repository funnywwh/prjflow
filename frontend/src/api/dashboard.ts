import request from '../utils/request'

export interface DashboardData {
  tasks: {
    todo: number  // 对应wait状态
    in_progress: number  // 对应doing状态
    done: number
  }
  bugs: {
    active: number   // 激活状态
    resolved: number // 已解决
    closed: number   // 已关闭
  }
  requirements: {
    in_progress: number  // 对应active状态
    completed: number  // 对应closed状态
  }
  projects: Array<{
    id: number
    name: string
    code: string
    role: string
  }>
  reports: {
    pending: number
    submitted: number
    pending_approval: number
  }
  statistics: {
    total_tasks: number
    total_bugs: number
    total_requirements: number
    total_projects: number
    week_hours: number
    month_hours: number
  }
}

export const getDashboard = async (): Promise<DashboardData> => {
  return request.get('/dashboard')
}

