<template>
  <div class="bug-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="Bug管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增Bug
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="Bug标题/描述"
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
                  <a-select-option value="open">待处理</a-select-option>
                  <a-select-option value="assigned">已分配</a-select-option>
                  <a-select-option value="in_progress">处理中</a-select-option>
                  <a-select-option value="resolved">已解决</a-select-option>
                  <a-select-option value="closed">已关闭</a-select-option>
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
              <a-form-item label="严重程度">
                <a-select
                  v-model:value="searchForm.severity"
                  placeholder="选择严重程度"
                  allow-clear
                  style="width: 120px"
                >
                  <a-select-option value="low">低</a-select-option>
                  <a-select-option value="medium">中</a-select-option>
                  <a-select-option value="high">高</a-select-option>
                  <a-select-option value="critical">严重</a-select-option>
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
                  title="总Bug数"
                  :value="statistics?.total || 0"
                  :value-style="{ color: '#ff4d4f' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="待处理"
                  :value="statistics?.open || 0"
                  :value-style="{ color: '#faad14' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="处理中"
                  :value="statistics?.in_progress || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="已解决"
                  :value="statistics?.resolved || 0"
                  :value-style="{ color: '#52c41a' }"
                />
              </a-card>
            </a-col>
          </a-row>

          <!-- 优先级和严重程度统计 -->
          <a-row :gutter="16" style="margin-bottom: 16px">
            <a-col :span="12">
              <a-card title="优先级统计" :bordered="false">
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
            </a-col>
            <a-col :span="12">
              <a-card title="严重程度统计" :bordered="false">
                <a-row :gutter="16">
                  <a-col :span="6">
                    <a-statistic
                      title="低"
                      :value="statistics?.low_severity || 0"
                      :value-style="{ color: '#8c8c8c' }"
                    />
                  </a-col>
                  <a-col :span="6">
                    <a-statistic
                      title="中"
                      :value="statistics?.medium_severity || 0"
                      :value-style="{ color: '#1890ff' }"
                    />
                  </a-col>
                  <a-col :span="6">
                    <a-statistic
                      title="高"
                      :value="statistics?.high_severity || 0"
                      :value-style="{ color: '#faad14' }"
                    />
                  </a-col>
                  <a-col :span="6">
                    <a-statistic
                      title="严重"
                      :value="statistics?.critical_severity || 0"
                      :value-style="{ color: '#ff4d4f' }"
                    />
                  </a-col>
                </a-row>
              </a-card>
            </a-col>
          </a-row>

          <a-card :bordered="false">
            <a-table
              :columns="columns"
              :data-source="bugs"
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
                <template v-else-if="column.key === 'severity'">
                  <a-tag :color="getSeverityColor(record.severity)">
                    {{ getSeverityText(record.severity) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'assignees'">
                  <a-tag
                    v-for="assignee in record.assignees || []"
                    :key="assignee.id"
                    style="margin-right: 4px"
                  >
                    {{ assignee.username }}{{ assignee.nickname ? `(${assignee.nickname})` : '' }}
                  </a-tag>
                  <span v-if="!record.assignees || record.assignees.length === 0">-</span>
                </template>
                <template v-else-if="column.key === 'requirement'">
                  {{ record.requirement?.title || '-' }}
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
                    <a-button type="link" size="small" @click="handleAssign(record)">
                      分配
                    </a-button>
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleOpenStatusModal(record, e.key as string)">
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

    <!-- Bug编辑/创建模态框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="800"
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
        <a-form-item label="Bug标题" name="title">
          <a-input v-model:value="formData.title" placeholder="请输入Bug标题" />
        </a-form-item>
        <a-form-item label="Bug描述" name="description">
          <MarkdownEditor
            v-model="formData.description"
            placeholder="请输入Bug描述（支持Markdown）"
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
        <a-form-item label="关联需求" name="requirement_id">
          <a-select
            v-model:value="formData.requirement_id"
            placeholder="选择关联需求（可选）"
            allow-clear
            show-search
            :filter-option="filterRequirementOption"
            :loading="requirementLoading"
            @focus="loadRequirementsForProject"
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
          <a-select v-model:value="formData.status">
            <a-select-option value="open">待处理</a-select-option>
            <a-select-option value="assigned">已分配</a-select-option>
            <a-select-option value="in_progress">处理中</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
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
        <a-form-item label="严重程度" name="severity">
          <a-select v-model:value="formData.severity">
            <a-select-option value="low">低</a-select-option>
            <a-select-option value="medium">中</a-select-option>
            <a-select-option value="high">高</a-select-option>
            <a-select-option value="critical">严重</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="分配人" name="assignee_ids">
          <a-select
            v-model:value="formData.assignee_ids"
            mode="multiple"
            placeholder="选择分配人（可选）"
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
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配（使用第一个分配人）</span>
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
      </a-form>
    </a-modal>

    <!-- Bug分配模态框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="分配Bug"
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

    <!-- Bug状态更新模态框 -->
    <a-modal
      v-model:open="statusModalVisible"
      title="更新Bug状态"
      :width="600"
      @ok="handleStatusSubmit"
      @cancel="handleStatusCancel"
    >
      <a-form
        ref="statusFormRef"
        :model="statusFormData"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="新状态">
          <a-select v-model:value="statusFormData.status" disabled>
            <a-select-option value="open">待处理</a-select-option>
            <a-select-option value="assigned">已分配</a-select-option>
            <a-select-option value="in_progress">处理中</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
          </a-select>
        </a-form-item>
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
              @focus="loadVersionsForProject"
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
              placeholder="请输入版本号（如：v1.0.0）"
            />
          </a-space>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime } from '@/utils/date'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getBugs,
  createBug,
  updateBug,
  deleteBug,
  updateBugStatus,
  assignBug,
  getBugStatistics,
  type Bug,
  type CreateBugRequest,
  type BugStatistics
} from '@/api/bug'
import { getProjects, type Project } from '@/api/project'
import { getUsers, type User } from '@/api/user'
import { getRequirements, type Requirement } from '@/api/requirement'
import { getVersions, type Version } from '@/api/version'

