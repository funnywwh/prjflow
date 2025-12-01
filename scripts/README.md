# 发布和编译脚本使用说明

## 脚本列表

- `release.sh` - 发布版本脚本
- `build.sh` - 编译脚本

## 发布版本脚本 (release.sh)

### 功能

- 自动更新版本号（在 `backend/cmd/server/main.go` 中）
- 提交更改并推送到远程仓库
- 创建或更新 Git tag
- 推送 tag 到远程仓库

### 使用方法

```bash
# 基本使用
./scripts/release.sh <版本号> [发布说明]

# 示例
./scripts/release.sh v0.5.0 "新功能发布"
./scripts/release.sh v0.5.0 "修复Bug和改进性能"
```

### 版本号格式

版本号必须符合格式：`vX.Y.Z`

- `v` - 版本前缀
- `X` - 主版本号
- `Y` - 次版本号
- `Z` - 修订版本号

示例：
- `v0.5.0` - 新功能发布
- `v0.5.1` - Bug修复
- `v1.0.0` - 正式版本

### 工作流程

1. **检查工作目录**：确保工作目录干净（或确认未提交的更改）
2. **检查分支**：确认在 main 分支（或确认继续）
3. **更新版本号**：自动更新 `backend/cmd/server/main.go` 中的版本号
4. **提交更改**：提交版本号更新
5. **推送代码**：推送到远程仓库
6. **创建/更新 Tag**：创建带注释的 tag
7. **推送 Tag**：推送 tag 到远程仓库

### 注意事项

- 脚本会检查工作目录是否有未提交的更改
- 如果 tag 已存在，会询问是否更新
- 建议在发布前先运行测试确保代码正常

## 编译脚本 (build.sh)

### 功能

- 自动构建前端（Vue.js）
- 复制前端文件到 embed 目录
- 构建后端（Go），支持多平台
- 注入版本信息（版本号、构建时间、Git提交哈希）
- 输出到 `releases/<版本号>/` 目录

### 使用方法

```bash
# 基本使用
./scripts/build.sh [版本号] [平台]

# 示例
./scripts/build.sh v0.5.0 linux
./scripts/build.sh v0.5.0 windows
./scripts/build.sh v0.5.0 all
```

### 参数说明

#### 版本号（可选）

- 如果不提供，会自动从 Git tag 获取
- 如果 Git tag 不存在，默认使用 `v0.4.9`

#### 平台选项

- `linux` - Linux (静态编译, 仅支持MySQL)
- `linux-sqlite` - Linux (支持SQLite, 需要CGO)
- `windows` - Windows
- `mac` - macOS (Intel)
- `mac-arm` - macOS (Apple Silicon)
- `all` - 构建所有平台

### 输出

编译后的文件会输出到：`releases/<版本号>/`

文件命名格式：
- `server-linux-amd64` - Linux
- `server-linux-amd64-sqlite` - Linux (SQLite)
- `server-windows-amd64.exe` - Windows
- `server-darwin-amd64` - macOS (Intel)
- `server-darwin-arm64` - macOS (Apple Silicon)

### 构建信息注入

编译时会自动注入以下信息：
- **版本号**：从参数或 Git tag 获取
- **构建时间**：当前时间
- **Git提交哈希**：当前提交的短哈希

这些信息可以通过 `./server --version` 查看。

### 完整发布流程示例

```bash
# 1. 发布版本（更新版本号、创建tag）
./scripts/release.sh v0.5.0 "新功能发布"

# 2. 编译所有平台
./scripts/build.sh v0.5.0 all

# 3. 在 GitHub 上创建 Release
# 访问: https://github.com/funnywwh/goproject/releases/new
# 选择 tag: v0.5.0
# 上传编译好的文件: releases/v0.5.0/*
```

## 依赖要求

### release.sh

- Git
- sed (Linux/macOS)

### build.sh

- Node.js 和 yarn/npm（前端构建）
- Go 1.21+（后端构建）
- 对于 `linux-sqlite` 平台，需要 CGO 支持

## 故障排除

### release.sh

**问题：版本号更新失败**
- 检查 `backend/cmd/server/main.go` 中版本号格式是否正确
- 确保有写入权限

**问题：Tag 已存在**
- 脚本会询问是否更新，选择 `y` 更新或 `N` 跳过

### build.sh

**问题：前端构建失败**
- 确保已安装 Node.js 和 yarn/npm
- 运行 `cd frontend && yarn install` 安装依赖

**问题：后端构建失败**
- 确保已安装 Go 1.21+
- 对于 `linux-sqlite`，确保系统支持 CGO

**问题：找不到版本号**
- 手动指定版本号：`./scripts/build.sh v0.5.0 linux`

## 最佳实践

1. **发布前检查**
   - 运行测试：`cd backend && go test ./...`
   - 检查代码：`go vet ./...`
   - 确保所有更改已提交

2. **版本号规范**
   - 主版本号：不兼容的 API 修改
   - 次版本号：向下兼容的功能性新增
   - 修订版本号：向下兼容的问题修正

3. **发布流程**
   - 先运行 `release.sh` 创建 tag
   - 再运行 `build.sh` 编译
   - 最后在 GitHub 上创建 Release 并上传文件

4. **版本说明**
   - 在 GitHub Release 中详细说明新功能和修复
   - 可以引用 `docs/vX.X.X发布说明.md` 文档

