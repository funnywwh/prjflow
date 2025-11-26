<template>
  <div class="bug-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="bug?.title || 'Bug详情'"
            @back="() => router.push('/bug')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleEdit">编辑</a-button>
                <a-button @click="handleAssign">指派</a-button>
                <a-button
                  v-if="bug?.status === 'active' && !bug?.confirmed"
                  @click="handleConfirm"
                >
                  确认
                </a-button>
                <a-button
                  v-if="bug?.status === 'active'"
                  @click="handleResolve"
                >
                  解决
                </a-button>
                <a-popconfirm
                  title="确定要删除这个Bug吗？"
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
                <a-descriptions-item label="Bug标题">{{ bug?.title }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-space>
                    <a-tag :color="getStatusColor(bug?.status || '')">
                      {{ getStatusText(bug?.status || '') }}
                    </a-tag>
                    <a-tag v-if="bug?.confirmed" color="green">已确认</a-tag>
                    <a-tag v-else-if="bug?.status === 'active'" color="orange">未确认</a-tag>
                  </a-space>
                </a-descriptions-item>
                <a-descriptions-item label="优先级">
                  <a-tag :color="getPriorityColor(bug?.priority || '')">
                    {{ getPriorityText(bug?.priority || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="严重程度">
                  <a-tag :color="getSeverityColor(bug?.severity || '')">
                    {{ getSeverityText(bug?.severity || '') }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="项目">
                  {{ bug?.project?.name || '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="关联需求">
                  <a v-if="bug?.requirement" @click="router.push(`/requirement/${bug.requirement.id}`)" style="cursor: pointer">
                    {{ bug.requirement.title }}
                  </a>
                  <span v-else>-</span>
                </a-descriptions-item>
                <a-descriptions-item label="指派给">
                  <a-space>
                    <a-tag
                      v-for="assignee in bug?.assignees || []"
                      :key="assignee.id"
                    >
                      {{ assignee.username }}{{ assignee.nickname ? `(${assignee.nickname})` : '' }}
                    </a-tag>
                    <span v-if="!bug?.assignees || bug.assignees.length === 0">-</span>
                  </a-space>
                </a-descriptions-item>
                <a-descriptions-item label="创建人">
                  {{ bug?.creator ? `${bug.creator.username}${bug.creator.nickname ? `(${bug.creator.nickname})` : ''}` : '-' }}
                </a-descriptions-item>
                <a-descriptions-item label="创建时间">
                  {{ formatDateTime(bug?.created_at) }}
                </a-descriptions-item>
                <a-descriptions-item label="更新时间">
                  {{ formatDateTime(bug?.updated_at) }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- Bug描述 -->
            <a-card title="Bug描述" :bordered="false" style="margin-bottom: 16px">
              <div v-if="bug?.description" class="markdown-content">
                <MarkdownEditor
                  :model-value="bug.description"
                  :readonly="true"
                />
              </div>
              <a-empty v-else description="暂无描述" />
            </a-card>
          </a-spin>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- Bug指派模态框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="指派Bug"
      :mask-closable="true"
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
        <a-form-item label="指派给" name="assignee_ids">
          <a-select
            v-model:value="assignFormData.assignee_ids"
            mode="multiple"
            placeholder="选择指派给"
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
      </a-form>
    </a-modal>

    <!-- Bug解决模态框 -->
    <a-modal
      v-model:open="statusModalVisible"
      title="解决Bug"
      :width="600"
      :mask-closable="true"
      @ok="handleStatusSubmit"
      @cancel="handleStatusCancel"
    >
      <a-form
        ref="statusFormRef"
        :model="statusFormData"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="解决方案" name="solution">
          <a-select
            v-model:value="statusFormData.solution"
            placeholder="选择解决方案（可选）"
            allow-clear
          >
            <a-select-option value="设计如此">设计如此</a-select-option>
            <a-select-option value="重复Bug">重复Bug</a-select-option>
            <a-select-option value="外部原因">外部原因</a-select-option>
            <a-select-option value="已解决">已解决</a-select-option>
            <a-select-option value="无法重现">无法重现</a-select-option>
            <a-select-option value="延期处理">延期处理</a-select-option>
            <a-select-option value="不予解决">不予解决</a-select-option>
            <a-select-option value="转为研发需求">转为研发需求</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="备注" name="solution_note">
          <a-textarea
            v-model:value="statusFormData.solution_note"
            placeholder="请输入备注（可选）"
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="预估工时" name="estimated_hours">
          <a-input-number
            v-model:value="statusFormData.estimated_hours"
            placeholder="预估工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="实际工时" name="actual_hours">
          <a-input-number
            v-model:value="statusFormData.actual_hours"
            placeholder="实际工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配</span>
        </a-form-item>
        <a-form-item label="工作日期" name="work_date" v-if="statusFormData.actual_hours">
          <a-date-picker
            v-model:value="statusFormData.work_date"
            placeholder="选择工作日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
          <span style="margin-left: 8px; color: #999">不填则使用今天</span>
        </a-form-item>
        <a-form-item label="解决版本" name="resolved_version_id">
          <a-space direction="vertical" style="width: 100%">
            <a-select
              v-model:value="statusFormData.resolved_version_id"
              placeholder="选择版本（可选）"
              allow-clear
              show-search
              :filter-option="filterVersionOption"
              :loading="versionLoading"
              :disabled="statusFormData.create_version"
              @focus="loadVersionsForProject(bug?.project_id || 0)"
            >
              <a-select-option
                v-for="version in versions"
                :key="version.id"
                :value="version.id"
              >
                {{ version.version_number }}
              </a-select-option>
            </a-select>
            <a-checkbox v-model:checked="statusFormData.create_version">
              创建新版本
            </a-checkbox>
            <a-input
              v-if="statusFormData.create_version"
              v-model:value="statusFormData.version_number"
              placeholder="输入版本号"
            />
          </a-space>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getBug,
  updateBugStatus,
  deleteBug,
  assignBug,
  confirmBug,
  type Bug
} from '@/api/bug'
import { getUsers, type User } from '@/api/user'
import { getVersions, type Version } from '@/api/version'
import type { Dayjs } from 'dayjs'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const bug = ref<Bug | null>(null)
const users = ref<User[]>([])
const assignModalVisible = ref(false)
const assignFormRef = ref()
const assignFormData = reactive({
  assignee_ids: [] as number[]
})

// 解决对话框相关
const statusModalVisible = ref(false)
const statusFormRef = ref()
const statusFormData = reactive({
  status: 'resolved' as string,
  solution: undefined as string | undefined,
  solution_note: undefined as string | undefined,
  estimated_hours: undefined as number | undefined,
  actual_hours: undefined as number | undefined,
  work_date: undefined as Dayjs | undefined,
  resolved_version_id: undefined as number | undefined,
  version_number: undefined as string | undefined,
  create_version: false
})
const versions = ref<Version[]>([])
const versionLoading = ref(false)

const assignFormRules = {
  assignee_ids: [{ required: true, message: '请选择指派给', trigger: 'change' }]
}

// 加载Bug详情
const loadBug = async () => {
  const id = Number(route.params.id)
  if (!id) {
    message.error('Bug ID无效')
    router.push('/bug')
    return
  }

  loading.value = true
  try {
    bug.value = await getBug(id)
  } catch (error: any) {
    message.error(error.message || '加载Bug详情失败')
    router.push('/bug')
  } finally {
    loading.value = false
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

// 编辑
const handleEdit = () => {
  router.push(`/bug?edit=${bug.value?.id}`)
}

// 指派
const handleAssign = () => {
  if (!bug.value) return
  assignFormData.assignee_ids = bug.value.assignees?.map(a => a.id) || []
  assignModalVisible.value = true
}

// 指派提交
const handleAssignSubmit = async () => {
  if (!bug.value) return
  try {
    await assignFormRef.value.validate()
    await assignBug(bug.value.id, { assignee_ids: assignFormData.assignee_ids })
    message.success('指派成功')
    assignModalVisible.value = false
    loadBug()
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
}

// 解决Bug（弹出对话框）
const handleResolve = () => {
  if (!bug.value) return
  handleOpenStatusModal('resolved')
}

// 打开解决对话框
const handleOpenStatusModal = (status: string) => {
  if (!bug.value) return
  statusFormData.status = status
  statusFormData.solution = bug.value.solution
  statusFormData.solution_note = bug.value.solution_note
  statusFormData.estimated_hours = bug.value.estimated_hours
  statusFormData.actual_hours = bug.value.actual_hours
  statusFormData.work_date = undefined
  statusFormData.resolved_version_id = bug.value.resolved_version_id
  statusFormData.version_number = undefined
  statusFormData.create_version = false
  // 加载项目下的版本列表
  if (bug.value.project_id) {
    loadVersionsForProject(bug.value.project_id)
  }
  statusModalVisible.value = true
}

// 加载项目下的版本列表
const loadVersionsForProject = async (projectId: number) => {
  if (!projectId) {
    versions.value = []
    return
  }
  try {
    versionLoading.value = true
    const response = await getVersions({ project_id: projectId, size: 1000 })
    versions.value = response.list || []
  } catch (error: any) {
    console.error('加载版本列表失败:', error)
  } finally {
    versionLoading.value = false
  }
}

// 版本筛选
const filterVersionOption = (input: string, option: any) => {
  const version = versions.value.find(v => v.id === option.value)
  if (!version) return false
  const searchText = input.toLowerCase()
  return version.version_number.toLowerCase().includes(searchText)
}

// 解决提交
const handleStatusSubmit = async () => {
  if (!bug.value) return
  try {
    const data: any = {
      status: statusFormData.status
    }
    if (statusFormData.solution) {
      data.solution = statusFormData.solution
    }
    if (statusFormData.solution_note) {
      data.solution_note = statusFormData.solution_note
    }
    if (statusFormData.estimated_hours !== undefined) {
      data.estimated_hours = statusFormData.estimated_hours
    }
    if (statusFormData.actual_hours !== undefined) {
      data.actual_hours = statusFormData.actual_hours
      if (statusFormData.work_date && statusFormData.work_date.isValid()) {
        data.work_date = statusFormData.work_date.format('YYYY-MM-DD')
      }
    }
    if (statusFormData.create_version && statusFormData.version_number) {
      data.create_version = true
      data.version_number = statusFormData.version_number
    } else if (statusFormData.resolved_version_id) {
      data.resolved_version_id = statusFormData.resolved_version_id
    }
    await updateBugStatus(bug.value.id, data)
    message.success('解决成功')
    statusModalVisible.value = false
    loadBug()
  } catch (error: any) {
    message.error(error.message || '解决失败')
  }
}

// 解决取消
const handleStatusCancel = () => {
  statusModalVisible.value = false
  statusFormData.status = 'resolved'
  statusFormData.solution = undefined
  statusFormData.solution_note = undefined
  statusFormData.estimated_hours = undefined
  statusFormData.actual_hours = undefined
  statusFormData.work_date = undefined
  statusFormData.resolved_version_id = undefined
  statusFormData.version_number = undefined
  statusFormData.create_version = false
}


// 确认Bug
const handleConfirm = async () => {
  if (!bug.value) return
  try {
    await confirmBug(bug.value.id)
    message.success('确认成功')
    loadBug()
  } catch (error: any) {
    message.error(error.message || '确认失败')
  }
}

// 删除
const handleDelete = async () => {
  if (!bug.value) return
  try {
    await deleteBug(bug.value.id)
    message.success('删除成功')
    router.push('/bug')
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    active: 'orange',
    resolved: 'green',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string | undefined) => {
  if (!status) return '-'
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return texts[status.toLowerCase()] || status
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

// 获取严重程度颜色
const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    critical: 'red'
  }
  return colors[severity] || 'default'
}

// 获取严重程度文本
const getSeverityText = (severity: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return texts[severity] || severity
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
  loadBug()
  loadUsers()
})
</script>

<style scoped>
.bug-detail {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content-inner {
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
}

.markdown-content {
  min-height: 200px;
}
</style>

