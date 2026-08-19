<template>
  <div class="users-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h2 class="page-title">用户管理</h2>
        <p class="page-subtitle">管理所有家属、老人和机构用户</p>
      </div>
      <HopeBtn variant="filled" size="md" @click="handleAddUser">
        <template #icon>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </template>
        手动创建用户
      </HopeBtn>
    </div>

    <!-- User Type Tabs -->
    <HopeTabs
      :model-value="activeTab as string"
      :tabs="tabItems"
      pill-style
      @update:model-value="(v: string | number) => { activeTab = typeof v === 'string' ? v : String(v); }"
    />

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="stats.totalUsers"
        label="总用户数"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.monthlyActive"
        label="月活跃用户"
        icon-color="success"
        gradient="linear-gradient(135deg, #1aa053 0%, #22c55e 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.paidSubscriptions"
        label="付费订阅"
        icon-color="accent"
        gradient="linear-gradient(135deg, #8C57FF 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.todayNew"
        label="今日新增"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938 0%, #f59e0b 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.institutionUsers"
        label="机构用户"
        icon-color="info"
        gradient="linear-gradient(135deg, #079aa2 0%, #14b8a6 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21h18M3 10h18M5 6l7-3 7 3M4 10v11M20 10v11M8 14v3M12 14v3M16 14v3"/></svg></el-icon>
        </template>
      </HopeStatCard>
    </div>

    <!-- Filter Bar — HopeCard -->
    <HopeCard>
      <template #header>
        <span class="filter-title">筛选条件</span>
      </template>
      <div class="filter-row">
        <div class="filter-item">
          <label class="filter-label">角色</label>
          <el-select v-model="filters.role" placeholder="全部角色" clearable class="hope-select">
            <el-option label="家属" value="family" />
            <el-option label="老人" value="elderly" />
            <el-option label="机构管理员" value="institution" />
          </el-select>
        </div>
        <div class="filter-item">
          <label class="filter-label">套餐等级</label>
          <el-select v-model="filters.tier" placeholder="全部等级" clearable class="hope-select">
            <el-option label="Pro" value="pro" />
            <el-option label="Plus" value="plus" />
            <el-option label="Starter" value="starter" />
          </el-select>
        </div>
        <div class="filter-item">
          <label class="filter-label">注册时间</label>
          <el-select v-model="filters.registerTime" placeholder="全部时间" clearable class="hope-select">
            <el-option label="今天" value="today" />
            <el-option label="本周" value="week" />
            <el-option label="本月" value="month" />
          </el-select>
        </div>
        <div class="filter-item">
          <label class="filter-label">订阅状态</label>
          <el-select v-model="filters.subscription" placeholder="全部状态" clearable class="hope-select">
            <el-option label="已付费" value="paid" />
            <el-option label="免费" value="free" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </div>
        <div class="filter-item filter-item--search">
          <label class="filter-label">搜索</label>
          <el-input v-model="filters.search" placeholder="用户名、手机号..." clearable class="hope-input" />
        </div>
        <div class="filter-actions">
          <HopeBtn variant="plain" size="sm" @click="handleResetFilters">重置</HopeBtn>
          <HopeBtn variant="filled" size="sm" @click="handleSearch">搜索</HopeBtn>
        </div>
      </div>
    </HopeCard>

    <!-- User Cards Grid -->
    <div class="user-grid">
      <div
        v-for="user in paginatedUsers" :key="user.id"
        class="user-card"
        @click="openSidePanel(user)"
      >
        <div class="user-card__top">
          <div class="user-card__avatar-row">
            <HopeAvatar
              :name="user.name"
              :size="user.gender === 'male' ? 'lg' : 'lg'"
              :style="{ '--hope-avatar-bg': user.gender === 'male' ? '#DBEAFE' : '#FCE7F3', '--hope-avatar-text-color': user.gender === 'male' ? '#6E9FC4' : '#D48EC0' }"
            />
            <div class="user-card__info">
              <div class="user-card__name">{{ user.name }}</div>
              <div class="user-card__phone">{{ maskPhone(user.phone) }}</div>
            </div>
          </div>
          <HopeBadge v-if="user.tier" :color="tierBadgeColor(user.tier)">
            {{ tierLabel(user.tier) }}
          </HopeBadge>
        </div>

        <div class="user-card__stats">
          <div class="stat-cell">
            <span class="stat-val">{{ user.elderlyCount }}</span>
            <span class="stat-lbl">关联老人</span>
          </div>
          <div class="stat-cell">
            <span class="stat-val">{{ user.subscriptionDays }}</span>
            <span class="stat-lbl">订阅剩余</span>
          </div>
          <div class="stat-cell">
            <span class="stat-val" :class="statStatusClass(user.statusText)">{{ user.statusText }}</span>
            <span class="stat-lbl">状态</span>
          </div>
        </div>

        <div class="user-card__tags">
          <HopeBadge v-if="user.verified" color="success" type="text">已实名认证</HopeBadge>
          <HopeBadge v-if="user.paid" color="primary" type="text">付费用户</HopeBadge>
          <HopeBadge
            v-if="(user as any).alerts && (user as any).alerts > 0"
            color="error"
            type="text"
          >
            {{ (user as any).alerts }}条未读告警
          </HopeBadge>
        </div>

        <div class="user-card__actions">
          <HopeBtn variant="text" size="sm" @click.stop="openSidePanel(user)">详情</HopeBtn>
          <HopeBtn variant="text" size="sm" @click.stop="handleEditUser(user)">编辑</HopeBtn>
          <HopeBtn variant="text" size="sm" @click.stop="handleSendMessage(user)">消息</HopeBtn>
          <HopeBtn variant="error" size="sm" @click.stop="handleDisableUser(user)">禁用</HopeBtn>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredUsers.length === 0" class="empty-state">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--hope-text-muted)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
      </svg>
      <p>暂无用户数据</p>
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
        <!-- Panel Header -->
        <div class="panel-header">
          <span class="panel-header__title">用户详情</span>
          <button class="panel-close" @click="showSidePanel = false" aria-label="关闭">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>

        <div class="panel-body">
          <!-- Profile -->
          <div class="panel-profile">
            <HopeAvatar
              :name="selectedUser?.name || ''"
              size="xl"
              :style="{ '--hope-avatar-bg': getPanelAvatarBg(selectedUser), '--hope-avatar-text-color': '#fff' }"
            />
            <div class="panel-profile__info">
              <div class="panel-profile__name">{{ selectedUser?.name }}</div>
              <div class="panel-profile__role">{{ roleLabel(selectedUser?.role || '') }} · {{ tierLabel(selectedUser?.tier || '') }}订阅</div>
            </div>
            <div class="panel-profile__badge">
              <HopeBadge v-if="selectedUser?.verified" color="success" type="text">已认证</HopeBadge>
              <HopeBadge v-else color="warning" type="text">未认证</HopeBadge>
            </div>
          </div>

          <!-- Personal Info -->
          <div class="panel-section">
            <div class="panel-section-title">个人信息</div>
            <div class="panel-rows">
              <div class="panel-row">
                <span class="panel-row-label">姓名</span>
                <span class="panel-row-value">{{ selectedUser?.name }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">手机号</span>
                <span class="panel-row-value">{{ selectedUser?.phone || '—' }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">邮箱</span>
                <span class="panel-row-value">{{ selectedUser?.email || '—' }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">注册时间</span>
                <span class="panel-row-value">{{ formatDate(selectedUser?.created_at) }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">最后登录</span>
                <span class="panel-row-value">{{ selectedUser?.last_login || '—' }}</span>
              </div>
            </div>
          </div>

          <!-- Subscription Info -->
          <div class="panel-section">
            <div class="panel-section-title">订阅信息</div>
            <div class="panel-rows">
              <div class="panel-row">
                <span class="panel-row-label">套餐</span>
                <span class="panel-row-value panel-row-value--highlight">{{ tierLabel(selectedUser?.tier || '') }} {{ (selectedUser as any).sub_type || '' }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">到期时间</span>
                <span class="panel-row-value">{{ formatDate((selectedUser as any).sub_expires) }}</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">月费</span>
                <span class="panel-row-value">¥{{ subAmount(selectedUser?.tier) }} / 月</span>
              </div>
              <div class="panel-row">
                <span class="panel-row-label">支付方式</span>
                <span class="panel-row-value">{{ (selectedUser as any).pay_method || '—' }}</span>
              </div>
            </div>
          </div>

          <!-- Related Elderly -->
          <div class="panel-section">
            <div class="panel-section-title">关联老人</div>
            <div
              v-for="(profile, i) in (selectedUser as any).elderly_profiles"
              :key="i"
              class="elderly-link"
              @click="viewElderlyProfile(profile)"
            >
              <HopeAvatar :name="profile.name" size="sm" />
              <span class="elderly-name">{{ profile.name }}</span>
              <span class="elderly-relation">{{ profile.relation }}</span>
              <span class="elderly-devices">{{ profile.devices || '无设备' }}</span>
            </div>
            <div v-if="!selectedUser || !(selectedUser as any).elderly_profiles?.length" class="empty-text">暂无关联老人</div>
          </div>

          <!-- Recent Activity — HopeTimeline -->
          <div class="panel-section">
            <div class="panel-section-title">最近活动</div>
            <HopeTimeline
              :items="timelineItems"
            />
          </div>

          <!-- Actions -->
          <div class="panel-section">
            <div class="panel-section-title">操作</div>
            <div class="panel-actions">
              <HopeBtn variant="filled" size="sm" style="flex:1;">发送通知</HopeBtn>
              <HopeBtn variant="outlined" size="sm" style="flex:1;">编辑信息</HopeBtn>
              <HopeBtn variant="outlined" size="sm" style="flex:1;">查看日志</HopeBtn>
              <HopeBtn variant="ghost" size="sm" style="flex:1;">禁用账号</HopeBtn>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add User Dialog -->
    <el-dialog v-model="showAddDialog" title="创建用户" width="480px" class="hope-dialog">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="addForm.name" placeholder="请输入姓名" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="addForm.phone" placeholder="请输入手机号" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="addForm.email" placeholder="请输入邮箱（可选）" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="addForm.password" type="password" placeholder="请输入密码" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="addForm.role" style="width: 100%;">
            <el-option label="家属" value="family" />
            <el-option label="老人" value="elderly" />
            <el-option label="机构管理员" value="institution" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" @click="showAddDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="confirmAddUser">创建</HopeBtn>
      </template>
    </el-dialog>

    <!-- Edit User Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="480px" class="hope-dialog">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="editForm.name" placeholder="请输入姓名" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="editForm.phone" placeholder="请输入手机号" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="editForm.email" placeholder="请输入邮箱（可选）" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" style="width: 100%;">
            <el-option label="家属" value="family" />
            <el-option label="老人" value="elderly" />
            <el-option label="机构管理员" value="institution" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <HopeBtn variant="plain" @click="showEditDialog = false">取消</HopeBtn>
        <HopeBtn variant="filled" @click="confirmEditUser">保存</HopeBtn>
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
import {
  HopeBtn,
  HopeStatCard,
  HopeBadge,
  HopeAvatar,
  HopeCard,
  HopeTabs,
  HopeTimeline,
} from '@/components/hope'

const usersStore = useUsersStore()
const activeTab = ref('all')

// Tab definitions with counts
const tabItems = computed(() => [
  { value: 'all', label: '全部用户', badge: usersStore.familyUsers.length + usersStore.elderlyProfiles.length },
  { value: 'family', label: '家属', badge: usersStore.familyUsers.length },
  { value: 'elderly', label: '老人', badge: usersStore.elderlyProfiles.length },
  { value: 'institution', label: '机构', badge: 0 },
])

// Stats
const stats = computed(() => ({
  totalUsers: usersStore.familyUsers.length + usersStore.elderlyProfiles.length,
  monthlyActive: Math.round(usersStore.familyUsers.length * 0.75),
  paidSubscriptions: usersStore.familyUsers.filter(u => u.tier === 'pro' || u.tier === 'plus').length,
  todayNew: 3,
  institutionUsers: 12,
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
        role: 'operator' as any,
        created_at: (e as any).created_at || '',
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

function tierBadgeColor(tier?: string): 'primary' | 'accent' | 'info' {
  if (tier === 'pro') return 'accent'
  if (tier === 'plus') return 'primary'
  return 'info'
}

function statStatusClass(status: string): string {
  if (status === '活跃' || status === '正常') return 'stat-ok'
  if (status.includes('未活')) return 'stat-warn'
  return 'stat-muted'
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
  if (!user) return '#3a57e8'
  if ((user as any).gender === 'male') return '#3a57e8'
  if ((user as any).gender === 'female') return '#8C57FF'
  return '#3a57e8'
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
const timelineItems = computed(() => [
  { title: '登录家属APP', meta: '2小时前', color: 'success' as const },
  { title: '收到SOS告警通知', meta: '昨天', color: 'error' as const },
  { title: '修改用药规则', meta: '3天前', color: 'info' as const },
  { title: '登录家属APP', meta: '5天前', color: 'success' as const },
])

// Add User Dialog
const showAddDialog = ref(false)
const addForm = ref({ name: '', phone: '', email: '', role: 'family', password: '' })
const showEditDialog = ref(false)
const editForm = ref({ id: '', name: '', phone: '', email: '', role: 'family' })

function handleAddUser() {
  addForm.value = { name: '', phone: '', email: '', role: 'family', password: '' }
  showAddDialog.value = true
}

async function confirmAddUser() {
  if (!addForm.value.name || !addForm.value.phone || !addForm.value.password) {
    ElMessage.warning('请填写姓名、手机号和密码')
    return
  }
  try {
    await usersApi.create({
      name: addForm.value.name,
      phone: addForm.value.phone,
      email: addForm.value.email || undefined,
      role: addForm.value.role,
      password: addForm.value.password,
    })
    ElMessage.success('用户创建成功')
    addForm.value = { name: '', phone: '', email: '', role: 'family', password: '' }
    showAddDialog.value = false
    await usersStore.fetchFamily({ page_size: 50 })
    await usersStore.fetchElderly({ page_size: 50 })
  } catch {
    ElMessage.error('创建用户失败，请重试')
  }
}

function handleEditUser(user: any) {
  editForm.value = {
    id: user.id,
    name: user.name,
    phone: user.phone || '',
    email: user.email || '',
    role: user.role || 'family',
  }
  showEditDialog.value = true
}

async function confirmEditUser() {
  if (!editForm.value.name) {
    ElMessage.warning('请填写姓名')
    return
  }
  try {
    await usersApi.update(editForm.value.id, {
      name: editForm.value.name,
      phone: editForm.value.phone || undefined,
      email: editForm.value.email || undefined,
      role: editForm.value.role,
    })
    ElMessage.success('用户信息已更新')
    showEditDialog.value = false
    await usersStore.fetchFamily({ page_size: 50 })
    await usersStore.fetchElderly({ page_size: 50 })
  } catch {
    ElMessage.error('更新失败，请重试')
  }
}

function handleSendMessage(user: any) {
  ElMessage.info(`发送消息给: ${user.name}`)
}

async function handleDisableUser(user: any) {
  try {
    await ElMessageBox.confirm(`确定要禁用用户 "${user.name}" 吗？`, '确认', { type: 'warning' })
    await usersApi.delete(user.id)
    ElMessage.success('用户已禁用')
    await usersStore.fetchFamily({ page_size: 50 })
    await usersStore.fetchElderly({ page_size: 50 })
  } catch { /* cancelled */ }
}

onMounted(async () => {
  await usersStore.fetchFamily({ page_size: 50 })
  await usersStore.fetchElderly({ page_size: 50 })
})
</script>

<style scoped>
.users-page {
  padding: 0;
}

/* ─── Page Header ─── */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
}

.page-header__left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--hope-text);
  margin: 0;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin: 0;
  font-weight: 500;
}

/* ─── KPI Grid ─── */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 20px;
}

/* ─── Filter Bar ─── */
.filter-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--hope-text);
}

.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 14px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  flex: 1;
  min-width: 140px;
}

.filter-item--search {
  flex: 2;
  min-width: 220px;
}

.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.filter-actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
  padding-bottom: 1px;
}

/* Hope UI Select/Input overrides */
:deep(.hope-select) {
  width: 100%;
}
:deep(.hope-select .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
  padding: 5px 11px !important;
}
:deep(.hope-select .el-input__wrapper:hover) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}
:deep(.hope-select .el-input__wrapper.is-focus) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}

