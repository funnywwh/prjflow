<template>
  <div class="test-report-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="测试报告管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增测试报告
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="报告标题/内容"
                  allow-clear
                  style="width: 200px"
                />
              </a-form-item>
              <a-form-item label="结果">
                <a-select
                  v-model:value="searchForm.result"
                  placeholder="选择结果"
                  allow-clear
                  style="width: 120px"
                >
                  <a-select-option value="passed">通过</a-select-option>
                  <a-select-option value="failed">失败</a-select-option>
                  <a-select-option value="blocked">阻塞</a-select-option>
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
                  title="总报告数"
                  :value="statistics?.total || 0"
                  :value-style="{ color: '#1890ff' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="通过"
                  :value="statistics?.passed || 0"
                  :value-style="{ color: '#52c41a' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="失败"
                  :value="statistics?.failed || 0"
                  :value-style="{ color: '#ff4d4f' }"
                />
              </a-card>
            </a-col>
            <a-col :span="6">
              <a-card :bordered="false">
                <a-statistic
                  title="阻塞"
                  :value="statistics?.blocked || 0"
                  :value-style="{ color: '#faad14' }"
                />
              </a-card>
            </a-col>
          </a-row>

          <a-card :bordered="false" class="table-card">
            <a-table
              :columns="columns"
              :data-source="testReports"
              :loading="loading"
              :pagination="pagination"
              :scroll="{ y: tableScrollHeight }"
              row-key="id"
              @change="handleTableChange"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'result'">
                  <a-tag :color="getResultColor(record.result)">
                    {{ getResultText(record.result) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'creator'">
                  {{ record.creator ? `${record.creator.username}${record.creator.nickname ? `(${record.creator.nickname})` : ''}` : '-' }}
                </template>
                <template v-else-if="column.key === 'test_cases'">
                  <a-tag v-for="tc in record.test_cases?.slice(0, 2)" :key="tc.id" style="margin-right: 4px">
                    {{ tc.name }}
                  </a-tag>
                  <a-tag v-if="record.test_cases && record.test_cases.length > 2" color="blue">
                    +{{ record.test_cases.length - 2 }}
                  </a-tag>
                  <span v-if="!record.test_cases || record.test_cases.length === 0">-</span>
                </template>
                <template v-else-if="column.key === 'created_at'">
                  {{ formatDateTime(record.created_at) }}
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-popconfirm
                      title="确定要删除这个测试报告吗？"
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

    <!-- 测试报告编辑/创建模态框 -->
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
        <a-form-item label="报告标题" name="title">
          <a-input v-model:value="formData.title" placeholder="请输入报告标题" />
        </a-form-item>
        <a-form-item label="测试摘要" name="summary">
          <a-textarea v-model:value="formData.summary" placeholder="请输入测试摘要" :rows="3" />
        </a-form-item>
        <a-form-item label="报告内容" name="content">
          <MarkdownEditor
            v-model="formData.content"
            placeholder="请输入报告内容（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="测试结果" name="result">
          <a-select v-model:value="formData.result" placeholder="选择测试结果（可选）" allow-clear>
            <a-select-option value="passed">通过</a-select-option>
            <a-select-option value="failed">失败</a-select-option>
            <a-select-option value="blocked">阻塞</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关联测试单" name="test_case_ids">
          <a-select
            v-model:value="formData.test_case_ids"
            mode="multiple"
            placeholder="选择测试单（可选）"
            show-search
            :filter-option="filterTestCaseOption"
            style="width: 100%"
          >
            <a-select-option
              v-for="testCase in availableTestCases"
              :key="testCase.id"
              :value="testCase.id"
            >
              {{ testCase.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
// import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import {
  getTestReports,
  createTestReport,
  updateTestReport,
  deleteTestReport,
  getTestReportStatistics,
  type TestReport,
  type CreateTestReportRequest,
  type TestReportStatistics
} from '@/api/testReport'
import { getTestCases, type TestCase } from '@/api/testCase'

// const router = useRouter()
const loading = ref(false)
const testReports = ref<TestReport[]>([])
const availableTestCases = ref<TestCase[]>([])
const statistics = ref<TestReportStatistics | null>(null)

const searchForm = reactive({
  keyword: '',
  result: undefined as string | undefined
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
  { title: '报告标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '测试结果', key: 'result', width: 100 },
  { title: '关联测试单', key: 'test_cases', width: 200 },
  { title: '创建人', key: 'creator', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增测试报告')
const formRef = ref()
const formData = reactive<CreateTestReportRequest & { id?: number }>({
  title: '',
  content: '',
  result: undefined,
  summary: '',
  test_case_ids: []
})

const formRules = {
  title: [{ required: true, message: '请输入报告标题', trigger: 'blur' }]
}

// 加载测试报告列表
const loadTestReports = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.result) params.result = searchForm.result

    const res = await getTestReports(params)
    testReports.value = res.list
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
    statistics.value = await getTestReportStatistics(params)
  } catch (error: any) {
    console.error('加载统计信息失败:', error)
  }
}

// 加载可用测试单列表
const loadAvailableTestCases = async () => {
  try {
    const res = await getTestCases({ page: 1, size: 1000 })
    availableTestCases.value = res.list
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载测试单失败')
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadTestReports()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.result = undefined
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadTestReports()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增测试报告'
  formData.id = undefined
  formData.title = ''
  formData.content = ''
  formData.result = undefined
  formData.summary = ''
  formData.test_case_ids = []
  loadAvailableTestCases()
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: TestReport) => {
  modalTitle.value = '编辑测试报告'
  formData.id = record.id
  formData.title = record.title
  formData.content = record.content || ''
  formData.result = record.result as any
  formData.summary = record.summary || ''
  formData.test_case_ids = record.test_cases?.map((tc: any) => tc.id) || []
  loadAvailableTestCases()
  modalVisible.value = true
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const data: CreateTestReportRequest = {
      title: formData.title,
      content: formData.content,
      result: formData.result,
      summary: formData.summary,
      test_case_ids: formData.test_case_ids
    }

    if (formData.id) {
      await updateTestReport(formData.id, data)
      message.success('更新成功')
    } else {
      await createTestReport(data)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadTestReports()
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
    await deleteTestReport(id)
    message.success('删除成功')
    loadTestReports()
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 结果颜色
const getResultColor = (result?: string) => {
  if (!result) return 'default'
  const colors: Record<string, string> = {
    passed: 'success',
    failed: 'error',
    blocked: 'warning'
  }
  return colors[result] || 'default'
}

// 结果文本
const getResultText = (result?: string) => {
  if (!result) return '-'
  const texts: Record<string, string> = {
    passed: '通过',
    failed: '失败',
    blocked: '阻塞'
  }
  return texts[result] || result
}

// 筛选函数
const filterTestCaseOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

onMounted(() => {
  loadTestReports()
})
</script>

<style scoped>
.test-report-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.test-report-management :deep(.ant-layout) {
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
  width: 100%;
  margin: 0 auto;
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
  width: 100%;
  margin: 0 auto;
}

.table-card {
  margin-top: 16px;
}
</style>

