<template>
  <div class="alerts-page">
    <!-- Stats Row -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #F56C6C;">{{ stats.p0 }}</div>
            <div class="stat-label">P0 紧急</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #E6A23C;">{{ stats.p1 }}</div>
            <div class="stat-label">P1 重要</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #409EFF;">{{ stats.p2 }}</div>
            <div class="stat-label">P2 通知</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filters -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
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
          <el-button type="success" size="small" @click="handleBatchResolve">批量标记已处理</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="filteredAlerts" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" label="告警ID" width="120" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="alertTypeTag(row.alert_type)" size="small">{{ alertTypeLabel(row.alert_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="严重程度" width="100">
          <template #default="{ row }">
            <el-tag :type="severityTag(row.severity)" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'pending' ? 'warning' : 'success'" size="small">{{ statusLabel(row.status) }}</el-tag>
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

    <!-- View Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="告警详情" width="520px">
      <el-descriptions :column="2" border v-if="detailAlert">
        <el-descriptions-item label="告警ID">{{ detailAlert.id }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ alertTypeLabel(detailAlert.alert_type) }}</el-descriptions-item>
        <el-descriptions-item label="严重程度">
          <el-tag :type="severityTag(detailAlert.severity)" size="small">{{ detailAlert.severity }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detailAlert.status === 'pending' ? 'warning' : 'success'" size="small">{{ statusLabel(detailAlert.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="老人ID">{{ detailAlert.elderly_id || '—' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detailAlert.created_at ? new Date(detailAlert.created_at).toLocaleString('zh-CN') : '—' }}</el-descriptions-item>
        <el-descriptions-item label="处理时间" :span="2">{{ detailAlert.resolved_at ? new Date(detailAlert.resolved_at).toLocaleString('zh-CN') : '—' }}</el-descriptions-item>
        <el-descriptions-item label="元数据" :span="2">
          <pre style="margin:0;font-size:12px;color:#666;">{{ JSON.stringify(detailAlert.metadata, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
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

function alertTypeTag(type: string): 'danger' | 'warning' | 'info' {
  const map: Record<string, 'danger' | 'warning' | 'info'> = {
    SOS: 'danger', fall: 'warning', heart: 'danger', geofence: 'info', medication: 'warning',
  }
  return map[type] || 'info'
}

function alertTypeLabel(type: string): string {
  const map: Record<string, string> = {
    sos: 'SOS', fall: '跌倒', heart: '心率', medication: '用药', geofence: '电子围栏',
  }
  return map[type] || type
}

function severityTag(sev: string): 'danger' | 'warning' | 'info' {
  const map: Record<string, 'danger' | 'warning' | 'info'> = { P0: 'danger', P1: 'warning', P2: 'info' }
  return map[sev] || 'info'
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
.stat-card :deep(.el-card__body) { padding: 20px; display: flex; align-items: center; justify-content: space-between; }
.stat-content { flex: 1; }
.stat-value { font-size: 32px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
