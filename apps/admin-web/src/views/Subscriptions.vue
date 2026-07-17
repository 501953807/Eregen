<template>
  <div class="subscriptions-page">
    <!-- Overview Stats -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.total.toLocaleString() }}</div>
            <div class="stat-label">订阅总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #67C23A;">{{ stats.active.toLocaleString() }}</div>
            <div class="stat-label">活跃订阅</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #E6A23C;">{{ stats.expiring.toLocaleString() }}</div>
            <div class="stat-label">即将到期</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #F56C6C;">{{ stats.expired.toLocaleString() }}</div>
            <div class="stat-label">已过期</div>
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
