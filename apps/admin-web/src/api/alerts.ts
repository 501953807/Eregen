import apiClient from './client'
import type { Alert } from '@/types'

export const alertsApi = {
  list(params: { severity?: string; status?: string; limit?: number }) {
    return apiClient.get<{ data: Alert[] }>('/admin/alerts', { params })
  },
  markResolved(id: string) {
    return apiClient.post(`/admin/alerts/${id}/resolve`, {})
  },
  acknowledge(id: string) {
    return apiClient.post(`/admin/alerts/${id}/acknowledge`, {})
  },
}
