// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"code.gitea.io/gitea/modules/ai_plugin/config"
	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/services"
)

// Service implements the services.NotificationService interface.
type Service struct {
	config       *config.NotificationConfig
	telegramBot  *TelegramBot
	notifRepo    *services.NotificationRepository
	templateRepo *TemplateRepository
}

// TemplateRepository provides access to notification templates.
type TemplateRepository struct {
	db interface{} // Will be injected
}

// NewService creates a new notification service.
func NewService(cfg *config.NotificationConfig, notifRepo *services.NotificationRepository) *Service {
	svc := &Service{
		config:    cfg,
		notifRepo: notifRepo,
	}

	if cfg.Telegram.Enabled && cfg.Telegram.BotToken != "" {
		svc.telegramBot = NewTelegramBot(&cfg.Telegram)
	}

	return svc
}

// Send sends a notification to the specified channel.
func (s *Service) Send(ctx context.Context, req *services.NotificationRequest) (*services.NotificationResponse, error) {
	// Render content if using template
	content := req.Content
	if req.TemplateName != "" {
		rendered, err := s.renderTemplate(ctx, req.TemplateName, req.Channel, req.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render template: %w", err)
		}
		content = rendered
	}

	// Create notification record
	notif := &models.Notification{
		EventID:      req.EventID,
		EvaluationID: req.EvaluationID,
		Channel:      req.Channel,
		TemplateID:   req.TemplateID,
		Recipient:    req.Recipient,
		Content:      content,
		Status:       models.NotificationStatusPending,
		CreatedAt:    time.Now(),
	}

	if s.notifRepo != nil {
		if err := s.notifRepo.Create(ctx, notif); err != nil {
			return nil, fmt.Errorf("failed to create notification record: %w", err)
		}
	}

	// Send notification based on channel
	var messageID string
	var sendErr error

	switch req.Channel {
	case models.ChannelTelegram:
		messageID, sendErr = s.sendTelegram(ctx, req.Recipient, content)
	case models.ChannelGiteaComment:
		messageID, sendErr = s.sendGiteaComment(ctx, req.Recipient, content)
	case models.ChannelWebhook:
		messageID, sendErr = s.sendWebhook(ctx, req.Recipient, content)
	case models.ChannelEmail:
		messageID, sendErr = s.sendEmail(ctx, req.Recipient, content)
	case models.ChannelSlack:
		messageID, sendErr = s.sendSlack(ctx, req.Recipient, content)
	default:
		sendErr = fmt.Errorf("unsupported channel: %s", req.Channel)
	}

	// Update notification status
	if sendErr != nil {
		notif.Status = models.NotificationStatusFailed
		notif.ErrorMessage = sendErr.Error()
	} else {
		notif.Status = models.NotificationStatusSent
		notif.MessageID = messageID
		now := time.Now()
		notif.SentAt = &now
	}

	if s.notifRepo != nil {
		if err := s.notifRepo.Update(ctx, notif); err != nil {
			// Log but don't fail
		}
	}

	return &services.NotificationResponse{
		ID:        notif.ID,
		Status:    notif.Status,
		MessageID: messageID,
		Error:     notif.ErrorMessage,
	}, sendErr
}

// SendBatch sends multiple notifications.
func (s *Service) SendBatch(ctx context.Context, reqs []*services.NotificationRequest) ([]*services.NotificationResponse, error) {
	responses := make([]*services.NotificationResponse, len(reqs))

	for i, req := range reqs {
		resp, err := s.Send(ctx, req)
		if err != nil {
			responses[i] = &services.NotificationResponse{
				Status: models.NotificationStatusFailed,
				Error:  err.Error(),
			}
		} else {
			responses[i] = resp
		}
	}

	return responses, nil
}

// GetNotification retrieves a notification by ID.
func (s *Service) GetNotification(ctx context.Context, id int64) (*models.Notification, error) {
	// Implementation would use notifRepo
	return nil, fmt.Errorf("not implemented")
}

// ListNotifications lists notifications with filtering.
func (s *Service) ListNotifications(ctx context.Context, filter *services.NotificationFilter) ([]*models.Notification, int64, error) {
	// Implementation would use notifRepo
	return nil, 0, fmt.Errorf("not implemented")
}

// RetryFailed retries failed notifications.
func (s *Service) RetryFailed(ctx context.Context) (int, error) {
	if s.notifRepo == nil {
		return 0, nil
	}

	failed, err := s.notifRepo.GetFailed(ctx, 100)
	if err != nil {
		return 0, err
	}

	retried := 0
	for _, notif := range failed {
		// Retry sending
		req := &services.NotificationRequest{
			Channel:   notif.Channel,
			Recipient: notif.Recipient,
			Content:   notif.Content,
		}

		_, sendErr := s.Send(ctx, req)
		if sendErr == nil {
			retried++
		} else {
			notif.RetryCount++
			notif.ErrorMessage = sendErr.Error()
			s.notifRepo.Update(ctx, notif)
		}
	}

	return retried, nil
}

