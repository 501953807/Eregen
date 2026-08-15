<template>
  <div class="analytics-page">
    <!-- Page Header -->
    <div class="hope-page-header">
      <div>
        <h1 class="hope-page-header__title">数据分析</h1>
        <p class="hope-page-header__subtitle">设备在线率 · 告警趋势 · 用户增长 · 机构活跃度</p>
      </div>
    </div>

    <!-- Key Metrics — HopeStatCard -->
    <div class="metrics-grid">
      <HopeStatCard :value="keyMetrics[0].value" label="在线设备总数" icon-color="primary" gradient="linear-gradient(135deg, #3a57e8, #6f42c1)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+12.5% 较上周</span>
        </template>
      </HopeStatCard>
      <HopeStatCard :value="keyMetrics[1].value" label="活跃用户数" icon-color="success" gradient="linear-gradient(135deg, #22c55e, #16a34a)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+8.3% 较上周</span>
        </template>
      </HopeStatCard>
      <HopeStatCard :value="keyMetrics[2].value" label="今日告警数" icon-color="warning" gradient="linear-gradient(135deg, #f59e0b, #d97706)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-down">-3.1% 较上周</span>
        </template>
      </HopeStatCard>
      <HopeStatCard :value="keyMetrics[3].value" label="机构接入数" icon-color="accent" gradient="linear-gradient(135deg, #8C57FF, #6f42c1)">
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg></el-icon>
        </template>
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">+2.0% 较上周</span>
        </template>
      </HopeStatCard>
    </div>

    <!-- Charts Row 1 -->
    <div class="charts-row">
      <HopeCard title="设备在线率趋势">
        <div ref="deviceChartRef" class="chart-container"></div>
        <div class="chart-footer">日均在线率 <strong>95.9%</strong> ｜ 峰值 97.1% ｜ 最低 94.2%</div>
      </HopeCard>
      <HopeCard title="告警类型分布">
        <div ref="alertChartRef" class="chart-container"></div>
        <div class="chart-footer">本周告警总计 <strong>{{ alertTotal }}</strong> 起</div>
      </HopeCard>
    </div>

    <!-- Charts Row 2 -->
    <div class="charts-row">
      <HopeCard title="用药依从性趋势">
        <div ref="medChartRef" class="chart-container"></div>
        <div class="chart-footer">周平均依从率 <strong>90.7%</strong></div>
      </HopeCard>
      <HopeCard title="用户增长趋势">
        <div ref="growthChartRef" class="chart-container"></div>
        <div class="chart-footer">用户累计 <strong>{{ totalGrowth }}</strong></div>
      </HopeCard>
    </div>

    <!-- Institution Activity Table -->
    <HopeCard title="机构活跃度排行">
      <template #header>
        <div class="period-selector">
          <HopeBtn
            v-for="p in periodOptions"
            :key="p.value"
            :variant="institutionPeriod === p.value ? 'filled' : 'plain'"
            size="sm"
            @click="institutionPeriod = p.value; loadDashboard()"
          >{{ p.label }}</HopeBtn>
        </div>
      </template>
      <HopeTable
        :columns="institutionColumns"
        :data="institutionList"
        :striped="true"
      >
        <template #col-name="{ row }">
          <div class="inst-name">{{ row.name }}</div>
        </template>
        <template #col-type="{ row }">
          <HopeBadge :color="institutionTypeBadgeColor(row.type)">{{ row.typeLabel }}</HopeBadge>
        </template>
        <template #col-elderlyCount="{ row }">
          <span class="num-right">{{ row.elderlyCount.toLocaleString() }}</span>
        </template>
        <template #col-dataIngested="{ row }">
          <span class="num-right">{{ formatNumber(row.dataIngested) }}</span>
        </template>
        <template #col-activityScore="{ row }">
          <div class="activity-bar-wrap">
            <div class="activity-bar-bg">
              <div class="activity-bar-fill" :style="{ width: row.activityScore + '%', background: row.activityColor }"></div>
            </div>
            <span class="activity-pct">{{ row.activityScore }}%</span>
          </div>
        </template>
      </HopeTable>
    </HopeCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { dashboardApi } from '@/api/dashboard'
import type { AlertDistributionItem, UserGrowthPoint } from '@/api/dashboard'
import { HopeStatCard, HopeCard, HopeTable, HopeBtn, HopeBadge } from '@/components/hope'
import * as echarts from 'echarts'

// ── Key Metrics ────────────────────────────────────────────
const keyMetrics = ref([
  { label: '在线设备总数', value: '0', trend: '+12.5%', trendUp: true },
  { label: '活跃用户数', value: '0', trend: '+8.3%', trendUp: true },
  { label: '今日告警数', value: '0', trend: '-3.1%', trendUp: false },
  { label: '机构接入数', value: '0', trend: '+2.0%', trendUp: true },
])

