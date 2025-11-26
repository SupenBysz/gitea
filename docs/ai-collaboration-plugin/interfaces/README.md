# AI 协同工作规范插件 - 模块接口规范

## 1. 概述

本文档定义了 AI 协同工作规范插件各核心模块的接口规范，包括 Go 接口定义、数据结构和使用示例。

---

## 2. 核心接口定义

### 2.1 规则引擎接口 (Rule Engine)

```go
package rule

import (
    "context"
    "time"
)

// RuleEngine 规则引擎主接口
type RuleEngine interface {
    // LoadRules 从指定源加载规则
    LoadRules(ctx context.Context, source RuleSource) error

    // ReloadRules 重新加载所有规则
    ReloadRules(ctx context.Context) error

    // GetRule 根据 ID 获取规则
    GetRule(ctx context.Context, ruleID string) (*Rule, error)

    // ListRules 列出所有规则
    ListRules(ctx context.Context, filter *RuleFilter) ([]*Rule, error)

    // CreateRule 创建规则
    CreateRule(ctx context.Context, rule *Rule) error

    // UpdateRule 更新规则
    UpdateRule(ctx context.Context, rule *Rule) error

    // DeleteRule 删除规则
    DeleteRule(ctx context.Context, ruleID string) error

    // Evaluate 评估事件是否匹配规则并执行动作
    Evaluate(ctx context.Context, event *Event) (*EvaluationResult, error)

    // ValidateRule 校验规则配置
    ValidateRule(ctx context.Context, rule *Rule) error
}

// RuleSource 规则源类型
type RuleSource string

const (
    RuleSourceFile     RuleSource = "file"     // YAML 文件
    RuleSourceDatabase RuleSource = "database" // 数据库
    RuleSourceAPI      RuleSource = "api"      // 远程 API
)

// Rule 规则定义
type Rule struct {
    ID          string            `json:"id" yaml:"id"`
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description" yaml:"description"`
    Enabled     bool              `json:"enabled" yaml:"enabled"`
    Priority    int               `json:"priority" yaml:"priority"`
    Trigger     *RuleTrigger      `json:"trigger" yaml:"trigger"`
    Conditions  *ConditionGroup   `json:"conditions" yaml:"conditions"`
    Actions     []*RuleAction     `json:"actions" yaml:"actions"`
    Metadata    map[string]string `json:"metadata" yaml:"metadata"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// RuleTrigger 规则触发器配置
type RuleTrigger struct {
    EventTypes []string       `json:"event_types" yaml:"event_types"`
    Filters    []*EventFilter `json:"filters" yaml:"filters"`
}

// EventFilter 事件过滤器
type EventFilter struct {
    Field    string      `json:"field" yaml:"field"`
    Operator string      `json:"operator" yaml:"operator"` // eq, ne, in, not_in, contains, regex, gt, lt, gte, lte
    Value    interface{} `json:"value" yaml:"value"`
}

// ConditionGroup 条件组
type ConditionGroup struct {
    Operator string       `json:"operator" yaml:"operator"` // and, or
    Items    []*Condition `json:"items" yaml:"items"`
}

// Condition 单个条件
type Condition struct {
    Type   string                 `json:"type" yaml:"type"`
    Params map[string]interface{} `json:"params" yaml:"params"`
    Negate bool                   `json:"negate" yaml:"negate"` // 取反
}

// RuleAction 规则动作
type RuleAction struct {
    Type   string                 `json:"type" yaml:"type"`     // block, warn, remind, notify, auto_fix
    When   string                 `json:"when" yaml:"when"`     // condition_passed, condition_failed, always
    Params map[string]interface{} `json:"params" yaml:"params"`
}

// EvaluationResult 规则评估结果
type EvaluationResult struct {
    RuleID         string           `json:"rule_id"`
    Matched        bool             `json:"matched"`
    ConditionsMet  bool             `json:"conditions_met"`
    ActionsResults []*ActionResult  `json:"actions_results"`
    Violations     []*Violation     `json:"violations"`
    Duration       time.Duration    `json:"duration"`
    Timestamp      time.Time        `json:"timestamp"`
}

// ActionResult 动作执行结果
type ActionResult struct {
    ActionType string        `json:"action_type"`
    Success    bool          `json:"success"`
    Message    string        `json:"message"`
    Error      string        `json:"error,omitempty"`
    Duration   time.Duration `json:"duration"`
}

// Violation 违规记录
type Violation struct {
    RuleID      string    `json:"rule_id"`
    RuleName    string    `json:"rule_name"`
    Message     string    `json:"message"`
    Suggestion  string    `json:"suggestion"`
    Severity    string    `json:"severity"` // critical, high, medium, low
    Timestamp   time.Time `json:"timestamp"`
}

// RuleFilter 规则过滤条件
type RuleFilter struct {
    Enabled    *bool    `json:"enabled"`
    Category   string   `json:"category"`
    Tags       []string `json:"tags"`
    EventTypes []string `json:"event_types"`
}

// ConditionEvaluator 条件评估器接口
type ConditionEvaluator interface {
    // Type 返回条件类型标识
    Type() string

    // Evaluate 评估条件
    Evaluate(ctx context.Context, event *Event, params map[string]interface{}) (bool, error)

    // Validate 校验参数
    Validate(params map[string]interface{}) error
}

// ActionExecutor 动作执行器接口
type ActionExecutor interface {
    // Type 返回动作类型标识
    Type() string

    // Execute 执行动作
    Execute(ctx context.Context, event *Event, params map[string]interface{}) error

    // Validate 校验参数
    Validate(params map[string]interface{}) error
}
```

