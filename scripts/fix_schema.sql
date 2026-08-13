-- ============================================================
-- Eregen Schema Unification Migration
-- Phase 1-8: unify health_records, medication_rules, alerts,
-- location_history, users, drop v2 tables, clear test data
-- ============================================================

-- Phase 1: health_records — add new fields
ALTER TABLE health_records ADD COLUMN person_id TEXT;
ALTER TABLE health_records ADD COLUMN business_chain TEXT CHECK (business_chain IN ('self','hospital','community'));
ALTER TABLE health_records ADD COLUMN record_type TEXT;
ALTER TABLE health_records ADD COLUMN source TEXT CHECK (source IN ('device','nurse','community_staff','his','manual'));
ALTER TABLE health_records ADD COLUMN device_id TEXT;
ALTER TABLE health_records ADD COLUMN recorded_at DATETIME;
ALTER TABLE health_records ADD COLUMN blood_pressure_sys INTEGER;
ALTER TABLE health_records ADD COLUMN blood_pressure_dia INTEGER;
ALTER TABLE health_records ADD COLUMN blood_glucose_fasting REAL;
ALTER TABLE health_records ADD COLUMN blood_glucose_postprandial REAL;
ALTER TABLE health_records ADD COLUMN uric_acid REAL;
ALTER TABLE health_records ADD COLUMN creatinine REAL;
ALTER TABLE health_records ADD COLUMN hemoglobin_a1c REAL;
ALTER TABLE health_records ADD COLUMN respiratory_rate INTEGER;
ALTER TABLE health_records ADD COLUMN pulse_rate INTEGER;
ALTER TABLE health_records ADD COLUMN weight REAL;
ALTER TABLE health_records ADD COLUMN height REAL;
ALTER TABLE health_records ADD COLUMN bmi REAL;
ALTER TABLE health_records ADD COLUMN exercise_minutes INTEGER;
ALTER TABLE health_records ADD COLUMN notes TEXT;
UPDATE health_records SET person_id = elderly_id WHERE elderly_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_health_records_person ON health_records(person_id, business_chain, recorded_at);

-- Phase 2: medication_rules — add new fields
ALTER TABLE medication_rules ADD COLUMN person_id TEXT;
ALTER TABLE medication_rules ADD COLUMN business_chain TEXT CHECK (business_chain IN ('self','hospital','community'));
ALTER TABLE medication_rules ADD COLUMN source_type TEXT CHECK (source_type IN ('custom','doctor_order','care_plan'));
ALTER TABLE medication_rules ADD COLUMN source_id TEXT;
ALTER TABLE medication_rules ADD COLUMN drug_name TEXT;
ALTER TABLE medication_rules ADD COLUMN generic_name TEXT;
ALTER TABLE medication_rules ADD COLUMN drug_category TEXT CHECK (drug_category IN ('prescription','otc','supplement','tcm'));
ALTER TABLE medication_rules ADD COLUMN dosage TEXT;
ALTER TABLE medication_rules ADD COLUMN frequency TEXT;
ALTER TABLE medication_rules ADD COLUMN route TEXT DEFAULT 'oral' CHECK (route IN ('oral','injection','topical','inhalation','other'));
ALTER TABLE medication_rules ADD COLUMN schedule_time1 TEXT;
ALTER TABLE medication_rules ADD COLUMN schedule_time2 TEXT;
ALTER TABLE medication_rules ADD COLUMN schedule_time3 TEXT;
ALTER TABLE medication_rules ADD COLUMN duration TEXT;
ALTER TABLE medication_rules ADD COLUMN pre_meal INTEGER DEFAULT 0;
ALTER TABLE medication_rules ADD COLUMN post_meal INTEGER DEFAULT 0;
ALTER TABLE medication_rules ADD COLUMN special_instructions TEXT;
ALTER TABLE medication_rules ADD COLUMN prescribed_by TEXT;
ALTER TABLE medication_rules ADD COLUMN prescribed_at DATETIME;
UPDATE medication_rules SET person_id = elderly_id WHERE elderly_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_med_rules_person ON medication_rules(person_id, business_chain);

-- Phase 3: alerts — add chain support
ALTER TABLE alerts ADD COLUMN person_id TEXT;
ALTER TABLE alerts ADD COLUMN business_chain TEXT CHECK (business_chain IN ('self','hospital','community'));
ALTER TABLE alerts ADD COLUMN rule_id TEXT;
ALTER TABLE alerts ADD COLUMN data_details TEXT;
UPDATE alerts SET person_id = elderly_id WHERE elderly_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_person ON alerts(person_id, business_chain);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity, business_chain, status);

-- Phase 4: location_history — add person_id
ALTER TABLE location_history ADD COLUMN person_id TEXT;
UPDATE location_history SET person_id = elderly_id WHERE elderly_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_location_history_person ON location_history(person_id, timestamp);

-- Phase 5: users — rebuild with extended role enum
CREATE TABLE IF NOT EXISTS users_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    phone TEXT UNIQUE,
    role TEXT DEFAULT 'family' CHECK (role IN (
        'super_admin','operator','hospital_doc','nurse',
        'community_doctor','community_staff','regulator',
        'family','elderly','institution','admin'
    )),
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new SELECT * FROM users;
DROP TABLE IF EXISTS users;
ALTER TABLE users_new RENAME TO users;

-- Phase 7: Drop v2 tables
DROP TABLE IF EXISTS health_records_v2;
DROP TABLE IF EXISTS medication_rules_v2;

-- Phase 8: Clear all test data (keep table structures and rule definitions)
DELETE FROM compliance_checks;
DELETE FROM health_reports;
DELETE FROM health_guidance_deliveries;
DELETE FROM medication_executions;
DELETE FROM location_history;
DELETE FROM alerts;
DELETE FROM health_records;
DELETE FROM medication_rules;
DELETE FROM device_bindings;
DELETE FROM elderly_devices;
DELETE FROM devices;
DELETE FROM persons;
DELETE FROM person_profiles;
DELETE FROM person_welfare_tags;
DELETE FROM hospital_admissions;
DELETE FROM medical_wristband_patients;
DELETE FROM medical_bindings;
DELETE FROM medical_verifications;
DELETE FROM medical_daily_entries;
DELETE FROM medical_expenses;
DELETE FROM medical_medications;
DELETE FROM medical_test_results;
DELETE FROM community_elders;
DELETE FROM community_signin_records;
DELETE FROM community_elder_welfare;
DELETE FROM community_elder_bindings;
DELETE FROM community_pharmacy_logs;
DELETE FROM community_batch_payments;
DELETE FROM community_minzheng_sync;
DELETE FROM regulatory_alerts;
DELETE FROM regulatory_location_logs;
