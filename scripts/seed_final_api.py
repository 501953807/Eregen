#!/usr/bin/env python3
"""
Eregen 全业务链数据模拟 - 通过API注入（最终版）
"""
import requests
import uuid
import random
import json
import time
from datetime import datetime, timedelta

BASE_URL = "http://localhost:8089/api/v1"
AUTH_URL = f"{BASE_URL}/auth/login"

ROLES = {
    "super_admin": {"email": "admin@eregen.com", "password": "Admin@123"},
}

ELDERS = [
    # Self chain (9 persons)
    {"id_card": "310101195401010001", "name": "张建国", "gender": 1, "age": 72, "tier": "pro", "risk": "high", "has_pillbox": True},
    {"id_card": "310101195801020002", "name": "李秀芳", "gender": 2, "age": 68, "tier": "plus", "risk": "medium", "has_pillbox": True},
    {"id_card": "310101195101030003", "name": "王德明", "gender": 1, "age": 75, "tier": "starter", "risk": "low", "has_pillbox": False},
    {"id_card": "310101195601040004", "name": "赵美华", "gender": 2, "age": 70, "tier": "pro_plus", "risk": "critical", "has_pillbox": True},
    {"id_card": "310101196101050005", "name": "陈志强", "gender": 1, "age": 65, "tier": "plus", "risk": "medium", "has_pillbox": False},
    {"id_card": "310101194801060006", "name": "刘淑珍", "gender": 2, "age": 78, "tier": "pro", "risk": "high", "has_pillbox": False},
    {"id_card": "310101196301070007", "name": "孙伟民", "gender": 1, "age": 63, "tier": "starter", "risk": "low", "has_pillbox": False},
    {"id_card": "310101195601080008", "name": "周海涛", "gender": 1, "age": 70, "tier": "pro", "risk": "high", "has_pillbox": True},
    {"id_card": "310101195801090009", "name": "吴雪梅", "gender": 2, "age": 68, "tier": "pro_plus", "risk": "critical", "has_pillbox": True},
    # Hospital chain (5 persons)
    {"id_card": "310101195501100010", "name": "郑国华", "gender": 1, "age": 71, "dept": "呼吸科", "blood_type": "A"},
    {"id_card": "310101196001110011", "name": "钱丽华", "gender": 2, "age": 66, "dept": "骨科", "blood_type": "B"},
    {"id_card": "310101195201120012", "name": "杨建国", "gender": 1, "age": 74, "dept": "神经内科", "blood_type": "O"},
    {"id_card": "310101195701130013", "name": "黄美玲", "gender": 2, "age": 69, "dept": "普外科", "blood_type": "AB"},
    {"id_card": "310101195301140014", "name": "林志强", "gender": 1, "age": 73, "dept": "心内科", "blood_type": "A"},
    # Community chain (4 persons)
    {"id_card": "310101195001150015", "name": "马德海", "gender": 1, "age": 76, "welfare": ["高龄补贴"]},
    {"id_card": "310101194601160016", "name": "朱秀英", "gender": 2, "age": 80, "welfare": ["特困供养", "高龄补贴"]},
    {"id_card": "310101195401170017", "name": "何国强", "gender": 1, "age": 72, "welfare": ["低保户"]},
    {"id_card": "310101196101180018", "name": "高美华", "gender": 2, "age": 65, "welfare": ["残疾补助"]},
]

MEDICATIONS = [
    ("氨氯地平", "Amlodipine", "5mg", "每日1次", "08:00", "prescription"),
    ("二甲双胍", "Metformin", "500mg", "每日2次", "08:00;20:00", "prescription"),
    ("阿司匹林", "Aspirin", "100mg", "每日1次", "20:00", "otc"),
    ("辛伐他汀", "Simvastatin", "20mg", "每晚1次", "22:00", "prescription"),
]

ALERT_TYPES = ["high_hr", "low_spo2", "fall", "sos", "medication_missed", "abnormal_bp"]
SEVERITIES = ["p0", "p1", "p2"]
ALERT_STATUSES = ["pending", "acknowledged", "resolved"]