### 2.2 事件监听器接口 (Event Listener)

```go
package event

import (
    "context"
    "time"
)

// EventListener 事件监听器主接口
type EventListener interface {
    // Start 启动监听器
    Start(ctx context.Context) error

    // Stop 停止监听器
    Stop(ctx context.Context) error

    // RegisterHandler 注册事件处理器
    RegisterHandler(eventType string, handler EventHandler)

    // UnregisterHandler 注销事件处理器
    UnregisterHandler(eventType string, handler EventHandler)

    // HandleWebhook 处理 Webhook 请求
    HandleWebhook(ctx context.Context, payload []byte, headers map[string]string) error

    // GetEventStats 获取事件统计
    GetEventStats(ctx context.Context, filter *EventStatsFilter) (*EventStats, error)
}

// EventHandler 事件处理器接口
type EventHandler interface {
    // Handle 处理事件
    Handle(ctx context.Context, event *Event) error

    // SupportedEvents 返回支持的事件类型列表
    SupportedEvents() []string

    // Priority 返回处理优先级（数字越小优先级越高）
    Priority() int
}

// Event Gitea 事件
type Event struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`        // pull_request.opened, issue.created, etc.
    Action     string                 `json:"action"`      // opened, closed, merged, etc.
    Repository *Repository            `json:"repository"`
    Sender     *User                  `json:"sender"`
    Payload    map[string]interface{} `json:"payload"`     // 原始 Webhook 数据
    Timestamp  time.Time              `json:"timestamp"`
    Metadata   map[string]string      `json:"metadata"`
}

// Repository 仓库信息
type Repository struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    FullName  string `json:"full_name"`
    Owner     *User  `json:"owner"`
    CloneURL  string `json:"clone_url"`
    HTMLURL   string `json:"html_url"`
    Private   bool   `json:"private"`
}

// User 用户信息
type User struct {
    ID        int64  `json:"id"`
    Login     string `json:"login"`
    FullName  string `json:"full_name"`
    Email     string `json:"email"`
    AvatarURL string `json:"avatar_url"`
}

// PullRequest PR 信息
type PullRequest struct {
    ID        int64  `json:"id"`
    Number    int    `json:"number"`
    Title     string `json:"title"`
    Body      string `json:"body"`
    State     string `json:"state"`
    User      *User  `json:"user"`
    Head      *Branch `json:"head"`
    Base      *Branch `json:"base"`
    Mergeable bool   `json:"mergeable"`
    HTMLURL   string `json:"html_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Branch 分支信息
type Branch struct {
    Ref  string `json:"ref"`
    SHA  string `json:"sha"`
    Repo *Repository `json:"repo"`
}

// Issue Issue 信息
type Issue struct {
    ID        int64  `json:"id"`
    Number    int    `json:"number"`
    Title     string `json:"title"`
    Body      string `json:"body"`
    State     string `json:"state"`
    User      *User  `json:"user"`
    Labels    []*Label `json:"labels"`
    Assignees []*User  `json:"assignees"`
    HTMLURL   string `json:"html_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Label 标签
type Label struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color"`
}

