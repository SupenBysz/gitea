# AI 协同工作规范插件 - API 接口文档

## 1. 概述

### 1.1 基本信息

- **Base URL**: `/api/v1/ai-plugin`
- **协议**: HTTPS
- **认证方式**: Bearer Token / API Key
- **数据格式**: JSON

### 1.2 认证

所有 API 请求需要在 Header 中携带认证信息：

```http
Authorization: Bearer <access_token>
```

或使用 API Key：

```http
X-API-Key: <api_key>
```

### 1.3 通用响应格式

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

#### 错误响应

```json
{
  "code": 1001,
  "message": "Invalid request",
  "error": "E1001",
  "details": {
    "field": "name",
    "reason": "Name is required"
  }
}
```

### 1.4 分页参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码 |
| `page_size` | int | 20 | 每页数量（最大 100）|

### 1.5 分页响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

---

## 2. 规则管理 API

### 2.1 获取规则列表

```http
GET /api/v1/ai-plugin/rules
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `enabled` | bool | 否 | 过滤启用状态 |
| `category` | string | 否 | 过滤分类 |
| `tags` | string | 否 | 过滤标签（逗号分隔）|
| `event_types` | string | 否 | 过滤事件类型（逗号分隔）|
| `page` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "pr-comment-required",
        "description": "PR 必须包含开始处理评论",
        "enabled": true,
        "priority": 100,
        "trigger": {
          "event_types": ["pull_request.opened"],
          "filters": []
        },
        "conditions_operator": "and",
        "metadata": {
          "category": "compliance",
          "tags": ["ai-employee", "pr-workflow"]
        },
        "conditions": [...],
        "actions": [...],
        "created_at": "2025-01-15T10:00:00Z",
        "updated_at": "2025-01-15T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 10,
      "total_pages": 1
    }
  }
}
```

### 2.2 获取单个规则

```http
GET /api/v1/ai-plugin/rules/{rule_id}
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| `rule_id` | int64 | 规则 ID |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "pr-comment-required",
    "description": "PR 必须包含开始处理评论",
    "enabled": true,
    "priority": 100,
    "trigger": {
      "event_types": ["pull_request.opened", "pull_request.synchronize"],
      "filters": [
        {
          "field": "pull_request.user.login",
          "operator": "in",
          "value": ["ArchitectAi", "DeveloperAi"]
        }
      ]
    },
    "conditions_operator": "and",
    "conditions": [
      {
        "id": 1,
        "type": "comment_exists",
        "params": {
          "pattern": "开始处理",
          "author_same_as": "pr_author"
        },
        "negate": false,
        "order_index": 0
      }
    ],
    "actions": [
      {
        "id": 1,
        "type": "block",
        "when_trigger": "condition_failed",
        "params": {
          "message": "请先在 Issue 评论「开始处理」"
        },
        "order_index": 0
      }
    ],
    "metadata": {
      "category": "compliance",
      "tags": ["ai-employee"]
    },
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
}
```

### 2.3 创建规则

```http
POST /api/v1/ai-plugin/rules
```

**请求体**

```json
{
  "name": "pr-comment-required",
  "description": "PR 必须包含开始处理评论",
  "enabled": true,
  "priority": 100,
  "trigger": {
    "event_types": ["pull_request.opened"],
    "filters": []
  },
  "conditions_operator": "and",
  "conditions": [
    {
      "type": "comment_exists",
      "params": {
        "pattern": "开始处理"
      },
      "negate": false
    }
  ],
  "actions": [
    {
      "type": "block",
      "when_trigger": "condition_failed",
      "params": {
        "message": "请先在 Issue 评论「开始处理」"
      }
    }
  ],
  "metadata": {
    "category": "compliance",
    "tags": ["ai-employee"]
  }
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "pr-comment-required",
    ...
  }
}
```

### 2.4 更新规则

```http
PUT /api/v1/ai-plugin/rules/{rule_id}
```

