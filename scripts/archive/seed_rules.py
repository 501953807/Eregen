#!/usr/bin/env python3
"""Seed alert rules, compliance rules via API."""
import requests
import sys

API = "http://localhost:8089/api/v1"
token = None

def login():
    global token
    r = requests.post(f"{API}/auth/login", json={
        "method": "email", "credential": "admin@eregen.com", "secret": "Admin@123"
    }, timeout=10)
    if r.status_code != 200:
        print(f"LOGIN FAILED: {r.text}"); sys.exit(1)
    token = r.json()["data"]["token"]
    print("✅ Logged in as admin")

def main():
    login()
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}

    # Seed alert rules
    alert_rules = [
        {"name": "心率过高", "business_chain": "self", "alert_type": "abnormal_hr", "severity": "p1",
         "condition_field": "hr", "condition_operator": ">", "condition_threshold": 120},
        {"name": "心率过低", "business_chain": "self", "alert_type": "abnormal_hr", "severity": "p1",
         "condition_field": "hr", "condition_operator": "<", "condition_threshold": 50},
        {"name": "血氧过低", "business_chain": "self", "alert_type": "abnormal_spo2", "severity": "p1",
         "condition_field": "spo2", "condition_operator": "<", "condition_threshold": 92},
        {"name": "跌倒检测", "business_chain": "self", "alert_type": "fall", "severity": "p0",
         "condition_field": "spo2", "condition_operator": "=", "condition_threshold": 0},
        {"name": "SOS按钮", "business_chain": "self", "alert_type": "sos", "severity": "p0",
         "condition_field": "spo2", "condition_operator": "=", "condition_threshold": 0},
        {"name": "漏服药物", "business_chain": "self", "alert_type": "med_missed", "severity": "p2",
         "condition_field": "spo2", "condition_operator": "=", "condition_threshold": 0},
        {"name": "生命体征异常", "business_chain": "hospital", "alert_type": "abnormal_vitals", "severity": "p1",
         "condition_field": "hr", "condition_operator": ">", "condition_threshold": 110},
        {"name": "跌倒住院", "business_chain": "hospital", "alert_type": "fall", "severity": "p0",
         "condition_field": "spo2", "condition_operator": "=", "condition_threshold": 0},
        {"name": "签到漏签", "business_chain": "community", "alert_type": "signin_miss", "severity": "p1",
         "condition_field": "spo2", "condition_operator": "=", "condition_threshold": 0},
    ]

    for rule in alert_rules:
        r = requests.post(f"{API}/admin/alert-rules", headers=headers, json=rule)
        if r.status_code in (200, 201):
            print(f"  ✅ Alert rule: {rule['name']} [{rule['business_chain']}]")
        else:
            print(f"  ❌ Alert rule {rule['name']}: {r.text[:80]}")

    # Seed compliance rules
    compliance_rules = [
        {"rule_code": "R_S01", "name": "心跳持续异常", "description": "连续3次心率>120",
         "business_chain": "self", "rule_type": "medication",
         "condition_sql": "SELECT COUNT(*) FROM health_records WHERE hr > 120 AND recorded_at > datetime('now', '-30 min')",
         "severity": "p1", "action_required": "通知家属"},
        {"rule_code": "R_S02", "name": "血糖超标", "description": "空腹血糖>7.0mmol/L",
         "business_chain": "self", "rule_type": "medication",
         "condition_sql": "SELECT COUNT(*) FROM health_records WHERE blood_glucose_fasting > 7.0",
         "severity": "p1", "action_required": "建议就医"},
        {"rule_code": "R_H01", "name": "住院超期", "description": "住院天数>30天且无延长申请",
         "business_chain": "hospital", "rule_type": "length_of_stay",
         "condition_sql": "SELECT COUNT(*) FROM hospital_admissions WHERE julianday('now') - julianday(admission_date) > 30",
         "severity": "p1", "action_required": "核查"},
        {"rule_code": "R_C01", "name": "重复签到", "description": "同一老人在不同社区医院同日内签到",
         "business_chain": "community", "rule_type": "billing",
         "condition_sql": "SELECT COUNT(*) FROM community_signin_records WHERE date(signin_time) = date('now') GROUP BY elder_id HAVING COUNT(DISTINCT hospital_id) > 1",
         "severity": "p1", "action_required": "核查"},
    ]

    for rule in compliance_rules:
        r = requests.post(f"{API}/admin/compliance-rules", headers=headers, json=rule)
        if r.status_code in (200, 201):
            print(f"  ✅ Compliance rule: {rule['rule_code']}")
        else:
            print(f"  ❌ Compliance rule {rule['rule_code']}: {r.text[:80]}")

    print("\nRules seeding complete.")

if __name__ == "__main__":
    main()
