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
}
