<template>
  <div class="analytics-page">
    <!-- Page Header -->
    <div class="hope-page-header">
      <div>
        <h1 class="hope-page-header__title">数据分析</h1>
        <p class="hope-page-header__subtitle">订阅收入 · 用户增长 · 业务经营指标</p>
      </div>
      <HopeBtn variant="filled" size="sm" @click="handleExport">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
        导出数据
      </HopeBtn>
    </div>

    <!-- KPI Cards (5) -->
    <div class="kpi-grid">
      <HopeStatCard :value="kpi.totalOrders" label="总订单数" icon-color="primary" gradient="linear-gradient(135deg, #3a57e8, #6f42c1)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4zM3 6h18M16 10a4 4 0 01-8 0"/></svg></el-icon>
        </template>
        <template #trend><span class="trend-up">+18.2% 较上月</span></template>
      </HopeStatCard>
      <HopeStatCard :value="kpi.totalUsers" label="总用户数" icon-color="success" gradient="linear-gradient(135deg, #1aa053, #0d9e6a)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon>
        </template>
        <template #trend><span class="trend-up">+12.5% 较上月</span></template>
      </HopeStatCard>
      <HopeStatCard :value="kpi.totalRevenue" label="总收入 (CNY)" icon-color="warning" gradient="linear-gradient(135deg, #FAA938, #f59e0b)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6"/></svg></el-icon>
        </template>
        <template #trend><span class="trend-up">+22.1% 较上月</span></template>
      </HopeStatCard>
      <HopeStatCard :value="kpi.avgOrderValue" label="平均客单价" icon-color="info" gradient="linear-gradient(135deg, #079aa2, #14b8a6)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg></el-icon>
        </template>
        <template #trend><span class="trend-down">-3.2% 较上月</span></template>
      </HopeStatCard>
      <HopeStatCard :value="kpi.renewalRate" label="续费率" icon-color="accent" gradient="linear-gradient(135deg, #8C57FF, #6f42c1)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg></el-icon>
        </template>
        <template #trend><span class="trend-up">+5.8% 较上月</span></template>
      </HopeStatCard>
    </div>

    <!-- Charts Row 1: Revenue Trend + Subscription Distribution -->
    <div class="charts-row">
      <HopeCard title="收入趋势">
        <div ref="revenueChartRef" class="chart-container"></div>
        <div class="chart-footer">按日期聚合 ｜ 本月累计 <strong>¥{{ revenueMonth }}</strong></div>
      </HopeCard>
      <HopeCard title="订阅分布">
        <div ref="subscriptionChartRef" class="chart-container"></div>
        <div class="chart-footer">按套餐档位分布</div>
      </HopeCard>
    </div>

    <!-- Charts Row 2: User Growth -->
    <div class="charts-row">
      <HopeCard title="用户增长趋势" :col-span="2">
        <div ref="growthChartRef" class="chart-container tall"></div>
        <div class="chart-footer">按月新增用户 ｜ 累计 <strong>{{ totalNewUsers }}</strong> 人</div>
      </HopeCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { HopeStatCard, HopeCard, HopeBtn } from '@/components/hope'
import * as echarts from 'echarts'
import 'echarts/theme/macarons.js'

// ── KPI Data ───────────────────────────────────────────────
const kpi = ref({
  totalOrders: '1,284',
  totalUsers: '3,592',
  totalRevenue: '¥286,400',
  avgOrderValue: '¥223',
  renewalRate: '87.3%',
})

// ── Subscriptions ──────────────────────────────────────────
const subscriptionData = ref([
  { name: 'Starter 入门版', value: 420 },
  { name: 'Plus 中端版', value: 310 },
  { name: 'Pro 高端版', value: 156 },
  { name: '免费版', value: 706 },
])
const revenueMonth = ref('28,640')

// ── User Growth ────────────────────────────────────────────
const growthData = ref([
  { month: '2月', newUsers: 1200 },
  { month: '3月', newUsers: 2100 },
  { month: '4月', newUsers: 3400 },
  { month: '5月', newUsers: 4800 },
  { month: '6月', newUsers: 6200 },
  { month: '7月', newUsers: 8234 },
])
const totalNewUsers = computed(() => growthData.value.reduce((s, g) => s + g.newUsers, 0).toLocaleString())

