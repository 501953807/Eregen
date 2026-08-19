<template>
  <div class="persons-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">统一人员管理</h2>
        <p class="page-subtitle">管理所有老人、家属和关联档案信息</p>
      </div>
      <HopeBtn variant="filled" size="md" @click="showCreateDialog = true">
        <template #icon>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </template>
        新增人员
      </HopeBtn>
    </div>

    <!-- Filter Bar — HopeCard -->
    <HopeCard>
      <template #header>
        <span class="filter-title">筛选条件</span>
      </template>
      <div class="filter-row">
        <div class="filter-item">
          <label class="filter-label">业务链</label>
          <el-select v-model="filters.business_chain" placeholder="全部" clearable class="hope-filter-select">
            <el-option label="自营链" value="self" />
            <el-option label="住院链" value="hospital" />
            <el-option label="社区链" value="community" />
          </el-select>
        </div>
        <div class="filter-item">
          <label class="filter-label">状态</label>
          <el-select v-model="filters.status" placeholder="全部" clearable class="hope-filter-select">
            <el-option label="活跃" value="active" />
            <el-option label="暂停" value="suspended" />
            <el-option label="已故" value="deceased" />
          </el-select>
        </div>
        <div class="filter-item filter-item--search">
          <label class="filter-label">搜索</label>
          <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable class="hope-filter-input" />
        </div>
        <div class="filter-actions">
          <HopeBtn variant="plain" size="sm" @click="resetFilters">重置</HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="fetchPersons">查询</HopeBtn>
        </div>
      </div>
    </HopeCard>

    <!-- Data Table — HopeTable -->
    <HopeCard class="table-card" style="margin-top: 16px;">
      <template #header>
        <span class="filter-title">人员列表</span>
        <span class="table-count">{{ pagination.total }} 条记录</span>
      </template>
      <HopeTable
        :columns="tableColumns"
        :data="persons"
        :loading="loading"
        :striped="true"
        :row-key="rowKeyFn"
      >
        <template #col-gender="{ row }">
          <span class="gender-cell">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}</span>
        </template>
        <template #col-phone="{ row }">
          <span class="phone-cell">{{ row.phone || '—' }}</span>
        </template>
        <template #col-status="{ row }">
          <HopeBadge :color="statusBadgeColor(row.status)">{{ statusLabel(row.status) }}</HopeBadge>
        </template>
        <template #col-business_chains="{ row }">
          <div class="chains-cell">
            <HopeBadge v-for="chain in getChains(row.id)" :key="chain" :color="chainBadgeColor(chain)" class="chain-badge">
              {{ chainLabel(chain) }}
            </HopeBadge>
          </div>
        </template>
        <template #col-created_at="{ row }">
          <span class="date-cell">{{ formatDate(row.created_at) }}</span>
        </template>
        <template #col-actions="{ row }">
          <div class="actions-cell">
            <HopeBtn variant="text" size="sm" @click="viewDetail(row)">详情</HopeBtn>
            <HopeBtn variant="text" size="sm" @click="editPerson(row)">编辑</HopeBtn>
          </div>
        </template>
      </HopeTable>

      <!-- Pagination -->
      <template #footer>
        <el-pagination
          background
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
          :current-page="pagination.page"
          :page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          @size-change="(v: number) => { pagination.pageSize = v; fetchPersons(); }"
          @current-change="(v: number) => { pagination.page = v; fetchPersons(); }"
        />
      </template>
    </HopeCard>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingPerson ? '编辑人员' : '新增人员'"
      width="500px"
      class="hope-dialog"
    >
      <el-form :model="form" label-width="100px">
        <el-form-item label="身份证号" required>
          <el-input v-model="form.id_card" :disabled="!!editingPerson" placeholder="请输入身份证号" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" style="width: 100%;">
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="出生日期">
          <el-date-picker
            v-model="form.birth_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="请选择出生日期"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="紧急联系人">
          <el-input v-model="form.emergency_contact" placeholder="请输入紧急联系人" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" type="textarea" :rows="2" placeholder="请输入地址" />
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" @click="showCreateDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="savePerson" :loading="saving">保存</HopeBtn>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { personApi } from '@/api/business-chains'
import type { Person } from '@/types'
import { HopeCard, HopeTable, HopeBtn, HopeBadge } from '@/components/hope'

