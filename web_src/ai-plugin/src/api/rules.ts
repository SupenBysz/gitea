import { get, post, put, del, patch } from './request'
import type { Rule, PaginatedResponse } from '@/types'

// Rule List
export interface RuleListParams {
  page?: number
  page_size?: number
  enabled?: boolean
  event_type?: string
  keyword?: string
}

export function getRuleList(params?: RuleListParams): Promise<PaginatedResponse<Rule>> {
  return get<PaginatedResponse<Rule>>('/rules', { params })
}

// Get Single Rule
export function getRule(id: string): Promise<Rule> {
  return get<Rule>(`/rules/${id}`)
}

// Create Rule
export interface CreateRuleData {
  name: string
  description: string
  enabled: boolean
  priority: number
  event_types: string[]
  conditions: Rule['conditions']
  actions: Rule['actions']
}

export function createRule(data: CreateRuleData): Promise<Rule> {
  return post<Rule>('/rules', data)
}

// Update Rule
export function updateRule(id: string, data: Partial<CreateRuleData>): Promise<Rule> {
  return put<Rule>(`/rules/${id}`, data)
}

// Delete Rule
export function deleteRule(id: string): Promise<void> {
  return del<void>(`/rules/${id}`)
}

// Toggle Rule Enable/Disable
export function toggleRule(id: string, enabled: boolean): Promise<Rule> {
  return patch<Rule>(`/rules/${id}/toggle`, { enabled })
}

// Validate Rule
export interface ValidateRuleResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export function validateRule(data: CreateRuleData): Promise<ValidateRuleResult> {
  return post<ValidateRuleResult>('/rules/validate', data)
}

// Test Rule with Sample Event
export interface TestRuleResult {
  matched: boolean
  conditions_results: {
    condition: string
    passed: boolean
    actual_value: unknown
  }[]
  actions_to_execute: string[]
}

export function testRule(ruleId: string, eventPayload: Record<string, unknown>): Promise<TestRuleResult> {
  return post<TestRuleResult>(`/rules/${ruleId}/test`, { payload: eventPayload })
}
