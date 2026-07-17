import { defineStore } from 'pinia'
import { ref } from 'vue'
import { devicesApi } from '@/api/devices'
import type { Device } from '@/types'

export const useDeviceStore = defineStore('device', () => {
  const devices = ref<Device[]>([])
  const loading = ref(false)
  const total = ref(0)
  const stats = ref<{ bracelet_count: number; pillbox_count: number; online_rate: number }>({
    bracelet_count: 0, pillbox_count: 0, online_rate: 0,
  })

  async function fetchList(params?: Record<string, any>) {
    loading.value = true
    try {
      const res = await devicesApi.list(params || {})
      devices.value = res.data.data || []
      total.value = (res.data as any).total || devices.value.length
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const res = await devicesApi.list({ page_size: 1 })
      const list = res.data.data || []
      const bracelets = list.filter((d: Device) => d.device_type === 'bracelet')
      const pillboxes = list.filter((d: Device) => d.device_type === 'pillbox')
      const onlineCount = list.filter((d: Device) => d.status === 'online').length
      stats.value = {
        bracelet_count: bracelets.length,
        pillbox_count: pillboxes.length,
        online_rate: list.length ? Math.round((onlineCount / list.length) * 1000) / 10 : 0,
      }
    } catch {
      stats.value = { bracelet_count: 0, pillbox_count: 0, online_rate: 0 }
    }
  }

  return { devices, loading, total, stats, fetchList, fetchStats }
})
