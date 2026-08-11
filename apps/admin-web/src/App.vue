<template>
  <div v-if="authStore.user" style="height: 100vh;">
    <!-- Full app layout (only shown when logged in) -->
    <el-container style="height: 100%;">
      <!-- Sidebar -->
      <el-aside width="260px" :class="['sidebar', { collapsed: isCollapsed }]" ref="sidebarRef">
        <div class="sidebar-logo">
          <span class="logo-brand">Eregen</span>
          <span class="logo-cn">颐贞</span>
        </div>
        <el-menu :default-active="activeMenu" background-color="transparent" text-color="var(--text-sidebar)" active-text-color="var(--color-primary-dark)" router class="sidebar-menu">
          <div class="nav-section-title">概览</div>
          <el-menu-item index="/dashboard">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12h4l3-9 4 18 3-9h4"/></svg></el-icon><span>健康仪表盘</span>
          </el-menu-item>
          <div class="nav-section-title">设备管理</div>
          <el-menu-item index="/devices">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/></svg></el-icon><span>手环设备</span>
          </el-menu-item>
          <el-menu-item index="/pillboxes">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="6" width="18" height="12" rx="2"/><path d="M8 6V4M16 6V4M3 10h18"/><circle cx="8" cy="14" r="1.5"/><circle cx="16" cy="14" r="1.5"/></svg></el-icon><span>药盒设备</span>
          </el-menu-item>
          <el-menu-item index="/ota">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12m0 0l-4-4m4 4l4-4M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2"/></svg></el-icon><span>固件OTA</span>
          </el-menu-item>
          <div class="nav-section-title">用药管理</div>
          <el-menu-item index="/medication">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 3h6v6h-6zM9 15h6v6h-6zM3 9h6v6H3zM15 9h6v6h-6z"/></svg></el-icon><span>用药规则</span>
          </el-menu-item>
          <div class="nav-section-title">用户管理</div>
          <el-menu-item index="/users">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H6a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon><span>家属用户</span>
          </el-menu-item>
          <el-menu-item index="/elderly">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg></el-icon><span>老人档案</span>
          </el-menu-item>
          <el-menu-item index="/institutions">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21h18M3 7v14M21 7v14M6 11h4M6 15h4M14 11h4M14 15h4M9 3h6v4H9z"/></svg></el-icon><span>机构管理</span>
          </el-menu-item>
          <div class="nav-section-title">医疗管理</div>
          <el-menu-item index="/medical">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4.8 2.3A.3.3 0 105 2H4a2 2 0 00-2 2v5a6 6 0 006 6 6 6 0 006-6V4a2 2 0 00-2-2h-1a.2.2 0 00.3.3"/><path d="M8 15v4M12 15v4M6 23h8"/></svg></el-icon><span>医疗腕带</span>
          </el-menu-item>
          <el-menu-item index="/regulatory">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg></el-icon><span>监管看板</span>
          </el-menu-item>
          <el-menu-item index="/community-wb">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg></el-icon><span>社区老人</span>
          </el-menu-item>
          <div class="nav-section-title">运营</div>
          <el-menu-item index="/alerts">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon><span>告警中心</span>
          </el-menu-item>
          <el-menu-item index="/subscriptions">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg></el-icon><span>订阅管理</span>
          </el-menu-item>
          <el-menu-item index="/analytics">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12V7H5a2 2 0 010-4h14v4M3 5v14a2 2 0 002 2h16v-5M18 14v6"/></svg></el-icon><span>数据分析</span>
          </el-menu-item>
          <div class="nav-section-title">系统</div>
          <el-menu-item index="/settings">
            <el-icon><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 01-2.83 2.83l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg></el-icon><span>系统设置</span>
          </el-menu-item>
        </el-menu>
        <div class="sidebar-footer">
          <el-avatar size="small" style="background: linear-gradient(135deg, #4A7C5F, #6FAF8F);">管</el-avatar>
          <div>
            <div class="footer-name">{{ authStore.user.name }}</div>
            <div class="footer-role">{{ authStore.user.role === 'super_admin' ? '超级管理员' : '管理员' }}</div>
          </div>
        </div>
        <!-- Collapse button -->
        <button class="sidebar-collapse-btn" @click="isCollapsed = !isCollapsed" :title="isCollapsed ? '展开' : '收起'">
          <el-icon :size="14"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg></el-icon>
        </button>
      </el-aside>

      <el-container>
        <!-- Top bar -->
        <el-header class="topbar">
          <div class="breadcrumb">
            <span class="breadcrumb-dot"></span>
            {{ currentBreadcrumb }}
          </div>
          <div class="topbar-right">
            <div class="topbar-icon" title="搜索">
              <el-icon :size="18"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></el-icon>
            </div>
            <el-badge :value="3" :max="99">
              <div class="topbar-icon" title="通知">
                <el-icon :size="18"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
              </div>
            </el-badge>
            <!-- Theme toggle -->
            <button class="theme-toggle-btn" :title="isDark ? '切换至浅色模式' : '切换至深色模式'" @click="toggleTheme">
              <el-icon :size="16">
                <svg v-if="!isDark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/></svg>
                <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
              </el-icon>
            </button>
            <el-button type="danger" @click="handleLogout" plain size="small" class="logout-btn">退出</el-button>
          </div>
        </el-header>

        <!-- Main content -->
        <el-main class="main-content">
          <router-view v-slot="{ Component }">
            <transition name="page-fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
  <div v-else class="login-page-wrapper">
    <LoginView />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import LoginView from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const { isDark, toggle: toggleTheme } = useTheme()

