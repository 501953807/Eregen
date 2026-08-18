-- Seed data for Eregen Admin API (SQLite MVP)
-- Generated: 2026-07-28
-- Purpose: Initialize demo data for development and testing

-- ============================================
-- 1. Insert default admin user (for dual-login email method)
-- Note: Password is stored with bcrypt hash for "Admin@123"
-- ============================================
INSERT OR IGNORE INTO users (id, name, email, role, password_hash, phone) VALUES
('admin-user-001', '系统管理员', 'admin@example.com', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgbflz5YxVbM1VdR46oHXzXPlI3G', '+8613800000000');

-- ============================================
-- 2. Insert elderly user + profile + device linkage
-- (Family member / Operator user + linked elder)
-- ============================================
-- Create family user (phone login method)
INSERT OR IGNORE INTO users (id, name, email, role, password_hash, phone) VALUES
('user-fam-001', '张建国家属', 'family@example.com', 'family', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgbflz5YxVbM1VdR46oHXzXPlI3G', '+8613900000000');

-- Create elder profile (same person in some cases, or separate)
INSERT OR IGNORE INTO elderly_profiles (id, name, user_id, birth_date, health_tiers, avatar_url) VALUES
('elderly-001', '张建国', 'user-fam-001', '1950-01-01', '{"cardiovascular":true,"diabetes":true},"hypertension":true}', 'https://example.com/avatar.jpg');

-- Link elder to devices
-- Device 1: bracelet (entry tier)
INSERT OR IGNORE INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings) VALUES
('device-br-001', 'BR-ZJC001', 'bracelet', 'starter', 'online', datetime('now'), 'user-fam-001', '{"fw_version":"v1.0.0","interval":30}');

-- Device 2: pillbox (smart tier)
INSERT OR IGNORE INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings) VALUES
('device-px-001', 'PX-ZJC001', 'pillbox', 'smart', 'online', datetime('now'), 'user-fam-001', '{"fw_version":"v2.1.0","has_audio":true}');

-- Link elderly to devices
INSERT OR IGNORE INTO elderly_devices (id, elderly_id, device_id) VALUES
('eld-dev-001', 'elderly-001', 'device-br-001'),
('eld-dev-002', 'elderly-001', 'device-px-001');

-- ============================================
-- 3. Insert sample alerts
-- ============================================
INSERT OR IGNORE INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id) VALUES
('alert-sos-001', 'elderly-001', 'sos', 'high', 'pending', '老人按下SOS按钮', 'device-br-001'),
('alert-fall-001', 'elderly-001', 'fall', 'medium', 'resolved', '检测到跌倒，已处理', 'device-br-001'),
('alert-med-001', 'elderly-001', 'med_dose_missed', 'low', 'pending', '漏服降压药提醒', 'device-px-001');

-- ============================================
-- 4. Insert sample health records
-- ============================================
WITH RECURSIVE t(n) AS (VALUES(0) UNION ALL SELECT n+1 FROM t WHERE n < 4)
INSERT OR IGNORE INTO health_records (id, elderly_id, timestamp, hr, spo2, steps, sleep_hours)
SELECT
    'hr-' || lower(hex(randomblob(8))) || '-' || n,
    'elderly-001',
    strftime('%Y-%m-%d %H:%M:%S', datetime('-5 hours', '+' || n || ' hours')),
    72 + (abs(random()) % 10),
    95 + (abs(random()) % 6),
    (abs(random()) % 5000) + 2000,
    CAST((random() / 1000.0) * 8 AS REAL)
FROM t;

-- ============================================
-- 5. Insert sample medication rule
-- ============================================
INSERT OR IGNORE INTO medication_rules (id, elderly_id, schedule_time, pill_type, dose_count, days_of_week, active) VALUES
('med-rule-001', 'elderly-001', '08:00', 'tablet', 2, '[0,1,2,3,4,5,6]', 1),
('med-rule-002', 'elderly-001', '12:00', 'tablet', 1, '[0,1,2,3,4,5,6]', 1),
('med-rule-003', 'elderly-001', '20:00', 'capsule', 1, '[0,1,2,3,4,5,6]', 1);

