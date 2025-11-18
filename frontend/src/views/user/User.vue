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
                <a-select
                  v-model:value="searchForm.department_id"
                  placeholder="选择部门"
                  allow-clear
                  style="width: 200px"
                >
                  <a-select-option
                    v-for="dept in departments"
                    :key="dept.id"
                    :value="dept.id"
                  >
                    {{ dept.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleSearch">查询</a-button>
                <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
              </a-form-item>
            </a-form>
          </a-card>

          <!-- 用户列表 -->
          <a-card :bordered="false">
            <a-table
              :columns="columns"
              :data-source="users"
              :loading="loading"
              :pagination="pagination"
              @change="handleTableChange"
              row-key="id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'avatar'">
                  <a-avatar :src="record.avatar" :size="40">
                    {{ record.username?.charAt(0).toUpperCase() }}
                  </a-avatar>
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
          <a-input v-model:value="formData.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="邮箱" name="email">
          <a-input v-model:value="formData.email" placeholder="请输入邮箱" />
        </a-form-item>
        <a-form-item label="手机号" name="phone">
          <a-input v-model:value="formData.phone" placeholder="请输入手机号" />
        </a-form-item>
        <a-form-item label="部门" name="department_id">
          <a-select
            v-model:value="formData.department_id"
            placeholder="选择部门"
            allow-clear
          >
            <a-select-option
              v-for="dept in departments"
              :key="dept.id"
              :value="dept.id"
            >
              {{ dept.name }}
            </a-select-option>
          </a-select>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const roleSubmitting = ref(false)
const users = ref<User[]>([])
const departments = ref<Department[]>([])
const roles = ref<Role[]>([])

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
const formData = reactive<CreateUserRequest & { id?: number }>({
  username: '',
  email: '',
  phone: '',
  department_id: undefined,
  status: 1
})

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }]
}

const roleModalVisible = ref(false)
const selectedRoleIds = ref<number[]>([])
const currentUserId = ref<number>()

const scanAddUserModalVisible = ref(false)
const scanAddUserQRCodeRef = ref<InstanceType<typeof WeChatQRCode>>()

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
    
    if (formData.id) {
      await updateUser(formData.id, formData)
      message.success('更新成功')
    } else {
      await createUser(formData)
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
  const callbackPath = '/auth/wechat/add-user/callback'
  
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
      // 用户已通过后端API创建，这里只需要刷新列表
      message.success('用户添加成功')
      scanAddUserModalVisible.value = false
      loadUsers()
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

onMounted(() => {
  loadUsers()
  loadDepartments()
  loadRoles()
})
</script>

<style scoped>
.user-management {
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

