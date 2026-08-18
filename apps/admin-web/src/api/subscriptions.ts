import apiClient from './client'
import type { Subscription } from '@/types'

export interface SubscriptionListParams {
  page?: number
  page_size?: number
  status?: string
  plan_tier?: string
}

export interface SubscriptionStatItem {
  tier: string
  count: number
  pct: number
}

export const subscriptionsApi = {
  stats() {
    return apiClient.get('/admin/subscriptions/stats')
  },
  list(params?: SubscriptionListParams) {
    return apiClient.get('/admin/subscriptions', { params })
  },
}
