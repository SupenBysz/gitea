// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package services provides core service interfaces for the AI collaboration plugin.
package services

import (
	"context"

	"code.gitea.io/gitea/modules/ai_plugin/models"
)

// RuleEngine defines the interface for rule evaluation and management.
type RuleEngine interface {
	// LoadRules loads rules from the configured source.
	LoadRules(ctx context.Context) error

	// ReloadRules reloads rules from the source.
	ReloadRules(ctx context.Context) error

	// Evaluate evaluates all applicable rules against an event.
	Evaluate(ctx context.Context, event *models.Event) (*EvaluationResult, error)

	// GetRule retrieves a rule by ID.
	GetRule(ctx context.Context, id int64) (*models.Rule, error)

	// GetRuleByName retrieves a rule by name.
	GetRuleByName(ctx context.Context, name string) (*models.Rule, error)

	// ListRules lists all rules with optional filtering.
	ListRules(ctx context.Context, filter *RuleFilter) ([]*models.Rule, int64, error)

	// CreateRule creates a new rule.
	CreateRule(ctx context.Context, rule *models.Rule) error

	// UpdateRule updates an existing rule.
	UpdateRule(ctx context.Context, rule *models.Rule) error

	// DeleteRule deletes a rule by ID.
	DeleteRule(ctx context.Context, id int64) error

	// ValidateRule validates a rule configuration.
	ValidateRule(ctx context.Context, rule *models.Rule) error

	// TestRule tests a rule against sample data.
	TestRule(ctx context.Context, rule *models.Rule, sampleEvent *models.Event) (*EvaluationResult, error)
}

// RuleFilter defines filtering options for listing rules.
type RuleFilter struct {
	Enabled    *bool
	EventTypes []string
	Category   string
	Tags       []string
	Search     string
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
}

// EvaluationResult represents the result of evaluating rules against an event.
type EvaluationResult struct {
	EventID       int64                    `json:"event_id"`
	RulesMatched  int                      `json:"rules_matched"`
	RulesEvaluated int                     `json:"rules_evaluated"`
	Passed        bool                     `json:"passed"`
	Violations    []models.Violation       `json:"violations"`
	ActionsResults []models.ActionResult   `json:"actions_results"`
	Evaluations   []*models.Evaluation     `json:"evaluations"`
	DurationMs    int                      `json:"duration_ms"`
}

// EventListener defines the interface for receiving and processing webhook events.
type EventListener interface {
	// Start starts the event listener.
	Start(ctx context.Context) error

	// Stop stops the event listener.
	Stop(ctx context.Context) error

	// HandleWebhook processes an incoming webhook request.
	HandleWebhook(ctx context.Context, eventType string, payload []byte, signature string) error

	// RegisterHandler registers a handler for a specific event type.
	RegisterHandler(eventType string, handler EventHandler)

	// GetStats returns listener statistics.
	GetStats() *ListenerStats
}

// EventHandler is a function that handles a specific event type.
type EventHandler func(ctx context.Context, event *models.Event) error

// ListenerStats contains statistics about the event listener.
type ListenerStats struct {
	EventsReceived  int64 `json:"events_received"`
	EventsProcessed int64 `json:"events_processed"`
	EventsFailed    int64 `json:"events_failed"`
	QueueSize       int   `json:"queue_size"`
	WorkersActive   int   `json:"workers_active"`
}

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	// Send sends a notification to the specified channel.
	Send(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error)

	// SendBatch sends multiple notifications.
	SendBatch(ctx context.Context, reqs []*NotificationRequest) ([]*NotificationResponse, error)

	// GetNotification retrieves a notification by ID.
	GetNotification(ctx context.Context, id int64) (*models.Notification, error)

	// ListNotifications lists notifications with filtering.
	ListNotifications(ctx context.Context, filter *NotificationFilter) ([]*models.Notification, int64, error)

	// RetryFailed retries failed notifications.
	RetryFailed(ctx context.Context) (int, error)
}

