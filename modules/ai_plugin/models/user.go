// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"time"

	"gorm.io/datatypes"
)

// User represents a plugin user synchronized from Gitea.
type User struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	GiteaUserID int64          `gorm:"uniqueIndex;not null" json:"gitea_user_id"`
	Username    string         `gorm:"size:255;uniqueIndex;not null" json:"username"`
	Email       string         `gorm:"size:255" json:"email"`
	AvatarURL   string         `gorm:"size:500" json:"avatar_url"`
	Settings    datatypes.JSON `gorm:"type:jsonb;default:{}" json:"settings"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	LastLoginAt *time.Time     `json:"last_login_at"`

	// Relations
	Roles []UserRole `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"roles"`
}

// TableName returns the table name for User model.
func (User) TableName() string {
	return "ai_plugin_user"
}

// UserSettings represents user-specific settings.
type UserSettings struct {
	NotifyChannels []string `json:"notify_channels,omitempty"`
	Language       string   `json:"language,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	TelegramChatID int64    `json:"telegram_chat_id,omitempty"`
	SlackUserID    string   `json:"slack_user_id,omitempty"`
}

// UserRole represents a role assignment for a user.
type UserRole struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"index;not null" json:"user_id"`
	Role       string    `gorm:"size:50;index;not null" json:"role"`
	Scope      string    `gorm:"size:50;default:global;not null" json:"scope"`
	ResourceID string    `gorm:"size:255" json:"resource_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for UserRole model.
func (UserRole) TableName() string {
	return "ai_plugin_user_role"
}

// User roles.
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleOperator   = "operator"
	RoleViewer     = "viewer"
	RoleAPIClient  = "api_client"
)

// Role scopes.
const (
	ScopeGlobal       = "global"
	ScopeOrganization = "organization"
	ScopeRepository   = "repository"
)

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r.Role == role {
			return true
		}
	}
	return false
}

// HasRoleInScope checks if the user has a specific role within a given scope.
func (u *User) HasRoleInScope(role, scope, resourceID string) bool {
	for _, r := range u.Roles {
		if r.Role == role && r.Scope == scope && r.ResourceID == resourceID {
			return true
		}
	}
	return false
}

// IsSuperAdmin checks if the user is a super admin.
func (u *User) IsSuperAdmin() bool {
	return u.HasRole(RoleSuperAdmin)
}

// IsAdmin checks if the user is an admin or super admin.
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleSuperAdmin) || u.HasRole(RoleAdmin)
}
