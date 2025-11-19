<template>
  <div class="task-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="任务管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增任务
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="任务标题/描述"
                  allow-clear
                  style="width: 200px"
                />
              </a-form-item>
              <a-form-item label="项目">
                <a-select
                  v-model:value="searchForm.project_id"
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
              <a-form-item label="状态">
                <a-select
                  v-model:value="searchForm.status"
                  placeholder="选择状态"
                  allow-clear
                  style="width: 120px"
                >
                  <a-select-option value="todo">待办</a-select-option>
                  <a-select-option value="in_progress">进行中</a-select-option>
                  <a-select-option value="done">已完成</a-select-option>
                  <a-select-option value="cancelled">已取消</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="优先级">
                <a-select
                  v-model:value="searchForm.priority"
                  placeholder="选择优先级"
                  allow-clear
                  style="width: 120px"
                >
                  <a-select-option value="low">低</a-select-option>
                  <a-select-option value="medium">中</a-select-option>
                  <a-select-option value="high">高</a-select-option>
                  <a-select-option value="urgent">紧急</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-card :bordered="false">
            <a-table
              :columns="columns"
              :data-source="tasks"
              :loading="loading"
              :pagination="pagination"
              row-key="id"
              @change="handleTableChange"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <a-tag :color="getStatusColor(record.status)">
                    {{ getStatusText(record.status) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'priority'">
                  <a-tag :color="getPriorityColor(record.priority)">
                    {{ getPriorityText(record.priority) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'requirement'">
                  {{ record.requirement?.title || '-' }}
                </template>
                <template v-else-if="column.key === 'assignee'">
                  {{ record.assignee ? `${record.assignee.username}${record.assignee.nickname ? `(${record.assignee.nickname})` : ''}` : '-' }}
                </template>
                <template v-else-if="column.key === 'progress'">
                  <a-progress :percent="record.progress" :status="record.status === 'done' ? 'success' : 'active'" />
                </template>
                <template v-else-if="column.key === 'hours'">
                  <div>
                    <div v-if="record.estimated_hours">预估: {{ record.estimated_hours.toFixed(2) }}h</div>
                    <div v-if="record.actual_hours">实际: {{ record.actual_hours.toFixed(2) }}h</div>
                    <span v-if="!record.estimated_hours && !record.actual_hours">-</span>
                  </div>
                </template>
                <template v-else-if="column.key === 'dates'">
                  <div>
                    <div v-if="record.start_date">开始: {{ formatDate(record.start_date) }}</div>
                    <div v-if="record.end_date">结束: {{ formatDate(record.end_date) }}</div>
                    <div v-if="record.due_date" :style="{ color: isOverdue(record.due_date, record.status) ? 'red' : '' }">
                      截止: {{ formatDate(record.due_date) }}
                    </div>
                  </div>
                </template>
                <template v-else-if="column.key === 'created_at'">
                  {{ formatDateTime(record.created_at) }}
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleView(record)">
                      详情
                    </a-button>
                    <a-button type="link" size="small" @click="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-button type="link" size="small" @click="handleUpdateProgress(record)">
                      进度
                    </a-button>
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleStatusChange(record.id, e.key as string)">
                          <a-menu-item key="todo">待办</a-menu-item>
                          <a-menu-item key="in_progress">进行中</a-menu-item>
                          <a-menu-item key="done">已完成</a-menu-item>
                          <a-menu-item key="cancelled">已取消</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个任务吗？"
                      @confirm="handleDelete(record.id)"
                    >
                      <a-button type="link" size="small" danger>删除</a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </template>
            </a-table>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 任务编辑/创建模态框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="900"
      @ok="handleSubmit"
      @cancel="handleCancel"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="任务标题" name="title">
          <a-input v-model:value="formData.title" placeholder="请输入任务标题" />
        </a-form-item>
        <a-form-item label="任务描述" name="description">
          <MarkdownEditor
            v-model="formData.description"
            placeholder="请输入任务描述（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="选择项目"
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
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status">
            <a-select-option value="todo">待办</a-select-option>
            <a-select-option value="in_progress">进行中</a-select-option>
            <a-select-option value="done">已完成</a-select-option>
            <a-select-option value="cancelled">已取消</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="优先级" name="priority">
          <a-select v-model:value="formData.priority">
            <a-select-option value="low">低</a-select-option>
            <a-select-option value="medium">中</a-select-option>
            <a-select-option value="high">高</a-select-option>
            <a-select-option value="urgent">紧急</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="负责人" name="assignee_id">
          <a-select
            v-model:value="formData.assignee_id"
            placeholder="选择负责人（可选）"
            allow-clear
            show-search
            :filter-option="filterUserOption"
          >
            <a-select-option
              v-for="user in users"
              :key="user.id"
              :value="user.id"
            >
              {{ user.username }}{{ user.nickname ? `(${user.nickname})` : '' }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="开始日期" name="start_date">
          <a-date-picker
            v-model:value="formData.start_date"
            placeholder="选择开始日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="结束日期" name="end_date">
          <a-date-picker
            v-model:value="formData.end_date"
            placeholder="选择结束日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="截止日期" name="due_date">
          <a-date-picker
            v-model:value="formData.due_date"
            placeholder="选择截止日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="进度" name="progress">
          <a-slider
            v-model:value="formData.progress"
            :min="0"
            :max="100"
            :marks="{ 0: '0%', 50: '50%', 100: '100%' }"
          />
          <span style="margin-left: 8px">{{ formData.progress }}%</span>
        </a-form-item>
        <a-form-item label="预估工时" name="estimated_hours">
          <a-input-number
            v-model:value="formData.estimated_hours"
            placeholder="预估工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="实际工时" name="actual_hours">
          <a-input-number
            v-model:value="formData.actual_hours"
            placeholder="实际工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配</span>
        </a-form-item>
        <a-form-item label="工作日期" name="work_date" v-if="formData.actual_hours">
          <a-date-picker
            v-model:value="formData.work_date"
            placeholder="选择工作日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
          <span style="margin-left: 8px; color: #999">不填则使用任务开始日期或今天</span>
        </a-form-item>
        <a-form-item label="依赖任务" name="dependency_ids">
          <a-select
            v-model:value="formData.dependency_ids"
            mode="multiple"
            placeholder="选择依赖任务（可选）"
            allow-clear
            show-search
            :filter-option="filterTaskOption"
            :loading="taskLoading"
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
import { ref, reactive, onMounted, watch, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime, formatDate } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getTasks,
  getTask,
  createTask,
  updateTask,
  deleteTask,
  updateTaskStatus,
  updateTaskProgress,
  type Task,
  type CreateTaskRequest,
  type UpdateTaskProgressRequest
} from '@/api/task'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, type Requirement } from '@/api/requirement'
import { getUsers, type User } from '@/api/user'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const tasks = ref<Task[]>([])
const projects = ref<Project[]>([])
const requirements = ref<Requirement[]>([])
const users = ref<User[]>([])
const availableTasks = ref<Task[]>([])
const taskLoading = ref(false)

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  priority: undefined as string | undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const columns = [
  { title: '任务标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '项目', key: 'project', width: 120 },
  { title: '需求', key: 'requirement', width: 150 },
  { title: '状态', key: 'status', width: 100 },
  { title: '优先级', key: 'priority', width: 100 },
  { title: '负责人', key: 'assignee', width: 150 },
  { title: '进度', key: 'progress', width: 150 },
  { title: '工时', key: 'hours', width: 150 },
  { title: '日期', key: 'dates', width: 200 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 300, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增任务')
const formRef = ref()
const formData = reactive<CreateTaskRequest & { id?: number; start_date?: Dayjs; end_date?: Dayjs; due_date?: Dayjs; actual_hours?: number; work_date?: Dayjs }>({
  title: '',
  description: '',
  status: 'todo',
  priority: 'medium',
  project_id: 0,
  requirement_id: undefined,
  assignee_id: undefined,
  start_date: undefined,
  end_date: undefined,
  due_date: undefined,
  progress: 0,
  estimated_hours: undefined,
  actual_hours: undefined,
  work_date: undefined,
  dependency_ids: []
})

const formRules = {
  title: [{ required: true, message: '请输入任务标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

const progressModalVisible = ref(false)
const progressFormRef = ref()
const progressFormData = reactive<{
  task_id: number
  progress?: number
  estimated_hours?: number
  actual_hours?: number
  work_date?: Dayjs
}>({
  task_id: 0,
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

// 加载任务列表
const loadTasks = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    if (searchForm.keyword) {
      params.keyword = searchForm.keyword
    }
    if (searchForm.project_id) {
      params.project_id = searchForm.project_id
    }
    if (searchForm.status) {
      params.status = searchForm.status
    }
    if (searchForm.priority) {
      params.priority = searchForm.priority
    }
    const response = await getTasks(params)
    tasks.value = response.list
    pagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载任务列表失败')
  } finally {
    loading.value = false
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

// 加载用户列表
const loadUsers = async () => {
  try {
    const response = await getUsers()
    users.value = response.list || []
  } catch (error: any) {
    console.error('加载用户列表失败:', error)
  }
}

// 加载任务列表（用于依赖选择）
const loadTasksForProject = async () => {
  if (!formData.project_id) {
    availableTasks.value = []
    return
  }
  taskLoading.value = true
  try {
    const response = await getTasks({ project_id: formData.project_id })
    // 排除当前任务（如果是编辑模式）
    availableTasks.value = response.list.filter(t => t.id !== formData.id)
  } catch (error: any) {
    console.error('加载任务列表失败:', error)
  } finally {
    taskLoading.value = false
  }
}

// 监听项目变化，重新加载任务
watch(() => formData.project_id, () => {
  formData.dependency_ids = []
  if (formData.project_id) {
    loadTasksForProject()
  } else {
    availableTasks.value = []
  }
})

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadTasks()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.priority = undefined
  pagination.current = 1
  loadTasks()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadTasks()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增任务'
  formData.id = undefined
  formData.title = ''
  formData.description = ''
  formData.status = 'todo'
  formData.priority = 'medium'
  // 如果有路由查询参数中的 project_id，使用它；否则重置为 0
  const projectIdFromQuery = route.query.project_id
  formData.project_id = projectIdFromQuery ? Number(projectIdFromQuery) : 0
  formData.requirement_id = undefined
  formData.assignee_id = undefined
  formData.start_date = undefined
  formData.end_date = undefined
  formData.due_date = undefined
  formData.progress = 0
  formData.estimated_hours = undefined
  formData.dependency_ids = []
  modalVisible.value = true
  // 如果预填充了项目ID，加载该项目的任务列表（用于依赖任务选择）
  if (formData.project_id) {
    loadTasksForProject()
  }
}

// 编辑
const handleEdit = (record: Task) => {
  modalTitle.value = '编辑任务'
  formData.id = record.id
  formData.title = record.title
  formData.description = record.description || ''
  formData.status = record.status
  formData.priority = record.priority
  formData.project_id = record.project_id
  formData.requirement_id = record.requirement_id
  formData.assignee_id = record.assignee_id
  // 解析日期，确保日期有效
  if (record.start_date) {
    const startDate = dayjs(record.start_date)
    formData.start_date = startDate.isValid() ? startDate : undefined
  } else {
    formData.start_date = undefined
  }
  if (record.end_date) {
    const endDate = dayjs(record.end_date)
    formData.end_date = endDate.isValid() ? endDate : undefined
  } else {
    formData.end_date = undefined
  }
  if (record.due_date) {
    const dueDate = dayjs(record.due_date)
    formData.due_date = dueDate.isValid() ? dueDate : undefined
  } else {
    formData.due_date = undefined
  }
  formData.progress = record.progress
  formData.estimated_hours = record.estimated_hours
  formData.actual_hours = record.actual_hours
  formData.work_date = undefined
  formData.dependency_ids = record.dependencies?.map(d => d.id) || []
  modalVisible.value = true
  if (formData.project_id) {
    loadTasksForProject()
  }
}

// 查看详情
const handleView = (record: Task) => {
  router.push(`/task/${record.id}`)
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    // 验证项目ID
    if (!formData.project_id || formData.project_id === 0) {
      message.error('请选择项目')
      return
    }
    const data: CreateTaskRequest = {
      title: formData.title,
      description: formData.description,
      status: formData.status,
      priority: formData.priority,
      project_id: formData.project_id,
      requirement_id: formData.requirement_id,
      assignee_id: formData.assignee_id,
      start_date: formData.start_date && formData.start_date.isValid() ? formData.start_date.format('YYYY-MM-DD') : undefined,
      end_date: formData.end_date && formData.end_date.isValid() ? formData.end_date.format('YYYY-MM-DD') : undefined,
      due_date: formData.due_date && formData.due_date.isValid() ? formData.due_date.format('YYYY-MM-DD') : undefined,
      progress: formData.progress,
      estimated_hours: formData.estimated_hours,
      actual_hours: formData.actual_hours,
      work_date: formData.work_date && formData.work_date.isValid() ? formData.work_date.format('YYYY-MM-DD') : undefined,
      dependency_ids: formData.dependency_ids
    }
    if (formData.id) {
      await updateTask(formData.id, data)
      message.success('更新成功')
    } else {
      await createTask(data)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadTasks()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  }
}

// 取消
const handleCancel = () => {
  formRef.value?.resetFields()
  availableTasks.value = []
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteTask(id)
    message.success('删除成功')
    loadTasks()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 状态变更
const handleStatusChange = async (id: number, status: string) => {
  try {
    await updateTaskStatus(id, { status: status as any })
    message.success('状态更新成功')
    loadTasks()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 更新进度
const handleUpdateProgress = (record: Task) => {
  progressFormData.task_id = record.id
  progressFormData.progress = record.progress
  progressFormData.estimated_hours = record.estimated_hours
  progressFormData.actual_hours = record.actual_hours // 显示当前实际工时
  progressFormData.work_date = dayjs() // 默认今天
  progressModalVisible.value = true
}

// 进度提交
const handleProgressSubmit = async () => {
  try {
    await progressFormRef.value.validate()
    const data: UpdateTaskProgressRequest = {
      progress: progressFormData.progress,
      estimated_hours: progressFormData.estimated_hours,
      actual_hours: progressFormData.actual_hours,
      work_date: progressFormData.work_date && progressFormData.work_date.isValid() ? progressFormData.work_date.format('YYYY-MM-DD') : undefined
    }
    await updateTaskProgress(progressFormData.task_id, data)
    message.success('进度更新成功')
    progressModalVisible.value = false
    loadTasks()
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
const isOverdue = (dueDate: string, status: string) => {
  if (status === 'done' || status === 'cancelled') {
    return false
  }
  const due = dayjs(dueDate)
  const now = dayjs()
  return due.isBefore(now, 'day')
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

// 用户筛选
const filterUserOption = (input: string, option: any) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  const searchText = input.toLowerCase()
  return (
    user.username.toLowerCase().includes(searchText) ||
    (user.nickname && user.nickname.toLowerCase().includes(searchText))
  )
}

onMounted(() => {
  loadTasks()
  loadProjects()
  loadUsers()
  
  // 检查是否有项目ID参数（从看板页面跳转过来）
  const projectId = route.query.project_id
  if (projectId) {
    formData.project_id = Number(projectId)
    searchForm.project_id = Number(projectId)
    handleSearch()
    // 如果是从看板跳转过来，自动打开创建任务模态框
    if (!route.query.edit) {
      nextTick(() => {
        handleCreate()
      })
    }
  }
  
  // 检查是否有编辑ID参数
  const editId = route.query.edit
  if (editId) {
    // 加载任务详情并打开编辑模态框
    getTask(Number(editId)).then(task => {
      handleEdit(task)
    }).catch(() => {
      message.error('加载任务失败')
    })
  }
})
</script>

<style scoped>
.task-management {
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
</style>

