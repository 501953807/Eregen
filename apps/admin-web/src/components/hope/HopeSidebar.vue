<template>
  <aside class="hope-sidebar" :class="{ 'collapsed': collapsed }">
    <!-- Logo -->
    <div class="sidebar-header">
      <router-link to="/dashboard" class="sidebar-brand">
        <div class="logo-main">
          <svg class="logo-icon" viewBox="0 0 30 30" fill="none">
            <rect x="-0.757" y="19.243" width="28" height="4" rx="2" transform="rotate(-45 -0.757 19.243)" fill="currentColor"/>
            <rect x="7.728" y="27.728" width="28" height="4" rx="2" transform="rotate(-45 7.728 27.728)" fill="currentColor"/>
            <rect x="10.537" y="16.395" width="16" height="4" rx="2" transform="rotate(45 10.537 16.395)" fill="currentColor"/>
            <rect x="10.556" y="-0.556" width="28" height="4" rx="2" transform="rotate(45 10.556 -0.556)" fill="currentColor"/>
          </svg>
        </div>
        <span class="logo-title" :class="{ 'hidden': collapsed }">Eregen</span>
      </router-link>
      <button class="sidebar-toggle" @click="$emit('toggle')" :title="collapsed ? '展开' : '收起'">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6"/>
        </svg>
      </button>
    </div>

    <!-- Navigation -->
    <nav class="sidebar-body">
      <ul class="nav-menu">
        <template v-for="category in menuConfig" :key="category.key">
          <!-- Category label (static) -->
          <li class="nav-item static-item">
            <span class="nav-link disabled">
              <span class="default-icon">{{ category.label }}</span>
            </span>
          </li>

          <!-- Group items -->
          <template v-for="group in category.groups" :key="group.key">
            <li class="nav-item" :class="{ 'has-sub': group.items.length > 1 }">
              <a
                v-if="group.items.length > 1"
                class="nav-link"
                :class="{ 'active': isGroupActive(group) }"
                :aria-expanded="openGroups[group.key] ? 'true' : 'false'"
                @click="toggleGroup(group.key)"
              >
                <i class="icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </i>
                <span class="item-name">{{ group.label }}</span>
                <i class="right-icon">
                  <svg class="icon-18" xmlns="http://www.w3.org/2000/svg" width="18" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                  </svg>
                </i>
              </a>
              <router-link
                v-else
                :to="group.items[0].path"
                class="nav-link"
                :class="{ 'active': isActive(group.items[0].path) }"
              >
                <i class="icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </i>
                <span class="item-name">{{ group.items[0].label }}</span>
              </router-link>

              <!-- Submenu -->
              <ul v-if="group.items.length > 1" class="sub-nav" :class="{ 'show': openGroups[group.key] }">
                <li v-for="item in group.items" :key="item.path">
                  <router-link :to="item.path" class="nav-link" :class="{ 'active': isActive(item.path) }">
                    <i class="icon">
                      <svg class="icon-10" xmlns="http://www.w3.org/2000/svg" width="10" viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="12" r="8"/>
                      </svg>
                    </i>
                    <span class="sidenav-mini-icon">{{ item.label.charAt(0) }}</span>
                    <span class="item-name">{{ item.label }}</span>
                    <span v-if="item.badge" class="menu-badge" :class="item.badgeClass || 'badge--red'">{{ item.badge }}</span>
                  </router-link>
                </li>
              </ul>
            </li>
          </template>
        </template>
      </ul>
    </nav>

    <!-- User Footer -->
    <div class="sidebar-footer">
      <div class="user-avatar">
        <span>{{ authStore.user?.name?.charAt(0) || '管' }}</span>
      </div>
      <div class="user-info" :class="{ 'hidden': collapsed }">
        <div class="user-name">{{ authStore.user?.name }}</div>
        <div class="user-role">{{ authStore.user?.role === 'super_admin' ? '超级管理员' : '管理员' }}</div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

defineProps<{ collapsed: boolean }>()
defineEmits<{ (e: 'toggle'): void }>()

const route = useRoute()
const authStore = useAuthStore()

interface MenuItem {
  path: string
  label: string
  badge?: string
  badgeClass?: string
}

interface MenuGroup {
  key: string
  label: string
  items: MenuItem[]
}

interface MenuCategory {
  key: string
  label: string
  groups: MenuGroup[]
}

