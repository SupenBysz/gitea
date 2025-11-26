// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package ai_plugin provides the AI collaboration plugin for Gitea.
// It implements compliance checking, rule-based automation, and notification
// services for AI employee collaboration workflows.
package ai_plugin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"code.gitea.io/gitea/modules/ai_plugin/api"
	"code.gitea.io/gitea/modules/ai_plugin/config"
	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/notification"
	"code.gitea.io/gitea/modules/ai_plugin/rules"
	"code.gitea.io/gitea/modules/ai_plugin/services"
	"code.gitea.io/gitea/modules/ai_plugin/webhook"
)

// Plugin represents the AI collaboration plugin instance.
type Plugin struct {
	config        *config.Config
	database      *services.Database
	ruleEngine    *rules.Engine
	eventListener *webhook.EventListener
	notification  *notification.Service
	router        *api.Router
	server        *http.Server
}

// New creates a new Plugin instance with the given configuration.
func New(cfg *config.Config) (*Plugin, error) {
	p := &Plugin{
		config: cfg,
	}

	// Initialize database
	db, err := services.NewDatabase(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	p.database = db

	// Initialize repositories
	ruleRepo := services.NewRuleRepository(db)
	eventRepo := services.NewEventRepository(db)
	evalRepo := services.NewEvaluationRepository(db)
	notifRepo := services.NewNotificationRepository(db)

	// Initialize rule engine
	p.ruleEngine = rules.NewEngine(&cfg.RuleEngine, ruleRepo, evalRepo)

	// Initialize event listener
	p.eventListener = webhook.NewEventListener(&cfg.Webhook, eventRepo)

	// Initialize notification service
	p.notification = notification.NewService(&cfg.Notification, notifRepo)

	// Initialize API router
	p.router = api.NewRouter(p.ruleEngine, p.eventListener, p.notification, nil)

	return p, nil
}

// NewWithDefaults creates a new Plugin instance with default configuration.
func NewWithDefaults() (*Plugin, error) {
	return New(config.DefaultConfig())
}

// Start starts the plugin services.
func (p *Plugin) Start(ctx context.Context) error {
	// Load rules from database
	if err := p.ruleEngine.LoadRules(ctx); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	// Register the rule engine as an event handler
	p.eventListener.RegisterHandler("*", func(ctx context.Context, event *models.Event) error {
		// Evaluate rules for all events
		_, err := p.ruleEngine.Evaluate(ctx, event)
		return err
	})

	// Start event listener
	if err := p.eventListener.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event listener: %w", err)
	}

	// Create and start HTTP server
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())

	// Register API routes
	apiGroup := engine.Group("/api/v1/ai-plugin")
	p.router.RegisterRoutes(apiGroup)

	addr := fmt.Sprintf("%s:%d", p.config.Server.Host, p.config.Server.Port)
	p.server = &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  p.config.Server.ReadTimeout,
		WriteTimeout: p.config.Server.WriteTimeout,
		IdleTimeout:  p.config.Server.IdleTimeout,
	}

	// Start server in background
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error
		}
	}()

	return nil
}

// Stop gracefully stops the plugin services.
func (p *Plugin) Stop(ctx context.Context) error {
	// Stop HTTP server
	if p.server != nil {
		if err := p.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	// Stop event listener
	if err := p.eventListener.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop event listener: %w", err)
	}

	// Close database connection
	if err := p.database.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// RuleEngine returns the rule engine instance.
func (p *Plugin) RuleEngine() *rules.Engine {
	return p.ruleEngine
}

// EventListener returns the event listener instance.
func (p *Plugin) EventListener() *webhook.EventListener {
	return p.eventListener
}

// NotificationService returns the notification service instance.
func (p *Plugin) NotificationService() *notification.Service {
	return p.notification
}

// Database returns the database instance.
func (p *Plugin) Database() *services.Database {
	return p.database
}

// corsMiddleware returns a CORS middleware handler.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
