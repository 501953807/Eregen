<template>
  <div class="dashboard">
    <!-- KPI Cards Row — Hope UI style with circle progress -->
    <div class="kpi-grid">
      <HopeStatCard
        value="560K"
        label="Total Devices"
        :icon-color="'primary'"
        :gradient="'linear-gradient(135deg, #3a57e8, #6f42c1)'"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
          </svg>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+2.3%</span>
        </template>
        <template #progress>
          <CircleProgress :value="78" :color="'#3a57e8'" :bg-color="'rgba(58,87,232,0.12)'"/>
        </template>
      </HopeStatCard>
      <HopeStatCard
        value="312"
        label="Active Families"
        icon-color="success"
        :gradient="'linear-gradient(135deg, #22c55e, #16a34a)'"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/>
            <path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
          </svg>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+5.1%</span>
        </template>
        <template #progress>
          <CircleProgress :value="92" :color="'#22c55e'" :bg-color="'rgba(34,197,94,0.12)'"/>
        </template>
      </HopeStatCard>
      <HopeStatCard
        value="12"
        label="P0 Alerts"
        icon-color="error"
        :gradient="'linear-gradient(135deg, #c03221, #e74c3c)'"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/>
            <path d="M13.73 21a2 2 0 01-3.46 0"/>
          </svg>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-down">-12.5%</span>
        </template>
        <template #progress>
          <CircleProgress :value="15" :color="'#c03221'" :bg-color="'rgba(192,50,33,0.12)'"/>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="deviceOnlineRate"
        label="Online Rate"
        icon-color="info"
        :gradient="'linear-gradient(135deg, #079aa2, #0ea5e9)'"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/>
          </svg>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+1.2%</span>
        </template>
        <template #progress>
          <CircleProgress :value="88" :color="'#079aa2'" :bg-color="'rgba(7,154,162,0.12)'"/>
        </template>
      </HopeStatCard>
    </div>

    <!-- Charts Row 1: Line chart + 2 donut charts -->
    <div class="charts-row">
      <HopeCard title="Device Trend" class="chart-card-wide">
        <template #header>
          <div class="chart-header-right">
            <select class="chart-period-select">
              <option>This Week</option>
              <option>Last Week</option>
              <option>This Month</option>
            </select>
          </div>
        </template>
        <div ref="lineChartRef" class="chart-container"></div>
      </HopeCard>
      <HopeCard title="Device Distribution" class="chart-card">
        <div ref="donutChartRef" class="chart-container chart-medium"></div>
      </HopeCard>
      <HopeCard title="Alert Priority" class="chart-card">
        <div ref="alertPriorityChartRef" class="chart-container chart-medium"></div>
      </HopeCard>
    </div>

    <!-- Charts Row 2: Earnings + Conversions -->
    <div class="charts-row-row2">
      <HopeCard title="Earnings Overview">
        <template #header>
          <div class="chart-header-right">
            <select class="chart-period-select">
              <option>This Week</option>
              <option>Last Week</option>
            </select>
          </div>
        </template>
        <div ref="pieChartRef" class="chart-container chart-wide"></div>
      </HopeCard>
      <HopeCard title="Alert Distribution">
        <template #header>
          <div class="chart-header-right">
            <select class="chart-period-select">
              <option>This Week</option>
              <option>Last Week</option>
            </select>
          </div>
        </template>
        <div ref="barChartRef" class="chart-container chart-wide"></div>
      </HopeCard>
      <!-- Right side: Credit card widget + visitor stats -->
      <div class="right-widgets">
        <div class="vip-card-widget">
          <div class="vip-card-bg"></div>
          <div class="vip-card-content">
            <div class="vip-card-top">
              <span class="vip-card-label">PREMIUM ACCOUNT</span>
              <div class="vip-card-chips">
                <div class="chip"></div>
                <div class="chip chip-alt"></div>
              </div>
            </div>
            <div class="vip-card-number">5789 **** **** 2847</div>
            <div class="vip-card-bottom">
              <div class="vip-card-holder">
                <span class="vip-card-holder-label">Card holder</span>
                <span class="vip-card-holder-name">{{ authStore.user?.name || 'Admin' }}</span>
              </div>
              <div class="vip-card-expiry">
                <span class="vip-card-holder-label">Expires</span>
                <span class="vip-card-expiry-val">06/28</span>
              </div>
            </div>
          </div>
        </div>
        <div class="stat-pair">
          <div class="stat-pair-item">
            <div class="stat-pair-value">{{ store.stats.online_devices || '—' }}</div>
            <div class="stat-pair-label">Online Devices</div>
          </div>
          <div class="stat-pair-item">
            <div class="stat-pair-value">{{ store.stats.total_users || '—' }}</div>
            <div class="stat-pair-label">Total Users</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Row: Alerts table + User growth -->
    <div class="bottom-row">
      <HopeCard title="Recent Alerts" class="alert-card">
        <template #header>
          <HopeBtn variant="text" size="sm">View All →</HopeBtn>
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
      <HopeCard title="User Growth" class="chart-card">
        <template #header>
          <HopeBtn variant="text" size="sm">Details →</HopeBtn>
        </template>
        <div ref="growthChartRef" class="chart-container chart-tall"></div>
      </HopeCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch, nextTick, computed, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { useDashboardStore } from '@/stores/dashboard'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import type { Alert } from '@/types'
