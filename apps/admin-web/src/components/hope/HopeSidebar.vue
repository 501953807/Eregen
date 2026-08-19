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
    </div>

    <!-- Navigation -->
    <nav class="sidebar-body">
      <ul class="nav-menu">
        <template v-for="category in menuConfig" :key="category.key">
          <!-- Category label: icon only when collapsed -->
          <li class="nav-item static-item" :class="{ 'collapsed': collapsed }">
            <span class="nav-link disabled">
              <i class="icon" v-html="getIcon(category.key, '')"></i>
              <span class="item-name" :class="{ 'hidden': collapsed }">{{ category.label }}</span>
            </span>
          </li>

          <!-- Group items -->
          <template v-for="group in category.groups" :key="group.key">
            <li
              class="nav-item"
              :class="{
                'has-sub': group.items.length > 1,
                'expanded': openGroups[group.key]
              }"
              ref="(el) => groupRefs[group.key] = el as HTMLElement | null"
              @mouseenter="onGroupEnter(group.key, $event.target as HTMLElement)"
              @mouseleave="onGroupLeave(group.key)"
            >
              <!-- Multi-item group: clickable toggle (expanded only), hover in collapsed mode -->
              <a
                v-if="group.items.length > 1"
                class="nav-link"
                :class="{ 'active': isGroupActive(group) }"
                @click.stop="!collapsed && toggleGroup(group.key)"
                :title="collapsed ? group.label : ''"
              >
                <i class="icon" v-html="getIcon(category.key, group.key)"></i>
                <span class="item-name" :class="{ 'hidden': collapsed }">{{ group.label }}</span>
                <i class="right-icon" :class="{ 'rotated': openGroups[group.key] }">
                  <svg class="icon-18" xmlns="http://www.w3.org/2000/svg" width="18" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                  </svg>
                </i>
              </a>
              <!-- Single-item group: direct router-link -->
              <router-link
                v-else
                :to="group.items[0].path"
                class="nav-link"
                :class="{ 'active': isActive(group.items[0].path) }"
                :title="collapsed ? group.items[0].label : ''"
              >
                <i class="icon" v-html="getIcon(category.key, group.key, group.items[0].path)"></i>
                <span class="item-name" :class="{ 'hidden': collapsed }">{{ group.items[0].label }}</span>
              </router-link>

              <!-- Inline submenu (visible when expanded) -->
              <ul
                v-if="group.items.length > 1 && !collapsed"
                class="sub-nav"
                :class="{ 'show': openGroups[group.key] }"
              >
                <li v-for="item in group.items" :key="item.path">
                  <router-link
                    :to="item.path"
                    class="nav-link"
                    :class="{ 'active': isActive(item.path) }"
                  >
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

    <!-- Floating panel portal — rendered outside overflow:hidden constraints -->
    <Teleport to="body">
      <Transition name="float-panel">
        <div
          v-if="floating && floating.group"
          class="float-panel"
          :style="floatPanelStyle"
          @mouseenter="onGroupEnter(floating.group.key)"
          @mouseleave="onGroupLeave(floating.group.key)"
        >
          <ul class="float-panel-list">
            <li v-for="item in floating.group.items" :key="item.path">
              <router-link
                :to="item.path"
                class="float-item"
                :class="{ 'active': isActive(item.path) }"
                @click="handleSubItemClick(item.path)"
              >
                <i class="float-icon" v-html="getIcon(floating.category.key, floating.group.key, item.path)"></i>
                <span>{{ item.label }}</span>
                <span v-if="item.badge" class="menu-badge" :class="item.badgeClass || 'badge--red'">{{ item.badge }}</span>
              </router-link>
            </li>
          </ul>
        </div>
      </Transition>
    </Teleport>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps<{ collapsed: boolean }>()
const emit = defineEmits<{ (e: 'toggle'): void }>()

const route = useRoute()

// Internal reactive collapsed state
const collapsedInternal = ref(props.collapsed)
watch(() => props.collapsed, (v) => { collapsedInternal.value = v })
const collapsed = computed({
  get: () => collapsedInternal.value,
  set: (v) => { collapsedInternal.value = v; emit('toggle') }
})

// Track which multi-item groups are expanded (click-based)
const openGroups = ref<Record<string, boolean>>({
  overview: true,
  device: false,
  medication: false,
  user: false,
  medical: false,
  chain: false,
  operation: false,
  system: false,
})

function toggleGroup(key: string) {
  openGroups.value[key] = !openGroups.value[key]
}

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

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}

function isGroupActive(group: MenuGroup): boolean {
  return group.items.some(item => isActive(item.path))
}