:deep(.hope-input .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
}
:deep(.hope-input .el-input__wrapper:hover) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}
:deep(.hope-input .el-input__wrapper.is-focus) {
  box-shadow: var(--hope-shadow-input-focus) !important;
}

/* ─── User Grid ─── */
.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
  margin-bottom: 20px;
}

.user-card {
  background: var(--hope-surface);
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-lg);
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--hope-shadow-sm);
}

.user-card:hover {
  border-color: var(--hope-primary);
  box-shadow: var(--hope-shadow-md);
  transform: translateY(-2px);
}

.user-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.user-card__avatar-row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.user-card__info {
  min-width: 0;
}

.user-card__name {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-card__phone {
  font-size: 12px;
  color: var(--hope-text-muted);
  font-family: monospace;
  margin-top: 2px;
}

.user-card__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin-bottom: 12px;
  padding: 10px 0;
  border-top: 1px solid var(--hope-border);
  border-bottom: 1px solid var(--hope-border);
}

.stat-cell {
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-val {
  font-size: 15px;
  font-weight: 700;
  color: var(--hope-text);
}

.stat-lbl {
  font-size: 11px;
  color: var(--hope-text-muted);
  font-weight: 500;
}

.stat-ok { color: var(--hope-success); }
.stat-warn { color: var(--hope-warning); }
.stat-muted { color: var(--hope-text-muted); }

.user-card__tags {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.user-card__actions {
  display: flex;
  gap: 4px;
}

/* ─── Empty State ─── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: var(--hope-text-muted);
}

.empty-state p {
  font-size: 14px;
  font-weight: 500;
  margin: 0;
}

/* ─── Pagination ─── */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
  margin-bottom: 24px;
}

:deep(.el-pagination) {
  --el-pagination-button-bg-color: var(--hope-surface);
  --el-pagination-button-border-radius: var(--hope-radius-md);
}

:deep(.el-pagination .btn-prev),
:deep(.el-pagination .btn-next),
:deep(.el-pagination .el-pager li) {
  border-radius: var(--hope-radius-md);
  border: 1px solid var(--hope-border);
  background: var(--hope-surface);
  color: var(--hope-text-secondary);
  font-weight: 600;
}

:deep(.el-pagination .el-pager li.active) {
  background: var(--hope-primary);
  border-color: var(--hope-primary);
  color: #fff;
}

:deep(.el-pagination .el-pager li:hover) {
  color: var(--hope-primary);
}

/* ─── Side Panel ─── */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(26,26,46,0.4);
  backdrop-filter: blur(4px);
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
  background: var(--hope-surface);
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(var(--hope-primary-rgb), 0.15);
  display: flex;
  flex-direction: column;
  animation: slideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

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

.panel-header__title {
  font-size: 16px;
  font-weight: 700;
  color: var(--hope-text);
}

.panel-close {
  width: 32px;
  height: 32px;
  border-radius: var(--hope-radius-md);
  border: 1px solid var(--hope-border);
  background: var(--hope-surface-light);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  color: var(--hope-text-muted);
}
.panel-close:hover {
  background: var(--hope-primary-light);
  border-color: var(--hope-primary);
  color: var(--hope-primary);
}

.panel-body {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 22px;
}

/* Panel Profile */
.panel-profile {
  display: flex;
  align-items: center;
  gap: 16px;
}

.panel-profile__info {
  flex: 1;
  min-width: 0;
}

.panel-profile__name {
  font-size: 18px;
  font-weight: 700;
  color: var(--hope-text);
}

.panel-profile__role {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin-top: 2px;
  font-weight: 500;
}

.panel-profile__badge {
  flex-shrink: 0;
}

/* Panel Sections */
.panel-section {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.panel-section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 2px solid var(--hope-border);
}

.panel-rows {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--hope-border);
  font-size: 13px;
  gap: 12px;
}

