import apiClient from './client'
import type { Alert } from '@/types'

export const alertsApi = {
  list(params: { severity?: string; status?: string; limit?: number }) {
    return apiClient.get<{ data: Alert[] }>('/alerts', { params })
  },
  markResolved(id: string) {
    return apiClient.put(`/alerts/${id}/status`, { status: 'resolved' })
  },
}
