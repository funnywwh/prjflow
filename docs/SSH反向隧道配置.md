# SSH 反向隧道配置指南

## 概述

当内网服务器无法被公网直接访问时，可以使用 SSH 反向隧道（-R），让内网服务器主动连接到公网服务器，建立隧道。

## 架构说明

```
内网服务器 → SSH 反向隧道 → 公网服务器 (ng.smartxy.com.cn) → nginx → 用户
```

- **内网服务器**：运行后端服务（8080）和前端服务（3001）
- **公网服务器**：nginx 服务器（ng.smartxy.com.cn）
- **SSH 反向隧道**：内网服务器主动连接到公网服务器

## 配置步骤

### 1. 在公网服务器上配置 SSH

#### 1.1 允许远程端口转发

编辑 `/etc/ssh/sshd_config`：

```bash
sudo nano /etc/ssh/sshd_config
```

确保以下配置：

```config
# 允许远程端口转发
GatewayPorts yes
# 或指定特定用户
# GatewayPorts clientspecified

# 允许 TCP 转发
AllowTcpForwarding yes
```

重启 SSH 服务：

```bash
sudo systemctl restart sshd
```

#### 1.2 配置 SSH 密钥认证（推荐）

在公网服务器上生成密钥对（如果还没有）：

```bash
# 在公网服务器上
ssh-keygen -t rsa -b 4096
```

将公钥添加到 `~/.ssh/authorized_keys`，或让内网服务器使用密钥连接。

### 2. 在内网服务器上建立反向隧道

#### 2.1 基本命令

```bash
# 在内网服务器上执行
ssh -N -R 8080:localhost:8080 user@ng.smartxy.com.cn
```

参数说明：
- `-N`：不执行远程命令，只建立隧道
- `-R 8080:localhost:8080`：反向隧道
  - `8080`：公网服务器上监听的端口
  - `localhost:8080`：内网服务器上的目标端口
- `user@ng.smartxy.com.cn`：公网服务器的 SSH 地址

#### 2.2 使用 autossh 保持连接（推荐）

安装 autossh：

```bash
# Ubuntu/Debian
sudo apt-get install autossh

# CentOS/RHEL
sudo yum install autossh
```

建立反向隧道：

```bash
# 使用 autossh 自动重连
autossh -M 20000 -N -R 8080:localhost:8080 user@ng.smartxy.com.cn
```

参数说明：
- `-M 20000`：监控端口，用于检测连接状态
- 其他参数与 ssh 相同

#### 2.3 配置为系统服务（推荐）

创建 systemd 服务文件：

```bash
sudo nano /etc/systemd/system/ssh-tunnel.service
```

内容：

```ini
[Unit]
Description=SSH Reverse Tunnel to Public Server
After=network.target

[Service]
Type=simple
User=你的用户名
ExecStart=/usr/bin/autossh -M 20000 -N -R 8080:localhost:8080 user@ng.smartxy.com.cn
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
# 重新加载 systemd
sudo systemctl daemon-reload

# 启用服务（开机自启）
sudo systemctl enable ssh-tunnel.service

# 启动服务
sudo systemctl start ssh-tunnel.service

# 查看状态
sudo systemctl status ssh-tunnel.service
```

### 3. 配置多个端口转发

如果需要同时转发多个端口（如后端 8080 和前端 3001）：

```bash
# 方式1：多个 -R 参数
autossh -M 20000 -N \
  -R 8080:localhost:8080 \
  -R 3001:localhost:3001 \
  user@ng.smartxy.com.cn

# 方式2：使用 SSH 配置文件
```

SSH 配置文件 `~/.ssh/config`：

```
Host ng-server
    HostName ng.smartxy.com.cn
    User user
    RemoteForward 8080 localhost:8080
    RemoteForward 3001 localhost:3001
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

然后使用：

```bash
autossh -M 20000 -N ng-server
```

### 4. 验证隧道

在公网服务器上测试：

```bash
# 检查端口是否监听
sudo netstat -tlnp | grep 8080
# 或
sudo ss -tlnp | grep 8080

# 测试连接
curl http://localhost:8080/health
```

### 5. 配置 nginx

nginx 配置使用本地端口：

```nginx
location /api {
    proxy_pass http://localhost:8080;
    # ... 其他配置
}
```

## 常见问题

### 问题1：连接断开

**原因**：SSH 连接不稳定或超时

**解决方案**：
1. 使用 autossh 自动重连
2. 配置 SSH 保活参数：
   ```bash
   ssh -o ServerAliveInterval=60 -o ServerAliveCountMax=3 -N -R 8080:localhost:8080 user@ng.smartxy.com.cn
   ```

### 问题2：端口已被占用

**原因**：公网服务器上的 8080 端口已被其他服务占用

**解决方案**：
1. 使用其他端口：
   ```bash
   ssh -N -R 18080:localhost:8080 user@ng.smartxy.com.cn
   ```
2. nginx 配置使用新端口：
   ```nginx
   proxy_pass http://localhost:18080;
   ```

### 问题3：权限 denied

**原因**：SSH 认证失败

**解决方案**：
1. 检查 SSH 密钥是否正确配置
2. 使用密码认证（不推荐，仅用于测试）：
   ```bash
   ssh -N -R 8080:localhost:8080 -o PreferredAuthentications=password user@ng.smartxy.com.cn
   ```

### 问题4：GatewayPorts 配置

如果需要在公网服务器上监听所有接口（0.0.0.0）而不是仅 localhost：

```bash
# 在公网服务器上配置
sudo nano /etc/ssh/sshd_config
# 设置：GatewayPorts yes
sudo systemctl restart sshd

# 在内网服务器上使用
ssh -N -R *:8080:localhost:8080 user@ng.smartxy.com.cn
```

## 安全建议

1. **使用密钥认证**：避免使用密码认证
2. **限制 SSH 访问**：配置防火墙，只允许特定 IP 访问 SSH
3. **使用非标准端口**：修改 SSH 默认端口（22）
4. **定期检查连接**：监控 SSH 隧道状态
5. **日志记录**：记录 SSH 连接日志

## 相关文档

- [Nginx 反向代理配置](./nginx反向代理配置.md)
- [微信登录配置指南](./微信登录配置指南.md)





