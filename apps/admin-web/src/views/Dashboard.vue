<template>
  <div class="dashboard">
    <!-- Welcome Header -->
    <div class="welcome-header">
      <div class="welcome-text">
        <h1 class="welcome-title">
          <span class="greeting">{{ timeGreeting }}</span>，管理员
          <span class="wave">👋</span>
        </h1>
        <p class="welcome-sub">今日健康概览 · 颐贞康养中心管理平台</p>
      </div>
      <div class="welcome-meta">
        <div class="meta-date">{{ currentDate }}</div>
        <div class="meta-status">
          <span class="status-dot green"></span>
          系统正常运行
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-primary">
          <div class="kpi-content">
            <div class="kpi-icon-wrap green">
              <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ store.stats.online_devices.toLocaleString() }}</div>
              <div class="kpi-label">在线设备</div>
              <div class="kpi-trend up">较昨日 +2.3%</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-success">
          <div class="kpi-content">
            <div class="kpi-icon-wrap blue">
              <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ store.stats.total_users.toLocaleString() }}</div>
              <div class="kpi-label">活跃家属</div>
              <div class="kpi-trend up">较昨日 +5.1%</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-content">
            <div class="kpi-icon-wrap orange">
              <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ store.stats.active_alerts }}</div>
              <div class="kpi-label">待处理告警</div>
              <div class="kpi-trend down">较昨日 -12.5%</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-info">
          <div class="kpi-content">
            <div class="kpi-icon-wrap green">
              <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ store.stats.total_devices ? Math.round((store.stats.online_devices / store.stats.total_devices) * 100) + '%' : '—' }}</div>
              <div class="kpi-label">设备在线率</div>
              <div class="kpi-trend up">较上周 +1.2%</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span class="card-title">设备类型分布</span></template>
          <div ref="donutChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span class="card-title">套餐订阅分布</span></template>
          <div ref="planChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header><span class="card-title">告警优先级分布</span></template>
          <div ref="alertPriorityChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Main Charts Row -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <span class="card-title">设备在线趋势</span>
          </template>
          <div ref="lineChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <span class="card-title">告警分布</span>
          </template>
          <div ref="pieChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Bottom Row -->
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <div class="card-header-with-action">
              <span class="card-title">最新告警</span>
              <el-link type="primary" :underline="'never'" style="color: var(--color-primary);">查看全部 →</el-link>
            </div>
          </template>
          <el-table :data="alertTableData" stripe style="width: 100%">
            <el-table-column prop="created_at" label="时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="alert_type" label="类型" width="100">
              <template #default="{ row }">
                <span class="status-badge" :class="alertBadgeClass(row.alert_type)">
                  <span class="status-dot" :class="alertDotClass(row.alert_type)"></span>
                  {{ row.alert_type }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="设备" width="120">
              <template #default="{ row }">
                {{ row.metadata?.device_id || '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <span class="status-badge" :class="statusBadgeClass(row.status)">
                  <span class="status-dot" :class="statusDotClass(row.status)"></span>
                  {{ statusLabel(row.status) }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <div class="card-header-with-action">
              <span class="card-title">用户增长</span>
              <el-link type="primary" :underline="'never'" style="color: var(--color-primary);">详情 →</el-link>
            </div>
          </template>
          <div ref="barChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch, nextTick, computed } from 'vue'
import * as echarts from 'echarts'
import { useDashboardStore } from '@/stores/dashboard'
import type { Alert } from '@/types'

const store = useDashboardStore()
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

// Time-based greeting
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

const wellnessColors = ['#5C8D73', '#7BAF8C', '#A8C3B0', '#D9A441', '#D77B72', '#6E9FC4', '#6FAF8F']

const alertTableData = ref<Array<Alert & { created_at: string }>>([])

watch(
  () => store.recentAlerts,
  (alerts) => {
    alertTableData.value = alerts.map(a => ({ ...a, created_at: a.created_at || '' }))
  },
  { immediate: true },
)

function formatTime(dateStr?: string): string {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function alertBadgeClass(type: string): string {
  if (['SOS', 'heart'].includes(type)) return 'badge-danger'
  if (['fall', 'medication'].includes(type)) return 'badge-warning'
  return 'badge-primary'
}
function alertDotClass(type: string): string {
  if (['SOS', 'heart'].includes(type)) return 'dot-danger'
  if (['fall', 'medication'].includes(type)) return 'dot-warning'
  return 'dot-primary'
}
function statusBadgeClass(status: string): string {
  return status === 'pending' ? 'badge-danger' : status === 'resolved' ? 'badge-success' : 'badge-warning'
}
function statusDotClass(status: string): string {
  return status === 'pending' ? 'dot-danger' : status === 'resolved' ? 'dot-success' : 'dot-warning'
}
function statusLabel(status: string): string {
  return status === 'pending' ? '未处理' : status === 'resolved' ? '已处理' : '处理中'
}

function renderLineChart() {
  if (!lineChartRef.value) return
  if (!lineChart) lineChart = echarts.init(lineChartRef.value)
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
      { name: '药盒', type: 'line', smooth: true, data: pillbox.length ? pillbox : [0], itemStyle: { color: '#6E9FC4' }, areaStyle: { opacity: 0.08 } },
    ],
  })
}

function renderPieChart() {
  if (!pieChartRef.value) return
  if (!pieChart) pieChart = echarts.init(pieChartRef.value)
  const items = store.chartData.alertDistribution
  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '告警类型', type: 'pie', radius: '60%',
      data: items.length
        ? items.map(i => ({ value: i.value, name: i.name, itemStyle: { color: i.color } }))
        : [
            { value: 35, name: 'SOS', itemStyle: { color: '#D77B72' } },
            { value: 28, name: '跌倒检测', itemStyle: { color: '#D9A441' } },
            { value: 22, name: '心率异常', itemStyle: { color: '#5C8D73' } },
            { value: 15, name: '漏服药物', itemStyle: { color: '#6FAF8F' } },
          ],
    }],
  })
}

