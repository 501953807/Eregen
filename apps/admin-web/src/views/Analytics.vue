<template>
  <div class="analytics-page">
    <el-row :gutter="12">
      <!-- Key Metrics -->
      <el-col :span="6" v-for="metric in keyMetrics" :key="metric.label">
        <el-card shadow="hover" class="metric-card" :class="'kpi-' + metric.colorClass">
          <div class="metric-header">
            <span class="metric-label">{{ metric.label }}</span>
            <el-icon :size="24" :color="metric.iconColor"><component :is="metric.icon" /></el-icon>
          </div>
          <div class="metric-value">{{ metric.value }}</div>
          <div class="metric-trend" :style="{ color: metric.trendUp ? '#16A34A' : '#EF4444' }">
            {{ metric.trendUp ? '↑' : '↓' }} {{ metric.trend }} 较上周
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" style="margin-top: 16px;">
      <!-- Device Online Rate Chart -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">设备在线率趋势</span>
          </template>
          <div id="device-chart-area"></div>
        </el-card>
      </el-col>

      <!-- Health Alert Distribution -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">告警类型分布</span>
          </template>
          <div id="alert-chart-area"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" style="margin-top: 16px;">
      <!-- Medication Adherence Trend -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">用药依从性趋势</span>
          </template>
          <div id="medication-chart-area"></div>
        </el-card>
      </el-col>

      <!-- User Growth -->
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600;">用户增长趋势</span>
          </template>
          <div id="user-growth-chart-area"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" style="margin-top: 16px;">
      <!-- Institution Activity Table -->
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span style="font-weight: 600;">机构活跃度排行</span>
              <el-select v-model="institutionPeriod" placeholder="选择时间段" size="default" style="width: 140px;">
                <el-option label="近7天" value="7" />
                <el-option label="近30天" value="30" />
                <el-option label="近90天" value="90" />
              </el-select>
            </div>
          </template>
          <el-table :data="institutionList" stripe style="width: 100%">
            <el-table-column type="index" label="排名" width="60" align="center">
              <template #default="{ $index }">
                <el-badge :value="$index + 1" :max="9"
                  :type="$index === 0 ? 'danger' : ($index === 1 ? 'warning' : 'info')" />
              </template>
            </el-table-column>
            <el-table-column prop="name" label="机构名称" min-width="180" />
            <el-table-column prop="type" label="类型" width="120" align="center">
              <template #default="{ row }">
                <span class="status-badge" :class="institutionTypeBadge(row.type)">
                  <span class="status-dot" :class="institutionTypeDot(row.type)"></span>
                  {{ row.typeLabel }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="elderlyCount" label="关联老人" width="100" align="right" />
            <el-table-column prop="dataIngested" label="数据接入量" width="120" align="right">
              <template #default="{ row }">{{ formatNumber(row.dataIngested) }}</template>
            </el-table-column>
            <el-table-column prop="lastActive" label="最后活跃" width="140" />
            <el-table-column label="活跃度" width="140">
              <template #default="{ row }">
                <el-progress :percentage="row.activityScore" :color="row.activityColor" :stroke-width="8"
                  :show-text="false" style="width: 100px;" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import {
  Monitor, UserFilled, TrendCharts, BellFilled,
  Connection, VideoCamera, DataLine, Calendar
} from '@element-plus/icons-vue'

// Key metrics
const keyMetrics = ref([
  { label: '在线设备总数', value: '12,847', trend: '12.5%', trendUp: true, icon: Connection, colorClass: 'blue', iconColor: '#2563EB' },
  { label: '活跃用户数', value: '8,234', trend: '8.3%', trendUp: true, icon: UserFilled, colorClass: 'green', iconColor: '#16A34A' },
  { label: '今日告警数', value: '342', trend: '5.1%', trendUp: false, icon: BellFilled, colorClass: 'warning', iconColor: '#F59E0B' },
  { label: '机构接入数', value: '128', trend: '22.0%', trendUp: true, icon: Monitor, colorClass: 'purple', iconColor: '#7C3AED' },
])

// Institution data
const institutionPeriod = ref('7')
const institutionList = ref([
  { name: '上海市第一中心医院', type: 'hospital', typeLabel: '三甲医院', elderlyCount: 1250, dataIngested: 45200, lastActive: '2分钟前', activityScore: 95, activityColor: '#16A34A' },
  { name: '浦东新区社区服务中心', type: 'community', typeLabel: '社区', elderlyCount: 890, dataIngested: 28300, lastActive: '15分钟前', activityScore: 82, activityColor: '#2563EB' },
  { name: '北京协和医院', type: 'hospital', typeLabel: '三甲医院', elderlyCount: 780, dataIngested: 24100, lastActive: '1小时前', activityScore: 76, activityColor: '#2563EB' },
  { name: '朝阳区养老服务站', type: 'station', typeLabel: '服务站', elderlyCount: 420, dataIngested: 12600, lastActive: '3小时前', activityScore: 65, activityColor: '#F59E0B' },
  { name: '广州医科大学附属第一医院', type: 'hospital', typeLabel: '三甲医院', elderlyCount: 360, dataIngested: 10800, lastActive: '5小时前', activityScore: 58, activityColor: '#F59E0B' },
  { name: '深圳市南山区养老院', type: 'nursing', typeLabel: '养老院', elderlyCount: 280, dataIngested: 8400, lastActive: '1天前', activityScore: 45, activityColor: '#EF4444' },
])

function formatNumber(n: number): string {
  return n >= 10000 ? `${(n / 10000).toFixed(1)}万` : n.toLocaleString()
}

function institutionTypeBadge(type: string): string {
  if (type === 'hospital') return 'badge-primary'
  if (type === 'community') return 'badge-success'
  if (type === 'station') return 'badge-warning'
  return 'badge-gray'
}
function institutionTypeDot(type: string): string {
  if (type === 'hospital') return 'dot-primary'
  if (type === 'community') return 'dot-success'
  if (type === 'station') return 'dot-warning'
  return 'dot-gray'
}

onMounted(async () => {
  await nextTick()
  renderDeviceChart()
  renderAlertChart()
  renderMedicationChart()
  renderUserGrowthChart()
})

function renderDeviceChart() {
  const el = document.querySelector('#device-chart-area')
  if (!el) return
  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const online = [96.2, 95.8, 97.1, 96.5, 95.9, 94.2, 95.5]
  const total = [12500, 12600, 12700, 12750, 12800, 12820, 12847]

  let html = '<div style="display:flex;justify-content:space-between;align-items:flex-end;height:260px;padding:10px 0;">'
  const maxOnline = Math.max(...online)
  const minOnline = Math.min(...online) - 2
  days.forEach((day, i) => {
    const barHeight = ((online[i] - minOnline) / (maxOnline - minOnline)) * 200
    html += `<div style="flex:1;text-align:center;">
      <div style="margin-bottom:4px;font-size:12px;color:#16A34A;font-weight:600;">${online[i]}%</div>
      <div style="background:linear-gradient(180deg,#2563EB,#7C3AED);border-radius:4px 4px 0 0;height:${barHeight}px;width:60%;margin:0 auto;min-height:20px;"></div>
      <div style="margin-top:8px;font-size:11px;color:var(--el-text-color-secondary);">${day}</div>
    </div>`
  })
  html += '</div>'
  html += `<div style="text-align:center;margin-top:12px;font-size:12px;color:var(--el-text-color-secondary);">日均在线率 <strong style="color:#16A34A;">95.9%</strong> ｜ 峰值 ${Math.max(...online)}% ｜ 最低 ${Math.min(...online)}%</div>`
  el.innerHTML = html
}

function renderAlertChart() {
  const el = document.querySelector('#alert-chart-area')
  if (!el) return
  const types = [
    { name: 'SOS紧急呼叫', count: 45, color: '#EF4444' },
    { name: '跌倒检测', count: 32, color: '#F59E0B' },
    { name: '心率异常', count: 78, color: '#6B7280' },
    { name: '电子围栏', count: 28, color: '#2563EB' },
    { name: '漏服药物', count: 156, color: '#16A34A' },
    { name: '设备离线', count: 3, color: '#F59E0B' },
  ]
  const max = Math.max(...types.map(t => t.count))

  let html = '<div style="padding:10px 0;">'
  types.forEach(t => {
    const pct = (t.count / max * 100).toFixed(0)
    html += `<div style="display:flex;align-items:center;margin-bottom:16px;gap:12px;">
      <span style="width:80px;font-size:13px;text-align:right;color:var(--el-text-color-primary);">${t.name}</span>
      <div style="flex:1;background:#f0f2f5;border-radius:4px;height:24px;overflow:hidden;">
        <div style="width:${pct}%;height:100%;background:${t.color};border-radius:4px;display:flex;align-items:center;justify-content:flex-end;padding-right:8px;">
          <span style="font-size:12px;color:#fff;font-weight:600;">${t.count}</span>
        </div>
      </div>
      <span style="width:40px;font-size:12px;color:var(--el-text-color-secondary);">${(t.count / 442 * 100).toFixed(0)}%</span>
    </div>`
  })
  html += '</div>'
  html += `<div style="text-align:center;font-size:12px;color:var(--el-text-color-secondary);">本周告警总计 <strong style="color:var(--el-text-color-primary);">442</strong> 起</div>`
  el.innerHTML = html
}

function renderMedicationChart() {
  const el = document.querySelector('#medication-chart-area')
  if (!el) return
  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const adherence = [92.3, 89.5, 91.8, 93.2, 90.1, 87.6, 91.0]

  let html = '<div style="display:flex;justify-content:space-between;align-items:flex-end;height:200px;padding:10px 0;">'
  const max = 100
  const min = 80
  days.forEach((day, i) => {
    const barHeight = ((adherence[i] - min) / (max - min)) * 160
    const color = adherence[i] >= 90 ? '#16A34A' : (adherence[i] >= 85 ? '#F59E0B' : '#EF4444')
    html += `<div style="flex:1;text-align:center;">
      <div style="margin-bottom:4px;font-size:12px;font-weight:600;color:${color};">${adherence[i]}%</div>
      <div style="background:${color};border-radius:4px 4px 0 0;height:${barHeight}px;width:60%;margin:0 auto;min-height:20px;"></div>
      <div style="margin-top:8px;font-size:11px;color:var(--el-text-color-secondary);">${day}</div>
    </div>`
  })
  html += '</div>'
  html += `<div style="text-align:center;margin-top:12px;font-size:12px;color:var(--el-text-color-secondary);">周平均依从率 <strong style="color:#16A34A;">90.7%</strong></div>`
  el.innerHTML = html
}

function renderUserGrowthChart() {
  const el = document.querySelector('#user-growth-chart-area')
  if (!el) return
  const months = ['2月', '3月', '4月', '5月', '6月', '7月']
  const familyUsers = [1200, 2100, 3400, 4800, 6200, 8234]
  const elderlyProfiles = [800, 1500, 2400, 3500, 4800, 6500]

  let html = '<div style="display:flex;justify-content:space-between;align-items:flex-end;height:200px;padding:10px 0;">'
  const max = Math.max(...familyUsers)
  months.forEach((month, i) => {
    const barH = (familyUsers[i] / max * 160).toFixed(0)
    html += `<div style="flex:1;text-align:center;">
      <div style="margin-bottom:4px;font-size:12px;font-weight:600;color:#2563EB;">${familyUsers[i].toLocaleString()}</div>
      <div style="background:linear-gradient(180deg,#2563EB,#7C3AED);border-radius:4px 4px 0 0;height:${barH}px;width:50%;margin:0 auto;min-height:20px;"></div>
      <div style="margin-top:8px;font-size:11px;color:var(--el-text-color-secondary);">${month}</div>
    </div>`
  })
  html += '</div>'
  html += `<div style="text-align:center;margin-top:12px;font-size:12px;color:var(--el-text-color-secondary);">家属用户累计 <strong style="color:#2563EB;">8,234</strong> ｜ 老人档案 <strong style="color:#16A34A;">6,500</strong></div>`
  el.innerHTML = html
}
</script>

<style scoped>
.analytics-page {
  padding: 0;
}
.metric-card {
  margin-bottom: 0;
}
.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.metric-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.metric-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}
.kpi-blue .metric-value { color: #2563EB; }
.kpi-green .metric-value { color: #16A34A; }
.kpi-warning .metric-value { color: #F59E0B; }
.kpi-purple .metric-value { color: #7C3AED; }
.metric-trend {
  font-size: 12px;
}

/* Status badges with dots */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.badge-warning { background: #FFFBEB; color: #D97706; }
.badge-primary { background: #EFF6FF; color: #2563EB; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
.dot-warning { background: #D97706; }
.dot-primary { background: #2563EB; }
.dot-gray { background: #6B7280; }
</style>
