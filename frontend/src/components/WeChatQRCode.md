# WeChatQRCode 组件使用文档

## 概述

`WeChatQRCode` 是一个通用的微信扫码组件，封装了二维码生成、WebSocket 连接、状态管理等功能，可以在多个场景中复用。

## 功能特性

- ✅ 自动获取和生成二维码
- ✅ WebSocket 实时状态通知
- ✅ 可自定义状态文本和提示
- ✅ 支持成功/错误/信息回调
- ✅ 支持自定义消息处理
- ✅ 自动清理资源

## Props

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `fetchQRCode` | `() => Promise<QRCodeResponse>` | 必填 | 获取二维码的函数 |
| `initialStatusText` | `string` | `'请使用微信扫码'` | 初始状态文本 |
| `hint` | `string` | `'扫码后会在微信内打开授权页面'` | 提示文本 |
| `showAuthUrl` | `boolean` | `false` | 是否显示授权URL |
| `getButtonText` | `string` | `'获取二维码'` | 获取按钮文本 |
| `refreshButtonText` | `string` | `'刷新二维码'` | 刷新按钮文本 |
| `autoFetch` | `boolean` | `true` | 是否自动获取二维码 |
| `qrCodeWidth` | `number` | `200` | 二维码宽度（像素） |
| `onMessage` | `(msg: any) => void` | `undefined` | 自定义消息处理函数 |
| `onSuccess` | `(data: any) => void` | `undefined` | 成功回调 |
| `onError` | `(error: string) => void` | `undefined` | 错误回调 |
| `onInfo` | `(message: string) => void` | `undefined` | 信息回调 |

## Events

| 事件名 | 参数 | 说明 |
|--------|------|------|
| `success` | `data: any` | 扫码成功时触发 |
| `error` | `error: string` | 发生错误时触发 |
| `info` | `message: string` | 收到信息时触发 |
| `ticket` | `ticket: string` | 获取到ticket时触发 |

## 方法

通过 `ref` 可以调用以下方法：

| 方法名 | 参数 | 说明 |
|--------|------|------|
| `refresh()` | - | 刷新二维码 |
| `getTicket()` | - | 获取当前ticket |
| `getQRCodeUrl()` | - | 获取二维码图片URL |

## 使用示例

### 1. 登录页面

```vue
<template>
  <WeChatQRCode
    :fetch-qr-code="getWeChatQRCode"
    initial-status-text="请使用微信扫码"
    hint="扫码后会在微信内打开授权页面"
    :show-auth-url="false"
    @success="handleLoginSuccess"
  />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { getWeChatQRCode } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const router = useRouter()
const authStore = useAuthStore()

const handleLoginSuccess = (data: any) => {
  if (data.token && data.user) {
    authStore.setToken(data.token)
    authStore.setUser(data.user)
    router.push('/dashboard')
  }
}
</script>
```

### 2. 初始化页面

```vue
<template>
  <WeChatQRCode
    :fetch-qr-code="getInitQRCode"
    initial-status-text="请使用微信扫码"
    hint="扫码后会在微信内打开授权页面"
    :show-auth-url="true"
    @success="handleInitSuccess"
  />
</template>

<script setup lang="ts">
import { getInitQRCode } from '@/api/init'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const handleInitSuccess = (data: any) => {
  // 处理初始化成功
}
</script>
```

### 3. 扫码添加用户（示例）

```vue
<template>
  <a-modal
    v-model:open="modalVisible"
    title="扫码添加用户"
    @cancel="handleCancel"
    width="500px"
  >
    <WeChatQRCode
      ref="qrCodeRef"
      :fetch-qr-code="getAddUserQRCode"
      initial-status-text="请使用微信扫码"
      hint="扫码后会在微信内打开授权页面，确认后将添加该用户"
      :auto-fetch="true"
      @success="handleAddUserSuccess"
      @error="handleError"
    />
  </a-modal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { getWeChatQRCode } from '@/api/auth'
import { createUser } from '@/api/user'
import WeChatQRCode from '@/components/WeChatQRCode.vue'

const modalVisible = ref(false)
const qrCodeRef = ref<InstanceType<typeof WeChatQRCode>>()

// 获取添加用户的二维码（可以使用登录的API，或者创建专门的API）
const getAddUserQRCode = async () => {
  // 这里可以调用专门的API，或者复用登录API
  return await getWeChatQRCode()
}

const handleAddUserSuccess = async (data: any) => {
  if (data.user) {
    try {
      // 创建用户（如果用户不存在）
      await createUser({
        username: data.user.username,
        email: data.user.email || '',
        avatar: data.user.avatar,
        // 其他字段...
      })
      message.success('用户添加成功')
      modalVisible.value = false
      // 刷新用户列表
    } catch (error: any) {
      message.error(error.message || '添加用户失败')
    }
  }
}

const handleError = (error: string) => {
  message.error(error)
}

const handleCancel = () => {
  modalVisible.value = false
}
</script>
```

### 4. 自定义消息处理

```vue
<template>
  <WeChatQRCode
    :fetch-qr-code="getQRCode"
    :on-message="handleCustomMessage"
    @success="handleSuccess"
  />
</template>

<script setup lang="ts">
const handleCustomMessage = (msg: any) => {
  // 自定义处理所有WebSocket消息
  console.log('收到消息:', msg)
  
  // 可以在这里添加自定义逻辑
  if (msg.type === 'custom') {
    // 处理自定义消息类型
  }
}
</script>
```

## 注意事项

1. **API 函数要求**：`fetchQRCode` 函数必须返回符合 `QRCodeResponse` 接口的对象：
   ```typescript
   {
     ticket: string
     qrCodeUrl: string
     authUrl?: string
     expireSeconds: number
   }
   ```

2. **WebSocket 连接**：
   - 开发环境：自动连接到 `ws://localhost:8080/ws`
   - 生产环境：使用当前协议和域名（`wss://` 或 `ws://`）

3. **状态管理**：组件内部管理二维码状态，父组件通过事件监听结果

4. **资源清理**：组件会在 `onUnmounted` 时自动关闭 WebSocket 连接

5. **错误处理**：WebSocket 连接失败不会影响扫码流程（因为登录/初始化通过回调页面完成）

## 应用场景

- ✅ 系统初始化（扫码创建管理员）
- ✅ 用户登录（扫码登录）
- ✅ 添加用户（扫码添加新用户）
- ✅ 其他需要微信扫码的场景


