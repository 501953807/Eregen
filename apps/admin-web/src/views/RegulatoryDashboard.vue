<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { handleApiError, handleApiSuccess } from '@/utils/error'
import { regulatoryApi, type RegulatoryAlert, type RuleConfig, type ComplianceReport } from '@/api/regulatory'
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

onMounted(async () => {
  await Promise.all([loadOverview(), loadPatients(), loadAlerts(), loadRuleConfigs()])
})
</script>

<template>
  <div class="regulatory-page">
    <div class="page-header">
      <h2 class="page-title">监管总览看板</h2>
      <div class="header-actions">
        <el-button @click="loadOverview" size="default"><el-icon><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg></el-icon> 刷新</el-button>
        <el-button type="primary" @click="exportReport" size="default"><el-icon><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg></el-icon> 导出报表</el-button>
      </div>
    </div>

    <KpiCards
      :overview="overview"
      :today-admissions="todayAdmissions"
      :offline-devices="offlineDevices"
      :fence-violations="fenceViolations"
      :auto-handle-rate="autoHandleRate"
    />

    <el-card shadow="never" class="filter-card">
      <el-row :gutter="12" align="middle">
        <el-col :span="5">
          <el-select v-model="filters.department" placeholder="全部科室" clearable filterable style="width: 100%;">
            <el-option v-for="d in departments" :key="d" :label="d" :value="d" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-select v-model="filters.severity" placeholder="告警等级" clearable style="width: 100%;">
            <el-option label="P0 - 紧急" value="high" />
            <el-option label="P1 - 重要" value="medium" />
            <el-option label="P2 - 一般" value="low" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-input v-model="filters.search" placeholder="搜索患者姓名/ID..." clearable />
        </el-col>
        <el-col :span="3">
          <el-button type="primary" @click="handleSearch" size="default"><el-icon><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></el-icon> 查询</el-button>
        </el-col>
        <el-col :span="3" style="text-align: right;"><el-button @click="handleResetFilters">重置</el-button></el-col>
      </el-row>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="16">
        <el-card shadow="never" class="content-panel">
          <template #header>
            <div class="panel-header">
              <span class="panel-title">实时异常告警列表</span>
              <el-button size="small" @click="refreshAlerts">刷新状态</el-button>
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
        </el-card>
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

    <el-card shadow="never" style="margin-top: 20px;">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="在院患者列表" name="patients">
          <PatientList :patient-list="patientList" :loading="loading.patients" @audit="viewAuditTrail" @detail="viewPatientDetail" />
        </el-tab-pane>
        <el-tab-pane label="规则配置" name="rules">
          <RuleConfigPanel
            :rule-status-list="[]"
            :department-stats="[]"
            :rule-configs="ruleConfigs"
            :loading="loading.rules"
            @update-rule="updateRule"
            @edit-rule="editRuleConfig"
            @save-rule="saveRuleConfig"
          />
        </el-tab-pane>
        <el-tab-pane label="合规报表" name="compliance">
          <CompliancePanel :report="complianceReport" @generate="loadComplianceReport" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <PatientDetailPanel v-model="showPatientDetail" :patient="selectedPatient" />
  </div>
</template>

<style scoped>
.regulatory-page { padding: 0; }
.regulatory-page :deep(.el-card:not(.filter-card):not(.content-panel)) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.regulatory-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-title { font-size: 22px; font-weight: 800; color: var(--el-text-color-primary); margin: 0; }
.header-actions { display: flex; gap: 12px; }
.filter-card :deep(.el-card__body) { padding: 16px; }
.content-panel { margin-bottom: 0; }
.panel-header { display: flex; justify-content: space-between; align-items: center; }
.panel-title { font-size: 15px; font-weight: 700; color: var(--el-text-color-primary); border-left: 3px solid #5C8D73; padding-left: 8px; }
:deep(.el-tabs--border-card) { border: none; }
:deep(.el-tabs--border-card > .el-tabs__header) { border-bottom: 1px solid var(--el-border-color-light); margin: 0; }

/* Responsive */
@media (max-width: 768px) {
  .regulatory-page :deep(.el-table) { font-size: 12px; }
  .regulatory-page :deep(.el-table th),
  .regulatory-page :deep(.el-table td) { padding: 6px 4px; }
}
</style>
