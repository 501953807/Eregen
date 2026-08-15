<template>
  <div class="self-chain-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">自营链管理</h2>
        <p class="page-subtitle">自营链用户、设备及健康数据管理</p>
      </div>
      <div class="page-header__actions">
        <HopeBtn variant="filled" size="md" @click="showCreateDialog = true">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </template>
          新增人员
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="kpis.self_active"
        label="自营用户"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.self_devices"
        label="绑定设备"
        icon-color="success"
        gradient="linear-gradient(135deg, #1aa053 0%, #22c55e 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="9" y1="7" x2="15" y2="7"/><line x1="9" y1="11" x2="15" y2="11"/><line x1="9" y1="15" x2="15" y2="15"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.self_alerts"
        label="待处理告警"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938 0%, #f59e0b 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="kpis.self_subs"
        label="订阅有效"
        icon-color="accent"
        gradient="linear-gradient(135deg, #8C57FF 0%, #a78bfa 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 13V4a2 2 0 00-2-2H6a2 2 0 00-2 2v9"/><path d="M3 21h18"/><path d="M3 7v1a3 3 0 006 0V7m0 1a3 3 0 006 0V7m0 1a3 3 0 006 0V7H9"/></svg></el-icon>
        </template>
      </HopeStatCard>
    </div>

    <!-- Main Table Card -->
    <HopeCard title="自营链人员管理" :subtitle="`共 ${pagination.total} 条记录`">
      <template #header>
        <div class="toolbar">
          <el-form :inline="true" class="filter-form">
            <el-form-item label="搜索">
              <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable style="width:180px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchElderly">查询</el-button>
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

      <el-table :data="elderlyList" v-loading="loading" stripe class="hope-table-custom">
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="birth_date" label="出生日期" width="110">
          <template #default="{ row }">{{ row.birth_date || '—' }}</template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'suspended' ? 'warning' : 'danger'" size="small">
              {{ row.status === 'active' ? '活跃' : row.status === 'suspended' ? '暂停' : '已故' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="health_risk_level" label="健康风险" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.health_risk_level" :type="riskTagType(row.health_risk_level)" size="small">
              {{ riskLabel(row.health_risk_level) }}
            </el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="subscription_tier" label="订阅层级" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.subscription_tier" type="info" size="small">{{ row.subscription_tier }}</el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="viewHealth(row)">健康报告</el-button>
            <el-button link type="primary" @click="viewGuidance(row)">健康指导</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchElderly"
        class="hope-pagination"
      />
    </HopeCard>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="'人员详情 — ' + (detailPerson?.name || '')" width="640px" class="hope-dialog">
      <div v-if="detailPerson">
        <el-descriptions :column="2" border class="hope-descriptions">
          <el-descriptions-item label="姓名">{{ detailPerson.name }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ detailPerson.id_card }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ genderLabel(detailPerson.gender) }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ detailPerson.phone || '—' }}</el-descriptions-item>
          <el-descriptions-item label="紧急联系人">{{ detailPerson.emergency_contact || '—' }}</el-descriptions-item>
          <el-descriptions-item label="地址" :span="2">{{ detailPerson.address || '—' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detailPerson.status === 'active' ? 'success' : 'warning'">{{ statusLabel(detailPerson.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailPerson.created_at }}</el-descriptions-item>
        </el-descriptions>
        <el-divider class="hope-divider">福利标签</el-divider>
        <el-table :data="welfareTags" size="small" stripe v-loading="welfareLoading" class="hope-table-custom">
          <el-table-column prop="tag_code" label="标签代码" width="150" />
          <el-table-column prop="valid_from" label="生效日期" width="120" />
          <el-table-column prop="valid_to" label="到期日期" width="120" />
        </el-table>
      </div>
      <template #footer>
        <HopeBtn variant="plain" size="sm" @click="showDetailDialog = false">关闭</HopeBtn>
      </template>
    </el-dialog>

    <!-- Health Report Dialog -->
    <el-dialog v-model="showHealthDialog" title="健康报告" width="700px" destroy-on-close class="hope-dialog">
      <div v-loading="healthLoading">
        <el-alert v-if="healthData" type="success" :closable="false" show-icon style="margin-bottom:16px">
          <template #title>健康摘要</template>
          <template #default>
            <div>心率: {{ healthData.avg_hr || '—' }} bpm | 血氧: {{ healthData.avg_spo2 || '—' }}% | 步数: {{ healthData.total_steps || 0 }}</div>
          </template>
        </el-alert>
        <el-table :data="healthRecords" size="small" stripe max-height="300" class="hope-table-custom">
          <el-table-column prop="recorded_at" label="记录时间" width="160" />
          <el-table-column prop="record_type" label="类型" width="100" />
          <el-table-column prop="hr" label="心率" width="80" />
          <el-table-column prop="spo2" label="血氧" width="80" />
          <el-table-column prop="steps" label="步数" width="80" />
        </el-table>
      </div>
    </el-dialog>

    <!-- Guidance Dialog -->
    <el-dialog v-model="showGuidanceDialog" title="健康指导" width="600px" destroy-on-close class="hope-dialog">
      <div v-loading="guidanceLoading">
        <el-alert v-if="guidanceData && guidanceData.length > 0" type="info" :closable="false" show-icon>
          <template #title>共 {{ guidanceData.length }} 条健康指导</template>
          <template #default>
            <div v-for="(g, i) in guidanceData" :key="i" style="margin-top:8px;padding:12px 14px;background:var(--el-fill-color-light);border-radius:8px;border:1px solid var(--hope-border)">
              <div style="font-weight:600;color:var(--hope-text)">{{ g.rule_name || g.rule_code }}</div>
              <div style="font-size:13px;color:var(--el-text-color-secondary);margin-top:4px">{{ g.description || g.content }}</div>
            </div>
          </template>
        </el-alert>
        <el-empty v-else description="暂无健康指导规则" :image-size="80" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { personApi, selfApi } from '@/api/business-chains'
import { HopeBtn, HopeCard, HopeStatCard } from '@/components/hope'
import type { Person } from '@/types'

const loading = ref(false)
const elderlyList = ref<Person[]>([])
const kpis = ref({ self_active: 0, self_devices: 0, self_alerts: 0, self_subs: 0 })

const filters = reactive({ search: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const showHealthDialog = ref(false)
const showGuidanceDialog = ref(false)
const detailPerson = ref<Person | null>(null)
const welfareTags = ref<any[]>([])
const welfareLoading = ref(false)
const healthRecords = ref<any[]>([])
const healthLoading = ref(false)
const healthData = ref<any>(null)
const guidanceData = ref<any[]>([])
const guidanceLoading = ref(false)

const riskLabel = (level: string) => ({ low: '低', medium: '中', high: '高', critical: '危' }[level] || level)
const riskTagType = (level: string) => ({ low: 'success', medium: 'warning', high: 'danger', critical: 'danger' }[level] || 'info')
const genderLabel = (g: number) => g === 1 ? '男' : g === 2 ? '女' : '—'
const statusLabel = (s: string) => ({ active: '活跃', suspended: '暂停', deceased: '已故' }[s] || s)

const fetchElderly = async () => {
  loading.value = true
  try {
    const res: any = await personApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      business_chain: 'self',
      status: 'active',
    })
    elderlyList.value = res.data || []
    pagination.total = res.page ? res.page * pagination.pageSize : elderlyList.value.length
    kpis.value.self_active = elderlyList.value.length
    kpis.value.self_alerts = 3
  } catch {
    kpis.value.self_active = 0
  } finally {
    loading.value = false
  }
}

const viewDetail = async (person: Person) => {
  detailPerson.value = person
  showDetailDialog.value = true
  welfareLoading.value = true
  try {
    const res: any = await personApi.listWelfareTags(person.id)
    welfareTags.value = res.data || []
  } catch {
    welfareTags.value = []
  } finally {
    welfareLoading.value = false
  }
}

const viewHealth = async (person: Person) => {
  showHealthDialog.value = true
  healthLoading.value = true
  try {
    const res: any = await selfApi.getHealthReport(person.id)
    healthData.value = res?.data?.summary || null
    healthRecords.value = res?.data?.records || []
  } catch {
    ElMessage.warning('健康数据暂未生成')
  } finally {
    healthLoading.value = false
  }
}

const viewGuidance = async (person: Person) => {
  showGuidanceDialog.value = true
  guidanceLoading.value = true
  try {
    const res: any = await selfApi.getGuidance(person.id)
    guidanceData.value = res?.data?.guidance || []
  } catch {
    ElMessage.warning('健康指导暂未生成')
  } finally {
    guidanceLoading.value = false
  }
}

const refreshAll = () => {
  fetchElderly()
}

onMounted(() => {
  fetchElderly()
})
</script>

<style scoped>
.self-chain-page {
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
  border-bottom: 1px solid rgba(26,46,38,0.06) !important;
  padding: 14px 18px !important;
  color: var(--hope-text);
}
.hope-table-custom :deep(.el-table__row:hover) td {
  background: rgba(58,87,232,0.04) !important;
}
.hope-table-custom :deep(.el-table__row:nth-child(even)) td {
  background: rgba(26,46,38,0.02);
}
.hope-table-custom :deep(.el-table__row:nth-child(even):hover) td {
  background: rgba(58,87,232,0.06);
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
.hope-pagination :deep(.el-pagination__sizes .el-select .el-input__wrapper) {
  border-radius: var(--hope-radius-md);
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

/* Descriptions */
.hope-descriptions :deep(.el-descriptions__label) {
  color: var(--hope-text-secondary);
  font-weight: 500;
  background: var(--hope-surface-light);
}
.hope-descriptions :deep(.el-descriptions__content) {
  color: var(--hope-text);
}

/* Divider */
.hope-divider {
  margin: 20px 0 !important;
  border-color: var(--hope-border) !important;
}

.text-muted { color: var(--el-text-color-placeholder); }
</style>
