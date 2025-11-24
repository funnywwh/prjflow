<template>
  <div class="progress-view">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="`进度跟踪 - ${project?.name || ''}`"
            @back="() => router.push(`/project/${projectId}`)"
          >
            <template #extra>
              <a-button @click="handleRefresh">刷新</a-button>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <!-- 统计概览 -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总任务数"
                    :value="progressData?.statistics?.total_tasks || 0"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="已完成任务"
                    :value="progressData?.statistics?.done_tasks || 0"
                    :value-style="{ color: '#52c41a' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总Bug数"
                    :value="progressData?.statistics?.total_bugs || 0"
                    :value-style="{ color: '#ff4d4f' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总需求数"
                    :value="progressData?.statistics?.total_requirements || 0"
                    :value-style="{ color: '#722ed1' }"
                  />
                </a-card>
              </a-col>
            </a-row>

            <!-- 任务状态分布 -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :span="12">
                <a-card title="任务状态分布" :bordered="false">
                  <div ref="taskStatusChartRef" style="height: 300px"></div>
                </a-card>
              </a-col>
              <a-col :span="12">
                <a-card title="任务优先级分布" :bordered="false">
                  <div ref="taskPriorityChartRef" style="height: 300px"></div>
                </a-card>
              </a-col>
            </a-row>

            <!-- 任务完成率趋势 -->
            <a-card title="任务完成率趋势（按周）" :bordered="false" style="margin-bottom: 16px">
              <div ref="completionTrendChartRef" style="height: 300px"></div>
            </a-card>

            <!-- 任务进度趋势 -->
            <a-card title="任务进度趋势（最近30天）" :bordered="false" style="margin-bottom: 16px">
              <div ref="progressTrendChartRef" style="height: 300px"></div>
            </a-card>

            <!-- Bug和需求趋势 -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :span="12">
                <a-card title="Bug趋势（最近30天）" :bordered="false">
                  <div ref="bugTrendChartRef" style="height: 300px"></div>
                </a-card>
              </a-col>
              <a-col :span="12">
                <a-card title="需求趋势（最近30天）" :bordered="false">
                  <div ref="requirementTrendChartRef" style="height: 300px"></div>
                </a-card>
              </a-col>
            </a-row>

            <!-- 成员工作量统计 -->
            <a-card title="成员工作量统计" :bordered="false">
              <a-table
                :scroll="{ x: 'max-content' }"
                :columns="memberColumns"
                :data-source="progressData?.member_workload || []"
                :pagination="false"
                row-key="user_id"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'username'">
                    {{ record.username }}{{ record.nickname ? `(${record.nickname})` : '' }}
                  </template>
                  <template v-else-if="column.key === 'completion_rate'">
                    <a-progress
                      :percent="record.total > 0 ? Math.round((record.completed / record.total) * 100) : 0"
                      :status="record.total > 0 && record.completed === record.total ? 'success' : 'active'"
                    />
                  </template>
                </template>
              </a-table>
            </a-card>
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import * as echarts from 'echarts'
import AppHeader from '@/components/AppHeader.vue'
import { getProjectProgress, getProject, type Project, type ProjectProgressData } from '@/api/project'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const progressData = ref<ProjectProgressData | null>(null)
const project = ref<Project | null>(null)
const projectId = ref<number>(0)

// 图表引用
const taskStatusChartRef = ref<HTMLElement>()
const taskPriorityChartRef = ref<HTMLElement>()
const completionTrendChartRef = ref<HTMLElement>()
const progressTrendChartRef = ref<HTMLElement>()
const bugTrendChartRef = ref<HTMLElement>()
const requirementTrendChartRef = ref<HTMLElement>()

// 图表实例
let taskStatusChart: echarts.ECharts | null = null
let taskPriorityChart: echarts.ECharts | null = null
let completionTrendChart: echarts.ECharts | null = null
let progressTrendChart: echarts.ECharts | null = null
let bugTrendChart: echarts.ECharts | null = null
let requirementTrendChart: echarts.ECharts | null = null

const memberColumns = [
  { title: '成员', key: 'username', width: 200 },
  { title: '总任务数', dataIndex: 'total', key: 'total', width: 100 },
  { title: '进行中', dataIndex: 'in_progress', key: 'in_progress', width: 100 },
  { title: '已完成', dataIndex: 'completed', key: 'completed', width: 100 },
  { title: '完成率', key: 'completion_rate', width: 200 }
]

// 任务状态分布图表
const taskStatusChartOption = computed(() => {
  if (!progressData.value?.task_status_distribution) {
    return {}
  }

  const data = progressData.value.task_status_distribution.map(item => ({
    value: item.count,
    name: getStatusText(item.status)
  }))

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '任务状态',
        type: 'pie',
        radius: '50%',
        data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
})

// 任务优先级分布图表
const taskPriorityChartOption = computed(() => {
  if (!progressData.value?.task_priority_distribution) {
    return {}
  }

  const data = progressData.value.task_priority_distribution.map(item => ({
    value: item.count,
    name: getPriorityText(item.priority)
  }))

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '任务优先级',
        type: 'pie',
        radius: '50%',
        data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
})

