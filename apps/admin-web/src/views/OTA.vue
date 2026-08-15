<template>
  <div class="ota-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">OTA 固件管理</h2>
        <p class="page-subtitle">管理手环和药盒固件版本 · 发布 OTA 推送任务</p>
      </div>
      <HopeBtn variant="filled" size="md" @click="showCreateDialog = true">
        <template #icon>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </template>
        创建固件版本
      </HopeBtn>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="firmwares.length"
        label="固件版本"
        icon-color="primary"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 15 21 19 17 23"/><polyline points="7 10 3 10 3 14"/><path d="M21 3l-9 9-9-9"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="bracelets"
        label="手环设备"
        icon-color="success"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="pillboxes"
        label="药盒设备"
        icon-color="accent"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M10.5 1.5H3.5A2.5 2.5 0 001 4v4a2.5 2.5 0 002.5 2.5h7A2.5 2.5 0 0013 8V4a2.5 2.5 0 00-2.5-2.5z"/><path d="M13.5 1.5h7A2.5 2.5 0 0123 4v4a2.5 2.5 0 01-2.5 2.5h-7A2.5 2.5 0 0111 8V4a2.5 2.5 0 012.5-2.5z"/><line x1="12" y1="8" x2="12" y2="22"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="activeJobs"
        label="活跃任务"
        icon-color="warning"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
        </template>
      </HopeStatCard>
    </div>

    <!-- Firmware Table — HopeCard + HopeTable -->
    <HopeCard title="固件版本列表">
      <HopeTable
        :columns="tableColumns"
        :data="firmwares"
        :loading="loading"
        striped
        :row-key="(row: FirmwareRelease) => row.id"
      >
        <template #col-type="{ row }">
          <HopeBadge :color="row.type === 'bracelet' ? 'success' : 'accent'">
            {{ row.type === 'bracelet' ? '手环' : '药盒' }}
          </HopeBadge>
        </template>
        <template #col-tier="{ row }">
          <span class="tier-tag" :class="tierClass(row.tier)">{{ tierLabel(row.tier) }}</span>
        </template>
        <template #col-version="{ row }">
          <span class="version-tag" :class="{ outdated: !isLatest(row) }">{{ row.version }}</span>
          <span v-if="isLatest(row)" class="latest-dot" title="最新版本">✓</span>
        </template>
        <template #col-sha256="{ row }">
          <span class="mono">{{ row.sha256_hash?.slice(0, 12) + '…' || '—' }}</span>
        </template>
        <template #col-created_at="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
        <template #col-actions="{ row }">
          <div class="action-group">
            <HopeBtn variant="text" size="sm" @click.stop="handlePush(row)">推送升级</HopeBtn>
            <HopeBtn variant="text" size="sm" @click.stop="handleVerify(row)" :loading="verifyingId === row.id">验证签名</HopeBtn>
            <HopeBtn v-if="jobMap[row.id]?.length" variant="text" size="sm" @click.stop="handleShowJobs(row.id)">查看进度</HopeBtn>
          </div>
        </template>
      </HopeTable>
    </HopeCard>

    <!-- Create Firmware Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建固件版本" width="550px" class="hope-dialog" destroy-on-close>
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
        <HopeBtn variant="plain" @click="showCreateDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="handleCreateFirmware" :loading="creating">创建</HopeBtn>
      </template>
    </el-dialog>

    <!-- Push OTA Dialog -->
    <el-dialog v-model="showPushDialog" title="推送OTA升级" width="550px" class="hope-dialog" destroy-on-close>
      <p style="margin-bottom: 16px; color: var(--hope-text-secondary); font-size: 14px;">
        目标固件: <strong style="color: var(--hope-text);">{{ selectedFirmware?.version }}</strong>
        ({{ deviceTypeLabel(selectedFirmware?.type ?? '') }}/{{ tierLabel(selectedFirmware?.tier ?? '') }})
      </p>
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
        <HopeBtn variant="plain" @click="showPushDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="handlePushOTA" :loading="pushing">确认推送</HopeBtn>
      </template>
    </el-dialog>

    <!-- Job Progress Side Panel -->
    <div class="side-panel-overlay" :class="{ show: showJobPanel }" @click="showJobPanel = false" />
    <div class="side-panel" :class="{ open: showJobPanel }">
      <div class="panel-header">
        <div>
          <div class="panel-title">推送进度</div>
          <div class="panel-subtitle">{{ selectedFirmware?.version }}</div>
        </div>
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
          <el-descriptions-item label="成功"><span style="color:var(--hope-success);font-weight:700;">{{ currentJob.progress.succeeded }}</span></el-descriptions-item>
          <el-descriptions-item label="失败"><span style="color:var(--hope-danger);font-weight:700;">{{ currentJob.progress.failed }}</span></el-descriptions-item>
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
            class="hope-progress"
          />
        </div>

        <div class="job-actions">
          <HopeBtn variant="error" size="sm" @click="cancelJob">取消任务</HopeBtn>
          <HopeBtn variant="plain" size="sm" @click="refreshJob">刷新状态</HopeBtn>
        </div>
      </div>
      <div v-else class="panel-empty">
        <el-icon :size="28" style="color:var(--hope-text-muted);margin-bottom:8px;display:block;"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 12a9 9 0 11-6.219-8.56"/><polyline points="21 3 21 9 15 9"/></svg></el-icon>
        加载中...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { otaApi, type FirmwareRelease, type OTAJob, type CreateFirmwareRequest } from '@/api/ota'
