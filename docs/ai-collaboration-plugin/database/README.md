# AI 协同工作规范插件 - 数据库模型设计

## 1. 概述

本文档定义了 AI 协同工作规范插件的数据库模型设计，包括 ER 图、表结构、索引策略和数据迁移方案。

### 1.1 数据库选型

| 存储类型 | 数据库 | 用途 |
|----------|--------|------|
| 关系型数据 | PostgreSQL 15+ | 核心业务数据 |
| 时序数据 | TimescaleDB | 事件日志、指标数据 |
| 缓存 | Redis 7+ | 缓存、会话、锁 |
| 文件存储 | MinIO | 附件、报告文件 |

### 1.2 命名规范

- 表名使用小写下划线格式：`ai_plugin_xxx`
- 字段名使用小写下划线格式：`field_name`
- 主键统一使用 `id`（BIGINT GENERATED ALWAYS AS IDENTITY）
- 外键命名：`fk_{表名}_{引用表名}`
- 索引命名：`idx_{表名}_{字段名}`
- 唯一约束命名：`uk_{表名}_{字段名}`

---

## 2. ER 图

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Entity Relationship Diagram                         │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│   ai_plugin_rule │       │ai_plugin_rule_   │       │ai_plugin_rule_   │
│                  │       │   condition      │       │    action        │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │──┐    │ id (PK)          │       │ id (PK)          │
│ name             │  │    │ rule_id (FK)     │───┐   │ rule_id (FK)     │───┐
│ description      │  └───▶│ type             │   │   │ type             │   │
│ enabled          │       │ params (JSONB)   │   │   │ when_trigger     │   │
│ priority         │       │ negate           │   │   │ params (JSONB)   │   │
│ trigger (JSONB)  │       │ order_index      │   │   │ order_index      │   │
│ metadata (JSONB) │       │ created_at       │   │   │ created_at       │   │
│ created_at       │       └──────────────────┘   │   └──────────────────┘   │
│ updated_at       │                              │                          │
│ created_by       │       ┌──────────────────────┼──────────────────────────┘
└──────────────────┘       │                      │
         │                 │                      │
         │                 ▼                      ▼
         │       ┌──────────────────┐   ┌──────────────────┐
         │       │ai_plugin_event   │   │ai_plugin_        │
         │       │                  │   │  evaluation      │
         │       ├──────────────────┤   ├──────────────────┤
         │       │ id (PK)          │   │ id (PK)          │
         │       │ event_type       │   │ rule_id (FK)     │──────────────┐
         │       │ action           │   │ event_id (FK)    │─────┐        │
         │       │ repository       │   │ matched          │     │        │
         │       │ sender           │   │ conditions_met   │     │        │
         │       │ payload (JSONB)  │   │ violations (JSONB)│    │        │
         │       │ signature        │   │ duration_ms      │     │        │
         │       │ processed        │   │ created_at       │     │        │
         │       │ created_at       │   └──────────────────┘     │        │
         │       └──────────────────┘                            │        │
         │                 │                                     │        │
         │                 │                                     │        │
         │                 ▼                                     │        │
         │       ┌──────────────────┐                            │        │
         │       │ai_plugin_        │                            │        │
         │       │  notification    │                            │        │
         │       ├──────────────────┤                            │        │
         │       │ id (PK)          │                            │        │
         │       │ event_id (FK)    │────────────────────────────┘        │
         │       │ evaluation_id(FK)│                                     │
         │       │ channel          │                                     │
         │       │ template_id      │                                     │
         │       │ recipient        │                                     │
         │       │ content          │                                     │
         │       │ status           │                                     │
         │       │ message_id       │                                     │
         │       │ error_message    │                                     │
         │       │ created_at       │                                     │
         │       │ sent_at          │                                     │
         │       └──────────────────┘                                     │
         │                                                                │
         │       ┌──────────────────┐       ┌──────────────────┐          │
         │       │ai_plugin_        │       │ai_plugin_        │          │
         │       │  llm_request     │       │  llm_provider    │          │
         │       ├──────────────────┤       ├──────────────────┤          │
         │       │ id (PK)          │       │ id (PK)          │          │
         │       │ provider_id (FK) │──────▶│ name             │          │
         │       │ model            │       │ type             │          │
         │       │ evaluation_id(FK)│───────│ config (JSONB)   │          │
         │       │ messages (JSONB) │       │ enabled          │          │
         │       │ response         │       │ created_at       │          │
         │       │ prompt_tokens    │       │ updated_at       │          │
         │       │ completion_tokens│       └──────────────────┘          │
         │       │ cost             │                                     │
         │       │ duration_ms      │                                     │
         │       │ status           │                                     │
         │       │ error_message    │                                     │
         │       │ created_at       │                                     │
         │       └──────────────────┘                                     │
         │                                                                │
         │       ┌──────────────────┐       ┌──────────────────┐          │
         └──────▶│ai_plugin_        │       │ai_plugin_        │          │
                 │  audit_log       │       │  scheduled_task  │          │
                 ├──────────────────┤       ├──────────────────┤          │
                 │ id (PK)          │       │ id (PK)          │          │
                 │ event_type       │       │ name             │          │
                 │ actor_type       │       │ description      │          │
                 │ actor_id         │       │ cron_expression  │          │
                 │ actor_name       │       │ task_type        │          │
                 │ resource_type    │       │ params (JSONB)   │          │
                 │ resource_id      │       │ enabled          │          │
                 │ action           │       │ last_run_at      │          │
                 │ result           │       │ next_run_at      │          │
                 │ details (JSONB)  │       │ created_at       │          │
                 │ ip_address       │       │ updated_at       │          │
                 │ user_agent       │       └──────────────────┘          │
                 │ request_id       │                                     │
                 │ duration_ms      │       ┌──────────────────┐          │
                 │ created_at       │       │ai_plugin_        │          │
                 └──────────────────┘       │  task_execution  │          │
                                            ├──────────────────┤          │
                                            │ id (PK)          │          │
                                            │ task_id (FK)     │──────────┘
                                            │ start_time       │
                                            │ end_time         │
                                            │ status           │
                                            │ output           │
                                            │ error_message    │
                                            └──────────────────┘


┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ai_plugin_        │       │ai_plugin_user    │       │ai_plugin_        │
│  template        │       │                  │       │  user_role       │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │◀──────│ id (PK)          │
│ name             │       │ gitea_user_id    │       │ user_id (FK)     │
│ description      │       │ username         │       │ role             │
│ channels         │       │ email            │       │ scope            │
│ content (JSONB)  │       │ avatar_url       │       │ resource_id      │
│ variables        │       │ settings (JSONB) │       │ created_at       │
│ created_at       │       │ created_at       │       └──────────────────┘
│ updated_at       │       │ updated_at       │
│ created_by       │       │ last_login_at    │
└──────────────────┘       └──────────────────┘


┌──────────────────────────────────────────────────────────────────────┐
│                      TimescaleDB Hypertables                          │
└──────────────────────────────────────────────────────────────────────┘

┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ai_plugin_        │       │ai_plugin_        │       │ai_plugin_        │
│  event_metrics   │       │  compliance_     │       │  llm_metrics     │
│  (hypertable)    │       │  metrics         │       │  (hypertable)    │
├──────────────────┤       │  (hypertable)    │       ├──────────────────┤
│ time             │       ├──────────────────┤       │ time             │
│ repository       │       │ time             │       │ provider         │
│ event_type       │       │ repository       │       │ model            │
│ count            │       │ rule_id          │       │ requests         │
│ processed        │       │ passed           │       │ tokens           │
│ failed           │       │ failed           │       │ cost             │
│ avg_duration_ms  │       │ violations       │       │ avg_latency_ms   │
└──────────────────┘       └──────────────────┘       └──────────────────┘
```

---

## 3. 表结构定义

### 3.1 规则管理表

#### 3.1.1 ai_plugin_rule (规则表)

```sql
CREATE TABLE ai_plugin_rule (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 100,
    trigger JSONB NOT NULL,
    -- trigger 结构: {"event_types": ["pull_request.opened"], "filters": [...]}
    conditions_operator VARCHAR(10) NOT NULL DEFAULT 'and', -- and, or
    metadata JSONB DEFAULT '{}',
    -- metadata 结构: {"category": "compliance", "tags": ["ai-employee"]}
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),

    CONSTRAINT uk_ai_plugin_rule_name UNIQUE (name)
);

-- 索引
CREATE INDEX idx_ai_plugin_rule_enabled ON ai_plugin_rule(enabled);
CREATE INDEX idx_ai_plugin_rule_priority ON ai_plugin_rule(priority);
CREATE INDEX idx_ai_plugin_rule_trigger ON ai_plugin_rule USING GIN (trigger);
CREATE INDEX idx_ai_plugin_rule_metadata ON ai_plugin_rule USING GIN (metadata);

