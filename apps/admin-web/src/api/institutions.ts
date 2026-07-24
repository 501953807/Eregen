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
  created_at: string
  updated_at: string
}

export interface APIKeyResult {
  key_id: string
  key_value: string
  expires: string
}

export const institutionsApi = {
  list(params?: { page?: number; page_size?: number; type?: string; status?: string }) {
    return apiClient.get<{ data: B2BInstitution[]; total: number; page: number }>(
      '/b2b/institutions',
      { params },
    )
  },

  get(id: string) {
    return apiClient.get<{ data: B2BInstitution }>(`/b2b/institutions/${id}`)
  },

  create(data: { name: string; code: string; type: string; contact_name?: string; contact_phone?: string; access_level?: string }) {
    return apiClient.post<{ data: B2BInstitution }>('/b2b/institutions', data)
  },

  update(id: string, data: Partial<B2BInstitution>) {
    return apiClient.put<{ data: B2BInstitution }>(`/b2b/institutions/${id}`, data)
  },

  generateApiKey(id: string, name: string, expiresIn: number) {
    return apiClient.post<{ data: APIKeyResult }>(`/b2b/institutions/${id}/api-keys`, { name, expires_in: expiresIn })
  },
}
