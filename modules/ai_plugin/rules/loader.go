// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.gitea.io/gitea/modules/ai_plugin/models"

	"gopkg.in/yaml.v3"
)

// RuleYAML represents a rule definition in YAML format.
type RuleYAML struct {
	Name               string          `yaml:"name"`
	Description        string          `yaml:"description"`
	Enabled            bool            `yaml:"enabled"`
	Priority           int             `yaml:"priority"`
	Trigger            TriggerYAML     `yaml:"trigger"`
	ConditionsOperator string          `yaml:"conditions_operator"`
	Conditions         []ConditionYAML `yaml:"conditions"`
	Actions            []ActionYAML    `yaml:"actions"`
}

// TriggerYAML represents a trigger definition in YAML format.
type TriggerYAML struct {
	EventTypes   []string `yaml:"event_types"`
	Actions      []string `yaml:"actions"`
	Repositories []string `yaml:"repositories,omitempty"`
	Branches     []string `yaml:"branches,omitempty"`
}

// ConditionYAML represents a condition definition in YAML format.
type ConditionYAML struct {
	Field    string                 `yaml:"field"`
	Operator string                 `yaml:"operator"`
	Value    interface{}            `yaml:"value"`
	Params   map[string]interface{} `yaml:"params,omitempty"`
}

// ActionYAML represents an action definition in YAML format.
type ActionYAML struct {
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`
}

// Loader handles loading rules from YAML files.
type Loader struct {
	rulesPath string
}

// NewLoader creates a new rule loader.
func NewLoader(rulesPath string) *Loader {
	return &Loader{rulesPath: rulesPath}
}

// LoadAll loads all rules from the rules directory.
func (l *Loader) LoadAll() ([]*models.Rule, error) {
	var rules []*models.Rule

	// Load from builtin directory
	builtinPath := filepath.Join(l.rulesPath, "builtin")
	if _, err := os.Stat(builtinPath); err == nil {
		builtinRules, err := l.loadFromDirectory(builtinPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load builtin rules: %w", err)
		}
		rules = append(rules, builtinRules...)
	}

	// Load from custom directory
	customPath := filepath.Join(l.rulesPath, "custom")
	if _, err := os.Stat(customPath); err == nil {
		customRules, err := l.loadFromDirectory(customPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom rules: %w", err)
		}
		rules = append(rules, customRules...)
	}

	// Also load from root rules directory
	rootRules, err := l.loadFromDirectory(l.rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load root rules: %w", err)
	}
	rules = append(rules, rootRules...)

	return rules, nil
}

// loadFromDirectory loads all YAML rules from a directory.
func (l *Loader) loadFromDirectory(dir string) ([]*models.Rule, error) {
	var rules []*models.Rule

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		filePath := filepath.Join(dir, name)
		rule, err := l.LoadFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load rule from %s: %w", filePath, err)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// LoadFromFile loads a rule from a YAML file.
func (l *Loader) LoadFromFile(filePath string) (*models.Rule, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return l.Parse(data)
}

// Parse parses a rule from YAML data.
func (l *Loader) Parse(data []byte) (*models.Rule, error) {
	var ruleYAML RuleYAML
	if err := yaml.Unmarshal(data, &ruleYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return l.convertToModel(&ruleYAML)
}

// convertToModel converts a YAML rule to a database model.
func (l *Loader) convertToModel(ruleYAML *RuleYAML) (*models.Rule, error) {
	// Convert trigger to JSON
	triggerJSON, err := json.Marshal(map[string]interface{}{
		"event_types":  ruleYAML.Trigger.EventTypes,
		"actions":      ruleYAML.Trigger.Actions,
		"repositories": ruleYAML.Trigger.Repositories,
		"branches":     ruleYAML.Trigger.Branches,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger: %w", err)
	}

	rule := &models.Rule{
		Name:               ruleYAML.Name,
		Description:        ruleYAML.Description,
		Enabled:            ruleYAML.Enabled,
		Priority:           ruleYAML.Priority,
		Trigger:            triggerJSON,
		ConditionsOperator: ruleYAML.ConditionsOperator,
	}

	// Convert conditions
	for i, condYAML := range ruleYAML.Conditions {
		// Merge field, operator, value into params for unified handling
		params := make(map[string]interface{})
		for k, v := range condYAML.Params {
			params[k] = v
		}
		params["field"] = condYAML.Field
		params["operator"] = condYAML.Operator
		params["value"] = condYAML.Value

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal condition params: %w", err)
		}

		rule.Conditions = append(rule.Conditions, models.RuleCondition{
			Type:       condYAML.Field,
			Params:     paramsJSON,
			OrderIndex: i,
		})
	}

	// Convert actions
	for i, actYAML := range ruleYAML.Actions {
		paramsJSON, err := json.Marshal(actYAML.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal action params: %w", err)
		}

		rule.Actions = append(rule.Actions, models.RuleAction{
			Type:       actYAML.Type,
			Params:     paramsJSON,
			OrderIndex: i,
		})
	}

	return rule, nil
}

// Validate validates a rule YAML structure.
func (l *Loader) Validate(data []byte) error {
	var ruleYAML RuleYAML
	if err := yaml.Unmarshal(data, &ruleYAML); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	if ruleYAML.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if len(ruleYAML.Trigger.EventTypes) == 0 {
		return fmt.Errorf("at least one event type is required in trigger")
	}

	if len(ruleYAML.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	if len(ruleYAML.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	// Validate conditions
	for i, cond := range ruleYAML.Conditions {
		if cond.Field == "" {
			return fmt.Errorf("condition %d: field is required", i)
		}
		if cond.Operator == "" {
			return fmt.Errorf("condition %d: operator is required", i)
		}
	}

	// Validate actions
	for i, act := range ruleYAML.Actions {
		if act.Type == "" {
			return fmt.Errorf("action %d: type is required", i)
		}
	}

	return nil
}
