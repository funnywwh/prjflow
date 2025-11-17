# 后端构建指南

## GLIBC版本问题

如果遇到以下错误：
```
./main: /lib64/libc.so.6: version `GLIBC_2.33' not found
./main: /lib64/libc.so.6: version `GLIBC_2.34' not found
```

这是因为编译环境的GLIBC版本比运行环境的GLIBC版本新。解决方案是使用**静态编译**。

## 构建方式

### 方式1：使用构建脚本（推荐）

**Linux版本（静态编译）**：
```bash
./build.sh
```

**Windows版本**：
```bash
./build-windows.sh
```

### 方式2：使用Makefile

```bash
# Linux版本（静态编译，推荐用于服务器部署）
make build-linux

# Windows版本
make build-windows

# macOS版本
make build-mac

# 当前平台
make build

# 清理构建文件
make clean
```

### 方式3：手动编译

**Linux静态编译（推荐）**：
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server cmd/server/main.go
```

**Windows静态编译**：
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o server.exe cmd/server/main.go
```

## 编译参数说明

- `CGO_ENABLED=0`: 禁用CGO，静态编译，不依赖系统库（解决GLIBC问题）
- `GOOS=linux`: 目标操作系统（linux/windows/darwin）
- `GOARCH=amd64`: 目标架构（amd64/arm64等）
- `-ldflags="-s -w"`: 减小二进制文件大小
  - `-s`: 去掉符号表
  - `-w`: 去掉调试信息

## 部署说明

### 静态编译的优势

1. **不依赖系统库**：编译后的二进制文件包含所有依赖，不依赖GLIBC等系统库
2. **跨平台兼容**：可以在不同版本的Linux系统上运行
3. **部署简单**：只需上传二进制文件和配置文件即可

### 部署步骤

1. **构建二进制文件**：
   ```bash
   make build-linux
   ```

2. **上传到服务器**：
   - 上传 `server` 二进制文件
   - 上传 `config.yaml` 配置文件
   - 上传数据库文件（如果使用SQLite）

3. **设置执行权限**：
   ```bash
   chmod +x server
   ```

4. **运行服务**：
   ```bash
   ./server
   ```

## 数据库选择

### SQLite vs MySQL

**SQLite（纯Go实现，推荐）**：
- ✅ 无需单独安装数据库服务器
- ✅ 适合开发和小型项目
- ✅ **支持静态编译（`CGO_ENABLED=0`）** - 使用纯Go驱动
- ✅ **不依赖系统库（解决GLIBC问题）** - 使用 `github.com/glebarez/go-sqlite`
- ✅ 可在任何Linux系统运行
- ⚠️ 性能略低于MySQL，但对于大多数应用场景足够

**MySQL（生产环境可选）**：
- ✅ 支持静态编译（`CGO_ENABLED=0`）
- ✅ 不依赖系统库（解决GLIBC问题）
- ✅ 性能更好，适合大型项目
- ❌ 需要单独安装MySQL服务器

### 构建选项

**选项1：静态编译 + SQLite（推荐）**
```bash
make build-linux
# 或
./build.sh
```
- ✅ 不依赖GLIBC，可在任何Linux系统运行
- ✅ 支持SQLite数据库（纯Go实现）
- ✅ 无需安装数据库服务器

**选项2：静态编译 + MySQL**
```bash
make build-linux
# 或
./build.sh
```
- ✅ 不依赖GLIBC，可在任何Linux系统运行
- ✅ 支持MySQL数据库
- ❌ 需要单独安装MySQL服务器

**选项3：动态编译 + SQLite（旧方式，不推荐）**
```bash
make build-linux-sqlite
# 或
./build-with-cgo.sh
```
- ✅ 支持SQLite数据库（CGO版本）
- ❌ 需要服务器GLIBC版本匹配
- ❌ 如果GLIBC不匹配会报错

## 注意事项

1. **SQLite驱动**：现在使用纯Go实现的 `github.com/glebarez/go-sqlite`，支持静态编译
2. **推荐配置**：SQLite + 静态编译，既简单又无需GLIBC匹配
3. **生产环境**：SQLite适合中小型项目，MySQL适合大型项目

2. **配置文件**：确保 `config.yaml` 文件在运行目录，或通过环境变量配置

3. **数据库文件**：如果使用SQLite，确保数据库文件路径正确，且有读写权限

## 故障排查

### 问题1：静态编译后SQLite无法使用（已解决）

**错误信息**（旧版本）：
```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work.
```

**原因**：旧版本的SQLite驱动（`github.com/mattn/go-sqlite3`）需要CGO支持

**解决方案**：已切换到纯Go实现的SQLite驱动（`github.com/glebarez/go-sqlite`）
- ✅ 支持静态编译（`CGO_ENABLED=0`）
- ✅ 无需CGO和系统库
- ✅ 可在任何Linux系统运行

**当前状态**：SQLite + 静态编译已完全支持，无需额外配置

### 问题2：文件太大

**原因**：包含调试信息

**解决方案**：使用 `-ldflags="-s -w"` 参数减小文件大小

### 问题3：运行时找不到配置文件

**原因**：配置文件路径不正确

**解决方案**：
- 确保 `config.yaml` 在运行目录
- 或通过环境变量指定配置文件路径

