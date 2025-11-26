// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package rules provides the rule engine for evaluating compliance rules.
package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"code.gitea.io/gitea/modules/ai_plugin/config"
	"code.gitea.io/gitea/modules/ai_plugin/models"
	"code.gitea.io/gitea/modules/ai_plugin/services"
)

// Engine implements the services.RuleEngine interface.
type Engine struct {
	config         *config.RuleEngineConfig
	ruleRepo       *services.RuleRepository
	evalRepo       *services.EvaluationRepository
	rules          []*models.Rule
	rulesByEvent   map[string][]*models.Rule
	conditions     map[string]ConditionEvaluator
	actions        map[string]ActionExecutor
	mu             sync.RWMutex
}

// ConditionEvaluator is a function that evaluates a single condition.
type ConditionEvaluator func(ctx context.Context, event *models.Event, params map[string]interface{}) (bool, error)

// ActionExecutor is a function that executes a single action.
type ActionExecutor func(ctx context.Context, event *models.Event, params map[string]interface{}) error

// NewEngine creates a new rule engine instance.
func NewEngine(cfg *config.RuleEngineConfig, ruleRepo *services.RuleRepository, evalRepo *services.EvaluationRepository) *Engine {
	e := &Engine{
		config:       cfg,
		ruleRepo:     ruleRepo,
		evalRepo:     evalRepo,
		rules:        make([]*models.Rule, 0),
		rulesByEvent: make(map[string][]*models.Rule),
		conditions:   make(map[string]ConditionEvaluator),
		actions:      make(map[string]ActionExecutor),
	}

	// Register built-in condition evaluators
	e.registerBuiltinConditions()

	// Register built-in action executors
	e.registerBuiltinActions()

	return e
}

// registerBuiltinConditions registers the built-in condition evaluators.
func (e *Engine) registerBuiltinConditions() {
	e.conditions[models.ConditionTypeCommentExists] = evaluateCommentExists
	e.conditions[models.ConditionTypeBranchNaming] = evaluateBranchNaming
	e.conditions[models.ConditionTypeFilePattern] = evaluateFilePattern
	e.conditions[models.ConditionTypeLabelCheck] = evaluateLabelCheck
	e.conditions[models.ConditionTypeUserInGroup] = evaluateUserInGroup
	e.conditions[models.ConditionTypePRTitle] = evaluatePRTitle
	e.conditions[models.ConditionTypePRDescription] = evaluatePRDescription
	e.conditions[models.ConditionTypeCommitMessage] = evaluateCommitMessage
	e.conditions[models.ConditionTypeFileCount] = evaluateFileCount
	e.conditions[models.ConditionTypeLineCount] = evaluateLineCount
	e.conditions[models.ConditionTypeReviewApproved] = evaluateReviewApproved
	e.conditions[models.ConditionTypeChecksPassed] = evaluateChecksPassed
	e.conditions[models.ConditionTypePRHasLinkedIssue] = evaluatePRHasLinkedIssue
	e.conditions[models.ConditionTypePRSubmitterIsAssignee] = evaluatePRSubmitterIsIssueAssignee
	e.conditions[models.ConditionTypeCommentContainsPattern] = evaluateCommentContainsPattern
	e.conditions[models.ConditionTypeIssueHasWorkSummary] = evaluateIssueHasWorkSummary
}

// registerBuiltinActions registers the built-in action executors.
func (e *Engine) registerBuiltinActions() {
	e.actions[models.ActionTypeBlock] = executeBlock
	e.actions[models.ActionTypeWarn] = executeWarn
	e.actions[models.ActionTypeRemind] = executeRemind
	e.actions[models.ActionTypeNotify] = executeNotify
	e.actions[models.ActionTypeAutoFix] = executeAutoFix
	e.actions[models.ActionTypeLabel] = executeLabel
	e.actions[models.ActionTypeAssign] = executeAssign
	e.actions[models.ActionTypeComment] = executeComment
}

