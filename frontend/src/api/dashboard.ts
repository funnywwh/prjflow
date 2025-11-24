import request from '../utils/request'

export interface DashboardData {
  tasks: {
    todo: number
    in_progress: number
    done: number
  }
  bugs: {
    open: number
    in_progress: number
    resolved: number
  }
  requirements: {
    in_progress: number
    completed: number
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

