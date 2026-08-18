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
    return apiClient.get('/admin/stats/overview')
  },
  alertTrend(params?: Record<string, any>) {
    return apiClient.get('/admin/stats/alert-trend', { params })
  },
  alertDistribution() {
    return apiClient.get('/admin/stats/alert-distribution')
  },
  userGrowth(params?: Record<string, any>) {
    return apiClient.get('/admin/stats/user-growth', { params })
  },
  recentAlerts(params?: Record<string, any>) {
    return apiClient.get('/admin/alerts', { params: { ...params, page: 1, page_size: 10 } })
  },
}
