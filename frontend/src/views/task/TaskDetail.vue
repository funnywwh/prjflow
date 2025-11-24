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
                  {{ formatDateTime(task?.created_at) }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ formatDateTime(task?.updated_at) }}
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
      :mask-closable="true"
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
            :disabled="autoProgress"
          />
          <span style="margin-left: 8px">{{ progressFormData.progress || 0 }}%</span>
          <span v-if="autoProgress" style="margin-left: 8px; color: #999">（根据工时自动计算）</span>
        </a-form-item>
        <a-form-item label="预估工时" name="estimated_hours">
          <a-input-number
            v-model:value="progressFormData.estimated_hours"
            placeholder="预估工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="实际工时" name="actual_hours">
          <a-input-number
            v-model:value="progressFormData.actual_hours"
            placeholder="实际工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配并计算进度</span>
        </a-form-item>
        <a-form-item label="工作日期" name="work_date" v-if="progressFormData.actual_hours">
          <a-date-picker
            v-model:value="progressFormData.work_date"
            placeholder="选择工作日期（默认今天）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getTask,
  updateTaskStatus,
  deleteTask,
  updateTaskProgress,
  type Task,
  type UpdateTaskProgressRequest
} from '@/api/task'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<Task | null>(null)
const progressModalVisible = ref(false)
const progressFormRef = ref()
const progressFormData = reactive<{
  progress?: number
  estimated_hours?: number
  actual_hours?: number
  work_date?: Dayjs
}>({
  progress: undefined,
  estimated_hours: undefined,
  actual_hours: undefined,
  work_date: undefined
})

const progressFormRules = {
  // progress不再是必填项，因为可以通过工时自动计算
}

// 自动计算进度（实际工时/预估工时 * 100）
const autoProgress = computed(() => {
  if (progressFormData.estimated_hours && progressFormData.estimated_hours > 0 && progressFormData.actual_hours) {
    const progress = Math.min(100, Math.max(0, Math.round((progressFormData.actual_hours / progressFormData.estimated_hours) * 100)))
    progressFormData.progress = progress
    return true
  }
  return false
})

// 监听实际工时和预估工时的变化，自动计算进度
watch([() => progressFormData.actual_hours, () => progressFormData.estimated_hours], () => {
  if (progressFormData.estimated_hours && progressFormData.estimated_hours > 0 && progressFormData.actual_hours) {
    const progress = Math.min(100, Math.max(0, Math.round((progressFormData.actual_hours / progressFormData.estimated_hours) * 100)))
    progressFormData.progress = progress
  }
})

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
  progressFormData.estimated_hours = task.value.estimated_hours
  progressFormData.actual_hours = task.value.actual_hours // 显示当前实际工时
  progressFormData.work_date = dayjs() // 默认今天
  progressModalVisible.value = true
}

// 进度提交
const handleProgressSubmit = async () => {
  if (!task.value) return
  try {
    await progressFormRef.value.validate()
    const data: UpdateTaskProgressRequest = {
      progress: progressFormData.progress,
      estimated_hours: progressFormData.estimated_hours,
      actual_hours: progressFormData.actual_hours,
      work_date: progressFormData.work_date && progressFormData.work_date.isValid() ? progressFormData.work_date.format('YYYY-MM-DD') : undefined
    }
    await updateTaskProgress(task.value.id, data)
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
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
}

.markdown-content {
  min-height: 200px;
}
</style>

