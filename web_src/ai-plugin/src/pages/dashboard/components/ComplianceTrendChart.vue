<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import * as echarts from 'echarts'
import { useDashboardStore } from '@/stores'
import dayjs from 'dayjs'

const dashboardStore = useDashboardStore()
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const dateRange = ref<[Date, Date]>([
  dayjs().subtract(7, 'day').toDate(),
  dayjs().toDate(),
])

const granularity = ref<'hour' | 'day' | 'week' | 'month'>('day')

const granularityOptions = [
  { label: '小时', value: 'hour' },
  { label: '天', value: 'day' },
  { label: '周', value: 'week' },
  { label: '月', value: 'month' },
]

const chartOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'cross',
    },
  },
  legend: {
    data: ['合规率', '事件数', '违规数'],
    bottom: 0,
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '15%',
    containLabel: true,
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: dashboardStore.complianceTrend.map((item) =>
      dayjs(item.date).format(granularity.value === 'hour' ? 'MM-DD HH:mm' : 'MM-DD')
    ),
  },
  yAxis: [
    {
      type: 'value',
      name: '合规率 (%)',
      min: 0,
      max: 100,
      position: 'left',
      axisLabel: {
        formatter: '{value}%',
      },
    },
    {
      type: 'value',
      name: '数量',
      position: 'right',
    },
  ],
  series: [
    {
      name: '合规率',
      type: 'line',
      smooth: true,
      yAxisIndex: 0,
      data: dashboardStore.complianceTrend.map((item) => item.compliance_rate.toFixed(1)),
      itemStyle: { color: '#67c23a' },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
          { offset: 1, color: 'rgba(103, 194, 58, 0.05)' },
        ]),
      },
    },
    {
      name: '事件数',
      type: 'bar',
      yAxisIndex: 1,
      data: dashboardStore.complianceTrend.map((item) => item.total_events),
      itemStyle: { color: '#409eff' },
    },
    {
      name: '违规数',
      type: 'bar',
      yAxisIndex: 1,
      data: dashboardStore.complianceTrend.map((item) => item.violations),
      itemStyle: { color: '#f56c6c' },
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

async function fetchData() {
  const [start, end] = dateRange.value
  await dashboardStore.fetchComplianceTrend({
    start_date: dayjs(start).format('YYYY-MM-DD'),
    end_date: dayjs(end).format('YYYY-MM-DD'),
    granularity: granularity.value,
  })
  updateChart()
}

watch([dateRange, granularity], fetchData)

watch(
  () => dashboardStore.complianceTrend,
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
      <span class="title">合规趋势</span>
      <div style="display: flex; gap: 12px">
        <el-select v-model="granularity" style="width: 100px" size="small">
          <el-option
            v-for="opt in granularityOptions"
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
          size="small"
          style="width: 240px"
        />
      </div>
    </div>
    <div class="dashboard-card__body">
      <div ref="chartRef" style="width: 100%; height: 300px" />
    </div>
  </div>
</template>
