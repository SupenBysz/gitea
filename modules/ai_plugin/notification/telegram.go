// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package notification provides notification services for various channels.
package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"code.gitea.io/gitea/modules/ai_plugin/config"
)

// TelegramBot provides Telegram notification functionality.
type TelegramBot struct {
	token         string
	defaultChatID int64
	client        *http.Client
	baseURL       string
}

// NewTelegramBot creates a new TelegramBot instance.
func NewTelegramBot(cfg *config.TelegramConfig) *TelegramBot {
	return &TelegramBot{
		token:         cfg.BotToken,
		defaultChatID: cfg.DefaultChatID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.telegram.org",
	}
}

// TelegramMessage represents a message to send via Telegram.
type TelegramMessage struct {
	ChatID                int64       `json:"chat_id"`
	Text                  string      `json:"text"`
	ParseMode             string      `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool        `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool        `json:"disable_notification,omitempty"`
	ReplyToMessageID      int64       `json:"reply_to_message_id,omitempty"`
	ReplyMarkup           interface{} `json:"reply_markup,omitempty"`
}

// TelegramResponse represents the response from Telegram API.
type TelegramResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

// TelegramMessageResult represents a sent message result.
type TelegramMessageResult struct {
	MessageID int64 `json:"message_id"`
	Chat      struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Date int64 `json:"date"`
}

// SendMessage sends a text message to a Telegram chat.
func (b *TelegramBot) SendMessage(ctx context.Context, msg *TelegramMessage) (*TelegramMessageResult, error) {
	if msg.ChatID == 0 {
		msg.ChatID = b.defaultChatID
	}

	if msg.ChatID == 0 {
		return nil, fmt.Errorf("chat_id is required")
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", b.baseURL, b.token)

	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(respBody, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", telegramResp.Description, telegramResp.ErrorCode)
	}

	var result TelegramMessageResult
	if err := json.Unmarshal(telegramResp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	return &result, nil
}

// SendText is a convenience method to send a simple text message.
func (b *TelegramBot) SendText(ctx context.Context, chatID int64, text string) (*TelegramMessageResult, error) {
	return b.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	})
}

// SendMarkdown sends a message with Markdown formatting.
func (b *TelegramBot) SendMarkdown(ctx context.Context, chatID int64, text string) (*TelegramMessageResult, error) {
	return b.SendMessage(ctx, &TelegramMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "MarkdownV2",
	})
}

// InlineKeyboard represents an inline keyboard for Telegram messages.
type InlineKeyboard struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents a button in an inline keyboard.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// SendWithButtons sends a message with inline keyboard buttons.
func (b *TelegramBot) SendWithButtons(ctx context.Context, chatID int64, text string, buttons [][]InlineKeyboardButton) (*TelegramMessageResult, error) {
	keyboard := &InlineKeyboard{
		InlineKeyboard: buttons,
	}

	return b.SendMessage(ctx, &TelegramMessage{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	})
}

// EditMessageText edits an existing message.
func (b *TelegramBot) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string) error {
	url := fmt.Sprintf("%s/bot%s/editMessageText", b.baseURL, b.token)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(respBody, &telegramResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return fmt.Errorf("telegram API error: %s", telegramResp.Description)
	}

	return nil
}

// DeleteMessage deletes a message.
func (b *TelegramBot) DeleteMessage(ctx context.Context, chatID int64, messageID int64) error {
	url := fmt.Sprintf("%s/bot%s/deleteMessage", b.baseURL, b.token)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetMe retrieves information about the bot.
func (b *TelegramBot) GetMe(ctx context.Context) (*BotInfo, error) {
	url := fmt.Sprintf("%s/bot%s/getMe", b.baseURL, b.token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(respBody, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return nil, fmt.Errorf("telegram API error: %s", telegramResp.Description)
	}

	var info BotInfo
	if err := json.Unmarshal(telegramResp.Result, &info); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	return &info, nil
}

// BotInfo represents information about the Telegram bot.
type BotInfo struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	CanJoinGroups bool `json:"can_join_groups"`
	CanReadAllGroupMessages bool `json:"can_read_all_group_messages"`
	SupportsInlineQueries bool `json:"supports_inline_queries"`
}