import { HopeCard, HopeBtn, HopeTable, HopeBadge, HopeStatCard } from '@/components/hope'
import CircleProgress from '@/components/common/CircleProgress.vue'

const store = useDashboardStore()
const authStore = useAuthStore()
const { isDark } = useTheme()
const lineChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()
const growthChartRef = ref<HTMLElement>()
const donutChartRef = ref<HTMLElement>()
const alertPriorityChartRef = ref<HTMLElement>()

let lineChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null
let barChart: echarts.ECharts | null = null
let growthChart: echarts.ECharts | null = null
let donutChart: echarts.ECharts | null = null
let alertPriorityChart: echarts.ECharts | null = null

const alertTableData = ref<Array<Alert & { created_at: string }>>([])

watch(
  () => store.recentAlerts,
  (alerts) => {
    alertTableData.value = alerts.map(a => ({ ...a, created_at: a.created_at || '' }))
  },
  { immediate: true },
)

const alertColumns = [
  { prop: 'created_at', label: 'Time', sortable: false },
  { prop: 'alert_type', label: 'Type', sortable: false },
  { prop: 'dev_id', label: 'Device', sortable: false },
  { prop: 'status', label: 'Status', sortable: false },
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
  return status === 'pending' ? 'Pending' : status === 'resolved' ? 'Resolved' : 'Processing'
}

const deviceOnlineRate = computed(() => {
  if (!store.stats.total_devices) return '—'
  return Math.round((store.stats.online_devices / store.stats.total_devices) * 100) + '%'
})

/** Hope UI ECharts theme */
const hopeUIEChartsTheme = {
  color: ['#3a57e8', '#8C57FF', '#22c55e', '#f59e0b', '#079aa2', '#c04a42', '#6366f1', '#14b8a6'],
  backgroundColor: 'transparent',
  textStyle: { fontFamily: 'Inter, "Noto Sans SC", sans-serif', fontSize: 13, fontWeight: 400 },
  title: { textStyle: { fontSize: 16, fontWeight: 600, color: 'var(--hope-text)' } },
  legend: { textStyle: { fontSize: 13, color: 'var(--hope-text-secondary)', fontWeight: 500 }, icon: 'roundRect', itemWidth: 14, itemHeight: 14, itemGap: 16 },
  tooltip: { backgroundColor: '#FFFFFF', borderColor: 'rgba(26,46,38,0.12)', borderWidth: 1, borderRadius: 10, padding: [10, 14], textStyle: { color: 'var(--hope-text)', fontSize: 13 }, extraCssText: 'box-shadow: 0 4px 16px rgba(26,46,38,0.14);' },
  axis: { axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)', width: 1 } }, axisLabel: { textStyle: { color: '#8a8d93', fontSize: 12 } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)', type: 'solid' } } },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  line: { symbol: 'circle', symbolSize: 6, lineStyle: { width: 2.5, type: 'solid' }, itemStyle: { borderWidth: 2 }, areaStyle: {}, emphasis: { focus: 'series' } },
  bar: { categoryAxis: { axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, splitLine: { show: false } }, valueAxis: { axisLine: { show: false }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } }, emphasis: { focus: 'series' } },
  pie: { roseType: false, avoidLabelOverlap: true, itemStyle: { borderRadius: 6, borderColor: '#FFFFFF', borderWidth: 2 }, label: { formatter: '{b}: {c}\n({d}%)', fontSize: 12, color: '#616161' }, emphasis: { label: { fontSize: 13, fontWeight: 600 }, itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.2)' } } },
}

