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

export interface SubscriptionStat {
  tier: string
  count: number
  pct: number
}

export interface DashboardOverview {
  online_devices: number
  total_devices: number
  active_alerts: number
  total_users: number
  active_subscriptions: number
  p0: number
  p1: number
  p2: number
}

export const dashboardApi = {
  overview(): Promise<any> {
    return apiClient.get('/admin/stats/overview')
  },
  subscriptionStats(): Promise<any> {
    return apiClient.get('/admin/stats/subscriptions')
  },
  userGrowth(params?: Record<string, any>): Promise<any> {
    return apiClient.get('/admin/stats/user-growth', { params })
  },
  alertTrend(params?: Record<string, any>): Promise<any> {
    return apiClient.get('/admin/stats/alert-trend', { params })
  },
  alertDistribution(): Promise<any> {
    return apiClient.get('/admin/stats/alert-distribution')
  },
  recentAlerts(params?: Record<string, any>): Promise<any> {
    return apiClient.get('/admin/alerts', { params: { ...params, page: 1, page_size: 10 } })
  },
}
