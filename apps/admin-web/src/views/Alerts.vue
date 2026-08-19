<template>
  <div class="alerts-page">
    <!-- Stats Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card kpi-primary">
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总告警数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card kpi-warning">
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card kpi-success">
          <div class="stat-content">
            <div class="stat-value">{{ stats.resolved }}</div>
            <div class="stat-label">已处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card kpi-danger">
          <div class="stat-content">
            <div class="stat-value">{{ stats.sos }}</div>
            <div class="stat-label">SOS 紧急</div>
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
            <el-option label="高 (P0)" value="p0" />
            <el-option label="中 (P1)" value="p1" />
            <el-option label="低 (P2)" value="p2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px;">
            <el-option label="待处理" value="pending" />
            <el-option label="已确认" value="acknowledged" />
            <el-option label="已处理" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px;">
            <el-option label="SOS" value="sos" />
            <el-option label="跌倒" value="fall" />
            <el-option label="越界" value="geofence" />
            <el-option label="健康异常" value="heart" />
            <el-option label="设备离线" value="offline" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <HopeBtn variant="plain" size="sm" @click="handleReset">重置</HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="handleSearch">查询</HopeBtn>
        </el-form-item>
      </el-form>
    </div>

    <!-- Alert Table -->
    <HopeCard title="告警列表" class="table-card">
      <template #header-actions>
        <HopeBtn variant="outlined" size="sm" @click="handleBatchResolve">批量标记已处理</HopeBtn>
      </template>
      <el-table v-loading="loading" :data="filteredAlerts" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" label="告警ID" width="120">
          <template #default="{ row }"><span class="mono">{{ row.id }}</span></template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <HopeBadge :color="alertBadgeColor(row.alert_type)">{{ alertTypeLabel(row.alert_type) }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="严重程度" width="100">
          <template #default="{ row }">
            <HopeBadge :color="severityBadgeColor(row.severity)">{{ row.severity }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <HopeBadge :color="statusBadgeColor(row.status)">{{ statusLabel(row.status) }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="老人ID" width="120">
          <template #default="{ row }">
            {{ row.elderly_id || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="设备ID" width="120">
          <template #default="{ row }">
            <span class="mono">{{ row.device_id || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" min-width="160">
          <template #default="{ row }">
            <HopeBtn variant="text" size="sm" @click="handleView(row)">查看</HopeBtn>
            <HopeBtn variant="text" size="sm" :disabled="row.status !== 'pending'" @click="handleAcknowledge(row)">标记已读</HopeBtn>
            <HopeBtn variant="success" size="sm" :disabled="row.status === 'resolved'" @click="handleResolve(row)">标记已处理</HopeBtn>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <div style="display: flex; align-items: center; gap: 12px;">
          <el-tag :type="sseConnected ? 'success' : 'danger'" size="small" effect="plain" style="border-radius:20px;">
            <span :style="{ background: sseConnected ? 'var(--hope-success)' : 'var(--hope-error)', display: 'inline-block', width: 6, height: 6, borderRadius: '50%', marginRight: 6 }"></span>
            {{ sseConnected ? '实时推送已连接' : '推送未连接' }}
          </el-tag>
        </div>
        <el-pagination background layout="prev, pager, next" :total="allAlerts.length" :page-size="20" />
      </template>
    </HopeCard>

    <!-- View Detail Side Panel -->
    <div class="side-panel-overlay" :class="{ show: showDetailDialog }" @click="showDetailDialog = false" />
    <div class="side-panel" :class="{ open: showDetailDialog }">
      <div class="panel-header">
        <span class="panel-title">告警详情</span>
        <HopeBtn variant="ghost" size="sm" icon-only @click="showDetailDialog = false">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </HopeBtn>
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
              <HopeBadge :color="severityBadgeColor(detailAlert.severity)">{{ detailAlert.severity }}</HopeBadge>
            </span>
          </div>
          <div class="panel-row">
            <span class="panel-label">状态</span>
            <span class="panel-value">
              <HopeBadge :color="statusBadgeColor(detailAlert.status)">{{ statusLabel(detailAlert.status) }}</HopeBadge>
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
import { HopeCard, HopeBtn, HopeBadge } from '@/components/hope'

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
  total: allAlerts.value.length,
  pending: allAlerts.value.filter(a => a.status === 'pending').length,
  resolved: allAlerts.value.filter(a => a.status === 'resolved').length,
  sos: allAlerts.value.filter(a => a.alert_type === 'sos' && a.status === 'pending').length,
}))

function alertBadgeColor(type: string): 'error' | 'warning' | 'primary' {
  if (['SOS', 'heart'].includes(type)) return 'error'
  if (['fall', 'medication'].includes(type)) return 'warning'
  return 'primary'
}

function severityBadgeColor(sev: string): 'error' | 'warning' | 'primary' {
  if (sev === 'p0') return 'error'
  if (sev === 'p1') return 'warning'
  return 'primary'
}

function statusBadgeColor(status: string): 'error' | 'success' | 'warning' {
  return status === 'pending' ? 'warning' : status === 'resolved' ? 'success' : 'warning'
}

function alertTypeLabel(type: string): string {
  const map: Record<string, string> = {
    sos: 'SOS', fall: '跌倒', heart: '健康异常', medication: '用药', geofence: '越界', offline: '设备离线',
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
    allAlerts.value = (res.data || []) as Alert[]
  } catch {
    ElMessage.warning('加载失败，使用模拟数据')
  } finally {
    loading.value = false
  }
}

function connectSSE() {
  const token = localStorage.getItem('admin_token')
  const url = `/ws/alerts${token ? `?token=${encodeURIComponent(token)}` : ''}`
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
          ElMessage.warning({ message: `新告警: ${alertTypeLabel(a.alert_type)} (${a.severity})`, duration: 5000 })
        })
      }
    } else if (data.type === 'new' && data.alert) {
      const a = data.alert as Alert
      if (!allAlerts.value.find(x => x.id === a.id)) {
        allAlerts.value.unshift(a)
        ElMessage.warning({ message: `新告警: ${alertTypeLabel(a.alert_type)} (${a.severity})`, duration: 5000 })
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
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border-radius: var(--hope-radius-lg) !important;
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
  box-shadow: var(--hope-shadow-md) !important;
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
  letter-spacing: -0.03em;
  line-height: 1;
}
.kpi-danger .stat-value { color: var(--hope-error); }
.kpi-warning .stat-value { color: var(--hope-warning); }
.kpi-info .stat-value { color: var(--hope-info); }
.stat-label {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin-top: 4px;
  font-weight: 600;
}

.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
}
.filter-form :deep(.el-select) {
  width: 100%;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text-secondary);
  white-space: nowrap;
}

/* Table card */
.table-card {
  margin-bottom: 0;
}
.table-card :deep(.hope-card-header__title) {
  color: var(--hope-text);
  font-weight: 700;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--hope-text-muted);
}

/* ========== Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(26,26,46,0.3);
  backdrop-filter: blur(4px);
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
  background: var(--hope-surface);
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(58,87,232,0.10);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: var(--hope-surface);
  z-index: 1;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
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
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--hope-border);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}
.panel-label {
  font-size: 13px;
  color: var(--hope-text-muted);
  font-weight: 500;
}
.panel-value {
  font-size: 13px;
  color: var(--hope-text);
  font-weight: 600;
}
.metadata-pre {
  margin: 0;
  font-size: 12px;
  color: var(--hope-text-secondary);
  background: var(--hope-surface-light);
  padding: 12px;
  border-radius: var(--hope-radius-md);
  overflow-x: auto;
}

/* Responsive */
@media (max-width: 768px) {
  .alerts-page :deep(.el-col) { width: 100% !important; flex: 0 0 100% !important; }
  .alerts-page :deep(.el-table) { font-size: 12px; }
  .alerts-page :deep(.el-table th),
  .alerts-page :deep(.el-table td) { padding: 6px 4px; }
}
</style>
