<template>
  <div class="dashboard">
    <!-- Welcome Hero Banner — Hope UI 蓝紫渐变 -->
    <div class="welcome-hero">
      <div class="welcome-hero__left">
        <span class="welcome-hero__wave">👋</span>
        <h1 class="welcome-hero__title">{{ timeGreeting }}，管理员</h1>
        <p class="welcome-hero__subtitle">今日健康概览 · 颐贞康养中心管理平台</p>
      </div>
      <div class="welcome-hero__right">
        <div class="meta-date">{{ currentDate }}</div>
        <div class="meta-status">
          <span class="status-dot-green"></span>
          <span>系统正常运行</span>
        </div>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard 组件 -->
    <div class="kpi-grid">
      <HopeStatCard
        value="online_devices"
        label="在线设备"
        :icon-color="'primary'"
        :gradient="'linear-gradient(135deg, #3a57e8, #6f42c1)'"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+2.3% 较昨日</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        value="total_users"
        label="活跃家属"
        icon-color="success"
        :gradient="'linear-gradient(135deg, #22c55e, #16a34a)'"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+5.1% 较昨日</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        value="active_alerts"
        label="待处理告警"
        icon-color="warning"
        :gradient="'linear-gradient(135deg, #f59e0b, #d97706)'"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-down">-12.5% 较昨日</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="deviceOnlineRate"
        label="设备在线率"
        icon-color="info"
        :gradient="'linear-gradient(135deg, #079aa2, #0ea5e9)'"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+1.2% 较上周</span>
        </template>
      </HopeStatCard>
    </div>

    <!-- Charts Row 1 -->
    <div class="charts-row">
      <HopeCard title="设备类型分布">
        <div ref="donutChartRef" class="chart-container"></div>
      </HopeCard>
      <HopeCard title="套餐订阅分布">
        <div ref="planChartRef" class="chart-container"></div>
      </HopeCard>
      <HopeCard title="告警优先级分布">
        <div ref="alertPriorityChartRef" class="chart-container"></div>
      </HopeCard>
    </div>

    <!-- Charts Row 2 -->
    <div class="charts-row">
      <HopeCard title="设备在线趋势" class="chart-card-wide">
        <div ref="lineChartRef" class="chart-container chart-tall"></div>
      </HopeCard>
      <HopeCard title="告警分布">
        <div ref="pieChartRef" class="chart-container chart-tall"></div>
      </HopeCard>
    </div>

    <!-- Bottom Row -->
    <div class="charts-row">
      <HopeCard title="最新告警" class="alert-card">
        <template #header>
          <HopeBtn variant="text" size="sm">查看全部 →</HopeBtn>
        </template>
        <HopeTable :columns="alertColumns" :data="alertTableData" :loading="false" class="alert-table">
          <template #col-alert_type="{ row }">
            <HopeBadge :color="alertBadgeColor(row.alert_type)">{{ row.alert_type }}</HopeBadge>
          </template>
          <template #col-status="{ row }">
            <HopeBadge :color="statusBadgeColor(row.status)">{{ statusLabel(row.status) }}</HopeBadge>
          </template>
        </HopeTable>
      </HopeCard>
      <HopeCard title="用户增长" class="chart-card">
        <template #header>
          <HopeBtn variant="text" size="sm">详情 →</HopeBtn>
        </template>
        <div ref="barChartRef" class="chart-container chart-tall"></div>
      </HopeCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch, nextTick, computed, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { useDashboardStore } from '@/stores/dashboard'
import { useTheme } from '@/composables/useTheme'
import type { Alert } from '@/types'
import { HopeCard, HopeBtn, HopeTable, HopeBadge, HopeStatCard } from '@/components/hope'

const store = useDashboardStore()
const { isDark } = useTheme()
const lineChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()
const donutChartRef = ref<HTMLElement>()
const planChartRef = ref<HTMLElement>()
const alertPriorityChartRef = ref<HTMLElement>()

let lineChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null
let barChart: echarts.ECharts | null = null
let donutChart: echarts.ECharts | null = null
let planChart: echarts.ECharts | null = null
let alertPriorityChart: echarts.ECharts | null = null

const timeGreeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '上午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const currentDate = computed(() => {
  return new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })
})

const alertTableData = ref<Array<Alert & { created_at: string }>>([])

watch(
  () => store.recentAlerts,
  (alerts) => {
    alertTableData.value = alerts.map(a => ({ ...a, created_at: a.created_at || '' }))
  },
  { immediate: true },
)

const alertColumns = [
  { prop: 'created_at', label: '时间', sortable: false },
  { prop: 'alert_type', label: '类型', sortable: false },
  { prop: 'dev_id', label: '设备', sortable: false },
  { prop: 'status', label: '状态', sortable: false },
]

function formatTime(dateStr?: string): string {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function alertBadgeColor(type: string): 'error' | 'warning' | 'primary' {
  if (['SOS', 'heart'].includes(type)) return 'error'
  if (['fall', 'medication'].includes(type)) return 'warning'
  return 'primary'
}

function statusBadgeColor(status: string): 'error' | 'success' | 'warning' {
  return status === 'pending' ? 'error' : status === 'resolved' ? 'success' : 'warning'
}

function statusLabel(status: string): string {
  return status === 'pending' ? '未处理' : status === 'resolved' ? '已处理' : '处理中'
}

/** Hope UI 蓝紫色 ECharts 主题 */
const hopeUIEChartsTheme = {
  color: [
    '#3a57e8', // primary 蓝紫
    '#8C57FF', // accent 紫
    '#22c55e', // success 绿
    '#f59e0b', // warning 橙
    '#079aa2', // info 青
    '#c04a42', // error 红
    '#6366f1', // indigo
    '#14b8a6', // teal
  ],
  backgroundColor: 'transparent',
  textStyle: {
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans SC", Roboto, sans-serif',
    fontSize: 13,
    fontWeight: 400,
  },
  title: {
    textStyle: { fontSize: 16, fontWeight: 600, color: 'var(--hope-text)' },
    subtextStyle: { fontSize: 13, color: 'var(--hope-text-muted)' },
  },
  legend: {
    textStyle: { fontSize: 13, color: 'var(--hope-text-secondary)', fontWeight: 500 },
    icon: 'roundRect',
    itemWidth: 14,
    itemHeight: 14,
    itemGap: 16,
  },
  tooltip: {
    backgroundColor: '#FFFFFF',
    borderColor: 'rgba(26,46,38,0.12)',
    borderWidth: 1,
    borderRadius: 10,
    padding: [10, 14],
    textStyle: { color: 'var(--hope-text)', fontSize: 13 },
    extraCssText: 'box-shadow: 0 4px 16px rgba(26,46,38,0.14);',
  },
  axis: {
    axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)', width: 1 } },
    axisLabel: { textStyle: { color: '#8a8d93', fontSize: 12 } },
    splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)', type: 'solid' } },
    splitArea: { areaStyle: { color: ['rgba(244,246,250,0.5)', 'transparent'] } },
  },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  line: {
    symbol: 'circle',
    symbolSize: 6,
    lineStyle: { width: 2.5, type: 'solid' },
    itemStyle: { borderWidth: 2 },
    areaStyle: {},
    emphasis: { focus: 'series' },
  },
  bar: {
    categoryAxis: {
      axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } },
      splitLine: { show: false },
    },
    valueAxis: {
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } },
    },
    emphasis: { focus: 'series' },
  },
  pie: {
    roseType: false,
    avoidLabelOverlap: true,
    itemStyle: { borderRadius: 6, borderColor: '#FFFFFF', borderWidth: 2 },
    label: { formatter: '{b}: {c}\n({d}%)', fontSize: 12, color: '#616161' },
    emphasis: {
      label: { fontSize: 13, fontWeight: 600 },
      itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.2)' },
    },
  },
}

