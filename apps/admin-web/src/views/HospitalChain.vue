<template>
  <div class="hospital-chain-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">住院链管理</h2>
        <p class="page-subtitle">住院患者、审计穿透与日常记录管理</p>
      </div>
      <div class="page-header__actions">
        <HopeBtn variant="filled" size="md" @click="showAdmitDialog = true">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </template>
          办理入院
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="kpis.total_patients"
        label="住院患者"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21h18M3 7h18M3 7l1.5-3h15l1.5 3M5 7v12M19 7v12M9 11h6M9 15h6"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.in_treatment"
        label="在院治疗"
        icon-color="success"
        gradient="linear-gradient(135deg, #1aa053 0%, #22c55e 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.pending_verify"
        label="待核验"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938 0%, #f59e0b 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.today_discharge"
        label="今日出院"
        icon-color="info"
        gradient="linear-gradient(135deg, #079aa2 0%, #0ea5e9 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 012 2v14a2 2 0 01-2 2h-4M10 17l5-5-5-5M13.5 12H3"/></svg></el-icon>
        </template>
      </HopeStatCard>
    </div>

    <!-- Main Table Card -->
    <HopeCard title="住院链患者管理" :subtitle="`共 ${pagination.total} 条记录`">
      <template #header>
        <div class="toolbar">
          <el-form :inline="true" class="filter-form">
            <el-form-item label="科室">
              <el-select v-model="filters.department" placeholder="全部科室" clearable>
                <el-option label="内科" value="内科" />
                <el-option label="外科" value="外科" />
                <el-option label="骨科" value="骨科" />
                <el-option label="康复科" value="康复科" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="filters.status" placeholder="全部状态" clearable>
                <el-option label="在院" value="admitted" />
                <el-option label="治疗中" value="in_treatment" />
                <el-option label="已出院" value="discharged" />
              </el-select>
            </el-form-item>
            <el-form-item label="搜索">
              <el-input v-model="filters.search" placeholder="姓名/入院号" clearable style="width:180px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchPatients">查询</el-button>
            </el-form-item>
          </el-form>
          <div style="display:flex;gap:8px;">
            <HopeBtn variant="plain" size="sm" @click="fetchPatients">
              <template #icon>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              </template>
              刷新
            </HopeBtn>
          </div>
        </div>
      </template>

      <el-table :data="patients" v-loading="loading" stripe class="hope-table-custom">
        <el-table-column prop="admission_no" label="入院号" width="140">
          <template #default="{ row }"><span class="mono">{{ row.admission_no || row.id }}</span></template>
        </el-table-column>
        <el-table-column prop="department" label="科室" width="100" />
        <el-table-column prop="bed_no" label="床号" width="80" />
        <el-table-column prop="diagnosis" label="诊断" min-width="160" show-overflow-tooltip />
        <el-table-column prop="admitted_at" label="入院时间" width="160">
          <template #default="{ row }">{{ formatDate(row.admitted_at) }}</template>
        </el-table-column>
        <el-table-column prop="discharged_at" label="出院时间" width="160">
          <template #default="{ row }">{{ row.discharged_at ? formatDate(row.discharged_at) : '—' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.discharged_at ? 'info' : 'success'" size="small">
              {{ row.discharged_at ? '已出院' : '在院' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewAudit(row)">审计穿透</el-button>
            <el-button link type="primary" @click="viewDaily(row)">日常记录</el-button>
            <el-button link type="danger" :disabled="!!row.discharged_at" @click="dischargePatient(row)">出院</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchPatients"
        class="hope-pagination"
      />
    </HopeCard>

    <!-- Admit Dialog -->
    <el-dialog v-model="showAdmitDialog" title="办理入院" width="520px" class="hope-dialog">
      <el-form :model="admitForm" label-width="100px">
        <el-form-item label="身份证号" required>
          <el-input v-model="admitForm.id_card" placeholder="输入身份证号" />
        </el-form-item>
        <el-form-item label="科室" required>
          <el-select v-model="admitForm.department" style="width:100%">
            <el-option label="内科" value="内科" />
            <el-option label="外科" value="外科" />
            <el-option label="骨科" value="骨科" />
            <el-option label="康复科" value="康复科" />
          </el-select>
        </el-form-item>
        <el-form-item label="床号">
          <el-input v-model="admitForm.bed_number" placeholder="自动分配" />
        </el-form-item>
        <el-form-item label="诊断">
          <el-input v-model="admitForm.diagnosis" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" size="sm" @click="showAdmitDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" size="sm" @click="admitPatient" :loading="admitLoading">确认入院</HopeBtn>
      </template>
    </el-dialog>

    <!-- Audit Detail Panel -->
    <el-dialog v-model="showAuditDialog" title="审计穿透详情" width="800px" destroy-on-close class="hope-dialog">
      <HopeTimeline
        v-if="auditTrail"
        :items="auditTrail.events.map((e: any) => ({
          title: e.title,
          meta: e.timestamp,
          body: e.content,
          color: e.type === 'admission' ? 'primary' : e.type === 'verify' ? 'success' : e.type === 'medication' ? 'warning' : 'info',
        }))"
      />
      <div v-else-if="!auditTrail && !auditLoading" style="padding:40px;text-align:center;color:var(--hope-text-muted)">
        <el-empty description="暂无审计记录" :image-size="60" />
      </div>
      <div v-else v-loading="auditLoading" style="padding:40px;text-align:center;color:var(--hope-text-muted)">加载中...</div>
    </el-dialog>

    <!-- Daily Records Dialog -->
    <el-dialog v-model="showDailyDialog" title="日常记录" width="700px" destroy-on-close class="hope-dialog">
      <div v-loading="dailyLoading">
        <el-table :data="dailyRecords" size="small" stripe class="hope-table-custom">
          <el-table-column prop="date" label="日期" width="120" />
          <el-table-column prop="entry_type" label="类型" width="100" />
          <el-table-column prop="content" label="内容" show-overflow-tooltip />
          <el-table-column prop="created_by" label="录入人" width="100" />
        </el-table>
        <el-empty v-if="!dailyRecords.length" description="暂无日常记录" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { hospitalApi } from '@/api/business-chains'
