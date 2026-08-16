<template>
  <div class="medication-page">
    <!-- Page Header -->
    <div class="hope-page-header">
      <div>
        <h1 class="hope-page-header__title">
          用药管理{{ currentElder?.name ? ` — ${currentElder.name}` : '' }}
        </h1>
        <p class="hope-page-header__subtitle">管理老人用药规则、服药提醒与依从性统计</p>
      </div>
      <div class="hope-page-header__actions">
        <HopeBtn variant="filled" @click="openCreateDialog">
          <template #icon>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </template>
          添加用药规则
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="hope-grid-4" style="margin-bottom: 24px;">
      <HopeStatCard
        value="stats.activeRules"
        label="今日规则数"
        icon-color="success"
        gradient="linear-gradient(135deg, #1aa053, #22c55e)"
      >
        <template #icon>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="6" width="18" height="12" rx="2"/><path d="M8 6V4M16 6V4M3 10h18"/><circle cx="8" cy="14" r="1.5"/><circle cx="16" cy="14" r="1.5"/>
          </svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.missedCount"
        label="今日漏服"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938, #f59e0b)"
      >
        <template #icon>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/>
          </svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="`${stats.adherenceRate}%`"
        label="按时服药率"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8, #6f42c1)"
      >
        <template #icon>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/>
          </svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.pendingActions"
        label="待处理提醒"
        icon-color="error"
        gradient="linear-gradient(135deg, #c03221, #ef4444)"
      >
        <template #icon>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4.8 2.3A.3.3 0 105 2H4a2 2 0 00-2 2v5a6 6 0 006 6 6 6 0 006-6V4a2 2 0 00-2-2h-1a.2.2 0 00.3.3"/><path d="M8 15v4M12 15v4M6 23h8"/>
          </svg>
        </template>
      </HopeStatCard>
    </div>

    <!-- Medication Rules Table -->
    <HopeCard title="用药规则列表">
      <template #header>
        <HopeInput
          v-model="searchQuery"
          placeholder="搜索药品名称..."
          size="sm"
          style="width: 220px;"
          @update:modelValue="handleSearch"
        >
          <template #prefix>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
            </svg>
          </template>
        </HopeInput>
      </template>
      <HopeTable
        :columns="tableColumns"
        :data="filteredRules"
        :loading="loading.rules"
        striped
      >
        <template #col-pillType="{ row }">
          <span style="font-weight: 600; color: var(--hope-text);">{{ row.pillType }}</span>
        </template>
        <template #col-scheduleTime="{ row }">
          <span style="color: var(--hope-text-secondary);">{{ row.scheduleTime }}</span>
        </template>
        <template #col-doseCount="{ row }">
          <span style="font-weight: 600;">{{ row.doseCount }}</span>
          <span style="color: var(--hope-text-muted); font-size: 12px;">粒</span>
        </template>
        <template #col-daysOfWeek="{ row }">
          <div style="display: flex; flex-wrap: wrap; gap: 4px;">
            <span v-for="day in (row.daysOfWeek || [])" :key="day" class="hope-chip hope-chip--primary" style="font-size: 11px; padding: 2px 8px;">
              {{ { mon: '一', tue: '二', wed: '三', thu: '四', fri: '五', sat: '六', sun: '日' }[day] }}
            </span>
            <span v-if="!row.daysOfWeek?.length" style="color: var(--hope-text-muted); font-size: 13px;">—</span>
          </div>
        </template>
        <template #col-active="{ row }">
          <HopeBadge :color="row.active ? 'success' : 'error'">
            {{ row.active ? '启用' : '停用' }}
          </HopeBadge>
        </template>
        <template #col-__actions="{ row }">
          <div style="display: flex; gap: 6px;">
            <HopeBtn variant="text" size="sm" @click="handleEdit(row)">编辑</HopeBtn>
            <HopeBtn variant="text" size="sm" :class="'hope-btn--error'" style="color: var(--hope-error);" @click="handleDelete(row.id!)">删除</HopeBtn>
          </div>
        </template>
      </HopeTable>
      <template #footer>
        <div v-if="filteredRules.length === 0" style="text-align: center; padding: 40px 20px; color: var(--hope-text-muted);">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" style="opacity: 0.4; margin-bottom: 12px; display: block; margin-left: auto; margin-right: auto;">
            <circle cx="12" cy="12" r="10"/><path d="M8 15s1.5-2 4-2 4 2 4 2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/>
          </svg>
          <div>暂无用药规则，点击"添加用药规则"创建第一条规则</div>
        </div>
      </template>
    </HopeCard>

    <!-- Create/Edit Modal -->
    <HopeModal
      :model-value="showDialog"
      :title="form.id ? '编辑用药规则' : '添加用药规则'"
      @update:model-value="showDialog = $event"
    >
      <div style="display: flex; flex-direction: column; gap: 18px;">
        <div class="hope-field" :class="{ 'focused': formFocused === 'pillType', 'has-value': form.pillType }">
          <label class="hope-label">药品名称</label>
          <HopeInput
            v-model="form.pillType"
            placeholder="如：降压药"
            size="lg"
          />
        </div>
        <div class="hope-field" :class="{ 'focused': formFocused === 'doseCount', 'has-value': form.doseCount && form.doseCount > 0 }">
          <label class="hope-label">剂量</label>
          <HopeInput
            :model-value="String(form.doseCount ?? 1)"
            type="number"
            placeholder="如：1"
            size="lg"
            @update:model-value="form.doseCount = parseInt($event) || 1"
          >
            <template #suffix><span style="color: var(--hope-text-muted); font-size: 13px;">粒</span></template>
          </HopeInput>
        </div>
        <div class="hope-field" :class="{ 'focused': formFocused === 'scheduleTime', 'has-value': form.scheduleTime }">
          <label class="hope-label">服用时间</label>
          <el-time-picker
            v-model="form.scheduleTime"
            format="HH:mm"
            placeholder="选择时间"
            value-format="HH:mm"
            style="width: 100%;"
            @focus="formFocused = 'scheduleTime'"
            @blur="formFocused = ''"
          />
        </div>
        <div class="hope-field" :class="{ 'has-value': form.daysOfWeek?.length }">
          <label class="hope-label">执行周期</label>
          <el-select
            v-model="form.daysOfWeek"
            multiple
            placeholder="请选择执行周期"
            style="width: 100%;"
          >
            <el-option v-for="day in dayOptions" :key="day.value" :label="day.label" :value="day.value" />
          </el-select>
        </div>
        <div class="hope-field">
          <label class="hope-label">启用状态</label>
          <el-switch v-model="form.active" active-text="启用" inactive-text="停用" />
        </div>
      </div>
      <template #footer>
        <HopeBtn variant="plain" @click="showDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="saveRule">保存</HopeBtn>
      </template>
    </HopeModal>
  </div>