// Comment 评论
type Comment struct {
    ID        int64  `json:"id"`
    Body      string `json:"body"`
    User      *User  `json:"user"`
    HTMLURL   string `json:"html_url"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// EventStatsFilter 事件统计过滤条件
type EventStatsFilter struct {
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    EventTypes []string  `json:"event_types"`
    Repository string    `json:"repository"`
}

// EventStats 事件统计
type EventStats struct {
    TotalEvents     int64            `json:"total_events"`
    EventsByType    map[string]int64 `json:"events_by_type"`
    EventsByRepo    map[string]int64 `json:"events_by_repo"`
    EventsByHour    map[int]int64    `json:"events_by_hour"`
    ProcessedEvents int64            `json:"processed_events"`
    FailedEvents    int64            `json:"failed_events"`
}
```

### 2.3 LLM 网关接口 (LLM Gateway)

```go
package llm

import (
    "context"
    "io"
    "time"
)

// LLMGateway LLM 网关主接口
type LLMGateway interface {
    // Chat 发送聊天请求
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream 发送流式聊天请求
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)

    // GetModels 获取可用模型列表
    GetModels(ctx context.Context) ([]*Model, error)

    // GetUsage 获取使用统计
    GetUsage(ctx context.Context, filter *UsageFilter) (*UsageStats, error)

    // HealthCheck 健康检查
    HealthCheck(ctx context.Context) error
}

// ChatRequest 聊天请求
type ChatRequest struct {
    Model       string            `json:"model"`        // 模型标识
    Messages    []*Message        `json:"messages"`     // 消息列表
    MaxTokens   int               `json:"max_tokens"`   // 最大 Token 数
    Temperature float64           `json:"temperature"`  // 温度参数
    TopP        float64           `json:"top_p"`        // Top-P 参数
    Stream      bool              `json:"stream"`       // 是否流式
    Metadata    map[string]string `json:"metadata"`     // 自定义元数据
}

// Message 消息
type Message struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"` // 消息内容
}

// ChatResponse 聊天响应
type ChatResponse struct {
    ID           string    `json:"id"`
    Model        string    `json:"model"`
    Content      string    `json:"content"`
    FinishReason string    `json:"finish_reason"` // stop, length, content_filter
    Usage        *Usage    `json:"usage"`
    CreatedAt    time.Time `json:"created_at"`
}

// Usage Token 使用统计
type Usage struct {
    PromptTokens     int     `json:"prompt_tokens"`
    CompletionTokens int     `json:"completion_tokens"`
    TotalTokens      int     `json:"total_tokens"`
    Cost             float64 `json:"cost"` // 预估费用（美元）
}

// StreamChunk 流式响应块
type StreamChunk struct {
    Delta        string `json:"delta"`         // 增量内容
    FinishReason string `json:"finish_reason"` // 完成原因
    Error        error  `json:"error"`         // 错误信息
}

// Model 模型信息
type Model struct {
    ID               string  `json:"id"`
    Provider         string  `json:"provider"`          // anthropic, openai, etc.
    Name             string  `json:"name"`
    Description      string  `json:"description"`
    MaxTokens        int     `json:"max_tokens"`
    CostPer1KInput   float64 `json:"cost_per_1k_input"`
    CostPer1KOutput  float64 `json:"cost_per_1k_output"`
    Available        bool    `json:"available"`
}

// UsageFilter 使用统计过滤条件
type UsageFilter struct {
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
    Model     string    `json:"model"`
    Provider  string    `json:"provider"`
}

// UsageStats 使用统计
type UsageStats struct {
    TotalRequests    int64            `json:"total_requests"`
    TotalTokens      int64            `json:"total_tokens"`
    TotalCost        float64          `json:"total_cost"`
    RequestsByModel  map[string]int64 `json:"requests_by_model"`
    TokensByModel    map[string]int64 `json:"tokens_by_model"`
    CostByModel      map[string]float64 `json:"cost_by_model"`
    AverageLatency   time.Duration    `json:"average_latency"`
}

// LLMAdapter LLM 适配器接口（用于扩展新的 LLM 提供商）
type LLMAdapter interface {
    // Provider 返回提供商标识
    Provider() string

    // Chat 发送聊天请求
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream 发送流式聊天请求
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)

    // GetModels 获取可用模型列表
    GetModels(ctx context.Context) ([]*Model, error)

    // HealthCheck 健康检查
    HealthCheck(ctx context.Context) error
}
```

### 2.4 智能分析接口 (Smart Analysis)

```go
package analysis

import (
    "context"
    "time"
)

// AnalysisService 智能分析服务接口
type AnalysisService interface {
    // AnalyzeCode 代码分析
    AnalyzeCode(ctx context.Context, req *CodeAnalysisRequest) (*CodeAnalysisResult, error)

    // AnalyzeComment 评论质量分析
    AnalyzeComment(ctx context.Context, req *CommentAnalysisRequest) (*CommentAnalysisResult, error)

    // AnalyzeDependencies 依赖分析
    AnalyzeDependencies(ctx context.Context, req *DependencyAnalysisRequest) (*DependencyAnalysisResult, error)

    // GenerateSummary 生成摘要
    GenerateSummary(ctx context.Context, req *SummaryRequest) (*SummaryResult, error)

    // AnalyzeConflict 冲突分析
    AnalyzeConflict(ctx context.Context, req *ConflictAnalysisRequest) (*ConflictAnalysisResult, error)
}

// CodeAnalysisRequest 代码分析请求
type CodeAnalysisRequest struct {
    Repository string   `json:"repository"`
    PRNumber   int      `json:"pr_number"`
    Diff       string   `json:"diff"`        // PR Diff 内容
    Files      []string `json:"files"`       // 变更文件列表
    Language   string   `json:"language"`    // 主要编程语言
    Context    string   `json:"context"`     // 额外上下文
}

