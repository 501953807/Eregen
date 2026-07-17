import { defineStore } from 'pinia'
import { ref } from 'vue'
import { subscriptionsApi } from '@/api/subscriptions'
import type { Subscription } from '@/types'

export const useSubscriptionStore = defineStore('subscription', () => {
  const renewals = ref<Subscription[]>([])
  const stats = ref<{ total: number; active: number; expiring: number; expired: number }>({
    total: 0, active: 0, expiring: 0, expired: 0,
  })
  const loading = ref(false)

  async function fetchList(params?: Record<string, any>) {
    loading.value = true
    try {
      const res = await subscriptionsApi.list(params)
      renewals.value = (res.data.data || res.data) as Subscription[]
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const res = await subscriptionsApi.stats()
      stats.value = res.data.data || res.data || stats.value
    } catch {
      // Keep defaults
    }
  }

  return { renewals, stats, loading, fetchList, fetchStats }
})
