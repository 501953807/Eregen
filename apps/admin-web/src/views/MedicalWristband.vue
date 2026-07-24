<template>
  <div class="medical-page">
    <!-- KPI Cards -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ stats.active_patients }}</div>
          <div class="kpi-label">在院患者</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ stats.today_admitted }}</div>
          <div class="kpi-label">今日入院</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ stats.bound_devices }}</div>
          <div class="kpi-label">已绑定腕带</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-value">{{ todayStats.matched }}/{{ todayStats.total }}</div>
          <div class="kpi-label">今日核验匹配</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Tabs -->
    <el-tabs v-model="activeTab" type="border-card">
      <!-- Patient Registration -->
      <el-tab-pane label="入院登记" name="patients">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="12">
            <el-input v-model="patientForm.admission_no" placeholder="住院号" clearable />
          </el-col>
          <el-col :span="8">
            <el-input v-model="patientForm.name" placeholder="姓名" clearable />
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="searchByAdmission">查询</el-button>
          </el-col>
        </el-row>

        <el-table :data="patients" v-loading="loading.patients" stripe>
          <el-table-column prop="admission_no" label="住院号" width="140">
            <template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="姓名" width="100">
            <template #default="{ row }">
              <div class="patient-cell">
                <div class="patient-avatar" :class="row.gender === '男' ? 'avatar-blue' : 'avatar-pink'">{{ row.name?.[0] || '?' }}</div>
                <strong>{{ row.name }}</strong>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="gender" label="性别" width="60" />
          <el-table-column prop="age" label="年龄" width="60" />
          <el-table-column prop="department" label="科室" width="120" />
          <el-table-column prop="bed_number" label="床号" width="80" />
          <el-table-column prop="blood_type" label="血型" width="60" />
          <el-table-column prop="allergies" label="过敏史" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.status === 'admitted' ? 'badge-success' : 'badge-gray'">
                <span class="status-dot" :class="row.status === 'admitted' ? 'dot-success' : 'dot-gray'"></span>
                {{ row.status === 'admitted' ? '在院' : '已出院' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" link @click="editPatient(row)">编辑</el-button>
              <el-button size="small" type="warning" link @click="bindDialogVisible = true; bindTarget = row">绑定腕带</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-dialog v-model="showPatientForm" title="编辑/新增患者" width="600px">
          <el-form :model="patientForm" label-width="80px">
            <el-form-item label="住院号"><el-input v-model="patientForm.admission_no" /></el-form-item>
            <el-form-item label="姓名"><el-input v-model="patientForm.name" /></el-form-item>
            <el-form-item label="性别">
              <el-radio-group v-model="patientForm.gender">
                <el-radio value="男">男</el-radio>
                <el-radio value="女">女</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="年龄"><el-input-number v-model="patientForm.age" :min="0" :max="150" /></el-form-item>
            <el-form-item label="科室"><el-input v-model="patientForm.department" /></el-form-item>
            <el-form-item label="床号"><el-input v-model="patientForm.bed_number" /></el-form-item>
            <el-form-item label="血型"><el-input v-model="patientForm.blood_type" /></el-form-item>
            <el-form-item label="过敏史"><el-input v-model="patientForm.allergies" type="textarea" /></el-form-item>
            <el-form-item label="特殊状况"><el-input v-model="patientForm.special_conditions" type="textarea" /></el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showPatientForm = false">取消</el-button>
            <el-button type="primary" @click="savePatient">保存</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <!-- Wristband Binding -->
      <el-tab-pane label="腕带管理" name="wristbands">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="6">
            <el-select v-model="wristbandFilter.status" placeholder="状态筛选" clearable @change="loadWristbands" style="width: 100%;">
              <el-option label="空闲" value="idle" />
              <el-option label="已绑定" value="bound" />
              <el-option label="已清空" value="cleared" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="loadWristbands">刷新</el-button>
          </el-col>
        </el-row>

        <el-table :data="wristbands" v-loading="loading.wristbands" stripe>
          <el-table-column prop="device_id" label="设备ID" width="160">
            <template #default="{ row }"><span class="mono">{{ row.device_id }}</span></template>
          </el-table-column>
          <el-table-column prop="firmware_version" label="固件版本" width="120" />
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="wristbandStatusClass(row.status)">
                <span class="status-dot" :class="wristbandDotClass(row.status)"></span>
                {{ row.status }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="bound_patient_id" label="绑定患者" show-overflow-tooltip />
          <el-table-column label="操作" width="240">
            <template #default="{ row }">
              <el-button size="small" link @click="clearWristband(row.device_id)">清空数据</el-button>
              <el-button size="small" link type="info" @click="writeToFirmware(row.device_id)">写入配置</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Verification Records -->
      <el-tab-pane label="核验记录" name="verifications">
        <el-table :data="verifications" v-loading="loading.verifications" stripe>
          <el-table-column prop="timestamp" label="时间" width="180" />
          <el-table-column prop="patient_id" label="患者ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.patient_id }}</span></template>
          </el-table-column>
          <el-table-column prop="device_id" label="腕带设备" width="140">
            <template #default="{ row }"><span class="mono">{{ row.device_id }}</span></template>
          </el-table-column>
          <el-table-column prop="scan_type" label="类型" width="100">
            <template #default="{ row }">
              <span class="status-badge badge-primary">
                <span class="status-dot dot-primary"></span>
                {{ scanTypeLabel(row.scan_type) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="result" label="结果" width="100">
            <template #default="{ row }">
              <span class="status-badge" :class="resultBadgeClass(row.result)">
                <span class="status-dot" :class="resultDotClass(row.result)"></span>
                {{ resultLabel(row.result) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="verified_by" label="操作人" width="100" />
          <el-table-column prop="notes" label="备注" show-overflow-tooltip />
        </el-table>
      </el-tab-pane>

      <!-- Daily Entries -->
      <el-tab-pane label="每日录入" name="daily">
        <el-date-picker v-model="dailyDate" type="date" placeholder="选择日期" style="margin-bottom: 16px; width: 100%;" />
        <el-table :data="dailyEntries" v-loading="loading.daily" stripe>
          <el-table-column prop="timestamp" label="时间" width="180" />
          <el-table-column prop="patient_id" label="患者ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.patient_id }}</span></template>
          </el-table-column>
          <el-table-column prop="entry_type" label="类型" width="100" />
          <el-table-column prop="content" label="内容" show-overflow-tooltip />
          <el-table-column prop="created_by" label="录入人" width="100" />
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { medicalApi, type Patient, type WristbandDevice, type VerificationRecord } from '@/api/medical'

const activeTab = ref('patients')

// Stats
const stats = ref({ active_patients: 0, today_admitted: 0, bound_devices: 0, total_devices: 0 })
const todayStats = ref({ matched: 0, total: 0, unmatched: 0 })

// Patients
const patients = ref<Patient[]>([])
const patientForm = ref<Partial<Patient>>({ status: 'admitted' })
const showPatientForm = ref(false)
const bindTarget = ref<Patient | null>(null)
const bindDialogVisible = ref(false)

// Wristbands
const wristbands = ref<WristbandDevice[]>([])
const wristbandFilter = ref({ status: '' })

// Verifications
const verifications = ref<VerificationRecord[]>([])

// Daily
const dailyDate = ref(new Date())
const dailyEntries = ref<any[]>([])

// Loading states
const loading = ref({
  patients: false,
  wristbands: false,
  verifications: false,
  daily: false,
})

onMounted(async () => {
  await Promise.all([loadOverview(), loadPatients(), loadWristbands(), loadVerifications()])
})

async function loadOverview() {
  try {
    const res = await medicalApi.getOverview()
    stats.value = res.data?.data || {}
  } catch { /* ignore */ }
}

async function loadPatients() {
  loading.value.patients = true
  try {
    const res = await medicalApi.listPatients({ page: 1, page_size: 50 })
    patients.value = res.data?.data || []
  } finally {
    loading.value.patients = false
  }
}

async function searchByAdmission() {
  try {
    const res = await medicalApi.getByAdmissionNo(patientForm.value.admission_no!)
    patients.value = [res.data?.data]
  } catch {
    ElMessage.error('未找到该住院号')
  }
}

function editPatient(row: Patient) {
  patientForm.value = { ...row }
  showPatientForm.value = true
}

async function savePatient() {
  try {
    if (patientForm.value.id) {
      await medicalApi.updatePatient(patientForm.value.id!, patientForm.value)
      ElMessage.success('更新成功')
    } else {
      await medicalApi.createPatient(patientForm.value)
      ElMessage.success('创建成功')
    }
    showPatientForm.value = false
    await loadPatients()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function loadWristbands() {
  loading.value.wristbands = true
  try {
    const res = await medicalApi.listWristbands({
      page: 1,
      page_size: 50,
      status: wristbandFilter.value.status || undefined,
    })
    wristbands.value = res.data?.data || []
  } finally {
    loading.value.wristbands = false
  }
}

async function clearWristband(deviceId: string) {
  try {
    await medicalApi.clearWristband(deviceId)
    ElMessage.success('腕带已清空')
    await loadWristbands()
  } catch {
    ElMessage.error('清空失败')
  }
}

async function writeToFirmware(deviceId: string) {
  try {
    await medicalApi.writeToFirmware(deviceId, JSON.stringify({ config: 'default' }))
    ElMessage.success('写入成功')
  } catch {
    ElMessage.error('写入失败')
  }
}

async function loadVerifications() {
  loading.value.verifications = true
  try {
    const res = await medicalApi.listVerifications({ page: 1, page_size: 50 })
    verifications.value = res.data?.data || []
    const statsRes = await medicalApi.getTodayStats()
    todayStats.value = statsRes.data?.data || { matched: 0, total: 0, unmatched: 0 }
  } finally {
    loading.value.verifications = false
  }
}

function scanTypeLabel(type: string) {
  const map: Record<string, string> = {
    round: '巡房', medication: '用药', treatment: '治疗', discharge: '出院'
  }
  return map[type] || type
}

function resultLabel(result: string) {
  const map: Record<string, string> = {
    matched: '匹配', unmatched: '不匹配', not_found: '未找到'
  }
  return map[result] || result
}

function wristbandStatusClass(status: string): string {
  if (status === 'bound') return 'badge-success'
  if (status === 'idle') return 'badge-info'
  return 'badge-gray'
}
function wristbandDotClass(status: string): string {
  if (status === 'bound') return 'dot-success'
  if (status === 'idle') return 'dot-info'
  return 'dot-gray'
}

function resultBadgeClass(result: string): string {
  return result === 'matched' ? 'badge-success' : 'badge-danger'
}
function resultDotClass(result: string): string {
  return result === 'matched' ? 'dot-success' : 'dot-danger'
}
</script>

<style scoped>
.medical-page {
  padding: 0;
}

/* KPI Cards */
.kpi-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-radius: 14px;
}
.kpi-value {
  font-size: 28px;
  font-weight: 800;
  line-height: 1.2;
}
.kpi-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 6px;
  font-weight: 600;
}
.kpi-blue .kpi-value { color: #2563EB; }
.kpi-green .kpi-value { color: #16A34A; }
.kpi-purple .kpi-value { color: #7C3AED; }
.kpi-warning .kpi-value { color: #F59E0B; }

/* Patient cell */
.patient-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.patient-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}
.avatar-blue { background: #DBEAFE; color: #2563EB; }
.avatar-pink { background: #FCE7F3; color: #EC4899; }

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
</style>
