#!/bin/bash

# Gitea Issue 创建脚本
# 用法: GITEA_TOKEN=your_token ./create-issue.sh

set -e

GITEA_URL="https://gitea.ktyun.cc"
REPO_OWNER="Kysion"
REPO_NAME="entai-gitea-custom"

# 检查 token
if [ -z "$GITEA_TOKEN" ]; then
    echo "错误: 请设置 GITEA_TOKEN 环境变量"
    echo "用法: GITEA_TOKEN=your_token ./create-issue.sh"
    echo ""
    echo "获取 token 的步骤:"
    echo "1. 访问 ${GITEA_URL}/user/settings/applications"
    echo "2. 创建新的 Personal Access Token"
    echo "3. 权限勾选: repo (完整权限)"
    exit 1
fi

# 读取 issue 模板
ISSUE_BODY=$(cat .issue-template.md)

# 创建 JSON payload
PAYLOAD=$(jq -n \
    --arg title "CI/CD 自动化发布系统实施计划" \
    --arg body "$ISSUE_BODY" \
    '{
        title: $title,
        body: $body,
        labels: [1, 2]
    }')

# 调用 Gitea API 创建 issue
echo "正在创建 Issue..."
RESPONSE=$(curl -s -X POST \
    "${GITEA_URL}/api/v1/repos/${REPO_OWNER}/${REPO_NAME}/issues" \
    -H "Authorization: token ${GITEA_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")

# 检查结果
ISSUE_NUMBER=$(echo "$RESPONSE" | jq -r '.number // empty')
if [ -n "$ISSUE_NUMBER" ]; then
    ISSUE_URL=$(echo "$RESPONSE" | jq -r '.html_url')
    echo "✅ Issue 创建成功!"
    echo "编号: #${ISSUE_NUMBER}"
    echo "URL: ${ISSUE_URL}"
else
    echo "❌ Issue 创建失败"
    echo "响应: $RESPONSE"
    exit 1
fi
