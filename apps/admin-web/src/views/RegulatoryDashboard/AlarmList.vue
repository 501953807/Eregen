<template>
  <HopeTable
    :columns="alertColumns"
    :data="computedAlerts"
    :loading="loading"
    :compact="true"
    class="alarm-table"
  >
    <template #toolbar>
      <div class="alarm-toolbar">
        <span class="alarm-count">共 {{ computedAlerts.length }} 条告警</span>
      </div>
    </template>

    <template #col-triggered_at="{ row }">
      <span class="time-cell">{{ formatTime(row.triggered_at) }}</span>
    </template>

    <template #col-patient_name="{ row }">
      <div class="patient-cell">
        <div class="patient-avatar" :class="row.patient_id?.endsWith('1') ? 'avatar-blue' : 'avatar-pink'">
          {{ (row.patient_name || '?')[0] }}
        </div>
        <div>
          <div class="patient-name">{{ row.patient_name || row.patient_id }}</div>
          <div class="patient-id">ID: {{ row.patient_id }}</div>
        </div>
      </div>
    </template>

    <template #col-alert_type="{ row }">
      <HopeBadge :color="alertBadgeColor(row.alert_type)">{{ row.alert_type || row.detail?.slice(0, 20) || '—' }}</HopeBadge>
    </template>

    <template #col-severity="{ row }">
      <span :class="['sev-badge', 'sev-' + row.severity]">{{ severityLabel(row.severity) }}</span>
    </template>

    <template #col-detail="{ row }">
      <span class="detail-cell">{{ row.detail || '—' }}</span>
    </template>

    <template #col-actions="{ row }">
      <div class="action-btns">
        <HopeBtn variant="text" size="sm" @click="$emit('location', row)">定位</HopeBtn>
        <HopeBtn v-if="row.status === 'pending'" variant="text" size="sm" color="success" @click="$emit('acknowledge', row.id)">确认</HopeBtn>
        <HopeBtn v-if="row.status !== 'resolved'" variant="text" size="sm" color="warning" @click="$emit('resolve', row.id)">解决</HopeBtn>
      </div>
    </template>
  </HopeTable>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { HopeTable, HopeBadge, HopeBtn } from '@/components/hope'

const props = defineProps<{
  alerts: any[]
  loading: boolean
  severity: string
  search: string
}>()

const emit = defineEmits<{
  location: [alert: any]
  acknowledge: [id: string]
  resolve: [id: string]
}>()

const alertColumns = [
  { prop: 'triggered_at', label: '时间' },
  { prop: 'patient_name', label: '患者' },
  { prop: 'alert_type', label: '告警类型' },
  { prop: 'severity', label: '等级' },
  { prop: 'detail', label: '详情' },
  { prop: 'actions', label: '操作' },
]

const computedAlerts = computed(() => {
  let result = props.alerts
  if (props.severity) {
    result = result.filter((a: any) => a.severity === props.severity)
  }
  if (props.search) {
    const q = props.search.toLowerCase()
    result = result.filter((a: any) =>
      (a.patient_name || '').toLowerCase().includes(q) ||
      (a.alert_type || '').toLowerCase().includes(q) ||
      (a.detail || '').toLowerCase().includes(q)
    )
  }
  return result
})

function formatTime(ts?: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function severityLabel(sev?: string): string {
  if (sev === 'high') return 'P0'
  if (sev === 'medium') return 'P1'
  return 'P2'
}

function alertBadgeColor(type?: string): 'error' | 'warning' | 'info' | 'primary' {
  if (!type) return 'info'
  const t = type.toLowerCase()
  if (t.includes('围栏') || t.includes('越界') || t.includes('跌倒')) return 'error'
  if (t.includes('心率') || t.includes('生命') || t.includes('用药') || t.includes('漏服')) return 'warning'
  return 'primary'
}
</script>

<style scoped>
.alarm-toolbar { display: flex; justify-content: flex-end; margin-bottom: 4px; }
.alarm-count { font-size: 13px; color: var(--hope-text-muted); font-weight: 500; }
.time-cell { font-size: 12px; color: var(--hope-text-muted); font-variant-numeric: tabular-nums; }
.patient-cell { display: flex; align-items: center; gap: 8px; }
.patient-avatar {
  width: 30px; height: 30px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 600; flex-shrink: 0;
}
.avatar-blue  { background: rgba(var(--hope-info-rgb, 7,154,162), 0.12); color: var(--hope-info, #079aa2); }
.avatar-pink  { background: rgba(var(--hope-accent-rgb, 140,87,255), 0.12); color: var(--hope-accent, #8C57FF); }
.patient-name { font-weight: 600; font-size: 13px; color: var(--hope-text); }
.patient-id   { font-size: 11px; color: var(--hope-text-muted); }
.detail-cell  { font-size: 13px; color: var(--hope-text-secondary); }
.sev-badge { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: var(--hope-radius-pill); font-size: 12px; font-weight: 700; letter-spacing: 0.02em; }
.sev-high   { background: rgba(var(--hope-error-rgb), 0.10); color: var(--hope-error); }
.sev-medium { background: rgba(var(--hope-warning-rgb), 0.12); color: var(--hope-warning); }
.sev-low    { background: rgba(var(--hope-success-rgb), 0.10); color: var(--hope-success); }
.action-btns { display: flex; gap: 4px; flex-wrap: wrap; }
</style>
