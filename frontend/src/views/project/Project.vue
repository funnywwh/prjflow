<template>
  <div class="project-management">
    <a-layout>
      <a-layout-header class="header">
        <div class="logo">项目管理系统</div>
        <a-menu
          mode="horizontal"
          :selected-keys="selectedKeys"
          :style="{ lineHeight: '64px' }"
        >
          <a-menu-item key="dashboard" @click="$router.push('/dashboard')">
            工作台
          </a-menu-item>
          <a-menu-item key="user" @click="$router.push('/user')">
            用户管理
          </a-menu-item>
          <a-menu-item key="permission" @click="$router.push('/permission')">
            权限管理
          </a-menu-item>
          <a-menu-item key="department" @click="$router.push('/department')">
            部门管理
          </a-menu-item>
          <a-menu-item key="product" @click="$router.push('/product')">
            产品管理
          </a-menu-item>
          <a-menu-item key="project" @click="$router.push('/project')">
            项目管理
          </a-menu-item>
        </a-menu>
      </a-layout-header>
      <a-layout-content class="content">
        <div class="content-inner">
          <a-tabs v-model:activeKey="activeTab">
            <!-- 项目集管理 -->
            <a-tab-pane key="projectGroups" tab="项目集管理">
              <a-page-header title="项目集管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreateProjectGroup">
                    <template #icon><PlusOutlined /></template>
                    新增项目集
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false" style="margin-bottom: 16px">
                <a-form layout="inline" :model="projectGroupSearchForm">
                  <a-form-item label="关键词">
                    <a-input
                      v-model:value="projectGroupSearchForm.keyword"
                      placeholder="项目集名称/描述"
                      allow-clear
                      style="width: 200px"
                    />
                  </a-form-item>
                  <a-form-item>
                    <a-button type="primary" @click="handleSearchProjectGroup">查询</a-button>
                    <a-button style="margin-left: 8px" @click="handleResetProjectGroup">重置</a-button>
                  </a-form-item>
                </a-form>
              </a-card>

              <a-card :bordered="false">
                <a-table
                  :columns="projectGroupColumns"
                  :data-source="projectGroups"
                  :loading="projectGroupLoading"
                  row-key="id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? '正常' : '禁用' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleEditProjectGroup(record)">
                          编辑
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个项目集吗？"
                          @confirm="handleDeleteProjectGroup(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <!-- 项目管理 -->
            <a-tab-pane key="projects" tab="项目管理">
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
                  <a-form-item label="项目集">
                    <a-select
                      v-model:value="projectSearchForm.project_group_id"
                      placeholder="选择项目集"
                      allow-clear
                      style="width: 200px"
                    >
                      <a-select-option
                        v-for="group in projectGroups"
                        :key="group.id"
                        :value="group.id"
                      >
                        {{ group.name }}
                      </a-select-option>
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
                    <template v-else-if="column.key === 'project_group'">
                      {{ record.project_group?.name || '-' }}
                    </template>
                    <template v-else-if="column.key === 'product'">
                      {{ record.product?.name || '-' }}
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
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
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-layout-content>
    </a-layout>

    <!-- 项目集编辑对话框 -->
    <a-modal
      v-model:open="projectGroupModalVisible"
      :title="projectGroupModalTitle"
      @ok="handleProjectGroupSubmit"
      @cancel="handleProjectGroupCancel"
      :confirm-loading="projectGroupSubmitting"
    >
      <a-form
        ref="projectGroupFormRef"
        :model="projectGroupFormData"
        :rules="projectGroupFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="项目集名称" name="name">
          <a-input v-model:value="projectGroupFormData.name" placeholder="请输入项目集名称" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="projectGroupFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="projectGroupFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 项目编辑对话框 -->
    <a-modal
      v-model:open="projectModalVisible"
      :title="projectModalTitle"
      @ok="handleProjectSubmit"
      @cancel="handleProjectCancel"
      :confirm-loading="projectSubmitting"
      width="800px"
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
        <a-form-item label="项目集" name="project_group_id">
          <a-select
            v-model:value="projectFormData.project_group_id"
            placeholder="选择项目集"
          >
            <a-select-option
              v-for="group in projectGroups"
              :key="group.id"
              :value="group.id"
            >
              {{ group.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关联产品">
          <a-select
            v-model:value="projectFormData.product_id"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import {
  getProjectGroups,
  createProjectGroup,
  updateProjectGroup,
  deleteProjectGroup,
  getProjects,
  createProject,
  updateProject,
  deleteProject,
  getProjectMembers,
  addProjectMembers,
  updateProjectMember,
  removeProjectMember,
  type ProjectGroup,
  type Project,
  type ProjectMember,
  type CreateProjectGroupRequest,
  type CreateProjectRequest
} from '@/api/project'
import { getProducts } from '@/api/product'
import { getUsers, type User } from '@/api/user'

const route = useRoute()
const router = useRouter()
const selectedKeys = ref([route.name as string])
const activeTab = ref('projectGroups')

const projectGroupLoading = ref(false)
const projectLoading = ref(false)
const memberLoading = ref(false)
const projectGroupSubmitting = ref(false)
const projectSubmitting = ref(false)

const projectGroups = ref<ProjectGroup[]>([])
const projects = ref<Project[]>([])
const products = ref<any[]>([])
const users = ref<User[]>([])
const projectMembers = ref<ProjectMember[]>([])
const currentProjectId = ref<number>()

const projectGroupSearchForm = reactive({
  keyword: ''
})

const projectSearchForm = reactive({
  keyword: '',
  project_group_id: undefined as number | undefined
})

const projectPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: (total: number) => `共 ${total} 条`,
  showSizeChanger: true,
  showQuickJumper: true
})

