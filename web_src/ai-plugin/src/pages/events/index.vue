<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardStore } from '@/stores'
import type { EventType } from '@/types'
import dayjs from 'dayjs'

const dashboardStore = useDashboardStore()

const filterType = ref<EventType | ''>('')
const filterStatus = ref<string>('')
const dateRange = ref<[Date, Date] | null>(null)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

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
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
]

async function fetchEvents() {
  const params: Record<string, unknown> = {
    page: page.value,
    page_size: pageSize.value,
  }
  if (filterType.value) params.type = filterType.value
  if (filterStatus.value) params.status = filterStatus.value
  if (dateRange.value) {
    params.start_date = dayjs(dateRange.value[0]).format('YYYY-MM-DD')
    params.end_date = dayjs(dateRange.value[1]).format('YYYY-MM-DD')
  }
  await dashboardStore.fetchEventStream(params as { page?: number; page_size?: number })
}

function handleSearch() {
  page.value = 1
  fetchEvents()
}

function handlePageChange(p: number) {
  page.value = p
  fetchEvents()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  page.value = 1
  fetchEvents()
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
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function formatDuration(ms: number | undefined) {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

onMounted(() => {
  fetchEvents()
})
</script>

<template>
  <div class="events-page">
    <div class="page-header">
      <h1 class="page-header__title">事件日志</h1>
      <p class="page-header__desc">查看所有 Webhook 事件及其处理状态</p>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-select v-model="filterType" placeholder="事件类型" style="width: 160px" clearable>
        <el-option
          v-for="opt in eventTypeOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-select v-model="filterStatus" placeholder="处理状态" style="width: 120px" clearable>
        <el-option
          v-for="opt in statusOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        style="width: 280px"
      />
      <el-button type="primary" @click="handleSearch">
        <el-icon><Search /></el-icon>
        搜索
      </el-button>
    </div>

    <!-- Events Table -->
    <div class="dashboard-card mt-20">
      <el-table
        v-loading="dashboardStore.loading"
        :data="dashboardStore.eventStream"
        style="width: 100%"
        row-key="id"
      >
        <el-table-column label="事件 ID" width="220">
          <template #default="{ row }">
            <el-tooltip :content="row.id" placement="top">
              <span class="event-id">{{ row.id.slice(0, 8) }}...</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="140">
          <template #default="{ row }">
            <el-tag size="small" :class="`event-type-tag ${row.type}`">
              {{ row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="180">
          <template #default="{ row }">
            <span class="text-truncate" :title="row.source">{{ row.source }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="处理时间" width="180">
          <template #default="{ row }">
            {{ row.processed_at ? formatTime(row.processed_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="200">
          <template #default="{ row }">
            <span v-if="row.error_message" class="error-message text-truncate" :title="row.error_message">
              {{ row.error_message }}
            </span>
            <span v-else class="no-error">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 16px 20px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.event-id {
  font-family: monospace;
  color: #606266;
}

.error-message {
  color: #f56c6c;
  max-width: 200px;
  display: inline-block;
}

.no-error {
  color: #909399;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>