// CodeAnalysisResult 代码分析结果
type CodeAnalysisResult struct {
    OverallScore   float64           `json:"overall_score"`   // 0-100
    Issues         []*CodeIssue      `json:"issues"`
    Suggestions    []*CodeSuggestion `json:"suggestions"`
    Summary        string            `json:"summary"`
    AnalyzedFiles  int               `json:"analyzed_files"`
    AnalyzedLines  int               `json:"analyzed_lines"`
    Duration       time.Duration     `json:"duration"`
}

// CodeIssue 代码问题
type CodeIssue struct {
    File        string `json:"file"`
    Line        int    `json:"line"`
    EndLine     int    `json:"end_line"`
    Severity    string `json:"severity"`    // critical, high, medium, low
    Category    string `json:"category"`    // security, performance, style, logic
    Message     string `json:"message"`
    Suggestion  string `json:"suggestion"`
    Rule        string `json:"rule"`        // 关联的规则
}

// CodeSuggestion 代码建议
type CodeSuggestion struct {
    File           string `json:"file"`
    Line           int    `json:"line"`
    OriginalCode   string `json:"original_code"`
    SuggestedCode  string `json:"suggested_code"`
    Reason         string `json:"reason"`
    Confidence     float64 `json:"confidence"` // 0-1
}

// CommentAnalysisRequest 评论分析请求
type CommentAnalysisRequest struct {
    Comment    string   `json:"comment"`      // 评论内容
    Context    string   `json:"context"`      // 上下文（Issue/PR 信息）
    Author     string   `json:"author"`       // 评论作者
    AuthorRole string   `json:"author_role"`  // ai_employee, human
}

// CommentAnalysisResult 评论分析结果
type CommentAnalysisResult struct {
    QualityScore    float64           `json:"quality_score"`    // 0-100
    Clarity         float64           `json:"clarity"`          // 清晰度
    Completeness    float64           `json:"completeness"`     // 完整性
    Professionalism float64           `json:"professionalism"`  // 专业性
    Issues          []*CommentIssue   `json:"issues"`
    Improvements    []string          `json:"improvements"`
    PassesStandard  bool              `json:"passes_standard"`
}

// CommentIssue 评论问题
type CommentIssue struct {
    Type    string `json:"type"`    // missing_context, unclear, incomplete, unprofessional
    Message string `json:"message"`
    Excerpt string `json:"excerpt"` // 问题片段
}

// DependencyAnalysisRequest 依赖分析请求
type DependencyAnalysisRequest struct {
    Repository   string `json:"repository"`
    ManifestFile string `json:"manifest_file"` // package.json, go.mod, etc.
    Content      string `json:"content"`       // 文件内容
}

// DependencyAnalysisResult 依赖分析结果
type DependencyAnalysisResult struct {
    TotalDependencies  int                   `json:"total_dependencies"`
    Vulnerabilities    []*Vulnerability      `json:"vulnerabilities"`
    OutdatedPackages   []*OutdatedPackage    `json:"outdated_packages"`
    LicenseIssues      []*LicenseIssue       `json:"license_issues"`
    RecommendedUpdates []*RecommendedUpdate  `json:"recommended_updates"`
}

// Vulnerability 安全漏洞
type Vulnerability struct {
    Package     string `json:"package"`
    Version     string `json:"version"`
    Severity    string `json:"severity"`     // critical, high, medium, low
    CVE         string `json:"cve"`
    Description string `json:"description"`
    FixVersion  string `json:"fix_version"`
}

// OutdatedPackage 过时包
type OutdatedPackage struct {
    Package        string `json:"package"`
    CurrentVersion string `json:"current_version"`
    LatestVersion  string `json:"latest_version"`
    UpdateType     string `json:"update_type"` // major, minor, patch
}

// LicenseIssue 许可证问题
type LicenseIssue struct {
    Package string `json:"package"`
    License string `json:"license"`
    Issue   string `json:"issue"`
}

// RecommendedUpdate 推荐更新
type RecommendedUpdate struct {
    Package        string `json:"package"`
    CurrentVersion string `json:"current_version"`
    NewVersion     string `json:"new_version"`
    Reason         string `json:"reason"`
    Breaking       bool   `json:"breaking"` // 是否有破坏性变更
}

// SummaryRequest 摘要请求
type SummaryRequest struct {
    Type    string `json:"type"`    // pr, issue, commit
    Title   string `json:"title"`
    Body    string `json:"body"`
    Diff    string `json:"diff,omitempty"`
    Changes string `json:"changes,omitempty"`
}

