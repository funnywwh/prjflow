<template>
  <div class="task-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="task?.title || '任务详情'"
            @back="() => router.push('/task')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEdit">编辑</a-button>
                <a-button @click="handleUpdateProgress">更新进度</a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="todo">待办</a-menu-item>
                      <a-menu-item key="in_progress">进行中</a-menu-item>
                      <a-menu-item key="done">已完成</a-menu-item>
                      <a-menu-item key="cancelled">已取消</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
                <a-popconfirm
                  title="确定要删除这个任务吗？"
                  @confirm="handleDelete"
                >
                  <a-button danger>删除</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <!-- 基本信息 -->
            <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
              <a-descriptions :column="2" bordered>
                <a-descriptions-item label="任务标题">{{ task?.title }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="getStatusColor(task?.status || '')">
                    {{ getStatusText(task?.status || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="优先级">
                  <a-tag :color="getPriorityColor(task?.priority || '')">
                    {{ getPriorityText(task?.priority || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="进度">
                  <a-progress :percent="task?.progress || 0" :status="task?.status === 'done' ? 'success' : 'active'" />
                </a-descriptions-item>
                <a-descriptions-item label="项目">
                  {{ task?.project?.name || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="负责人">
                  {{ task?.assignee ? `${task.assignee.username}${task.assignee.nickname ? `(${task.assignee.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="开始日期">
                  {{ task?.start_date || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="结束日期">
                  {{ task?.end_date || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="截止日期">
                  <span :style="{ color: isOverdue(task?.due_date, task?.status) ? 'red' : '' }">
                    {{ task?.due_date || '-' }}
                  </span>
                </a-descriptions-item>
                <a-descriptions-item label="创建人">
                  {{ task?.creator ? `${task.creator.username}${task.creator.nickname ? `(${task.creator.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建时间">
                  {{ task?.created_at ? new Date(task.created_at).toLocaleString('zh-CN') : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ task?.updated_at ? new Date(task.updated_at).toLocaleString('zh-CN') : '-' }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 任务描述 -->
            <a-card title="任务描述" :bordered="false" style="margin-bottom: 16px">
              <div v-if="task?.description" class="markdown-content">
                <MarkdownEditor
                  :model-value="task.description"
                  :readonly="true"
                />
              </div>
              <a-empty v-else description="暂无描述" />
            </a-card>

            <!-- 依赖任务 -->
            <a-card title="依赖任务" :bordered="false" v-if="task?.dependencies && task.dependencies.length > 0">
              <a-list :data-source="task.dependencies" :bordered="false">
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #title>
                        <a @click="router.push(`/task/${item.id}`)" style="cursor: pointer">
                          {{ item.title }}
                        </a>
                      </template>
                      <template #description>
                        <a-tag :color="getStatusColor(item.status)">{{ getStatusText(item.status) }}</a-tag>
                        <a-tag :color="getPriorityColor(item.priority)" style="margin-left: 8px">
                          {{ getPriorityText(item.priority) }}
                        </a-tag>
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

    <!-- 更新进度模态框 -->
    <a-modal
      v-model:open="progressModalVisible"
      title="更新任务进度"
      @ok="handleProgressSubmit"
      @cancel="handleProgressCancel"
    >
      <a-form
        ref="progressFormRef"
        :model="progressFormData"
        :rules="progressFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="进度" name="progress">
          <a-slider
            v-model:value="progressFormData.progress"
            :min="0"
            :max="100"
            :marks="{ 0: '0%', 50: '50%', 100: '100%' }"
          />
          <span style="margin-left: 8px">{{ progressFormData.progress }}%</span>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getTask,
  updateTaskStatus,
  deleteTask,
  updateTaskProgress,
  type Task
} from '@/api/task'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<Task | null>(null)
const progressModalVisible = ref(false)
const progressFormRef = ref()
const progressFormData = reactive({
  progress: 0
})

const progressFormRules = {
  progress: [{ required: true, message: '请设置进度', trigger: 'change' }]
}

// 加载任务详情
const loadTask = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('任务ID无效')
    router.push('/task')
    return
  }

  loading.value = true
  try {
    task.value = await getTask(id)
  } catch (error: any) {
    message.error(error.message || '加载任务详情失败')
    router.push('/task')
  } finally {
    loading.value = false
  }
}

// 编辑
const handleEdit = () => {
  router.push(`/task?edit=${task.value?.id}`)
}

// 更新进度
const handleUpdateProgress = () => {
  if (!task.value) return
  progressFormData.progress = task.value.progress
  progressModalVisible.value = true
}

// 进度提交
const handleProgressSubmit = async () => {
  if (!task.value) return
  try {
    await progressFormRef.value.validate()
    await updateTaskProgress(task.value.id, { progress: progressFormData.progress })
    message.success('进度更新成功')
    progressModalVisible.value = false
    loadTask()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '进度更新失败')
  }
}

// 进度取消
const handleProgressCancel = () => {
  progressFormRef.value?.resetFields()
}

// 状态变更
const handleStatusChange = async (status: string) => {
  if (!task.value) return
  try {
    await updateTaskStatus(task.value.id, { status: status as any })
    message.success('状态更新成功')
    loadTask()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 删除
const handleDelete = async () => {
  if (!task.value) return
  try {
    await deleteTask(task.value.id)
    message.success('删除成功')
    router.push('/task')
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    todo: 'orange',
    in_progress: 'blue',
    done: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    todo: '待办',
    in_progress: '进行中',
    done: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || status
}

// 获取优先级颜色
const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    urgent: 'red'
  }
  return colors[priority] || 'default'
}

// 获取优先级文本
const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

// 判断是否逾期
const isOverdue = (dueDate?: string, status?: string) => {
  if (!dueDate || status === 'done' || status === 'cancelled') {
    return false
  }
  const due = dayjs(dueDate)
  const now = dayjs()
  return due.isBefore(now, 'day')
}

onMounted(() => {
  loadTask()
})
</script>

<style scoped>
.task-detail {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 1400px;
  margin: 0 auto;
}

.markdown-content {
  min-height: 200px;
}
</style>