function renderBarChart() {
  if (!barChartRef.value) return
  if (!barChart) barChart = echarts.init(barChartRef.value)
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
  await store.refreshAll()
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
  if (!donutChart) donutChart = echarts.init(donutChartRef.value)
  donutChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '设备类型', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 480, name: '手环-入门版', itemStyle: { color: '#5C8D73' } },
        { value: 312, name: '手环-中端版', itemStyle: { color: '#7BAF8C' } },
        { value: 148, name: '手环-高端版', itemStyle: { color: '#A8C3B0' } },
        { value: 220, name: '药盒-智能版', itemStyle: { color: '#6E9FC4' } },
        { value: 85, name: '药盒-自动版', itemStyle: { color: '#D9A441' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}' },
    }],
  })
}

function renderPlanChart() {
  if (!planChartRef.value) return
  if (!planChart) planChart = echarts.init(planChartRef.value)
  planChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '套餐', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 189, name: 'Starter ¥29/月', itemStyle: { color: '#A8C3B0' } },
        { value: 312, name: 'Plus ¥59/月', itemStyle: { color: '#5C8D73' } },
        { value: 148, name: 'Pro ¥99/月', itemStyle: { color: '#6E9FC4' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{d}%' },
    }],
  })
}

function renderAlertPriorityChart() {
  if (!alertPriorityChartRef.value) return
  if (!alertPriorityChart) alertPriorityChart = echarts.init(alertPriorityChartRef.value)
  alertPriorityChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '告警优先级', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 12, name: 'P0 紧急', itemStyle: { color: '#D77B72' } },
        { value: 38, name: 'P1 重要', itemStyle: { color: '#D9A441' } },
        { value: 156, name: 'P2 一般', itemStyle: { color: '#A8C3B0' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}条' },
    }],
  })
}

onMounted(() => {
  initCharts()
})

function handleResize() {
  lineChart?.resize()
  pieChart?.resize()
  barChart?.resize()
  donutChart?.resize()
  planChart?.resize()
  alertPriorityChart?.resize()
}

window.addEventListener('resize', handleResize)
</script>

<style scoped>
.dashboard {
  padding: 0;
}

/* Welcome Header */
.welcome-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
}
.welcome-title {
  font-size: 24px;
  font-weight: 700;
  color: #29404A;
  margin: 0 0 4px;
  letter-spacing: -0.01em;
}
.wave { font-size: 22px; }
.welcome-sub {
  font-size: 13px;
  color: #6B8980;
  margin: 0;
}
.welcome-meta {
  text-align: right;
}
.meta-date {
  font-size: 14px;
  font-weight: 600;
  color: #4A6260;
  margin-bottom: 4px;
}
.meta-status {
  font-size: 12px;
  color: #6B8980;
  display: flex;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
}
.status-dot.green {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #5C8D73;
  box-shadow: 0 0 6px rgba(92,141,115,0.4);
}

/* KPI Cards */
.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all var(--duration-normal) var(--easing);
}

.kpi-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(ellipse at top left, rgba(255,255,255,0.6) 0%, transparent 60%);
  pointer-events: none;
}

.kpi-card:hover {
  transform: translateY(-3px);
}

.kpi-card :deep(.el-card__body) {
  padding: 18px 20px;
  border: 1px solid var(--border-light);
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), var(--shadow-card);
  border-radius: var(--radius-lg);
}
.kpi-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.kpi-icon-wrap.blue { background: linear-gradient(135deg, #5C8D73, #7BAF8C); color: #fff; }
.kpi-icon-wrap.green { background: linear-gradient(135deg, #6FAF8F, #8BC4A8); color: #fff; }
.kpi-icon-wrap.orange { background: linear-gradient(135deg, #D9A441, #E8BC6A); color: #fff; }
.kpi-icon-wrap.red { background: linear-gradient(135deg, #D77B72, #E09890); color: #fff; }
.kpi-icon-wrap.purple { background: linear-gradient(135deg, #7C3AED, #A78BFA); color: #fff; }
.kpi-icon { display: none; }
.kpi-info { flex: 1; }
.kpi-value {
  font-size: 32px;
  font-weight: 800;
  color: #29404A;
  letter-spacing: -0.03em;
  line-height: 1;
  margin-bottom: 4px;
}
.kpi-label {
  font-size: 13px;
  color: #6B8980;
  margin-top: 2px;
}
.kpi-trend {
  font-size: 12px;
  margin-top: 4px;
}
.kpi-trend.up { color: #4A8A6A; }
.kpi-trend.down { color: #D77B72; }

/* Card titles */
.card-title {
  font-size: 14px;
  font-weight: 600;
  color: #29404A;
}
.card-header-with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* Status badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}
.badge-success { background: #E8F4EC; color: #4A8A6A; }
.badge-danger { background: #FDF0EE; color: #B85C54; }
.badge-warning { background: #FEF7E8; color: #B8860B; }
.badge-primary { background: #DDEBE1; color: #47745C; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: #6FAF8F; }
.dot-danger { background: #D77B72; }
.dot-warning { background: #D9A441; }
.dot-primary { background: #5C8D73; }
</style>
