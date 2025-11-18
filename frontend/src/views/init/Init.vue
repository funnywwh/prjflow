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
        <div class="qr-section">
          <a-divider orientation="left">管理员登录</a-divider>
          <p class="hint">请使用微信扫码，扫码后会在微信内打开授权页面，确认后系统将自动创建管理员账号</p>
          
          <WeChatQRCode
            :fetchQRCode="getInitQRCode"
            initial-status-text="请使用微信扫码"
            hint="扫码后会在微信内打开授权页面"
            :show-auth-url="true"
            @success="handleInitSuccess"
          />
        </div>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { 
  checkInitStatus, 
  saveWeChatConfig, 
  getInitQRCode, 
  type WeChatConfigRequest 
} from '@/api/init'
import { useAuthStore } from '@/stores/auth'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const router = useRouter()
const authStore = useAuthStore()

const step = ref(1) // 1: 配置微信, 2: 扫码登录
const loading = ref(false)

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
  } catch (error: any) {
    message.error(error.message || '保存配置失败')
  } finally {
    loading.value = false
  }
}

// 处理初始化成功
const handleInitSuccess = (data: any) => {
  if (data.token && data.user) {
    // 保存token和用户信息
    authStore.setToken(data.token)
    authStore.setUser(data.user)
    
    message.success('系统初始化成功！')
    
    // 跳转到工作台
    setTimeout(() => {
      router.push('/dashboard')
    }, 1000)
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

:deep(.ant-divider) {
  margin: 24px 0 16px 0;
}
</style>
