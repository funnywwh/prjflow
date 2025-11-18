<template>
  <div class="gantt-view">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="`甘特图 - ${project?.name || ''}`"
            @back="() => router.push(`/project/${projectId}`)"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleRefresh">刷新</a-button>
                <a-button @click="handleViewTask" v-if="selectedTask">查看任务</a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <div class="gantt-container" v-if="tasks.length > 0">
              <!-- 甘特图时间轴 -->
              <div class="gantt-header">
                <div class="gantt-task-column" style="width: 300px; border-right: 1px solid #d9d9d9;">
                  <div class="gantt-header-cell" style="height: 60px; border-bottom: 1px solid #d9d9d9;">
                    <strong>任务</strong>
                  </div>
                </div>
                <div class="gantt-timeline" ref="timelineRef">
                  <div class="gantt-header-cell" style="height: 60px; border-bottom: 1px solid #d9d9d9;">
                    <div class="timeline-months">
                      <div
                        v-for="month in months"
                        :key="month.key"
                        class="timeline-month"
                        :style="{ width: month.width + 'px' }"
                      >
                        {{ month.label }}
                      </div>
                    </div>
                    <div class="timeline-days">
                      <div
                        v-for="day in days"
                        :key="day.key"
                        class="timeline-day"
                        :style="{ width: dayWidth + 'px' }"
                        :class="{ 'is-weekend': day.isWeekend, 'is-today': day.isToday }"
                      >
                        {{ day.date }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 甘特图内容 -->
              <div class="gantt-body">
                <div class="gantt-task-column" style="width: 300px; border-right: 1px solid #d9d9d9;">
                  <div
                    v-for="task in tasks"
                    :key="task.id"
                    class="gantt-task-row"
                    :class="{ 'selected': selectedTask?.id === task.id }"
                    @click="handleSelectTask(task)"
                  >
                    <div class="task-info">
                      <div class="task-title">{{ task.title }}</div>
                      <div class="task-meta">
                        <a-tag :color="getStatusColor(task.status)" size="small">
                          {{ getStatusText(task.status) }}
                        </a-tag>
                        <a-tag :color="getPriorityColor(task.priority)" size="small" style="margin-left: 4px">
                          {{ getPriorityText(task.priority) }}
                        </a-tag>
                        <span v-if="task.assignee" class="task-assignee">{{ task.assignee }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="gantt-timeline" ref="timelineBodyRef">
                  <div
                    v-for="task in tasks"
                    :key="task.id"
                    class="gantt-task-row"
                    :class="{ 'selected': selectedTask?.id === task.id }"
                    @click="handleSelectTask(task)"
                  >
                    <div class="gantt-bars">
                      <!-- 任务条 -->
                      <div
                        v-if="task.start_date && task.end_date"
                        class="gantt-bar"
                        :class="getTaskBarClass(task)"
                        :style="getTaskBarStyle(task)"
                        :title="getTaskTooltip(task)"
                      >
                        <div class="gantt-bar-progress" :style="getProgressStyle(task)"></div>
                        <div class="gantt-bar-label">{{ task.title }}</div>
                      </div>
                      <!-- 依赖关系线 -->
                      <svg
                        v-if="task.dependencies && task.dependencies.length > 0"
                        class="dependency-lines"
                        :style="getDependencyStyle(task)"
                      >
                        <line
                          v-for="depId in task.dependencies"
                          :key="depId"
                          :x1="getDependencyX1(depId)"
                          :y1="getDependencyY1(depId)"
                          :x2="getDependencyX2(task)"
                          :y2="getDependencyY2(task)"
                          stroke="#1890ff"
                          stroke-width="2"
                          marker-end="url(#arrowhead)"
                        />
                        <defs>
                          <marker id="arrowhead" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                            <polygon points="0 0, 10 3, 0 6" fill="#1890ff" />
                          </marker>
                        </defs>
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <a-empty v-else description="暂无任务数据" />
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import { getProjectGantt, getProject, type GanttTask, type Project } from '@/api/project'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const tasks = ref<GanttTask[]>([])
const project = ref<Project | null>(null)
const projectId = ref<number>(0)
const selectedTask = ref<GanttTask | null>(null)
const timelineRef = ref<HTMLElement>()
const timelineBodyRef = ref<HTMLElement>()

const dayWidth = 30 // 每天的宽度（像素）
const rowHeight = 60 // 每行的高度（像素）

// 计算时间范围
const timeRange = computed(() => {
  if (tasks.value.length === 0) {
    const today = dayjs()
    return {
      start: today.subtract(1, 'month'),
      end: today.add(3, 'months')
    }
  }

  let minDate: dayjs.Dayjs | null = null
  let maxDate: dayjs.Dayjs | null = null

  tasks.value.forEach(task => {
    if (task.start_date) {
      const start = dayjs(task.start_date)
      if (!minDate || start.isBefore(minDate)) {
        minDate = start
      }
    }
    if (task.end_date) {
      const end = dayjs(task.end_date)
      if (!maxDate || end.isAfter(maxDate)) {
        maxDate = end
      }
    }
    if (task.due_date) {
      const due = dayjs(task.due_date)
      if (!maxDate || due.isAfter(maxDate)) {
        maxDate = due
      }
    }
  })

  if (!minDate) minDate = dayjs().subtract(1, 'month')
  if (!maxDate) maxDate = dayjs().add(3, 'months')

  return { start: minDate, end: maxDate }
})

// 计算月份
const months = computed(() => {
  const result: Array<{ key: string; label: string; width: number; start: dayjs.Dayjs }> = []
  let current = timeRange.value.start.startOf('month')
  const end = timeRange.value.end.endOf('month')

  while (current.isBefore(end) || current.isSame(end, 'month')) {
    const daysInMonth = current.daysInMonth()
    const width = daysInMonth * dayWidth
    result.push({
      key: current.format('YYYY-MM'),
      label: current.format('YYYY年MM月'),
      width,
      start: current
    })
    current = current.add(1, 'month')
  }

  return result
})

// 计算日期
const days = computed(() => {
  const result: Array<{ key: string; date: number; isWeekend: boolean; isToday: boolean }> = []
  let current = timeRange.value.start
  const end = timeRange.value.end
  const today = dayjs()

  while (current.isBefore(end) || current.isSame(end, 'day')) {
    result.push({
      key: current.format('YYYY-MM-DD'),
      date: current.date(),
      isWeekend: current.day() === 0 || current.day() === 6,
      isToday: current.isSame(today, 'day')
    })
    current = current.add(1, 'day')
  }

  return result
})

// 计算任务条样式
const getTaskBarStyle = (task: GanttTask) => {
  if (!task.start_date || !task.end_date) return {}
  
  const start = dayjs(task.start_date)
  const end = dayjs(task.end_date)
  const rangeStart = timeRange.value.start
  
  const left = start.diff(rangeStart, 'day') * dayWidth
  const width = end.diff(start, 'day') * dayWidth + dayWidth
  
  return {
    left: `${left}px`,
    width: `${width}px`
  }
}

// 计算进度样式
const getProgressStyle = (task: GanttTask) => {
  return {
    width: `${task.progress}%`
  }
}

// 获取任务条类名
const getTaskBarClass = (task: GanttTask) => {
  const classes = []
  if (task.status === 'done') classes.push('status-done')
  else if (task.status === 'in_progress') classes.push('status-in-progress')
  else if (task.status === 'cancelled') classes.push('status-cancelled')
  else classes.push('status-todo')
  
  if (task.priority === 'urgent') classes.push('priority-urgent')
  else if (task.priority === 'high') classes.push('priority-high')
  
  return classes.join(' ')
}

// 获取任务提示
const getTaskTooltip = (task: GanttTask) => {
  let tooltip = `${task.title}\n`
  tooltip += `状态: ${getStatusText(task.status)}\n`
  tooltip += `优先级: ${getPriorityText(task.priority)}\n`
  tooltip += `进度: ${task.progress}%\n`
  if (task.start_date) tooltip += `开始: ${task.start_date}\n`
  if (task.end_date) tooltip += `结束: ${task.end_date}\n`
  if (task.due_date) tooltip += `截止: ${task.due_date}\n`
  if (task.assignee) tooltip += `负责人: ${task.assignee}`
  return tooltip
}

// 计算依赖关系样式
const getDependencyStyle = (task: GanttTask) => {
  return {
    position: 'absolute' as const,
    top: '0',
    left: '0',
    width: '100%',
    height: '100%',
    pointerEvents: 'none' as const
  }
}

// 计算依赖关系的起点和终点
const getDependencyX1 = (depId: number) => {
  const depTask = tasks.value.find(t => t.id === depId)
  if (!depTask || !depTask.end_date) return 0
  
  const end = dayjs(depTask.end_date)
  const rangeStart = timeRange.value.start
  const x = end.diff(rangeStart, 'day') * dayWidth + dayWidth
  
  return x
}

const getDependencyY1 = (depId: number) => {
  const index = tasks.value.findIndex(t => t.id === depId)
  return index * rowHeight + rowHeight / 2
}

const getDependencyX2 = (task: GanttTask) => {
  if (!task.start_date) return 0
  
  const start = dayjs(task.start_date)
  const rangeStart = timeRange.value.start
  const x = start.diff(rangeStart, 'day') * dayWidth
  
  return x
}

const getDependencyY2 = (task: GanttTask) => {
  const index = tasks.value.findIndex(t => t.id === task.id)
  return index * rowHeight + rowHeight / 2
}

// 加载项目信息
const loadProject = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('项目ID无效')
    router.push('/project')
    return
  }
  projectId.value = id
  try {
    const response = await getProject(id)
    project.value = response.project
  } catch (error: any) {
    message.error(error.message || '加载项目信息失败')
  }
}

