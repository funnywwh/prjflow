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
                      <a-menu-item key="wait">未开始</a-menu-item>
                      <a-menu-item key="doing">进行中</a-menu-item>
                      <a-menu-item key="done">已完成</a-menu-item>
                      <a-menu-item key="pause">已暂停</a-menu-item>
                      <a-menu-item key="cancel">已取消</a-menu-item>
                      <a-menu-item key="closed">已关闭</a-menu-item>
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

    <!-- 任务编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑任务"
      :width="800"
      :mask-closable="false"
      @ok="handleEditSubmit"
      @cancel="handleEditCancel"
    >
      <a-form
        ref="editFormRef"
        :model="editFormData"
        :rules="editFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="任务标题" name="title">
          <a-input v-model:value="editFormData.title" placeholder="请输入任务标题" />
        </a-form-item>
        <a-form-item label="任务描述" name="description">
          <MarkdownEditor
            ref="editDescriptionEditorRef"
            v-model="editFormData.description"
            placeholder="请输入任务描述（支持Markdown）"
            :rows="8"
            :project-id="editFormData.project_id || 0"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="editFormData.project_id"
            placeholder="选择项目"
            show-search
            :filter-option="filterProjectOption"
            @change="handleEditFormProjectChange"
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
        <a-form-item label="关联需求" name="requirement_id">
          <a-select
            v-model:value="editFormData.requirement_id"
            placeholder="选择关联需求（可选）"
            allow-clear
            show-search
            :filter-option="filterRequirementOption"
            @focus="loadRequirements(editFormData.project_id || 0)"
          >
            <a-select-option
              v-for="requirement in requirements"
              :key="requirement.id"
              :value="requirement.id"
            >
              {{ requirement.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="editFormData.status">
            <a-select-option value="wait">未开始</a-select-option>
            <a-select-option value="doing">进行中</a-select-option>
            <a-select-option value="done">已完成</a-select-option>
            <a-select-option value="pause">已暂停</a-select-option>
            <a-select-option value="cancel">已取消</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="优先级" name="priority">
          <a-select v-model:value="editFormData.priority">
            <a-select-option value="low">低</a-select-option>
            <a-select-option value="medium">中</a-select-option>
            <a-select-option value="high">高</a-select-option>
            <a-select-option value="urgent">紧急</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="负责人" name="assignee_id">
          <a-select
            v-model:value="editFormData.assignee_id"
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
            v-model:value="editFormData.start_date"
            placeholder="选择开始日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="结束日期" name="end_date">
          <a-date-picker
            v-model:value="editFormData.end_date"
            placeholder="选择结束日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="截止日期" name="due_date">
          <a-date-picker
            v-model:value="editFormData.due_date"
            placeholder="选择截止日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="进度" name="progress">
          <a-slider
            v-model:value="editFormData.progress"
            :min="0"
            :max="100"
            :marks="{ 0: '0%', 50: '50%', 100: '100%' }"
          />
          <span style="margin-left: 8px">{{ editFormData.progress || 0 }}%</span>
        </a-form-item>
        <a-form-item label="预估工时" name="estimated_hours">
          <a-input-number
            v-model:value="editFormData.estimated_hours"
            placeholder="预估工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 添加备注模态框 -->
    <a-modal
      v-model:open="noteModalVisible"
      title="添加备注"
      :mask-closable="true"
      @ok="handleNoteSubmit"
      @cancel="handleNoteCancel"
    >
      <a-form
        ref="noteFormRef"
        :model="noteFormData"
        :rules="noteFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="备注" name="comment">
          <a-textarea
            v-model:value="noteFormData.comment"
            placeholder="请输入备注"
            :rows="4"
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
  updateTask,
  updateTaskStatus,
  deleteTask,
  updateTaskProgress,
  getTaskHistory,
  addTaskHistoryNote,
  type Task,
  type CreateTaskRequest,
  type UpdateTaskProgressRequest,
  type Action
} from '@/api/task'
import { getUsers, type User } from '@/api/user'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, type Requirement } from '@/api/requirement'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<Task | null>(null)
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const requirements = ref<Requirement[]>([])
const progressModalVisible = ref(false)

