<script setup lang="ts">
import { ref, computed } from 'vue'
import { useDashboardStore } from '@/stores'
import type { EventType } from '@/types'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const dashboardStore = useDashboardStore()
const filterType = ref<EventType | ''>('')
const filterStatus = ref<string>('')

const eventTypeOptions: { label: string; value: EventType | '' }[] = [
  { label: '全部类型', value: '' },
  { label: 'Push', value: 'push' },
  { label: 'Pull Request', value: 'pull_request' },
  { label: 'Issues', value: 'issues' },
  { label: 'Issue Comment', value: 'issue_comment' },
  { label: 'PR Review', value: 'pull_request_review' },
  { label: 'Create', value: 'create' },
  { label: 'Delete', value: 'delete' },
]

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '处理中', value: 'processing' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '待处理', value: 'pending' },
]

const filteredEvents = computed(() => {
  return dashboardStore.eventStream.filter((event) => {
    if (filterType.value && event.type !== filterType.value) return false
    if (filterStatus.value && event.status !== filterStatus.value) return false
    return true
  })
})

function getTypeTagClass(type: string) {
  return `event-type-tag ${type}`
}

function getStatusType(status: string) {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger',
  }
  return map[status] ?? ''
}

function getStatusText(status: string) {
  const map: Record<string, string> = {
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    failed: '失败',
  }
  return map[status] ?? status
}

function formatTime(time: string) {
  return dayjs(time).fromNow()
}

function formatFullTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function getEventSummary(event: typeof dashboardStore.eventStream[0]) {
  const payload = event.payload as Record<string, unknown>
  switch (event.type) {
    case 'push':
      return `推送到 ${payload.ref || 'unknown'}`
    case 'pull_request':
      return `PR: ${payload.action || ''} - ${(payload.pull_request as Record<string, unknown>)?.title || ''}`
    case 'issues':
      return `Issue: ${payload.action || ''} - ${(payload.issue as Record<string, unknown>)?.title || ''}`
    case 'issue_comment':
      return `评论: ${(payload.comment as Record<string, unknown>)?.body?.toString().slice(0, 50) || ''}`
    default:
      return `${event.type} 事件`
  }
}
</script>

<template>
  <div class="dashboard-card">
    <div class="dashboard-card__header">
      <div style="display: flex; align-items: center; gap: 8px">
        <span class="title">实时事件流</span>
        <el-tag v-if="dashboardStore.wsConnected" type="success" size="small" effect="light">
          <span class="pulse" style="margin-right: 4px">●</span>
          实时连接
        </el-tag>
        <el-tag v-else type="danger" size="small" effect="light">
          断开连接
        </el-tag>
      </div>
      <div style="display: flex; gap: 12px">
        <el-select v-model="filterType" placeholder="事件类型" size="small" style="width: 140px">
          <el-option
            v-for="opt in eventTypeOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
        <el-select v-model="filterStatus" placeholder="状态" size="small" style="width: 100px">
          <el-option
            v-for="opt in statusOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </div>
    </div>
    <div class="dashboard-card__body">
      <div class="event-stream">
        <div
          v-for="event in filteredEvents"
          :key="event.id"
          class="event-item"
          :class="{ failed: event.status === 'failed' }"
        >
          <div class="event-time">
            <span :title="formatFullTime(event.created_at)">
              {{ formatTime(event.created_at) }}
            </span>
          </div>
          <div class="event-content">
            <div class="event-header">
              <el-tag size="small" :class="getTypeTagClass(event.type)">
                {{ event.type }}
              </el-tag>
              <el-tag :type="getStatusType(event.status)" size="small">
                {{ getStatusText(event.status) }}
              </el-tag>
              <span class="event-source">{{ event.source }}</span>
            </div>
            <div class="event-summary">
              {{ getEventSummary(event) }}
            </div>
            <div v-if="event.error_message" class="event-error">
              <el-icon><Warning /></el-icon>
              {{ event.error_message }}
            </div>
          </div>
        </div>
        <el-empty v-if="filteredEvents.length === 0" description="暂无事件" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.event-stream {
  max-height: 400px;
  overflow-y: auto;
}

.event-item {
  display: flex;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    border-bottom: none;
  }

  &.failed {
    background-color: #fef0f0;
    margin: 0 -20px;
    padding: 12px 20px;
  }
}

.event-time {
  width: 80px;
  flex-shrink: 0;
  font-size: 12px;
  color: #909399;
}

.event-content {
  flex: 1;
  min-width: 0;
}

.event-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.event-source {
  font-size: 12px;
  color: #606266;
}

.event-summary {
  font-size: 14px;
  color: #303133;
  line-height: 1.5;
}

.event-error {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: 12px;
  color: #f56c6c;
}
</style>
