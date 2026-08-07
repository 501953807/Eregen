<template>
  <div class="institutions-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">机构管理</h2>
      <el-button type="primary" @click="showDialog = true" size="default">+ 新增机构</el-button>
    </div>

    <!-- KPI Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ total }}</div>
          <div class="kpi-label">机构总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ activeCount }}</div>
          <div class="kpi-label">已激活</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-value">{{ pendingCount }}</div>
          <div class="kpi-label">待审核</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ apiKeyCount }}</div>
          <div class="kpi-label">API 密钥</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter Bar -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchForm.name" placeholder="搜索机构名称/编码..." clearable style="width: 240px;" />
        <el-select v-model="searchForm.type" placeholder="机构类型" clearable style="width: 150px;">
          <el-option label="医院" value="hospital" />
          <el-option label="社区" value="community_center" />
          <el-option label="养老院" value="nursing_home" />
          <el-option label="诊所" value="clinic" />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 130px;">
          <el-option label="已激活" value="active" />
          <el-option label="待审核" value="pending" />
          <el-option label="已停用" value="suspended" />
        </el-select>
        <span class="filter-spacer"></span>
        <el-button @click="resetSearch">重置</el-button>
        <el-button type="primary" @click="loadInstitutions">搜索</el-button>
      </div>
    </el-card>

    <!-- Institution Table -->
    <el-card shadow="never" class="table-card">
      <el-table
        :data="pagedInstitutions"
        stripe
        v-loading="loading"
        class="inst-table"
        @row-click="viewDetail"
        highlight-current-row
      >
        <el-table-column label="机构信息" min-width="200">
          <template #default="{ row }">
            <div class="inst-cell">
              <div class="inst-icon" :class="'type-' + row.type">
                {{ typeEmoji(row.type) }}
              </div>
              <div>
                <div class="inst-name">{{ row.name }}</div>
                <div class="inst-code">{{ row.code }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="联系人" width="100">
          <template #default="{ row }">{{ row.contact_name || '—' }}</template>
        </el-table-column>
        <el-table-column label="权限" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.access_level === 'read_write' ? 'success' : (row.access_level === 'emergency_only' ? 'danger' : '')">
              {{ getAccessLevelLabel(row.access_level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="statusClass(row.status)">
              <span class="status-dot" :class="statusClass(row.status)"></span>
              {{ getStatusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="120">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click.stop="viewDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click.stop="generateKey(row)">生成密钥</el-button>
            <el-button :link="true" :type="row.status === 'active' ? 'warning' : 'success'" size="small" @click.stop="toggleStatus(row)">
              {{ row.status === 'active' ? '停用' : '启用' }}
            </el-button>
            <el-button link type="danger" size="small" @click.stop="deleteInstitution(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
          :page-size="pagination.pageSize"
          :current-page="pagination.page"
          :page-sizes="[10, 20, 50]"
          @size-change="(v: number) => { pagination.pageSize = v; loadInstitutions(); }"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- Side Panel (Detail) -->
    <div class="side-panel-overlay" :class="{ show: detailPanelOpen }" @click="detailPanelOpen = false" />
    <div class="side-panel" :class="{ open: detailPanelOpen }">
      <div class="panel-header">
        <span class="panel-title">机构详情</span>
        <button class="panel-close" @click="detailPanelOpen = false">&#10005;</button>
      </div>
      <div class="panel-body" v-if="detailData">
        <div class="inst-header">
          <div class="inst-icon large" :class="'type-' + detailData.type">{{ typeEmoji(detailData.type) }}</div>
          <div>
            <div style="font-size:18px;font-weight:700;">{{ detailData.name }}</div>
            <div style="font-size:12px;color:var(--el-text-color-secondary);">{{ detailData.code }}</div>
          </div>
        </div>

        <div class="panel-section">
          <div class="panel-section-title">基本信息</div>
          <div class="panel-row"><span class="panel-row-label">机构编码</span><span class="panel-row-value mono">{{ detailData.code }}</span></div>
          <div class="panel-row"><span class="panel-row-label">类型</span><span class="panel-row-value">{{ getTypeLabel(detailData.type) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">联系人</span><span class="panel-row-value">{{ detailData.contact_name || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">联系电话</span><span class="panel-row-value">{{ detailData.contact_phone || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">访问权限</span><span class="panel-row-value">{{ getAccessLevelLabel(detailData.access_level) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">状态</span><span class="panel-row-value">
            <span class="status-badge" :class="statusClass(detailData.status)">
              <span class="status-dot" :class="statusClass(detailData.status)"></span>
              {{ getStatusLabel(detailData.status) }}
            </span>
          </span></div>
        </div>

        <div class="panel-section">
          <div class="panel-section-title">时间信息</div>
          <div class="panel-row"><span class="panel-row-label">创建时间</span><span class="panel-row-value">{{ formatDate(detailData.created_at) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">更新时间</span><span class="panel-row-value">{{ formatDate(detailData.updated_at) }}</span></div>
        </div>

        <div class="panel-section">
          <div class="panel-section-title">操作</div>
          <div class="panel-actions">
            <el-button size="small" type="primary" @click="generateKey(detailData)">生成 API 密钥</el-button>
            <el-button size="small" @click="toggleStatus(detailData)">{{ detailData.status === 'active' ? '停用' : '启用' }}</el-button>
            <el-button size="small" type="danger" plain @click="deleteInstitution(detailData)">删除机构</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Institution Dialog -->
    <el-dialog v-model="showDialog" title="新增机构" width="500px" destroy-on-close>
      <el-form :model="form" label-width="80px">
        <el-form-item label="机构名称" required>
          <el-input v-model="form.name" placeholder="如：上海市第一中心医院" />
        </el-form-item>
        <el-form-item label="机构编码" required>
          <el-input v-model="form.code" placeholder="如：SH-YXY-001" />
        </el-form-item>
        <el-form-item label="机构类型" required>
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%;">
            <el-option label="医院" value="hospital" />
            <el-option label="社区服务中心" value="community_center" />
            <el-option label="养老院" value="nursing_home" />
            <el-option label="诊所" value="clinic" />
          </el-select>
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form.contactName" placeholder="联系人姓名" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contactPhone" placeholder="联系电话" />
        </el-form-item>
        <el-form-item label="访问权限">
          <el-select v-model="form.accessLevel" style="width: 100%;">
            <el-option label="仅紧急" value="emergency_only" />
            <el-option label="只读" value="read" />
            <el-option label="读写" value="read_write" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { institutionsApi, type B2BInstitution } from '@/api/institutions'

const loading = ref(false)
const showDialog = ref(false)
const detailPanelOpen = ref(false)
const detailData = ref<B2BInstitution | null>(null)

const institutions = ref<B2BInstitution[]>([])

const total = computed(() => institutions.value.length)
const activeCount = computed(() => institutions.value.filter(i => i.status === 'active').length)
const pendingCount = computed(() => institutions.value.filter(i => i.status === 'pending').length)
const apiKeyCount = computed(() => institutions.value.length)

const searchForm = ref({ name: '', type: '', status: '' })
const pagination = ref({ page: 1, pageSize: 10, total: 0 })

const form = ref({
  name: '', code: '', type: 'hospital', contactName: '', contactPhone: '', accessLevel: 'read',
})

const B2B_BASE = import.meta.env.VITE_B2B_URL || 'http://localhost:8082/api/v2'

async function loadInstitutions() {
  loading.value = true
  try {
    const res = await institutionsApi.list({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      ...(searchForm.value.type ? { type: searchForm.value.type } : {}),
      ...(searchForm.value.status ? { status: searchForm.value.status } : {}),
    })
    institutions.value = (res.data as B2BInstitution[]) ?? []
    pagination.value.total = institutions.value.length
  } catch (err: any) {
    console.error('load institutions failed:', err)
    institutions.value = []
    pagination.value.total = 0
  } finally {
    loading.value = false
  }
}

function resetSearch() {
  searchForm.value = { name: '', type: '', status: '' }
  pagination.value.page = 1
  loadInstitutions()
}

function handlePageChange(page: number) {
  pagination.value.page = page
  loadInstitutions()
}

const filteredInstitutions = computed(() => {
  let list = institutions.value
  if (searchForm.value.name) {
    list = list.filter(i => i.name.includes(searchForm.value.name) || i.code.includes(searchForm.value.name))
  }
  return list
})

const pagedInstitutions = computed(() => {
  const start = (pagination.value.page - 1) * pagination.value.pageSize
  return filteredInstitutions.value.slice(start, start + pagination.value.pageSize)
})

function getTypeTagType(type: string): string {
  const map: Record<string, string> = { hospital: '', community_center: 'success', clinic: 'warning', nursing_home: 'info' }
  return map[type] || ''
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = { hospital: '医院', community_center: '社区', clinic: '诊所', nursing_home: '养老院' }
  return map[type] || type
}

function getAccessLevelLabel(level: string): string {
  const map: Record<string, string> = { read: '只读', read_write: '读写', emergency_only: '紧急' }
  return map[level] || level
}

function getStatusLabel(status: string): string {
  const map: Record<string, string> = { active: '已激活', pending: '待审核', suspended: '已停用' }
  return map[status] || status
}

function statusClass(status: string): string {
  const map: Record<string, string> = { active: 'status-active', pending: 'status-pending', suspended: 'status-suspended' }
  return map[status] || 'status-active'
}

function typeEmoji(type: string): string {
  const map: Record<string, string> = { hospital: '🏥', community_center: '🏘️', clinic: '💉', nursing_home: '🏠' }
  return map[type] || '🏢'
}

function formatDate(ts?: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleDateString('zh-CN')
}

function viewDetail(row: B2BInstitution) {
  detailData.value = { ...row }
  detailPanelOpen.value = true
}

function generateKey(row: B2BInstitution) {
  ElMessage.info('正在生成 API 密钥...')
  institutionsApi.generateApiKey(row.id, row.name)
    .then(res => {
      ElMessageBox.alert(`密钥值（请妥善保存，仅显示一次）：<br><code style="font-family:monospace;">${res.data?.key_value || ''}</code>`, 'API 密钥', {
        dangerouslyUseHTMLString: true,
        confirmButtonText: '已复制',
        type: 'success'
      }).catch(() => {})
    })
    .catch(err => {
      console.error('generate key failed:', err)
      ElMessage.error('生成密钥失败')
    })
}

function copyKey() {
  ElMessage.info('密钥已在生成弹窗中展示，请妥善保存')
}

function toggleStatus(row: B2BInstitution) {
  const newStatus = row.status === 'active' ? 'suspended' : 'active'
  institutionsApi.update(row.id, { status: newStatus })
    .then(() => {
      ElMessage.success(newStatus === 'active' ? '已启用' : '已停用')
      loadInstitutions()
    })
    .catch(err => {
      console.error('toggle status failed:', err)
      ElMessage.error('操作失败')
    })
}

async function deleteInstitution(row: B2BInstitution) {
  try {
    await ElMessageBox.confirm(`确定要删除机构「${row.name}」吗？此操作不可恢复。`, '警告', { type: 'warning' })
    await institutionsApi.delete(row.id)
    ElMessage.success('删除成功')
    loadInstitutions()
  } catch (err: any) {
    if (err !== 'cancel') {
      console.error('delete institution failed:', err)
      ElMessage.error('删除失败')
    }
  }
}

async function handleAdd() {
  if (!form.value.name || !form.value.code) {
    ElMessage.warning('请填写机构名称和编码')
    return
  }
  try {
    await institutionsApi.create({
      name: form.value.name,
      code: form.value.code,
      type: form.value.type,
      contact_name: form.value.contactName,
      contact_phone: form.value.contactPhone,
      access_level: form.value.accessLevel,
      status: 'pending'
    })
    ElMessage.success('创建成功')
    showDialog.value = false
    form.value = { name: '', code: '', type: 'hospital', contactName: '', contactPhone: '', accessLevel: 'read' }
    loadInstitutions()
  } catch (err) {
    console.error('create institution failed:', err)
    ElMessage.error('创建失败')
  }
}

onMounted(() => {
  loadInstitutions()
})
</script>

<style scoped>
.institutions-page {
  padding: 0;
}
.institutions-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.institutions-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.institutions-page :deep(.el-card__header) {
  background: linear-gradient(180deg, rgba(255,255,255,0.6) 0%, transparent 100%);
  border-bottom: 1px solid var(--el-border-color-lighter);
  padding: 16px 20px;
  border-radius: 12px 12px 0 0 !important;
}
.institutions-page :deep(.el-card__body) {
  padding: 20px;
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

/* KPI Cards */
.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
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
  color: var(--el-text-color-secondary);
  margin-top: 6px;
  font-weight: 600;
}
.kpi-blue .kpi-value { color: #5C8D73; }
.kpi-green .kpi-value { color: #6FAF8F; }
.kpi-warning .kpi-value { color: #D9A441; }
.kpi-purple .kpi-value { color: #7BAF8C; }

/* Filter card */
.filter-card :deep(.el-card__body) {
  padding: 12px 16px;
}
.filter-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.filter-spacer {
  flex: 1;
}

/* Table */
.table-card :deep(.el-card__header) {
  padding: 0;
}
.inst-table {
  width: 100%;
}
.inst-table :deep(.el-table__row) {
  cursor: pointer;
}
.inst-table :deep(.el-table__row:hover) {
  background-color: var(--el-fill-color-light) !important;
}

/* Institution cell with icon */
.inst-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.inst-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}
.inst-icon.type-hospital { background: #DBEAFE; }
.inst-icon.type-community_center { background: #F0FDF4; }
.inst-icon.type-clinic { background: #FFF7ED; }
.inst-icon.type-nursing_home { background: #F3E8FF; }
.inst-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.inst-code {
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
.status-active { background: #F0FDF4; color: #16A34A; }
.status-pending { background: #FFFBEB; color: #D97706; }
.status-suspended { background: #F3F4F6; color: #6B7280; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.status-active .status-dot { background: #16A34A; }
.status-pending .status-dot { background: #D97706; }
.status-suspended .status-dot { background: #6B7280; }

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
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
  transition: background 0.15s;
}
.panel-close:hover {
  background: var(--el-border-color-light);
}
.panel-body {
  padding: 20px 24px;
}

.inst-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}
.inst-header .inst-icon.large {
  width: 52px;
  height: 52px;
  font-size: 24px;
  border-radius: 14px;
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
.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
