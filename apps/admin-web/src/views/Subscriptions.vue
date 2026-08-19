<template>
  <div class="subscriptions-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h2 class="page-title">订阅管理</h2>
        <p class="page-subtitle">管理所有用户订阅套餐、续费与取消记录</p>
      </div>
      <HopeBtn variant="filled" size="md" @click="handleCreatePlan">
        + 创建订阅计划
      </HopeBtn>
    </div>

    <!-- Revenue KPI Row -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="'¥' + revenue.mth.toLocaleString()"
        label="本月收入"
        icon-color="success"
        gradient="linear-gradient(135deg, #16a34a, #22c55e)"
      >
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">↑ {{ revenue.mth_change }}% vs 上月</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="'¥' + revenue.mrr.toLocaleString()"
        label="MRR"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8, #6366f1)"
      >
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">↑ {{ revenue.mrr_change }}% vs 上月</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.active.toLocaleString()"
        label="活跃订阅"
        icon-color="accent"
        gradient="linear-gradient(135deg, #8C57FF, #a78bfa)"
      >
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">↑ {{ revenue.active_change }}% vs 上月</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="revenue.churn_rate + '%'"
        label="Churn 率"
        icon-color="error"
        gradient="linear-gradient(135deg, #c04a42, #e11d48)"
      >
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">↓ {{ revenue.churn_improve }}% 改善</span>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="revenue.renewal_rate + '%'"
        label="续费率"
        icon-color="info"
        gradient="linear-gradient(135deg, #079aa2, #14b8a6)"
      >
        <template #trend>
          <span class="hope-stat-card__trend hope-stat-card__trend-up">↑ {{ revenue.renewal_change }}% vs 上月</span>
        </template>
      </HopeStatCard>
    </div>

    <!-- Conversion Funnel -->
    <HopeCard title="订阅转化漏斗" subtitle="从注册到付费的完整转化路径">
      <div class="funnel">
        <div v-for="(step, i) in funnelSteps" :key="i" class="funnel-step" :style="{ animationDelay: i * 0.08 + 's' }">
          <div class="funnel-step-label">{{ step.label }}</div>
          <div class="funnel-bar-wrap">
            <div class="funnel-bar" :style="{ width: step.width + '%', background: step.gradient }"></div>
            <div class="funnel-stat">
              <span class="funnel-count">{{ step.count.toLocaleString() }}</span>
              <span class="funnel-pct">{{ step.percent }}%</span>
            </div>
          </div>
        </div>
      </div>
    </HopeCard>

    <!-- Tier Comparison -->
    <div class="tier-compare">
      <div class="tier-card" v-for="tier in tiers" :key="tier.name" :class="{ recommended: tier.recommended }">
        <div v-if="tier.recommended" class="tier-rec-badge">最受欢迎</div>
        <div class="tier-name">{{ tier.name }}</div>
        <div class="tier-price">{{ tier.price }} <span>/月</span></div>
        <div class="tier-users">{{ tier.sub_count }} 订阅中</div>
        <div class="tier-features">
          <div v-for="(f, fi) in tier.features" :key="fi" class="tier-feature" :class="{ disabled: !f.active }">
            <span class="feature-icon">{{ f.active ? '✓' : '—' }}</span>
            {{ f.text }}
          </div>
        </div>
      </div>
    </div>

    <!-- Subscription Table -->
    <HopeCard title="订阅列表" subtitle="共 {{ subStore.total }} 条订阅记录">
      <template #toolbar>
        <div class="filter-bar">
          <span class="filter-label">筛选：</span>
          <HopeInput v-model="tableFilters.status" placeholder="状态筛选" clearable style="width:140px;" @input="onFilterChange">
            <template #prepend>状态</template>
          </HopeInput>
          <el-select v-model="tableFilters.plan" placeholder="全部套餐" clearable @change="searchTable" style="width:130px;">
            <el-option label="免费" value="free" />
            <el-option label="Plus" value="premium" />
            <el-option label="Pro" value="enterprise" />
          </el-select>
          <el-select v-model="tableFilters.renewal" placeholder="续费时间" clearable @change="searchTable" style="width:140px;">
            <el-option label="即将到期(7天)" value="soon" />
            <el-option label="本月到期" value="this_month" />
            <el-option label="已过期" value="expired" />
          </el-select>
          <span class="filter-spacer"></span>
          <HopeBtn variant="plain" size="sm" @click="resetTableFilters">重置</HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="searchTable">搜索</HopeBtn>
        </div>
      </template>

      <HopeTable
        :columns="tableColumns"
        :data="filteredSubscriptions"
        :loading="subStore.loading"
        :striped="true"
        compact
        class="sub-table"
      >
        <template #col-user_name="{ row }">
          <div class="user-cell">
            <strong>{{ row.user_name || '—' }}</strong>
            <span class="user-phone">{{ row.user_phone || '' }}</span>
          </div>
        </template>
        <template #col-plan_tier="{ row }">
          <HopeBadge :color="tierBadgeColor(row.plan_tier || '')">{{ tierLabel(row.plan_tier || '') }}</HopeBadge>
        </template>
        <template #col-billing_cycle="{ row }">
          <span class="plan-tag">{{ row.billing_cycle === 'annual' ? '年度' : '月度' }}</span>
        </template>
        <template #col-status="{ row }">
          <HopeBadge :color="statusBadgeColor(row.status)">{{ statusLabel(row.status) }}</HopeBadge>
        </template>
        <template #col-end_date="{ row }">
          <span class="renewal-count" :class="{ critical: isCritical(row.end_date) }">
            {{ formatRenewalDate(row.end_date) }}
          </span>
        </template>
        <template #col-plan_price="{ row }">
          ¥{{ planPrice(row.plan_tier || '') }}/月
        </template>
        <template #col-cancellation_reason="{ row }">
          <span v-if="row.cancellation_reason" class="churn-reason">{{ row.cancellation_reason }}</span>
          <span v-else class="text-muted">—</span>
        </template>
        <template #col-actions="{ row }">
          <HopeBtn variant="text" size="sm" @click.stop="openPanel(row)">详情</HopeBtn>
        </template>
      </HopeTable>

      <template #footer>
        <div class="pagination-wrapper">
          <el-pagination
            background
            layout="total, sizes, prev, pager, next, jumper"
            :total="subStore.total"
            :page-size="tableFilters.pageSize"
            :current-page="tableFilters.page"
            :page-sizes="[10, 20, 50, 100]"
            @size-change="handleSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </template>
    </HopeCard>

    <!-- Side Panel (Subscription Detail) -->
    <div class="side-panel-overlay" :class="{ show: panelOpen }" @click="closePanel" />
    <div class="side-panel" :class="{ open: panelOpen }">
      <div class="panel-header">
        <span class="panel-title">订阅详情</span>
        <HopeBtn variant="ghost" size="sm" @click="closePanel">✕</HopeBtn>
      </div>
      <div class="panel-body" v-if="panelSub">
        <!-- User header -->
        <div class="panel-user-row">
          <div class="panel-user-avatar">{{ userEmoji(panelSub.user_name) }}</div>
          <div>
            <div class="panel-user-name">{{ panelSub.user_name || '—' }}</div>
            <div class="panel-user-plan">{{ tierLabel(panelSub.plan_tier || '') }} 订阅</div>
          </div>
        </div>

        <!-- Subscription info -->
        <div class="panel-section">
          <div class="panel-section-title">订阅信息</div>
          <div class="panel-row">
            <span class="panel-row-label">状态</span>
            <span class="panel-row-value">
              <HopeBadge :color="statusBadgeColor(panelSub.status)">{{ statusLabel(panelSub.status) }}</HopeBadge>
            </span>
          </div>
          <div class="panel-row">
            <span class="panel-row-label">套餐等级</span>
            <span class="panel-row-value">
              <HopeBadge :color="tierBadgeColor(panelSub.plan_tier || '')">{{ tierLabel(panelSub.plan_tier || '') }}</HopeBadge>
            </span>
          </div>
          <div class="panel-row"><span class="panel-row-label">计费周期</span><span class="panel-row-value">{{ panelSub.billing_cycle === 'annual' ? '年度' : '月度' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">开始时间</span><span class="panel-row-value">{{ formatDate(panelSub.start_date) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">到期时间</span><span class="panel-row-value">{{ formatDate(panelSub.end_date) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">月均费用</span><span class="panel-row-value">¥{{ planPrice(panelSub.plan_tier || '') }}/月</span></div>
          <div class="panel-row"><span class="panel-row-label">累计消费</span><span class="panel-row-value">¥{{ panelSub.total_spent?.toLocaleString() || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">支付方式</span><span class="panel-row-value">微信支付</span></div>
        </div>

        <!-- Related devices -->
        <div class="panel-section">
          <div class="panel-section-title">关联设备</div>
          <div v-if="panelSub.devices?.length" class="device-list">
            <HopeBadge v-for="(dev, i) in panelSub.devices" :key="i" color="info" type="dot">{{ dev }}</HopeBadge>
          </div>
          <div v-else class="text-muted panel-empty">暂无关联设备</div>
        </div>

        <!-- Billing timeline -->
        <div class="panel-section">
          <div class="panel-section-title">账单记录</div>
          <HopeTimeline :items="timelineItems" />
        </div>

        <!-- Actions -->
        <div class="panel-section">
          <div class="panel-section-title">操作</div>
          <div class="panel-actions">
            <HopeBtn variant="filled" size="sm" @click="manualRenew">手动续费</HopeBtn>
            <HopeBtn variant="outlined" size="sm" @click="changePlan">变更套餐</HopeBtn>
            <HopeBtn variant="outlined" size="sm" @click="sendReminder">发送提醒</HopeBtn>
            <HopeBtn variant="error" size="sm" @click="forceCancel">强制取消</HopeBtn>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSubscriptionStore } from '@/stores/subscription'
import type { Subscription } from '@/types'
import { HopeCard, HopeStatCard, HopeTable, HopeBadge, HopeBtn, HopeTimeline, HopeInput } from '@/components/hope'

const subStore = useSubscriptionStore()

// Revenue KPI data
const revenue = ref({
  mth: 28450,
  mrr: 87200,
  arr: 1046400,
  churn_rate: 3.2,
  renewal_rate: 94.1,
  mth_change: 12.3,
  mrr_change: 8.1,
  active_change: 5.2,
  churn_improve: 0.8,
  renewal_change: 1.5,
})

// Funnel steps — v2 prototype enhancement
const funnelSteps = ref([
  { label: '注册账号', count: 5234, width: 100, percent: 100, gradient: 'linear-gradient(90deg, #3a57e8, #6366f1)' },
  { label: '绑定设备', count: 3558, width: 68, percent: 68.0, gradient: 'linear-gradient(90deg, #6366f1, #8C57FF)' },
  { label: '使用7天+', count: 1832, width: 35, percent: 35.0, gradient: 'linear-gradient(90deg, #8C57FF, #a78bfa)' },
  { label: '试用开始', count: 945, width: 18, percent: 18.1, gradient: 'linear-gradient(90deg, #a78bfa, #c4b5fd)' },
  { label: '付费订阅', count: 487, width: 9, percent: 9.3, gradient: 'linear-gradient(90deg, #16a34a, #22c55e)' },
])

// Tier comparison data — v2 prototype
const tiers = ref([
  {
    name: 'Starter', price: '¥29', sub_count: 189, recommended: false,
    features: [
      { text: '心率/血氧监测', active: true },
      { text: 'SOS紧急呼叫', active: true },
      { text: '基础定位', active: true },
      { text: '电子围栏', active: false },
      { text: 'ECG心电分析', active: false },
    ],
  },
  {
    name: 'Plus', price: '¥59', sub_count: 312, recommended: true,
    features: [
      { text: 'Starter全部功能', active: true },
      { text: '电子围栏', active: true },
      { text: '跌倒检测', active: true },
      { text: '用药管理', active: true },
      { text: 'ECG心电分析', active: false },
    ],
  },
  {
    name: 'Pro', price: '¥99', sub_count: 148, recommended: false,
    features: [
      { text: 'Plus全部功能', active: true },
      { text: 'ECG心电分析', active: true },
      { text: 'AI健康报告', active: true },
      { text: '在线问诊', active: true },
      { text: '优先客服', active: true },
    ],
  },
])

// Stats
const stats = computed(() => ({
  total: subStore.total,
  active: subStore.subscriptions.filter(s => s.status === 'active').length,
  expiring: subStore.subscriptions.filter(s => {
    if (!s.end_date) return false
    const days = Math.ceil((new Date(s.end_date).getTime() - Date.now()) / 86400000)
    return days > 0 && days <= 7
  }).length,
  expired: subStore.subscriptions.filter(s => s.status === 'expired' || s.status === 'past_due').length,
}))

// Table filters
const tableFilters = ref({
  status: '',
  plan: '',
  renewal: '',
  page: 1,
  pageSize: 20,
})

const filteredSubscriptions = computed(() => {
  let list = subStore.subscriptions
  if (tableFilters.value.status) list = list.filter(s => s.status === tableFilters.value.status)
  if (tableFilters.value.plan) list = list.filter(s => s.plan_tier === tableFilters.value.plan)
  return list
})

// HopeTable columns
const tableColumns = [
  { prop: 'user_name', label: '用户' },
  { prop: 'plan_tier', label: '套餐' },
  { prop: 'billing_cycle', label: '计费周期' },
  { prop: 'status', label: '状态' },
  { prop: 'end_date', label: '到期时间' },
  { prop: 'plan_price', label: '月费' },
  { prop: 'cancellation_reason', label: '取消原因' },
  { prop: 'actions', label: '操作' },
]

// Helpers
function tierLabel(tier: string): string {
  const map: Record<string, string> = { free: '免费', premium: 'Plus', enterprise: 'Pro' }
  return map[tier] || tier
}

function tierClass(tier: string): string {
  const map: Record<string, string> = { free: 'tier-basic', premium: 'tier-plus', enterprise: 'tier-pro' }
  return map[tier] || 'tier-basic'
}

function tierBadgeColor(tier: string): 'primary' | 'success' | 'accent' | 'info' {
  const map: Record<string, 'primary' | 'success' | 'accent' | 'info'> = {
    free: 'info', premium: 'primary', enterprise: 'accent'
  }
  return map[tier] || 'info'
}

function planPrice(tier: string): number {
  const map: Record<string, number> = { free: 0, premium: 59, enterprise: 99 }
  return map[tier] || 29
}

function statusBadgeColor(status: string): 'success' | 'warning' | 'info' | 'error' | 'accent' {
  const map: Record<string, 'success' | 'warning' | 'info' | 'error' | 'accent'> = {
    active: 'success', trial: 'warning', expired: 'info',
    cancelled: 'error', past_due: 'accent',
  }
  return map[status] || 'success'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    active: '活跃', trial: '试用中', expired: '已过期',
    cancelled: '已取消', past_due: '逾期',
  }
  return map[status] || '活跃'
}

function isCritical(dateStr?: string): boolean {
  if (!dateStr) return false
  const days = Math.ceil((new Date(dateStr).getTime() - Date.now()) / 86400000)
  return days <= 0
}

function formatRenewalDate(dateStr?: string): string {
  if (!dateStr) return '—'
  const days = Math.ceil((new Date(dateStr).getTime() - Date.now()) / 86400000)
  if (days < 0) return `逾期${Math.abs(days)}天`
  if (days === 0) return '今天到期'
  if (days <= 7) return `剩余${days}天`
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

function formatDate(ts?: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleDateString('zh-CN')
}

function userEmoji(name?: string): string {
  if (!name) return '👤'
  return ['👨', '👩', '👴', '👵'][name.length % 4]
}

// Billing timeline mapped to HopeTimeline items
const timelineItems = ref([
  { title: '年度续费成功', meta: '2025-01-15', body: '¥1,188.00', color: 'success' as const },
  { title: '年度续费成功', meta: '2024-08-15', body: '¥1,188.00', color: 'success' as const },
  { title: '月度扣款成功', meta: '2024-07-01', body: '¥99.00', color: 'success' as const },
  { title: '下次自动续费', meta: '2025-08-15', body: '¥1,188.00', color: 'warning' as const },
])

// Side Panel
const panelOpen = ref(false)
const panelSub = ref<Subscription | null>(null)

function openPanel(row: Subscription) {
  panelSub.value = { ...row }
  panelOpen.value = true
}

function closePanel() {
  panelOpen.value = false
}

// Table actions
function resetTableFilters() {
  tableFilters.value = { status: '', plan: '', renewal: '', page: 1, pageSize: 20 }
}

function onFilterChange() {
  // debounce handled by el-select
}

function searchTable() {
  tableFilters.value.page = 1
  subStore.fetchList()
}

function handleSizeChange(size: number) { tableFilters.value.pageSize = size; subStore.fetchList() }
function handlePageChange(page: number) { tableFilters.value.page = page; subStore.fetchList() }

async function handleCreatePlan() {
  ElMessage.info('创建订阅计划功能需要管理员权限，请联系超级管理员操作')
}

function manualRenew() {
  ElMessageBox.confirm('确认手动续费？', '提示', { type: 'info' })
    .then(() => ElMessage.success('续费成功'))
    .catch(() => {})
}

async function changePlan() {
  ElMessage.info('变更套餐功能需要在弹窗中选择新档位并确认')
}

function sendReminder() {
  ElMessage.info('提醒发送成功')
}

async function forceCancel() {
  try {
    await ElMessageBox.confirm('确认强制取消该订阅？此操作不可恢复。', '警告', { type: 'warning' })
    ElMessage.success('订阅已取消')
    closePanel()
  } catch { /* cancelled */ }
}

onMounted(async () => {
  await Promise.all([subStore.fetchList(), subStore.fetchStats()])
})
</script>

<style scoped>
.subscriptions-page {
  padding: 0;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 24px;
}
.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}
.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

/* Funnel */
.funnel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.funnel-step {
  display: flex;
  align-items: center;
  gap: 16px;
  animation: funnelSlideIn 0.4s ease both;
}
@keyframes funnelSlideIn {
  from { opacity: 0; transform: translateX(-8px); }
  to { opacity: 1; transform: translateX(0); }
}
.funnel-step-label {
  width: 90px;
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text);
  flex-shrink: 0;
}
.funnel-bar-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}
.funnel-bar {
  height: 28px;
  border-radius: 6px;
  transition: width 0.6s ease;
  min-width: 32px;
}
.funnel-stat {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  min-width: 56px;
}
.funnel-count {
  font-size: 14px;
  font-weight: 700;
  color: var(--hope-text);
}
.funnel-pct {
  font-size: 11px;
  color: var(--hope-text-muted);
}

/* Tier Cards */
.tier-compare {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}
.tier-card {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-xl);
  padding: 24px;
  border: 1px solid var(--hope-border);
  text-align: center;
  position: relative;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}
.tier-card:hover {
  box-shadow: var(--hope-shadow-md);
  transform: translateY(-2px);
}
.tier-card.recommended {
  border-color: var(--hope-primary);
  box-shadow: 0 0 0 1px var(--hope-primary), var(--hope-shadow-primary);
}
.tier-rec-badge {
  position: absolute;
  top: -10px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--hope-primary-gradient);
  color: var(--hope-white);
  font-size: 11px;
  font-weight: 700;
  padding: 3px 14px;
  border-radius: 12px;
}
.tier-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--hope-text);
  margin-bottom: 6px;
}
.tier-price {
  font-size: 32px;
  font-weight: 800;
  color: var(--hope-primary);
  margin: 10px 0;
  letter-spacing: -0.03em;
}
.tier-price span {
  font-size: 14px;
  font-weight: 400;
  color: var(--hope-text-muted);
}
.tier-users {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin-bottom: 16px;
}
.tier-features {
  text-align: left;
  font-size: 13px;
}
.tier-feature {
  padding: 6px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--hope-text);
}
.tier-feature .feature-icon {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}
.tier-feature:not(.disabled) .feature-icon {
  background: rgba(var(--hope-success-rgb), 0.12);
  color: var(--hope-success);
}
.tier-feature.disabled {
  color: var(--hope-text-muted);
}
.tier-feature.disabled .feature-icon {
  background: rgba(var(--hope-text-muted-rgb, 148,169,162), 0.12);
  color: var(--hope-gray-400);
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
  padding: 16px 22px 0;
}
.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text-muted);
  white-space: nowrap;
}
.filter-spacer {
  flex: 1;
}