.panel-row:last-child {
  border-bottom: none;
}

.panel-row-label {
  color: var(--hope-text-muted);
  font-weight: 500;
  flex-shrink: 0;
}

.panel-row-value {
  font-weight: 600;
  color: var(--hope-text);
  text-align: right;
  word-break: break-all;
}

.panel-row-value--highlight {
  color: var(--hope-primary);
  font-weight: 700;
}

/* Elderly Links */
.elderly-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--hope-radius-md);
  cursor: pointer;
  transition: background 0.15s;
}
.elderly-link:hover {
  background: var(--hope-primary-light);
}

.elderly-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text);
  flex: 1;
}

.elderly-relation {
  font-size: 12px;
  color: var(--hope-text-muted);
  background: var(--hope-bg);
  padding: 2px 8px;
  border-radius: var(--hope-radius-pill);
  font-weight: 500;
}

.elderly-devices {
  font-size: 12px;
  color: var(--hope-text-muted);
}

.empty-text {
  font-size: 13px;
  color: var(--hope-text-muted);
  padding: 8px 0;
  font-weight: 500;
}

/* Panel Actions */
.panel-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* ─── Dialog ─── */
:deep(.hope-dialog .el-dialog) {
  border-radius: var(--hope-radius-xl) !important;
  border: 1px solid var(--hope-border) !important;
  box-shadow: var(--hope-shadow-lg) !important;
}

