import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { SystemSettings, NotificationChannelConfig } from '@/types'
import {
  getSystemSettings,
  updateSystemSettings,
  testNotificationChannel,
  regenerateWebhookSecret,
} from '@/api'

export const useSettingsStore = defineStore('settings', () => {
  // State
  const settings = ref<SystemSettings | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchSettings() {
    loading.value = true
    error.value = null
    try {
      settings.value = await getSystemSettings()
    } catch (e) {
      error.value = '获取系统设置失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function saveSettings(data: Partial<SystemSettings>) {
    saving.value = true
    error.value = null
    try {
      settings.value = await updateSystemSettings(data)
    } catch (e) {
      error.value = '保存系统设置失败'
      throw e
    } finally {
      saving.value = false
    }
  }

  async function testChannel(channel: string, config: Record<string, string>) {
    error.value = null
    try {
      return await testNotificationChannel(channel, config)
    } catch (e) {
      error.value = '测试通知渠道失败'
      throw e
    }
  }

  async function regenerateSecret() {
    error.value = null
    try {
      const result = await regenerateWebhookSecret()
      if (settings.value) {
        settings.value.webhook_secret = result.secret
      }
      return result.secret
    } catch (e) {
      error.value = '重新生成密钥失败'
      throw e
    }
  }

  function updateChannelConfig(channel: string, config: Partial<NotificationChannelConfig>) {
    if (!settings.value) return
    const index = settings.value.notification_channels.findIndex((c) => c.channel === channel)
    if (index !== -1) {
      settings.value.notification_channels[index] = {
        ...settings.value.notification_channels[index],
        ...config,
      }
    }
  }

  return {
    // State
    settings,
    loading,
    saving,
    error,
    // Actions
    fetchSettings,
    saveSettings,
    testChannel,
    regenerateSecret,
    updateChannelConfig,
  }
})