// 历史记录相关
const historyLoading = ref(false)
const historyList = ref<Action[]>([])
const expandedHistoryIds = ref<Set<number>>(new Set())
const noteModalVisible = ref(false)
const noteFormRef = ref()
const noteFormData = reactive({
  comment: ''
})
const noteFormRules = {
  comment: [{ required: true, message: '请输入备注', trigger: 'blur' }]
}

// 编辑模态框相关
const editModalVisible = ref(false)
const editFormRef = ref()
const editDescriptionEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
const editFormData = reactive<Omit<CreateTaskRequest, 'start_date' | 'end_date' | 'due_date'> & { 
  start_date?: Dayjs | undefined
  end_date?: Dayjs | undefined
  due_date?: Dayjs | undefined
}>({
  title: '',
  description: '',
  status: 'wait',
  priority: 'medium',
  project_id: 0,
  requirement_id: undefined,
  assignee_id: undefined,
  start_date: undefined,
  end_date: undefined,
  due_date: undefined,
  progress: 0,
  estimated_hours: undefined,
  dependency_ids: []
})
const editFormRules = {
  title: [{ required: true, message: '请输入任务标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}
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
    await loadTaskHistory(id) // 加载历史记录
  } catch (error: any) {
    message.error(error.message || '加载任务详情失败')
    router.push('/task')
  } finally {
    loading.value = false
  }
}

// 加载历史记录
const loadTaskHistory = async (taskId?: number) => {
  const id = taskId || Number(route.params.id)
  if (!id) return

  historyLoading.value = true
  try {
    const response = await getTaskHistory(id)
    historyList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    historyLoading.value = false
  }
}

// 编辑
const handleEdit = async () => {
  if (!task.value) return
  
  editFormData.title = task.value.title
  editFormData.description = task.value.description || ''
  editFormData.status = task.value.status
  editFormData.priority = task.value.priority
  editFormData.project_id = task.value.project_id
  editFormData.requirement_id = task.value.requirement_id
  editFormData.assignee_id = task.value.assignee_id
  editFormData.start_date = task.value.start_date ? dayjs(task.value.start_date) : undefined
  editFormData.end_date = task.value.end_date ? dayjs(task.value.end_date) : undefined
  editFormData.due_date = task.value.due_date ? dayjs(task.value.due_date) : undefined
  editFormData.progress = task.value.progress
  editFormData.estimated_hours = task.value.estimated_hours
  editFormData.dependency_ids = task.value.dependencies?.map(d => d.id) || []
  
  editModalVisible.value = true
  if (users.value.length === 0) {
    await loadUsers()
  }
  if (projects.value.length === 0) {
    await loadProjects()
  }
  if (task.value.project_id && requirements.value.length === 0) {
    await loadRequirements(task.value.project_id)
  }
}

// 编辑提交
const handleEditSubmit = async () => {
  if (!task.value) return
  
  try {
    await editFormRef.value.validate()
    
    // 获取最新的描述内容
    let description = editFormData.description || ''
    
    // 如果有项目ID，尝试上传本地图片（如果有的话）
    if (editDescriptionEditorRef.value && editFormData.project_id) {
      try {
        const uploadedDescription = await editDescriptionEditorRef.value.uploadLocalImages(async (file: File, projectId: number) => {
          const { uploadFile } = await import('@/api/attachment')
          const attachment = await uploadFile(file, projectId)
          return attachment
        })
        description = uploadedDescription
      } catch (error: any) {
        console.error('上传图片失败:', error)
        message.warning('部分图片上传失败，请检查')
        description = editFormData.description || ''
      }
    }
    
    const data: Partial<CreateTaskRequest> = {
      title: editFormData.title,
      description: description || '',
      status: editFormData.status,
      priority: editFormData.priority,
      requirement_id: editFormData.requirement_id,
      assignee_id: editFormData.assignee_id,
      start_date: editFormData.start_date && typeof editFormData.start_date !== 'string' && 'isValid' in editFormData.start_date && (editFormData.start_date as Dayjs).isValid() ? (editFormData.start_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.start_date === 'string' ? editFormData.start_date : undefined),
      end_date: editFormData.end_date && typeof editFormData.end_date !== 'string' && 'isValid' in editFormData.end_date && (editFormData.end_date as Dayjs).isValid() ? (editFormData.end_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.end_date === 'string' ? editFormData.end_date : undefined),
      due_date: editFormData.due_date && typeof editFormData.due_date !== 'string' && 'isValid' in editFormData.due_date && (editFormData.due_date as Dayjs).isValid() ? (editFormData.due_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.due_date === 'string' ? editFormData.due_date : undefined),
      progress: editFormData.progress,
      estimated_hours: editFormData.estimated_hours,
      dependency_ids: editFormData.dependency_ids
    }
    
    await updateTask(task.value.id, data)
    
    message.success('更新成功')
    editModalVisible.value = false
    await loadTask() // 重新加载任务详情（会自动加载历史记录）
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '更新失败')
  }
}

