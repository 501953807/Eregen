<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { handleApiError, handleApiSuccess } from '@/utils/error'
import { regulatoryApi, type RegulatoryAlert, type RuleConfig, type ComplianceReport } from '@/api/regulatory'
import { HopeCard, HopeBtn, HopeTabs, HopeBadge } from '@/components/hope'
import KpiCards from './RegulatoryDashboard/KpiCards.vue'
import AlarmList from './RegulatoryDashboard/AlarmList.vue'
import PatientList from './RegulatoryDashboard/PatientList.vue'
import RuleConfigPanel from './RegulatoryDashboard/RuleConfigPanel.vue'
import CompliancePanel from './RegulatoryDashboard/CompliancePanel.vue'
import PatientDetailPanel from './RegulatoryDashboard/PatientDetailPanel.vue'

const activeTab = ref('patients')

// Filters
const filters = ref({ department: '', severity: '', search: '' })
const departments = ref<string[]>(['心内科', '康复科', '老年病科', '神经内科'])

// Overview stats
const overview = ref({ total_patients: 1248, wearable_count: 1180, today_alerts: 8, rule_triggers: 24 })
const todayAdmissions = ref(12)
const offlineDevices = ref(3)
const fenceViolations = ref(3)
const autoHandleRate = ref(92)

// Alerts data
const alerts = ref<RegulatoryAlert[]>([])
const loading = ref({ patients: false, rules: false })
const showPatientDetail = ref(false)
const selectedPatient = ref<any>(null)

const filteredAlerts = computed(() => {
  let list = alerts.value
  if (filters.value.severity) list = list.filter(a => a.severity === filters.value.severity)
  if (filters.value.search) {
    const q = filters.value.search.toLowerCase()
    list = list.filter(a => (a.patient_name || '').toLowerCase().includes(q) || (a.patient_id || '').toLowerCase().includes(q))
  }
  return list
})

// Rule configs
const ruleConfigs = ref<RuleConfig[]>([])
const showRuleEdit = ref(false)
const editingRule = ref<Partial<RuleConfig>>({})
const editingRuleConfigJson = ref('')

// Compliance report
const reportDateRange = ref<[Date, Date] | null>(null)
const complianceReport = ref<ComplianceReport | null>(null)

// Rule engine status list
const ruleStatusList = [
  { code: 'R01', name: 'R01: 越界警报', desc: '患者离开设定电子围栏', riskLevel: 'high', triggerCount: 3, triggerText: '已触发 3次', triggerColor: '#EF4444' },
  { code: 'R02', name: 'R02: 生命体征异常', desc: '心率/血压超出安全阈值', riskLevel: 'med', triggerCount: 1, triggerText: '已触发 1次', triggerColor: '#F59E0B' },
  { code: 'R03', name: 'R03: SOS一键呼叫', desc: '患者主动触发紧急求救', riskLevel: 'low', triggerCount: 0, triggerText: '运行正常', triggerColor: '#16A34A' },
  { code: 'R04', name: 'R04: 用药依从性监测', desc: '漏服/多服药物提醒', riskLevel: 'med', triggerCount: 5, triggerText: '已触发 5次', triggerColor: '#F59E0B' },
  { code: 'R05', name: 'R05: 跌倒检测', desc: 'IMU传感器识别跌倒动作', riskLevel: 'low', triggerCount: 0, triggerText: '运行正常', triggerColor: '#16A34A' },
]

// Department stats
const departmentStats = computed(() => {
  const data = [
    { name: '心内科', count: 420 },
    { name: '康复科', count: 315 },
    { name: '老年病科', count: 288 },
    { name: '神经内科', count: 225 },
  ]
  const max = Math.max(...data.map(d => d.count))
  return data.map(d => ({ ...d, barWidth: Math.round((d.count / max) * 100) }))
})

// Actions
function handleSearch() { loadAlerts(); loadPatients() }
function handleResetFilters() { filters.value = { department: '', severity: '', search: '' } }

async function refreshAlerts() {
  await loadAlerts()
  ElMessage.success('告警状态已刷新')
}

async function loadOverview() {
  try {
    const res = await regulatoryApi.getDashboardOverview(filters.value.department ? { department: filters.value.department } : undefined)
    overview.value = res.data?.data || overview.value
  } catch { /* ignore */ }
}

async function loadPatients() {
  loading.value.patients = true
  try {
    const res = await regulatoryApi.getPatientList(filters.value.department ? { department: filters.value.department } : undefined)
    patientList.value = res.data?.data || []
  } finally { loading.value.patients = false }
}

