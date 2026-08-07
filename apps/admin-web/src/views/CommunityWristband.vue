<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { communityApi, type CommunityElder, type CommunityDevice, type WelfareTagConfig, type CommunityPharmacyLog } from '@/api/community'
import { handleApiError, handleApiSuccess } from '@/utils/error'
import ElderList from './CommunityWristband/ElderList.vue'
import ElderDetailPanel from './CommunityWristband/ElderDetailPanel.vue'
import ElderFormDialog from './CommunityWristband/ElderFormDialog.vue'
import TabPanels from './CommunityWristband/TabPanels.vue'

const activeTab = ref('elders')

// KPI stats
const stats = ref({ total_elders: 0, active_devices: 0, welfare_tags_count: 0 })
const todaySignin = ref(0)
const pendingPayments = ref(0)
const minzhengPending = ref(0)
const pendingAlerts = ref(0)

// Elders
const elders = ref<CommunityElder[]>([])
const loading = ref({ elders: false, devices: false, welfare: false, signin: false, payments: false, minzheng: false, pharmacy: false, detail: false, tag_elders: false })

// Detail panel
const showDetailDialog = ref(false)
const detailElder = ref<CommunityElder | null>(null)

// Add/Edit
const showAddElder = ref(false)
const editingElder = ref<CommunityElder | null>(null)
const elderForm = ref<Partial<CommunityElder>>({ gender: 1, status: 'active' })

// Tag elders dialog
const showTagEldersDialog = ref(false)
const selectedTagName = ref('')

// Devices
const devices = ref<CommunityDevice[]>([])

// Welfare
const welfareTags = ref<WelfareTagConfig[]>([])

// Minzheng
const minzhengSyncs = ref<any[]>([])
const signinRecords = ref<any[]>([])

// Pharmacy
const pharmacyLogs = ref<any[]>([])

// Payments
const batchPayments = ref<any[]>([])

const filteredElders = computed(() => elders.value)

onMounted(() => {
  loadElders()
  loadDevices()
  loadWelfareTags()
  loadSigninRecords()
  loadBatchPayments()
  loadMinzhengSync()
  loadPharmacyLogs()
})

function resetElderForm() { elderForm.value = { gender: 1, status: 'active' } }

async function loadElders() {
  loading.value.elders = true
  try {
    const res = await communityApi.listElders({ page: 1, page_size: 50 })
    const list = res.data?.data || []
    elders.value = list
    stats.value.total_elders = list.length
  } finally { loading.value.elders = false }
}

async function saveElder(form: Record<string, any>) {
  try {
    if (editingElder.value?.id) {
      await communityApi.updateElder(editingElder.value.id, form as any)
    ElMessage.success('更新成功')
    } else {
      await communityApi.createElder(form as any)
      ElMessage.success('创建成功')
    }
    showAddElder.value = false
    editingElder.value = null
    resetElderForm()
    await loadElders()
  } catch (e: any) { handleApiError(e) }
}

function editElder(row: CommunityElder) { editingElder.value = row; elderForm.value = { ...row }; showAddElder.value = true }

async function viewElderDetail(row: CommunityElder) { detailElder.value = row; showDetailDialog.value = true }

async function loadDevices() {
  loading.value.devices = true
  try {
    const res = await communityApi.listDevices({ page: 1, page_size: 50 })
    devices.value = res.data?.data || []
    stats.value.active_devices = devices.value.filter(d => d.status === 'active').length
  } finally { loading.value.devices = false }
}

async function loadWelfareTags() {
  loading.value.welfare = true
  try {
    const res = await communityApi.listWelfareTags()
    welfareTags.value = res.data?.data || []
    stats.value.welfare_tags_count = welfareTags.value.length
  } finally { loading.value.welfare = false }
}

async function viewTagElders(tagCode: string, tagName: string) { selectedTagName.value = tagName; showTagEldersDialog.value = true }

async function loadSigninRecords() {
  loading.value.signin = true
  try {
    const period = dayjs().format('YYYY-MM')
    const res = await communityApi.listSigninRecords({ period })
    signinRecords.value = res.data?.data || []
    todaySignin.value = signinRecords.value.filter((r: any) => r.signin_time?.startsWith(dayjs().format('YYYY-MM-DD'))).length
  } finally { loading.value.signin = false }
}

