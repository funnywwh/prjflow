# 配置微信 redirect_uri 环境变量

## 问题

微信扫码登录失败，提示 `redirect_uri` 和后台配置不一致。

## 原因

微信 OAuth 授权要求：
1. **`redirect_uri` 必须使用公网可访问的域名**（微信服务器需要访问）
2. **`redirect_uri` 的域名必须与微信后台配置的授权回调域名一致**

如果前端使用内网地址（如 `http://192.168.3.8:3001`），微信服务器无法访问，会导致授权失败。

## 解决方案

### 1. 创建环境变量文件

在 `frontend` 目录下创建 `.env` 文件（或 `.env.development`、`.env.production`）：

```bash
# 前端地址配置
# 必须与微信后台配置的授权回调域名一致
VITE_FRONTEND_URL=https://project.smartxy.com.cn
```

### 2. 配置说明

#### 生产环境

如果通过 nginx 访问，使用公网域名：

```bash
VITE_FRONTEND_URL=https://project.smartxy.com.cn
```

#### 开发环境

**选项 1：通过 nginx 访问（推荐）**

如果开发环境也通过 nginx 访问公网域名：

```bash
VITE_FRONTEND_URL=https://project.smartxy.com.cn
```

**选项 2：使用内网穿透**

如果直接运行 `yarn dev`，需要使用内网穿透工具（如 ngrok）：

1. 安装 ngrok：
   ```bash
   # 下载 ngrok
   # https://ngrok.com/download
   ```

2. 启动内网穿透：
   ```bash
   ngrok http 3001
   ```

3. 获取公网地址，例如：`https://abc123.ngrok.io`

4. 配置环境变量：
   ```bash
   VITE_FRONTEND_URL=https://abc123.ngrok.io
   ```

5. 在微信后台配置授权回调域名：`abc123.ngrok.io`

**选项 3：使用公网服务器**

如果开发环境部署在公网服务器上：

```bash
VITE_FRONTEND_URL=https://dev.project.smartxy.com.cn
```

### 3. 确保微信后台配置正确

1. 登录微信公众平台（或开放平台）
2. 进入"开发" → "接口权限" → "网页授权域名"
3. 配置授权回调域名，例如：`project.smartxy.com.cn`（只填域名，不填协议和路径）

### 4. 重启前端服务

修改环境变量后，需要重启前端服务：

```bash
cd frontend
yarn dev
```

## 验证

### 1. 检查环境变量

在浏览器控制台查看：

```javascript
console.log(import.meta.env.VITE_FRONTEND_URL)
```

### 2. 检查生成的 redirect_uri

在浏览器 Network 标签中，查看 `/api/auth/wechat/qrcode` 请求的 Query Parameters：

```
redirect_uri=https://project.smartxy.com.cn/auth/wechat/callback
```

### 3. 检查微信授权 URL

查看后端返回的 `auth_url`，确认 `redirect_uri` 参数是否正确：

```
https://open.weixin.qq.com/connect/oauth2/authorize?
  appid=你的应用AppID&
  redirect_uri=https%3A%2F%2Fproject.smartxy.com.cn%2Fauth%2Fwechat%2Fcallback&
  response_type=code&
  scope=snsapi_userinfo&
  state=ticket:xxx#wechat_redirect
```

解码后的 `redirect_uri` 应该是：`https://project.smartxy.com.cn/auth/wechat/callback`

## 注意事项

1. ✅ **环境变量必须以 `VITE_` 开头**，Vite 才会将其暴露给前端代码
2. ✅ **域名必须与微信后台配置一致**（不区分大小写）
3. ✅ **必须使用公网可访问的地址**，不能使用内网地址
4. ✅ **建议使用 HTTPS**（微信支持 HTTP，但 HTTPS 更安全）
5. ⚠️ **修改环境变量后需要重启前端服务**

## 故障排查

### 问题 1：环境变量未生效

**原因**：Vite 只在启动时读取环境变量

**解决**：重启前端服务

### 问题 2：仍然使用内网地址

**原因**：环境变量未配置或配置错误

**解决**：
1. 检查 `.env` 文件是否存在
2. 检查环境变量名称是否正确（`VITE_FRONTEND_URL`）
3. 检查环境变量值是否正确（公网域名）
4. 重启前端服务

### 问题 3：微信后台配置的域名不匹配

**原因**：微信后台配置的授权回调域名与 `redirect_uri` 的域名不一致

**解决**：
1. 检查微信后台配置的授权回调域名
2. 确保与 `VITE_FRONTEND_URL` 的域名一致（不包含协议和路径）

## 示例配置

### 开发环境（使用内网穿透）

```bash
# frontend/.env.development
VITE_FRONTEND_URL=https://abc123.ngrok.io
```

### 生产环境

```bash
# frontend/.env.production
VITE_FRONTEND_URL=https://project.smartxy.com.cn
```

### 本地开发（通过 nginx）

```bash
# frontend/.env.local
VITE_FRONTEND_URL=https://project.smartxy.com.cn
```

