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
                  :scroll="{ x: 'max-content' }"
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
                  :scroll="{ x: 'max-content' }"
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
                    <template v-else-if="column.key === 'operation'">
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
      :mask-closable="true"
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
      :mask-closable="true"
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
            <a-popover
              v-model:open="iconSelectorVisible"
              trigger="click"
              placement="bottomLeft"
              :overlay-style="{ width: '600px', maxHeight: '400px', overflow: 'auto' }"
            >
              <template #content>
                <div class="icon-selector">
                  <div class="icon-grid">
                    <div
                      v-for="iconName in commonIcons"
                      :key="iconName"
                      class="icon-item"
                      :class="{ 'icon-item-selected': permissionFormData.menu_icon === iconName }"
                      @click="handleSelectIcon(iconName)"
                    >
                      <component :is="getIconComponent(iconName)" style="font-size: 20px;" />
                      <span class="icon-name">{{ iconName }}</span>
                    </div>
                  </div>
                </div>
              </template>
              <a-input
                v-model:value="permissionFormData.menu_icon"
                placeholder="点击选择图标或输入图标名称"
                readonly
                style="cursor: pointer"
                @click="iconSelectorVisible = true"
              >
                <template #prefix>
                  <component
                    v-if="permissionFormData.menu_icon"
                    :is="getIconComponent(permissionFormData.menu_icon)"
                    style="font-size: 16px; color: #1890ff;"
                  />
                </template>
              </a-input>
            </a-popover>
            <span style="margin-left: 8px; color: #999">点击输入框选择图标</span>
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
      :mask-closable="true"
      @ok="handleAssignSubmit"
      @cancel="handleAssignCancel"
      :confirm-loading="assignSubmitting"
      :width="1200"
    >
      <a-row :gutter="24">
        <!-- 左侧：全部权限树（可勾选） -->
        <a-col :span="12">
          <a-card title="权限树" size="small" :bordered="false">
            <a-spin :spinning="permissionLoading">
              <a-tree
                v-model:checkedKeys="selectedPermissionIds"
                checkable
                :tree-data="allPermissionTreeData"
                :field-names="{ children: 'children', title: 'title', key: 'key' }"
                :default-expand-all="true"
                :check-strictly="false"
                @check="handleAssignPermissionCheck"
              >
                <template #title="{ title, key, icon, isMenu }">
                  <span class="permission-tree-node">
                    <a-space>
                      <component 
                        v-if="isMenu && icon" 
                        :is="getIconComponent(icon)" 
                        style="font-size: 14px; color: #1890ff;"
                      />
                      <span>{{ title }}</span>
                    </a-space>
                    <a-space style="margin-left: 8px">
                      <a-button
                        type="link"
                        size="small"
                        @click.stop="handleEditPermissionById(key)"
                      >
                        编辑
                      </a-button>
                    </a-space>
                  </span>
                </template>
              </a-tree>
              <a-empty v-if="!permissionLoading && allPermissionTreeData.length === 0" description="暂无权限" />
            </a-spin>
          </a-card>
        </a-col>

        <!-- 右侧：动态生成的菜单树 -->
        <a-col :span="12">
          <a-card title="菜单树（根据左侧勾选动态生成）" size="small" :bordered="false">
            <a-spin :spinning="menuGenerating">
              <a-tree
                :tree-data="assignMenuTreeData"
                :field-names="{ children: 'children', title: 'title', key: 'key' }"
                :default-expand-all="true"
                :expanded-keys="expandedAssignMenuKeys"
                :show-line="{ showLeafIcon: false }"
              >
                <template #title="{ title }">
                  <span>{{ title }}</span>
                </template>
              </a-tree>
              <a-empty v-if="!menuGenerating && assignMenuTreeData.length === 0" description="请先勾选权限" />
            </a-spin>
          </a-card>
        </a-col>
      </a-row>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
// import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { PlusOutlined, MenuOutlined } from '@ant-design/icons-vue'
import * as Icons from '@ant-design/icons-vue'
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
  type Permission,
  type MenuItem
} from '@/api/permission'

// const route = useRoute()
// const router = useRouter()
const activeTab = ref('roles')

const roleLoading = ref(false)
const permissionLoading = ref(false)
const menuGenerating = ref(false)
const roleSubmitting = ref(false)
const permissionSubmitting = ref(false)
const assignSubmitting = ref(false)

