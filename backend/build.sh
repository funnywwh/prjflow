#!/bin/bash

# 后端构建脚本
# 使用静态编译，避免GLIBC版本依赖问题

set -e

echo "开始构建后端服务..."

# 设置编译参数
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 编译参数说明：
# - CGO_ENABLED=0: 禁用CGO，静态编译，不依赖系统库
# - GOOS=linux: 目标操作系统
# - GOARCH=amd64: 目标架构
# - -ldflags="-s -w": 减小二进制文件大小（-s去掉符号表，-w去掉调试信息）

echo "编译参数："
echo "  CGO_ENABLED=$CGO_ENABLED"
echo "  GOOS=$GOOS"
echo "  GOARCH=$GOARCH"
echo ""

# 编译
go build -ldflags="-s -w" -o server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✓ 构建成功！"
    echo "  输出文件: ./server"
    echo "  文件大小: $(du -h server | cut -f1)"
    echo ""
    echo "提示："
    echo "  - 这是一个静态编译的二进制文件，不依赖系统GLIBC"
    echo "  - 可以直接在Linux服务器上运行，无需安装Go环境"
    echo "  - 支持SQLite（纯Go实现）和MySQL数据库"
    echo "  - 记得同时上传 config.yaml 和数据库文件（如果使用SQLite）"
else
    echo "✗ 构建失败！"
    exit 1
fi

