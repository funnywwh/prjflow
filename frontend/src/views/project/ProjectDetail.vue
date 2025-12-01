<template>
  <div class="project-detail">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header
            :title="project?.name || '项目详情'"
            :sub-title="project?.code"
            @back="() => router.push('/project')"
          >
            <template #extra>
              <a-space>
                <a-button @click="handleManageRequirements">需求管理</a-button>
                <a-button @click="handleManageTasks">任务管理</a-button>
                <a-button @click="handleManageBugs">Bug管理</a-button>
                <a-button @click="handleViewBoards">看板</a-button>
                <a-button @click="handleViewGantt">甘特图</a-button>
                <a-button @click="handleViewProgress">进度跟踪</a-button>
                <a-button @click="handleViewResourceStatistics">资源统计</a-button>
                <a-button @click="handleManageModules">功能模块</a-button>
                <a-button @click="handleEdit">编辑</a-button>
                <a-button @click="handleManageMembers">成员管理</a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-spin :spinning="loading">
            <!-- 项目基本信息 -->
            <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
              <a-descriptions :column="2" bordered>
                <a-descriptions-item label="项目名称">{{ project?.name }}</a-descriptions-item>
                <a-descriptions-item label="项目编码">{{ project?.code }}</a-descriptions-item>
                <!-- <a-descriptions-item label="项目集">{{ project?.project_group?.name || '-' }}</a-descriptions-item> -->
                <!-- <a-descriptions-item label="关联产品">{{ project?.product?.name || '-' }}</a-descriptions-item> -->
                <a-descriptions-item label="开始日期">{{ project?.start_date || '-' }}</a-descriptions-item>
                <a-descriptions-item label="结束日期">{{ project?.end_date || '-' }}</a-descriptions-item>
                <a-descriptions-item label="状态">
                  <a-tag :color="project?.status === 'doing' || project?.status === 'wait' ? 'green' : 'red'">
                    {{ project?.status === 'doing' || project?.status === 'wait' ? '正常' : '禁用' }}
                  </a-tag>
                </a-descriptions-item>
                <a-descriptions-item label="成员数">{{ statistics?.total_members || 0 }} 人</a-descriptions-item>
                <a-descriptions-item label="描述" :span="2">
                  <div v-if="project?.description" class="description-content">
                    <MarkdownEditor
                      :model-value="project.description"
                      :readonly="true"
                    />
                  </div>
                  <span v-else>-</span>
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 项目描述 -->
            <a-card title="项目描述" :bordered="false" style="margin-bottom: 16px" v-if="project?.description">
              <div class="markdown-content">
                <MarkdownEditor
                  :model-value="project.description"
                  :readonly="true"
                />
              </div>
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

            <!-- 统计概览 -->
            <a-row :gutter="16" style="margin-bottom: 16px">
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总任务数"
                    :value="statistics?.total_tasks || 0"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总Bug数"
                    :value="statistics?.total_bugs || 0"
                    :value-style="{ color: '#ff4d4f' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="总需求数"
                    :value="statistics?.total_requirements || 0"
                    :value-style="{ color: '#52c41a' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card :bordered="false">
                  <a-statistic
                    title="项目成员"
                    :value="statistics?.total_members || 0"
                    suffix="人"
                    :value-style="{ color: '#722ed1' }"
                  />
                </a-card>
              </a-col>
            </a-row>

            <!-- 任务统计 -->
            <a-card title="任务统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('todo')">
                    <a-statistic
                      title="待办"
                      :value="statistics?.todo_tasks || 0"
                      :value-style="{ color: '#1890ff' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('in_progress')">
                    <a-statistic
                      title="进行中"
                      :value="statistics?.in_progress_tasks || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToTasks('done')">
                    <a-statistic
                      title="已完成"
                      :value="statistics?.done_tasks || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- Bug统计 -->
            <a-card title="Bug统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('open')">
                    <a-statistic
                      title="待处理"
                      :value="statistics?.open_bugs || 0"
                      :value-style="{ color: '#ff4d4f' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('in_progress')">
                    <a-statistic
                      title="处理中"
                      :value="statistics?.in_progress_bugs || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="8">
                  <a-card class="stat-card" @click="goToBugs('resolved')">
                    <a-statistic
                      title="已解决"
                      :value="statistics?.resolved_bugs || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- 需求统计 -->
            <a-card title="需求统计" :bordered="false" style="margin-bottom: 16px">
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-card class="stat-card" @click="goToRequirements('in_progress')">
                    <a-statistic
                      title="进行中"
                      :value="statistics?.in_progress_requirements || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-card>
                </a-col>
                <a-col :span="12">
                  <a-card class="stat-card" @click="goToRequirements('completed')">
                    <a-statistic
                      title="已完成"
                      :value="statistics?.completed_requirements || 0"
                      :value-style="{ color: '#52c41a' }"
                    />
                  </a-card>
                </a-col>
              </a-row>
            </a-card>

            <!-- 项目成员 -->
            <a-card title="项目成员" :bordered="false">
              <template #extra>
                <a-button type="link" @click="handleManageMembers">成员管理</a-button>
              </template>
              <a-list
                :data-source="project?.members || []"
                :loading="loading"
              >
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #avatar>
                        <a-avatar :src="item.user?.avatar">
                          {{ (item.user?.nickname || item.user?.username)?.charAt(0).toUpperCase() }}
                        </a-avatar>
                      </template>
                      <template #title>
                        {{ item.user?.username }}{{ item.user?.nickname ? `(${item.user.nickname})` : '' }}
                      </template>
                      <template #description>
                        <a-tag>{{ item.role }}</a-tag>
                        <span v-if="item.user?.department" style="margin-left: 8px; color: #999">
                          {{ item.user.department.name }}
                        </span>
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

    <!-- 项目成员管理对话框 -->
    <a-modal
      v-model:open="memberModalVisible"
      title="项目成员管理"
      :mask-closable="true"
      @cancel="handleCloseMemberModal"
      @ok="handleCloseMemberModal"
      ok-text="关闭"
      width="800px"
    >
      <a-spin :spinning="memberLoading">
        <div style="margin-bottom: 16px">
          <a-space>
            <a-select
              v-model:value="selectedUserIds"
              mode="multiple"
              placeholder="选择用户"
              style="width: 300px"
              show-search
              :filter-option="(input: string, option: any) => {
                const user = users.find(u => u.id === option.value)
                if (!user) return false
                const searchText = input.toLowerCase()
                return user.username.toLowerCase().includes(searchText) ||
                  (user.nickname && user.nickname.toLowerCase().includes(searchText))
              }"
            >
              <a-select-option
                v-for="user in users"
                :key="user.id"
                :value="user.id"
              >
                {{ user.username }}{{ user.nickname ? `(${user.nickname})` : '' }}
              </a-select-option>
            </a-select>
            <a-select
              v-model:value="memberRole"
              placeholder="选择角色"
              style="width: 150px"
            >
              <a-select-option value="owner">负责人</a-select-option>
              <a-select-option value="member">成员</a-select-option>
              <a-select-option value="viewer">查看者</a-select-option>
            </a-select>
            <a-button type="primary" @click="handleAddMembers">添加成员</a-button>
          </a-space>
        </div>
        <a-table
          :columns="memberColumns"
          :data-source="projectMembers"
          :scroll="{ x: 'max-content' }"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'user'">
              {{ record.user?.username || '-' }}{{ record.user?.nickname ? `(${record.user.nickname})` : '' }}
            </template>
            <template v-else-if="column.key === 'role'">
              <a-select
                :value="record.role"
                @change="(value: any) => handleUpdateMemberRole(record.id, value)"
                style="width: 120px"
              >
                <a-select-option value="owner">负责人</a-select-option>
                <a-select-option value="member">成员</a-select-option>
                <a-select-option value="viewer">查看者</a-select-option>
              </a-select>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-popconfirm
                title="确定要移除这个成员吗？"
                @confirm="handleRemoveMember(record.id)"
              >
                <a-button type="link" size="small" danger>移除</a-button>
              </a-popconfirm>
            </template>
          </template>
        </a-table>
      </a-spin>
    </a-modal>

    <!-- 功能模块管理对话框 -->
    <a-modal
      v-model:open="moduleManageModalVisible"
      title="功能模块管理（系统资源）"
      :mask-closable="true"
      @cancel="handleCloseModuleModal"
      width="900px"
      :footer="null"
    >
      <ModuleManagement :show-card="false" />
    </a-modal>

    <!-- 项目编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑项目"
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
        <a-form-item label="项目名称" name="name">
          <a-input v-model:value="editFormData.name" placeholder="请输入项目名称" />
        </a-form-item>
        <a-form-item label="项目编码" name="code">
          <a-input v-model:value="editFormData.code" placeholder="请输入项目编码" />
        </a-form-item>
        <a-form-item label="项目描述" name="description">
          <MarkdownEditor
            ref="editDescriptionEditorRef"
            v-model="editFormData.description"
            placeholder="请输入项目描述（支持Markdown）"
            :rows="8"
            :project-id="project?.id || 0"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="editFormData.status">
            <a-select-option value="wait">等待</a-select-option>
            <a-select-option value="doing">进行中</a-select-option>
            <a-select-option value="suspended">已暂停</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
            <a-select-option value="done">已完成</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="标签" name="tag_ids">
          <a-select
            v-model:value="editFormData.tag_ids"
            mode="multiple"
            placeholder="选择标签（支持多选）"
            allow-clear
            :options="tagOptions"
            :field-names="{ label: 'name', value: 'id' }"
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
import { message } from 'ant-design-vue'
import { formatDateTime } from '@/utils/date'
import dayjs, { type Dayjs } from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import ModuleManagement from '@/components/ModuleManagement.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import { 
  getProject, 
  updateProject,
  getProjectMembers,
  addProjectMembers,
  updateProjectMember,
  removeProjectMember,
  getProjectHistory,
  addProjectHistoryNote,
  type ProjectDetailResponse, 
  type Project,
  type ProjectMember,
  type CreateProjectRequest,
  type Action
} from '@/api/project'
import { getUsers, type User } from '@/api/user'
import { getTags, type Tag } from '@/api/tag'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const project = ref<Project>()
const statistics = ref<any>()

