<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import * as echarts from 'echarts'
import { useDashboardStore } from '@/stores'

const dashboardStore = useDashboardStore()
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const chartOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: {c} ({d}%)',
  },
  legend: {
    orient: 'vertical',
    right: 10,
    top: 'center',
    formatter: (name: string) => {
      const item = dashboardStore.violationStats.find((s) => s.rule_name === name)
      return item ? `${name} (${item.count})` : name
    },
  },
  series: [
    {
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 4,
        borderColor: '#fff',
        borderWidth: 2,
      },
      label: {
        show: false,
        position: 'center',
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 16,
          fontWeight: 'bold',
        },
      },
      labelLine: {
        show: false,
      },
      data: dashboardStore.violationStats.map((item, index) => ({
        name: item.rule_name,
        value: item.count,
        itemStyle: {
          color: [
            '#f56c6c',
            '#e6a23c',
            '#409eff',
            '#67c23a',
            '#909399',
          ][index % 5],
        },
      })),
    },
  ],
}))

function initChart() {
  if (chartRef.value) {
    chartInstance = echarts.init(chartRef.value)
    chartInstance.setOption(chartOption.value)
  }
}

function updateChart() {
  if (chartInstance) {
    chartInstance.setOption(chartOption.value)
  }
}

watch(
  () => dashboardStore.violationStats,
  () => updateChart(),
  { deep: true }
)

onMounted(() => {
  initChart()
  window.addEventListener('resize', () => chartInstance?.resize())
})
</script>

<template>
  <div class="dashboard-card">
    <div class="dashboard-card__header">
      <span class="title">违规分布</span>
    </div>
    <div class="dashboard-card__body">
      <div v-if="dashboardStore.violationStats.length === 0" class="empty-state">
        <el-empty description="暂无违规数据" />
      </div>
      <div v-else ref="chartRef" style="width: 100%; height: 300px" />
    </div>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 300px;
}
</style>
