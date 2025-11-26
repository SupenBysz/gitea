<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useDashboardStore } from '@/stores'
import OverviewCards from './components/OverviewCards.vue'
import ComplianceTrendChart from './components/ComplianceTrendChart.vue'
import ViolationStatsChart from './components/ViolationStatsChart.vue'
import BlockedIssuesList from './components/BlockedIssuesList.vue'
import AIEmployeeBoard from './components/AIEmployeeBoard.vue'
import EventStreamList from './components/EventStreamList.vue'

const dashboardStore = useDashboardStore()

onMounted(async () => {
  await dashboardStore.fetchAll()
  dashboardStore.connectWebSocket()
})

onUnmounted(() => {
  dashboardStore.disconnectWebSocket()
})
</script>

<template>
  <div class="dashboard-page">
    <div class="page-header">
      <h1 class="page-header__title">仪表盘</h1>
      <p class="page-header__desc">AI 协同工作规范实时监控与分析</p>
    </div>

    <!-- Overview Cards -->
    <OverviewCards :loading="dashboardStore.loading" />

    <!-- Charts Row -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="16">
        <ComplianceTrendChart />
      </el-col>
      <el-col :span="8">
        <ViolationStatsChart />
      </el-col>
    </el-row>

    <!-- Data Row -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="12">
        <BlockedIssuesList />
      </el-col>
      <el-col :span="12">
        <AIEmployeeBoard />
      </el-col>
    </el-row>

    <!-- Event Stream -->
    <div class="mt-20">
      <EventStreamList />
    </div>
  </div>
</template>

<style scoped>
.dashboard-page {
  min-height: 100%;
}
</style>
