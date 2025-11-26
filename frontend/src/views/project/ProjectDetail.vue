<template>
  <div class="project-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="project?.name || '项目详情'"
            :sub-title="project?.code"
            @back="() => router.push('/project')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleManageRequirements">需求管理</a-button>
                <a-button @click="handleManageTasks">任务管理</a-button>
                <a-button @click="handleManageBugs">Bug管理</a-button>
                <a-button @click="handleViewBoards">看板</a-button>
                <a-button @click="handleViewGantt">甘特图</a-button>
                <a-button @click="handleViewProgress">进度跟踪</a-button>
                <a-button @click="handleManageModules">功能模块</a-button>
                <a-button @click="handleEdit">编辑</a-button>
                <a-button @click="handleManageMembers">成员管理</a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <!-- 项目基本信息 -->
            <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
              <a-descriptions :column="2" bordered>
                <a-descriptions-item label="项目名称">{{ project?.name }}</a-descriptions-item>
                <a-descriptions-item label="项目编码">{{ project?.code }}</a-descriptions-item>
                <!-- <a-descriptions-item label="项目集">{{ project?.project_group?.name || '-' }}</a-descriptions-item> -->
                <!-- <a-descriptions-item label="关联产品">{{ project?.product?.name || '-' }}</a-descriptions-item> -->
                <a-descriptions-item label="开始日期">{{ project?.start_date || '-' }}</a-descriptions-item>
                <a-descriptions-item label="结束日期">{{ project?.end_date || '-' }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="project?.status === 1 ? 'green' : 'red'">
                    {{ project?.status === 1 ? '正常' : '禁用' }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="成员数">{{ statistics?.total_members || 0 }} 人</a-descriptions-item>
                <a-descriptions-item label="描述" :span="2">
                  {{ project?.description || '-' }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 统计概览 -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总任务数"
                    :value="statistics?.total_tasks || 0"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总Bug数"
                    :value="statistics?.total_bugs || 0"
                    :value-style="{ color: '#ff4d4f' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总需求数"
                    :value="statistics?.total_requirements || 0"
                    :value-style="{ color: '#52c41a' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="项目成员"
                    :value="statistics?.total_members || 0"
                    suffix="人"
                    :value-style="{ color: '#722ed1' }"
                  />
                </a-card>
              </a-col>
            </a-row>

            <!-- 任务统计 -->
            <a-card title="任务统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('todo')">
                    <a-statistic
                      title="待办"
                      :value="statistics?.todo_tasks || 0"
                      :value-style="{ color: '#1890ff' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('in_progress')">
                    <a-statistic
                      title="进行中"
                      :value="statistics?.in_progress_tasks || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('done')">
                    <a-statistic
                      title="已完成"
                      :value="statistics?.done_tasks || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- Bug统计 -->
            <a-card title="Bug统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('open')">
                    <a-statistic
                      title="待处理"
                      :value="statistics?.open_bugs || 0"
                      :value-style="{ color: '#ff4d4f' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('in_progress')">
                    <a-statistic
                      title="处理中"
                      :value="statistics?.in_progress_bugs || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('resolved')">
                    <a-statistic
                      title="已解决"
                      :value="statistics?.resolved_bugs || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- 需求统计 -->
            <a-card title="需求统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-card class="stat-card" @click="goToRequirements('in_progress')">
                    <a-statistic
                      title="进行中"
                      :value="statistics?.in_progress_requirements || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="12">
                  <a-card class="stat-card" @click="goToRequirements('completed')">
                    <a-statistic
                      title="已完成"
                      :value="statistics?.completed_requirements || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- 项目成员 -->
            <a-card title="项目成员" :bordered="false">
              <a-list
                :data-source="project?.members || []"
                :loading="loading"
              >
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #avatar>
                        <a-avatar :src="item.user?.avatar">
                          {{ (item.user?.nickname || item.user?.username)?.charAt(0).toUpperCase() }}
                        </a-avatar>
                      </template>
                      <template #title>
                        {{ item.user?.username }}{{ item.user?.nickname ? `(${item.user.nickname})` : '' }}
                      </template>
                      <template #description>
                        <a-tag>{{ item.role }}</a-tag>
                        <span v-if="item.user?.department" style="margin-left: 8px; color: #999">
                          {{ item.user.department.name }}
                        </span>
                      </template>
                    </a-list-item-meta>
                  </a-list-item>
                </template>
              </a-list>
            </a-card>
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import AppHeader from '@/components/AppHeader.vue'
import { getProject, type ProjectDetailResponse, type Project } from '@/api/project'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const project = ref<Project>()
const statistics = ref<any>()

// 加载项目详情
const loadProject = async () => {
  const projectId = Number(route.params.id)
  if (!projectId) {
    message.error('项目ID无效')
    router.push('/project')
    return
  }

  loading.value = true
  try {
    const data: ProjectDetailResponse = await getProject(projectId)
    project.value = data.project
    statistics.value = data.statistics
  } catch (error: any) {
    message.error(error.message || '加载项目详情失败')
    router.push('/project')
  } finally {
    loading.value = false
  }
}

// 查看看板
const handleViewBoards = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/boards`)
}

// 查看甘特图
const handleViewGantt = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/gantt`)
}

// 查看进度跟踪
const handleViewProgress = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/progress`)
}

// 编辑项目
const handleEdit = () => {
  router.push({
    path: '/project',
    query: { edit: project.value?.id }
  })
}

// 成员管理
const handleManageMembers = () => {
  router.push({
    path: '/project',
    query: { manageMembers: project.value?.id }
  })
}

// 功能模块管理
const handleManageModules = () => {
  router.push({
    path: '/project',
    query: { manageModules: project.value?.id }
  })
}

// 需求管理
const handleManageRequirements = () => {
  if (!project.value) return
  router.push({
    path: '/requirement',
    query: { project_id: project.value.id }
  })
}

// 任务管理
const handleManageTasks = () => {
  if (!project.value) return
  router.push({
    path: '/task',
    query: { project_id: project.value.id }
  })
}

// Bug管理
const handleManageBugs = () => {
  if (!project.value) return
  router.push({
    path: '/bug',
    query: { project_id: project.value.id }
  })
}

// 跳转到任务列表
const goToTasks = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/task',
    query: { status, project_id: project.value.id }
  })
}

// 跳转到Bug列表
const goToBugs = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/bug',
    query: { status, project_id: project.value.id }
  })
}

// 跳转到需求列表
const goToRequirements = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/requirement',
    query: { status, project_id: project.value.id }
  })
}

onMounted(() => {
  loadProject()
})
</script>

<style scoped>
.project-detail {
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

.stat-card {
  cursor: pointer;
  transition: all 0.3s;
  text-align: center;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
</style>

