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
                <a-button @click="handleAssign">分配</a-button>
                <a-dropdown>
                  <a-button>
                    状态 <DownOutlined />
                  </a-button>
                  <template #overlay>
                    <a-menu @click="(e: any) => handleStatusChange(e.key as string)">
                      <a-menu-item key="open">待处理</a-menu-item>
                      <a-menu-item key="assigned">已分配</a-menu-item>
                      <a-menu-item key="in_progress">处理中</a-menu-item>
                      <a-menu-item key="resolved">已解决</a-menu-item>
                      <a-menu-item key="closed">已关闭</a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
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
                  <a-tag :color="getStatusColor(bug?.status || '')">
                    {{ getStatusText(bug?.status || '') }}
                  </a-tag>
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
                <a-descriptions-item label="分配人">
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

    <!-- Bug分配模态框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="分配Bug"
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
        <a-form-item label="分配人" name="assignee_ids">
          <a-select
            v-model:value="assignFormData.assignee_ids"
            mode="multiple"
            placeholder="选择分配人"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getBug,
  updateBugStatus,
  deleteBug,
  assignBug,
  type Bug
} from '@/api/bug'
import { getUsers, type User } from '@/api/user'

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

const assignFormRules = {
  assignee_ids: [{ required: true, message: '请选择分配人', trigger: 'change' }]
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

// 分配
const handleAssign = () => {
  if (!bug.value) return
  assignFormData.assignee_ids = bug.value.assignees?.map(a => a.id) || []
  assignModalVisible.value = true
}

// 分配提交
const handleAssignSubmit = async () => {
  if (!bug.value) return
  try {
    await assignFormRef.value.validate()
    await assignBug(bug.value.id, { assignee_ids: assignFormData.assignee_ids })
    message.success('分配成功')
    assignModalVisible.value = false
    loadBug()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '分配失败')
  }
}

// 分配取消
const handleAssignCancel = () => {
  assignFormRef.value?.resetFields()
}

// 状态变更
const handleStatusChange = async (status: string) => {
  if (!bug.value) return
  try {
    await updateBugStatus(bug.value.id, { status: status as any })
    message.success('状态更新成功')
    loadBug()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
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
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    assigned: '已分配',
    in_progress: '处理中',
    resolved: '已解决',
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

