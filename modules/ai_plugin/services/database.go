// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"context"
	"fmt"
	"time"

	"code.gitea.io/gitea/modules/ai_plugin/config"
	"code.gitea.io/gitea/modules/ai_plugin/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection with additional functionality.
type Database struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// NewDatabase creates a new database connection.
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	logLevel := logger.Silent
	if cfg.SSLMode == "debug" {
		logLevel = logger.Info
	}

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	database := &Database{
		db:     db,
		config: cfg,
	}

	if cfg.AutoMigrate {
		if err := database.Migrate(); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return database, nil
}

// DB returns the underlying GORM database instance.
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Migrate runs database migrations for all models.
func (d *Database) Migrate() error {
	return models.AutoMigrate(d.db)
}

// Ping checks the database connection.
func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithContext returns a new Database instance with the given context.
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

// Transaction executes a function within a database transaction.
func (d *Database) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(fn)
}

// RuleRepository provides database operations for rules.
type RuleRepository struct {
	db *Database
}

// NewRuleRepository creates a new RuleRepository.
func NewRuleRepository(db *Database) *RuleRepository {
	return &RuleRepository{db: db}
}

// Create creates a new rule.
func (r *RuleRepository) Create(ctx context.Context, rule *models.Rule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// Get retrieves a rule by ID.
func (r *RuleRepository) Get(ctx context.Context, id int64) (*models.Rule, error) {
	var rule models.Rule
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Actions").
		First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetByName retrieves a rule by name.
func (r *RuleRepository) GetByName(ctx context.Context, name string) (*models.Rule, error) {
	var rule models.Rule
	err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Actions").
		Where("name = ?", name).
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// List lists rules with optional filtering.
func (r *RuleRepository) List(ctx context.Context, filter *RuleFilter) ([]*models.Rule, int64, error) {
	var rules []*models.Rule
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Rule{})

	if filter != nil {
		if filter.Enabled != nil {
			query = query.Where("enabled = ?", *filter.Enabled)
		}
		if filter.Search != "" {
			query = query.Where("name ILIKE ? OR description ILIKE ?",
				"%"+filter.Search+"%", "%"+filter.Search+"%")
		}
		if len(filter.EventTypes) > 0 {
			query = query.Where("trigger->>'event_types' ?| ?", filter.EventTypes)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter != nil {
		if filter.SortBy != "" {
			order := filter.SortBy
			if filter.SortOrder == "desc" {
				order += " DESC"
			}
			query = query.Order(order)
		} else {
			query = query.Order("priority ASC, created_at DESC")
		}

		if filter.PageSize > 0 {
			query = query.Limit(filter.PageSize)
			if filter.Page > 0 {
				query = query.Offset((filter.Page - 1) * filter.PageSize)
			}
		}
	}

	err := query.
		Preload("Conditions").
		Preload("Actions").
		Find(&rules).Error

	return rules, total, err
}

// Update updates a rule.
func (r *RuleRepository) Update(ctx context.Context, rule *models.Rule) error {
	rule.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(rule).Error
}

// Delete deletes a rule by ID.
func (r *RuleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Rule{}, id).Error
}

// GetEnabled retrieves all enabled rules.
func (r *RuleRepository) GetEnabled(ctx context.Context) ([]*models.Rule, error) {
	var rules []*models.Rule
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("priority ASC").
		Preload("Conditions").
		Preload("Actions").
		Find(&rules).Error
	return rules, err
}

// EventRepository provides database operations for events.
type EventRepository struct {
	db *Database
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *Database) *EventRepository {
	return &EventRepository{db: db}
}

// Create creates a new event.
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// Get retrieves an event by ID.
func (r *EventRepository) Get(ctx context.Context, id int64) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetByEventID retrieves an event by Gitea event ID.
func (r *EventRepository) GetByEventID(ctx context.Context, eventID string) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// MarkProcessed marks an event as processed.
func (r *EventRepository) MarkProcessed(ctx context.Context, id int64, err error) error {
	now := time.Now()
	updates := map[string]interface{}{
		"processed":    true,
		"processed_at": &now,
	}
	if err != nil {
		updates["process_error"] = err.Error()
	}
	return r.db.WithContext(ctx).Model(&models.Event{}).Where("id = ?", id).Updates(updates).Error
}

// EvaluationRepository provides database operations for evaluations.
type EvaluationRepository struct {
	db *Database
}

// NewEvaluationRepository creates a new EvaluationRepository.
func NewEvaluationRepository(db *Database) *EvaluationRepository {
	return &EvaluationRepository{db: db}
}

// Create creates a new evaluation.
func (r *EvaluationRepository) Create(ctx context.Context, eval *models.Evaluation) error {
	return r.db.WithContext(ctx).Create(eval).Error
}

// CreateBatch creates multiple evaluations.
func (r *EvaluationRepository) CreateBatch(ctx context.Context, evals []*models.Evaluation) error {
	return r.db.WithContext(ctx).Create(&evals).Error
}

// GetByEventID retrieves evaluations for an event.
func (r *EvaluationRepository) GetByEventID(ctx context.Context, eventID int64) ([]*models.Evaluation, error) {
	var evals []*models.Evaluation
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Preload("Rule").
		Find(&evals).Error
	return evals, err
}

// NotificationRepository provides database operations for notifications.
type NotificationRepository struct {
	db *Database
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *Database) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification.
func (r *NotificationRepository) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

// Update updates a notification.
func (r *NotificationRepository) Update(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Save(notif).Error
}

// GetPending retrieves pending notifications.
func (r *NotificationRepository) GetPending(ctx context.Context, limit int) ([]*models.Notification, error) {
	var notifs []*models.Notification
	err := r.db.WithContext(ctx).
		Where("status = ?", models.NotificationStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifs).Error
	return notifs, err
}

// GetFailed retrieves failed notifications that can be retried.
func (r *NotificationRepository) GetFailed(ctx context.Context, limit int) ([]*models.Notification, error) {
	var notifs []*models.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < ?", models.NotificationStatusFailed, models.MaxRetryCount).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifs).Error
	return notifs, err
}