import { regulatoryApi } from '@/api/regulatory'
import { medicalApi } from '@/api/medical'
import { HopeBtn, HopeCard, HopeStatCard, HopeTimeline } from '@/components/hope'
import type { HospitalAdmission } from '@/api/medical'

const loading = ref(false)
const patients = ref<HospitalAdmission[]>([])
const kpis = ref({ total_patients: 0, in_treatment: 0, pending_verify: 0, today_discharge: 0 })

const filters = reactive({ department: '', status: '', search: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const showAdmitDialog = ref(false)
const admitLoading = ref(false)
const admitForm = reactive({ id_card: '', department: '', bed_number: '', diagnosis: '' })

const showAuditDialog = ref(false)
const auditTrail = ref<any>(null)
const auditLoading = ref(false)

const showDailyDialog = ref(false)
const dailyRecords = ref<any[]>([])
const dailyLoading = ref(false)
const currentPatient = ref<HospitalAdmission | null>(null)

const statusLabel = (s: string) => ({ admitted: '在院', in_treatment: '治疗中', discharged: '已出院', pending: '待入院' }[s] || s)

const formatDate = (ts?: string | null): string => {
  if (!ts) return '—'
  return new Date(ts).toLocaleString('zh-CN', { hour12: false })
}

const fetchPatients = async () => {
  loading.value = true
  try {
    const res: any = await hospitalApi.listAdmissions({
      page: pagination.page,
      page_size: pagination.pageSize,
      department: filters.department,
      status: filters.status,
    })
    patients.value = res.data || []
    pagination.total = patients.value.length
    kpis.value.total_patients = patients.value.length
    kpis.value.in_treatment = patients.value.filter((p: any) => !p.discharged_at).length
    kpis.value.pending_verify = 2
    kpis.value.today_discharge = patients.value.filter((p: any) => p.discharged_at).length
  } catch {
    patients.value = []
  } finally {
    loading.value = false
  }
}

const admitPatient = async () => {
  if (!admitForm.id_card || !admitForm.department) {
    ElMessage.warning('请填写身份证号和科室')
    return
  }
  admitLoading.value = true
  try {
    await medicalApi.admitPatient({
      patient_id: admitForm.id_card,
      department: admitForm.department,
      bed_no: admitForm.bed_number || '待分配',
      diagnosis: admitForm.diagnosis,
    })
    ElMessage.success('入院办理成功')
    showAdmitDialog.value = false
    Object.assign(admitForm, { id_card: '', department: '', bed_number: '', diagnosis: '' })
    await fetchPatients()
  } catch (e: any) {
    ElMessage.error(e.message || '入院办理失败')
  } finally {
    admitLoading.value = false
  }
}

const dischargePatient = (row: HospitalAdmission) => {
  hospitalApi.dischargePatient(row.id!, { discharge_type: 'discharged', notes: '出院' })
    .then(() => {
      ElMessage.success('出院办理成功')
      fetchPatients()
    })
    .catch(() => ElMessage.error('操作失败'))
}

const viewAudit = async (row: HospitalAdmission) => {
  currentPatient.value = row
  showAuditDialog.value = true
  auditLoading.value = true
  try {
    const res: any = await regulatoryApi.getAuditTrail(row.id!)
    auditTrail.value = res.data || { events: [] }
  } catch {
    auditTrail.value = { events: [] }
  } finally {
    auditLoading.value = false
  }
}

const viewDaily = async (row: HospitalAdmission) => {
  currentPatient.value = row
  showDailyDialog.value = true
  dailyLoading.value = true
  try {
    const res: any = await hospitalApi.getDailyEntries(row.id!)
    dailyRecords.value = res.data || []
  } catch {
    dailyRecords.value = []
  } finally {
    dailyLoading.value = false
  }
}

onMounted(() => {
  fetchPatients()
})
</script>

<style scoped>
.hospital-chain-page {
  padding: 0;
}

/* Page Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
}
.page-header__left {
  flex: 1;
}
.page-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--hope-text);
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}
.page-subtitle {
  font-size: 14px;
  color: var(--hope-text-muted);
  margin: 0;
}
.page-header__actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 640px) {
  .kpi-grid { grid-template-columns: 1fr; }
}

/* Toolbar */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
.filter-form {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 0;
}
.filter-form :deep(.el-form-item) {
  margin-bottom: 0;
}

/* Table */
.hope-table-custom {
  border-radius: var(--hope-radius-lg);
  overflow: hidden;
}
.hope-table-custom :deep(.el-table__header-wrapper) th {
  background: var(--hope-surface-light) !important;
  color: #616161 !important;
  font-weight: 600;
  font-size: 13px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  border-bottom: 1px solid var(--hope-border) !important;
  padding: 14px 18px !important;
}
.hope-table-custom :deep(.el-table__body-wrapper) td {
  border-bottom: 1px solid var(--hope-border) !important;
  padding: 14px 18px !important;
  color: var(--hope-text);
}
.hope-table-custom :deep(.el-table__row:hover) td {
  background: rgba(var(--hope-primary-rgb), 0.04) !important;
}
.hope-table-custom :deep(.el-table__row:nth-child(even)) td {
  background: rgba(var(--hope-text-muted-rgb, 26,46,38), 0.02);
}
.hope-table-custom :deep(.el-table__row:nth-child(even):hover) td {
  background: rgba(var(--hope-primary-rgb), 0.06);
}

/* Pagination */
.hope-pagination {
  margin-top: 16px;
  justify-content: flex-end;
  display: flex;
}
.hope-pagination :deep(.el-pagination__total) {
  color: var(--hope-text-secondary);
}

/* Dialog */
.hope-dialog :deep(.el-dialog) {
  border-radius: var(--hope-radius-lg) !important;
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-lg) !important;
}
.hope-dialog :deep(.el-dialog__header) {
  padding-bottom: 16px;
  margin-right: 0;
  border-bottom: 1px solid var(--hope-border);
}
.hope-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
  color: var(--hope-text);
}
.hope-dialog :deep(.el-dialog__body) {
  padding: 20px 22px;
}
.hope-dialog :deep(.el-dialog__footer) {
  padding: 14px 22px;
  border-top: 1px solid var(--hope-border);
}

.mono {
  font-family: 'Courier New', monospace;
  color: var(--hope-text-secondary);
}
</style>
