<template>
  <header class="hope-topbar">
    <div class="topbar-left">
      <!-- Breadcrumb -->
      <nav class="topbar-breadcrumb" aria-label="breadcrumb">
        <ol class="breadcrumb-list">
          <li class="breadcrumb-item">
            <router-link to="/dashboard">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              Home
            </router-link>
          </li>
          <li v-for="(crumb, idx) in breadcrumbs" :key="idx" class="breadcrumb-item">
            <span class="breadcrumb-sep">/</span>
            <router-link v-if="crumb.path && idx < breadcrumbs.length - 1" :to="crumb.path">
              {{ crumb.label }}
            </router-link>
            <span v-else class="breadcrumb-active">{{ crumb.label }}</span>
          </li>
        </ol>
      </nav>
    </div>

    <div class="topbar-right">
      <!-- Search -->
      <div class="search-box">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <path d="M21 21l-4.35-4.35"/>
        </svg>
        <input type="text" placeholder="Search..." class="search-input" />
      </div>

      <!-- Language Toggle -->
      <button class="topbar-btn" title="Language">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
          <circle cx="12" cy="12" r="10"/>
          <line x1="2" y1="12" x2="22" y2="12"/>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
        </svg>
      </button>

      <!-- Notifications -->
      <button class="topbar-btn" title="Notifications">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 01-3.46 0"/>
        </svg>
        <span class="badge badge--dot"></span>
      </button>

      <!-- Messages -->
      <button class="topbar-btn" title="Messages">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>
        </svg>
      </button>

      <!-- Theme Toggle -->
      <button class="topbar-btn" title="Toggle theme" @click="toggleTheme">
        <svg v-if="!isDark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
          <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
          <circle cx="12" cy="12" r="5" stroke-linecap="round" stroke-linejoin="round"/>
          <line x1="12" y1="1" x2="12" y2="3" stroke-linecap="round"/>
          <line x1="12" y1="21" x2="12" y2="23" stroke-linecap="round"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" stroke-linecap="round"/>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" stroke-linecap="round"/>
          <line x1="1" y1="12" x2="3" y2="12" stroke-linecap="round"/>
          <line x1="21" y1="12" x2="23" y2="12" stroke-linecap="round"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" stroke-linecap="round"/>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" stroke-linecap="round"/>
        </svg>
      </button>

      <!-- User Dropdown -->
      <el-dropdown trigger="click" @command="handleCommand" placement="bottom-end">
        <div class="user-menu">
          <div class="user-avatar">
            <span>{{ authStore.user?.name?.charAt(0) || '管' }}</span>
          </div>
          <div class="user-meta">
            <span class="user-name">{{ authStore.user?.name }}</span>
            <span class="user-role">{{ authStore.user?.role === 'super_admin' ? '超级管理员' : '管理员' }}</span>
          </div>
          <svg class="dropdown-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </div>
        <template #dropdown>
          <el-dropdown-menu class="topbar-dropdown">
            <el-dropdown-item command="profile">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2" stroke-linecap="round" stroke-linejoin="round"/>
                <circle cx="12" cy="7" r="4" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              个人资料
            </el-dropdown-item>
            <el-dropdown-item command="settings">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                <circle cx="12" cy="12" r="3" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              系统设置
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { isDark, toggle: toggleTheme } = useTheme()

// Build breadcrumb from route
const breadcrumbs = computed(() => {
  const path = route.path
  if (path === '/dashboard' || path === '/') return []

  const parts = path.split('/').filter(Boolean)
  const map: Record<string, { label: string; path?: string }> = {
    dashboard: { label: '仪表盘' },
    devices: { label: '设备管理', path: '/devices' },
    pillboxes: { label: '药盒设备', path: '/devices' },
    subscriptions: { label: '订阅管理' },
    users: { label: '用户管理', path: '/users' },
    institutions: { label: '机构管理' },
    alerts: { label: '告警中心' },
    analytics: { label: '数据分析' },
    settings: { label: '系统设置' },
    ota: { label: '固件OTA' },
    medication: { label: '用药管理' },
    elderly: { label: '老人档案' },
    persons: { label: '人员档案' },
    self: { label: '自证链' },
    hospital: { label: '医院链' },
    community: { label: '社区链' },
    regulatory: { label: '监管看板' },
    medical: { label: '医疗腕带' },
    audit: { label: '审计详情', path: '/audit' },
    'community-wb': { label: '社区腕带' },
  }

  const result: { label: string; path?: string }[] = []
  let accumulated = ''
  for (const part of parts) {
    accumulated += '/' + part
    const info = map[part] || map[accumulated.replace('/', '')]
    if (info) {
      result.push({ label: info.label, path: accumulated })
    }
  }
  return result
})

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    ElMessage.info('已安全退出')
    router.push('/login')
  } else if (command === 'settings') {
    router.push('/settings')
  } else if (command === 'profile') {
    ElMessage.info('个人资料页面开发中...')
  }
}
</script>

