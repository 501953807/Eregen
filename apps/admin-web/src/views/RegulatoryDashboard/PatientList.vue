<template>
  <HopeTable
    :columns="patientColumns"
    :data="patientList"
    :loading="loading"
    :compact="true"
    class="patient-table"
  >
    <template #col-name="{ row }">
      <div class="patient-cell">
        <div class="patient-avatar avatar-blue">{{ row.name?.[0] || '?' }}</div>
        <span class="patient-name-cell">{{ row.name }}</span>
      </div>
    </template>

    <template #col-admission_no="{ row }">
      <span class="mono">{{ row.admission_no }}</span>
    </template>

    <template #col-department="{ row }">
      <HopeBadge color="info">{{ row.department }}</HopeBadge>
    </template>

    <template #col-bed_number="{ row }">
      <span class="bed-cell">{{ row.bed_number || '—' }}</span>
    </template>

    <template #col-last_verify="{ row }">
      <span class="verify-cell">{{ row.last_verify || '' }}</span>
      <span v-if="!row.last_verify" class="empty-hint">未核验</span>
    </template>

    <template #col-verify_gap_hours="{ row }">
      <span :class="['verify-tag', row.verify_gap_hours > 12 ? 'tag-danger' : row.verify_gap_hours > 6 ? 'tag-warning' : '']">
        {{ row.verify_gap_hours }}
      </span>
    </template>

    <template #col-fence_status="{ row }">
      <span class="status-badge" :class="row.fence_status === 'inside' ? 'badge-success' : 'badge-danger'">
        <span class="status-dot" :class="row.fence_status === 'inside' ? 'dot-success' : 'dot-danger'"></span>
        {{ row.fence_status === 'inside' ? '在院内' : '已越界' }}
      </span>
    </template>

    <template #col-actions="{ row }">
      <div class="action-btns">
        <HopeBtn variant="text" size="sm" @click="$emit('audit', row.id)">审计追踪</HopeBtn>
        <HopeBtn variant="text" size="sm" @click="$emit('detail', row)">详情</HopeBtn>
      </div>
    </template>
  </HopeTable>
</template>

<script setup lang="ts">
import { HopeTable, HopeBadge, HopeBtn } from '@/components/hope'

defineProps<{ patientList: any[]; loading: boolean }>()
defineEmits<{ audit: [id: string]; detail: [row: any] }>()

const patientColumns = [
  { prop: 'name', label: '姓名' },
  { prop: 'admission_no', label: '住院号' },
  { prop: 'department', label: '科室' },
  { prop: 'bed_number', label: '床号' },
  { prop: 'last_verify', label: '最后核验' },
  { prop: 'verify_gap_hours', label: '距上次核验(h)' },
  { prop: 'fence_status', label: '围栏状态' },
  { prop: 'actions', label: '操作' },
]
</script>

<style scoped>
.patient-cell { display: flex; align-items: center; gap: 8px; }
.patient-avatar {
  width: 30px; height: 30px; border-radius: 50%;
  background: rgba(7,154,162,0.12); color: #079aa2;
  display: flex; align-items: center; justify-content: center;
  font-size: 13px; font-weight: 600; flex-shrink: 0;
}
.patient-name-cell { font-weight: 600; font-size: 13px; color: var(--hope-text); }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; color: var(--hope-text-secondary); }
.bed-cell { font-size: 13px; color: var(--hope-text-secondary); }
.verify-cell { font-size: 13px; color: var(--hope-text-secondary); }
.empty-hint { color: var(--hope-text-muted); font-style: italic; }
.verify-tag { font-size: 12px; font-weight: 600; }
.tag-danger { color: #c03221; }
.tag-warning { color: #b8860b; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: var(--hope-radius-pill); font-size: 12px; font-weight: 600; }
.badge-success { background: rgba(26,160,83,0.10); color: #1aa053; }
.badge-danger  { background: rgba(192,50,33,0.10); color: #c03221; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #1aa053; }
.dot-danger  { background: #c03221; }
.action-btns { display: flex; gap: 4px; }
</style>
