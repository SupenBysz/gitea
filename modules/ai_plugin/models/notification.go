// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// Template represents a notification template for various channels.
type Template struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"size:255;uniqueIndex;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Channels    datatypes.JSON `gorm:"type:jsonb;default:[]" json:"channels"`
	Content     datatypes.JSON `gorm:"type:jsonb;not null" json:"content"`
	Variables   datatypes.JSON `gorm:"type:jsonb;default:[]" json:"variables"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy   string         `gorm:"size:255" json:"created_by"`
}

// TableName returns the table name for Template model.
func (Template) TableName() string {
	return "ai_plugin_template"
}

// TemplateContent represents the content for each notification channel.
type TemplateContent struct {
	Telegram     string `json:"telegram,omitempty"`
	GiteaComment string `json:"gitea_comment,omitempty"`
	Webhook      string `json:"webhook,omitempty"`
	Email        string `json:"email,omitempty"`
	Slack        string `json:"slack,omitempty"`
}

// Notification represents a notification record sent to a recipient.
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

// TableName returns the table name for Notification model.
func (Notification) TableName() string {
	return "ai_plugin_notification"
}

// Notification channels.
const (
	ChannelTelegram     = "telegram"
	ChannelGiteaComment = "gitea_comment"
	ChannelWebhook      = "webhook"
	ChannelEmail        = "email"
	ChannelSlack        = "slack"
)

// Notification statuses.
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)

// MaxRetryCount defines the maximum number of retry attempts for notifications.
const MaxRetryCount = 3
