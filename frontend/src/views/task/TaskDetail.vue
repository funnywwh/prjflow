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
                <a-button @click="handleAssign">指派</a-button>
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
                <a-button @click="handleConvertToRequirement">
                  转需求
                </a-button>
                <a-button @click="handleConvertToBug">
                  转Bug
                </a-button>
                <a-popconfirm
                  title="确定要删除这个任务吗？"
                  @confirm="handleDelete"
                >
                  <a-button danger>删除</a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-page-header>

          <TaskDetailContent
            :task="task"
            :loading="loading"
            :history-list="historyList"
            :history-loading="historyLoading"
            @add-note="handleAddNote"
            @go-to-task="(taskId) => router.push(`/task/${taskId}`)"
          />
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 更新进度模态框 -->
    <a-modal
      v-model:open="progressModalVisible"
      title="更新任务进度"
      :mask-closable="true"
      :z-index="2000"
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
            :getPopupContainer="getPopupContainer"
            :popupStyle="{ zIndex: 2100 }"
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
          <ProjectMemberSelect
            v-model="editFormData.assignee_id"
            :project-id="task?.project_id"
            :multiple="false"
            placeholder="选择负责人（可选）"
            :show-role="true"
            :show-hint="!task?.project_id"
          />
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
        <a-form-item label="附件">
          <AttachmentUpload
            v-if="editFormData.project_id && editFormData.project_id > 0"
            :project-id="editFormData.project_id"
            :model-value="editFormData.attachment_ids"
            :existing-attachments="taskAttachments"
            @update:modelValue="(value) => { editFormData.attachment_ids = value }"
            @attachment-deleted="handleAttachmentDeleted"
          />
          <span v-else style="color: #999;">请先选择项目后再上传附件</span>
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

    <!-- 任务指派模态框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="指派任务"
      :mask-closable="true"
      :z-index="2100"
      @ok="handleAssignSubmit"
      @cancel="handleAssignCancel"
    >
      <a-form
        ref="assignFormRef"
        :model="assignFormData"
        :rules="assignFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="指派给" name="assignee_id">
          <ProjectMemberSelect
            v-model="assignFormData.assignee_id"
            :project-id="task?.project_id"
            :multiple="false"
            placeholder="选择指派给"
            :show-role="true"
            :get-popup-container="(triggerNode: HTMLElement) => triggerNode.parentElement || triggerNode.ownerDocument.body"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select
            v-model:value="assignFormData.status"
            placeholder="选择状态（可选，不选择则自动修改）"
            allow-clear
            :get-popup-container="(triggerNode: HTMLElement) => triggerNode.parentElement || triggerNode.ownerDocument.body"
          >
            <a-select-option value="wait">未开始</a-select-option>
            <a-select-option value="doing">进行中</a-select-option>
            <a-select-option value="done">已完成</a-select-option>
            <a-select-option value="pause">已暂停</a-select-option>
            <a-select-option value="cancel">已取消</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="备注" name="comment">
          <a-textarea
            v-model:value="assignFormData.comment"
            placeholder="请输入备注（可选）"
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
import { message, Modal } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import ProjectMemberSelect from '@/components/ProjectMemberSelect.vue'
import TaskDetailContent from '@/components/TaskDetailContent.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import { getAttachments, type Attachment } from '@/api/attachment'
import {
  getTask,
  updateTask,
  updateTaskStatus,
  deleteTask,
  updateTaskProgress,
  getTaskHistory,
  addTaskHistoryNote,
  assignTask,
  type Task,
  type CreateTaskRequest,
  type UpdateTaskProgressRequest,
  type Action
} from '@/api/task'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, createRequirement, type Requirement, type CreateRequirementRequest } from '@/api/requirement'
import { createBug, type CreateBugRequest } from '@/api/bug'
import { getVersions } from '@/api/version'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<Task | null>(null)
const projects = ref<Project[]>([])
const requirements = ref<Requirement[]>([])
const progressModalVisible = ref(false)

// 历史记录相关
const historyLoading = ref(false)
const historyList = ref<Action[]>([])
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
  attachment_ids?: number[]
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
  attachment_ids: []
})
const taskAttachments = ref<Attachment[]>([])
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

