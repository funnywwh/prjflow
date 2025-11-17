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
            <p>请使用微信扫码</p>
            <p class="hint-small">扫码后会在微信内打开授权页面</p>
            <a-button @click="getQRCode">刷新二维码</a-button>
          </div>
        </a-spin>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'
import { getWeChatQRCode } from '@/api/auth'

const loading = ref(false)
const qrCodeUrl = ref('')
const ticket = ref('')
let pollTimer: number | null = null

const getQRCode = async () => {
  loading.value = true
  try {
    const data = await getWeChatQRCode()
    ticket.value = data.ticket
    
    // 将授权URL转换为二维码图片
    if (data.authUrl || data.qrCodeUrl) {
      const authUrl = data.authUrl || data.qrCodeUrl
      try {
        // 使用 qrcode 库生成二维码图片
        const qrCodeDataURL = await QRCode.toDataURL(authUrl, {
          width: 200,
          margin: 2
        })
        qrCodeUrl.value = qrCodeDataURL
      } catch (qrError) {
        console.error('生成二维码失败:', qrError)
        message.error('生成二维码图片失败')
        // 如果生成失败，显示授权URL
        qrCodeUrl.value = authUrl
      }
    } else {
      message.error('未获取到授权URL')
    }
    
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

.hint-small {
  color: #999;
  font-size: 12px;
  margin: 0;
}
</style>

