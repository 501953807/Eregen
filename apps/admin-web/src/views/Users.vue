<template>
  <div class="users-page">
    <!-- Page Header -->
    <div class="page-header">
      <h2 class="page-title">用户管理</h2>
      <el-button type="primary" @click="handleAddUser">+ 手动创建用户</el-button>
    </div>

    <!-- User Type Tabs -->
    <div class="user-tabs">
      <el-button
        v-for="tab in userTabs" :key="tab.name"
        :class="{ active: activeTab === tab.name }"
        @click="activeTab = tab.name"
      >
        {{ tab.label }}
        <span class="tab-count">{{ tab.count }}</span>
      </el-button>
    </div>

    <!-- KPI Row (4 columns) -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ stats.totalUsers }}</div>
          <div class="kpi-label">总用户数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ stats.monthlyActive }}</div>
          <div class="kpi-label">月活跃</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ stats.paidSubscriptions }}</div>
          <div class="kpi-label">付费订阅</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-red">
          <div class="kpi-value">{{ stats.todayNew }}</div>
          <div class="kpi-label">今日新增</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter Bar -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="12" align="middle">
        <el-col :span="4">
          <el-select v-model="filters.role" placeholder="全部角色" clearable>
            <el-option label="家属" value="family" />
            <el-option label="老人" value="elderly" />
            <el-option label="机构管理员" value="institution" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.tier" placeholder="全部等级" clearable>
            <el-option label="Pro" value="pro" />
            <el-option label="Plus" value="plus" />
            <el-option label="Starter" value="starter" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.registerTime" placeholder="注册时间" clearable>
            <el-option label="今天" value="today" />
            <el-option label="本周" value="week" />
            <el-option label="本月" value="month" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.subscription" placeholder="订阅状态" clearable>
            <el-option label="已付费" value="paid" />
            <el-option label="免费" value="free" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-input v-model="filters.search" placeholder="搜索用户名、手机号..." clearable />
        </el-col>
        <el-col :span="3" style="text-align: right;">
          <el-button @click="handleResetFilters">重置</el-button>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- User Cards Grid -->
    <div class="user-grid">
      <el-card
        v-for="user in paginatedUsers" :key="user.id"
        shadow="hover" class="user-card"
        @click="openSidePanel(user)"
      >
        <div class="user-card-header">
          <div class="user-avatar" :class="user.gender">
            {{ user.name.charAt(0) }}
          </div>
          <div class="user-name-info">
            <div class="user-name">{{ user.name }}</div>
            <div class="user-phone">{{ maskPhone(user.phone) }}</div>
          </div>
          <el-tag v-if="user.tier" size="small" :class="'tier-' + user.tier" effect="light">
            {{ tierLabel(user.tier) }}
          </el-tag>
        </div>

        <div class="user-stats">
          <div class="user-stat">
            <div class="user-stat-val">{{ user.elderlyCount }}</div>
            <div class="user-stat-lbl">关联老人</div>
          </div>
          <div class="user-stat">
            <div class="user-stat-val">{{ user.subscriptionDays }}</div>
            <div class="user-stat-lbl">订阅剩余</div>
          </div>
          <div class="user-stat">
            <div class="user-stat-val" :style="{ color: userStatColor(user.status) }">
              {{ user.statusText }}
            </div>
            <div class="user-stat-lbl">状态</div>
          </div>
        </div>

        <div class="user-tags">
          <el-tag v-if="user.verified" size="small" type="success" effect="plain" round>已实名认证</el-tag>
          <el-tag v-if="user.paid" size="small" type="primary" effect="plain" round>付费用户</el-tag>
          <el-tag v-if="(user as any).alerts && (user as any).alerts > 0" size="small" type="danger" effect="plain" round>{{ (user as any).alerts }}条未读告警</el-tag>
        </div>

        <div class="user-card-actions">
          <el-button link type="primary" size="small" @click.stop="openSidePanel(user)">详情</el-button>
          <el-button link type="primary" size="small" @click.stop="handleEditUser(user)">编辑</el-button>
          <el-button link type="primary" size="small" @click.stop="handleSendMessage(user)">消息</el-button>
          <el-button link type="danger" size="small" @click.stop="handleDisableUser(user)">禁用</el-button>
        </div>
      </el-card>
    </div>

    <!-- Pagination -->
    <div class="pagination-wrapper">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="filteredUsers.length"
        :page-size="pageSize"
        :current-page="currentPage"
        :page-sizes="[12, 24, 48]"
        @size-change="(v: number) => { pageSize = v; }"
        @current-change="(v: number) => { currentPage = v; }"
      />
    </div>

    <!-- Side Panel Overlay -->
    <div v-if="showSidePanel" class="side-panel-overlay" @click.self="showSidePanel = false">
      <div class="side-panel">
        <div class="panel-header">
          <span style="font-size:15px;font-weight:700;">用户详情</span>
          <button class="panel-close" @click="showSidePanel = false">&#10005;</button>
        </div>

        <div class="panel-body">
          <!-- Profile -->
          <div class="panel-profile">
            <div class="panel-avatar" :style="{ background: getPanelAvatarBg(selectedUser) }">
              {{ selectedUser?.name.charAt(0) }}
            </div>
            <div>
              <div class="panel-name">{{ selectedUser?.name }}</div>
              <div class="panel-role">{{ roleLabel(selectedUser?.role || '') }} · {{ tierLabel(selectedUser?.tier || '') }}订阅</div>
            </div>
          </div>

          <!-- Personal Info -->
          <div class="panel-section">
            <div class="panel-section-title">个人信息</div>
            <div class="panel-row"><span class="panel-row-label">姓名</span><span class="panel-row-value">{{ selectedUser?.name }}</span></div>
            <div class="panel-row"><span class="panel-row-label">手机号</span><span class="panel-row-value">{{ selectedUser?.phone || '—' }}</span></div>
            <div class="panel-row"><span class="panel-row-label">邮箱</span><span class="panel-row-value">{{ selectedUser?.email || '—' }}</span></div>
            <div class="panel-row"><span class="panel-row-label">注册时间</span><span class="panel-row-value">{{ formatDate(selectedUser?.created_at) }}</span></div>
            <div class="panel-row"><span class="panel-row-label">最后登录</span><span class="panel-row-value">{{ selectedUser?.last_login || '—' }}</span></div>
            <div class="panel-row">
              <span class="panel-row-label">实名状态</span>
              <span class="panel-row-value" :style="{ color: selectedUser?.verified ? '#16A34A' : '' }">
                {{ selectedUser?.verified ? '✓ 已认证' : '未认证' }}
              </span>
            </div>
          </div>

          <!-- Subscription Info -->
          <div class="panel-section">
            <div class="panel-section-title">订阅信息</div>
            <div class="panel-row">
              <span class="panel-row-label">套餐</span>
              <span class="panel-row-value" :style="{ color: selectedUser?.tier === 'pro' ? '#7C3AED' : '#2563EB' }">
                {{ tierLabel(selectedUser?.tier || '') }} {{ (selectedUser as any).sub_type || '' }}
              </span>
            </div>
            <div class="panel-row"><span class="panel-row-label">到期时间</span><span class="panel-row-value">{{ formatDate((selectedUser as any).sub_expires) }}</span></div>
            <div class="panel-row"><span class="panel-row-label">月费</span><span class="panel-row-value">¥{{ subAmount(selectedUser?.tier) }}/月</span></div>
            <div class="panel-row"><span class="panel-row-label">支付方式</span><span class="panel-row-value">{{ (selectedUser as any).pay_method || '—' }}</span></div>
          </div>

          <!-- Related Elderly -->
          <div class="panel-section">
            <div class="panel-section-title">关联老人</div>
            <div
              v-for="(profile, i) in (selectedUser as any).elderly_profiles"
              :key="i"
              class="panel-row"
              style="cursor:pointer;color:#2563EB;"
              @click="viewElderlyProfile(profile)"
            >
              {{ profile.name }}（{{ profile.relation }}）· {{ profile.devices || '无设备' }}
            </div>
            <div v-if="!selectedUser || !(selectedUser as any).elderly_profiles?.length" style="color:#909399;font-size:13px;padding:6px 0;">暂无关联老人</div>
          </div>

          <!-- Recent Activity -->
          <div class="panel-section">
            <div class="panel-section-title">最近活动</div>
            <div class="activity-list">
              <div v-for="(item, i) in activityTimeline" :key="i" class="activity-item">
                <div class="activity-dot" :class="item.dotClass"></div>
                <div class="activity-text">{{ item.text }}</div>
                <div class="activity-time">{{ item.time }}</div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="panel-section">
            <div class="panel-section-title">操作</div>
            <div class="panel-actions">
              <el-button type="primary" size="small" style="flex:1;">发送通知</el-button>
              <el-button size="small" style="flex:1;">编辑信息</el-button>
              <el-button size="small" style="flex:1;">查看日志</el-button>
              <el-button size="small" style="flex:1;color:#EF4444;border-color:#EF4444;">禁用账号</el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add User Dialog -->
    <el-dialog v-model="showAddDialog" title="创建用户" width="480px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="addForm.name" placeholder="请输入姓名" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="addForm.phone" placeholder="请输入手机号" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="addForm.email" placeholder="请输入邮箱（可选）" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="addForm.role" style="width: 100%;">
            <el-option label="家属" value="family" />
            <el-option label="老人" value="elderly" />
            <el-option label="机构管理员" value="institution" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddUser">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUsersStore } from '@/stores/users'
