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
              <div class="kpi-value">1,247</div>
              <div class="kpi-label">在线设备</div>
              <div class="kpi-trend up">↑ 12.5% 较昨日</div>
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
              <div class="kpi-value">856</div>
              <div class="kpi-label">活跃家属</div>
              <div class="kpi-trend up">↑ 8.3% 较昨日</div>
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
              <div class="kpi-value">19</div>
              <div class="kpi-label">待处理告警</div>
              <div class="kpi-trend down">↓ 5.2% 较昨日</div>
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
              <div class="kpi-value">94.2%</div>
              <div class="kpi-label">设备在线率</div>
              <div class="kpi-trend up">↑ 0.8% 较上周</div>
            </div>
          </div>
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
          <el-table :data="recentAlerts" stripe style="width: 100%">
            <el-table-column prop="time" label="时间" width="160" />
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.typeTag" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="device" label="设备" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.statusTag" size="small">{{ row.status }}</el-tag>
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
import { onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import { Monitor, UserFilled, Bell, TrendCharts } from '@element-plus/icons-vue'

const lineChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()

interface AlertItem {
  time: string
  type: string
  typeTag: 'danger' | 'warning' | 'info'
  device: string
  status: string
  statusTag: 'danger' | 'warning' | 'success'
}

const recentAlerts: AlertItem[] = [
  { time: '2026-07-16 14:32:01', type: 'SOS', typeTag: 'danger', device: 'BR-0042', status: '未处理', statusTag: 'danger' },
  { time: '2026-07-16 13:18:45', type: '跌倒', typeTag: 'warning', device: 'BR-0017', status: '处理中', statusTag: 'warning' },
  { time: '2026-07-16 12:05:22', type: '心率异常', typeTag: 'danger', device: 'BR-0089', status: '已处理', statusTag: 'success' },
  { time: '2026-07-16 11:42:10', type: '电子围栏', typeTag: 'info', device: 'BR-0033', status: '未处理', statusTag: 'danger' },
  { time: '2026-07-16 10:15:33', type: '漏服药物', typeTag: 'warning', device: 'PX-0012', status: '已处理', statusTag: 'success' },
]

onMounted(() => {
  // Line chart - device online trend
  if (lineChartRef.value) {
    const chart = echarts.init(lineChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['手环', '药盒'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['7/10', '7/11', '7/12', '7/13', '7/14', '7/15', '7/16'],
      },
      yAxis: { type: 'value' },
      series: [
        {
          name: '手环', type: 'line', smooth: true, data: [980, 1020, 990, 1050, 1100, 1150, 1180],
          itemStyle: { color: '#4A90D9' }, areaStyle: { opacity: 0.1 },
        },
        {
          name: '药盒', type: 'line', smooth: true, data: [180, 195, 200, 210, 215, 220, 230],
          itemStyle: { color: '#67C23A' }, areaStyle: { opacity: 0.1 },
        },
      ],
    })
  }

  // Pie chart - alert distribution
  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'item' },
      legend: { orient: 'vertical', left: 'left' },
      series: [{
        name: '告警类型', type: 'pie', radius: '60%',
        data: [
          { value: 35, name: 'SOS', itemStyle: { color: '#F56C6C' } },
          { value: 28, name: '跌倒检测', itemStyle: { color: '#E6A23C' } },
          { value: 22, name: '心率异常', itemStyle: { color: '#4A90D9' } },
          { value: 15, name: '漏服药物', itemStyle: { color: '#67C23A' } },
        ],
        emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' } },
      }],
    })
  }

  // Bar chart - user growth
  if (barChartRef.value) {
    const chart = echarts.init(barChartRef.value)
    chart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: ['2月', '3月', '4月', '5月', '6月', '7月'],
      },
      yAxis: { type: 'value' },
      series: [{
        name: '新增用户', type: 'bar', barWidth: '40%',
        data: [120, 180, 250, 320, 410, 520],
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#4A90D9' },
            { offset: 1, color: '#357ABD' },
          ]),
        },
      }],
    })
  }
})
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
.card-header-with-action { display: flex; justify-content: space-between; align-items: center; }
</style>
