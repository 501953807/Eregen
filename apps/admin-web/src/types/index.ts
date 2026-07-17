export interface User {
  id: string
  email?: string
  phone?: string
  name: string
  role: 'family' | 'elderly' | 'institution' | 'admin' | 'operator'
  created_at: string
  updated_at: string
}

export interface ElderlyProfile {
  id: string
  user_id: string
  name: string
  birth_date?: string
  avatar_url?: string
  health_tiers: string[]
  created_at: string
  updated_at: string
}

export interface Device {
  id: string
  device_id: string
  device_type: 'bracelet' | 'pillbox'
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto'
  status: 'online' | 'offline'
  last_seen?: string
  owner_user_id: string
  settings?: Record<string, any>
  created_at: string
}

export interface HealthRecord {
  id: string
  elderly_id: string
  timestamp: string
  hr?: number
  spo2?: number
  steps?: number
  sleep_hours?: number
  bp_systolic?: number
  bp_diastolic?: number
}

export interface LocationRecord {
  id: string
  elderly_id: string
  timestamp: string
  lat: number
  lon: number
  accuracy?: number
}

export interface MedicationRule {
  id: string
  elderly_id: string
  schedule_time: string
  dose_count: number
  pill_type: string
  days_of_week: number[]
  active: boolean
  created_at: string
}

export interface Alert {
  id: string
  elderly_id: string
  alert_type: string
  severity: 'P0' | 'P1' | 'P2'
  status: 'pending' | 'resolved'
  metadata?: Record<string, any>
  created_at: string
  resolved_at?: string
}

export interface Subscription {
  id: string
  user_id: string
  plan_tier: 'free' | 'premium' | 'enterprise'
  status: string
  start_date: string
  end_date: string
}

export interface DashboardStats {
  online_devices: number
  total_devices: number
  active_alerts: number
  total_users: number
  active_subscriptions: number
  alert_trend: TrendPoint[]
}

export interface TrendPoint {
  date: string
  value: number
}
