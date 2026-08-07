<template>
  <div class="audit-detail-page">
    <!-- Breadcrumb Header -->
    <div class="page-header">
      <div>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/regulatory' }">监管总览看板</el-breadcrumb-item>
          <el-breadcrumb-item>穿透审计详情</el-breadcrumb-item>
        </el-breadcrumb>
        <h2 class="page-title" style="margin-top: 8px;">穿透审计详情</h2>
      </div>
      <div class="header-actions">
        <el-button @click="handleRefresh"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>刷新状态</el-button>
        <el-button type="primary" @click="exportReport"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>导出审计报告</el-button>
      </div>
    </div>

    <!-- Patient Info Card -->
    <el-card shadow="hover" class="patient-card">
      <div class="patient-card-inner">
        <div class="patient-avatar" :class="patientData.gender === '女' ? 'avatar-pink' : 'avatar-blue'">{{ patientName.charAt(0) }}</div>
        <div class="patient-details">
          <div class="patient-name-row">
            <span class="patient-name">{{ patientName }}</span>
            <span class="patient-id-badge">ID: {{ patientId }}</span>
          </div>
          <div class="patient-meta-grid">
            <div class="meta-item"><span class="meta-label">性别:</span> <span class="meta-value">{{ patientData.gender }}</span></div>
            <div class="meta-item"><span class="meta-label">年龄:</span> <span class="meta-value">{{ patientData.age }}岁</span></div>
            <div class="meta-item"><span class="meta-label">科室:</span> <span class="meta-value">{{ patientData.department }}</span></div>
            <div class="meta-item"><span class="meta-label">入院日期:</span> <span class="meta-value">{{ formatDate(patientData.admissionDate) }}</span></div>
            <div class="meta-item"><span class="meta-label">主治医生:</span> <span class="meta-value">{{ patientData.doctor }}</span></div>
            <div class="meta-item"><span class="meta-label">腕带状态:</span>
              <span class="status-badge" :class="patientData.wearableStatus === '在线正常' ? 'badge-success' : 'badge-danger'">
                <span class="status-dot" :class="patientData.wearableStatus === '在线正常' ? 'dot-success' : 'dot-danger'"></span>
                {{ patientData.wearableStatus }}
              </span>
            </div>
          </div>
        </div>
        <div class="patient-actions">
          <el-button type="primary" @click="viewRealtimeLocation">查看实时定位</el-button>
          <el-button @click="contactNurseStation">联系护士站</el-button>
        </div>
      </div>
    </el-card>

    <!-- Audit Timeline -->
    <el-card shadow="never" class="timeline-card">
      <template #header>
        <div class="timeline-header">
          <span class="panel-title">全链路数据追溯</span>
          <span class="timeline-meta">共 {{ timeline.length }} 条记录 | 最后更新: {{ lastUpdateTime }}</span>
        </div>
      </template>

      <div class="audit-timeline">
        <div v-for="(item, idx) in timeline" :key="idx" class="timeline-node" :class="item.type">
          <div class="timeline-dot" :class="item.type"></div>
          <div class="timeline-content" :class="item.type">
            <div class="content-title">
              <span>{{ item.icon }} {{ item.title }}</span>
              <span class="content-time">{{ formatTime(item.time) }}</span>
            </div>
            <div class="content-body" v-if="item.bodyHtml">
              <div v-html="item.bodyHtml"></div>
            </div>
            <div class="content-body" v-else>
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
                        <span class="status-badge" :class="cell.tagType === 'success' ? 'badge-success' : cell.tagType === 'warning' ? 'badge-warning' : 'badge-danger'">
                          <span class="status-dot" :class="cell.tagType === 'success' ? 'dot-success' : cell.tagType === 'warning' ? 'dot-warning' : 'dot-danger'"></span>
                          {{ cell.text }}
                        </span>
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
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { regulatoryApi } from '@/api/regulatory'
import type { AuditTrail } from '@/api/regulatory'

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

