import apiClient from './client'
import type { Alert } from '@/types'

export interface AlertTrendPoint {
  date: string
  bracelet_count: number
  pillbox_count: number
}

export interface AlertDistributionItem {
  name: string
  value: number
  color: string
}

export interface UserGrowthPoint {
  month: string
  new_users: number
}

export const dashboardApi = {
  overview() {
    return apiClient.get<{ data: any }>('/admin/stats/overview')
  },
  alertTrend(params?: Record<string, any>) {
    return apiClient.get<{ data: AlertTrendPoint[] }>('/admin/stats/alert-trend', { params })
  },
  alertDistribution() {
    return apiClient.get<{ data: AlertDistributionItem[] }>('/admin/stats/alert-distribution')
  },
  userGrowth() {
    return apiClient.get<{ data: UserGrowthPoint[] }>('/admin/stats/user-growth')
  },
  recentAlerts(params?: Record<string, any>) {
    return apiClient.get<{ data: Alert[] }>('/alerts', { params: { ...params, limit: 10 } })
  },
}