// SummaryResult 摘要结果
type SummaryResult struct {
    Summary       string   `json:"summary"`        // 简短摘要
    DetailedDesc  string   `json:"detailed_desc"`  // 详细描述
    KeyChanges    []string `json:"key_changes"`    // 关键变更点
    ImpactAreas   []string `json:"impact_areas"`   // 影响范围
    TestingNeeded []string `json:"testing_needed"` // 需要测试的点
}

// ConflictAnalysisRequest 冲突分析请求
type ConflictAnalysisRequest struct {
    Repository string   `json:"repository"`
    PRNumber   int      `json:"pr_number"`
    Conflicts  []string `json:"conflicts"` // 冲突文件列表
}

// ConflictAnalysisResult 冲突分析结果
type ConflictAnalysisResult struct {
    ConflictCount    int                  `json:"conflict_count"`
    Resolutions      []*ConflictResolution `json:"resolutions"`
    AutoResolvable   int                  `json:"auto_resolvable"`
    ManualRequired   int                  `json:"manual_required"`
}

// ConflictResolution 冲突解决建议
type ConflictResolution struct {
    File           string  `json:"file"`
    ConflictType   string  `json:"conflict_type"` // content, rename, delete
    Suggestion     string  `json:"suggestion"`
    ResolvedCode   string  `json:"resolved_code,omitempty"`
    Confidence     float64 `json:"confidence"`
    AutoResolvable bool    `json:"auto_resolvable"`
}
```

### 2.5 通知模块接口 (Notification)

```go
package notification

import (
    "context"
    "time"
)

// NotificationService 通知服务主接口
type NotificationService interface {
    // Send 发送通知
    Send(ctx context.Context, req *NotificationRequest) error

    // SendBatch 批量发送通知
    SendBatch(ctx context.Context, reqs []*NotificationRequest) ([]*NotificationResult, error)

    // GetTemplates 获取模板列表
    GetTemplates(ctx context.Context) ([]*Template, error)

    // GetTemplate 获取单个模板
    GetTemplate(ctx context.Context, templateID string) (*Template, error)

    // CreateTemplate 创建模板
    CreateTemplate(ctx context.Context, template *Template) error

    // UpdateTemplate 更新模板
    UpdateTemplate(ctx context.Context, template *Template) error

    // DeleteTemplate 删除模板
    DeleteTemplate(ctx context.Context, templateID string) error

    // GetHistory 获取通知历史
    GetHistory(ctx context.Context, filter *HistoryFilter) ([]*NotificationRecord, error)
}

// NotificationRequest 通知请求
type NotificationRequest struct {
    Channels   []string               `json:"channels"`    // telegram, gitea_comment, webhook, email
    Template   string                 `json:"template"`    // 模板 ID
    Data       map[string]interface{} `json:"data"`        // 模板数据
    Recipients []string               `json:"recipients"`  // 接收者（用户名或 ID）
    Priority   string                 `json:"priority"`    // high, normal, low
    Metadata   map[string]string      `json:"metadata"`
}

