<template>
  <div>
    <!-- KPI Row -->
    <el-row :gutter="12" style="margin-bottom: 20px;">
      <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-primary"><div class="kpi-value">{{ kpis.total }}</div><div class="kpi-label">登记老人</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">{{ kpis.wearable }}</div><div class="kpi-label">在线腕带</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="never" class="kpi-card"><div class="kpi-value">{{ kpis.welfareTags }}</div><div class="kpi-label">福利标签</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-info"><div class="kpi-value">{{ kpis.todaySignin }}</div><div class="kpi-label">今日签到</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-warning"><div class="kpi-value">{{ kpis.pendingReview }}</div><div class="kpi-label">待审核民政</div></el-card></el-col>
      <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-danger"><div class="kpi-value">{{ kpis.alerts }}</div><div class="kpi-label">异常告警</div></el-card></el-col>
    </el-row>

    <div class="filter-bar">
      <el-button type="primary" @click="$emit('add')">＋ 新增老人</el-button>
      <el-input v-model="searchQuery" placeholder="搜索姓名 / 身份证号 / 手机号" clearable style="width: 300px;" @input="emitSearch">
        <template #prefix><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></template>
      </el-input>
    </div>

    <el-card shadow="never" class="table-card">
      <el-table :data="filteredRows" stripe v-loading="loading">
        <el-table-column prop="name" label="姓名" width="90">
          <template #default="{ row }"><strong>{{ row.name }}</strong></template>
        </el-table-column>
        <el-table-column label="身份证号" width="180">
          <template #default="{ row }"><span class="mono">{{ row.id_card || '—' }}</span></template>
        </el-table-column>
        <el-table-column label="年龄" width="60">
          <template #default="{ row }">{{ calculateAge(row.birth_date) }}</template>
        </el-table-column>
        <el-table-column label="性别" width="60">
          <template #default="{ row }">{{ row.gender || '—' }}</template>
        </el-table-column>
        <el-table-column label="紧急联系人" width="110">
          <template #default="{ row }">{{ row.emergency_contact || '—' }}</template>
        </el-table-column>
        <el-table-column label="福利标签" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="tag in (row.welfare_tags || [])" :key="tag.code" :type="welfareTagType(tag.code)" size="small" style="margin-right: 4px;">{{ tag.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <span class="status-badge" :class="row.status === '正常' ? 'badge-success' : 'badge-gray'">
              <span class="status-dot" :class="row.status === '正常' ? 'dot-success' : 'dot-gray'"></span>{{ row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="$emit('detail', row)">详情</el-button>
            <el-button link type="primary" size="small" @click="$emit('edit', row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrapper">
        <el-pagination background layout="total, prev, pager, next" :total="total" :current-page="page" :page-size="pageSize" @current-change="handlePageChange" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface WelfareTag { code: string; name: string; issuer?: string; start_date?: string; end_date?: string }
interface ElderlyRow {
  id: string; name: string; id_card?: string; birth_date?: string; gender?: string
  emergency_contact?: string; welfare_tags: WelfareTag[]; status: string
  address?: string; wearable_id?: string; wearable_online?: boolean
}

interface Kpis { total: number; wearable: number; welfareTags: number; todaySignin: number; pendingReview: number; alerts: number }

const props = defineProps<{
  rows: ElderlyRow[]
  loading: boolean
  total: number
  page: number
  pageSize: number
  kpis: Kpis
}>()

const emit = defineEmits<{
  'update:search': [q: string]
  'add': []
  'detail': [row: ElderlyRow]
  'edit': [row: ElderlyRow]
  'page-change': [page: number]
}>()

const searchQuery = ref('')
const filteredRows = computed(() => {
  if (!searchQuery.value) return props.rows
  const q = searchQuery.value.toLowerCase()
  return props.rows.filter(e => e.name.toLowerCase().includes(q) || (e.id_card && e.id_card.includes(q)))
})

function emitSearch() { emit('update:search', searchQuery.value) }

function calculateAge(birthDate?: string): number {
  if (!birthDate) return 0
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function welfareTagType(code: string): string {
  const map: Record<string, string> = {
    orphan: 'danger', poverty_1: 'warning', poverty_2: 'warning',
    disability_1: 'primary', disability_2: 'primary', disability_3: 'primary',
    special_disease: 'info', bus_discount: 'success', medical_assist: 'primary',
  }
  return map[code] || 'info'
}

function handlePageChange(p: number) { emit('page-change', p) }
</script>

<style scoped>
.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-sm) !important;
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
  box-shadow: var(--hope-shadow-md) !important;
}
.kpi-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-radius: 14px;
}
.kpi-value {
  font-size: 32px;
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1;
  margin-bottom: 4px;
}
.kpi-label {
  font-size: 12px;
  color: var(--hope-text-muted);
  margin-top: 6px;
  font-weight: 600;
}
.kpi-primary .kpi-value { color: var(--hope-primary); }
.kpi-success .kpi-value { color: var(--hope-success); }
.kpi-info .kpi-value { color: var(--hope-info); }
.kpi-warning .kpi-value { color: var(--hope-warning); }
.kpi-danger .kpi-value { color: var(--hope-error); }

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.table-card {
  margin-bottom: 0;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--hope-text-muted);
}
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: var(--hope-radius-pill);
  font-size: 12px;
  font-weight: 600;
}
.badge-success { background: var(--hope-success-light); color: var(--hope-success); }
.badge-gray { background: var(--hope-surface-light); color: var(--hope-text-muted); }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: var(--hope-success); }
.dot-gray { background: var(--hope-text-muted); }
</style>
