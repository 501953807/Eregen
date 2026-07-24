<template>
  <div class="regulatory-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">监管总览看板</h2>
      <div class="header-actions">
        <el-button @click="loadOverview" size="default">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
        <el-button type="primary" @click="exportReport" size="default">
          <el-icon><Download /></el-icon> 导出报表
        </el-button>
      </div>
    </div>

    <!-- KPI Row (4 columns) -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-icon-wrap">🏥</div>
          <div class="kpi-value">{{ overview.total_patients }}</div>
          <div class="kpi-label">在院患者总数</div>
          <div class="kpi-trend trend-up">↑ {{ todayAdmissions }} 今日入院</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-icon-wrap">💍</div>
          <div class="kpi-value">{{ overview.wearable_count }}</div>
          <div class="kpi-label">佩戴腕带设备</div>
          <div class="kpi-trend trend-down">↓ {{ offlineDevices }} 离线</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-danger">
          <div class="kpi-icon-wrap">⚠️</div>
          <div class="kpi-value">{{ overview.today_alerts }}</div>
          <div class="kpi-label">今日异常告警</div>
          <div class="kpi-trend trend-down">↑ {{ fenceViolations }} 越界</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-icon-wrap">⚙️</div>
          <div class="kpi-value">{{ overview.rule_triggers }}</div>
          <div class="kpi-label">规则引擎触发</div>
          <div class="kpi-trend">自动处理率 {{ autoHandleRate }}%</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter Bar -->
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
          <el-button type="primary" @click="handleSearch" size="default">
            <el-icon><Search /></el-icon> 查询
          </el-button>
        </el-col>
        <el-col :span="3" style="text-align: right;">
          <el-button @click="handleResetFilters">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- Content Grid: Alarm List + Rule Engine -->
    <el-row :gutter="20">
      <!-- Real-time Alarm List (2/3 width) -->
      <el-col :span="16">
        <el-card shadow="never" class="content-panel">
          <template #header>
            <div class="panel-header">
              <span class="panel-title">实时异常告警列表</span>
              <el-button size="small" @click="refreshAlerts">刷新状态</el-button>
            </div>
          </template>

          <el-table :data="filteredAlerts" stripe class="alarm-table" v-loading="loading.patients">
            <el-table-column prop="triggered_at" label="时间" width="140">
              <template #default="{ row }">
                {{ formatTime(row.triggered_at) }}
              </template>
            </el-table-column>
            <el-table-column label="患者" width="130">
              <template #default="{ row }">
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
            </el-table-column>
            <el-table-column label="告警类型" min-width="150">
              <template #default="{ row }">
                <span class="alert-type-badge" :class="alertTypeClass(row.alert_type)">
                  {{ row.alert_type || row.detail?.slice(0, 20) || '—' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="等级" width="80">
              <template #default="{ row }">
                <el-tag :type="severityTag(row.severity)" size="small" effect="light">
                  {{ severityLabel(row.severity) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="详情" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.detail || '—' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="viewPatientLocation(row)">查看定位</el-button>
                <el-button v-if="row.status === 'pending'" link type="success" size="small" @click="acknowledgeAlert(row.id)">确认</el-button>
                <el-button v-if="row.status !== 'resolved'" link type="warning" size="small" @click="resolveAlert(row.id)">解决</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- Rule Engine Status (1/3 width) -->
      <el-col :span="8">
        <el-card shadow="never" class="content-panel rule-panel">
          <template #header>
            <div class="panel-header">
              <span class="panel-title">规则引擎状态</span>
            </div>
          </template>
          <div class="rule-list">
            <div
              v-for="rule in ruleStatusList" :key="rule.code"
              class="rule-item"
              :class="'risk-' + rule.riskLevel"
            >
              <div>
                <div class="rule-name">{{ rule.name }}</div>
                <div class="rule-desc">{{ rule.desc }}</div>
              </div>
              <span class="rule-trigger-count" :style="{ color: rule.triggerColor }">
                {{ rule.triggerText }}
              </span>
            </div>
          </div>
        </el-card>

        <!-- Department Distribution -->
        <el-card shadow="never" class="content-panel dept-panel" style="margin-top: 20px;">
          <template #header>
            <div class="panel-header">
              <span class="panel-title">今日科室分布</span>
            </div>
          </template>
          <div class="dept-list">
            <div v-for="dept in departmentStats" :key="dept.name" class="dept-item">
              <div class="dept-name">{{ dept.name }}</div>
              <div class="dept-bar-wrap">
                <div class="dept-bar" :style="{ width: dept.barWidth + '%' }"></div>
              </div>
              <div class="dept-count">{{ dept.count }} 人</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Tabs for Patient List, Audit Trail, Rules Config, Compliance -->
    <el-card shadow="never" style="margin-top: 20px;">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="在院患者列表" name="patients">
          <el-table :data="patientList" v-loading="loading.patients" stripe class="patient-table">
            <el-table-column prop="name" label="姓名" width="100">
              <template #default="{ row }">
                <div class="patient-cell">
                  <div class="patient-avatar avatar-blue">{{ row.name[0] }}</div>
                  <span>{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="admission_no" label="住院号" width="140">
              <template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template>
            </el-table-column>
            <el-table-column prop="department" label="科室" width="120">
              <template #default="{ row }">
                <span class="dept-badge">{{ row.department }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="bed_number" label="床号" width="80" />
            <el-table-column prop="last_verify" label="最后核验" width="180">
              <template #default="{ row }">{{ row.last_verify || '未核验' }}</template>
            </el-table-column>
            <el-table-column prop="verify_gap_hours" label="距上次核验(h)" width="120">
              <template #default="{ row }">
                <span class="verify-tag" :class="row.verify_gap_hours > 12 ? 'tag-danger' : row.verify_gap_hours > 6 ? 'tag-warning' : ''">
                  {{ row.verify_gap_hours }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="fence_status" label="围栏状态" width="100">
              <template #default="{ row }">
                <span class="status-badge" :class="row.fence_status === 'inside' ? 'badge-success' : 'badge-danger'">
                  <span class="status-dot" :class="row.fence_status === 'inside' ? 'dot-success' : 'dot-danger'"></span>
                  {{ row.fence_status === 'inside' ? '在院内' : '已越界' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="viewAuditTrail(row.id)">审计追踪</el-button>
                <el-button link type="primary" size="small" @click="viewPatientDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="规则配置" name="rules">
          <el-table :data="ruleConfigs" v-loading="loading.rules" stripe>
            <el-table-column prop="code" label="规则代码" width="100">
              <template #default="{ row }"><span class="mono">{{ row.code }}</span></template>
            </el-table-column>
            <el-table-column prop="name" label="规则名称" width="160" />
            <el-table-column prop="enabled" label="启用" width="80">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" @change="updateRule(row)" />
              </template>
            </el-table-column>
            <el-table-column prop="config" label="配置">
              <template #default="{ row }">
                <el-button size="small" @click="editRuleConfig(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-dialog v-model="showRuleEdit" title="编辑规则配置" width="600px" destroy-on-close>
            <el-form :model="editingRule" label-width="100px">
              <el-form-item label="规则代码"><el-input v-model="editingRule.code" disabled /></el-form-item>
              <el-form-item label="配置(JSON)">
                <el-input v-model="editingRule.configJson" type="textarea" :rows="10" />
              </el-form-item>
            </el-form>
            <template #footer>
              <el-button @click="showRuleEdit = false">取消</el-button>
              <el-button type="primary" @click="saveRuleConfig">保存</el-button>
            </template>
          </el-dialog>
        </el-tab-pane>

        <el-tab-pane label="合规报表" name="compliance">
          <el-row :gutter="16" style="margin-bottom: 16px;">
            <el-col :span="8">
              <el-date-picker v-model="reportDateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 100%;" />
            </el-col>
            <el-col :span="4">
              <el-button type="primary" @click="loadComplianceReport">生成报表</el-button>
            </el-col>
          </el-row>
          <div v-if="complianceReport" class="compliance-report">
            <el-descriptions title="总体概览" :column="3" border class="report-desc">
              <el-descriptions-item label="期间患者总数">{{ complianceReport.summary.total_patients_period }}</el-descriptions-item>
              <el-descriptions-item label="平均住院天数">{{ complianceReport.summary.avg_stay_days }}</el-descriptions-item>
              <el-descriptions-item label="合规率">{{ complianceReport.summary.compliance_rate }}%</el-descriptions-item>
              <el-descriptions-item label="围栏违规">{{ complianceReport.summary.fence_violations }}</el-descriptions-item>
              <el-descriptions-item label="未核验告警">{{ complianceReport.summary.no_verify_alerts }}</el-descriptions-item>
              <el-descriptions-item label="费用异常">{{ complianceReport.summary.expense_anomalies }}</el-descriptions-item>
            </el-descriptions>
            <h4 style="margin-top: 20px;">科室合规率</h4>
            <el-table :data="complianceReport.department_breakdown" stripe>
              <el-table-column prop="name" label="科室" width="150" />
              <el-table-column prop="total_patients" label="患者数" width="100" />
              <el-table-column prop="alerts" label="告警数" width="100" />
              <el-table-column prop="compliance_rate" label="合规率" width="100">
                <template #default="{ row }">
                  <el-progress :percentage="row.compliance_rate" :format="() => row.compliance_rate + '%'" :stroke-width="10" />
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Patient Detail Side Panel -->
    <div class="side-panel-overlay" :class="{ show: showPatientDetail }" @click="showPatientDetail = false" />
    <div class="side-panel" :class="{ open: showPatientDetail }">
      <div class="panel-header">
        <span class="panel-title">患者详情 — {{ selectedPatient?.name || '' }}</span>
        <button class="panel-close" @click="showPatientDetail = false">&#10005;</button>
      </div>
      <div class="panel-body" v-if="selectedPatient">
        <div class="patient-detail-header">
          <div class="patient-avatar-large avatar-blue">{{ selectedPatient.name?.[0] || '?' }}</div>
          <div>
            <div class="patient-detail-name">{{ selectedPatient.name }}</div>
            <div class="patient-detail-id">住院号: <span class="mono">{{ selectedPatient.admission_no }}</span></div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">基本信息</div>
          <div class="panel-row">
            <span class="panel-label">科室</span>
            <span class="panel-value">{{ selectedPatient.department }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">床号</span>
            <span class="panel-value">{{ selectedPatient.bed_number }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">最后核验</span>
            <span class="panel-value">{{ selectedPatient.last_verify || '未核验' }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">围栏状态</span>
            <span class="panel-value">
              <span class="status-badge" :class="selectedPatient.fence_status === 'inside' ? 'badge-success' : 'badge-danger'">
                <span class="status-dot" :class="selectedPatient.fence_status === 'inside' ? 'dot-success' : 'dot-danger'"></span>
                {{ selectedPatient.fence_status === 'inside' ? '在院内' : '已越界' }}
              </span>
            </span>
          </div>
        </div>

        <div class="info-section" v-if="selectedPatient.alert_tags?.length">
          <div class="section-title">告警标签</div>
          <div class="alert-tags-wrap">
            <el-tag v-for="tag in selectedPatient.alert_tags" :key="tag" size="small" type="warning" effect="light" style="margin: 2px 4px 2px 0;">{{ tag }}</el-tag>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Download, Search } from '@element-plus/icons-vue'
import { regulatoryApi, type RegulatoryAlert, type RuleConfig, type ComplianceReport } from '@/api/regulatory'

const activeTab = ref('patients')

// Filters
const filters = ref({
  department: '',
  severity: '',
  search: '',
})

// Departments
const departments = ref<string[]>(['心内科', '康复科', '老年病科', '神经内科'])

// Overview stats
const overview = ref({
  total_patients: 1248,
  wearable_count: 1180,
  today_alerts: 8,
  rule_triggers: 24,
})

const todayAdmissions = ref(12)
const offlineDevices = ref(3)
const fenceViolations = ref(3)
const autoHandleRate = ref(92)

// Alerts data
const alerts = ref<RegulatoryAlert[]>([])

const filteredAlerts = computed(() => {
  let list = alerts.value
  if (filters.value.severity) {
    list = list.filter(a => a.severity === filters.value.severity)
  }
  if (filters.value.search) {
    const q = filters.value.search.toLowerCase()
    list = list.filter(a =>
      (a.patient_name || '').toLowerCase().includes(q) ||
      (a.patient_id || '').toLowerCase().includes(q)
    )
  }
  return list
})

// Patient list
const patientList = ref<any[]>([])
const loading = ref({ patients: false, rules: false })
const showPatientDetail = ref(false)
const selectedPatient = ref<any>(null)

// Rule configs
const ruleConfigs = ref<RuleConfig[]>([])
const showRuleEdit = ref(false)
const editingRule = ref<Partial<RuleConfig>>({})
const editingRuleConfigJson = ref('')

// Compliance report
const reportDateRange = ref<[Date, Date] | null>(null)
const complianceReport = ref<ComplianceReport | null>(null)

// Rule engine status list (hardcoded from prototype)
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

// Helpers
function formatTime(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
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

// Actions
function handleSearch() {
  // Triggered by filter change; data is reactive via computed
}

function handleResetFilters() {
  filters.value = { department: '', severity: '', search: '' }
}

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
  } finally {
    loading.value.patients = false
  }
}

async function loadAlerts() {
  try {
    const params: Record<string, any> = {}
    if (filters.value.severity) params.severity = filters.value.severity
    const res = await regulatoryApi.listAlerts(params)
    alerts.value = res.data?.data || []
  } catch {
    // Mock data for demo
    alerts.value = [
      { id: '1', patient_name: '李秀英', patient_id: '8842', department: '心内科', alert_type: '电子围栏越界', severity: 'high', detail: '离开病区范围 50m', status: 'pending', triggered_at: new Date().toISOString() },
      { id: '2', patient_name: '王建国', patient_id: '7731', department: '康复科', alert_type: '心率异常', severity: 'medium', detail: '持续心率 > 110bpm', status: 'pending', triggered_at: new Date(Date.now() - 300000).toISOString() },
      { id: '3', patient_name: '赵淑华', patient_id: '9921', department: '老年病科', alert_type: '跌倒检测', severity: 'high', detail: 'IMU检测到剧烈震动', status: 'acknowledged', triggered_at: new Date(Date.now() - 900000).toISOString() },
      { id: '4', patient_name: '陈志强', patient_id: '6654', department: '神经内科', alert_type: '用药提醒漏服', severity: 'low', detail: '早餐药未确认服用', status: 'pending', triggered_at: new Date(Date.now() - 1800000).toISOString() },
      { id: '5', patient_name: '刘美兰', patient_id: '5523', department: '心内科', alert_type: '夜间离床超时', severity: 'medium', detail: '离床超过 15分钟', status: 'acknowledged', triggered_at: new Date(Date.now() - 3600000).toISOString() },
    ]
  }
}

async function acknowledgeAlert(id: string) {
  try {
    await regulatoryApi.acknowledgeAlert(id, 'current-user')
    ElMessage.success('已确认')
    await loadAlerts()
  } catch {
    ElMessage.success('已确认（模拟）')
    await loadAlerts()
  }
}

async function resolveAlert(id: string) {
  try {
    await regulatoryApi.resolveAlert(id, 'current-user', '已核实处理')
    ElMessage.success('已解决')
    await loadAlerts()
  } catch {
    ElMessage.success('已解决（模拟）')
    await loadAlerts()
  }
}

function viewPatientLocation(alert: any) {
  ElMessage.info(`查看 ${alert.patient_name || alert.patient_id} 的实时定位`)
}

function viewPatientDetail(patient: any) {
  selectedPatient.value = patient
  showPatientDetail.value = true
}

async function viewAuditTrail(patientId: string) {
  ElMessage.info(`审计追踪: ${patientId}`)
}

async function loadRuleConfigs() {
  loading.value.rules = true
  try {
    const res = await regulatoryApi.listRuleConfigs()
    ruleConfigs.value = res.data?.data || []
  } finally {
    loading.value.rules = false
  }
}

function editRuleConfig(row: RuleConfig) {
  editingRule.value = { ...row }
  editingRuleConfigJson.value = JSON.stringify(row.config || {}, null, 2)
  showRuleEdit.value = true
}

async function updateRule(row: RuleConfig) {
  try {
    await regulatoryApi.updateRuleConfig(row.code, row.config || {})
    ElMessage.success('更新成功')
  } catch {
    ElMessage.error('更新失败')
  }
}

async function saveRuleConfig() {
  try {
    const config = JSON.parse(editingRuleConfigJson.value)
    await regulatoryApi.updateRuleConfig(editingRule.value.code!, config)
    showRuleEdit.value = false
    ElMessage.success('保存成功')
    await loadRuleConfigs()
  } catch (e: any) {
    ElMessage.error(e.message || 'JSON 解析失败')
  }
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
  } catch {
    ElMessage.info('报表生成中（模拟）')
  }
}

function exportReport() {
  ElMessage.info('导出功能开发中...')
}

onMounted(async () => {
  await Promise.all([loadOverview(), loadPatients(), loadAlerts(), loadRuleConfigs()])
})
</script>

<style scoped>
.regulatory-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

/* KPI Cards */
.kpi-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-radius: 14px;
}
.kpi-icon-wrap {
  font-size: 24px;
  margin-bottom: 8px;
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
.kpi-trend {
  font-size: 11px;
  margin-top: 4px;
}
.trend-up { color: #16A34A; }
.trend-down { color: #EF4444; }

.kpi-blue .kpi-value { color: #2563EB; }
.kpi-green .kpi-value { color: #16A34A; }
.kpi-danger .kpi-value { color: #EF4444; }
.kpi-purple .kpi-value { color: #7C3AED; }

/* Filter Card */
.filter-card :deep(.el-card__body) {
  padding: 12px 16px;
}

/* Content Panels */
.content-panel :deep(.el-card__header) {
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  border-left: 3px solid #2563EB;
  padding-left: 8px;
}

/* Alarm Table */
.alarm-table {
  font-size: 14px;
}
.alarm-table :deep(.el-table__header th) {
  background: #fafafa;
  color: var(--el-text-color-secondary);
  font-weight: 600;
}
.alarm-table :deep(.el-table__row:hover) {
  background: #f5f7fa;
}

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
.patient-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.patient-id {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

/* Alert type badges */
.alert-type-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}
.badge-danger { background: #FEF2F2; color: #DC2626; }
.badge-warning { background: #FFFBEB; color: #D97706; }
.badge-primary { background: #EFF6FF; color: #2563EB; }
.badge-info { background: #F3F4F6; color: #6B7280; }

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
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }

/* Verify tags */
.verify-tag {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
}
.tag-danger { background: #FEF2F2; color: #DC2626; }
.tag-warning { background: #FFFBEB; color: #D97706; }

/* Dept badge */
.dept-badge {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

/* Mono font */
.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

/* Rule List */
.rule-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.rule-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #fafafa;
  border-radius: 6px;
  border-left: 3px solid transparent;
  transition: all 0.2s;
}
.rule-item:hover {
  background: #f5f7fa;
}
.rule-item.risk-high { border-left-color: #EF4444; }
.rule-item.risk-med { border-left-color: #F59E0B; }
.rule-item.risk-low { border-left-color: #16A34A; }
.rule-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.rule-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}
.rule-trigger-count {
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

/* Department List */
.dept-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.dept-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}
.dept-name {
  width: 70px;
  flex-shrink: 0;
  color: var(--el-text-color-primary);
}
.dept-bar-wrap {
  flex: 1;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}
.dept-bar {
  height: 100%;
  background: linear-gradient(90deg, #2563EB, #7C3AED);
  border-radius: 4px;
  transition: width 0.3s;
}
.dept-count {
  width: 60px;
  text-align: right;
  font-weight: 600;
  color: var(--el-text-color-primary);
  flex-shrink: 0;
}

/* Compliance Report */
.compliance-report h4 {
  margin-top: 20px;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}
.report-desc :deep(.el-descriptions__label) {
  font-weight: 600;
}

/* Tab styles */
:deep(.el-tabs--border-card) {
  border: none;
  box-shadow: none;
}
:deep(.el-tabs--border-card > .el-tabs__header) {
  background: #fafafa;
  border-bottom: 1px solid #e8e8e8;
}

/* ========== Patient Detail Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show {
  display: block;
}
.side-panel {
  position: fixed;
  top: 0;
  right: -520px;
  bottom: 0;
  width: 520px;
  background: white;
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(0,0,0,0.1);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: white;
  z-index: 1;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: var(--el-fill-color-light);
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.panel-close:hover {
  background: var(--el-border-color-light);
}
.panel-body {
  padding: 20px 24px;
}

.patient-detail-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.patient-avatar-large {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  flex-shrink: 0;
}
.patient-detail-name {
  font-size: 17px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.patient-detail-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.info-section {
  margin-bottom: 20px;
}
.section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-regular);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}
.panel-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}
.panel-value {
  font-size: 13px;
  color: var(--el-text-color-primary);
  font-weight: 600;
}
</style>
