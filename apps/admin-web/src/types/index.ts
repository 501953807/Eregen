export interface User {
  id: string;
  phone: string;
  email?: string;
  name: string;
  role: 'elder' | 'family' | 'operator' | 'nurse' | 'admin' | 'super_admin' | 'elderly';
  tier?: 'starter' | 'plus' | 'pro';
  verified?: boolean;
  paid?: boolean;
  last_login?: string;
  created_at: string;
}

export interface Device {
  id: string;
  device_id: string;
  device_type: 'bracelet' | 'pillbox' | 'medical_wristband';
  type?: 'bracelet' | 'pillbox' | 'medical_wristband';
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto';
  status: 'online' | 'offline' | 'pending_upgrade' | 'fault' | 'ota_updating';
  firmware_version: string;
  last_seen: string;
  elder_id?: string;
  owner_name?: string;
  institution?: string;
  mode?: 'family' | 'admin' | 'community' | 'medical';
  battery_pct?: number;
  rssi?: number;
  ota_progress?: number;
  ota_status?: string;
  ota_speed?: string;
  ota_eta?: string;
  created_at?: string;
  settings?: Record<string, any>;
}

export interface Elderly {
  id: string;
  user_id: string;
  name: string;
  birth_date: string;
  gender: 'male' | 'female';
  conditions: string[];
  emergency_contacts: string[];
  avatar_url?: string;
  tier?: string;
  verified?: boolean;
  paid?: boolean;
  last_login?: string;
  created_at?: string;
}

export type ElderlyProfile = Elderly;

export interface HealthRecord {
  id: string;
  elderly_id: string;
  timestamp: string;
  hr?: number;
  spo2?: number;
  steps?: number;
  sleep_hours?: number;
  bp_systolic?: number;
  bp_diastolic?: number;
}

export interface LocationRecord {
  id: string;
  elderly_id: string;
  timestamp: string;
  lat: number;
  lon: number;
  accuracy?: number;
}

export interface MedicationRule {
  id?: string;
  elderly_id: string;
  schedule_time: string;
  dose_count: number;
  pill_type: string;
  days_of_week: number[];
  active: boolean;
  created_at?: string;
  pillType?: string;
  doseCount?: number;
  scheduleTime?: string;
  daysOfWeek?: number[];
}

export interface Subscription {
  id: string;
  user_id: string;
  user_name?: string;
  plan: 'free' | 'pro' | 'enterprise';
  plan_tier?: 'starter' | 'plus' | 'pro';
  status: 'active' | 'expired' | 'cancelled' | 'pending_renewal' | 'past_due';
  billing_cycle?: 'monthly' | 'annual';
  start_date: string;
  end_date: string;
  downgrade_reason?: string;
  per_device?: boolean;
  total_spent?: number;
  devices?: string[];
}

export interface Alert {
  id: string;
  dev_id?: string;
  elderly_id?: string;
  alert_type: 'sos' | 'fall' | 'heart' | 'medication' | 'geofence' | 'battery';
  severity: 'p0' | 'p1' | 'p2' | 'high' | 'medium' | 'low';
  status: 'pending' | 'acknowledged' | 'resolved';
  location?: { lat: number; lon: number };
  created_at: string;
  updated_at?: string;
  resolved_at?: string;
  metadata?: Record<string, any>;
}

export interface DashboardStats {
  online_devices: number;
  total_devices: number;
  active_alerts: number;
  total_users: number;
  active_subscriptions: number;
  p0: number;
  p1: number;
  p2: number;
  alert_trend: Array<{ date: string; bracelet_count: number; pillbox_count: number }>;
}

export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

export interface LoginRequest {
  method: 'email' | 'phone';
  credential: string;
  secret: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface AuthState {
  token: string | null;
  user: User | null;
  expiresAt: number | null;
}

export interface WbPatient {
  id: string;
  patient_id: string;
  wb_id: string;
  ward: string;
  bed_number: string;
  doctor_id: string;
  admission_date: string;
  discharge_date?: string;
  status: 'admitted' | 'discharged' | 'transferred';
}

export interface WbDevice {
  id: string;
  device_id: string;
  serial_number: string;
  binding_time: string;
  last_sync: string;
  status: 'active' | 'inactive' | 'fault';
}

export interface WbBinding {
  id: string;
  wb_patient_id: string;
  wb_device_id: string;
  bound_by: string;
  bind_time: string;
  unbind_time?: string;
}

export interface WbMedication {
  id: string;
  patient_id: string;
  medication_name: string;
  dosage: string;
  administered_time: string;
  operator_id: string;
}

export interface WbVerification {
  id: string;
  wb_binding_id: string;
  verified_at: string;
  verifier_id: string;
  verification_result: 'success' | 'failed';
  notes?: string;
}

export interface AlertTrendPoint {
  date: string;
  value: number;
}

export interface AlertDistributionItem {
  alert_type: string;
  count: number;
}

export interface UserGrowthPoint {
  date: string;
  new_users: number;
}