// RegisterCondition registers a custom condition evaluator.
func (e *Engine) RegisterCondition(condType string, evaluator ConditionEvaluator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.conditions[condType] = evaluator
}

// RegisterAction registers a custom action executor.
func (e *Engine) RegisterAction(actionType string, executor ActionExecutor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.actions[actionType] = executor
}

// LoadRules loads rules from the database.
func (e *Engine) LoadRules(ctx context.Context) error {
	rules, err := e.ruleRepo.GetEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = rules
	e.rulesByEvent = make(map[string][]*models.Rule)

	for _, rule := range rules {
		eventTypes := e.extractEventTypes(rule)
		for _, eventType := range eventTypes {
			e.rulesByEvent[eventType] = append(e.rulesByEvent[eventType], rule)
		}
	}

	return nil
}

// ReloadRules reloads rules from the database.
func (e *Engine) ReloadRules(ctx context.Context) error {
	return e.LoadRules(ctx)
}

// extractEventTypes extracts event types from a rule's trigger configuration.
func (e *Engine) extractEventTypes(rule *models.Rule) []string {
	var trigger models.RuleTrigger
	if err := json.Unmarshal(rule.Trigger, &trigger); err != nil {
		return nil
	}
	return trigger.EventTypes
}

// Evaluate evaluates all applicable rules against an event.
func (e *Engine) Evaluate(ctx context.Context, event *models.Event) (*services.EvaluationResult, error) {
	startTime := time.Now()

	e.mu.RLock()
	applicableRules := e.getApplicableRules(event.EventType)
	e.mu.RUnlock()

	result := &services.EvaluationResult{
		EventID:        event.ID,
		RulesEvaluated: len(applicableRules),
		Passed:         true,
		Violations:     make([]models.Violation, 0),
		ActionsResults: make([]models.ActionResult, 0),
		Evaluations:    make([]*models.Evaluation, 0),
	}

	for _, rule := range applicableRules {
		evalResult, err := e.evaluateRule(ctx, rule, event)
		if err != nil {
			// Log error but continue with other rules
			continue
		}

		result.Evaluations = append(result.Evaluations, evalResult)

		if evalResult.Matched {
			result.RulesMatched++
		}

		if !evalResult.ConditionsMet {
			result.Passed = false
			// Extract violations from the evaluation
			var violations []models.Violation
			if err := json.Unmarshal(evalResult.Violations, &violations); err == nil {
				result.Violations = append(result.Violations, violations...)
			}
		}

		// Extract action results
		var actionResults []models.ActionResult
		if err := json.Unmarshal(evalResult.ActionsResults, &actionResults); err == nil {
			result.ActionsResults = append(result.ActionsResults, actionResults...)
		}
	}

	result.DurationMs = int(time.Since(startTime).Milliseconds())

	return result, nil
}

// getApplicableRules returns rules that apply to the given event type.
func (e *Engine) getApplicableRules(eventType string) []*models.Rule {
	rules := e.rulesByEvent[eventType]
	// Also include rules with wildcard event type
	rules = append(rules, e.rulesByEvent["*"]...)
	return rules
}

// evaluateRule evaluates a single rule against an event.
func (e *Engine) evaluateRule(ctx context.Context, rule *models.Rule, event *models.Event) (*models.Evaluation, error) {
	startTime := time.Now()

	// Check if rule trigger matches the event
	matched, err := e.matchesTrigger(ctx, rule, event)
	if err != nil {
		return nil, err
	}

	evaluation := &models.Evaluation{
		RuleID:    rule.ID,
		EventID:   event.ID,
		Matched:   matched,
		CreatedAt: time.Now(),
	}

	if !matched {
		evaluation.ConditionsMet = true // No conditions to evaluate if not matched
		evaluation.DurationMs = int(time.Since(startTime).Milliseconds())
		return evaluation, nil
	}

	// Evaluate conditions
	conditionsMet, violations := e.evaluateConditions(ctx, rule, event)
	evaluation.ConditionsMet = conditionsMet

	violationsJSON, _ := json.Marshal(violations)
	evaluation.Violations = violationsJSON

	// Execute actions based on condition results
	actionResults := e.executeActions(ctx, rule, event, conditionsMet)
	actionsJSON, _ := json.Marshal(actionResults)
	evaluation.ActionsResults = actionsJSON

	evaluation.DurationMs = int(time.Since(startTime).Milliseconds())

	// Store evaluation result
	if e.evalRepo != nil {
		if err := e.evalRepo.Create(ctx, evaluation); err != nil {
			// Log error but don't fail
		}
	}

	return evaluation, nil
}

