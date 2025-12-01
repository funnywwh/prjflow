<template>
  <div class="login-container">
    <a-card class="login-card" :bordered="false">
      <template #title>
        <h2>项目管理系统</h2>
      </template>
      <div class="login-content">
        <a-tabs v-model:activeKey="loginType" centered>
          <a-tab-pane key="password" tab="账号登录">
            <a-form
              :model="loginForm"
              :rules="loginRules"
              @finish="handlePasswordLogin"
              layout="vertical"
            >
              <a-form-item name="username" label="用户名">
                <a-input v-model:value="loginForm.username" placeholder="请输入用户名" size="large" />
              </a-form-item>
              <a-form-item name="password" label="密码">
                <a-input-password v-model:value="loginForm.password" placeholder="请输入密码" size="large" />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" html-type="submit" block size="large" :loading="loading">
                  登录
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>
          <a-tab-pane key="wechat" tab="微信登录">
            <WeChatQRCode
              :fetchQRCode="getWeChatQRCode"
              initial-status-text="请使用微信扫码"
              hint="扫码后会在微信内打开授权页面"
              :show-auth-url="false"
              @success="handleLoginSuccess"
            />
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { getWeChatQRCode, passwordLogin } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { usePermissionStore } from '@/stores/permission'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const router = useRouter()
const authStore = useAuthStore()
const permissionStore = usePermissionStore()

const loginType = ref('password')
const loading = ref(false)
const loginForm = ref({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// 用户名密码登录
const handlePasswordLogin = async () => {
  try {
    loading.value = true
    const response = await passwordLogin({
      username: loginForm.value.username,
      password: loginForm.value.password
    })
    handleLoginSuccess(response)
  } catch (error: any) {
    message.error(error.message || '登录失败')
  } finally {
    loading.value = false
  }
}

// 处理登录成功
const handleLoginSuccess = async (data: any) => {
  if (data.token && data.user) {
    // 保存token和用户信息
    authStore.setToken(data.token)
    // 将 is_first_login 添加到 user 对象中，以便 setUser 可以正确设置状态
    const userData = { ...data.user, is_first_login: data.is_first_login }
    authStore.setUser(userData)
    
    // 加载用户权限
    try {
      await permissionStore.loadPermissions()
    } catch (error) {
      console.error('加载权限失败:', error)
    }
    
    message.success('登录成功！')
    
    // 如果是首次登录，跳转到修改密码页面
    if (data.is_first_login) {
      setTimeout(() => {
        router.push('/auth/change-password')
      }, 1000)
    } else {
      // 否则跳转到工作台
      setTimeout(() => {
        router.push('/dashboard')
      }, 1000)
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  text-align: center;
}

.login-content {
  padding: 20px 0;
}

.qr-placeholder {
  padding: 40px;
}

.qr-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.qr-container img {
  width: 200px;
  height: 200px;
}

.hint-small {
  color: #999;
  font-size: 12px;
  margin: 0;
}

.auth-url-container {
  width: 100%;
  max-width: 400px;
  margin: 10px 0;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
  text-align: left;
}

.auth-url-label {
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

.auth-url-text {
  word-break: break-all;
  font-size: 11px;
  color: #333;
  margin: 0;
}
</style>