const isCollapsed = ref(false)
const sidebarRef = ref<HTMLElement | null>(null)

const activeMenu = computed(() => route.path)

const currentBreadcrumb = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '首页 / 健康仪表盘',
    '/devices': '设备管理 / 手环设备',
    '/pillboxes': '设备管理 / 药盒设备',
    '/ota': '设备管理 / 固件OTA',
    '/medication': '用药管理 / 用药规则',
    '/subscriptions': '运营管理 / 订阅管理',
    '/users': '用户管理 / 全部用户',
    '/institutions': '用户管理 / 机构管理',
    '/elderly': '用户管理 / 老人档案',
    '/alerts': '告警中心 / 告警列表',
    '/analytics': '数据分析 / 概览',
    '/settings': '系统设置 / 配置',
    '/medical': '医疗管理 / 医疗腕带',
    '/regulatory': '医疗管理 / 监管看板',
    '/community-wb': '医疗管理 / 社区老人',
  }
  return map[route.path] || route.path
})

async function handleLogout() {
  await authStore.logout()
  ElMessage.info('已安全退出')
}

onMounted(async () => {
  if (!authStore.checkLoggedIn()) {
    if (route.path !== '/login') {
      router.push({ path: '/login', query: { redirect: route.path } })
    }
  }
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Noto+Sans+SC:wght@400;500;600;700&display=swap');

/* Global base — handled by admin-theme.scss */

/* ==================== SIDEBAR ==================== */
.sidebar {
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  position: relative;
  box-shadow: var(--shadow-sidebar);
  z-index: var(--z-sidebar);
  width: 260px;
  transition: width 350ms cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  overflow: hidden;
}

.sidebar.collapsed {
  width: 68px;
}

.sidebar::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--color-primary-gradient);
  opacity: 0;
  transition: opacity 250ms ease;
  border-radius: 0 2px 2px 0;
}

.sidebar:hover::before {
  opacity: 1;
}

.sidebar.collapsed::before {
  opacity: 0;
}

.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--sidebar-border);
  padding: 0 20px;
  flex-shrink: 0;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar.collapsed .sidebar-logo {
  justify-content: center;
  padding: 0 10px;
}

.logo-brand {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--color-primary-dark);
}

.logo-cn {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  margin-left: 6px;
  letter-spacing: 0.04em;
}

.sidebar.collapsed .logo-cn { display: none; }

.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
}

.sidebar-menu :deep(.el-menu) {
  border: none !important;
  background: transparent !important;
}

.sidebar-menu :deep(.el-menu-item) {
  height: 40px !important;
  line-height: 40px !important;
  margin: 2px 8px !important;
  border-radius: var(--radius-md) !important;
  font-size: 13px !important;
  color: var(--text-sidebar) !important;
  transition: all 250ms ease !important;
  position: relative;
  padding-left: 16px !important;
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 10px;
}

.sidebar.collapsed .sidebar-menu :deep(.el-menu-item) {
  justify-content: center;
  padding: 0 !important;
  margin: 2px 6px !important;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: var(--sidebar-hover) !important;
  color: var(--color-primary-dark) !important;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: var(--sidebar-active) !important;
  color: var(--color-primary-dark) !important;
  font-weight: 600;
}

