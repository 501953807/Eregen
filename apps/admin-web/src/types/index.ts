export interface User {
  id: string;
  phone: string;
  email?: string;
  name: string;
  role: ChainRole;
  tier?: 'starter' | 'plus' | 'pro';
  verified?: boolean;
  paid?: boolean;
  last_login?: string;
  created_at: string;
}

export interface Device {
  id: string;
  device_id: string;
  device_type: 'bracelet' | 'pillbox' | 'medical_wristband' | 'community_wristband';
  type?: 'bracelet' | 'pillbox' | 'medical_wristband' | 'community_wristband';
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto';
  status: 'online' | 'offline' | 'pending_upgrade' | 'fault' | 'ota_updating';
  firmware_version: string;
  last_seen: string;
  elder_id?: string;
  owner_name?: string;
  institution?: string;
  mode?: 'family' | 'admin' | 'community' | 'medical' | 'collection' | 'guard';
  battery_pct?: number;
  rssi?: number;
  hr?: number;
  spo2?: number;
  steps?: number;
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
  user_phone?: string;
  plan: 'free' | 'pro' | 'enterprise';
  plan_tier?: 'starter' | 'plus' | 'pro';
  status: 'active' | 'expired' | 'cancelled' | 'pending_renewal' | 'past_due';
  billing_cycle?: 'monthly' | 'annual';
  start_date: string;
  end_date: string;
  downgrade_reason?: string;
  cancellation_reason?: string;
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
  identifier: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token: string;
  user_id: string;
  role: string;
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

// Business Chain Types
export type BusinessChain = 'self' | 'hospital' | 'community' | 'regulatory'

export type ChainRole = 'super_admin' | 'operator' | 'hospital_doc' | 'nurse' | 'community_staff' | 'regulator'

export interface Person {
  id: string
  id_card: string
  name: string
  gender: 0 | 1 | 2  // 0=unknown, 1=male, 2=female
  birth_date?: string
  phone?: string
  emergency_contact?: string
  address?: string
  avatar_url?: string
  status: 'active' | 'suspended' | 'deceased'
  created_at: string
  updated_at: string
}

export interface PersonProfile {
  person_id: string
  business_chain: BusinessChain
  subscription_tier?: 'starter' | 'plus' | 'pro' | 'pro_plus'
  subscription_status?: 'trial' | 'active' | 'expired' | 'cancelled'
  subscription_start?: string
  subscription_end?: string
  health_risk_level?: 'low' | 'medium' | 'high' | 'critical'
  admission_no?: string
  department?: string
  bed_number?: string
  blood_type?: 'A' | 'B' | 'AB' | 'O' | 'unknown'
  attending_doctor?: string
  diagnosis?: string
  admission_date?: string
  expected_discharge_date?: string
  discharge_date?: string
  discharge_type?: 'recovered' | 'transferred' | 'refused' | 'deceased'
  hospital_id?: string
  hospital_id_community?: string
  minzheng_certified?: number
  subsidy_type?: string
  certification_date?: string
  certification_doc?: string
  next_review_date?: string
  linked_person_id?: string
  created_at: string
  updated_at: string
}

export interface PersonWelfareTag {
  person_id: string
  tag_code: string
  valid_from: string
  valid_to: string
}

export interface PersonRoleBinding {
  id: string
  user_id: string
  business_chain: BusinessChain
  role: ChainRole
  institution_id?: string
  granted_by?: string
  granted_at: string
  expires_at?: string
  active: number
  created_at: string
}

export interface AlertRule {
  id: string
  name: string
  business_chain: BusinessChain
  alert_type: string
  severity: 'p0' | 'p1' | 'p2'
  condition_field: string
  condition_operator: string
  condition_threshold?: number
  condition_duration_min?: number
  notify_roles: string
  notify_channels: string
  notify_institution_ids: string
  escalation_timeout_min?: number
  escalation_roles: string
  auto_action: string
  active: number
  created_at: string
  updated_at: string
}

export interface HealthRecordV2 {
  id: string
  person_id: string
  business_chain: BusinessChain
  record_type: string
  source: 'device' | 'nurse' | 'community_staff' | 'his' | 'manual'
  device_id?: string
  recorded_at: string
  heart_rate?: number
  blood_pressure_sys?: number
  blood_pressure_dia?: number
  spo2?: number
  temperature?: number
  respiratory_rate?: number
  pulse_rate?: number
  blood_glucose_fasting?: number
  blood_glucose_postprandial?: number
  uric_acid?: number
  creatinine?: number
  hemoglobin_a1c?: number
  weight?: number
  height?: number
  bmi?: number
  steps?: number
  sleep_hours?: number
  exercise_minutes?: number
  notes?: string
  created_at: string
}

export interface PersonHealthSummary {
  person_id: string
  business_chain: BusinessChain
  latest_hr?: number
  latest_spo2?: number
  latest_bp_sys?: number
  latest_bp_dia?: number
  latest_glucose_fasting?: number
  latest_uric_acid?: number
  latest_steps?: number
  latest_sleep_hours?: number
  risk_score?: number
  trend_direction?: 'improving' | 'stable' | 'worsening'
  last_updated: string
  ai_recommendation?: string
}

export interface MedicationRuleV2 {
  id: string
  person_id: string
  business_chain: BusinessChain
  source_type: 'custom' | 'doctor_order' | 'care_plan'
  source_id?: string
  drug_name: string
  generic_name?: string
  drug_category?: 'prescription' | 'otc' | 'supplement' | 'tcm'
  dosage: string
  frequency: string
  route?: string
  schedule_time1?: string
  schedule_time2?: string
  schedule_time3?: string
  days_of_week?: string
  duration?: string
  pre_meal?: number
  post_meal?: number
  special_instructions?: string
  prescribed_by?: string
  prescribed_at?: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface MedicationExecution {
  id: string
  person_id: string
  business_chain: BusinessChain
  rule_id: string
  scheduled_time: string
  actual_time?: string
  status: 'pending' | 'taken' | 'missed' | 'skipped' | 'alerted' | 'refused'
  taken_by?: 'self' | 'family' | 'nurse' | 'community_staff' | 'pillbox_auto'
  device_id?: string
  verification_method?: 'manual' | 'optical' | 'nfc' | 'barcode'
  notes?: string
  created_at: string
}

export interface HealthGuidanceRule {
  id: string
  name: string
  business_chain: BusinessChain
  trigger_condition: string
  condition_field?: string
  condition_operator?: string
  condition_threshold?: number
  guidance_type: 'diet' | 'exercise' | 'medication' | 'lifestyle' | 'education'
  title: string
  content: string
  priority?: number
  enabled: number
  created_at: string
  updated_at: string
}

export interface HealthGuidanceDelivery {
  id: string
  person_id: string
  business_chain: BusinessChain
  rule_id: string
  guidance_type: string
  title: string
  content: string
  channel: 'app_push' | 'wechat' | 'sms' | 'in_app'
  delivered_at: string
  read_status: number
  feedback?: string
}

export interface HealthReportTemplate {
  id: string
  name: string
  business_chain: BusinessChain
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'annual'
  template_type: 'summary' | 'detailed' | 'trend' | 'alert'
  include_sections?: string
  enabled: number
  created_at: string
}

export interface HealthReport {
  id: string
  person_id: string
  business_chain: BusinessChain
  template_id?: string
  report_period_start: string
  report_period_end: string
  generated_at: string
  report_data?: string
  delivered_channels?: string
  status: 'generated' | 'delivered' | 'failed'
  created_at: string
}

export interface ComplianceRule {
  id: string
  rule_code: string
  name: string
  description: string
  business_chain: BusinessChain
  rule_type: 'fence' | 'medication' | 'billing' | 'length_of_stay'
  condition_sql: string
  severity?: 'p0' | 'p1' | 'p2'
  action_required?: string
  enabled: number
  created_at: string
  updated_at: string
}

export interface ComplianceCheck {
  id: string
  rule_id: string
  person_id: string
  check_time: string
  violated: number
  violation_details?: string
  reviewed_by?: string
  reviewed_at?: string
  review_result?: 'confirmed' | 'false_alarm' | 'investigating'
  action_taken?: string
  created_at: string
}

export interface DeviceBinding {
  id: string
  device_id: string
  person_id: string
  business_chain: BusinessChain
  bound_at: string
  unbound_at?: string
  binding_type?: 'self' | 'hospital' | 'community'
  created_at: string
}

export interface NotificationTemplate {
  id: string
  name: string
  business_chain: BusinessChain
  channel: 'push' | 'sms' | 'wechat' | 'email' | 'in_app'
  subject?: string
  body_template: string
  enabled: number
  created_at: string
}

export interface NotificationLog {
  id: string
  person_id: string
  business_chain: BusinessChain
  template_id?: string
  recipient_role?: string
  recipient_id?: string
  channel: string
  status: 'pending' | 'sent' | 'failed' | 'read'
  sent_at?: string
  read_at?: string
  error_message?: string
  created_at: string
}
