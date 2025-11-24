<template>
  <div class="wechat-settings">
    <a-layout>
      <!-- 如果是从初始化页面跳转过来的，不显示AppHeader -->
      <AppHeader v-if="route.query.from !== 'init'" />
      <a-layout-content class="content" :class="{ 'no-header': route.query.from === 'init' }">
        <div class="content-inner">
          <a-page-header title="微信设置">
            <template #extra>
              <a-button type="primary" @click="handleSave" :loading="loading">
                保存配置
              </a-button>
            </template>
          </a-page-header>

          <a-card :bordered="false">
            <a-form
              :model="wechatConfig"
              :rules="wechatRules"
              @finish="handleSave"
              layout="vertical"
            >
              <a-divider orientation="left">微信配置</a-divider>
              
              <a-alert
                message="配置说明"
                description="请填写微信开放平台的AppID和AppSecret。如果使用公众号，请填写公众号的AppID和AppSecret。"
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
                  :loading="loading"
                >
                  保存配置
                </a-button>
              </a-form-item>
            </a-form>
          </a-card>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { getWeChatConfig, saveWeChatConfig, type WeChatConfigRequest } from '@/api/wechat'
import AppHeader from '@/components/AppHeader.vue'
import { checkInitStatus } from '@/api/init'

const router = useRouter()
const route = useRoute()

const loading = ref(false)

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

// 加载微信配置
const loadWeChatConfig = async () => {
  loading.value = true
  try {
    const config = await getWeChatConfig()
    wechatConfig.value = {
      wechat_app_id: config.wechat_app_id || '',
      wechat_app_secret: config.wechat_app_secret || ''
    }
  } catch (error: any) {
    // 如果配置不存在，使用空值（允许创建新配置）
    if (error.response?.status !== 404) {
      message.error('加载配置失败: ' + (error.response?.data?.message || error.message))
    }
  } finally {
    loading.value = false
  }
}

// 保存微信配置
const handleSave = async () => {
  loading.value = true
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
    
    // 如果是从初始化页面跳转过来的，返回初始化页面
    if (route.query.from === 'init') {
      setTimeout(() => {
        router.push({
          path: '/init',
          query: { wechatConfigured: 'true' }
        })
      }, 1000)
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || error.message || '保存配置失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // 检查系统是否已初始化
  try {
    const status = await checkInitStatus()
    // 如果系统未初始化，且不是从初始化页面跳转过来的，跳转到初始化页面
    if (!status.initialized && route.query.from !== 'init') {
      router.push('/init')
      return
    }
    // 加载微信配置
    await loadWeChatConfig()
  } catch (error) {
    console.error('检查初始化状态失败:', error)
    // 如果检查失败，尝试加载配置（可能是网络问题）
    loadWeChatConfig()
  }
})
</script>

<style scoped>
.wechat-settings {
  min-height: 100vh;
}

.content {
  padding: 24px;
  background: #f0f2f5;
}

.content.no-header {
  padding-top: 24px;
  min-height: 100vh;
}

.content-inner {
  max-width: 1200px;
  margin: 0 auto;
  background: #fff;
  padding: 24px;
  border-radius: 8px;
}

:deep(.ant-divider) {
  margin: 24px 0 16px 0;
}
</style>

