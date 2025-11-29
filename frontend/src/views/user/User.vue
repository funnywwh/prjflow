<template>
  <div class="user-management">
    <a-layout>
      <AppHeader />
      <a-layout-content class="content">
        <div class="content-inner">
          <a-page-header title="用户管理">
            <template #extra>
              <a-space>
                <a-button @click="handleScanAddUser">
                  <template #icon><QrcodeOutlined /></template>
                  扫码添加用户
                </a-button>
                <a-button type="primary" @click="handleCreate">
                  <template #icon><PlusOutlined /></template>
                  新增用户
                </a-button>
              </a-space>
            </template>
          </a-page-header>

          <!-- 搜索栏 -->
          <a-card :bordered="false" style="margin-bottom: 16px">
            <a-form layout="inline" :model="searchForm">
              <a-form-item label="关键词">
                <a-input
                  v-model:value="searchForm.keyword"
                  placeholder="用户名/邮箱"
                  allow-clear
                  style="width: 200px"
                />
              </a-form-item>
              <a-form-item label="部门">
                <a-tree-select
                  v-model:value="searchForm.department_id"
                  :tree-data="departmentTreeData"
                  placeholder="选择部门"
                  allow-clear
                  style="width: 200px"
                  :field-names="{ children: 'children', label: 'name', value: 'id' }"
                  tree-default-expand-all
                  show-search
                  :tree-node-filter-prop="'name'"
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <!-- 用户列表 -->
          <a-card :bordered="false" class="table-card">
            <a-table
              :columns="columns"
              :data-source="users"
              :loading="loading"
              :scroll="{ x: 'max-content', y: tableScrollHeight }"
              :pagination="pagination"
              @change="handleTableChange"
              row-key="id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'avatar'">
                  <a-avatar :src="record.avatar" :size="40">
                    {{ (record.nickname || record.username)?.charAt(0).toUpperCase() }}
                  </a-avatar>
                </template>
                <template v-else-if="column.key === 'nickname'">
                  {{ record.username }}({{ record.nickname }})
                </template>
                <template v-else-if="column.key === 'status'">
                  <a-tag :color="record.status === 1 ? 'green' : 'red'">
                    {{ record.status === 1 ? '正常' : '禁用' }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'department'">
                  {{ record.department?.name || '-' }}
                </template>
                <template v-else-if="column.key === 'roles'">
                  <a-tag
                    v-for="role in record.roles"
                    :key="role.id"
                    style="margin-right: 4px"
                  >
                    {{ role.name }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleEdit(record)">
                      编辑
                    </a-button>
                    <a-button type="link" size="small" @click="handleAssignRoles(record)">
                      分配角色
                    </a-button>
                    <a-popconfirm
                      title="确定要删除这个用户吗？"
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

    <!-- 用户编辑对话框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :mask-closable="false"
      @ok="handleSubmit"
      @cancel="handleCancel"
      :confirm-loading="submitting"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="用户名" name="username">
          <a-input v-model:value="formData.username" placeholder="请输入用户名（用于登录）" />
        </a-form-item>
        <a-form-item label="昵称" name="nickname">
          <a-input v-model:value="formData.nickname" placeholder="请输入昵称（必填，用于显示）" />
        </a-form-item>
        <a-form-item 
          label="密码" 
          name="password"
          :rules="formData.id ? passwordRules : passwordRules"
        >
          <a-input-password 
            v-model:value="formData.password" 
            :placeholder="formData.id ? '留空则不修改密码，否则需包含大小写字母和数字' : '请输入密码（可选，需包含大小写字母和数字）'" 
          />
          <template v-if="formData.password && formData.password.length > 0" #help>
            <div style="font-size: 12px; color: #999;">
              密码要求：至少6位，必须包含大写字母、小写字母和数字
            </div>
          </template>
        </a-form-item>
        <a-form-item label="邮箱" name="email">
          <a-input v-model:value="formData.email" placeholder="请输入邮箱" />
        </a-form-item>
        <a-form-item label="手机号" name="phone">
          <a-input v-model:value="formData.phone" placeholder="请输入手机号" />
        </a-form-item>
        <a-form-item label="部门" name="department_id">
          <a-tree-select
            v-model:value="formData.department_id"
            :tree-data="departmentTreeData"
            placeholder="选择部门（树形结构）"
            allow-clear
            :field-names="{ children: 'children', label: 'name', value: 'id' }"
            tree-default-expand-all
            show-search
            :tree-node-filter-prop="'name'"
          />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="formData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 分配角色对话框 -->
    <a-modal
      v-model:open="roleModalVisible"
      title="分配角色"
      :mask-closable="true"
      @ok="handleRoleSubmit"
      @cancel="roleModalVisible = false"
      :confirm-loading="roleSubmitting"
      
      >
      <a-checkbox-group v-model:value="selectedRoleIds" style="width: 100%">
        <a-row>
          <a-col :span="12" v-for="role in roles" :key="role.id">
            <a-checkbox :value="role.id">{{ role.name }}</a-checkbox>
          </a-col>
        </a-row>
      </a-checkbox-group>
    </a-modal>

    <!-- 扫码添加用户对话框 -->
    <a-modal
      v-model:open="scanAddUserModalVisible"
      title="扫码添加用户"
      :footer="null"
      width="500px"
      :mask-closable="true"
      @cancel="handleCloseScanAddUserModal"
    >
      <WeChatQRCode
        ref="scanAddUserQRCodeRef"
        :fetchQRCode="getAddUserQRCode"
        initial-status-text="请使用微信扫码"
        hint="扫码后会在微信内打开授权页面，确认后将添加该用户"
        :auto-fetch="true"
        :show-auth-url="false"
        @success="handleScanAddUserSuccess"
        @error="handleScanAddUserError"
      />
    </a-modal>

    <!-- 修改昵称对话框 -->
    <a-modal
      v-model:open="nicknameModalVisible"
      title="设置用户昵称"
      :mask-closable="true"
      @ok="handleNicknameSubmit"
      @cancel="handleNicknameCancel"
      :confirm-loading="nicknameSubmitting"
      
      >
      <a-form
        ref="nicknameFormRef"
        :model="nicknameFormData"
        :rules="nicknameFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="用户名" name="username">
          <a-input v-model:value="nicknameFormData.username" disabled />
        </a-form-item>
        <a-form-item label="昵称" name="nickname">
          <a-input v-model:value="nicknameFormData.nickname" placeholder="请输入昵称（必填，用于前端显示）" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
// import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, QrcodeOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import WeChatQRCode from '@/components/WeChatQRCode.vue'
import {
  getUsers,
  createUser,
  updateUser,
  deleteUser,
  type User,
  type CreateUserRequest
} from '@/api/user'
import { getDepartments, type Department } from '@/api/department'
import { getRoles, assignUserRoles, type Role } from '@/api/permission'
import request from '@/utils/request'

// const route = useRoute()
// const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const roleSubmitting = ref(false)
const users = ref<User[]>([])
const departments = ref<Department[]>([])
const roles = ref<Role[]>([])

// 部门树形数据（用于树形选择器）
const departmentTreeData = computed(() => {
  return departments.value || []
})

const searchForm = reactive({
  keyword: '',
  department_id: undefined as number | undefined
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
  {
    title: '头像',
    key: 'avatar',
    width: 80
  },
  {
    title: '用户名',
    dataIndex: 'username',
    key: 'username'
  },
  {
    title: '昵称',
    dataIndex: 'nickname',
    key: 'nickname'
  },
  {
    title: '邮箱',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: '手机号',
    dataIndex: 'phone',
    key: 'phone'
  },
  {
    title: '部门',
    key: 'department',
    width: 120
  },
  {
    title: '角色',
    key: 'roles',
    width: 200
  },
  {
    title: '状态',
    key: 'status',
    width: 80
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    fixed: 'right' as const
  }
]

const modalVisible = ref(false)
const modalTitle = ref('新增用户')
const formRef = ref()
const formData = reactive<CreateUserRequest & { id?: number; password?: string }>({
  username: '',
  nickname: '',
  password: '',
  email: '',
  phone: '',
  department_id: undefined,
  status: 1
})

// 验证密码强度：必须包含大小写字母和数字
const validatePasswordStrength = (_rule: any, value: string) => {
  // 如果密码为空，且是编辑模式，则允许（留空不修改）
  if (!value || value.trim() === '') {
    return Promise.resolve()
  }
  if (value.length < 6) {
    return Promise.reject('密码长度至少6位')
  }
  const hasUpper = /[A-Z]/.test(value)
  const hasLower = /[a-z]/.test(value)
  const hasDigit = /[0-9]/.test(value)
  
  if (!hasUpper) {
    return Promise.reject('密码必须包含至少一个大写字母')
  }
  if (!hasLower) {
    return Promise.reject('密码必须包含至少一个小写字母')
  }
  if (!hasDigit) {
    return Promise.reject('密码必须包含至少一个数字')
  }
  return Promise.resolve()
}

const passwordRules = [
  {
    validator: validatePasswordStrength,
    trigger: 'blur'
  }
]

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }],
  password: passwordRules
}

