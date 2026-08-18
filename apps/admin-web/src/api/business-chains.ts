import apiClient from './client'
import type { Person, PersonProfile, PersonWelfareTag, MedicationRuleV2, MedicationExecution, PersonRoleBinding, AlertRule, HealthRecordV2, PersonHealthSummary, HealthGuidanceRule, HealthGuidanceDelivery, HealthReportTemplate, HealthReport, ComplianceRule, ComplianceCheck, DeviceBinding, NotificationTemplate, NotificationLog } from '@/types'

export const personApi = {
  list(params?: { page?: number; page_size?: number; business_chain?: string; status?: string }) {
    return apiClient.get<{ data: Person[]; page: number; page_size: number }>('/admin/persons', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: Person }>(`/admin/persons/${id}`)
  },
  create(data: Partial<Person>) {
    return apiClient.post('/admin/persons', data)
  },
  update(id: string, data: Partial<Person>) {
    return apiClient.put(`/admin/persons/${id}`, data)
  },
  delete(id: string) {
    return apiClient.delete(`/admin/persons/${id}`)
  },
  createProfile(data: Partial<PersonProfile>) {
    return apiClient.post('/admin/persons/profile', data)
  },
  getProfile(personId: string, chain: string) {
    return apiClient.get<{ data: PersonProfile }>(`/admin/persons/${personId}/profile?chain=${chain}`)
  },
  assignWelfareTag(data: Partial<PersonWelfareTag>) {
    return apiClient.post('/admin/persons/welfare-tags', data)
  },
  revokeWelfareTag(personId: string, tagCode: string) {
    return apiClient.delete(`/admin/persons/${personId}/welfare-tags/${tagCode}`)
  },
  listWelfareTags(personId: string) {
    return apiClient.get<{ data: PersonWelfareTag[] }>(`/admin/persons/${personId}/welfare-tags`)
  },
}

export const selfApi = {
  listElderly(params?: any) {
    return apiClient.get('/self/elderly', { params })
  },
  getHealthReport(personId: string) {
    return apiClient.get(`/self/elderly/${personId}/health-report`)
  },
  getGuidance(personId: string) {
    return apiClient.post(`/self/elderly/${personId}/guidance`)
  },
}

export const hospitalApi = {
  listPatients(params?: any) {
    return apiClient.get('/admin/hospital/patients', { params })
  },
  getPatient(id: string) {
    return apiClient.get(`/admin/hospital/patients/${id}`)
  },
  admitPatient(data: any) {
    return apiClient.post('/admin/hospital/admissions', data)
  },
  dischargePatient(id: string, data: any) {
    return apiClient.post(`/admin/hospital/admissions/${id}/discharge`, data)
  },
  getDailyEntries(personId: string, date?: string) {
    return apiClient.get(`/admin/hospital/patients/${personId}/daily`, { params: { date } })
  },
  verifyPatient(personId: string, data: any) {
    return apiClient.post(`/admin/hospital/patients/${personId}/verify`, data)
  },
  listAdmissions(params?: any) {
    return apiClient.get('/admin/medical/admissions', { params })
  },
  getWardRounds(personId: string) {
    return apiClient.get(`/admin/medical/patients/${personId}/ward-round`)
  },
  completeWardRound(personId: string, data: any) {
    return apiClient.post(`/admin/medical/patients/${personId}/ward-round`, data)
  },
}

export const communityApi = {
  listElders(params?: any) {
    return apiClient.get('/admin/community/elders', { params })
  },
  createElder(data: any) {
    return apiClient.post('/admin/community/elders', data)
  },
  updateElder(id: string, data: any) {
    return apiClient.put(`/admin/community/elders/${id}`, data)
  },
  signin(personId: string, data: any) {
    return apiClient.post(`/admin/community/elders/${personId}/signin`, data)
  },
  assignWelfare(personId: string, data: any) {
    return apiClient.post(`/admin/community/elders/${personId}/welfare`, data)
  },
  getStats(personId: string) {
    return apiClient.get(`/admin/community/elders/${personId}/stats`)
  },
  revokeWelfareTag(elderId: string, tagCode: string) {
    return apiClient.delete(`/admin/community-wb/elders/${elderId}/welfare/${tagCode}`)
  },
}

export const regulatoryApi = {
  getCompliance(params?: any) {
    return apiClient.get('/regulatory/compliance', { params })
  },
  runCompliance(params?: any) {
    return apiClient.post('/regulatory/compliance/run', params)
  },
  getAudit(personId: string) {
    return apiClient.get(`/regulatory/audit/patient/${personId}`)
  },
  getReports(params?: any) {
    return apiClient.get('/regulatory/reports', { params })
  },
}

export const medicationApi = {
  listRules(personId: string, chain?: string) {
    return apiClient.get(`/admin/persons/${personId}/medications`, { params: { chain } })
  },
  createRule(personId: string, data: Partial<MedicationRuleV2>) {
    return apiClient.post(`/admin/persons/${personId}/medications`, data)
  },
  updateRule(personId: string, ruleId: string, data: any) {
    return apiClient.put(`/admin/persons/${personId}/medications/${ruleId}`, data)
  },
  deleteRule(personId: string, ruleId: string) {
    return apiClient.delete(`/admin/persons/${personId}/medications/${ruleId}`)
  },
  createExecution(personId: string, data: Partial<MedicationExecution>) {
    return apiClient.post(`/admin/persons/${personId}/medications/executions`, data)
  },
  listExecutions(personId: string, chain?: string, limit = 50) {
    return apiClient.get(`/admin/persons/${personId}/medications/executions`, { params: { chain, limit } })
  },
}

export const healthApi = {
  listRecords(personId: string, chain?: string, recordType?: string, limit = 50) {
    return apiClient.get(`/admin/persons/${personId}/health`, { params: { chain, record_type: recordType, limit } })
  },
  getSummary(personId: string, chain: string) {
    return apiClient.get(`/admin/persons/${personId}/health/summary`, { params: { chain } })
  },
  createRecord(personId: string, data: Partial<HealthRecordV2>) {
    return apiClient.post(`/admin/persons/${personId}/health`, data)
  },
}

export const lifecycleApi = {
  transitionStatus(personId: string, data: { business_chain: string; new_status: string; reason?: string }) {
    return apiClient.put(`/admin/persons/${personId}/status`, data)
  },
  linkPersons(data: { person_id_1: string; person_id_2: string; business_chain_1: string; business_chain_2: string }) {
    return apiClient.post('/admin/persons/link', data)
  },
}
