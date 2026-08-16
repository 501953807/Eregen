<template>
  <aside class="hope-sidebar" :class="{ 'collapsed': collapsed }">
    <!-- Logo -->
    <div class="sidebar-logo">
      <div class="logo-icon-wrap">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
        </svg>
      </div>
      <span class="logo-text">Eregen</span>
    </div>

    <!-- Menu -->
    <nav class="sidebar-nav">
      <ul class="nav-list">
        <li v-for="item in menuItems" :key="item.path">
          <router-link :to="item.path" class="nav-link" :class="{ 'active': isActive(item.path) }">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 12h4l3-9 4 18 3-9h4"/>
            </svg>
            <span class="nav-text">{{ item.label }}</span>
          </router-link>
        </li>
      </ul>
    </nav>

    <!-- User Info -->
    <div class="sidebar-footer">
      <div class="user-avatar">
        <span>{{ authStore.user?.name?.charAt(0) || '管' }}</span>
      </div>
      <div class="user-info">
        <div class="user-name">{{ authStore.user?.name }}</div>
        <div class="user-role">{{ authStore.user?.role === 'super_admin' ? '超级管理员' : '管理员' }}</div>
      </div>
    </div>

    <!-- Collapse Button -->
    <button class="collapse-btn" @click="$emit('toggle')" :title="collapsed ? '展开' : '收起'">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline :class="['collapse-icon', { 'rotated': collapsed }]" points="15 18 9 12 15 6"/>
      </svg>
    </button>
  </aside>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  (e: 'toggle'): void
}>()

const route = useRoute()
const authStore = useAuthStore()

function isActive(path: string) {
  return route.path === path || route.path.startsWith(path + '/')
}

const menuItems = [
  { path: '/dashboard', label: '健康仪表盘' },
  { path: '/devices', label: '手环设备' },
  { path: '/pillboxes', label: '药盒设备' },
  { path: '/ota', label: '固件OTA' },
  { path: '/medication', label: '用药规则' },
  { path: '/users', label: '家属用户' },
  { path: '/elderly', label: '老人档案' },
  { path: '/institutions', label: '机构管理' },
  { path: '/medical', label: '医疗腕带' },
  { path: '/regulatory', label: '监管看板' },
  { path: '/community-wb', label: '社区老人' },
  { path: '/alerts', label: '告警中心' },
  { path: '/subscriptions', label: '订阅管理' },
  { path: '/analytics', label: '数据分析' },
  { path: '/settings', label: '系统设置' },
]
</script>

<style scoped>
.hope-sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 16.2rem;
  background: var(--hope-surface);
  border-right: 1px solid var(--hope-border);
  display: flex;
  flex-direction: column;
  transition: width 0.4s ease-in-out;
  z-index: 100;
  box-shadow: 2px 0 16px rgba(17,38,146,0.06);
}

.hope-sidebar.collapsed {
  width: 4.8rem;
}

/* Logo */
.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 1.25rem;
  border-bottom: 1px solid var(--hope-border);
}

.logo-icon-wrap {
  width: 32px;
  height: 32px;
  background: var(--hope-primary-gradient);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(58,87,232,0.25);
}

.logo-icon-wrap svg {
  width: 18px;
  height: 18px;
}

.logo-text {
  margin-left: 0.75rem;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--hope-text);
  white-space: nowrap;
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.hope-sidebar.collapsed .logo-text {
  opacity: 0;
  transform: translateX(-100%);
  width: 0;
}

/* Navigation */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 0.75rem 0;
}

.nav-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.nav-link {
  display: flex;
  align-items: center;
  padding: 0.625rem 1rem;
  margin: 0.125rem 0.5rem;
  border-radius: var(--hope-radius-md);
  color: var(--hope-text-secondary);
  text-decoration: none;
  transition: background-color 0.2s ease, color 0.2s ease;
  white-space: nowrap;
  overflow: hidden;
  font-weight: 500;
}

.nav-link:hover {
  background: var(--hope-primary-lighter);
  color: var(--hope-primary);
}

.nav-link.active {
  background: var(--hope-primary-gradient);
  color: var(--hope-text-inverse);
  box-shadow: var(--hope-shadow-active);
}

.nav-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.nav-text {
  margin-left: 0.75rem;
  font-size: 0.875rem;
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.hope-sidebar.collapsed .nav-text {
  opacity: 0;
  transform: translateX(-100%);
  width: 0;
}

/* Footer */
.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--hope-surface-light);
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--hope-primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.875rem;
  font-weight: 600;
  flex-shrink: 0;
}

.user-info {
  overflow: hidden;
}

.user-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--hope-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 0.6875rem;
  color: var(--hope-text-muted);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.hope-sidebar.collapsed .user-info {
  display: none;
}

/* Collapse Button */
.collapse-btn {
  position: absolute;
  right: -12px;
  top: 16px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: var(--hope-primary);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 8px rgba(58,87,232,0.3);
  transition: transform 0.4s ease-in-out;
  z-index: 10;
}

.collapse-btn:hover {
  transform: scale(1.1);
}

.collapse-icon {
  width: 14px;
  height: 14px;
  transition: transform 0.4s ease-in-out;
}

.collapse-icon.rotated {
  transform: rotate(180deg);
}

.hope-sidebar.collapsed .collapse-btn {
  right: -12px;
}

/* Responsive */
@media (max-width: 768px) {
  .hope-sidebar {
    transform: translateX(-100%);
  }

  .hope-sidebar.mobile-open {
    transform: translateX(0);
  }
}
</style>
