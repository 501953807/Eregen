import apiClient from './client'

export interface B2BInstitution {
  id: string
  name: string
  type: 'hospital' | 'nursing_home' | 'community_center' | 'clinic'
  code: string
  contact_name?: string
  contact_phone?: string
  access_level: string
  status: 'pending' | 'active' | 'suspended'
  api_key_count?: number
  created_at: string
  updated_at: string
}

export interface APIKeyResult {
  key_id: string
  key_value: string
  expires: string
}

export const institutionsApi = {
  list(params?: { page?: number; page_size?: number; name?: string; type?: string; status?: string }) {
    return apiClient.get('/admin/institutions', { params })
  },

  get(id: string) {
    return apiClient.get(`/admin/institutions/${id}`)
  },

  create(data: { name: string; code: string; type: string; contact_name?: string; contact_phone?: string; access_level?: string; status?: string }) {
    return apiClient.post('/admin/institutions', data)
  },

  update(id: string, data: Partial<B2BInstitution>) {
    return apiClient.put(`/admin/institutions/${id}`, data)
  },

  delete(id: string) {
    return apiClient.delete(`/admin/institutions/${id}`)
  },

  generateApiKey(id: string, name: string, expiresIn?: number) {
    return apiClient.post(`/admin/institutions/${id}/api-keys`, { name })
  },

  revokeApiKey(id: string, keyId: string) {
    return apiClient.delete(`/admin/institutions/${id}/api-keys/${keyId}`)
  },
}