**请求体** (同创建规则)

### 2.5 删除规则

```http
DELETE /api/v1/ai-plugin/rules/{rule_id}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success"
}
```

### 2.6 验证规则

```http
POST /api/v1/ai-plugin/rules/validate
```

**请求体** (规则配置)

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "errors": []
  }
}
```

### 2.7 测试规则

```http
POST /api/v1/ai-plugin/rules/{rule_id}/test
```

**请求体**

```json
{
  "event": {
    "type": "pull_request.opened",
    "action": "opened",
    "repository": "org/repo",
    "payload": {
      "pull_request": {
        "number": 123,
        "title": "feat: new feature"
      }
    }
  }
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "rule_id": 1,
    "matched": true,
    "conditions_met": false,
    "violations": [
      {
        "rule_id": "1",
        "rule_name": "pr-comment-required",
        "message": "缺少开始处理评论",
        "severity": "high"
      }
    ],
    "actions_results": [],
    "duration_ms": 45
  }
}
```

---

## 3. 事件管理 API

### 3.1 Webhook 接收端点

```http
POST /api/v1/ai-plugin/webhook
```

**Headers**

| Header | 说明 |
|--------|------|
| `X-Gitea-Delivery` | 事件 ID |
| `X-Gitea-Event` | 事件类型 |
| `X-Gitea-Signature` | HMAC 签名 |

**请求体**: Gitea Webhook Payload

**响应示例**

```json
{
  "code": 0,
  "message": "accepted",
  "data": {
    "event_id": "abc123"
  }
}
```

### 3.2 获取事件列表

```http
GET /api/v1/ai-plugin/events
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `event_type` | string | 否 | 事件类型 |
| `repository` | string | 否 | 仓库名称 |
| `processed` | bool | 否 | 处理状态 |
| `start_time` | string | 否 | 开始时间 (ISO 8601) |
| `end_time` | string | 否 | 结束时间 (ISO 8601) |
| `page` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "event_id": "abc123",
        "event_type": "pull_request.opened",
        "action": "opened",
        "repository": "org/repo",
        "sender": "ArchitectAi",
        "processed": true,
        "created_at": "2025-01-15T10:00:00Z",
        "processed_at": "2025-01-15T10:00:01Z"
      }
    ],
    "pagination": {...}
  }
}
```

### 3.3 获取事件详情

```http
GET /api/v1/ai-plugin/events/{event_id}
```

### 3.4 重新处理事件

```http
POST /api/v1/ai-plugin/events/{event_id}/reprocess
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "event_id": "abc123",
    "reprocessed": true,
    "evaluation_id": 456
  }
}
```

### 3.5 获取事件统计

```http
GET /api/v1/ai-plugin/events/stats
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `repository` | string | 否 | 仓库名称 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_events": 1000,
    "events_by_type": {
      "pull_request.opened": 300,
      "pull_request.merged": 200,
      "issue.opened": 500
    },
    "events_by_repository": {
      "org/repo-a": 600,
      "org/repo-b": 400
    },
    "processed_events": 950,
    "failed_events": 50
  }
}
```

---

## 4. 评估结果 API

### 4.1 获取评估列表

```http
GET /api/v1/ai-plugin/evaluations
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `rule_id` | int64 | 否 | 规则 ID |
| `event_id` | int64 | 否 | 事件 ID |
| `matched` | bool | 否 | 匹配状态 |
| `conditions_met` | bool | 否 | 条件满足状态 |
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `page` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "rule_id": 1,
        "rule_name": "pr-comment-required",
        "event_id": 100,
        "matched": true,
        "conditions_met": false,
        "violations": [
          {
            "rule_id": "1",
            "rule_name": "pr-comment-required",
            "message": "缺少开始处理评论",
            "severity": "high"
          }
        ],
        "actions_results": [
          {
            "action_type": "notify",
            "success": true,
            "message": "Notification sent"
          }
        ],
        "duration_ms": 45,
        "created_at": "2025-01-15T10:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

