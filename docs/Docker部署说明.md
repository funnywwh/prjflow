# Docker 部署说明

本文档说明如何使用 Docker 部署项目管理系统。

## 概述

项目提供了完整的 Docker 支持，使用多阶段构建生成包含前端和后端的生产镜像。镜像使用纯 Go SQLite（modernc.org/sqlite），支持静态编译（CGO_ENABLED=0），无需 CGO 和系统库依赖。

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+（可选，用于 docker-compose 部署）

## 快速开始

### 方式一：使用 Docker Compose（推荐）

1. **准备数据目录**

```bash
mkdir -p data
```

2. **修改环境变量**

编辑 `docker-compose.yml`，修改 `JWT_SECRET` 环境变量：

```yaml
environment:
  - JWT_SECRET=your-secret-key-change-in-production  # 修改为你的密钥
```

3. **构建并启动**

```bash
docker-compose up -d
```

4. **查看日志**

```bash
docker-compose logs -f
```

5. **访问应用**

打开浏览器访问：http://localhost:8080

### 方式二：使用 Docker 命令

1. **构建镜像**

```bash
docker build -t prjflow:latest .
```

或者指定版本信息：

```bash
docker build \
  --build-arg VERSION=v0.5.6 \
  --build-arg BUILD_TIME="$(date -u +'%Y-%m-%d %H:%M:%S')" \
  --build-arg GIT_COMMIT="$(git rev-parse --short HEAD)" \
  -t prjflow:latest .
```

2. **创建数据目录**

```bash
mkdir -p data
```

3. **运行容器**

```bash
docker run -d \
  --name prjflow \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e JWT_SECRET=your-secret-key-change-in-production \
  prjflow:latest
```

4. **查看日志**

```bash
docker logs -f prjflow
```

## 配置说明

### 环境变量

可以通过环境变量配置应用，主要环境变量如下：

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `SERVER_PORT` | 服务器端口 | `8080` |
| `DATABASE_TYPE` | 数据库类型 | `sqlite` |
| `DATABASE_DSN` | 数据库连接字符串 | `/app/data/data.db` |
| `JWT_SECRET` | JWT 密钥（必须修改） | 无 |
| `UPLOAD_STORAGE_PATH` | 上传文件存储路径 | `/app/data/uploads` |
| `TZ` | 时区 | `Asia/Shanghai` |

### 配置文件

如果需要使用配置文件，可以挂载 `config.yaml`：

```bash
docker run -d \
  --name prjflow \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  prjflow:latest
```

注意：配置文件中的路径应该使用绝对路径，例如：
- 数据库：`/app/data/data.db`
- 上传文件：`/app/data/uploads`

## 数据持久化

容器中的数据存储在 `/app/data` 目录，建议使用数据卷挂载到宿主机：

```bash
-v $(pwd)/data:/app/data
```

数据目录结构：
```
data/
├── data.db          # 主数据库（SQLite）
├── audit.db         # 审计数据库（SQLite）
├── uploads/         # 上传文件
├── logs/            # 日志文件
└── backups/         # 备份文件
```

## 健康检查

容器包含健康检查，默认每 30 秒检查一次。可以通过以下命令查看健康状态：

```bash
docker inspect --format='{{.State.Health.Status}}' prjflow
```

## 常用操作

### 查看容器状态

```bash
docker ps | grep prjflow
```

### 查看日志

```bash
# 实时日志
docker logs -f prjflow

# 最近 100 行日志
docker logs --tail 100 prjflow
```

### 停止容器

```bash
docker stop prjflow
```

### 启动容器

```bash
docker start prjflow
```

### 重启容器

```bash
docker restart prjflow
```

### 删除容器

```bash
docker stop prjflow
docker rm prjflow
```

### 进入容器

```bash
docker exec -it prjflow sh
```

### 备份数据库

```bash
# 方式一：直接复制数据库文件
docker cp prjflow:/app/data/data.db ./backup/data_$(date +%Y%m%d_%H%M%S).db

# 方式二：使用容器内的备份命令
docker exec prjflow /app/prjflow --backup
```

### 更新镜像

```bash
# 1. 停止并删除旧容器
docker stop prjflow
docker rm prjflow

# 2. 构建新镜像
docker build -t prjflow:latest .

# 3. 启动新容器
docker-compose up -d
# 或
docker run -d --name prjflow -p 8080:8080 -v $(pwd)/data:/app/data -e JWT_SECRET=your-secret-key prjflow:latest
```

## 生产环境部署建议

### 1. 使用反向代理

建议使用 Nginx 或 Traefik 作为反向代理，处理 HTTPS 和域名：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 设置强密码和密钥

- 修改 `JWT_SECRET` 为强随机字符串
- 使用环境变量文件（`.env`）管理敏感信息

### 3. 数据备份

定期备份数据目录：

```bash
# 创建备份脚本
#!/bin/bash
BACKUP_DIR="/backup/prjflow"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR
docker cp prjflow:/app/data $BACKUP_DIR/data_$DATE
tar -czf $BACKUP_DIR/data_$DATE.tar.gz -C $BACKUP_DIR data_$DATE
rm -rf $BACKUP_DIR/data_$DATE
```

### 4. 资源限制

在 `docker-compose.yml` 中添加资源限制：

```yaml
services:
  prjflow:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

### 5. 日志管理

配置日志轮转，避免日志文件过大：

```yaml
services:
  prjflow:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 故障排查

### 容器无法启动

1. 查看容器日志：
```bash
docker logs prjflow
```

2. 检查端口是否被占用：
```bash
netstat -tlnp | grep 8080
```

3. 检查数据目录权限：
```bash
ls -la data/
```

### 数据库连接失败

1. 检查数据库文件是否存在：
```bash
docker exec prjflow ls -la /app/data/
```

2. 检查数据库文件权限：
```bash
docker exec prjflow ls -la /app/data/data.db
```

### 上传文件失败

1. 检查上传目录权限：
```bash
docker exec prjflow ls -la /app/data/uploads/
```

2. 检查磁盘空间：
```bash
docker exec prjflow df -h
```

## 常见问题

### Q: 如何修改配置？

A: 有两种方式：
1. 使用环境变量（推荐）
2. 挂载配置文件

### Q: 数据会丢失吗？

A: 只要正确挂载了数据卷（`-v $(pwd)/data:/app/data`），数据会持久化到宿主机，容器删除后数据不会丢失。

### Q: 如何升级版本？

A: 参考"更新镜像"章节，注意备份数据。

### Q: 支持 MySQL 吗？

A: 当前 Dockerfile 使用纯 Go SQLite，如需 MySQL 支持，可以修改环境变量 `DATABASE_TYPE=mysql` 并配置 MySQL 连接信息。

### Q: 镜像体积多大？

A: 使用 Alpine Linux 和多阶段构建，最终镜像体积约 30-50MB（不包含数据）。

## 技术支持

如有问题，请查看：
- 项目 README
- GitHub Issues
- 项目文档

