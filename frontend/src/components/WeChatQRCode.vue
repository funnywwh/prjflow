<template>
  <div class="wechat-qrcode">
    <a-spin :spinning="loading">
      <div v-if="!qrCodeUrl" class="qr-placeholder">
        <a-button type="primary" @click="handleGetQRCode" :loading="loading">
          {{ getButtonText }}
        </a-button>
      </div>
      <div v-else class="qr-container">
        <img :src="qrCodeUrl" alt="微信登录二维码" />
        <p class="status-text">{{ statusText }}</p>
        <p v-if="hint" class="hint-small">{{ hint }}</p>
        <div v-if="showAuthUrl && authUrl" class="auth-url-container">
          <p class="auth-url-label">连接地址：</p>
          <a-typography-paragraph 
            :copyable="{ text: authUrl }" 
            class="auth-url-text"
          >
            {{ authUrl }}
          </a-typography-paragraph>
        </div>
        <a-button @click="handleGetQRCode" :loading="loading">
          {{ refreshButtonText }}
        </a-button>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import QRCode from 'qrcode'

export interface QRCodeResponse {
  ticket: string
  qrCodeUrl: string
  authUrl?: string
  expireSeconds: number
}

const props = withDefaults(defineProps<{
  // 获取二维码的函数（在模板中使用 fetch-qr-code）
  fetchQRCode: () => Promise<QRCodeResponse>
  // 初始状态文本
  initialStatusText?: string
  // 提示文本
  hint?: string
  // 是否显示授权URL
  showAuthUrl?: boolean
  // 获取按钮文本
  getButtonText?: string
  // 刷新按钮文本
  refreshButtonText?: string
  // 是否自动获取二维码
  autoFetch?: boolean
  // 二维码宽度
  qrCodeWidth?: number
  // WebSocket消息处理函数
  onMessage?: (msg: any) => void
  // 成功回调
  onSuccess?: (data: any) => void
  // 错误回调
  onError?: (error: string) => void
  // 信息回调
  onInfo?: (message: string) => void
}>(), {
  initialStatusText: '请使用微信扫码',
  hint: '扫码后会在微信内打开授权页面',
  showAuthUrl: false,
  getButtonText: '获取二维码',
  refreshButtonText: '刷新二维码',
  autoFetch: true,
  qrCodeWidth: 200,
  onMessage: undefined,
  onSuccess: undefined,
  onError: undefined,
  onInfo: undefined
})

const emit = defineEmits<{
  success: [data: any]
  error: [error: string]
  info: [message: string]
  ticket: [ticket: string]
}>()

const loading = ref(false)
const qrCodeUrl = ref('')
const authUrl = ref('')
const ticket = ref('')
const statusText = ref(props.initialStatusText)
let ws: WebSocket | null = null

// 获取二维码
const handleGetQRCode = async () => {
  if (!props.fetchQRCode || typeof props.fetchQRCode !== 'function') {
    const errorMsg = 'fetchQRCode prop is required and must be a function'
    console.error('WeChatQRCode:', errorMsg)
    message.error('配置错误：缺少获取二维码的函数')
    return
  }
  
  loading.value = true
  statusText.value = '正在生成二维码...'
  try {
    const data = await props.fetchQRCode()
    ticket.value = data.ticket
    emit('ticket', data.ticket)
    
    // 将授权URL转换为二维码图片
    if (data.authUrl || data.qrCodeUrl) {
      const url = data.authUrl || data.qrCodeUrl
      authUrl.value = url
      try {
        // 使用 qrcode 库生成二维码图片
        const qrCodeDataURL = await QRCode.toDataURL(url, {
          width: props.qrCodeWidth,
          margin: 2
        })
        qrCodeUrl.value = qrCodeDataURL
        statusText.value = props.initialStatusText
      } catch (qrError) {
        console.error('生成二维码失败:', qrError)
        message.error('生成二维码图片失败')
        // 如果生成失败，显示授权URL
        qrCodeUrl.value = url
      }
    } else {
      message.error('未获取到授权URL')
      statusText.value = '获取二维码失败'
      return
    }
    
    // 建立WebSocket连接
    startWebSocket()
  } catch (error: any) {
    const errorMsg = error.message || '获取二维码失败'
    message.error(errorMsg)
    statusText.value = '获取二维码失败'
    emit('error', errorMsg)
    if (props.onError) {
      props.onError(errorMsg)
    }
  } finally {
    loading.value = false
  }
}

// WebSocket连接
const startWebSocket = () => {
  if (!ticket.value) {
    return
  }

  // 关闭旧连接
  if (ws) {
    ws.close()
    ws = null
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
    statusText.value = '等待扫码...'
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
    // WebSocket连接失败不影响流程，只是无法实时显示扫码状态
  }

  ws.onclose = () => {
    console.log('WebSocket连接已关闭')
  }
}

// 处理WebSocket消息
const handleWebSocketMessage = (msg: any) => {
  // 如果提供了自定义消息处理函数，优先使用
  if (props.onMessage) {
    props.onMessage(msg)
    return
  }

  // 默认消息处理
  if (msg.type === 'info') {
    // 信息提示（如：已扫码、正在授权等）
    const infoMsg = msg.message || '处理中...'
    statusText.value = infoMsg
    message.info(infoMsg)
    emit('info', infoMsg)
    if (props.onInfo) {
      props.onInfo(infoMsg)
    }
  } else if (msg.type === 'success') {
    // 成功
    const data = msg.data
    statusText.value = msg.message || '操作成功'
    message.success(msg.message || '操作成功')
    emit('success', data)
    if (props.onSuccess) {
      props.onSuccess(data)
    }
    
    // 关闭WebSocket连接
    if (ws) {
      ws.close()
      ws = null
    }
  } else if (msg.type === 'error') {
    // 错误
    const errorMsg = msg.message || '操作失败'
    message.error(errorMsg)
    statusText.value = errorMsg
    loading.value = false
    emit('error', errorMsg)
    if (props.onError) {
      props.onError(errorMsg)
    }
  }
}

// 监听状态文本变化
watch(() => props.initialStatusText, (newVal) => {
  if (!qrCodeUrl.value) {
    statusText.value = newVal
  }
})

onMounted(() => {
  if (props.autoFetch) {
    handleGetQRCode()
  }
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

// 暴露方法供父组件调用
defineExpose({
  refresh: handleGetQRCode,
  getTicket: () => ticket.value,
  getQRCodeUrl: () => qrCodeUrl.value
})
</script>

<style scoped>
.wechat-qrcode {
  width: 100%;
}

.qr-placeholder {
  padding: 40px;
  text-align: center;
}

.qr-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.qr-container img {
  width: 200px;
  height: 200px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
}

.status-text {
  font-size: 14px;
  color: #333;
  margin: 0;
  font-weight: 500;
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

