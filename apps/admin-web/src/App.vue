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

        <!-- Content -->
        <main class="hope-content">
          <router-view v-slot="{ Component }">
            <transition name="page-fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </main>
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

function handleUserAction(command: string) {
  if (command === 'logout') handleLogout()
  else ElMessage.info('功能开发中...')
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
  transition: margin-left 0.4s ease-in-out;
  overflow: hidden;
}

.hope-main.collapsed {
  margin-left: 0;
}

.hope-content {
  flex: 1;
  overflow-y: auto;
  background: var(--hope-bg);
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