// ─── Floating panel (collapsed state) ─────────────────────────────────────
const groupRefs = ref<Record<string, HTMLElement | null>>({})
const floating = ref<{ group: MenuGroup; category: MenuCategory } | null>(null)
const floatTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const floatPanelStyle = ref<{ top: string; left: string }>({ top: '0px', left: '0px' })

function onGroupEnter(key: string, el?: HTMLElement | null) {
  if (!collapsed.value) return
  // Cancel any pending close
  if (floatTimer.value) { clearTimeout(floatTimer.value); floatTimer.value = null }

  const group = findGroup(key)
  if (!group) return

  // Auto-open hovered group, close others
  for (const k of Object.keys(openGroups.value)) {
    openGroups.value[k] = (k === key)
  }

  // Position panel
  floating.value = { group, category: findCategory(key)! }

  requestAnimationFrame(() => {
    const refEl = el || groupRefs.value[key]
    if (refEl) {
      const rect = refEl.getBoundingClientRect()
      floatPanelStyle.value = {
        top: rect.top + 'px',
        left: (rect.right + 2) + 'px',
      }
    }
  })
}

function onGroupLeave(key: string) {
  if (!collapsed.value) return
  // Delay close so mouse can travel to panel
  floatTimer.value = setTimeout(() => {
    floating.value = null
    openGroups.value[key] = false
    floatTimer.value = null
  }, 180)
}

function handleSubItemClick(_path: string) {
  if (collapsed.value) {
    collapsed.value = false
  }
}

// ─── Helpers ───────────────────────────────────────────────────────────────
function findGroup(key: string): MenuGroup | null {
  for (const cat of menuConfig) {
    for (const g of cat.groups) {
      if (g.key === key) return g
    }
  }
  return null
}

function findCategory(key: string): MenuCategory | null {
  for (const cat of menuConfig) {
    for (const g of cat.groups) {
      if (g.key === key) return cat
    }
  }
  return null
}

onMounted(() => {
  // Reposition floating panel on scroll
  window.addEventListener('scroll', repositionFloating, true)
})
onUnmounted(() => {
  if (floatTimer.value) clearTimeout(floatTimer.value)
  window.removeEventListener('scroll', repositionFloating, true)
})

function repositionFloating() {
  if (floating.value) {
    const refEl = groupRefs.value[floating.value!.group.key]
    if (refEl) {
      const rect = refEl.getBoundingClientRect()
      floatPanelStyle.value = {
        top: rect.top + 'px',
        left: (rect.right + 2) + 'px',
      }
    }
  }
}