import { usersApi } from '@/api/users'
import type { User } from '@/types'

const usersStore = useUsersStore()
const activeTab = ref('all')

// Tab definitions with counts
const userTabs = computed(() => [
  { name: 'all', label: '全部用户', count: usersStore.familyUsers.length + usersStore.elderlyProfiles.length },
  { name: 'family', label: '家属', count: usersStore.familyUsers.length },
  { name: 'elderly', label: '老人', count: usersStore.elderlyProfiles.length },
  { name: 'institution', label: '机构', count: 0 },
])

// Stats
const stats = computed(() => ({
  totalUsers: usersStore.familyUsers.length + usersStore.elderlyProfiles.length,
  monthlyActive: Math.round(usersStore.familyUsers.length * 0.75),
  paidSubscriptions: usersStore.familyUsers.filter(u => u.tier === 'pro' || u.tier === 'plus').length,
  todayNew: 3,
}))

// Filters
const filters = ref({
  role: '',
  tier: '',
  registerTime: '',
  subscription: '',
  search: '',
})

interface UserCard extends User {
  gender?: string
  elderlyCount: number
  subscriptionDays: string
  statusText: string
}

const filteredUsers = computed<UserCard[]>(() => {
  let list: UserCard[] = usersStore.familyUsers.map(u => ({
    ...u,
    gender: 'male',
    elderlyCount: (u as any).elderly_profiles?.length || 0,
    subscriptionDays: u.tier === 'pro' ? '14天' : u.tier === 'plus' ? '28天' : '—',
    statusText: '活跃',
  }))

  if (activeTab.value === 'all' || activeTab.value === 'elderly') {
    list = list.concat(
      usersStore.elderlyProfiles.map(e => ({
        id: e.id || '',
        name: e.name,
        phone: '',
        email: '',
        role: 'elderly',
        created_at: e.created_at,
        gender: 'female',
        elderlyCount: 1,
        subscriptionDays: '—',
        statusText: '正常',
      } as UserCard))
    )
  }

  if (filters.value.search) {
    const q = filters.value.search.toLowerCase()
    list = list.filter(u => u.name.toLowerCase().includes(q) || (u.phone || '').includes(q))
  }

  return list
})

