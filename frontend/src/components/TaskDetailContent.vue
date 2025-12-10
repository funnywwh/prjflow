<template>
  <div class="task-detail-content">
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

      <!-- 附件 -->
      <AttachmentList :attachments="task?.attachments || []" />

      <!-- 历史记录 -->
      <a-card :bordered="false" style="margin-bottom: 16px">
        <template #title>
          <span>历史记录</span>
          <a-button 
            type="link" 
            size="small"
            @click.stop="handleAddNote" 
            :disabled="historyLoading"
            style="margin-left: 8px; padding: 0"
          >
            添加备注
          </a-button>
        </template>
        <a-spin :spinning="historyLoading" :style="{ minHeight: '100px' }">
          <a-timeline v-if="historyList.length > 0">
            <a-timeline-item
              v-for="(action, index) in historyList"
              :key="action.id"
            >
              <template #dot>
                <span style="font-weight: bold; color: #1890ff">{{ historyList.length - index }}</span>
              </template>
              <div>
                <div style="margin-bottom: 8px">
                  <span style="color: #666; margin-right: 8px">{{ formatDateTime(action.date) }}</span>
                  <span>{{ getActionDescription(action) }}</span>
                  <a-button
                    v-if="hasHistoryDetails(action)"
                    type="link"
                    size="small"
                    @click="toggleHistoryDetail(action.id)"
                    style="padding: 0; height: auto; margin-left: 8px"
                  >
                    {{ expandedHistoryIds.has(action.id) ? '收起' : '展开' }}
                  </a-button>
                </div>
                <!-- 字段变更详情和备注内容（可折叠） -->
                <div
                  v-show="expandedHistoryIds.has(action.id)"
                  style="margin-left: 24px; margin-top: 8px"
                >
                  <!-- 字段变更详情 -->
                  <div v-if="action.histories && action.histories.length > 0">
                    <div
                      v-for="history in action.histories"
                      :key="history.id"
                      style="margin-bottom: 8px; color: #666"
                    >
                      <div>修改了{{ getFieldDisplayName(history.field) }}</div>
                      <div style="margin-left: 16px; margin-top: 4px;">
                        <div>旧值："{{ history.old_value || history.old || '-' }}"</div>
                        <div>新值："{{ history.new_value || history.new || '-' }}"</div>
                      </div>
                    </div>
                  </div>
                  <!-- 备注内容 -->
                  <div v-if="action.comment" style="margin-top: 8px; color: #666">
                    {{ action.comment }}
                  </div>
                </div>
              </div>
            </a-timeline-item>
          </a-timeline>
          <a-empty v-else description="暂无历史记录" />
        </a-spin>
      </a-card>

      <!-- 依赖任务 -->
      <a-card title="依赖任务" :bordered="false" v-if="task?.dependencies && task.dependencies.length > 0">
        <a-list :data-source="task.dependencies" :bordered="false">
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta>
                <template #title>
                  <a @click="handleGoToTask(item.id)" style="cursor: pointer">
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
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { formatDateTime } from '@/utils/date'
import dayjs from 'dayjs'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentList from '@/components/AttachmentList.vue'
import type { Task, Action } from '@/api/task'

interface Props {
  task?: Task | null
  loading?: boolean
  historyList?: Action[]
  historyLoading?: boolean
}

withDefaults(defineProps<Props>(), {
  loading: false,
  historyList: () => [],
  historyLoading: false
})

const emit = defineEmits<{
  addNote: []
  goToTask: [taskId: number]
}>()

const expandedHistoryIds = ref<Set<number>>(new Set())

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    wait: 'orange',
    doing: 'blue',
    done: 'green',
    pause: 'purple',
    cancel: 'default',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    wait: '未开始',
    doing: '进行中',
    done: '已完成',
    pause: '已暂停',
    cancel: '已取消',
    closed: '已关闭'
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
const isOverdue = (dueDate: string | undefined, status: string | undefined): boolean => {
  if (!dueDate || status === 'done' || status === 'closed' || status === 'cancel') {
    return false
  }
  const due = dayjs(dueDate)
  const now = dayjs()
  return due.isBefore(now, 'day')
}

// 获取字段显示名称
const getFieldDisplayName = (fieldName: string): string => {
  const fieldNames: Record<string, string> = {
    title: '任务标题',
    description: '任务描述',
    status: '状态',
    priority: '优先级',
    project_id: '项目',
    assignee_id: '负责人',
    start_date: '开始日期',
    end_date: '结束日期',
    due_date: '截止日期',
    estimated_hours: '预估工时',
    actual_hours: '实际工时',
    progress: '进度'
  }
  return fieldNames[fieldName] || fieldName
}

// 获取操作描述
const getActionDescription = (action: Action): string => {
  const actorName = action.actor
    ? `${action.actor.username}${action.actor.nickname ? `(${action.actor.nickname})` : ''}`
    : '系统'

  switch (action.action) {
    case 'created':
      return `由 ${actorName} 创建。`
    case 'edited':
      return `由 ${actorName} 编辑。`
    case 'commented':
      return `由 ${actorName} 添加了备注：${action.comment || ''}`
    default:
      return `由 ${actorName} 执行了 ${action.action} 操作。`
  }
}

// 判断历史记录是否有详情
const hasHistoryDetails = (action: Action): boolean => {
  return !!(action.histories && action.histories.length > 0) || !!action.comment
}

// 切换历史记录详情展开/收起
const toggleHistoryDetail = (actionId: number) => {
  const newSet = new Set(expandedHistoryIds.value)
  if (newSet.has(actionId)) {
    newSet.delete(actionId)
  } else {
    newSet.add(actionId)
  }
  expandedHistoryIds.value = newSet
}

// 事件处理
const handleAddNote = () => {
  emit('addNote')
}


const handleGoToTask = (taskId: number) => {
  emit('goToTask', taskId)
}
</script>

<style scoped>
.task-detail-content {
  width: 100%;
}

.markdown-content {
  max-height: 500px;
  overflow-y: auto;
}

</style>

