// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rules

import (
	"context"
	"fmt"

	"code.gitea.io/gitea/modules/ai_plugin/models"
)

// ActionContext provides context for action execution.
type ActionContext struct {
	NotificationService interface{} // Will be injected
	GiteaClient         interface{} // Will be injected
}

// Global action context - will be set during initialization
var actionCtx *ActionContext

// SetActionContext sets the global action context.
func SetActionContext(ctx *ActionContext) {
	actionCtx = ctx
}

// executeBlock prevents the action from proceeding (e.g., block PR merge).
func executeBlock(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	message, _ := params["message"].(string)
	if message == "" {
		message = "Action blocked by compliance rule"
	}

	// In a real implementation, this would:
	// 1. Add a failing commit status to block merge
	// 2. Post a comment explaining why it's blocked

	// For now, log the block action
	fmt.Printf("[BLOCK] Event %s: %s\n", event.EventID, message)

	return nil
}

// executeWarn issues a warning without blocking.
func executeWarn(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	message, _ := params["message"].(string)
	if message == "" {
		message = "Warning: potential compliance issue detected"
	}

	// In a real implementation, this would:
	// 1. Add a warning commit status
	// 2. Post a comment with the warning

	fmt.Printf("[WARN] Event %s: %s\n", event.EventID, message)

	return nil
}

// executeRemind sends a reminder to relevant parties.
func executeRemind(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	message, _ := params["message"].(string)
	targets, _ := params["targets"].([]interface{})

	if message == "" {
		message = "Reminder: Please review the compliance requirements"
	}

	// In a real implementation, this would send reminders via configured channels
	fmt.Printf("[REMIND] Event %s to %v: %s\n", event.EventID, targets, message)

	return nil
}

// executeNotify sends a notification to the configured channel.
func executeNotify(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	channel, _ := params["channel"].(string)
	message, _ := params["message"].(string)
	templateName, _ := params["template"].(string)

	if channel == "" {
		channel = "telegram" // Default channel
	}

	// In a real implementation, this would use the notification service
	fmt.Printf("[NOTIFY] Channel %s, Event %s, Template %s: %s\n", channel, event.EventID, templateName, message)

	return nil
}

// executeAutoFix attempts to automatically fix the compliance issue.
func executeAutoFix(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	fixType, _ := params["type"].(string)

	// Auto-fix types could include:
	// - add_label: Add missing labels
	// - update_title: Fix title format
	// - add_reviewer: Request required reviewers
	// - fix_branch_name: Suggest correct branch name

	fmt.Printf("[AUTO_FIX] Event %s, Fix type: %s\n", event.EventID, fixType)

	switch fixType {
	case "add_label":
		return executeAddLabel(ctx, event, params)
	case "add_reviewer":
		return executeAddReviewer(ctx, event, params)
	default:
		return fmt.Errorf("unknown auto-fix type: %s", fixType)
	}
}

// executeLabel adds or removes labels.
func executeLabel(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	addLabels, _ := params["add"].([]interface{})
	removeLabels, _ := params["remove"].([]interface{})

	// In a real implementation, this would call Gitea API
	fmt.Printf("[LABEL] Event %s, Add: %v, Remove: %v\n", event.EventID, addLabels, removeLabels)

	return nil
}

// executeAssign assigns users to the issue/PR.
func executeAssign(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	assignees, _ := params["assignees"].([]interface{})

	// In a real implementation, this would call Gitea API
	fmt.Printf("[ASSIGN] Event %s, Assignees: %v\n", event.EventID, assignees)

	return nil
}

// executeComment posts a comment on the issue/PR.
func executeComment(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	body, _ := params["body"].(string)
	templateName, _ := params["template"].(string)

	if body == "" && templateName == "" {
		return fmt.Errorf("comment body or template is required")
	}

	// In a real implementation, this would:
	// 1. Render template if specified
	// 2. Call Gitea API to post comment

	fmt.Printf("[COMMENT] Event %s, Template: %s, Body: %s\n", event.EventID, templateName, body)

	return nil
}

// executeAddLabel adds labels to the issue/PR.
func executeAddLabel(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	labels, _ := params["labels"].([]interface{})

	// In a real implementation, this would call Gitea API
	fmt.Printf("[ADD_LABEL] Event %s, Labels: %v\n", event.EventID, labels)

	return nil
}

// executeAddReviewer requests reviewers for the PR.
func executeAddReviewer(ctx context.Context, event *models.Event, params map[string]interface{}) error {
	reviewers, _ := params["reviewers"].([]interface{})

	// In a real implementation, this would call Gitea API
	fmt.Printf("[ADD_REVIEWER] Event %s, Reviewers: %v\n", event.EventID, reviewers)

	return nil
}