const router = useRouter()
const loading = ref(false)
const bugs = ref<Bug[]>([])
const projects = ref<Project[]>([])
const users = ref<User[]>([])
const requirements = ref<Requirement[]>([])
const requirementLoading = ref(false)
const versions = ref<Version[]>([])
const versionLoading = ref(false)
const statistics = ref<BugStatistics | null>(null)

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  priority: undefined as string | undefined,
  severity: undefined as string | undefined
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
  { title: 'Bug标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '项目', key: 'project', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '优先级', key: 'priority', width: 100 },
  { title: '严重程度', key: 'severity', width: 100 },
  { title: '分配人', key: 'assignees', width: 200 },
  { title: '工时', key: 'hours', width: 150 },
  { title: '关联需求', key: 'requirement', width: 150, ellipsis: true },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 300, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增Bug')
const formRef = ref()
const formData = reactive<CreateBugRequest & { id?: number }>({
  title: '',
  description: '',
  status: 'open',
  priority: 'medium',
  severity: 'medium',
  project_id: 0,
  requirement_id: undefined,
  assignee_ids: [],
  estimated_hours: undefined
})

const formRules = {
  title: [{ required: true, message: '请输入Bug标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

const assignModalVisible = ref(false)
const assignFormRef = ref()
const assignFormData = reactive({
  bug_id: 0,
  assignee_ids: [] as number[]
})

const statusModalVisible = ref(false)
const statusFormRef = ref()
const statusFormData = reactive({
  bug_id: 0,
  status: 'open' as string,
  solution: undefined as string | undefined,
  solution_note: undefined as string | undefined,
  resolved_version_id: undefined as number | undefined,
  version_number: undefined as string | undefined,
  create_version: false
})

const assignFormRules = {
  assignee_ids: [{ required: true, message: '请选择分配人', trigger: 'change' }]
}

// 加载Bug列表
const loadBugs = async () => {
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
    if (searchForm.severity) {
      params.severity = searchForm.severity
    }
    const response = await getBugs(params)
    bugs.value = response.list
    pagination.total = response.total
    // 加载统计信息
    await loadStatistics()
  } catch (error: any) {
    message.error(error.message || '加载Bug列表失败')
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
    statistics.value = await getBugStatistics(params)
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

// 加载需求列表（根据项目）
const loadRequirementsForProject = async () => {
  if (!formData.project_id) {
    requirements.value = []
    return
  }
  requirementLoading.value = true
  try {
    const response = await getRequirements({ project_id: formData.project_id })
    requirements.value = response.list || []
  } catch (error: any) {
    console.error('加载需求列表失败:', error)
  } finally {
    requirementLoading.value = false
  }
}

// 监听项目变化，重新加载需求
watch(() => formData.project_id, () => {
  formData.requirement_id = undefined
  if (formData.project_id) {
    loadRequirementsForProject()
  } else {
    requirements.value = []
  }
})

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadBugs()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.priority = undefined
  searchForm.severity = undefined
  pagination.current = 1
  loadBugs()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadBugs()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增Bug'
  formData.id = undefined
  formData.title = ''
  formData.description = ''
  formData.status = 'open'
  formData.priority = 'medium'
  formData.severity = 'medium'
  formData.project_id = 0
  formData.requirement_id = undefined
  formData.assignee_ids = []
  formData.estimated_hours = undefined
  formData.actual_hours = undefined
  formData.work_date = undefined
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: Bug) => {
  modalTitle.value = '编辑Bug'
  formData.id = record.id
  formData.title = record.title
  formData.description = record.description || ''
  formData.status = record.status
  formData.priority = record.priority
  formData.severity = record.severity
  formData.project_id = record.project_id
  formData.requirement_id = record.requirement_id
  formData.assignee_ids = record.assignees?.map(a => a.id) || []
  formData.estimated_hours = record.estimated_hours
  formData.actual_hours = record.actual_hours
  formData.work_date = undefined
  modalVisible.value = true
  if (formData.project_id) {
    loadRequirementsForProject()
  }
}

// 查看详情
const handleView = (record: Bug) => {
  router.push(`/bug/${record.id}`)
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const data: CreateBugRequest = {
      title: formData.title,
      description: formData.description,
      status: formData.status,
      priority: formData.priority,
      severity: formData.severity,
      project_id: formData.project_id,
      requirement_id: formData.requirement_id,
      assignee_ids: formData.assignee_ids,
      estimated_hours: formData.estimated_hours,
      actual_hours: formData.actual_hours,
      work_date: formData.work_date && formData.work_date.isValid() ? formData.work_date.format('YYYY-MM-DD') : undefined
    }
    if (formData.id) {
      await updateBug(formData.id, data)
      message.success('更新成功')
    } else {
      await createBug(data)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadBugs()
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
  requirements.value = []
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteBug(id)
    message.success('删除成功')
    loadBugs()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 状态变更
// 打开状态更新对话框
const handleOpenStatusModal = (record: Bug, status: string) => {
  statusFormData.bug_id = record.id
  statusFormData.status = status
  statusFormData.solution = record.solution
  statusFormData.solution_note = record.solution_note
  statusFormData.resolved_version_id = record.resolved_version_id
  statusFormData.version_number = undefined
  statusFormData.create_version = false
  // 加载项目下的版本列表
  if (record.project_id) {
    loadVersionsForProject(record.project_id)
  }
  statusModalVisible.value = true
}

// 加载项目下的版本列表
const loadVersionsForProject = async (projectId?: number) => {
  let pid = projectId
  if (!pid && statusFormData.bug_id) {
    const bug = bugs.value.find(b => b.id === statusFormData.bug_id)
    pid = bug?.project_id
  }
  if (!pid) {
    versions.value = []
    return
  }
  try {
    versionLoading.value = true
    const response = await getVersions({ project_id: pid, page_size: 1000 })
    versions.value = response.list || []
  } catch (error: any) {
    console.error('加载版本列表失败:', error)
  } finally {
    versionLoading.value = false
  }
}

// 状态更新提交
const handleStatusSubmit = async () => {
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
    if (statusFormData.create_version && statusFormData.version_number) {
      data.create_version = true
      data.version_number = statusFormData.version_number
    } else if (statusFormData.resolved_version_id) {
      data.resolved_version_id = statusFormData.resolved_version_id
    }
    await updateBugStatus(statusFormData.bug_id, data)
    message.success('状态更新成功')
    statusModalVisible.value = false
    loadBugs()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 状态更新取消
const handleStatusCancel = () => {
  statusModalVisible.value = false
  statusFormData.bug_id = 0
  statusFormData.status = 'open'
  statusFormData.solution = undefined
  statusFormData.solution_note = undefined
  statusFormData.resolved_version_id = undefined
  statusFormData.version_number = undefined
  statusFormData.create_version = false
}

// 版本筛选
const filterVersionOption = (input: string, option: any) => {
  const version = versions.value.find(v => v.id === option.value)
  if (!version) return false
  const searchText = input.toLowerCase()
  return version.version_number.toLowerCase().includes(searchText)
}

// 分配
const handleAssign = (record: Bug) => {
  assignFormData.bug_id = record.id
  assignFormData.assignee_ids = record.assignees?.map(a => a.id) || []
  assignModalVisible.value = true
}

// 分配提交
const handleAssignSubmit = async () => {
  try {
    await assignFormRef.value.validate()
    await assignBug(assignFormData.bug_id, { assignee_ids: assignFormData.assignee_ids })
    message.success('分配成功')
    assignModalVisible.value = false
    loadBugs()
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

// 获取状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    open: 'orange',
    assigned: 'blue',
    in_progress: 'processing',
    resolved: 'green',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    open: '待处理',
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

// 需求筛选
const filterRequirementOption = (input: string, option: any) => {
  const requirement = requirements.value.find(r => r.id === option.value)
  if (!requirement) return false
  const searchText = input.toLowerCase()
  return requirement.title.toLowerCase().includes(searchText)
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
  loadBugs()
  loadProjects()
  loadUsers()
})
</script>

<style scoped>
.bug-management {
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

