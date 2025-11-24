<template>
  <div class="init-container">
    <a-card class="init-card" :bordered="false">
      <template #title>
        <h2>系统初始化</h2>
        <p class="subtitle">欢迎使用项目管理系统，请完成初始配置</p>
      </template>
      
      <!-- 登录创建管理员 -->
      <div>
        <a-tabs v-model:activeKey="loginType" centered>
          <a-tab-pane key="wechat" tab="微信登录">
            <!-- 如果未配置微信，显示微信设置表单 -->
            <div v-if="!wechatConfigured" class="wechat-config-section">
              <a-form
                :model="wechatConfig"
                :rules="wechatRules"
                @finish="handleSaveWeChatConfig"
                layout="vertical"
              >
                <a-divider orientation="left">微信配置</a-divider>
                
                <a-alert
                  message="配置说明"
                  description="请填写微信开放平台的AppID和AppSecret。如果使用公众号，请填写公众号的AppID和AppSecret。配置完成后即可使用微信扫码登录。"
                  type="info"
                  show-icon
                  style="margin-bottom: 24px"
                />

                <a-form-item label="微信AppID" name="wechat_app_id">
                  <a-input
                    v-model:value="wechatConfig.wechat_app_id"
                    placeholder="请输入微信开放平台AppID或公众号AppID"
                    size="large"
                  />
                </a-form-item>

                <a-form-item label="微信AppSecret" name="wechat_app_secret">
                  <a-input-password
                    v-model:value="wechatConfig.wechat_app_secret"
                    placeholder="请输入微信开放平台AppSecret或公众号AppSecret"
                    size="large"
                  />
                </a-form-item>

                <a-form-item>
                  <a-button
                    type="primary"
                    html-type="submit"
                    size="large"
                    :loading="wechatConfigLoading"
                    block
                  >
                    保存配置
                  </a-button>
                </a-form-item>
              </a-form>
            </div>
            <!-- 如果已配置微信，显示扫码 -->
            <div v-else class="qr-section">
              <p class="hint">请使用微信扫码，扫码后会在微信内打开授权页面，确认后系统将自动创建管理员账号</p>
              <WeChatQRCode
                :fetchQRCode="getInitQRCode"
                initial-status-text="请使用微信扫码"
                hint="扫码后会在微信内打开授权页面"
                :show-auth-url="true"
                @success="handleInitSuccess"
              />
            </div>
          </a-tab-pane>
          <a-tab-pane key="password" tab="账号登录">
            <a-form
              :model="initForm"
              :rules="initRules"
              @finish="handlePasswordInit"
              layout="vertical"
            >
              <p class="hint">请输入管理员账号信息，系统将自动创建管理员账号</p>
              
              <a-form-item name="username" label="用户名">
                <a-input v-model:value="initForm.username" placeholder="请输入用户名" size="large" />
              </a-form-item>
              <a-form-item name="password" label="密码">
                <a-input-password v-model:value="initForm.password" placeholder="请输入密码" size="large" />
              </a-form-item>
              <a-form-item name="nickname" label="昵称">
                <a-input v-model:value="initForm.nickname" placeholder="请输入昵称" size="large" />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" html-type="submit" block size="large" :loading="loading">
                  创建管理员并登录
                </a-button>
              </a-form-item>
            </a-form>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { 
  checkInitStatus, 
  getInitQRCode, 
  initSystemWithPassword
} from '@/api/init'
import { getWeChatConfig, saveWeChatConfig, type WeChatConfigRequest } from '@/api/wechat'
import { useAuthStore } from '@/stores/auth'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const wechatConfigLoading = ref(false)
const loginType = ref('wechat') // 默认微信登录
const wechatConfigured = ref(false) // 微信是否已配置

const wechatConfig = ref<WeChatConfigRequest>({
  wechat_app_id: '',
  wechat_app_secret: ''
})

