<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useSettingsStore } from '@/stores'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { NotificationChannel } from '@/types'

const settingsStore = useSettingsStore()
const formRef = ref<FormInstance>()
const activeTab = ref('general')

const formData = reactive({
  webhook_secret: '',
  default_llm_provider: '',
  audit_retention_days: 90,
  metrics_retention_days: 365,
  max_concurrent_evaluations: 10,
})

const rules: FormRules = {
  audit_retention_days: [
    { required: true, message: '请输入审计日志保留天数', trigger: 'blur' },
    { type: 'number', min: 7, max: 3650, message: '保留天数需在 7-3650 之间', trigger: 'blur' },
  ],
  metrics_retention_days: [
    { required: true, message: '请输入指标数据保留天数', trigger: 'blur' },
    { type: 'number', min: 30, max: 3650, message: '保留天数需在 30-3650 之间', trigger: 'blur' },
  ],
  max_concurrent_evaluations: [
    { required: true, message: '请输入最大并发评估数', trigger: 'blur' },
    { type: 'number', min: 1, max: 100, message: '并发数需在 1-100 之间', trigger: 'blur' },
  ],
}

const llmProviders = [
  { label: 'Claude (Anthropic)', value: 'claude' },
  { label: 'GPT-4 (OpenAI)', value: 'gpt4' },
  { label: 'ChatGPT (OpenAI)', value: 'chatgpt' },
]

const channelLabels: Record<NotificationChannel, string> = {
  telegram: 'Telegram',
  gitea_comment: 'Gitea 评论',
  webhook: 'Webhook',
  email: '邮件',
  slack: 'Slack',
}

function syncFormData() {
  if (settingsStore.settings) {
    formData.webhook_secret = settingsStore.settings.webhook_secret
    formData.default_llm_provider = settingsStore.settings.default_llm_provider
    formData.audit_retention_days = settingsStore.settings.audit_retention_days
    formData.metrics_retention_days = settingsStore.settings.metrics_retention_days
    formData.max_concurrent_evaluations = settingsStore.settings.max_concurrent_evaluations
  }
}

async function handleSave() {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    await settingsStore.saveSettings({
      default_llm_provider: formData.default_llm_provider,
      audit_retention_days: formData.audit_retention_days,
      metrics_retention_days: formData.metrics_retention_days,
      max_concurrent_evaluations: formData.max_concurrent_evaluations,
    })
    ElMessage.success('设置已保存')
  } catch {
    // Validation failed or save error handled in store
  }
}

