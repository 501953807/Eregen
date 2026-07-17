import apiClient from './client'
import type { User, ElderlyProfile } from '@/types'

export const usersApi = {
  list(params: { page?: number; page_size?: number; role?: string }) {
    return apiClient.get<{ data: User[] }>('/users', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: User }>(`/users/${id}`)
  },
  updateRole(id: string, role: string) {
    return apiClient.put(`/users/${id}/role`, { role })
  },
  listElderly(params?: Record<string, any>) {
    return apiClient.get<{ data: ElderlyProfile[] }>('/elderly', { params })
  },
}
