import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { subscriptionsApi } from '@/api/subscriptions'
import type { Subscription } from '@/types'

export const useSubscriptionStore = defineStore('subscription', () => {
  const subscriptions = ref<Subscription[]>([])
  const stats = ref<{ total: number; active: number; expiring: number; expired: number }>({
    total: 0, active: 0, expiring: 0, expired: 0,
  })
  const loading = ref(false)

  const total = computed(() => stats.value.total)

  async function fetchList() {
    loading.value = true
    try {
      // Backend does not expose a list endpoint yet, use mock data
      subscriptions.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const res = await subscriptionsApi.stats()
      const tiers = res.data.data || []
      let total = 0, active = 0, expiring = 0, expired = 0
      for (const s of tiers) {
        total += s.count
        if (s.tier === 'starter' || s.tier === 'plus' || s.tier === 'pro') active += s.count
        if (s.tier === 'expired' || s.tier === 'past_due') expired += s.count
      }
      stats.value = { total, active, expiring, expired }
    } catch {
      // Keep defaults
    }
  }

  return { subscriptions, total, stats, loading, fetchList, fetchStats }
})
