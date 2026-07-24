import apiClient from './client'

export interface RegulatoryFenceConfig {
  id: string
  hospital_id: string
  hospital_name: string
  center_lat: number
  center_lng: number
  radius_meters: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface RegulatoryAlert {
  id: string
  rule_code: string
  patient_id?: string
  patient_name?: string
  hospital_id: string
  department: string
  severity: 'low' | 'medium' | 'high'
  alert_type: string
  detail: string
  status: 'pending' | 'acknowledged' | 'resolved' | 'false_positive'
  triggered_at: string
  acknowledged_at?: string
  acknowledged_by?: string
  resolved_at?: string
  resolved_by?: string
  notes?: string
}

export interface RegulatoryPatientRow {
  id: string
  name: string
  admission_no: string
  department: string
  bed_number: string
  bound_at: string
  last_verify?: string
  verify_gap_hours: number
  fence_status: string
  fence_exit_duration_sec: number
  alert_tags: string[]
  alerts_triggered: string[]
}

export interface RuleConfig {
  code: string
  name: string
  enabled: boolean
  config: Record<string, any>
}

export interface ComplianceSummary {
  total_patients_period: number
  avg_stay_days: number
  fence_violations: number
  no_verify_alerts: number
  expense_anomalies: number
  med_verify_mismatch: number
  compliance_rate: number
}

export interface ComplianceDeptBreakdown {
  name: string
  total_patients: number
  alerts: number
  compliance_rate: number
}

export interface ComplianceReport {
  summary: ComplianceSummary
  department_breakdown: ComplianceDeptBreakdown[]
}

export interface AuditTrail {
  patient: any
  binding: any
  verifications: any[]
  medications: any[]
  expenses: any[]
  daily_entries: any[]
  fence_logs: any[]
  alerts_generated: RegulatoryAlert[]
}

export const regulatoryApi = {
  getDashboardOverview(params?: { department?: string }) {
    return apiClient.get<{ data: any }>('/admin/regulatory/dashboard/patient-overview', { params })
  },
  getPatientList(params?: { department?: string; status?: string; page?: number; page_size?: number }) {
    return apiClient.get<{ data: RegulatoryPatientRow[]; page: number; page_size: number }>('/admin/regulatory/dashboard/patient-list', { params })
  },
  listAlerts(params?: { rule_code?: string; level?: string; status?: string; department?: string; page?: number; page_size?: number }) {
    return apiClient.get<{ data: RegulatoryAlert[]; page: number; page_size: number }>('/admin/regulatory/alerts', { params })
  },
  getAlert(id: string) {
    return apiClient.get<{ data: RegulatoryAlert }>(`/admin/regulatory/alerts/${id}`)
  },
  acknowledgeAlert(id: string, userId: string) {
    return apiClient.post(`/admin/regulatory/alerts/${id}/acknowledge`, { user_id: userId })
  },
  resolveAlert(id: string, userId: string, notes?: string) {
    return apiClient.post(`/admin/regulatory/alerts/${id}/resolve`, { user_id: userId, notes })
  },
  getAuditTrail(patientId: string) {
    return apiClient.get<{ data: AuditTrail }>(`/admin/regulatory/audit/patient/${patientId}`)
  },
  listRuleConfigs() {
    return apiClient.get<{ data: RuleConfig[] }>('/admin/regulatory/rules')
  },
  updateRuleConfig(code: string, config: Record<string, any>) {
    return apiClient.put(`/admin/regulatory/rules/${code}/config`, { config })
  },
  configureFence(data: Omit<RegulatoryFenceConfig, 'id' | 'created_at' | 'updated_at'>) {
    return apiClient.post('/admin/regulatory/fence/config', data)
  },
  getFenceConfig(hospitalId: string) {
    return apiClient.get<{ data: RegulatoryFenceConfig }>('/admin/regulatory/fence/config', { params: { hospital_id: hospitalId } })
  },
  getComplianceReport(params: { hospital_id?: string; start_date?: string; end_date?: string }) {
    return apiClient.get<{ data: ComplianceReport }>('/admin/regulatory/compliance/report', { params })
  },
}
