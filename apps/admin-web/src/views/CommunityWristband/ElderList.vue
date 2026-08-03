<template>
  <div>
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="3">
        <el-button type="primary" @click="$emit('add')">新增老人</el-button>
      </el-col>
      <el-col :span="7">
        <el-input v-model="searchQuery" placeholder="搜索姓名 / 身份证号 / 手机号" clearable />
      </el-col>
    </el-row>

    <el-table :data="filteredRows" v-loading="loading" stripe class="elder-table">
      <el-table-column prop="name" label="姓名" width="90">
        <template #default="{ row }">
          <div class="patient-cell">
            <div class="patient-avatar" :class="row.gender === 1 ? 'avatar-blue' : 'avatar-pink'">{{ row.name[0] || '?' }}</div>
            <strong>{{ row.name }}</strong>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="id_card" label="身份证号" width="190">
        <template #default="{ row }"><span class="mono">{{ row.id_card }}</span></template>
      </el-table-column>
      <el-table-column prop="age" label="年龄" width="55" />
      <el-table-column prop="gender" label="性别" width="50">
        <template #default="{ row }">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}</template>
      </el-table-column>
      <el-table-column prop="emergency_contact" label="紧急联系人" width="120" />
      <el-table-column prop="welfare_tags" label="福利标签" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="tag in (row as any)._welfareTags || []" :key="tag.tag_code" size="small" :class="'welfare-tag-' + welfareTagClass(tag.tag_code)" effect="light" style="margin-right: 4px;">
            {{ tag.tag_name }}
          </el-tag>
          <span v-if="!((row as any)._welfareTags?.length)" style="color:var(--el-text-color-placeholder);">无</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="70">
        <template #default="{ row }">
          <span class="status-badge" :class="row.status === 'active' ? 'badge-success' : 'badge-gray'">
            <span class="status-dot" :class="row.status === 'active' ? 'dot-success' : 'dot-gray'"></span>
            {{ row.status === 'active' ? '正常' : '停用' }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="$emit('detail', row)">详情</el-button>
          <el-button size="small" link @click="$emit('edit', row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { CommunityElder } from '@/api/community'

const props = defineProps<{ rows: CommunityElder[]; loading: boolean }>()
const emit = defineEmits<{ 'add': []; 'detail': [row: CommunityElder]; 'edit': [row: CommunityElder] }>()

const searchQuery = ref('')
const filteredRows = computed(() => {
  if (!searchQuery.value) return props.rows
  const q = searchQuery.value.toLowerCase()
  return props.rows.filter(e =>
    e.name?.toLowerCase().includes(q) ||
    e.id_card?.toLowerCase().includes(q) ||
    e.emergency_contact?.toLowerCase().includes(q)
  )
})

function welfareTagClass(code: string): string {
  const map: Record<string, string> = {
    orphan: 'orphan', poverty_level_1: 'poverty', poverty_level_2: 'poverty',
    disability_1: 'disability', disability_2: 'disability', disability_3: 'disability',
    special_disease: 'special', bus_discount: 'bus', medical_assistance: 'medical',
  }
  return map[code] || ''
}
</script>

<style scoped>
.patient-cell { display: flex; align-items: center; gap: 8px; }
.patient-avatar { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; flex-shrink: 0; }
.avatar-blue { background: #DBEAFE; color: #165DFF; }
.avatar-pink { background: #FCE7F3; color: #D48EC0; }
.welfare-tag-orphan { background: #FEF2F2; color: #DC2626; }
.welfare-tag-poverty { background: #FFFBEB; color: #D97706; }
.welfare-tag-disability { background: #EFF6FF; color: #165DFF; }
.welfare-tag-special { background: #FEF2F2; color: #DC2626; }
.welfare-tag-bus { background: #F0FDF4; color: #16A34A; }
.welfare-tag-medical { background: #EDE9FE; color: #9B8ED8; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-gray { background: #6B7280; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
</style>
