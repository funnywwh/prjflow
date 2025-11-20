<template>
  <a-layout-header class="header">
    <div class="header-left">
      <div class="logo">项目管理系统</div>
      <a-menu
        mode="horizontal"
        :selected-keys="selectedKeys"
        :style="{ lineHeight: '64px', flex: 1 }"
      >
        <template v-for="item in filteredMenu" :key="item.key">
          <!-- 有子菜单的情况 -->
          <a-sub-menu v-if="item.children && item.children.length > 0" :key="item.key">
            <template #icon>
              <component v-if="item.icon" :is="getIconComponent(item.icon)" />
            </template>
            <template #title>{{ item.title }}</template>
            <a-menu-item
              v-for="child in item.children"
              :key="child.key"
              @click="handleMenuClick(child)"
            >
              {{ child.title }}
            </a-menu-item>
          </a-sub-menu>
          <!-- 没有子菜单的情况 -->
          <a-menu-item v-else :key="item.key" @click="handleMenuClick(item)">
            <template v-if="item.icon" #icon>
              <component :is="getIconComponent(item.icon)" />
            </template>
            {{ item.title }}
          </a-menu-item>
        </template>
      </a-menu>
    </div>
    <div class="header-right">
      <a-dropdown>
        <template #overlay>
          <a-menu>
            <a-menu-item>
              <a-avatar :src="authStore.user?.avatar" :size="24" style="margin-right: 8px">
                {{ (authStore.user?.nickname || authStore.user?.username)?.charAt(0).toUpperCase() }}
              </a-avatar>
              {{ authStore.user?.username }}{{ authStore.user?.nickname ? `(${authStore.user.nickname})` : '' }}
            </a-menu-item>
            <a-menu-divider />
            <a-menu-item key="changePassword" @click="handleChangePassword">
              <template #icon><LockOutlined /></template>
              修改密码
            </a-menu-item>
            <a-menu-divider />
            <a-menu-item key="logout" @click="handleLogout">
              <template #icon><LogoutOutlined /></template>
              退出登录
            </a-menu-item>
          </a-menu>
        </template>
        <a-space class="user-info" style="cursor: pointer">
          <a-avatar :src="authStore.user?.avatar" :size="32">
            {{ (authStore.user?.nickname || authStore.user?.username)?.charAt(0).toUpperCase() }}
          </a-avatar>
          <span class="username">{{ authStore.user?.username || '用户' }}{{ authStore.user?.nickname ? `(${authStore.user.nickname})` : '' }}</span>
        </a-space>
      </a-dropdown>
    </div>

    <!-- 修改密码弹窗 -->
    <a-modal
      v-model:open="changePasswordVisible"
      title="修改密码"
      @ok="handleChangePasswordSubmit"
      :confirm-loading="changePasswordLoading"
    >
      <a-form :model="changePasswordForm" layout="vertical">
        <a-form-item 
          label="旧密码" 
          :required="hasPassword"
          :help="hasPassword ? '' : '您还没有设置密码，可以直接设置新密码'"
        >
          <a-input-password 
            v-model:value="changePasswordForm.old_password" 
            :placeholder="hasPassword ? '请输入旧密码' : '留空即可（首次设置密码）'" 
          />
        </a-form-item>
        <a-form-item label="新密码" required>
          <a-input-password v-model:value="changePasswordForm.new_password" placeholder="请输入新密码（至少6位）" />
        </a-form-item>
        <a-form-item label="确认新密码" required>
          <a-input-password v-model:value="changePasswordForm.confirm_password" placeholder="请再次输入新密码" />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-layout-header>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  LogoutOutlined,
  LockOutlined,
  DashboardOutlined,
  ProjectOutlined,
  TeamOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { usePermissionStore } from '@/stores/permission'
import { changePassword } from '@/api/auth'
import { menuConfig, filterMenuByPermission, type MenuItem as ConfigMenuItem } from '@/config/menu'
import { getMenus, type MenuItem } from '@/api/permission'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const permissionStore = usePermissionStore()

