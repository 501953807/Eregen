import apiClient from './client'

export interface SubscriptionItem {
  tier: string
  count: number
  pct: number
}

export const subscriptionsApi = {
  stats() {
    return apiClient.get<{ data: SubscriptionItem[] }>('/admin/stats/subscriptions')
  },
}
