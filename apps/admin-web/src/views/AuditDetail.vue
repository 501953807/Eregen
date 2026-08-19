<template>
  <div class="audit-detail-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/regulatory' }">监管总览看板</el-breadcrumb-item>
          <el-breadcrumb-item>穿透审计详情</el-breadcrumb-item>
        </el-breadcrumb>
        <h2 class="page-title">穿透审计详情</h2>
      </div>
      <div class="page-header__actions">
        <HopeBtn variant="plain" size="md" @click="handleRefresh">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
          </template>
          刷新状态
        </HopeBtn>
        <HopeBtn variant="filled" size="md" @click="exportReport">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </template>
          导出审计报告
        </HopeBtn>
      </div>
    </div>

    <!-- Patient Info Card -->
    <HopeCard>
      <div class="patient-card">
        <div class="patient-card__avatar-wrap">
          <HopeAvatar
            :name="patientName"
            :size="'lg'"
            :style="{ background: patientData.gender === '女' ? 'linear-gradient(135deg, #fce7f3, #fbcfe8)' : 'linear-gradient(135deg, #ddebfa, #d0e8f7)', color: patientData.gender === '女' ? '#d48ec0' : '#3a57e8' }"
          />
          <span class="patient-card__wearable-status" :class="patientData.wearableStatus === '在线正常' ? 'wearable-online' : 'wearable-offline'">
            <span class="wearable-dot"></span>
            {{ patientData.wearableStatus }}
          </span>
        </div>
        <div class="patient-card__info">
          <div class="patient-card__name-row">
            <span class="patient-card__name">{{ patientName }}</span>
            <span class="patient-card__id">ID: {{ patientId }}</span>
          </div>
          <div class="patient-card__meta">
            <div class="meta-item">
              <span class="meta-label">性别</span>
              <span class="meta-value">{{ patientData.gender }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">年龄</span>
              <span class="meta-value">{{ patientData.age }}岁</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">科室</span>
              <span class="meta-value">{{ patientData.department }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">入院日期</span>
              <span class="meta-value">{{ formatDate(patientData.admissionDate) }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">主治医生</span>
              <span class="meta-value">{{ patientData.doctor }}</span>
            </div>
          </div>
        </div>
        <div class="patient-card__actions">
          <HopeBtn variant="filled" size="md" @click="viewRealtimeLocation">查看实时定位</HopeBtn>
          <HopeBtn variant="outlined" size="md" @click="contactNurseStation">联系护士站</HopeBtn>
        </div>
      </div>
    </HopeCard>

    <!-- Audit Timeline Card -->
    <HopeCard :title="`全链路数据追溯`" :subtitle="`共 ${timeline.length} 条记录 | 最后更新: ${lastUpdateTime}`">
      <div class="audit-timeline">
        <div v-for="(item, idx) in timeline" :key="idx" class="timeline-node" :class="`timeline-node--${item.type}`">
          <div class="timeline-node__line"></div>
          <div class="timeline-node__dot" :class="`timeline-node__dot--${item.type}`"></div>
          <div class="timeline-node__content" :class="`timeline-node__content--${item.type}`">
            <div class="timeline-node__header">
              <span class="timeline-node__icon">{{ item.icon }}</span>
              <span class="timeline-node__title">{{ item.title }}</span>
              <span class="timeline-node__time">{{ formatTime(item.time) }}</span>
            </div>
            <div class="timeline-node__body" v-if="item.bodyHtml">
              <div v-html="item.bodyHtml"></div>
            </div>
            <div class="timeline-node__body" v-else>
              <div v-for="(line, lIdx) in item.lines" :key="lIdx" class="data-line">
                <strong>{{ line.label }}：</strong>{{ renderCell(line.value) }}
              </div>
            </div>
            <div v-if="item.table" class="data-table-wrap">
              <table class="audit-table">
                <thead>
                  <tr v-for="(col, cIdx) in item.table.headers" :key="cIdx">
                    <th>{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, rIdx) in item.table.rows" :key="rIdx">
                    <td v-for="(cell, cIdx2) in row" :key="cIdx2">
                      <template v-if="typeof cell === 'object' && cell !== null">
                        <HopeBadge :color="cell.tagType === 'success' ? 'success' : cell.tagType === 'warning' ? 'warning' : 'error'">
                          {{ cell.text }}
                        </HopeBadge>
                      </template>
                      <template v-else>{{ cell }}</template>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </HopeCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { regulatoryApi } from '@/api/regulatory'
import type { AuditTrail } from '@/api/regulatory'
import { HopeCard, HopeBtn, HopeBadge, HopeAvatar } from '@/components/hope'

const route = useRoute()
const patientId = computed(() => route.params.patientId as string)

const patientName = ref('李秀英')
const patientData = ref({
  gender: '女',
  age: 76,
  department: '心内科',
  admissionDate: '2026-07-15',
  doctor: '张医生',
  wearableStatus: '在线正常',
})

const lastUpdateTime = ref(formatTime(new Date().toISOString()))

// Real API-driven audit trail data
const auditTrail = ref<AuditTrail | null>(null)

function renderCell(value: any): string {
  if (typeof value === 'object' && value !== null) return value.text || ''
  return String(value ?? '')
}

// Build timeline from real audit trail data
const timeline = computed(() => {
  const items: any[] = []
  if (!auditTrail.value) return []

  const at = auditTrail.value

  // Node 1: Admission (入院登记)
  if (at.patient) {
    items.push({
      type: 'inbound',
      icon: '📋',
      title: '入院登记',
      time: at.patient.created_at ? new Date(at.patient.created_at).toLocaleString('zh-CN') : '',
      lines: [
        { label: '入院编号', value: at.patient.admission_no || '—' },
        { label: '入院科室', value: at.patient.department || '—' },
        { label: '床位号', value: at.patient.bed_number || '—' },
        { label: '腕带绑定', value: at.binding ? `设备 ID ${at.binding.device_id || '—'} 已绑定` : '未绑定腕带' },
      ],
    })
  }

  // Node 2: Verification records
  if (at.verifications && at.verifications.length > 0) {
    const latest = at.verifications[0]
    items.push({
      type: 'verify',
      icon: '✅',
      title: '身份核验记录',
      time: latest.verified_at ? new Date(latest.verified_at).toLocaleString('zh-CN') : '',
      lines: [
        { label: '核验方式', value: latest.verification_type || '—' },
        { label: '核验结果', value: latest.matched ? '匹配成功' : latest.result === 'unmatched' ? '不匹配' : '未找到' },
        { label: '核验人员', value: latest.verified_by || '—' },
      ],
    })
  }

  // Node 3: Medication records
  if (at.medications && at.medications.length > 0) {
    items.push({
      type: 'medication',
      icon: '💊',
      title: '用药记录',
      time: at.medications[0]?.created_at ? new Date(at.medications[0].created_at).toLocaleString('zh-CN') : '',
      table: {
        headers: ['时间', '药品名称', '剂量', '频率', '给药途径'],
        rows: at.medications.slice(0, 5).map((m: any) => [
          m.created_at ? new Date(m.created_at).toLocaleString('zh-CN') : '—',
          m.name || '—',
          m.dosage || '—',
          m.frequency || '—',
          m.route || '—',
        ]),
      },
    })
  }

  // Node 4: Geofence alerts
  if (at.alerts_generated && at.alerts_generated.length > 0) {
    const alert = at.alerts_generated[0]
    items.push({
      type: 'geofence',
      icon: '⚠️',
      title: '电子围栏告警',
      time: alert.triggered_at || '',
      lines: [
        { label: '告警等级', value: { text: alert.severity || 'P1', tagType: alert.severity === 'high' ? 'danger' : 'warning' } },
        { label: '触发规则', value: alert.rule_code || '—' },
        { label: '处理状态', value: { text: alert.status === 'resolved' ? '已处理' : '处理中', tagType: alert.status === 'resolved' ? 'success' : 'warning' } },
      ],
    })
  }

  // Node 5: Ward round / vital signs
  if (at.daily_entries && at.daily_entries.length > 0) {
    const entry = at.daily_entries.find((e: any) => e.entry_type === 'ward_round' || e.entry_type === 'vitals') || at.daily_entries[0]
    items.push({
      type: 'verify',
      icon: '❤️',
      title: '生命体征摘要',
      time: entry.timestamp || '',
      lines: [
        { label: '心率', value: entry.content || '—' },
        { label: '血压', value: '—' },
        { label: '血氧', value: '—' },
      ],
    })
  }

  // Node 6: Discharge
  if (at.patient?.discharge_date) {
    items.push({
      type: 'discharge',
      icon: '🚪',
      title: '出院记录',
      time: at.patient.discharge_date,
      lines: [
        { label: '出院类型', value: at.patient.discharge_type || '—' },
        { label: '后续随访', value: '—' },
      ],
    })
  }

  return items
})

function formatDate(dateStr: string): string {
  if (!dateStr) return '—'
  return dateStr
}

function formatTime(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function handleRefresh() {
  loadAuditTrail()
  ElMessage.success('审计数据已刷新')
}

async function loadAuditTrail() {
  try {
    const res = await regulatoryApi.getAuditTrail(patientId.value)
    auditTrail.value = res.data?.data || null
    if (auditTrail.value?.patient) {
      patientName.value = auditTrail.value.patient.name || patientName.value
      const p = auditTrail.value.patient
      patientData.value = {
        ...patientData.value,
        gender: p.gender || patientData.value.gender,
        age: p.age || patientData.value.age,
        department: p.department || patientData.value.department,
        admissionDate: p.created_at || patientData.value.admissionDate,
      }
    }
  } catch (e: any) {
    ElMessage.error('加载审计数据失败: ' + (e.message || 'unknown error'))
  }
}

function viewRealtimeLocation() {
  ElMessage.info(`查看 ${patientName.value} 的实时定位`)
}

function contactNurseStation() {
  ElMessage.info(`正在连接 ${patientData.value.department} 护士站...`)
}

function exportReport() {
  ElMessage.info('导出功能开发中...')
}

onMounted(() => {
  loadAuditTrail()
})
</script>

<style scoped>
.audit-detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ── Page Header ── */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.page-header__left :deep(.el-breadcrumb) {
  margin-bottom: 6px;
}

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0;
  letter-spacing: -0.02em;
}

.page-header__actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
  align-items: flex-start;
  padding-top: 4px;
}

/* ── Patient Card ── */
.patient-card {
  display: flex;
  align-items: flex-start;
  gap: 24px;
}

.patient-card__avatar-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.patient-card__wearable-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: var(--hope-radius-pill);
  white-space: nowrap;
}

.wearable-online {
  background: rgba(26, 160, 83, 0.12);
  color: #1aa053;
}

.wearable-offline {
  background: rgba(192, 50, 33, 0.12);
  color: #c03221;
}

.wearable-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.patient-card__info {
  flex: 1;
  min-width: 0;
}

.patient-card__name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.patient-card__name {
  font-size: 20px;
  font-weight: 700;
  color: var(--hope-text);
  letter-spacing: -0.01em;
}

.patient-card__id {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: var(--hope-radius-pill);
  background: var(--hope-primary-lighter);
  color: var(--hope-primary);
}

.patient-card__meta {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px 16px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 12px;
  color: var(--hope-text-muted);
  font-weight: 500;
}

.meta-value {
  font-size: 14px;
  color: var(--hope-text);
  font-weight: 500;
}

.patient-card__actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

/* ── Audit Timeline ── */
.audit-timeline {
  position: relative;
  padding-left: 36px;
}

.audit-timeline::before {
  content: '';
  position: absolute;
  left: 14px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: var(--hope-border);
  border-radius: 1px;
}

.timeline-node {
  position: relative;
  margin-bottom: 24px;
}

.timeline-node:last-child {
  margin-bottom: 0;
}

.timeline-node__line {
  position: absolute;
  left: -22px;
  top: 20px;
  bottom: -8px;
  width: 2px;
  background: transparent;
}

.timeline-node__dot {
  position: absolute;
  left: -28px;
  top: 4px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid var(--hope-surface);
  box-shadow: 0 0 0 2px currentColor;
}

/* Timeline node colors */
.timeline-node--inbound .timeline-node__dot { background: #1aa053; color: #1aa053; }
.timeline-node--verify .timeline-node__dot { background: #3a57e8; color: #3a57e8; }
.timeline-node--medication .timeline-node__dot { background: #faa938; color: #faa938; }
.timeline-node--geofence .timeline-node__dot { background: #c03221; color: #c03221; }
.timeline-node--discharge .timeline-node__dot { background: #949aab; color: #949aab; }

.timeline-node__content {
  background: var(--hope-surface-light);
  border: 1px solid var(--hope-border);
  border-left: 3px solid transparent;
  border-radius: var(--hope-radius-md);
  padding: 14px 16px;
  transition: all 0.2s ease;
}

.timeline-node__content:hover {
  background: var(--hope-bg);
  border-color: var(--hope-border-strong);
  box-shadow: var(--hope-shadow-sm);
}

.timeline-node--inbound .timeline-node__content { border-left-color: #1aa053; }
.timeline-node--verify .timeline-node__content { border-left-color: #3a57e8; }
.timeline-node--medication .timeline-node__content { border-left-color: #faa938; }
.timeline-node--geofence .timeline-node__content { border-left-color: #c03221; }
.timeline-node--discharge .timeline-node__content { border-left-color: #949aab; }

.timeline-node__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.timeline-node__icon {
  font-size: 16px;
  line-height: 1;
}

.timeline-node__title {
  font-weight: 700;
  font-size: 14px;
  color: var(--hope-text);
}

.timeline-node__time {
  font-size: 12px;
  color: var(--hope-text-muted);
  margin-left: auto;
}

.timeline-node__body {
  font-size: 13px;
  line-height: 1.8;
  color: var(--hope-text-secondary);
}

.data-line {
  margin-bottom: 2px;
}

.data-line strong {
  color: var(--hope-text);
  font-weight: 600;
}

/* Data Table */
.data-table-wrap {
  margin-top: 12px;
  overflow-x: auto;
  border-radius: var(--hope-radius-sm);
  border: 1px solid var(--hope-border);
}

.audit-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.audit-table th,
.audit-table td {
  padding: 9px 12px;
  text-align: left;
  border-bottom: 1px solid var(--hope-border);
}

.audit-table th {
  background: var(--hope-bg);
  color: var(--hope-text-secondary);
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.audit-table td {
  color: var(--hope-text);
}

.audit-table tr:last-child td {
  border-bottom: none;
}

.audit-table tr:hover td {
  background: var(--hope-primary-lighter);
}

/* Responsive */
@media (max-width: 1200px) {
  .patient-card {
    flex-wrap: wrap;
  }

  .patient-card__meta {
    grid-template-columns: repeat(2, 1fr);
  }

  .patient-card__actions {
    flex-direction: row;
    width: 100%;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
  }

  .page-header__actions {
    width: 100%;
  }

  .page-header__actions .hope-btn {
    flex: 1;
  }

  .patient-card__meta {
    grid-template-columns: 1fr;
  }

  .audit-timeline {
    padding-left: 28px;
  }

  .audit-timeline::before {
    left: 10px;
  }

  .timeline-node__dot {
    left: -24px;
  }
}
</style>