// 成员管理相关
const memberModalVisible = ref(false)
const memberLoading = ref(false)
const users = ref<User[]>([])
const projectMembers = ref<ProjectMember[]>([])
const selectedUserIds = ref<number[]>([])
const memberRole = ref('member')

// 功能模块管理相关
const moduleManageModalVisible = ref(false)

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
const editFormData = reactive<Omit<CreateProjectRequest, 'start_date' | 'end_date'> & { 
  start_date?: Dayjs | undefined
  end_date?: Dayjs | undefined
}>({
  name: '',
  code: '',
  description: '',
  status: 'wait',
  tag_ids: [],
  start_date: undefined,
  end_date: undefined
})
const editFormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入项目编码', trigger: 'blur' }]
}
const tags = ref<Tag[]>([])
const tagOptions = ref<Array<{ id: number; name: string; color?: string }>>([])

const memberColumns = [
  { title: '用户', key: 'user', width: 150 },
  { title: '角色', key: 'role', width: 150 },
  { title: '操作', key: 'action', width: 100 }
]

// 加载项目详情
const loadProject = async () => {
  const projectId = Number(route.params.id)
  if (!projectId) {
    message.error('项目ID无效')
    router.push('/project')
    return
  }

  loading.value = true
  try {
    const data: ProjectDetailResponse = await getProject(projectId)
    project.value = data.project
    statistics.value = data.statistics
    await loadProjectHistory(projectId) // 加载历史记录
  } catch (error: any) {
    message.error(error.message || '加载项目详情失败')
    router.push('/project')
  } finally {
    loading.value = false
  }
}