// 指派模态框相关
const assignModalVisible = ref(false)
const assignFormRef = ref()
const assignFormData = reactive({
  assignee_id: undefined as number | undefined,
  status: undefined as string | undefined,
  comment: undefined as string | undefined
})
const assignFormRules = {
  assignee_id: [{ required: true, message: '请选择指派人', trigger: 'change' }]
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
  
  // 加载任务附件
  try {
    if (task.value.attachments && task.value.attachments.length > 0) {
      taskAttachments.value = task.value.attachments
      editFormData.attachment_ids = task.value.attachments.map((a: any) => a.id)
    } else {
      taskAttachments.value = await getAttachments({ task_id: task.value.id })
      editFormData.attachment_ids = taskAttachments.value.map(a => a.id)
    }
  } catch (error: any) {
    console.error('加载附件失败:', error)
    taskAttachments.value = []
    editFormData.attachment_ids = []
  }
  editFormData.due_date = task.value.due_date ? dayjs(task.value.due_date) : undefined
  editFormData.progress = task.value.progress
  editFormData.estimated_hours = task.value.estimated_hours
  editFormData.dependency_ids = task.value.dependencies?.map(d => d.id) || []
  
  editModalVisible.value = true
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
    
    const data: any = {
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
    
    // 始终发送 attachment_ids，如果为 undefined 或 null，发送空数组
    const attachmentIdsValue = editFormData.attachment_ids
    if (attachmentIdsValue === undefined || attachmentIdsValue === null) {
      data.attachment_ids = []
    } else {
      data.attachment_ids = Array.isArray(attachmentIdsValue) ? attachmentIdsValue : []
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

// 处理附件删除事件
const handleAttachmentDeleted = (attachmentId: number) => {
  // 从taskAttachments中移除已删除的附件
  taskAttachments.value = taskAttachments.value.filter(a => a.id !== attachmentId)
  // 同时从 editFormData.attachment_ids 中移除
  if (editFormData.attachment_ids) {
    editFormData.attachment_ids = editFormData.attachment_ids.filter(id => id !== attachmentId)
  }
}

// 编辑取消
const handleEditCancel = () => {
  editFormRef.value?.resetFields()
}

// 指派
const handleAssign = () => {
  if (!task.value) return
  // 设置默认值：如果当前状态是 "wait"，默认选择 "doing"；否则默认不选择（自动修改）
  if (task.value.status === 'wait') {
    assignFormData.status = 'doing'
  } else {
    assignFormData.status = undefined
  }
  assignFormData.assignee_id = task.value.assignee_id // 预填充当前指派人
  assignFormData.comment = undefined // 清空备注
  assignModalVisible.value = true
}

// 指派提交
const handleAssignSubmit = async () => {
  if (!task.value) return
  try {
    await assignFormRef.value.validate()
    const requestData: any = { assignee_id: assignFormData.assignee_id }
    if (assignFormData.status) {
      requestData.status = assignFormData.status
    }
    if (assignFormData.comment) {
      requestData.comment = assignFormData.comment
    }
    await assignTask(task.value.id, requestData)
    message.success('指派成功')
    assignModalVisible.value = false
    await loadTask() // 重新加载任务详情（会自动加载历史记录）
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '指派失败')
  }
}

// 指派取消
const handleAssignCancel = () => {
  assignFormRef.value?.resetFields()
  assignFormData.status = undefined
  assignFormData.comment = undefined
  assignFormData.assignee_id = undefined
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

// 任务转需求
const handleConvertToRequirement = async () => {
  if (!task.value) return
  
  // 确认对话框
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此任务转为需求吗？转换后将创建新需求，并关联到此任务。',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        resolve(true)
        modal.destroy()
      },
      onCancel: () => {
        resolve(false)
        modal.destroy()
      }
    })
  })
  
  if (!confirmed) return
  
  try {
    // 创建新需求，基于任务的信息
    const requirementData: CreateRequirementRequest = {
      title: `[转需求] ${task.value.title}`,
      description: task.value.description 
        ? `${task.value.description}\n\n---\n\n*由任务 #${task.value.id}转换而来*`
        : `*由任务 #${task.value.id}转换而来*`,
      project_id: task.value.project_id,
      priority: task.value.priority,
      status: 'draft', // 需求默认草稿状态
      assignee_id: task.value.assignee_id,
      estimated_hours: task.value.estimated_hours
    }
    
    // 创建需求
    const requirement = await createRequirement(requirementData)
    
    message.success(`转换成功，已创建需求 #${requirement.id}`)
    
    // 刷新任务详情
    await loadTask()
    
    // 可选：跳转到新创建的需求详情页
    // router.push(`/requirement/${requirement.id}`)
  } catch (error: any) {
    message.error(error.message || '转换失败')
  }
}

// 任务转Bug
const handleConvertToBug = async () => {
  if (!task.value) return
  
  // 确认对话框
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此任务转为Bug吗？转换后将创建新Bug，并关联到此任务。',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        resolve(true)
        modal.destroy()
      },
      onCancel: () => {
        resolve(false)
        modal.destroy()
      }
    })
  })
  
  if (!confirmed) return
  
  try {
    // 获取项目下的版本列表
    const versionsResponse = await getVersions({ project_id: task.value.project_id, size: 1000 })
    const versions = versionsResponse.list || []
    
    if (versions.length === 0) {
      message.error('该项目下没有版本，请先创建版本后再转换')
      return
    }
    
    const firstVersion = versions[0]
    if (!firstVersion) {
      message.error('无法获取版本信息')
      return
    }
    
    // 创建新Bug，基于任务的信息
    const bugData: CreateBugRequest = {
      title: `[转Bug] ${task.value.title}`,
      description: task.value.description 
        ? `${task.value.description}\n\n---\n\n*由任务 #${task.value.id}转换而来*`
        : `*由任务 #${task.value.id}转换而来*`,
      project_id: task.value.project_id,
      priority: task.value.priority,
      severity: 'medium', // Bug默认严重程度
      status: 'active', // Bug默认激活状态
      // 如果任务有负责人，作为Bug的指派人员
      assignee_ids: task.value.assignee_id 
        ? [task.value.assignee_id] 
        : undefined,
      version_ids: [firstVersion.id], // 使用第一个版本
      estimated_hours: task.value.estimated_hours
    }
    
    // 创建Bug
    const bug = await createBug(bugData)
    
    message.success(`转换成功，已创建Bug #${bug.id}`)
    
    // 刷新任务详情
    await loadTask()
    
    // 可选：跳转到新创建的Bug详情页
    // router.push(`/bug/${bug.id}`)
  } catch (error: any) {
    message.error(error.message || '转换失败')
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


onMounted(() => {
  loadTask()
})

// 获取下拉框容器（用于解决模态框中下拉框被遮挡的问题）
const getPopupContainer = (triggerNode: HTMLElement): HTMLElement => {
  return triggerNode.parentElement || document.body
}
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

