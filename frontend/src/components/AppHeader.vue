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
        <a-menu-item key="requirement" @click="$router.push('/requirement')">
          需求管理
        </a-menu-item>
        <a-menu-item key="bug" @click="$router.push('/bug')">
          Bug管理
        </a-menu-item>
        <a-menu-item key="task" @click="$router.push('/task')">
          任务管理
        </a-menu-item>
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
import { message, Modal } from 'ant-design-vue'
import { LogoutOutlined, LockOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { changePassword } from '@/api/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const selectedKeys = computed(() => [route.name as string])

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

