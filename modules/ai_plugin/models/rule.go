// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// Rule represents a compliance rule configuration.
// Rules define what events to listen for and what actions to take when conditions are met.
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

	// Relations
	Conditions []RuleCondition `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"conditions"`
	Actions    []RuleAction    `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"actions"`
}

// TableName returns the table name for Rule model.
func (Rule) TableName() string {
	return "ai_plugin_rule"
}

// RuleTrigger represents the trigger configuration for a rule.
type RuleTrigger struct {
	EventTypes []string          `json:"event_types"`
	Filters    []RuleTriggerFilter `json:"filters,omitempty"`
}

// RuleTriggerFilter represents a filter condition for triggering rules.
type RuleTriggerFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, contains, starts_with, ends_with, regex, in, not_in
	Value    interface{} `json:"value"`
}

// RuleMetadata represents additional metadata for a rule.
type RuleMetadata struct {
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Author   string   `json:"author,omitempty"`
	Version  string   `json:"version,omitempty"`
}

// RuleCondition represents a single condition that must be evaluated for a rule.
type RuleCondition struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID     int64          `gorm:"index;not null" json:"rule_id"`
	Type       string         `gorm:"size:100;not null" json:"type"`
	Params     datatypes.JSON `gorm:"type:jsonb;default:{}" json:"params"`
	Negate     bool           `gorm:"default:false;not null" json:"negate"`
	OrderIndex int            `gorm:"default:0;not null" json:"order_index"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for RuleCondition model.
func (RuleCondition) TableName() string {
	return "ai_plugin_rule_condition"
}

// Condition types supported by the rule engine.
const (
	ConditionTypeCommentExists          = "comment_exists"
	ConditionTypeBranchNaming           = "branch_naming"
	ConditionTypeFilePattern            = "file_pattern"
	ConditionTypeLabelCheck             = "label_check"
	ConditionTypeUserInGroup            = "user_in_group"
	ConditionTypePRTitle                = "pr_title"
	ConditionTypePRDescription          = "pr_description"
	ConditionTypeCommitMessage          = "commit_message"
	ConditionTypeFileCount              = "file_count"
	ConditionTypeLineCount              = "line_count"
	ConditionTypeReviewApproved         = "review_approved"
	ConditionTypeChecksPassed           = "checks_passed"
	ConditionTypeCustomExpression       = "custom_expression"
	ConditionTypePRHasLinkedIssue       = "pr_has_linked_issue"
	ConditionTypePRSubmitterIsAssignee  = "pr_submitter_is_issue_assignee"
	ConditionTypeCommentContainsPattern = "comment_contains_pattern"
	ConditionTypeIssueHasWorkSummary    = "issue_has_work_summary"
)

// RuleAction represents an action to be executed when rule conditions are evaluated.
type RuleAction struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID      int64          `gorm:"index;not null" json:"rule_id"`
	Type        string         `gorm:"size:100;not null" json:"type"`
	WhenTrigger string         `gorm:"size:50;default:condition_failed;not null" json:"when_trigger"`
	Params      datatypes.JSON `gorm:"type:jsonb;default:{}" json:"params"`
	OrderIndex  int            `gorm:"default:0;not null" json:"order_index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for RuleAction model.
func (RuleAction) TableName() string {
	return "ai_plugin_rule_action"
}

// Action types supported by the rule engine.
const (
	ActionTypeBlock    = "block"
	ActionTypeWarn     = "warn"
	ActionTypeRemind   = "remind"
	ActionTypeNotify   = "notify"
	ActionTypeAutoFix  = "auto_fix"
	ActionTypeLabel    = "label"
	ActionTypeAssign   = "assign"
	ActionTypeComment  = "comment"
	ActionTypeApprove  = "approve"
	ActionTypeReject   = "reject"
)

// Action trigger timings.
const (
	WhenConditionPassed = "condition_passed"
	WhenConditionFailed = "condition_failed"
	WhenAlways          = "always"
)

// Condition operators.
const (
	OperatorAnd = "and"
	OperatorOr  = "or"
)