const loading = ref(false)
const saving = ref(false)
const persons = ref<Person[]>([])
const showCreateDialog = ref(false)
const editingPerson = ref<Person | null>(null)

const filters = reactive({
  business_chain: '',
  status: '',
  search: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  id_card: '',
  name: '',
  gender: 0 as 0 | 1 | 2,
  birth_date: '',
  phone: '',
  emergency_contact: '',
  address: '',
})

const chainMap: Record<string, string> = {
  self: '自营',
  hospital: '住院',
  community: '社区',
}

const getChains = (personId: string) => {
  // In production, this would fetch from person_profiles
  return ['self'] as string[]
}

const rowKeyFn = (row: Person) => row.id
const tableColumns: Array<{ prop: string; label: string }> = [
  { prop: 'id_card', label: '身份证号' },
  { prop: 'name', label: '姓名' },
  { prop: 'gender', label: '性别' },
  { prop: 'phone', label: '电话' },
  { prop: 'status', label: '状态' },
  { prop: 'business_chains', label: '业务链' },
  { prop: 'created_at', label: '创建时间' },
  { prop: 'actions', label: '操作' },
]

const chainLabel = (chain: string) => chainMap[chain] || chain
const chainBadgeColor = (chain: string): 'primary' | 'accent' | 'info' => {
  if (chain === 'self') return 'primary'
  if (chain === 'hospital') return 'accent'
  return 'info'
}
const statusBadgeColor = (status: string): 'success' | 'warning' | 'error' => {
  if (status === 'active') return 'success'
  if (status === 'suspended') return 'warning'
  return 'error'
}
const statusLabel = (status: string) => {
  if (status === 'active') return '活跃'
  if (status === 'suspended') return '暂停'
  return '已故'
}
const formatDate = (dateStr?: string): string => {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const fetchPersons = async () => {
  loading.value = true
  try {
    const res = await personApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      business_chain: filters.business_chain,
      status: filters.status,
    })
    if (res.data) {
      persons.value = (res as unknown as { data: Person[]; page: number; page_size: number }).data
    }
  } catch (e) {
    console.error('Failed to fetch persons:', e)
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.business_chain = ''
  filters.status = ''
  filters.search = ''
  pagination.page = 1
  fetchPersons()
}

const viewDetail = (person: Person) => {
  // Navigate to detail page
  console.log('View detail:', person)
}

const editPerson = (person: Person) => {
  editingPerson.value = person
  form.id_card = person.id_card
  form.name = person.name
  form.gender = person.gender
  form.birth_date = person.birth_date || ''
  form.phone = person.phone || ''
  form.emergency_contact = person.emergency_contact || ''
  form.address = person.address || ''
  showCreateDialog.value = true
}

const savePerson = async () => {
  if (!form.id_card || !form.name) {
    ElMessage.warning('请填写必填项')
    return
  }
  saving.value = true
  try {
    if (editingPerson.value) {
      await personApi.update(editingPerson.value.id, form)
      ElMessage.success('更新成功')
    } else {
      await personApi.create(form)
      ElMessage.success('创建成功')
    }
    showCreateDialog.value = false
    editingPerson.value = null
    resetForm()
    fetchPersons()
  } catch (e) {
    console.error('Failed to save person:', e)
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  form.id_card = ''
  form.name = ''
  form.gender = 0
  form.birth_date = ''
  form.phone = ''
  form.emergency_contact = ''
  form.address = ''
}

onMounted(() => {
  fetchPersons()
})
</script>

<style scoped>
.persons-page {
  padding: 0;
}

/* ─── Page Header ─── */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
}

