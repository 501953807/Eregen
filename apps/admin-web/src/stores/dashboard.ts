import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dashboardApi } from '@/api/dashboard'
import type { DashboardStats, Alert, TrendPoint, AlertTrendPoint, AlertDistributionItem, UserGrowthPoint } from '@/api/dashboard'

export const useDashboardStore = defineStore('dashboard', () => {
  const stats = ref<DashboardStats>({
    online_devices: 0, total_devices: 0, active_alerts: 0,
    total_users: 0, active_subscriptions: 0, alert_trend: [],
  })
  const chartData = ref<{
    alertTrend: AlertTrendPoint[]
    alertDistribution: AlertDistributionItem[]
    userGrowth: UserGrowthPoint[]
  }>({
    alertTrend: [],
    alertDistribution: [],
    userGrowth: [],
  })
  const recentAlerts = ref<Alert[]>([])
  const loading = ref(false)

  async function fetchOverview() {
    loading.value = true
    try {
      const res = await dashboardApi.overview()
      stats.value = res.data.data || res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchCharts() {
    loading.value = true
    try {
      const [trendRes, distRes, growthRes] = await Promise.all([
        dashboardApi.alertTrend(),
        dashboardApi.alertDistribution(),
        dashboardApi.userGrowth(),
      ])
      chartData.value.alertTrend = trendRes.data.data || []
      chartData.value.alertDistribution = distRes.data.data || []
      chartData.value.userGrowth = growthRes.data.data || []
    } finally {
      loading.value = false
    }
  }

  async function fetchRecentAlerts() {
    try {
      const res = await dashboardApi.recentAlerts({ limit: 10 })
      recentAlerts.value = res.data.data || []
    } catch {
      recentAlerts.value = []
    }
  }

  async function refreshAll() {
    await Promise.all([fetchOverview(), fetchCharts(), fetchRecentAlerts()])
  }

  return { stats, chartData, recentAlerts, loading, fetchOverview, fetchCharts, fetchRecentAlerts, refreshAll }
})
