import apiClient from './client'
import type { Alert } from '@/types'

export const alertsApi = {
  list(params: { severity?: string; status?: string; limit?: number }) {
    return apiClient.get('/alerts', { params })
  },
  markResolved(id: string) {
    return apiClient.put(`/alerts/${id}/resolve`, {})
  },
  acknowledge(id: string) {
    return apiClient.put(`/alerts/${id}`, { status: 'acknowledged' })
  },
}
