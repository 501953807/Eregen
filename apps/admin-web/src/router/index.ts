import { createRouter, createWebHistory, NavigationGuard } from 'vue-router'
// Removed top-level import of useAuthStore to avoid circular dependency during module initialization

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', component: () => import('@/views/Dashboard.vue') },
    { path: '/devices', component: () => import('@/views/Devices.vue') },
    { path: '/pillboxes', redirect: '/devices?type=pillbox' },
    { path: '/subscriptions', component: () => import('@/views/Subscriptions.vue') },
    { path: '/users', component: () => import('@/views/Users.vue') },
    { path: '/institutions', component: () => import('@/views/Institutions.vue') },
    { path: '/alerts', component: () => import('@/views/Alerts.vue') },
    { path: '/analytics', component: () => import('@/views/Analytics.vue') },
    { path: '/settings', component: () => import('@/views/Settings.vue') },
    { path: '/ota', component: () => import('@/views/OTA.vue') },
    { path: '/elderly', component: () => import('@/views/Elderly.vue') },
    { path: '/persons', component: () => import('@/views/Persons.vue') },
    { path: '/self', component: () => import('@/views/SelfChain.vue') },
    { path: '/hospital', component: () => import('@/views/HospitalChain.vue') },
    { path: '/community', component: () => import('@/views/CommunityChain.vue') },
    { path: '/regulatory', component: () => import('@/views/RegulatoryDashboard.vue') },
    { path: '/medication', component: () => import('@/views/Medication.vue'), name: 'Medication' },
    { path: '/medical', component: () => import('@/views/MedicalWristband.vue') },
    { path: '/audit/:patientId', name: 'AuditDetail', component: () => import('@/views/AuditDetail.vue') },
    { path: '/community-wb', component: () => import('@/views/CommunityWristband.vue') },
    // Login route
    { path: '/login', component: () => import('@/views/Login.vue') },
  ],
})

// Add navigation guard to protect routes - use lazy store access
const canAccessProtectedRoute: NavigationGuard = async (to, from, next) => {
  // All routes except /login require authentication - lazily load store
  const { useAuthStore } = await import('@/stores/auth')
  const authStore = useAuthStore()

  if (to.path !== '/login' && !authStore.checkLoggedIn()) {
    const redirectPath = from.path === '/' ? to.path : from.path
    return next({ path: '/login', query: { redirect: redirectPath } })
  }

  if (to.path === '/login' && authStore.checkLoggedIn()) {
    // Already logged in, redirect to dashboard
    return next('/dashboard')
  }

  next()
}

router.beforeEach(canAccessProtectedRoute)

export default router
