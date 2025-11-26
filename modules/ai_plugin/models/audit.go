// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// AuditLog represents an audit log entry for tracking system activities.
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
	IPAddress    string         `gorm:"size:45" json:"ip_address"`
	UserAgent    string         `gorm:"size:500" json:"user_agent"`
	RequestID    string         `gorm:"size:100" json:"request_id"`
	DurationMs   *int           `json:"duration_ms"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName returns the table name for AuditLog model.
func (AuditLog) TableName() string {
	return "ai_plugin_audit_log"
}

// Audit event types.
const (
	AuditEventRuleEvaluation   = "rule_evaluation"
	AuditEventActionExecution  = "action_execution"
	AuditEventConfigChange     = "config_change"
	AuditEventUserLogin        = "user_login"
	AuditEventRuleCRUD         = "rule_crud"
	AuditEventTemplateCRUD     = "template_crud"
	AuditEventAPICall          = "api_call"
	AuditEventWebhookReceived  = "webhook_received"
	AuditEventNotificationSent = "notification_sent"
	AuditEventLLMRequest       = "llm_request"
)

// Actor types.
const (
	ActorTypeUser      = "user"
	ActorTypeSystem    = "system"
	ActorTypeAPIClient = "api_client"
)

// Resource types.
const (
	ResourceTypeRule         = "rule"
	ResourceTypeEvent        = "event"
	ResourceTypeEvaluation   = "evaluation"
	ResourceTypeTemplate     = "template"
	ResourceTypeNotification = "notification"
	ResourceTypeConfig       = "config"
	ResourceTypeUser         = "user"
	ResourceTypeTask         = "task"
	ResourceTypeLLMProvider  = "llm_provider"
)

// Audit actions.
const (
	AuditActionCreate  = "create"
	AuditActionRead    = "read"
	AuditActionUpdate  = "update"
	AuditActionDelete  = "delete"
	AuditActionExecute = "execute"
	AuditActionLogin   = "login"
	AuditActionLogout  = "logout"
)

// Audit results.
const (
	AuditResultSuccess = "success"
	AuditResultFailure = "failure"
)

// AuditLogDetails represents additional details stored in the audit log.
type AuditLogDetails struct {
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Extra    interface{} `json:"extra,omitempty"`
}