// Track which groups are expanded
const openGroups = ref<Record<string, boolean>>({
  overview: true,
  device: false,
  medication: false,
  user: false,
  medical: false,
  operation: false,
  system: false,
})

function toggleGroup(key: string) {
  openGroups.value[key] = !openGroups.value[key]
}

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}

function isGroupActive(group: MenuGroup): boolean {
  return group.items.some(item => isActive(item.path))
}

// Menu configuration — three levels: category → group → page
const menuConfig: MenuCategory[] = [
  {
    key: 'overview',
    label: '概览',
    groups: [
      { key: 'overview', label: '仪表盘', items: [{ path: '/dashboard', label: '健康仪表盘' }] },
    ],
  },
  {
    key: 'device',
    label: '设备管理',
    groups: [
      {
        key: 'device',
        label: '设备列表',
        items: [
          { path: '/devices', label: '手环设备' },
          { path: '/pillboxes', label: '药盒设备' },
          { path: '/ota', label: '固件OTA' },
        ],
      },
    ],
  },
  {
    key: 'medication',
    label: '用药管理',
    groups: [
      { key: 'medication', label: '用药规则', items: [{ path: '/medication', label: '用药管理' }] },
    ],
  },
  {
    key: 'user',
    label: '用户管理',
    groups: [
      {
        key: 'user',
        label: '用户列表',
        items: [
          { path: '/users', label: '家属用户' },
          { path: '/elderly', label: '老人档案' },
          { path: '/institutions', label: '机构管理' },
        ],
      },
    ],
  },
  {
    key: 'medical',
    label: '医疗管理',
    groups: [
      {
        key: 'medical',
        label: '医疗应用',
        items: [
          { path: '/medical', label: '医疗腕带' },
          { path: '/regulatory', label: '监管看板' },
          { path: '/community-wb', label: '社区老人' },
        ],
      },
    ],
  },
  {
    key: 'operation',
    label: '运营',
    groups: [
      {
        key: 'operation',
        label: '运营管理',
        items: [
          { path: '/alerts', label: '告警中心', badge: '3', badgeClass: 'badge--red' },
          { path: '/subscriptions', label: '订阅管理' },
          { path: '/analytics', label: '数据分析' },
        ],
      },
    ],
  },
  {
    key: 'system',
    label: '系统',
    groups: [
      { key: 'system', label: '系统设置', items: [{ path: '/settings', label: '系统设置' }] },
    ],
  },
]
</script>

<style scoped>
/* ─── Hope UI 侧边栏核心样式 ─── */
.hope-sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 16.2rem;
  background: #ffffff;
  border-right: 1px solid rgba(138, 146, 166, 0.20);
  display: flex;
  flex-direction: column;
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 200;
  overflow: hidden;
}

.hope-sidebar.collapsed {
  width: 4.8rem;
}

/* ─── Header: Logo + Toggle ─── */
.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 0.75rem;
  border-bottom: 1px solid rgba(138, 146, 166, 0.15);
  flex-shrink: 0;
  position: relative;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  text-decoration: none;
  flex: 1;
  min-width: 0;
  gap: 0.6rem;
}

.logo-main {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
}

.logo-icon {
  width: 30px;
  height: 30px;
  color: var(--hope-primary);
}

.logo-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--hope-text);
  white-space: nowrap;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.logo-title.hidden {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

/* ─── Toggle Button: 右侧半圆形 ─── */
.sidebar-toggle {
  position: absolute;
  right: -10px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ffffff;
  border: 1px solid rgba(138, 146, 166, 0.25);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--hope-text-muted);
  box-shadow: 0 2px 6px rgba(17,38,146,0.12);
  transition: all 0.2s ease;
  z-index: 10;
  padding: 0;
}

.sidebar-toggle:hover {
  background: var(--hope-primary);
  color: #ffffff;
  border-color: var(--hope-primary);
  transform: translateY(-50%) scale(1.1);
}

.sidebar-toggle svg {
  width: 10px;
  height: 10px;
  transition: transform 0.3s ease;
}

.hope-sidebar.collapsed .sidebar-toggle svg {
  transform: rotate(180deg);
}

