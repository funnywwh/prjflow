<template>
  <div class="test-case-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="测试单管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增测试单
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="测试单名称/描述"
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
                  <a-select-option value="pending">待测试</a-select-option>
                  <a-select-option value="running">测试中</a-select-option>
                  <a-select-option value="passed">通过</a-select-option>
                  <a-select-option value="failed">失败</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="类型">
                <a-select
                  v-model:value="searchForm.type"
                  placeholder="选择类型"
                  allow-clear
                  style="width: 120px"
                >
                  <a-select-option value="functional">功能测试</a-select-option>
                  <a-select-option value="performance">性能测试</a-select-option>
                  <a-select-option value="security">安全测试</a-select-option>
                  <a-select-option value="integration">集成测试</a-select-option>
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
            <a-col :span="5">
              <a-card :bordered="false">
                <a-statistic
                  title="总测试单数"
                  :value="statistics?.total || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="5">
              <a-card :bordered="false">
                <a-statistic
                  title="待测试"
                  :value="statistics?.pending || 0"
                  :value-style="{ color: '#faad14' }"
                />
              </a-card>
            </a-col>
            <a-col :span="5">
              <a-card :bordered="false">
                <a-statistic
                  title="测试中"
                  :value="statistics?.running || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="5">
              <a-card :bordered="false">
                <a-statistic
                  title="通过"
                  :value="statistics?.passed || 0"
                  :value-style="{ color: '#52c41a' }"
                />
              </a-card>
            </a-col>
            <a-col :span="4">
              <a-card :bordered="false">
                <a-statistic
                  title="失败"
                  :value="statistics?.failed || 0"
                  :value-style="{ color: '#ff4d4f' }"
                />
              </a-card>
            </a-col>
          </a-row>

          <!-- 覆盖率分析 -->
          <a-row :gutter="16" style="margin-bottom: 16px" v-if="statistics && statistics.total > 0">
            <a-col :span="12">
              <a-card title="测试通过率" :bordered="false">
                <a-statistic
                  :value="statistics.pass_rate || 0"
                  suffix="%"
                  :precision="2"
                  :value-style="{ color: statistics.pass_rate >= 80 ? '#52c41a' : statistics.pass_rate >= 60 ? '#faad14' : '#ff4d4f' }"
                />
                <div style="margin-top: 16px">
                  <a-progress
                    :percent="statistics.pass_rate || 0"
                    :stroke-color="statistics.pass_rate >= 80 ? '#52c41a' : statistics.pass_rate >= 60 ? '#faad14' : '#ff4d4f'"
                    :format="(percent: number) => `${percent?.toFixed(2)}%`"
                  />
                </div>
              </a-card>
            </a-col>
            <a-col :span="12">
              <a-card title="测试失败率" :bordered="false">
                <a-statistic
                  :value="statistics.fail_rate || 0"
                  suffix="%"
                  :precision="2"
                  :value-style="{ color: '#ff4d4f' }"
                />
                <div style="margin-top: 16px">
                  <a-progress
                    :percent="statistics.fail_rate || 0"
                    stroke-color="#ff4d4f"
                    :format="(percent: number) => `${percent?.toFixed(2)}%`"
                  />
                </div>
              </a-card>
            </a-col>
          </a-row>

          <!-- 按项目统计 -->
          <a-card title="按项目统计" :bordered="false" style="margin-bottom: 16px" v-if="statistics?.project_stats && statistics.project_stats.length > 0">
            <a-table
              :columns="projectStatsColumns"
              :data-source="statistics.project_stats"
              :pagination="false"
              :scroll="{ x: 'max-content' }"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'pass_rate'">
                  <span :style="{ color: record.pass_rate >= 80 ? '#52c41a' : record.pass_rate >= 60 ? '#faad14' : '#ff4d4f' }">
                    {{ record.pass_rate.toFixed(2) }}%
                  </span>
                </template>
              </template>
            </a-table>
          </a-card>

          <!-- 按类型统计 -->
          <a-card title="按测试类型统计" :bordered="false" style="margin-bottom: 16px" v-if="statistics?.type_stats && statistics.type_stats.length > 0">
            <a-table
              :columns="typeStatsColumns"
              :data-source="statistics.type_stats"
              :pagination="false"
              :scroll="{ x: 'max-content' }"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'type'">
                  {{ getTypeText(record.type) }}
                </template>
                <template v-else-if="column.key === 'pass_rate'">
                  <span :style="{ color: record.pass_rate >= 80 ? '#52c41a' : record.pass_rate >= 60 ? '#faad14' : '#ff4d4f' }">
                    {{ record.pass_rate.toFixed(2) }}%
                  </span>
                </template>
              </template>
            </a-table>
          </a-card>

          <a-card :bordered="false" class="table-card">
            <a-table
              :columns="columns"
              :data-source="testCases"
              :loading="loading"
              :pagination="pagination"
              :scroll="{ x: 'max-content', y: tableScrollHeight }"
              row-key="id"
              @change="handleTableChange"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <a-tag :color="getStatusColor(record.status)">
                    {{ getStatusText(record.status) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'type'">
                  <div v-if="record.types && record.types.length > 0" style="display: flex; flex-wrap: wrap; gap: 4px;">
                    <a-tag v-for="type in record.types" :key="type" style="margin: 0;">
                      {{ getTypeText(type) }}
                    </a-tag>
                  </div>
                  <span v-else>-</span>
                </template>
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'creator'">
                  {{ record.creator ? `${record.creator.username}${record.creator.nickname ? `(${record.creator.nickname})` : ''}` : '-' }}
                </template>
                <template v-else-if="column.key === 'bugs'">
                  <a-tag v-for="bug in record.bugs?.slice(0, 2)" :key="bug.id" style="margin-right: 4px" color="red">
                    {{ bug.title }}
                  </a-tag>
                  <a-tag v-if="record.bugs && record.bugs.length > 2" color="blue">
                    +{{ record.bugs.length - 2 }}
                  </a-tag>
                  <span v-if="!record.bugs || record.bugs.length === 0">-</span>
                </template>
                <template v-else-if="column.key === 'created_at'">
                  {{ formatDateTime(record.created_at) }}
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleStatusChange(record.id, e.key as string)">
                          <a-menu-item key="pending">待测试</a-menu-item>
                          <a-menu-item key="running">测试中</a-menu-item>
                          <a-menu-item key="passed">通过</a-menu-item>
                          <a-menu-item key="failed">失败</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个测试单吗？"
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

    <!-- 测试单编辑/创建模态框 -->
    <a-modal
      :mask-closable="true"
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
        <a-form-item label="测试单名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入测试单名称" />
        </a-form-item>
        <a-form-item label="测试描述" name="description">
          <a-textarea v-model:value="formData.description" placeholder="请输入测试描述" :rows="3" />
        </a-form-item>
        <a-form-item label="测试步骤" name="test_steps">
          <MarkdownEditor
            v-model="formData.test_steps"
            placeholder="请输入测试步骤（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="选择项目"
            show-search
            :filter-option="filterProjectOption"
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
        <a-form-item label="测试类型" name="types">
          <a-select
            v-model:value="formData.types"
            mode="multiple"
            placeholder="选择测试类型（可选，可多选）"
            allow-clear
            style="width: 100%"
          >
            <a-select-option value="functional">功能测试</a-select-option>
            <a-select-option value="performance">性能测试</a-select-option>
            <a-select-option value="security">安全测试</a-select-option>
            <a-select-option value="integration">集成测试</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status">
            <a-select-option value="pending">待测试</a-select-option>
            <a-select-option value="running">测试中</a-select-option>
            <a-select-option value="passed">通过</a-select-option>
            <a-select-option value="failed">失败</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关联Bug" name="bug_ids">
          <a-select
            v-model:value="formData.bug_ids"
            mode="multiple"
            placeholder="选择Bug（可选）"
            show-search
            :filter-option="filterBugOption"
            style="width: 100%"
          >
            <a-select-option
              v-for="bug in availableBugs"
              :key="bug.id"
              :value="bug.id"
            >
              {{ bug.title }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
// import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, DownOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getTestCases,
  createTestCase,
  updateTestCase,
  deleteTestCase,
  updateTestCaseStatus,
  getTestCaseStatistics,
  type TestCase,
  type CreateTestCaseRequest,
  type TestCaseStatistics
} from '@/api/testCase'
import { getProjects, type Project } from '@/api/project'
import { getBugs, type Bug } from '@/api/bug'

// const router = useRouter()
const loading = ref(false)
const testCases = ref<TestCase[]>([])
const projects = ref<Project[]>([])
const availableBugs = ref<Bug[]>([])
const statistics = ref<TestCaseStatistics | null>(null)

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  type: undefined as string | undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const projectStatsColumns = [
  { title: '项目名称', dataIndex: 'project_name', key: 'project_name' },
  { title: '总测试单数', dataIndex: 'total', key: 'total' },
  { title: '通过', dataIndex: 'passed', key: 'passed' },
  { title: '失败', dataIndex: 'failed', key: 'failed' },
  { title: '通过率', key: 'pass_rate', width: 120 }
]

const typeStatsColumns = [
  { title: '测试类型', key: 'type', width: 120 },
  { title: '总测试单数', dataIndex: 'total', key: 'total' },
  { title: '通过', dataIndex: 'passed', key: 'passed' },
  { title: '失败', dataIndex: 'failed', key: 'failed' },
  { title: '通过率', key: 'pass_rate', width: 120 }
]

// 计算表格滚动高度
const tableScrollHeight = computed(() => {
  return 'calc(100vh - 400px)'
})

const columns = [
  { title: '测试单名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '项目', key: 'project', width: 120 },
  { title: '类型', key: 'type', width: 200 },
  { title: '状态', key: 'status', width: 100 },
  { title: '关联Bug', key: 'bugs', width: 200 },
  { title: '创建人', key: 'creator', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增测试单')
const formRef = ref()
const formData = reactive<CreateTestCaseRequest & { id?: number }>({
  name: '',
  description: '',
  test_steps: '',
  types: [],
  status: 'pending',
  project_id: 0,
  bug_ids: []
})

const formRules = {
  name: [{ required: true, message: '请输入测试单名称', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载测试单列表
const loadTestCases = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (searchForm.status) params.status = searchForm.status
    if (searchForm.type) params.type = searchForm.type

    const res = await getTestCases(params)
    testCases.value = res.list
    pagination.total = res.total
    await loadStatistics()
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

// 加载统计信息
const loadStatistics = async () => {
  try {
    const params: any = {}
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.project_id) params.project_id = searchForm.project_id
    statistics.value = await getTestCaseStatistics(params)
  } catch (error: any) {
    console.error('加载统计信息失败:', error)
  }
}

// 加载项目列表
const loadProjects = async () => {
  try {
    const res = await getProjects({ page: 1, size: 1000 })
    projects.value = res.list
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载项目失败')
  }
}

// 加载可用Bug列表
const loadAvailableBugs = async () => {
  try {
    const res = await getBugs({ page: 1, size: 1000 })
    availableBugs.value = res.list
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载Bug失败')
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadTestCases()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_testcase_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_testcase_project_form', value || 0)
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.type = undefined
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_testcase_project_search', undefined)
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadTestCases()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增测试单'
  formData.id = undefined
  formData.name = ''
  formData.description = ''
  formData.test_steps = ''
  formData.types = []
  formData.status = 'pending'
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_testcase_project_form')
  formData.project_id = lastProjectId || 0
  formData.bug_ids = []
  loadAvailableBugs()
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: TestCase) => {
  modalTitle.value = '编辑测试单'
  formData.id = record.id
  formData.name = record.name
  formData.description = record.description || ''
  formData.test_steps = record.test_steps || ''
  formData.types = record.types || []
  formData.status = record.status
  formData.project_id = record.project_id
  formData.bug_ids = record.bugs?.map((b: any) => b.id) || []
  loadAvailableBugs()
  modalVisible.value = true
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    // 验证项目ID
    if (!formData.project_id || formData.project_id === 0) {
      message.error('请选择项目')
      return
    }
    
    const data: CreateTestCaseRequest = {
      name: formData.name,
      description: formData.description,
      test_steps: formData.test_steps,
      types: formData.types,
      status: formData.status,
      project_id: formData.project_id,
      bug_ids: formData.bug_ids
    }

    if (formData.id) {
      await updateTestCase(formData.id, data)
      message.success('更新成功')
    } else {
      await createTestCase(data)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadTestCases()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.response?.data?.message || '操作失败')
  }
}

// 取消
const handleCancel = () => {
  modalVisible.value = false
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteTestCase(id)
    message.success('删除成功')
    loadTestCases()
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 状态变更
const handleStatusChange = async (id: number, status: string) => {
  try {
    await updateTestCaseStatus(id, status)
    message.success('状态更新成功')
    loadTestCases()
  } catch (error: any) {
    message.error(error.response?.data?.message || '状态更新失败')
  }
}

// 状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'default',
    running: 'processing',
    passed: 'success',
    failed: 'error'
  }
  return colors[status] || 'default'
}

// 状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待测试',
    running: '测试中',
    passed: '通过',
    failed: '失败'
  }
  return texts[status] || status
}

// 类型文本
const getTypeText = (type?: string) => {
  if (!type) return '-'
  const texts: Record<string, string> = {
    functional: '功能测试',
    performance: '性能测试',
    security: '安全测试',
    integration: '集成测试'
  }
  return texts[type] || type
}

// 筛选函数
const filterProjectOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

const filterBugOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_testcase_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadProjects()
  loadTestCases()
})
</script>

<style scoped>
.test-case-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.test-case-management :deep(.ant-layout) {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content {
  padding: 24px;
  flex: 1;
  height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 0;
}

.table-card {
  margin-top: 16px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
}

.table-card :deep(.ant-table-wrapper) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-spin-nested-loading) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-spin-container) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.table-card :deep(.ant-table) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-card :deep(.ant-table-container) {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.content-inner {
  max-width: 100%;
  margin: 0 auto;
  width: 100%;
}

.table-card {
  margin-top: 16px;
}
</style>

