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
  list(_params?: { page?: number; page_size?: number; type?: string; status?: string }) {
    return Promise.resolve({ data: [] as B2BInstitution[] })
  },

  get(_id: string) {
    return Promise.resolve({ data: null as B2BInstitution | null })
  },

  create(_data: { name: string; code: string; type: string; contact_name?: string; contact_phone?: string; access_level?: string }) {
    return Promise.resolve({ data: null as B2BInstitution | null })
  },

  update(_id: string, _data: Partial<B2BInstitution>) {
    return Promise.resolve({ data: null as B2BInstitution | null })
  },

  generateApiKey(_id: string, _name: string, _expiresIn: number) {
    return Promise.resolve({ data: null as APIKeyResult | null })
  },
}
