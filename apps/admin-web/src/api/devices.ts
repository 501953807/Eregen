import apiClient from './client'
import type { Device } from '@/types'

export const devicesApi = {
  list(params: { page?: number; page_size?: number; status?: string; type?: string; tier?: string }) {
    return apiClient.get('/admin/devices', { params })
  },
  detail(id: string) {
    return apiClient.get(`/admin/devices/${id}`)
  },
  updateSettings(id: string, settings: Record<string, any>) {
    return apiClient.post(`/admin/devices/${id}/config`, { settings })
  },
  triggerOTA(id: string, firmwareUrl: string, hash: string) {
    return apiClient.post(`/admin/devices/${id}/ota`, { url: firmwareUrl, hash })
  },
  updateConfig(deviceId: string, config: Record<string, any>) {
    return apiClient.post(`/admin/devices/${deviceId}/config`, config)
  },
  batchOtaPush(deviceIds: string[], firmwareUrl: string, hash: string) {
    return apiClient.post('/admin/devices/batch-ota', { device_ids: deviceIds, url: firmwareUrl, hash })
  },
  unbind(id: string) {
    return apiClient.delete(`/admin/devices/${id}/unbind`)
  },
}