// computed: device online rate
const deviceOnlineRate = computed(() => {
  if (!store.stats.total_devices) return '—'
  return Math.round((store.stats.online_devices / store.stats.total_devices) * 100) + '%'
})

function renderLineChart() {
  if (!lineChartRef.value) return
  if (!lineChart) lineChart = echarts.init(lineChartRef.value, hopeUIEChartsTheme)
  const trend = store.chartData.alertTrend
  const dates = trend.map(d => d.date)
  const bracelet = trend.map(d => d.bracelet_count)
  const pillbox = trend.map(d => d.pillbox_count)
  lineChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['手环', '药盒'], bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '12%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: dates.length ? dates : ['暂无数据'] },
    yAxis: { type: 'value' },
    series: [
      { name: '手环', type: 'line', smooth: true, data: bracelet.length ? bracelet : [0], itemStyle: { color: '#3a57e8' }, areaStyle: { opacity: 0.08 } },
      { name: '药盒', type: 'line', smooth: true, data: pillbox.length ? pillbox : [0], itemStyle: { color: '#8C57FF' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

function renderPieChart() {
  if (!pieChartRef.value) return
  if (!pieChart) pieChart = echarts.init(pieChartRef.value, hopeUIEChartsTheme)
  const items = store.chartData.alertDistribution
  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '告警类型', type: 'pie', radius: '60%',
      data: items.length
        ? items.map(i => ({ value: i.value, name: i.name, itemStyle: { color: i.color } }))
        : [
            { value: 35, name: 'SOS', itemStyle: { color: '#c04a42' } },
            { value: 28, name: '跌倒检测', itemStyle: { color: '#f59e0b' } },
            { value: 22, name: '心率异常', itemStyle: { color: '#3a57e8' } },
            { value: 15, name: '漏服药物', itemStyle: { color: '#8C57FF' } },
          ],
    }],
  })
}

function renderBarChart() {
  if (!barChartRef.value) return
  if (!barChart) barChart = echarts.init(barChartRef.value, hopeUIEChartsTheme)
  const growth = store.chartData.userGrowth
  barChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: growth.length ? growth.map(g => g.month) : ['2月', '3月', '4月', '5月', '6月', '7月'],
    },
    yAxis: { type: 'value' },
    series: [{
      name: '新增用户', type: 'bar', barWidth: '40%',
      data: growth.length ? growth.map(g => g.new_users) : [120, 180, 250, 320, 410, 520],
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3a57e8' },
          { offset: 1, color: '#8C57FF' },
        ]),
      },
    }],
  })
}

async function initCharts() {
  const ok = await store.refreshAll().catch(e => { console.warn('Dashboard API failed, using mock data:', e); return null })
  await nextTick()
  renderLineChart()
  renderPieChart()
  renderBarChart()
  renderDonutChart()
  renderPlanChart()
  renderAlertPriorityChart()
}

function renderDonutChart() {
  if (!donutChartRef.value) return
  if (!donutChart) donutChart = echarts.init(donutChartRef.value, hopeUIEChartsTheme)
  donutChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '设备类型', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 480, name: '手环-入门版', itemStyle: { color: '#3a57e8' } },
        { value: 312, name: '手环-中端版', itemStyle: { color: '#8C57FF' } },
        { value: 148, name: '手环-高端版', itemStyle: { color: '#6366f1' } },
        { value: 220, name: '药盒-智能版', itemStyle: { color: '#22c55e' } },
        { value: 85, name: '药盒-自动版', itemStyle: { color: '#f59e0b' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}' },
    }],
  })
}