</template>

<script setup lang='ts'>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import type { MedicationRule } from '@/types'
import { medicationApi } from '@/api/medication'
import {
  HopeCard,
  HopeBtn,
  HopeInput,
  HopeTable,
  HopeBadge,
  HopeStatCard,
  HopeModal,
} from '@/components/hope'

// State
const loading = ref({ rules: true })
const showDialog = ref(false)
const searchQuery = ref('')
const formFocused = ref('')

// Form state for create/edit (using camelCase for local state)
const form = ref<Omit<MedicationRule, 'id' | 'elderly_id' | 'created_at'> & { id?: string; active?: boolean }>({
  pill_type: '',
  dose_count: 1,
  schedule_time: '',
  days_of_week: [],
  active: true,
})

// Rules data
const rules = ref<MedicationRule[]>([])

// Selected elderly (from route or store)
const currentElder = ref<any>({ name: '张大爷', id: 'elderly_123' })

// Statistics
const stats = computed(() => {
  const now = new Date()
  const currentHour = now.getHours()
  const active = rules.value.filter(r => r.active === true).length
  const missed = rules.value.filter(r => {
    const scheduleTime = (r as any).scheduleTime || r.schedule_time
    return scheduleTime && parseInt(scheduleTime.split(':')[0]) < currentHour
  }).length
  return {
    activeRules: active,
    missedCount: missed,
    adherenceRate: 85,
    pendingActions: 3
  }
})

// Day options for dropdown
const dayOptions = [
  { value: 'mon', label: '周一' },
  { value: 'tue', label: '周二' },
  { value: 'wed', label: '周三' },
  { value: 'thu', label: '周四' },
  { value: 'fri', label: '周五' },
  { value: 'sat', label: '周六' },
  { value: 'sun', label: '周日' }
]