// matchesTrigger checks if the rule's trigger matches the event.
func (e *Engine) matchesTrigger(ctx context.Context, rule *models.Rule, event *models.Event) (bool, error) {
	var trigger models.RuleTrigger
	if err := json.Unmarshal(rule.Trigger, &trigger); err != nil {
		return false, err
	}

	// Check event type
	eventTypeMatched := false
	for _, et := range trigger.EventTypes {
		if et == event.EventType || et == "*" {
			eventTypeMatched = true
			break
		}
		// Support event type with action, e.g., "pull_request.opened"
		if et == event.EventType+"."+event.Action {
			eventTypeMatched = true
			break
		}
	}

	if !eventTypeMatched {
		return false, nil
	}

	// Check filters
	for _, filter := range trigger.Filters {
		matched, err := e.evaluateFilter(ctx, event, &filter)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}

	return true, nil
}

// evaluateFilter evaluates a single trigger filter.
func (e *Engine) evaluateFilter(ctx context.Context, event *models.Event, filter *models.RuleTriggerFilter) (bool, error) {
	// Get the value from the event payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return false, err
	}

	value := getNestedValue(payload, filter.Field)

	// Compare based on operator
	switch filter.Operator {
	case "eq":
		return value == filter.Value, nil
	case "ne":
		return value != filter.Value, nil
	case "contains":
		if strVal, ok := value.(string); ok {
			if filterVal, ok := filter.Value.(string); ok {
				return containsString(strVal, filterVal), nil
			}
		}
		return false, nil
	case "in":
		if arr, ok := filter.Value.([]interface{}); ok {
			for _, v := range arr {
				if value == v {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown filter operator: %s", filter.Operator)
	}
}

// evaluateConditions evaluates all conditions for a rule.
func (e *Engine) evaluateConditions(ctx context.Context, rule *models.Rule, event *models.Event) (bool, []models.Violation) {
	violations := make([]models.Violation, 0)

	if len(rule.Conditions) == 0 {
		return true, violations
	}

	isAnd := rule.ConditionsOperator == models.OperatorAnd || rule.ConditionsOperator == ""

	allMet := true
	anyMet := false

	for _, condition := range rule.Conditions {
		evaluator, ok := e.conditions[condition.Type]
		if !ok {
			violations = append(violations, models.Violation{
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Unknown condition type: %s", condition.Type),
				Severity: models.SeverityError,
			})
			if isAnd {
				allMet = false
			}
			continue
		}

		var params map[string]interface{}
		json.Unmarshal(condition.Params, &params)

		met, err := evaluator(ctx, event, params)
		if err != nil {
			violations = append(violations, models.Violation{
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Condition evaluation error: %v", err),
				Severity: models.SeverityError,
			})
			if isAnd {
				allMet = false
			}
			continue
		}

		// Apply negation if configured
		if condition.Negate {
			met = !met
		}

		if met {
			anyMet = true
		} else {
			if isAnd {
				allMet = false
			}
			violations = append(violations, models.Violation{
				Rule:     rule.Name,
				Message:  fmt.Sprintf("Condition '%s' not met", condition.Type),
				Severity: models.SeverityWarning,
			})
		}
	}

	if isAnd {
		return allMet, violations
	}
	return anyMet, violations
}

// executeActions executes actions based on condition results.
func (e *Engine) executeActions(ctx context.Context, rule *models.Rule, event *models.Event, conditionsMet bool) []models.ActionResult {
	results := make([]models.ActionResult, 0)

	for _, action := range rule.Actions {
		// Check when to trigger
		shouldExecute := false
		switch action.WhenTrigger {
		case models.WhenConditionPassed:
			shouldExecute = conditionsMet
		case models.WhenConditionFailed:
			shouldExecute = !conditionsMet
		case models.WhenAlways:
			shouldExecute = true
		}

		if !shouldExecute {
			continue
		}

		executor, ok := e.actions[action.Type]
		if !ok {
			results = append(results, models.ActionResult{
				Type:    action.Type,
				Success: false,
				Error:   fmt.Sprintf("Unknown action type: %s", action.Type),
			})
			continue
		}

		var params map[string]interface{}
		json.Unmarshal(action.Params, &params)

		err := executor(ctx, event, params)
		result := models.ActionResult{
			Type:    action.Type,
			Success: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	return results
}

// GetRule retrieves a rule by ID.
func (e *Engine) GetRule(ctx context.Context, id int64) (*models.Rule, error) {
	return e.ruleRepo.Get(ctx, id)
}

// GetRuleByName retrieves a rule by name.
func (e *Engine) GetRuleByName(ctx context.Context, name string) (*models.Rule, error) {
	return e.ruleRepo.GetByName(ctx, name)
}

// ListRules lists rules with filtering.
func (e *Engine) ListRules(ctx context.Context, filter *services.RuleFilter) ([]*models.Rule, int64, error) {
	return e.ruleRepo.List(ctx, filter)
}

// CreateRule creates a new rule.
func (e *Engine) CreateRule(ctx context.Context, rule *models.Rule) error {
	if err := e.ValidateRule(ctx, rule); err != nil {
		return err
	}
	return e.ruleRepo.Create(ctx, rule)
}

// UpdateRule updates an existing rule.
func (e *Engine) UpdateRule(ctx context.Context, rule *models.Rule) error {
	if err := e.ValidateRule(ctx, rule); err != nil {
		return err
	}
	return e.ruleRepo.Update(ctx, rule)
}

// DeleteRule deletes a rule.
func (e *Engine) DeleteRule(ctx context.Context, id int64) error {
	return e.ruleRepo.Delete(ctx, id)
}

// ValidateRule validates a rule configuration.
func (e *Engine) ValidateRule(ctx context.Context, rule *models.Rule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	var trigger models.RuleTrigger
	if err := json.Unmarshal(rule.Trigger, &trigger); err != nil {
		return fmt.Errorf("invalid trigger configuration: %w", err)
	}

	if len(trigger.EventTypes) == 0 {
		return fmt.Errorf("at least one event type is required")
	}

	// Validate conditions
	for _, cond := range rule.Conditions {
		if _, ok := e.conditions[cond.Type]; !ok {
			return fmt.Errorf("unknown condition type: %s", cond.Type)
		}
	}

	// Validate actions
	for _, action := range rule.Actions {
		if _, ok := e.actions[action.Type]; !ok {
			return fmt.Errorf("unknown action type: %s", action.Type)
		}
	}

	return nil
}

// TestRule tests a rule against sample data.
func (e *Engine) TestRule(ctx context.Context, rule *models.Rule, sampleEvent *models.Event) (*services.EvaluationResult, error) {
	// Temporarily add the rule for testing
	e.mu.Lock()
	originalRules := e.rules
	e.rules = []*models.Rule{rule}
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.rules = originalRules
		e.mu.Unlock()
	}()

	return e.Evaluate(ctx, sampleEvent)
}

// Helper functions

func getNestedValue(data map[string]interface{}, path string) interface{} {
	// Simple implementation for now - could be enhanced with JSONPath
	return data[path]
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s[1:], substr))
}