// NotificationRequest represents a request to send a notification.
type NotificationRequest struct {
	Channel      string                 `json:"channel"`
	Recipient    string                 `json:"recipient"`
	TemplateID   *int64                 `json:"template_id,omitempty"`
	TemplateName string                 `json:"template_name,omitempty"`
	Content      string                 `json:"content,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	EventID      *int64                 `json:"event_id,omitempty"`
	EvaluationID *int64                 `json:"evaluation_id,omitempty"`
}

// NotificationResponse represents the result of sending a notification.
type NotificationResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NotificationFilter defines filtering options for listing notifications.
type NotificationFilter struct {
	Channel  string
	Status   string
	EventID  *int64
	FromTime *int64
	ToTime   *int64
	Page     int
	PageSize int
}

// LLMGateway defines the interface for interacting with LLM providers.
type LLMGateway interface {
	// Chat sends a chat request to an LLM provider.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// StreamChat sends a streaming chat request.
	StreamChat(ctx context.Context, req *ChatRequest) (<-chan *ChatChunk, error)

	// GetProviders lists available LLM providers.
	GetProviders(ctx context.Context) ([]*models.LLMProvider, error)

	// GetProvider retrieves a provider by name.
	GetProvider(ctx context.Context, name string) (*models.LLMProvider, error)

	// CreateProvider creates a new LLM provider.
	CreateProvider(ctx context.Context, provider *models.LLMProvider) error

	// UpdateProvider updates an existing provider.
	UpdateProvider(ctx context.Context, provider *models.LLMProvider) error

	// DeleteProvider deletes a provider.
	DeleteProvider(ctx context.Context, id int64) error

	// GetUsageStats retrieves usage statistics for LLM providers.
	GetUsageStats(ctx context.Context, filter *UsageFilter) (*UsageStats, error)
}

// ChatRequest represents a request to an LLM.
type ChatRequest struct {
	ProviderName string               `json:"provider_name,omitempty"`
	Model        string               `json:"model"`
	Messages     []models.ChatMessage `json:"messages"`
	MaxTokens    int                  `json:"max_tokens,omitempty"`
	Temperature  float64              `json:"temperature,omitempty"`
	TaskType     string               `json:"task_type,omitempty"`
	EvaluationID *int64               `json:"evaluation_id,omitempty"`
}

// ChatResponse represents a response from an LLM.
type ChatResponse struct {
	Content          string `json:"content"`
	FinishReason     string `json:"finish_reason"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	Model            string `json:"model"`
	Provider         string `json:"provider"`
	DurationMs       int    `json:"duration_ms"`
}

// ChatChunk represents a single chunk in a streaming response.
type ChatChunk struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
	Error        error  `json:"error,omitempty"`
}

// UsageFilter defines filtering options for usage statistics.
type UsageFilter struct {
	Provider string
	Model    string
	FromTime int64
	ToTime   int64
}

// UsageStats represents LLM usage statistics.
type UsageStats struct {
	TotalRequests    int64   `json:"total_requests"`
	TotalTokens      int64   `json:"total_tokens"`
	TotalCost        float64 `json:"total_cost"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	ErrorRate        float64 `json:"error_rate"`
	ByProvider       map[string]*ProviderStats `json:"by_provider"`
}

// ProviderStats represents statistics for a single provider.
type ProviderStats struct {
	Provider     string  `json:"provider"`
	Requests     int64   `json:"requests"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
}

// AuditService defines the interface for audit logging.
type AuditService interface {
	// Log creates an audit log entry.
	Log(ctx context.Context, entry *AuditEntry) error

	// Query queries audit logs with filtering.
	Query(ctx context.Context, filter *AuditFilter) ([]*models.AuditLog, int64, error)

	// GetLog retrieves a single audit log by ID.
	GetLog(ctx context.Context, id int64) (*models.AuditLog, error)
}

// AuditEntry represents an audit log entry to be created.
type AuditEntry struct {
	EventType    string
	ActorType    string
	ActorID      string
	ActorName    string
	ResourceType string
	ResourceID   string
	ResourceName string
	Action       string
	Result       string
	Details      map[string]interface{}
	IPAddress    string
	UserAgent    string
	RequestID    string
	DurationMs   *int
}

// AuditFilter defines filtering options for querying audit logs.
type AuditFilter struct {
	EventType    string
	ActorType    string
	ActorID      string
	ResourceType string
	ResourceID   string
	Action       string
	Result       string
	FromTime     int64
	ToTime       int64
	Page         int
	PageSize     int
	SortBy       string
	SortOrder    string
}
