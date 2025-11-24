#!/bin/bash

set -e  # 遇到错误立即退出

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}=== Gitea 快速构建脚本 ===${NC}"
echo ""

# 获取版本
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
echo -e "${BLUE}版本:${NC} ${VERSION}"

# 设置环境变量
export CGO_ENABLED=0
export GOEXPERIMENT=jsonv2
export BROWSERSLIST_IGNORE_OLD_DATA=true

# 构建标签
TAGS="bindata sqlite sqlite_unlock_notify"

echo ""
echo -e "${BLUE}[1/3]${NC} 安装前端依赖..."
pnpm install --frozen-lockfile

echo ""
echo -e "${BLUE}[2/3]${NC} 编译前端..."
pnpm exec webpack --disable-interpret

echo ""
echo -e "${BLUE}[3/3]${NC} 编译后端..."
go build -v -tags "${TAGS}" -ldflags "-s -w -X main.Version=${VERSION}" -o gitea

echo ""
echo -e "${GREEN}✓ 构建完成!${NC}"
echo -e "  二进制文件: $(pwd)/gitea"
echo -e "  文件大小: $(du -h gitea | cut -f1)"
echo ""
echo -e "运行命令: ${BLUE}./gitea web${NC}"
echo ""
