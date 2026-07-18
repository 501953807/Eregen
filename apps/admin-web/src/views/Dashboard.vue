<template>
  <div class="dashboard">
    <!-- KPI Cards -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #4A90D9;">
              <el-icon :size="28"><Monitor /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.online_devices.toLocaleString() }}</div>
              <div class="kpi-label">在线设备</div>
              <div class="kpi-trend up">较昨日 +2.3%</div>
              <svg class="sparkline" viewBox="0 0 120 30">
                <polyline :points="sparkLinePoints(lineSparkData)" fill="none" stroke="#4A90D9" stroke-width="1.5"/>
                <circle v-for="(p, i) in lineSparkData" :key="i" :cx="sparkX(i, lineSparkData.length)" :cy="sparkY(p, lineSparkData)" r="2" fill="#4A90D9" opacity="0.6"/>
              </svg>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #67C23A;">
              <el-icon :size="28"><UserFilled /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.total_users.toLocaleString() }}</div>
              <div class="kpi-label">活跃家属</div>
              <div class="kpi-trend up">较昨日 +5.1%</div>
              <svg class="sparkline" viewBox="0 0 120 30">
                <polyline :points="sparkLinePoints(userSparkData)" fill="none" stroke="#67C23A" stroke-width="1.5"/>
                <circle v-for="(p, i) in userSparkData" :key="i" :cx="sparkX(i, userSparkData.length)" :cy="sparkY(p, userSparkData)" r="2" fill="#67C23A" opacity="0.6"/>
              </svg>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #E6A23C;">
              <el-icon :size="28"><Bell /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.active_alerts }}</div>
              <div class="kpi-label">待处理告警</div>
              <div class="kpi-trend down">较昨日 -12.5%</div>
              <svg class="sparkline" viewBox="0 0 120 30">
                <polyline :points="sparkLinePoints(alertSparkData)" fill="none" stroke="#E6A23C" stroke-width="1.5"/>
                <circle v-for="(p, i) in alertSparkData" :key="i" :cx="sparkX(i, alertSparkData.length)" :cy="sparkY(p, alertSparkData)" r="2" fill="#E6A23C" opacity="0.6"/>
              </svg>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #F56C6C;">
              <el-icon :size="28"><TrendCharts /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.total_devices ? Math.round((stats.online_devices / stats.total_devices) * 100) + '%' : '—' }}</div>
              <div class="kpi-label">设备在线率</div>
              <div class="kpi-trend up">较上周 +1.2%</div>
              <svg class="sparkline" viewBox="0 0 120 30">
                <polyline :points="sparkLinePoints(onlineRateSparkData)" fill="none" stroke="#F56C6C" stroke-width="1.5"/>
                <circle v-for="(p, i) in onlineRateSparkData" :key="i" :cx="sparkX(i, onlineRateSparkData.length)" :cy="sparkY(p, onlineRateSparkData)" r="2" fill="#F56C6C" opacity="0.6"/>
              </svg>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Device Type Donut Chart Row -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600;">设备类型分布</span></template>
          <div ref="donutChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600;">套餐订阅分布</span></template>
          <div ref="planChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600;">告警优先级分布</span></template>
          <div ref="alertPriorityChartRef" style="height: 260px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">设备在线趋势</span>
          </template>
          <div ref="lineChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">告警分布</span>
          </template>
          <div ref="pieChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Bottom Row -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header-with-action">
              <span style="font-weight: 600;">最新告警</span>
              <el-link type="primary" :underline="false">查看全部 →</el-link>
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
                <el-tag :type="alertTypeTag(row.alert_type)" size="small">{{ row.alert_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="设备" width="120">
              <template #default="{ row }">
                {{ row.metadata?.device_id || '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header-with-action">
              <span style="font-weight: 600;">用户增长</span>
              <el-link type="primary" :underline="false">详情 →</el-link>
            </div>
          </template>
          <div ref="barChartRef" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { Monitor, UserFilled, Bell, TrendCharts } from '@element-plus/icons-vue'
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

// Sparkline data for KPI cards — v2 prototype enhancement
const lineSparkData = [42, 45, 43, 48, 50, 47, 52]
const userSparkData = [120, 125, 132, 128, 135, 142, 150]
const alertSparkData = [28, 25, 30, 22, 26, 20, 18]
const onlineRateSparkData = [91, 92, 90, 93, 94, 93, 95]

function sparkX(i: number, total: number): number {
  return (i / (total - 1)) * 115 + 2.5
}
function sparkY(v: number, data: number[]): string {
  const min = Math.min(...data)
  const max = Math.max(...data)
  const range = max - min || 1
  return 28 - ((v - min) / range) * 24
}
function sparkLinePoints(data: number[]): string {
  const total = data.length
  return data.map((_, i) => `${sparkX(i, total)},${sparkY(data[i], data)}`).join(' ')
}

function formatTime(dateStr?: string): string {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function alertTypeTag(type: string): 'danger' | 'warning' | 'info' {
  const map: Record<string, 'danger' | 'warning' | 'info'> = {
    SOS: 'danger', fall: 'warning', heart: 'danger', geofence: 'info', medication: 'warning',
  }
  return map[type] || 'info'
}

function statusTag(status: string): 'danger' | 'warning' | 'success' {
  return status === 'pending' ? 'danger' : status === 'resolved' ? 'success' : 'warning'
}

function statusLabel(status: string): string {
  return status === 'pending' ? '未处理' : status === 'resolved' ? '已处理' : '处理中'
}

const alertTableData = ref<Array<Alert & { created_at: string }>>([])

watch(
  () => store.recentAlerts,
  (alerts) => {
    alertTableData.value = alerts.map(a => ({ ...a, created_at: a.created_at || '' }))
  },
  { immediate: true },
)

function renderLineChart() {
  if (!lineChartRef.value) return
  if (!lineChart) lineChart = echarts.init(lineChartRef.value)

  const trend = store.chartData.alertTrend
  const dates = trend.map(d => d.date)
  const bracelet = trend.map(d => d.bracelet_count)
  const pillbox = trend.map(d => d.pillbox_count)

  lineChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['手环', '药盒'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: dates.length ? dates : ['暂无数据'] },
    yAxis: { type: 'value' },
    series: [
      {
        name: '手环', type: 'line', smooth: true, data: bracelet.length ? bracelet : [0],
        itemStyle: { color: '#4A90D9' }, areaStyle: { opacity: 0.1 },
      },
      {
        name: '药盒', type: 'line', smooth: true, data: pillbox.length ? pillbox : [0],
        itemStyle: { color: '#67C23A' }, areaStyle: { opacity: 0.1 },
      },
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
            { value: 35, name: 'SOS', itemStyle: { color: '#F56C6C' } },
            { value: 28, name: '跌倒检测', itemStyle: { color: '#E6A23C' } },
            { value: 22, name: '心率异常', itemStyle: { color: '#4A90D9' } },
            { value: 15, name: '漏服药物', itemStyle: { color: '#67C23A' } },
          ],
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } },
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
          { offset: 0, color: '#4A90D9' },
          { offset: 1, color: '#357ABD' },
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

// Device type donut — v2 prototype enhancement
function renderDonutChart() {
  if (!donutChartRef.value) return
  if (!donutChart) donutChart = echarts.init(donutChartRef.value)
  donutChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '设备类型', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 480, name: '手环-入门版', itemStyle: { color: '#4A90D9' } },
        { value: 312, name: '手环-中端版', itemStyle: { color: '#6366F1' } },
        { value: 148, name: '手环-高端版', itemStyle: { color: '#8B5CF6' } },
        { value: 220, name: '药盒-智能版', itemStyle: { color: '#67C23A' } },
        { value: 85, name: '药盒-自动版', itemStyle: { color: '#E6A23C' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}' },
    }],
  })
}

