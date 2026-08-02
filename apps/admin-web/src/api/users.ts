import apiClient from './client'
import type { User, ElderlyProfile } from '@/types'

export interface CreateUserBody {
  name: string
  email?: string
  phone?: string
  role: string
  password: string
}

export interface UpdateUserBody {
  name?: string
  email?: string
  phone?: string
  role?: string
}

export const usersApi = {
  list(params?: { page?: number; page_size?: number; role?: string }) {
    return apiClient.get<{ data: User[]; page?: number; page_size?: number }>('/admin/users', { params })
  },
  create(data: CreateUserBody) {
    return apiClient.post<{ data: { id: string; name: string; role: string } }>('/admin/users', data)
  },
  update(id: string, data: UpdateUserBody) {
    return apiClient.put<{ message: string }>(`/admin/users/${id}`, data)
  },
  delete(id: string) {
    return apiClient.delete<{ message: string }>(`/admin/users/${id}`)
  },
  updateRole(id: string, role: string) {
    return apiClient.post<{ message: string }>(`/admin/users/${id}/role`, { role })
  },
  listElderly(params?: Record<string, any>) {
    return apiClient.get<{ data: ElderlyProfile[] }>('/admin/elderly', { params })
  },
}