def log(msg):
    print(msg, flush=True)

def login():
    resp = requests.post(AUTH_URL, json={
        "method": "email",
        "credential": ROLES["super_admin"]["email"],
        "secret": ROLES["super_admin"]["password"]
    }, timeout=10)
    if resp.status_code == 200:
        token = resp.json()["data"]["token"]
        log("✓ Login successful")
        return token
    log(f"✗ Login failed: {resp.status_code}")
    return None

def api_request(token, method, path, data=None, max_retries=2):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    url = f"{BASE_URL}{path}"
    for attempt in range(max_retries):
        try:
            if method == "GET":
                resp = requests.get(url, headers=headers, timeout=30)
            elif method == "POST":
                resp = requests.post(url, headers=headers, json=data, timeout=30)
            else:
                return None
            if resp.status_code in [200, 201]:
                return resp.json()
            elif resp.status_code == 429:
                log(f"  ⚠ Rate limited, waiting...")
                time.sleep(2)
            else:
                log(f"  ✗ {method} {path} -> {resp.status_code}: {resp.text[:100]}")
                return None
        except Exception as e:
            log(f"  ✗ API error: {e}")
            if attempt < max_retries - 1:
                time.sleep(1)
    return None

def main():
    log("=" * 70)
    log("Eregen API Data Seed (Final)")
    log("=" * 70)

    # Check service
    try:
        resp = requests.get(f"{BASE_URL}/health", timeout=5)
        if resp.status_code != 200:
            log(f"Service not healthy: {resp.status_code}")
            return 1
        log("✓ Service is running")
    except Exception as e:
        log(f"✗ Cannot connect: {e}")
        return 1

    # Login
    log("\n[1] Authenticating...")
    token = login()
    if not token:
        return 1

    # Check existing data
    log("\n[2] Checking existing data...")
    data = api_request(token, "GET", "/admin/persons")
    persons = data.get("data", []) if data else []
    log(f"  Current persons: {len(persons)}")

    person_ids = {}
    for elder in ELDERS:
        existing = None
        for p in persons:
            if p.get("id_card") == elder["id_card"]:
                existing = p
                break
        if existing:
            person_ids[elder["id_card"]] = existing["id"]
        else:
            payload = {
                "id_card": elder["id_card"],
                "name": elder["name"],
                "gender": elder["gender"],
                "birth_date": f"{datetime.now().year - elder['age']}-01-01",
                "phone": f"138{random.randint(10000000, 99999999)}",
                "address": f"上海市浦东新区{random.randint(1, 99)}号",
                "status": "active"
            }
            resp = api_request(token, "POST", "/admin/persons", payload)
            if resp and resp.get("code") == "OK":
                person_ids[elder["id_card"]] = resp.get("data", {}).get("id")
                log(f"  ✓ Created {elder['name']}")
            time.sleep(0.2)

    # Create health records
    log("\n[3] Creating health records...")
    total_health = 0
    for elder in ELDERS:
        if elder.get("tier") not in ("starter", "plus", "pro", "pro_plus"):
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        count = random.randint(30, 50)
        hr_base = 82 if elder.get("risk") in ("high", "critical") else 75
        success = 0
        for i in range(count):
            payload = {
                "person_id": pid,
                "business_chain": "self",
                "record_type": "vitals",
                "source": "device",
                "heart_rate": max(50, min(130, hr_base + random.randint(-15, 15))),
                "spo2": max(88, min(100, 97 + random.randint(-3, 2))),
                "steps": random.randint(500, 12000),
                "sleep_hours": round(random.uniform(5.0, 9.5), 1),
                "blood_pressure_sys": random.randint(100, 180),
                "blood_pressure_dia": random.randint(60, 110)
            }
            resp = api_request(token, "POST", "/admin/health-records", payload, max_retries=1)
            if resp and resp.get("code") == "OK":
                success += 1
            if i % 10 == 9:
                time.sleep(0.3)
        total_health += success
        log(f"  ✓ {elder['name']}: {success}/{count} records")
    log(f"  Total health records: {total_health}")

    # Create medication rules
    log("\n[4] Creating medication rules...")
    total_meds = 0
    for elder in ELDERS:
        if elder.get("tier") not in ("starter", "plus", "pro", "pro_plus") or not elder.get("has_pillbox"):
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for med in random.sample(MEDICATIONS, min(2, len(MEDICATIONS))):
            payload = {
                "person_id": pid,
                "business_chain": "self",
                "drug_name": med[0],
                "generic_name": med[1],
                "dosage": med[2],
                "frequency": med[3],
                "schedule_time1": med[4],
                "schedule_time": med[4],
                "drug_category": med[5],
                "active": True
            }
            resp = api_request(token, "POST", "/admin/medications", payload)
            if resp and resp.get("code") == "OK":
                total_meds += 1
                log(f"  ✓ Added {med[0]} for {elder['name']}")
            time.sleep(0.1)
    log(f"  Total medication rules: {total_meds}")

    # Create alerts
    log("\n[5] Creating alerts...")
    total_alerts = 0
    for elder in ELDERS:
        if elder.get("tier") not in ("starter", "plus", "pro", "pro_plus"):
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for _ in range(random.randint(2, 4)):
            alert_type = random.choice(ALERT_TYPES)
            severity = random.choice(SEVERITIES)
            status = random.choice(ALERT_STATUSES)
            payload = {
                "elderly_id": pid,
                "alert_type": alert_type,
                "severity": severity,
                "device_id": ""
            }
            resp = api_request(token, "POST", "/admin/alerts", payload)
            if resp and resp.get("code") == "OK":
                total_alerts += 1
            time.sleep(0.1)
    log(f"  Total alerts: {total_alerts}")

    # Create hospital patients
    log("\n[6] Creating hospital patients...")
    total_patients = 0
    for elder in ELDERS:
        if "dept" not in elder:
            continue
        patient_id = str(uuid.uuid4())
        payload = {
            "id": patient_id,
            "admission_no": f"H{random.randint(10000, 99999)}",
            "name": elder["name"],
            "gender": "male" if elder["gender"] == 1 else "female",
            "age": elder["age"],
            "department": elder.get("dept", "内科"),
            "blood_type": elder.get("blood_type", "O"),
            "status": "admitted"
        }
        resp = api_request(token, "POST", "/admin/medical/patients", payload)
        if resp and resp.get("code") == "OK":
            total_patients += 1
            log(f"  ✓ Created patient: {elder['name']}")
        time.sleep(0.2)
    log(f"  Total hospital patients: {total_patients}")

    # Create community elders
    log("\n[7] Creating community elders...")
    total_community = 0
    for elder in ELDERS:
        if "welfare" not in elder:
            continue
        payload = {
            "name": elder["name"],
            "id_card": elder["id_card"],
            "gender": elder["gender"],
            "age": elder["age"],
            "hospital_id": "INST-001",
            "status": "active"
        }
        resp = api_request(token, "POST", "/admin/community-wb/elders", payload)
        if resp and resp.get("code") == "OK":
            total_community += 1
            log(f"  ✓ Created community elder: {elder['name']}")
        time.sleep(0.2)
    log(f"  Total community elders: {total_community}")

    # Final counts
    log("\n" + "=" * 70)
    log("✅ DATA SIMULATION COMPLETE")
    log("=" * 70)

    checks = [
        ("/admin/persons", "persons"),
        ("/admin/health-records?chain=self&limit=1", "health_records"),
        ("/admin/medications?chain=self&page=1&page_size=1", "medications"),
        ("/admin/alerts?page=1&page_size=1", "alerts"),
        ("/admin/medical/patients?page=1&page_size=1", "hospital_patients"),
        ("/admin/community-wb/elders?page=1&page_size=1", "community_elders"),
    ]
    for path, name in checks:
        data = api_request(token, "GET", path)
        if data:
            d = data.get("data", [])
            if isinstance(d, list):
                log(f"  {name}: {len(d)}")
            else:
                log(f"  {name}: OK")
        else:
            log(f"  {name}: ERROR")

    log("=" * 70)
    return 0

if __name__ == "__main__":
    import sys
    sys.exit(main())
