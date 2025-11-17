<template>
  <div class="login-container">
    <a-card class="login-card" :bordered="false">
      <template #title>
        <h2>项目管理系统</h2>
      </template>
      <div class="login-content">
        <a-spin :spinning="loading">
          <div v-if="!qrCodeUrl" class="qr-placeholder">
            <a-button type="primary" @click="getQRCode">获取二维码</a-button>
          </div>
          <div v-else class="qr-container">
            <img :src="qrCodeUrl" alt="微信登录二维码" />
            <p>请使用微信扫码登录</p>
            <a-button @click="getQRCode">刷新二维码</a-button>
          </div>
        </a-spin>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { getWeChatQRCode, login } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const qrCodeUrl = ref('')
const ticket = ref('')
let pollTimer: number | null = null

const getQRCode = async () => {
  loading.value = true
  try {
    const data = await getWeChatQRCode()
    qrCodeUrl.value = data.qrCodeUrl
    ticket.value = data.ticket
    startPolling()
  } catch (error) {
    message.error('获取二维码失败')
  } finally {
    loading.value = false
  }
}

const startPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
  
  pollTimer = window.setInterval(async () => {
    try {
      // 这里应该调用检查登录状态的API
      // 暂时使用模拟逻辑
    } catch (error) {
      console.error('Polling error:', error)
    }
  }, 2000)
}

const handleLogin = async (code: string) => {
  try {
    const response = await login({ code })
    authStore.setToken(response.token)
    authStore.setUser(response.user)
    
    const redirect = route.query.redirect as string || '/dashboard'
    router.push(redirect)
  } catch (error) {
    message.error('登录失败')
  }
}

onMounted(() => {
  getQRCode()
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
})
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
</style>

