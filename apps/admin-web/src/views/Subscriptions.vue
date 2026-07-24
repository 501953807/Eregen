<template>
  <div class="subscriptions-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">订阅管理</h2>
      <el-button type="primary" @click="handleCreatePlan" size="default">+ 创建订阅计划</el-button>
    </div>

    <!-- Revenue KPI Row (5 columns) -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="5">
        <el-card shadow="hover" class="rev-card">
          <div class="rev-label">本月收入</div>
          <div class="rev-value" style="color: var(--el-color-success);">¥{{ revenue.mth.toLocaleString() }}</div>
          <div class="rev-change up">↑ {{ revenue.mth_change }}% vs 上月</div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="rev-card">
          <div class="rev-label">MRR</div>
          <div class="rev-value">¥{{ revenue.mrr.toLocaleString() }}</div>
          <div class="rev-change up">↑ {{ revenue.mrr_change }}%</div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="rev-card">
          <div class="rev-label">活跃订阅</div>
          <div class="rev-value">{{ stats.active.toLocaleString() }}</div>
          <div class="rev-change up">↑ {{ revenue.active_change }}%</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="rev-card">
          <div class="rev-label">Churn 率</div>
          <div class="rev-value" style="color: var(--el-color-danger);">{{ revenue.churn_rate }}%</div>
          <div class="rev-change down">↓ {{ revenue.churn_improve }}% 改善</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="rev-card">
          <div class="rev-label">续费率</div>
          <div class="rev-value" style="color: var(--el-color-primary);">{{ revenue.renewal_rate }}%</div>
          <div class="rev-change up">↑ {{ revenue.renewal_change }}%</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Conversion Funnel -->
    <el-card shadow="hover" class="funnel-card" style="margin-bottom: 16px;">
      <template #header><span class="section-title">订阅转化漏斗</span></template>
      <div class="funnel">
        <div v-for="(step, i) in funnelSteps" :key="i" class="funnel-step">
          <div class="funnel-bar-wrap">
            <div class="funnel-bar" :style="{ width: step.width + '%' }">{{ step.label }} · {{ step.count.toLocaleString() }}</div>
            <div class="funnel-pct">{{ step.percent }}%</div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- Tier Comparison -->
    <div class="tier-compare" style="margin-bottom: 16px;">
      <div class="tier-card" v-for="tier in tiers" :key="tier.name" :class="{ recommended: tier.recommended }">
        <div v-if="tier.recommended" class="tier-rec-badge">最受欢迎</div>
        <div class="tier-name">{{ tier.name }}</div>
        <div class="tier-price">{{ tier.price }} <span>/月</span></div>
        <div class="tier-users">{{ tier.sub_count }} 订阅中</div>
        <div class="tier-features">
          <div v-for="(f, fi) in tier.features" :key="fi" class="tier-feature" :class="{ disabled: !f.active }">
            {{ f.text }}
          </div>
        </div>
      </div>
    </div>

    <!-- Subscription Table -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="filter-bar" style="border:none;padding:0;margin:0;background:transparent;">
          <span class="filter-label">筛选：</span>
          <el-select v-model="tableFilters.status" placeholder="全部状态" clearable class="filter-select" style="width:120px;">
            <el-option label="活跃" value="active" />
            <el-option label="试用中" value="trial" />
            <el-option label="已过期" value="expired" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="逾期" value="past_due" />
          </el-select>
          <el-select v-model="tableFilters.plan" placeholder="全部套餐" clearable class="filter-select" style="width:120px;">
            <el-option label="Starter" value="free" />
            <el-option label="Plus" value="premium" />
            <el-option label="Pro" value="enterprise" />
          </el-select>
          <el-select v-model="tableFilters.renewal" placeholder="续费时间" clearable class="filter-select" style="width:130px;">
            <el-option label="即将到期(7天)" value="soon" />
            <el-option label="本月到期" value="this_month" />
            <el-option label="已过期" value="expired" />
          </el-select>
          <span class="filter-spacer"></span>
          <el-button size="small" @click="resetTableFilters">重置</el-button>
          <el-button size="small" type="primary" @click="searchTable">搜索</el-button>
        </div>
      </template>

      <el-table
        v-loading="subStore.loading"
        :data="filteredSubscriptions"
        stripe
        class="sub-table"
        @row-click="openPanel"
        highlight-current-row
      >
        <el-table-column type="selection" width="30" />
        <el-table-column label="用户" min-width="130">
          <template #default="{ row }">
            <div class="user-cell">
              <strong>{{ row.user_name || '—' }}</strong>
              <span class="user-phone">{{ row.user_phone || '' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="套餐" width="90">
          <template #default="{ row }">
            <span class="tier-tag" :class="tierClass(row.plan_tier)">{{ tierLabel(row.plan_tier) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="计费周期" width="80">
          <template #default="{ row }">
            <span class="plan-tag">{{ row.billing_cycle === 'annual' ? '年度' : '月度' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="statusBadgeClass(row.status)">
              <span class="status-dot" :class="statusBadgeClass(row.status)"></span>
              {{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="到期时间" width="110">
          <template #default="{ row }">
            <span class="renewal-count" :class="{ critical: isCritical(row.end_date) }">
              {{ formatRenewalDate(row.end_date) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="月费" width="80">
          <template #default="{ row }">
            ¥{{ planPrice(row.plan_tier) }}/月
          </template>
        </el-table-column>
        <el-table-column label="取消原因" min-width="100">
          <template #default="{ row }">
            <span v-if="row.cancellation_reason" class="churn-reason">{{ row.cancellation_reason }}</span>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="openPanel(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

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
    </el-card>

    <!-- Side Panel (Subscription Detail) -->
    <div class="side-panel-overlay" :class="{ show: panelOpen }" @click="closePanel" />
    <div class="side-panel" :class="{ open: panelOpen }">
      <div class="panel-header">
        <span class="panel-title">订阅详情</span>
        <button class="panel-close" @click="closePanel">&#10005;</button>
      </div>
      <div class="panel-body" v-if="panelSub">
        <!-- User header -->
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:20px;">
          <div class="panel-user-avatar">{{ userEmoji(panelSub.user_name) }}</div>
          <div>
            <div style="font-size:18px;font-weight:700;">{{ panelSub.user_name || '—' }}</div>
            <div style="font-size:12px;color:var(--el-text-color-secondary);">{{ tierLabel(panelSub.plan_tier) }} 订阅</div>
          </div>
        </div>

        <!-- Subscription info -->
        <div class="panel-section">
          <div class="panel-section-title">订阅信息</div>
          <div class="panel-row"><span class="panel-row-label">状态</span><span class="panel-row-value">
            <span class="status-badge" :class="statusBadgeClass(panelSub.status)">
              <span class="status-dot" :class="statusBadgeClass(panelSub.status)"></span>
              {{ statusLabel(panelSub.status) }}
            </span>
          </span></div>
          <div class="panel-row"><span class="panel-row-label">套餐等级</span><span class="panel-row-value">
            <span class="tier-tag" :class="tierClass(panelSub.plan_tier)">{{ tierLabel(panelSub.plan_tier) }}</span>
          </span></div>
          <div class="panel-row"><span class="panel-row-label">计费周期</span><span class="panel-row-value">{{ panelSub.billing_cycle === 'annual' ? '年度' : '月度' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">开始时间</span><span class="panel-row-value">{{ formatDate(panelSub.start_date) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">到期时间</span><span class="panel-row-value">{{ formatDate(panelSub.end_date) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">月均费用</span><span class="panel-row-value">¥{{ planPrice(panelSub.plan_tier) }}/月</span></div>
          <div class="panel-row"><span class="panel-row-label">累计消费</span><span class="panel-row-value">¥{{ panelSub.total_spent?.toLocaleString() || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">支付方式</span><span class="panel-row-value">微信支付</span></div>
        </div>

        <!-- Related devices -->
        <div class="panel-section">
          <div class="panel-section-title">关联设备</div>
          <div class="panel-row" v-for="(dev, i) in (panelSub.devices || [])" :key="i" style="cursor:pointer;color:var(--el-color-primary);">
            <span class="panel-row-value">{{ dev }}</span>
          </div>
          <div v-if="!panelSub.devices?.length" class="panel-row">
            <span class="panel-row-value" style="color:var(--el-text-color-placeholder);">暂无关联设备</span>
          </div>
        </div>

        <!-- Billing timeline -->
        <div class="panel-section">
          <div class="panel-section-title">账单记录</div>
          <div class="billing-timeline">
            <div v-for="(item, i) in billingTimeline" :key="i" class="timeline-item" :class="item.state">
              <div class="timeline-date">{{ item.date }}</div>
              <div class="timeline-desc">{{ item.desc }}</div>
              <div class="timeline-amount">{{ item.amount }}</div>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="panel-section">
          <div class="panel-section-title">操作</div>
          <div class="panel-actions">
            <el-button size="small" type="primary" @click="manualRenew">手动续费</el-button>
            <el-button size="small" @click="changePlan">变更套餐</el-button>
            <el-button size="small" @click="sendReminder">发送提醒</el-button>
            <el-button size="small" type="danger" plain @click="forceCancel">强制取消</el-button>
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
  { label: '注册账号', count: 5234, width: 100, percent: 100, gradient: 'linear-gradient(90deg, #2563EB, #7C3AED)' },
  { label: '绑定设备', count: 3558, width: 68, percent: 68.0, gradient: 'linear-gradient(90deg, #3B82F6, #6366F1)' },
  { label: '使用7天+', count: 1832, width: 35, percent: 35.0, gradient: 'linear-gradient(90deg, #6366F1, #8B5CF6)' },
  { label: '试用开始', count: 945, width: 18, percent: 18.1, gradient: 'linear-gradient(90deg, #8B5CF6, #A78BFA)' },
  { label: '付费订阅', count: 487, width: 9, percent: 9.3, gradient: 'linear-gradient(90deg, #A78BFA, #C4B5FD)' },
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

// Helpers
function tierLabel(tier: string): string {
  const map: Record<string, string> = { free: '免费', premium: 'Plus', enterprise: 'Pro' }
  return map[tier] || tier
}

function tierClass(tier: string): string {
  const map: Record<string, string> = { free: 'tier-basic', premium: 'tier-plus', enterprise: 'tier-pro' }
  return map[tier] || 'tier-basic'
}

function planPrice(tier: string): number {
  const map: Record<string, number> = { free: 0, premium: 59, enterprise: 99 }
  return map[tier] || 29
}

function statusBadgeClass(status: string): string {
  const map: Record<string, string> = {
    active: 'status-active', trial: 'status-trial', expired: 'status-expired',
    cancelled: 'status-cancelled', past_due: 'status-past-due',
  }
  return map[status] || 'status-active'
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

// Billing timeline mock data
const billingTimeline = ref([
  { date: '2025-01-15', desc: '年度续费成功', amount: '¥1,188.00', state: 'paid' },
  { date: '2024-08-15', desc: '年度续费成功', amount: '¥1,188.00', state: 'paid' },
  { date: '2024-07-01', desc: '月度扣款成功', amount: '¥99.00', state: 'paid' },
  { date: '2025-08-15', desc: '下次自动续费', amount: '¥1,188.00', state: 'pending' },
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

function searchTable() {
  tableFilters.value.page = 1
  subStore.fetchList()
}

function handleSizeChange(size: number) { tableFilters.value.pageSize = size; subStore.fetchList() }
function handlePageChange(page: number) { tableFilters.value.page = page; subStore.fetchList() }

function handleCreatePlan() {
  ElMessage.info('创建订阅计划功能开发中...')
}

function manualRenew() {
  ElMessageBox.confirm('确认手动续费？', '提示', { type: 'info' })
    .then(() => ElMessage.success('续费成功'))
    .catch(() => {})
}

function changePlan() {
  ElMessage.info('变更套餐功能开发中...')
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

/* Page header */
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

/* Revenue KPI cards */
.rev-card :deep(.el-card__body) {
  padding: 16px;
  text-align: left;
}
.rev-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}
.rev-value {
  font-size: 24px;
  font-weight: 800;
  line-height: 1.2;
}
.rev-change {
  font-size: 11px;
  margin-top: 4px;
  font-weight: 600;
}
.rev-change.up { color: var(--el-color-success); }
.rev-change.down { color: var(--el-color-danger); }

/* Funnel card */
.funnel-card :deep(.el-card__body) {
  padding: 20px;
}
.section-title {
  font-size: 15px;
  font-weight: 700;
}
.funnel {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.funnel-step {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  transition: transform 0.15s;
  cursor: pointer;
  width: 100%;
}
.funnel-step:hover {
  transform: scale(1.02);
}
.funnel-bar-wrap {
  flex: 1;
  max-width: 600px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.funnel-bar {
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  padding: 0 14px;
  font-size: 13px;
  font-weight: 700;
  color: white;
  transition: width 0.6s ease;
}
.funnel-pct {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  width: 60px;
  text-align: right;
}

/* Tier comparison */
.tier-compare {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.tier-card {
  background: white;
  border-radius: 14px;
  padding: 20px;
  border: 1px solid var(--el-border-color-light);
  text-align: center;
  position: relative;
}
.tier-card.recommended {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px var(--el-color-primary);
}
.tier-rec-badge {
  position: absolute;
  top: -10px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--el-color-primary);
  color: white;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 12px;
  border-radius: 10px;
}
.tier-name {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 4px;
}
.tier-price {
  font-size: 28px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  margin: 8px 0;
}
.tier-price span {
  font-size: 14px;
  font-weight: 400;
  color: var(--el-text-color-placeholder);
}
.tier-users {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 12px;
}
.tier-features {
  text-align: left;
  font-size: 12px;
  color: var(--el-text-color-regular);
}
.tier-feature {
  padding: 4px 0;
  display: flex;
  align-items: center;
  gap: 6px;
}
.tier-feature::before {
  content: '✓';
  color: var(--el-color-success);
  font-weight: 700;
}
.tier-feature.disabled {
  color: var(--el-text-color-placeholder);
}
.tier-feature.disabled::before {
  content: '—';
  color: var(--el-text-color-placeholder);
}

/* Filter bar inside table header */
.filter-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
.filter-select {
  width: 120px;
}
.filter-spacer {
  flex: 1;
}

/* Table */
.table-card :deep(.el-card__header) {
  padding: 0;
}
.sub-table {
  width: 100%;
}
.sub-table :deep(.el-table__row) {
  cursor: pointer;
}
.sub-table :deep(.el-table__row:hover) {
  background-color: var(--el-fill-color-light) !important;
}

.user-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.user-phone {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  font-family: monospace;
}

/* Status badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}
.status-active { background: #F0FDF4; color: var(--el-color-success); }
.status-trial { background: #FFFBEB; color: #D97706; }
.status-expired { background: var(--el-fill-color-light); color: var(--el-text-color-secondary); }
.status-cancelled { background: #FEF2F2; color: var(--el-color-danger); }
.status-past-due { background: #FFF7ED; color: #C2410C; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.status-active .status-dot { background: var(--el-color-success); }
.status-trial .status-dot { background: #D97706; }
.status-expired .status-dot { background: var(--el-text-color-placeholder); }
.status-cancelled .status-dot { background: var(--el-color-danger); }
.status-past-due .status-dot { background: #C2410C; animation: pulse 1.5s infinite; }
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.tier-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
}
.tier-pro { background: #EDE9FE; color: #7C3AED; }
.tier-plus { background: #DBEAFE; color: #2563EB; }
.tier-basic { background: var(--el-fill-color-light); color: var(--el-text-color-secondary); }

.plan-tag {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.renewal-count {
  font-size: 12px;
  font-weight: 600;
}
.renewal-count.critical {
  color: var(--el-color-danger);
}

.churn-reason {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
  padding: 2px 8px;
  border-radius: 6px;
}

/* Pagination */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 14px 20px;
  border-top: 1px solid var(--el-border-color-light);
}

/* ========== Side Panel ========== */
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
}
.panel-close:hover {
  background: var(--el-border-color-light);
}

.panel-body {
  padding: 20px 24px;
}
.panel-user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 24px;
  background: var(--el-color-primary-light-9);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.panel-section {
  margin-bottom: 20px;
}
.panel-section-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--el-text-color-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}
.panel-row-label {
  color: var(--el-text-color-secondary);
}
.panel-row-value {
  font-weight: 600;
}

/* Billing timeline */
.billing-timeline {
  position: relative;
  padding-left: 20px;
}
.billing-timeline::before {
  content: '';
  position: absolute;
  left: 6px;
  top: 4px;
  bottom: 4px;
  width: 2px;
  background: var(--el-border-color-light);
}
.timeline-item {
  position: relative;
  padding-bottom: 14px;
}
.timeline-item::before {
  content: '';
  position: absolute;
  left: -17px;
  top: 5px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid var(--el-border-color-base);
  background: white;
}
.timeline-item.paid::before {
  border-color: var(--el-color-success);
  background: var(--el-color-success);
}
.timeline-item.pending::before {
  border-color: var(--el-color-warning);
  background: var(--el-color-warning);
}
.timeline-item.failed::before {
  border-color: var(--el-color-danger);
  background: var(--el-color-danger);
}
.timeline-date {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}
.timeline-desc {
  font-size: 12px;
  font-weight: 600;
}
.timeline-amount {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

/* Panel actions */
.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
