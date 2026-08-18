import apiClient from './client'
import type { ElderlyProfile } from '@/types'

export const elderlyApi = {
  list(params?: Record<string, any>) {
    return apiClient.get('/admin/elderly', { params })
  },
  detail(id: string) {
    return apiClient.get(`/admin/elderly/${id}`)
  },
  update(id: string, data: Partial<ElderlyProfile>) {
    return apiClient.put(`/admin/elderly/${id}`, data)
  },
}
