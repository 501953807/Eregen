import apiClient from './client'
import type { ElderlyProfile } from '@/types'

export const elderlyApi = {
  list(params?: Record<string, any>) {
    return apiClient.get<{ data: ElderlyProfile[] }>('/admin/elderly', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: ElderlyProfile }>(`/admin/elderly/${id}`)
  },
  create(data: Partial<ElderlyProfile>) {
    return apiClient.post('/admin/elderly', data)
  },
  update(id: string, data: Partial<ElderlyProfile>) {
    return apiClient.put(`/admin/elderly/${id}`, data)
  },
  delete(id: string) {
    return apiClient.delete(`/admin/elderly/${id}`)
  },
}
