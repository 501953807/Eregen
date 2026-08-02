<template>
  <el-table :data="patientList" v-loading="loading" stripe class="patient-table">
    <el-table-column prop="name" label="姓名" width="100">
      <template #default="{ row }">
        <div class="patient-cell">
          <div class="patient-avatar avatar-blue">{{ row.name[0] }}</div>
          <span>{{ row.name }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="admission_no" label="住院号" width="140"><template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template></el-table-column>
    <el-table-column prop="department" label="科室" width="120"><template #default="{ row }"><span class="dept-badge">{{ row.department }}</span></template></el-table-column>
    <el-table-column prop="bed_number" label="床号" width="80" />
    <el-table-column prop="last_verify" label="最后核验" width="180"><template #default="{ row }">{{ row.last_verify || '未核验' }}</template></el-table-column>
    <el-table-column prop="verify_gap_hours" label="距上次核验(h)" width="120">
      <template #default="{ row }">
        <span class="verify-tag" :class="row.verify_gap_hours > 12 ? 'tag-danger' : row.verify_gap_hours > 6 ? 'tag-warning' : ''">{{ row.verify_gap_hours }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="fence_status" label="围栏状态" width="100">
      <template #default="{ row }">
        <span class="status-badge" :class="row.fence_status === 'inside' ? 'badge-success' : 'badge-danger'">
          <span class="status-dot" :class="row.fence_status === 'inside' ? 'dot-success' : 'dot-danger'"></span>
          {{ row.fence_status === 'inside' ? '在院内' : '已越界' }}
        </span>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="160" fixed="right">
      <template #default="{ row }">
        <el-button link type="primary" size="small" @click="$emit('audit', row.id)">审计追踪</el-button>
        <el-button link type="primary" size="small" @click="$emit('detail', row)">详情</el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
defineProps<{ patientList: any[]; loading: boolean }>()
defineEmits<{ audit: [id: string]; detail: [row: any] }>()
</script>

<style scoped>
.patient-cell { display: flex; align-items: center; gap: 8px; }
.patient-avatar { width: 28px; height: 28px; border-radius: 50%; background: #DBEAFE; color: #2563EB; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; flex-shrink: 0; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
.dept-badge { background: #EFF6FF; color: #2563EB; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.verify-tag { font-size: 12px; }
.tag-danger { color: #DC2626; font-weight: 600; }
.tag-warning { color: #D97706; font-weight: 600; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
</style>