-- ============================================
-- 6. Insert sample location history
-- ============================================
INSERT OR IGNORE INTO location_history (id, elderly_id, lat, lon, accuracy, timestamp) VALUES
('loc-001', 'elderly-001', 39.9042, 116.4074, 15, datetime('now')),
('loc-002', 'elderly-001', 39.9142, 116.4174, 10, datetime('-1 hour')),
('loc-003', 'elderly-001', 39.9242, 116.4274, 20, datetime('-2 hours'));

-- ============================================
-- 7. Insert default regulatory rule configs
-- ============================================
INSERT OR IGNORE INTO regulatory_rule_config (rule_code, rule_name, enabled, config_json) VALUES
('R01', '挂床住院', 1, '{"max_verify_gap_hours":24,"severity":"high"}'),
('R02', '电子围栏越界', 1, '{"max_fence_exit_minutes":30,"severity":"high"}'),
('R03', '虚假入院', 1, '{"bind_duration_hours":48,"severity":"medium"}'),
('R04', '费用突增', 1, '{"expense_multiplier":3,"severity":"medium"}'),
('R05', '用药与核验不匹配', 1, '{"severity":"medium"}'),
('R06', '频繁转科', 1, '{"transfers_per_week":3,"severity":"low"}'),
('R07', '腕带异常断开', 1, '{"disconnect_hours":2,"severity":"high"}'),
('R08', '长期不在院', 1, '{"severity":"low"}'),
('R_C01', '重复领取福利', 1, '{"overlap_days":30,"severity":"high"}'),
('R_C02', '跨社区医院互认', 1, '{"enabled":1,"severity":"low"}'),
('R_C03', '冒领嫌疑', 1, '{"id_card_mismatch":1,"severity":"high"}'),
('R_C04', '福利标签超期未续', 1, '{"grace_days":7,"severity":"medium"}'),
('R_C05', '签到-发药时间差异常', 1, '{"max_gap_hours":24,"severity":"medium"}'),
('R_C06', '批量发放失败重试超限', 1, '{"max_retries":3,"severity":"high"}'),
('R_C07', '僵尸账户', 1, '{"inactive_days":180,"severity":"low"}'),
('R_C08', '死亡后仍激活', 1, '{"severity":"high"}');

-- ============================================
-- 8. Insert community welfare tags
-- ============================================
INSERT OR IGNORE INTO community_welfare_tag_config (tag_code, tag_name, issuer, renewal_period_days, benefit_amount) VALUES
('orphan', '孤寡老人', '民政局', 365, 0),
('poverty_level_1', '特困一级', '民政局', 365, 800),
('poverty_level_2', '特困二级', '民政局', 365, 500),
('disability_level_1', '残疾一级', '残联', 365, 1200),
('disability_level_2', '残疾二级', '残联', 365, 800),
('special_disease', '特病补助', '医保局', 180, 2000),
('bus_discount', '乘车补贴', '民政局', 365, 360),
('elder_care_subsidy', '高龄津贴', '民政局', 365, 200),
('nursing_subsidy', '护理补贴', '民政局', 365, 600);

-- ============================================
-- 9. Insert sample hospital admission data (medical wristband workflow)
-- ============================================
INSERT OR IGNORE INTO medical_wristband_patients (id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, status) VALUES
('wb-patient-001', 'H202607001', '李大伯', 1, 72, '心内科', 'A05', 'O型', '青霉素过敏', '冠心病史, 糖尿病', 'admitted'),
('wb-patient-002', 'H202607002', '王阿姨', 2, 68, '内分泌科', 'B12', 'A型', '头孢过敏', '高血压', 'admitted');

-- Medical wristband device binding
INSERT OR IGNORE INTO medical_wristband_devices (id, device_id, firmware_version, status, bound_patient_id) VALUES
('wb-device-001', 'WB-BED001', 'v1.2.3', 'idle', NULL),
('wb-device-002', 'WB-BED002', 'v1.2.3', 'active', 'wb-patient-001');

-- Medical binding
-- INSERT for medical_bindings
INSERT OR IGNORE INTO medical_bindings (id, patient_id, device_id, bound_at) VALUES
('wb-bind-001', 'wb-patient-001', 'wb-device-002', datetime('now'));

-- ============================================
-- ============================================
-- 完成数据初始化（Development Only） -- 生产环境请使用真实数据库初始化流程 --
-- ============================================
-- END OF SEED DATA