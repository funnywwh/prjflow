<template>
  <div class="project-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <!-- 项目管理 -->
          <div>
              <a-page-header title="项目管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreateProject">
                    <template #icon><PlusOutlined /></template>
                    新增项目
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="projectSearchForm">
                  <a-form-item label="关键词">
                    <a-input
                      v-model:value="projectSearchForm.keyword"
                      placeholder="项目名称/编码"
                      allow-clear
                      style="width: 200px"
                    />
                  </a-form-item>
                  <a-form-item label="标签">
                    <a-select
                      v-model:value="projectSearchForm.tags"
                      mode="multiple"
                      placeholder="选择标签（支持多选）"
                      allow-clear
                      style="width: 300px"
                      :options="tagOptions"
                      :field-names="{ label: 'name', value: 'id' }"
                    >
                    </a-select>
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleSearchProject">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleResetProject">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :columns="projectColumns"
                  :data-source="projects"
                  :loading="projectLoading"
                  :pagination="projectPagination"
                  @change="handleProjectTableChange"
                  row-key="id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? '正常' : '禁用' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'tags'">
                      <div v-if="record.tags && record.tags.length > 0" style="display: flex; flex-wrap: wrap; gap: 4px;">
                        <a-tag v-for="tag in record.tags" :key="tag.id" :color="tag.color || 'blue'" style="margin: 0;">
                          {{ tag.name }}
                        </a-tag>
                      </div>
                      <span v-else>-</span>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleViewDetail(record)">
                          详情
                        </a-button>
                        <a-button type="link" size="small" @click="handleEditProject(record)">
                          编辑
                        </a-button>
                        <a-button type="link" size="small" @click="handleManageMembers(record)">
                          成员管理
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个项目吗？"
                          @confirm="handleDeleteProject(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
          </div>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 项目编辑对话框 -->
    <a-modal
      v-model:open="projectModalVisible"
      :title="projectModalTitle"
      @ok="handleProjectSubmit"
      @cancel="handleProjectCancel"
      :confirm-loading="projectSubmitting"
      width="800px"
      :mask-closable="false"
    >
      <a-form
        ref="projectFormRef"
        :model="projectFormData"
        :rules="projectFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="项目名称" name="name">
          <a-input v-model:value="projectFormData.name" placeholder="请输入项目名称" />
        </a-form-item>
        <a-form-item label="项目编码" name="code">
          <a-input v-model:value="projectFormData.code" placeholder="请输入项目编码" />
        </a-form-item>
        <a-form-item label="标签">
          <a-space style="width: 100%" direction="vertical">
            <a-select
              v-model:value="projectFormData.tag_ids"
              mode="multiple"
              placeholder="选择标签（支持多选）"
              allow-clear
              :options="tagOptions"
              :field-names="{ label: 'name', value: 'id' }"
              :filter-option="false"
              :show-search="true"
              @search="handleTagSearch"
              @dropdown-visible-change="handleTagDropdownVisibleChange"
            >
              <template #notFoundContent>
                <div style="padding: 8px; text-align: center;">
                  <a-button type="link" size="small" @click="handleCreateNewTag">
                    <template #icon><PlusOutlined /></template>
                    创建标签 "{{ tagSearchKeyword }}"
                  </a-button>
                </div>
              </template>
            </a-select>
            <a-button type="link" size="small" @click="handleOpenTagManageModal" style="padding: 0;">
              <template #icon><PlusOutlined /></template>
              管理标签
            </a-button>
          </a-space>
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始日期">
              <a-date-picker
                v-model:value="projectFormData.start_date"
                placeholder="选择开始日期"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束日期">
              <a-date-picker
                v-model:value="projectFormData.end_date"
                placeholder="选择结束日期"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="projectFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="projectFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 项目成员管理对话框 -->
    <a-modal
      v-model:open="memberModalVisible"
      title="项目成员管理"
      @cancel="handleCloseMemberModal"
      @ok="handleCloseMemberModal"
      ok-text="关闭"
      width="800px"
    >
      <a-spin :spinning="memberLoading">
        <div style="margin-bottom: 16px">
          <a-space>
            <a-select
              v-model:value="selectedUserIds"
              mode="multiple"
              placeholder="选择用户"
              style="width: 300px"
            >
              <a-select-option
                v-for="user in users"
                :key="user.id"
                :value="user.id"
              >
                {{ user.username }}
              </a-select-option>
            </a-select>
            <a-select
              v-model:value="memberRole"
              placeholder="选择角色"
              style="width: 150px"
            >
              <a-select-option value="owner">负责人</a-select-option>
              <a-select-option value="member">成员</a-select-option>
              <a-select-option value="viewer">查看者</a-select-option>
            </a-select>
            <a-button type="primary" @click="handleAddMembers">添加成员</a-button>
          </a-space>
        </div>
        <a-table
          :columns="memberColumns"
          :data-source="projectMembers"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'user'">
              {{ record.user?.username || '-' }}
            </template>
            <template v-else-if="column.key === 'role'">
              <a-select
                :value="record.role"
                @change="(value) => handleUpdateMemberRole(record.id, value)"
                style="width: 120px"
              >
                <a-select-option value="owner">负责人</a-select-option>
                <a-select-option value="member">成员</a-select-option>
                <a-select-option value="viewer">查看者</a-select-option>
              </a-select>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-popconfirm
                title="确定要移除这个成员吗？"
                @confirm="handleRemoveMember(record.id)"
              >
                <a-button type="link" size="small" danger>移除</a-button>
              </a-popconfirm>
            </template>
          </template>
        </a-table>
      </a-spin>
    </a-modal>

    <!-- 标签管理对话框 -->
    <a-modal
      v-model:open="tagManageModalVisible"
      title="创建标签"
      @ok="handleTagManageSubmit"
      @cancel="tagManageModalVisible = false"
      :confirm-loading="tagSubmitting"
      width="600px"
    >
      <a-form
        ref="tagFormRef"
        :model="tagFormData"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="标签名称" name="name" :rules="[{ required: true, message: '请输入标签名称' }]">
          <a-input v-model:value="tagFormData.name" placeholder="请输入标签名称" />
        </a-form-item>
        <a-form-item label="标签描述" name="description">
          <a-textarea v-model:value="tagFormData.description" placeholder="请输入标签描述" :rows="2" />
        </a-form-item>
        <a-form-item label="标签颜色" name="color">
          <a-select v-model:value="tagFormData.color" placeholder="选择颜色">
            <a-select-option value="blue">蓝色</a-select-option>
            <a-select-option value="green">绿色</a-select-option>
            <a-select-option value="red">红色</a-select-option>
            <a-select-option value="orange">橙色</a-select-option>
            <a-select-option value="purple">紫色</a-select-option>
            <a-select-option value="cyan">青色</a-select-option>
            <a-select-option value="magenta">品红色</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import {
  getProjects,
  createProject,
  updateProject,
  deleteProject,
  getProjectMembers,
  addProjectMembers,
  updateProjectMember,
  removeProjectMember,
  type Project,
  type ProjectMember,
  type CreateProjectRequest
} from '@/api/project'
import { getUsers, type User } from '@/api/user'
import { getTags, createTag, type Tag } from '@/api/tag'

