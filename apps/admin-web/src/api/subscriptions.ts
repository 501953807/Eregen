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
    return apiClient.get<{ data: SubscriptionStatItem[] }>('/admin/stats/subscriptions')
  },
  list(params?: SubscriptionListParams) {
    return apiClient.get<{ data: Subscription[]; page: number; page_size: number }>('/admin/subscriptions', { params })
  },
  renew(id: string, endDate: string) {
    return apiClient.put(`/admin/subscriptions/${id}/renew`, { expires_at: endDate })
  },
}
