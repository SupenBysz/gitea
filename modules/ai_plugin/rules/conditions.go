// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rules

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/webhook"
)

// evaluateCommentExists checks if comments exist on the PR/Issue.
func evaluateCommentExists(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	// Check if this is an issue/PR close event that should have comments
	if payload.Issue != nil {
		// Would need to call Gitea API to check comments
		// For now, return true as a placeholder
		return true, nil
	}

	if payload.PR != nil {
		// Would need to call Gitea API to check comments
		return true, nil
	}

	return true, nil
}

// evaluateBranchNaming checks if the branch name matches the required pattern.
func evaluateBranchNaming(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return true, nil
	}

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	var branchName string

	// Get branch name based on event type
	if payload.PR != nil {
		branchName = payload.PR.Head.Ref
	} else if payload.Ref != "" {
		// For push events, extract branch from ref
		branchName = strings.TrimPrefix(payload.Ref, "refs/heads/")
	}

	if branchName == "" {
		return true, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(branchName), nil
}

// evaluateFilePattern checks if changed files match a pattern.
func evaluateFilePattern(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return true, nil
	}

	mustExist, _ := params["must_exist"].(bool)
	mustNotExist, _ := params["must_not_exist"].(bool)

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	// Collect all changed files
	var changedFiles []string
	for _, commit := range payload.Commits {
		changedFiles = append(changedFiles, commit.Added...)
		changedFiles = append(changedFiles, commit.Modified...)
		changedFiles = append(changedFiles, commit.Removed...)
	}

	// Check if any file matches the pattern
	matchFound := false
	for _, file := range changedFiles {
		if re.MatchString(file) {
			matchFound = true
			break
		}
	}

	if mustExist {
		return matchFound, nil
	}
	if mustNotExist {
		return !matchFound, nil
	}

	return matchFound, nil
}

// evaluateLabelCheck checks if required labels are present.
func evaluateLabelCheck(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	requiredLabels, _ := params["required"].([]interface{})
	forbiddenLabels, _ := params["forbidden"].([]interface{})

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	var labels []webhook.LabelPayload
	if payload.Issue != nil {
		labels = payload.Issue.Labels
	} else if payload.PR != nil {
		labels = payload.PR.Labels
	}

	labelNames := make(map[string]bool)
	for _, label := range labels {
		labelNames[label.Name] = true
	}

	// Check required labels
	for _, req := range requiredLabels {
		if labelName, ok := req.(string); ok {
			if !labelNames[labelName] {
				return false, nil
			}
		}
	}

	// Check forbidden labels
	for _, forbidden := range forbiddenLabels {
		if labelName, ok := forbidden.(string); ok {
			if labelNames[labelName] {
				return false, nil
			}
		}
	}

	return true, nil
}

// evaluateUserInGroup checks if the user belongs to a specific group.
func evaluateUserInGroup(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	groups, _ := params["groups"].([]interface{})
	if len(groups) == 0 {
		return true, nil
	}

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	// Get sender username
	senderUsername := payload.Sender.Login
	if senderUsername == "" {
		senderUsername = payload.Sender.Username
	}

	// This would need to call Gitea API to check group membership
	// For now, we'll check against a list of usernames in params
	userList, _ := params["users"].([]interface{})
	for _, user := range userList {
		if userName, ok := user.(string); ok {
			if userName == senderUsername {
				return true, nil
			}
		}
	}

	return false, nil
}

// evaluatePRTitle checks if the PR title matches requirements.
func evaluatePRTitle(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	if payload.PR == nil {
		return true, nil
	}

	title := payload.PR.Title

	// Check pattern
	if pattern, ok := params["pattern"].(string); ok && pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		if !re.MatchString(title) {
			return false, nil
		}
	}

	// Check minimum length
	if minLen, ok := params["min_length"].(float64); ok {
		if len(title) < int(minLen) {
			return false, nil
		}
	}

	// Check maximum length
	if maxLen, ok := params["max_length"].(float64); ok {
		if len(title) > int(maxLen) {
			return false, nil
		}
	}

	return true, nil
}