async function loadAlerts() {
  try {
    const params: Record<string, any> = {}
    if (filters.value.severity) params.severity = filters.value.severity
    const res = await regulatoryApi.listAlerts(params)
    alerts.value = res.data?.data || []
  } catch {
    alerts.value = [
      { id: '1', patient_name: '李秀英', patient_id: '8842', department: '心内科', alert_type: '电子围栏越界', severity: 'high', detail: '离开病区范围 50m', status: 'pending', triggered_at: new Date().toISOString(), rule_code: 'R02', hospital_id: 'h001' },
      { id: '2', patient_name: '王建国', patient_id: '7731', department: '康复科', alert_type: '心率异常', severity: 'medium', detail: '持续心率 > 110bpm', status: 'pending', triggered_at: new Date(Date.now() - 300000).toISOString(), rule_code: 'R04', hospital_id: 'h002' },
      { id: '3', patient_name: '赵淑华', patient_id: '9921', department: '老年病科', alert_type: '跌倒检测', severity: 'high', detail: 'IMU检测到剧烈震动', status: 'acknowledged', triggered_at: new Date(Date.now() - 900000).toISOString(), rule_code: 'R05', hospital_id: 'h003' },
      { id: '4', patient_name: '陈志强', patient_id: '6654', department: '神经内科', alert_type: '用药提醒漏服', severity: 'low', detail: '早餐药未确认服用', status: 'pending', triggered_at: new Date(Date.now() - 1800000).toISOString(), rule_code: 'R06', hospital_id: 'h004' },
      { id: '5', patient_name: '刘美兰', patient_id: '5523', department: '心内科', alert_type: '夜间离床超时', severity: 'medium', detail: '离床超过 15分钟', status: 'acknowledged', triggered_at: new Date(Date.now() - 3600000).toISOString(), rule_code: 'R08', hospital_id: 'h001' },
    ]
  }
}

async function acknowledgeAlert(id: string) {
  try { await regulatoryApi.acknowledgeAlert(id, 'current-user'); ElMessage.success('已确认') }
  catch { ElMessage.success('已确认（模拟）') }
  await loadAlerts()
}

async function resolveAlert(id: string) {
  try { await regulatoryApi.resolveAlert(id, 'current-user', '已核实处理'); ElMessage.success('已解决') }
  catch { ElMessage.success('已解决（模拟）') }
  await loadAlerts()
}

function viewPatientLocation(alert: any) { ElMessage.info(`查看 ${alert.patient_name || alert.patient_id} 的实时定位`) }
function viewPatientDetail(patient: any) { selectedPatient.value = patient; showPatientDetail.value = true }
async function viewAuditTrail(patientId: string) { ElMessage.info(`审计追踪: ${patientId}`) }

async function loadRuleConfigs() {
  loading.value.rules = true
  try {
    const res = await regulatoryApi.listRuleConfigs()
    ruleConfigs.value = res.data?.data || []
  } finally { loading.value.rules = false }
}

function editRuleConfig(row: RuleConfig) {
  editingRule.value = { ...row }
  editingRuleConfigJson.value = JSON.stringify(row.config || {}, null, 2)
  showRuleEdit.value = true
}

async function updateRule(row: RuleConfig) {
  try { await regulatoryApi.updateRuleConfig(row.code, row.config || {}); handleApiSuccess('更新成功') }
  catch (e) { handleApiError(e) }
}

async function saveRuleConfig() {
  try {
    const config = JSON.parse(editingRuleConfigJson.value)
    await regulatoryApi.updateRuleConfig(editingRule.value.code!, config)
    showRuleEdit.value = false
    ElMessage.success('保存成功')
    await loadRuleConfigs()
  } catch (e: any) { handleApiError(e, 'JSON 解析失败') }
}

async function loadComplianceReport() {
  try {
    const params: Record<string, any> = {}
    if (reportDateRange.value) {
      params.start_date = reportDateRange.value[0]?.toISOString().slice(0, 10)
      params.end_date = reportDateRange.value[1]?.toISOString().slice(0, 10)
    }
    const res = await regulatoryApi.getComplianceReport(params)
    complianceReport.value = res.data?.data || null
  } catch { ElMessage.info('报表生成中（模拟）') }
}

function exportReport() { ElMessage.info('导出功能开发中...') }

let patientList = ref<any[]>([])

const tabs = [
  { label: '在院患者列表', value: 'patients' },
  { label: '规则配置', value: 'rules' },
  { label: '合规报表', value: 'compliance' },
]

onMounted(async () => {
  await Promise.all([loadOverview(), loadPatients(), loadAlerts(), loadRuleConfigs()])
})
</script>

