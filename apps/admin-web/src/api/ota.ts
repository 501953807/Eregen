import apiClient from './client'

export interface FirmwareRelease {
  id: string
  device_type: 'bracelet' | 'pillbox'
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto'
  version: string
  url: string
  sha256_hash: string
  changelog: string
  min_app_version?: string
  force_update: boolean
  active: boolean
  created_at: string
  updated_at: string
}

export interface OTAJobProgress {
  total: number
  pending: number
  downloading: number
  succeeding: number
  succeeded: number
  failed: number
}

export interface OTAJob {
  id: string
  firmware_id: string
  target_devices: string[]
  progress: OTAJobProgress
  created_at: string
  updated_at: string
}

export interface PushOTARequest {
  firmware_id: string
  device_ids?: string[]
}

export interface CreateFirmwareRequest {
  device_type: string
  tier: string
  version: string
  url: string
  sha256_hash: string
  changelog?: string
  min_app_version?: string
  force_update?: boolean
}

function firmwareParams(params?: Record<string, any>): Record<string, any> {
  const p: Record<string, any> = {}
  if (params?.device_type) p.device_type = params.device_type
  if (params?.tier) p.tier = params.tier
  return p
}

export const otaApi = {
  listFirmware(params?: Record<string, any>) {
    return apiClient.get<{ data: FirmwareRelease[] }>('/admin/firmware', { params: firmwareParams(params) })
  },
  getFirmware(id: string) {
    return apiClient.get<{ data: FirmwareRelease }>(`/admin/firmware/${id}`)
  },
  createFirmware(data: CreateFirmwareRequest) {
    return apiClient.post<{ data: FirmwareRelease }>('/admin/firmware', data)
  },
  verifyFirmware(id: string) {
    return apiClient.post<{ data: { valid: boolean; status: string } }>(`/admin/firmware/${id}/verify`)
  },
  pushOTA(data: PushOTARequest) {
    return apiClient.post<{ data: { job_id: string; target_count: number; firmware_version: string } }>('/admin/ota/push', data)
  },
  getOTAJob(id: string) {
    return apiClient.get<{ data: OTAJob }>(`/admin/ota/jobs/${id}`)
  },
}
