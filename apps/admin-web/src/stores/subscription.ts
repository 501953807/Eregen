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
  const page = ref(1)
  const pageSize = ref(20)
  const total = ref(0)

  const activeCount = computed(() => subscriptions.value.filter(s => s.status === 'active').length)
  const expiringCount = computed(() => subscriptions.value.filter(s => {
    if (!s.end_date) return false
    const days = Math.ceil((new Date(s.end_date).getTime() - Date.now()) / 86400000)
    return days > 0 && days <= 7
  }).length)

  async function fetchList(status?: string, planTier?: string) {
    loading.value = true
    try {
      const res = await subscriptionsApi.list({ page: page.value, page_size: pageSize.value, status, plan_tier: planTier })
      const d = res as any
      subscriptions.value = d.data || []
      total.value = d.page_size || d.data?.length || 0
    } catch (error) {
      console.error('Failed to fetch subscriptions:', error)
      subscriptions.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const res = await subscriptionsApi.stats()
      const tiers = (res as any)?.tiers || (res as any)?.data?.tiers || []
      let total = 0, active = 0, expired = 0
      for (const s of tiers) {
        total += s.count
        if (s.tier === 'starter' || s.tier === 'plus' || s.tier === 'pro') active += s.count
        if (s.tier === 'expired' || s.tier === 'past_due') expired += s.count
      }
      stats.value = { total, active, expiring: expiringCount.value, expired }
    } catch {
      // Keep defaults
    }
  }

  return { subscriptions, stats, loading, page, pageSize, total, activeCount, expiringCount, fetchList, fetchStats }
})