// 任务完成率趋势图表
const completionTrendChartOption = computed(() => {
  if (!progressData.value?.task_completion_trend) {
    return {}
  }

  const weeks = progressData.value.task_completion_trend.map(item => item.week)
  const completionRates = progressData.value.task_completion_trend.map(item => item.completion_rate.toFixed(2))

  return {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: weeks
    },
    yAxis: {
      type: 'value',
      name: '完成率(%)',
      max: 100
    },
    series: [
      {
        name: '完成率',
        type: 'line',
        data: completionRates,
        smooth: true,
        itemStyle: {
          color: '#52c41a'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(82, 196, 26, 0.3)' },
              { offset: 1, color: 'rgba(82, 196, 26, 0.1)' }
            ]
          }
        }
      }
    ]
  }
})

// 任务进度趋势图表
const progressTrendChartOption = computed(() => {
  if (!progressData.value?.task_progress_trend) {
    return {}
  }

  const dates = progressData.value.task_progress_trend.map(item => item.date)
  const averages = progressData.value.task_progress_trend.map(item => item.average.toFixed(2))

  return {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value',
      name: '平均进度(%)',
      max: 100
    },
    series: [
      {
        name: '平均进度',
        type: 'line',
        data: averages,
        smooth: true,
        itemStyle: {
          color: '#1890ff'
        }
      }
    ]
  }
})

// Bug趋势图表
const bugTrendChartOption = computed(() => {
  if (!progressData.value?.bug_trend) {
    return {}
  }

  const dates = progressData.value.bug_trend.map(item => item.date)
  const counts = progressData.value.bug_trend.map(item => item.count)

  return {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value',
      name: 'Bug数量'
    },
    series: [
      {
        name: 'Bug数量',
        type: 'bar',
        data: counts,
        itemStyle: {
          color: '#ff4d4f'
        }
      }
    ]
  }
})

// 需求趋势图表
const requirementTrendChartOption = computed(() => {
  if (!progressData.value?.requirement_trend) {
    return {}
  }

  const dates = progressData.value.requirement_trend.map(item => item.date)
  const counts = progressData.value.requirement_trend.map(item => item.count)

  return {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value',
      name: '需求数量'
    },
    series: [
      {
        name: '需求数量',
        type: 'bar',
        data: counts,
        itemStyle: {
          color: '#722ed1'
        }
      }
    ]
  }
})

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

// 初始化图表
const initCharts = () => {
  if (taskStatusChartRef.value) {
    taskStatusChart = echarts.init(taskStatusChartRef.value)
  }
  if (taskPriorityChartRef.value) {
    taskPriorityChart = echarts.init(taskPriorityChartRef.value)
  }
  if (completionTrendChartRef.value) {
    completionTrendChart = echarts.init(completionTrendChartRef.value)
  }
  if (progressTrendChartRef.value) {
    progressTrendChart = echarts.init(progressTrendChartRef.value)
  }
  if (bugTrendChartRef.value) {
    bugTrendChart = echarts.init(bugTrendChartRef.value)
  }
  if (requirementTrendChartRef.value) {
    requirementTrendChart = echarts.init(requirementTrendChartRef.value)
  }
}

// 更新图表
const updateCharts = () => {
  if (taskStatusChart && taskStatusChartOption.value) {
    taskStatusChart.setOption(taskStatusChartOption.value)
  }
  if (taskPriorityChart && taskPriorityChartOption.value) {
    taskPriorityChart.setOption(taskPriorityChartOption.value)
  }
  if (completionTrendChart && completionTrendChartOption.value) {
    completionTrendChart.setOption(completionTrendChartOption.value)
  }
  if (progressTrendChart && progressTrendChartOption.value) {
    progressTrendChart.setOption(progressTrendChartOption.value)
  }
  if (bugTrendChart && bugTrendChartOption.value) {
    bugTrendChart.setOption(bugTrendChartOption.value)
  }
  if (requirementTrendChart && requirementTrendChartOption.value) {
    requirementTrendChart.setOption(requirementTrendChartOption.value)
  }
}

// 销毁图表
const destroyCharts = () => {
  taskStatusChart?.dispose()
  taskPriorityChart?.dispose()
  completionTrendChart?.dispose()
  progressTrendChart?.dispose()
  bugTrendChart?.dispose()
  requirementTrendChart?.dispose()
  taskStatusChart = null
  taskPriorityChart = null
  completionTrendChart = null
  progressTrendChart = null
  bugTrendChart = null
  requirementTrendChart = null
}

// 加载进度数据
const loadProgressData = async () => {
  if (!projectId.value) return
  loading.value = true
  try {
    progressData.value = await getProjectProgress(projectId.value)
    // 等待DOM更新后更新图表
    setTimeout(() => {
      updateCharts()
    }, 100)
  } catch (error: any) {
    message.error(error.message || '加载进度数据失败')
  } finally {
    loading.value = false
  }
}

// 刷新
const handleRefresh = () => {
  loadProgressData()
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

// 监听数据变化，更新图表
watch(() => progressData.value, () => {
  nextTick(() => {
    updateCharts()
  })
}, { deep: true })

onMounted(() => {
  loadProject()
  nextTick(() => {
    initCharts()
    loadProgressData()
  })
})

onBeforeUnmount(() => {
  destroyCharts()
})
</script>

<style scoped>
.progress-view {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
}
</style>

