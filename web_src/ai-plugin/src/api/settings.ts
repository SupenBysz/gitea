import { get, put } from './request'
import type { SystemSettings, NotificationChannelConfig } from '@/types'

// Get System Settings
export function getSystemSettings(): Promise<SystemSettings> {
  return get<SystemSettings>('/settings')
}

// Update System Settings
export function updateSystemSettings(data: Partial<SystemSettings>): Promise<SystemSettings> {
  return put<SystemSettings>('/settings', data)
}

// Test Notification Channel
export interface TestNotificationResult {
  success: boolean
  message: string
}

export function testNotificationChannel(
  channel: string,
  config: Record<string, string>
): Promise<TestNotificationResult> {
  return get<TestNotificationResult>('/settings/notifications/test', {
    params: { channel, ...config },
  })
}

// Get Notification Channel Config
export function getNotificationChannelConfig(channel: string): Promise<NotificationChannelConfig> {
  return get<NotificationChannelConfig>(`/settings/notifications/${channel}`)
}

// Update Notification Channel Config
export function updateNotificationChannelConfig(
  channel: string,
  config: Partial<NotificationChannelConfig>
): Promise<NotificationChannelConfig> {
  return put<NotificationChannelConfig>(`/settings/notifications/${channel}`, config)
}

// Regenerate Webhook Secret
export function regenerateWebhookSecret(): Promise<{ secret: string }> {
  return get<{ secret: string }>('/settings/webhook-secret/regenerate')
}
