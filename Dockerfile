# 多阶段构建 Dockerfile - 纯 Go SQLite 版本（静态编译）

# Stage 1: 前端构建
FROM node:20-alpine AS frontend-builder

WORKDIR /build

# 复制前端文件
COPY frontend/package*.json ./
COPY frontend/yarn.lock* ./

# 安装依赖
RUN if [ -f yarn.lock ]; then \
      apk add --no-cache yarn && \
      yarn install --frozen-lockfile; \
    else \
      npm ci; \
    fi

# 复制前端源代码
COPY frontend/ ./

# 构建前端
RUN if [ -f yarn.lock ]; then \
      yarn build; \
    else \
      npm run build; \
    fi

# 验证构建结果
RUN test -f dist/index.html || (echo "前端构建失败：未找到 dist/index.html" && exit 1)

# Stage 2: 后端构建
FROM golang:1.24-alpine AS backend-builder

WORKDIR /build

# 安装必要的工具
RUN apk add --no-cache git

# 复制 go mod 文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖（利用 Docker 缓存）
RUN go mod download

# 复制后端源代码
COPY backend/ ./

# 从前端构建阶段复制前端构建产物
COPY --from=frontend-builder /build/dist ./cmd/server/frontend-dist

# 获取版本信息（从 git 或使用默认值）
ARG VERSION=unknown
ARG BUILD_TIME
ARG GIT_COMMIT

# 构建 Go 程序（静态编译，CGO_ENABLED=0，使用纯 Go SQLite）
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 构建参数（如果没有提供，自动获取）
RUN BUILD_TIME_VAR=${BUILD_TIME:-$(date -u +"%Y-%m-%d %H:%M:%S")} && \
    GIT_COMMIT_VAR=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")} && \
    go build \
      -ldflags="-s -w -X main.Version=${VERSION} -X 'main.BuildTime=${BUILD_TIME_VAR}' -X main.GitCommit=${GIT_COMMIT_VAR}" \
      -o prjflow \
      ./cmd/server/main.go

# 验证构建结果
RUN test -f prjflow || (echo "后端构建失败：未找到 prjflow 二进制文件" && exit 1)

# Stage 3: 运行时镜像
FROM alpine:latest

# 安装必要的运行时依赖（ca-certificates 用于 HTTPS 请求，wget 用于健康检查）
RUN apk add --no-cache ca-certificates tzdata wget && \
    rm -rf /var/cache/apk/*

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /build/prjflow /app/prjflow

# 创建数据目录
RUN mkdir -p /app/data /app/data/uploads /app/data/logs /app/data/backups && \
    chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV SERVER_PORT=8080
ENV DATABASE_TYPE=sqlite
ENV DATABASE_DSN=/app/data/data.db
ENV UPLOAD_STORAGE_PATH=/app/data/uploads

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT}/health || exit 1

# 启动命令
CMD ["/app/prjflow"]

