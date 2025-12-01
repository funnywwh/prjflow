<template>
  <div class="build-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="构建管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增构建
              </a-button>
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
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="构建号/分支/提交"
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
                  <a-select-option value="pending">待构建</a-select-option>
                  <a-select-option value="building">构建中</a-select-option>
                  <a-select-option value="success">成功</a-select-option>
                  <a-select-option value="failed">失败</a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <a-card :bordered="false" class="table-card">
            <a-table
              :columns="columns"
              :data-source="builds"
              :loading="loading"
              :scroll="{ x: 'max-content', y: tableScrollHeight }"
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
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'creator'">
                  {{ record.creator ? `${record.creator.username}${record.creator.nickname ? `(${record.creator.nickname})` : ''}` : '-' }}
                </template>
                <template v-else-if="column.key === 'version'">
                  <a-button v-if="record.version" type="link" @click="router.push(`/version/${record.version.id}`)">
                    {{ record.version.version_number }}
                  </a-button>
                  <span v-else>-</span>
                </template>
                <template v-else-if="column.key === 'build_time'">
                  {{ formatDateTime(record.build_time) }}
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
                          <a-menu-item key="pending">待构建</a-menu-item>
                          <a-menu-item key="building">构建中</a-menu-item>
                          <a-menu-item key="success">成功</a-menu-item>
                          <a-menu-item key="failed">失败</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个构建吗？"
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

    <!-- 构建编辑/创建模态框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="600"
      :mask-closable="false"
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
        <a-form-item label="构建号" name="build_number">
          <a-input v-model:value="formData.build_number" placeholder="请输入构建号" />
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
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status">
            <a-select-option value="pending">待构建</a-select-option>
            <a-select-option value="building">构建中</a-select-option>
            <a-select-option value="success">成功</a-select-option>
            <a-select-option value="failed">失败</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="分支" name="branch">
          <a-input v-model:value="formData.branch" placeholder="请输入分支" />
        </a-form-item>
        <a-form-item label="提交" name="commit">
          <a-input v-model:value="formData.commit" placeholder="请输入提交哈希" />
        </a-form-item>
        <a-form-item label="构建时间" name="build_time">
          <a-date-picker
            v-model:value="formData.build_time"
            show-time
            placeholder="选择构建时间"
            style="width: 100%"
            format="YYYY-MM-DD HH:mm:ss"
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
import { PlusOutlined, DownOutlined, UpOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import {
  getBuilds,
  createBuild,
  updateBuild,
  deleteBuild,
  updateBuildStatus,
  type Build,
  type CreateBuildRequest
} from '@/api/build'
import { getProjects, type Project } from '@/api/project'

const router = useRouter()
const loading = ref(false)
const searchFormVisible = ref(false) // 搜索栏显示/隐藏状态，默认折叠
const builds = ref<Build[]>([])
const projects = ref<Project[]>([])

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined
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
  { title: '构建号', dataIndex: 'build_number', key: 'build_number' },
  { title: '项目', key: 'project', width: 150 },
  { title: '状态', key: 'status', width: 100 },
  { title: '分支', dataIndex: 'branch', key: 'branch', width: 150 },
  { title: '提交', dataIndex: 'commit', key: 'commit', width: 120, ellipsis: true },
  { title: '构建时间', dataIndex: 'build_time', key: 'build_time', width: 180 },
  { title: '版本', key: 'version', width: 120 },
  { title: '创建人', key: 'creator', width: 150 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增构建')
const formRef = ref()
const formData = reactive<Omit<CreateBuildRequest, 'build_time'> & { id?: number; build_time?: Dayjs | undefined }>({
  build_number: '',
  status: 'pending',
  branch: '',
  commit: '',
  build_time: undefined,
  project_id: 0
})

const formRules = {
  build_number: [{ required: true, message: '请输入构建号', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载构建列表
const loadBuilds = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (searchForm.status) params.status = searchForm.status

    const res = await getBuilds(params)
    builds.value = res.list
    pagination.total = res.total
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
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

// 切换搜索栏显示/隐藏
const toggleSearchForm = () => {
  searchFormVisible.value = !searchFormVisible.value
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadBuilds()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_build_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_build_project_form', value || 0)
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_build_project_search', undefined)
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadBuilds()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增构建'
  formData.id = undefined
  formData.build_number = ''
  formData.status = 'pending'
  formData.branch = ''
  formData.commit = ''
  formData.build_time = undefined
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_build_project_form')
  formData.project_id = lastProjectId || 0
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: Build) => {
  modalTitle.value = '编辑构建'
  formData.id = record.id
  formData.build_number = record.build_number
  formData.status = record.status
  formData.branch = record.branch || ''
  formData.commit = record.commit || ''
  formData.project_id = record.project_id
  if (record.build_time) {
    formData.build_time = dayjs(record.build_time) as Dayjs | undefined
  } else {
    formData.build_time = undefined
  }
  modalVisible.value = true
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    const data: CreateBuildRequest = {
      build_number: formData.build_number,
      status: formData.status,
      branch: formData.branch,
      commit: formData.commit,
      project_id: formData.project_id,
      build_time: formData.build_time && formData.build_time.isValid() ? formData.build_time.format('YYYY-MM-DD HH:mm:ss') : undefined
    }

    if (formData.id) {
      await updateBuild(formData.id, data)
      message.success('更新成功')
    } else {
      await createBuild(data)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadBuilds()
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
    await deleteBuild(id)
    message.success('删除成功')
    loadBuilds()
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 状态变更
const handleStatusChange = async (id: number, status: string) => {
  try {
    await updateBuildStatus(id, status)
    message.success('状态更新成功')
    loadBuilds()
  } catch (error: any) {
    message.error(error.response?.data?.message || '状态更新失败')
  }
}

// 状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    pending: 'default',
    building: 'processing',
    success: 'success',
    failed: 'error'
  }
  return colors[status] || 'default'
}

// 状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待构建',
    building: '构建中',
    success: '成功',
    failed: '失败'
  }
  return texts[status] || status
}

// 项目筛选
const filterProjectOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_build_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadProjects()
  loadBuilds()
})
</script>

<style scoped>
.build-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.build-management :deep(.ant-layout) {
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