// NotificationResult 通知结果
type NotificationResult struct {
    Channel   string    `json:"channel"`
    Success   bool      `json:"success"`
    MessageID string    `json:"message_id"`
    Error     string    `json:"error,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

// Template 通知模板
type Template struct {
    ID          string                       `json:"id"`
    Name        string                       `json:"name"`
    Description string                       `json:"description"`
    Channels    []string                     `json:"channels"`
    Content     map[string]string            `json:"content"` // channel -> template content
    Variables   []string                     `json:"variables"` // 模板变量列表
    CreatedAt   time.Time                    `json:"created_at"`
    UpdatedAt   time.Time                    `json:"updated_at"`
}

// HistoryFilter 历史记录过滤条件
type HistoryFilter struct {
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    Channel    string    `json:"channel"`
    Template   string    `json:"template"`
    Recipient  string    `json:"recipient"`
    Success    *bool     `json:"success"`
    Limit      int       `json:"limit"`
    Offset     int       `json:"offset"`
}

// NotificationRecord 通知记录
type NotificationRecord struct {
    ID         string    `json:"id"`
    Channel    string    `json:"channel"`
    Template   string    `json:"template"`
    Recipient  string    `json:"recipient"`
    Content    string    `json:"content"`
    Success    bool      `json:"success"`
    Error      string    `json:"error,omitempty"`
    MessageID  string    `json:"message_id,omitempty"`
    Timestamp  time.Time `json:"timestamp"`
}

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
    // Name 返回渠道名称
    Name() string

    // Send 发送消息
    Send(ctx context.Context, message *ChannelMessage) error

    // Validate 校验配置
    Validate(config map[string]interface{}) error

    // HealthCheck 健康检查
    HealthCheck(ctx context.Context) error
}

// ChannelMessage 渠道消息
type ChannelMessage struct {
    Recipient string            `json:"recipient"`
    Subject   string            `json:"subject,omitempty"`
    Content   string            `json:"content"`
    Format    string            `json:"format"` // text, markdown, html
    Metadata  map[string]string `json:"metadata"`
}
```

### 2.6 自动化模块接口 (Automation)

```go
package automation

import (
    "context"
    "time"
)

// AutomationService 自动化服务主接口
type AutomationService interface {
    // ExecuteAction 执行自动化动作
    ExecuteAction(ctx context.Context, action *AutomationAction) (*ActionResult, error)

    // CreateScheduledTask 创建定时任务
    CreateScheduledTask(ctx context.Context, task *ScheduledTask) error

    // UpdateScheduledTask 更新定时任务
    UpdateScheduledTask(ctx context.Context, task *ScheduledTask) error

    // DeleteScheduledTask 删除定时任务
    DeleteScheduledTask(ctx context.Context, taskID string) error

    // GetScheduledTasks 获取定时任务列表
    GetScheduledTasks(ctx context.Context) ([]*ScheduledTask, error)

    // TriggerTask 手动触发任务
    TriggerTask(ctx context.Context, taskID string) (*ActionResult, error)

    // GetTaskHistory 获取任务执行历史
    GetTaskHistory(ctx context.Context, taskID string, limit int) ([]*TaskExecution, error)
}

// AutomationAction 自动化动作
type AutomationAction struct {
    Type       string                 `json:"type"`   // auto_fix, comment, label, assign, close, reopen
    Target     *ActionTarget          `json:"target"` // 目标对象
    Params     map[string]interface{} `json:"params"`
    DryRun     bool                   `json:"dry_run"` // 模拟执行
}

// ActionTarget 动作目标
type ActionTarget struct {
    Type       string `json:"type"`       // issue, pull_request, repository
    Repository string `json:"repository"` // owner/repo
    Number     int    `json:"number"`     // Issue/PR 编号
}

// ActionResult 动作执行结果
type ActionResult struct {
    Success   bool          `json:"success"`
    Message   string        `json:"message"`
    Changes   []string      `json:"changes"` // 变更描述列表
    Error     string        `json:"error,omitempty"`
    Duration  time.Duration `json:"duration"`
    Timestamp time.Time     `json:"timestamp"`
}

// ScheduledTask 定时任务
type ScheduledTask struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Cron        string                 `json:"cron"`     // Cron 表达式
    Enabled     bool                   `json:"enabled"`
    TaskType    string                 `json:"task_type"` // compliance_report, data_cleanup, sync
    Params      map[string]interface{} `json:"params"`
    NextRun     time.Time              `json:"next_run"`
    LastRun     time.Time              `json:"last_run"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// TaskExecution 任务执行记录
type TaskExecution struct {
    ID        string        `json:"id"`
    TaskID    string        `json:"task_id"`
    StartTime time.Time     `json:"start_time"`
    EndTime   time.Time     `json:"end_time"`
    Duration  time.Duration `json:"duration"`
    Status    string        `json:"status"` // running, success, failed, cancelled
    Output    string        `json:"output"`
    Error     string        `json:"error,omitempty"`
}
```

### 2.7 审计日志接口 (Audit Logger)

```go
package audit

import (
    "context"
    "time"
)

// AuditService 审计服务接口
type AuditService interface {
    // Log 记录审计日志
    Log(ctx context.Context, entry *AuditEntry) error

    // Query 查询审计日志
    Query(ctx context.Context, filter *AuditFilter) (*AuditQueryResult, error)

    // GetEntry 获取单条日志
    GetEntry(ctx context.Context, entryID string) (*AuditEntry, error)

    // Export 导出审计日志
    Export(ctx context.Context, filter *AuditFilter, format string) ([]byte, error)

    // GetStats 获取审计统计
    GetStats(ctx context.Context, filter *AuditFilter) (*AuditStats, error)
}

// AuditEntry 审计日志条目
type AuditEntry struct {
    ID           string                 `json:"id"`
    Timestamp    time.Time              `json:"timestamp"`
    EventType    string                 `json:"event_type"`    // rule_evaluation, action_execution, config_change, etc.
    Actor        *AuditActor            `json:"actor"`
    Resource     *AuditResource         `json:"resource"`
    Action       string                 `json:"action"`        // create, read, update, delete, execute
    Result       string                 `json:"result"`        // success, failure
    Details      map[string]interface{} `json:"details"`
    IPAddress    string                 `json:"ip_address"`
    UserAgent    string                 `json:"user_agent"`
    RequestID    string                 `json:"request_id"`
    Duration     time.Duration          `json:"duration"`
}

// AuditActor 操作者
type AuditActor struct {
    Type     string `json:"type"`      // user, system, api_client
    ID       string `json:"id"`
    Name     string `json:"name"`
    Role     string `json:"role"`
}

// AuditResource 资源
type AuditResource struct {
    Type       string `json:"type"`       // rule, config, repository, etc.
    ID         string `json:"id"`
    Name       string `json:"name"`
    Repository string `json:"repository,omitempty"`
}

// AuditFilter 审计日志过滤条件
type AuditFilter struct {
    StartTime    time.Time `json:"start_time"`
    EndTime      time.Time `json:"end_time"`
    EventTypes   []string  `json:"event_types"`
    ActorID      string    `json:"actor_id"`
    ActorType    string    `json:"actor_type"`
    ResourceType string    `json:"resource_type"`
    ResourceID   string    `json:"resource_id"`
    Action       string    `json:"action"`
    Result       string    `json:"result"`
    Keyword      string    `json:"keyword"` // 全文搜索
    Limit        int       `json:"limit"`
    Offset       int       `json:"offset"`
}

// AuditQueryResult 查询结果
type AuditQueryResult struct {
    Total   int64         `json:"total"`
    Entries []*AuditEntry `json:"entries"`
}

// AuditStats 审计统计
type AuditStats struct {
    TotalEntries     int64            `json:"total_entries"`
    EntriesByType    map[string]int64 `json:"entries_by_type"`
    EntriesByResult  map[string]int64 `json:"entries_by_result"`
    EntriesByActor   map[string]int64 `json:"entries_by_actor"`
    TopActors        []*ActorStats    `json:"top_actors"`
    TopResources     []*ResourceStats `json:"top_resources"`
}

// ActorStats 操作者统计
type ActorStats struct {
    ActorID   string `json:"actor_id"`
    ActorName string `json:"actor_name"`
    Count     int64  `json:"count"`
}

// ResourceStats 资源统计
type ResourceStats struct {
    ResourceType string `json:"resource_type"`
    ResourceID   string `json:"resource_id"`
    Count        int64  `json:"count"`
}
```

### 2.8 缓存管理接口 (Cache Manager)

```go
package cache

import (
    "context"
    "time"
)

// CacheService 缓存服务接口
type CacheService interface {
    // Get 获取缓存
    Get(ctx context.Context, key string) ([]byte, error)

    // Set 设置缓存
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

    // Delete 删除缓存
    Delete(ctx context.Context, key string) error

    // Exists 检查是否存在
    Exists(ctx context.Context, key string) (bool, error)

    // GetMulti 批量获取
    GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)

    // SetMulti 批量设置
    SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error

    // DeleteMulti 批量删除
    DeleteMulti(ctx context.Context, keys []string) error

    // Clear 清空所有缓存（指定前缀）
    Clear(ctx context.Context, prefix string) error

    // GetStats 获取缓存统计
    GetStats(ctx context.Context) (*CacheStats, error)
}

