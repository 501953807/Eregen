import apiClient from './client'

export const settingsApi = {
  changePassword(data: { old_password: string; new_password: string }) {
    return apiClient.post('/admin/settings/password', data)
  },
  getNotificationSettings() {
    return apiClient.get('/admin/settings/notifications')
  },
  updateNotificationSettings(data: Record<string, any>) {
    return apiClient.put('/admin/settings/notifications', data)
  },
  getApiKeyList() {
    return apiClient.get('/admin/settings/api-keys')
  },
  createApiKey(data: { name: string; expires_at?: string }) {
    return apiClient.post('/admin/settings/api-keys', data)
  },
  revokeApiKey(id: string) {
    return apiClient.delete(`/admin/settings/api-keys/${id}`)
  },
}