// Plan distribution donut — v2 prototype enhancement
function renderPlanChart() {
  if (!planChartRef.value) return
  if (!planChart) planChart = echarts.init(planChartRef.value)
  planChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '套餐', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 189, name: 'Starter ¥29/月', itemStyle: { color: '#A78BFA' } },
        { value: 312, name: 'Plus ¥59/月', itemStyle: { color: '#409EFF' } },
        { value: 148, name: 'Pro ¥99/月', itemStyle: { color: '#6366F1' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{d}%' },
    }],
  })
}

// Alert priority donut — v2 prototype enhancement
function renderAlertPriorityChart() {
  if (!alertPriorityChartRef.value) return
  if (!alertPriorityChart) alertPriorityChart = echarts.init(alertPriorityChartRef.value)
  alertPriorityChart.setOption({
    tooltip: { trigger: 'item' },
    series: [{
      name: '告警优先级', type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { value: 12, name: 'P0 紧急', itemStyle: { color: '#F56C6C' } },
        { value: 38, name: 'P1 重要', itemStyle: { color: '#E6A23C' } },
        { value: 156, name: 'P2 一般', itemStyle: { color: '#909399' } },
      ],
      label: { fontSize: 11, formatter: '{b}\n{c}条' },
    }],
  })
}

// Resize handler
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
.kpi-card :deep(.el-card__body) { padding: 16px 20px; }
.kpi-content { display: flex; align-items: center; gap: 16px; }
.kpi-icon { width: 56px; height: 56px; border-radius: 12px; display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0; }
.kpi-info { flex: 1; }
.kpi-value { font-size: 28px; font-weight: 700; color: #303133; }
.kpi-label { font-size: 13px; color: #909399; margin-top: 2px; }
.kpi-trend { font-size: 12px; margin-top: 4px; }
.kpi-trend.up { color: #67C23A; }
.kpi-trend.down { color: #F56C6C; }
.sparkline { width: 120px; height: 30px; margin-top: 4px; }
.card-header-with-action { display: flex; justify-content: space-between; align-items: center; }
</style>