.sidebar-menu :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--color-primary);
  border-radius: 0 2px 2px 0;
}

.nav-section-title {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  padding: 16px 20px 6px;
  margin: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.sidebar.collapsed .nav-section-title {
  text-align: center;
  padding: 16px 8px 6px;
  font-size: 0;
}

.sidebar.collapsed .nav-section-title::after {
  content: '· · ·';
  font-size: 10px;
  letter-spacing: 4px;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--sidebar-border);
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-sidebar-footer);
  flex-shrink: 0;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar.collapsed .sidebar-footer {
  justify-content: center;
  padding: 12px 8px;
}

.sidebar-footer :deep(.el-avatar) {
  width: 32px !important;
  height: 32px !important;
  border-radius: var(--radius-sm) !important;
  background: var(--color-primary-gradient) !important;
  font-size: 13px !important;
  font-weight: 700;
  flex-shrink: 0;
}

.sidebar.collapsed .sidebar-footer :deep(.el-avatar) {
  margin: 0 auto;
}

.footer-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer-role {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 1px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar.collapsed .footer-name,
.sidebar.collapsed .footer-role {
  display: none;
}

/* Sidebar collapse button */
.sidebar-collapse-btn {
  position: absolute;
  right: -14px;
  top: 72px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--bg-surface);
  border: 1px solid var(--border-light);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  transition: all 150ms ease;
  color: var(--text-muted);
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.sidebar-collapse-btn:hover {
  background: var(--color-primary-lighter);
  border-color: var(--color-primary);
  color: var(--color-primary);
  transform: scale(1.1);
}

.sidebar.collapsed .sidebar-collapse-btn {
  right: -14px;
  transform: rotate(180deg);
}

.sidebar.collapsed .sidebar-collapse-btn:hover {
  transform: rotate(180deg) scale(1.1);
}

/* ==================== TOPBAR ==================== */
.topbar {
  height: 64px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: var(--z-topbar);
  box-shadow: var(--shadow-topbar);
  transition: background 250ms ease, border-color 250ms ease;
}
.breadcrumb {
  font-size: 13px;
  color: var(--text-tertiary);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}
.breadcrumb-dot {
  width: 6px;
  height: 6px;
  min-width: 6px;
  border-radius: 50%;
  background: var(--color-primary);
  box-shadow: 0 0 6px rgba(74,124,95,0.35);
  animation: eregen-pulse 2.5s ease-in-out infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.15); }
}
.breadcrumb span:last-child {
  color: var(--text-primary);
  font-weight: 600;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.topbar-icon {
  cursor: pointer;
  color: var(--text-tertiary);
  transition: all 150ms ease;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-light);
  background: var(--bg-surface);
}
.topbar-icon:hover {
  color: var(--color-primary);
  background: var(--color-primary-lighter);
  border-color: var(--color-primary-light);
}
.logout-btn {
  margin-left: 8px;
  border-radius: var(--radius-md) !important;
  font-size: 13px;
  padding: 6px 14px !important;
  border: 1px solid var(--border-base) !important;
  color: var(--text-secondary) !important;
  background: var(--bg-surface) !important;
}

.el-badge__content {
  background: var(--color-danger) !important;
  font-size: 10px !important;
  padding: 1px 5px !important;
  min-width: 0 !important;
  height: auto !important;
  border-radius: 10px !important;
  border: 2px solid var(--bg-surface);
  box-shadow: 0 1px 3px rgba(192,74,66,0.3);
}

/* ==================== MAIN CONTENT ==================== */
.main-content {
  background: var(--bg-page);
  background-image: var(--bg-page-gradient);
  padding: 24px 28px;
  overflow-y: auto;
  height: calc(100vh - 64px);
  transition: background 250ms ease;
}

/* ==================== LOGIN PAGE WRAPPER ==================== */
.login-page-wrapper {
  min-height: 100vh;
}

/* ==================== RESPONSIVE ==================== */
@media (max-width: 768px) {
  .sidebar { width: 68px !important; }
  .sidebar-logo .logo-brand,
  .sidebar-logo .logo-cn,
  .sidebar .el-menu-item span,
  .nav-section-title { display: none; }
  .sidebar-menu :deep(.el-menu-item) {
    justify-content: center !important;
    padding: 0 !important;
    margin: 2px 4px !important;
  }
  .sidebar-footer { justify-content: center; }
  .main-content { padding: 16px; }
  .topbar { padding: 0 16px; }
}
</style>