// 加载甘特图数据
const loadGanttData = async () => {
  if (!projectId.value) return
  loading.value = true
  try {
    const response = await getProjectGantt(projectId.value)
    tasks.value = response.tasks || []
  } catch (error: any) {
    message.error(error.message || '加载甘特图数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新
const handleRefresh = () => {
  loadGanttData()
}

// 选择任务
const handleSelectTask = (task: GanttTask) => {
  selectedTask.value = task
}

// 查看任务
const handleViewTask = () => {
  if (!selectedTask.value) return
  router.push(`/task/${selectedTask.value.id}`)
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

onMounted(() => {
  loadProject()
  loadGanttData()
})
</script>

<style scoped>
.gantt-view {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  background: white;
  border-radius: 4px;
  padding: 16px;
}

.gantt-container {
  margin-top: 16px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  overflow: auto;
}

.gantt-header {
  display: flex;
  position: sticky;
  top: 0;
  background: white;
  z-index: 10;
}

.gantt-body {
  display: flex;
  position: relative;
}

.gantt-task-column {
  flex-shrink: 0;
  background: #fafafa;
}

.gantt-timeline {
  flex: 1;
  min-width: 0;
  position: relative;
}

.gantt-header-cell {
  display: flex;
  flex-direction: column;
  background: white;
}

.timeline-months {
  display: flex;
  height: 30px;
  border-bottom: 1px solid #d9d9d9;
}

.timeline-month {
  border-right: 1px solid #d9d9d9;
  padding: 4px 8px;
  font-weight: 600;
  text-align: center;
}

.timeline-days {
  display: flex;
  height: 30px;
}

.timeline-day {
  border-right: 1px solid #e8e8e8;
  padding: 2px 4px;
  text-align: center;
  font-size: 12px;
  flex-shrink: 0;
}

.timeline-day.is-weekend {
  background: #f5f5f5;
}

.timeline-day.is-today {
  background: #e6f7ff;
  font-weight: 600;
}

.gantt-task-row {
  height: 60px;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.gantt-task-row:hover {
  background: #f5f5f5;
}

.gantt-task-row.selected {
  background: #e6f7ff;
}

.task-info {
  padding: 8px 12px;
  width: 100%;
}

.task-title {
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-meta {
  display: flex;
  align-items: center;
  font-size: 12px;
}

.task-assignee {
  margin-left: 8px;
  color: #666;
}

.gantt-bars {
  position: relative;
  height: 100%;
  width: 100%;
}

.gantt-bar {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  height: 24px;
  background: #1890ff;
  border-radius: 4px;
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  padding: 0 8px;
  color: white;
  font-size: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.gantt-bar.status-todo {
  background: #faad14;
}

.gantt-bar.status-in-progress {
  background: #1890ff;
}

.gantt-bar.status-done {
  background: #52c41a;
}

.gantt-bar.status-cancelled {
  background: #ff4d4f;
  opacity: 0.6;
}

.gantt-bar.priority-urgent {
  border: 2px solid #ff4d4f;
}

.gantt-bar.priority-high {
  border: 2px solid #ff7a45;
}

.gantt-bar-progress {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: rgba(255, 255, 255, 0.3);
  transition: width 0.3s;
}

.gantt-bar-label {
  position: relative;
  z-index: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dependency-lines {
  pointer-events: none;
}
</style>