// 验证微信AppID格式
const validateWeChatAppID = (_rule: any, value: string) => {
  if (!value) {
    return Promise.reject('请输入微信AppID')
  }
  // 去除空格
  const trimmedValue = value.trim()
  if (!trimmedValue) {
    return Promise.reject('请输入微信AppID')
  }
  
  // 微信AppID格式验证：
  // 1. 公众号AppID：以wx开头，后跟16个十六进制字符（共18位）
  // 2. 开放平台AppID：通常也是18位，由字母和数字组成
  // 3. 长度检查：通常为18位，但也可能略有不同
  if (trimmedValue.length < 10 || trimmedValue.length > 32) {
    return Promise.reject('微信AppID长度应在10-32位之间')
  }
  
  // 如果以wx开头，验证格式为wx + 16位十六进制
  if (trimmedValue.startsWith('wx')) {
    const wxPattern = /^wx[a-fA-F0-9]{16}$/
    if (!wxPattern.test(trimmedValue)) {
      return Promise.reject('公众号AppID格式不正确，应为wx开头后跟16位十六进制字符（共18位）')
    }
  } else {
    // 开放平台AppID：由字母和数字组成
    const openPlatformPattern = /^[a-zA-Z0-9]+$/
    if (!openPlatformPattern.test(trimmedValue)) {
      return Promise.reject('开放平台AppID只能包含字母和数字')
    }
  }
  
  return Promise.resolve()
}

// 验证微信AppSecret格式
const validateWeChatAppSecret = (_rule: any, value: string) => {
  if (!value) {
    return Promise.reject('请输入微信AppSecret')
  }
  // 去除空格
  const trimmedValue = value.trim()
  if (!trimmedValue) {
    return Promise.reject('请输入微信AppSecret')
  }
  // AppSecret通常是32位字符（字母和数字），但长度可能略有不同
  if (trimmedValue.length < 16 || trimmedValue.length > 64) {
    return Promise.reject('微信AppSecret长度应在16-64位之间')
  }
  // AppSecret只能包含字母、数字和部分特殊字符
  const secretPattern = /^[a-zA-Z0-9_-]+$/
  if (!secretPattern.test(trimmedValue)) {
    return Promise.reject('微信AppSecret只能包含字母、数字、下划线和连字符')
  }
  return Promise.resolve()
}

const wechatRules = {
  wechat_app_id: [
    { required: true, message: '请输入微信AppID', trigger: 'blur' },
    { validator: validateWeChatAppID, trigger: 'blur' }
  ],
  wechat_app_secret: [
    { required: true, message: '请输入微信AppSecret', trigger: 'blur' },
    { validator: validateWeChatAppSecret, trigger: 'blur' }
  ]
}

const initForm = ref({
  username: '',
  password: '',
  nickname: ''
})

const initRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }]
}

// 处理初始化成功（微信登录）
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

// 处理密码登录初始化
const handlePasswordInit = async () => {
  loading.value = true
  try {
    const response = await initSystemWithPassword({
      username: initForm.value.username,
      password: initForm.value.password,
      nickname: initForm.value.nickname
    })
    
    if (response.token && response.user) {
      // 保存token和用户信息
      authStore.setToken(response.token)
      authStore.setUser(response.user)
      
      message.success('系统初始化成功！')
      
      // 跳转到工作台
      setTimeout(() => {
        router.push('/dashboard')
      }, 1000)
    }
  } catch (error: any) {
    message.error(error.message || '初始化失败')
  } finally {
    loading.value = false
  }
}

// 检查微信配置
const checkWeChatConfig = async () => {
  try {
    const config = await getWeChatConfig()
    wechatConfigured.value = !!(config.wechat_app_id && config.wechat_app_secret)
    // 如果已配置，同时更新表单数据
    if (wechatConfigured.value) {
      wechatConfig.value = {
        wechat_app_id: config.wechat_app_id || '',
        wechat_app_secret: config.wechat_app_secret || ''
      }
    }
  } catch (error: any) {
    // 如果配置不存在（404）或其他错误，认为未配置
    wechatConfigured.value = false
  }
}

// 保存微信配置
const handleSaveWeChatConfig = async () => {
  wechatConfigLoading.value = true
  try {
    // 去除空格后保存
    const configToSave = {
      wechat_app_id: wechatConfig.value.wechat_app_id.trim(),
      wechat_app_secret: wechatConfig.value.wechat_app_secret.trim()
    }
    await saveWeChatConfig(configToSave)
    message.success('微信配置保存成功')
    // 更新本地配置（去除空格后的值）
    wechatConfig.value = configToSave
    // 重新检查配置状态
    await checkWeChatConfig()
  } catch (error: any) {
    message.error(error.response?.data?.message || error.message || '保存配置失败')
  } finally {
    wechatConfigLoading.value = false
  }
}

onMounted(async () => {
  // 检查是否已经初始化
  try {
    const status = await checkInitStatus()
    if (status.initialized) {
      // 如果已初始化，跳转到登录页
      router.push('/login')
      return
    }
    // 检查微信配置
    await checkWeChatConfig()
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

.wechat-config-section {
  padding: 0 20px;
}

:deep(.ant-divider) {
  margin: 24px 0 16px 0;
}
</style>