function renderLineChart() {
  if (!lineChartRef.value) return
  if (!lineChart) lineChart = echarts.init(lineChartRef.value, hopeUIEChartsTheme)
  const trend = store.chartData.alertTrend
  lineChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['Bracelet', 'Pillbox'], bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '12%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: trend.length ? trend.map(d => d.date) : ['暂无数据'] },
    yAxis: { type: 'value' },
    series: [
      { name: 'Bracelet', type: 'line', smooth: true, data: trend.length ? trend.map(d => d.bracelet_count) : [0], itemStyle: { color: '#3a57e8' }, areaStyle: { opacity: 0.08 } },
      { name: 'Pillbox', type: 'line', smooth: true, data: trend.length ? trend.map(d => d.pillbox_count) : [0], itemStyle: { color: '#8C57FF' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

function renderPieChart() {
  if (!pieChartRef.value) return
  if (!pieChart) pieChart = echarts.init(pieChartRef.value, hopeUIEChartsTheme)
  const items = store.chartData.alertDistribution
  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', right: '5%', top: 'center' },
    series: [{
      name: 'Alert Type', type: 'pie', radius: ['0%', '65%'], center: ['35%', '50%'],
      data: items.length
        ? items.map(i => ({ value: i.value, name: i.name, itemStyle: { color: i.color } }))
        : [
            { value: 35, name: 'SOS', itemStyle: { color: '#c04a42' } },
            { value: 28, name: 'Fall', itemStyle: { color: '#f59e0b' } },
            { value: 22, name: 'Heart', itemStyle: { color: '#3a57e8' } },
            { value: 15, name: 'Medication', itemStyle: { color: '#8C57FF' } },
          ],
    }],
  })
}

function renderBarChart() {
  if (!barChartRef.value) return
  if (!barChart) barChart = echarts.init(barChartRef.value, hopeUIEChartsTheme)
  const dist = store.chartData.alertDistribution
  barChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: dist.length ? dist.map(d => d.name) : ['SOS', 'Fall', 'Heart', 'Med'] },
    yAxis: { type: 'value' },
    series: [{
      name: 'Alerts', type: 'bar', barWidth: '40%',
      data: dist.length ? dist.map(d => d.value) : [35, 28, 22, 15],
      itemStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: '#c04a42' }, { offset: 1, color: '#8C57FF' }]), borderRadius: [4, 4, 0, 0] },
    }],
  })
}

function renderGrowthChart() {
  if (!growthChartRef.value) return
  if (!growthChart) growthChart = echarts.init(growthChartRef.value, hopeUIEChartsTheme)
  const growth = store.chartData.userGrowth
  growthChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: growth.length ? growth.map(g => g.month) : ['Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul'] },
    yAxis: { type: 'value' },
    series: [{
      name: 'New Users', type: 'bar', barWidth: '40%',
      data: growth.length ? growth.map(g => g.new_users) : [120, 180, 250, 320, 410, 520],
      itemStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: '#3a57e8' }, { offset: 1, color: '#8C57FF' }]), borderRadius: [4, 4, 0, 0] },
    }],
  })
}

function renderDonutChart() {
  if (!donutChartRef.value) return
  if (!donutChart) donutChart = echarts.init(donutChartRef.value, hopeUIEChartsTheme)
  donutChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: 'Device Type', type: 'pie', radius: ['42%', '72%'], center: ['50%', '50%'],
      data: [
        { value: 480, name: 'Bracelet-Entry', itemStyle: { color: '#3a57e8' } },
        { value: 312, name: 'Bracelet-Plus', itemStyle: { color: '#8C57FF' } },
        { value: 148, name: 'Bracelet-Pro', itemStyle: { color: '#6366f1' } },
        { value: 220, name: 'Pillbox-Smart', itemStyle: { color: '#22c55e' } },
        { value: 85, name: 'Pillbox-Auto', itemStyle: { color: '#f59e0b' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}' },
    }],
  })
}

function renderAlertPriorityChart() {
  if (!alertPriorityChartRef.value) return
  if (!alertPriorityChart) alertPriorityChart = echarts.init(alertPriorityChartRef.value, hopeUIEChartsTheme)
  alertPriorityChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: 'Alert Priority', type: 'pie', radius: ['42%', '72%'], center: ['50%', '50%'],
      data: [
        { value: 12, name: 'P0 Urgent', itemStyle: { color: '#c04a42' } },
        { value: 38, name: 'P1 Important', itemStyle: { color: '#f59e0b' } },
        { value: 156, name: 'P2 Normal', itemStyle: { color: '#8a8d93' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c} items' },
    }],
  })
}

