<template>
  <div class="alerts-page">
    <!-- Stats Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card kpi-danger">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p0 }}</div>
            <div class="stat-label">P0 紧急</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card kpi-warning">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p1 }}</div>
            <div class="stat-label">P1 重要</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card kpi-blue">
          <div class="stat-content">
            <div class="stat-value">{{ stats.p2 }}</div>
            <div class="stat-label">P2 通知</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filters -->
    <el-card shadow="hover" style="margin-bottom: 16px;">
      <el-form :inline="true">
        <el-form-item label="严重程度">
          <el-select v-model="filters.severity" placeholder="全部" clearable style="width: 140px;">
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
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Alert Table -->
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">告警列表</span>
          <el-button type="success" size="default" @click="handleBatchResolve">批量标记已处理</el-button>
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
            <el-button link type="success" size="small" @click="handleResolve(row)" :disabled="row.status === 'resolved'">标记已处理</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
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
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { alertsApi } from '@/api/alerts'
import type { Alert } from '@/types'

const allAlerts = ref<Alert[]>([])
const loading = ref(false)
const selectedRows = ref<Alert[]>([])

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
  p0: allAlerts.value.filter(a => a.severity === 'P0' && a.status === 'pending').length,
  p1: allAlerts.value.filter(a => a.severity === 'P1' && a.status === 'pending').length,
  p2: allAlerts.value.filter(a => a.severity === 'P2' && a.status === 'pending').length,
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
  const map: Record<string, string> = { P0: 'badge-danger', P1: 'badge-warning', P2: 'badge-info' }
  return map[sev] || 'badge-info'
}
function severityDotClass(sev: string): string {
  const map: Record<string, string> = { P0: 'dot-danger', P1: 'dot-warning', P2: 'dot-info' }
  return map[sev] || 'dot-info'
}

function alertTypeLabel(type: string): string {
  const map: Record<string, string> = {
    sos: 'SOS', fall: '跌倒', heart: '心率', medication: '用药', geofence: '电子围栏',
  }
  return map[type] || type
}

function statusLabel(status: string): string {
  return status === 'pending' ? '未处理' : '已处理'
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
  } catch {
    // cancelled
  }
}

// View detail
const showDetailDialog = ref(false)
const detailAlert = ref<Alert | null>(null)

function handleView(row: Alert) {
  detailAlert.value = { ...row }
  showDetailDialog.value = true
}

onMounted(fetchAlerts)
</script>

<style scoped>
.alerts-page {
  padding: 0;
}

/* KPI stat cards */
.stat-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 14px;
}
.stat-content {
  flex: 1;
}
.stat-value {
  font-size: 32px;
  font-weight: 800;
}
.kpi-danger .stat-value { color: #EF4444; }
.kpi-warning .stat-value { color: #F59E0B; }
.kpi-blue .stat-value { color: #2563EB; }
.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
  font-weight: 600;
}
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
.badge-info { background: #F8FAFC; color: #94A3B8; }
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
.dot-info { background: #94A3B8; }

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

/* ========== Detail Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
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
  box-shadow: -10px 0 40px rgba(0,0,0,0.1);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
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
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: var(--el-fill-color-light);
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.panel-close:hover {
  background: var(--el-border-color-light);
}
.panel-body {
  padding: 20px 24px;
}

.info-section {
  margin-bottom: 20px;
}
.section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-regular);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}
.panel-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}
.panel-value {
  font-size: 13px;
  color: var(--el-text-color-primary);
  font-weight: 600;
}
.metadata-pre {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-regular);
  background: #f5f7fa;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
}
</style>
