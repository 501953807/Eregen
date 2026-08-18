<template>
  <HopeTheme v-if="authStore.isLoggedIn">
    <div class="hope-app">
      <!-- Sidebar -->
      <HopeSidebar :collapsed="isCollapsed" @toggle="isCollapsed = !isCollapsed" />

      <!-- Main Container -->
      <div class="hope-main" :class="{ 'collapsed': isCollapsed }">
        <!-- Topbar -->
        <HopeTopbar />

        <!-- Hero Banner (Dashboard only) -->
        <HopeHeroBanner v-if="route.path === '/dashboard'" />

        <!-- Content wrapper -->
        <div class="content-wrapper">
          <main class="hope-content" :class="{ 'has-hero': route.path === '/dashboard' }">
            <div class="page-content" :class="{ 'page-content--no-hero': route.path !== '/dashboard' }">
              <router-view v-slot="{ Component }">
                <transition name="page-fade" mode="out-in">
                  <component :is="Component" />
                </transition>
              </router-view>
            </div>
          </main>
        </div>
      </div>
    </div>
  </HopeTheme>
  <!-- Sidebar toggle — placed OUTSIDE .hope-app so it's never clipped by overflow:hidden -->
  <button
    v-if="authStore.isLoggedIn"
    class="sidebar-toggle"
    :class="{ expanded: !isCollapsed }"
    @click="isCollapsed = !isCollapsed"
    :title="isCollapsed ? '展开菜单' : '收起菜单'"
    aria-label="切换侧边栏"
  >
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
      <polyline :points="isCollapsed ? '9 18 15 12 9 6' : '15 18 9 12 15 6'"/>
    </svg>
  </button>
  <div v-else class="login-page-wrapper">
    <LoginView />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import LoginView from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { HopeTheme, HopeSidebar, HopeTopbar, HopeHeroBanner } from '@/components/hope'

const authStore = useAuthStore()
const route = useRoute()
const isCollapsed = ref(false)
</script>

<style scoped>
.hope-app {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.hope-main {
  flex: 1;
  margin-left: 16.2rem;
  display: flex;
  flex-direction: column;
  transition: margin-left 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.hope-main.collapsed {
  margin-left: 4.8rem;
}

.hope-content {
  flex: 1;
  overflow-y: auto;
  background: var(--hope-bg);
  position: relative;
  z-index: 1;
}

.hope-content.has-hero {
  margin-top: -3rem;
  padding: 0;
}

.hope-content:not(.has-hero) {
  margin-top: 0;
  padding-top: 1.5rem;
}

.content-wrapper {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.page-content {
  padding: 0 1.5rem 2rem;
  max-width: 100%;
}

.page-content--no-hero {
  padding-top: 1.5rem;
}

/* ─── Sidebar toggle ───
   Fixed position, outside .hope-app, not clipped by any overflow:hidden.
   Moves with the sidebar edge as it collapses/expands.
*/
.sidebar-toggle {
  position: fixed;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 48px;
  border-radius: 0 12px 12px 0;
  background: #ffffff;
  border: 1px solid rgba(138, 146, 166, 0.25);
  border-left: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8e99a9;
  box-shadow: 2px 0 8px rgba(17, 38, 146, 0.10);
  transition:
    left 0.4s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.2s,
    color 0.2s;
  z-index: 9999;
  padding: 0;
}

.sidebar-toggle:hover {
  background: var(--hope-primary);
  color: #ffffff;
  border-color: var(--hope-primary);
}

.sidebar-toggle svg {
  width: 10px;
  height: 10px;
  transition: transform 0.3s ease;
  display: block;
}

/* Expanded (sidebar open) → button at sidebar right edge (16.2rem) */
.sidebar-toggle.expanded {
  left: 16.2rem;
}

/* Collapsed (sidebar closed) → button at sidebar right edge (4.8rem) */
.sidebar-toggle:not(.expanded) {
  left: 4.8rem;
}

/* Page transition */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.2s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}

/* Login page */
.login-page-wrapper {
  position: fixed;
  inset: 0;
  z-index: 1000;
  overflow: auto;
}

@media (max-width: 768px) {
  .hope-main {
    margin-left: 0 !important;
  }
}
</style>
