<template>
  <div class="dashboard">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="个人工作台" />
          
          <a-spin :spinning="loading">
            <!-- 统计概览 -->
            <a-row :gutter="16" class="stats-row">
              <a-col :span="6">
                <a-card
                  class="stat-card clickable-stat-card"
                  @click="goToAllTasks"
                >
                  <a-statistic title="总任务数" :value="statistics.total_tasks" />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card
                  class="stat-card clickable-stat-card"
                  @click="goToAllBugs"
                >
                  <a-statistic title="总Bug数" :value="statistics.total_bugs" />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card
                  class="stat-card clickable-stat-card"
                  @click="goToAllRequirements"
                >
                  <a-statistic title="总需求数" :value="statistics.total_requirements" />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card
                  class="stat-card clickable-stat-card"
                  @click="goToAllProjects"
                >
                  <a-statistic title="参与项目" :value="statistics.total_projects" />
                </a-card>
              </a-col>
            </a-row>

            <a-row :gutter="16" class="stats-row">
              <a-col :span="12">
                <a-card class="stat-card" @click="goToResourceAllocations('week')">
                  <a-statistic title="本周工时" :value="statistics.week_hours" suffix="小时" :precision="1" />
                </a-card>
              </a-col>
              <a-col :span="12">
                <a-card class="stat-card" @click="goToResourceAllocations('month')">
                  <a-statistic title="本月工时" :value="statistics.month_hours" suffix="小时" :precision="1" />
                </a-card>
              </a-col>
            </a-row>

            <!-- Tab标签页 -->
            <a-card class="dashboard-card" :bordered="false">
              <a-tabs v-model:activeKey="activeTab" type="card" class="centered-tabs">
                <a-tab-pane key="projects" tab="我的项目">
                  <div class="project-list-container">
                    <a-list
                      :data-source="projects"
                      :loading="loading"
                    >
                      <template #renderItem="{ item }">
                        <a-list-item @click="goToProject(item.id)">
                          <a-list-item-meta>
                            <template #title>
                              {{ item.name }}
                            </template>
                            <template #description>
                              <a-tag>{{ item.role }}</a-tag>
                              <span style="margin-left: 8px;">{{ item.code }}</span>
                            </template>
                          </a-list-item-meta>
                        </a-list-item>
                      </template>
                    </a-list>
                  </div>
                </a-tab-pane>

                <a-tab-pane key="bugs" tab="我的Bug">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-card
                        class="stat-card"
                        @click="goToBugs('active')"
                      >
                        <a-statistic
                          title="激活"
                          :value="bugs.active"
                          :value-style="{ color: '#ff4d4f' }"
                        />
                      </a-card>
                    </a-col>
                    <a-col :span="8">
                      <a-card
                        class="stat-card"
                        @click="goToBugs('resolved')"
                      >
                        <a-statistic
                          title="已解决"
                          :value="bugs.resolved"
                          :value-style="{ color: '#faad14' }"
                        />
                      </a-card>
                    </a-col>
                    <a-col :span="8">
                      <a-card
                        class="stat-card"
                        @click="goToBugs('closed')"
                      >
                        <a-statistic
                          title="已关闭"
                          :value="bugs.closed"
                          :value-style="{ color: '#52c41a' }"
                        />
                      </a-card>
                    </a-col>
                  </a-row>
                </a-tab-pane>

                <a-tab-pane key="tasks" tab="我的任务">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-card
                        class="stat-card todo-card"
                        @click="goToTasks('todo')"
                      >
                        <a-statistic
                          title="待办"
                          :value="tasks.todo"
                          :value-style="{ color: '#1890ff' }"
                        />
                      </a-card>
                    </a-col>
                    <a-col :span="8">
                      <a-card
                        class="stat-card in-progress-card"
                        @click="goToTasks('in_progress')"
                      >
                        <a-statistic
                          title="进行中"
                          :value="tasks.in_progress"
                          :value-style="{ color: '#faad14' }"
                        />
                      </a-card>
                    </a-col>
                    <a-col :span="8">
                      <a-card
                        class="stat-card done-card"
                        @click="goToTasks('done')"
                      >
                        <a-statistic
                          title="已完成"
                          :value="tasks.done"
                          :value-style="{ color: '#52c41a' }"
                        />
                      </a-card>
                    </a-col>
                  </a-row>
                </a-tab-pane>

                <a-tab-pane key="resources" tab="我的资源分配">
                  <div>
                    <a-row :gutter="16" style="margin-bottom: 16px;">
                      <a-col :span="12">
                        <a-card class="stat-card" @click="goToResourceAllocations('week')">
                          <a-statistic
                            title="本周工时"
                            :value="statistics.week_hours"
                            suffix="小时"
                            :precision="1"
                            :value-style="{ color: '#1890ff' }"
                          />
                        </a-card>
                      </a-col>
                      <a-col :span="12">
                        <a-card class="stat-card" @click="goToResourceAllocations('month')">
                          <a-statistic
                            title="本月工时"
                            :value="statistics.month_hours"
                            suffix="小时"
                            :precision="1"
                            :value-style="{ color: '#52c41a' }"
                          />
                        </a-card>
                      </a-col>
                    </a-row>
                    
                    <!-- 最近的资源分配记录 -->
                    <a-list
                      :data-source="resourceAllocations"
                      :loading="resourceLoading"
                      :pagination="false"
                      size="small"
                    >
                      <template #renderItem="{ item }">
                        <a-list-item>
                          <a-list-item-meta>
                            <template #title>
                              <span>{{ formatDate(item.date) }}</span>
                              <a-tag color="blue" style="margin-left: 8px;">{{ item.hours }}小时</a-tag>
                            </template>
                            <template #description>
                              <div>
                                <span v-if="item.project?.name" style="margin-right: 8px;">
                                  <a-tag>{{ item.project.name }}</a-tag>
                                </span>
                                <span v-if="item.task?.title" style="margin-right: 8px;">
                                  任务: {{ item.task.title }}
                                </span>
                                <span v-if="item.bug?.title" style="margin-right: 8px;">
                                  Bug: {{ item.bug.title }}
                                </span>
                                <div v-if="item.description" style="margin-top: 4px; color: #666; font-size: 12px;">
                                  {{ item.description.substring(0, 50) }}{{ item.description.length > 50 ? '...' : '' }}
                                </div>
                              </div>
                            </template>
                          </a-list-item-meta>
                        </a-list-item>
                      </template>
                      <template #empty>
                        <a-empty description="暂无资源分配记录" />
                      </template>
                    </a-list>
                    <div style="text-align: center; margin-top: 16px;">
                      <a-button type="link" @click="goToResourceAllocations('all')">查看全部资源分配</a-button>
                    </div>
                  </div>
                </a-tab-pane>

                <a-tab-pane key="reports">
                  <template #tab>
                    <a-badge 
                      :count="reports.pending_approval" 
                      :number-style="{ backgroundColor: '#ff4d4f' }"
                      :show-zero="false"
                    >
                      <span>工作报告</span>
                    </a-badge>
                  </template>
                  <div>
                    <a-row :gutter="16" style="margin-bottom: 16px;">
                      <a-col :span="8">
                        <a-card
                          class="stat-card"
                          @click="goToReports('pending')"
                        >
                          <a-statistic
                            title="待提交"
                            :value="reports.pending"
                            :value-style="{ color: '#faad14' }"
                          />
                        </a-card>
                      </a-col>
                      <a-col :span="8">
                        <a-card
                          class="stat-card"
                          @click="goToReports('submitted')"
                        >
                          <a-statistic
                            title="已提交"
                            :value="reports.submitted"
                            :value-style="{ color: '#52c41a' }"
                          />
                        </a-card>
                      </a-col>
                      <a-col :span="8">
                        <a-card
                          class="stat-card"
                          @click="goToReports('approval')"
                        >
                          <a-statistic
                            title="待审批"
                            :value="reports.pending_approval"
                            :value-style="{ color: '#1890ff' }"
                          />
                        </a-card>
                      </a-col>
                    </a-row>
                    
                    <!-- 报告列表 -->
                    <a-tabs v-model:activeKey="reportTab" size="small">
                      <a-tab-pane key="daily" tab="日报">
                        <a-list
                          :data-source="dailyReports"
                          :loading="reportLoading"
                          :pagination="false"
                          size="small"
                        >
                          <template #renderItem="{ item }">
                            <a-list-item @click="goToReportDetail('daily', item.id)">
                              <a-list-item-meta>
                                <template #title>
                                  <span>{{ formatDate(item.date) }}</span>
                                  <a-tag :color="getReportStatusColor(item.status)" style="margin-left: 8px;">
                                    {{ getReportStatusText(item.status) }}
                                  </a-tag>
                                </template>
                                <template #description>
                                  <div v-if="item.content" style="max-height: 40px; overflow: hidden; text-overflow: ellipsis;">
                                    {{ item.content.replace(/[#*`]/g, '').substring(0, 50) }}{{ item.content.length > 50 ? '...' : '' }}
                                  </div>
                                  <div v-else style="color: #999;">暂无内容</div>
                                </template>
                              </a-list-item-meta>
                            </a-list-item>
                          </template>
                          <template #empty>
                            <a-empty description="暂无日报" />
                          </template>
                        </a-list>
                        <div style="text-align: center; margin-top: 16px;">
                          <a-button type="link" @click="goToReports('draft')">查看全部日报</a-button>
                        </div>
                      </a-tab-pane>
                      <a-tab-pane key="weekly" tab="周报">
                        <a-list
                          :data-source="weeklyReports"
                          :loading="reportLoading"
                          :pagination="false"
                          size="small"
                        >
                          <template #renderItem="{ item }">
                            <a-list-item @click="goToReportDetail('weekly', item.id)">
                              <a-list-item-meta>
                                <template #title>
                                  <span>{{ formatDate(item.week_start) }} ~ {{ formatDate(item.week_end) }}</span>
                                  <a-tag :color="getReportStatusColor(item.status)" style="margin-left: 8px;">
                                    {{ getReportStatusText(item.status) }}
                                  </a-tag>
                                </template>
                                <template #description>
                                  <div v-if="item.summary" style="max-height: 40px; overflow: hidden; text-overflow: ellipsis;">
                                    {{ item.summary.replace(/[#*`]/g, '').substring(0, 50) }}{{ item.summary.length > 50 ? '...' : '' }}
                                  </div>
                                  <div v-else style="color: #999;">暂无内容</div>
                                </template>
                              </a-list-item-meta>
                            </a-list-item>
                          </template>
                          <template #empty>
                            <a-empty description="暂无周报" />
                          </template>
                        </a-list>
                        <div style="text-align: center; margin-top: 16px;">
                          <a-button type="link" @click="goToReports('draft')">查看全部周报</a-button>
                        </div>
                      </a-tab-pane>
                    </a-tabs>
                  </div>
                </a-tab-pane>
              </a-tabs>
            </a-card>
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { getDashboard, type DashboardData } from '@/api/dashboard'
import { getDailyReports, getWeeklyReports, type DailyReport, type WeeklyReport } from '@/api/report'
import { getResourceAllocations, type ResourceAllocation } from '@/api/resource'
import { useAuthStore } from '@/stores/auth'
import AppHeader from '@/components/AppHeader.vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const reportLoading = ref(false)
const resourceLoading = ref(false)
const activeTab = ref('projects')
const reportTab = ref('daily')
const dailyReports = ref<DailyReport[]>([])
const weeklyReports = ref<WeeklyReport[]>([])
const resourceAllocations = ref<ResourceAllocation[]>([])
const dashboardData = ref<DashboardData>({
  tasks: { todo: 0, in_progress: 0, done: 0 },
  bugs: { active: 0, resolved: 0, closed: 0 },
  requirements: { in_progress: 0, completed: 0 },
  projects: [],
  reports: { pending: 0, submitted: 0, pending_approval: 0 },
  statistics: {
    total_tasks: 0,
    total_bugs: 0,
    total_requirements: 0,
    total_projects: 0,
    week_hours: 0,
    month_hours: 0
  }
})

const tasks = ref(dashboardData.value.tasks)
const bugs = ref(dashboardData.value.bugs)
const requirements = ref(dashboardData.value.requirements)
const projects = ref(dashboardData.value.projects)
const reports = ref(dashboardData.value.reports)
const statistics = ref(dashboardData.value.statistics)

const loadDashboard = async () => {
  loading.value = true
  try {
    const data = await getDashboard()
    dashboardData.value = data
    tasks.value = data.tasks
    bugs.value = data.bugs
    requirements.value = data.requirements
    projects.value = data.projects
    reports.value = data.reports
    statistics.value = data.statistics
  } catch (error) {
    message.error('加载工作台数据失败')
  } finally {
    loading.value = false
  }
}

// 跳转到任务列表
const goToTasks = (status: string) => {
  router.push({
    path: '/task',
    query: { status, assignee: 'me' }
  })
}

// 跳转到Bug列表
const goToBugs = (status: string) => {
  router.push({
    path: '/bug',
    query: { status, assignee: 'me' }
  })
}

// 跳转到项目详情
const goToProject = (projectId: number) => {
  router.push({
    path: `/project/${projectId}`
  })
}

// 跳转到工作报告
const goToReports = (status: string) => {
  if (status === 'approval') {
    router.push({
      path: '/reports',
      query: { tab: 'approval' }
    })
  } else {
    router.push({
      path: '/reports',
      query: { status }
    })
  }
}

// 跳转到所有任务
const goToAllTasks = () => {
  router.push({
    path: '/task',
    query: { assignee: 'me' }
  })
}

// 跳转到所有Bug
const goToAllBugs = () => {
  router.push({
    path: '/bug',
    query: { assignee: 'me' }
  })
}

// 跳转到所有需求
const goToAllRequirements = () => {
  router.push({
    path: '/requirement',
    query: { assignee: 'me' }
  })
}

// 跳转到所有项目
const goToAllProjects = () => {
  router.push({
    path: '/project'
  })
}

// 加载报告列表
const loadReports = async () => {
  reportLoading.value = true
  try {
    // 加载最近的日报（最多5条）
    const dailyRes = await getDailyReports({ page: 1, size: 5 })
    dailyReports.value = dailyRes.list

    // 加载最近的周报（最多5条）
    const weeklyRes = await getWeeklyReports({ page: 1, size: 5 })
    weeklyReports.value = weeklyRes.list
  } catch (error) {
    // 静默失败，不影响主界面
  } finally {
    reportLoading.value = false
  }
}

// 格式化日期
const formatDate = (date: string | Date) => {
  if (!date) return ''
  return dayjs(date).format('YYYY-MM-DD')
}

// 获取报告状态颜色
const getReportStatusColor = (status: string) => {
  const colorMap: Record<string, string> = {
    draft: 'default',
    submitted: 'processing',
    approved: 'success',
    rejected: 'error'
  }
  return colorMap[status] || 'default'
}

// 获取报告状态文本
const getReportStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    draft: '草稿',
    submitted: '已提交',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return textMap[status] || status
}

// 跳转到报告详情
const goToReportDetail = (type: 'daily' | 'weekly', id: number) => {
  if (type === 'daily') {
    router.push(`/reports/daily/${id}`)
  } else {
    router.push(`/reports/weekly/${id}`)
  }
}

// 监听报告Tab切换，加载对应数据
watch(reportTab, () => {
  if (activeTab.value === 'reports') {
    loadReports()
  }
})

// 加载资源分配列表
const loadResourceAllocations = async () => {
  resourceLoading.value = true
  try {
    // 获取当前用户ID
    const userId = authStore.user?.id
    if (!userId) return

    // 计算本周开始和结束日期
    const now = dayjs()
    const weekStart = now.startOf('week').add(1, 'day') // 周一开始
    const weekEnd = weekStart.add(6, 'days')

    // 加载本周的资源分配（最多10条）
    const res = await getResourceAllocations({
      user_id: userId,
      start_date: weekStart.format('YYYY-MM-DD'),
      end_date: weekEnd.format('YYYY-MM-DD'),
      page: 1,
      size: 10
    })
    resourceAllocations.value = res.list
  } catch (error) {
    // 静默失败，不影响主界面
  } finally {
    resourceLoading.value = false
  }
}

// 跳转到资源分配页面
const goToResourceAllocations = (type: 'week' | 'month' | 'all') => {
  if (type === 'week') {
    router.push({
      path: '/resource-allocation',
      query: { period: 'week' }
    })
  } else if (type === 'month') {
    router.push({
      path: '/resource-allocation',
      query: { period: 'month' }
    })
  } else {
    router.push('/resource-allocation')
  }
}

// 监听主Tab切换
watch(activeTab, (newTab) => {
  if (newTab === 'reports') {
    loadReports()
  } else if (newTab === 'resources') {
    loadResourceAllocations()
  }
})

onMounted(() => {
  loadDashboard()
  // 加载用户信息
  if (!authStore.user && authStore.isAuthenticated) {
    authStore.loadUserInfo()
  }
})
</script>

<style scoped>
.dashboard {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dashboard :deep(.ant-layout) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  flex: 1;
  height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-inner {
  background: white;
  padding: 24px;
  border-radius: 4px;
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  height: 0;
}

.stats-row {
  margin-bottom: 24px;
}

.dashboard-card {
  margin-bottom: 24px;
}

.stat-card {
  cursor: pointer;
  transition: all 0.3s;
  text-align: center;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.clickable-stat-card {
  cursor: pointer;
  transition: all 0.3s;
}

.clickable-stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
  border-color: #1890ff;
}

:deep(.ant-list-item) {
  cursor: pointer;
  transition: background-color 0.3s;
}

:deep(.ant-list-item):hover {
  background-color: #f5f5f5;
}

.todo-card:hover {
  border-color: #1890ff;
}

.in-progress-card:hover {
  border-color: #faad14;
}

.done-card:hover {
  border-color: #52c41a;
}

.centered-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 16px;
}

.centered-tabs :deep(.ant-tabs-nav-list) {
  width: 100%;
  display: flex;
  justify-content: center;
}

.project-list-container {
  max-height: calc(100vh - 500px);
  overflow-y: auto;
}
</style>