function renderPlanChart() {
  if (!planChartRef.value) return
  if (!planChart) planChart = echarts.init(planChartRef.value, hopeUIEChartsTheme)
  planChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '套餐', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 189, name: 'Starter ¥29/月', itemStyle: { color: '#6366f1' } },
        { value: 312, name: 'Plus ¥59/月', itemStyle: { color: '#3a57e8' } },
        { value: 148, name: 'Pro ¥99/月', itemStyle: { color: '#8C57FF' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{d}%' },
    }],
  })
}

function renderAlertPriorityChart() {
  if (!alertPriorityChartRef.value) return
  if (!alertPriorityChart) alertPriorityChart = echarts.init(alertPriorityChartRef.value, hopeUIEChartsTheme)
  alertPriorityChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '告警优先级', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 12, name: 'P0 紧急', itemStyle: { color: '#c04a42' } },
        { value: 38, name: 'P1 重要', itemStyle: { color: '#f59e0b' } },
        { value: 156, name: 'P2 一般', itemStyle: { color: '#8a8d93' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}条' },
    }],
  })
}

watch(isDark, () => {
  nextTick(() => { initCharts() })
})

function handleResize() {
  lineChart?.resize(); pieChart?.resize(); barChart?.resize()
  donutChart?.resize(); planChart?.resize(); alertPriorityChart?.resize()
}

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  lineChart?.dispose(); pieChart?.dispose(); barChart?.dispose()
  donutChart?.dispose(); planChart?.dispose(); alertPriorityChart?.dispose()
})

onMounted(() => {
  initCharts()
  window.addEventListener('resize', handleResize)
})
</script>

<style scoped>
.dashboard { padding: 0; }

/* Welcome Hero — Hope UI 蓝紫渐变 */
.welcome-hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
  padding: 24px 28px;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 60%, #8C57FF 100%);
  border-radius: var(--hope-radius-xl);
  color: #fff;
  position: relative;
  overflow: hidden;
}
.welcome-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 80% 20%, rgba(255,255,255,0.12) 0%, transparent 60%);
  pointer-events: none;
}
.welcome-hero__left { position: relative; z-index: 1; }
.welcome-hero__wave { font-size: 28px; margin-right: 8px; }
.welcome-hero__title {
  font-size: 26px;
  font-weight: 800;
  margin: 0 0 4px;
  letter-spacing: -0.02em;
  color: #fff;
}
.welcome-hero__subtitle {
  font-size: 13px;
  color: rgba(255,255,255,0.75);
  margin: 0;
}
.welcome-hero__right {
  position: relative;
  z-index: 1;
  text-align: right;
}
.meta-date { font-size: 14px; font-weight: 600; color: rgba(255,255,255,0.9); margin-bottom: 4px; }
.meta-status {
  font-size: 12px;
  color: rgba(255,255,255,0.7);
  display: flex;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
}
.status-dot-green {
  width: 6px; height: 6px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 0 6px rgba(255,255,255,0.5);
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.15); }
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
.kpi-grid :deep(.hope-stat-card) { cursor: default; }

/* Charts Row */
.charts-row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}
.charts-row:nth-of-type(1),
.charts-row:nth-of-type(2) { grid-template-columns: repeat(3, 1fr); }
.charts-row:nth-of-type(3) { grid-template-columns: 2fr 1fr; }

.chart-container { height: 260px; }
.chart-container.tall { height: 300px; }

/* Alert Table */
.alert-table { margin-top: 8px; }
.alert-table :deep(.hope-table) { font-size: 13px; }
.alert-table :deep(.hope-table th) { font-size: 12px; }
.alert-table :deep(.hope-table td) { font-size: 13px; }

/* Responsive */
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-row:nth-of-type(1),
  .charts-row:nth-of-type(2) { grid-template-columns: 1fr; }
  .charts-row:nth-of-type(3) { grid-template-columns: 1fr; }
}

@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: 1fr; }
  .welcome-hero { flex-direction: column; align-items: flex-start; gap: 12px; padding: 16px 20px; }
  .welcome-hero__right { text-align: left; }
}
</style>
