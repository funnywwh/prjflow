#!/bin/bash

# Windows构建脚本（在WSL或Git Bash中运行）

set -e

echo "开始构建Windows版本..."

# 设置编译参数
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64

echo "编译参数："
echo "  CGO_ENABLED=$CGO_ENABLED"
echo "  GOOS=$GOOS"
echo "  GOARCH=$GOARCH"
echo ""

# 编译
go build -ldflags="-s -w" -o server.exe cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✓ 构建成功！"
    echo "  输出文件: ./server.exe"
else
    echo "✗ 构建失败！"
    exit 1
fi