// Period labels helper
function periodLabels(days: string[]): string {
  if (days.length === 7) return '每天'
  return days.map(d => ({ mon: '一', tue: '二', wed: '三', thu: '四', fri: '五', sat: '六', sun: '日' })[d]).join('、')
}

// Table columns definition for HopeTable
const tableColumns = [
  { prop: 'pillType', label: '药品名称', sortable: false },
  { prop: 'scheduleTime', label: '服用时间', sortable: false },
  { prop: 'doseCount', label: '剂量', sortable: false },
  { prop: 'daysOfWeek', label: '执行周期', sortable: false },
  { prop: 'active', label: '状态', sortable: false },
  { prop: '__actions', label: '操作', sortable: false },
]

// Filtered rules based on search
const filteredRules = computed(() => {
  if (!searchQuery.value) return rules.value
  return rules.value.filter(r => r.pillType && r.pillType.toLowerCase().includes(searchQuery.value.toLowerCase()))
})

// Load data on mount
onMounted(async () => {
  loading.value.rules = true
  try {
    const res = await medicationApi.listRules(currentElder.value.id as string)
    if (res.data && Array.isArray(res.data)) {
      rules.value = res.data.map((r: any) => ({
        ...r,
        daysOfWeek: r.daysOfWeek || []
      }))
    }
  } catch (error) {
    console.error('Failed to load medication rules:', error)
    ElMessage.error('加载用药规则失败，请重试')
  } finally {
    loading.value.rules = false
  }
})

// Open create dialog
function openCreateDialog() {
  form.value = {
    pill_type: '',
    dose_count: 1,
    schedule_time: '',
    days_of_week: [],
    active: true,
  } as Omit<MedicationRule, 'id' | 'elderly_id' | 'created_at'> & { id?: string }
  showDialog.value = true
}

// Handle edit
function handleEdit(rule: MedicationRule) {
  form.value = {
    pill_type: rule.pill_type,
    dose_count: rule.dose_count,
    schedule_time: rule.schedule_time,
    days_of_week: rule.days_of_week || [],
    active: rule.active
  }
  showDialog.value = true
}

// Handle delete
async function handleDelete(id: string) {
  const confirmText = await ElMessageBox.confirm(`确定要删除用药规则 "${id}" 吗？`, '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).catch(() => 'cancel')

  if (confirmText === 'cancel') return

  try {
    await medicationApi.deleteRule(currentElder.value.id as string, id)
    rules.value = rules.value.filter(r => r.id !== id)
    ElMessage.success('删除成功')
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// Save rule (create or update)
async function saveRule() {
  if (!form.value.pillType || !form.value.scheduleTime) {
    ElNotification({
      title: '必填项缺失',
      message: '药品名称和服用时间为必填字段',
      type: 'error'
    })
    return
  }

  loading.value.rules = true
  try {
    if (form.value.id) {
      await medicationApi.updateRule(currentElder.value.id as string, form.value.id!, form.value)
      ElMessage.success('更新成功')
    } else {
      await medicationApi.createRule(currentElder.value.id as string, form.value)
      ElMessage.success('创建成功')
    }
    const res = await medicationApi.listRules(currentElder.value.id as string)
    if (res.data && Array.isArray(res.data)) {
      rules.value = res.data.map((r: any) => ({
        ...r,
        daysOfWeek: r.daysOfWeek || []
      }))
    }
  } catch (error) {
    console.error('Failed to save medication rule:', error)
    ElMessage.error('保存失败，请检查输入')
  } finally {
    loading.value.rules = false
    showDialog.value = false
  }
}

// Handle search
function handleSearch() {
  // Filtering handled by computed property
}
</script>

<style scoped>
.medication-page {
  padding: 0;
}

/* Stat card value override for computed display */
.hope-stat-card__value {
  font-size: 30px;
  font-weight: 800;
  color: var(--hope-text);
  line-height: 1.1;
  margin-bottom: 4px;
  letter-spacing: -0.03em;
}

/* Responsive */
@media (max-width: 1200px) {
  .hope-grid-4 { grid-template-columns: repeat(2, 1fr) !important; }
}

@media (max-width: 768px) {
  .hope-grid-4 { grid-template-columns: 1fr !important; }
  .hope-page-header { flex-direction: column; align-items: flex-start; gap: 12px; }
}
</style>
