import apiClient from './client'

export const subscriptionsApi = {
  list(params?: Record<string, any>) {
    return apiClient.get('/subscriptions', { params })
  },
  stats() {
    return apiClient.get('/subscriptions/stats')
  },
}