const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const expandedAssignMenuKeys = ref<(string | number)[]>([])

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
  { title: '操作', key: 'operation', width: 150, fixed: 'right' as const }
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
  icon?: string      // 图标名称
  isMenu?: boolean   // 是否是菜单
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
      title: `${perm.name} (${perm.action || perm.code})`,
      icon: perm.is_menu ? (perm.menu_icon || 'MenuOutlined') : undefined,
      isMenu: perm.is_menu || false
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

// 获取图标组件
const getIconComponent = (iconName?: string) => {
  if (!iconName) return MenuOutlined
  // 从 Icons 对象中获取图标组件
  const IconComponent = (Icons as any)[iconName]
  return IconComponent || MenuOutlined
}

// 常用图标列表（用于图标选择器）
const commonIcons = [
  'DashboardOutlined', 'ProjectOutlined', 'TeamOutlined', 'SettingOutlined',
  'UserOutlined', 'FileOutlined', 'FolderOutlined', 'AppstoreOutlined',
  'DatabaseOutlined', 'BugOutlined', 'CheckCircleOutlined', 'ClockCircleOutlined',
  'CodeOutlined', 'BuildOutlined', 'ExperimentOutlined', 'BarChartOutlined',
  'CalendarOutlined', 'MailOutlined', 'BellOutlined', 'HomeOutlined',
  'SearchOutlined', 'EditOutlined', 'DeleteOutlined', 'SaveOutlined',
  'ReloadOutlined', 'DownloadOutlined', 'UploadOutlined', 'EyeOutlined',
  'LockOutlined', 'UnlockOutlined', 'KeyOutlined', 'SafetyOutlined',
  'SecurityScanOutlined', 'ToolOutlined', 'ApiOutlined', 'CloudOutlined',
  'DesktopOutlined', 'MobileOutlined', 'TabletOutlined', 'GlobalOutlined',
  'LinkOutlined', 'ShareAltOutlined', 'HeartOutlined', 'StarOutlined',
  'MessageOutlined', 'CommentOutlined', 'NotificationOutlined', 'SoundOutlined',
  'VideoCameraOutlined', 'PictureOutlined', 'FileImageOutlined', 'FilePdfOutlined',
  'FileWordOutlined', 'FileExcelOutlined', 'FileZipOutlined', 'FileTextOutlined',
  'FolderOpenOutlined', 'InboxOutlined', 'ShoppingCartOutlined', 'ShoppingOutlined',
  'GiftOutlined', 'TrophyOutlined', 'CrownOutlined', 'FireOutlined',
  'ThunderboltOutlined', 'RocketOutlined', 'CarOutlined', 'BankOutlined',
  'ShopOutlined', 'EnvironmentOutlined', 'CompassOutlined', 'FlagOutlined',
  'PushpinOutlined', 'TagsOutlined', 'TagOutlined', 'BookOutlined',
  'ReadOutlined', 'BookmarkOutlined', 'ContactsOutlined', 'IdcardOutlined',
  'SolutionOutlined', 'UsergroupAddOutlined', 'UserAddOutlined', 'UserDeleteOutlined',
  'CustomerServiceOutlined', 'QuestionCircleOutlined', 'InfoCircleOutlined',
  'ExclamationCircleOutlined', 'CloseCircleOutlined', 'CloseOutlined',
  'PlusCircleOutlined', 'MinusCircleOutlined', 'WarningOutlined',
  'MenuOutlined', 'MenuFoldOutlined', 'MenuUnfoldOutlined', 'BarsOutlined',
  'TableOutlined', 'UnorderedListOutlined', 'OrderedListOutlined', 'FilterOutlined',
  'ZoomInOutlined', 'ZoomOutOutlined', 'ExpandOutlined', 'CompressOutlined',
  'FullscreenOutlined', 'FullscreenExitOutlined', 'SortAscendingOutlined',
  'SortDescendingOutlined', 'SortOutlined', 'SyncOutlined', 'LoadingOutlined'
]

// 图标选择器相关
const iconSelectorVisible = ref(false)
const selectedIconName = ref<string>('')

