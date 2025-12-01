<template>
  <div class="resource-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="资源管理">
            <template #extra>
              <a-space>
                <a-button @click="handleViewCalendar">资源日历</a-button>
                <a-button @click="handleViewStatistics">资源统计</a-button>
                <a-button @click="handleViewUtilization">利用率分析</a-button>
                <a-button type="primary" @click="handleCreate">
                  <template #icon><PlusOutlined /></template>
                  新增资源
                </a-button>
              </a-space>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <template #title>
              <a-space>
                <span>搜索条件</span>
                <a-button type="text" size="small" @click="toggleSearchForm">
                  <template #icon>
                    <UpOutlined v-if="searchFormVisible" />
                    <DownOutlined v-else />
                  </template>
                  {{ searchFormVisible ? '收起' : '展开' }}
                </a-button>
              </a-space>
            </template>
            <a-form v-show="searchFormVisible" layout="inline" :model="searchForm">
              <a-form-item label="用户">
                <a-select
                  v-model:value="searchForm.user_id"
                  placeholder="选择用户"
                  allow-clear
                  show-search
                  :filter-option="filterUserOption"
                  style="width: 150px"
                >
                  <a-select-option
                    v-for="user in users"
                    :key="user.id"
                    :value="user.id"
                  >
                    {{ user.nickname || user.username }}({{ user.username }})
                  </a-select-option>
                </a-select>
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
              <a-form-item label="角色">
                <a-input
                  v-model:value="searchForm.role"
                  placeholder="角色"
                  allow-clear
                  style="width: 120px"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-table
            :columns="columns"
            :data-source="resources"
            :loading="loading"
            :scroll="{ x: 'max-content', y: tableScrollHeight }"
            :pagination="pagination"
            @change="handleTableChange"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'user'">
                <span>{{ record.user?.nickname || record.user?.username }}({{ record.user?.username }})</span>
              </template>
              <template v-else-if="column.key === 'project'">
                <span>{{ record.project?.name }}</span>
              </template>
              <template v-else-if="column.key === 'created_at'">
                {{ formatDateTime(record.created_at) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space>
                  <a-button type="link" size="small" @click="handleManageAllocations(record)">
                    分配管理
                  </a-button>
                  <a-button type="link" size="small" @click="handleEdit(record)">
                    编辑
                  </a-button>
                  <a-popconfirm
                    title="确定要删除这个资源吗？"
                    @confirm="handleDelete(record.id!)"
                  >
                    <a-button type="link" size="small" danger>删除</a-button>
                  </a-popconfirm>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 资源创建/编辑模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="600"
      @ok="handleSubmit"
      @cancel="handleCancel"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="用户" name="user_id">
          <a-select
            v-model:value="formData.user_id"
            placeholder="选择用户"
            show-search
            :filter-option="filterUserOption"
            @change="handleUserChange"
          >
            <a-select-option
              v-for="user in users"
              :key="user.id"
              :value="user.id"
            >
              {{ user.nickname || user.username }}({{ user.username }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="选择项目"
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
        <a-form-item label="角色" name="role">
          <a-input
            v-model:value="formData.role"
            placeholder="资源角色（可选）"
            allow-clear
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 资源分配管理模态框 -->
    <a-modal
      v-model:open="allocationModalVisible"
      title="资源分配管理"
      :width="1000"
      :mask-closable="true"
      :footer="null"
      
      >
      <div style="margin-bottom: 16px">
        <a-space>
          <span>资源：{{ currentResource?.user?.nickname || currentResource?.user?.username }}({{ currentResource?.user?.username }})</span>
          <span>项目：{{ currentResource?.project?.name }}</span>
          <a-button type="primary" @click="handleCreateAllocation">
            <template #icon><PlusOutlined /></template>
            新增分配
          </a-button>
        </a-space>
      </div>
      <a-table
        :columns="allocationColumns"
        :data-source="allocations"
        :loading="allocationLoading"
        :scroll="{ x: 'max-content' }"
        :pagination="allocationPagination"
        @change="handleAllocationTableChange"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'date'">
            <span>{{ record.date }}</span>
          </template>
          <template v-else-if="column.key === 'hours'">
            <span>{{ record.hours }} 小时</span>
          </template>
          <template v-else-if="column.key === 'task'">
            <span v-if="record.task">{{ record.task.name }}</span>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.key === 'bug'">
            <span v-if="record.bug">{{ record.bug.title }}</span>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.key === 'project'">
            <span v-if="record.project">{{ record.project.name }}</span>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="handleEditAllocation(record)">
                编辑
              </a-button>
              <a-popconfirm
                title="确定要删除这个分配吗？"
                @confirm="handleDeleteAllocation(record.id!)"
              >
                <a-button type="link" size="small" danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-modal>

    <!-- 资源分配创建/编辑模态框 -->
    <a-modal
      :mask-closable="true"
      v-model:open="allocationFormVisible"
      :title="allocationModalTitle"
      :width="600"
      @ok="handleAllocationSubmit"
      @cancel="handleAllocationCancel"
    >
      <a-form
        ref="allocationFormRef"
        :model="allocationFormData"
        :rules="allocationRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="日期" name="date">
          <a-date-picker
            v-model:value="allocationFormData.date"
            placeholder="选择日期"
            style="width: 100%"
            :disabled-date="disabledDate"
          />
        </a-form-item>
        <a-form-item label="工时" name="hours">
          <a-input-number
            v-model:value="allocationFormData.hours"
            placeholder="工时（小时）"
            :min="0"
            :max="24"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="关联任务" name="task_id">
          <a-select
            v-model:value="allocationFormData.task_id"
            placeholder="选择任务（可选）"
            allow-clear
            show-search
            :filter-option="filterTaskOption"
            :loading="tasksLoading"
            @change="handleTaskChange"
          >
            <a-select-option
              v-for="task in tasks"
              :key="task.id"
              :value="task.id"
            >
              {{ task.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关联Bug" name="bug_id">
          <a-select
            v-model:value="allocationFormData.bug_id"
            placeholder="选择Bug（可选）"
            allow-clear
            show-search
            :filter-option="filterBugOption"
            :loading="bugsLoading"
            @change="handleBugChange"
          >
            <a-select-option
              v-for="bug in bugs"
              :key="bug.id"
              :value="bug.id"
            >
              {{ bug.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关联项目" name="project_id">
          <a-select
            v-model:value="allocationFormData.project_id"
            placeholder="选择项目（可选）"
            allow-clear
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
        <a-form-item label="工作描述" name="description">
          <a-textarea
            v-model:value="allocationFormData.description"
            placeholder="工作描述（可选）"
            :rows="4"
            allow-clear
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import {
  getResources,
  createResource,
  updateResource,
  deleteResource,
  getResourceAllocations,
  createResourceAllocation,
  updateResourceAllocation,
  deleteResourceAllocation,
  type Resource,
  type ResourceAllocation,
  type CreateResourceRequest,
  type UpdateResourceRequest,
  type CreateResourceAllocationRequest,
  type UpdateResourceAllocationRequest
} from '@/api/resource'
import { getUsers } from '@/api/user'
import { getProjects } from '@/api/project'
import { getTasks } from '@/api/task'
import { getBugs } from '@/api/bug'
import type { User } from '@/api/user'
import type { Project } from '@/api/project'
import type { Task } from '@/api/task'
import type { Bug } from '@/api/bug'
import { PlusOutlined, DownOutlined, UpOutlined } from '@ant-design/icons-vue'

const router = useRouter()

const loading = ref(false)
const searchFormVisible = ref(false) // 搜索栏显示/隐藏状态，默认折叠
const resources = ref<Resource[]>([])
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const tasks = ref<Task[]>([])
const tasksLoading = ref(false)
const bugs = ref<Bug[]>([])
const bugsLoading = ref(false)

const searchForm = reactive({
  user_id: undefined as number | undefined,
  project_id: undefined as number | undefined,
  role: undefined as string | undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

// 计算表格滚动高度
const tableScrollHeight = computed(() => {
  return 'calc(100vh - 400px)'
})

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '用户', key: 'user', width: 200 },
  { title: '项目', key: 'project', width: 200 },
  { title: '角色', dataIndex: 'role', key: 'role', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' }
]

const modalVisible = ref(false)
const modalTitle = ref('新增资源')
const formRef = ref<FormInstance>()
const formData = reactive<CreateResourceRequest & { id?: number }>({
  user_id: 0,
  project_id: 0,
  role: ''
})

const rules = {
  user_id: [{ required: true, message: '请选择用户', trigger: 'change' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 资源分配相关
const allocationModalVisible = ref(false)
const allocationLoading = ref(false)
const allocations = ref<ResourceAllocation[]>([])
const currentResource = ref<Resource | null>(null)

const allocationPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const allocationColumns = [
  { title: '日期', key: 'date', width: 120 },
  { title: '工时', key: 'hours', width: 100 },
  { title: '关联任务', key: 'task', width: 200 },
  { title: '关联Bug', key: 'bug', width: 200 },
  { title: '关联项目', key: 'project', width: 200 },
  { title: '工作描述', dataIndex: 'description', key: 'description' },
  { title: '操作', key: 'action', width: 150, fixed: 'right' }
]

const allocationFormVisible = ref(false)
const allocationModalTitle = ref('新增分配')
const allocationFormRef = ref<FormInstance>()
const allocationFormData = reactive<Omit<CreateResourceAllocationRequest, 'date'> & { id?: number; date?: Dayjs | undefined }>({
  resource_id: 0,
  date: undefined,
  hours: 0,
  task_id: undefined,
  bug_id: undefined,
  project_id: undefined,
  description: ''
})

const allocationRules = {
  resource_id: [{ required: true, message: '资源ID不能为空', trigger: 'change' }],
  date: [{ required: true, message: '请选择日期', trigger: 'change' }],
  hours: [
    { required: true, message: '请输入工时', trigger: 'blur' },
    { type: 'number', min: 0.01, max: 24, message: '工时必须在0.01-24小时之间', trigger: 'blur' }
  ]
}

const filterUserOption = (input: string, option: any) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  const nickname = user.nickname || ''
  const username = user.username || ''
  return nickname.toLowerCase().includes(input.toLowerCase()) ||
    username.toLowerCase().includes(input.toLowerCase())
}

const filterTaskOption = (input: string, option: any) => {
  const task = tasks.value.find(t => t.id === option.value)
  if (!task) return false
  return task.title?.toLowerCase().includes(input.toLowerCase()) || false
}

const filterBugOption = (input: string, option: any) => {
  const bug = bugs.value.find(b => b.id === option.value)
  if (!bug) return false
  return bug.title?.toLowerCase().includes(input.toLowerCase()) || false
}

const disabledDate = (current: Dayjs) => {
  return current && current > dayjs().endOf('day')
}

const loadUsers = async () => {
  try {
    const res = await getUsers({ size: 1000 })
    users.value = res.list || []
  } catch (error: any) {
    message.error('加载用户列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadProjects = async () => {
  try {
    const res = await getProjects({ size: 1000 })
    projects.value = res.list || []
  } catch (error: any) {
    message.error('加载项目列表失败: ' + (error.response?.data?.message || error.message))
  }
}

const loadTasksForProject = async (projectId?: number) => {
  if (!projectId) return
  tasksLoading.value = true
  try {
    const res = await getTasks({ project_id: projectId, size: 1000 })
    tasks.value = res.list || []
  } catch (error: any) {
    message.error('加载任务列表失败: ' + (error.response?.data?.message || error.message))
  } finally {
    tasksLoading.value = false
  }
}

const loadBugsForProject = async (projectId?: number) => {
  if (!projectId) return
  bugsLoading.value = true
  try {
    const res = await getBugs({ project_id: projectId, size: 1000 })
    bugs.value = res.list || []
  } catch (error: any) {
    message.error('加载Bug列表失败: ' + (error.response?.data?.message || error.message))
  } finally {
    bugsLoading.value = false
  }
}

const handleTaskChange = () => {
  // 当选择任务时，可以自动填充项目ID（如果任务有项目）
  // 这里可以根据需要添加逻辑
}

const handleBugChange = () => {
  // 当选择Bug时，可以自动填充项目ID（如果Bug有项目）
  // 这里可以根据需要添加逻辑
}

const loadResources = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.user_id) params.user_id = searchForm.user_id
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (searchForm.role) params.role = searchForm.role

    const res = await getResources(params)
    resources.value = res.list || []
    pagination.total = res.total || 0
  } catch (error: any) {
    message.error('加载资源列表失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

const loadAllocations = async () => {
  if (!currentResource.value?.id) return
  allocationLoading.value = true
  try {
    const params: any = {
      resource_id: currentResource.value.id,
      page: allocationPagination.current,
      size: allocationPagination.pageSize
    }
    const res = await getResourceAllocations(params)
    allocations.value = res.list || []
    allocationPagination.total = res.total || 0
  } catch (error: any) {
    message.error('加载分配列表失败: ' + (error.response?.data?.message || error.message))
  } finally {
    allocationLoading.value = false
  }
}

// 切换搜索栏显示/隐藏
const toggleSearchForm = () => {
  searchFormVisible.value = !searchFormVisible.value
}

const handleSearch = () => {
  pagination.current = 1
  loadResources()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_resource_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_resource_project_form', value || 0)
  // 原有的 handleProjectChange 逻辑
  handleProjectChange()
}

const handleReset = () => {
  searchForm.user_id = undefined
  searchForm.project_id = undefined
  searchForm.role = undefined
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_resource_project_search', undefined)
  handleSearch()
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadResources()
}

const handleCreate = () => {
  modalTitle.value = '新增资源'
  formData.id = undefined
  formData.user_id = 0
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_resource_project_form')
  formData.project_id = lastProjectId || 0
  formData.role = ''
  modalVisible.value = true
}

const handleEdit = (record: Resource) => {
  modalTitle.value = '编辑资源'
  formData.id = record.id
  formData.user_id = record.user_id
  formData.project_id = record.project_id
  formData.role = record.role || ''
  modalVisible.value = true
}

const handleUserChange = () => {
  // 可以在这里添加用户变更的逻辑
}

const handleProjectChange = () => {
  // 可以在这里添加项目变更的逻辑
}

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()

    if (!formData.user_id || formData.user_id === 0) {
      message.error('请选择用户')
      return
    }
    if (!formData.project_id || formData.project_id === 0) {
      message.error('请选择项目')
      return
    }

    if (formData.id) {
      const updateData: UpdateResourceRequest = {}
      if (formData.role !== undefined) updateData.role = formData.role
      await updateResource(formData.id, updateData)
      message.success('更新成功')
    } else {
      const createData: CreateResourceRequest = {
        user_id: formData.user_id,
        project_id: formData.project_id,
        role: formData.role || undefined
      }
      await createResource(createData)
      message.success('创建成功')
    }

    modalVisible.value = false
    loadResources()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error('操作失败: ' + (error.response?.data?.message || error.message))
  }
}

const handleCancel = () => {
  modalVisible.value = false
  formRef.value?.resetFields()
}

const handleDelete = async (id: number) => {
  try {
    await deleteResource(id)
    message.success('删除成功')
    loadResources()
  } catch (error: any) {
    message.error('删除失败: ' + (error.response?.data?.message || error.message))
  }
}

const handleManageAllocations = (record: Resource) => {
  currentResource.value = record
  allocationModalVisible.value = true
  allocationPagination.current = 1
  loadAllocations()
}

const handleAllocationTableChange = (pag: any) => {
  allocationPagination.current = pag.current
  allocationPagination.pageSize = pag.pageSize
  loadAllocations()
}

const handleCreateAllocation = () => {
  if (!currentResource.value?.id) {
    message.error('资源信息不存在')
    return
  }
  allocationModalTitle.value = '新增分配'
  allocationFormData.id = undefined
  allocationFormData.resource_id = currentResource.value.id
  allocationFormData.date = undefined as Dayjs | undefined
  allocationFormData.hours = 0
  allocationFormData.task_id = undefined
  allocationFormData.bug_id = undefined
  allocationFormData.project_id = currentResource.value.project_id
  allocationFormData.description = ''
  if (currentResource.value.project_id) {
    loadTasksForProject(currentResource.value.project_id)
    loadBugsForProject(currentResource.value.project_id)
  }
  allocationFormVisible.value = true
}

const handleEditAllocation = (record: ResourceAllocation) => {
  allocationModalTitle.value = '编辑分配'
  allocationFormData.id = record.id
  allocationFormData.resource_id = record.resource_id
  if (record.date) {
    allocationFormData.date = dayjs(record.date) as Dayjs | undefined
  } else {
    allocationFormData.date = undefined as Dayjs | undefined
  }
  allocationFormData.hours = record.hours
  allocationFormData.task_id = record.task_id
  allocationFormData.bug_id = record.bug_id
  allocationFormData.project_id = record.project_id
  allocationFormData.description = record.description || ''
  if (record.project_id) {
    loadTasksForProject(record.project_id)
    loadBugsForProject(record.project_id)
  }
  allocationFormVisible.value = true
}

const handleAllocationSubmit = async () => {
  try {
    await allocationFormRef.value?.validate()

    if (!allocationFormData.date || !(allocationFormData.date as Dayjs).isValid()) {
      message.error('请选择有效的日期')
      return
    }

    if (allocationFormData.id) {
      const updateData: UpdateResourceAllocationRequest = {
        date: (allocationFormData.date as Dayjs).format('YYYY-MM-DD'),
        hours: allocationFormData.hours,
        task_id: allocationFormData.task_id,
        bug_id: allocationFormData.bug_id,
        project_id: allocationFormData.project_id,
        description: allocationFormData.description || undefined
      }
      await updateResourceAllocation(allocationFormData.id, updateData)
      message.success('更新成功')
    } else {
      const createData: CreateResourceAllocationRequest = {
        resource_id: allocationFormData.resource_id,
        date: (allocationFormData.date as Dayjs).format('YYYY-MM-DD'),
        hours: allocationFormData.hours,
        task_id: allocationFormData.task_id,
        bug_id: allocationFormData.bug_id,
        project_id: allocationFormData.project_id,
        description: allocationFormData.description || undefined
      }
      await createResourceAllocation(createData)
      message.success('创建成功')
    }

    allocationFormVisible.value = false
    loadAllocations()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error('操作失败: ' + (error.response?.data?.message || error.message))
  }
}

const handleAllocationCancel = () => {
  allocationFormVisible.value = false
  allocationFormRef.value?.resetFields()
}

const handleDeleteAllocation = async (id: number) => {
  try {
    await deleteResourceAllocation(id)
    message.success('删除成功')
    loadAllocations()
  } catch (error: any) {
    message.error('删除失败: ' + (error.response?.data?.message || error.message))
  }
}

const handleViewCalendar = () => {
  router.push('/resource/calendar')
}

const handleViewStatistics = () => {
  router.push('/resource/statistics')
}

const handleViewUtilization = () => {
  router.push('/resource/utilization')
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_resource_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadUsers()
  loadProjects()
  loadResources()
})
</script>

<style scoped>
.resource-management {
  min-height: 100vh;
}

.resource-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.resource-management :deep(.ant-layout) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  flex: 1;
  height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-inner {
  background: white;
  padding: 24px;
  border-radius: 4px;
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  height: 0;
}
</style>

