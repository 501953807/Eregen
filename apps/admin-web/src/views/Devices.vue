<template>
  <div class="devices-page">
    <!-- Stats Row -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.bracelet_count.toLocaleString() }}</div>
            <div class="stat-label">手环总数</div>
          </div>
          <el-icon :size="40" style="color: #4A90D9;"><Watch /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.pillbox_count.toLocaleString() }}</div>
            <div class="stat-label">药盒总数</div>
          </div>
          <el-icon :size="40" style="color: #67C23A;"><PieChart /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">{{ stats.online_rate }}%</div>
            <div class="stat-label">在线率</div>
          </div>
          <el-icon :size="40" style="color: #E6A23C;"><Connection /></el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filters -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <el-form :inline="true">
        <el-form-item label="设备类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px;">
            <el-option label="手环入门版" value="bracelet-starter" />
            <el-option label="手环中端版" value="bracelet-plus" />
            <el-option label="手环高端版" value="bracelet-pro" />
            <el-option label="药盒基础版" value="pillbox-basic" />
            <el-option label="药盒智能版" value="pillbox-smart" />
            <el-option label="药盒自动版" value="pillbox-auto" />
          </el-select>
        </el-form-item>
        <el-form-item label="在线状态">
          <el-select v-model="filters.online" placeholder="全部" clearable style="width: 120px;">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
          </el-select>
        </el-form-item>
        <el-form-item label="固件版本">
          <el-input v-model="filters.firmware" placeholder="输入版本" clearable style="width: 140px;" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Device Table -->
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">设备列表</span>
          <el-button type="primary" size="small">批量OTA升级</el-button>
        </div>
      </template>
      <el-table v-loading="deviceStore.loading" :data="displayDevices" stripe style="width: 100%">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="device_id" label="设备ID" width="120" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.device_type, row.tier)" size="small">{{ deviceLabel(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关联老人" width="120">
          <template #default="{ row }">
            {{ row.owner_user_id || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="绑定家属" width="120">
          <template #default="{ row }">
            {{ row.owner_user_id || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="固件版本" width="100">
          <template #default="{ row }">
            {{ row.settings?.firmware_version || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small">{{ row.status === 'online' ? '在线' : '离线' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后上线" width="160">
          <template #default="{ row }">
            {{ row.last_seen || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleOTA(row)">OTA升级</el-button>
            <el-button link type="primary" size="small" @click="handleConfig(row)">远程配置</el-button>
            <el-button link type="danger" size="small">解绑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
        <el-pagination background layout="prev, pager, next" :total="deviceStore.total" :page-size="20" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Watch, PieChart, Connection } from '@element-plus/icons-vue'
import { useDeviceStore } from '@/stores/device'
import type { Device } from '@/types'

const deviceStore = useDeviceStore()

const filters = ref({
  type: '',
  online: '',
  firmware: '',
})

const displayDevices = computed(() => {
  let list = deviceStore.devices
  if (filters.value.online) {
    list = list.filter(d => d.status === filters.value.online)
  }
  return list
})

function deviceLabel(d: Device): string {
  const labels: Record<string, Record<string, string>> = {
    bracelet: { starter: '手环Starter', plus: '手环Plus', pro: '手环Pro' },
    pillbox: { basic: '药盒Basic', smart: '药盒Smart', auto: '药盒Auto' },
  }
  return labels[d.device_type]?.[d.tier] || `${d.device_type}-${d.tier}`
}

function typeTag(type: string, tier: string): 'primary' | 'success' | 'warning' {
  if (type === 'bracelet') return 'primary'
  return 'success'
}

async function handleSearch() {
  const params: Record<string, any> = {}
  if (filters.value.online) params.status = filters.value.online
  await deviceStore.fetchList(params)
}

function handleReset() {
  filters.value = { type: '', online: '', firmware: '' }
  deviceStore.fetchList()
}

function handleOTA(_row: Device) {
  ElMessage.info('OTA升级功能开发中')
}

function handleConfig(_row: Device) {
  ElMessage.info('远程配置功能开发中')
}

onMounted(() => {
  Promise.all([deviceStore.fetchList(), deviceStore.fetchStats()])
})
</script>

<style scoped>
.stat-card :deep(.el-card__body) { padding: 20px; display: flex; align-items: center; justify-content: space-between; }
.stat-content { flex: 1; }
.stat-value { font-size: 32px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