// ── Chart refs ─────────────────────────────────────────────
const revenueChartRef = ref<HTMLElement>()
const subscriptionChartRef = ref<HTMLElement>()
const growthChartRef = ref<HTMLElement>()

let revenueChart: echarts.ECharts | null = null
let subscriptionChart: echarts.ECharts | null = null
let growthChart: echarts.ECharts | null = null

// ── Render Functions ───────────────────────────────────────

function renderRevenueChart() {
  if (!revenueChartRef.value) return
  if (!revenueChart) revenueChart = echarts.init(revenueChartRef.value)
  const days = ['07-01','07-02','07-03','07-04','07-05','07-06','07-07','07-08','07-09','07-10','07-11','07-12','07-13','07-14']
  const data = [8200,9100,7800,10200,11500,9800,8600,10100,11200,12400,10800,9600,11000,12600]
  revenueMonth.value = data.reduce((s, v) => s + v, 0).toLocaleString()
  revenueChart.setOption({
    tooltip: { trigger: 'axis' },
    color: ['#3a57e8'],
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: days, axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, axisLabel: { textStyle: { color: '#8a8d93', fontSize: 11 } } },
    yAxis: { type: 'value', axisLabel: { formatter: (v: number) => `¥${(v/1000).toFixed(0)}k`, textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    series: [{
      name: '收入', type: 'line', smooth: true, symbol: 'circle', symbolSize: 6,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(58,87,232,0.20)' },
          { offset: 1, color: 'rgba(58,87,232,0.02)' },
        ]),
      },
      lineStyle: { width: 2.5 },
      itemStyle: { color: '#3a57e8' },
      data,
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, sans-serif' },
    backgroundColor: 'transparent',
  })
}

function renderSubscriptionChart() {
  if (!subscriptionChartRef.value) return
  if (!subscriptionChart) subscriptionChart = echarts.init(subscriptionChartRef.value)
  const colors = ['#3a57e8', '#1aa053', '#FAA938', '#8C57FF']
  subscriptionChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { orient: 'vertical', right: '5%', top: 'center', textStyle: { color: '#616161', fontSize: 12 } },
    color: colors,
    series: [{
      type: 'pie', radius: ['42%', '68%'], center: ['40%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: subscriptionData.value.map((d, i) => ({ ...d, itemStyle: { color: colors[i], borderRadius: 6 } })),
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, sans-serif' },
    backgroundColor: 'transparent',
  })
}

function renderGrowthChart() {
  if (!growthChartRef.value) return
  if (!growthChart) growthChart = echarts.init(growthChartRef.value)
  growthChart.setOption({
    tooltip: { trigger: 'axis' },
    color: ['#3a57e8'],
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: growthData.value.map(d => d.month), axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, axisLabel: { textStyle: { color: '#8a8d93' } } },
    yAxis: { type: 'value', axisLabel: { formatter: (v: number) => `${(v/1000).toFixed(0)}k`, textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    series: [{
      name: '新增用户', type: 'bar', barWidth: '40%',
      data: growthData.value.map(d => d.newUsers),
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3a57e8' }, { offset: 1, color: '#8C57FF' },
        ]),
        borderRadius: [6, 6, 0, 0],
      },
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, sans-serif' },
    backgroundColor: 'transparent',
  })
}

function handleExport() {
  ElMessage.info('导出功能开发中...')
}

// ── Lifecycle ──────────────────────────────────────────────

function handleResize() {
  revenueChart?.resize()
  subscriptionChart?.resize()
  growthChart?.resize()
}

onMounted(async () => {
  await nextTick()
  setTimeout(async () => {
    await nextTick()
    renderRevenueChart()
    renderSubscriptionChart()
    renderGrowthChart()
    window.addEventListener('resize', handleResize)
  }, 100)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  revenueChart?.dispose()
  subscriptionChart?.dispose()
  growthChart?.dispose()
})
</script>

<style scoped>
.analytics-page { padding: 0; }

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.charts-row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
  grid-template-columns: repeat(2, 1fr);
}

.chart-container { height: 260px; }
.chart-container.tall { height: 300px; }
.chart-footer {
  text-align: center; margin-top: 8px; font-size: 12px;
  color: var(--hope-text-muted);
}
.chart-footer strong { color: var(--hope-text); }

/* HopePageHeader override */
.hope-page-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 20px;
}

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-row { grid-template-columns: 1fr; }
}
</style>
