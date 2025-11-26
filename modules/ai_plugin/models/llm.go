// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// LLMProvider represents an LLM provider configuration (e.g., Claude, GPT).
type LLMProvider struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Type      string         `gorm:"size:50;not null" json:"type"`
	Config    datatypes.JSON `gorm:"type:jsonb;not null" json:"config"`
	Enabled   bool           `gorm:"default:true;not null" json:"enabled"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for LLMProvider model.
func (LLMProvider) TableName() string {
	return "ai_plugin_llm_provider"
}

// LLM provider types.
const (
	LLMProviderTypeAnthropic = "anthropic"
	LLMProviderTypeOpenAI    = "openai"
	LLMProviderTypeAzure     = "azure"
	LLMProviderTypeCustom    = "custom"
)

// LLMProviderConfig represents the configuration for an LLM provider.
type LLMProviderConfig struct {
	APIKey       string   `json:"api_key"`
	BaseURL      string   `json:"base_url,omitempty"`
	Models       []string `json:"models,omitempty"`
	DefaultModel string   `json:"default_model,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	Timeout      int      `json:"timeout,omitempty"` // in seconds
}

// LLMRequest represents a request made to an LLM provider.
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

	// Relations
	Provider *LLMProvider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

// TableName returns the table name for LLMRequest model.
func (LLMRequest) TableName() string {
	return "ai_plugin_llm_request"
}

// LLM task types.
const (
	LLMTaskTypeCodeReview      = "code_review"
	LLMTaskTypeCommentAnalysis = "comment_analysis"
	LLMTaskTypeSummary         = "summary"
	LLMTaskTypeConflictAnalysis = "conflict_analysis"
	LLMTaskTypeSecurityCheck   = "security_check"
	LLMTaskTypeCustom          = "custom"
)

// LLMRequest statuses.
const (
	LLMRequestStatusPending = "pending"
	LLMRequestStatusSuccess = "success"
	LLMRequestStatusFailed  = "failed"
)

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// Message roles.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)
