<template>
  <el-table :data="filteredAlerts" stripe class="alarm-table" v-loading="loading">
    <el-table-column prop="triggered_at" label="时间" width="140">
      <template #default="{ row }">{{ formatTime(row.triggered_at) }}</template>
    </el-table-column>
    <el-table-column label="患者" width="130">
      <template #default="{ row }">
        <div class="patient-cell">
          <div class="patient-avatar" :class="row.patient_id?.endsWith('1') ? 'avatar-blue' : 'avatar-pink'">{{ (row.patient_name || '?')[0] }}</div>
          <div>
            <div class="patient-name">{{ row.patient_name || row.patient_id }}</div>
            <div class="patient-id">ID: {{ row.patient_id }}</div>
          </div>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="告警类型" min-width="150">
      <template #default="{ row }">
        <span class="alert-type-badge" :class="alertTypeClass(row.alert_type)">{{ row.alert_type || row.detail?.slice(0, 20) || '—' }}</span>
      </template>
    </el-table-column>
    <el-table-column label="等级" width="80">
      <template #default="{ row }">
        <el-tag :type="severityTag(row.severity)" size="small" effect="light">{{ severityLabel(row.severity) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="详情" show-overflow-tooltip><template #default="{ row }">{{ row.detail || '—' }}</template></el-table-column>
    <el-table-column label="操作" width="180" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" size="small" @click="$emit('location', row)">查看定位</el-button>
        <el-button v-if="row.status === 'pending'" link type="success" size="small" @click="$emit('acknowledge', row.id)">确认</el-button>
        <el-button v-if="row.status !== 'resolved'" link type="warning" size="small" @click="$emit('resolve', row.id)">解决</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
const props = defineProps<{
  alerts: any[]
  loading: boolean
  severity: string
  search: string
}>()

function formatTime(ts?: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function severityTag(sev?: string): 'danger' | 'warning' | 'info' {
  if (sev === 'high') return 'danger'
  if (sev === 'medium') return 'warning'
  return 'info'
}

function severityLabel(sev?: string): string {
  if (sev === 'high') return 'P0'
  if (sev === 'medium') return 'P1'
  return 'P2'
}

function alertTypeClass(type?: string): string {
  if (!type) return 'badge-info'
  const t = type.toLowerCase()
  if (t.includes('围栏') || t.includes('越界')) return 'badge-danger'
  if (t.includes('心率') || t.includes('生命')) return 'badge-warning'
  if (t.includes('跌倒')) return 'badge-danger'
  if (t.includes('用药') || t.includes('漏服')) return 'badge-warning'
  return 'badge-primary'
}
</script>

<style scoped>
.patient-cell { display: flex; align-items: center; gap: 8px; }
.patient-avatar { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; flex-shrink: 0; }
.avatar-blue { background: #DBEAFE; color: #2563EB; }
.avatar-pink { background: #FCE7F3; color: #EC4899; }
.patient-name { font-weight: 600; font-size: 13px; }
.patient-id { font-size: 11px; color: var(--el-text-color-secondary); }
.alert-type-badge { font-size: 12px; padding: 2px 8px; border-radius: 4px; }
.badge-primary { background: #EFF6FF; color: #2563EB; }
.badge-warning { background: #FFFBEB; color: #D97706; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.badge-info { background: #F8FAFC; color: #94A3B8; }
</style>