// 选择图标
const handleSelectIcon = (iconName: string) => {
  selectedIconName.value = iconName
  permissionFormData.menu_icon = iconName
  iconSelectorVisible.value = false
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

// 分配权限对话框中的权限勾选变化
const handleAssignPermissionCheck = () => {
  // 菜单树会自动根据 selectedPermissionIds 更新
}

// 取消分配权限
const handleAssignCancel = () => {
  assignModalVisible.value = false
  selectedPermissionIds.value = []
  currentRoleId.value = undefined
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

// 全部权限树数据（用于左侧显示）
const allPermissionTreeData = computed<PermissionTreeNode[]>(() => {
  return buildPermissionTree(permissions.value)
})

// 分配权限对话框中的动态菜单树数据（根据勾选的权限生成）
const assignMenuTreeData = computed<MenuTreeNode[]>(() => {
  // 获取所有勾选的权限ID（过滤掉资源节点的key）
  const checkedIds = selectedPermissionIds.value
    .filter(id => typeof id === 'number')
    .map(id => id as number)
  
  // 从permissions中筛选出勾选的、且是菜单的权限
  const menuPermissions = permissions.value.filter(
    p => checkedIds.includes(p.id) && p.is_menu && p.status === 1
  )
  
  if (menuPermissions.length === 0) {
    expandedAssignMenuKeys.value = []
    return []
  }
  
  // 构建菜单树
  const menuTree = buildMenuTreeFromPermissions(menuPermissions)
  
  // 自动展开所有菜单节点
  const getAllKeys = (nodes: MenuTreeNode[]): (string | number)[] => {
    const keys: (string | number)[] = []
    nodes.forEach(node => {
      keys.push(node.key)
      if (node.children && node.children.length > 0) {
        keys.push(...getAllKeys(node.children))
      }
    })
    return keys
  }
  expandedAssignMenuKeys.value = getAllKeys(menuTree)
  
  return menuTree
})

// 菜单树节点
interface MenuTreeNode {
  key: string | number
  title: string
  id?: number
  icon?: string
  path?: string
  children?: MenuTreeNode[]
}

// 从权限列表构建菜单树
const buildMenuTreeFromPermissions = (menuPerms: Permission[]): MenuTreeNode[] => {
  // 创建权限映射
  const permMap = new Map<number, Permission>()
  menuPerms.forEach(perm => {
    permMap.set(perm.id, perm)
  })
  
  // 创建菜单项映射
  const menuMap = new Map<number, MenuTreeNode>()
  const rootMenuSet = new Set<number>()
  
  // 第一遍：创建所有菜单项
  menuPerms.forEach(perm => {
    const menuTitle = perm.menu_title || perm.name
    const menuNode: MenuTreeNode = {
      key: perm.id,
      id: perm.id,
      title: menuTitle,
      icon: perm.menu_icon,
      path: perm.menu_path,
      children: []
    }
    menuMap.set(perm.id, menuNode)
    
    if (!perm.parent_menu_id) {
      rootMenuSet.add(perm.id)
    }
  })
  
  // 第二遍：建立父子关系
  menuPerms.forEach(perm => {
    if (perm.parent_menu_id) {
      const parent = menuMap.get(perm.parent_menu_id)
      const child = menuMap.get(perm.id)
      if (parent && child) {
        parent.children = parent.children || []
        parent.children.push(child)
      } else {
        // 父菜单不在勾选列表中，作为根菜单
        rootMenuSet.add(perm.id)
      }
    }
  })
  
  // 提取根菜单并排序
  const rootMenus: MenuTreeNode[] = []
  rootMenuSet.forEach(id => {
    const menu = menuMap.get(id)
    if (menu) {
      rootMenus.push(menu)
    }
  })
  
  // 排序
  rootMenus.sort((a, b) => {
    const aPerm = menuPerms.find(p => p.id === a.id)
    const bPerm = menuPerms.find(p => p.id === b.id)
    return (aPerm?.menu_order || 0) - (bPerm?.menu_order || 0)
  })
  
  // 递归排序子菜单
  const sortChildren = (nodes: MenuTreeNode[]) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        node.children.sort((a, b) => {
          const aPerm = menuPerms.find(p => p.id === a.id)
          const bPerm = menuPerms.find(p => p.id === b.id)
          return (aPerm?.menu_order || 0) - (bPerm?.menu_order || 0)
        })
        sortChildren(node.children)
      }
    })
  }
  sortChildren(rootMenus)
  
  return rootMenus
}


// 根据ID编辑权限
const handleEditPermissionById = (permissionId: number | string) => {
  // 从permissions中找到对应的权限
  const perm = permissions.value.find(p => p.id === permissionId)
  if (perm) {
    handleEditPermission(perm)
  } else {
    message.warning('未找到对应的权限')
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

.permission-tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.permission-tree-node:hover {
  background-color: #f5f5f5;
}

.icon-selector {
  padding: 8px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px 8px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.icon-item:hover {
  border-color: #1890ff;
  background-color: #e6f7ff;
}

.icon-item-selected {
  border-color: #1890ff;
  background-color: #bae7ff;
}

.icon-name {
  margin-top: 4px;
  font-size: 11px;
  color: #666;
  text-align: center;
  word-break: break-all;
  line-height: 1.2;
}
</style>

