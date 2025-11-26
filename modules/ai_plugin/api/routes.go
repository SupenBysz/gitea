// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package api provides HTTP API handlers for the AI collaboration plugin.
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code.gitea.io/gitea/modules/ai_plugin/services"
)

// Router holds all dependencies for API handlers.
type Router struct {
	ruleEngine   services.RuleEngine
	eventListener services.EventListener
	notification services.NotificationService
	audit        services.AuditService
}

// NewRouter creates a new API router with all handlers.
func NewRouter(
	ruleEngine services.RuleEngine,
	eventListener services.EventListener,
	notification services.NotificationService,
	audit services.AuditService,
) *Router {
	return &Router{
		ruleEngine:   ruleEngine,
		eventListener: eventListener,
		notification: notification,
		audit:        audit,
	}
}

// RegisterRoutes registers all API routes on the given Gin router group.
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	// Health check
	rg.GET("/health", r.healthCheck)

	// Webhook endpoint
	rg.POST("/webhook", r.handleWebhook)

	// Rules API
	rules := rg.Group("/rules")
	{
		rules.GET("", r.listRules)
		rules.POST("", r.createRule)
		rules.GET("/:id", r.getRule)
		rules.PUT("/:id", r.updateRule)
		rules.DELETE("/:id", r.deleteRule)
		rules.POST("/:id/test", r.testRule)
		rules.POST("/validate", r.validateRule)
	}

	// Events API
	events := rg.Group("/events")
	{
		events.GET("", r.listEvents)
		events.GET("/:id", r.getEvent)
		events.GET("/:id/evaluations", r.getEventEvaluations)
		events.POST("/:id/replay", r.replayEvent)
	}

	// Evaluations API
	evaluations := rg.Group("/evaluations")
	{
		evaluations.GET("", r.listEvaluations)
		evaluations.GET("/:id", r.getEvaluation)
	}

	// Notifications API
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", r.listNotifications)
		notifications.POST("", r.sendNotification)
		notifications.GET("/:id", r.getNotification)
		notifications.POST("/retry", r.retryNotifications)
	}

	// Templates API
	templates := rg.Group("/templates")
	{
		templates.GET("", r.listTemplates)
		templates.POST("", r.createTemplate)
		templates.GET("/:id", r.getTemplate)
		templates.PUT("/:id", r.updateTemplate)
		templates.DELETE("/:id", r.deleteTemplate)
		templates.POST("/:id/preview", r.previewTemplate)
	}

	// Dashboard API
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/stats", r.getDashboardStats)
		dashboard.GET("/compliance", r.getComplianceStats)
		dashboard.GET("/events/recent", r.getRecentEvents)
		dashboard.GET("/violations/recent", r.getRecentViolations)
	}

	// Audit API
	audit := rg.Group("/audit")
	{
		audit.GET("/logs", r.listAuditLogs)
		audit.GET("/logs/:id", r.getAuditLog)
	}

	// System API
	system := rg.Group("/system")
	{
		system.POST("/rules/reload", r.reloadRules)
		system.GET("/config", r.getConfig)
		system.GET("/stats", r.getSystemStats)
	}
}

// healthCheck handles GET /health
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   gin.H{},
	})
}

// Response helpers

// APIResponse represents a standard API response.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIError represents an API error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// APIMeta represents pagination and other metadata.
type APIMeta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func successWithMeta(c *gin.Context, data interface{}, meta *APIMeta) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func createdResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func errorResponse(c *gin.Context, status int, code string, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

func errorWithDetails(c *gin.Context, status int, code string, message string, details interface{}) {
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Common error codes
const (
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeDuplicate      = "DUPLICATE"
	ErrCodeRateLimited    = "RATE_LIMITED"
)