### 4.2 获取评估详情

```http
GET /api/v1/ai-plugin/evaluations/{evaluation_id}
```

### 4.3 获取合规统计

```http
GET /api/v1/ai-plugin/evaluations/compliance
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `repository` | string | 否 | 仓库名称 |
| `rule_id` | int64 | 否 | 规则 ID |
| `granularity` | string | 否 | 粒度：hour/day/week |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "overall_compliance_rate": 85.5,
    "total_evaluations": 1000,
    "passed": 855,
    "failed": 145,
    "by_rule": [
      {
        "rule_id": 1,
        "rule_name": "pr-comment-required",
        "total": 500,
        "passed": 450,
        "failed": 50,
        "compliance_rate": 90.0
      }
    ],
    "trend": [
      {
        "time": "2025-01-15T00:00:00Z",
        "compliance_rate": 88.0,
        "total": 100
      }
    ]
  }
}
```

---

## 5. LLM 网关 API

### 5.1 发送聊天请求

```http
POST /api/v1/ai-plugin/llm/chat
```

**请求体**

```json
{
  "model": "claude-3-sonnet",
  "messages": [
    {"role": "system", "content": "You are a code reviewer."},
    {"role": "user", "content": "Review this code..."}
  ],
  "max_tokens": 1000,
  "temperature": 0.7,
  "stream": false
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "chat-123",
    "model": "claude-3-sonnet",
    "content": "Here is my code review...",
    "finish_reason": "stop",
    "usage": {
      "prompt_tokens": 150,
      "completion_tokens": 500,
      "total_tokens": 650,
      "cost": 0.00195
    },
    "created_at": "2025-01-15T10:00:00Z"
  }
}
```

### 5.2 流式聊天请求

```http
POST /api/v1/ai-plugin/llm/chat/stream
```

**请求体** (同 5.1，`stream` 设为 `true`)

**响应**: Server-Sent Events (SSE)

```
data: {"delta": "Here ", "finish_reason": null}

data: {"delta": "is my ", "finish_reason": null}

data: {"delta": "code review...", "finish_reason": null}

data: {"delta": "", "finish_reason": "stop"}
```

### 5.3 获取可用模型

```http
GET /api/v1/ai-plugin/llm/models
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "claude-3-opus",
      "provider": "anthropic",
      "name": "Claude 3 Opus",
      "description": "Most capable model",
      "max_tokens": 4096,
      "cost_per_1k_input": 0.015,
      "cost_per_1k_output": 0.075,
      "available": true
    },
    {
      "id": "claude-3-sonnet",
      "provider": "anthropic",
      "name": "Claude 3 Sonnet",
      "description": "Balanced model",
      "max_tokens": 4096,
      "cost_per_1k_input": 0.003,
      "cost_per_1k_output": 0.015,
      "available": true
    }
  ]
}
```

### 5.4 获取 LLM 使用统计

```http
GET /api/v1/ai-plugin/llm/usage
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `provider` | string | 否 | 提供商 |
| `model` | string | 否 | 模型 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_requests": 1000,
    "total_tokens": 500000,
    "total_cost": 15.50,
    "requests_by_model": {
      "claude-3-sonnet": 800,
      "claude-3-opus": 200
    },
    "tokens_by_model": {
      "claude-3-sonnet": 400000,
      "claude-3-opus": 100000
    },
    "cost_by_model": {
      "claude-3-sonnet": 6.00,
      "claude-3-opus": 9.50
    },
    "average_latency_ms": 1500
  }
}
```

---

## 6. 智能分析 API

### 6.1 代码分析

```http
POST /api/v1/ai-plugin/analysis/code
```

**请求体**

