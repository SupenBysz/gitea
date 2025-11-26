// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// Event represents a Gitea webhook event received by the plugin.
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

// TableName returns the table name for Event model.
func (Event) TableName() string {
	return "ai_plugin_event"
}

// Gitea webhook event types.
const (
	EventTypeCreate             = "create"
	EventTypeDelete             = "delete"
	EventTypeFork               = "fork"
	EventTypePush               = "push"
	EventTypeIssues             = "issues"
	EventTypeIssueAssign        = "issue_assign"
	EventTypeIssueLabel         = "issue_label"
	EventTypeIssueMilestone     = "issue_milestone"
	EventTypeIssueComment       = "issue_comment"
	EventTypePullRequest        = "pull_request"
	EventTypePullRequestAssign  = "pull_request_assign"
	EventTypePullRequestLabel   = "pull_request_label"
	EventTypePullRequestMilestone = "pull_request_milestone"
	EventTypePullRequestComment = "pull_request_comment"
	EventTypePullRequestReview  = "pull_request_review"
	EventTypePullRequestSync    = "pull_request_sync"
	EventTypeRepository         = "repository"
	EventTypeRelease            = "release"
	EventTypePackage            = "package"
	EventTypeWiki               = "wiki"
)

// Common webhook actions.
const (
	ActionOpened      = "opened"
	ActionClosed      = "closed"
	ActionReopened    = "reopened"
	ActionEdited      = "edited"
	ActionAssigned    = "assigned"
	ActionUnassigned  = "unassigned"
	ActionLabeled     = "labeled"
	ActionUnlabeled   = "unlabeled"
	ActionMilestoned  = "milestoned"
	ActionDemilestoned = "demilestoned"
	ActionSynchronize = "synchronize"
	ActionCreated     = "created"
	ActionDeleted     = "deleted"
	ActionReviewed    = "reviewed"
	ActionApproved    = "approved"
	ActionRejected    = "rejected"
	ActionCommented   = "commented"
	ActionMerged      = "merged"
)

// Evaluation represents the result of evaluating a rule against an event.
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

	// Relations
	Rule  *Rule  `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Event *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

// TableName returns the table name for Evaluation model.
func (Evaluation) TableName() string {
	return "ai_plugin_evaluation"
}

// Violation represents a single rule violation found during evaluation.
type Violation struct {
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning, info
	Field    string `json:"field,omitempty"`
}

// ActionResult represents the result of executing a single action.
type ActionResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Violation severity levels.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)