const route = useRoute()
const router = useRouter()

const projectLoading = ref(false)
const memberLoading = ref(false)
const projectSubmitting = ref(false)

const projects = ref<Project[]>([])
const users = ref<User[]>([])
const projectMembers = ref<ProjectMember[]>([])
const currentProjectId = ref<number>()

const projectSearchForm = reactive({
  keyword: '',
  tags: [] as number[] // 改为标签ID数组
})

const tags = ref<Tag[]>([])
const tagOptions = computed(() => tags.value.map(tag => ({ label: tag.name, value: tag.id, color: tag.color })))

const projectPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const projectColumns = [
  { title: '项目名称', dataIndex: 'name', key: 'name' },
  { title: '项目编码', dataIndex: 'code', key: 'code' },
  { title: '标签', key: 'tags', width: 200 },
  { title: '开始日期', dataIndex: 'start_date', key: 'start_date', width: 120 },
  { title: '结束日期', dataIndex: 'end_date', key: 'end_date', width: 120 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const memberColumns = [
  { title: '用户', key: 'user', width: 150 },
  { title: '角色', key: 'role', width: 150 },
  { title: '操作', key: 'action', width: 100 }
]

const projectModalVisible = ref(false)
const projectModalTitle = ref('新增项目')
const projectFormRef = ref()
const projectFormData = reactive<CreateProjectRequest & { id?: number; start_date?: Dayjs; end_date?: Dayjs }>({
  name: '',
  code: '',
  description: '',
  status: 1,
  tag_ids: [] as number[] // 改为标签ID数组
})

const projectFormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入项目编码', trigger: 'blur' }]
}

const memberModalVisible = ref(false)
const selectedUserIds = ref<number[]>([])
const memberRole = ref('member')

// 标签管理相关
const tagManageModalVisible = ref(false)
const tagSubmitting = ref(false)
const tagFormRef = ref()
const tagFormData = reactive({
  name: '',
  description: '',
  color: 'blue'
})
const tagSearchKeyword = ref('')

// 加载项目列表
const loadProjects = async () => {
  projectLoading.value = true
  try {
    const params: any = {
      page: projectPagination.current,
      size: projectPagination.pageSize
    }
    if (projectSearchForm.keyword) {
      params.keyword = projectSearchForm.keyword
    }
    if (projectSearchForm.tags && projectSearchForm.tags.length > 0) {
      params.tags = projectSearchForm.tags
    }
    const response = await getProjects(params)
    projects.value = response.list
    projectPagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载项目列表失败')
  } finally {
    projectLoading.value = false
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

// 加载标签列表
const loadTags = async () => {
  try {
    tags.value = await getTags()
  } catch (error: any) {
    console.error('加载标签列表失败:', error)
    message.error('加载标签列表失败')
  }
}

// 加载项目成员
const loadProjectMembers = async (projectId: number) => {
  memberLoading.value = true
  try {
    projectMembers.value = await getProjectMembers(projectId)
  } catch (error: any) {
    message.error(error.message || '加载项目成员失败')
  } finally {
    memberLoading.value = false
  }
}

// 项目搜索
const handleSearchProject = () => {
  projectPagination.current = 1
  loadProjects()
}

// 项目重置
const handleResetProject = () => {
  projectSearchForm.keyword = ''
  projectSearchForm.tags = [] as number[]
  handleSearchProject()
}

// 项目表格变化
const handleProjectTableChange = (pag: any) => {
  projectPagination.current = pag.current
  projectPagination.pageSize = pag.pageSize
  loadProjects()
}

// 新增项目
const handleCreateProject = () => {
  console.log('handleCreateProject 被调用')
  try {
    projectModalTitle.value = '新增项目'
    // 重置表单数据
    projectFormData.name = ''
    projectFormData.code = ''
    projectFormData.description = ''
    projectFormData.status = 1
    projectFormData.tag_ids = []
    if (projectFormData.id) {
      delete projectFormData.id
    }
    projectFormData.start_date = undefined
    projectFormData.end_date = undefined
    // 打开对话框
    console.log('设置 projectModalVisible 为 true')
    projectModalVisible.value = true
    console.log('projectModalVisible 当前值:', projectModalVisible.value)
    // 使用 nextTick 确保 DOM 更新后再重置表单
    nextTick(() => {
      if (projectFormRef.value) {
        projectFormRef.value.resetFields()
      }
    })
  } catch (error) {
    console.error('handleCreateProject 出错:', error)
    message.error('打开对话框失败: ' + (error as Error).message)
  }
}

// 查看项目详情
const handleViewDetail = (record: Project) => {
  router.push(`/project/${record.id}`)
}

// 编辑项目
const handleEditProject = (record: Project) => {
  projectModalTitle.value = '编辑项目'
  Object.assign(projectFormData, {
    id: record.id,
    name: record.name,
    code: record.code,
    description: record.description || '',
    status: record.status,
    tag_ids: record.tags ? record.tags.map(tag => tag.id) : []
  })
  if (record.start_date) {
    projectFormData.start_date = dayjs(record.start_date)
  }
  if (record.end_date) {
    projectFormData.end_date = dayjs(record.end_date)
  }
  projectModalVisible.value = true
}

// 提交项目
const handleProjectSubmit = async () => {
  try {
    await projectFormRef.value.validate()
    projectSubmitting.value = true

    const data: any = {
      name: projectFormData.name,
      code: projectFormData.code,
      description: projectFormData.description,
      status: projectFormData.status,
      tag_ids: projectFormData.tag_ids || []
    }
    if (projectFormData.start_date) {
      data.start_date = projectFormData.start_date.format('YYYY-MM-DD')
    }
    if (projectFormData.end_date) {
      data.end_date = projectFormData.end_date.format('YYYY-MM-DD')
    }

    if (projectFormData.id) {
      await updateProject(projectFormData.id, data)
      message.success('更新成功')
    } else {
      await createProject(data)
      message.success('创建成功')
    }

    projectModalVisible.value = false
    loadProjects()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    projectSubmitting.value = false
  }
}

// 取消项目
const handleProjectCancel = () => {
  projectModalVisible.value = false
  projectFormRef.value?.resetFields()
}

// 删除项目
const handleDeleteProject = async (id: number) => {
  try {
    await deleteProject(id)
    message.success('删除成功')
    loadProjects()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 管理成员
const handleManageMembers = async (record: Project) => {
  currentProjectId.value = record.id
  memberModalVisible.value = true
  selectedUserIds.value = []
  memberRole.value = 'member'
  await loadProjectMembers(record.id)
}

// 添加成员
const handleAddMembers = async () => {
  if (!currentProjectId.value || selectedUserIds.value.length === 0) {
    message.warning('请选择用户')
    return
  }
  try {
    await addProjectMembers(currentProjectId.value, {
      user_ids: selectedUserIds.value,
      role: memberRole.value
    })
    message.success('添加成功')
    selectedUserIds.value = []
    await loadProjectMembers(currentProjectId.value)
  } catch (error: any) {
    message.error(error.message || '添加失败')
  }
}

// 更新成员角色
const handleUpdateMemberRole = async (memberId: number, role: string) => {
  if (!currentProjectId.value) return
  try {
    await updateProjectMember(currentProjectId.value, memberId, role)
    message.success('更新成功')
    await loadProjectMembers(currentProjectId.value)
  } catch (error: any) {
    message.error(error.message || '更新失败')
  }
}

// 移除成员
const handleRemoveMember = async (memberId: number) => {
  if (!currentProjectId.value) return
  try {
    await removeProjectMember(currentProjectId.value, memberId)
    message.success('移除成功')
    await loadProjectMembers(currentProjectId.value)
  } catch (error: any) {
    message.error(error.message || '移除失败')
  }
}

// 关闭成员管理对话框
const handleCloseMemberModal = () => {
  memberModalVisible.value = false
  selectedUserIds.value = []
  memberRole.value = 'member'
}

// 标签搜索
const handleTagSearch = (value: string) => {
  tagSearchKeyword.value = value
}

// 标签下拉框显示/隐藏
const handleTagDropdownVisibleChange = (open: boolean) => {
  if (!open) {
    tagSearchKeyword.value = ''
  }
}

// 创建新标签（从搜索框）
const handleCreateNewTag = async () => {
  if (!tagSearchKeyword.value.trim()) {
    message.warning('请输入标签名称')
    return
  }
  
  try {
    tagSubmitting.value = true
    const newTag = await createTag({
      name: tagSearchKeyword.value.trim(),
      color: 'blue'
    })
    // 添加到标签列表
    tags.value.push(newTag)
    // 自动选中新创建的标签
    if (!projectFormData.tag_ids) {
      projectFormData.tag_ids = []
    }
    projectFormData.tag_ids.push(newTag.id)
    tagSearchKeyword.value = ''
    message.success('标签创建成功')
  } catch (error: any) {
    message.error(error.message || '创建标签失败')
  } finally {
    tagSubmitting.value = false
  }
}

// 打开标签管理对话框
const handleOpenTagManageModal = () => {
  tagFormData.name = ''
  tagFormData.description = ''
  tagFormData.color = 'blue'
  tagManageModalVisible.value = true
  nextTick(() => {
    tagFormRef.value?.resetFields()
  })
}

// 提交标签管理
const handleTagManageSubmit = async () => {
  try {
    await tagFormRef.value.validate()
    tagSubmitting.value = true
    
    const newTag = await createTag({
      name: tagFormData.name,
      description: tagFormData.description,
      color: tagFormData.color
    })
    
    // 添加到标签列表
    tags.value.push(newTag)
    // 自动选中新创建的标签
    if (!projectFormData.tag_ids) {
      projectFormData.tag_ids = []
    }
    projectFormData.tag_ids.push(newTag.id)
    
    tagManageModalVisible.value = false
    message.success('标签创建成功')
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '创建标签失败')
  } finally {
    tagSubmitting.value = false
  }
}

onMounted(() => {
  loadProjects()
  loadUsers()
  loadTags()
})
</script>

<style scoped>
.project-management {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
  min-height: calc(100vh - 64px);
}

.content-inner {
  background: white;
  padding: 24px;
  border-radius: 4px;
}
</style>

