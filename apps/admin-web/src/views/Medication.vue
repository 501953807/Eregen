<template>
  <div class="medication-page">
    <!-- Header -->
    <el-card shadow="hover" class="card-header">
      <div class="header-content">
        <h1>{{ currentElder ? `用药管理 — ${currentElder.name}` : '用药管理' }}</h1>
        <el-button type="primary" @click="openCreateDialog">+ 添加用药规则</el-button>
      </div>
    </el-card>

    <!-- KPI Cards -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-content">
            <div class="kpi-icon-wrap green">
              <el-icon :size="28"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="6" width="18" height="12" rx="2"/><path d="M8 6V4M16 6V4M3 10h18"/><circle cx="8" cy="14" r="1.5"/><circle cx="16" cy="14" r="1.5"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.activeRules }}</div>
              <div class="kpi-label">今日规则数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-content">
            <div class="kpi-icon-wrap orange">
              <el-icon :size="28"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.missedCount }}</div>
              <div class="kpi-label">今日漏服</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-content">
            <div class="kpi-icon-wrap blue">
              <el-icon :size="28"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.adherenceRate }}%</div>
              <div class="kpi-label">按时服药率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-danger">
          <div class="kpi-content">
            <div class="kpi-icon-wrap red">
              <el-icon :size="28"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4.8 2.3A.3.3 0 105 2H4a2 2 0 00-2 2v5a6 6 0 006 6 6 6 0 006-6V4a2 2 0 00-2-2h-1a.2.2 0 00.3.3"/><path d="M8 15v4M12 15v4M6 23h8"/></svg></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.pendingActions }}</div>
              <div class="kpi-label">待处理提醒</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Medication Rules Table -->
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">用药规则列表</span>
          <el-input
            v-model="searchQuery"
            placeholder="搜索药品名称..."
            clearable
            style="width: 200px;"
            @input="handleSearch"
          >
            <template #prefix><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></template>
          </el-input>
        </div>
      </template>

      <el-table :data="filteredRules" stripe v-loading="loading.rules">
        <el-table-column prop="pillType" label="药品名称" width="180"></el-table-column>
        <el-table-column prop="scheduleTime" label="服用时间" width="120"></el-table-column>
        <el-table-column prop="doseCount" label="剂量" width="80">
          <template #default="{ row }">
            {{ row.doseCount }} 粒
          </template>
        </el-table-column>
        <el-table-column prop="daysOfWeek" label="执行周期" width="140">
          <template #default="{ row }">
            {{ periodLabels(row.daysOfWeek) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.active ? 'success' : 'danger'">
              {{ row.active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button size="danger" @click="handleDelete(row.id)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="filteredRules.length === 0" class="empty-state">
        <el-icon :size="48" style="color: #9CA3AF; margin-bottom: 12px;">Empty</el-icon>
        <div>暂无用药规则，点击"添加用药规则"创建第一条规则</div>
      </div>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showDialog" title="用药规则" width="500px">
      <el-form :model="form" label-width="100px" label-position="right">
        <el-form-item label="药品名称">
          <el-input v-model="form.pillType" placeholder="如：降压药" />
        </el-form-item>
        <el-form-item label="剂量">
          <el-input-number v-model="form.doseCount" min="1" max="99" style="width: 100%;" placeholder="如：1" />
        </el-form-item>
        <el-form-item label="服用时间">
          <el-time-picker
            v-model="form.scheduleTime"
            format="HH:mm"
            placeholder="选择时间"
            value-format="HH:mm"
            style="width: 100%;"
          ></el-time-picker>
        </el-form-item>
        <el-form-item label="执行周期">
          <el-select v-model="form.daysOfWeek" multiple placeholder="请选择">
            <el-option v-for="day in dayOptions" :key="day.value" :label="day.label" :value="day.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="form.active" active-text="启用" inactive-text="停用"></el-switch>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang='ts'>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElNotification, ElLoading } from 'element-plus'
import type { MedicationRule } from '@/types'
import { medicationApi } from '@/api/medication'

// State
const loading = ref({ rules: true })
const showDialog = ref(false)
const searchQuery = ref('')

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
    if (res.data && Array.isArray(res.data.data)) {
      rules.value = res.data.data.map((r: any) => ({
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
  // Reset form
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
    if (showDialog.value && form.value.id) {
      // Update mode
      await medicationApi.updateRule(currentElder.value.id as string, form.value.id!, form.value)
      ElMessage.success('更新成功')
    } else {
      // Create mode
      await medicationApi.createRule(currentElder.value.id as string, form.value)
      ElMessage.success('创建成功')
    }
    // Reload data
    const res = await medicationApi.listRules(currentElder.value.id as string)
    if (res.data && Array.isArray(res.data.data)) {
      rules.value = res.data.data.map((r: any) => ({
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
.medication-page { padding: 0; }
.medication-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.medication-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}

.card-header {
  padding: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all var(--duration-normal) var(--easing);
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

.kpi-card :deep(.el-card__body) {
  padding: 16px 20px;
  border: 1px solid var(--border-light);
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), var(--shadow-card);
  border-radius: var(--radius-lg);
}

.kpi-content {
  display: flex;
  align-items: center;
  gap: 14px;
}

.kpi-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.kpi-icon-wrap.blue { background: linear-gradient(135deg, #5C8D73, #7BAF8C); color: #fff; }
.kpi-icon-wrap.green { background: linear-gradient(135deg, #6FAF8F, #8BC4A8); color: #fff; }
.kpi-icon-wrap.orange { background: linear-gradient(135deg, #D9A441, #E8BC6A); color: #fff; }
.kpi-icon-wrap.red { background: linear-gradient(135deg, #D77B72, #E09890); color: #fff; }
.kpi-icon { display: none; }
.kpi-info { flex: 1; }

.kpi-value {
  font-size: 32px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  letter-spacing: -0.03em;
  line-height: 1;
  margin-bottom: 4px;
}

.kpi-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #9CA3AF;
}

:deep(.el-table__header th) {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

:deep(.el-table__row td) {
  color: var(--el-text-color-secondary);
}

.el-form-item__label {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .kpi-card :deep(.el-card__body) {
    padding: 12px 16px;
  }

  .kpi-value {
    font-size: 22px;
  }
}
</style>
