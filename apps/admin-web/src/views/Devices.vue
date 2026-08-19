<template>
  <div class="devices-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h2 class="page-title">设备管理</h2>
        <p class="page-subtitle">管理所有硬件设备 · 手环与药盒状态总览</p>
      </div>
      <div class="header-actions">
        <HopeBtn variant="filled" @click="handleRegister" size="md">
          + 注册设备
        </HopeBtn>
        <HopeBtn variant="plain" @click="handleRefresh" size="md" iconOnly>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="deviceStore.total"
        label="设备总数"
        icon-color="primary"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.online_devices"
        label="在线设备"
        icon-color="success"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.offline_devices"
        label="离线设备"
        icon-color="info"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.outdated_firmware"
        label="待升级固件"
        icon-color="warning"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 15 21 19 17 23"/><polyline points="7 10 3 10 3 14"/><path d="M21 3l-9 9-9-9"/></svg>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.fault_count"
        label="故障设备"
        icon-color="error"
      >
        <template #icon>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        </template>
      </HopeStatCard>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <span class="filter-label">筛选：</span>
      <el-select v-model="filters.type" placeholder="全部类型" clearable class="filter-select">
        <el-option label="手环" value="bracelet" />
        <el-option label="药盒" value="pillbox" />
        <el-option label="医用腕带" value="medical_wristband" />
        <el-option label="社区腕带" value="community_wristband" />
      </el-select>
      <el-select v-model="filters.status" placeholder="全部状态" clearable class="filter-select">
        <el-option label="在线" value="online" />
        <el-option label="离线" value="offline" />
        <el-option label="故障" value="fault" />
      </el-select>
      <el-select v-model="filters.mode" placeholder="全部模式" clearable class="filter-select">
        <el-option label="采集" value="collection" />
        <el-option label="守护" value="guard" />
      </el-select>
      <span class="filter-spacer"></span>
      <el-input v-model="filters.search" placeholder="搜索设备ID、名称、老人姓名..." clearable class="filter-search" />
      <HopeBtn variant="plain" size="sm" @click="handleReset">重置</HopeBtn>
      <HopeBtn variant="filled" size="sm" @click="handleSearch">搜索</HopeBtn>
    </div>

    <!-- Bulk Selection Banner -->
    <div class="bulk-banner" :class="{ show: selectedIds.length > 0 }">
      <el-checkbox :model-value="allSelected" :model-enabled="allSelected" @change="handleToggleSelectAll" />
      <span><strong class="bulk-count">{{ selectedIds.length }}</strong> 项已选中</span>
      <div class="bulk-actions">
        <HopeBtn variant="filled" size="sm" @click="handleBatchOta">批量OTA</HopeBtn>
        <HopeBtn variant="plain" size="sm" @click="handleBatchConfig">批量配置</HopeBtn>
        <HopeBtn variant="error" size="sm" @click="handleBatchUnbind">批量注销</HopeBtn>
        <HopeBtn variant="ghost" size="sm" @click="clearSelection">取消</HopeBtn>
      </div>
    </div>

    <!-- Device Table -->
    <HopeCard subtitle="所有设备列表 · 点击行查看详情">
      <template #header>
        <div class="table-toolbar">
          <span class="table-title">设备列表</span>
          <div class="table-actions">
            <HopeBtn variant="plain" size="sm" @click="exportDevices">导出CSV</HopeBtn>
            <HopeBtn variant="plain" size="sm" @click="handleRefresh">刷新</HopeBtn>
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
        <el-table-column type="selection" width="40" />
        <el-table-column label="设备ID" width="130">
          <template #default="{ row }">
            <span class="mono">{{ row.device_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <HopeBadge :color="row.type === 'bracelet' ? 'primary' : row.type === 'pillbox' ? 'warning' : 'info'">
              {{ row.type === 'bracelet' ? '手环' : row.type === 'pillbox' ? '药盒' : row.type || '—' }}
            </HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="档位" width="90">
          <template #default="{ row }">
            <HopeBadge color="success">{{ row.tier || '—' }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <HopeBadge :color="badgeColor(row.status)">{{ statusLabel(row.status) }}</HopeBadge>
          </template>
        </el-table-column>
        <el-table-column label="绑定老人" width="100">
          <template #default="{ row }">
            {{ row.owner_name || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="最后上线" width="110">
          <template #default="{ row }">
            {{ formatLastSeen(row.last_seen) }}
          </template>
        </el-table-column>
        <el-table-column label="固件版本" width="100">
          <template #default="{ row }">
            <span class="version-tag" :class="{ outdated: isOutdated(row) }">
              {{ row.firmware_version || '—' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="action-links">
              <a class="action-link" @click.stop="handleOTA(row)">OTA升级</a>
              <a class="action-link" @click.stop="handleConfig(row)">配置</a>
              <a class="action-link" @click.stop="handleReboot(row)">重启</a>
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
          :page-size="pState.pageSize"
          :current-page="pState.page"
          :page-sizes="[10, 20, 50, 100]"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </HopeCard>

    <!-- Side Panel -->
    <div class="side-panel-overlay" :class="{ show: panelOpen }" @click="closePanel" />
    <div class="side-panel" :class="{ open: panelOpen }">
      <div class="panel-header">
        <span class="panel-title">设备详情</span>
        <HopeBtn variant="ghost" size="sm" iconOnly @click="closePanel">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </HopeBtn>
      </div>
      <div class="panel-body" v-if="panelDevice">
        <div class="panel-device-header">
          <div class="panel-device-icon" :class="panelDevice.type === 'bracelet' ? 'icon-bracelet' : 'icon-pillbox'">
            {{ panelDevice.type === 'bracelet' ? '📱' : '💊' }}
          </div>
          <div>
            <div class="panel-device-name">{{ deviceLabel(panelDevice) }}</div>
            <div class="panel-device-id">{{ panelDevice.device_id }}</div>
          </div>
        </div>

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

        <div class="panel-section">
          <div class="panel-section-title">实时状态</div>
          <div class="panel-row"><span class="panel-row-label">连接状态</span><span class="panel-row-value">
            <HopeBadge :color="badgeColor(panelDevice.status)">{{ statusLabel(panelDevice.status) }}</HopeBadge>
          </span></div>
          <div class="panel-row"><span class="panel-row-label">信号强度</span><span class="panel-row-value">{{ signalStrength(panelDevice) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">电量</span><span class="panel-row-value">{{ panelDevice.battery_pct ?? '—' }}%</span></div>
          <div class="panel-row"><span class="panel-row-label">最后心跳</span><span class="panel-row-value">{{ formatLastSeen(panelDevice.last_seen) }}</span></div>
          <div class="panel-row"><span class="panel-row-label">最近定位</span><span class="panel-row-value panel-link" @click="goToMap">查看地图 →</span></div>
        </div>

        <div class="panel-section" v-if="panelDevice.type === 'bracelet'">
          <div class="panel-section-title">健康数据摘要</div>
          <div class="panel-row"><span class="panel-row-label">心率</span><span class="panel-row-value">{{ panelDevice.hr ?? '—' }} bpm</span></div>
          <div class="panel-row"><span class="panel-row-label">血氧</span><span class="panel-row-value">{{ panelDevice.spo2 ?? '—' }}%</span></div>
          <div class="panel-row"><span class="panel-row-label">步数</span><span class="panel-row-value">{{ panelDevice.steps ?? '—' }}</span></div>
        </div>

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

        <div class="panel-actions">
          <HopeBtn variant="filled" @click="handleOTA(panelDevice)" style="flex:1;">OTA升级</HopeBtn>
          <HopeBtn variant="plain" @click="handleConfig(panelDevice)" style="flex:1;">远程配置</HopeBtn>
          <HopeBtn variant="plain" @click="handleReboot(panelDevice)" style="flex:1;">远程重启</HopeBtn>
          <HopeBtn variant="error" @click="handleUnbind(panelDevice)" style="flex:1;">注销设备</HopeBtn>
        </div>
      </div>
    </div>

    <!-- Config Dialog -->
    <el-dialog v-model="showConfigDialog" :title="`远程配置 — ${configDevice?.device_id || ''}`" width="480px" destroy-on-close>
      <el-form :model="configForm" label-width="130px">
        <el-form-item label="心跳间隔（秒）">
          <el-input-number v-model="configForm.interval" :min="5" :max="300" style="width:100%;" />
        </el-form-item>
        <el-form-item label="音量（%）">
          <el-slider v-model="configForm.volume" :min="0" :max="100" show-input style="width:100%;" />
        </el-form-item>
        <el-form-item label="GPS 定位">
          <el-switch v-model="configForm.gps_enabled" />
        </el-form-item>
        <el-form-item label="SOS 按钮">
          <el-switch v-model="configForm.sos_enabled" />
        </el-form-item>
        <el-form-item label="跌倒检测">
          <el-switch v-model="configForm.fall_detect" />
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" @click="showConfigDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="confirmConfig" :loading="false">确认下发</HopeBtn>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useDeviceStore } from '@/stores/device'
import { devicesApi } from '@/api/devices'
import type { Device } from '@/types'
import { usePagination, useFilters, useSelection } from '@/composables'
import { HopeCard, HopeStatCard, HopeBadge, HopeBtn } from '@/components/hope'

const deviceStore = useDeviceStore()

const pagination = usePagination(20)
const { state: pState, setPageSize: setPageSizeFn, setTotal } = pagination
const { filters, setFilter, reset: resetFilters } = useFilters({
  type: '',
  status: '',
  mode: '',
  search: '',
})
const { selectedIds, toggleSelectAll: toggleSelectAllFn, toggleRow, clearSelection, isSelected, allSelected } = useSelection<string>()

const filteredDevices = computed(() => {
  let list = deviceStore.devices
  if (filters.value.status) list = list.filter(d => d.status === filters.value.status)
  if (filters.value.type) list = list.filter(d => d.type === filters.value.type)
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

const latestFw = 'v2.4.1'
const stats = computed(() => ({
  online_devices: deviceStore.devices.filter(d => d.status === 'online').length,
  offline_devices: deviceStore.devices.filter(d => d.status !== 'online').length,
  outdated_firmware: deviceStore.devices.filter(d => d.firmware_version && d.firmware_version !== latestFw).length,
  fault_count: deviceStore.devices.filter(d => d.status === 'fault').length,
}))

function handleSelectionChange(rows: Device[]) {
  rows.forEach(r => toggleRow(r.id, true))
}
function handleToggleSelectAll(val: boolean) {
  toggleSelectAllFn(val, filteredDevices.value.map(d => d.id))
}
function clearSelectionBtn() {
  clearSelection()
}

function deviceLabel(d: Device): string {
  const labels: Record<string, Record<string, string>> = {
    bracelet: { starter: '手环 Starter', plus: '手环 Plus', pro: '手环 Pro' },
    pillbox: { basic: '药盒 Basic', smart: '药盒 Smart', auto: '药盒 Auto' },
  }
  return (d.type ? labels[d.type]?.[d.tier || ''] : '') || `${d.type || ''}-${d.tier || ''}`
}

function chipLabel(d: Device): string {
  if (d.type === 'bracelet') return 'GD32E230C8T3'
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

function badgeColor(s: string): 'success' | 'error' | 'info' {
  if (s === 'online') return 'success'
  if (s === 'offline') return 'info'
  return 'error'
}

function modeLabel(m?: string): string {
  const map: Record<string, string> = { family: '家属', admin: '后台', community: '社区', medical: '医疗', collection: '采集', guard: '守护' }
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

async function handleSearch() {
  pState.value.page = 1
  await deviceStore.fetchList({ status: filters.value.status })
}

function handleReset() {
  resetFilters()
  deviceStore.fetchList()
}

function handleRefresh() {
  Promise.all([deviceStore.fetchList(), deviceStore.fetchStats()])
}

function handleSizeChange(size: number) {
  setPageSizeFn(size)
  deviceStore.fetchList()
}

function handlePageChange(page: number) {
  pState.value.page = page
  deviceStore.fetchList()
}

function exportDevices() { ElMessage.info('导出功能开发中...') }
function handleRegister() { ElMessage.info('设备注册功能开发中...') }

// Side Panel
const panelOpen = ref(false)
const panelDevice = ref<Device | null>(null)

function handleRowClick(row: Device) {
  panelDevice.value = row
  panelOpen.value = true
}

function closePanel() { panelOpen.value = false }
function goToMap() { ElMessage.info('地图功能开发中...') }

// OTA
function handleOTA(row: Device) {
  closePanel()
  ElMessage.info(`准备对 ${row.device_id} 进行OTA升级`)
}
function handleBatchOta() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  ElMessage.info(`准备对 ${selectedIds.value.length} 台设备进行批量OTA`)
}

// Config dialog
const showConfigDialog = ref(false)
const configDevice = ref<Device | null>(null)
const configForm = ref({ interval: 30, volume: 80, gps_enabled: true, sos_enabled: true, fall_detect: true })

function handleConfig(row: Device) {
  closePanel()
  configDevice.value = row
  configForm.value = {
    interval: row.settings?.interval ?? 30,
    volume: row.settings?.volume ?? 80,
    gps_enabled: row.settings?.gps_enabled ?? true,
    sos_enabled: row.settings?.sos_enabled ?? true,
    fall_detect: row.settings?.fall_detect ?? true,
  }
  showConfigDialog.value = true
}

async function confirmConfig() {
  if (!configDevice.value) return
  try {
    await devicesApi.updateConfig(configDevice.value.device_id, {
      interval: configForm.value.interval,
      volume: configForm.value.volume,
      gps_enabled: configForm.value.gps_enabled,
      sos_enabled: configForm.value.sos_enabled,
      fall_detect: configForm.value.fall_detect,
    })
    ElMessage.success('配置已下发')
    showConfigDialog.value = false
    await deviceStore.fetchList({
      page: pState.value.page,
      page_size: pState.value.pageSize,
      status: filters.value.status,
      type: filters.value.type,
    })
  } catch {
    ElMessage.error('配置下发失败')
  }
}

function handleBatchConfig() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  ElMessage.info(`准备对 ${selectedIds.value.length} 台设备进行批量配置`)
}

async function handleReboot(row: Device) {
  try {
    await ElMessageBox.confirm(`确认重启设备 ${row.device_id}？`, '确认', { type: 'warning' })
    ElMessage.success('重启指令已发送')
  } catch { /* cancelled */ }
}

async function handleUnbind(row: Device) {
  try {
    await ElMessageBox.confirm(`确定要解绑设备 ${row.device_id} 吗？`, '确认', { type: 'warning' })
    await devicesApi.unbind(row.id)
    deviceStore.devices = deviceStore.devices.filter(d => d.id !== row.id)
    ElMessage.success('已解绑')
    closePanel()
  } catch (e: any) {
    if (e?.code !== 'cancel') ElMessage.error('解绑失败')
  }
}

async function handleBatchUnbind() {
  if (!selectedIds.value.length) { ElMessage.warning('请先选择设备'); return }
  try {
    await ElMessageBox.confirm(`确定要解绑选中的 ${selectedIds.value.length} 台设备吗？`, '确认', { type: 'warning' })
    await Promise.all(selectedIds.value.map(id => devicesApi.unbind(id)))
    deviceStore.devices = deviceStore.devices.filter(d => !selectedIds.value.includes(d.id))
    selectedIds.value = []
    ElMessage.success('已批量解绑')
  } catch (e: any) {
    if (e?.code !== 'cancel') ElMessage.error('批量解绑失败')
  }
}

onMounted(() => {
  Promise.all([deviceStore.fetchList(), deviceStore.fetchStats()])
})
</script>

<style scoped>
.devices-page {
  padding: 0;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}
.page-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0 0 4px;
  letter-spacing: -0.02em;
}
.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
}
.header-actions {
  display: flex;
  gap: 8px;
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}
.kpi-grid .hope-stat-card {
  cursor: default;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.filter-select { width: 130px; }
.filter-search { width: 240px; }
.filter-spacer { flex: 1; }

/* Bulk Banner */
.bulk-banner {
  background: var(--hope-primary-lighter);
  border: 1px solid var(--hope-primary-light);
  border-radius: var(--hope-radius-lg);
  padding: 10px 16px;
  margin-bottom: 16px;
  display: none;
  align-items: center;
  gap: 12px;
  font-size: 13px;
}
.bulk-banner.show {
  display: flex;
  animation: eregen-list-enter 0.2s ease-out;
}
.bulk-count {
  font-weight: 700;
  color: var(--hope-primary);
}
.bulk-actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

/* Table Toolbar */
.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.table-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--hope-text);
}
.table-actions {
  display: flex;
  gap: 6px;
}

/* Device Cell */
.device-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.device-thumb {
  width: 36px;
  height: 36px;
  border-radius: var(--hope-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}
.thumb-bracelet { background: var(--hope-primary-lighter); }
.thumb-pillbox { background: var(--hope-surface-light); }
.device-name {
  font-weight: 600;
  font-size: 13px;
  color: var(--hope-text);
}
.device-model {
  font-size: 11px;
  color: var(--hope-text-muted);
}

/* Version Tag */
.mono {
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 12px;
  color: var(--hope-text-secondary);
}
.version-tag {
  font-family: 'SF Mono', Consolas, monospace;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--hope-radius-sm);
  background: var(--hope-surface-light);
  font-weight: 500;
  color: var(--hope-text-secondary);
}
.version-tag.outdated {
  background: var(--hope-warning-light);
  color: #926C0E;
  border: 1px solid rgba(var(--hope-warning-rgb), 0.3);
}

/* Action Links */
.action-links {
  display: flex;
  gap: 12px;
}
.action-link {
  color: var(--hope-primary);
  font-size: 12px;
  cursor: pointer;
  font-weight: 600;
  text-decoration: none;
  transition: color 0.15s;
}
.action-link:hover { color: var(--hope-primary-hover); text-decoration: underline; }
.action-link.danger { color: var(--hope-danger); }
.action-link.danger:hover { color: #9B3A33; }

/* Pagination */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 14px 22px;
}

/* ========== Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(26,26,46,0.3);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show { display: block; }
.side-panel {
  position: fixed;
  top: 0;
  right: -480px;
  bottom: 0;
  width: 480px;
  background: var(--hope-surface);
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: var(--hope-shadow-lg);
  border-left: 1px solid var(--hope-border);
}
.side-panel.open { right: 0; }
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
.panel-body { padding: 20px 24px; }

.panel-device-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}
.panel-device-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--hope-radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
}
.icon-bracelet { background: var(--hope-primary-lighter); }
.icon-pillbox { background: var(--hope-surface-light); }
.panel-device-name { font-size: 18px; font-weight: 700; color: var(--hope-text); }
.panel-device-id { font-size: 12px; color: var(--hope-text-muted); font-family: monospace; }

.panel-section { margin-bottom: 20px; }
.panel-section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
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
.panel-row-label { color: var(--hope-text-muted); }
.panel-row-value { font-weight: 600; color: var(--hope-text); }
.panel-link { color: var(--hope-primary); cursor: pointer; }
.panel-link:hover { text-decoration: underline; }

/* OTA Progress */
.panel-progress { margin-top: 8px; }
.progress-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin-bottom: 4px;
  color: var(--hope-text-muted);
}
.progress-bar {
  height: 8px;
  background: var(--hope-surface-light);
  border-radius: 4px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s;
}
.progress-fill.success { background: var(--hope-primary-gradient); }
.progress-fill.running {
  background: var(--hope-primary-gradient);
  animation: progressPulse 1.5s infinite;
}
@keyframes progressPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
.progress-meta {
  font-size: 11px;
  color: var(--hope-text-muted);
  margin-top: 6px;
}

.panel-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--hope-border);
}

/* Responsive */
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .devices-page :deep(.el-table) { font-size: 12px; }
  .devices-page :deep(.el-table th),
  .devices-page :deep(.el-table td) { padding: 6px 4px; }
  .side-panel { width: 100%; right: -100%; }
}
</style>
