import { createRouter, createWebHistory, NavigationGuard, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: () => import('@/views/Dashboard.vue'), meta: { chains: ['self'] } },
  { path: '/devices', component: () => import('@/views/Devices.vue'), meta: { chains: ['self'] } },
  { path: '/pillboxes', redirect: '/devices?type=pillbox' },
  { path: '/subscriptions', component: () => import('@/views/Subscriptions.vue'), meta: { chains: ['self'] } },
  { path: '/users', component: () => import('@/views/Users.vue'), meta: { chains: ['self'] } },
  { path: '/institutions', component: () => import('@/views/Institutions.vue'), meta: { chains: ['self'] } },
  { path: '/alerts', component: () => import('@/views/Alerts.vue'), meta: { chains: ['self', 'hospital', 'community'] } },
  { path: '/analytics', component: () => import('@/views/Analytics.vue'), meta: { chains: ['self'] } },
  { path: '/settings', component: () => import('@/views/Settings.vue'), meta: { chains: ['self', 'hospital', 'community', 'regulatory'] } },
  { path: '/ota', component: () => import('@/views/OTA.vue'), meta: { chains: ['self'] } },
  { path: '/elderly', component: () => import('@/views/Elderly.vue'), meta: { chains: ['self'] } },
  { path: '/persons', component: () => import('@/views/Persons.vue'), meta: { chains: ['self', 'hospital', 'community'] } },
  { path: '/self', component: () => import('@/views/SelfChain.vue'), meta: { chains: ['self'] } },
  { path: '/hospital', component: () => import('@/views/HospitalChain.vue'), meta: { chains: ['hospital'] } },
  { path: '/community', component: () => import('@/views/CommunityChain.vue'), meta: { chains: ['community'] } },
  { path: '/regulatory', component: () => import('@/views/RegulatoryDashboard.vue'), meta: { chains: ['hospital', 'community', 'regulatory'] } },
  { path: '/medication', component: () => import('@/views/Medication.vue'), name: 'Medication', meta: { chains: ['self', 'hospital', 'community'] } },
  { path: '/medical', component: () => import('@/views/MedicalWristband.vue'), meta: { chains: ['hospital'] } },
  { path: '/medical/workstation', redirect: '/medical' },
  { path: '/audit/:patientId', name: 'AuditDetail', component: () => import('@/views/AuditDetail.vue'), meta: { chains: ['hospital', 'community', 'regulatory'] } },
  { path: '/community-wb', component: () => import('@/views/CommunityWristband.vue'), meta: { chains: ['community'] } },
  { path: '/login', component: () => import('@/views/Login.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Route → business chain mapping for access control
const routeChainMap: Record<string, string[]> = {
  '/dashboard': ['self'],
  '/devices': ['self'],
  '/pillboxes': ['self'],
  '/ota': ['self'],
  '/subscriptions': ['self'],
  '/users': ['self'],
  '/institutions': ['self'],
  '/alerts': ['self', 'hospital', 'community'],
  '/analytics': ['self'],
  '/settings': ['self', 'hospital', 'community', 'regulatory'],
  '/elderly': ['self'],
  '/persons': ['self', 'hospital', 'community'],
  '/self': ['self'],
  '/hospital': ['hospital'],
  '/community': ['community'],
  '/regulatory': ['hospital', 'community', 'regulatory'],
  '/audit': ['hospital', 'community', 'regulatory'],
  '/community-wb': ['community'],
  '/medication': ['self', 'hospital', 'community'],
  '/medical': ['hospital'],
}

// Role → allowed business chains (mirrors backend ChainPermissions)
const roleChainMap: Record<string, string[]> = {
  super_admin: ['self', 'hospital', 'community', 'regulatory'],
  operator: ['self'],
  hospital_doc: ['hospital'],
  nurse: ['hospital'],
  community_staff: ['community'],
  regulator: ['hospital', 'community', 'regulatory'],
}

const canAccessRoute: NavigationGuard = async (to, from, next) => {
  const { useAuthStore } = await import('@/stores/auth')
  const authStore = useAuthStore()

  // Login handling
  if (to.path === '/login') {
    if (authStore.checkLoggedIn()) return next('/dashboard')
    return next()
  }

  // Auth check
  if (!authStore.checkLoggedIn()) {
    const redirectPath = from.path === '/' || from.path === '' ? to.path : from.path
    return next({ path: '/login', query: { redirect: redirectPath } })
  }

  // Chain permission check for protected routes
  const role = authStore.getUser?.role
  if (!role) return next('/login')

  const allowedChains = roleChainMap[role] ?? []

  // Check all matched routes for chain access
  const matchedPaths = routes.filter(r => {
    const path = r.path.replace(/\/:.*$/, '')
    return to.path.startsWith(path)
  })

  const hasAccess = matchedPaths.some(r => {
    const chains = (r.meta?.chains as string[]) ?? routeChainMap[r.path.replace(/\/:.*$/, '')] ?? []
    return chains.some(c => allowedChains.includes(c))
  })

  if (!hasAccess) {
    // Redirect to first accessible route or dashboard
    return next('/dashboard')
  }

  next()
}

router.beforeEach(canAccessRoute)

export default router
