<template>
  <div class="init-container">
    <a-card class="init-card" :bordered="false">
      <template #title>
        <h2>系统初始化</h2>
        <p class="subtitle">欢迎使用项目管理系统，请完成初始配置</p>
      </template>
      
      <!-- 第一步：配置微信 -->
      <div v-if="step === 1">
        <a-form
          :model="wechatConfig"
          :rules="wechatRules"
          @finish="handleSaveWeChatConfig"
          layout="vertical"
        >
          <a-divider orientation="left">微信配置</a-divider>
          
          <a-form-item label="微信AppID" name="wechat_app_id">
            <a-input
              v-model:value="wechatConfig.wechat_app_id"
              placeholder="请输入微信开放平台AppID"
              size="large"
            />
          </a-form-item>

          <a-form-item label="微信AppSecret" name="wechat_app_secret">
            <a-input-password
              v-model:value="wechatConfig.wechat_app_secret"
              placeholder="请输入微信开放平台AppSecret"
              size="large"
            />
          </a-form-item>

          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              size="large"
              :loading="loading"
              block
            >
              保存配置并继续
            </a-button>
          </a-form-item>
        </a-form>
      </div>

      <!-- 第二步：扫码登录创建管理员 -->
      <div v-else-if="step === 2">
        <a-spin :spinning="loading">
          <div class="qr-section">
            <a-divider orientation="left">管理员登录</a-divider>
            <p class="hint">请使用微信扫码，扫码后会在微信内打开授权页面，确认后系统将自动创建管理员账号</p>
            
            <div v-if="!qrCodeUrl" class="qr-placeholder">
              <a-button type="primary" @click="getQRCode" :loading="qrLoading">
                获取二维码
              </a-button>
            </div>
            
            <div v-else class="qr-container">
              <img :src="qrCodeUrl" alt="微信登录二维码" />
              <p>请使用微信扫码</p>
              <p class="hint-small">扫码后会在微信内打开授权页面</p>
              <a-button @click="getQRCode" :loading="qrLoading">刷新二维码</a-button>
            </div>
          </div>
        </a-spin>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'
import { 
  checkInitStatus, 
  saveWeChatConfig, 
  getInitQRCode, 
  initSystem,
  type WeChatConfigRequest 
} from '@/api/init'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const step = ref(1) // 1: 配置微信, 2: 扫码登录
const loading = ref(false)
const qrLoading = ref(false)
const qrCodeUrl = ref('')
const ticket = ref('')
let pollTimer: number | null = null
let ws: WebSocket | null = null

const wechatConfig = ref<WeChatConfigRequest>({
  wechat_app_id: '',
  wechat_app_secret: ''
})

const wechatRules = {
  wechat_app_id: [
    { required: true, message: '请输入微信AppID', trigger: 'blur' }
  ],
  wechat_app_secret: [
    { required: true, message: '请输入微信AppSecret', trigger: 'blur' }
  ]
}

// 保存微信配置
const handleSaveWeChatConfig = async () => {
  loading.value = true
  try {
    await saveWeChatConfig(wechatConfig.value)
    message.success('微信配置保存成功')
    step.value = 2
    // 自动获取二维码
    await getQRCode()
  } catch (error: any) {
    message.error(error.message || '保存配置失败')
  } finally {
    loading.value = false
  }
}

// 获取二维码
const getQRCode = async () => {
  qrLoading.value = true
  try {
    const data = await getInitQRCode()
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
      return
    }
    
    // 建立WebSocket连接
    startWebSocket()
  } catch (error: any) {
    message.error(error.message || '获取二维码失败')
  } finally {
    qrLoading.value = false
  }
}

// WebSocket连接
let ws: WebSocket | null = null

// 建立WebSocket连接
const startWebSocket = () => {
  if (!ticket.value) {
    return
  }

  // 关闭旧连接
  if (ws) {
    ws.close()
  }

  // 建立新连接
  // 开发环境：直接连接到后端（因为WebSocket不能通过HTTP代理）
  // 生产环境：使用当前域名
  const isDev = import.meta.env.DEV
  let wsUrl: string
  
  if (isDev) {
    // 开发环境：直接连接到后端服务器
    wsUrl = `ws://localhost:8080/ws?ticket=${ticket.value}`
  } else {
    // 生产环境：使用当前协议和域名
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    wsUrl = `${protocol}//${window.location.host}/ws?ticket=${ticket.value}`
  }
  
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('WebSocket连接已建立')
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      handleWebSocketMessage(msg)
    } catch (error) {
      console.error('解析WebSocket消息失败:', error)
    }
  }

  ws.onerror = (error) => {
    console.error('WebSocket错误:', error)
    message.error('WebSocket连接错误')
  }

  ws.onclose = () => {
    console.log('WebSocket连接已关闭')
  }
}

// 处理WebSocket消息
const handleWebSocketMessage = (msg: any) => {
  if (msg.type === 'success') {
    // 初始化成功
    const data = msg.data as any
    if (data.token && data.user) {
      // 保存token和用户信息
      authStore.setToken(data.token)
      authStore.setUser(data.user)
      
      message.success(msg.message || '系统初始化成功！')
      
      // 关闭WebSocket连接
      if (ws) {
        ws.close()
        ws = null
      }
      
      // 跳转到工作台
      setTimeout(() => {
        router.push('/dashboard')
      }, 1000)
    }
  } else if (msg.type === 'error') {
    // 初始化失败
    message.error(msg.message || '初始化失败')
    loading.value = false
  } else if (msg.type === 'info') {
    // 信息提示
    message.info(msg.message || '')
  }
}

// 处理微信登录回调（当用户扫码后，微信会返回code）
// 保留用于后续实现微信回调处理
// @ts-ignore
const _handleWeChatLogin = async (code: string) => {
  loading.value = true
  try {
    const response = await initSystem({ code })
    
    // 保存token和用户信息
    authStore.setToken(response.token)
    authStore.setUser(response.user)
    
    message.success('系统初始化成功！')
    
    // 停止轮询
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    
    // 跳转到工作台
    setTimeout(() => {
      router.push('/dashboard')
    }, 1000)
  } catch (error: any) {
    message.error(error.message || '初始化失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // 检查是否已经初始化
  try {
    const status = await checkInitStatus()
    if (status.initialized) {
      // 如果已初始化，跳转到登录页
      router.push('/login')
    } else {
      // 检查是否已配置微信
      // 如果已配置，直接进入第二步
      // 这里可以添加检查逻辑
    }
  } catch (error) {
    console.error('检查初始化状态失败:', error)
  }
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
  if (ws) {
    ws.close()
    ws = null
  }
})
</script>

<style scoped>
.init-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.init-card {
  width: 100%;
  max-width: 600px;
}

.subtitle {
  color: #666;
  font-size: 14px;
  margin-top: 8px;
  margin-bottom: 0;
}

.hint {
  color: #666;
  margin-bottom: 20px;
}

.hint-small {
  color: #999;
  font-size: 12px;
  margin: 0;
}

.qr-section {
  text-align: center;
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

:deep(.ant-divider) {
  margin: 24px 0 16px 0;
}
</style>