<style scoped>
/* ─── Hope UI 顶栏 ─── */
.hope-topbar {
  height: 64px;
  background: var(--hope-surface);
  border-bottom: 1px solid rgba(138, 146, 166, 0.20);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  position: sticky;
  top: 0;
  z-index: 90;
  box-shadow: 0 4px 14px rgba(17,38,146,0.06);
}

/* Left: Breadcrumb */
.topbar-left { display: flex; align-items: center; min-width: 0; flex: 1; }

.topbar-breadcrumb {
  font-size: 0.8125rem;
  overflow: hidden;
}
.breadcrumb-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0;
  flex-wrap: wrap;
}
.breadcrumb-item {
  display: flex;
  align-items: center;
  white-space: nowrap;
}
.breadcrumb-item a {
  color: var(--hope-text-muted);
  text-decoration: none;
  transition: color 0.15s ease;
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}
.breadcrumb-item a:hover { color: var(--hope-primary); }
.breadcrumb-item a svg { width: 14px; height: 14px; }
.breadcrumb-sep {
  color: var(--hope-text-muted);
  margin: 0 6px;
  font-size: 0.75rem;
}
.breadcrumb-active {
  color: var(--hope-text);
  font-weight: 600;
}

/* Right actions */
.topbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

/* ─── Search Box: Hope UI 风格 ─── */
.search-box { position: relative; display: flex; align-items: center; }
.search-icon {
  position: absolute;
  left: 14px;
  width: 16px;
  height: 16px;
  color: var(--hope-text-muted);
  pointer-events: none;
}
.search-input {
  width: 220px;
  height: 40px;
  padding: 0 14px 0 42px;
  border: 1px solid var(--hope-border-strong);
  border-radius: var(--hope-radius-lg);
  background: var(--hope-surface-light);
  font-size: 0.9375rem;
  color: var(--hope-text);
  transition: all 0.2s ease;
  outline: none;
  font-family: inherit;
}
.search-input:focus {
  border-color: var(--hope-primary);
  box-shadow: 0 0 0 3px rgba(58,87,232,0.12);
  width: 260px;
  background: var(--hope-surface);
}
.search-input::placeholder { color: var(--hope-text-muted); }

/* ─── Topbar Buttons: Hope UI 圆角风格 ─── */
.topbar-btn {
  width: 40px;
  height: 40px;
  border-radius: var(--hope-radius-md);
  border: 1px solid transparent;
  background: transparent;
  color: var(--hope-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  position: relative;
}
.topbar-btn:hover {
  background: rgba(var(--hope-primary-rgb), 0.08);
  color: var(--hope-primary);
  border-color: rgba(var(--hope-primary-rgb), 0.15);
}
.topbar-btn svg { width: 18px; height: 18px; }

/* Badge dot */
.badge {
  position: absolute;
  top: 6px; right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.badge--dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--hope-error);
  box-shadow: 0 0 0 2px var(--hope-surface);
}

/* ─── User Menu: Hope UI 风格 ─── */
.user-menu {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 6px 10px 6px 6px;
  border-radius: var(--hope-radius-lg);
  cursor: pointer;
  transition: background 0.15s ease;
  border: 1px solid var(--hope-border);
}
.user-menu:hover {
  background: rgba(var(--hope-primary-rgb), 0.06);
  border-color: rgba(var(--hope-primary-rgb), 0.18);
}
.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.8125rem;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(58, 87, 232, 0.22);
}
.user-meta { display: flex; flex-direction: column; line-height: 1.2; }
.user-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--hope-text);
  white-space: nowrap;
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.user-role {
  font-size: 0.6875rem;
  color: var(--hope-text-muted);
}
.dropdown-arrow {
  width: 14px;
  height: 14px;
  color: var(--hope-text-muted);
  transition: transform 0.2s ease;
}

/* Responsive */
@media (max-width: 768px) {
  .search-box { display: none; }
  .user-meta { display: none; }
  .dropdown-arrow { display: none; }
}
</style>
