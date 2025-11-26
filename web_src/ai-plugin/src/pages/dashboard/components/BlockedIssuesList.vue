<script setup lang="ts">
import { useDashboardStore } from '@/stores'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const dashboardStore = useDashboardStore()

function getSeverityType(severity: string) {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    critical: 'danger',
    high: 'warning',
    medium: 'info',
    low: '',
  }
  return map[severity] ?? ''
}

function formatTime(time: string) {
  return dayjs(time).fromNow()
}

function viewIssue(issue: typeof dashboardStore.blockedIssues[0]) {
  window.open(`/${issue.repo_owner}/${issue.repo_name}/issues/${issue.issue_number}`, '_blank')
}
</script>

<template>
  <div class="dashboard-card">
    <div class="dashboard-card__header">
      <span class="title">阻塞工单</span>
      <el-button type="primary" link size="small">
        查看全部
      </el-button>
    </div>
    <div class="dashboard-card__body">
      <el-table
        :data="dashboardStore.blockedIssues"
        style="width: 100%"
        max-height="300"
        :show-header="true"
        size="small"
      >
        <el-table-column label="工单" min-width="180">
          <template #default="{ row }">
            <div class="issue-info">
              <span class="issue-title text-truncate" :title="row.title">
                {{ row.title }}
              </span>
              <span class="issue-repo">
                {{ row.repo_owner }}/{{ row.repo_name }}#{{ row.issue_number }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="严重级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)" size="small">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="阻塞原因" min-width="150">
          <template #default="{ row }">
            <span class="text-truncate" :title="row.blocked_reason">
              {{ row.blocked_reason }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="阻塞时间" width="100">
          <template #default="{ row }">
            <span :title="row.blocked_at">{{ formatTime(row.blocked_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewIssue(row)">
              查看
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty
        v-if="dashboardStore.blockedIssues.length === 0"
        description="暂无阻塞工单"
      />
    </div>
  </div>
</template>

<style scoped>
.issue-info {
  display: flex;
  flex-direction: column;
}

.issue-title {
  font-weight: 500;
  color: #303133;
  max-width: 200px;
}

.issue-repo {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
</style>
