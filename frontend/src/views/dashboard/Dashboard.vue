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
                <a-statistic title="总任务数" :value="statistics.total_tasks" />
              </a-col>
              <a-col :span="6">
                <a-statistic title="总Bug数" :value="statistics.total_bugs" />
              </a-col>
              <a-col :span="6">
                <a-statistic title="总需求数" :value="statistics.total_requirements" />
              </a-col>
              <a-col :span="6">
                <a-statistic title="参与项目" :value="statistics.total_projects" />
              </a-col>
            </a-row>

            <a-row :gutter="16" class="stats-row">
              <a-col :span="12">
                <a-statistic title="本周工时" :value="statistics.week_hours" suffix="小时" :precision="1" />
              </a-col>
              <a-col :span="12">
                <a-statistic title="本月工时" :value="statistics.month_hours" suffix="小时" :precision="1" />
              </a-col>
            </a-row>

            <!-- Tab标签页 -->
            <a-card class="dashboard-card" :bordered="false">
              <a-tabs v-model:activeKey="activeTab" type="card" class="centered-tabs">
                <a-tab-pane key="projects" tab="我的项目">
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
                </a-tab-pane>

                <a-tab-pane key="bugs" tab="我的Bug">
                  <a-row :gutter="16">
                    <a-col :span="8">
                      <a-card
                        class="stat-card"
                        @click="goToBugs('open')"
                      >
                        <a-statistic
                          title="待处理"
                          :value="bugs.open"
                          :value-style="{ color: '#ff4d4f' }"
                        />
                      </a-card>
                    </a-col>
                    <a-col :span="8">
                      <a-card
                        class="stat-card"
                        @click="goToBugs('in_progress')"
                      >
                        <a-statistic
                          title="处理中"
                          :value="bugs.in_progress"
                          :value-style="{ color: '#faad14' }"
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
                  <a-row :gutter="16">
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Badge } from 'ant-design-vue'
import { getDashboard, type DashboardData } from '@/api/dashboard'
import { useAuthStore } from '@/stores/auth'
import AppHeader from '@/components/AppHeader.vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const activeTab = ref('projects')
const dashboardData = ref<DashboardData>({
  tasks: { todo: 0, in_progress: 0, done: 0 },
  bugs: { open: 0, in_progress: 0, resolved: 0 },
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
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 64px);
}

.content-inner {
  background: white;
  padding: 24px;
  border-radius: 4px;
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

:deep(.ant-list-item) {
  cursor: pointer;
  transition: all 0.3s;
}

:deep(.ant-list-item):hover {
  background-color: #f5f5f5;
  transform: translateX(4px);
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
</style>