.page-header__left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
  font-weight: 500;
}

/* ─── Filter Bar ─── */
.filter-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--hope-text);
}

.table-count {
  font-size: 13px;
  color: var(--hope-text-muted);
  font-weight: 500;
}

.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  flex: 1;
  min-width: 140px;
}

.filter-item--search {
  flex: 2;
  min-width: 220px;
}

.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
  padding-bottom: 1px;
}

/* Hope UI Select/Input overrides */
:deep(.hope-filter-select) {
  width: 100%;
}
:deep(.hope-filter-select .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
  padding: 5px 11px !important;
}
:deep(.hope-filter-select .el-input__wrapper:hover) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}
:deep(.hope-filter-select .el-input__wrapper.is-focus) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}

:deep(.hope-filter-input .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
}
:deep(.hope-filter-input .el-input__wrapper:hover) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}
:deep(.hope-filter-input .el-input__wrapper.is-focus) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}

/* ─── Table Card ─── */
.table-card :deep(.hope-content-card__body) {
  padding: 0;
}

/* ─── Table Cells ─── */
.gender-cell {
  font-size: 14px;
  color: var(--hope-text);
  font-weight: 500;
}

.phone-cell {
  font-size: 13px;
  color: var(--hope-text-secondary);
  font-family: monospace;
}

.chains-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.chain-badge {
  font-size: 12px;
}

.date-cell {
  font-size: 13px;
  color: var(--hope-text-muted);
}

.actions-cell {
  display: flex;
  gap: 4px;
}

/* ─── Pagination ─── */
:deep(.hope-content-card__footer) {
  justify-content: flex-end;
  padding: 14px 22px;
}

:deep(.el-pagination) {
  --el-pagination-button-bg-color: var(--hope-surface);
  --el-pagination-button-border-radius: var(--hope-radius-md);
}

:deep(.el-pagination .btn-prev),
:deep(.el-pagination .btn-next),
:deep(.el-pagination .el-pager li) {
  border-radius: var(--hope-radius-md);
  border: 1px solid var(--hope-border);
  background: var(--hope-surface);
  color: var(--hope-text-secondary);
  font-weight: 600;
}

:deep(.el-pagination .el-pager li.active) {
  background: var(--hope-primary);
  border-color: var(--hope-primary);
  color: #fff;
}

:deep(.el-pagination .el-pager li:hover) {
  color: var(--hope-primary);
}

/* ─── Dialog ─── */
:deep(.hope-dialog .el-dialog) {
  border-radius: var(--hope-radius-xl) !important;
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-lg) !important;
}

:deep(.hope-dialog .el-dialog__header) {
  padding: 20px 24px 16px !important;
  border-bottom: 1px solid var(--hope-border) !important;
  margin-right: 0 !important;
}

:deep(.hope-dialog .el-dialog__title) {
  font-size: 16px !important;
  font-weight: 700 !important;
  color: var(--hope-text) !important;
}

:deep(.hope-dialog .el-dialog__body) {
  padding: 20px 24px !important;
}

:deep(.hope-dialog .el-dialog__footer) {
  padding: 16px 24px 20px !important;
  border-top: 1px solid var(--hope-border) !important;
}

:deep(.hope-dialog .el-form-item__label) {
  font-weight: 600 !important;
  color: var(--hope-text-secondary) !important;
}

:deep(.hope-dialog .el-input__wrapper),
:deep(.hope-dialog .el-select .el-input__wrapper),
:deep(.hope-dialog .el-date-editor.el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
}

/* ─── Responsive ─── */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
  }

  .filter-row {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-item, .filter-item--search {
    min-width: 100%;
  }

  .filter-actions {
    justify-content: flex-end;
  }

  .actions-cell {
    flex-direction: column;
    gap: 2px;
  }
}
</style>
