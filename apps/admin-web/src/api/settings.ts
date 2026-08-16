import apiClient from './client'

export const settingsApi = {
  changePassword(data: { old_password: string; new_password: string }) {
    return apiClient.post('/auth/change-password', data)
  },
}