<template>
  <div class="regulatory-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h2 class="page-title">监管总览看板</h2>
        <p class="page-subtitle">医疗监管规则引擎 · 在院患者异常追踪与合规审计</p>
      </div>
      <div class="header-actions">
        <HopeBtn variant="plain" size="md" @click="loadOverview">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
          </template>
          刷新
        </HopeBtn>
        <HopeBtn variant="filled" size="md" @click="exportReport">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </template>
          导出报表
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards -->
    <KpiCards
      :overview="overview"
      :today-admissions="todayAdmissions"
      :offline-devices="offlineDevices"
      :fence-violations="fenceViolations"
      :auto-handle-rate="autoHandleRate"
    />

    <!-- Filters -->
    <HopeCard class="filter-card" style="margin-bottom: 16px;">
      <template #header>
        <div class="filter-bar">
          <el-select v-model="filters.department" placeholder="全部科室" clearable filterable style="width: 180px;">
            <el-option v-for="d in departments" :key="d" :label="d" :value="d" />
          </el-select>
          <el-select v-model="filters.severity" placeholder="告警等级" clearable style="width: 140px;">
            <el-option label="P0 - 紧急" value="high" />
            <el-option label="P1 - 重要" value="medium" />
            <el-option label="P2 - 一般" value="low" />
          </el-select>
          <el-input v-model="filters.search" placeholder="搜索患者姓名/ID..." clearable style="width: 220px;" />
          <div class="filter-actions">
            <HopeBtn variant="filled" size="sm" @click="handleSearch">查询</HopeBtn>
            <HopeBtn variant="plain" size="sm" @click="handleResetFilters">重置</HopeBtn>
          </div>
        </div>
      </template>
    </HopeCard>

    <!-- Alerts + Rules Row -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="16">
        <HopeCard title="实时异常告警列表">
          <template #header>
            <div class="panel-header">
              <span class="panel-title">实时异常告警列表</span>
              <HopeBtn variant="text" size="sm" @click="refreshAlerts">刷新状态</HopeBtn>
            </div>
          </template>
          <AlarmList
            :alerts="filteredAlerts"
            :loading="loading.patients"
            :severity="filters.severity"
            :search="filters.search"
            @location="viewPatientLocation"
            @acknowledge="acknowledgeAlert"
            @resolve="resolveAlert"
          />
        </HopeCard>
      </el-col>
      <el-col :span="8">
        <RuleConfigPanel
          :rule-status-list="ruleStatusList"
          :department-stats="departmentStats"
          :rule-configs="ruleConfigs"
          :loading="loading.rules"
          @update-rule="updateRule"
          @edit-rule="editRuleConfig"
          @save-rule="saveRuleConfig"
        />
      </el-col>
    </el-row>

    <!-- Tabs: Patients / Rules / Compliance -->
    <HopeCard>
      <HopeTabs
        v-model="activeTab"
        :tabs="tabs"
        :animated="true"
      />
      <div class="tab-content">
        <PatientList
          v-if="activeTab === 'patients'"
          :patient-list="patientList"
          :loading="loading.patients"
          @audit="viewAuditTrail"
          @detail="viewPatientDetail"
        />
        <div v-else-if="activeTab === 'rules'" class="tab-pane-inner">
          <RuleConfigPanel
            :rule-status-list="[]"
            :department-stats="[]"
            :rule-configs="ruleConfigs"
            :loading="loading.rules"
            @update-rule="updateRule"
            @edit-rule="editRuleConfig"
            @save-rule="saveRuleConfig"
          />
        </div>
        <div v-else-if="activeTab === 'compliance'" class="tab-pane-inner">
          <CompliancePanel :report="complianceReport" @generate="loadComplianceReport" />
        </div>
      </div>
    </HopeCard>

    <PatientDetailPanel v-model="showPatientDetail" :patient="selectedPatient" />
  </div>
</template>

<style scoped>
.regulatory-page { padding: 0; }

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding: 24px 28px;
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: 1px solid var(--hope-border);
  box-shadow: var(--hope-shadow-card);
}
.page-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0 0 4px 0;
  letter-spacing: -0.02em;
}
.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
}
.header-actions { display: flex; gap: 10px; align-items: center; }

/* Filter bar */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.filter-actions { display: flex; gap: 8px; margin-left: auto; }

/* Panel header */
.panel-header { display: flex; justify-content: space-between; align-items: center; }
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
  border-left: 3px solid #5C8D73;
  padding-left: 8px;
}

/* Tabs */
.tab-content { padding: 20px 22px 0; }
.tab-pane-inner { padding-top: 4px; }

/* HopeCard override for filter card */
.filter-card :deep(.hope-content-card__body) { padding: 14px 22px; }

/* Responsive */
@media (max-width: 768px) {
  .regulatory-page .page-header { flex-direction: column; gap: 12px; padding: 16px; }
  .regulatory-page :deep(.el-table) { font-size: 12px; }
  .regulatory-page :deep(.el-table th),
  .regulatory-page :deep(.el-table td) { padding: 6px 4px; }
  .filter-bar { flex-direction: column; align-items: stretch; }
  .filter-actions { margin-left: 0; }
}
</style>
