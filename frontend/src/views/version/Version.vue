<template>
  <div class="version-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="版本管理">
            <template #extra>
              <a-button type="primary" @click="handleCreate">
                <template #icon><PlusOutlined /></template>
                新增版本
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
                  placeholder="版本号/发布说明"
                  allow-clear
                  style="width: 200px"
                />
              </a-form-item>
              <a-form-item label="项目">
                <a-select
                  v-model:value="searchForm.project_id"
                  placeholder="选择项目"
                  allow-clear
                  show-search
                  :filter-option="filterProjectOption"
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
                  <a-select-option value="draft">草稿</a-select-option>
                  <a-select-option value="released">已发布</a-select-option>
                  <a-select-option value="archived">已归档</a-select-option>
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
              :data-source="versions"
              :loading="loading"
              :scroll="{ x: 'max-content', y: tableScrollHeight }"
              :pagination="pagination"
              row-key="id"
              @change="handleTableChange"
              :custom-row="(record: Version) => ({
                onClick: () => handleView(record),
                class: 'table-row-clickable'
              })"
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
                <template v-else-if="column.key === 'requirements'">
                  <a-tag v-for="req in record.requirements?.slice(0, 2)" :key="req.id" style="margin-right: 4px">
                    {{ req.title }}
                  </a-tag>
                  <a-tag v-if="record.requirements && record.requirements.length > 2" color="blue">
                    +{{ record.requirements.length - 2 }}
                  </a-tag>
                  <span v-if="!record.requirements || record.requirements.length === 0">-</span>
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
                <template v-else-if="column.key === 'release_date'">
                  {{ formatDateTime(record.release_date) }}
                </template>
                <template v-else-if="column.key === 'created_at'">
                  {{ formatDateTime(record.created_at) }}
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space @click.stop>
                    <a-button type="link" size="small" @click.stop="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-button v-if="record.status === 'wait'" type="link" size="small" @click.stop="handleRelease(record.id)">
                      发布
                    </a-button>
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleStatusChange(record, e.key as string)">
                          <a-menu-item key="draft">草稿</a-menu-item>
                          <a-menu-item key="released">已发布</a-menu-item>
                          <a-menu-item key="archived">已归档</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个版本吗？"
                      @confirm="handleDelete(record.id)"
                    >
                      <a-button type="link" size="small" danger @click.stop>删除</a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </template>
            </a-table>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 版本编辑/创建模态框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="800"
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
        <a-form-item label="版本号" name="version_number">
          <a-input v-model:value="formData.version_number" placeholder="请输入版本号" />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="选择项目"
            show-search
            :filter-option="filterProjectOption"
            :disabled="!!formData.id"
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
            <a-select-option value="wait">未开始</a-select-option>
            <a-select-option value="normal">已发布</a-select-option>
            <a-select-option value="fail">发布失败</a-select-option>
            <a-select-option value="terminate">停止维护</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="发布日期" name="release_date">
          <a-date-picker
            v-model:value="formData.release_date"
            placeholder="选择发布日期"
            style="width: 100%"
            format="YYYY-MM-DD"
          />
        </a-form-item>
        <a-form-item label="发布说明" name="release_notes">
          <MarkdownEditor
            v-model="formData.release_notes"
            placeholder="请输入发布说明（支持Markdown）"
            :rows="8"
          />
        </a-form-item>
        <a-form-item label="关联需求" name="requirement_ids">
          <a-select
            v-model:value="formData.requirement_ids"
            mode="multiple"
            placeholder="选择需求（可选）"
            show-search
            :filter-option="filterRequirementOption"
            style="width: 100%"
          >
            <a-select-option
              v-for="requirement in availableRequirements"
              :key="requirement.id"
              :value="requirement.id"
            >
              {{ requirement.title }}
            </a-select-option>
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
        <a-form-item label="附件">
          <AttachmentUpload
            v-if="formData.project_id && formData.project_id > 0"
            :project-id="formData.project_id"
            :model-value="formData.attachment_ids"
            :existing-attachments="versionAttachments"
            @update:modelValue="(value) => { formData.attachment_ids = value }"
            @attachment-deleted="handleAttachmentDeleted"
          />
          <span v-else style="color: #999;">请先选择项目后再上传附件</span>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 版本详情弹窗 -->
    <a-modal
      v-model:open="detailModalVisible"
      :width="1200"
      :mask-closable="true"
      :footer="null"
      @cancel="handleDetailCancel"
    >
      <template #title>
        <div style="width: 100%;">
          <div style="text-align: center;">版本详情</div>
          <div v-if="detailVersion" style="font-size: 14px; color: #666; margin-top: 4px; text-align: left;">
            {{ detailVersion.version_number }}
          </div>
        </div>
      </template>
      <a-spin :spinning="detailLoading">
        <div v-if="detailVersion" style="max-height: 70vh; overflow-y: auto">
          <!-- 操作按钮 -->
          <div style="margin-bottom: 16px; text-align: right">
            <a-space>
              <a-button @click="handleDetailEdit">编辑</a-button>
              <a-button v-if="detailVersion.status === 'wait'" type="primary" @click="handleDetailRelease">
                发布
              </a-button>
              <a-dropdown>
                <a-button>
                  状态 <DownOutlined />
                </a-button>
                <template #overlay>
                  <a-menu @click="(e: any) => handleDetailStatusChange(e.key as string)">
                    <a-menu-item key="draft">草稿</a-menu-item>
                    <a-menu-item key="released">已发布</a-menu-item>
                    <a-menu-item key="archived">已归档</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
              <a-popconfirm
                title="确定要删除这个版本吗？"
                @confirm="handleDetailDelete"
              >
                <a-button danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </div>

          <!-- 基本信息 -->
          <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
            <a-descriptions :column="2" bordered>
              <a-descriptions-item label="版本号">{{ detailVersion.version_number }}</a-descriptions-item>
              <a-descriptions-item label="状态">
                <a-tag :color="getStatusColor(detailVersion.status || '')">
                  {{ getStatusText(detailVersion.status || '') }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="项目">
                <a v-if="detailVersion.project" @click="router.push(`/project/${detailVersion.project.id}`)" style="cursor: pointer">
                  {{ detailVersion.project.name }}
                </a>
                <span v-else>-</span>
              </a-descriptions-item>
              <a-descriptions-item label="发布日期">
                {{ formatDateTime(detailVersion.release_date) }}
              </a-descriptions-item>
              <a-descriptions-item label="创建时间">
                {{ formatDateTime(detailVersion.created_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="更新时间">
                {{ formatDateTime(detailVersion.updated_at) }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <!-- 发布说明 -->
          <a-card title="发布说明" :bordered="false" style="margin-bottom: 16px">
            <div v-if="detailVersion.release_notes" class="markdown-content">
              <MarkdownEditor
                :model-value="detailVersion.release_notes"
                :readonly="true"
              />
            </div>
            <a-empty v-else description="暂无发布说明" />
          </a-card>

          <!-- 关联需求 -->
          <a-card title="关联需求" :bordered="false" style="margin-bottom: 16px">
            <a-list
              v-if="detailVersion.requirements && detailVersion.requirements.length > 0"
              :data-source="detailVersion.requirements"
              :pagination="false"
            >
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta>
                    <template #title>
                      <a-button type="link" @click="router.push(`/requirement/${item.id}`)">
                        {{ item.title }}
                      </a-button>
                    </template>
                    <template #description>
                      <a-space>
                        <a-tag :color="getRequirementStatusColor(item.status)">
                          {{ getRequirementStatusText(item.status) }}
                        </a-tag>
                        <a-tag :color="getPriorityColor(item.priority)">
                          {{ getPriorityText(item.priority) }}
                        </a-tag>
                      </a-space>
                    </template>
                  </a-list-item-meta>
                </a-list-item>
              </template>
            </a-list>
            <a-empty v-else description="暂无关联需求" />
          </a-card>

          <!-- 关联Bug -->
          <a-card title="关联Bug" :bordered="false">
            <a-list
              v-if="detailVersion.bugs && detailVersion.bugs.length > 0"
              :data-source="detailVersion.bugs"
              :pagination="false"
            >
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta>
                    <template #title>
                      <a-button type="link" @click="router.push(`/bug/${item.id}`)">
                        {{ item.title }}
                      </a-button>
                    </template>
                    <template #description>
                      <a-space>
                        <a-tag :color="getBugStatusColor(item.status)">
                          {{ getBugStatusText(item.status) }}
                        </a-tag>
                        <a-tag :color="getPriorityColor(item.priority)">
                          {{ getPriorityText(item.priority) }}
                        </a-tag>
                        <a-tag :color="getSeverityColor(item.severity)">
                          {{ getSeverityText(item.severity) }}
                        </a-tag>
                      </a-space>
                    </template>
                  </a-list-item-meta>
                </a-list-item>
              </template>
            </a-list>
            <a-empty v-else description="暂无关联Bug" />
          </a-card>
        </div>
      </a-spin>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed, watch } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
import { message } from 'ant-design-vue'
import { PlusOutlined, DownOutlined, UpOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { useRouter } from 'vue-router'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import { getAttachments, type Attachment } from '@/api/attachment'
import {
  getVersions,
  getVersion,
  createVersion,
  updateVersion,
  deleteVersion,
  releaseVersion,
  type Version,
  type CreateVersionRequest,
  type UpdateVersionRequest
} from '@/api/version'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, type Requirement } from '@/api/requirement'
import { getBugs, type Bug } from '@/api/bug'

const router = useRouter()
const loading = ref(false)
const searchFormVisible = ref(false) // 搜索栏显示/隐藏状态，默认折叠
const versions = ref<Version[]>([])
const projects = ref<Project[]>([])
const availableRequirements = ref<Requirement[]>([])
const availableBugs = ref<Bug[]>([])

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
  { title: '版本号', dataIndex: 'version_number', key: 'version_number' },
  { title: '项目', key: 'project', width: 150 },
  { title: '状态', key: 'status', width: 100 },
  { title: '关联需求', key: 'requirements', width: 200 },
  { title: '关联Bug', key: 'bugs', width: 200 },
  { title: '发布日期', dataIndex: 'release_date', key: 'release_date', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 300, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增版本')
const formRef = ref()

// 详情弹窗相关
const detailModalVisible = ref(false)
const detailLoading = ref(false)
const detailVersion = ref<Version | null>(null)
const shouldKeepDetailOpen = ref(false)
const formData = reactive<Omit<CreateVersionRequest, 'release_date'> & { id?: number; release_date?: Dayjs | undefined; attachment_ids?: number[] }>({
  version_number: '',
  release_notes: '',
  status: 'wait',
  project_id: 0,
  release_date: undefined,
  requirement_ids: [],
  bug_ids: [],
  attachment_ids: []
})
const versionAttachments = ref<Attachment[]>([])

// 处理附件删除事件
const handleAttachmentDeleted = (attachmentId: number) => {
  // 从versionAttachments中移除已删除的附件
  versionAttachments.value = versionAttachments.value.filter(a => a.id !== attachmentId)
}

const formRules = {
  version_number: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载版本列表
const loadVersions = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.project_id) params.project_id = searchForm.project_id
    if (searchForm.status) params.status = searchForm.status

    const res = await getVersions(params)
    versions.value = res.list
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

// 加载可用需求和Bug
const loadAvailableRequirementsAndBugs = async () => {
  try {
    const [reqRes, bugRes] = await Promise.all([
      getRequirements({ page: 1, size: 1000 }),
      getBugs({ page: 1, size: 1000 })
    ])
    availableRequirements.value = reqRes.list
    availableBugs.value = bugRes.list
  } catch (error: any) {
    message.error(error.response?.data?.message || '加载失败')
  }
}

// 切换搜索栏显示/隐藏
const toggleSearchForm = () => {
  searchFormVisible.value = !searchFormVisible.value
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadVersions()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_version_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_version_project_form', value || 0)
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_version_project_search', undefined)
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadVersions()
}

// 创建
const handleCreate = () => {
  modalTitle.value = '新增版本'
  formData.id = undefined
  formData.version_number = ''
  formData.release_notes = ''
  formData.status = 'wait'
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_version_project_form')
  formData.project_id = lastProjectId || 0
  formData.release_date = undefined
  formData.requirement_ids = []
  formData.bug_ids = []
  formData.attachment_ids = []
  versionAttachments.value = []
  loadAvailableRequirementsAndBugs()
  modalVisible.value = true
}

// 编辑
const handleEdit = async (record: Version) => {
  modalTitle.value = '编辑版本'
  formData.id = record.id
  formData.version_number = record.version_number
  formData.release_notes = record.release_notes || ''
  // 直接使用后端状态值，不进行转换
  formData.status = record.status || 'wait'
  formData.project_id = record.project_id
  if (record.release_date) {
    formData.release_date = dayjs(record.release_date)
  } else {
    formData.release_date = undefined
  }
  formData.requirement_ids = record.requirements?.map((r: any) => r.id) || []
  formData.bug_ids = record.bugs?.map((b: any) => b.id) || []
  
  // 加载版本附件
  try {
    if (record.attachments && record.attachments.length > 0) {
      versionAttachments.value = record.attachments
      formData.attachment_ids = record.attachments.map((a: any) => a.id)
    } else {
      versionAttachments.value = await getAttachments({ version_id: record.id })
      formData.attachment_ids = versionAttachments.value.map(a => a.id)
    }
  } catch (error: any) {
    console.error('加载附件失败:', error)
    versionAttachments.value = []
    formData.attachment_ids = []
  }
  
  await loadAvailableRequirementsAndBugs()
  modalVisible.value = true
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (formData.id) {
      // 更新版本
      const updateData: UpdateVersionRequest = {
        version_number: formData.version_number,
        release_notes: formData.release_notes,
        status: formData.status,
        release_date: formData.release_date && formData.release_date.isValid() ? formData.release_date.format('YYYY-MM-DD') : undefined,
        requirement_ids: formData.requirement_ids || [],
        bug_ids: formData.bug_ids || []
      }
      
      // 始终发送 attachment_ids，如果为 undefined 或 null，发送空数组
      const attachmentIdsValue = formData.attachment_ids
      if (attachmentIdsValue === undefined || attachmentIdsValue === null) {
        updateData.attachment_ids = []
      } else {
        updateData.attachment_ids = Array.isArray(attachmentIdsValue) ? attachmentIdsValue : []
      }
      
      await updateVersion(formData.id, updateData)
      message.success('更新成功')
    } else {
      // 创建版本
      const createData: CreateVersionRequest = {
        version_number: formData.version_number,
        release_notes: formData.release_notes,
        status: formData.status,
        project_id: formData.project_id,
        release_date: formData.release_date && formData.release_date.isValid() ? formData.release_date.format('YYYY-MM-DD') : undefined,
        requirement_ids: formData.requirement_ids || [],
        bug_ids: formData.bug_ids || []
      }
      
      // 始终发送 attachment_ids，如果为 undefined 或 null，发送空数组
      const attachmentIdsValue = formData.attachment_ids
      if (attachmentIdsValue === undefined || attachmentIdsValue === null) {
        createData.attachment_ids = []
      } else {
        createData.attachment_ids = Array.isArray(attachmentIdsValue) ? attachmentIdsValue : []
      }
      
      await createVersion(createData)
      message.success('创建成功')
    }
    modalVisible.value = false
    loadVersions()
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
    await deleteVersion(id)
    message.success('删除成功')
    loadVersions()
  } catch (error: any) {
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 状态变更（弹出编辑界面）
const handleStatusChange = async (record: Version, status: string) => {
  // 打开编辑对话框，并设置新状态
  await handleEdit(record)
  // 使用 nextTick 确保表单已加载后再设置状态
  await nextTick()
  formData.status = status as any
}

// 发布版本
const handleRelease = async (id: number) => {
  try {
    await releaseVersion(id)
    message.success('发布成功')
    loadVersions()
  } catch (error: any) {
    message.error(error.response?.data?.message || '发布失败')
  }
}

// 状态颜色
const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    wait: 'orange',
    normal: 'green',
    fail: 'red',
    terminate: 'default'
  }
  return colors[status] || 'default'
}

// 状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    wait: '未开始',
    normal: '已发布',
    fail: '发布失败',
    terminate: '停止维护'
  }
  return texts[status] || status
}

// 筛选函数
const filterProjectOption = (input: string, option: any) => {
  const project = projects.value.find(p => p.id === option.value)
  if (!project) return false
  const searchText = input.toLowerCase()
  return (
    project.name.toLowerCase().includes(searchText) ||
    (project.code && project.code.toLowerCase().includes(searchText))
  )
}

const filterRequirementOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

const filterBugOption = (input: string, option: any) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 查看详情
const handleView = async (record: Version) => {
  detailModalVisible.value = true
  await loadVersionDetail(record.id)
}

// 加载版本详情
const loadVersionDetail = async (versionId: number) => {
  detailLoading.value = true
  try {
    detailVersion.value = await getVersion(versionId)
  } catch (error: any) {
    message.error(error.message || '加载版本详情失败')
    detailModalVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

// 详情弹窗取消
const handleDetailCancel = () => {
  detailVersion.value = null
}

// 详情页编辑
const handleDetailEdit = async () => {
  if (!detailVersion.value) return
  shouldKeepDetailOpen.value = true
  detailModalVisible.value = false
  await nextTick()
  handleEdit(detailVersion.value)
}

// 详情页发布
const handleDetailRelease = async () => {
  if (!detailVersion.value) return
  try {
    await releaseVersion(detailVersion.value.id)
    message.success('发布成功')
    await loadVersionDetail(detailVersion.value.id)
    loadVersions()
  } catch (error: any) {
    message.error(error.message || '发布失败')
  }
}

// 详情页状态变更
const handleDetailStatusChange = async (status: string) => {
  if (!detailVersion.value) return
  try {
    await handleStatusChange(detailVersion.value, status)
    await loadVersionDetail(detailVersion.value.id)
    loadVersions()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 详情页删除
const handleDetailDelete = async () => {
  if (!detailVersion.value) return
  try {
    await deleteVersion(detailVersion.value.id)
    message.success('删除成功')
    detailModalVisible.value = false
    loadVersions()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 需求状态颜色
const getRequirementStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    draft: 'default',
    reviewing: 'blue',
    active: 'green',
    changing: 'orange',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 需求状态文本
const getRequirementStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    reviewing: '评审中',
    active: '激活',
    changing: '变更中',
    closed: '已关闭'
  }
  return texts[status] || status
}

// Bug状态颜色
const getBugStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    active: 'orange',
    resolved: 'green',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// Bug状态文本
const getBugStatusText = (status: string) => {
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return texts[status] || status
}

// 严重程度颜色
const getSeverityColor = (severity: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    critical: 'red'
  }
  return colors[severity] || 'default'
}

// 严重程度文本
const getSeverityText = (severity: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '严重'
  }
  return texts[severity] || severity
}

// 优先级颜色
const getPriorityColor = (priority: string) => {
  const colors: Record<string, string> = {
    low: 'default',
    medium: 'blue',
    high: 'orange',
    urgent: 'red'
  }
  return colors[priority] || 'default'
}

// 优先级文本
const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '中',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

// 监听编辑模态框关闭，重新打开详情弹窗
watch(modalVisible, (visible, prevVisible) => {
  if (prevVisible && !visible && shouldKeepDetailOpen.value && detailVersion.value) {
    shouldKeepDetailOpen.value = false
    nextTick(() => {
      detailModalVisible.value = true
      loadVersionDetail(detailVersion.value!.id)
      loadVersions()
    })
  }
})

onMounted(() => {
  // 从 localStorage 恢复最后选择的搜索项目
  const lastSearchProjectId = getLastSelected<number>('last_selected_version_project_search')
  if (lastSearchProjectId) {
    searchForm.project_id = lastSearchProjectId
  }
  loadProjects()
  loadVersions()
})
</script>

<style scoped>
.version-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.version-management :deep(.ant-layout) {
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

/* 详情弹窗样式 */
.markdown-content {
  min-height: 200px;
}

/* 表格行可点击样式 */
.table-card :deep(.ant-table-tbody > tr.table-row-clickable) {
  cursor: pointer;
}

.table-card :deep(.ant-table-tbody > tr.table-row-clickable:hover) {
  background-color: #f5f5f5;
}
</style>

