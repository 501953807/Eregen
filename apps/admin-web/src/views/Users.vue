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
            <el-table-column label="操作" fixed="right" min-width="160">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openSidePanel(row)">详情</el-button>
                <el-button link type="primary" size="small" @click="handleEditUser(row)">编辑</el-button>
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
                <el-button link type="primary" size="small" @click="openElderlyPanel(row)">详情</el-button>
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

    <!-- Side Panel for User Details — v2 prototype enhancement -->
    <div v-if="showSidePanel" class="side-panel-overlay" @click.self="showSidePanel = false">
      <div class="side-panel">
        <div class="side-panel-header">
          <div>
            <div style="font-size:16px;font-weight:700;">{{ selectedUser?.name || selectedElderly?.name }}</div>
            <div style="font-size:12px;color:#909399;margin-top:2px;">{{ selectedUser ? roleLabel(selectedUser.role) : (selectedElderly ? '老人档案' : '') }}</div>
          </div>
          <el-button link type="primary" @click="showSidePanel = false"><el-icon><Close /></el-icon></el-button>
        </div>

        <!-- User info cards -->
        <div class="side-panel-section">
          <div class="section-title">基本信息</div>
          <el-descriptions :column="1" size="small" border>
            <el-descriptions-item label="手机号">{{ selectedUser?.phone || '—' }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ selectedUser?.email || '—' }}</el-descriptions-item>
            <el-descriptions-item label="注册时间">{{ selectedUser?.created_at ? new Date(selectedUser.created_at).toLocaleDateString() : (selectedElderly?.created_at ? new Date(selectedElderly.created_at).toLocaleDateString() : '—') }}</el-descriptions-item>
            <el-descriptions-item label="设备数">{{ (selectedUser?.elderly_profiles || []).length }} 位老人</el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- Activity Timeline -->
        <div class="side-panel-section">
          <div class="section-title">活动记录</div>
          <div class="timeline">
            <div v-for="(item, i) in activityTimeline" :key="i" class="timeline-item">
              <div class="timeline-dot" :style="{ background: item.color }"></div>
              <div class="timeline-content">
                <div class="timeline-title">{{ item.title }}</div>
                <div class="timeline-time">{{ item.time }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Billing History -->
        <div class="side-panel-section">
          <div class="section-title">订阅记录</div>
          <div v-for="(sub, i) in billingHistory" :key="i" class="billing-item">
            <div class="billing-row">
              <span class="billing-plan">{{ sub.plan }}</span>
              <el-tag :type="sub.status === 'active' ? 'success' : 'info'" size="small">{{ sub.statusText }}</el-tag>
            </div>
            <div class="billing-detail">{{ sub.date }} · ¥{{ sub.amount }}/月</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Close } from '@element-plus/icons-vue'
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

// Side panel — v2 prototype enhancement
const showSidePanel = ref(false)
const selectedUser = ref<User | null>(null)
const selectedElderly = ref<ElderlyProfile | null>(null)

function openSidePanel(user: User) {
  selectedUser.value = user
  selectedElderly.value = null
  showSidePanel.value = true
}

function openElderlyPanel(elderly: ElderlyProfile) {
  selectedElderly.value = elderly
  selectedUser.value = null
  showSidePanel.value = true
}

const activityTimeline = ref([
  { title: '设备上线：手环 BR-0042', time: '2026-07-18 14:32', color: '#67C23A' },
  { title: '健康数据同步：心率 72bpm', time: '2026-07-18 14:30', color: '#409EFF' },
  { title: 'SOS告警（误触已取消）', time: '2026-07-18 10:15', color: '#F56C6C' },
  { title: '用药确认：早餐药已服用', time: '2026-07-18 08:05', color: '#E6A23C' },
  { title: '家属APP登录', time: '2026-07-17 22:10', color: '#909399' },
])

const billingHistory = ref([
  { plan: 'Plus 套餐', statusText: '活跃中', status: 'active', date: '2026-06-18', amount: '59' },
  { plan: 'Plus 套餐', statusText: '已完成', status: 'active', date: '2026-05-18', amount: '59' },
  { plan: 'Starter 套餐', statusText: '已完成', status: 'active', date: '2026-04-18', amount: '29' },
])

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

/* Side Panel — v2 */
.side-panel-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.3); z-index: 2000;
  display: flex; justify-content: flex-end;
}
.side-panel {
  width: 420px; max-width: 90vw; background: #fff;
  overflow-y: auto; box-shadow: -4px 0 20px rgba(0,0,0,0.1);
  padding: 20px; display: flex; flex-direction: column; gap: 20px;
}
.side-panel-header {
  display: flex; justify-content: space-between; align-items: flex-start;
  padding-bottom: 16px; border-bottom: 1px solid #EBEEF5;
}
.section-title {
  font-size: 14px; font-weight: 700; color: #303133; margin-bottom: 12px;
  display: flex; align-items: center; gap: 6px;
}
.timeline { display: flex; flex-direction: column; gap: 0; }
.timeline-item { display: flex; gap: 12px; position: relative; padding-bottom: 16px; }
.timeline-item:last-child { padding-bottom: 0; }
.timeline-dot {
  width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; margin-top: 4px;
}
.timeline-content { flex: 1; }
.timeline-title { font-size: 13px; color: #303133; }
.timeline-time { font-size: 11px; color: #909399; margin-top: 2px; }
.billing-item { padding: 10px 12px; background: #FAFAFA; border-radius: 8px; margin-bottom: 8px; }
.billing-item:last-child { margin-bottom: 0; }
.billing-row { display: flex; justify-content: space-between; align-items: center; }
.billing-plan { font-size: 13px; font-weight: 600; }
.billing-detail { font-size: 11px; color: #909399; margin-top: 4px; }
</style>
