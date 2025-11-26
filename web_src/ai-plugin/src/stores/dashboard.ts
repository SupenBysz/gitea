import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  DashboardOverview,
  ComplianceTrend,
  ViolationStats,
  BlockedIssue,
  Event,
  WSMessage,
} from '@/types'
import {
  getDashboardOverview,
  getComplianceTrend,
  getViolationStats,
  getBlockedIssues,
  getEventStream,
  createDashboardWebSocket,
} from '@/api'

export const useDashboardStore = defineStore('dashboard', () => {
  // State
  const overview = ref<DashboardOverview | null>(null)
  const complianceTrend = ref<ComplianceTrend[]>([])
  const violationStats = ref<ViolationStats[]>([])
  const blockedIssues = ref<BlockedIssue[]>([])
  const eventStream = ref<Event[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const wsConnection = ref<WebSocket | null>(null)
  const wsConnected = ref(false)

  // Computed
  const complianceRate = computed(() => overview.value?.compliance_rate ?? 0)
  const totalEvents = computed(() => overview.value?.total_events ?? 0)
  const blockedCount = computed(() => overview.value?.blocked_count ?? 0)
  const activeRules = computed(() => overview.value?.active_rules ?? 0)

  // Actions
  async function fetchOverview() {
    try {
      overview.value = await getDashboardOverview()
    } catch (e) {
      error.value = '获取概览数据失败'
      throw e
    }
  }

  async function fetchComplianceTrend(params?: { start_date?: string; end_date?: string; granularity?: 'hour' | 'day' | 'week' | 'month' }) {
    try {
      complianceTrend.value = await getComplianceTrend(params)
    } catch (e) {
      error.value = '获取合规趋势失败'
      throw e
    }
  }

  async function fetchViolationStats(params?: { start_date?: string; end_date?: string; limit?: number }) {
    try {
      violationStats.value = await getViolationStats(params)
    } catch (e) {
      error.value = '获取违规统计失败'
      throw e
    }
  }

  async function fetchBlockedIssues(params?: { page?: number; page_size?: number }) {
    try {
      const response = await getBlockedIssues(params)
      blockedIssues.value = response.items
    } catch (e) {
      error.value = '获取阻塞工单失败'
      throw e
    }
  }

  async function fetchEventStream(params?: { page?: number; page_size?: number }) {
    try {
      const response = await getEventStream(params)
      eventStream.value = response.items
    } catch (e) {
      error.value = '获取事件流失败'
      throw e
    }
  }

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      await Promise.all([
        fetchOverview(),
        fetchComplianceTrend(),
        fetchViolationStats(),
        fetchBlockedIssues(),
        fetchEventStream({ page_size: 20 }),
      ])
    } finally {
      loading.value = false
    }
  }

  // WebSocket
  function connectWebSocket() {
    if (wsConnection.value) {
      wsConnection.value.close()
    }

    const ws = createDashboardWebSocket()

    ws.onopen = () => {
      wsConnected.value = true
    }

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WSMessage
        handleWSMessage(message)
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    ws.onclose = () => {
      wsConnected.value = false
      // Attempt reconnection after 5 seconds
      setTimeout(() => {
        if (!wsConnected.value) {
          connectWebSocket()
        }
      }, 5000)
    }

    ws.onerror = (err) => {
      console.error('WebSocket error:', err)
    }

    wsConnection.value = ws
  }

  function handleWSMessage(message: WSMessage) {
    switch (message.type) {
      case 'event':
        // Prepend new event to stream
        eventStream.value = [message.data as Event, ...eventStream.value.slice(0, 19)]
        // Update overview event count
        if (overview.value) {
          overview.value.total_events++
          overview.value.events_today++
        }
        break
      case 'evaluation':
        // Update compliance rate if evaluation failed
        if (overview.value) {
          fetchOverview() // Refresh overview to get updated stats
        }
        break
      case 'alert':
        // Add alert to overview
        if (overview.value) {
          overview.value.recent_alerts = [
            message.data as typeof overview.value.recent_alerts[0],
            ...overview.value.recent_alerts.slice(0, 4),
          ]
        }
        break
      case 'status':
        // Update AI employee status
        if (overview.value) {
          const status = message.data as typeof overview.value.ai_employees[0]
          const index = overview.value.ai_employees.findIndex((e) => e.id === status.id)
          if (index !== -1) {
            overview.value.ai_employees[index] = status
          }
        }
        break
    }
  }

  function disconnectWebSocket() {
    if (wsConnection.value) {
      wsConnection.value.close()
      wsConnection.value = null
      wsConnected.value = false
    }
  }

  return {
    // State
    overview,
    complianceTrend,
    violationStats,
    blockedIssues,
    eventStream,
    loading,
    error,
    wsConnected,
    // Computed
    complianceRate,
    totalEvents,
    blockedCount,
    activeRules,
    // Actions
    fetchOverview,
    fetchComplianceTrend,
    fetchViolationStats,
    fetchBlockedIssues,
    fetchEventStream,
    fetchAll,
    connectWebSocket,
    disconnectWebSocket,
  }
})
