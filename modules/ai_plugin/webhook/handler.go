// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package webhook provides Gitea webhook event handling functionality.
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"code.gitea.io/gitea/modules/ai_plugin/config"
	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/services"

	"github.com/google/uuid"
)

// EventListener implements the services.EventListener interface.
type EventListener struct {
	config       *config.WebhookConfig
	eventRepo    *services.EventRepository
	handlers     map[string][]services.EventHandler
	eventQueue   chan *models.Event
	workerWg     sync.WaitGroup
	stopCh       chan struct{}
	mu           sync.RWMutex
	stats        *listenerStats
}

type listenerStats struct {
	eventsReceived  int64
	eventsProcessed int64
	eventsFailed    int64
}

// NewEventListener creates a new EventListener instance.
func NewEventListener(cfg *config.WebhookConfig, eventRepo *services.EventRepository) *EventListener {
	return &EventListener{
		config:     cfg,
		eventRepo:  eventRepo,
		handlers:   make(map[string][]services.EventHandler),
		eventQueue: make(chan *models.Event, cfg.QueueSize),
		stopCh:     make(chan struct{}),
		stats:      &listenerStats{},
	}
}

// Start starts the event listener workers.
func (l *EventListener) Start(ctx context.Context) error {
	workers := l.config.Workers
	if workers <= 0 {
		workers = 5
	}

	for i := 0; i < workers; i++ {
		l.workerWg.Add(1)
		go l.worker(ctx, i)
	}

	return nil
}

// Stop stops the event listener.
func (l *EventListener) Stop(ctx context.Context) error {
	close(l.stopCh)
	l.workerWg.Wait()
	return nil
}

// worker processes events from the queue.
func (l *EventListener) worker(ctx context.Context, workerID int) {
	defer l.workerWg.Done()

	for {
		select {
		case <-l.stopCh:
			return
		case <-ctx.Done():
			return
		case event := <-l.eventQueue:
			l.processEvent(ctx, event)
		}
	}
}

// processEvent processes a single event.
func (l *EventListener) processEvent(ctx context.Context, event *models.Event) {
	l.mu.RLock()
	handlers := l.handlers[event.EventType]
	allHandlers := l.handlers["*"] // wildcard handlers
	l.mu.RUnlock()

	var processErr error

	// Execute type-specific handlers
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			processErr = err
			atomic.AddInt64(&l.stats.eventsFailed, 1)
		}
	}

	// Execute wildcard handlers
	for _, handler := range allHandlers {
		if err := handler(ctx, event); err != nil {
			if processErr == nil {
				processErr = err
			}
			atomic.AddInt64(&l.stats.eventsFailed, 1)
		}
	}

	// Mark event as processed
	if l.eventRepo != nil {
		if err := l.eventRepo.MarkProcessed(ctx, event.ID, processErr); err != nil {
			// Log error but don't fail
		}
	}

	atomic.AddInt64(&l.stats.eventsProcessed, 1)
}

// HandleWebhook processes an incoming webhook request.
func (l *EventListener) HandleWebhook(ctx context.Context, eventType string, payload []byte, signature string) error {
	atomic.AddInt64(&l.stats.eventsReceived, 1)

	// Validate signature if configured
	if l.config.ValidateSignature && l.config.Secret != "" {
		if !l.validateSignature(payload, signature) {
			return fmt.Errorf("invalid webhook signature")
		}
	}

	// Parse the payload to extract common fields
	var rawPayload map[string]interface{}
	if err := json.Unmarshal(payload, &rawPayload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract repository and sender information
	repository := extractRepository(rawPayload)
	sender := extractSender(rawPayload)
	action := extractAction(rawPayload)

	// Create event record
	event := &models.Event{
		EventID:    generateEventID(),
		EventType:  eventType,
		Action:     action,
		Repository: repository,
		Sender:     sender,
		Payload:    payload,
		Signature:  signature,
		Processed:  false,
		CreatedAt:  time.Now(),
	}

	// Store event in database
	if l.eventRepo != nil {
		if err := l.eventRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}
	}

	// Queue event for processing
	select {
	case l.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event queue is full")
	}
}

// RegisterHandler registers a handler for a specific event type.
// Use "*" as eventType to register a handler for all event types.
func (l *EventListener) RegisterHandler(eventType string, handler services.EventHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.handlers[eventType] = append(l.handlers[eventType], handler)
}

// GetStats returns listener statistics.
func (l *EventListener) GetStats() *services.ListenerStats {
	return &services.ListenerStats{
		EventsReceived:  atomic.LoadInt64(&l.stats.eventsReceived),
		EventsProcessed: atomic.LoadInt64(&l.stats.eventsProcessed),
		EventsFailed:    atomic.LoadInt64(&l.stats.eventsFailed),
		QueueSize:       len(l.eventQueue),
		WorkersActive:   l.config.Workers,
	}
}

