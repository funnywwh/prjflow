<template>
  <div class="callback-container">
    <a-spin :spinning="loading" tip="正在处理微信授权...">
      <div class="callback-content">
        <p>{{ messageText }}</p>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { addUserByWeChat } from '@/api/user'

const route = useRoute()

const loading = ref(true)
const messageText = ref('正在处理微信授权...')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code) {
    message.error('未获取到授权码')
    messageText.value = '授权失败：未获取到授权码'
    loading.value = false
    return
  }

  try {
    await addUserByWeChat({ code, state })
    
    message.success('用户添加成功！')
    messageText.value = '用户添加成功，请返回用户管理页面查看'
    
    // 3秒后自动关闭窗口（如果是在新窗口打开）
    setTimeout(() => {
      window.close()
    }, 3000)
  } catch (error: any) {
    message.error(error.message || '添加用户失败')
    messageText.value = '添加用户失败：' + (error.message || '未知错误')
    loading.value = false
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