```json
{
  "repository": "org/repo",
  "pr_number": 123,
  "diff": "--- a/file.go\n+++ b/file.go\n...",
  "files": ["file.go"],
  "language": "go",
  "context": "This PR adds a new feature..."
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "overall_score": 85.0,
    "issues": [
      {
        "file": "file.go",
        "line": 42,
        "severity": "medium",
        "category": "performance",
        "message": "Consider using a more efficient algorithm",
        "suggestion": "Use a map instead of nested loops"
      }
    ],
    "suggestions": [
      {
        "file": "file.go",
        "line": 50,
        "original_code": "for i := 0; i < len(arr); i++ {",
        "suggested_code": "for _, item := range arr {",
        "reason": "Range loop is more idiomatic in Go",
        "confidence": 0.9
      }
    ],
    "summary": "Overall good code quality with minor suggestions",
    "analyzed_files": 1,
    "analyzed_lines": 150,
    "duration_ms": 2500
  }
}
```

### 6.2 评论质量分析

```http
POST /api/v1/ai-plugin/analysis/comment
```

**请求体**

```json
{
  "comment": "开始处理\n\n分支名称: feat/issue-14",
  "context": "Issue #14: Add new feature",
  "author": "ArchitectAi",
  "author_role": "ai_employee"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "quality_score": 75.0,
    "clarity": 80.0,
    "completeness": 70.0,
    "professionalism": 85.0,
    "issues": [
      {
        "type": "incomplete",
        "message": "Missing work plan details",
        "excerpt": ""
      }
    ],
    "improvements": [
      "Add estimated completion time",
      "List specific tasks to be done"
    ],
    "passes_standard": true
  }
}
```

### 6.3 生成 PR 摘要

```http
POST /api/v1/ai-plugin/analysis/summary
```

**请求体**

```json
{
  "type": "pr",
  "title": "feat: Add user authentication",
  "body": "This PR adds user authentication...",
  "diff": "...",
  "changes": "Modified 5 files, +200 -50 lines"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "summary": "Implements JWT-based user authentication with login/logout endpoints",
    "detailed_desc": "This PR introduces a complete authentication system...",
    "key_changes": [
      "Added JWT token generation and validation",
      "Created login and logout API endpoints",
      "Added authentication middleware"
    ],
    "impact_areas": [
      "User service",
      "API gateway",
      "Database schema"
    ],
    "testing_needed": [
      "Test login with valid/invalid credentials",
      "Test token expiration",
      "Test protected endpoint access"
    ]
  }
}
```

---

## 7. 通知管理 API

### 7.1 发送通知

```http
POST /api/v1/ai-plugin/notifications/send
```

**请求体**

```json
{
  "channels": ["telegram", "gitea_comment"],
  "template": "pr_compliance_check",
  "data": {
    "Repository": {"FullName": "org/repo"},
    "PullRequest": {"Number": 123, "Title": "feat: new feature"},
    "CheckPassed": false,
    "Violations": [...]
  },
  "recipients": ["user1", "user2"],
  "priority": "high"
}
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "results": [
      {
        "channel": "telegram",
        "success": true,
        "message_id": "msg-123"
      },
      {
        "channel": "gitea_comment",
        "success": true,
        "message_id": "comment-456"
      }
    ]
  }
}
```

### 7.2 获取通知模板列表

```http
GET /api/v1/ai-plugin/notifications/templates
```

### 7.3 创建通知模板

```http
POST /api/v1/ai-plugin/notifications/templates
```

**请求体**

```json
{
  "name": "pr_compliance_check",
  "description": "PR 合规检查通知模板",
  "channels": ["telegram", "gitea_comment"],
  "content": {
    "telegram": "🔍 *PR 合规检查结果*\n...",
    "gitea_comment": "## 🤖 自动合规检查结果\n..."
  },
  "variables": ["Repository", "PullRequest", "CheckPassed", "Violations"]
}
```

### 7.4 获取通知历史

