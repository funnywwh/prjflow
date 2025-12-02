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

          <a-tabs v-model:activeKey="activeTab">
            <!-- 统计标签页 -->
            <a-tab-pane key="statistics" tab="统计">
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
              <a-card title="优先级统计" :bordered="false" class="priority-statistics-card">
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
            </a-tab-pane>

            <!-- 列表标签页 -->
            <a-tab-pane key="list" tab="列表">
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
                      placeholder="需求标题/描述"
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
                      <a-select-option value="reviewing">评审中</a-select-option>
                      <a-select-option value="active">激活</a-select-option>
                      <a-select-option value="changing">变更中</a-select-option>
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
                  <a-form-item>
                    <a-button type="primary" @click="handleSearch">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false" class="table-card">
                <a-table
                  :columns="columns"
                  :data-source="requirements"
                  :loading="loading"
                  :pagination="pagination"
                  :scroll="{ x: 'max-content', y: tableScrollHeight }"
                  row-key="id"
                  @change="handleTableChange"
                  :custom-row="(record: Requirement) => ({
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
                <template v-else-if="column.key === 'priority'">
                  <a-tag :color="getPriorityColor(record.priority)">
                    {{ getPriorityText(record.priority) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'project'">
                  {{ record.project?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'assignee'">
                  {{ record.assignee ? `${record.assignee.username}${record.assignee.nickname ? `(${record.assignee.nickname})` : ''}` : '-' }}
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
                  <a-space @click.stop>
                    <a-button type="link" size="small" @click.stop="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-dropdown>
                      <a-button type="link" size="small">
                        状态 <DownOutlined />
                      </a-button>
                      <template #overlay>
                        <a-menu @click="(e: any) => handleStatusChange(record.id, e.key as string)">
                          <a-menu-item key="pending">待处理</a-menu-item>
                          <a-menu-item key="draft">草稿</a-menu-item>
                          <a-menu-item key="reviewing">评审中</a-menu-item>
                          <a-menu-item key="active">激活</a-menu-item>
                          <a-menu-item key="changing">变更中</a-menu-item>
                          <a-menu-item key="closed">已关闭</a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                    <a-popconfirm
                      title="确定要删除这个需求吗？"
                      @confirm="handleDelete(record.id)"
                    >
                      <a-button type="link" size="small" danger @click.stop>删除</a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </template>
            </a-table>
          </a-card>
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 需求编辑/创建模态框 -->
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
        <a-form-item label="需求标题" name="title">
          <a-input v-model:value="formData.title" placeholder="请输入需求标题" />
        </a-form-item>
        <a-form-item label="需求描述" name="description">
          <MarkdownEditor
            ref="descriptionEditorRef"
            v-model="formData.description"
            placeholder="请输入需求描述（支持Markdown）"
            :rows="8"
            :project-id="formData.project_id || 0"
          />
        </a-form-item>
        <a-form-item label="项目" name="project_id">
          <a-select
            v-model:value="formData.project_id"
            placeholder="请选择项目"
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
            <a-select-option value="draft">草稿</a-select-option>
            <a-select-option value="reviewing">评审中</a-select-option>
            <a-select-option value="active">激活</a-select-option>
            <a-select-option value="changing">变更中</a-select-option>
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
          <span style="margin-left: 8px; color: #999">更新实际工时会自动创建资源分配</span>
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
        <a-form-item label="附件">
          <AttachmentUpload
            v-if="formData.project_id && (formData.id || formData.project_id)"
            :project-id="formData.project_id"
            v-model="formData.attachment_ids"
            :existing-attachments="requirementAttachments"
          />
          <span v-else style="color: #999;">请先选择项目后再上传附件</span>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 需求详情弹窗 -->
    <a-modal
      v-model:open="detailModalVisible"
      :title="detailRequirement?.title || '需求详情'"
      :width="1200"
      :mask-closable="true"
      :footer="null"
      @cancel="handleDetailCancel"
    >
      <a-spin :spinning="detailLoading">
        <div v-if="detailRequirement" style="max-height: 70vh; overflow-y: auto">
          <!-- 操作按钮 -->
          <div style="margin-bottom: 16px; text-align: right">
            <a-space>
              <a-button @click="handleDetailEdit">编辑</a-button>
              <a-dropdown>
                <a-button>
                  状态 <DownOutlined />
                </a-button>
                <template #overlay>
                  <a-menu @click="(e: any) => handleDetailStatusChange(e.key as string)">
                    <a-menu-item key="pending">待处理</a-menu-item>
                    <a-menu-item key="draft">草稿</a-menu-item>
                    <a-menu-item key="reviewing">评审中</a-menu-item>
                    <a-menu-item key="active">激活</a-menu-item>
                    <a-menu-item key="changing">变更中</a-menu-item>
                    <a-menu-item key="closed">已关闭</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
              <a-button @click="handleDetailConvertToBug">
                需求转Bug
              </a-button>
              <a-popconfirm
                title="确定要删除这个需求吗？"
                @confirm="handleDetailDelete"
              >
                <a-button danger>删除</a-button>
              </a-popconfirm>
            </a-space>
          </div>

          <!-- 基本信息 -->
          <a-card title="基本信息" :bordered="false" style="margin-bottom: 16px">
            <a-descriptions :column="2" bordered>
              <a-descriptions-item label="需求标题">{{ detailRequirement.title }}</a-descriptions-item>
              <a-descriptions-item label="状态">
                <a-tag :color="getStatusColor(detailRequirement.status || '')">
                  {{ getStatusText(detailRequirement.status || '') }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="优先级">
                <a-tag :color="getPriorityColor(detailRequirement.priority || '')">
                  {{ getPriorityText(detailRequirement.priority || '') }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item label="项目">
                {{ detailRequirement.project?.name || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="负责人">
                {{ detailRequirement.assignee ? `${detailRequirement.assignee.username}${detailRequirement.assignee.nickname ? `(${detailRequirement.assignee.nickname})` : ''}` : '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="创建人">
                {{ detailRequirement.creator ? `${detailRequirement.creator.username}${detailRequirement.creator.nickname ? `(${detailRequirement.creator.nickname})` : ''}` : '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="创建时间">
                {{ formatDateTime(detailRequirement.created_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="更新时间">
                {{ formatDateTime(detailRequirement.updated_at) }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <!-- 需求描述 -->
          <a-card title="需求描述" :bordered="false" style="margin-bottom: 16px">
            <div v-if="detailRequirement.description" class="markdown-content">
              <MarkdownEditor
                :model-value="detailRequirement.description"
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
                @click.stop="handleDetailAddNote" 
                :disabled="detailHistoryLoading"
                style="margin-left: 8px; padding: 0"
              >
                添加备注
              </a-button>
            </template>
            <a-spin :spinning="detailHistoryLoading" :style="{ minHeight: '100px' }">
              <a-timeline v-if="detailHistoryList.length > 0">
                <a-timeline-item
                  v-for="(action, index) in detailHistoryList"
                  :key="action.id"
                >
                  <template #dot>
                    <span style="font-weight: bold; color: #1890ff">{{ detailHistoryList.length - index }}</span>
                  </template>
                  <div>
                    <div style="margin-bottom: 8px">
                      <span style="color: #666; margin-right: 8px">{{ formatDateTime(action.date) }}</span>
                      <span>{{ getDetailActionDescription(action) }}</span>
                      <a-button
                        v-if="hasDetailHistoryDetails(action)"
                        type="link"
                        size="small"
                        @click="toggleDetailHistoryDetail(action.id)"
                        style="padding: 0; height: auto; margin-left: 8px"
                      >
                        {{ detailExpandedHistoryIds.has(action.id) ? '收起' : '展开' }}
                      </a-button>
                    </div>
                    <!-- 字段变更详情和备注内容（可折叠） -->
                    <div
                      v-show="detailExpandedHistoryIds.has(action.id)"
                      style="margin-left: 24px; margin-top: 8px"
                    >
                      <!-- 字段变更详情 -->
                      <div v-if="action.histories && action.histories.length > 0">
                        <div
                          v-for="history in action.histories"
                          :key="history.id"
                          style="margin-bottom: 8px; color: #666"
                        >
                          <div>修改了{{ getDetailFieldDisplayName(history.field) }}</div>
                          <div style="margin-left: 16px; margin-top: 4px;">
                            <div>旧值："{{ history.old_value || history.old || '-' }}"</div>
                            <div>新值："{{ history.new_value || history.new || '-' }}"</div>
                          </div>
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
        </div>
      </a-spin>
    </a-modal>

    <!-- 详情页添加备注模态框 -->
    <a-modal
      v-model:open="detailNoteModalVisible"
      title="添加备注"
      :mask-closable="true"
      @ok="handleDetailNoteSubmit"
      @cancel="handleDetailNoteCancel"
    >
      <a-form
        ref="detailNoteFormRef"
        :model="detailNoteFormData"
        :rules="detailNoteFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="备注" name="comment">
          <a-textarea
            v-model:value="detailNoteFormData.comment"
            placeholder="请输入备注"
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch, nextTick } from 'vue'
import { saveLastSelected, getLastSelected } from '@/utils/storage'
import { useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, DownOutlined, UpOutlined } from '@ant-design/icons-vue'
import { formatDateTime } from '@/utils/date'
import { type Dayjs } from 'dayjs'
import AppHeader from '@/components/AppHeader.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentUpload from '@/components/AttachmentUpload.vue'
import { useAuthStore } from '@/stores/auth'
import {
  getRequirements,
  getRequirement,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  updateRequirementStatus,
  getRequirementStatistics,
  getRequirementHistory,
  addRequirementHistoryNote,
  type Requirement,
  type CreateRequirementRequest,
  type RequirementStatistics,
  type Action
} from '@/api/requirement'
import { getProjects, type Project } from '@/api/project'
import { getUsers, type User } from '@/api/user'
import { createBug, type CreateBugRequest } from '@/api/bug'
import { getAttachments, attachToEntity, uploadFile, type Attachment } from '@/api/attachment'

const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const requirements = ref<Requirement[]>([])
const projects = ref<Project[]>([])
const users = ref<User[]>([])
const statistics = ref<RequirementStatistics | null>(null)
const activeTab = ref<string>('list')
const searchFormVisible = ref(false) // 搜索栏显示/隐藏状态，默认折叠

// 详情弹窗相关
const detailModalVisible = ref(false)
const detailLoading = ref(false)
const detailRequirement = ref<Requirement | null>(null)
const detailHistoryLoading = ref(false)
const detailHistoryList = ref<Action[]>([])
const detailExpandedHistoryIds = ref<Set<number>>(new Set())
const detailNoteModalVisible = ref(false)
const detailNoteFormRef = ref()
const detailNoteFormData = reactive({
  comment: ''
})
const detailNoteFormRules = {
  comment: [{ required: true, message: '请输入备注', trigger: 'blur' }]
}
const shouldKeepDetailOpen = ref(false)

const searchForm = reactive({
  keyword: '',
  project_id: undefined as number | undefined,
  status: undefined as string | undefined,
  priority: undefined as string | undefined,
  assignee_id: undefined as number | undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

// 计算表格滚动高度（考虑 tab 标签页的高度）
const tableScrollHeight = computed(() => {
  return 'calc(100vh - 550px)'
})

const columns = [
  { title: '需求标题', dataIndex: 'title', key: 'title', ellipsis: true },
  { title: '项目', key: 'project', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '优先级', key: 'priority', width: 100 },
  { title: '负责人', key: 'assignee', width: 150 },
  { title: '工时', key: 'hours', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 180 },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const }
]

const modalVisible = ref(false)
const modalTitle = ref('新增需求')
const formRef = ref()
const descriptionEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
const formData = reactive<Partial<CreateRequirementRequest> & { id?: number; actual_hours?: number; work_date?: Dayjs; attachment_ids?: number[] }>({
  title: '',
  description: '',
  status: 'draft',
  priority: 'medium',
  project_id: undefined, // 表单中可以为undefined，提交时会验证
  assignee_id: undefined,
  estimated_hours: undefined,
  actual_hours: undefined,
  work_date: undefined,
  attachment_ids: [] as number[]
})

const requirementAttachments = ref<Attachment[]>([]) // 需求附件列表

const formRules = {
  title: [{ required: true, message: '请输入需求标题', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

// 加载需求列表
const loadRequirements = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
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
    if (searchForm.assignee_id) {
      params.assignee_id = searchForm.assignee_id
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
    if (searchForm.project_id) {
      params.project_id = searchForm.project_id
    }
    statistics.value = await getRequirementStatistics(params)
  } catch (error: any) {
    console.error('加载统计信息失败:', error)
  }
}

// 加载项目列表
const loadProjects = async () => {
  try {
    // 获取所有项目（不分页），用于下拉选择器
    const response = await getProjects({ size: 1000 })
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

// 切换搜索栏显示/隐藏
const toggleSearchForm = () => {
  searchFormVisible.value = !searchFormVisible.value
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadRequirements()
}

// 搜索表单项目选择改变
const handleSearchProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_requirement_project_search', value)
}

// 编辑表单项目选择改变
const handleFormProjectChange = (value: number | undefined) => {
  saveLastSelected('last_selected_requirement_project_form', value)
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.project_id = undefined
  searchForm.status = undefined
  searchForm.priority = undefined
  searchForm.assignee_id = undefined
  pagination.current = 1
  // 清除保存的搜索项目选择
  saveLastSelected('last_selected_requirement_project_search', undefined)
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
  formData.status = 'draft'
  formData.priority = 'medium'
  // 从 localStorage 恢复最后选择的项目
  const lastProjectId = getLastSelected<number>('last_selected_requirement_project_form')
  formData.project_id = lastProjectId
  formData.assignee_id = undefined
  formData.estimated_hours = undefined
  formData.actual_hours = undefined
  formData.work_date = undefined
  formData.attachment_ids = []
  requirementAttachments.value = []
  modalVisible.value = true
}

// 编辑
const handleEdit = async (record: Requirement) => {
  modalTitle.value = '编辑需求'
  formData.id = record.id
  formData.title = record.title
  formData.description = record.description || ''
  formData.status = record.status
  formData.priority = record.priority
  formData.project_id = record.project_id
  formData.assignee_id = record.assignee_id
  formData.estimated_hours = record.estimated_hours
  formData.actual_hours = record.actual_hours
  formData.work_date = undefined
  
  // 加载需求附件
  try {
    requirementAttachments.value = await getAttachments({ requirement_id: record.id })
    formData.attachment_ids = requirementAttachments.value.map(a => a.id)
  } catch (error: any) {
    console.error('加载附件失败:', error)
    requirementAttachments.value = []
    formData.attachment_ids = []
  }
  
  modalVisible.value = true
}

// 查看详情
const handleView = async (record: Requirement) => {
  detailModalVisible.value = true
  await loadRequirementDetail(record.id)
}

// 提交
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    // 验证项目ID必填
    if (!formData.project_id) {
      message.error('请选择项目')
      return
    }
    // 上传Markdown编辑器中的本地图片
    let description = formData.description || ''
    if (descriptionEditorRef.value && formData.project_id) {
      try {
        description = await descriptionEditorRef.value.uploadLocalImages(async (file: File, projectId: number) => {
          const attachment = await uploadFile(file, projectId)
          console.log('图片上传成功，附件信息:', attachment)
          return attachment
        })
        console.log('上传图片后的description:', description)
      } catch (error: any) {
        console.error('上传图片失败:', error)
        message.warning('部分图片上传失败，请检查')
      }
    }
    
    const data: any = {
      title: formData.title,
      description: description, // 使用已上传图片的description
      status: formData.status,
      priority: formData.priority,
      project_id: formData.project_id, // 必填
      assignee_id: formData.assignee_id,
      estimated_hours: formData.estimated_hours,
      actual_hours: formData.actual_hours,
      work_date: formData.work_date ? formData.work_date.format('YYYY-MM-DD') : undefined
    }
    // 调试：检查提交的数据
    console.log('提交的数据:', {
      description: data.description,
      hasImages: data.description?.includes('/uploads/')
    })
    
    let requirementId: number
    if (formData.id) {
      requirementId = formData.id
      await updateRequirement(requirementId, data)
      message.success('更新成功')
    } else {
      const newRequirement = await createRequirement(data)
      requirementId = newRequirement.id
      message.success('创建成功')
      
      // 创建需求后，如果有待上传的附件，需要关联到需求
      // 附件上传组件会在上传时自动关联到项目，这里需要额外关联到需求
      if (formData.attachment_ids && formData.attachment_ids.length > 0 && formData.project_id) {
        try {
          for (const attachmentId of formData.attachment_ids) {
            await attachToEntity(attachmentId, { requirement_id: requirementId })
          }
        } catch (error: any) {
          console.error('关联附件到需求失败:', error)
        }
      }
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
    draft: 'orange',
    reviewing: 'purple',
    active: 'blue',
    changing: 'cyan',
    closed: 'default'
  }
  return colors[status] || 'default'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    reviewing: '评审中',
    active: '激活',
    changing: '变更中',
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

// 加载需求详情
const loadRequirementDetail = async (requirementId: number) => {
  detailLoading.value = true
  try {
    detailRequirement.value = await getRequirement(requirementId)
    await loadRequirementDetailHistory(requirementId)
  } catch (error: any) {
    message.error(error.message || '加载需求详情失败')
    detailModalVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

// 加载需求详情历史记录
const loadRequirementDetailHistory = async (requirementId: number) => {
  detailHistoryLoading.value = true
  try {
    const response = await getRequirementHistory(requirementId)
    detailHistoryList.value = response.list || []
  } catch (error: any) {
    console.error('加载历史记录失败:', error)
  } finally {
    detailHistoryLoading.value = false
  }
}

// 详情弹窗取消
const handleDetailCancel = () => {
  detailRequirement.value = null
  detailHistoryList.value = []
  detailExpandedHistoryIds.value = new Set()
}

// 详情页编辑
const handleDetailEdit = async () => {
  if (!detailRequirement.value) return
  shouldKeepDetailOpen.value = true
  detailModalVisible.value = false
  await nextTick()
  handleEdit(detailRequirement.value)
}

// 详情页状态变更
const handleDetailStatusChange = async (status: string) => {
  if (!detailRequirement.value) return
  try {
    await updateRequirementStatus(detailRequirement.value.id, { status: status as any })
    message.success('状态更新成功')
    await loadRequirementDetail(detailRequirement.value.id)
    loadRequirements()
  } catch (error: any) {
    message.error(error.message || '状态更新失败')
  }
}

// 详情页需求转Bug
const handleDetailConvertToBug = async () => {
  if (!detailRequirement.value) return
  
  const confirmed = await new Promise<boolean>((resolve) => {
    const modal = Modal.confirm({
      title: '确认转换',
      content: '确定要将此需求转为Bug吗？转换后将创建新Bug，并将需求状态更新为"已关闭"。',
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
    const bugData: CreateBugRequest = {
      title: `[需求转Bug] ${detailRequirement.value.title}`,
      description: detailRequirement.value.description 
        ? `${detailRequirement.value.description}\n\n---\n\n*由需求 #${detailRequirement.value.id}转换而来*`
        : `*由需求 #${detailRequirement.value.id}转换而来*`,
      project_id: detailRequirement.value.project_id,
      priority: detailRequirement.value.priority,
      status: 'active',
      severity: 'medium',
      assignee_ids: detailRequirement.value.assignee_id ? [detailRequirement.value.assignee_id] : [],
      estimated_hours: detailRequirement.value.estimated_hours
    }
    
    const bug = await createBug(bugData)
    
    await updateRequirementStatus(detailRequirement.value.id, {
      status: 'closed'
    })
    
    message.success(`转换成功，已创建Bug #${bug.id}`)
    await loadRequirementDetail(detailRequirement.value.id)
    loadRequirements()
  } catch (error: any) {
    message.error(error.message || '转换失败')
  }
}

// 详情页删除
const handleDetailDelete = async () => {
  if (!detailRequirement.value) return
  try {
    await deleteRequirement(detailRequirement.value.id)
    message.success('删除成功')
    detailModalVisible.value = false
    loadRequirements()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 详情页添加备注
const handleDetailAddNote = () => {
  if (!detailRequirement.value) {
    message.warning('需求信息未加载完成，请稍候再试')
    return
  }
  detailNoteFormData.comment = ''
  detailNoteModalVisible.value = true
}

// 详情页提交备注
const handleDetailNoteSubmit = async () => {
  if (!detailRequirement.value) return
  try {
    await detailNoteFormRef.value.validate()
    await addRequirementHistoryNote(detailRequirement.value.id, { comment: detailNoteFormData.comment })
    message.success('添加备注成功')
    detailNoteModalVisible.value = false
    await loadRequirementDetailHistory(detailRequirement.value.id)
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '添加备注失败')
  }
}

// 详情页取消添加备注
const handleDetailNoteCancel = () => {
  detailNoteFormRef.value?.resetFields()
}

// 获取详情页操作描述
const getDetailActionDescription = (action: Action): string => {
  const actorName = action.actor
    ? `${action.actor.username}${action.actor.nickname ? `(${action.actor.nickname})` : ''}`
    : '系统'

  switch (action.action) {
    case 'created':
      return `由 ${actorName} 创建。`
    case 'edited':
      return `由 ${actorName} 编辑。`
    case 'status_changed':
      return `由 ${actorName} 变更状态。`
    case 'commented':
      return `由 ${actorName} 添加了备注：${action.comment || ''}`
    default:
      return `由 ${actorName} 执行了 ${action.action} 操作。`
  }
}

// 获取详情页字段显示名称
const getDetailFieldDisplayName = (fieldName: string): string => {
  const fieldNames: Record<string, string> = {
    title: '需求标题',
    description: '需求描述',
    status: '状态',
    priority: '优先级',
    project_id: '项目',
    assignee_id: '负责人',
    estimated_hours: '预估工时',
    actual_hours: '实际工时'
  }
  return fieldNames[fieldName] || fieldName
}

// 判断详情页历史记录是否有详情
const hasDetailHistoryDetails = (action: Action): boolean => {
  return !!(action.histories && action.histories.length > 0) || !!action.comment
}

// 切换详情页历史记录详情展开/收起
const toggleDetailHistoryDetail = (actionId: number) => {
  const newSet = new Set(detailExpandedHistoryIds.value)
  if (newSet.has(actionId)) {
    newSet.delete(actionId)
  } else {
    newSet.add(actionId)
  }
  detailExpandedHistoryIds.value = newSet
}

// 监听编辑模态框关闭，重新打开详情弹窗
watch(modalVisible, (visible, prevVisible) => {
  if (prevVisible && !visible && shouldKeepDetailOpen.value && detailRequirement.value) {
    shouldKeepDetailOpen.value = false
    nextTick(() => {
      detailModalVisible.value = true
      loadRequirementDetail(detailRequirement.value!.id)
      loadRequirements()
    })
  }
})

// 监听 tab 切换，切换到统计 tab 时加载统计信息
watch(activeTab, (newTab) => {
  if (newTab === 'statistics') {
    loadStatistics()
  }
})

onMounted(async () => {
  // 先加载项目列表，确保项目选择器有数据
  await loadProjects()
  loadUsers()
  
  // 读取路由查询参数（优先级高于 localStorage）
  if (route.query.project_id) {
    searchForm.project_id = Number(route.query.project_id)
  } else {
    // 从 localStorage 恢复最后选择的搜索项目
    const lastSearchProjectId = getLastSelected<number>('last_selected_requirement_project_search')
    if (lastSearchProjectId) {
      searchForm.project_id = lastSearchProjectId
    }
  }
  
  if (route.query.status) {
    searchForm.status = route.query.status as string
  }
  if (route.query.assignee === 'me' && authStore.user) {
    searchForm.assignee_id = authStore.user.id
  }
  
  // 使用 nextTick 确保项目列表已渲染后再加载需求
  nextTick(() => {
    loadRequirements()
  })
})
</script>

<style scoped>
.requirement-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.requirement-management :deep(.ant-layout) {
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

.content-inner :deep(.ant-tabs) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.content-inner :deep(.ant-tabs-content-holder) {
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.content-inner :deep(.ant-tabs-tabpane) {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  max-width: 100%;
  box-sizing: border-box;
}

.content-inner :deep(.ant-tabs-tabpane) > * {
  max-width: 100%;
  box-sizing: border-box;
}

.content-inner :deep(.ant-tabs-tabpane) .ant-row {
  margin-left: 0 !important;
  margin-right: 0 !important;
  max-width: 100%;
}

.content-inner :deep(.ant-tabs-tabpane) .ant-col {
  padding-left: 8px;
  padding-right: 8px;
  max-width: 100%;
  box-sizing: border-box;
}

.content-inner :deep(.ant-tabs-tabpane) .ant-card {
  max-width: 100%;
  box-sizing: border-box;
}

/* 优先级统计卡片对齐 - 让白色背景左边与上方卡片对齐 */
.priority-statistics-card {
  margin-left: 8px; /* 与 col 的左边 padding 对齐（gutter 的一半） */
  margin-right: 8px; /* 与 col 的右边 padding 对齐 */
}

.priority-statistics-card :deep(.ant-card-head) {
  padding-left: 16px; /* 恢复标题的左边 padding */
  padding-right: 16px;
}

.priority-statistics-card :deep(.ant-card-body) {
  padding-left: 16px; /* 恢复 body 的左边 padding，与上方卡片内容对齐 */
  padding-right: 16px;
  padding-top: 16px;
  padding-bottom: 16px;
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