// sendTelegram sends a message via Telegram.
func (s *Service) sendTelegram(ctx context.Context, recipient string, content string) (string, error) {
	if s.telegramBot == nil {
		return "", fmt.Errorf("telegram bot is not configured")
	}

	// Parse recipient as chat ID
	chatID := s.config.Telegram.DefaultChatID
	if recipient != "" {
		if id, err := strconv.ParseInt(recipient, 10, 64); err == nil {
			chatID = id
		}
	}

	result, err := s.telegramBot.SendText(ctx, chatID, content)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(result.MessageID, 10), nil
}

// sendGiteaComment posts a comment on Gitea.
func (s *Service) sendGiteaComment(ctx context.Context, recipient string, content string) (string, error) {
	// recipient format: "owner/repo/issues/123" or "owner/repo/pulls/123"
	// Would need to call Gitea API
	return "", fmt.Errorf("gitea comment not implemented")
}

// sendWebhook sends a notification via webhook.
func (s *Service) sendWebhook(ctx context.Context, recipient string, content string) (string, error) {
	// Would send HTTP POST to the webhook URL
	return "", fmt.Errorf("webhook not implemented")
}

// sendEmail sends a notification via email.
func (s *Service) sendEmail(ctx context.Context, recipient string, content string) (string, error) {
	// Would use SMTP to send email
	return "", fmt.Errorf("email not implemented")
}

// sendSlack sends a notification via Slack.
func (s *Service) sendSlack(ctx context.Context, recipient string, content string) (string, error) {
	// Would use Slack webhook or API
	return "", fmt.Errorf("slack not implemented")
}

// renderTemplate renders a notification template with variables.
func (s *Service) renderTemplate(ctx context.Context, templateName string, channel string, variables map[string]interface{}) (string, error) {
	// For now, use built-in templates
	tmplContent, ok := builtinTemplates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Get channel-specific content
	channelContent, ok := tmplContent[channel]
	if !ok {
		// Fallback to default channel
		channelContent, ok = tmplContent["default"]
		if !ok {
			return "", fmt.Errorf("no template content for channel: %s", channel)
		}
	}

	// Parse and execute template
	tmpl, err := template.New(templateName).Parse(channelContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// builtinTemplates contains built-in notification templates.
var builtinTemplates = map[string]map[string]string{
	"rule_violation": {
		"telegram": `⚠️ <b>Rule Violation Detected</b>

<b>Rule:</b> {{.RuleName}}
<b>Repository:</b> {{.Repository}}
<b>Event:</b> {{.EventType}}
<b>User:</b> {{.Sender}}

<b>Violation:</b>
{{.Message}}

{{if .URL}}<a href="{{.URL}}">View Details</a>{{end}}`,
		"gitea_comment": `### ⚠️ Rule Violation: {{.RuleName}}

{{.Message}}

---
*This comment was generated automatically by the AI Collaboration Plugin.*`,
		"default": `Rule Violation: {{.RuleName}} - {{.Message}}`,
	},
	"compliance_passed": {
		"telegram": `✅ <b>Compliance Check Passed</b>

<b>Repository:</b> {{.Repository}}
<b>Event:</b> {{.EventType}}
<b>User:</b> {{.Sender}}

All compliance rules have passed.`,
		"default": `Compliance Check Passed for {{.Repository}}`,
	},
	"pr_opened": {
		"telegram": `📝 <b>New Pull Request</b>

<b>Repository:</b> {{.Repository}}
<b>Title:</b> {{.Title}}
<b>Author:</b> {{.Author}}
<b>Branch:</b> {{.SourceBranch}} → {{.TargetBranch}}

{{if .Description}}{{.Description}}{{end}}

<a href="{{.URL}}">View PR</a>`,
		"default": `New PR: {{.Title}} by {{.Author}}`,
	},
	"issue_closed": {
		"telegram": `🔒 <b>Issue Closed</b>

<b>Repository:</b> {{.Repository}}
<b>Issue:</b> #{{.Number}} - {{.Title}}
<b>Closed by:</b> {{.ClosedBy}}

<a href="{{.URL}}">View Issue</a>`,
		"default": `Issue #{{.Number}} closed: {{.Title}}`,
	},
	"work_summary_required": {
		"telegram": `📋 <b>Work Summary Required</b>

<b>Issue:</b> #{{.Number}} - {{.Title}}
<b>Repository:</b> {{.Repository}}
<b>Assignee:</b> {{.Assignee}}

This issue is being closed without a work summary comment. Please add a summary of the work done before closing.

<a href="{{.URL}}">View Issue</a>`,
		"gitea_comment": `### 📋 Work Summary Required

This issue is being closed without a work summary. Please add a comment describing:

1. What was accomplished
2. Any relevant implementation details
3. Next steps (if applicable)

---
*This comment was generated automatically by the AI Collaboration Plugin.*`,
		"default": `Work summary required for issue #{{.Number}}`,
	},
}