const roleModalVisible = ref(false)
const selectedRoleIds = ref<number[]>([])
const currentUserId = ref<number>()

const scanAddUserModalVisible = ref(false)
const scanAddUserQRCodeRef = ref<InstanceType<typeof WeChatQRCode>>()

const nicknameModalVisible = ref(false)
const nicknameSubmitting = ref(false)
const nicknameFormRef = ref()
const nicknameFormData = reactive({
  id: 0,
  username: '',
  nickname: ''
})
const nicknameFormRules = {
  nickname: [{ required: true, message: '请输入昵称（不能为空）', trigger: 'blur' }]
}

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.current,
      size: pagination.pageSize
    }
    if (searchForm.keyword) {
      params.keyword = searchForm.keyword
    }
    if (searchForm.department_id) {
      params.department_id = searchForm.department_id
    }
    const response = await getUsers(params)
    users.value = response.list
    pagination.total = response.total
  } catch (error: any) {
    message.error(error.message || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 加载部门列表
const loadDepartments = async () => {
  try {
    departments.value = await getDepartments()
  } catch (error: any) {
    console.error('加载部门列表失败:', error)
  }
}

// 加载角色列表
const loadRoles = async () => {
  try {
    roles.value = await getRoles()
  } catch (error: any) {
    console.error('加载角色列表失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadUsers()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  searchForm.department_id = undefined
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadUsers()
}

// 新增
const handleCreate = () => {
  modalTitle.value = '新增用户'
  Object.assign(formData, {
    username: '',
    nickname: '',
    password: '',
    email: '',
    phone: '',
    department_id: undefined,
    status: 1
  })
  delete formData.id
  modalVisible.value = true
}

// 编辑
const handleEdit = (record: User) => {
  modalTitle.value = '编辑用户'
  Object.assign(formData, {
    id: record.id,
    username: record.username,
    nickname: record.nickname || '',
    password: '', // 编辑时不显示密码，留空则不修改
    email: record.email || '',
    phone: record.phone || '',
    department_id: record.department_id,
    status: record.status
  })
  modalVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    
    // 准备提交数据，如果密码为空则删除该字段
    const submitData: any = { ...formData }
    if (!submitData.password || submitData.password.trim() === '') {
      delete submitData.password
    }
    
    if (formData.id) {
      await updateUser(formData.id, submitData)
      message.success('更新成功')
    } else {
      await createUser(submitData)
      message.success('创建成功')
    }
    
    modalVisible.value = false
    loadUsers()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// 取消
const handleCancel = () => {
  modalVisible.value = false
  formRef.value?.resetFields()
}

// 删除
const handleDelete = async (id: number) => {
  try {
    await deleteUser(id)
    message.success('删除成功')
    loadUsers()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 分配角色
const handleAssignRoles = (record: User) => {
  currentUserId.value = record.id
  selectedRoleIds.value = record.roles?.map(r => r.id) || []
  roleModalVisible.value = true
}

// 提交角色分配
const handleRoleSubmit = async () => {
  if (!currentUserId.value) return
  
  try {
    roleSubmitting.value = true
    await assignUserRoles(currentUserId.value, selectedRoleIds.value)
    message.success('分配成功')
    roleModalVisible.value = false
    loadUsers()
  } catch (error: any) {
    message.error(error.message || '分配失败')
  } finally {
    roleSubmitting.value = false
  }
}

// 扫码添加用户
const handleScanAddUser = () => {
  scanAddUserModalVisible.value = true
}

// 获取添加用户的二维码（使用特殊的回调地址）
const getAddUserQRCode = async () => {
  // 后端会优先使用配置文件中的 callback_domain
  // 但添加用户的回调路径不同，需要传递 redirect_uri（只传路径部分）
  // 后端会自动使用配置的域名拼接路径
  // 注意：回调路径需要包含 /api 前缀，因为后端路由都加了 /api 前缀
  const callbackPath = '/api/auth/wechat/add-user/callback'
  
  // 调用API时传递回调路径（后端会使用配置的域名）
  const data: any = await request.get('/auth/wechat/qrcode', {
    params: { redirect_uri: callbackPath }
  })
  return {
    ticket: data.ticket || '',
    qrCodeUrl: data.qr_code_url || data.auth_url || '',
    authUrl: data.auth_url || data.qr_code_url || '',
    expireSeconds: data.expire_seconds || 600
  }
}

// 处理扫码添加用户成功
const handleScanAddUserSuccess = async (data: any) => {
  if (data.user) {
    try {
      // 用户已通过后端API创建，显示修改昵称对话框
      scanAddUserModalVisible.value = false
      nicknameFormData.id = data.user.id
      nicknameFormData.username = data.user.username
      nicknameFormData.nickname = data.user.nickname || ''
      nicknameModalVisible.value = true
    } catch (error: any) {
      message.error(error.message || '添加用户失败')
    }
  }
}

// 处理扫码添加用户错误
const handleScanAddUserError = (error: string) => {
  message.error(error)
}

// 关闭扫码添加用户对话框
const handleCloseScanAddUserModal = () => {
  scanAddUserModalVisible.value = false
}

// 提交昵称修改
const handleNicknameSubmit = async () => {
  try {
    await nicknameFormRef.value.validate()
    nicknameSubmitting.value = true
    
    await updateUser(nicknameFormData.id, {
      nickname: nicknameFormData.nickname
    })
    
    message.success('昵称设置成功')
    nicknameModalVisible.value = false
    loadUsers()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '设置昵称失败')
  } finally {
    nicknameSubmitting.value = false
  }
}

// 取消昵称修改
const handleNicknameCancel = () => {
  nicknameModalVisible.value = false
  nicknameFormRef.value?.resetFields()
  // 即使取消，也刷新列表，因为用户已经创建了
  loadUsers()
}

onMounted(() => {
  loadUsers()
  loadDepartments()
  loadRoles()
})
</script>

<style scoped>
.user-management {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.user-management :deep(.ant-layout) {
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
  background: white;
  padding: 24px;
  border-radius: 4px;
}

.table-card {
  margin-top: 16px;
}
</style>

