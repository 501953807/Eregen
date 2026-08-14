<template>
  <div class="dashboard">
    <!-- Welcome Hero Banner -->
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

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <div class="hope-stat-card">
        <div class="hope-stat-card__icon hope-stat-card__icon--primary">
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></el-icon>
        </div>
        <div class="hope-stat-card__value">{{ store.stats.online_devices.toLocaleString() }}</div>
        <div class="hope-stat-card__label">在线设备</div>
        <div class="hope-stat-card__trend hope-stat-card__trend-up">+2.3% 较昨日</div>
      </div>
      <div class="hope-stat-card">
        <div class="hope-stat-card__icon hope-stat-card__icon--success">
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon>
        </div>
        <div class="hope-stat-card__value">{{ store.stats.total_users.toLocaleString() }}</div>
        <div class="hope-stat-card__label">活跃家属</div>
        <div class="hope-stat-card__trend hope-stat-card__trend-up">+5.1% 较昨日</div>
      </div>
      <div class="hope-stat-card">
        <div class="hope-stat-card__icon hope-stat-card__icon--warning">
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
        </div>
        <div class="hope-stat-card__value">{{ store.stats.active_alerts }}</div>
        <div class="hope-stat-card__label">待处理告警</div>
        <div class="hope-stat-card__trend hope-stat-card__trend-down">-12.5% 较昨日</div>
      </div>
      <div class="hope-stat-card">
        <div class="hope-stat-card__icon hope-stat-card__icon--info">
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/></svg></el-icon>
        </div>
        <div class="hope-stat-card__value">{{ store.stats.total_devices ? Math.round((store.stats.online_devices / store.stats.total_devices) * 100) + '%' : '—' }}</div>
        <div class="hope-stat-card__label">设备在线率</div>
        <div class="hope-stat-card__trend hope-stat-card__trend-up">+1.2% 较上周</div>
      </div>
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
import { eregenGreenEChartsTheme } from '@/utils/echarts-theme'
import type { Alert } from '@/types'
import { HopeCard, HopeBtn, HopeTable, HopeBadge } from '@/components/hope'

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
  { prop: 'device_id', label: '设备', sortable: false },
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

// (slots used via template rendering)

function renderLineChart() {
  if (!lineChartRef.value) return
  if (!lineChart) lineChart = echarts.init(lineChartRef.value, eregenGreenEChartsTheme)
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
      { name: '手环', type: 'line', smooth: true, data: bracelet.length ? bracelet : [0], itemStyle: { color: '#5C8D73' }, areaStyle: { opacity: 0.08 } },
      { name: '药盒', type: 'line', smooth: true, data: pillbox.length ? pillbox : [0], itemStyle: { color: '#6FAF8F' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

function renderPieChart() {
  if (!pieChartRef.value) return
  if (!pieChart) pieChart = echarts.init(pieChartRef.value, eregenGreenEChartsTheme)
  const items = store.chartData.alertDistribution
  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '告警类型', type: 'pie', radius: '60%',
      data: items.length
        ? items.map(i => ({ value: i.value, name: i.name, itemStyle: { color: i.color } }))
        : [
            { value: 35, name: 'SOS', itemStyle: { color: '#C04A42' } },
            { value: 28, name: '跌倒检测', itemStyle: { color: '#D9A441' } },
            { value: 22, name: '心率异常', itemStyle: { color: '#5C8D73' } },
            { value: 15, name: '漏服药物', itemStyle: { color: '#6FAF8F' } },
          ],
    }],
  })
}

function renderBarChart() {
  if (!barChartRef.value) return
  if (!barChart) barChart = echarts.init(barChartRef.value, eregenGreenEChartsTheme)
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
          { offset: 0, color: '#5C8D73' },
          { offset: 1, color: '#A8C3B0' },
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
  if (!donutChart) donutChart = echarts.init(donutChartRef.value, eregenGreenEChartsTheme)
  donutChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '设备类型', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 480, name: '手环-入门版', itemStyle: { color: '#5C8D73' } },
        { value: 312, name: '手环-中端版', itemStyle: { color: '#7BAF8C' } },
        { value: 148, name: '手环-高端版', itemStyle: { color: '#A8C3B0' } },
        { value: 220, name: '药盒-智能版', itemStyle: { color: '#6FAF8F' } },
        { value: 85, name: '药盒-自动版', itemStyle: { color: '#D9A441' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}' },
    }],
  })
}

function renderPlanChart() {
  if (!planChartRef.value) return
  if (!planChart) planChart = echarts.init(planChartRef.value, eregenGreenEChartsTheme)
  planChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '套餐', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 189, name: 'Starter ¥29/月', itemStyle: { color: '#A8C3B0' } },
        { value: 312, name: 'Plus ¥59/月', itemStyle: { color: '#5C8D73' } },
        { value: 148, name: 'Pro ¥99/月', itemStyle: { color: '#6FAF8F' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{d}%' },
    }],
  })
}

function renderAlertPriorityChart() {
  if (!alertPriorityChartRef.value) return
  if (!alertPriorityChart) alertPriorityChart = echarts.init(alertPriorityChartRef.value, eregenGreenEChartsTheme)
  alertPriorityChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '告警优先级', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 12, name: 'P0 紧急', itemStyle: { color: '#C04A42' } },
        { value: 38, name: 'P1 重要', itemStyle: { color: '#D9A441' } },
        { value: 156, name: 'P2 一般', itemStyle: { color: '#A8C3B0' } },
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

/* Welcome Hero */
.welcome-hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
  padding: 24px 28px;
  background: linear-gradient(135deg, #4A7C5F 0%, #6FAF8F 60%, #8FB89A 100%);
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
.kpi-grid .hope-stat-card { cursor: default; }

/* Charts Row */
.charts-row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}
.charts-row:nth-of-type(1),
.charts-row:nth-of-type(2) { grid-template-columns: repeat(3, 1fr); }
.charts-row:nth-of-type(3) { grid-template-columns: 2fr 1fr; }
.charts-row:nth-of-type(4) { grid-template-columns: 1fr 1fr; }

.chart-card-wide { grid-column: span 2; }
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
  .charts-row:nth-of-type(2),
  .charts-row:nth-of-type(4) { grid-template-columns: 1fr; }
  .charts-row:nth-of-type(3) { grid-template-columns: 1fr; }
  .chart-card-wide { grid-column: span 1; }
}

@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: 1fr; }
  .welcome-hero { flex-direction: column; align-items: flex-start; gap: 12px; padding: 16px 20px; }
  .welcome-hero__right { text-align: left; }
}
</style>