async function initCharts() {
  await store.refreshAll().catch(e => { console.warn('Dashboard API failed:', e); return null })
  await nextTick()
  renderLineChart()
  renderPieChart()
  renderBarChart()
  renderGrowthChart()
  renderDonutChart()
  renderAlertPriorityChart()
}

watch(isDark, () => { nextTick(() => { initCharts() }) })

function handleResize() {
  lineChart?.resize(); pieChart?.resize(); barChart?.resize(); growthChart?.resize()
  donutChart?.resize(); alertPriorityChart?.resize()
}

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  lineChart?.dispose(); pieChart?.dispose(); barChart?.dispose(); growthChart?.dispose()
  donutChart?.dispose(); alertPriorityChart?.dispose()
})

onMounted(() => { initCharts(); window.addEventListener('resize', handleResize) })
</script>

<style scoped>
.dashboard { padding: 0; }

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin: 0 1.5rem 20px;
}
.kpi-grid :deep(.hope-stat-card) { cursor: default; }

/* Charts Row 1 */
.charts-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 20px;
  margin: 0 1.5rem 20px;
}

/* Charts Row 2 */
.charts-row-row2 {
  display: grid;
  grid-template-columns: 1fr 1fr 320px;
  gap: 20px;
  margin: 0 1.5rem 20px;
}

/* Bottom Row */
.bottom-row {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 20px;
  margin: 0 1.5rem 20px;
}

/* Chart containers */
.chart-container { height: 260px; }
.chart-container.medium { height: 200px; }
.chart-container.tall { height: 300px; }
.chart-container.wide { height: 240px; }
.chart-card-wide { min-height: 320px; }

/* Chart header right */
.chart-header-right { display: flex; align-items: center; gap: 8px; }
.chart-period-select {
  padding: 4px 10px;
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-sm);
  background: var(--hope-surface-light);
  font-size: 12px;
  color: var(--hope-text-secondary);
  cursor: pointer;
  outline: none;
}
.chart-period-select:focus { border-color: var(--hope-primary); }

/* Alert table */
.alert-table { margin-top: 8px; }
.alert-table :deep(.hope-table) { font-size: 13px; }
.alert-table :deep(.hope-table th) { font-size: 12px; }
.alert-table :deep(.hope-table td) { font-size: 13px; }

/* Right widgets */
.right-widgets { display: flex; flex-direction: column; gap: 16px; }

/* VIP Card Widget */
.vip-card-widget {
  position: relative;
  border-radius: var(--hope-radius-lg);
  overflow: hidden;
  padding: 20px;
  color: #fff;
  min-height: 140px;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 50%, #8C57FF 100%);
}
.vip-card-bg {
  position: absolute;
  top: -30px;
  right: -30px;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: rgba(255,255,255,0.1);
}
.vip-card-bg::before {
  content: '';
  position: absolute;
  top: 20px;
  left: 20px;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(255,255,255,0.06);
}
.vip-card-content { position: relative; z-index: 1; }
.vip-card-top { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.vip-card-label { font-size: 11px; font-weight: 600; letter-spacing: 0.1em; opacity: 0.8; }
.vip-card-chips { display: flex; gap: 4px; }
.chip { width: 28px; height: 28px; border-radius: 50%; background: rgba(255,255,255,0.25); }
.chip-alt { background: rgba(255,255,255,0.15); }
.vip-card-number { font-size: 18px; font-weight: 600; letter-spacing: 0.15em; margin-bottom: 16px; }
.vip-card-bottom { display: flex; justify-content: space-between; }
.vip-card-holder-label, .vip-card-expiry-label { font-size: 10px; opacity: 0.6; text-transform: uppercase; letter-spacing: 0.05em; }
.vip-card-holder-name { font-size: 13px; font-weight: 600; }
.vip-card-expiry-val { font-size: 13px; font-weight: 600; }

/* Stat pair */
.stat-pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.stat-pair-item {
  background: var(--hope-surface);
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-md);
  padding: 14px 16px;
  text-align: center;
}
.stat-pair-value {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  letter-spacing: -0.02em;
}
.stat-pair-label {
  font-size: 11px;
  color: var(--hope-text-muted);
  margin-top: 2px;
  font-weight: 500;
}

/* Responsive */
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-row { grid-template-columns: 1fr; }
  .charts-row-row2 { grid-template-columns: 1fr; }
  .bottom-row { grid-template-columns: 1fr; }
  .right-widgets { flex-direction: row; }
}

@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: 1fr; }
  .right-widgets { flex-direction: column; }
}
</style>
