<template>
  <a-layout-header class="header">
    <div class="header-left">
      <div class="logo">项目管理系统</div>
      <a-menu
        mode="horizontal"
        :selected-keys="selectedKeys"
        :style="{ lineHeight: '64px', flex: 1 }"
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
    </div>
    <div class="header-right">
      <a-dropdown>
        <template #overlay>
          <a-menu>
            <a-menu-item>
              <a-avatar :src="authStore.user?.avatar" :size="24" style="margin-right: 8px">
                {{ authStore.user?.username?.charAt(0).toUpperCase() }}
              </a-avatar>
              {{ authStore.user?.username }}
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
            {{ authStore.user?.username?.charAt(0).toUpperCase() }}
          </a-avatar>
          <span class="username">{{ authStore.user?.username || '用户' }}</span>
        </a-space>
      </a-dropdown>
    </div>
  </a-layout-header>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { LogoutOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const selectedKeys = computed(() => [route.name as string])

// 退出登录
const handleLogout = async () => {
  try {
    await authStore.logout()
    message.success('退出登录成功')
    router.push('/login')
  } catch (error: any) {
    message.error(error.message || '退出登录失败')
  }
}

// 加载用户信息
onMounted(() => {
  if (!authStore.user && authStore.isAuthenticated) {
    authStore.loadUserInfo()
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

