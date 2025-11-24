<template>
  <div class="requirement-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="需求管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增需求
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="需求标题/描述"
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
                  @change="handleSearchProjectChange"
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
                  <a-select-option value="pending">待处理</a-select-option>
                  <a-select-option value="in_progress">进行中</a-select-option>
                  <a-select-option value="completed">已完成</a-select-option>
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

          <!-- 统计概览 -->
          <a-row :gutter="16" style="margin-bottom: 16px">
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="总需求数"
                  :value="statistics?.total || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="待处理"
                  :value="statistics?.pending || 0"
                  :value-style="{ color: '#faad14' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="进行中"
                  :value="statistics?.in_progress || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="已完成"
                  :value="statistics?.completed || 0"
                  :value-style="{ color: '#52c41a' }"
                />
              </a-card>
            </a-col>
          </a-row>

          <!-- 优先级统计 -->
          <a-card title="优先级统计" :bordered="false" style="margin-bottom: 16px">
            <a-row :gutter="16">
              <a-col :span="6">
                <a-statistic
                  title="低"
                  :value="statistics?.low_priority || 0"
                  :value-style="{ color: '#8c8c8c' }"
                />
              </a-col>
              <a-col :span="6">
                <a-statistic
                  title="中"
                  :value="statistics?.medium_priority || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-col>
              <a-col :span="6">
                <a-statistic
                  title="高"
                  :value="statistics?.high_priority || 0"
                  :value-style="{ color: '#faad14' }"
                />
              </a-col>
              <a-col :span="6">
                <a-statistic
                  title="紧急"
                  :value="statistics?.urgent_priority || 0"
                  :value-style="{ color: '#ff4d4f' }"
                />
              </a-col>
            </a-row>
          </a-card>

          <a-card :bordered="false">
            <a-table
              :columns="columns"
              :data-source="requirements"
              :loading="loading"
              :pagination="pagination"
              :scroll="{ x: 'max-content' }"
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
                <template v-else-if="column.key === 'assignee'">
                  {{ record.assignee ? `${record.assignee.username}${record.assignee.nickname ? `(${record.assignee.nickname})` : ''}` : '-' }}
                </template>
                <template v-else-if="column.key === 'hours'">
                  <div>
                    <div v-if="record.estimated_hours">预估: {{ record.estimated_hours.toFixed(2) }}h</div>
                    <div v-if="record.actual_hours">实际: {{ record.actual_hours.toFixed(2) }}h</div>
                    <span v-if="!record.estimated_hours && !record.actual_hours">-</span>
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
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleStatusChange(record.id, e.key as string)">
                          <a-menu-item key="pending">待处理</a-menu-item>
                          <a-menu-item key="in_progress">进行中</a-menu-item>
                          <a-menu-item key="completed">已完成</a-menu-item>
                          <a-menu-item key="cancelled">已取消</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个需求吗？"
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

    <!-- 需求编辑/创建模态框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="800"
      :mask-closable="false"
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
        <a-form-item label="需求标题" name="title">
          <a-input v-model:value="formData.title" placeholder="请输入需求标题" />
        </a-form-item>
        <a-form-item label="需求描述" name="description">
          <MarkdownEditor
            ref="descriptionEditorRef"
            v-model="formData.description"
            placeholder="请输入需求描述（支持Markdown）"
            :rows="8"
            :project-id="formData.project_id || 0"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="请选择项目"
            @change="handleFormProjectChange"
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
            <a-select-option value="pending">待处理</a-select-option>
            <a-select-option value="in_progress">进行中</a-select-option>
            <a-select-option value="completed">已完成</a-select-option>
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
          <span style="margin-left: 8px; color: #999">不填则使用今天</span>
        </a-form-item>
        <a-form-item label="附件">
          <AttachmentUpload
            v-if="formData.project_id && (formData.id || formData.project_id)"
            :project-id="formData.project_id"
            v-model="formData.attachment_ids"
            :existing-attachments="requirementAttachments"
          />
          <span v-else style="color: #999;">请先选择项目后再上传附件</span>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import { type Dayjs } from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import { useAuthStore } from '@/stores/auth'
