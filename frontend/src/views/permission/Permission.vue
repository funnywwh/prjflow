<template>
  <div class="permission-management">
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
        </a-menu>
      </a-layout-header>
      <a-layout-content class="content">
        <div class="content-inner">
          <a-tabs v-model:activeKey="activeTab">
            <!-- 角色管理 -->
            <a-tab-pane key="roles" tab="角色管理">
              <a-page-header title="角色管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreateRole">
                    <template #icon><PlusOutlined /></template>
                    新增角色
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false">
                <a-table
                  :columns="roleColumns"
                  :data-source="roles"
                  :loading="roleLoading"
                  row-key="id"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'status'">
                      <a-tag :color="record.status === 1 ? 'green' : 'red'">
                        {{ record.status === 1 ? '正常' : '禁用' }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'permissions'">
                      <a-tag
                        v-for="perm in record.permissions"
                        :key="perm.id"
                        style="margin-right: 4px"
                      >
                        {{ perm.name }}
                      </a-tag>
                    </template>
                    <template v-else-if="column.key === 'action'">
                      <a-space>
                        <a-button type="link" size="small" @click="handleEditRole(record)">
                          编辑
                        </a-button>
                        <a-button type="link" size="small" @click="handleAssignPermissions(record)">
                          分配权限
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个角色吗？"
                          @confirm="handleDeleteRole(record.id)"
                        >
                          <a-button type="link" size="small" danger>删除</a-button>
                        </a-popconfirm>
                      </a-space>
                    </template>
                  </template>
                </a-table>
              </a-card>
            </a-tab-pane>

            <!-- 权限管理 -->
            <a-tab-pane key="permissions" tab="权限管理">
              <a-page-header title="权限管理">
                <template #extra>
                  <a-button type="primary" @click="handleCreatePermission">
                    <template #icon><PlusOutlined /></template>
                    新增权限
                  </a-button>
                </template>
              </a-page-header>

              <a-card :bordered="false">
                <a-table
                  :columns="permissionColumns"
                  :data-source="permissions"
                  :loading="permissionLoading"
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
                        <a-button type="link" size="small" @click="handleEditPermission(record)">
                          编辑
                        </a-button>
                        <a-popconfirm
                          title="确定要删除这个权限吗？"
                          @confirm="handleDeletePermission(record.id)"
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

    <!-- 角色编辑对话框 -->
    <a-modal
      v-model:open="roleModalVisible"
      :title="roleModalTitle"
      @ok="handleRoleSubmit"
      @cancel="handleRoleCancel"
      :confirm-loading="roleSubmitting"
    >
      <a-form
        ref="roleFormRef"
        :model="roleFormData"
        :rules="roleFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="角色名称" name="name">
          <a-input v-model:value="roleFormData.name" placeholder="请输入角色名称" />
        </a-form-item>
        <a-form-item label="角色代码" name="code">
          <a-input v-model:value="roleFormData.code" placeholder="请输入角色代码" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="roleFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="roleFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 权限编辑对话框 -->
    <a-modal
      v-model:open="permissionModalVisible"
      :title="permissionModalTitle"
      @ok="handlePermissionSubmit"
      @cancel="handlePermissionCancel"
      :confirm-loading="permissionSubmitting"
    >
      <a-form
        ref="permissionFormRef"
        :model="permissionFormData"
        :rules="permissionFormRules"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
      >
        <a-form-item label="权限代码" name="code">
          <a-input v-model:value="permissionFormData.code" placeholder="请输入权限代码" />
        </a-form-item>
        <a-form-item label="权限名称" name="name">
          <a-input v-model:value="permissionFormData.name" placeholder="请输入权限名称" />
        </a-form-item>
        <a-form-item label="资源类型" name="resource">
          <a-input v-model:value="permissionFormData.resource" placeholder="请输入资源类型" />
        </a-form-item>
        <a-form-item label="操作类型" name="action">
          <a-input v-model:value="permissionFormData.action" placeholder="请输入操作类型" />
        </a-form-item>
        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="permissionFormData.description" placeholder="请输入描述" :rows="3" />
        </a-form-item>
        <a-form-item label="状态" name="status">
          <a-select v-model:value="permissionFormData.status" placeholder="选择状态">
            <a-select-option :value="1">正常</a-select-option>
            <a-select-option :value="0">禁用</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 分配权限对话框 -->
    <a-modal
      v-model:open="assignModalVisible"
      title="分配权限"
      @ok="handleAssignSubmit"
      @cancel="assignModalVisible = false"
      :confirm-loading="assignSubmitting"
    >
      <a-checkbox-group v-model:value="selectedPermissionIds" style="width: 100%">
        <a-row>
          <a-col :span="12" v-for="perm in permissions" :key="perm.id">
            <a-checkbox :value="perm.id">{{ perm.name }}</a-checkbox>
          </a-col>
        </a-row>
      </a-checkbox-group>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import {
  getRoles,
  createRole,
  updateRole,
  deleteRole,
  getPermissions,
  createPermission,
  assignRolePermissions,
  type Role,
  type Permission
} from '@/api/permission'

const route = useRoute()
const router = useRouter()
const selectedKeys = ref([route.name as string])
const activeTab = ref('roles')

const roleLoading = ref(false)
const permissionLoading = ref(false)
const roleSubmitting = ref(false)
const permissionSubmitting = ref(false)
const assignSubmitting = ref(false)

const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])

