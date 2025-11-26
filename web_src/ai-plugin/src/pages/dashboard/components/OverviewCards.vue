<script setup lang="ts">
import { computed } from 'vue'
import { useDashboardStore } from '@/stores'
import { TrendCharts, Warning, Checked, User } from '@element-plus/icons-vue'

defineProps<{
  loading: boolean
}>()

const dashboardStore = useDashboardStore()

const cards = computed(() => [
  {
    title: '今日事件',
    value: dashboardStore.overview?.events_today ?? 0,
    total: dashboardStore.overview?.total_events ?? 0,
    icon: TrendCharts,
    type: 'primary',
    suffix: '/ ' + (dashboardStore.overview?.total_events ?? 0) + ' 总计',
  },
  {
    title: '合规率',
    value: dashboardStore.complianceRate.toFixed(1),
    icon: Checked,
    type: 'success',
    suffix: '%',
  },
  {
    title: '阻塞工单',
    value: dashboardStore.blockedCount,
    icon: Warning,
    type: 'danger',
    suffix: '个待处理',
  },
  {
    title: '活跃规则',
    value: dashboardStore.activeRules,
    total: dashboardStore.overview?.total_rules ?? 0,
    icon: User,
    type: 'warning',
    suffix: '/ ' + (dashboardStore.overview?.total_rules ?? 0) + ' 总计',
  },
])
</script>

<template>
  <el-row :gutter="20">
    <el-col v-for="card in cards" :key="card.title" :span="6">
      <el-skeleton :loading="loading" animated>
        <template #template>
          <div class="stat-card">
            <el-skeleton-item variant="circle" style="width: 64px; height: 64px" />
            <div class="stat-card__content" style="margin-left: 16px">
              <el-skeleton-item variant="h3" style="width: 60%" />
              <el-skeleton-item variant="text" style="width: 40%; margin-top: 8px" />
            </div>
          </div>
        </template>
        <template #default>
          <div class="stat-card">
            <div class="stat-card__icon" :class="card.type">
              <el-icon :size="28"><component :is="card.icon" /></el-icon>
            </div>
            <div class="stat-card__content">
              <div class="value">
                {{ card.value }}
                <span v-if="card.suffix" style="font-size: 14px; color: #909399; font-weight: normal">
                  {{ card.suffix }}
                </span>
              </div>
              <div class="label">{{ card.title }}</div>
            </div>
          </div>
        </template>
      </el-skeleton>
    </el-col>
  </el-row>
</template>
