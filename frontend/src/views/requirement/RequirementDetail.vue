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
                <a-button @click="handleAssign">指派</a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="draft">草稿</a-menu-item>
                      <a-menu-item key="reviewing">评审中</a-menu-item>
                      <a-menu-item key="active">激活</a-menu-item>
                      <a-menu-item key="changing">变更中</a-menu-item>
                      <a-menu-item key="closed">已关闭</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
                <a-button @click="handleConvertToTask">
                  转任务
                </a-button>
                <a-button @click="handleConvertToBug">
                  转Bug
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

          <RequirementDetailContent
            :requirement="requirement"
            :loading="loading"
            :history-list="historyList"
            :history-loading="historyLoading"
            @add-note="handleAddNote"
          />
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
          <ProjectMemberSelect
            v-model="editFormData.assignee_id"
            :project-id="requirement?.project_id"
            :multiple="false"
            placeholder="选择负责人（可选）"
            :show-role="true"
            :show-hint="!requirement?.project_id"
          />
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

    <!-- 需求指派模态框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="指派需求"
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
            :project-id="requirement?.project_id"
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
            <a-select-option value="draft">草稿</a-select-option>
            <a-select-option value="reviewing">评审中</a-select-option>
            <a-select-option value="active">激活</a-select-option>
            <a-select-option value="changing">变更中</a-select-option>
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
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import ProjectMemberSelect from '@/components/ProjectMemberSelect.vue'
import RequirementDetailContent from '@/components/RequirementDetailContent.vue'
import {
  getRequirement,
  updateRequirement,
  updateRequirementStatus,
  deleteRequirement,
  getRequirementHistory,
  addRequirementHistoryNote,
  assignRequirement,
  type Requirement,
  type CreateRequirementRequest,
  type Action
} from '@/api/requirement'
import { createBug, type CreateBugRequest } from '@/api/bug'
import { createTask, type CreateTaskRequest } from '@/api/task'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const requirement = ref<Requirement | null>(null)

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

// 指派
const handleAssign = () => {
  if (!requirement.value) return
  // 设置默认值：如果当前状态是 "draft" 或 "reviewing"，默认选择 "active"；否则默认不选择（自动修改）
  if (requirement.value.status === 'draft' || requirement.value.status === 'reviewing') {
    assignFormData.status = 'active'
  } else {
    assignFormData.status = undefined
  }
  assignFormData.assignee_id = requirement.value.assignee_id // 预填充当前指派人
  assignFormData.comment = undefined // 清空备注
  assignModalVisible.value = true
}

// 指派提交
const handleAssignSubmit = async () => {
  if (!requirement.value) return
  try {
    await assignFormRef.value.validate()
    const requestData: any = { assignee_id: assignFormData.assignee_id }
    if (assignFormData.status) {
      requestData.status = assignFormData.status
    }
    if (assignFormData.comment) {
      requestData.comment = assignFormData.comment
    }
    await assignRequirement(requirement.value.id, requestData)
    message.success('指派成功')
    assignModalVisible.value = false
    await loadRequirement() // 重新加载需求详情（会自动加载历史记录）
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

// 需求转任务
const handleConvertToTask = async () => {
  if (!requirement.value) return
  
  // 确认对话框
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此需求转为任务吗？转换后将创建新任务，并关联到此需求。',
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
    // 创建新任务，基于需求的信息
    const taskData: CreateTaskRequest = {
      title: `[转任务] ${requirement.value.title}`,
      description: requirement.value.description 
        ? `${requirement.value.description}\n\n---\n\n*由需求 #${requirement.value.id}转换而来*`
        : `*由需求 #${requirement.value.id}转换而来*`,
      project_id: requirement.value.project_id,
      priority: requirement.value.priority,
      status: 'wait', // 任务默认未开始状态
      requirement_id: requirement.value.id, // 关联原需求
      assignee_id: requirement.value.assignee_id,
      estimated_hours: requirement.value.estimated_hours
    }
    
    // 创建任务
    const task = await createTask(taskData)
    
    message.success(`转换成功，已创建任务 #${task.id}`)
    
    // 刷新需求详情
    await loadRequirement()
    
    // 可选：跳转到新创建的任务详情页
    // router.push(`/task/${task.id}`)
  } catch (error: any) {
    message.error(error.message || '转换失败')
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
      title: `[转Bug] ${requirement.value.title}`,
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

