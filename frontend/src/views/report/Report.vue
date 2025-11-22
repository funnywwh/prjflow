<template>
  <div class="report-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="工作报告">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增{{ activeTab === 'daily' ? '日报' : '周报' }}
              </a-button>
            </template>
          </a-page-header>

          <a-tabs v-model:activeKey="activeTab" @change="handleTabChange">
            <a-tab-pane key="daily" tab="日报">
              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="dailySearchForm">
                  <a-form-item label="状态">
                    <a-select
                      v-model:value="dailySearchForm.status"
                      placeholder="选择状态"
                      allow-clear
                      style="width: 120px"
                    >
                      <a-select-option value="draft">草稿</a-select-option>
                      <a-select-option value="submitted">已提交</a-select-option>
                      <a-select-option value="approved">已审批</a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="开始日期">
                    <a-date-picker
                      v-model:value="dailySearchForm.start_date"
                      placeholder="选择开始日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="结束日期">
                    <a-date-picker
                      v-model:value="dailySearchForm.end_date"
                      placeholder="选择结束日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="项目">
                    <a-select
                      v-model:value="dailySearchForm.project_id"
                      placeholder="选择项目"
                      allow-clear
                      style="width: 150px"
                    >
                      <a-select-option
                        v-for="project in projects"
                        :key="project.id"
                        :value="project.id"
                      >
                        {{ project.name }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleDailySearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleDailyReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :scroll="{ x: 'max-content' }"
                  :columns="dailyColumns"
                  :data-source="dailyReports"
                  :loading="dailyLoading"
                  :pagination="dailyPagination"
                  row-key="id"
                  @change="handleDailyTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="getStatusColor(record.status)">
                        {{ getStatusText(record.status) }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'date'">
                      {{ formatDate(record.date) }}
                    </template>
                    <template v-else-if="column.key === 'project'">
                      {{ record.project?.name || '-' }}
                    </template>
                    <template v-else-if="column.key === 'task'">
                      {{ record.task?.title || '-' }}
                    </template>
                    <template v-else-if="column.key === 'created_at'">
                      {{ formatDateTime(record.created_at) }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleDailyEdit(record)">
                          编辑
                        </a-button>
                        <a-button
                          v-if="record.status === 'draft'"
                          type="link"
                          size="small"
                          @click="handleDailySubmit(record)"
                        >
                          提交
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个日报吗？"
                          @confirm="handleDailyDelete(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <a-tab-pane key="weekly" tab="周报">
              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="weeklySearchForm">
                  <a-form-item label="状态">
                    <a-select
                      v-model:value="weeklySearchForm.status"
                      placeholder="选择状态"
                      allow-clear
                      style="width: 120px"
                    >
                      <a-select-option value="draft">草稿</a-select-option>
                      <a-select-option value="submitted">已提交</a-select-option>
                      <a-select-option value="approved">已审批</a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item label="开始日期">
                    <a-date-picker
                      v-model:value="weeklySearchForm.start_date"
                      placeholder="选择开始日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="结束日期">
                    <a-date-picker
                      v-model:value="weeklySearchForm.end_date"
                      placeholder="选择结束日期"
                      style="width: 150px"
                      format="YYYY-MM-DD"
                    />
                  </a-form-item>
                  <a-form-item label="项目">
                    <a-select
                      v-model:value="weeklySearchForm.project_id"
                      placeholder="选择项目"
                      allow-clear
                      style="width: 150px"
                    >
                      <a-select-option
                        v-for="project in projects"
                        :key="project.id"
                        :value="project.id"
                      >
                        {{ project.name }}
                      </a-select-option>
                    </a-select>
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleWeeklySearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleWeeklyReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :scroll="{ x: 'max-content' }"
                  :columns="weeklyColumns"
                  :data-source="weeklyReports"
                  :loading="weeklyLoading"
                  :pagination="weeklyPagination"
                  row-key="id"
                  @change="handleWeeklyTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="getStatusColor(record.status)">
                        {{ getStatusText(record.status) }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'week'">
                      {{ formatDate(record.week_start) }} ~ {{ formatDate(record.week_end) }}
                    </template>
                    <template v-else-if="column.key === 'project'">
                      {{ record.project?.name || '-' }}
                    </template>
                    <template v-else-if="column.key === 'task'">
                      {{ record.task?.title || '-' }}
                    </template>
                    <template v-else-if="column.key === 'created_at'">
                      {{ formatDateTime(record.created_at) }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleWeeklyEdit(record)">
                          编辑
                        </a-button>
                        <a-button
                          v-if="record.status === 'draft'"
                          type="link"
                          size="small"
                          @click="handleWeeklySubmit(record)"
                        >
                          提交
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个周报吗？"
                          @confirm="handleWeeklyDelete(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 日报编辑/创建模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="dailyModalVisible"
      :title="dailyModalTitle"
      :width="800"
      @ok="handleDailySubmitForm"
      @cancel="handleDailyCancel"
    >
      <a-form
        ref="dailyFormRef"
        :model="dailyFormData"
        :rules="dailyFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="日期" name="date">
          <a-date-picker
            v-model:value="dailyFormData.date"
            placeholder="选择日期"
            style="width: 100%"
            format="YYYY-MM-DD"
            :disabled="!!dailyFormData.id"
          />
        </a-form-item>
        <a-form-item label="工作内容" name="content">
          <MarkdownEditor
            v-model="dailyFormData.content"
            placeholder="请输入工作内容（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="工时" name="hours">
          <a-input-number
            v-model:value="dailyFormData.hours"
            placeholder="工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="dailyFormData.project_id"
            placeholder="选择项目（可选）"
            allow-clear
            show-search
            :filter-option="filterProjectOption"
          >
            <a-select-option
              v-for="project in projects"
              :key="project.id"
              :value="project.id"
            >
              {{ project.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="任务" name="task_id">
          <a-select
            v-model:value="dailyFormData.task_id"
            placeholder="选择任务（可选）"
            allow-clear
            show-search
            :filter-option="filterTaskOption"
            :loading="taskLoading"
            :disabled="!dailyFormData.project_id"
            @focus="loadTasksForProject"
          >
            <a-select-option
              v-for="task in availableTasks"
              :key="task.id"
              :value="task.id"
            >
              {{ task.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 周报编辑/创建模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="weeklyModalVisible"
      :title="weeklyModalTitle"
      :width="800"
      @ok="handleWeeklySubmitForm"
      @cancel="handleWeeklyCancel"
    >
      <a-form
        ref="weeklyFormRef"
        :model="weeklyFormData"
        :rules="weeklyFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="周开始日期" name="week_start">
          <a-date-picker
            v-model:value="weeklyFormData.week_start"
            placeholder="选择周开始日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="周结束日期" name="week_end">
          <a-date-picker
            v-model:value="weeklyFormData.week_end"
            placeholder="选择周结束日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="工作总结" name="summary">
          <MarkdownEditor
            v-model="weeklyFormData.summary"
            placeholder="请输入工作总结（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="下周计划" name="next_week_plan">
          <MarkdownEditor
            v-model="weeklyFormData.next_week_plan"
            placeholder="请输入下周计划（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="weeklyFormData.project_id"
            placeholder="选择项目（可选）"
            allow-clear
            show-search
            :filter-option="filterProjectOption"
          >
            <a-select-option
              v-for="project in projects"
              :key="project.id"
              :value="project.id"
            >
              {{ project.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="任务" name="task_id">
          <a-select
            v-model:value="weeklyFormData.task_id"
            placeholder="选择任务（可选）"
            allow-clear
            show-search
            :filter-option="filterTaskOption"
            :loading="taskLoading"
            :disabled="!weeklyFormData.project_id"
            @focus="loadTasksForProject"
          >
            <a-select-option
              v-for="task in availableTasks"
              :key="task.id"
              :value="task.id"
            >
              {{ task.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime, formatDate } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getDailyReports,
  getDailyReport,
  createDailyReport,
  updateDailyReport,
  deleteDailyReport,
  updateDailyReportStatus,
  getWeeklyReports,
  getWeeklyReport,
  createWeeklyReport,
  updateWeeklyReport,
  deleteWeeklyReport,
  updateWeeklyReportStatus,
  type DailyReport,
  type WeeklyReport,
  type CreateDailyReportRequest,
  type CreateWeeklyReportRequest
} from '@/api/report'
import { getProjects, type Project } from '@/api/project'
import { getTasks, type Task } from '@/api/task'

const route = useRoute()
const activeTab = ref<'daily' | 'weekly'>('daily')

// 日报相关
const dailyLoading = ref(false)
const dailyReports = ref<DailyReport[]>([])
const dailySearchForm = reactive({
  status: undefined as string | undefined,
  start_date: undefined as Dayjs | undefined,
  end_date: undefined as Dayjs | undefined,
  project_id: undefined as number | undefined
})
const dailyPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})
const dailyColumns = [
  { title: '日期', key: 'date', width: 120 },
  { title: '工作内容', dataIndex: 'content', key: 'content', ellipsis: true },
  { title: '工时', dataIndex: 'hours', key: 'hours', width: 100 },
  { title: '状态', key: 'status', width: 100 },
  { title: '项目', key: 'project', width: 120 },
  { title: '任务', key: 'task', width: 150, ellipsis: true },
  { title: '创建时间', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const dailyModalVisible = ref(false)
const dailyModalTitle = ref('新增日报')
const dailyFormRef = ref()
const dailyFormData = reactive<{
  id?: number
  date?: Dayjs
  content?: string
  hours?: number
  project_id?: number
  task_id?: number
}>({
  date: undefined,
  content: '',
  hours: 0,
  project_id: undefined,
  task_id: undefined
})
const dailyFormRules = {
  date: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

// 周报相关
const weeklyLoading = ref(false)
const weeklyReports = ref<WeeklyReport[]>([])
const weeklySearchForm = reactive({
  status: undefined as string | undefined,
  start_date: undefined as Dayjs | undefined,
  end_date: undefined as Dayjs | undefined,
  project_id: undefined as number | undefined
})
const weeklyPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})
const weeklyColumns = [
  { title: '周期', key: 'week', width: 200 },
  { title: '工作总结', dataIndex: 'summary', key: 'summary', ellipsis: true },
  { title: '下周计划', dataIndex: 'next_week_plan', key: 'next_week_plan', ellipsis: true },
  { title: '状态', key: 'status', width: 100 },
  { title: '项目', key: 'project', width: 120 },
  { title: '任务', key: 'task', width: 150, ellipsis: true },
  { title: '创建时间', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const weeklyModalVisible = ref(false)
const weeklyModalTitle = ref('新增周报')
const weeklyFormRef = ref()
const weeklyFormData = reactive<{
  id?: number
  week_start?: Dayjs
  week_end?: Dayjs
  summary?: string
  next_week_plan?: string
  project_id?: number
  task_id?: number
}>({
  week_start: undefined,
  week_end: undefined,
  summary: '',
  next_week_plan: '',
  project_id: undefined,
  task_id: undefined
})
const weeklyFormRules = {
  week_start: [{ required: true, message: '请选择周开始日期', trigger: 'change' }],
  week_end: [{ required: true, message: '请选择周结束日期', trigger: 'change' }]
}

// 公共数据
const projects = ref<Project[]>([])
const availableTasks = ref<Task[]>([])
const taskLoading = ref(false)

// 加载日报列表
const loadDailyReports = async () => {
  dailyLoading.value = true
  try {
    const params: any = {
      page: dailyPagination.current,
      size: dailyPagination.pageSize
    }
    if (dailySearchForm.status) {
      params.status = dailySearchForm.status
    }
    if (dailySearchForm.start_date && dailySearchForm.start_date.isValid()) {
      params.start_date = dailySearchForm.start_date.format('YYYY-MM-DD')
    }
    if (dailySearchForm.end_date && dailySearchForm.end_date.isValid()) {
      params.end_date = dailySearchForm.end_date.format('YYYY-MM-DD')
    }
    if (dailySearchForm.project_id) {
      params.project_id = dailySearchForm.project_id
    }
    const response = await getDailyReports(params)
    dailyReports.value = response.list
    dailyPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载日报列表失败')
  } finally {
    dailyLoading.value = false
  }
}

// 加载周报列表
const loadWeeklyReports = async () => {
  weeklyLoading.value = true
  try {
    const params: any = {
      page: weeklyPagination.current,
      size: weeklyPagination.pageSize
    }
    if (weeklySearchForm.status) {
      params.status = weeklySearchForm.status
    }
    if (weeklySearchForm.start_date && weeklySearchForm.start_date.isValid()) {
      params.start_date = weeklySearchForm.start_date.format('YYYY-MM-DD')
    }
    if (weeklySearchForm.end_date && weeklySearchForm.end_date.isValid()) {
      params.end_date = weeklySearchForm.end_date.format('YYYY-MM-DD')
    }
    if (weeklySearchForm.project_id) {
      params.project_id = weeklySearchForm.project_id
    }
    const response = await getWeeklyReports(params)
    weeklyReports.value = response.list
    weeklyPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载周报列表失败')
  } finally {
    weeklyLoading.value = false
  }
}

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await getProjects()
    projects.value = response.list || []
  } catch (error: any) {
    console.error('加载项目列表失败:', error)
  }
}

// 加载任务列表（用于关联选择）
const loadTasksForProject = async () => {
  const projectId = dailyFormData.project_id || weeklyFormData.project_id
  if (!projectId) {
    availableTasks.value = []
    return
  }
  taskLoading.value = true
  try {
    const response = await getTasks({ project_id: projectId, size: 1000 })
    availableTasks.value = response.list
  } catch (error: any) {
    console.error('加载任务列表失败:', error)
  } finally {
    taskLoading.value = false
  }
}

// 标签页切换
const handleTabChange = (key: string) => {
  activeTab.value = key as 'daily' | 'weekly'
  if (key === 'daily') {
    loadDailyReports()
  } else {
    loadWeeklyReports()
  }
}

// 日报搜索
const handleDailySearch = () => {
  dailyPagination.current = 1
  loadDailyReports()
}

// 日报重置
const handleDailyReset = () => {
  dailySearchForm.status = undefined
  dailySearchForm.start_date = undefined
  dailySearchForm.end_date = undefined
  dailySearchForm.project_id = undefined
  dailyPagination.current = 1
  loadDailyReports()
}

// 日报表格变化
const handleDailyTableChange = (pag: any) => {
  dailyPagination.current = pag.current
  dailyPagination.pageSize = pag.pageSize
  loadDailyReports()
}

// 创建日报
const handleCreate = () => {
  if (activeTab.value === 'daily') {
    dailyModalTitle.value = '新增日报'
    dailyFormData.id = undefined
    dailyFormData.date = dayjs()
    dailyFormData.content = ''
    dailyFormData.hours = 0
    dailyFormData.project_id = undefined
    dailyFormData.task_id = undefined
    dailyModalVisible.value = true
  } else {
    weeklyModalTitle.value = '新增周报'
    weeklyFormData.id = undefined
    // 默认设置为本周
    const today = dayjs()
    weeklyFormData.week_start = today.startOf('week').add(1, 'day') // 周一
    weeklyFormData.week_end = today.endOf('week').add(1, 'day') // 周日
    weeklyFormData.summary = ''
    weeklyFormData.next_week_plan = ''
    weeklyFormData.project_id = undefined
    weeklyFormData.task_id = undefined
    weeklyModalVisible.value = true
  }
}

// 编辑日报
const handleDailyEdit = (record: DailyReport) => {
  dailyModalTitle.value = '编辑日报'
  dailyFormData.id = record.id
  dailyFormData.date = dayjs(record.date)
  dailyFormData.content = record.content || ''
  dailyFormData.hours = record.hours
  dailyFormData.project_id = record.project_id
  dailyFormData.task_id = record.task_id
  dailyModalVisible.value = true
  if (dailyFormData.project_id) {
    loadTasksForProject()
  }
}

// 提交日报表单
const handleDailySubmitForm = async () => {
  try {
    await dailyFormRef.value.validate()
    const data: CreateDailyReportRequest = {
      date: dailyFormData.date!.format('YYYY-MM-DD'),
      content: dailyFormData.content,
      hours: dailyFormData.hours,
      project_id: dailyFormData.project_id,
      task_id: dailyFormData.task_id
    }
    if (dailyFormData.id) {
      await updateDailyReport(dailyFormData.id, data)
      message.success('更新成功')
    } else {
      await createDailyReport(data)
      message.success('创建成功')
    }
    dailyModalVisible.value = false
    loadDailyReports()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 提交日报（状态改为已提交）
const handleDailySubmit = async (record: DailyReport) => {
  try {
    await updateDailyReportStatus(record.id, { status: 'submitted' })
    message.success('提交成功')
    loadDailyReports()
  } catch (error: any) {
    message.error(error.message || '提交失败')
  }
}

// 删除日报
const handleDailyDelete = async (id: number) => {
  try {
    await deleteDailyReport(id)
    message.success('删除成功')
    loadDailyReports()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 取消日报表单
const handleDailyCancel = () => {
  dailyFormRef.value?.resetFields()
  availableTasks.value = []
}

// 周报搜索
const handleWeeklySearch = () => {
  weeklyPagination.current = 1
  loadWeeklyReports()
}

// 周报重置
const handleWeeklyReset = () => {
  weeklySearchForm.status = undefined
  weeklySearchForm.start_date = undefined
  weeklySearchForm.end_date = undefined
  weeklySearchForm.project_id = undefined
  weeklyPagination.current = 1
  loadWeeklyReports()
}

// 周报表格变化
const handleWeeklyTableChange = (pag: any) => {
  weeklyPagination.current = pag.current
  weeklyPagination.pageSize = pag.pageSize
  loadWeeklyReports()
}

// 编辑周报
const handleWeeklyEdit = (record: WeeklyReport) => {
  weeklyModalTitle.value = '编辑周报'
  weeklyFormData.id = record.id
  weeklyFormData.week_start = dayjs(record.week_start)
  weeklyFormData.week_end = dayjs(record.week_end)
  weeklyFormData.summary = record.summary || ''
  weeklyFormData.next_week_plan = record.next_week_plan || ''
  weeklyFormData.project_id = record.project_id
  weeklyFormData.task_id = record.task_id
  weeklyModalVisible.value = true
  if (weeklyFormData.project_id) {
    loadTasksForProject()
  }
}

// 提交周报表单
const handleWeeklySubmitForm = async () => {
  try {
    await weeklyFormRef.value.validate()
    const data: CreateWeeklyReportRequest = {
      week_start: weeklyFormData.week_start!.format('YYYY-MM-DD'),
      week_end: weeklyFormData.week_end!.format('YYYY-MM-DD'),
      summary: weeklyFormData.summary,
      next_week_plan: weeklyFormData.next_week_plan,
      project_id: weeklyFormData.project_id,
      task_id: weeklyFormData.task_id
    }
    if (weeklyFormData.id) {
      await updateWeeklyReport(weeklyFormData.id, data)
      message.success('更新成功')
    } else {
      await createWeeklyReport(data)
      message.success('创建成功')
    }
    weeklyModalVisible.value = false
    loadWeeklyReports()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 提交周报（状态改为已提交）
const handleWeeklySubmit = async (record: WeeklyReport) => {
  try {
    await updateWeeklyReportStatus(record.id, { status: 'submitted' })
    message.success('提交成功')
    loadWeeklyReports()
  } catch (error: any) {
    message.error(error.message || '提交失败')
  }
}

// 删除周报
const handleWeeklyDelete = async (id: number) => {
  try {
    await deleteWeeklyReport(id)
    message.success('删除成功')
    loadWeeklyReports()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 取消周报表单
const handleWeeklyCancel = () => {
  weeklyFormRef.value?.resetFields()
  availableTasks.value = []
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'default',
    submitted: 'processing',
    approved: 'success'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    submitted: '已提交',
    approved: '已审批'
  }
  return texts[status] || status
}

// 项目筛选
const filterProjectOption = (input: string, option: any) => {
  const project = projects.value.find(p => p.id === option.value)
  if (!project) return false
  const searchText = input.toLowerCase()
  return (
    project.name.toLowerCase().includes(searchText) ||
    (project.code && project.code.toLowerCase().includes(searchText))
  )
}

// 任务筛选
const filterTaskOption = (input: string, option: any) => {
  const task = availableTasks.value.find(t => t.id === option.value)
  if (!task) return false
  const searchText = input.toLowerCase()
  return task.title.toLowerCase().includes(searchText)
}

// 监听项目变化，重新加载任务
watch(() => dailyFormData.project_id, () => {
  dailyFormData.task_id = undefined
  if (dailyFormData.project_id) {
    loadTasksForProject()
  } else {
    availableTasks.value = []
  }
})

watch(() => weeklyFormData.project_id, () => {
  weeklyFormData.task_id = undefined
  if (weeklyFormData.project_id) {
    loadTasksForProject()
  } else {
    availableTasks.value = []
  }
})

onMounted(() => {
  // 读取路由查询参数
  if (route.query.status) {
    // 将工作台传递的 pending 映射为 draft（草稿）
    let status = route.query.status as string
    if (status === 'pending') {
      status = 'draft'
    }
    if (activeTab.value === 'daily') {
      dailySearchForm.status = status
    } else {
      weeklySearchForm.status = status
    }
  }
  
  loadDailyReports()
  loadProjects()
})
</script>

<style scoped>
.report-management {
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

