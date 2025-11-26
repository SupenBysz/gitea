// API Response Types
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
  request_id?: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Dashboard Types
export interface DashboardOverview {
  total_events: number
  events_today: number
  compliance_rate: number
  blocked_count: number
  active_rules: number
  total_rules: number
  ai_employees: AIEmployeeStatus[]
  recent_alerts: Alert[]
}

export interface AIEmployeeStatus {
  id: string
  name: string
  username: string
  avatar_url: string
  status: 'online' | 'offline' | 'busy'
  current_task?: string
  compliance_rate: number
  last_active: string
}

export interface Alert {
  id: string
  severity: 'critical' | 'high' | 'medium' | 'low'
  message: string
  source: string
  timestamp: string
  acknowledged: boolean
}

export interface ComplianceTrend {
  date: string
  compliance_rate: number
  total_events: number
  violations: number
}

export interface ViolationStats {
  rule_id: string
  rule_name: string
  count: number
  percentage: number
  trend: 'up' | 'down' | 'stable'
}

export interface BlockedIssue {
  id: number
  issue_number: number
  repo_owner: string
  repo_name: string
  title: string
  blocked_reason: string
  blocked_at: string
  assignee: string
  severity: 'critical' | 'high' | 'medium' | 'low'
}

// Event Types
export interface Event {
  id: string
  type: EventType
  source: string
  payload: Record<string, unknown>
  status: 'pending' | 'processing' | 'completed' | 'failed'
  created_at: string
  processed_at?: string
  error_message?: string
}

export type EventType =
  | 'push'
  | 'pull_request'
  | 'pull_request_review'
  | 'pull_request_review_comment'
  | 'issues'
  | 'issue_comment'
  | 'create'
  | 'delete'
  | 'fork'
  | 'release'

// Rule Types
export interface Rule {
  id: string
  name: string
  description: string
  enabled: boolean
  priority: number
  event_types: EventType[]
  conditions: RuleCondition[]
  actions: RuleAction[]
  created_at: string
  updated_at: string
  created_by: string
}

export interface RuleCondition {
  field: string
  operator: ConditionOperator
  value: string | number | boolean
  logic?: 'AND' | 'OR'
}

export type ConditionOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'starts_with'
  | 'ends_with'
  | 'matches'
  | 'greater_than'
  | 'less_than'
  | 'in'
  | 'not_in'

export interface RuleAction {
  type: ActionType
  config: Record<string, unknown>
}

export type ActionType =
  | 'block'
  | 'warn'
  | 'remind'
  | 'notify'
  | 'auto_fix'
  | 'llm_analyze'

// Evaluation Types
export interface Evaluation {
  id: string
  event_id: string
  rule_id: string
  rule_name: string
  passed: boolean
  action_taken: ActionType
  details: string
  duration_ms: number
  created_at: string
}

// Notification Types
export interface NotificationTemplate {
  id: string
  name: string
  channel: NotificationChannel
  subject: string
  body: string
  variables: string[]
  enabled: boolean
  created_at: string
  updated_at: string
}

export type NotificationChannel =
  | 'telegram'
  | 'gitea_comment'
  | 'webhook'
  | 'email'
  | 'slack'

// System Settings
export interface SystemSettings {
  webhook_secret: string
  default_llm_provider: string
  notification_channels: NotificationChannelConfig[]
  audit_retention_days: number
  metrics_retention_days: number
  max_concurrent_evaluations: number
}

export interface NotificationChannelConfig {
  channel: NotificationChannel
  enabled: boolean
  config: Record<string, string>
}

// WebSocket Message Types
export interface WSMessage {
  type: 'event' | 'evaluation' | 'alert' | 'status'
  data: Event | Evaluation | Alert | AIEmployeeStatus
  timestamp: string
}
