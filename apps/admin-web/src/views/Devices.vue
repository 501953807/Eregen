<template>
  <div class="devices-page">
    <!-- Stats Row -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">1,247</div>
            <div class="stat-label">手环总数</div>
          </div>
          <el-icon :size="40" style="color: #4A90D9;"><Watch /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">230</div>
            <div class="stat-label">药盒总数</div>
          </div>
          <el-icon :size="40" style="color: #67C23A;"><PieChart /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">94.2%</div>
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
          <el-button type="primary">查询</el-button>
          <el-button>重置</el-button>
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
      <el-table :data="devices" stripe style="width: 100%">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" label="设备ID" width="120" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.typeTag" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="elderly" label="关联老人" width="120" />
        <el-table-column prop="family" label="绑定家属" width="120" />
        <el-table-column prop="firmware" label="固件版本" width="100" />
        <el-table-column prop="online" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.online ? 'success' : 'info'" size="small">{{ row.online ? '在线' : '离线' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastSeen" label="最后上线" width="160" />
        <el-table-column label="操作" fixed="right" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" size="small">OTA升级</el-button>
            <el-button link type="primary" size="small">远程配置</el-button>
            <el-button link type="danger" size="small">解绑</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
        <el-pagination background layout="prev, pager, next" :total="1247" :page-size="20" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Watch, PieChart, Connection } from '@element-plus/icons-vue'

const filters = ref({
  type: '',
  online: '',
  firmware: '',
})

interface Device {
  id: string
  type: string
  typeTag: 'primary' | 'success' | 'warning'
  elderly: string
  family: string
  firmware: string
  online: boolean
  lastSeen: string
}

const devices: Device[] = [
  { id: 'BR-0042', type: '手环Pro', typeTag: 'primary', elderly: '张建国', family: '张伟', firmware: 'v2.1.0', online: true, lastSeen: '2分钟前' },
  { id: 'BR-0017', type: '手环Plus', typeTag: 'success', elderly: '李秀英', family: '李芳', firmware: 'v2.0.8', online: true, lastSeen: '5分钟前' },
  { id: 'BR-0089', type: '手环Starter', typeTag: 'warning', elderly: '王德明', family: '王磊', firmware: 'v2.1.0', online: false, lastSeen: '3小时前' },
  { id: 'PX-0012', type: '药盒Auto', typeTag: 'primary', elderly: '赵淑华', family: '赵敏', firmware: 'v1.3.2', online: true, lastSeen: '1分钟前' },
  { id: 'PX-0008', type: '药盒Smart', typeTag: 'success', elderly: '陈志强', family: '陈刚', firmware: 'v1.3.2', online: false, lastSeen: '1天前' },
]
</script>

<style scoped>
.stat-card :deep(.el-card__body) { padding: 20px; display: flex; align-items: center; justify-content: space-between; }
.stat-value { font-size: 32px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