import HopeCard from '@/components/hope/HopeCard.vue'
import HopeBtn from '@/components/hope/HopeBtn.vue'
import HopeTable from '@/components/hope/HopeTable.vue'
import HopeStatCard from '@/components/hope/HopeStatCard.vue'
import HopeBadge from '@/components/hope/HopeBadge.vue'

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

/* ---------- HopeTable columns ---------- */

const tableColumns = [
  { prop: 'type', label: '设备类型' },
  { prop: 'tier', label: '档位' },
  { prop: 'version', label: '版本号' },
  { prop: 'changelog', label: '更新说明' },
  { prop: 'sha256_hash', label: 'SHA256' },
  { prop: 'created_at', label: '创建时间' },
  { prop: 'actions', label: '操作' },
]

/* ---------- Computed ---------- */

const bracelets = computed(() => firmwares.value.filter(f => f.type === 'bracelet').length)
const pillboxes = computed(() => firmwares.value.filter(f => f.type === 'pillbox').length)
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
/* ─── Page layout ─── */
.ota-page {
  padding: 0;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
}
.page-header__left { display: flex; flex-direction: column; gap: 4px; }

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0;
  letter-spacing: -0.02em;
}
.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
  font-weight: 500;
}

/* ─── KPI Grid ─── */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

/* ─── Table section ─── */
.hope-content-card__body { padding: 0 !important; }
.hope-content-card__header { padding: 20px 22px 0 !important; }

/* ─── Tier tags ─── */
.tier-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
}
.tier-pro  { background: var(--hope-primary-light); color: var(--hope-primary); }
.tier-plus { background: var(--hope-success-light); color: var(--hope-success); }
.tier-basic { background: rgba(148,169,162,0.12); color: var(--hope-text-muted); }

/* ─── Version tag ─── */
.version-tag {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(148,169,162,0.10);
  color: var(--hope-text-secondary);
}
.version-tag.outdated {
  background: var(--hope-warning-light);
  color: #B8860B;
}
.latest-dot {
  display: inline-block;
  margin-left: 4px;
  font-size: 11px;
  color: var(--hope-success);
  font-weight: 700;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--hope-text-muted);
}

/* ─── Action group ─── */
.action-group {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-wrap: wrap;
}

/* ─── Hope progress ─── */
.hope-progress {
  :deep(.el-progress-bar__outer) {
    border-radius: 6px;
    background: var(--hope-border);
  }
  :deep(.el-progress-bar__inner) {
    border-radius: 6px;
    background: var(--hope-primary);
  }
}

/* ─── Job Side Panel ─── */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.3);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show { display: block; }

.side-panel {
  position: fixed;
  top: 0;
  right: -540px;
  bottom: 0;
  width: 540px;
  background: var(--hope-surface);
  z-index: 201;
  transition: right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow-y: auto;
  box-shadow: -12px 0 40px rgba(17,38,146,0.10);
}
.side-panel.open { right: 0; }

.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: var(--hope-surface);
  z-index: 1;
  gap: 12px;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
  margin: 0;
}
.panel-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin-top: 2px;
  font-family: 'SF Mono', 'Consolas', monospace;
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: var(--hope-radius-md);
  border: 1px solid var(--hope-border);
  background: var(--hope-bg);
  cursor: pointer;
  font-size: 16px;
  color: var(--hope-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex-shrink: 0;
}
.panel-close:hover {
  background: var(--hope-border);
  color: var(--hope-text);
}

.panel-body { padding: 20px 24px; }
.panel-empty {
  padding: 60px 24px;
  text-align: center;
  color: var(--hope-text-muted);
  font-size: 14px;
}

.job-info { margin-bottom: 16px; }
.job-id {
  font-size: 12px;
  color: var(--hope-text-muted);
}

.job-desc {
  :deep(.el-descriptions) {
    border-radius: var(--hope-radius-md);
    border: 1px solid var(--hope-border);
  }
  :deep(.el-descriptions__label) {
    width: 90px;
    font-weight: 600;
    color: var(--hope-text-secondary);
    background: var(--hope-surface-light);
  }
}

.progress-section { margin: 20px 0; }
.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--hope-text-secondary);
}
.progress-pct {
  font-size: 14px;
  color: var(--hope-primary);
  font-weight: 700;
}

.job-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--hope-border);
}

/* ─── Dialog (hope-dialog class applied via class attr) ─── */
:deep(.hope-dialog .el-dialog) {
  border-radius: var(--hope-radius-xl) !important;
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-lg) !important;
}
:deep(.hope-dialog .el-dialog__header) {
  padding: 20px 24px 16px !important;
  border-bottom: 1px solid var(--hope-border) !important;
  margin-right: 0 !important;
}
:deep(.hope-dialog .el-dialog__title) {
  font-size: 16px !important;
  font-weight: 700 !important;
  color: var(--hope-text) !important;
}
:deep(.hope-dialog .el-dialog__body) { padding: 20px 24px !important; }
:deep(.hope-dialog .el-dialog__footer) {
  padding: 16px 24px 20px !important;
  border-top: 1px solid var(--hope-border) !important;
}
:deep(.hope-dialog .el-form-item__label) {
  font-weight: 600 !important;
  color: var(--hope-text-secondary) !important;
}
:deep(.hope-dialog .el-input__wrapper),
:deep(.hope-dialog .el-select .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
}

/* ─── Responsive ─── */
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .page-header { flex-direction: column; }
  .kpi-grid { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .side-panel { width: 100%; right: -100%; }
}
</style>
