import request from '../utils/request'

export interface TestCase {
  id: number
  name: string
  description?: string
  test_steps?: string
  types?: string[] // 测试类型（多选）
  status: 'wait' | 'normal' | 'blocked' | 'investigate'
  project_id: number
  project?: any
  creator_id: number
  creator?: any
  result?: string  // 测试结果：passed, failed, blocked（合并自TestReport）
  summary?: string  // 测试摘要（合并自TestReport）
  bugs?: any[]
  created_at?: string
  updated_at?: string
}

export interface TestCaseListResponse {
  list: TestCase[]
  total: number
  page: number
  page_size: number
}

export interface CreateTestCaseRequest {
  name: string
  description?: string
  test_steps?: string
  types?: string[] // 测试类型（多选）
  status?: 'pending' | 'running' | 'passed' | 'failed'
  result?: string  // 测试结果：passed, failed, blocked（合并自TestReport）
  summary?: string  // 测试摘要（合并自TestReport）
  project_id: number
  bug_ids?: number[]
}

export interface UpdateTestCaseRequest {
  name?: string
  description?: string
  test_steps?: string
  types?: string[] // 测试类型（多选）
  status?: 'pending' | 'running' | 'passed' | 'failed'
  result?: string  // 测试结果：passed, failed, blocked（合并自TestReport）
  summary?: string  // 测试摘要（合并自TestReport）
  bug_ids?: number[]
}

export interface TestCaseStatistics {
  total: number
  pending: number
  running: number
  passed: number
  failed: number
  pass_rate: number
  fail_rate: number
  project_stats: {
    project_id: number
    project_name: string
    total: number
    passed: number
    failed: number
    pass_rate: number
  }[]
  type_stats: {
    type: string
    total: number
    passed: number
    failed: number
    pass_rate: number
  }[]
}

// 获取测试单列表
export const getTestCases = (params?: {
  keyword?: string
  project_id?: number
  status?: string
  type?: string
  creator_id?: number
  page?: number
  size?: number
}) => {
  return request.get<TestCaseListResponse>('/test-cases', { params })
}

// 获取测试单详情
export const getTestCase = (id: number) => {
  return request.get<TestCase>(`/test-cases/${id}`)
}

// 创建测试单
export const createTestCase = (data: CreateTestCaseRequest) => {
  return request.post<TestCase>('/test-cases', data)
}

// 更新测试单
export const updateTestCase = (id: number, data: UpdateTestCaseRequest) => {
  return request.put<TestCase>(`/test-cases/${id}`, data)
}

// 删除测试单
export const deleteTestCase = (id: number) => {
  return request.delete(`/test-cases/${id}`)
}

// 更新测试单状态
export const updateTestCaseStatus = (id: number, status: string) => {
  return request.patch<TestCase>(`/test-cases/${id}/status`, { status })
}

// 获取测试单统计
export const getTestCaseStatistics = (params?: {
  project_id?: number
  keyword?: string
}) => {
  return request.get<TestCaseStatistics>('/test-cases/statistics', { params })
}

