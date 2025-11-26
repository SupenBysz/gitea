// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package config provides configuration management for the AI collaboration plugin.
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration for the AI plugin.
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Redis        RedisConfig        `yaml:"redis"`
	Gitea        GiteaConfig        `yaml:"gitea"`
	Webhook      WebhookConfig      `yaml:"webhook"`
	Notification NotificationConfig `yaml:"notification"`
	RuleEngine   RuleEngineConfig   `yaml:"rule_engine"`
	LLM          LLMConfig          `yaml:"llm"`
	Security     SecurityConfig     `yaml:"security"`
	Logging      LoggingConfig      `yaml:"logging"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"` // postgres
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	AutoMigrate     bool          `yaml:"auto_migrate"`
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// GiteaConfig contains Gitea integration settings.
type GiteaConfig struct {
	BaseURL      string `yaml:"base_url"`
	APIToken     string `yaml:"api_token"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// WebhookConfig contains webhook listener settings.
type WebhookConfig struct {
	Secret           string        `yaml:"secret"`
	ValidateSignature bool          `yaml:"validate_signature"`
	QueueSize        int           `yaml:"queue_size"`
	Workers          int           `yaml:"workers"`
	Timeout          time.Duration `yaml:"timeout"`
}

// NotificationConfig contains notification settings for various channels.
type NotificationConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Email    EmailConfig    `yaml:"email"`
	Slack    SlackConfig    `yaml:"slack"`
	Webhook  WebhookNotifyConfig `yaml:"webhook"`
}

// TelegramConfig contains Telegram bot settings.
type TelegramConfig struct {
	Enabled    bool   `yaml:"enabled"`
	BotToken   string `yaml:"bot_token"`
	DefaultChatID int64 `yaml:"default_chat_id"`
}

// EmailConfig contains email notification settings.
type EmailConfig struct {
	Enabled   bool   `yaml:"enabled"`
	SMTPHost  string `yaml:"smtp_host"`
	SMTPPort  int    `yaml:"smtp_port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	From      string `yaml:"from"`
	TLS       bool   `yaml:"tls"`
}

// SlackConfig contains Slack notification settings.
type SlackConfig struct {
	Enabled     bool   `yaml:"enabled"`
	WebhookURL  string `yaml:"webhook_url"`
	BotToken    string `yaml:"bot_token"`
	DefaultChannel string `yaml:"default_channel"`
}

// WebhookNotifyConfig contains generic webhook notification settings.
type WebhookNotifyConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	Secret  string `yaml:"secret"`
}

// RuleEngineConfig contains rule engine settings.
type RuleEngineConfig struct {
	RulesPath       string        `yaml:"rules_path"`
	ReloadInterval  time.Duration `yaml:"reload_interval"`
	EnabledByDefault bool         `yaml:"enabled_by_default"`
	MaxConditions   int           `yaml:"max_conditions"`
	MaxActions      int           `yaml:"max_actions"`
	EvalTimeout     time.Duration `yaml:"eval_timeout"`
}

// LLMConfig contains LLM gateway settings.
type LLMConfig struct {
	Enabled        bool          `yaml:"enabled"`
	DefaultProvider string       `yaml:"default_provider"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
}

// SecurityConfig contains security settings.
type SecurityConfig struct {
	JWTSecret        string        `yaml:"jwt_secret"`
	JWTExpiration    time.Duration `yaml:"jwt_expiration"`
	APIKeyHeader     string        `yaml:"api_key_header"`
	RateLimitEnabled bool          `yaml:"rate_limit_enabled"`
	RateLimitPerMin  int           `yaml:"rate_limit_per_min"`
	CORSOrigins      []string      `yaml:"cors_origins"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"` // debug, info, warn, error
	Format string `yaml:"format"` // json, text
	Output string `yaml:"output"` // stdout, file
	File   string `yaml:"file"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8081,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			User:            "ai_plugin",
			Password:        "",
			Database:        "ai_plugin",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			AutoMigrate:     true,
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			DB:       0,
			PoolSize: 10,
		},
		Webhook: WebhookConfig{
			ValidateSignature: true,
			QueueSize:        1000,
			Workers:          5,
			Timeout:          30 * time.Second,
		},
		Notification: NotificationConfig{
			Telegram: TelegramConfig{
				Enabled: false,
			},
		},
		RuleEngine: RuleEngineConfig{
			RulesPath:        "./rules",
			ReloadInterval:   5 * time.Minute,
			EnabledByDefault: true,
			MaxConditions:    20,
			MaxActions:       10,
			EvalTimeout:      10 * time.Second,
		},
		LLM: LLMConfig{
			Enabled:        false,
			RequestTimeout: 60 * time.Second,
			MaxRetries:     3,
			RetryDelay:     1 * time.Second,
		},
		Security: SecurityConfig{
			JWTExpiration:    24 * time.Hour,
			APIKeyHeader:     "X-API-Key",
			RateLimitEnabled: true,
			RateLimitPerMin:  100,
			CORSOrigins:      []string{"*"},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables.
func (c *Config) loadFromEnv() {
	if v := os.Getenv("AI_PLUGIN_DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("AI_PLUGIN_DB_PORT"); v != "" {
		// Parse port from string
	}
	if v := os.Getenv("AI_PLUGIN_DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("AI_PLUGIN_DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("AI_PLUGIN_DB_NAME"); v != "" {
		c.Database.Database = v
	}
	if v := os.Getenv("AI_PLUGIN_REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("AI_PLUGIN_GITEA_URL"); v != "" {
		c.Gitea.BaseURL = v
	}
	if v := os.Getenv("AI_PLUGIN_GITEA_TOKEN"); v != "" {
		c.Gitea.APIToken = v
	}
	if v := os.Getenv("AI_PLUGIN_WEBHOOK_SECRET"); v != "" {
		c.Webhook.Secret = v
	}
	if v := os.Getenv("AI_PLUGIN_TELEGRAM_TOKEN"); v != "" {
		c.Notification.Telegram.BotToken = v
		c.Notification.Telegram.Enabled = true
	}
	if v := os.Getenv("AI_PLUGIN_JWT_SECRET"); v != "" {
		c.Security.JWTSecret = v
	}
}

// DSN returns the database connection string.
func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + string(rune(c.Port)) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Database +
		" sslmode=" + c.SSLMode
}