import {
  getRequirements,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  updateRequirementStatus,
  getRequirementStatistics,
  type Requirement,
  type CreateRequirementRequest,
  type RequirementStatistics
} from '@/api/requirement'
import { getProjects, type Project } from '@/api/project'
import { getUsers, type User } from '@/api/user'
import { getAttachments, attachToEntity, uploadFile, type Attachment } from '@/api/attachment'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const requirements = ref<Requirement[]>([])
const projects = ref<Project[]>([])
const users = ref<User[]>([])
const statistics = ref<RequirementStatistics | null>(null)

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  priority: undefined as string | undefined,
  assignee_id: undefined as number | undefined
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
  { title: '需求标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '项目', key: 'project', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '优先级', key: 'priority', width: 100 },
  { title: '负责人', key: 'assignee', width: 150 },
  { title: '工时', key: 'hours', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增需求')
const formRef = ref()
const descriptionEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
const formData = reactive<Partial<CreateRequirementRequest> & { id?: number; actual_hours?: number; work_date?: Dayjs; attachment_ids?: number[] }>({
  title: '',
  description: '',
  status: 'pending',
  priority: 'medium',
  project_id: undefined, // 表单中可以为undefined，提交时会验证
  assignee_id: undefined,
  estimated_hours: undefined,
  actual_hours: undefined,
  work_date: undefined,
  attachment_ids: [] as number[]
})

const requirementAttachments = ref<Attachment[]>([]) // 需求附件列表

const formRules = {
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载需求列表
const loadRequirements = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
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
    if (searchForm.assignee_id) {
      params.assignee_id = searchForm.assignee_id
    }
    const response = await getRequirements(params)
    requirements.value = response.list
    pagination.total = response.total
    // 加载统计信息
    await loadStatistics()
  } catch (error: any) {
    message.error(error.message || '加载需求列表失败')
  } finally {
    loading.value = false
  }
}

// 加载统计信息
const loadStatistics = async () => {
  try {
    const params: any = {}
    if (searchForm.keyword) {
      params.keyword = searchForm.keyword
    }
    if (searchForm.project_id) {
      params.project_id = searchForm.project_id
    }
    statistics.value = await getRequirementStatistics(params)
  } catch (error: any) {
    console.error('加载统计信息失败:', error)
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

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadRequirements()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_requirement_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_requirement_project_form', value)
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.priority = undefined
  searchForm.assignee_id = undefined
  pagination.current = 1
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_requirement_project_search', undefined)
  loadRequirements()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadRequirements()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增需求'
  formData.id = undefined
  formData.title = ''
  formData.description = ''
  formData.status = 'pending'
  formData.priority = 'medium'
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_requirement_project_form')
  formData.project_id = lastProjectId
  formData.assignee_id = undefined
  formData.estimated_hours = undefined
  formData.actual_hours = undefined
  formData.work_date = undefined
  formData.attachment_ids = []
  requirementAttachments.value = []
  modalVisible.value = true
}

// 编辑
const handleEdit = async (record: Requirement) => {
  modalTitle.value = '编辑需求'
  formData.id = record.id
  formData.title = record.title
  formData.description = record.description || ''
  formData.status = record.status
  formData.priority = record.priority
  formData.project_id = record.project_id
  formData.assignee_id = record.assignee_id
  formData.estimated_hours = record.estimated_hours
  formData.actual_hours = record.actual_hours
  formData.work_date = undefined
  
  // 加载需求附件
  try {
    requirementAttachments.value = await getAttachments({ requirement_id: record.id })
    formData.attachment_ids = requirementAttachments.value.map(a => a.id)
  } catch (error: any) {
    console.error('加载附件失败:', error)
    requirementAttachments.value = []
    formData.attachment_ids = []
  }
  
  modalVisible.value = true
}

// 查看详情
const handleView = (record: Requirement) => {
  router.push(`/requirement/${record.id}`)
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    // 验证项目ID必填
    if (!formData.project_id) {
      message.error('请选择项目')
      return
    }
    // 上传Markdown编辑器中的本地图片
    let description = formData.description || ''
    if (descriptionEditorRef.value && formData.project_id) {
      try {
        description = await descriptionEditorRef.value.uploadLocalImages(async (file: File, projectId: number) => {
          const attachment = await uploadFile(file, projectId)
          console.log('图片上传成功，附件信息:', attachment)
          return attachment
        })
        console.log('上传图片后的description:', description)
      } catch (error: any) {
        console.error('上传图片失败:', error)
        message.warning('部分图片上传失败，请检查')
      }
    }
    
    const data: any = {
      title: formData.title,
      description: description, // 使用已上传图片的description
      status: formData.status,
      priority: formData.priority,
      project_id: formData.project_id, // 必填
      assignee_id: formData.assignee_id,
      estimated_hours: formData.estimated_hours,
      actual_hours: formData.actual_hours,
      work_date: formData.work_date ? formData.work_date.format('YYYY-MM-DD') : undefined
    }
    // 调试：检查提交的数据
    console.log('提交的数据:', {
      description: data.description,
      hasImages: data.description?.includes('/uploads/')
    })
    
    let requirementId: number
    if (formData.id) {
      requirementId = formData.id
      await updateRequirement(requirementId, data)
      message.success('更新成功')
    } else {
      const newRequirement = await createRequirement(data)
      requirementId = newRequirement.id
      message.success('创建成功')
      
      // 创建需求后，如果有待上传的附件，需要关联到需求
      // 附件上传组件会在上传时自动关联到项目，这里需要额外关联到需求
      if (formData.attachment_ids && formData.attachment_ids.length > 0 && formData.project_id) {
        try {
          for (const attachmentId of formData.attachment_ids) {
            await attachToEntity(attachmentId, { requirement_id: requirementId })
          }
        } catch (error: any) {
          console.error('关联附件到需求失败:', error)
        }
      }
    }
    modalVisible.value = false
    loadRequirements()
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
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteRequirement(id)
    message.success('删除成功')
    loadRequirements()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 状态变更
const handleStatusChange = async (id: number, status: string) => {
  try {
    await updateRequirementStatus(id, { status: status as any })
    message.success('状态更新成功')
    loadRequirements()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'orange',
    in_progress: 'blue',
    completed: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待处理',
    in_progress: '进行中',
    completed: '已完成',
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
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_requirement_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  
  // 读取路由查询参数
  if (route.query.status) {
    searchForm.status = route.query.status as string
  }
  if (route.query.assignee === 'me' && authStore.user) {
    searchForm.assignee_id = authStore.user.id
  }
  
  loadRequirements()
  loadProjects()
  loadUsers()
})
</script>

<style scoped>
.requirement-management {
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