/* Table */
.sub-table {
  margin-top: 8px;
}
.sub-table :deep(.hope-table th) {
  font-size: 12px;
}
.sub-table :deep(.hope-table td) {
  font-size: 13px;
}
.user-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.user-phone {
  font-size: 11px;
  color: var(--hope-text-muted);
  font-family: monospace;
}
.plan-tag {
  font-size: 12px;
  color: var(--hope-text-muted);
}
.renewal-count {
  font-size: 12px;
  font-weight: 600;
}
.renewal-count.critical {
  color: var(--hope-danger);
}
.churn-reason {
  font-size: 11px;
  color: var(--hope-text-muted);
  background: rgba(148,169,162,0.1);
  padding: 2px 8px;
  border-radius: 6px;
}
.text-muted {
  color: var(--hope-text-muted);
}

/* Pagination */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 14px 0;
}

/* ========== Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.3);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show {
  display: block;
}
.side-panel {
  position: fixed;
  top: 0;
  right: -540px;
  bottom: 0;
  width: 540px;
  background: var(--hope-bg);
  z-index: 201;
  transition: right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(0,0,0,0.1);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  background: var(--hope-surface);
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 1;
}
.panel-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--hope-text);
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: var(--hope-bg);
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--hope-text-muted);
}
.panel-close:hover {
  background: var(--hope-border);
  color: var(--hope-text);
}

.panel-body {
  padding: 20px 24px;
}
.panel-user-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
  padding: 16px;
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: 1px solid var(--hope-border);
}
.panel-user-avatar {
  width: 52px;
  height: 52px;
  border-radius: 26px;
  background: var(--hope-primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}
.panel-user-name {
  font-size: 17px;
  font-weight: 700;
  color: var(--hope-text);
}
.panel-user-plan {
  font-size: 12px;
  color: var(--hope-text-muted);
  margin-top: 2px;
}

.panel-section {
  margin-bottom: 24px;
}
.panel-section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.6px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--hope-border);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--hope-border);
}
.panel-row:last-child {
  border-bottom: none;
}
.panel-row-label {
  color: var(--hope-text-muted);
}
.panel-row-value {
  font-weight: 600;
  color: var(--hope-text);
}

/* Device list in panel */
.device-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.panel-empty {
  font-size: 13px;
  color: var(--hope-text-muted);
}

/* Panel actions */
.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* Responsive */
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(3, 1fr); }
  .tier-compare { grid-template-columns: 1fr; }
}
@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .subscriptions-page .page-header { flex-direction: column; align-items: flex-start; gap: 12px; }
  .side-panel { width: 100%; right: -100%; }
}
</style>
