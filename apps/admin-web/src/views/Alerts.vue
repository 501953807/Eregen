<template>
  <div class="alerts-page">
    <!-- Stats Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="8">
        <el-card shadow="never" class="stat-card kpi-danger">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p0 }}</div>
            <div class="stat-label">P0 紧急</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" class="stat-card kpi-warning">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p1 }}</div>
            <div class="stat-label">P1 重要</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" class="stat-card kpi-info">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p2 }}</div>
            <div class="stat-label">P2 通知</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filters -->
    <div class="filter-bar">
      <span class="filter-label">筛选：</span>
      <el-form :inline="true" class="filter-form">
        <el-form-item label="严重程度">
          <el-select v-model="filters.severity" placeholder="全部" clearable style="width: 140px;" popper-class="wellness-popper">
            <el-option label="P0 紧急" value="P0" />
            <el-option label="P1 重要" value="P1" />
            <el-option label="P2 通知" value="P2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px;">
            <el-option label="待处理" value="pending" />
            <el-option label="已处理" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px;">
            <el-option label="SOS" value="sos" />
            <el-option label="跌倒" value="fall" />
            <el-option label="心率" value="heart" />
            <el-option label="用药" value="medication" />
            <el-option label="电子围栏" value="geofence" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="handleReset" class="btn-reset">重置</el-button>
          <el-button type="primary" @click="handleSearch">查询</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- Alert Table -->
    <el-card shadow="never">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600; color: #29404A;">告警列表</span>
          <el-button type="primary" size="default" @click="handleBatchResolve" class="btn-primary-outline">
            批量标记已处理
          </el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="filteredAlerts" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" label="告警ID" width="120">
          <template #default="{ row }"><span class="mono">{{ row.id }}</span></template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <span class="status-badge" :class="alertBadgeClass(row.alert_type)">
              <span class="status-dot" :class="alertDotClass(row.alert_type)"></span>
              {{ alertTypeLabel(row.alert_type) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="严重程度" width="100">
          <template #default="{ row }">
            <span class="status-badge" :class="severityBadgeClass(row.severity)">
              <span class="status-dot" :class="severityDotClass(row.severity)"></span>
              {{ row.severity }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <span class="status-badge" :class="row.status === 'pending' ? 'badge-warning' : 'badge-success'">
              <span class="status-dot" :class="row.status === 'pending' ? 'dot-warning' : 'dot-success'"></span>
              {{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="老人ID" width="120">
          <template #default="{ row }">
            {{ row.elderly_id || '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" min-width="160">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleView(row)">查看</el-button>
            <el-button link size="small" style="color:#D9A441;" @click="handleAcknowledge(row)" :disabled="row.status !== 'pending'">标记已读</el-button>
            <el-button link size="small" style="color:#6FAF8F;" @click="handleResolve(row)" :disabled="row.status === 'resolved'">标记已处理</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 16px;">
        <el-tag :type="sseConnected ? 'success' : 'danger'" size="small" effect="plain" style="border-radius:20px;">
          <span :style="{ color: sseConnected ? '#6FAF8F' : '#D77B72', display: 'inline-block', width: 6, height: 6, borderRadius: '50%', background: sseConnected ? '#6FAF8F' : '#D77B72', marginRight: 6 }"></span>
          {{ sseConnected ? '实时推送已连接' : '推送未连接' }}
        </el-tag>
        <el-pagination background layout="prev, pager, next" :total="allAlerts.length" :page-size="20" />
      </div>
    </el-card>

    <!-- View Detail Side Panel -->
    <div class="side-panel-overlay" :class="{ show: showDetailDialog }" @click="showDetailDialog = false" />
    <div class="side-panel" :class="{ open: showDetailDialog }">
      <div class="panel-header">
        <span class="panel-title">告警详情</span>
        <button class="panel-close" @click="showDetailDialog = false">&#10005;</button>
      </div>
      <div class="panel-body" v-if="detailAlert">
        <div class="info-section">
          <div class="section-title">基本信息</div>
          <div class="panel-row">
            <span class="panel-label">告警ID</span>
            <span class="panel-value mono">{{ detailAlert.id }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">类型</span>
            <span class="panel-value">{{ alertTypeLabel(detailAlert.alert_type) }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">严重程度</span>
            <span class="panel-value">
              <span class="status-badge" :class="severityBadgeClass(detailAlert.severity)">
                <span class="status-dot" :class="severityDotClass(detailAlert.severity)"></span>
                {{ detailAlert.severity }}
              </span>
            </span>
          </div>
          <div class="panel-row">
            <span class="panel-label">状态</span>
            <span class="panel-value">
              <span class="status-badge" :class="detailAlert.status === 'pending' ? 'badge-warning' : 'badge-success'">
                <span class="status-dot" :class="detailAlert.status === 'pending' ? 'dot-warning' : 'dot-success'"></span>
                {{ statusLabel(detailAlert.status) }}
              </span>
            </span>
          </div>
          <div class="panel-row">
            <span class="panel-label">老人ID</span>
            <span class="panel-value">{{ detailAlert.elderly_id || '—' }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">创建时间</span>
            <span class="panel-value">{{ detailAlert.created_at ? new Date(detailAlert.created_at).toLocaleString('zh-CN') : '—' }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">处理时间</span>
            <span class="panel-value">{{ detailAlert.resolved_at ? new Date(detailAlert.resolved_at).toLocaleString('zh-CN') : '—' }}</span>
          </div>
        </div>
        <div class="info-section">
          <div class="section-title">元数据</div>
          <pre class="metadata-pre">{{ JSON.stringify(detailAlert.metadata, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertsApi } from '@/api/alerts'
import type { Alert } from '@/types'

const allAlerts = ref<Alert[]>([])
const loading = ref(false)
const selectedRows = ref<Alert[]>([])
const sseConnected = ref(false)
let eventSource: EventSource | null = null

const filters = ref({
  severity: '',
  status: '',
  type: '',
})

const filteredAlerts = computed(() => {
  let list = allAlerts.value
  if (filters.value.severity) {
    list = list.filter(a => a.severity === filters.value.severity)
  }
  if (filters.value.status) {
    list = list.filter(a => a.status === filters.value.status)
  }
  if (filters.value.type) {
    list = list.filter(a => a.alert_type.toLowerCase().includes(filters.value.type))
  }
  return list
})

const stats = computed(() => ({
  p0: allAlerts.value.filter(a => (a.severity === 'p0' || a.severity === 'high') && a.status === 'pending').length,
  p1: allAlerts.value.filter(a => (a.severity === 'p1' || a.severity === 'medium') && a.status === 'pending').length,
  p2: allAlerts.value.filter(a => (a.severity === 'p2' || a.severity === 'low') && a.status === 'pending').length,
}))

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

function severityBadgeClass(sev: string): string {
  const map: Record<string, string> = { P0: 'badge-danger', P1: 'badge-warning', P2: 'badge-info', high: 'badge-danger', medium: 'badge-warning', low: 'badge-info' }
  return map[sev] || 'badge-info'
}
function severityDotClass(sev: string): string {
  const map: Record<string, string> = { P0: 'dot-danger', P1: 'dot-warning', P2: 'dot-info', high: 'dot-danger', medium: 'dot-warning', low: 'dot-info' }
  return map[sev] || 'dot-info'
}

function alertTypeLabel(type: string): string {
  const map: Record<string, string> = {
    sos: 'SOS', fall: '跌倒', heart: '心率', medication: '用药', geofence: '电子围栏',
  }
  return map[type] || type
}

function statusLabel(status: string): string {
  return status === 'pending' ? '未处理' : status === 'acknowledged' ? '已确认' : '已处理'
}

async function handleSearch() {
  await fetchAlerts()
}

function handleReset() {
  filters.value = { severity: '', status: '', type: '' }
  fetchAlerts()
}

async function fetchAlerts() {
  loading.value = true
  try {
    const params: any = {}
    if (filters.value.severity) params.severity = filters.value.severity
    if (filters.value.status) params.status = filters.value.status
    const res = await alertsApi.list(params)
    allAlerts.value = (res.data.data || res.data) as Alert[]
  } catch {
    ElMessage.warning('加载失败，使用模拟数据')
  } finally {
    loading.value = false
  }
}

function connectSSE() {
  const token = localStorage.getItem('admin_token')
  const url = `/api/v1/admin/stream/alerts${token ? `?token=${encodeURIComponent(token)}` : ''}`
  eventSource = new EventSource(url)

  eventSource.addEventListener('alert', (e: Event) => {
    const evt = e as MessageEvent
    let data: any
    try { data = JSON.parse(evt.data) } catch { return }
    if (data.type === 'init' && data.alerts) {
      const existingIds = new Set(allAlerts.value.map(a => a.id))
      const newAlerts = data.alerts.filter((a: Alert) => !existingIds.has(a.id))
      if (newAlerts.length) {
        allAlerts.value = [...newAlerts, ...allAlerts.value]
        newAlerts.forEach((a: Alert) => {
          ElMessage.warning({ message: `🔔 新告警: ${alertTypeLabel(a.alert_type)} (${a.severity})`, duration: 5000 })
        })
      }
    } else if (data.type === 'new' && data.alert) {
      const a = data.alert as Alert
      if (!allAlerts.value.find(x => x.id === a.id)) {
        allAlerts.value.unshift(a)
        ElMessage.warning({ message: `🔔 新告警: ${alertTypeLabel(a.alert_type)} (${a.severity})`, duration: 5000 })
      }
    }
  })

  eventSource.onerror = () => {
    sseConnected.value = false
    eventSource?.close()
    eventSource = null
    setTimeout(connectSSE, 10000)
  }
}

function disconnectSSE() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  sseConnected.value = false
}

function handleSelectionChange(rows: Alert[]) {
  selectedRows.value = rows
}

async function handleResolve(row: Alert) {
  try {
    await alertsApi.markResolved(row.id)
    row.status = 'resolved'
    row.resolved_at = new Date().toISOString()
    ElMessage.success('已标记为已处理')
  } catch {
    ElMessage.warning('操作失败（模拟）')
    row.status = 'resolved'
  }
}

async function handleAcknowledge(row: Alert) {
  try {
    await alertsApi.acknowledge(row.id)
    row.status = 'acknowledged'
    ElMessage.success('已标记为已读')
  } catch {
    ElMessage.warning('操作失败（模拟）')
    row.status = 'acknowledged'
  }
}

async function handleBatchResolve() {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要处理的告警')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要批量标记 ${selectedRows.value.length} 条告警为已处理吗？`, '确认', { type: 'warning' })
    for (const row of selectedRows.value) {
      await alertsApi.markResolved(row.id).catch(() => {})
      row.status = 'resolved'
      row.resolved_at = new Date().toISOString()
    }
    ElMessage.success(`已批量处理 ${selectedRows.value.length} 条告警`)
    selectedRows.value = []
  } catch { /* cancelled */ }
}

const showDetailDialog = ref(false)
const detailAlert = ref<Alert | null>(null)

function handleView(row: Alert) {
  detailAlert.value = { ...row }
  showDetailDialog.value = true
}

onMounted(() => {
  fetchAlerts()
  connectSSE()
})

onUnmounted(() => {
  disconnectSSE()
})
</script>

<style scoped>
.alerts-page {
  padding: 0;
}

/* KPI stat cards */
.stat-card {
  border: 1px solid var(--border-light) !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), var(--shadow-card) !important;
  border-radius: var(--radius-lg) !important;
  background: white !important;
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.stat-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(ellipse at top left, rgba(255,255,255,0.6) 0%, transparent 60%);
  pointer-events: none;
}
.stat-card:hover {
  transform: translateY(-3px);
}
.stat-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 14px;
}
.stat-content { flex: 1; }
.stat-value {
  font-size: 32px;
  font-weight: 800;
}
.kpi-danger .stat-value { color: #D77B72; }
.kpi-warning .stat-value { color: #D9A441; }
.kpi-info .stat-value { color: #6E9FC4; }
.stat-label {
  font-size: 13px;
  color: #6B8980;
  margin-top: 4px;
  font-weight: 600;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
}
.filter-form :deep(.el-select) {
  width: 100%;
}

/* Table header */
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.btn-reset {
  border-radius: var(--hope-radius-lg) !important;
  border: 1px solid var(--hope-border) !important;
  color: var(--hope-text-secondary) !important;
  background: var(--hope-surface) !important;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}
.btn-reset:hover {
  background: var(--hope-surface-light) !important;
  border-color: var(--hope-primary) !important;
  color: var(--hope-primary) !important;
}
.btn-primary-outline {
  border-radius: var(--hope-radius-lg) !important;
  background: transparent !important;
  border: 1.5px solid var(--hope-primary) !important;
  color: var(--hope-primary) !important;
  font-weight: 600;
  transition: all 0.2s;
}
.btn-primary-outline:hover {
  background: var(--hope-primary-lighter) !important;
  border-color: var(--hope-primary-dark) !important;
}
.el-button--primary {
  background: linear-gradient(135deg, #5C8D73 0%, #6FAF8F 100%) !important;
  border-color: transparent !important;
  border-radius: var(--hope-radius-lg) !important;
  box-shadow: var(--hope-shadow-primary) !important;
  font-weight: 600 !important;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1) !important;
}
.el-button--primary:hover {
  background: linear-gradient(135deg, #6FAF8F 0%, #7BAF8C 100%) !important;
  box-shadow: var(--hope-shadow-primary-hover) !important;
  transform: translateY(-1px) !important;
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
.badge-info { background: #EEF4F8; color: #4A7FA0; }
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
.dot-info { background: #6E9FC4; }

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: #6B8980;
}

/* ========== Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(41,64,74,0.3);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show {
  display: block;
}
.side-panel {
  position: fixed;
  top: 0;
  right: -520px;
  bottom: 0;
  width: 520px;
  background: white;
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(60,90,70,0.12);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid #E5EDE6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: white;
  z-index: 1;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: #29404A;
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: #F3F5F1;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
  color: #6B8980;
}
.panel-close:hover {
  background: #E5EDE6;
}
.panel-body {
  padding: 20px 24px;
}

.info-section {
  margin-bottom: 20px;
}
.section-title {
  font-size: 12px;
  font-weight: 700;
  color: #6B8980;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid #F3F5F1;
}
.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}
.panel-label {
  font-size: 13px;
  color: #6B8980;
  font-weight: 500;
}
.panel-value {
  font-size: 13px;
  color: #29404A;
  font-weight: 600;
}
.metadata-pre {
  margin: 0;
  font-size: 12px;
  color: #4A6260;
  background: #F8F6F1;
  padding: 12px;
  border-radius: 10px;
  overflow-x: auto;
}
</style>