// ── Institution Data ───────────────────────────────────────
const institutionPeriod = ref('7')
const periodOptions = [
  { label: '近7天', value: '7' },
  { label: '近30天', value: '30' },
  { label: '近90天', value: '90' },
]
const institutionList = ref([
  { name: '上海市第一中心医院', type: 'hospital', typeLabel: '三甲医院', elderlyCount: 0, dataIngested: 0, lastActive: '—', activityScore: 0, activityColor: '#22c55e' },
  { name: '浦东新区社区服务中心', type: 'community', typeLabel: '社区', elderlyCount: 0, dataIngested: 0, lastActive: '—', activityScore: 0, activityColor: '#3a57e8' },
])
const institutionColumns = [
  { prop: 'name', label: '机构名称', sortable: false },
  { prop: 'type', label: '类型', sortable: false },
  { prop: 'elderlyCount', label: '关联老人', sortable: false },
  { prop: 'dataIngested', label: '数据接入量', sortable: false },
  { prop: 'lastActive', label: '最后活跃', sortable: false },
  { prop: 'activityScore', label: '活跃度', sortable: false },
]

function institutionTypeBadgeColor(type: string): 'primary' | 'success' | 'warning' | 'info' {
  if (type === 'hospital') return 'primary'
  if (type === 'community') return 'success'
  if (type === 'station') return 'warning'
  return 'info'
}

function formatNumber(n: number): string {
  return n >= 10000 ? `${(n / 10000).toFixed(1)}万` : n.toLocaleString()
}

// ── Chart Data ─────────────────────────────────────────────
const alertDistribution = ref<AlertDistributionItem[]>([])
const userGrowth = ref<UserGrowthPoint[]>([])
const alertTotal = ref(442)
const totalGrowth = ref(25900)

// ── Chart refs ─────────────────────────────────────────────
const deviceChartRef = ref<HTMLElement>()
const alertChartRef = ref<HTMLElement>()
const medChartRef = ref<HTMLElement>()
const growthChartRef = ref<HTMLElement>()

let deviceChart: echarts.ECharts | null = null
let alertChart: echarts.ECharts | null = null
let medChart: echarts.ECharts | null = null
let growthChart: echarts.ECharts | null = null

// ── Hope UI ECharts theme ──────────────────────────────────
const hopeUIEChartsTheme: echarts.EChartsType = null as any

// ── Render functions ───────────────────────────────────────

