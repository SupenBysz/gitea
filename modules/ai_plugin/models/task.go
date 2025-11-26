// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// ScheduledTask represents a scheduled background task.
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

// TableName returns the table name for ScheduledTask model.
func (ScheduledTask) TableName() string {
	return "ai_plugin_scheduled_task"
}

// Scheduled task types.
const (
	TaskTypeComplianceReport = "compliance_report"
	TaskTypeDataCleanup      = "data_cleanup"
	TaskTypeSync             = "sync"
	TaskTypeMetricsAggregate = "metrics_aggregate"
	TaskTypeCustom           = "custom"
)

// TaskExecution represents a single execution record of a scheduled task.
type TaskExecution struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID       int64      `gorm:"index;not null" json:"task_id"`
	StartTime    time.Time  `gorm:"not null;index" json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Status       string     `gorm:"size:20;default:running;index;not null" json:"status"`
	Output       string     `gorm:"type:text" json:"output"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`

	// Relations
	Task *ScheduledTask `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// TableName returns the table name for TaskExecution model.
func (TaskExecution) TableName() string {
	return "ai_plugin_task_execution"
}

// Task execution statuses.
const (
	TaskStatusRunning   = "running"
	TaskStatusSuccess   = "success"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)