:deep(.hope-dialog .el-dialog__header) {
  padding: 20px 24px 16px !important;
  border-bottom: 1px solid var(--hope-border) !important;
  margin-right: 0 !important;
}

:deep(.hope-dialog .el-dialog__title) {
  font-size: 16px !important;
  font-weight: 700 !important;
  color: var(--hope-text) !important;
}

:deep(.hope-dialog .el-dialog__body) {
  padding: 20px 24px !important;
}

:deep(.hope-dialog .el-dialog__footer) {
  padding: 16px 24px 20px !important;
  border-top: 1px solid var(--hope-border) !important;
}

:deep(.hope-dialog .el-form-item__label) {
  font-weight: 600 !important;
  color: var(--hope-text-secondary) !important;
}

:deep(.hope-dialog .el-input__wrapper),
:deep(.hope-dialog .el-select .el-input__wrapper) {
  border-radius: var(--hope-radius-md) !important;
  box-shadow: var(--hope-shadow-sm) !important;
  border: 1px solid var(--hope-border) !important;
}

/* ─── Responsive ─── */
@media (max-width: 1200px) {
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .filter-row {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-item, .filter-item--search {
    min-width: 100%;
  }

  .user-grid {
    grid-template-columns: 1fr;
  }

  .pagination-wrapper {
    justify-content: center;
  }

  .panel-actions {
    flex-direction: column;
  }

  .panel-actions .hope-btn {
    width: 100%;
  }
}
</style>