/* ─── Body: Scrollable nav ─── */
.sidebar-body {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.5rem 0;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

.sidebar-body:hover {
  scrollbar-color: rgba(138, 146, 166, 0.3) transparent;
}

/* ─── Nav Menu ─── */
.nav-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.nav-item {
  position: relative;
  margin-top: 2px;
  color: #6c757d;
}

.nav-item.static-item {
  padding: 0.5rem 1rem 0.25rem;
  margin-top: 0.35rem;
}

/* ─── Nav Link: Hope UI 圆角导航样式 ─── */
.nav-link {
  display: flex;
  align-items: center;
  padding: 0.625rem 1rem;
  border-radius: 0.5rem;
  color: #6c757d;
  text-decoration: none;
  transition: all 0.2s ease;
  white-space: nowrap;
  overflow: hidden;
  font-weight: 400;
  font-size: 0.9375rem;
  position: relative;
  cursor: pointer;
  line-height: 1.4;
}

.nav-link:hover {
  color: #2c4aba;
  background: #f0f4ff;
}

.nav-link.disabled {
  color: var(--hope-text-muted);
  cursor: default;
  padding: 0.25rem 1rem;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.nav-link.active {
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%);
  color: #ffffff;
  box-shadow: 0 6px 16px rgba(58, 87, 232, 0.32), 0 2px 6px rgba(58, 87, 232, 0.18);
  font-weight: 500;
}

.nav-link.active:hover {
  color: #ffffff;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%);
}

/* ─── Icons ─── */
.nav-link .icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.72;
}

.nav-link.active .icon,
.nav-link:hover .icon {
  opacity: 1;
}

.nav-link .icon svg {
  width: 100%;
  height: 100%;
}

.nav-link .icon-10 svg {
  width: 10px;
  height: 10px;
}

.nav-link .icon-18 svg {
  width: 18px;
  height: 18px;
}

/* ─── Item name ─── */
.item-name {
  flex: 1;
  margin-left: 0.625rem;
  font-size: 0.9375rem;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.hope-sidebar.collapsed .item-name {
  opacity: 0;
  width: 0;
  margin-left: 0;
  overflow: hidden;
}

/* ─── Mini icon (collapsed state) ─── */
.sidenav-mini-icon {
  display: none;
  font-size: 0.8125rem;
  font-weight: 600;
  color: inherit;
}

.hope-sidebar.collapsed .sidenav-mini-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

/* ─── Right chevron ─── */
.right-icon {
  display: flex;
  transition: transform 0.3s ease;
  color: currentColor;
  opacity: 0.55;
}

.nav-link[aria-expanded="true"] .right-icon {
  transform: rotate(90deg);
}

/* ─── Submenu ─── */
.sub-nav {
  list-style: none;
  padding: 0.125rem 0;
  margin: 0;
  overflow: hidden;
  max-height: 0;
  transition: max-height 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.sub-nav.show {
  max-height: 500px;
}

.sub-nav .nav-link {
  padding: 0.4375rem 1rem 0.4375rem 2.5rem;
  font-size: 0.875rem;
}

.sub-nav .nav-link .icon {
  width: 10px;
  height: 10px;
  opacity: 0.5;
}

.sub-nav .nav-link.active .icon {
  opacity: 1;
}

.hope-sidebar.collapsed .sub-nav {
  max-height: 0 !important;
}

/* ─── Badge ─── */
.menu-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  border-radius: 50%;
  font-size: 0.625rem;
  font-weight: 700;
  margin-left: auto;
  padding: 0 4px;
  flex-shrink: 0;
}

.badge--red {
  background: #dc3545;
  color: white;
}

/* ─── Footer: User card ─── */
.sidebar-footer {
  padding: 0.875rem 1rem;
  border-top: 1px solid rgba(138, 146, 166, 0.15);
  display: flex;
  align-items: center;
  gap: 0.625rem;
  background: #fafbfd;
  flex-shrink: 0;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.875rem;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 3px 8px rgba(58, 87, 232, 0.25);
}

.user-info {
  overflow: hidden;
  transition: opacity 0.2s ease;
  flex: 1;
  min-width: 0;
}

.user-info.hidden {
  opacity: 0;
  width: 0;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--hope-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 0.6875rem;
  color: var(--hope-text-muted);
  margin-top: 1px;
  white-space: nowrap;
}

/* ─── Responsive ─── */
@media (max-width: 768px) {
  .hope-sidebar {
    transform: translateX(-100%);
  }
  .hope-sidebar.mobile-open {
    transform: translateX(0);
  }
}
</style>