// Icon rendering
const iconMap: Record<string, string> = {
  overview: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>`,
  dashboard_overview: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>`,
  device: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>`,
  bracelet_device: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/></svg>`,
  pillbox_device: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="7" width="18" height="13" rx="2"/><path d="M8 7V5a1 1 0 011-1h6a1 1 0 011 1v2"/><line x1="12" y1="7" x2="12" y2="20"/></svg>`,
  ota_device: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>`,
  medication: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M10.5 1.5H8A6.5 6.5 0 001.5 8v8A6.5 6.5 0 008 22.5h8a6.5 6.5 0 006.5-6.5v-2.5M12 7v5M9.5 9.5h5"/></svg>`,
  user: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>`,
  users_user: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>`,
  elderly_user: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`,
  institutions_user: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 21h18M3 10h18M5 6l7-3 7 3M4 10v11M20 10v11M8 14v3M12 14v3M16 14v3"/></svg>`,
  medical: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>`,
  regulatory_medical: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`,
  community_medical: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>`,
  chain_self: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`,
  chain_hospital: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 21h18M3 7v14M21 7v14M6 11h4M14 11h4M6 15h4M14 15h4"/><path d="M9 7V3h6v4"/></svg>`,
  chain_community: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/></svg>`,
  persons_chain: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>`,
  operation: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/><line x1="2" y1="20" x2="22" y2="20"/></svg>`,
  alerts_operation: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg>`,
  subscriptions_operation: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>`,
  analytics_operation: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/><line x1="2" y1="20" x2="22" y2="20"/></svg>`,
  system: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06-.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/></svg>`,
}

function getIcon(category: string, group: string, path?: string): string {
  if (path) {
    const itemKey = path.replace('/', '') + '_' + category
    if (iconMap[itemKey]) return iconMap[itemKey]
  }
  if (iconMap[group]) return iconMap[group]
  if (iconMap[category]) return iconMap[category]
  return iconMap['overview']
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
    key: 'chain',
    label: '业务链',
    groups: [
      {
        key: 'chain',
        label: '业务链',
        items: [
          { path: '/persons', label: '人员档案' },
          { path: '/self', label: '自营链' },
          { path: '/hospital', label: '医院链' },
          { path: '/community', label: '社区链' },
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
/* ─── Sidebar core ─── */
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

/* ─── Header: Logo ─── */
.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 0.75rem;
  border-bottom: 1px solid rgba(138, 146, 166, 0.15);
  flex-shrink: 0;
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
  transition: opacity 0.2s ease;
}

.logo-title.hidden {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

/* ─── Body: scrollable nav ─── */
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
}

/* ─── Category labels ─── */
.nav-item.static-item {
  margin-top: 0.75rem;
}

.nav-item.static-item.collapsed {
  display: flex;
  justify-content: center;
  padding: 0.375rem 0;
}

.nav-link.disabled {
  display: flex;
  align-items: center;
  padding: 0.125rem 0.75rem;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--hope-text-muted);
  opacity: 0.55;
  cursor: default;
  pointer-events: none;
  white-space: nowrap;
  overflow: hidden;
}

.hope-sidebar.collapsed .nav-link.disabled {
  padding: 0.25rem 0;
  font-size: 0;
  justify-content: center;
  opacity: 0.4;
}

.nav-link.disabled .item-name.hidden {
  display: none;
}

/* ─── Nav Link (regular items) ─── */
.nav-link {
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  margin: 0;
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
  min-width: 0;
  width: 100%;
}

.nav-link:hover {
  color: #2c4aba;
  background: #f0f4ff;
}

.nav-link.active {
  background: var(--hope-primary-gradient);
  color: #ffffff;
  box-shadow: var(--hope-shadow-primary), 0 2px 6px rgba(58, 87, 232, 0.18);
  font-weight: 500;
}

.nav-link.active:hover {
  color: #ffffff;
  background: var(--hope-primary-gradient-hover);
}

/* ─── Collapsed state: center icons ─── */
.hope-sidebar.collapsed .nav-item {
  display: flex;
  justify-content: center;
  padding: 0;
}

.hope-sidebar.collapsed .nav-link {
  justify-content: center;
  padding: 0.5625rem 0;
  width: 100%;
}

.hope-sidebar.collapsed .nav-link .icon {
  margin: 0;
}

.hope-sidebar.collapsed .nav-link .right-icon {
  display: none;
}

.hope-sidebar.collapsed .nav-link .item-name {
  opacity: 0;
  width: 0;
  overflow: hidden;
  display: none;
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
  margin-right: 0.625rem;
}

.nav-link.active .icon,
.nav-link:hover .icon {
  opacity: 1;
}

.nav-link .icon svg {
  width: 100%;
  height: 100%;
}

/* ─── Right chevron ─── */
.right-icon {
  display: flex;
  transition: transform 0.3s ease;
  color: currentColor;
  opacity: 0.55;
  flex-shrink: 0;
}

.right-icon.rotated {
  transform: rotate(90deg);
}

/* ─── Inline Submenu (expanded state) ─── */
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
  padding: 0.4375rem 0.75rem 0.4375rem 2rem;
  font-size: 0.875rem;
}

/* ─── Floating Panel (Teleported to body) ─── */
.float-panel {
  position: fixed;
  top: 0;
  left: 0;
  min-width: 14rem;
  max-width: 20rem;
  background: #ffffff;
  border: 1px solid rgba(138, 146, 166, 0.20);
  border-radius: 0 0.75rem 0.75rem 0;
  box-shadow: 4px 4px 20px rgba(17, 38, 146, 0.12);
  z-index: 9999;
  overflow: hidden;
}

.float-panel-list {
  list-style: none;
  padding: 0.25rem 0;
  margin: 0;
}

.float-item {
  display: flex;
  align-items: center;
  padding: 0.5rem 1rem;
  color: #4a5568;
  text-decoration: none;
  transition: background 0.15s ease, color 0.15s ease;
  font-size: 0.9375rem;
  cursor: pointer;
  white-space: nowrap;
}

.float-item:hover {
  background: #f0f4ff;
  color: #2c4aba;
}

.float-item.active {
  background: var(--hope-primary-gradient);
  color: #ffffff;
}

.float-item.active:hover {
  color: #ffffff;
}

.float-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  opacity: 0.7;
  margin-right: 0.625rem;
}

.float-icon svg {
  width: 100%;
  height: 100%;
}

/* Float panel transition */
.float-panel-enter-active,
.float-panel-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.float-panel-enter-from,
.float-panel-leave-to {
  opacity: 0;
  transform: translateX(-4px);
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