```http
GET /api/v1/ai-plugin/notifications/history
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `channel` | string | 否 | 通知渠道 |
| `template` | string | 否 | 模板 ID |
| `status` | string | 否 | 状态：pending/sent/failed |
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `page` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |

---

## 8. 定时任务 API

### 8.1 获取任务列表

```http
GET /api/v1/ai-plugin/tasks
```

### 8.2 创建定时任务

```http
POST /api/v1/ai-plugin/tasks
```

**请求体**

```json
{
  "name": "daily_compliance_report",
  "description": "每日合规报告",
  "cron_expression": "0 9 * * *",
  "task_type": "compliance_report",
  "params": {
    "scope": "organization",
    "period": "last_24h",
    "notify_channels": ["telegram"]
  },
  "enabled": true
}
```

### 8.3 更新定时任务

```http
PUT /api/v1/ai-plugin/tasks/{task_id}
```

### 8.4 删除定时任务

```http
DELETE /api/v1/ai-plugin/tasks/{task_id}
```

### 8.5 手动触发任务

```http
POST /api/v1/ai-plugin/tasks/{task_id}/trigger
```

### 8.6 获取任务执行历史

```http
GET /api/v1/ai-plugin/tasks/{task_id}/executions
```

---

## 9. 审计日志 API

### 9.1 查询审计日志

```http
GET /api/v1/ai-plugin/audit
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `event_type` | string | 否 | 事件类型 |
| `actor_id` | string | 否 | 操作者 ID |
| `resource_type` | string | 否 | 资源类型 |
| `resource_id` | string | 否 | 资源 ID |
| `action` | string | 否 | 操作类型 |
| `result` | string | 否 | 结果：success/failure |
| `keyword` | string | 否 | 关键词搜索 |
| `start_time` | string | 否 | 开始时间 |
| `end_time` | string | 否 | 结束时间 |
| `page` | int | 否 | 页码 |
| `page_size` | int | 否 | 每页数量 |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "event_type": "rule_evaluation",
        "actor_type": "system",
        "actor_id": "rule-engine",
        "actor_name": "Rule Engine",
        "resource_type": "rule",
        "resource_id": "1",
        "resource_name": "pr-comment-required",
        "action": "execute",
        "result": "success",
        "details": {
          "event_id": 100,
          "matched": true
        },
        "ip_address": "10.0.0.1",
        "duration_ms": 45,
        "created_at": "2025-01-15T10:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

### 9.2 导出审计日志

```http
GET /api/v1/ai-plugin/audit/export
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `format` | string | 否 | 格式：csv/json/xlsx |
| (其他参数同 9.1)

### 9.3 获取审计统计

```http
GET /api/v1/ai-plugin/audit/stats
```

---

## 10. 仪表盘 API

### 10.1 获取概览数据

```http
GET /api/v1/ai-plugin/dashboard/overview
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "summary": {
      "total_events_today": 150,
      "compliance_rate": 92.5,
      "active_rules": 25,
      "violations_today": 12
    },
    "cards": [
      {
        "title": "今日事件",
        "value": 150,
        "trend": 15.5,
        "trend_direction": "up"
      },
      {
        "title": "合规率",
        "value": "92.5%",
        "trend": 2.3,
        "trend_direction": "up"
      },
      {
        "title": "活跃规则",
        "value": 25,
        "trend": 0,
        "trend_direction": "stable"
      },
      {
        "title": "今日违规",
        "value": 12,
        "trend": -5.2,
        "trend_direction": "down"
      }
    ]
  }
}
```

### 10.2 获取合规趋势

```http
GET /api/v1/ai-plugin/dashboard/compliance-trend
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `period` | string | 否 | 时间范围：7d/30d/90d |
| `granularity` | string | 否 | 粒度：hour/day/week |

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "period": "7d",
    "data_points": [
      {
        "time": "2025-01-09",
        "compliance_rate": 90.0,
        "total": 100,
        "passed": 90,
        "failed": 10
      },
      {
        "time": "2025-01-10",
        "compliance_rate": 92.0,
        "total": 120,
        "passed": 110,
        "failed": 10
      }
    ]
  }
}
```

### 10.3 获取事件流

```http
GET /api/v1/ai-plugin/dashboard/event-stream
```

**Query 参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `limit` | int | 否 | 返回数量（默认 20）|

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 100,
      "event_type": "pull_request.opened",
      "repository": "org/repo",
      "title": "feat: Add new feature",
      "author": "ArchitectAi",
      "compliance_status": "passed",
      "timestamp": "2025-01-15T10:00:00Z"
    }
  ]
}
```