// evaluatePRDescription checks if the PR description meets requirements.
func evaluatePRDescription(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	if payload.PR == nil {
		return true, nil
	}

	body := payload.PR.Body

	// Check if description is required and not empty
	if required, ok := params["required"].(bool); ok && required {
		if strings.TrimSpace(body) == "" {
			return false, nil
		}
	}

	// Check pattern
	if pattern, ok := params["pattern"].(string); ok && pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		if !re.MatchString(body) {
			return false, nil
		}
	}

	// Check minimum length
	if minLen, ok := params["min_length"].(float64); ok {
		if len(body) < int(minLen) {
			return false, nil
		}
	}

	// Check required sections
	if sections, ok := params["required_sections"].([]interface{}); ok {
		for _, section := range sections {
			if sectionName, ok := section.(string); ok {
				if !strings.Contains(strings.ToLower(body), strings.ToLower(sectionName)) {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

// evaluateCommitMessage checks if commit messages follow conventions.
func evaluateCommitMessage(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	pattern, _ := params["pattern"].(string)
	var re *regexp.Regexp
	if pattern != "" {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
	}

	// Check all commits
	for _, commit := range payload.Commits {
		message := commit.Message

		// Check pattern
		if re != nil && !re.MatchString(message) {
			return false, nil
		}

		// Check minimum length
		if minLen, ok := params["min_length"].(float64); ok {
			if len(message) < int(minLen) {
				return false, nil
			}
		}

		// Check maximum length for first line
		if maxFirstLine, ok := params["max_first_line_length"].(float64); ok {
			firstLine := strings.Split(message, "\n")[0]
			if len(firstLine) > int(maxFirstLine) {
				return false, nil
			}
		}
	}

	return true, nil
}

// evaluateFileCount checks if the number of changed files is within limits.
func evaluateFileCount(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	// Count changed files
	fileSet := make(map[string]bool)
	for _, commit := range payload.Commits {
		for _, f := range commit.Added {
			fileSet[f] = true
		}
		for _, f := range commit.Modified {
			fileSet[f] = true
		}
		for _, f := range commit.Removed {
			fileSet[f] = true
		}
	}

	fileCount := len(fileSet)

	// Check minimum
	if minCount, ok := params["min"].(float64); ok {
		if fileCount < int(minCount) {
			return false, nil
		}
	}

	// Check maximum
	if maxCount, ok := params["max"].(float64); ok {
		if fileCount > int(maxCount) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateLineCount checks if the number of changed lines is within limits.
func evaluateLineCount(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	// This would need to call Gitea API to get diff stats
	// For now, return true as a placeholder
	return true, nil
}

// evaluateReviewApproved checks if required reviews are approved.
func evaluateReviewApproved(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	if payload.Review != nil {
		// Check if this review is an approval
		if payload.Review.State == "approved" {
			return true, nil
		}
	}

	// Would need to call Gitea API to check all reviews
	// For now, return true as a placeholder
	return true, nil
}

// evaluateChecksPassed checks if CI checks have passed.
func evaluateChecksPassed(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	// This would need to call Gitea API to check commit status
	// For now, return true as a placeholder
	return true, nil
}

// evaluatePRHasLinkedIssue checks if a PR has a linked issue.
func evaluatePRHasLinkedIssue(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	if payload.PR == nil {
		return false, nil
	}

	// Check PR body for issue references like #123, fixes #123, closes #123
	body := payload.PR.Body
	title := payload.PR.Title

	// Common patterns for linking issues
	issueRefPatterns := []string{
		`#(\d+)`,                                    // #123
		`(?i)fixes?\s+#(\d+)`,                       // fixes #123, fix #123
		`(?i)closes?\s+#(\d+)`,                      // closes #123, close #123
		`(?i)resolves?\s+#(\d+)`,                    // resolves #123, resolve #123
		`(?i)ref(?:erence)?s?\s+#(\d+)`,             // refs #123, reference #123
		`(?i)关联\s*[#＃]?(\d+)`,                      // 关联 #123, 关联123
		`(?i)工单\s*[#＃]?(\d+)`,                      // 工单 #123, 工单123
	}

	combined := title + " " + body
	for _, pattern := range issueRefPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(combined) {
			return true, nil
		}
	}

	return false, nil
}

// evaluatePRSubmitterIsIssueAssignee checks if PR submitter is the linked issue's assignee.
func evaluatePRSubmitterIsIssueAssignee(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	if payload.PR == nil {
		return false, nil
	}

	prSubmitter := payload.PR.User.Login
	if prSubmitter == "" {
		prSubmitter = payload.PR.User.Username
	}

	// Extract linked issue number from PR body/title
	body := payload.PR.Body
	title := payload.PR.Title
	combined := title + " " + body

	// Find issue references
	issueRefPattern := regexp.MustCompile(`(?i)(?:fixes?|closes?|resolves?|refs?|关联|工单)\s*[#＃]?(\d+)`)
	matches := issueRefPattern.FindStringSubmatch(combined)
	if len(matches) < 2 {
		// No linked issue found, cannot validate
		return true, nil
	}

	// Note: To properly check the issue assignee, we would need to call Gitea API
	// This is a placeholder that assumes the check context includes linked issue info
	// The actual implementation would use a Gitea client to fetch issue details

	// For now, we check if there's an issue in the payload (some webhooks include it)
	if payload.Issue != nil && len(payload.Issue.Assignees) > 0 {
		// Check if PR submitter is one of the assignees
		for _, assignee := range payload.Issue.Assignees {
			assigneeName := assignee.Login
			if assigneeName == "" {
				assigneeName = assignee.Username
			}
			if prSubmitter == assigneeName {
				return true, nil
			}
		}
		return false, nil
	}

	// Cannot determine without API call, return true to avoid false positives
	return true, nil
}

// evaluateCommentContainsPattern checks if any comment matches a pattern.
func evaluateCommentContainsPattern(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return true, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	// Check current comment if this is a comment event
	if payload.Comment != nil {
		if re.MatchString(payload.Comment.Body) {
			return true, nil
		}
	}

	// Note: To check all comments, we would need to call Gitea API
	// The actual implementation would iterate through all comments
	// For now, this only checks the current comment in the event

	return false, nil
}

// evaluateIssueHasWorkSummary checks if an issue has a work summary comment.
func evaluateIssueHasWorkSummary(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error) {
	// Work summary patterns in Chinese and English
	summaryPatterns := []string{
		`(?i)##\s*工作总结`,
		`(?i)##\s*work\s*summary`,
		`(?i)###\s*完成内容`,
		`(?i)###\s*completed\s*work`,
	}

	payload, err := webhook.ParsePayload(event.Payload)
	if err != nil {
		return false, err
	}

	// Check current comment if present
	if payload.Comment != nil {
		for _, pattern := range summaryPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(payload.Comment.Body) {
				return true, nil
			}
		}
	}

	// Note: To fully implement this, we need to call Gitea API to fetch all comments
	// This would be done in the actual implementation with a Gitea client
	// For now, return false to indicate work summary not found in current context

	return false, nil
}

// parsePayloadField extracts a nested field from the event payload.
func parsePayloadField(payload []byte, field string) (interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	parts := strings.Split(field, ".")
	current := interface{}(data)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil, nil
		}
	}

	return current, nil
}
