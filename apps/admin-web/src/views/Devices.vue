<template>
  <div class="devices-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">设备管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleRegister" size="default">+ 注册设备</el-button>
        <el-button @click="handleRefresh">刷新</el-button>
      </div>
    </div>

    <!-- KPI Cards (5 columns) -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="5">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-num">{{ deviceStore.total }}</div>
          <div class="kpi-label">设备总数</div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="kpi-card kpi-online">
          <div class="kpi-num">{{ stats.online_devices }}</div>
          <div class="kpi-label">在线</div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="kpi-card kpi-offline">
          <div class="kpi-num">{{ stats.offline_devices }}</div>
          <div class="kpi-label">离线</div>
        </el-card>
      </el-col>
      <el-col :span="5">
        <el-card shadow="hover" class="kpi-card kpi-upgrade">
          <div class="kpi-num">{{ stats.outdated_firmware }}</div>
          <div class="kpi-label">待升级</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-fault">
          <div class="kpi-num">{{ stats.fault_count }}</div>
          <div class="kpi-label">故障</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <span class="filter-label">筛选：</span>
      <el-select v-model="filters.type" placeholder="全部类型" clearable filterable class="filter-select">
        <el-option label="手环 Starter" value="bracelet-starter" />
        <el-option label="手环 Plus" value="bracelet-plus" />
        <el-option label="手环 Pro" value="bracelet-pro" />
        <el-option label="药盒 Basic" value="pillbox-basic" />
        <el-option label="药盒 Smart" value="pillbox-smart" />
        <el-option label="药盒 Auto" value="pillbox-auto" />
      </el-select>
      <el-select v-model="filters.status" placeholder="全部状态" clearable class="filter-select">
        <el-option label="在线" value="online" />
        <el-option label="离线" value="offline" />
      </el-select>
      <el-select v-model="filters.mode" placeholder="全部模式" clearable class="filter-select">
        <el-option label="家属APP" value="family" />
        <el-option label="管理后台" value="admin" />
        <el-option label="社区老人" value="community" />
        <el-option label="医疗腕带" value="medical" />
      </el-select>
      <span class="filter-spacer"></span>
      <el-input v-model="filters.search" placeholder="搜索设备ID、名称、老人姓名..." clearable class="filter-search" />
      <el-button @click="handleReset">重置</el-button>
      <el-button type="primary" @click="handleSearch">搜索</el-button>
    </div>

    <!-- Bulk Selection Banner -->
    <div class="bulk-banner" :class="{ show: selectedIds.length > 0 }">
      <el-checkbox :model-value="allSelected" :model-enabled="allSelected" @change="toggleSelectAll" />
      <span><strong class="bulk-count">{{ selectedIds.length }}</strong> 项已选中</span>
      <div class="bulk-actions">
        <el-button size="small" @click="handleBatchOta">批量OTA</el-button>
        <el-button size="small" @click="handleBatchConfig">批量配置</el-button>
        <el-button size="small" type="danger" plain @click="handleBatchUnbind">批量注销</el-button>
        <el-button size="small" @click="clearSelection">取消选择</el-button>
      </div>
    </div>

    <!-- Device Table -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="table-toolbar">
          <span class="table-title">设备列表</span>
          <div class="table-actions">
            <el-button size="small" @click="exportDevices">导出CSV</el-button>
            <el-button size="small" @click="handleRefresh">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="deviceStore.loading"
        :data="filteredDevices"
        stripe
        class="device-table"
        @selection-change="handleSelectionChange"
        @row-click="handleRowClick"
        highlight-current-row
      >
        <el-table-column type="selection" width="40" :selectable="row => row.device_type !== 'pillbox_basic'" />
        <el-table-column label="设备信息" min-width="160">
          <template #default="{ row }">
            <div class="device-cell">
              <div class="device-thumb" :class="row.device_type === 'bracelet' ? 'thumb-bracelet' : 'thumb-pillbox'">
                {{ row.device_type === 'bracelet' ? '📱' : '💊' }}
              </div>
              <div>
                <div class="device-name">{{ deviceLabel(row) }}</div>
                <div class="device-model">{{ chipLabel(row) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="设备ID" width="130">
          <template #default="{ row }">
            <span class="mono">{{ row.device_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="statusClass(row.status)">
              <span class="status-dot" :class="statusClass(row.status)"></span>
              {{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="固件" width="100">
          <template #default="{ row }">
            <span class="version-tag" :class="{ outdated: isOutdated(row) }">
              {{ row.firmware_version || '—' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="绑定老人" width="100">
          <template #default="{ row }">
            {{ row.owner_name || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="最后在线" width="110">
          <template #default="{ row }">
            {{ formatLastSeen(row.last_seen) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="action-links">
              <a class="action-link" @click.stop="handleOTA(row)">OTA升级</a>
              <a class="action-link" @click.stop="handleConfig(row)">配置</a>
              <a class="action-link danger" @click.stop="handleUnbind(row)">解绑</a>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next, jumper"
          :total="deviceStore.total"
          :page-size="filters.pageSize"
          :current-page="filters.page"
          :page-sizes="[10, 20, 50, 100]"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- Side Panel (Device Detail) -->
    <div class="side-panel-overlay" :class="{ show: panelOpen }" @click="closePanel" />
    <div class="side-panel" :class="{ open: panelOpen }">
      <div class="panel-header">
        <span class="panel-title">设备详情</span>
        <button class="panel-close" @click="closePanel">&#10005;</button>
      </div>
      <div class="panel-body" v-if="panelDevice">
        <!-- Device Header -->
        <div class="panel-device-header">
          <div class="panel-device-icon" :class="panelDevice.device_type === 'bracelet' ? 'icon-bracelet' : 'icon-pillbox'">
            {{ panelDevice.device_type === 'bracelet' ? '📱' : '💊' }}
          </div>
          <div>
            <div class="panel-device-name">{{ deviceLabel(panelDevice) }}</div>
            <div class="panel-device-id">{{ panelDevice.device_id }}</div>
          </div>
        </div>

        <!-- Basic Info -->
        <div class="panel-section">
          <div class="panel-section-title">基本信息</div>
          <div class="panel-row"><span class="panel-row-label">型号芯片</span><span class="panel-row-value">{{ chipLabel(panelDevice) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">固件版本</span><span class="panel-row-value">
            <span class="version-tag" :class="{ outdated: isOutdated(panelDevice) }">{{ panelDevice.firmware_version || '—' }}</span>
          </span></div>
          <div class="panel-row"><span class="panel-row-label">注册时间</span><span class="panel-row-value">{{ formatDate(panelDevice.created_at) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">绑定老人</span><span class="panel-row-value">{{ panelDevice.owner_name || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">所属机构</span><span class="panel-row-value">{{ panelDevice.institution || '—' }}</span></div>
          <div class="panel-row"><span class="panel-row-label">运行模式</span><span class="panel-row-value">{{ modeLabel(panelDevice.mode) }}</span></div>
        </div>

        <!-- Real-time Status -->
        <div class="panel-section">
          <div class="panel-section-title">实时状态</div>
          <div class="panel-row"><span class="panel-row-label">连接状态</span><span class="panel-row-value">
            <span class="status-badge" :class="statusClass(panelDevice.status)">
              <span class="status-dot" :class="statusClass(panelDevice.status)"></span>
              {{ statusLabel(panelDevice.status) }}
            </span>
          </span></div>
          <div class="panel-row"><span class="panel-row-label">信号强度</span><span class="panel-row-value">{{ signalStrength(panelDevice) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">电量</span><span class="panel-row-value">{{ panelDevice.battery_pct ?? '—' }}%</span></div>
          <div class="panel-row"><span class="panel-row-label">最后心跳</span><span class="panel-row-value">{{ formatLastSeen(panelDevice.last_seen) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">最近定位</span><span class="panel-row-value" style="color:var(--el-color-primary);cursor:pointer;" @click="goToMap">查看地图 →</span></div>
        </div>

        <!-- OTA Progress -->
        <div class="panel-section" v-if="panelDevice.ota_progress != null">
          <div class="panel-section-title">OTA 升级进度</div>
          <div class="panel-progress">
            <div class="progress-header">
              <span>{{ otaStatusText(panelDevice.ota_status) }}</span>
              <strong>{{ panelDevice.ota_progress }}%</strong>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" :class="panelDevice.ota_status === 'downloading' ? 'running' : 'success'" :style="{ width: panelDevice.ota_progress + '%' }"></div>
            </div>
            <div class="progress-meta" v-if="panelDevice.ota_speed">
              速度 {{ panelDevice.ota_speed }} · 预计剩余 {{ panelDevice.ota_eta || '未知' }}
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="panel-actions">
          <el-button size="default" @click="handleOTA(panelDevice)" style="flex:1;">OTA升级</el-button>
          <el-button size="default" @click="handleConfig(panelDevice)" style="flex:1;">远程配置</el-button>
          <el-button size="default" @click="handleReboot(panelDevice)" style="flex:1;">远程重启</el-button>
          <el-button size="default" type="danger" plain @click="handleUnbind(panelDevice)" style="flex:1;">注销设备</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Upload } from '@element-plus/icons-vue'
import { useDeviceStore } from '@/stores/device'
import { devicesApi } from '@/api/devices'
import type { Device } from '@/types'

const deviceStore = useDeviceStore()

// Filters
const filters = ref({
  type: '',
  status: '',
  mode: '',
  search: '',
  page: 1,
  pageSize: 20,
})

const filteredDevices = computed(() => {
  let list = deviceStore.devices
  if (filters.value.status) list = list.filter(d => d.status === filters.value.status)
  if (filters.value.type) list = list.filter(d => `${d.device_type}-${d.tier}` === filters.value.type)
  if (filters.value.mode) list = list.filter(d => d.mode === filters.value.mode)
  if (filters.value.search) {
    const q = filters.value.search.toLowerCase()
    list = list.filter(d =>
      d.device_id.toLowerCase().includes(q) ||
      (d.owner_name && d.owner_name.toLowerCase().includes(q))
    )
  }
  return list
})

// Stats
const latestFw = 'v2.4.1'
const stats = computed(() => ({
  online_devices: deviceStore.devices.filter(d => d.status === 'online').length,
  offline_devices: deviceStore.devices.filter(d => d.status !== 'online').length,
  outdated_firmware: deviceStore.devices.filter(d => d.firmware_version && d.firmware_version !== latestFw).length,
  fault_count: deviceStore.devices.filter(d => d.fault === true).length,
}))

// Selection
const selectedIds = ref<string[]>([])
const allSelected = computed(() => {
  const selectable = filteredDevices.value.filter(d => d.device_type !== 'pillbox_basic')
  return selectable.length > 0 && selectable.every(d => selectedIds.value.includes(d.id))
})

function handleSelectionChange(rows: Device[]) {
  selectedIds.value = rows.map(r => r.id)
}
function toggleSelectAll(val: boolean) {
  if (val) {
    const ids = filteredDevices.value.filter(d => d.device_type !== 'pillbox_basic').map(d => d.id)
    selectedIds.value = [...new Set([...selectedIds.value, ...ids])]
  } else {
    const removable = new Set(filteredDevices.value.filter(d => d.device_type !== 'pillbox_basic').map(d => d.id))
    selectedIds.value = selectedIds.value.filter(id => !removable.has(id))
  }
}
function clearSelection() {
  selectedIds.value = []
}

// Helpers
function deviceLabel(d: Device): string {
  const labels: Record<string, Record<string, string>> = {
    bracelet: { starter: '手环 Starter', plus: '手环 Plus', pro: '手环 Pro' },
    pillbox: { basic: '药盒 Basic', smart: '药盒 Smart', auto: '药盒 Auto' },
  }
  return labels[d.device_type]?.[d.tier] || `${d.device_type}-${d.tier}`
}

function chipLabel(d: Device): string {
  if (d.device_type === 'bracelet') return 'GD32E230C8T3'
  if (d.tier === 'auto') return 'ESP32-C3 + 电机驱动'
  if (d.tier === 'smart') return 'ESP32-C3 + TTS'
  return '无MCU (纯机械)'
}

function statusClass(s: string): string {
  if (s === 'online') return 'online'
  if (s === 'offline') return 'offline'
  return 'fault'
}

function statusLabel(s: string): string {
  return s === 'online' ? '在线' : s === 'offline' ? '离线' : '故障'
}

function modeLabel(m?: string): string {
  const map: Record<string, string> = { family: '家属', admin: '后台', community: '社区', medical: '医疗' }
  return map[m || ''] || m || '—'
}

function isOutdated(d: Device): boolean {
  return !!(d.firmware_version && d.firmware_version !== latestFw)
}

function signalStrength(d: Device): string {
  if (d.rssi) return `${d.rssi} dBm (${d.rssi! > -70 ? '良好' : '一般'})`
  return '—'
}

function formatDate(ts?: string): string {
  if (!ts) return '—'
  return new Date(ts).toLocaleDateString('zh-CN')
}

function formatLastSeen(ts?: string): string {
  if (!ts) return '—'
  const d = new Date(ts)
  const now = new Date()
  const diff = Math.floor((now.getTime() - d.getTime()) / 60000)
  if (diff < 1) return '刚刚'
  if (diff < 60) return `${diff}分钟前`
  if (diff < 1440) return `${Math.floor(diff / 60)}小时前`
  return d.toLocaleDateString('zh-CN')
}

function otaStatusText(status?: string): string {
  const map: Record<string, string> = {
    idle: '就绪', downloading: '下载中...', verifying: '验证中...', installing: '安装中...', success: '升级成功', failed: '升级失败',
  }
  return map[status || ''] || '—'
}

// Actions
async function handleSearch() {
  filters.value.page = 1
  await deviceStore.fetchList({ status: filters.value.status })
}

function handleReset() {
  filters.value = { type: '', status: '', mode: '', search: '', page: 1, pageSize: 20 }
  deviceStore.fetchList()
}

function handleRefresh() {
  Promise.all([deviceStore.fetchList(), deviceStore.fetchStats()])
}

function handleSizeChange(size: number) { filters.value.pageSize = size; deviceStore.fetchList() }
function handlePageChange(page: number) { filters.value.page = page; deviceStore.fetchList() }

function exportDevices() {
  ElMessage.info('导出功能开发中...')
}

function handleRegister() {
  ElMessage.info('设备注册功能开发中...')
}

// Side Panel
const panelOpen = ref(false)
const panelDevice = ref<Device | null>(null)

function handleRowClick(row: Device) {
  panelDevice.value = row
  panelOpen.value = true
}

function closePanel() {
  panelOpen.value = false
}

function goToMap() {
  ElMessage.info('地图功能开发中...')
}

// OTA (uses side panel actions)
const otaForm = ref({ firmwareUrl: '', hash: '', force: false, changelog: '' })

function handleOTA(row: Device) {
  closePanel()
  ElMessage.info(`准备对 ${row.device_id} 进行OTA升级`)
}

function handleBatchOta() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  ElMessage.info(`准备对 ${selectedIds.value.length} 台设备进行批量OTA`)
}

// Config
function handleConfig(row: Device) {
  closePanel()
  ElMessage.info(`准备配置 ${row.device_id}`)
}

function handleBatchConfig() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  ElMessage.info(`准备对 ${selectedIds.value.length} 台设备进行批量配置`)
}

// Reboot
async function handleReboot(row: Device) {
  try {
    await ElMessageBox.confirm(`确认重启设备 ${row.device_id}？`, '确认', { type: 'warning' })
    ElMessage.success('重启指令已发送')
  } catch { /* cancelled */ }
}

// Unbind
async function handleUnbind(row: Device) {
  try {
    await ElMessageBox.confirm(`确定要解绑设备 ${row.device_id} 吗？`, '确认', { type: 'warning' })
    deviceStore.devices = deviceStore.devices.filter(d => d.id !== row.id)
    ElMessage.success('已解绑')
    closePanel()
  } catch { /* cancelled */ }
}

async function handleBatchUnbind() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  try {
    await ElMessageBox.confirm(`确定要解绑选中的 ${selectedIds.value.length} 台设备吗？`, '确认', { type: 'warning' })
    deviceStore.devices = deviceStore.devices.filter(d => !selectedIds.value.includes(d.id))
    selectedIds.value = []
    ElMessage.success('已批量解绑')
  } catch { /* cancelled */ }
}

onMounted(() => {
  Promise.all([deviceStore.fetchList(), deviceStore.fetchStats()])
})
</script>

<style scoped>
.devices-page {
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
.header-actions {
  display: flex;
  gap: 8px;
}

/* KPI cards */
.kpi-card :deep(.el-card__body) {
  padding: 16px;
  text-align: center;
}
.kpi-num {
  font-size: 26px;
  font-weight: 800;
  line-height: 1;
}
.kpi-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
.kpi-online .kpi-num { color: var(--el-color-success); }
.kpi-offline .kpi-num { color: var(--el-color-info); }
.kpi-upgrade .kpi-num { color: var(--el-color-warning); }
.kpi-fault .kpi-num { color: var(--el-color-danger); }

/* Filter bar */
.filter-bar {
  background: white;
  border-radius: 12px;
  padding: 14px 18px;
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-light);
  flex-wrap: wrap;
  align-items: center;
}
.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}
.filter-select {
  width: 130px;
}
.filter-search {
  width: 240px;
}
.filter-spacer {
  flex: 1;
}

/* Bulk banner */
.bulk-banner {
  background: var(--el-color-primary-light-9);
  border: 1px solid var(--el-color-primary);
  border-radius: 10px;
  padding: 10px 16px;
  margin-bottom: 12px;
  display: none;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}
.bulk-banner.show {
  display: flex;
}
.bulk-count {
  font-weight: 700;
  color: var(--el-color-primary);
}
.bulk-actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

/* Table card */
.table-card :deep(.el-card__header) {
  padding: 0;
}
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
}
.table-title {
  font-size: 15px;
  font-weight: 700;
}
.table-actions {
  display: flex;
  gap: 8px;
}

/* Device cell */
.device-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.device-thumb {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}
.thumb-bracelet { background: #DBEAFE; }
.thumb-pillbox { background: #FCE7F3; }
.device-name {
  font-weight: 600;
  font-size: 13px;
}
.device-model {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

/* Status badges */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
}
.status-badge.online { background: #F0FDF4; color: var(--el-color-success); }
.status-badge.offline { background: var(--el-fill-color-light); color: var(--el-text-color-secondary); }
.status-badge.fault { background: #FEF2F2; color: var(--el-color-danger); }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.status-dot.online { background: var(--el-color-success); }
.status-dot.offline { background: var(--el-text-color-placeholder); }
.status-dot.fault { background: var(--el-color-danger); }

/* Version tag */
.version-tag {
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-weight: 500;
}
.version-tag.outdated {
  background: #FFFBEB;
  color: #D97706;
  border: 1px solid #FDE68A;
}

/* Action links */
.action-links {
  display: flex;
  gap: 12px;
}
.action-link {
  color: var(--el-color-primary);
  font-size: 12px;
  cursor: pointer;
  font-weight: 500;
  text-decoration: none;
}
.action-link:hover { text-decoration: underline; }
.action-link.danger { color: var(--el-color-danger); }

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
  right: -480px;
  bottom: 0;
  width: 480px;
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
.panel-close:hover { background: var(--el-border-color-light); }

.panel-body {
  padding: 20px 24px;
}
.panel-device-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}
.panel-device-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}
.icon-bracelet { background: #DBEAFE; }
.icon-pillbox { background: #FCE7F3; }
.panel-device-name {
  font-size: 18px;
  font-weight: 700;
}
.panel-device-id {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  font-family: monospace;
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

/* OTA progress */
.panel-progress {
  margin-top: 8px;
}
.progress-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin-bottom: 4px;
}
.progress-bar {
  height: 8px;
  background: var(--el-border-color-light);
  border-radius: 4px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s;
}
.progress-fill.success { background: var(--el-color-success); }
.progress-fill.running {
  background: var(--el-color-primary);
  animation: progressPulse 1.5s infinite;
}
@keyframes progressPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
.progress-meta {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  margin-top: 6px;
}

/* Panel action buttons */
.panel-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-light);
}
</style>
