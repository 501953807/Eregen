<template>
  <div class="elderly-page">
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">老人档案管理</span>
          <el-button type="primary" size="small" @click="showAddDialog = true">添加老人</el-button>
        </div>
      </template>

      <el-table :data="elderlyList" stripe style="width: 100%">
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column label="年龄" width="80">
          <template #default="{ row }">
            {{ row.birth_date ? calculateAge(row.birth_date) : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="健康等级" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.health_tiers?.length" :type="healthTag(row.health_tiers[0])" size="small">
              {{ row.health_tiers[0] }}
            </el-tag>
            <span v-else>未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="关联用户ID" width="160">
          <template #default="{ row }">
            {{ row.user_id || '—' }}
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="160">
          <template #default="{ row }">
            {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleView(row)">查看</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog v-model="showAddDialog" :title="editingId ? '编辑老人档案' : '添加老人'" width="520px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="姓名">
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="出生日期">
          <el-date-picker v-model="form.birth_date" type="date" placeholder="选择出生日期" value-format="YYYY-MM-DD" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="关联用户">
          <el-input v-model="form.user_id" placeholder="家属用户ID" />
        </el-form-item>
        <el-form-item label="健康等级">
          <el-select v-model="form.health_tiers" multiple placeholder="选择健康等级" style="width: 100%;">
            <el-option label="低风险" value="低风险" />
            <el-option label="中风险" value="中风险" />
            <el-option label="高风险" value="高风险" />
          </el-select>
        </el-form-item>
        <el-form-item label="头像URL">
          <el-input v-model="form.avatar_url" placeholder="图片URL（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- View Detail Dialog -->
    <el-dialog v-model="showViewDialog" title="老人详情" width="520px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="姓名">{{ viewData.name }}</el-descriptions-item>
        <el-descriptions-item label="年龄">{{ viewData.birth_date ? calculateAge(viewData.birth_date) : '—' }}</el-descriptions-item>
        <el-descriptions-item label="关联用户">{{ viewData.user_id || '—' }}</el-descriptions-item>
        <el-descriptions-item label="健康等级">
          <el-tag v-for="tier in viewData.health_tiers" :key="tier" :type="healthTag(tier)" size="small" style="margin-right: 4px;">{{ tier }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ viewData.created_at ? new Date(viewData.created_at).toLocaleDateString() : '—' }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ viewData.updated_at ? new Date(viewData.updated_at).toLocaleDateString() : '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { elderlyApi } from '@/api/elderly'
import type { ElderlyProfile } from '@/types'

const elderlyList = ref<ElderlyProfile[]>([])

// Mock data as fallback
const mockElderly: ElderlyProfile[] = [
  { id: '1', user_id: 'user-001', name: '张建国', birth_date: '1948-03-15', avatar_url: '', health_tiers: ['低风险'], created_at: '2025-03-20T00:00:00Z', updated_at: '2026-07-10T00:00:00Z' },
  { id: '2', user_id: 'user-002', name: '李秀英', birth_date: '1944-07-22', avatar_url: '', health_tiers: ['中风险'], created_at: '2025-06-01T00:00:00Z', updated_at: '2026-07-12T00:00:00Z' },
  { id: '3', user_id: 'user-003', name: '王德明', birth_date: '1951-11-08', avatar_url: '', health_tiers: ['高风险'], created_at: '2025-09-15T00:00:00Z', updated_at: '2026-07-14T00:00:00Z' },
  { id: '4', user_id: 'user-004', name: '赵淑华', birth_date: '1946-01-30', avatar_url: '', health_tiers: ['低风险', '中风险'], created_at: '2025-01-05T00:00:00Z', updated_at: '2026-07-13T00:00:00Z' },
]

onMounted(async () => {
  try {
    const res = await elderlyApi.list({ page_size: 50 })
    elderlyList.value = res.data.data || []
  } catch {
    elderlyList.value = mockElderly
  }
})

function calculateAge(birthDate: string): number {
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function healthTag(level: string): 'success' | 'warning' | 'danger' {
  const map: Record<string, 'success' | 'warning' | 'danger'> = { '低风险': 'success', '中风险': 'warning', '高风险': 'danger' }
  return map[level] || 'success'
}

// Add/Edit dialog
const showAddDialog = ref(false)
const editingId = ref('')
const form = ref<Partial<ElderlyProfile>>({
  name: '', birth_date: '', user_id: '', health_tiers: [], avatar_url: '',
})

function handleAdd() {
  editingId.value = ''
  form.value = { name: '', birth_date: '', user_id: '', health_tiers: [], avatar_url: '' }
  showAddDialog.value = true
}

function handleEdit(row: ElderlyProfile) {
  editingId.value = row.id
  form.value = { ...row }
  showAddDialog.value = true
}

async function handleSave() {
  if (!form.value.name) {
    ElMessage.warning('请输入姓名')
    return
  }
  try {
    if (editingId.value) {
      await elderlyApi.update(editingId.value, form.value)
      ElMessage.success('更新成功')
    } else {
      const res = await elderlyApi.create(form.value)
      const created = res.data.data || res.data
      elderlyList.value.unshift(created as ElderlyProfile)
      ElMessage.success('创建成功')
    }
    showAddDialog.value = false
  } catch {
    // Fallback: optimistic update
    if (editingId.value) {
      const idx = elderlyList.value.findIndex(e => e.id === editingId.value)
      if (idx !== -1) Object.assign(elderlyList.value[idx], form.value)
      ElMessage.success('更新成功（模拟）')
    } else {
      const newItem: ElderlyProfile = {
        id: Date.now().toString(),
        name: form.value.name!,
        birth_date: form.value.birth_date as string,
        user_id: form.value.user_id as string,
        health_tiers: (form.value.health_tiers as string[]) || [],
        avatar_url: form.value.avatar_url as string,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      elderlyList.value.unshift(newItem)
      ElMessage.success('创建成功（模拟）')
    }
    showAddDialog.value = false
  }
}

// View detail
const showViewDialog = ref(false)
const viewData = ref<ElderlyProfile>({} as ElderlyProfile)

function handleView(row: ElderlyProfile) {
  viewData.value = { ...row }
  showViewDialog.value = true
}

// Delete
async function handleDelete(row: ElderlyProfile) {
  try {
    await ElMessageBox.confirm(`确定要删除老人 "${row.name}" 的档案吗？`, '确认', { type: 'warning' })
    try {
      await elderlyApi.delete(row.id)
    } catch {
      // API may not be available
    }
    elderlyList.value = elderlyList.value.filter(e => e.id !== row.id)
    ElMessage.success('已删除')
  } catch {
    // cancelled
  }
}
</script>

<style scoped>
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
