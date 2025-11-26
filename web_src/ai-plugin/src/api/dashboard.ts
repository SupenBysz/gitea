import { get } from './request'
import type {
  DashboardOverview,
  ComplianceTrend,
  ViolationStats,
  BlockedIssue,
  Event,
  PaginatedResponse,
} from '@/types'

// Dashboard Overview
export function getDashboardOverview(): Promise<DashboardOverview> {
  return get<DashboardOverview>('/dashboard/overview')
}

// Compliance Trend
export interface ComplianceTrendParams {
  start_date?: string
  end_date?: string
  granularity?: 'hour' | 'day' | 'week' | 'month'
}

export function getComplianceTrend(params?: ComplianceTrendParams): Promise<ComplianceTrend[]> {
  return get<ComplianceTrend[]>('/dashboard/compliance-trend', { params })
}

// Violation Statistics
export interface ViolationStatsParams {
  start_date?: string
  end_date?: string
  limit?: number
}

export function getViolationStats(params?: ViolationStatsParams): Promise<ViolationStats[]> {
  return get<ViolationStats[]>('/dashboard/violation-stats', { params })
}

// Blocked Issues
export interface BlockedIssuesParams {
  page?: number
  page_size?: number
  severity?: string
  repo?: string
}

export function getBlockedIssues(params?: BlockedIssuesParams): Promise<PaginatedResponse<BlockedIssue>> {
  return get<PaginatedResponse<BlockedIssue>>('/dashboard/blocked-issues', { params })
}

// Event Stream
export interface EventStreamParams {
  page?: number
  page_size?: number
  type?: string
  status?: string
  start_date?: string
  end_date?: string
}

export function getEventStream(params?: EventStreamParams): Promise<PaginatedResponse<Event>> {
  return get<PaginatedResponse<Event>>('/dashboard/event-stream', { params })
}

// WebSocket connection for real-time updates
export function createDashboardWebSocket(): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/v1/ai-plugin/dashboard/ws`
  return new WebSocket(wsUrl)
}
