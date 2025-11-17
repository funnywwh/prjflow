<template>
  <div class="callback-container">
    <a-spin :spinning="loading" tip="正在处理微信登录...">
      <div class="callback-content">
        <p>{{ message }}</p>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const loading = ref(true)
const messageText = ref('正在处理微信登录...')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code) {
    message.error('未获取到授权码')
    messageText.value = '登录失败：未获取到授权码'
    loading.value = false
    setTimeout(() => {
      router.push('/login')
    }, 2000)
    return
  }

  try {
    const response = await login({ code, state })
    
    // 保存token和用户信息
    authStore.setToken(response.token)
    authStore.setUser(response.user)
    
    message.success('登录成功！')
    messageText.value = '登录成功，正在跳转...'
    
    // 跳转到工作台或之前要访问的页面
    const redirect = route.query.redirect as string || '/dashboard'
    setTimeout(() => {
      router.push(redirect)
    }, 1000)
  } catch (error: any) {
    message.error(error.message || '登录失败')
    messageText.value = '登录失败：' + (error.message || '未知错误')
    loading.value = false
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  }
})
</script>

<style scoped>
.callback-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.callback-content {
  text-align: center;
  color: white;
  font-size: 16px;
}
</style>

