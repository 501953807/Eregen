import apiClient from './client'
import type { MedicationRule } from '@/types'

export const medicationApi = {
  // GET /api/v1/elderly/{elderly_id}/medication/rules
  listRules(elderlyId: string) {
    return apiClient.get<{ data: MedicationRule[] }>(`/api/v1/elderly/${elderlyId}/medication/rules`)
  },

  // POST /api/v1/elderly/{elderly_id}/medication/rules
  createRule(elderlyId: string, data: Omit<MedicationRule, 'id' | 'elderly_id' | 'active' | 'created_at'>) {
    return apiClient.post(`/api/v1/elderly/${elderlyId}/medication/rules`, data)
  },

  // PUT /api/v1/elderly/{elderly_id}/medication/rules/{rule_id}
  updateRule(elderlyId: string, ruleId: string, data: Partial<Omit<MedicationRule, 'id' | 'elderly_id'>>) {
    return apiClient.put(`/api/v1/elderly/${elderlyId}/medication/rules/${ruleId}`, data)
  },

  // DELETE /api/v1/elderly/{elderly_id}/medication_rules/{rule_id}
  deleteRule(elderlyId: string, ruleId: string) {
    return apiClient.delete(`/api/v1/elderly/${elderlyId}/medication/rules/${ruleId}`)
  },

  // GET /api/v1/elderly/{elderly_id}/medication/today
  getTodayStatus(elderlyId: string) {
    return apiClient.get(`/api/v1/elderly/${elderlyId}/medication/today`)
  },

  // GET /api/v1/elderly/{elderly_id}/medication/history
  getHistory(elderlyId: string, days: number = 30) {
    return apiClient.get(`/api/v1/elderly/${elderlyId}/medication/history`, { params: { days } })
  },

  // POST /api/v1/medication/:rule_id/take (manual mark by family member)
  takeRule(ruleId: string) {
    return apiClient.post(`/api/v1/medication/${ruleId}/take`)
  }
}
