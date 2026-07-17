import apiClient from './client'

export interface FirmwareVersion {
  id: string
  device_type: string
  tier: string
  version: string
  release_date: string
  download_url: string
  changelog: string
  is_latest: boolean
}

export const otaApi = {
  listVersions(params?: Record<string, any>) {
    return apiClient.get<{ data: FirmwareVersion[] }>('/admin/firmware-versions', { params })
  },
  createVersion(data: Partial<FirmwareVersion>) {
    return apiClient.post('/admin/firmware-versions', data)
  },
  pushUpgrade(deviceIds: string[], firmwareId: string) {
    return apiClient.post('/admin/ota/push', { device_ids: deviceIds, firmware_id: firmwareId })
  },
  deleteVersion(id: string) {
    return apiClient.delete(`/admin/firmware-versions/${id}`)
  },
}
