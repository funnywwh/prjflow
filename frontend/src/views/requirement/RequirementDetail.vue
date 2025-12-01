<template>
  <div class="requirement-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="requirement?.title || '需求详情'"
            @back="() => router.push('/requirement')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEdit">编辑</a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="pending">待处理</a-menu-item>
                      <a-menu-item key="in_progress">进行中</a-menu-item>
                      <a-menu-item key="completed">已完成</a-menu-item>
                      <a-menu-item key="cancelled">已取消</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
                <a-button @click="handleConvertToBug">
                  需求转Bug
                </a-button>
                <a-popconfirm
                  title="确定要删除这个需求吗？"
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
                <a-descriptions-item label="需求标题">{{ requirement?.title }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="getStatusColor(requirement?.status || '')">
                    {{ getStatusText(requirement?.status || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="优先级">
                  <a-tag :color="getPriorityColor(requirement?.priority || '')">
                    {{ getPriorityText(requirement?.priority || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="产品">
                  <!-- {{ requirement?.product?.name || '-' }} -->
                </a-descriptions-item>
                <a-descriptions-item label="项目">
                  {{ requirement?.project?.name || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="负责人">
                  {{ requirement?.assignee ? `${requirement.assignee.username}${requirement.assignee.nickname ? `(${requirement.assignee.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建人">
                  {{ requirement?.creator ? `${requirement.creator.username}${requirement.creator.nickname ? `(${requirement.creator.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建时间">
                  {{ formatDateTime(requirement?.created_at) }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ formatDateTime(requirement?.updated_at) }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 需求描述 -->
            <a-card title="需求描述" :bordered="false" style="margin-bottom: 16px">
              <div v-if="requirement?.description" class="markdown-content">
                <MarkdownEditor
                  :model-value="requirement.description"
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
                            style="margin-bottom: 4px; color: #666"
                          >
                            修改了{{ getFieldDisplayName(history.field) }}, 旧值为"{{ history.old_value || history.old || '-' }}",新值为"{{ history.new_value || history.new || '-' }}"。
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
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 需求编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑需求"
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
        <a-form-item label="需求标题" name="title">
          <a-input v-model:value="editFormData.title" placeholder="请输入需求标题" />
        </a-form-item>
        <a-form-item label="需求描述" name="description">
          <MarkdownEditor
            ref="editDescriptionEditorRef"
            v-model="editFormData.description"
            placeholder="请输入需求描述（支持Markdown）"
            :rows="8"
            :project-id="requirement?.project_id || 0"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="editFormData.status">
            <a-select-option value="draft">草稿</a-select-option>
            <a-select-option value="reviewing">评审中</a-select-option>
            <a-select-option value="active">激活</a-select-option>
            <a-select-option value="changing">变更中</a-select-option>
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
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getRequirement,
  updateRequirement,
  updateRequirementStatus,
  deleteRequirement,
  getRequirementHistory,
  addRequirementHistoryNote,
  type Requirement,
  type CreateRequirementRequest,
  type Action
} from '@/api/requirement'
import { getUsers, type User } from '@/api/user'
import { createBug, type CreateBugRequest } from '@/api/bug'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const requirement = ref<Requirement | null>(null)
const users = ref<User[]>([])

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
const editFormData = reactive<CreateRequirementRequest>({
  title: '',
  description: '',
  status: 'draft',
  priority: 'medium',
  project_id: 0,
  assignee_id: undefined,
  estimated_hours: undefined
})
const editFormRules = {
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载需求详情
const loadRequirement = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('需求ID无效')
    router.push('/requirement')
    return
  }

  loading.value = true
  try {
    requirement.value = await getRequirement(id)
    await loadRequirementHistory(id) // 加载历史记录
  } catch (error: any) {
    message.error(error.message || '加载需求详情失败')
    router.push('/requirement')
  } finally {
    loading.value = false
  }
}

// 加载历史记录
const loadRequirementHistory = async (requirementId?: number) => {
  const id = requirementId || Number(route.params.id)
  if (!id) return

  historyLoading.value = true
  try {
    const response = await getRequirementHistory(id)
    historyList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    historyLoading.value = false
  }
}

// 编辑
const handleEdit = async () => {
  if (!requirement.value) return
  
  editFormData.title = requirement.value.title
  editFormData.description = requirement.value.description || ''
  editFormData.status = requirement.value.status
  editFormData.priority = requirement.value.priority
  editFormData.project_id = requirement.value.project_id
  editFormData.assignee_id = requirement.value.assignee_id
  editFormData.estimated_hours = requirement.value.estimated_hours
  
  editModalVisible.value = true
  if (users.value.length === 0) {
    await loadUsers()
  }
}

// 编辑提交
const handleEditSubmit = async () => {
  if (!requirement.value) return
  
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
    
    const data: Partial<CreateRequirementRequest> = {
      title: editFormData.title,
      description: description || '',
      status: editFormData.status,
      priority: editFormData.priority,
      assignee_id: editFormData.assignee_id,
      estimated_hours: editFormData.estimated_hours
    }
    
    await updateRequirement(requirement.value.id, data)
    
    message.success('更新成功')
    editModalVisible.value = false
    await loadRequirement() // 重新加载需求详情（会自动加载历史记录）
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

// 用户筛选
const filterUserOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
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
    title: '需求标题',
    description: '需求描述',
    status: '状态',
    priority: '优先级',
    assignee_id: '负责人',
    estimated_hours: '预估工时',
    actual_hours: '实际工时'
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
  if (!requirement.value) {
    message.warning('需求信息未加载完成，请稍候再试')
    return
  }
  noteFormData.comment = ''
  noteModalVisible.value = true
}

// 提交备注
const handleNoteSubmit = async () => {
  if (!requirement.value) return
  try {
    await noteFormRef.value.validate()
    await addRequirementHistoryNote(requirement.value.id, { comment: noteFormData.comment })
    message.success('添加备注成功')
    noteModalVisible.value = false
    await loadRequirementHistory(requirement.value.id)
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

// 状态变更
const handleStatusChange = async (status: string) => {
  if (!requirement.value) return
  try {
    await updateRequirementStatus(requirement.value.id, { status: status as any })
    message.success('状态更新成功')
    loadRequirement()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 需求转Bug
const handleConvertToBug = async () => {
  if (!requirement.value) return
  
  // 确认对话框
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此需求转为Bug吗？转换后将创建新Bug，并关联到此需求。',
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
    // 创建新Bug，基于需求的信息
    const bugData: CreateBugRequest = {
      title: `[需求转Bug] ${requirement.value.title}`,
      description: requirement.value.description 
        ? `${requirement.value.description}\n\n---\n\n*由需求 #${requirement.value.id}转换而来*`
        : `*由需求 #${requirement.value.id}转换而来*`,
      project_id: requirement.value.project_id,
      priority: requirement.value.priority,
      severity: 'medium', // Bug默认严重程度
      status: 'active', // Bug默认激活状态
      requirement_id: requirement.value.id, // 关联原需求
      // 如果需求有负责人，作为Bug的指派人员
      assignee_ids: requirement.value.assignee_id 
        ? [requirement.value.assignee_id] 
        : undefined,
      estimated_hours: requirement.value.estimated_hours
    }
    
    // 创建Bug
    const bug = await createBug(bugData)
    
    message.success(`转换成功，已创建Bug #${bug.id}`)
    
    // 刷新需求详情
    await loadRequirement()
    
    // 可选：跳转到新创建的Bug详情页
    // router.push(`/bug/${bug.id}`)
  } catch (error: any) {
    message.error(error.message || '转换失败')
  }
}

// 删除
const handleDelete = async () => {
  if (!requirement.value) return
  try {
    await deleteRequirement(requirement.value.id)
    message.success('删除成功')
    router.push('/requirement')
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'orange',
    reviewing: 'purple',
    active: 'blue',
    changing: 'cyan',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    reviewing: '评审中',
    active: '激活',
    changing: '变更中',
    closed: '已关闭',
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

onMounted(() => {
  loadRequirement()
})
</script>

<style scoped>
.requirement-detail {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.requirement-detail :deep(.ant-layout) {
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

