#!/bin/bash

# 后端构建脚本（支持SQLite，需要CGO）
# 注意：此版本需要服务器GLIBC版本匹配

set -e

echo "开始构建后端服务（支持SQLite）..."
echo "⚠️  警告：此版本需要CGO支持，服务器GLIBC版本必须匹配编译环境"
echo ""

# 设置编译参数
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

echo "编译参数："
echo "  CGO_ENABLED=$CGO_ENABLED (启用CGO，支持SQLite)"
echo "  GOOS=$GOOS"
echo "  GOARCH=$GOARCH"
echo ""

# 检查是否有SQLite开发库
if ! pkg-config --exists sqlite3 2>/dev/null; then
    echo "⚠️  警告：未检测到SQLite开发库"
    echo "   Ubuntu/Debian: sudo apt-get install libsqlite3-dev"
    echo "   CentOS/RHEL: sudo yum install sqlite-devel"
    echo ""
fi

# 编译
go build -ldflags="-s -w" -o server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✓ 构建成功！"
    echo "  输出文件: ./server"
    echo "  文件大小: $(du -h server | cut -f1)"
    echo ""
    echo "提示："
    echo "  - 此版本支持SQLite数据库"
    echo "  - 但需要服务器GLIBC版本与编译环境匹配"
    echo "  - 如果遇到GLIBC错误，请使用 build.sh（静态编译，但仅支持MySQL）"
else
    echo "✗ 构建失败！"
    exit 1
fi