-- 注释
COMMENT ON TABLE ai_plugin_rule IS '规则配置表';
COMMENT ON COLUMN ai_plugin_rule.trigger IS '触发器配置，包含事件类型和过滤条件';
COMMENT ON COLUMN ai_plugin_rule.conditions_operator IS '条件组合运算符：and 或 or';
```

#### 3.1.2 ai_plugin_rule_condition (规则条件表)

```sql
CREATE TABLE ai_plugin_rule_condition (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    type VARCHAR(100) NOT NULL,
    -- 条件类型: comment_exists, branch_naming, file_pattern, label_check, etc.
    params JSONB NOT NULL DEFAULT '{}',
    -- params 结构根据 type 不同而不同
    negate BOOLEAN NOT NULL DEFAULT false,
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_rule_condition_rule FOREIGN KEY (rule_id)
        REFERENCES ai_plugin_rule(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_ai_plugin_rule_condition_rule ON ai_plugin_rule_condition(rule_id);
CREATE INDEX idx_ai_plugin_rule_condition_type ON ai_plugin_rule_condition(type);

-- 注释
COMMENT ON TABLE ai_plugin_rule_condition IS '规则条件配置表';
COMMENT ON COLUMN ai_plugin_rule_condition.type IS '条件类型标识';
COMMENT ON COLUMN ai_plugin_rule_condition.negate IS '是否取反条件结果';
```

#### 3.1.3 ai_plugin_rule_action (规则动作表)

```sql
CREATE TABLE ai_plugin_rule_action (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    type VARCHAR(100) NOT NULL,
    -- 动作类型: block, warn, remind, notify, auto_fix, label, assign, comment
    when_trigger VARCHAR(50) NOT NULL DEFAULT 'condition_failed',
    -- 触发时机: condition_passed, condition_failed, always
    params JSONB NOT NULL DEFAULT '{}',
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_rule_action_rule FOREIGN KEY (rule_id)
        REFERENCES ai_plugin_rule(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_ai_plugin_rule_action_rule ON ai_plugin_rule_action(rule_id);
CREATE INDEX idx_ai_plugin_rule_action_type ON ai_plugin_rule_action(type);

-- 注释
COMMENT ON TABLE ai_plugin_rule_action IS '规则动作配置表';
COMMENT ON COLUMN ai_plugin_rule_action.when_trigger IS '动作触发时机';
```

### 3.2 事件处理表

#### 3.2.1 ai_plugin_event (事件表)

```sql
CREATE TABLE ai_plugin_event (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id VARCHAR(100) NOT NULL,  -- Gitea 事件 ID
    event_type VARCHAR(100) NOT NULL,
    action VARCHAR(50),
    repository VARCHAR(255) NOT NULL,
    sender VARCHAR(255),
    payload JSONB NOT NULL,
    signature VARCHAR(255),  -- Webhook 签名
    processed BOOLEAN NOT NULL DEFAULT false,
    process_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uk_ai_plugin_event_id UNIQUE (event_id)
);

-- 索引
CREATE INDEX idx_ai_plugin_event_type ON ai_plugin_event(event_type);
CREATE INDEX idx_ai_plugin_event_repository ON ai_plugin_event(repository);
CREATE INDEX idx_ai_plugin_event_processed ON ai_plugin_event(processed);
CREATE INDEX idx_ai_plugin_event_created ON ai_plugin_event(created_at DESC);
CREATE INDEX idx_ai_plugin_event_payload ON ai_plugin_event USING GIN (payload);

-- 注释
COMMENT ON TABLE ai_plugin_event IS 'Gitea Webhook 事件记录表';
COMMENT ON COLUMN ai_plugin_event.payload IS '原始 Webhook 载荷';
```

#### 3.2.2 ai_plugin_evaluation (规则评估表)

```sql
CREATE TABLE ai_plugin_evaluation (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    event_id BIGINT NOT NULL,
    matched BOOLEAN NOT NULL,
    conditions_met BOOLEAN NOT NULL,
    violations JSONB DEFAULT '[]',
    -- violations 结构: [{"rule": "...", "message": "...", "severity": "..."}]
    actions_results JSONB DEFAULT '[]',
    -- actions_results 结构: [{"type": "...", "success": true, "message": "..."}]
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_evaluation_rule FOREIGN KEY (rule_id)
        REFERENCES ai_plugin_rule(id) ON DELETE SET NULL,
    CONSTRAINT fk_evaluation_event FOREIGN KEY (event_id)
        REFERENCES ai_plugin_event(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_ai_plugin_evaluation_rule ON ai_plugin_evaluation(rule_id);
CREATE INDEX idx_ai_plugin_evaluation_event ON ai_plugin_evaluation(event_id);
CREATE INDEX idx_ai_plugin_evaluation_matched ON ai_plugin_evaluation(matched);
CREATE INDEX idx_ai_plugin_evaluation_created ON ai_plugin_evaluation(created_at DESC);

-- 注释
COMMENT ON TABLE ai_plugin_evaluation IS '规则评估结果表';
```

### 3.3 通知管理表

#### 3.3.1 ai_plugin_template (通知模板表)

```sql
CREATE TABLE ai_plugin_template (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    channels VARCHAR(255)[] NOT NULL DEFAULT '{}',
    -- channels: telegram, gitea_comment, webhook, email, slack
    content JSONB NOT NULL,
    -- content 结构: {"telegram": "模板内容...", "gitea_comment": "模板内容..."}
    variables VARCHAR(255)[] NOT NULL DEFAULT '{}',
    -- 模板变量列表
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),

    CONSTRAINT uk_ai_plugin_template_name UNIQUE (name)
);

-- 注释
COMMENT ON TABLE ai_plugin_template IS '通知模板表';
COMMENT ON COLUMN ai_plugin_template.content IS '各渠道的模板内容，支持 Go template 语法';
```

#### 3.3.2 ai_plugin_notification (通知记录表)

```sql
CREATE TABLE ai_plugin_notification (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id BIGINT,
    evaluation_id BIGINT,
    channel VARCHAR(50) NOT NULL,
    template_id BIGINT,
    recipient VARCHAR(255),
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- status: pending, sent, failed
    message_id VARCHAR(255),  -- 外部消息 ID
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_notification_event FOREIGN KEY (event_id)
        REFERENCES ai_plugin_event(id) ON DELETE SET NULL,
    CONSTRAINT fk_notification_evaluation FOREIGN KEY (evaluation_id)
        REFERENCES ai_plugin_evaluation(id) ON DELETE SET NULL,
    CONSTRAINT fk_notification_template FOREIGN KEY (template_id)
        REFERENCES ai_plugin_template(id) ON DELETE SET NULL
);

-- 索引
CREATE INDEX idx_ai_plugin_notification_event ON ai_plugin_notification(event_id);
CREATE INDEX idx_ai_plugin_notification_channel ON ai_plugin_notification(channel);
CREATE INDEX idx_ai_plugin_notification_status ON ai_plugin_notification(status);
CREATE INDEX idx_ai_plugin_notification_created ON ai_plugin_notification(created_at DESC);

-- 注释
COMMENT ON TABLE ai_plugin_notification IS '通知发送记录表';
```

### 3.4 LLM 管理表

#### 3.4.1 ai_plugin_llm_provider (LLM 提供商表)

```sql
CREATE TABLE ai_plugin_llm_provider (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    -- type: anthropic, openai, azure, custom
    config JSONB NOT NULL,
    -- config 结构: {"api_key": "...", "base_url": "...", "models": [...]}
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uk_ai_plugin_llm_provider_name UNIQUE (name)
);

-- 注释
COMMENT ON TABLE ai_plugin_llm_provider IS 'LLM 提供商配置表';
COMMENT ON COLUMN ai_plugin_llm_provider.config IS '提供商配置，包含 API 密钥等敏感信息（加密存储）';
```

#### 3.4.2 ai_plugin_llm_request (LLM 请求记录表)

```sql
CREATE TABLE ai_plugin_llm_request (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    provider_id BIGINT NOT NULL,
    model VARCHAR(100) NOT NULL,
    evaluation_id BIGINT,
    task_type VARCHAR(100),
    -- task_type: code_review, comment_analysis, summary, etc.
    messages JSONB NOT NULL,
    -- messages 结构: [{"role": "...", "content": "..."}]
    response TEXT,
    finish_reason VARCHAR(50),
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cost DECIMAL(10, 6) NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- status: pending, success, failed
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_llm_request_provider FOREIGN KEY (provider_id)
        REFERENCES ai_plugin_llm_provider(id) ON DELETE SET NULL,
    CONSTRAINT fk_llm_request_evaluation FOREIGN KEY (evaluation_id)
        REFERENCES ai_plugin_evaluation(id) ON DELETE SET NULL
);

-- 索引
CREATE INDEX idx_ai_plugin_llm_request_provider ON ai_plugin_llm_request(provider_id);
CREATE INDEX idx_ai_plugin_llm_request_model ON ai_plugin_llm_request(model);
CREATE INDEX idx_ai_plugin_llm_request_status ON ai_plugin_llm_request(status);
CREATE INDEX idx_ai_plugin_llm_request_created ON ai_plugin_llm_request(created_at DESC);

-- 注释
COMMENT ON TABLE ai_plugin_llm_request IS 'LLM 请求记录表';
```

### 3.5 定时任务表

#### 3.5.1 ai_plugin_scheduled_task (定时任务表)

```sql
CREATE TABLE ai_plugin_scheduled_task (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cron_expression VARCHAR(100) NOT NULL,
    task_type VARCHAR(100) NOT NULL,
    -- task_type: compliance_report, data_cleanup, sync, custom
    params JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uk_ai_plugin_scheduled_task_name UNIQUE (name)
);

-- 索引
CREATE INDEX idx_ai_plugin_scheduled_task_enabled ON ai_plugin_scheduled_task(enabled);
CREATE INDEX idx_ai_plugin_scheduled_task_next_run ON ai_plugin_scheduled_task(next_run_at);

-- 注释
COMMENT ON TABLE ai_plugin_scheduled_task IS '定时任务配置表';
```

#### 3.5.2 ai_plugin_task_execution (任务执行记录表)

```sql
CREATE TABLE ai_plugin_task_execution (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id BIGINT NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    -- status: running, success, failed, cancelled
    output TEXT,
    error_message TEXT,

    CONSTRAINT fk_task_execution_task FOREIGN KEY (task_id)
        REFERENCES ai_plugin_scheduled_task(id) ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_ai_plugin_task_execution_task ON ai_plugin_task_execution(task_id);
CREATE INDEX idx_ai_plugin_task_execution_status ON ai_plugin_task_execution(status);
CREATE INDEX idx_ai_plugin_task_execution_start ON ai_plugin_task_execution(start_time DESC);

-- 注释
COMMENT ON TABLE ai_plugin_task_execution IS '任务执行记录表';
```

### 3.6 用户与权限表

#### 3.6.1 ai_plugin_user (用户表)

```sql
CREATE TABLE ai_plugin_user (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gitea_user_id BIGINT NOT NULL,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    avatar_url VARCHAR(500),
    settings JSONB DEFAULT '{}',
    -- settings 结构: {"notify_channels": ["telegram"], "language": "zh-CN"}
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uk_ai_plugin_user_gitea_id UNIQUE (gitea_user_id),
    CONSTRAINT uk_ai_plugin_user_username UNIQUE (username)
);

-- 索引
CREATE INDEX idx_ai_plugin_user_username ON ai_plugin_user(username);

-- 注释
COMMENT ON TABLE ai_plugin_user IS '插件用户表';
```

#### 3.6.2 ai_plugin_user_role (用户角色表)

```sql
CREATE TABLE ai_plugin_user_role (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL,
    -- role: super_admin, admin, operator, viewer, api_client
    scope VARCHAR(50) NOT NULL DEFAULT 'global',
    -- scope: global, organization, repository
    resource_id VARCHAR(255),
    -- resource_id: org name 或 repo full name
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user_role_user FOREIGN KEY (user_id)
        REFERENCES ai_plugin_user(id) ON DELETE CASCADE,
    CONSTRAINT uk_ai_plugin_user_role UNIQUE (user_id, role, scope, resource_id)
);

-- 索引
CREATE INDEX idx_ai_plugin_user_role_user ON ai_plugin_user_role(user_id);
CREATE INDEX idx_ai_plugin_user_role_role ON ai_plugin_user_role(role);

-- 注释
COMMENT ON TABLE ai_plugin_user_role IS '用户角色分配表';
```

### 3.7 审计日志表

#### 3.7.1 ai_plugin_audit_log (审计日志表)

```sql
CREATE TABLE ai_plugin_audit_log (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    -- event_type: rule_evaluation, action_execution, config_change, user_login, etc.
    actor_type VARCHAR(50) NOT NULL,
    -- actor_type: user, system, api_client
    actor_id VARCHAR(255),
    actor_name VARCHAR(255),
    resource_type VARCHAR(100),
    -- resource_type: rule, config, repository, etc.
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    action VARCHAR(50) NOT NULL,
    -- action: create, read, update, delete, execute
    result VARCHAR(20) NOT NULL,
    -- result: success, failure
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent VARCHAR(500),
    request_id VARCHAR(100),
    duration_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_ai_plugin_audit_log_event ON ai_plugin_audit_log(event_type);
CREATE INDEX idx_ai_plugin_audit_log_actor ON ai_plugin_audit_log(actor_id);
CREATE INDEX idx_ai_plugin_audit_log_resource ON ai_plugin_audit_log(resource_type, resource_id);
CREATE INDEX idx_ai_plugin_audit_log_action ON ai_plugin_audit_log(action);
CREATE INDEX idx_ai_plugin_audit_log_result ON ai_plugin_audit_log(result);
CREATE INDEX idx_ai_plugin_audit_log_created ON ai_plugin_audit_log(created_at DESC);
CREATE INDEX idx_ai_plugin_audit_log_details ON ai_plugin_audit_log USING GIN (details);

-- 注释
COMMENT ON TABLE ai_plugin_audit_log IS '审计日志表';
```

### 3.8 时序数据表 (TimescaleDB)

#### 3.8.1 ai_plugin_event_metrics (事件指标表)

```sql
-- 创建时序表
CREATE TABLE ai_plugin_event_metrics (
    time TIMESTAMPTZ NOT NULL,
    repository VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    processed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    avg_duration_ms DOUBLE PRECISION DEFAULT 0
);

-- 转换为 hypertable
SELECT create_hypertable('ai_plugin_event_metrics', 'time');

-- 创建索引
CREATE INDEX idx_event_metrics_repo ON ai_plugin_event_metrics(repository, time DESC);
CREATE INDEX idx_event_metrics_type ON ai_plugin_event_metrics(event_type, time DESC);

-- 创建连续聚合
CREATE MATERIALIZED VIEW ai_plugin_event_metrics_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    repository,
    event_type,
    SUM(count) AS total_count,
    SUM(processed) AS total_processed,
    SUM(failed) AS total_failed,
    AVG(avg_duration_ms) AS avg_duration
FROM ai_plugin_event_metrics
GROUP BY bucket, repository, event_type;

-- 设置数据保留策略（90天）
SELECT add_retention_policy('ai_plugin_event_metrics', INTERVAL '90 days');

-- 注释
COMMENT ON TABLE ai_plugin_event_metrics IS '事件处理指标时序表';
```

#### 3.8.2 ai_plugin_compliance_metrics (合规指标表)

```sql
CREATE TABLE ai_plugin_compliance_metrics (
    time TIMESTAMPTZ NOT NULL,
    repository VARCHAR(255) NOT NULL,
    rule_id BIGINT NOT NULL,
    passed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    violations INTEGER NOT NULL DEFAULT 0
);

-- 转换为 hypertable
SELECT create_hypertable('ai_plugin_compliance_metrics', 'time');

-- 创建索引
CREATE INDEX idx_compliance_metrics_repo ON ai_plugin_compliance_metrics(repository, time DESC);
CREATE INDEX idx_compliance_metrics_rule ON ai_plugin_compliance_metrics(rule_id, time DESC);

-- 创建连续聚合
CREATE MATERIALIZED VIEW ai_plugin_compliance_metrics_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    repository,
    rule_id,
    SUM(passed) AS total_passed,
    SUM(failed) AS total_failed,
    SUM(violations) AS total_violations,
    CASE
        WHEN SUM(passed) + SUM(failed) > 0
        THEN ROUND(SUM(passed)::NUMERIC / (SUM(passed) + SUM(failed)) * 100, 2)
        ELSE 0
    END AS compliance_rate
FROM ai_plugin_compliance_metrics
GROUP BY bucket, repository, rule_id;

-- 设置数据保留策略
SELECT add_retention_policy('ai_plugin_compliance_metrics', INTERVAL '180 days');

-- 注释
COMMENT ON TABLE ai_plugin_compliance_metrics IS '合规检查指标时序表';
```

#### 3.8.3 ai_plugin_llm_metrics (LLM 使用指标表)

```sql
CREATE TABLE ai_plugin_llm_metrics (
    time TIMESTAMPTZ NOT NULL,
    provider VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    requests INTEGER NOT NULL DEFAULT 0,
    tokens INTEGER NOT NULL DEFAULT 0,
    cost DECIMAL(10, 6) NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION DEFAULT 0,
    errors INTEGER NOT NULL DEFAULT 0
);

-- 转换为 hypertable
SELECT create_hypertable('ai_plugin_llm_metrics', 'time');

-- 创建索引
CREATE INDEX idx_llm_metrics_provider ON ai_plugin_llm_metrics(provider, time DESC);
CREATE INDEX idx_llm_metrics_model ON ai_plugin_llm_metrics(model, time DESC);

-- 创建连续聚合
CREATE MATERIALIZED VIEW ai_plugin_llm_metrics_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    provider,
    model,
    SUM(requests) AS total_requests,
    SUM(tokens) AS total_tokens,
    SUM(cost) AS total_cost,
    AVG(avg_latency_ms) AS avg_latency,
    SUM(errors) AS total_errors
FROM ai_plugin_llm_metrics
GROUP BY bucket, provider, model;

-- 设置数据保留策略
SELECT add_retention_policy('ai_plugin_llm_metrics', INTERVAL '365 days');

-- 注释
COMMENT ON TABLE ai_plugin_llm_metrics IS 'LLM 使用指标时序表';
```

---

## 4. Go 模型定义

```go
package models

import (
    "time"

    "gorm.io/datatypes"
)

// Rule 规则模型
type Rule struct {
    ID                 int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    Name               string         `gorm:"size:255;uniqueIndex;not null" json:"name"`
    Description        string         `gorm:"type:text" json:"description"`
    Enabled            bool           `gorm:"default:true;not null" json:"enabled"`
    Priority           int            `gorm:"default:100;not null" json:"priority"`
    Trigger            datatypes.JSON `gorm:"type:jsonb;not null" json:"trigger"`
    ConditionsOperator string         `gorm:"size:10;default:and;not null" json:"conditions_operator"`
    Metadata           datatypes.JSON `gorm:"type:jsonb;default:{}" json:"metadata"`
    CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    CreatedBy          string         `gorm:"size:255" json:"created_by"`

    // 关联
    Conditions []RuleCondition `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"conditions"`
    Actions    []RuleAction    `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"actions"`
}

func (Rule) TableName() string {
    return "ai_plugin_rule"
}

// RuleCondition 规则条件模型
type RuleCondition struct {
    ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    RuleID     int64          `gorm:"index;not null" json:"rule_id"`
    Type       string         `gorm:"size:100;not null" json:"type"`
    Params     datatypes.JSON `gorm:"type:jsonb;default:{}" json:"params"`
    Negate     bool           `gorm:"default:false;not null" json:"negate"`
    OrderIndex int            `gorm:"default:0;not null" json:"order_index"`
    CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (RuleCondition) TableName() string {
    return "ai_plugin_rule_condition"
}

// RuleAction 规则动作模型
type RuleAction struct {
    ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    RuleID      int64          `gorm:"index;not null" json:"rule_id"`
    Type        string         `gorm:"size:100;not null" json:"type"`
    WhenTrigger string         `gorm:"size:50;default:condition_failed;not null" json:"when_trigger"`
    Params      datatypes.JSON `gorm:"type:jsonb;default:{}" json:"params"`
    OrderIndex  int            `gorm:"default:0;not null" json:"order_index"`
    CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (RuleAction) TableName() string {
    return "ai_plugin_rule_action"
}

// Event 事件模型
type Event struct {
    ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    EventID      string         `gorm:"size:100;uniqueIndex;not null" json:"event_id"`
    EventType    string         `gorm:"size:100;index;not null" json:"event_type"`
    Action       string         `gorm:"size:50" json:"action"`
    Repository   string         `gorm:"size:255;index;not null" json:"repository"`
    Sender       string         `gorm:"size:255" json:"sender"`
    Payload      datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`
    Signature    string         `gorm:"size:255" json:"signature"`
    Processed    bool           `gorm:"default:false;index;not null" json:"processed"`
    ProcessError string         `gorm:"type:text" json:"process_error"`
    CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
    ProcessedAt  *time.Time     `json:"processed_at"`
}

func (Event) TableName() string {
    return "ai_plugin_event"
}

// Evaluation 评估结果模型
type Evaluation struct {
    ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    RuleID         int64          `gorm:"index;not null" json:"rule_id"`
    EventID        int64          `gorm:"index;not null" json:"event_id"`
    Matched        bool           `gorm:"not null" json:"matched"`
    ConditionsMet  bool           `gorm:"not null" json:"conditions_met"`
    Violations     datatypes.JSON `gorm:"type:jsonb;default:[]" json:"violations"`
    ActionsResults datatypes.JSON `gorm:"type:jsonb;default:[]" json:"actions_results"`
    DurationMs     int            `gorm:"default:0;not null" json:"duration_ms"`
    CreatedAt      time.Time      `gorm:"autoCreateTime;index" json:"created_at"`

    // 关联
    Rule  *Rule  `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
    Event *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

func (Evaluation) TableName() string {
    return "ai_plugin_evaluation"
}

// LLMProvider LLM 提供商模型
type LLMProvider struct {
    ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string         `gorm:"size:100;uniqueIndex;not null" json:"name"`
    Type      string         `gorm:"size:50;not null" json:"type"`
    Config    datatypes.JSON `gorm:"type:jsonb;not null" json:"config"`
    Enabled   bool           `gorm:"default:true;not null" json:"enabled"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LLMProvider) TableName() string {
    return "ai_plugin_llm_provider"
}

// LLMRequest LLM 请求记录模型
type LLMRequest struct {
    ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    ProviderID       int64          `gorm:"index;not null" json:"provider_id"`
    Model            string         `gorm:"size:100;index;not null" json:"model"`
    EvaluationID     *int64         `gorm:"index" json:"evaluation_id"`
    TaskType         string         `gorm:"size:100" json:"task_type"`
    Messages         datatypes.JSON `gorm:"type:jsonb;not null" json:"messages"`
    Response         string         `gorm:"type:text" json:"response"`
    FinishReason     string         `gorm:"size:50" json:"finish_reason"`
    PromptTokens     int            `gorm:"default:0;not null" json:"prompt_tokens"`
    CompletionTokens int            `gorm:"default:0;not null" json:"completion_tokens"`
    Cost             float64        `gorm:"type:decimal(10,6);default:0;not null" json:"cost"`
    DurationMs       int            `gorm:"default:0;not null" json:"duration_ms"`
    Status           string         `gorm:"size:20;default:pending;index;not null" json:"status"`
    ErrorMessage     string         `gorm:"type:text" json:"error_message"`
    CreatedAt        time.Time      `gorm:"autoCreateTime;index" json:"created_at"`

    // 关联
    Provider *LLMProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

func (LLMRequest) TableName() string {
    return "ai_plugin_llm_request"
}

// Template 通知模板模型
type Template struct {
    ID          int64                  `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string                 `gorm:"size:255;uniqueIndex;not null" json:"name"`
    Description string                 `gorm:"type:text" json:"description"`
    Channels    datatypes.JSONSlice[string] `gorm:"type:varchar(255)[]" json:"channels"`
    Content     datatypes.JSON         `gorm:"type:jsonb;not null" json:"content"`
    Variables   datatypes.JSONSlice[string] `gorm:"type:varchar(255)[]" json:"variables"`
    CreatedAt   time.Time              `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
    CreatedBy   string                 `gorm:"size:255" json:"created_by"`
}

func (Template) TableName() string {
    return "ai_plugin_template"
}

// Notification 通知记录模型
type Notification struct {
    ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    EventID      *int64     `gorm:"index" json:"event_id"`
    EvaluationID *int64     `gorm:"index" json:"evaluation_id"`
    Channel      string     `gorm:"size:50;index;not null" json:"channel"`
    TemplateID   *int64     `json:"template_id"`
    Recipient    string     `gorm:"size:255" json:"recipient"`
    Content      string     `gorm:"type:text;not null" json:"content"`
    Status       string     `gorm:"size:20;default:pending;index;not null" json:"status"`
    MessageID    string     `gorm:"size:255" json:"message_id"`
    ErrorMessage string     `gorm:"type:text" json:"error_message"`
    RetryCount   int        `gorm:"default:0;not null" json:"retry_count"`
    CreatedAt    time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
    SentAt       *time.Time `json:"sent_at"`
}

func (Notification) TableName() string {
    return "ai_plugin_notification"
}

// ScheduledTask 定时任务模型
type ScheduledTask struct {
    ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    Name           string         `gorm:"size:255;uniqueIndex;not null" json:"name"`
    Description    string         `gorm:"type:text" json:"description"`
    CronExpression string         `gorm:"size:100;not null" json:"cron_expression"`
    TaskType       string         `gorm:"size:100;not null" json:"task_type"`
    Params         datatypes.JSON `gorm:"type:jsonb;default:{}" json:"params"`
    Enabled        bool           `gorm:"default:true;index;not null" json:"enabled"`
    LastRunAt      *time.Time     `json:"last_run_at"`
    NextRunAt      *time.Time     `gorm:"index" json:"next_run_at"`
    CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ScheduledTask) TableName() string {
    return "ai_plugin_scheduled_task"
}

// TaskExecution 任务执行记录模型
type TaskExecution struct {
    ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    TaskID       int64      `gorm:"index;not null" json:"task_id"`
    StartTime    time.Time  `gorm:"not null;index" json:"start_time"`
    EndTime      *time.Time `json:"end_time"`
    Status       string     `gorm:"size:20;default:running;index;not null" json:"status"`
    Output       string     `gorm:"type:text" json:"output"`
    ErrorMessage string     `gorm:"type:text" json:"error_message"`

    // 关联
    Task *ScheduledTask `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

func (TaskExecution) TableName() string {
    return "ai_plugin_task_execution"
}

// User 用户模型
type User struct {
    ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    GiteaUserID int64          `gorm:"uniqueIndex;not null" json:"gitea_user_id"`
    Username    string         `gorm:"size:255;uniqueIndex;not null" json:"username"`
    Email       string         `gorm:"size:255" json:"email"`
    AvatarURL   string         `gorm:"size:500" json:"avatar_url"`
    Settings    datatypes.JSON `gorm:"type:jsonb;default:{}" json:"settings"`
    CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    LastLoginAt *time.Time     `json:"last_login_at"`

    // 关联
    Roles []UserRole `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"roles"`
}

func (User) TableName() string {
    return "ai_plugin_user"
}

// UserRole 用户角色模型
type UserRole struct {
    ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID     int64     `gorm:"index;not null" json:"user_id"`
    Role       string    `gorm:"size:50;index;not null" json:"role"`
    Scope      string    `gorm:"size:50;default:global;not null" json:"scope"`
    ResourceID string    `gorm:"size:255" json:"resource_id"`
    CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRole) TableName() string {
    return "ai_plugin_user_role"
}

// AuditLog 审计日志模型
type AuditLog struct {
    ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
    EventType    string         `gorm:"size:100;index;not null" json:"event_type"`
    ActorType    string         `gorm:"size:50;not null" json:"actor_type"`
    ActorID      string         `gorm:"size:255;index" json:"actor_id"`
    ActorName    string         `gorm:"size:255" json:"actor_name"`
    ResourceType string         `gorm:"size:100;index" json:"resource_type"`
    ResourceID   string         `gorm:"size:255;index" json:"resource_id"`
    ResourceName string         `gorm:"size:255" json:"resource_name"`
    Action       string         `gorm:"size:50;index;not null" json:"action"`
    Result       string         `gorm:"size:20;index;not null" json:"result"`
    Details      datatypes.JSON `gorm:"type:jsonb;default:{}" json:"details"`
    IPAddress    string         `gorm:"type:inet" json:"ip_address"`
    UserAgent    string         `gorm:"size:500" json:"user_agent"`
    RequestID    string         `gorm:"size:100" json:"request_id"`
    DurationMs   *int           `json:"duration_ms"`
    CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
}

func (AuditLog) TableName() string {
    return "ai_plugin_audit_log"
}
```

---

## 5. 数据迁移

### 5.1 迁移脚本结构

```
migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_create_indexes.up.sql
├── 000002_create_indexes.down.sql
├── 000003_create_hypertables.up.sql
├── 000003_create_hypertables.down.sql
└── ...
```

### 5.2 使用 golang-migrate

```bash
# 安装
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移
migrate create -ext sql -dir migrations -seq init_schema

# 执行迁移
migrate -path migrations -database "postgres://user:pass@host:5432/dbname?sslmode=disable" up

# 回滚
migrate -path migrations -database "postgres://user:pass@host:5432/dbname?sslmode=disable" down 1
```

---

## 6. 索引策略

### 6.1 索引设计原则

1. **主键索引**: 所有表使用自增 BIGINT 主键
2. **外键索引**: 所有外键字段创建索引
3. **查询索引**: 根据常用查询条件创建组合索引
4. **JSONB 索引**: 使用 GIN 索引支持 JSONB 字段查询
5. **时间索引**: 时间字段使用 BRIN 或 B-tree 索引（降序）

### 6.2 索引监控

```sql
-- 查看索引使用情况
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- 查看未使用的索引
SELECT
    schemaname,
    tablename,
    indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0
AND schemaname = 'public';
```

---

## 7. 数据保留策略

| 数据类型 | 保留时间 | 归档策略 |
|----------|----------|----------|
| 事件日志 | 90 天 | 压缩归档到 MinIO |
| 规则评估 | 180 天 | 聚合后删除明细 |
| LLM 请求 | 90 天 | 仅保留统计数据 |
| 审计日志 | 365 天 | 压缩归档 |
| 时序指标 | 按表配置 | TimescaleDB 自动管理 |

---

*文档版本: 1.0.0*
*最后更新: 2025-01-15*
*作者: ArchitectAi*
