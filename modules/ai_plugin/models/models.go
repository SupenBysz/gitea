// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package models provides database models for the AI collaboration plugin.
package models

import (
	"gorm.io/gorm"
)

// AllModels returns all model types for database migration.
func AllModels() []interface{} {
	return []interface{}{
		&Rule{},
		&RuleCondition{},
		&RuleAction{},
		&Event{},
		&Evaluation{},
		&Template{},
		&Notification{},
		&LLMProvider{},
		&LLMRequest{},
		&ScheduledTask{},
		&TaskExecution{},
		&User{},
		&UserRole{},
		&AuditLog{},
	}
}

// AutoMigrate runs auto-migration for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
