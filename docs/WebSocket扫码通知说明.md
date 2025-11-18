# WebSocket 扫码通知说明

## 概述

扫码组件 `WeChatQRCode.vue` **已实现 WebSocket 连接**，用于实时接收扫码结果通知。

## 实现情况

### ✅ 前端实现

**文件**：`frontend/src/components/WeChatQRCode.vue`

1. **WebSocket 连接**：
   - 在获取二维码后自动建立 WebSocket 连接
   - 开发环境：`ws://localhost:8080/ws?ticket=${ticket}`
   - 生产环境：`wss://${window.location.host}/ws?ticket=${ticket}`

2. **消息处理**：
   - `info` 类型：显示信息提示（如"已扫码，正在获取授权..."）
   - `success` 类型：触发成功回调，传递数据
   - `error` 类型：显示错误信息

3. **状态更新**：
   - 连接建立：显示"等待扫码..."
   - 收到消息：更新状态文本和显示提示

### ✅ 后端实现

**文件**：`backend/internal/websocket/`

1. **WebSocket 路由**：
   - 路由：`GET /ws?ticket=xxx`
   - 处理器：`websocket.HandleWebSocket`

2. **消息发送**：
   - 扫码时：`websocket.GetHub().SendMessage(ticket, "info", nil, "已扫码，正在获取授权...")`
   - 登录成功：`websocket.GetHub().SendMessage(ticket, "success", data, "登录成功")`
   - 错误时：`websocket.GetHub().SendMessage(ticket, "error", nil, "错误信息")`

## 开发环境配置

### 前端配置

**文件**：`frontend/src/components/WeChatQRCode.vue` (第 165-175 行)

```typescript
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
```

### 后端配置

**文件**：`backend/cmd/server/main.go` (第 59 行)

```go
// WebSocket路由
r.GET("/ws", websocket.HandleWebSocket)
```

## 工作流程

1. **获取二维码**：
   - 前端调用 `fetchQRCode()` 获取二维码
   - 后端返回 `ticket` 和授权 URL

2. **建立 WebSocket 连接**：
   - 前端使用 `ticket` 建立 WebSocket 连接
   - 连接 URL：`ws://localhost:8080/ws?ticket=${ticket}`

3. **扫码流程**：
   - 用户扫码 → 微信回调 → 后端处理
   - 后端在处理过程中发送 WebSocket 消息：
     - `info`: "已扫码，正在获取授权..."
     - `info`: "正在获取用户信息..."
     - `info`: "正在登录..."
     - `success`: 登录成功（包含 token 和用户信息）

4. **前端接收消息**：
   - 前端通过 WebSocket 接收消息
   - 更新 UI 状态（状态文本、提示信息）
   - 触发成功回调（如 `@success` 事件）

## 验证方法

### 1. 检查 WebSocket 连接

在浏览器开发者工具的 **Network** 标签中：
- 筛选 `WS`（WebSocket）
- 查看是否有到 `/ws?ticket=xxx` 的连接
- 连接状态应该是 `101 Switching Protocols`

### 2. 检查控制台日志

前端控制台应该看到：
```
WebSocket连接已建立
```

后端日志应该看到：
```
WebSocket连接已注册: ticket=xxx
```

### 3. 测试扫码流程

1. 打开登录页面
2. 获取二维码
3. 查看浏览器控制台，确认 WebSocket 连接已建立
4. 扫码后，应该看到：
   - 状态文本更新（"已扫码，正在获取授权..."）
   - 提示消息（Ant Design message）
   - 最终成功消息和跳转

## 可能的问题

### 1. WebSocket 连接失败

**原因**：
- 后端未启动或端口不正确
- 防火墙阻止连接
- CORS 配置问题

**解决**：
- 确认后端运行在 `localhost:8080`
- 检查后端日志是否有错误
- 检查浏览器控制台错误信息

### 2. 未收到消息

**原因**：
- `ticket` 不匹配
- 后端未发送消息
- WebSocket 连接已断开

**解决**：
- 检查后端日志，确认是否调用了 `SendMessage`
- 检查 `ticket` 是否一致
- 查看浏览器 Network 标签，确认 WebSocket 连接状态

### 3. 开发环境连接问题

如果前端运行在 `http://192.168.3.8:3001`，但 WebSocket 连接到 `ws://localhost:8080`：
- ✅ **这是正确的**：WebSocket 连接是从浏览器发起的，`localhost:8080` 指向本地后端服务器
- ⚠️ **注意**：如果后端运行在其他机器，需要修改为后端服务器的 IP 地址

## 总结

- ✅ **前端已实现 WebSocket 连接**
- ✅ **后端已实现 WebSocket 服务**
- ✅ **开发环境配置正确**（`ws://localhost:8080/ws`）
- ✅ **消息处理完整**（info、success、error）

扫码组件**已完整实现 WebSocket 通知功能**，可以实时显示扫码状态和结果。

