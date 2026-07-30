export interface User {
  id: string;             // UUID
  phone: string;          // UNIQUE
  email?: string;         // Optional, for email login
  name: string;           // 姓名
  role: 'elder' | 'family' | 'operator' | 'nurse' | 'admin';
  created_at: string;     // ISO timestamp
}

export interface Device {
  id: string;             // UUID
  device_id: string;      // BR-XXXX / PX-XXXX / WB-XXXX
  device_type: 'bracelet' | 'pillbox' | 'medical_wristband';
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto';
  status: 'online' | 'offline' | 'pending_upgrade' | 'fault';
  firmware_version: string;
  last_seen: string;      // ISO timestamp
  elder_id?: string;      // Optional FK to elders.id
  owner_name?: string;    // Display name of bound elder
  institution?: string;   // Assigned institution
  mode?: 'family' | 'admin' | 'community' | 'medical';
  battery_pct?: number;   // Battery percentage
  rssi?: number;          // Signal strength
}

export interface Elderly {
  id: string;
  user_id: string;        // FK to users.id
  name: string;
  birth_date: string;     // YYYY-MM-DD
  gender: 'male' | 'female';
  conditions: string[];   // JSON array of chronic disease tags
  emergency_contacts: string[];
  avatar_url?: string;
}

export interface Subscription {
  id: string;
  user_id: string;
  plan: 'free' | 'pro' | 'enterprise';
  status: 'active' | 'expired' | 'cancelled' | 'pending_renewal';
  start_date: string;
  end_date: string;
  downgrade_reason?: string;
  per_device?: boolean;   // Enterprise flag
}

export interface Alert {
  id: string;
  dev_id?: string;
  elderly_id?: string;
  alert_type: 'sos' | 'fall' | 'heart' | 'medication' | 'geofence' | 'battery';
  severity: 'p0' | 'p1' | 'p2';
  status: 'pending' | 'acknowledged' | 'resolved';
  location?: { lat: number; lon: number };
  created_at: string;
  updated_at?: string;
  metadata?: Record<string, any>;
}

export interface DashboardStats {
  online_devices: number;
  total_devices: number;
  active_alerts: number;
  total_users: number;
  active_subscriptions: number;
  alert_trend: Array<{ date: string; bracelet_count: number; pillbox_count: number }>;
}

// API response envelope (used consistently)
export interface ApiResponse<T> {
  code: number;           // HTTP status or custom code
  msg: string;            // Error/success message
  data: T;                // Actual payload
}

// Login request/response
export interface LoginRequest {
  method: 'email' | 'phone';
  credential: string;     // email or phone number
  secret: string;         // password or OTP
}

export interface LoginResponse {
  token: string;          // JWT
  user: User;
}

export interface AuthState {
  token: string | null;
  user: User | null;
  expiresAt: number | null; // Unix timestamp
}

// Medical wristband related interfaces (from design doc §4.5)
export interface WbPatient {
  id: string;
  patient_id: string;     // HIS patient ID
  wb_id: string;          // WB-XXXX medical wristband
  ward: string;
  bed_number: string;
  doctor_id: string;      // Attending nurse/doctor
  admission_date: string;
  discharge_date?: string;
  status: 'admitted' | 'discharged' | 'transferred';
}

export interface WbDevice {
  id: string;
  device_id: string;      // WB-XXXX
  serial_number: string;
  binding_time: string;
  last_sync: string;
  status: 'active' | 'inactive' | 'fault';
}

export interface WbBinding {
  id: string;
  wb_patient_id: string;
  wb_device_id: string;
  bound_by: string;       // Operator user ID
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

// For dashboard chart data (matching api/dashboard.ts)
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
