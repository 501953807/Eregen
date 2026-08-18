import apiClient from './client'
import type { Alert } from '@/types'

export const alertsApi = {
  list(params: { severity?: string; status?: string; page?: number; page_size?: number }) {
    return apiClient.get('/admin/alerts', { params })
  },
  markResolved(id: string) {
    return apiClient.post(`/admin/alerts/${id}/resolve`, {})
  },
  acknowledge(id: string) {
    return apiClient.post(`/admin/alerts/${id}/acknowledge`, {})
  },
}
