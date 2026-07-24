import apiClient from './client'

export interface Patient {
  id: string
  admission_no: string
  name: string
  gender?: string
  age?: number
  department?: string
  bed_number?: string
  blood_type?: string
  allergies?: string
  special_conditions?: string
  tag_ids?: string[]
  status: string
  created_at: string
  updated_at: string
}

export interface WristbandDevice {
  id: string
  device_id: string
  firmware_version: string
  status: string
  bound_patient_id?: string
  created_at: string
  updated_at: string
}

export interface VerificationRecord {
  id: string
  patient_id: string
  device_id: string
  scan_type: string
  result: string
  verified_by?: string
  lat?: number
  lon?: number
  notes?: string
  timestamp: string
}

export interface RegulatoryAlert {
  id: string
  rule_code: string
  severity: string
  patient_id?: string
  message: string
  triggered_at: string
  resolved: boolean
}

export interface HospitalAdmission {
  id: string
  patient_id: string
  admission_no: string
  bed_no: string
  department: string
  diagnosis?: string
  emergency_contact?: string
  allergies?: string
  admitted_at: string
  expected_discharge_at?: string
  discharged_at?: string
  discharge_type?: string
  transferred_to?: string
  notes?: string
}

export interface WardRoundEntry {
  id: string
  patient_id: string
  nurse_id: string
  blood_pressure?: string
  heart_rate?: number
  spo2?: number
  temperature?: number
  weight?: number
  notes?: string
  observations?: string
  completed_at: string
}

export const medicalApi = {
  // Patients
  listPatients(params: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get('/medical/patients', { params })
  },

  getPatient(id: string) {
    return apiClient.get(`/medical/patients/${id}`)
  },

  createPatient(data: Partial<Patient>) {
    return apiClient.post('/medical/patients', data)
  },

  updatePatient(id: string, data: Partial<Patient>) {
    return apiClient.put(`/medical/patients/${id}`, data)
  },

  deletePatient(id: string) {
    return apiClient.delete(`/medical/patients/${id}`)
  },

  getByAdmissionNo(admissionNo: string) {
    return apiClient.get('/medical/patients/by-admission', { params: { admission_no: admissionNo } })
  },

  batchImport(patients: Partial<Patient>[]) {
    return apiClient.post('/medical/patients/batch-import', patients)
  },

  getPatientHistory(patientId: string) {
    return apiClient.get(`/medical/patients/${patientId}/history`)
  },

  // Wristband devices
  listWristbands(params: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get('/medical/wristbands', { params })
  },

  bindWristband(patientId: string, deviceId: string) {
    return apiClient.post('/medical/wristbands/bind', { patient_id: patientId, device_id: deviceId })
  },

  unbindWristband(bindingId: string) {
    return apiClient.post(`/medical/wristbands/${bindingId}/unbind`)
  },

  clearWristband(deviceId: string) {
    return apiClient.post(`/medical/wristbands/${deviceId}/clear`)
  },

  writeToFirmware(deviceId: string, data: string) {
    return apiClient.post(`/medical/wristbands/${deviceId}/write`, { data })
  },

  getFirmware(deviceId: string) {
    return apiClient.get(`/medical/wristbands/${deviceId}/firmware`)
  },

  // Verifications
  listVerifications(params: { page?: number; page_size?: number }) {
    return apiClient.get('/medical/verifications', { params })
  },

  createVerification(data: Partial<VerificationRecord>) {
    return apiClient.post('/medical/verifications', data)
  },

  updateVerificationStatus(id: string, status: string) {
    return apiClient.put(`/medical/verifications/${id}/status`, { status })
  },

  getTodayStats() {
    return apiClient.get('/medical/verifications/stats/today')
  },

  // Stats
  getOverview() {
    return apiClient.get('/medical/stats/overview')
  },

  // Clinical workflow
  admitPatient(data: { patient_id: string; bed_no: string; department: string; diagnosis?: string; emergency_contact?: string; allergies?: string; expected_stay_days?: number }) {
    return apiClient.post('/medical/admissions', data)
  },

  listAdmissions(params?: { page?: number; page_size?: number; department?: string; status?: string }) {
    return apiClient.get('/medical/admissions', { params })
  },

  dischargePatient(id: string, data: { discharge_type: string; notes?: string; transferred_to?: string }) {
    return apiClient.post(`/medical/admissions/${id}/discharge`, data)
  },

  getWardRounds(patientId: string) {
    return apiClient.get(`/medical/patients/${patientId}/ward-round`)
  },

  completeWardRound(patientId: string, data: { nurse_id: string; blood_pressure?: string; heart_rate?: number; spo2?: number; temperature?: number; weight?: number; notes?: string; observations?: string }) {
    return apiClient.post(`/medical/patients/${patientId}/ward-round`, data)
  },

  getRegulatoryAlerts(params?: { rule_code?: string; severity?: string; status?: string; department?: string; page?: number; page_size?: number }) {
    return apiClient.get('/regulatory/alerts', { params })
  },

  resolveRegulatoryAlert(alertId: string, data: { user_id: string; notes?: string }) {
    return apiClient.put(`/regulatory/alerts/${alertId}/resolve`, data)
  },

  getAuditTrail(patientId: string) {
    return apiClient.get(`/regulatory/audit/patient/${patientId}`)
  },

  exportReport(patientId: string) {
    return apiClient.get(`/regulatory/compliance/report`, { params: { patient_id: patientId } })
  },
}
