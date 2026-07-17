import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/dashboard' },
    { path: '/dashboard', component: () => import('@/views/Dashboard.vue') },
    { path: '/devices', component: () => import('@/views/Devices.vue') },
    { path: '/subscriptions', component: () => import('@/views/Subscriptions.vue') },
    { path: '/users', component: () => import('@/views/Users.vue') },
    { path: '/settings', component: () => import('@/views/Settings.vue') },
    { path: '/ota', component: () => import('@/views/OTA.vue') },
    { path: '/elderly', component: () => import('@/views/Elderly.vue') },
  ],
})

export default router