function renderDeviceChart() {
  if (!deviceChartRef.value) return
  if (!deviceChart) deviceChart = echarts.init(deviceChartRef.value)
  deviceChart.setOption({
    tooltip: { trigger: 'axis' },
    color: ['#3a57e8'],
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'], axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, axisLabel: { textStyle: { color: '#8a8d93' } } },
    yAxis: { type: 'value', min: 90, max: 100, axisLabel: { formatter: '{value}%', textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    series: [{
      name: '在线率', type: 'bar', barWidth: '45%', data: [96.2, 95.8, 97.1, 96.5, 95.9, 94.2, 95.5],
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3a57e8' }, { offset: 1, color: '#6f42c1' },
        ]),
        borderRadius: [6, 6, 0, 0],
      },
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Noto Sans SC", sans-serif' },
    backgroundColor: 'transparent',
  })
}

function renderAlertChart() {
  if (!alertChartRef.value) return
  if (!alertChart) alertChart = echarts.init(alertChartRef.value)
  const items = alertDistribution.value.length > 0
    ? alertDistribution.value.map(d => ({ name: d.name, value: d.value }))
    : [
        { name: 'SOS紧急呼叫', value: 45 }, { name: '跌倒检测', value: 32 },
        { name: '心率异常', value: 78 }, { name: '电子围栏', value: 28 },
        { name: '漏服药物', value: 156 }, { name: '设备离线', value: 3 },
      ]
  alertTotal.value = items.reduce((s, i) => s + i.value, 0)
  alertChart.setOption({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '8%', bottom: '3%', containLabel: true },
    xAxis: { type: 'value', axisLabel: { textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    yAxis: { type: 'category', data: items.map(i => i.name).reverse(), axisLabel: { textStyle: { color: '#616161' } } },
    series: [{
      type: 'bar', data: items.map(i => i.value).reverse(), barWidth: '50%',
      itemStyle: { borderRadius: [0, 6, 6, 0], color: (params: any) => {
        const v = items[params.dataIndex]?.value ?? 0
        return v > 100 ? '#22c55e' : v > 30 ? '#3a57e8' : '#f59e0b'
      }},
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Noto Sans SC", sans-serif' },
    backgroundColor: 'transparent',
  })
}

function renderMedChart() {
  if (!medChartRef.value) return
  if (!medChart) medChart = echarts.init(medChartRef.value)
  medChart.setOption({
    tooltip: { trigger: 'axis' },
    color: ['#22c55e', '#f59e0b'],
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'], axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, axisLabel: { textStyle: { color: '#8a8d93' } } },
    yAxis: { type: 'value', min: 80, max: 100, axisLabel: { formatter: '{value}%', textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    series: [{
      name: '依从率', type: 'bar', barWidth: '45%',
      data: [
        { value: 92.3, itemStyle: { color: '#22c55e', borderRadius: [6, 6, 0, 0] } },
        { value: 89.5, itemStyle: { color: '#f59e0b', borderRadius: [6, 6, 0, 0] } },
        { value: 91.8, itemStyle: { color: '#22c55e', borderRadius: [6, 6, 0, 0] } },
        { value: 93.2, itemStyle: { color: '#22c55e', borderRadius: [6, 6, 0, 0] } },
        { value: 90.1, itemStyle: { color: '#f59e0b', borderRadius: [6, 6, 0, 0] } },
        { value: 87.6, itemStyle: { color: '#f59e0b', borderRadius: [6, 6, 0, 0] } },
        { value: 91.0, itemStyle: { color: '#22c55e', borderRadius: [6, 6, 0, 0] } },
      ],
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Noto Sans SC", sans-serif' },
    backgroundColor: 'transparent',
  })
}

function renderGrowthChart() {
  if (!growthChartRef.value) return
  if (!growthChart) growthChart = echarts.init(growthChartRef.value)
  const months = userGrowth.value.length > 0
    ? userGrowth.value.map(p => p.month)
    : ['2月', '3月', '4月', '5月', '6月', '7月']
  const values = userGrowth.value.length > 0
    ? userGrowth.value.map(p => p.new_users)
    : [1200, 2100, 3400, 4800, 6200, 8234]
  totalGrowth.value = values.reduce((s, v) => s + v, 0)
  growthChart.setOption({
    tooltip: { trigger: 'axis' },
    color: ['#3a57e8'],
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: months, axisLine: { lineStyle: { color: 'rgba(26,46,38,0.10)' } }, axisLabel: { textStyle: { color: '#8a8d93' } } },
    yAxis: { type: 'value', axisLabel: { textStyle: { color: '#8a8d93' } }, splitLine: { lineStyle: { color: 'rgba(26,46,38,0.06)' } } },
    series: [{
      name: '新增用户', type: 'bar', barWidth: '40%',
      data: values,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#3a57e8' }, { offset: 1, color: '#8C57FF' },
        ]),
        borderRadius: [6, 6, 0, 0],
      },
    }],
    textStyle: { fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Noto Sans SC", sans-serif' },
    backgroundColor: 'transparent',
  })
}

// ── Data loading ───────────────────────────────────────────
async function loadDashboard() {
  try {
    const [distRes, growthRes, overviewRes] = await Promise.all([
      dashboardApi.alertDistribution(),
      dashboardApi.userGrowth({ months: 6 }),
      dashboardApi.overview(),
    ])
    const dist = distRes.data?.data || []
    const growth = growthRes.data?.data || []
    const overview = overviewRes.data?.data || {}
    alertDistribution.value = dist
    userGrowth.value = growth
    keyMetrics.value[0].value = String(overview.total_devices ?? 0)
    keyMetrics.value[1].value = String(overview.total_users ?? 0)
    keyMetrics.value[2].value = String(overview.active_alerts ?? 0)
    keyMetrics.value[3].value = String(overview.total_institutions ?? 0)
  } catch {
    // Keep defaults
  }
}

// ── Lifecycle ──────────────────────────────────────────────
function handleResize() {
  deviceChart?.resize()
  alertChart?.resize()
  medChart?.resize()
  growthChart?.resize()
}

onMounted(async () => {
  await loadDashboard()
  await nextTick()
  setTimeout(async () => {
    await nextTick()
    renderDeviceChart()
    renderAlertChart()
    renderMedChart()
    renderGrowthChart()
    window.addEventListener('resize', handleResize)
  }, 100)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  deviceChart?.dispose()
  alertChart?.dispose()
  medChart?.dispose()
  growthChart?.dispose()
})
</script>

<style scoped>
.analytics-page { padding: 0; }

/* Metrics grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

/* Charts rows */
.charts-row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
  grid-template-columns: repeat(2, 1fr);
}

.chart-container { height: 260px; }

.chart-footer {
  text-align: center;
  margin-top: 8px;
  font-size: 12px;
  color: var(--hope-text-muted);
}
.chart-footer strong { color: var(--hope-text); }

/* Period selector */
.period-selector { display: flex; gap: 6px; }

/* Number right-align */
.num-right { text-align: right; font-variant-numeric: tabular-nums; }

/* Institution name */
.inst-name { font-weight: 500; color: var(--hope-text); }

/* Activity bar */
.activity-bar-wrap { display: flex; align-items: center; gap: 8px; }
.activity-bar-bg {
  flex: 1; height: 8px; background: var(--hope-surface-light);
  border-radius: 999px; overflow: hidden;
}
.activity-bar-fill {
  height: 100%; border-radius: 999px;
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
.activity-pct {
  font-size: 12px; color: var(--hope-text-muted);
  min-width: 32px; text-align: right; font-variant-numeric: tabular-nums;
}

/* Responsive */
@media (max-width: 1200px) {
  .metrics-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .metrics-grid { grid-template-columns: 1fr; }
  .charts-row { grid-template-columns: 1fr; }
}
</style>
