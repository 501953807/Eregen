<template>
  <div class="ota-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">OTA 固件管理</h2>
      <el-button type="primary" @click="showCreateDialog = true" size="default">+ 创建固件版本</el-button>
    </div>

    <!-- KPI Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ firmwares.length }}</div>
          <div class="kpi-label">固件版本</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ bracelets }}</div>
          <div class="kpi-label">手环设备</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ pillboxes }}</div>
          <div class="kpi-label">药盒设备</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-value">{{ activeJobs }}</div>
          <div class="kpi-label">活跃任务</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Firmware Table -->
    <el-card shadow="never" class="table-card">
      <template #header><span class="section-title">固件版本列表</span></template>
      <el-table
        :data="firmwares"
        stripe
        class="ota-table"
        v-loading="loading"
      >
        <el-table-column label="设备类型" width="100">
          <template #default="{ row }">
            <span class="device-type-badge" :class="row.device_type === 'bracelet' ? 'badge-bracelet' : 'badge-pillbox'">
              {{ row.device_type === 'bracelet' ? '📱' : '💊' }}
              <span>{{ deviceTypeLabel(row.device_type) }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="档位" width="90">
          <template #default="{ row }">
            <span class="tier-tag" :class="tierClass(row.tier)">{{ tierLabel(row.tier) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="版本号" width="130">
          <template #default="{ row }">
            <span class="version-tag" :class="{ outdated: !isLatest(row) }">{{ row.version }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="changelog" label="更新说明" min-width="180" show-overflow-tooltip />
        <el-table-column label="SHA256" width="100">
          <template #default="{ row }">
            <span class="mono">{{ row.sha256_hash?.slice(0, 12) + '…' || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="handlePush(row)">推送升级</el-button>
            <el-button link type="info" size="small" @click.stop="handleVerify(row)" :loading="verifyingId === row.id">验证签名</el-button>
            <el-button link type="primary" size="small" @click.stop="handleShowJobs(row.id)" v-if="jobMap[row.id]?.length">查看进度</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create Firmware Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建固件版本" width="550px" destroy-on-close>
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="设备类型" required>
          <el-select v-model="createForm.device_type" style="width: 100%;">
            <el-option label="手环" value="bracelet" />
            <el-option label="药盒" value="pillbox" />
          </el-select>
        </el-form-item>
        <el-form-item label="档位" required>
          <el-select v-model="createForm.tier" style="width: 100%;">
            <el-option label="入门版" value="starter" />
            <el-option label="中端版" value="plus" />
            <el-option label="高端版" value="pro" />
            <el-option label="基础版" value="basic" />
            <el-option label="智能版" value="smart" />
            <el-option label="自动版" value="auto" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本号" required>
          <el-input v-model="createForm.version" placeholder="如: v2.2.0" />
        </el-form-item>
        <el-form-item label="下载 URL" required>
          <el-input v-model="createForm.url" placeholder="https://cdn.example.com/firmware.bin" />
        </el-form-item>
        <el-form-item label="SHA256 Hash" required>
          <el-input v-model="createForm.sha256_hash" placeholder="64位十六进制哈希值" />
        </el-form-item>
        <el-form-item label="更新说明">
          <el-input v-model="createForm.changelog" type="textarea" :rows="3" placeholder="描述本次更新内容" />
        </el-form-item>
        <el-form-item label="强制更新">
          <el-switch v-model="createForm.force_update" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateFirmware" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- Push OTA Dialog -->
    <el-dialog v-model="showPushDialog" title="推送OTA升级" width="550px" destroy-on-close>
      <p style="margin-bottom: 12px;">目标固件: <strong>{{ selectedFirmware?.version }}</strong> ({{ deviceTypeLabel(selectedFirmware?.device_type ?? '') }}/{{ tierLabel(selectedFirmware?.tier ?? '') }})</p>
      <el-form :model="pushForm" label-width="100px">
        <el-form-item label="目标设备">
          <el-radio-group v-model="pushForm.mode">
            <el-radio label="all">全量推送（所有匹配设备）</el-radio>
            <el-radio label="manual">指定设备</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="pushForm.mode === 'manual'" label="设备ID列表">
          <el-input
            v-model="pushForm.deviceIdsStr"
            type="textarea"
            :rows="4"
            placeholder="每行一个设备ID，如：BR-0001"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="handlePushOTA" :loading="pushing">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Job Progress Side Panel -->
    <div class="side-panel-overlay" :class="{ show: showJobPanel }" @click="showJobPanel = false" />
    <div class="side-panel" :class="{ open: showJobPanel }">
      <div class="panel-header">
        <span class="panel-title">推送进度 — {{ selectedFirmware?.version }}</span>
        <button class="panel-close" @click="showJobPanel = false">&#10005;</button>
      </div>
      <div class="panel-body" v-if="currentJob">
        <div class="job-info">
          <div class="job-id">任务ID: <span class="mono">{{ currentJob.id }}</span></div>
        </div>

        <el-descriptions :column="2" border class="job-desc">
          <el-descriptions-item label="固件版本">{{ selectedFirmware?.version }}</el-descriptions-item>
          <el-descriptions-item label="总数">{{ currentJob.progress.total }}</el-descriptions-item>
          <el-descriptions-item label="已推送">{{ currentJob.progress.succeeded + currentJob.progress.failed }}</el-descriptions-item>
          <el-descriptions-item label="下载中">{{ currentJob.progress.downloading }}</el-descriptions-item>
          <el-descriptions-item label="待推送">{{ currentJob.progress.pending }}</el-descriptions-item>
          <el-descriptions-item label="成功"><span style="color:#16A34A;font-weight:700;">{{ currentJob.progress.succeeded }}</span></el-descriptions-item>
          <el-descriptions-item label="失败"><span style="color:#EF4444;font-weight:700;">{{ currentJob.progress.failed }}</span></el-descriptions-item>
        </el-descriptions>

        <div class="progress-section">
          <div class="progress-label">
            <span>整体进度</span>
            <span class="progress-pct">{{ progressPct }}%</span>
          </div>
          <el-progress
            :percentage="progressPct"
            :status="progressStatus"
            :stroke-width="12"
            :show-text="false"
          />
        </div>

        <div class="job-actions">
          <el-button size="small" type="danger" plain @click="cancelJob">取消任务</el-button>
          <el-button size="small" @click="refreshJob">刷新状态</el-button>
        </div>
      </div>
      <div v-else class="panel-empty">加载中...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { otaApi, type FirmwareRelease, type OTAJob, type CreateFirmwareRequest } from '@/api/ota'

/* ---------- Data ---------- */

const firmwares = ref<FirmwareRelease[]>([])
const jobMap = ref<Record<string, OTAJob[]>>({})
const showJobPanel = ref(false)
const verifyingId = ref('')
const creating = ref(false)
const pushing = ref(false)
const loading = ref(false)

const selectedFirmware = ref<FirmwareRelease | null>(null)
const currentJob = ref<OTAJob | null>(null)

/* ---------- Create Form ---------- */

const showCreateDialog = ref(false)
const createForm = ref<Partial<CreateFirmwareRequest>>({
  device_type: 'bracelet',
  tier: 'starter',
  version: '',
  url: '',
  sha256_hash: '',
  changelog: '',
  force_update: false,
})

/* ---------- Push Form ---------- */

const showPushDialog = ref(false)
const pushForm = ref({ mode: 'all', deviceIdsStr: '' })

/* ---------- Computed ---------- */

const bracelets = computed(() => firmwares.value.filter(f => f.device_type === 'bracelet').length)
const pillboxes = computed(() => firmwares.value.filter(f => f.device_type === 'pillbox').length)
const activeJobs = computed(() => Object.values(jobMap.value).flat().filter(j => {
  const p = j.progress
  return p.pending > 0 || p.downloading > 0
}).length)

const progressPct = computed(() => {
  if (!currentJob.value) return 0
  const p = currentJob.value.progress
  return Math.round(((p.succeeded + p.failed) / Math.max(p.total, 1)) * 100)
})

const progressStatus = computed(() => {
  if (!currentJob.value) return undefined
  const p = currentJob.value.progress
  if (p.failed >= p.total) return 'exception'
  if (p.succeeded + p.failed >= p.total) return 'success'
  return undefined
})

/* ---------- Lifecycle ---------- */

onMounted(() => {
  loadFirmwares()
})

async function loadFirmwares() {
  loading.value = true
  try {
    const res = await otaApi.listFirmware()
    firmwares.value = res.data?.data || []
  } catch {
    ElMessage.error('加载固件列表失败')
  } finally {
    loading.value = false
  }
}

/* ---------- Create ---------- */

async function handleCreateFirmware() {
  if (!createForm.value.version || !createForm.value.url || !createForm.value.sha256_hash) {
    ElMessage.warning('请填写必填项')
    return
  }
  creating.value = true
  try {
    await otaApi.createFirmware(createForm.value as CreateFirmwareRequest)
    ElMessage.success('固件版本创建成功')
    showCreateDialog.value = false
    await loadFirmwares()
    createForm.value = {
      device_type: 'bracelet',
      tier: 'starter',
      version: '',
      url: '',
      sha256_hash: '',
      changelog: '',
      force_update: false,
    }
  } catch {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

/* ---------- Verify ---------- */

async function handleVerify(row: FirmwareRelease) {
  verifyingId.value = row.id
  try {
    const res = await otaApi.verifyFirmware(row.id)
    const d = res.data?.data ?? {}
    ElMessage.success(`验证结果: ${d.status} (valid: ${d.valid})`)
  } catch {
    ElMessage.error('验证失败')
  } finally {
    verifyingId.value = ''
  }
}

/* ---------- Push OTA ---------- */

function handlePush(row: FirmwareRelease) {
  selectedFirmware.value = row
  pushForm.value = { mode: 'all', deviceIdsStr: '' }
  showPushDialog.value = true
}

async function handlePushOTA() {
  if (!selectedFirmware.value) return
  pushing.value = true
  try {
    const deviceIds = pushForm.value.mode === 'manual'
      ? pushForm.value.deviceIdsStr.split('\n').map(s => s.trim()).filter(Boolean)
      : []
    const res = await otaApi.pushOTA({ firmware_id: selectedFirmware.value.id, device_ids: deviceIds })
    const data = res.data?.data ?? {}
    ElMessage.success(`推送已发起，job_id: ${data.job_id}`)
    showPushDialog.value = false
    if (data.job_id) {
      startPolling(data.job_id)
    }
  } catch {
    ElMessage.error('推送失败')
  } finally {
    pushing.value = false
  }
}

/* ---------- Job Progress ---------- */

let pollTimer: ReturnType<typeof setInterval> | null = null

function handleShowJobs(firmwareId: string) {
  selectedFirmware.value = firmwares.value.find(f => f.id === firmwareId) ?? null
  showJobPanel.value = true
  const jobs = jobMap.value[firmwareId] ?? []
  if (jobs.length > 0) {
    currentJob.value = jobs[jobs.length - 1]
    startPolling(currentJob.value.id)
  } else {
    ElMessage.info('暂无推送记录')
  }
}

async function startPolling(jobId: string) {
  stopPolling()
  const fetchOne = async () => {
    try {
      const res = await otaApi.getOTAJob(jobId)
      currentJob.value = res.data?.data ?? null
      if (currentJob.value) {
        const p = currentJob.value.progress
        if (p.succeeded + p.failed >= p.total) {
          stopPolling()
        }
      }
    } catch {
      // ignore polling errors
    }
  }
  await fetchOne()
  pollTimer = setInterval(fetchOne, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function cancelJob() {
  ElMessage.info('取消任务功能开发中...')
}

function refreshJob() {
  if (currentJob.value) startPolling(currentJob.value.id)
}

/* ---------- Helpers ---------- */

function isLatest(fw: FirmwareRelease): boolean {
  return fw.is_latest ?? false
}

function deviceTypeLabel(type: string): string {
  return type === 'bracelet' ? '手环' : '药盒'
}

function tierLabel(tier: string): string {
  const map: Record<string, string> = {
    starter: '入门版', plus: '中端版', pro: '高端版',
    basic: '基础版', smart: '智能版', auto: '自动版',
  }
  return map[tier] || tier
}

function tierClass(tier: string): string {
  const map: Record<string, string> = { starter: 'tier-basic', plus: 'tier-plus', pro: 'tier-pro' }
  return map[tier] || 'tier-basic'
}

function formatDate(ts: string): string {
  return new Date(ts).toLocaleString('zh-CN')
}
</script>

<style scoped>
.ota-page {
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

/* Section title */
.section-title {
  font-size: 15px;
  font-weight: 700;
}

/* Table */
.table-card :deep(.el-card__header) {
  padding: 16px 20px;
}
.ota-table {
  width: 100%;
}

/* Device type badge */
.device-type-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}
.badge-bracelet { background: #DBEAFE; color: #2563EB; }
.badge-pillbox { background: #FCE7F3; color: #EC4899; }

/* Tier tag */
.tier-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
}
.tier-pro { background: #EDE9FE; color: #7C3AED; }
.tier-plus { background: #DBEAFE; color: #2563EB; }
.tier-basic { background: #F3F4F6; color: #6B7280; }

/* Version tag */
.version-tag {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  background: #F3F4F6;
}
.version-tag.outdated {
  background: #FFFBEB;
  color: #D97706;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

/* ========== Job Side Panel ========== */
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
.panel-empty {
  padding: 60px 24px;
  text-align: center;
  color: var(--el-text-color-placeholder);
}

.job-info {
  margin-bottom: 16px;
}
.job-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.job-desc :deep(.el-descriptions__label) {
  width: 100px;
  font-weight: 600;
}

.progress-section {
  margin: 20px 0;
}
.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
}
.progress-pct {
  font-size: 14px;
  color: var(--el-color-primary);
}

.job-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}
</style>
