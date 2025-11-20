<template>
  <div class="permission-management">
    <a-layout>
      <AppHeader />
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
                    <template v-else-if="column.key === 'menu'">
                      <a-tag v-if="record.is_menu" color="blue">是</a-tag>
                      <span v-else>-</span>
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
        <a-form-item label="是否显示在菜单" name="is_menu">
          <a-switch v-model:checked="permissionFormData.is_menu" />
          <span style="margin-left: 8px; color: #999">开启后，该权限对应的菜单将显示在导航栏</span>
        </a-form-item>
        <template v-if="permissionFormData.is_menu">
          <a-form-item label="菜单标题" name="menu_title">
            <a-input v-model:value="permissionFormData.menu_title" placeholder="菜单显示名称（留空则使用权限名称）" />
          </a-form-item>
          <a-form-item label="菜单路径" name="menu_path">
            <a-input v-model:value="permissionFormData.menu_path" placeholder="路由路径，如：/project" />
          </a-form-item>
          <a-form-item label="菜单图标" name="menu_icon">
            <a-input v-model:value="permissionFormData.menu_icon" placeholder="图标名称，如：ProjectOutlined" />
            <span style="margin-left: 8px; color: #999">Ant Design Vue 图标组件名称</span>
          </a-form-item>
          <a-form-item label="父菜单" name="parent_menu_id">
            <a-select
              v-model:value="permissionFormData.parent_menu_id"
              placeholder="选择父菜单（留空则为顶级菜单）"
              allow-clear
            >
              <a-select-option
                v-for="perm in permissions.filter(p => p.is_menu && p.id !== permissionFormData.id)"
                :key="perm.id"
                :value="perm.id"
              >
                {{ perm.menu_title || perm.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="菜单排序" name="menu_order">
            <a-input-number v-model:value="permissionFormData.menu_order" :min="0" placeholder="数字越小越靠前" />
          </a-form-item>
        </template>
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
      :width="600"
    >
      <a-tree
        v-model:checkedKeys="selectedPermissionIds"
        checkable
        :tree-data="permissionTreeData"
        :field-names="{ children: 'children', title: 'title', key: 'key' }"
        :default-expand-all="true"
        :check-strictly="false"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
// import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import AppHeader from '@/components/AppHeader.vue'
import {
  getRoles,
  createRole,
  updateRole,
  deleteRole,
  getPermissions,
  getRolePermissions,
  createPermission,
  updatePermission,
  deletePermission,
  assignRolePermissions,
  type Role,
  type Permission
} from '@/api/permission'

// const route = useRoute()
// const router = useRouter()
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
  status: 1,
  is_menu: false,
  menu_path: '',
  menu_icon: '',
  menu_title: '',
  parent_menu_id: undefined,
  menu_order: 0
})

const permissionFormRules = {
  code: [{ required: true, message: '请输入权限代码', trigger: 'blur' }],
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }]
}

const assignModalVisible = ref(false)
const selectedPermissionIds = ref<(number | string)[]>([])
const currentRoleId = ref<number>()

// 权限树数据
interface PermissionTreeNode {
  key: number | string
  title: string
  children?: PermissionTreeNode[]
}

const permissionTreeData = ref<PermissionTreeNode[]>([])

// 将权限列表转换为树形结构
const buildPermissionTree = (perms: Permission[]): PermissionTreeNode[] => {
  // 按资源分组
  const resourceMap = new Map<string, Permission[]>()
  
  perms.forEach(perm => {
    const resource = perm.resource || '其他'
    if (!resourceMap.has(resource)) {
      resourceMap.set(resource, [])
    }
    resourceMap.get(resource)!.push(perm)
  })
  
  // 构建树形结构
  const tree: PermissionTreeNode[] = []
  resourceMap.forEach((perms, resource) => {
    const children: PermissionTreeNode[] = perms.map(perm => ({
      key: perm.id,
      title: `${perm.name} (${perm.action || perm.code})`
    }))
    
    tree.push({
      key: `resource-${resource}`,
      title: getResourceName(resource),
      children
    })
  })
  
  // 按资源名称排序
  tree.sort((a, b) => a.title.localeCompare(b.title))
  
  return tree
}

// 获取资源的中文名称
const getResourceName = (resource: string): string => {
  const resourceNames: Record<string, string> = {
    project: '项目',
    requirement: '需求',
    bug: 'Bug',
    task: '任务',
    user: '用户',
    permission: '权限',
    department: '部门',
    resource: '资源',
    module: '模块',
    version: '版本',
    testcase: '测试用例',
    testreport: '测试报告',
    dailyreport: '日报',
    weeklyreport: '周报',
    plugin: '插件',
    entityrelation: '实体关系',
    systemconfig: '系统配置'
  }
  return resourceNames[resource] || resource
}

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
    // 构建权限树
    permissionTreeData.value = buildPermissionTree(permissions.value)
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
      await createRole({
        name: roleFormData.name || '',
        code: roleFormData.code || '',
        description: roleFormData.description,
        status: roleFormData.status
      })
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
const handleAssignPermissions = async (record: Role) => {
  currentRoleId.value = record.id
  try {
    // 获取角色的完整权限列表
    const rolePermissions = await getRolePermissions(record.id)
    selectedPermissionIds.value = rolePermissions.map(p => p.id)
  } catch (error: any) {
    // 如果获取失败，使用角色对象中的权限列表
    selectedPermissionIds.value = record.permissions?.map((p: any) => p.id) || []
  }
  assignModalVisible.value = true
}

// 提交权限分配
const handleAssignSubmit = async () => {
  if (!currentRoleId.value) return
  
  try {
    assignSubmitting.value = true
    // 过滤掉资源节点的 key（字符串类型），只保留权限ID（数字类型）
    const permissionIds = selectedPermissionIds.value
      .filter(id => typeof id === 'number')
      .map(id => id as number)
    await assignRolePermissions(currentRoleId.value, permissionIds)
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
    status: record.status,
    is_menu: record.is_menu || false,
    menu_path: record.menu_path || '',
    menu_icon: record.menu_icon || '',
    menu_title: record.menu_title || '',
    parent_menu_id: record.parent_menu_id,
    menu_order: record.menu_order || 0
  })
  permissionModalVisible.value = true
}

// 提交权限
const handlePermissionSubmit = async () => {
  try {
    await permissionFormRef.value.validate()
    permissionSubmitting.value = true
    
    if (permissionFormData.id) {
      await updatePermission(permissionFormData.id, {
        code: permissionFormData.code,
        name: permissionFormData.name,
        resource: permissionFormData.resource,
        action: permissionFormData.action,
        description: permissionFormData.description,
        status: permissionFormData.status,
        is_menu: permissionFormData.is_menu,
        menu_path: permissionFormData.menu_path,
        menu_icon: permissionFormData.menu_icon,
        menu_title: permissionFormData.menu_title,
        parent_menu_id: permissionFormData.parent_menu_id,
        menu_order: permissionFormData.menu_order
      })
      message.success('更新成功')
    } else {
      await createPermission({
        code: permissionFormData.code || '',
        name: permissionFormData.name || '',
        resource: permissionFormData.resource,
        action: permissionFormData.action,
        description: permissionFormData.description,
        status: permissionFormData.status,
        is_menu: permissionFormData.is_menu,
        menu_path: permissionFormData.menu_path,
        menu_icon: permissionFormData.menu_icon,
        menu_title: permissionFormData.menu_title,
        parent_menu_id: permissionFormData.parent_menu_id,
        menu_order: permissionFormData.menu_order
      })
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
    await deletePermission(id)
    message.success('删除成功')
    loadPermissions()
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

