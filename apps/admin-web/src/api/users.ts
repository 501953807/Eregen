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
    return apiClient.get('/users', { params })
  },
  create(data: CreateUserBody) {
    return apiClient.post('/users', data)
  },
  update(id: string, data: UpdateUserBody) {
    return apiClient.put(`/users/${id}`, data)
  },
  delete(id: string) {
    return apiClient.delete(`/users/${id}`)
  },
  updateRole(id: string, role: string) {
    return apiClient.post(`/admin/users/${id}/role`, { role })
  },
  listElderly(params?: Record<string, any>) {
    return apiClient.get('/elderly', { params })
  },
}