// validateSignature validates the webhook signature using HMAC-SHA256.
func (l *EventListener) validateSignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Gitea uses "sha256=" prefix for signatures
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	mac := hmac.New(sha256.New, []byte(l.config.Secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	return uuid.New().String()
}

// extractRepository extracts the repository full name from the payload.
func extractRepository(payload map[string]interface{}) string {
	if repo, ok := payload["repository"].(map[string]interface{}); ok {
		if fullName, ok := repo["full_name"].(string); ok {
			return fullName
		}
	}
	return ""
}

// extractSender extracts the sender username from the payload.
func extractSender(payload map[string]interface{}) string {
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if login, ok := sender["login"].(string); ok {
			return login
		}
		if username, ok := sender["username"].(string); ok {
			return username
		}
	}
	return ""
}

// extractAction extracts the action from the payload.
func extractAction(payload map[string]interface{}) string {
	if action, ok := payload["action"].(string); ok {
		return action
	}
	return ""
}

// GiteaWebhookPayload represents common fields in Gitea webhook payloads.
type GiteaWebhookPayload struct {
	Action     string                 `json:"action,omitempty"`
	Secret     string                 `json:"secret,omitempty"`
	Ref        string                 `json:"ref,omitempty"`
	Before     string                 `json:"before,omitempty"`
	After      string                 `json:"after,omitempty"`
	CompareURL string                 `json:"compare_url,omitempty"`
	Commits    []CommitPayload        `json:"commits,omitempty"`
	Repository RepositoryPayload      `json:"repository,omitempty"`
	Sender     UserPayload            `json:"sender,omitempty"`
	Issue      *IssuePayload          `json:"issue,omitempty"`
	PR         *PullRequestPayload    `json:"pull_request,omitempty"`
	Review     *ReviewPayload         `json:"review,omitempty"`
	Comment    *CommentPayload        `json:"comment,omitempty"`
}

// RepositoryPayload represents repository information in webhooks.
type RepositoryPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    UserPayload `json:"owner"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// UserPayload represents user information in webhooks.
type UserPayload struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// CommitPayload represents commit information in webhooks.
type CommitPayload struct {
	ID        string      `json:"id"`
	Message   string      `json:"message"`
	URL       string      `json:"url"`
	Author    AuthorPayload `json:"author"`
	Committer AuthorPayload `json:"committer"`
	Timestamp time.Time   `json:"timestamp"`
	Added     []string    `json:"added"`
	Removed   []string    `json:"removed"`
	Modified  []string    `json:"modified"`
}

// AuthorPayload represents author information in commits.
type AuthorPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// IssuePayload represents issue information in webhooks.
type IssuePayload struct {
	ID        int64       `json:"id"`
	Number    int64       `json:"number"`
	Title     string      `json:"title"`
	Body      string      `json:"body"`
	State     string      `json:"state"`
	HTMLURL   string      `json:"html_url"`
	User      UserPayload `json:"user"`
	Labels    []LabelPayload `json:"labels"`
	Milestone *MilestonePayload `json:"milestone"`
	Assignees []UserPayload `json:"assignees"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	ClosedAt  *time.Time  `json:"closed_at"`
}

// PullRequestPayload represents pull request information in webhooks.
type PullRequestPayload struct {
	ID             int64       `json:"id"`
	Number         int64       `json:"number"`
	Title          string      `json:"title"`
	Body           string      `json:"body"`
	State          string      `json:"state"`
	HTMLURL        string      `json:"html_url"`
	User           UserPayload `json:"user"`
	Head           BranchPayload `json:"head"`
	Base           BranchPayload `json:"base"`
	Labels         []LabelPayload `json:"labels"`
	Milestone      *MilestonePayload `json:"milestone"`
	Assignees      []UserPayload `json:"assignees"`
	RequestedReviewers []UserPayload `json:"requested_reviewers"`
	Merged         bool        `json:"merged"`
	MergedBy       *UserPayload `json:"merged_by"`
	MergeCommitSHA string      `json:"merge_commit_sha"`
	Mergeable      bool        `json:"mergeable"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	ClosedAt       *time.Time  `json:"closed_at"`
	MergedAt       *time.Time  `json:"merged_at"`
}

// BranchPayload represents branch information in pull requests.
type BranchPayload struct {
	Label string          `json:"label"`
	Ref   string          `json:"ref"`
	SHA   string          `json:"sha"`
	Repo  RepositoryPayload `json:"repo"`
}

// LabelPayload represents label information.
type LabelPayload struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// MilestonePayload represents milestone information.
type MilestonePayload struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	DueOn       *time.Time `json:"due_on"`
}

// ReviewPayload represents review information in webhooks.
type ReviewPayload struct {
	ID          int64       `json:"id"`
	Body        string      `json:"body"`
	State       string      `json:"state"`
	HTMLURL     string      `json:"html_url"`
	User        UserPayload `json:"user"`
	CommitID    string      `json:"commit_id"`
	SubmittedAt time.Time   `json:"submitted_at"`
}

// CommentPayload represents comment information in webhooks.
type CommentPayload struct {
	ID        int64       `json:"id"`
	Body      string      `json:"body"`
	HTMLURL   string      `json:"html_url"`
	User      UserPayload `json:"user"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ParsePayload parses a webhook payload into the appropriate struct.
func ParsePayload(payload []byte) (*GiteaWebhookPayload, error) {
	var p GiteaWebhookPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