// 菜单数据
const menuItems = ref<MenuItem[]>([])
const useBackendMenu = ref(false) // 标记是否使用后端菜单

// 加载菜单
const loadMenus = async () => {
  try {
    const menus = await getMenus()
    console.log('后端返回的菜单:', menus)
    console.log('用户权限:', permissionStore.permissions)
    console.log('用户角色:', permissionStore.roles)
    // 如果后端返回了菜单（即使为空），标记为使用后端菜单
    // 但如果为空数组，则回退到静态配置
    if (menus && menus.length > 0) {
      menuItems.value = menus
      useBackendMenu.value = true
      console.log('使用后端菜单，数量:', menus.length)
    } else {
      // 后端返回空数组，使用静态配置
      menuItems.value = []
      useBackendMenu.value = false
      console.log('后端菜单为空，使用静态配置')
    }
  } catch (error) {
    console.error('加载菜单失败:', error)
    // 如果后端菜单加载失败，使用静态配置作为后备
    menuItems.value = []
    useBackendMenu.value = false
  }
}

// 将静态配置转换为菜单项（后备方案）
const convertConfigToMenuItems = (config: ConfigMenuItem[]): MenuItem[] => {
  return config.map(item => ({
    key: item.key,
    title: item.title,
    icon: item.icon,
    path: item.path,
    permission: Array.isArray(item.permission) ? item.permission[0] : item.permission,
    order: item.order,
    children: item.children ? convertConfigToMenuItems(item.children) : undefined
  }))
}

// 根据权限过滤菜单（如果使用静态配置）
const filteredMenu = computed<(MenuItem | ConfigMenuItem)[]>(() => {
  // 如果从后端加载了菜单且不为空，合并后端菜单和静态配置
  if (useBackendMenu.value && menuItems.value.length > 0) {
    // 获取静态配置菜单（根据权限过滤）
    const staticMenus = filterMenuByPermission(menuConfig, (code: string) => {
      // 如果没有权限要求，始终显示（如工作台）
      if (!code) {
        return true
      }
      return permissionStore.hasPermission(code)
    })
    console.log('静态配置菜单（过滤后）:', staticMenus)
    
    // 创建后端菜单的key集合（包括所有子菜单的key）
    const backendMenuKeys = new Set<string>()
    const collectKeys = (menus: MenuItem[]) => {
      menus.forEach(menu => {
        backendMenuKeys.add(menu.key)
        if (menu.children) {
          collectKeys(menu.children)
        }
      })
    }
    collectKeys(menuItems.value)
    console.log('后端菜单keys:', Array.from(backendMenuKeys))
    
    // 从静态配置中找出不在后端菜单中的菜单（包括所有菜单，不仅仅是基础菜单）
    const additionalMenus = staticMenus.filter(menu => {
      // 检查菜单及其子菜单是否在后端菜单中
      const checkMenuInBackend = (m: ConfigMenuItem): boolean => {
        if (backendMenuKeys.has(m.key)) {
          return true
        }
        if (m.children) {
          return m.children.some(child => checkMenuInBackend(child))
        }
        return false
      }
      return !checkMenuInBackend(menu)
    })
    console.log('补充的静态菜单:', additionalMenus)
    
    // 合并菜单：静态配置的菜单在前，后端菜单在后
    // 这样可以确保即使后端没有配置某些菜单，静态配置的菜单也能显示
    const result = [...additionalMenus, ...menuItems.value]
    console.log('最终菜单:', result)
    return result
  }
  // 否则使用静态配置并过滤
  // 注意：filterMenuByPermission 函数已经处理了没有权限要求的情况（如工作台）
  const result = filterMenuByPermission(menuConfig, (code: string) => permissionStore.hasPermission(code))
  console.log('使用静态配置菜单:', result)
  console.log('用户权限列表:', permissionStore.permissions)
  console.log('用户角色列表:', permissionStore.roles)
  console.log('是否有权限 project:read:', permissionStore.hasPermission('project:read'))
  return result
})