async function executeBatchPayment() {
  try {
    const elderIds = elders.value.filter(e => e.status === 'active').map(e => e.id)
    if (!elderIds.length) { ElMessage.warning('没有可发放的老人'); return }
    await communityApi.executeBatchPayment({ batch_id: 'BATCH-' + Date.now(), period: dayjs().format('YYYY-MM'), pay_type: 'welfare', elder_ids: elderIds })
    ElMessage.success('已提交发放')
    await loadBatchPayments()
  } catch { ElMessage.error('发放失败') }
}

async function loadBatchPayments() {
  loading.value.payments = true
  try {
    const res = await communityApi.listBatchPayments()
    batchPayments.value = res.data?.data || []
  } finally { loading.value.payments = false }
}

async function loadMinzhengSync() {
  loading.value.minzheng = true
  try {
    const res = await communityApi.listMinzhengSync()
    minzhengSyncs.value = res.data?.data || []
    minzhengPending.value = minzhengSyncs.value.reduce((sum: number, s: any) => sum + (s.pending_review_count || 0), 0)
  } finally { loading.value.minzheng = false }
}

async function loadPharmacyLogs() {
  loading.value.pharmacy = true
  try {
    const period = dayjs().format('YYYY-MM')
    const res = await communityApi.listPharmacyLogs({ period })
    pharmacyLogs.value = (res.data?.data || []) as CommunityPharmacyLog[]
  } catch {
    pharmacyLogs.value = []
  } finally {
    loading.value.pharmacy = false
  }
}

function onFileUpload(file: File) { ElMessage.success(`文件 ${file.name} 已选择，正在上传...`) }
</script>

<template>
  <div class="community-page">
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-blue"><div class="kpi-value">{{ stats.total_elders }}</div><div class="kpi-label">登记老人</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">{{ stats.active_devices }}</div><div class="kpi-label">在线腕带</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-purple"><div class="kpi-value">{{ stats.welfare_tags_count }}</div><div class="kpi-label">福利标签</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">{{ todaySignin }}</div><div class="kpi-label">今日签到</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-danger"><div class="kpi-value">{{ minzhengPending }}</div><div class="kpi-label">待审核民政</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-orange"><div class="kpi-value">{{ pendingAlerts }}</div><div class="kpi-label">异常告警</div></el-card></el-col>
    </el-row>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="老人管理" name="elders">
        <ElderList :rows="filteredElders" :loading="loading.elders" @add="showAddElder = true" @detail="viewElderDetail" @edit="editElder" />
      </el-tab-pane>
      <el-tab-pane label="福利标签" name="welfare">
        <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @view-tag-elders="viewTagElders" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
      </el-tab-pane>
      <el-tab-pane label="签到总览" name="signin">
        <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
      </el-tab-pane>
      <el-tab-pane label="药房发药" name="pharmacy">
        <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
      </el-tab-pane>
      <el-tab-pane label="民政数据" name="minzheng">
        <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
      </el-tab-pane>
      <el-tab-pane label="批量发放" name="payments">
        <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
      </el-tab-pane>
    </el-tabs>

    <ElderFormDialog v-model="showAddElder" :editing="!!editingElder" :initial-form="elderForm" @save="saveElder" />
    <ElderDetailPanel v-model="showDetailDialog" :row="detailElder" />
  </div>
</template>

<style scoped>
.community-page { padding: 0; }
.community-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.community-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
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
.kpi-card:hover {
  transform: translateY(-3px);
}
.kpi-card :deep(.el-card__body) { padding: 18px; display: flex; flex-direction: column; align-items: center; text-align: center; border-radius: 14px; }
.kpi-value { font-size: 32px; font-weight: 800; letter-spacing: -0.03em; line-height: 1; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 6px; font-weight: 600; }
.kpi-blue .kpi-value { color: #5C8D73; }
.kpi-green .kpi-value { color: #6FAF8F; }
.kpi-purple .kpi-value { color: #7BAF8C; }
.kpi-warning .kpi-value { color: #F59E0B; }
.kpi-danger .kpi-value { color: #EF4444; }
.kpi-orange .kpi-value { color: #EA580C; }
</style>
