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
              <a-form-item label="产品">
                <a-select
                  v-model:value="searchForm.product_id"
                  placeholder="选择产品"
                  allow-clear
                  style="width: 150px"
                >
                  <a-select-option
                    v-for="product in products"
                    :key="product.id"
                    :value="product.id"
                  >
                    {{ product.name }}
                  </a-select-option>
                </a-select>
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
                <template v-else-if="column.key === 'product'">
                  {{ record.product?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'assignee'">
                  {{ record.assignee ? `${record.assignee.username}${record.assignee.nickname ? `(${record.assignee.nickname})` : ''}` : '-' }}
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
            v-model="formData.description"
            placeholder="请输入需求描述（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="产品" name="product_id">
          <a-select
            v-model:value="formData.product_id"
            placeholder="选择产品（可选）"
            allow-clear
          >
            <a-select-option
              v-for="product in products"
              :key="product.id"
              :value="product.id"
            >
              {{ product.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
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
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
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
import { getProducts, type Product } from '@/api/product'
import { getProjects, type Project } from '@/api/project'
import { getUsers, type User } from '@/api/user'

const loading = ref(false)
const requirements = ref<Requirement[]>([])
const products = ref<Product[]>([])
const projects = ref<Project[]>([])
const users = ref<User[]>([])
const statistics = ref<RequirementStatistics | null>(null)

const searchForm = reactive({
  keyword: '',
  product_id: undefined as number | undefined,
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  priority: undefined as string | undefined
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
  { title: '产品', key: 'product', width: 120 },
  { title: '项目', key: 'project', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '优先级', key: 'priority', width: 100 },
  { title: '负责人', key: 'assignee', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增需求')
const formRef = ref()
const formData = reactive<CreateRequirementRequest & { id?: number }>({
  title: '',
  description: '',
  status: 'pending',
  priority: 'medium',
  product_id: undefined,
  project_id: undefined,
  assignee_id: undefined
})

const formRules = {
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }]
}

// 加载需求列表
const loadRequirements = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    if (searchForm.keyword) {
      params.keyword = searchForm.keyword
    }
    if (searchForm.product_id) {
      params.product_id = searchForm.product_id
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
    if (searchForm.product_id) {
      params.product_id = searchForm.product_id
    }
    if (searchForm.project_id) {
      params.project_id = searchForm.project_id
    }
    statistics.value = await getRequirementStatistics(params)
  } catch (error: any) {
    console.error('加载统计信息失败:', error)
  }
}

// 加载产品列表
const loadProducts = async () => {
  try {
    const response = await getProducts()
    products.value = response.list || []
  } catch (error: any) {
    console.error('加载产品列表失败:', error)
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

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.product_id = undefined
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.priority = undefined
  pagination.current = 1
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
  formData.product_id = undefined
  formData.project_id = undefined
  formData.assignee_id = undefined
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: Requirement) => {
  modalTitle.value = '编辑需求'
  formData.id = record.id
  formData.title = record.title
  formData.description = record.description || ''
  formData.status = record.status
  formData.priority = record.priority
  formData.product_id = record.product_id
  formData.project_id = record.project_id
  formData.assignee_id = record.assignee_id
  modalVisible.value = true
}

// 查看详情
const handleView = (record: Requirement) => {
  // TODO: 实现详情页面
  message.info('详情功能待实现')
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const data: CreateRequirementRequest = {
      title: formData.title,
      description: formData.description,
      status: formData.status,
      priority: formData.priority,
      product_id: formData.product_id,
      project_id: formData.project_id,
      assignee_id: formData.assignee_id
    }
    if (formData.id) {
      await updateRequirement(formData.id, data)
      message.success('更新成功')
    } else {
      await createRequirement(data)
      message.success('创建成功')
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
  loadRequirements()
  loadProducts()
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
  max-width: 1400px;
  margin: 0 auto;
}
</style>

