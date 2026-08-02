import apiClient from './client'
import type { Device } from '@/types'

export const devicesApi = {
  list(params: { page?: number; page_size?: number; status?: string; type?: string; tier?: string }) {
    return apiClient.get<{ data: Device[] }>('/admin/devices', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: Device }>(`/admin/devices/${id}`)
  },
  updateSettings(id: string, settings: Record<string, any>) {
    return apiClient.put(`/admin/devices/${id}/settings`, { settings })
  },
  triggerOTA(id: string, firmwareUrl: string, hash: string) {
    return apiClient.post(`/admin/devices/${id}/ota`, { url: firmwareUrl, hash })
  },
  // Admin endpoints
  adminOtaPush(deviceId: string, firmwareUrl: string, hash: string) {
    return apiClient.post(`/admin/devices/${deviceId}/ota`, { url: firmwareUrl, hash })
  },
  adminUpdateConfig(deviceId: string, config: Record<string, any>) {
    return apiClient.post<{ message: string }>(`/admin/devices/${deviceId}/config`, config)
  },
  adminUnbindDevice(deviceId: string) {
    return apiClient.delete(`/admin/devices/${deviceId}/unbind`)
  },
  batchOtaPush(deviceIds: string[], firmwareUrl: string, hash: string) {
    return apiClient.post('/admin/devices/batch-ota', { device_ids: deviceIds, url: firmwareUrl, hash })
  },
}
