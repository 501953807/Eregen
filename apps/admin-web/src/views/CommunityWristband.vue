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
import { HopeBtn, HopeStatCard, HopeTabs } from '@/components/hope'

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
  <div class="hope-theme">
    <div class="hope-layout">
      <!-- Page Header -->
      <div class="hope-page-header">
        <div>
          <h1 class="hope-page-header__title">社区老人腕带管理</h1>
          <p class="hope-page-header__subtitle">管理社区老人档案、福利标签、签到记录和民政数据同步</p>
        </div>
        <div class="hope-page-header__actions">
          <HopeBtn variant="plain" size="sm">
            <template #icon>
              <el-icon :size="16"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg></el-icon>
            </template>
            导出
          </HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="showAddElder = true">
            <template #icon>
              <el-icon :size="16"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg></el-icon>
            </template>
            新增老人
          </HopeBtn>
        </div>
      </div>

      <!-- KPI Stats -->
      <el-row :gutter="12" style="margin-bottom: 16px;">
        <el-col :span="4">
          <HopeStatCard :value="stats.total_elders" label="登记老人" icon-color="primary" gradient="linear-gradient(135deg, #3a57e8, #6f42c1)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
        <el-col :span="4">
          <HopeStatCard :value="stats.active_devices" label="在线腕带" icon-color="success" gradient="linear-gradient(135deg, #1aa053, #2D5AA0)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
        <el-col :span="4">
          <HopeStatCard :value="stats.welfare_tags_count" label="福利标签" icon-color="accent" gradient="linear-gradient(135deg, #8C57FF, #5E3BB3)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L5 14.14l5-5a2 2 0 012.83 2.83l-2.83 2.83 7.17 7.17a2 2 0 002.83 0l4.24-4.24a2 2 0 000-2.83l-5-5a2 2 0 00-2.83 2.83z"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
        <el-col :span="4">
          <HopeStatCard :value="todaySignin" label="今日签到" icon-color="success" gradient="linear-gradient(135deg, #1aa053, #2D5AA0)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 11-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
        <el-col :span="4">
          <HopeStatCard :value="minzhengPending" label="待审核民政" icon-color="warning" gradient="linear-gradient(135deg, #FAA938, #D97706)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
        <el-col :span="4">
          <HopeStatCard :value="pendingAlerts" label="异常告警" icon-color="error" gradient="linear-gradient(135deg, #c03221, #8B2020)">
            <template #icon><el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg></el-icon></template>
          </HopeStatCard>
        </el-col>
      </el-row>

      <!-- Hope Tabs -->
      <div class="hope-content-card" style="margin-bottom: 16px;">
        <div class="hope-content-card__body" style="padding: 0;">
          <HopeTabs
            :model-value="activeTab"
            :tabs="[
              { label: '老人管理', value: 'elders' },
              { label: '福利标签', value: 'welfare' },
              { label: '签到总览', value: 'signin' },
              { label: '药房发药', value: 'pharmacy' },
              { label: '民政数据', value: 'minzheng' },
              { label: '批量发放', value: 'payments' },
            ]"
            animated
            @update:model-value="activeTab = $event as string"
          />
          <div style="padding: 20px 22px;">
            <el-tabs v-model="activeTab" type="border-card" style="display: none;">
              <!-- Hidden el-tabs to preserve event handling -->
            </el-tabs>

            <!-- Tab Content -->
            <div v-if="activeTab === 'elders'">
              <ElderList :rows="filteredElders" :loading="loading.elders" @add="showAddElder = true" @detail="viewElderDetail" @edit="editElder" />
            </div>
            <div v-else-if="activeTab === 'welfare'">
              <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @view-tag-elders="viewTagElders" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" @add-tag="showAddElder = true" />
            </div>
            <div v-else-if="activeTab === 'signin'">
              <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
            </div>
            <div v-else-if="activeTab === 'pharmacy'">
              <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
            </div>
            <div v-else-if="activeTab === 'minzheng'">
              <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
            </div>
            <div v-else-if="activeTab === 'payments'">
              <TabPanels :active-tab="activeTab" :elders="elders" :loading="loading" @execute-payment="executeBatchPayment" @file-upload="onFileUpload" />
            </div>
          </div>
        </div>
      </div>

      <!-- Dialogs -->
      <ElderFormDialog v-model="showAddElder" :editing="!!editingElder" :initial-form="elderForm" @save="saveElder" />
      <ElderDetailPanel v-model="showDetailDialog" :row="detailElder" />
    </div>
  </div>
</template>

<style scoped>
/* No additional custom styles needed — all Hope UI classes handled by the component library */
</style>
