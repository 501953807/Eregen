<template>
  <div class="subscriptions-page">
    <!-- Overview Stats -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.total.toLocaleString() }}</div>
            <div class="stat-label">订阅总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #67C23A;">{{ stats.active.toLocaleString() }}</div>
            <div class="stat-label">活跃订阅</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #E6A23C;">{{ stats.expiring.toLocaleString() }}</div>
            <div class="stat-label">即将到期</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #F56C6C;">{{ stats.expired.toLocaleString() }}</div>
            <div class="stat-label">已过期</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #409EFF;">94.1%</div>
            <div class="stat-label">续费率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Conversion Funnel -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <template #header><span style="font-weight: 600;">订阅转化漏斗</span></template>
      <div style="max-width: 600px; margin: 0 auto;">
        <div v-for="(step, i) in funnelSteps" :key="i" style="display:flex;align-items:center;gap:12px;margin-bottom:4px;">
          <div :style="{ width: `${step.width}%`, height: 36, borderRadius: 8, display: 'flex', alignItems: 'center', padding: '0 14px', fontSize: 13, fontWeight: 600, color: '#fff', background: step.gradient }">{{ step.label }} · {{ step.count.toLocaleString() }}</div>
          <span style="font-size: 12px; color: #909399; width: 60px; text-align: right;">{{ step.percent }}%</span>
        </div>
      </div>
    </el-card>

    <!-- Tier Comparison -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="8">
        <el-card shadow="hover" :class="{ 'recommended': false }" style="text-align:center;">
          <div style="font-size:16px;font-weight:700;margin-bottom:4px;">Starter</div>
          <div style="font-size:28px;font-weight:800;color:#303133;">¥29 <span style="font-size:14px;font-weight:400;color:#909399;">/月</span></div>
          <div style="font-size:13px;color:#909399;margin-bottom:12px;">189 订阅中</div>
          <div style="text-align:left;font-size:12px;color:#606266;">
            <div style="padding:4px 0;">✓ 心率/血氧监测</div>
            <div style="padding:4px 0;">✓ SOS紧急呼叫</div>
            <div style="padding:4px 0;">✓ 基础定位</div>
            <div style="padding:4px 0;color:#C0C4CC;">— 电子围栏</div>
            <div style="padding:4px 0;color:#C0C4CC;">— ECG心电分析</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" style="text-align:center;border-color:#409EFF;">
          <div style="background:#409EFF;color:#fff;font-size:10px;font-weight:700;padding:2px 12px;border-radius:10px;display:inline-block;margin-bottom:8px;">最受欢迎</div>
          <div style="font-size:16px;font-weight:700;margin-bottom:4px;">Plus</div>
          <div style="font-size:28px;font-weight:800;color:#303133;">¥59 <span style="font-size:14px;font-weight:400;color:#909399;">/月</span></div>
          <div style="font-size:13px;color:#909399;margin-bottom:12px;">312 订阅中</div>
          <div style="text-align:left;font-size:12px;color:#606266;">
            <div style="padding:4px 0;">✓ Starter全部功能</div>
            <div style="padding:4px 0;">✓ 电子围栏</div>
            <div style="padding:4px 0;">✓ 跌倒检测</div>
            <div style="padding:4px 0;">✓ 用药管理</div>
            <div style="padding:4px 0;color:#C0C4CC;">— ECG心电分析</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" style="text-align:center;">
          <div style="font-size:16px;font-weight:700;margin-bottom:4px;">Pro</div>
          <div style="font-size:28px;font-weight:800;color:#303133;">¥99 <span style="font-size:14px;font-weight:400;color:#909399;">/月</span></div>
          <div style="font-size:13px;color:#909399;margin-bottom:12px;">148 订阅中</div>
          <div style="text-align:left;font-size:12px;color:#606266;">
            <div style="padding:4px 0;">✓ Plus全部功能</div>
            <div style="padding:4px 0;">✓ ECG心电分析</div>
            <div style="padding:4px 0;">✓ AI健康报告</div>
            <div style="padding:4px 0;">✓ 在线问诊</div>
            <div style="padding:4px 0;">✓ 优先客服</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Renewal Records Table -->
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">续费记录</span>
          <el-button size="small" @click="handleExport">导出报表</el-button>
        </div>
      </template>
      <el-table v-loading="subStore.loading" :data="subStore.renewals" stripe style="width: 100%">
        <el-table-column prop="id" label="订阅ID" width="120" />
        <el-table-column label="套餐" width="120">
          <template #default="{ row }">
            <el-tag :type="planTag(row.plan_tier)" size="small">{{ planLabel(row.plan_tier) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始日期" width="140">
          <template #default="{ row }">
            {{ row.start_date ? new Date(row.start_date).toLocaleDateString() : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="结束日期" width="140">
          <template #default="{ row }">
            {{ row.end_date ? new Date(row.end_date).toLocaleDateString() : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
        <el-pagination background layout="prev, pager, next" :total="subStore.renewals.length" :page-size="20" />
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="订阅详情" width="520px">
      <el-descriptions :column="2" border v-if="detailSub">
        <el-descriptions-item label="订阅ID">{{ detailSub.id }}</el-descriptions-item>
        <el-descriptions-item label="套餐">
          <el-tag :type="planTag(detailSub.plan_tier)" size="small">{{ planLabel(detailSub.plan_tier) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detailSub.user_id || '—' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detailSub.status === 'active' ? 'success' : 'info'" size="small">{{ detailSub.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始日期">{{ detailSub.start_date ? new Date(detailSub.start_date).toLocaleDateString() : '—' }}</el-descriptions-item>
        <el-descriptions-item label="结束日期">{{ detailSub.end_date ? new Date(detailSub.end_date).toLocaleDateString() : '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useSubscriptionStore } from '@/stores/subscription'
import type { Subscription } from '@/types'

const subStore = useSubscriptionStore()
const showDetailDialog = ref(false)
const detailSub = ref<Subscription | null>(null)

// Conversion funnel data — v2 prototype enhancement
const funnelSteps = ref([
  { label: '注册', count: 12500, width: 100, percent: 100, gradient: 'linear-gradient(90deg, #6366F1, #8B5CF6)' },
  { label: '激活设备', count: 7800, width: 72, percent: 62.4, gradient: 'linear-gradient(90deg, #8B5CF6, #A78BFA)' },
  { label: '绑定家属', count: 5200, width: 50, percent: 41.6, gradient: 'linear-gradient(90deg, #A78BFA, #C4B5FD)' },
  { label: '使用7天', count: 3100, width: 34, percent: 24.8, gradient: 'linear-gradient(90deg, #C4B5FD, #DDD6FE)' },
  { label: '付费订阅', count: 649, width: 18, percent: 5.2, gradient: 'linear-gradient(90deg, #6366F1, #8B5CF6)' },
])

// Revenue KPIs — v2 prototype enhancement
const revenueStats = ref({
  mrr: 51651,
  arr: 619812,
  avgRevenuePerUser: 79.6,
  churnRate: 3.2,
})

function planLabel(tier: string): string {
  const map: Record<string, string> = { free: '免费版', premium: '专业版', enterprise: '企业版' }
  return map[tier] || tier
}

function planTag(tier: string): 'primary' | 'success' | 'warning' {
  const map: Record<string, 'primary' | 'success' | 'warning'> = { free: 'primary', premium: 'success', enterprise: 'warning' }
  return map[tier] || 'primary'
}

async function handleExport() {
  try {
    await subStore.fetchList({ page_size: 1 })
    ElMessage.success('报表导出成功（模拟）')
  } catch {
    ElMessage.success('报表导出成功（模拟）')
  }
}

function handleDetail(row: Subscription) {
  detailSub.value = { ...row }
  showDetailDialog.value = true
}

onMounted(async () => {
  await Promise.all([subStore.fetchList(), subStore.fetchStats()])
})
</script>

<style scoped>
.stat-card :deep(.el-card__body) { padding: 20px; display: flex; align-items: center; justify-content: space-between; }
.stat-content { flex: 1; }
.stat-value { font-size: 32px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
