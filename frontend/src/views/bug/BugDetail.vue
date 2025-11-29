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
                  <a-space>
                    <a-tag :color="getStatusColor(bug?.status || '')">
                      {{ getStatusText(bug?.status || '') }}
                    </a-tag>
                    <a-tag v-if="bug?.confirmed" color="green">已确认</a-tag>
                    <a-tag v-else-if="bug?.status === 'active'" color="orange">未确认</a-tag>
                  </a-space>
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
                <a-descriptions-item label="指派给">
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
          </a-spin>
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
      </a-form>
    </a-modal>

    <!-- Bug解决模态框 -->
    <a-modal
      v-model:open="statusModalVisible"
      title="解决Bug"
      :width="600"
      :mask-closable="true"
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
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { formatDateTime } from '@/utils/date'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import ProjectMemberSelect from '@/components/ProjectMemberSelect.vue'
import {
  getBug,
  updateBug,
  updateBugStatus,
  deleteBug,
  assignBug,
  confirmBug,
  getBugHistory,
  addBugHistoryNote,
  type Bug,
  type Action,
  type CreateBugRequest
} from '@/api/bug'
import { getUsers, type User } from '@/api/user'
import { getVersions, type Version } from '@/api/version'
import { getProjects, type Project } from '@/api/project'
import { getRequirements, type Requirement } from '@/api/requirement'
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
const assignModalVisible = ref(false)
const assignFormRef = ref()
const assignFormData = reactive({
  assignee_ids: [] as number[]
})

// 历史记录相关
const historyLoading = ref(false)
const historyList = ref<Action[]>([])
const expandedHistoryIds = ref<Set<number>>(new Set()) // 展开的历史记录ID集合
const noteModalVisible = ref(false)
const noteFormRef = ref()
const noteFormData = reactive({
  comment: ''
})
const noteFormRules = {
  comment: [{ required: true, message: '请输入备注', trigger: 'blur' }]
}

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
    // 加载历史记录
    await loadBugHistory(id)
  } catch (error: any) {
    message.error(error.message || '加载Bug详情失败')
    router.push('/bug')
  } finally {
    loading.value = false
  }
}

// 加载历史记录
const loadBugHistory = async (bugId?: number) => {
  const id = bugId || Number(route.params.id)
  if (!id) return

  historyLoading.value = true
  try {
    const response = await getBugHistory(id)
    historyList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    historyLoading.value = false
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

// 编辑模态框相关
const editModalVisible = ref(false)
const editFormRef = ref()
const editDescriptionEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
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
    
    const data: CreateBugRequest = {
      title: editFormData.title,
      description: description || '',
      status: editFormData.status,
      priority: editFormData.priority,
      severity: editFormData.severity,
      project_id: editFormData.project_id,
      requirement_id: editFormData.requirement_id,
      module_id: editFormData.module_id,
      assignee_ids: editFormData.assignee_ids,
      estimated_hours: editFormData.estimated_hours,
      actual_hours: editFormData.actual_hours,
      work_date: editFormData.work_date && typeof editFormData.work_date !== 'string' && 'isValid' in editFormData.work_date && (editFormData.work_date as Dayjs).isValid() ? (editFormData.work_date as Dayjs).format('YYYY-MM-DD') : (typeof editFormData.work_date === 'string' ? editFormData.work_date : undefined)
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
  assignFormData.assignee_ids = bug.value.assignees?.map(a => a.id) || []
  assignModalVisible.value = true
}

// 指派提交
const handleAssignSubmit = async () => {
  if (!bug.value) return
  try {
    await assignFormRef.value.validate()
    await assignBug(bug.value.id, { assignee_ids: assignFormData.assignee_ids })
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
const getStatusText = (status: string | undefined) => {
  if (!status) return '-'
  const texts: Record<string, string> = {
    active: '激活',
    resolved: '已解决',
    closed: '已关闭'
  }
  return texts[status.toLowerCase()] || status
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



// 获取操作描述
const getActionDescription = (action: Action): string => {
  const actorName = action.actor
    ? `${action.actor.username}${action.actor.nickname ? `(${action.actor.nickname})` : ''}`
    : '系统'

  switch (action.action) {
    case 'created':
      return `由 ${actorName} 创建。`
    case 'assigned':
      // 从histories中获取指派信息
      const assignHistory = action.histories?.find(h => h.field === 'assignee_ids')
      if (assignHistory) {
        return `由 ${actorName} 指派给 ${assignHistory.new_value || assignHistory.new || '-'}。`
      }
      return `由 ${actorName} 指派。`
    case 'resolved':
      // 从extra中获取解决方案（如果有）
      let solution = ''
      if (action.extra) {
        try {
          const extra = JSON.parse(action.extra)
          if (extra.solution) {
            solution = extra.solution
          }
        } catch (e) {
          // 解析失败，忽略
        }
      }
      return `由 ${actorName} 解决${solution ? `, 方案为 ${solution}。` : '。'}`
    case 'closed':
      return `由 ${actorName} 关闭。`
    case 'confirmed':
      return `由 ${actorName} 确认。`
    case 'edited':
      return `由 ${actorName} 编辑。`
    case 'commented':
      return `由 ${actorName} 添加了备注：${action.comment || ''}`
    default:
      return `由 ${actorName} 执行了 ${action.action} 操作。`
  }
}

// 获取字段显示名称
const getFieldDisplayName = (fieldName: string): string => {
  const fieldNames: Record<string, string> = {
    title: 'Bug标题',
    description: 'Bug描述',
    status: 'Bug状态',
    priority: '优先级',
    severity: '严重程度',
    confirmed: '是否确认',
    project_id: '项目',
    requirement_id: '关联需求',
    module_id: '功能模块',
    assignee_ids: '指派给',
    estimated_hours: '预估工时',
    actual_hours: '实际工时',
    solution: '解决方案',
    solution_note: '解决方案备注',
    resolved_version_id: '解决版本'
  }
  return fieldNames[fieldName] || fieldName
}

// 判断历史记录是否有详情（字段变更或备注）
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

// 添加备注
const handleAddNote = () => {
  if (!bug.value) {
    message.warning('Bug信息未加载完成，请稍候再试')
    return
  }
  noteFormData.comment = ''
  noteModalVisible.value = true
}

// 提交备注
const handleNoteSubmit = async () => {
  if (!bug.value) return
  try {
    await noteFormRef.value.validate()
    await addBugHistoryNote(bug.value.id, { comment: noteFormData.comment })
    message.success('添加备注成功')
    noteModalVisible.value = false
    await loadBugHistory(bug.value.id)
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
</style>

