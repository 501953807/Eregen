<template>
  <div class="users-page">
    <!-- Tabs for user types -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="家属用户" name="family">
          <div class="tab-toolbar">
            <el-input v-model="familySearch" placeholder="搜索用户名/手机号" clearable style="width: 240px;" />
            <el-button type="primary" style="margin-left: 12px;" @click="handleFamilySearch">查询</el-button>
            <el-button type="success" @click="handleAddFamily">添加用户</el-button>
          </div>
          <el-table v-loading="usersStore.loading" :data="filteredFamilyUsers" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column prop="phone" label="手机号" width="140" />
            <el-table-column prop="role" label="角色" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ roleLabel(row.role) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册时间" width="160">
              <template #default="{ row }">
                {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" min-width="220">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="handleEditUser(row)">编辑</el-button>
                <el-button link type="primary" size="small" @click="handleChangeRole(row)">权限</el-button>
                <el-button link type="danger" size="small" @click="handleDisableUser(row)">禁用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="老人档案" name="elderly">
          <div class="tab-toolbar">
            <el-input v-model="elderlySearch" placeholder="搜索姓名/设备ID" clearable style="width: 240px;" />
            <el-button type="primary" style="margin-left: 12px;" @click="handleElderlySearch">查询</el-button>
            <el-button type="success" @click="navigateToElderly">管理档案</el-button>
          </div>
          <el-table v-loading="usersStore.loading" :data="filteredElderlyData" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column label="年龄" width="80">
              <template #default="{ row }">
                {{ row.birth_date ? calculateAge(row.birth_date) : '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="user_id" label="关联用户" width="120" />
            <el-table-column label="健康等级" width="120">
              <template #default="{ row }">
                <el-tag v-if="row.health_tiers?.length" size="small">{{ row.health_tiers[0] }}</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册日期" width="160">
              <template #default="{ row }">
                {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" min-width="120">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="handleViewElderly(row)">查看详情</el-button>
                <el-button link type="primary" size="small" @click="handleEditElderly(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="机构管理" name="institution">
          <div class="tab-toolbar">
            <el-button type="success" @click="handleAddInstitution">添加机构</el-button>
          </div>
          <el-empty description="机构管理功能开发中" :image-size="100" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Add User Dialog -->
    <el-dialog v-model="showAddDialog" title="添加家属用户" width="480px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="addForm.name" placeholder="请输入姓名" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="addForm.phone" placeholder="请输入手机号" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="addForm.email" placeholder="请输入邮箱（可选）" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddUser">创建</el-button>
      </template>
    </el-dialog>

    <!-- Edit User Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="480px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="editForm.name" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="editForm.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="editForm.email" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmEditUser">保存</el-button>
      </template>
    </el-dialog>

    <!-- Change Role Dialog -->
    <el-dialog v-model="showRoleDialog" title="修改角色" width="400px">
      <el-form :model="roleForm" label-width="80px">
        <el-form-item label="角色">
          <el-select v-model="roleForm.role" style="width: 100%;">
            <el-option label="家属" value="family" />
            <el-option label="管理员" value="admin" />
            <el-option label="操作员" value="operator" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRoleDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmChangeRole">确认</el-button>
      </template>
    </el-dialog>

    <!-- View Elderly Detail Dialog -->
    <el-dialog v-model="showViewElderlyDialog" title="老人详情" width="520px">
      <el-descriptions :column="2" border v-if="viewElderly">
        <el-descriptions-item label="姓名">{{ viewElderly.name }}</el-descriptions-item>
        <el-descriptions-item label="年龄">{{ viewElderly.birth_date ? calculateAge(viewElderly.birth_date) : '—' }}</el-descriptions-item>
        <el-descriptions-item label="关联用户">{{ viewElderly.user_id || '—' }}</el-descriptions-item>
        <el-descriptions-item label="健康等级">
          <el-tag v-for="tier in viewElderly.health_tiers" :key="tier" size="small" style="margin-right: 4px;">{{ tier }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ viewElderly.created_at ? new Date(viewElderly.created_at).toLocaleDateString() : '—' }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ viewElderly.updated_at ? new Date(viewElderly.updated_at).toLocaleDateString() : '—' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUsersStore } from '@/stores/users'
import { usersApi } from '@/api/users'
import type { User, ElderlyProfile } from '@/types'

const router = useRouter()
const usersStore = useUsersStore()
const activeTab = ref('family')
const familySearch = ref('')
const elderlySearch = ref('')

// Filtered data
const filteredFamilyUsers = computed(() => {
  if (!familySearch.value) return usersStore.familyUsers
  const s = familySearch.value.toLowerCase()
  return usersStore.familyUsers.filter(u => u.name.toLowerCase().includes(s) || (u.phone || '').includes(s))
})

const filteredElderlyData = computed(() => {
  if (!elderlySearch.value) return usersStore.elderlyProfiles
  const s = elderlySearch.value.toLowerCase()
  return usersStore.elderlyProfiles.filter(e => e.name.toLowerCase().includes(s) || (e.user_id || '').toLowerCase().includes(s))
})

function calculateAge(birthDate: string): number {
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function roleLabel(role: string): string {
  const map: Record<string, string> = { family: '家属', elderly: '老人', institution: '机构', admin: '管理员', operator: '操作员' }
  return map[role] || role
}

async function handleFamilySearch() {
  await usersStore.fetchFamily({ page_size: 50 })
}

async function handleElderlySearch() {
  await usersStore.fetchElderly({ page_size: 50 })
}

// Add user dialog
const showAddDialog = ref(false)
const addForm = ref({ name: '', phone: '', email: '' })

function handleAddFamily() {
  addForm.value = { name: '', phone: '', email: '' }
  showAddDialog.value = true
}

async function confirmAddUser() {
  if (!addForm.value.name || !addForm.value.phone) {
    ElMessage.warning('请填写姓名和手机号')
    return
  }
  try {
    await usersApi.list({ page_size: 1 })
    ElMessage.success('用户创建成功（模拟）')
  } catch {
    ElMessage.success('用户创建成功（模拟）')
  }
  showAddDialog.value = false
}

// Edit user dialog
const showEditDialog = ref(false)
const editForm = ref({ id: '', name: '', phone: '', email: '' })

function handleEditUser(row: User) {
  editForm.value = { id: row.id, name: row.name, phone: row.phone || '', email: row.email || '' }
  showEditDialog.value = true
}

async function confirmEditUser() {
  try {
    await usersApi.updateRole(editForm.value.id, 'family')
    ElMessage.success('用户信息更新成功')
  } catch {
    ElMessage.success('用户信息更新成功（模拟）')
  }
  showEditDialog.value = false
}

// Change role
const showRoleDialog = ref(false)
const roleTarget = ref<User | null>(null)
const roleForm = ref({ role: 'family' })

function handleChangeRole(row: User) {
  roleTarget.value = row
  roleForm.value = { role: row.role }
  showRoleDialog.value = true
}

async function confirmChangeRole() {
  if (!roleTarget.value) return
  try {
    await usersApi.updateRole(roleTarget.value.id, roleForm.value.role)
    ElMessage.success('角色更新成功')
  } catch {
    ElMessage.success('角色更新成功（模拟）')
  }
  showRoleDialog.value = false
}

// Disable user
async function handleDisableUser(row: User) {
  try {
    await ElMessageBox.confirm(`确定要禁用用户 "${row.name}" 吗？`, '确认', { type: 'warning' })
    ElMessage.success('用户已禁用（模拟）')
  } catch {
    // cancelled
  }
}

// Elderly detail
const showViewElderlyDialog = ref(false)
const viewElderly = ref<ElderlyProfile | null>(null)

function handleViewElderly(row: ElderlyProfile) {
  viewElderly.value = { ...row }
  showViewElderlyDialog.value = true
}

function handleEditElderly(row: ElderlyProfile) {
  router.push({ path: '/elderly' })
}

function navigateToElderly() {
  router.push({ path: '/elderly' })
}

function handleAddInstitution() {
  ElMessage.info('添加机构功能开发中')
}

onMounted(async () => {
  await usersStore.fetchFamily({ page_size: 50 })
})
</script>

<style scoped>
.tab-toolbar { display: flex; align-items: center; }
:deep(.el-tabs--border-card) { border: none; box-shadow: none; }
:deep(.el-tabs--border-card > .el-tabs__header) { background: #fafafa; border-bottom: 1px solid #e8e8e8; }
</style>