// DistributedLock 分布式锁接口
type DistributedLock interface {
    // Acquire 获取锁
    Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

    // Release 释放锁
    Release(ctx context.Context, key string) error

    // Extend 延长锁时间
    Extend(ctx context.Context, key string, ttl time.Duration) error

    // WithLock 在锁内执行函数
    WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error
}

// CacheStats 缓存统计
type CacheStats struct {
    Hits       int64   `json:"hits"`
    Misses     int64   `json:"misses"`
    HitRate    float64 `json:"hit_rate"`
    Keys       int64   `json:"keys"`
    MemoryUsed int64   `json:"memory_used"`
    Uptime     int64   `json:"uptime"`
}
```

---

## 3. 错误处理规范

### 3.1 错误码定义

```go
package errors

// ErrorCode 错误码
type ErrorCode string

const (
    // 通用错误 (1xxx)
    ErrInternal          ErrorCode = "E1000" // 内部错误
    ErrInvalidRequest    ErrorCode = "E1001" // 无效请求
    ErrUnauthorized      ErrorCode = "E1002" // 未授权
    ErrForbidden         ErrorCode = "E1003" // 禁止访问
    ErrNotFound          ErrorCode = "E1004" // 资源不存在
    ErrConflict          ErrorCode = "E1005" // 资源冲突
    ErrRateLimited       ErrorCode = "E1006" // 请求限流
    ErrTimeout           ErrorCode = "E1007" // 请求超时

    // 规则引擎错误 (2xxx)
    ErrRuleNotFound      ErrorCode = "E2001" // 规则不存在
    ErrRuleInvalid       ErrorCode = "E2002" // 规则无效
    ErrRuleDisabled      ErrorCode = "E2003" // 规则已禁用
    ErrConditionFailed   ErrorCode = "E2004" // 条件评估失败
    ErrActionFailed      ErrorCode = "E2005" // 动作执行失败

    // 事件监听错误 (3xxx)
    ErrEventInvalid      ErrorCode = "E3001" // 事件无效
    ErrEventSignature    ErrorCode = "E3002" // 签名验证失败
    ErrHandlerNotFound   ErrorCode = "E3003" // 处理器不存在

    // LLM 网关错误 (4xxx)
    ErrModelNotFound     ErrorCode = "E4001" // 模型不存在
    ErrModelUnavailable  ErrorCode = "E4002" // 模型不可用
    ErrQuotaExceeded     ErrorCode = "E4003" // 配额超限
    ErrProviderError     ErrorCode = "E4004" // 提供商错误

    // 通知错误 (5xxx)
    ErrChannelNotFound   ErrorCode = "E5001" // 渠道不存在
    ErrTemplateNotFound  ErrorCode = "E5002" // 模板不存在
    ErrSendFailed        ErrorCode = "E5003" // 发送失败
)

