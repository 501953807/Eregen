<template>
  <HopeTheme v-if="authStore.user">
    <div class="hope-app">
      <!-- Sidebar -->
      <HopeSidebar :collapsed="isCollapsed" @toggle="isCollapsed = !isCollapsed" />

      <!-- Main Container -->
      <div class="hope-main" :class="{ 'collapsed': isCollapsed }">
        <!-- Topbar -->
        <HopeTopbar />

        <!-- Hero Banner (Dashboard only) -->
        <HopeHeroBanner v-if="route.path === '/dashboard'" />

        <!-- Content wrapper with overlap -->
        <div class="content-wrapper">
          <main class="hope-content" :class="{ 'has-hero': route.path === '/dashboard' }">
            <!-- Content card with proper padding for non-Dashboard pages -->
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
  <div v-else class="login-page-wrapper">
    <LoginView />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import LoginView from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import { HopeTheme, HopeSidebar, HopeTopbar, HopeHeroBanner } from '@/components/hope'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const { isDark, toggle: toggleTheme } = useTheme()

const isCollapsed = ref(false)

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
  margin-top: -3rem;
  padding: 0;
  position: relative;
  z-index: 1;
}

.hope-content.has-hero {
  padding-top: 0;
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
  min-height: 100vh;
}

/* Responsive */
@media (max-width: 768px) {
  .hope-main {
    margin-left: 0 !important;
  }
}
</style>
