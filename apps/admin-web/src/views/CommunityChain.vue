<template>
  <div class="community-chain-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">社区链管理</h2>
        <p class="page-subtitle">社区老人档案、福利标签与签到结算管理</p>
      </div>
      <div class="page-header__actions">
        <HopeBtn variant="filled" size="md" @click="showCreateDialog = true">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </template>
          新增老人
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="kpis.total_elders"
        label="社区老人"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.today_signin"
        label="今日签到"
        icon-color="success"
        gradient="linear-gradient(135deg, #1aa053 0%, #22c55e 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.welfare_tags"
        label="福利标签"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938 0%, #f59e0b 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.pending_payments"
        label="待结算"
        icon-color="accent"
        gradient="linear-gradient(135deg, #8C57FF 0%, #a78bfa 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg></el-icon>
        </template>
      </HopeStatCard>
    </div>

    <!-- Main Table Card -->
    <HopeCard title="社区老人管理" :subtitle="`共 ${pagination.total} 条记录`">
      <template #header>
        <div class="toolbar">
          <el-form :inline="true" class="filter-form">
            <el-form-item label="状态">
              <el-select v-model="filters.status" placeholder="全部" clearable>
                <el-option label="正常" value="active" />
                <el-option label="停用" value="inactive" />
                <el-option label="已退役" value="retired" />
              </el-select>
            </el-form-item>
            <el-form-item label="搜索">
              <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable style="width:180px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchElders">查询</el-button>
            </el-form-item>
          </el-form>
          <div style="display:flex;gap:8px;">
            <HopeBtn variant="plain" size="sm" @click="refreshAll">
              <template #icon>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
              </template>
              刷新
            </HopeBtn>
          </div>
        </div>
      </template>

      <el-table :data="elders" v-loading="loading" stripe class="hope-table-custom">
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="welfare_tags" label="福利标签" width="200">
          <template #default="{ row }">
            <el-tag v-for="tag in parseTags(row.welfare_tags)" :key="tag" size="small" class="mr-1">{{ tag }}</el-tag>
            <span v-if="!row.welfare_tags" class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="last_signin" label="最后签到" width="140">
          <template #default="{ row }">{{ row.last_signin || '—' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'inactive' ? 'warning' : 'info'" size="small">
              {{ row.status === 'active' ? '正常' : row.status === 'inactive' ? '停用' : '已退役' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="triggerSignin(row)">签到</el-button>
            <el-button link type="primary" @click="assignWelfare(row)">福利</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchElders"
        class="hope-pagination"
      />
    </HopeCard>

    <!-- Create Elder Dialog -->
    <el-dialog v-model="showCreateDialog" title="新增社区老人" width="520px" class="hope-dialog">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="身份证号" required>
          <el-input v-model="createForm.id_card" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="createForm.phone" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="createForm.address" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" size="sm" @click="showCreateDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" size="sm" @click="createElder" :loading="createLoading">创建</HopeBtn>
      </template>
    </el-dialog>

    <!-- Welfare Dialog -->
    <el-dialog v-model="showWelfareDialog" title="福利标签管理" width="480px" class="hope-dialog">
      <div v-if="currentElder">
        <div style="margin-bottom:16px;font-weight:600;color:var(--hope-text)">{{ currentElder.name }} — 福利标签</div>
        <el-table :data="currentWelfareTags" size="small" stripe class="hope-table-custom">
          <el-table-column prop="tag_code" label="标签代码" width="150" />
          <el-table-column prop="valid_from" label="生效日期" width="120" />
          <el-table-column prop="valid_to" label="到期日期" width="120" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="revokeTag(row.tag_code)">撤销</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <HopeBtn variant="plain" size="sm" @click="showWelfareDialog = false">关闭</HopeBtn>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { communityApi } from '@/api/business-chains'
import { HopeBtn, HopeCard, HopeStatCard } from '@/components/hope'

interface CommunityElder {
  id: string
  id_card: string
  name: string
  phone?: string
  address?: string
  welfare_tags?: string[]
  status: string
  last_signin?: string
  created_at: string
}

const loading = ref(false)
const elders = ref<CommunityElder[]>([])
const kpis = ref({ total_elders: 0, today_signin: 0, welfare_tags: 0, pending_payments: 0 })

const filters = reactive({ status: '', search: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const showCreateDialog = ref(false)
const createLoading = ref(false)
const createForm = reactive({ id_card: '', name: '', phone: '', address: '' })

const showWelfareDialog = ref(false)
const currentElder = ref<CommunityElder | null>(null)
const currentWelfareTags = ref<any[]>([])

const parseTags = (json: string | null | undefined): string[] => {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [] }
}

const fetchElders = async () => {
  loading.value = true
  try {
    const res: any = await communityApi.listElders({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: filters.status,
    })
    elders.value = res.data || []
    pagination.total = elders.value.length
    kpis.value.total_elders = elders.value.length
    kpis.value.today_signin = elders.value.filter((e: CommunityElder) => e.last_signin).length
    kpis.value.pending_payments = 5
  } catch {
    elders.value = []
  } finally {
    loading.value = false
  }
}

const createElder = async () => {
  if (!createForm.id_card || !createForm.name) {
    ElMessage.warning('请填写身份证号和姓名')
    return
  }
  createLoading.value = true
  try {
    await communityApi.createElder(createForm)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    Object.assign(createForm, { id_card: '', name: '', phone: '', address: '' })
    await fetchElders()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

const triggerSignin = (row: CommunityElder) => {
  communityApi.signin(row.id, { type: 'welfare' })
    .then(() => ElMessage.success('签到成功'))
    .catch(() => ElMessage.error('签到失败'))
}

const viewDetail = (row: CommunityElder) => {
  ElMessage.info(`查看 ${row.name} 详情`)
}

const assignWelfare = (row: CommunityElder) => {
  currentElder.value = row
  currentWelfareTags.value = []
  showWelfareDialog.value = true
  communityApi.listElderWelfareTags(row.id).then((res: any) => {
    currentWelfareTags.value = res.data || []
  }).catch(() => {})
}

const revokeTag = (tagCode: string) => {
  if (!currentElder.value) return
  communityApi.revokeWelfareTag(currentElder.value.id, tagCode)
    .then(() => {
      ElMessage.success('已撤销')
      currentWelfareTags.value = currentWelfareTags.value.filter((t: any) => t.tag_code !== tagCode)
    })
    .catch(() => ElMessage.error('撤销失败'))
}

const refreshAll = () => fetchElders()

onMounted(() => {
  fetchElders()
})
</script>

<style scoped>
.community-chain-page {
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

.mr-1 { margin-right: 4px; }
.text-muted { color: var(--el-text-color-placeholder); }
</style>
