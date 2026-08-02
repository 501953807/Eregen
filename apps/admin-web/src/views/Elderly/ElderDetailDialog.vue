<template>
  <el-dialog v-model="visible" :title="'老人档案详情 — ' + (row?.name || '')" width="640px" destroy-on-close>
    <div v-if="row">
      <h4 style="margin-bottom:12px;color:#303133;">基本信息</h4>
      <div class="detail-grid">
        <div class="detail-item"><span class="label">姓名：</span><span class="value">{{ row.name }}</span></div>
        <div class="detail-item"><span class="label">性别：</span><span class="value">{{ row.gender || '—' }}</span></div>
        <div class="detail-item"><span class="label">年龄：</span><span class="value">{{ row.birth_date ? calculateAge(row.birth_date) + ' 岁' : '—' }}</span></div>
        <div class="detail-item"><span class="label">身份证号：</span><span class="value mono">{{ row.id_card || '—' }}</span></div>
        <div class="detail-item"><span class="label">地址：</span><span class="value">{{ row.address || '—' }}</span></div>
        <div class="detail-item"><span class="label">紧急联系人：</span><span class="value">{{ row.emergency_contact || '—' }}</span></div>
        <div class="detail-item"><span class="label">腕带设备：</span><span class="value">{{ row.wearable_id || '—' }} <span v-if="row.wearable_online" class="status-badge badge-success"><span class="status-dot dot-success"></span>在线</span></span></div>
        <div class="detail-item"><span class="label">状态：</span><span class="value"><span class="status-badge" :class="row.status === '正常' ? 'badge-success' : 'badge-gray'"><span class="status-dot" :class="row.status === '正常' ? 'dot-success' : 'dot-gray'"></span>{{ row.status }}</span></span></div>
      </div>
      <h4 style="margin:20px 0 12px;color:#303133;">福利标签</h4>
      <el-table :data="(row.welfare_tags || []).map(t => ({ ...t, status: '有效' }))" stripe size="small">
        <el-table-column label="标签" width="100">
          <template #default="{ row: t }"><el-tag :type="welfareTagType(t.code)" size="small">{{ t.name }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="issuer" label="发放机构" width="100" />
        <el-table-column label="生效日期" width="110">
          <template #default="{ row: t }">{{ t.start_date }}</template>
        </el-table-column>
        <el-table-column label="到期日期" width="110">
          <template #default="{ row: t }">{{ t.end_date }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row: t }"><span class="status-badge badge-success"><span class="status-dot dot-success"></span>{{ t.status }}</span></template>
        </el-table-column>
      </el-table>
    </div>
    <template #footer>
      <el-button @click="visible = false">关闭</el-button>
      <el-button type="primary" @click="$emit('edit', row)">编辑档案</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface WelfareTag { code: string; name: string; issuer?: string; start_date?: string; end_date?: string }
interface ElderlyRow {
  id: string; name: string; id_card?: string; birth_date?: string; gender?: string
  emergency_contact?: string; welfare_tags: WelfareTag[]; status: string
  address?: string; wearable_id?: string; wearable_online?: boolean
}

const props = defineProps<{ modelValue: boolean; row: ElderlyRow | null }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean]; edit: [row: ElderlyRow] }>()

const visible = ref(false)
watch(() => props.modelValue, v => { visible.value = v })
watch(visible, v => { if (!v) emit('update:modelValue', false) })

function calculateAge(birthDate?: string): number {
  if (!birthDate) return 0
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function welfareTagType(code: string): string {
  const map: Record<string, string> = { orphan: 'danger', poverty_1: 'warning', poverty_2: 'warning', disability_1: 'primary', disability_2: 'primary', disability_3: 'primary', special_disease: 'info', bus_discount: 'success', medical_assist: 'primary' }
  return map[code] || 'info'
}
</script>

<style scoped>
.detail-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.detail-item { display: flex; font-size: 13px; }
.detail-item .label { width: 80px; color: var(--el-text-color-secondary); flex-shrink: 0; }
.detail-item .value { color: var(--el-text-color-primary); font-weight: 500; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-gray { background: #6B7280; }
@media (max-width: 1200px) { .detail-grid { grid-template-columns: 1fr; } }
</style>
