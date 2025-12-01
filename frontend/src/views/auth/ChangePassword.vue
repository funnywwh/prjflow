<template>
  <div class="change-password-container">
    <a-card class="change-password-card" :bordered="false">
      <template #title>
        <h2>修改密码</h2>
      </template>
      <div class="change-password-content">
        <!-- 显示用户信息 -->
        <div v-if="authStore.user" class="user-info-section">
          <a-descriptions :column="1" bordered size="small">
            <a-descriptions-item label="用户名">
              {{ authStore.user.username }}
            </a-descriptions-item>
            <a-descriptions-item v-if="authStore.user.nickname" label="昵称">
              {{ authStore.user.nickname }}
            </a-descriptions-item>
          </a-descriptions>
        </div>
        <a-alert
          v-if="!hasPassword"
          message="首次登录"
          description="这是您首次登录，请设置密码以便后续使用用户名密码登录"
          type="info"
          show-icon
          :closable="false"
          style="margin-bottom: 24px"
        />
        <a-form
          :model="changePasswordForm"
          :rules="changePasswordRules"
          @finish="handleChangePasswordSubmit"
          layout="vertical"
        >
          <a-form-item
            label="旧密码"
            name="old_password"
            :required="hasPassword"
            :help="hasPassword ? '' : '您还没有设置密码，可以直接设置新密码'"
          >
            <a-input-password
              v-model:value="changePasswordForm.old_password"
              :placeholder="hasPassword ? '请输入旧密码' : '留空即可（首次设置密码）'"
              size="large"
            />
          </a-form-item>
          <a-form-item label="新密码" name="new_password" required>
            <a-input-password
              v-model:value="changePasswordForm.new_password"
              placeholder="请输入新密码（至少6位，包含大小写字母和数字）"
              size="large"
            />
            <template #help>
              <div style="font-size: 12px; color: #999;">
                密码要求：至少6位，必须包含大写字母、小写字母和数字
              </div>
            </template>
          </a-form-item>
          <a-form-item label="确认新密码" name="confirm_password" required>
            <a-input-password
              v-model:value="changePasswordForm.confirm_password"
              placeholder="请再次输入新密码"
              size="large"
            />
          </a-form-item>
          <a-form-item>
            <a-button type="primary" html-type="submit" block size="large" :loading="loading">
              确认修改
            </a-button>
          </a-form-item>
        </a-form>
      </div>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { changePassword } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const hasPassword = ref(false)
const changePasswordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

// 验证规则
const validateConfirmPassword = (_rule: any, value: string) => {
  if (!value) {
    return Promise.reject('请再次输入新密码')
  }
  if (value !== changePasswordForm.value.new_password) {
    return Promise.reject('两次输入的密码不一致')
  }
  return Promise.resolve()
}

// 验证密码强度：必须包含大小写字母和数字
const validatePasswordStrength = (_rule: any, value: string) => {
  if (!value) {
    return Promise.reject('请输入新密码')
  }
  if (value.length < 6) {
    return Promise.reject('密码长度至少6位')
  }
  const hasUpper = /[A-Z]/.test(value)
  const hasLower = /[a-z]/.test(value)
  const hasDigit = /[0-9]/.test(value)
  
  if (!hasUpper) {
    return Promise.reject('密码必须包含至少一个大写字母')
  }
  if (!hasLower) {
    return Promise.reject('密码必须包含至少一个小写字母')
  }
  if (!hasDigit) {
    return Promise.reject('密码必须包含至少一个数字')
  }
  return Promise.resolve()
}

const changePasswordRules = {
  old_password: [
    {
      validator: (_rule: any, value: string) => {
        if (hasPassword.value && !value) {
          return Promise.reject('请输入旧密码')
        }
        return Promise.resolve()
      },
      trigger: 'blur'
    }
  ],
  new_password: [
    { required: true, validator: validatePasswordStrength, trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 检查用户是否已有密码
const checkUserPassword = async () => {
  try {
    // 如果 authStore 中没有用户信息，则加载
    if (!authStore.user) {
      await authStore.loadUserInfo()
    }
    // 如果用户有密码，后端会在GetUserInfo中返回相关信息
    // 这里我们通过尝试修改密码时的错误来判断
    // 但为了简化，我们可以假设首次登录的用户没有密码
    // 实际上，首次登录的用户（LoginCount == 1）通常没有密码
    hasPassword.value = false // 首次登录用户通常没有密码
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

// 提交修改密码
const handleChangePasswordSubmit = async () => {
  if (changePasswordForm.value.new_password !== changePasswordForm.value.confirm_password) {
    message.error('两次输入的密码不一致')
    return
  }
  // 密码强度验证（表单验证规则已包含，这里作为双重检查）
  const password = changePasswordForm.value.new_password
  if (password.length < 6) {
    message.error('新密码长度至少6位')
    return
  }
  const hasUpper = /[A-Z]/.test(password)
  const hasLower = /[a-z]/.test(password)
  const hasDigit = /[0-9]/.test(password)
  if (!hasUpper || !hasLower || !hasDigit) {
    message.error('密码必须包含大写字母、小写字母和数字')
    return
  }

  try {
    loading.value = true
    await changePassword({
      old_password: changePasswordForm.value.old_password || '', // 如果没有密码，传空字符串
      new_password: changePasswordForm.value.new_password
    })
    // 清除首次登录状态
    authStore.clearFirstLogin()
    message.success('密码设置成功')
    // 跳转到工作台
    setTimeout(() => {
      router.push('/dashboard')
    }, 1000)
  } catch (error: any) {
    // 如果错误提示需要旧密码，说明用户已有密码
    if (error.message && error.message.includes('旧密码')) {
      hasPassword.value = true
      message.error(error.message || '密码修改失败')
    } else {
      message.error(error.message || '密码设置失败')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  checkUserPassword()
})
</script>

<style scoped>
.change-password-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.change-password-card {
  width: 500px;
  text-align: center;
}

.change-password-content {
  padding: 20px 0;
}

.user-info-section {
  margin-bottom: 24px;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 500;
  width: 80px;
}
</style>