// 查看看板
const handleViewBoards = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/boards`)
}

// 查看甘特图
const handleViewGantt = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/gantt`)
}

// 查看进度跟踪
const handleViewProgress = () => {
  if (!project.value) return
  router.push(`/project/${project.value.id}/progress`)
}

// 查看资源统计
const handleViewResourceStatistics = () => {
  if (!project.value) return
  router.push({
    path: '/resource/statistics',
    query: { project_id: project.value.id }
  })
}

// 编辑项目
const handleEdit = async () => {
  if (!project.value) return
  
  editFormData.name = project.value.name
  editFormData.code = project.value.code
  editFormData.description = project.value.description || ''
  editFormData.status = project.value.status
  editFormData.tag_ids = project.value.tags?.map(t => t.id) || []
  editFormData.start_date = project.value.start_date ? dayjs(project.value.start_date) : undefined
  editFormData.end_date = project.value.end_date ? dayjs(project.value.end_date) : undefined
  
  editModalVisible.value = true
  await loadTags()
}

// 编辑提交
const handleEditSubmit = async () => {
  if (!project.value) return
  
  try {
    await editFormRef.value.validate()
    
    // 获取最新的描述内容
    let description = editFormData.description || ''
    
    // 如果有项目ID，尝试上传本地图片（如果有的话）
    if (editDescriptionEditorRef.value && project.value.id) {
      try {
        const uploadedDescription = await editDescriptionEditorRef.value.uploadLocalImages(async (file: File, projectId: number) => {
          // TODO: 需要实现文件上传API
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
    
    const data: Partial<CreateProjectRequest> = {
      name: editFormData.name,
      code: editFormData.code,
      description: description || '',
      status: editFormData.status,
      tag_ids: editFormData.tag_ids,
      start_date: editFormData.start_date && typeof editFormData.start_date !== 'string' && 'isValid' in editFormData.start_date && (editFormData.start_date as Dayjs).isValid() ? (editFormData.start_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.start_date === 'string' ? editFormData.start_date : undefined),
      end_date: editFormData.end_date && typeof editFormData.end_date !== 'string' && 'isValid' in editFormData.end_date && (editFormData.end_date as Dayjs).isValid() ? (editFormData.end_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.end_date === 'string' ? editFormData.end_date : undefined)
    }
    
    await updateProject(project.value.id, data)
    
    message.success('更新成功')
    editModalVisible.value = false
    await loadProject() // 重新加载项目详情（会自动加载历史记录）
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

// 加载标签
const loadTags = async () => {
  try {
    tags.value = await getTags()
    tagOptions.value = tags.value.map(t => ({ id: t.id, name: t.name, color: t.color }))
  } catch (error: any) {
    console.error('加载标签列表失败:', error)
  }
}

// 加载历史记录
const loadProjectHistory = async (projectId?: number) => {
  const id = projectId || Number(route.params.id)
  if (!id) return

  historyLoading.value = true
  try {
    const response = await getProjectHistory(id)
    historyList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    historyLoading.value = false
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
    name: '项目名称',
    code: '项目编码',
    description: '项目描述',
    status: '状态',
    start_date: '开始日期',
    end_date: '结束日期',
    tag_ids: '标签'
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
  if (!project.value) {
    message.warning('项目信息未加载完成，请稍候再试')
    return
  }
  noteFormData.comment = ''
  noteModalVisible.value = true
}

// 提交备注
const handleNoteSubmit = async () => {
  if (!project.value) return
  try {
    await noteFormRef.value.validate()
    await addProjectHistoryNote(project.value.id, { comment: noteFormData.comment })
    message.success('添加备注成功')
    noteModalVisible.value = false
    await loadProjectHistory(project.value.id)
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

// 加载用户列表
const loadUsers = async () => {
  try {
    const response = await getUsers({ size: 1000 })
    users.value = response.list || []
  } catch (error: any) {
    console.error('加载用户列表失败:', error)
  }
}

// 加载项目成员
const loadProjectMembers = async (projectId: number) => {
  memberLoading.value = true
  try {
    projectMembers.value = await getProjectMembers(projectId)
  } catch (error: any) {
    message.error(error.message || '加载项目成员失败')
  } finally {
    memberLoading.value = false
  }
}

// 成员管理
const handleManageMembers = async () => {
  if (!project.value) return
  memberModalVisible.value = true
  selectedUserIds.value = []
  memberRole.value = 'member'
  await loadProjectMembers(project.value.id)
  if (users.value.length === 0) {
    await loadUsers()
  }
}

// 关闭成员管理对话框
const handleCloseMemberModal = async () => {
  memberModalVisible.value = false
  selectedUserIds.value = []
  memberRole.value = 'member'
  // 重新加载项目详情以更新成员列表
  if (project.value) {
    await loadProject()
  }
}

// 添加成员
const handleAddMembers = async () => {
  if (!project.value || selectedUserIds.value.length === 0) {
    message.warning('请选择用户')
    return
  }
  try {
    await addProjectMembers(project.value.id, {
      user_ids: selectedUserIds.value,
      role: memberRole.value
    })
    message.success('添加成功')
    selectedUserIds.value = []
    await loadProjectMembers(project.value.id)
    // 重新加载项目详情
    await loadProject()
  } catch (error: any) {
    message.error(error.message || '添加失败')
  }
}

// 更新成员角色
const handleUpdateMemberRole = async (memberId: number, role: string) => {
  if (!project.value) return
  try {
    await updateProjectMember(project.value.id, memberId, role)
    message.success('更新成功')
    await loadProjectMembers(project.value.id)
    // 重新加载项目详情
    await loadProject()
  } catch (error: any) {
    message.error(error.message || '更新失败')
  }
}

// 移除成员
const handleRemoveMember = async (memberId: number) => {
  if (!project.value) return
  try {
    await removeProjectMember(project.value.id, memberId)
    message.success('移除成功')
    await loadProjectMembers(project.value.id)
    // 重新加载项目详情
    await loadProject()
  } catch (error: any) {
    message.error(error.message || '移除失败')
  }
}

// 功能模块管理
const handleManageModules = () => {
  moduleManageModalVisible.value = true
}

// 关闭模块管理对话框
const handleCloseModuleModal = () => {
  moduleManageModalVisible.value = false
}

// 需求管理
const handleManageRequirements = () => {
  if (!project.value) return
  router.push({
    path: '/requirement',
    query: { project_id: project.value.id }
  })
}

// 任务管理
const handleManageTasks = () => {
  if (!project.value) return
  router.push({
    path: '/task',
    query: { project_id: project.value.id }
  })
}

// Bug管理
const handleManageBugs = () => {
  if (!project.value) return
  router.push({
    path: '/bug',
    query: { project_id: project.value.id }
  })
}

// 跳转到任务列表
const goToTasks = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/task',
    query: { status, project_id: project.value.id }
  })
}

// 跳转到Bug列表
const goToBugs = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/bug',
    query: { status, project_id: project.value.id }
  })
}

// 跳转到需求列表
const goToRequirements = (status: string) => {
  if (!project.value) return
  router.push({
    path: '/requirement',
    query: { status, project_id: project.value.id }
  })
}

onMounted(() => {
  loadProject()
})
</script>

<style scoped>
.project-detail {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.project-detail :deep(.ant-layout) {
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
  background: white;
  padding: 24px;
  border-radius: 4px;
  min-height: fit-content;
}

.description-content {
  max-width: 100%;
}

.stat-card {
  cursor: pointer;
  transition: all 0.3s;
  text-align: center;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
</style>