### 10.4 WebSocket 实时更新

```http
GET /api/v1/ai-plugin/dashboard/ws
```

**WebSocket 消息格式**

```json
{
  "type": "event",
  "data": {
    "event_type": "pull_request.opened",
    "repository": "org/repo",
    ...
  }
}
```

```json
{
  "type": "evaluation",
  "data": {
    "rule_id": 1,
    "matched": true,
    ...
  }
}
```

```json
{
  "type": "stats_update",
  "data": {
    "compliance_rate": 92.5,
    "events_today": 151
  }
}
```

---

## 11. 用户管理 API

### 11.1 获取当前用户信息

```http
GET /api/v1/ai-plugin/user/me
```

### 11.2 更新用户设置

```http
PUT /api/v1/ai-plugin/user/settings
```

**请求体**

```json
{
  "notify_channels": ["telegram"],
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

### 11.3 获取用户列表（管理员）

```http
GET /api/v1/ai-plugin/users
```

### 11.4 分配用户角色（管理员）

```http
POST /api/v1/ai-plugin/users/{user_id}/roles
```

**请求体**

```json
{
  "role": "admin",
  "scope": "repository",
  "resource_id": "org/repo"
}
```

---

## 12. 系统配置 API

### 12.1 获取系统配置

```http
GET /api/v1/ai-plugin/config
```

### 12.2 更新系统配置

```http
PUT /api/v1/ai-plugin/config
```

**请求体**

```json
{
  "webhook_secret": "new-secret",
  "default_notification_channel": "telegram",
  "rate_limit": {
    "requests_per_minute": 100
  }
}
```

### 12.3 健康检查

```http
GET /api/v1/ai-plugin/health
```

**响应示例**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy",
    "components": {
      "database": "healthy",
      "redis": "healthy",
      "llm_gateway": "healthy"
    },
    "version": "1.0.0",
    "uptime": 86400
  }
}
```

---

## 13. 错误码参考

| 错误码 | HTTP 状态码 | 说明 |
|--------|-------------|------|
| E1000 | 500 | 内部错误 |
| E1001 | 400 | 无效请求 |
| E1002 | 401 | 未授权 |
| E1003 | 403 | 禁止访问 |
| E1004 | 404 | 资源不存在 |
| E1005 | 409 | 资源冲突 |
| E1006 | 429 | 请求限流 |
| E1007 | 408 | 请求超时 |
| E2001 | 404 | 规则不存在 |
| E2002 | 400 | 规则无效 |
| E2003 | 400 | 规则已禁用 |
| E2004 | 500 | 条件评估失败 |
| E2005 | 500 | 动作执行失败 |
| E3001 | 400 | 事件无效 |
| E3002 | 401 | 签名验证失败 |
| E3003 | 404 | 处理器不存在 |
| E4001 | 404 | 模型不存在 |
| E4002 | 503 | 模型不可用 |
| E4003 | 429 | 配额超限 |
| E4004 | 502 | 提供商错误 |
| E5001 | 404 | 渠道不存在 |
| E5002 | 404 | 模板不存在 |
| E5003 | 500 | 发送失败 |

---

*文档版本: 1.0.0*
*最后更新: 2025-01-15*
*作者: ArchitectAi*