const projectGroupColumns = [
  { title: '项目集名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

const projectColumns = [
  { title: '项目名称', dataIndex: 'name', key: 'name' },
  { title: '项目编码', dataIndex: 'code', key: 'code' },
  { title: '项目集', key: 'project_group', width: 120 },
  { title: '关联产品', key: 'product', width: 120 },
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

const projectGroupModalVisible = ref(false)
const projectGroupModalTitle = ref('新增项目集')
const projectGroupFormRef = ref()
const projectGroupFormData = reactive<CreateProjectGroupRequest & { id?: number }>({
  name: '',
  description: '',
  status: 1
})

const projectGroupFormRules = {
  name: [{ required: true, message: '请输入项目集名称', trigger: 'blur' }]
}

const projectModalVisible = ref(false)
const projectModalTitle = ref('新增项目')
const projectFormRef = ref()
const projectFormData = reactive<CreateProjectRequest & { id?: number; start_date?: Dayjs; end_date?: Dayjs }>({
  name: '',
  code: '',
  description: '',
  status: 1,
  project_group_id: 0
})

const projectFormRules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入项目编码', trigger: 'blur' }],
  project_group_id: [{ required: true, message: '请选择项目集', trigger: 'change' }]
}

const memberModalVisible = ref(false)
const selectedUserIds = ref<number[]>([])
const memberRole = ref('member')

// 加载项目集列表
const loadProjectGroups = async () => {
  projectGroupLoading.value = true
  try {
    projectGroups.value = await getProjectGroups()
  } catch (error: any) {
    message.error(error.message || '加载项目集列表失败')
  } finally {
    projectGroupLoading.value = false
  }
}

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
    if (projectSearchForm.project_group_id) {
      params.project_group_id = projectSearchForm.project_group_id
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

// 加载产品列表
const loadProducts = async () => {
  try {
    const response = await getProducts()
    products.value = response.list || []
  } catch (error: any) {
    console.error('加载产品列表失败:', error)
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

// 项目集搜索
const handleSearchProjectGroup = () => {
  loadProjectGroups()
}

// 项目集重置
const handleResetProjectGroup = () => {
  projectGroupSearchForm.keyword = ''
  loadProjectGroups()
}

// 项目搜索
const handleSearchProject = () => {
  projectPagination.current = 1
  loadProjects()
}

// 项目重置
const handleResetProject = () => {
  projectSearchForm.keyword = ''
  projectSearchForm.project_group_id = undefined
  handleSearchProject()
}

// 项目表格变化
const handleProjectTableChange = (pag: any) => {
  projectPagination.current = pag.current
  projectPagination.pageSize = pag.pageSize
  loadProjects()
}

// 新增项目集
const handleCreateProjectGroup = () => {
  projectGroupModalTitle.value = '新增项目集'
  Object.assign(projectGroupFormData, {
    name: '',
    description: '',
    status: 1
  })
  delete projectGroupFormData.id
  projectGroupModalVisible.value = true
}

// 编辑项目集
const handleEditProjectGroup = (record: ProjectGroup) => {
  projectGroupModalTitle.value = '编辑项目集'
  Object.assign(projectGroupFormData, {
    id: record.id,
    name: record.name,
    description: record.description || '',
    status: record.status
  })
  projectGroupModalVisible.value = true
}

// 提交项目集
const handleProjectGroupSubmit = async () => {
  try {
    await projectGroupFormRef.value.validate()
    projectGroupSubmitting.value = true

    if (projectGroupFormData.id) {
      await updateProjectGroup(projectGroupFormData.id, projectGroupFormData)
      message.success('更新成功')
    } else {
      await createProjectGroup(projectGroupFormData)
      message.success('创建成功')
    }

    projectGroupModalVisible.value = false
    loadProjectGroups()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    projectGroupSubmitting.value = false
  }
}

// 取消项目集
const handleProjectGroupCancel = () => {
  projectGroupModalVisible.value = false
  projectGroupFormRef.value?.resetFields()
}

// 删除项目集
const handleDeleteProjectGroup = async (id: number) => {
  try {
    await deleteProjectGroup(id)
    message.success('删除成功')
    loadProjectGroups()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 新增项目
const handleCreateProject = () => {
  projectModalTitle.value = '新增项目'
  Object.assign(projectFormData, {
    name: '',
    code: '',
    description: '',
    status: 1,
    project_group_id: projectGroups.value[0]?.id || 0
  })
  delete projectFormData.id
  projectFormData.start_date = undefined
  projectFormData.end_date = undefined
  projectModalVisible.value = true
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
    project_group_id: record.project_group_id,
    product_id: record.product_id
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
      project_group_id: projectFormData.project_group_id
    }
    if (projectFormData.product_id) {
      data.product_id = projectFormData.product_id
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

onMounted(() => {
  loadProjectGroups()
  loadProjects()
  loadProducts()
  loadUsers()
})
</script>

<style scoped>
.project-management {
  min-height: 100vh;
}

.header {
  background: #001529;
  color: white;
  display: flex;
  align-items: center;
  padding: 0 24px;
}

.logo {
  color: white;
  font-size: 20px;
  font-weight: bold;
  margin-right: 24px;
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

