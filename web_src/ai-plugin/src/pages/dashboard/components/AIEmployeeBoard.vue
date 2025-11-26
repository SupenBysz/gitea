<script setup lang="ts">
import { computed } from 'vue'
import { useDashboardStore } from '@/stores'

const dashboardStore = useDashboardStore()

const employees = computed(() => dashboardStore.overview?.ai_employees ?? [])

function getStatusText(status: string) {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    busy: '工作中',
  }
  return map[status] ?? status
}

function getComplianceColor(rate: number) {
  if (rate >= 90) return '#67c23a'
  if (rate >= 70) return '#e6a23c'
  return '#f56c6c'
}
</script>

<template>
  <div class="dashboard-card">
    <div class="dashboard-card__header">
      <span class="title">AI 员工状态</span>
      <div class="status-summary">
        <span class="online-count">
          {{ employees.filter((e) => e.status === 'online' || e.status === 'busy').length }}
          / {{ employees.length }} 在线
        </span>
      </div>
    </div>
    <div class="dashboard-card__body">
      <div v-if="employees.length === 0" class="empty-state">
        <el-empty description="暂无 AI 员工" />
      </div>
      <div v-else class="employee-grid">
        <div
          v-for="employee in employees"
          :key="employee.id"
          class="employee-card"
          :class="employee.status"
        >
          <div class="employee-header">
            <el-avatar :size="40" :src="employee.avatar_url">
              {{ employee.name.charAt(0) }}
            </el-avatar>
            <div class="employee-info">
              <span class="name">{{ employee.name }}</span>
              <span class="username">@{{ employee.username }}</span>
            </div>
            <span class="status-badge" :class="employee.status">
              {{ getStatusText(employee.status) }}
            </span>
          </div>
          <div class="employee-body">
            <div class="current-task" v-if="employee.current_task">
              <el-icon><Document /></el-icon>
              <span class="text-truncate" :title="employee.current_task">
                {{ employee.current_task }}
              </span>
            </div>
            <div class="compliance-rate">
              <span class="label">合规率</span>
              <el-progress
                :percentage="employee.compliance_rate"
                :color="getComplianceColor(employee.compliance_rate)"
                :stroke-width="6"
                :show-text="true"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.status-summary {
  font-size: 12px;
  color: #67c23a;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.employee-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.employee-card {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #ebeef5;

  &.online {
    border-left: 3px solid #67c23a;
  }

  &.offline {
    border-left: 3px solid #909399;
    opacity: 0.7;
  }

  &.busy {
    border-left: 3px solid #e6a23c;
  }
}

.employee-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.employee-info {
  flex: 1;
  display: flex;
  flex-direction: column;

  .name {
    font-weight: 600;
    font-size: 14px;
    color: #303133;
  }

  .username {
    font-size: 12px;
    color: #909399;
  }
}

.employee-body {
  margin-top: 12px;
}

.current-task {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #606266;
  padding: 6px 8px;
  background-color: #fff;
  border-radius: 4px;
  margin-bottom: 8px;

  .el-icon {
    color: #409eff;
    flex-shrink: 0;
  }
}

.compliance-rate {
  .label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 4px;
    display: block;
  }
}
</style>