// 获取当前选中的菜单项（包括子菜单）
const selectedKeys = computed(() => {
  const routeName = route.name as string
  const routePath = route.path
  const keys: string[] = []
  
  // 查找匹配的菜单项
  const findMenu = (items: (MenuItem | ConfigMenuItem)[], targetPath: string, targetName: string): string | undefined => {
    for (const item of items) {
      const itemPath = 'path' in item ? item.path : undefined
      // 检查路径或key是否匹配
      if (itemPath === targetPath || item.key === targetName) {
        keys.push(item.key)
        return item.key
      }
      if (item.children) {
        const found = findMenu(item.children, targetPath, targetName)
        if (found) {
          keys.push(item.key) // 添加父菜单
          return item.key
        }
      }
    }
    return undefined
  }
  
  findMenu(filteredMenu.value, routePath, routeName)
  return keys.length > 0 ? keys : [routeName]
})

// 获取图标组件
const getIconComponent = (iconName: string) => {
  const iconMap: Record<string, any> = {
    DashboardOutlined,
    ProjectOutlined,
    TeamOutlined,
    SettingOutlined
  }
  return iconMap[iconName] || null
}

// 处理菜单点击
const handleMenuClick = (item: { path?: string }) => {
  if (item.path) {
    router.push(item.path)
  }
}

const changePasswordVisible = ref(false)
const changePasswordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})
const changePasswordLoading = ref(false)

// 修改密码
const handleChangePassword = () => {
  changePasswordVisible.value = true
  changePasswordForm.value = {
    old_password: '',
    new_password: '',
    confirm_password: ''
  }
}

// 检查用户是否已有密码（通过检查是否有wechat_open_id但没有密码来判断）
// 注意：这个判断不准确，我们改为允许旧密码为空，后端会判断
const hasPassword = ref(true) // 默认假设有密码，如果后端返回需要旧密码的错误，再提示

const handleChangePasswordSubmit = async () => {
  if (changePasswordForm.value.new_password !== changePasswordForm.value.confirm_password) {
    message.error('两次输入的密码不一致')
    return
  }
  if (changePasswordForm.value.new_password.length < 6) {
    message.error('新密码长度至少6位')
    return
  }

  try {
    changePasswordLoading.value = true
    await changePassword({
      old_password: changePasswordForm.value.old_password || '', // 如果没有密码，传空字符串
      new_password: changePasswordForm.value.new_password
    })
    message.success('密码设置成功')
    changePasswordVisible.value = false
  } catch (error: any) {
    // 如果错误提示需要旧密码，说明用户已有密码
    if (error.message && error.message.includes('旧密码')) {
      hasPassword.value = true
      message.error(error.message || '密码修改失败')
    } else {
      message.error(error.message || '密码设置失败')
    }
  } finally {
    changePasswordLoading.value = false
  }
}

// 退出登录
const handleLogout = async () => {
  try {
    await authStore.logout()
    permissionStore.clearPermissions() // 清空权限
    message.success('退出登录成功')
    router.push('/login')
  } catch (error: any) {
    message.error(error.message || '退出登录失败')
  }
}

// 加载用户信息和权限
onMounted(async () => {
  if (!authStore.user && authStore.isAuthenticated) {
    await authStore.loadUserInfo()
  }
  // 加载用户权限
  if (authStore.isAuthenticated) {
    await permissionStore.loadPermissions()
    // 加载菜单
    await loadMenus()
  }
})
</script>

<style scoped>
.header {
  background: #001529;
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.logo {
  color: white;
  font-size: 20px;
  font-weight: bold;
  margin-right: 24px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  color: white;
}

.username {
  color: white;
  margin-left: 8px;
}
</style>