async function handleRegenerateSecret() {
  try {
    await ElMessageBox.confirm(
      '重新生成 Webhook 密钥后，需要更新所有已配置的 Webhook。确定要继续吗？',
      '重新生成密钥',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    const newSecret = await settingsStore.regenerateSecret()
    formData.webhook_secret = newSecret
    ElMessage.success('Webhook 密钥已重新生成')
  } catch (e) {
    if (e !== 'cancel') {
      // Error handled in store
    }
  }
}

async function handleTestChannel(channel: NotificationChannel) {
  const config = settingsStore.settings?.notification_channels.find((c) => c.channel === channel)?.config
  if (!config) {
    ElMessage.warning('请先配置该渠道')
    return
  }
  try {
    const result = await settingsStore.testChannel(channel, config)
    if (result.success) {
      ElMessage.success('测试消息已发送')
    } else {
      ElMessage.error(result.message)
    }
  } catch {
    // Error handled in store
  }
}

function handleToggleChannel(channel: NotificationChannel, enabled: boolean) {
  settingsStore.updateChannelConfig(channel, { enabled })
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

onMounted(async () => {
  await settingsStore.fetchSettings()
  syncFormData()
})
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <h1 class="page-header__title">系统设置</h1>
      <p class="page-header__desc">配置 AI 协同工作规范插件的系统参数</p>
    </div>

    <div class="dashboard-card">
      <el-tabs v-model="activeTab">
        <!-- General Settings -->
        <el-tab-pane label="基本设置" name="general">
          <el-form
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-width="180px"
            style="max-width: 600px"
          >
            <el-form-item label="Webhook 密钥">
              <el-input v-model="formData.webhook_secret" readonly>
                <template #append>
                  <el-button @click="copyToClipboard(formData.webhook_secret)">
                    复制
                  </el-button>
                </template>
              </el-input>
              <div class="form-tip">
                用于验证 Gitea Webhook 请求的签名
                <el-button type="primary" link @click="handleRegenerateSecret">
                  重新生成
                </el-button>
              </div>
            </el-form-item>

            <el-form-item label="默认 LLM 提供商">
              <el-select v-model="formData.default_llm_provider" style="width: 100%">
                <el-option
                  v-for="provider in llmProviders"
                  :key="provider.value"
                  :label="provider.label"
                  :value="provider.value"
                />
              </el-select>
              <div class="form-tip">用于智能分析的默认 LLM 模型</div>
            </el-form-item>

            <el-form-item label="审计日志保留天数" prop="audit_retention_days">
              <el-input-number
                v-model="formData.audit_retention_days"
                :min="7"
                :max="3650"
                style="width: 100%"
              />
              <div class="form-tip">超过此天数的审计日志将被自动清理</div>
            </el-form-item>

            <el-form-item label="指标数据保留天数" prop="metrics_retention_days">
              <el-input-number
                v-model="formData.metrics_retention_days"
                :min="30"
                :max="3650"
                style="width: 100%"
              />
              <div class="form-tip">时序指标数据的保留时间</div>
            </el-form-item>

            <el-form-item label="最大并发评估数" prop="max_concurrent_evaluations">
              <el-input-number
                v-model="formData.max_concurrent_evaluations"
                :min="1"
                :max="100"
                style="width: 100%"
              />
              <div class="form-tip">同时进行规则评估的最大数量</div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :loading="settingsStore.saving" @click="handleSave">
                保存设置
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Notification Settings -->
        <el-tab-pane label="通知设置" name="notification">
          <div v-if="settingsStore.settings" class="notification-channels">
            <div
              v-for="channel in settingsStore.settings.notification_channels"
              :key="channel.channel"
              class="channel-card"
            >
              <div class="channel-header">
                <div class="channel-info">
                  <span class="channel-name">{{ channelLabels[channel.channel] }}</span>
                  <el-tag :type="channel.enabled ? 'success' : 'info'" size="small">
                    {{ channel.enabled ? '已启用' : '已禁用' }}
                  </el-tag>
                </div>
                <el-switch
                  :model-value="channel.enabled"
                  @change="(val: boolean) => handleToggleChannel(channel.channel, val)"
                />
              </div>
              <div class="channel-body">
                <div v-if="channel.channel === 'telegram'" class="config-form">
                  <el-form label-width="100px" size="small">
                    <el-form-item label="Bot Token">
                      <el-input
                        :model-value="channel.config.bot_token"
                        type="password"
                        show-password
                        placeholder="请输入 Telegram Bot Token"
                      />
                    </el-form-item>
                    <el-form-item label="Chat ID">
                      <el-input
                        :model-value="channel.config.chat_id"
                        placeholder="请输入目标 Chat ID"
                      />
                    </el-form-item>
                  </el-form>
                </div>
                <div v-else-if="channel.channel === 'webhook'" class="config-form">
                  <el-form label-width="100px" size="small">
                    <el-form-item label="Webhook URL">
                      <el-input
                        :model-value="channel.config.url"
                        placeholder="请输入 Webhook URL"
                      />
                    </el-form-item>
                  </el-form>
                </div>
                <div v-else-if="channel.channel === 'email'" class="config-form">
                  <el-form label-width="100px" size="small">
                    <el-form-item label="SMTP 服务器">
                      <el-input
                        :model-value="channel.config.smtp_host"
                        placeholder="smtp.example.com"
                      />
                    </el-form-item>
                    <el-form-item label="发件人邮箱">
                      <el-input
                        :model-value="channel.config.from_email"
                        placeholder="noreply@example.com"
                      />
                    </el-form-item>
                  </el-form>
                </div>
                <div v-else class="config-placeholder">
                  <span>配置项开发中...</span>
                </div>
              </div>
              <div class="channel-footer">
                <el-button
                  size="small"
                  :disabled="!channel.enabled"
                  @click="handleTestChannel(channel.channel)"
                >
                  发送测试消息
                </el-button>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.notification-channels {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.channel-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
}

.channel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #fafafa;
  border-bottom: 1px solid #ebeef5;
}

.channel-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.channel-name {
  font-weight: 600;
  color: #303133;
}

.channel-body {
  padding: 16px;
  min-height: 120px;
}

.config-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100px;
  color: #909399;
}

.channel-footer {
  padding: 12px 16px;
  border-top: 1px solid #ebeef5;
  text-align: right;
}
</style>