// 编辑取消
const handleEditCancel = () => {
  editFormRef.value?.resetFields()
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

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await getProjects({ size: 1000 })
    projects.value = response.list || []
  } catch (error: any) {
    console.error('加载项目列表失败:', error)
  }
}

// 加载需求列表
const loadRequirements = async (projectId: number) => {
  try {
    const response = await getRequirements({ project_id: projectId })
    requirements.value = response.list || []
  } catch (error: any) {
    console.error('加载需求列表失败:', error)
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

// 获取字段显示名称
const getFieldDisplayName = (fieldName: string): string => {
  const fieldNames: Record<string, string> = {
    title: '任务标题',
    description: '任务描述',
    status: '状态',
    priority: '优先级',
    assignee_id: '负责人',
    start_date: '开始日期',
    end_date: '结束日期',
    due_date: '截止日期',
    progress: '进度',
    estimated_hours: '预估工时',
    actual_hours: '实际工时',
    requirement_id: '关联需求'
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

// 添加备注
const handleAddNote = () => {
  if (!task.value) {
    message.warning('任务信息未加载完成，请稍候再试')
    return
  }
  noteFormData.comment = ''
  noteModalVisible.value = true
}

// 提交备注
const handleNoteSubmit = async () => {
  if (!task.value) return
  try {
    await noteFormRef.value.validate()
    await addTaskHistoryNote(task.value.id, { comment: noteFormData.comment })
    message.success('添加备注成功')
    noteModalVisible.value = false
    await loadTaskHistory(task.value.id)
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '添加备注失败')
  }
}

// 取消添加备注
const handleNoteCancel = () => {
  noteFormRef.value?.resetFields()
}

// 项目筛选
const filterProjectOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 需求筛选
const filterRequirementOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 用户筛选
const filterUserOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 编辑表单项目选择改变
const handleEditFormProjectChange = () => {
  editFormData.requirement_id = undefined
  if (editFormData.project_id) {
    loadRequirements(editFormData.project_id)
  } else {
    requirements.value = []
  }
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
    wait: 'orange',
    doing: 'blue',
    done: 'green',
    pause: 'purple',
    cancel: 'red',
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
const isOverdue = (dueDate?: string, status?: string) => {
  if (!dueDate || status === 'done' || status === 'cancel' || status === 'closed') {
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
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.task-detail :deep(.ant-layout) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  flex: 1;
  padding: 24px;
  background: #f0f2f5;
  overflow-y: auto;
  overflow-x: hidden;
}

.content-inner {
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
  min-height: fit-content;
}

.markdown-content {
  min-height: 200px;
}
</style>