function getAvatarBg(): string {
  return 'linear-gradient(135deg, #fce7f3, #fbcfe8)'
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
      time: at.patient.admission_date || '',
      lines: [
        { label: '入院科室', value: at.patient.department || '—' },
        { label: '主治医生', value: at.patient.doctor || '—' },
        { label: '诊断结果', value: at.patient.diagnosis || '—' },
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
      time: latest.timestamp || '',
      lines: [
        { label: '核验方式', value: latest.scan_type || '—' },
        { label: '核验结果', value: latest.result === 'matched' ? '匹配成功' : latest.result === 'unmatched' ? '不匹配' : '未找到' },
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
      time: at.medications[0]?.time || '',
      table: {
        headers: ['时间', '药品名称', '剂量', '执行人', '状态'],
        rows: at.medications.slice(0, 5).map((m: any) => [
          m.time || '—',
          m.medication_name || '—',
          m.dosage || '—',
          m.administered_by || '—',
          { text: m.status === 'given' ? '已服用' : '待服用', tagType: m.status === 'given' ? 'success' : 'warning' },
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
    // Update patient info from trail
    if (auditTrail.value?.patient) {
      patientName.value = auditTrail.value.patient.name || patientName.value
      patientData.value = {
        ...patientData.value,
        gender: auditTrail.value.patient.gender || patientData.value.gender,
        age: auditTrail.value.patient.age || patientData.value.age,
        department: auditTrail.value.patient.department || patientData.value.department,
        admissionDate: auditTrail.value.patient.admission_date || patientData.value.admissionDate,
        doctor: auditTrail.value.patient.doctor || patientData.value.doctor,
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
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 8px;
}

/* Patient Card */
.patient-card :deep(.el-card__body) {
  padding: 20px;
}

.patient-card-inner {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.patient-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 700;
  flex-shrink: 0;
}
.avatar-blue { background: #DDEBE1; color: #47745C; }
.avatar-pink { background: #FCE7F3; color: #D48EC0; }

.patient-details {
  flex-grow: 1;
}

.patient-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.patient-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.patient-id-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 8px;
  background: #EFF6FF;
  color: #165DFF;
}

.patient-meta-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.meta-item {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.meta-label {
  color: var(--el-text-color-secondary);
}

.meta-value {
  color: var(--el-text-color-primary);
  font-weight: 500;
}

.patient-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
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
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
.dot-warning { background: #D97706; }

/* Timeline Card */
.timeline-card :deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  border-left: 3px solid #165DFF;
  padding-left: 8px;
}

.timeline-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Timeline */
.audit-timeline {
  position: relative;
  padding-left: 30px;
}

.audit-timeline::before {
  content: '';
  position: absolute;
  left: 8px;
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--el-border-color-light);
}

.timeline-node {
  position: relative;
  margin-bottom: 30px;
}

.timeline-node:last-child {
  margin-bottom: 0;
}

.timeline-dot {
  position: absolute;
  left: -26px;
  top: 4px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid #fff;
}

.timeline-dot.inbound { background: #6FAF8F; }
.timeline-dot.verify { background: #5C8D73; }
.timeline-dot.medication { background: #D9A441; }
.timeline-dot.geofence { background: #D77B72; }
.timeline-dot.discharge { background: #8FA8A0; }

.timeline-content {
  background: #fafafa;
  padding: 15px;
  border-radius: 6px;
  border-left: 3px solid transparent;
  transition: all 0.2s;
}

.timeline-content:hover {
  background: #f5f7fa;
}

.timeline-content.inbound { border-left-color: #16A34A; }
.timeline-content.verify { border-left-color: #165DFF; }
.timeline-content.medication { border-left-color: #F59E0B; }
.timeline-content.geofence { border-left-color: #EF4444; }
.timeline-content.discharge { border-left-color: #6B7280; }

.content-title {
  font-weight: 700;
  font-size: 15px;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.content-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-weight: normal;
}

.content-body {
  font-size: 13px;
  line-height: 1.8;
  color: var(--el-text-color-regular);
}

.data-line {
  margin-bottom: 2px;
}

.data-line strong {
  color: var(--el-text-color-primary);
}

/* Data Table */
.data-table-wrap {
  margin-top: 10px;
  overflow-x: auto;
}

.audit-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.audit-table th,
.audit-table td {
  padding: 8px;
  text-align: left;
  border-bottom: 1px solid var(--el-border-color-light);
}

.audit-table th {
  background: #f5f7fa;
  color: var(--el-text-color-secondary);
  font-weight: 600;
}

.audit-table td {
  color: var(--el-text-color-primary);
}

/* Responsive */
@media (max-width: 1200px) {
  .patient-meta-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