// AppError 应用错误
type AppError struct {
    Code    ErrorCode              `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
    Cause   error                  `json:"-"`
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError 创建错误
func NewError(code ErrorCode, message string) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
    }
}

// WithDetails 添加详情
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
    e.Details = details
    return e
}

// WithCause 添加原因
func (e *AppError) WithCause(cause error) *AppError {
    e.Cause = cause
    return e
}
```

---

## 4. 使用示例

### 4.1 规则引擎使用示例

```go
package main

import (
    "context"
    "fmt"
    "your-module/rule"
)

func main() {
    ctx := context.Background()

    // 创建规则引擎实例
    engine := rule.NewEngine(rule.Config{
        RuleDir: "/etc/ai-plugin/rules",
    })

    // 加载规则
    if err := engine.LoadRules(ctx, rule.RuleSourceFile); err != nil {
        panic(err)
    }

    // 模拟事件
    event := &rule.Event{
        Type:   "pull_request.opened",
        Action: "opened",
        Repository: &rule.Repository{
            FullName: "org/repo",
        },
        Payload: map[string]interface{}{
            "pull_request": map[string]interface{}{
                "number": 123,
                "title":  "feat: add new feature",
                "user": map[string]interface{}{
                    "login": "ArchitectAi",
                },
            },
        },
    }

    // 评估规则
    result, err := engine.Evaluate(ctx, event)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Rule matched: %v\n", result.Matched)
    fmt.Printf("Conditions met: %v\n", result.ConditionsMet)
    for _, v := range result.Violations {
        fmt.Printf("Violation: %s - %s\n", v.RuleName, v.Message)
    }
}
```

### 4.2 LLM 网关使用示例

```go
package main

import (
    "context"
    "fmt"
    "your-module/llm"
)

func main() {
    ctx := context.Background()

    // 创建 LLM 网关实例
    gateway := llm.NewGateway(llm.Config{
        Providers: []llm.ProviderConfig{
            {
                Name:    "claude",
                Type:    "anthropic",
                APIKey:  "your-api-key",
                BaseURL: "https://api.anthropic.com",
            },
        },
        DefaultModel: "claude-3-sonnet",
    })

    // 发送聊天请求
    req := &llm.ChatRequest{
        Model: "claude-3-sonnet",
        Messages: []*llm.Message{
            {Role: "system", Content: "You are a code reviewer."},
            {Role: "user", Content: "Review this code: func add(a, b int) int { return a + b }"},
        },
        MaxTokens:   1000,
        Temperature: 0.7,
    }

    resp, err := gateway.Chat(ctx, req)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

### 4.3 通知服务使用示例

```go
package main

import (
    "context"
    "your-module/notification"
)

func main() {
    ctx := context.Background()

    // 创建通知服务实例
    svc := notification.NewService(notification.Config{
        Telegram: notification.TelegramConfig{
            BotToken: "your-bot-token",
            ChatID:   "your-chat-id",
        },
    })

    // 发送通知
    req := &notification.NotificationRequest{
        Channels: []string{"telegram", "gitea_comment"},
        Template: "pr_compliance_check",
        Data: map[string]interface{}{
            "Repository": map[string]interface{}{
                "FullName": "org/repo",
            },
            "PullRequest": map[string]interface{}{
                "Number": 123,
                "Title":  "feat: add new feature",
            },
            "CheckPassed": false,
            "Violations": []map[string]interface{}{
                {
                    "Rule":    "pr-comment-required",
                    "Message": "缺少开始处理评论",
                },
            },
        },
    }

    if err := svc.Send(ctx, req); err != nil {
        panic(err)
    }
}
```

---

## 5. 接口版本管理

### 5.1 版本策略

- 接口版本遵循语义化版本规范（SemVer）
- 主版本号变更表示不兼容的 API 修改
- 次版本号变更表示向后兼容的功能新增
- 修订号变更表示向后兼容的问题修正

### 5.2 废弃策略

- 废弃接口将标记 `// Deprecated:` 注释
- 废弃接口将保留至少 2 个主版本周期
- 废弃前会提供迁移指南

---

*文档版本: 1.0.0*
*最后更新: 2025-01-15*
*作者: ArchitectAi*
