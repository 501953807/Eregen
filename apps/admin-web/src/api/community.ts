import apiClient from './client'

export interface CommunityElder {
  id: string
  name: string
  id_card: string
  gender: number
  age: number
  address: string
  emergency_contact: string
  hospital_id: string
  status: 'active' | 'deactivated' | 'deceased'
  created_at: string
  updated_at: string
}

export interface CommunityDevice {
  id: string
  device_id: string
  firmware_version: string
  mode: 'hospital' | 'community'
  status: 'active' | 'inactive' | 'retired'
  last_seen: string
  bound_elder_name?: string
  created_at: string
}

export interface WelfareTagConfig {
  id: string
  tag_code: string
  tag_name: string
  issuer: string
  renewal_period_days: number
  benefit_amount: number
  enabled: boolean
}

export interface SigninRecord {
  id: string
  elder_id: string
  elder_name?: string
  device_id: string
  hospital_id: string
  signin_time: string
  period: string
  is_medical_signin: boolean
  is_welfare_signin: boolean
}

export interface BatchPayment {
  id: string
  batch_id: string
  period: string
  pay_type: string
  elder_id: string
  amount: number
  status: 'pending' | 'success' | 'failed' | 'retrying'
  failure_reason?: string
  executed_at: string
  created_at: string
}

export const communityApi = {
  // Elders
  listElders(params?: { status?: string; page?: number; page_size?: number }) {
    return apiClient.get<{ data: CommunityElder[]; page: number; page_size: number }>('/admin/community-wb/elders', { params })
  },
  getElder(id: string) {
    return apiClient.get<{ data: CommunityElder }>(`/admin/community-wb/elders/${id}`)
  },
  createElder(data: Omit<CommunityElder, 'id' | 'created_at' | 'updated_at'>) {
    return apiClient.post('/admin/community-wb/elders', data)
  },
  updateElder(id: string, data: Partial<CommunityElder>) {
    return apiClient.put(`/admin/community-wb/elders/${id}`, data)
  },
  deleteElder(id: string) {
    return apiClient.delete(`/admin/community-wb/elders/${id}`)
  },
  getElderStats() {
    return apiClient.get<{ data: any }>('/admin/community-wb/elders/stats')
  },
  // Devices
  listDevices(params?: { status?: string; page?: number; page_size?: number }) {
    return apiClient.get<{ data: CommunityDevice[]; page: number; page_size: number }>('/admin/community-wb/devices', { params })
  },
  bindDevice(data: { elder_id: string; device_id: string }) {
    return apiClient.post('/admin/community-wb/devices/bind', data)
  },
  // Welfare tags
  listWelfareTags() {
    return apiClient.get<{ data: WelfareTagConfig[] }>('/admin/community-wb/welfare-tags')
  },
  assignWelfareTag(elderId: string, tagCode: string, data: { valid_from: string; valid_to: string; certified_by?: string }) {
    return apiClient.post(`/admin/community-wb/elders/${elderId}/welfare/${tagCode}`, data)
  },
  revokeWelfareTag(elderId: string, tagCode: string) {
    return apiClient.delete(`/admin/community-wb/elders/${elderId}/welfare/${tagCode}`)
  },
  // Sign-in
  triggerSignin(data: { elder_id: string; device_id: string; hospital_id: string; period: string; activated_tags?: string[]; is_medical_signin?: boolean; is_welfare_signin?: boolean }) {
    return apiClient.post('/admin/community-wb/signin/trigger', data)
  },
  listSigninRecords(params?: { elder_id?: string; period?: string; hospital_id?: string }) {
    return apiClient.get<{ data: SigninRecord[] }>('/admin/community-wb/signin/records', { params })
  },
  // Pharmacy
  dispenseMedicine(data: { elder_id: string; hospital_id: string; period: string; items: string[]; total_cost?: number; insurance_covered?: number; self_pay?: number }) {
    return apiClient.post('/admin/community-wb/pharmacy/dispense', data)
  },
  // Minzheng
  importMinzhengData(data: { source: string; filename?: string }) {
    return apiClient.post('/admin/community-wb/minzheng/import', data)
  },
  listMinzhengSync() {
    return apiClient.get<{ data: any[] }>('/admin/community-wb/minzheng/sync')
  },
  // Batch payments
  executeBatchPayment(data: { batch_id: string; period: string; pay_type: string; elder_ids: string[] }) {
    return apiClient.post('/admin/community-wb/batch-pay/execute', data)
  },
  listBatchPayments(params?: { batch_id?: string }) {
    return apiClient.get<{ data: BatchPayment[] }>('/admin/community-wb/batch-payments', { params })
  },
  // NFC authentication
  nfcAuth(data: { elder_id: string; serial_number?: string }) {
    return apiClient.post('/admin/community-wb/nfc-auth', data)
  },
}