// Pagination
const currentPage = ref(1)
const pageSize = ref(12)

const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return filteredUsers.value.slice(start, start + pageSize.value)
})

// Helpers
function maskPhone(phone?: string): string {
  if (!phone) return '—'
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

function tierLabel(tier?: string): string {
  const map: Record<string, string> = { pro: 'PRO', plus: 'PLUS', starter: '基础', free: '免费' }
  return map[tier || ''] || '—'
}

function roleLabel(role: string): string {
  const map: Record<string, string> = { family: '家属', elderly: '老人', institution: '机构', admin: '管理员' }
  return map[role] || role
}

function userStatColor(status: string): string {
  if (status === '活跃') return '#16A34A'
  if (status === '正常') return '#16A34A'
  if (status.includes('未活')) return '#F59E0B'
  return '#909399'
}

function subAmount(tier?: string): string {
  const map: Record<string, string> = { pro: '99', plus: '59', starter: '29' }
  return map[tier || ''] || '0'
}

function formatDate(date?: string): string {
  if (!date) return '—'
  return new Date(date).toLocaleDateString('zh-CN')
}

function getPanelAvatarBg(user?: User | null): string {
  if (!user) return '#f3f4f6'
  if ((user as any).gender === 'male') return '#DBEAFE'
  if ((user as any).gender === 'female') return '#FCE7F3'
  return '#F3E8FF'
}

function handleSearch() {
  currentPage.value = 1
}

function handleResetFilters() {
  filters.value = { role: '', tier: '', registerTime: '', subscription: '', search: '' }
  currentPage.value = 1
}

// Side Panel
const showSidePanel = ref(false)
const selectedUser = ref<UserCard & { elderly_profiles?: any[] } | null>(null)

function openSidePanel(user: UserCard & { elderly_profiles?: any[] }) {
  selectedUser.value = user
  showSidePanel.value = true
}

function viewElderlyProfile(profile: any) {
  ElMessage.info(`查看 ${profile.name} 的档案`)
}

// Activity Timeline
const activityTimeline = [
  { text: '登录家属APP', time: '2小时前', dotClass: 'login' },
  { text: '收到SOS告警通知', time: '昨天', dotClass: 'alert' },
  { text: '修改用药规则', time: '3天前', dotClass: 'config' },
  { text: '登录家属APP', time: '5天前', dotClass: 'login' },
]

// Add User Dialog
const showAddDialog = ref(false)
const addForm = ref({ name: '', phone: '', email: '', role: 'family' })

function handleAddUser() {
  addForm.value = { name: '', phone: '', email: '', role: 'family' }
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

function handleEditUser(user: any) {
  ElMessage.info(`编辑用户: ${user.name}`)
}

function handleSendMessage(user: any) {
  ElMessage.info(`发送消息给: ${user.name}`)
}

async function handleDisableUser(user: any) {
  try {
    await ElMessageBox.confirm(`确定要禁用用户 "${user.name}" 吗？`, '确认', { type: 'warning' })
    ElMessage.success('用户已禁用（模拟）')
  } catch { /* cancelled */ }
}

onMounted(async () => {
  await usersStore.fetchFamily({ page_size: 50 })
})
</script>

<style scoped>
.users-page {
  padding: 0;
}

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

/* User Type Tabs — pill style */
.user-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  background: white;
  border-radius: 14px;
  padding: 4px;
  border: 1px solid var(--el-border-color-light);
  width: fit-content;
  box-shadow: var(--el-box-shadow-light);
}

.user-tabs .el-button {
  border-radius: 12px;
  font-size: 13px;
  font-weight: 700;
  padding: 9px 22px;
  border: none;
  background: transparent;
  color: var(--el-text-color-secondary);
  transition: all 0.2s;
}

.user-tabs .el-button.active {
  background: linear-gradient(135deg, #2563EB, #7C3AED);
  color: white;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.25);
}

.tab-count {
  font-size: 11px;
  opacity: 0.7;
  margin-left: 6px;
  font-weight: 600;
}

/* KPI Cards — v2 blue/purple palette */
.kpi-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-radius: 14px;
}

