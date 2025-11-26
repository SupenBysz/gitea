// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package api

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/services"
)

// Webhook handler

// handleWebhook handles incoming Gitea webhook events.
func (r *Router) handleWebhook(c *gin.Context) {
	// Get event type from header
	eventType := c.GetHeader("X-Gitea-Event")
	if eventType == "" {
		eventType = c.GetHeader("X-GitHub-Event") // Fallback for compatibility
	}

	if eventType == "" {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "missing event type header")
		return
	}

	// Get signature for verification
	signature := c.GetHeader("X-Gitea-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Hub-Signature-256")
	}

	// Read payload
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "failed to read request body")
		return
	}

	// Process webhook
	if r.eventListener != nil {
		if err := r.eventListener.HandleWebhook(c.Request.Context(), eventType, payload, signature); err != nil {
			errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
			return
		}
	}

	successResponse(c, gin.H{"message": "webhook received"})
}

// Rules handlers

// listRules handles GET /rules
func (r *Router) listRules(c *gin.Context) {
	filter := &services.RuleFilter{
		Page:     getIntParam(c, "page", 1),
		PageSize: getIntParam(c, "page_size", 20),
		Search:   c.Query("search"),
		SortBy:   c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if enabled := c.Query("enabled"); enabled != "" {
		b := enabled == "true"
		filter.Enabled = &b
	}

	rules, total, err := r.ruleEngine.ListRules(c.Request.Context(), filter)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	successWithMeta(c, rules, &APIMeta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: totalPages,
	})
}

// createRule handles POST /rules
func (r *Router) createRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	if err := r.ruleEngine.CreateRule(c.Request.Context(), &rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	createdResponse(c, rule)
}

// getRule handles GET /rules/:id
func (r *Router) getRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "invalid rule ID")
		return
	}

	rule, err := r.ruleEngine.GetRule(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, ErrCodeNotFound, "rule not found")
		return
	}

	successResponse(c, rule)
}

// updateRule handles PUT /rules/:id
func (r *Router) updateRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "invalid rule ID")
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	rule.ID = id
	if err := r.ruleEngine.UpdateRule(c.Request.Context(), &rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	successResponse(c, rule)
}

// deleteRule handles DELETE /rules/:id
func (r *Router) deleteRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "invalid rule ID")
		return
	}

	if err := r.ruleEngine.DeleteRule(c.Request.Context(), id); err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "rule deleted"})
}

// testRule handles POST /rules/:id/test
func (r *Router) testRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeBadRequest, "invalid rule ID")
		return
	}

	var sampleEvent models.Event
	if err := c.ShouldBindJSON(&sampleEvent); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	rule, err := r.ruleEngine.GetRule(c.Request.Context(), id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, ErrCodeNotFound, "rule not found")
		return
	}

	result, err := r.ruleEngine.TestRule(c.Request.Context(), rule, &sampleEvent)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	successResponse(c, result)
}

// validateRule handles POST /rules/validate
func (r *Router) validateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	if err := r.ruleEngine.ValidateRule(c.Request.Context(), &rule); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	successResponse(c, gin.H{"valid": true})
}

// Events handlers

// listEvents handles GET /events
func (r *Router) listEvents(c *gin.Context) {
	// Implementation would query events from database
	successResponse(c, []interface{}{})
}

// getEvent handles GET /events/:id
func (r *Router) getEvent(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// getEventEvaluations handles GET /events/:id/evaluations
func (r *Router) getEventEvaluations(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// replayEvent handles POST /events/:id/replay
func (r *Router) replayEvent(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// Evaluations handlers

// listEvaluations handles GET /evaluations
func (r *Router) listEvaluations(c *gin.Context) {
	successResponse(c, []interface{}{})
}

// getEvaluation handles GET /evaluations/:id
func (r *Router) getEvaluation(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// Notifications handlers

// listNotifications handles GET /notifications
func (r *Router) listNotifications(c *gin.Context) {
	successResponse(c, []interface{}{})
}

// sendNotification handles POST /notifications
func (r *Router) sendNotification(c *gin.Context) {
	var req services.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, ErrCodeValidation, err.Error())
		return
	}

	if r.notification == nil {
		errorResponse(c, http.StatusServiceUnavailable, ErrCodeInternal, "notification service not available")
		return
	}

	resp, err := r.notification.Send(c.Request.Context(), &req)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	successResponse(c, resp)
}

// getNotification handles GET /notifications/:id
func (r *Router) getNotification(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// retryNotifications handles POST /notifications/retry
func (r *Router) retryNotifications(c *gin.Context) {
	if r.notification == nil {
		errorResponse(c, http.StatusServiceUnavailable, ErrCodeInternal, "notification service not available")
		return
	}

	count, err := r.notification.RetryFailed(c.Request.Context())
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	successResponse(c, gin.H{"retried": count})
}

// Templates handlers

func (r *Router) listTemplates(c *gin.Context) {
	successResponse(c, []interface{}{})
}

func (r *Router) createTemplate(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

func (r *Router) getTemplate(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

func (r *Router) updateTemplate(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

func (r *Router) deleteTemplate(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

func (r *Router) previewTemplate(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// Dashboard handlers

// getDashboardStats handles GET /dashboard/stats
func (r *Router) getDashboardStats(c *gin.Context) {
	stats := gin.H{
		"total_rules":        0,
		"active_rules":       0,
		"events_today":       0,
		"evaluations_today":  0,
		"violations_today":   0,
		"compliance_rate":    100.0,
	}
	successResponse(c, stats)
}

// getComplianceStats handles GET /dashboard/compliance
func (r *Router) getComplianceStats(c *gin.Context) {
	successResponse(c, gin.H{
		"overall_rate": 100.0,
		"by_rule":      []interface{}{},
		"by_repo":      []interface{}{},
		"trend":        []interface{}{},
	})
}

// getRecentEvents handles GET /dashboard/events/recent
func (r *Router) getRecentEvents(c *gin.Context) {
	successResponse(c, []interface{}{})
}

// getRecentViolations handles GET /dashboard/violations/recent
func (r *Router) getRecentViolations(c *gin.Context) {
	successResponse(c, []interface{}{})
}

// Audit handlers

func (r *Router) listAuditLogs(c *gin.Context) {
	successResponse(c, []interface{}{})
}

func (r *Router) getAuditLog(c *gin.Context) {
	errorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "not implemented")
}

// System handlers

// reloadRules handles POST /system/rules/reload
func (r *Router) reloadRules(c *gin.Context) {
	if err := r.ruleEngine.ReloadRules(c.Request.Context()); err != nil {
		errorResponse(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}
	successResponse(c, gin.H{"message": "rules reloaded"})
}

// getConfig handles GET /system/config
func (r *Router) getConfig(c *gin.Context) {
	// Return non-sensitive configuration
	successResponse(c, gin.H{
		"version": "0.1.0",
	})
}

// getSystemStats handles GET /system/stats
func (r *Router) getSystemStats(c *gin.Context) {
	var listenerStats *services.ListenerStats
	if r.eventListener != nil {
		listenerStats = r.eventListener.GetStats()
	}

	successResponse(c, gin.H{
		"listener": listenerStats,
	})
}

// Helper functions

func getIntParam(c *gin.Context, name string, defaultValue int) int {
	if v := c.Query(name); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getInt64Param(c *gin.Context, name string, defaultValue int64) int64 {
	if v := c.Query(name); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}