const roleColumns = [
  { title: '角色名称', dataIndex: 'name', key: 'name' },
  { title: '角色代码', dataIndex: 'code', key: 'code' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '权限', key: 'permissions', width: 300 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const }
]

const permissionColumns = [
  { title: '权限代码', dataIndex: 'code', key: 'code' },
  { title: '权限名称', dataIndex: 'name', key: 'name' },
  { title: '资源类型', dataIndex: 'resource', key: 'resource' },
  { title: '操作类型', dataIndex: 'action', key: 'action' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const }
]

const roleModalVisible = ref(false)
const roleModalTitle = ref('新增角色')
const roleFormRef = ref()
const roleFormData = reactive<Partial<Role> & { id?: number }>({
  name: '',
  code: '',
  description: '',
  status: 1
})

const roleFormRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色代码', trigger: 'blur' }]
}

const permissionModalVisible = ref(false)
const permissionModalTitle = ref('新增权限')
const permissionFormRef = ref()
const permissionFormData = reactive<Partial<Permission> & { id?: number }>({
  code: '',
  name: '',
  resource: '',
  action: '',
  description: '',
  status: 1
})

const permissionFormRules = {
  code: [{ required: true, message: '请输入权限代码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }]
}

const assignModalVisible = ref(false)
const selectedPermissionIds = ref<number[]>([])
const currentRoleId = ref<number>()

// 加载角色列表
const loadRoles = async () => {
  roleLoading.value = true
  try {
    roles.value = await getRoles()
  } catch (error: any) {
    message.error(error.message || '加载角色列表失败')
  } finally {
    roleLoading.value = false
  }
}

// 加载权限列表
const loadPermissions = async () => {
  permissionLoading.value = true
  try {
    permissions.value = await getPermissions()
  } catch (error: any) {
    message.error(error.message || '加载权限列表失败')
  } finally {
    permissionLoading.value = false
  }
}

// 新增角色
const handleCreateRole = () => {
  roleModalTitle.value = '新增角色'
  Object.assign(roleFormData, {
    name: '',
    code: '',
    description: '',
    status: 1
  })
  delete roleFormData.id
  roleModalVisible.value = true
}

// 编辑角色
const handleEditRole = (record: Role) => {
  roleModalTitle.value = '编辑角色'
  Object.assign(roleFormData, {
    id: record.id,
    name: record.name,
    code: record.code,
    description: record.description || '',
    status: record.status
  })
  roleModalVisible.value = true
}

// 提交角色
const handleRoleSubmit = async () => {
  try {
    await roleFormRef.value.validate()
    roleSubmitting.value = true
    
    if (roleFormData.id) {
      await updateRole(roleFormData.id, roleFormData)
      message.success('更新成功')
    } else {
      await createRole(roleFormData)
      message.success('创建成功')
    }
    
    roleModalVisible.value = false
    loadRoles()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    roleSubmitting.value = false
  }
}

// 取消角色
const handleRoleCancel = () => {
  roleModalVisible.value = false
  roleFormRef.value?.resetFields()
}

// 删除角色
const handleDeleteRole = async (id: number) => {
  try {
    await deleteRole(id)
    message.success('删除成功')
    loadRoles()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 分配权限
const handleAssignPermissions = (record: Role) => {
  currentRoleId.value = record.id
  selectedPermissionIds.value = record.permissions?.map(p => p.id) || []
  assignModalVisible.value = true
}

// 提交权限分配
const handleAssignSubmit = async () => {
  if (!currentRoleId.value) return
  
  try {
    assignSubmitting.value = true
    await assignRolePermissions(currentRoleId.value, selectedPermissionIds.value)
    message.success('分配成功')
    assignModalVisible.value = false
    loadRoles()
  } catch (error: any) {
    message.error(error.message || '分配失败')
  } finally {
    assignSubmitting.value = false
  }
}

// 新增权限
const handleCreatePermission = () => {
  permissionModalTitle.value = '新增权限'
  Object.assign(permissionFormData, {
    code: '',
    name: '',
    resource: '',
    action: '',
    description: '',
    status: 1
  })
  delete permissionFormData.id
  permissionModalVisible.value = true
}

// 编辑权限
const handleEditPermission = (record: Permission) => {
  permissionModalTitle.value = '编辑权限'
  Object.assign(permissionFormData, {
    id: record.id,
    code: record.code,
    name: record.name,
    resource: record.resource || '',
    action: record.action || '',
    description: record.description || '',
    status: record.status
  })
  permissionModalVisible.value = true
}

// 提交权限
const handlePermissionSubmit = async () => {
  try {
    await permissionFormRef.value.validate()
    permissionSubmitting.value = true
    
    if (permissionFormData.id) {
      // 更新权限（如果后端支持）
      message.warning('更新权限功能待实现')
    } else {
      await createPermission(permissionFormData)
      message.success('创建成功')
    }
    
    permissionModalVisible.value = false
    loadPermissions()
  } catch (error: any) {
    if (error.errorFields) {
      return
    }
    message.error(error.message || '操作失败')
  } finally {
    permissionSubmitting.value = false
  }
}

// 取消权限
const handlePermissionCancel = () => {
  permissionModalVisible.value = false
  permissionFormRef.value?.resetFields()
}

// 删除权限
const handleDeletePermission = async (id: number) => {
  try {
    // 删除权限（如果后端支持）
    message.warning('删除权限功能待实现')
    // await deletePermission(id)
    // message.success('删除成功')
    // loadPermissions()
  } catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

onMounted(() => {
  loadRoles()
  loadPermissions()
})
</script>

<style scoped>
.permission-management {
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

