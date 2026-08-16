import apiClient from './client'
import type { ElderlyProfile } from '@/types'

export const elderlyApi = {
  list(params?: Record<string, any>) {
    return apiClient.get('/elderly', { params })
  },
  detail(id: string) {
    return apiClient.get(`/elderly/${id}/profile`)
  },
  update(id: string, data: Partial<ElderlyProfile>) {
    return apiClient.put(`/elderly/${id}/profile`, data)
  },
}
