<template>
  <div class="institutions-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item>首页</el-breadcrumb-item>
          <el-breadcrumb-item>B2B 对接</el-breadcrumb-item>
          <el-breadcrumb-item>机构管理</el-breadcrumb-item>
        </el-breadcrumb>
        <h2 class="page-title">机构管理</h2>
      </div>
      <HopeBtn variant="filled" @click="showDialog = true">
        <template #icon>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </template>
        新增机构
      </HopeBtn>
    </div>

    <!-- KPI Row -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="never" class="kpi-card kpi-primary">
          <div class="kpi-value">{{ total }}</div>
          <div class="kpi-label">机构总数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="kpi-card kpi-success">
          <div class="kpi-value">{{ activeCount }}</div>
          <div class="kpi-label">已激活</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="kpi-card kpi-warning">
          <div class="kpi-value">{{ pendingCount }}</div>
          <div class="kpi-label">待审核</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="kpi-card kpi-info">
          <div class="kpi-value">{{ apiKeyCount }}</div>
          <div class="kpi-label">API 密钥</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter Bar -->
    <HopeCard class="filter-card">
      <template #default>
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
          <HopeBtn variant="plain" size="sm" @click="resetSearch">重置</HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="loadInstitutions">搜索</HopeBtn>
        </div>
      </template>
    </HopeCard>

    <!-- Institution Table -->
    <HopeCard title="机构列表" class="table-card">
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
            <HopeBadge :color="row.access_level === 'read_write' ? 'success' : (row.access_level === 'emergency_only' ? 'error' : 'primary')">
              {{ getAccessLevelLabel(row.access_level) }}
            </HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <HopeBadge :color="statusBadgeColor(row.status)">{{ getStatusLabel(row.status) }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="120">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <HopeBtn variant="text" size="sm" @click.stop="viewDetail(row)">详情</HopeBtn>
            <HopeBtn variant="text" size="sm" @click.stop="generateKey(row)">生成密钥</HopeBtn>
            <HopeBtn variant="warning" size="sm" :disabled="row.status !== 'active'" @click.stop="toggleStatus(row)">
              停用
            </HopeBtn>
            <HopeBtn variant="success" size="sm" :disabled="row.status === 'active'" @click.stop="toggleStatus(row)">
              启用
            </HopeBtn>
            <HopeBtn variant="text" size="sm" @click.stop="deleteInstitution(row)">删除</HopeBtn>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
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
      </template>
    </HopeCard>

    <!-- Side Panel (Detail) -->
    <div class="side-panel-overlay" :class="{ show: detailPanelOpen }" @click="detailPanelOpen = false" />
    <div class="side-panel" :class="{ open: detailPanelOpen }">
      <div class="panel-header">
        <span class="panel-title">机构详情</span>
        <HopeBtn variant="ghost" size="sm" icon-only @click="detailPanelOpen = false">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </HopeBtn>
      </div>
      <div class="panel-body" v-if="detailData">
        <div class="inst-header">
          <div class="inst-icon large" :class="'type-' + detailData.type">{{ typeEmoji(detailData.type) }}</div>
          <div>
            <div style="font-size:18px;font-weight:700;color:var(--hope-text);">{{ detailData.name }}</div>
            <div style="font-size:12px;color:var(--hope-text-muted);">{{ detailData.code }}</div>
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
            <HopeBadge :color="statusBadgeColor(detailData.status)">{{ getStatusLabel(detailData.status) }}</HopeBadge>
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
            <HopeBtn variant="filled" size="sm" @click="generateKey(detailData)">生成 API 密钥</HopeBtn>
            <HopeBtn variant="outlined" size="sm" @click="toggleStatus(detailData)">{{ detailData.status === 'active' ? '停用' : '启用' }}</HopeBtn>
            <HopeBtn variant="plain" size="sm" @click="deleteInstitution(detailData)">删除机构</HopeBtn>
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
        <HopeBtn variant="plain" size="sm" @click="showDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" size="sm" @click="handleAdd">确认添加</HopeBtn>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { institutionsApi, type B2BInstitution } from '@/api/institutions'
import { HopeCard, HopeBtn, HopeBadge } from '@/components/hope'

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

function statusBadgeColor(status: string): 'success' | 'warning' | 'primary' {
  if (status === 'active') return 'success'
  if (status === 'pending') return 'warning'
  return 'primary'
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
  institutionsApi.generateApiKey(row.id, row.name, 365)
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

/* Page header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 16px;
}
.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 8px 0 0;
  letter-spacing: -0.02em;
}

/* KPI Cards */
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
.kpi-warning .kpi-value { color: var(--hope-warning); }
.kpi-info .kpi-value { color: var(--hope-info); }

/* Filter card */
.filter-card :deep(.hope-content-card__body) {
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

/* Table card */
.table-card {
  margin-bottom: 0;
}
.table-card :deep(.hope-card-header__title) {
  color: var(--hope-text);
  font-weight: 700;
}

.inst-table {
  width: 100%;
}
.inst-table :deep(.el-table__row) {
  cursor: pointer;
}
.inst-table :deep(.el-table__row:hover) {
  background-color: rgba(58,87,232,0.04) !important;
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
  border-radius: var(--hope-radius-md);
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
  color: var(--hope-text);
}
.inst-code {
  font-size: 11px;
  color: var(--hope-text-muted);
  font-family: monospace;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

/* Pagination */
.table-card :deep(.hope-content-card__footer) {
  display: flex;
  justify-content: flex-end;
  padding: 14px 22px;
  border-top: 1px solid var(--hope-border);
}

/* ========== Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(26,26,46,0.3);
  backdrop-filter: blur(4px);
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
  background: var(--hope-surface);
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(58,87,232,0.10);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: var(--hope-surface);
  z-index: 1;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
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
  border-radius: var(--hope-radius-lg);
}

.panel-section {
  margin-bottom: 20px;
}
.panel-section-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--hope-border);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}
.panel-row-label {
  color: var(--hope-text-muted);
}
.panel-row-value {
  font-weight: 600;
  color: var(--hope-text);
}
.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* Responsive */
@media (max-width: 768px) {
  .institutions-page :deep(.el-col) { width: 100% !important; flex: 0 0 100% !important; }
  .institutions-page :deep(.el-table) { font-size: 12px; }
  .institutions-page :deep(.el-table th),
  .institutions-page :deep(.el-table td) { padding: 6px 4px; }
}
</style>
