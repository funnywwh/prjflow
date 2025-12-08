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
              <div style="display: flex; align-items: center; gap: 8px">
                <a-space>
                  <a-button
                    :disabled="!prevBugId || bugListLoading"
                    @click="handleNavigateToPrev"
                  >
                    ← 上一个
                  </a-button>
                  <a-button
                    :disabled="!nextBugId || bugListLoading"
                    @click="handleNavigateToNext"
                  >
                    下一个 →
                  </a-button>
                </a-space>
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
                  <a-button
                    @click="handleClose"
                    :disabled="bug?.status !== 'resolved'"
                  >
                    关闭
                  </a-button>
                  <a-button
                    v-if="bug?.status === 'active'"
                    @click="handleConvertToRequirement"
                  >
                    转需求
                  </a-button>
                  <a-popconfirm
                    title="确定要删除这个Bug吗？"
                    @confirm="handleDelete"
                  >
                    <a-button danger>删除</a-button>
                  </a-popconfirm>
                </a-space>
              </div>
            </template>
          </a-page-header>

          <BugDetailContent
            :bug="bug"
            :loading="loading"
            @refresh="handleRefresh"
            @requirement-click="handleRequirementClick"
          />
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
          <ProjectMemberSelect
            v-model="assignFormData.assignee_ids"
            :project-id="bug?.project_id"
            :multiple="true"
            placeholder="选择指派给"
            :show-role="true"
          />
        </a-form-item>
        <a-form-item label="备注" name="comment">
          <a-textarea
            v-model:value="assignFormData.comment"
            placeholder="请输入备注（可选）"
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select
            v-model:value="assignFormData.status"
            placeholder="选择状态（可选）"
            allow-clear
          >
            <a-select-option value="active">激活</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
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
      :z-index="2000"
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
            :getPopupContainer="getPopupContainer"
            :dropdownStyle="{ zIndex: 2100 }"
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
              :getPopupContainer="getPopupContainer"
              :dropdownStyle="{ zIndex: 2100 }"
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

    <!-- Bug编辑模态框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑Bug"
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
        <a-form-item label="Bug标题" name="title">
          <a-input v-model:value="editFormData.title" placeholder="请输入Bug标题" />
        </a-form-item>
        <a-form-item label="Bug描述" name="description">
          <MarkdownEditor
            ref="editDescriptionEditorRef"
            v-model="editFormData.description"
            placeholder="请输入Bug描述（支持Markdown）"
            :rows="8"
            :project-id="editFormData.project_id || 0"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="editFormData.project_id"
            placeholder="选择项目"
            show-search
            :filter-option="filterProjectOption"
            @change="handleEditFormProjectChange"
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
            v-model:value="editFormData.requirement_id"
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
        <a-form-item label="功能模块" name="module_id">
          <a-select
            v-model:value="editFormData.module_id"
            placeholder="选择功能模块（可选）"
            allow-clear
            show-search
            :filter-option="filterModuleOption"
            :loading="moduleLoading"
            @focus="loadModulesForProject"
          >
            <a-select-option
              v-for="module in modules"
              :key="module.id"
              :value="module.id"
            >
              {{ module.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="editFormData.status">
            <a-select-option value="active">激活</a-select-option>
            <a-select-option value="resolved">已解决</a-select-option>
            <a-select-option value="closed">已关闭</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="优先级" name="priority">
          <a-select v-model:value="editFormData.priority">
            <a-select-option value="low">低</a-select-option>
            <a-select-option value="medium">中</a-select-option>
            <a-select-option value="high">高</a-select-option>
            <a-select-option value="urgent">紧急</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="严重程度" name="severity">
          <a-select v-model:value="editFormData.severity">
            <a-select-option value="low">低</a-select-option>
            <a-select-option value="medium">中</a-select-option>
            <a-select-option value="high">高</a-select-option>
            <a-select-option value="critical">严重</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="指派给" name="assignee_ids">
          <ProjectMemberSelect
            v-model="editFormData.assignee_ids"
            :project-id="editFormData.project_id"
            :multiple="true"
            placeholder="选择指派给（可选）"
            :show-role="true"
          />
        </a-form-item>
        <a-form-item label="预估工时" name="estimated_hours">
          <a-input-number
            v-model:value="editFormData.estimated_hours"
            placeholder="预估工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="实际工时" name="actual_hours">
          <a-input-number
            v-model:value="editFormData.actual_hours"
            placeholder="实际工时（小时）"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配（使用第一个分配人）</span>
        </a-form-item>
        <a-form-item label="工作日期" name="work_date" v-if="editFormData.actual_hours">
          <a-date-picker
            v-model:value="editFormData.work_date"
            placeholder="选择工作日期（可选）"
            style="width: 100%"
            format="YYYY-MM-DD"
            :getPopupContainer="getPopupContainer"
            :popupStyle="{ zIndex: 2100 }"
          />
          <span style="margin-left: 8px; color: #999">不填则使用今天</span>
        </a-form-item>
        <a-form-item label="附件">
          <AttachmentUpload
            v-if="editFormData.project_id && editFormData.project_id > 0"
            :project-id="editFormData.project_id"
            v-model="editFormData.attachment_ids"
            :existing-attachments="bugAttachments"
          />
          <span v-else style="color: #999;">请先选择项目后再上传附件</span>
        </a-form-item>
      </a-form>
    </a-modal>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import AppHeader from '@/components/AppHeader.vue'
import BugDetailContent from '@/components/BugDetailContent.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import ProjectMemberSelect from '@/components/ProjectMemberSelect.vue'
import {
  getBug,
  getBugs,
  updateBug,
  updateBugStatus,
  deleteBug,
  assignBug,
  confirmBug,
  type Bug,
  type CreateBugRequest
} from '@/api/bug'
import { getUsers, type User } from '@/api/user'
import { getVersions, type Version } from '@/api/version'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, createRequirement, type Requirement, type CreateRequirementRequest } from '@/api/requirement'
import { getModules, type Module } from '@/api/module'
import { getAttachments, attachToEntity, uploadFile, type Attachment } from '@/api/attachment'
import type { Dayjs } from 'dayjs'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const bug = ref<Bug | null>(null)
const users = ref<User[]>([])
const projects = ref<Project[]>([])
const requirements = ref<Requirement[]>([])
const requirementLoading = ref(false)
const modules = ref<Module[]>([])
const moduleLoading = ref(false)

// 上一个/下一个bug导航
const prevBugId = ref<number | null>(null)
const nextBugId = ref<number | null>(null)
const bugListLoading = ref(false)
const assignModalVisible = ref(false)
const assignFormRef = ref()
const assignFormData = reactive({
  assignee_ids: [] as number[],
  status: undefined as string | undefined,
  comment: undefined as string | undefined
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
const loadBug = async (bugId?: number) => {
  const id = bugId || Number(route.params.id)
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
  
  // 异步加载相邻bug信息，不阻塞主流程
  loadAdjacentBugs(id).catch(err => {
    console.error('加载相邻bug失败:', err)
  })
}

// 加载相邻bug（上一个和下一个）
const loadAdjacentBugs = async (currentBugId: number) => {
  if (!currentBugId) return
  
  bugListLoading.value = true
  prevBugId.value = null
  nextBugId.value = null
  
  try {
    // 从URL查询参数获取筛选条件（如果存在）
    const baseParams: any = {}
    if (route.query.keyword) {
      baseParams.keyword = route.query.keyword as string
    }
    if (route.query.project_id) {
      baseParams.project_id = Number(route.query.project_id)
    }
    if (route.query.status) {
      baseParams.status = route.query.status as string
    }
    if (route.query.priority) {
      baseParams.priority = route.query.priority as string
    }
    if (route.query.severity) {
      baseParams.severity = route.query.severity as string
    }
    if (route.query.assignee_id) {
      baseParams.assignee_id = Number(route.query.assignee_id)
    }
    
    // 先获取总数，确定需要查询多少页
    const totalParams = { ...baseParams, page: 1, size: 100 }
    const totalResponse = await getBugs(totalParams)
    const total = totalResponse.total || 0
    const pageSize = 100 // 后端最大限制
    
    if (total === 0) {
      return
    }
    
    // 通过分页查询找到当前bug所在的页
    let currentPage = -1
    let currentIndex = -1
    const maxPages = Math.ceil(total / pageSize)
    
    // 线性查找当前bug所在的页
    for (let page = 1; page <= maxPages; page++) {
      const params = { ...baseParams, page, size: pageSize }
      const response = await getBugs(params)
      const bugs = response.list || []
      const index = bugs.findIndex(b => b.id === currentBugId)
      if (index !== -1) {
        currentPage = page
        currentIndex = index
        break
      }
    }
    
    if (currentPage === -1 || currentIndex === -1) {
      console.warn('当前bug不在列表中，ID:', currentBugId)
      return
    }
    
    // 获取当前页的bug列表
    const currentPageParams = { ...baseParams, page: currentPage, size: pageSize }
    const currentPageResponse = await getBugs(currentPageParams)
    const currentPageBugs = currentPageResponse.list || []
    
    // 获取上一个bug
    if (currentIndex > 0) {
      // 在当前页的前一个
      const prevBug = currentPageBugs[currentIndex - 1]
      if (prevBug) {
        prevBugId.value = prevBug.id
      }
    } else if (currentPage > 1) {
      // 在前一页的最后一个
      const prevPageParams = { ...baseParams, page: currentPage - 1, size: pageSize }
      const prevPageResponse = await getBugs(prevPageParams)
      const prevPageBugs = prevPageResponse.list || []
      if (prevPageBugs.length > 0) {
        const lastBug = prevPageBugs[prevPageBugs.length - 1]
        if (lastBug) {
          prevBugId.value = lastBug.id
        }
      }
    }
    
    // 获取下一个bug
    if (currentIndex < currentPageBugs.length - 1) {
      // 在当前页的下一个
      const nextBug = currentPageBugs[currentIndex + 1]
      if (nextBug) {
        nextBugId.value = nextBug.id
      }
    } else if (currentPage < maxPages) {
      // 在下一页的第一个
      const nextPageParams = { ...baseParams, page: currentPage + 1, size: pageSize }
      const nextPageResponse = await getBugs(nextPageParams)
      const nextPageBugs = nextPageResponse.list || []
      if (nextPageBugs.length > 0) {
        const firstBug = nextPageBugs[0]
        if (firstBug) {
          nextBugId.value = firstBug.id
        }
      }
    }
  } catch (error: any) {
    console.error('加载相邻bug失败:', error)
    message.error('加载相邻bug失败: ' + (error.message || '未知错误'))
  } finally {
    bugListLoading.value = false
  }
}

// 导航到上一个bug
const handleNavigateToPrev = () => {
  if (!prevBugId.value) {
    message.warning('没有上一个bug')
    return
  }
  // 保持当前的查询参数
  const query = { ...route.query }
  router.push({ path: `/bug/${prevBugId.value}`, query })
}

// 导航到下一个bug
const handleNavigateToNext = () => {
  if (!nextBugId.value) {
    message.warning('没有下一个bug')
    return
  }
  // 保持当前的查询参数
  const query = { ...route.query }
  router.push({ path: `/bug/${nextBugId.value}`, query })
}

// 处理刷新事件
const handleRefresh = async () => {
  if (bug.value?.id) {
    await loadBug(bug.value.id)
  }
}

// 处理需求点击事件
const handleRequirementClick = (requirementId: number) => {
  router.push(`/requirement/${requirementId}`)
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

// 编辑模态框相关
const editModalVisible = ref(false)
const editFormRef = ref()
const editDescriptionEditorRef = ref<any>(null)
const editFormData = reactive<CreateBugRequest & { id?: number; attachment_ids?: number[]; work_date?: Dayjs | undefined }>({
  title: '',
  description: '',
  status: 'active',
  priority: 'medium',
  severity: 'medium',
  project_id: 0,
  requirement_id: undefined,
  module_id: undefined,
  assignee_ids: [],
  estimated_hours: undefined,
  actual_hours: undefined,
  work_date: undefined,
  attachment_ids: [] as number[]
})
const editFormRules = {
  title: [{ required: true, message: '请输入Bug标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}
const bugAttachments = ref<Attachment[]>([])

// 编辑
const handleEdit = async () => {
  if (!bug.value) return
  
  editFormData.id = bug.value.id
  editFormData.title = bug.value.title
  editFormData.description = bug.value.description || ''
  editFormData.status = bug.value.status
  editFormData.priority = bug.value.priority
  editFormData.severity = bug.value.severity
  editFormData.project_id = bug.value.project_id
  editFormData.requirement_id = bug.value.requirement_id
  editFormData.module_id = bug.value.module_id
  editFormData.assignee_ids = bug.value.assignees?.map(a => a.id) || []
  editFormData.estimated_hours = bug.value.estimated_hours
  editFormData.actual_hours = bug.value.actual_hours
  editFormData.work_date = undefined
  
  // 加载Bug附件
  try {
    bugAttachments.value = await getAttachments({ bug_id: bug.value.id })
    editFormData.attachment_ids = bugAttachments.value.map(a => a.id)
  } catch (error: any) {
    console.error('加载附件失败:', error)
    bugAttachments.value = []
    editFormData.attachment_ids = []
  }
  
  editModalVisible.value = true
  if (editFormData.project_id) {
    loadRequirementsForProject()
  }
}

// 编辑提交
const handleEditSubmit = async () => {
  if (!bug.value) return
  
  try {
    await editFormRef.value.validate()
    
    // 获取最新的描述内容
    let description = editFormData.description || ''
    
    // 如果有项目ID，尝试上传本地图片（如果有的话）
    if (editDescriptionEditorRef.value && editFormData.project_id) {
      try {
        const uploadedDescription = await editDescriptionEditorRef.value.uploadLocalImages(async (file: File, projectId: number) => {
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
    
    // 构建请求数据，确保 requirement_id 和 module_id 始终发送（即使是0也要发送）
    const data: any = {
      title: editFormData.title,
      description: description || '',
      status: editFormData.status,
      priority: editFormData.priority,
      severity: editFormData.severity,
      project_id: editFormData.project_id,
      assignee_ids: editFormData.assignee_ids,
      estimated_hours: editFormData.estimated_hours,
      actual_hours: editFormData.actual_hours,
      work_date: editFormData.work_date && typeof editFormData.work_date !== 'string' && 'isValid' in editFormData.work_date && (editFormData.work_date as Dayjs).isValid() ? (editFormData.work_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.work_date === 'string' ? editFormData.work_date : undefined)
    }
    
    // 始终发送 requirement_id，如果为 undefined 或 null，发送 0 以清空关联需求（后端会将0转换为nil）
    // 注意：必须显式设置，不能依赖对象字面量，因为 undefined 值会被 JSON 序列化忽略
    const requirementIdValue = editFormData.requirement_id
    if (requirementIdValue === undefined || requirementIdValue === null || (typeof requirementIdValue === 'number' && isNaN(requirementIdValue))) {
      data.requirement_id = 0
    } else {
      data.requirement_id = requirementIdValue
    }
    
    // 始终发送 module_id，如果为 undefined 或 null，发送 0 以清空关联模块（后端会将0转换为nil）
    const moduleIdValue = editFormData.module_id
    if (moduleIdValue === undefined || moduleIdValue === null || (typeof moduleIdValue === 'number' && isNaN(moduleIdValue))) {
      data.module_id = 0
    } else {
      data.module_id = moduleIdValue
    }
    
    await updateBug(bug.value.id, data)
    
    // 处理附件关联
    if (editFormData.attachment_ids && editFormData.attachment_ids.length > 0 && editFormData.project_id) {
      try {
        for (const attachmentId of editFormData.attachment_ids) {
          await attachToEntity(attachmentId, { bug_id: bug.value.id })
        }
      } catch (error: any) {
        console.error('关联附件到Bug失败:', error)
      }
    }
    
    message.success('更新成功')
    editModalVisible.value = false
    await loadBug(bug.value.id) // 重新加载Bug详情（会自动加载历史记录）
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
  requirements.value = []
}

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await getProjects({ size: 1000 })
    projects.value = response.list || []
  } catch (error: any) {
    console.error('加载项目列表失败:', error)
  }
}

// 加载需求列表（根据项目）
const loadRequirementsForProject = async () => {
  if (!editFormData.project_id) {
    requirements.value = []
    return
  }
  requirementLoading.value = true
  try {
    const response = await getRequirements({ project_id: editFormData.project_id })
    requirements.value = response.list || []
  } catch (error: any) {
    console.error('加载需求列表失败:', error)
  } finally {
    requirementLoading.value = false
  }
}

// 加载模块列表
const loadModulesForProject = async () => {
  moduleLoading.value = true
  try {
    modules.value = await getModules()
  } catch (error: any) {
    console.error('加载模块列表失败:', error)
  } finally {
    moduleLoading.value = false
  }
}

// 监听编辑表单项目变化
watch(() => editFormData.project_id, () => {
  editFormData.requirement_id = undefined
  if (editFormData.project_id) {
    loadRequirementsForProject()
  } else {
    requirements.value = []
  }
})

// 项目筛选
const filterProjectOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 需求筛选
const filterRequirementOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 模块筛选
const filterModuleOption = (input: string, option: any) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

// 编辑表单项目选择改变
const handleEditFormProjectChange = () => {
  // watch会自动处理
}

// 指派
const handleAssign = () => {
  if (!bug.value) return
  assignFormData.assignee_ids = [] // 默认清空，不预填充当前值
  assignFormData.status = bug.value.status // 设置当前Bug的状态
  assignFormData.comment = undefined // 清空备注
  assignModalVisible.value = true
}

// 指派提交
const handleAssignSubmit = async () => {
  if (!bug.value) return
  try {
    await assignFormRef.value.validate()
    const requestData: any = { assignee_ids: assignFormData.assignee_ids }
    if (assignFormData.status) {
      requestData.status = assignFormData.status
    }
    if (assignFormData.comment) {
      requestData.comment = assignFormData.comment
    }
    await assignBug(bug.value.id, requestData)
    message.success('指派成功')
    assignModalVisible.value = false
    await loadBug(bug.value.id)
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
  assignFormData.status = undefined
  assignFormData.comment = undefined
  assignFormData.assignee_ids = []
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

// 获取下拉框容器（用于解决模态框中下拉框被遮挡的问题）
const getPopupContainer = (triggerNode: HTMLElement): HTMLElement => {
  return triggerNode.parentElement || document.body
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
    await loadBug(bug.value.id)
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
    await loadBug(bug.value.id)
  } catch (error: any) {
    message.error(error.message || '确认失败')
  }
}

// 关闭Bug
const handleClose = async () => {
  if (!bug.value) return
  
  // 只有resolved状态才能关闭
  if (bug.value.status !== 'resolved') {
    message.warning('只有已解决的Bug才能关闭')
    return
  }
  
  try {
    await updateBugStatus(bug.value.id, { status: 'closed' })
    message.success('关闭成功')
    await loadBug(bug.value.id)
  } catch (error: any) {
    message.error(error.message || '关闭失败')
  }
}

// Bug转需求
const handleConvertToRequirement = async () => {
  if (!bug.value) return
  
  // 确认对话框
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此Bug转为需求吗？转换后将创建新需求，并将Bug状态更新为"已解决"。',
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        resolve(true)
        modal.destroy()
      },
      onCancel: () => {
        resolve(false)
        modal.destroy()
      }
    })
  })
  
  if (!confirmed) return
  
  try {
    // 创建新需求，基于bug的信息
    const requirementData: CreateRequirementRequest = {
      title: `[转需求] ${bug.value.title}`,
      description: bug.value.description 
        ? `${bug.value.description}\n\n---\n\n*由Bug #${bug.value.id}转换而来*`
        : `*由Bug #${bug.value.id}转换而来*`,
      project_id: bug.value.project_id,
      priority: bug.value.priority,
      status: 'draft', // 默认草稿状态
      // 如果bug有指派人员，使用第一个作为需求的负责人
      assignee_id: bug.value.assignees && bug.value.assignees.length > 0 
        ? bug.value.assignees[0].id 
        : undefined,
      estimated_hours: bug.value.estimated_hours
    }
    
    // 创建需求
    const requirement = await createRequirement(requirementData)
    
    // 更新bug状态为resolved，解决方案为"转为研发需求"
    await updateBugStatus(bug.value.id, {
      status: 'resolved',
      solution: '转为研发需求',
      solution_note: `已转为需求 #${requirement.id}`
    })
    
    // 关联新创建的需求到bug
    await updateBug(bug.value.id, {
      requirement_id: requirement.id
    })
    
    message.success(`转换成功，已创建需求 #${requirement.id}`)
    
    // 刷新bug详情
    await loadBug(bug.value.id)
  } catch (error: any) {
    message.error(error.message || '转换失败')
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





// 监听路由变化，当bug ID改变时重新加载
watch(() => route.params.id, (newId) => {
  if (newId) {
    loadBug(Number(newId))
  }
}, { immediate: false })

onMounted(() => {
  loadBug()
  loadUsers()
  loadProjects()
  loadModulesForProject()
})
</script>

<style scoped>
.bug-detail {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.bug-detail :deep(.ant-layout) {
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
  max-width: 100%;
  width: 100%;
  margin: 0 auto;
  min-height: fit-content;
}

.markdown-content {
  min-height: 200px;
}

/* 让导航按钮固定在左侧 */
.bug-detail :deep(.ant-page-header-heading-extra) {
  display: flex;
  justify-content: flex-start;
  align-items: center;
}

.bug-detail :deep(.ant-page-header-heading-extra > div) {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