.kpi-value {
  font-size: 28px;
  font-weight: 800;
  line-height: 1.2;
}

.kpi-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 6px;
  font-weight: 600;
}

.kpi-blue .kpi-value { color: #2563EB; }
.kpi-green .kpi-value { color: #16A34A; }
.kpi-purple .kpi-value { color: #7C3AED; }
.kpi-red .kpi-value { color: #EF4444; }

/* Filter Card */
.filter-card :deep(.el-card__body) {
  padding: 12px 16px;
}

/* User Grid */
.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.user-card {
  cursor: pointer;
  transition: all 0.2s;
  border-radius: 14px;
  border: 1px solid var(--el-border-color-light);
  background: white;
}

.user-card:hover {
  border-color: #2563EB;
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.1);
  transform: translateY(-2px);
}

.user-card :deep(.el-card__body) {
  padding: 16px;
}

.user-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.user-avatar {
  width: 44px;
  height: 44px;
  border-radius: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  flex-shrink: 0;
  background: #F3F4F6;
  color: var(--el-text-color-placeholder);
}

.user-avatar.male { background: #DBEAFE; color: #2563EB; }
.user-avatar.female { background: #FCE7F3; color: #EC4899; }

.user-name-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.user-phone {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  font-family: monospace;
}

.tier-pro { background: #EDE9FE; color: #7C3AED; }
.tier-plus { background: #DBEAFE; color: #2563EB; }
.tier-starter { background: #F3F4F6; color: #6B7280; }

.user-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 12px;
}

.user-stat {
  text-align: center;
  padding: 8px;
  background: #F9FAFB;
  border-radius: 8px;
}

.user-stat-val {
  font-size: 16px;
  font-weight: 700;
}

.user-stat-lbl {
  font-size: 10px;
  color: var(--el-text-color-placeholder);
}

.user-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.user-card-actions {
  display: flex;
  gap: 6px;
  padding-top: 12px;
  border-top: 1px solid #F3F4F6;
}

/* Pagination */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

/* Side Panel */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 200;
  display: flex;
  justify-content: flex-end;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.side-panel {
  width: 520px;
  max-width: 90vw;
  background: white;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
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
  transition: background 0.15s;
}
.panel-close:hover {
  background: var(--el-border-color-light);
}

.panel-body {
  padding: 20px 24px;
}

.panel-profile {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.panel-avatar {
  width: 56px;
  height: 56px;
  border-radius: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  flex-shrink: 0;
}

.panel-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.panel-role {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
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

/* Activity Timeline */
.activity-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.activity-item {
  display: flex;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #F3F4F6;
  font-size: 12px;
  align-items: flex-start;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-top: 5px;
  flex-shrink: 0;
}

.activity-dot.login { background: #16A34A; }
.activity-dot.alert { background: #EF4444; }
.activity-dot.config { background: #2563EB; }

.activity-text {
  flex: 1;
  font-weight: 600;
}

.activity-time {
  color: var(--el-text-color-placeholder);
  white-space: nowrap;
  font-size: 11px;
}

.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
