<template>
  <div class="hospital-chain-page">
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ kpis.total_patients }}</div>
          <div class="kpi-label">住院患者</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ kpis.in_treatment }}</div>
          <div class="kpi-label">在院治疗</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-orange">
          <div class="kpi-value">{{ kpis.pending_verify }}</div>
          <div class="kpi-label">待核验</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ kpis.today_discharge }}</div>
          <div class="kpi-label">今日出院</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div class="card-header">
          <span>住院链患者管理</span>
          <div style="display:flex;gap:8px;">
            <el-button type="primary" size="small" @click="showAdmitDialog = true">+ 办理入院</el-button>
            <el-button size="small" @click="fetchPatients">刷新</el-button>
          </div>
        </div>
      </template>

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

      <el-table :data="patients" v-loading="loading" stripe>
        <el-table-column prop="admission_no" label="入院号" width="140">
          <template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template>
        </el-table-column>
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="department" label="科室" width="100" />
        <el-table-column prop="bed_number" label="床号" width="80" />
        <el-table-column prop="diagnosis" label="诊断" min-width="160" show-overflow-tooltip />
        <el-table-column prop="admitted_at" label="入院时间" width="160" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'admitted' || row.status === 'in_treatment' ? 'success' : 'info'" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_verify" label="最后核验" width="140" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewAudit(row)">审计穿透</el-button>
            <el-button link type="primary" @click="viewDaily(row)">日常记录</el-button>
            <el-button link type="danger" @click="dischargePatient(row)">出院</el-button>
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
        style="margin-top:16px;justify-content:flex-end;"
      />
    </el-card>

    <!-- Admit Dialog -->
    <el-dialog v-model="showAdmitDialog" title="办理入院" width="520px">
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
        <el-button @click="showAdmitDialog = false">取消</el-button>
        <el-button type="primary" @click="admitPatient" :loading="admitLoading">确认入院</el-button>
      </template>
    </el-dialog>

    <!-- Audit Detail Panel -->
    <el-dialog v-model="showAuditDialog" title="审计穿透详情" width="800px" destroy-on-close>
      <el-timeline v-if="auditTrail" style="padding:20px">
        <el-timeline-item
          v-for="event in auditTrail.events"
          :key="event.id"
          :timestamp="event.timestamp"
          :type="event.type === 'admission' ? 'primary' : event.type === 'verify' ? 'success' : event.type === 'medication' ? 'warning' : 'info'"
        >
          <el-card shadow="never">
            <div style="font-weight:600">{{ event.title }}</div>
            <div style="font-size:12px;color:var(--el-text-color-secondary)">{{ event.content }}</div>
          </el-card>
        </el-timeline-item>
        <el-timeline-item v-if="!auditTrail.events.length" timestamp="无记录" type="info">
          <el-empty description="暂无审计记录" :image-size="60" />
        </el-timeline-item>
      </el-timeline>
      <div v-else v-loading="auditLoading" style="padding:40px;text-align:center;color:var(--el-text-color-secondary)">加载中...</div>
    </el-dialog>

    <!-- Daily Records Dialog -->
    <el-dialog v-model="showDailyDialog" title="日常记录" width="700px" destroy-on-close>
      <div v-loading="dailyLoading">
        <el-table :data="dailyRecords" size="small" stripe>
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
import { hospitalApi, regulatoryApi } from '@/api/business-chains'
import { medicalApi } from '@/api/medical'
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

const fetchPatients = async () => {
  loading.value = true
  try {
    const res: any = await hospitalApi.listPatients({
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
    const res: any = await regulatoryApi.getAudit(row.id!)
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
.hospital-chain-page { padding: 0; }
.hospital-chain-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.hospital-chain-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-form { margin-bottom: 16px; }
.mono { font-family: 'Courier New', monospace; }
.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.kpi-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(ellipse at top left, rgba(255,255,255,0.6) 0%, transparent 60%);
  pointer-events: none;
}
.kpi-card:hover { transform: translateY(-3px); }
.kpi-card :deep(.el-card__body) { padding: 18px; display: flex; flex-direction: column; align-items: center; text-align: center; border-radius: 14px; }
.kpi-value { font-size: 32px; font-weight: 800; letter-spacing: -0.03em; line-height: 1; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 6px; font-weight: 600; }
</style>
