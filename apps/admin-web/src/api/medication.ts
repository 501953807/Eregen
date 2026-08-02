import apiClient from './client'
import type { MedicationRule } from '@/types'

export const medicationApi = {
  listRules(elderlyId: string) {
    return apiClient.get<{ data: MedicationRule[] }>(`/admin/elderly/${elderlyId}/medication-rules`)
  },
  createRule(elderlyId: string, data: Omit<MedicationRule, 'id' | 'elderly_id' | 'active' | 'created_at'>) {
    return apiClient.post<{ data: MedicationRule }>(`/admin/elderly/${elderlyId}/medication-rules`, data)
  },
  updateRule(elderlyId: string, ruleId: string, data: Partial<MedicationRule>) {
    return apiClient.put<{ message: string }>(`/admin/elderly/${elderlyId}/medication-rules/${ruleId}`, data)
  },
  deleteRule(elderlyId: string, ruleId: string) {
    return apiClient.delete<{ message: string }>(`/admin/elderly/${elderlyId}/medication-rules/${ruleId}`)
  },
}
