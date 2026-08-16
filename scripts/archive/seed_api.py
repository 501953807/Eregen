#!/usr/bin/env python3
"""
Eregen 全业务链数据模拟 - 通过API接口注入
完全符合业务流程，验证系统功能和发现bug
"""
import requests
import uuid
import random
import json
import time
from datetime import datetime, timedelta

BASE_URL = "http://localhost:8089/api/v1"
AUTH_URL = f"{BASE_URL}/auth/login"
ADMIN_URL = f"{BASE_URL}/admin"

# 角色账号密码
ROLES = {
    "super_admin": {"email": "admin@eregen.com", "password": "Admin@123", "name": "系统管理员"},
    "operator": {"email": "op@eregen.com", "password": "Op@123", "name": "自营运营员"},
    "nurse": {"email": "nurse01@eregen.com", "password": "Nurse@123", "name": "住院护士"},
    "community_doctor": {"email": "cd01@eregen.com", "password": "Cd@123", "name": "社区医生"},
    "community_staff": {"email": "cs01@eregen.com", "password": "Cs@123", "name": "社区干事"},
    "regulator": {"email": "reg01@eregen.com", "password": "Reg@123", "name": "监管人员"},
}

# 老人数据定义
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

def log(msg):
    print(msg, flush=True)

def login(role_key):
    """Login and get JWT token"""
    role = ROLES[role_key]
    try:
        resp = requests.post(AUTH_URL, json={
            "method": "email",
            "credential": role["email"],
            "secret": role["password"]
        }, timeout=10)
        if resp.status_code == 200:
            data = resp.json()
            if data.get("code") == 200:
                log(f"  ✓ Login as {role_key}: {role['name']}")
                return data["data"]["token"]
        log(f"  ✗ Login failed for {role_key}: {resp.status_code} - {resp.text}")
        return None
    except Exception as e:
        log(f"  ✗ Login error for {role_key}: {e}")
        return None

def api_request(token, method, path, data=None):
    """Make authenticated API request"""
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    url = f"{BASE_URL}{path}"
    try:
        if method == "GET":
            resp = requests.get(url, headers=headers, timeout=30)
        elif method == "POST":
            resp = requests.post(url, headers=headers, json=data, timeout=30)
        elif method == "PUT":
            resp = requests.put(url, headers=headers, json=data, timeout=30)
        elif method == "DELETE":
            resp = requests.delete(url, headers=headers, timeout=30)
        else:
            return None

        if resp.status_code in [200, 201, 204]:
            return resp.json()
        elif resp.status_code == 429:
            log(f"  ⚠ Rate limited, waiting...")
            time.sleep(2)
            return api_request(token, method, path, data)
        else:
            log(f"  ✗ {method} {path} -> {resp.status_code}: {resp.text[:200]}")
            return None
    except Exception as e:
        log(f"  ✗ API error: {e}")
        return None

def main():
    log("=" * 70)
    log("Eregen Full Chain Data Simulation (API-Based)")
    log("=" * 70)

    # Check service health
    try:
        resp = requests.get(f"{BASE_URL}/health", timeout=5)
        if resp.status_code != 200:
            log(f"Service not healthy: {resp.status_code}")
            return 1
        log("✓ Service is running")
    except Exception as e:
        log(f"✗ Cannot connect to service: {e}")
        return 1

    # Step 1: Login as admin
    log("\n[1] Authenticating as super_admin...")
    token = login("super_admin")
    if not token:
        log("✗ Failed to login. Aborting.")
        return 1

    # Step 2: Check current data
    log("\n[2] Checking current database state...")
    data = api_request(token, "GET", "/admin/persons")
    persons = data.get("data", []) if data else []
    log(f"  Current persons: {len(persons)}")

    data = api_request(token, "GET", "/admin/alerts")
    alerts = data.get("data", []) if data else []
    log(f"  Current alerts: {len(alerts)}")

    # Step 3: Create persons via API
    log("\n[3] Creating persons via API...")
    person_ids = {}
    for elder in ELDERS:
        chain = "self" if "tier" in elder else ("hospital" if "dept" in elder else "community")

        # Check if person exists
        existing = None
        for p in persons:
            if p.get("id_card") == elder["id_card"]:
                existing = p
                break

        if existing:
            person_ids[elder["id_card"]] = existing["id"]
            log(f"  ✓ {elder['name']} exists: {existing['id']}")
            continue

        # Create via API
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
            pid = resp.get("data", {}).get("id", str(uuid.uuid4()))
            person_ids[elder["id_card"]] = pid
            log(f"  ✓ Created {elder['name']} [{chain}]: {pid[:8]}...")
        else:
            log(f"  ✗ Failed to create {elder['name']}: {resp}")

        time.sleep(0.1)  # Rate limit protection

    # Step 4: Create profiles and devices
    log("\n[4] Creating profiles and devices...")
    device_ids = {}

    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue

        chain = "self" if "tier" in elder else ("hospital" if "dept" in elder else "community")

        # Create person profile
        profile_payload = {
            "person_id": pid,
            "business_chain": chain
        }

        if chain == "self":
            profile_payload.update({
                "subscription_tier": elder.get("tier", "starter"),
                "subscription_status": "active",
                "health_risk_level": elder.get("risk", "medium")
            })
        elif chain == "hospital":
            profile_payload.update({
                "admission_no": f"H{random.randint(10000, 99999)}",
                "department": elder.get("dept", "内科"),
                "blood_type": elder.get("blood_type", "O"),
                "status": "in_treatment"
            })
        elif chain == "community":
            profile_payload.update({
                "hospital_id_community": "INST-001",
                "minzheng_certified": 1,
                "status": "active"
            })

        resp = api_request(token, "POST", "/admin/persons/profile", profile_payload)
        if resp and resp.get("code") == "OK":
            log(f"  ✓ Profile created for {elder['name']} [{chain}]")

        # Create devices for self chain
        if chain == "self":
            # Bracelet
            bracelet_payload = {
                "device_id": f"BR-{uuid.uuid4().hex[:4].upper()}",
                "device_type": "bracelet",
                "tier": elder.get("tier", "starter"),
                "status": "active"
            }
            resp = api_request(token, "POST", "/admin/devices", bracelet_payload)
            if resp and resp.get("code") == "OK":
                dev_id = resp.get("data", {}).get("id")
                device_ids[f"bracelet_{elder['id_card']}"] = dev_id
                log(f"  ✓ Created bracelet: {bracelet_payload['device_id']}")

                # Bind device
                bind_payload = {
                    "device_id": dev_id,
                    "person_id": pid,
                    "business_chain": "self",
                    "binding_type": "self"
                }
                resp = api_request(token, "POST", "/admin/device-bindings", bind_payload)
                if resp and resp.get("code") == "OK":
                    log(f"  ✓ Bound bracelet to {elder['name']}")

            # Pillbox
            if elder.get("has_pillbox"):
                pillbox_payload = {
                    "device_id": f"PX-{uuid.uuid4().hex[:4].upper()}",
                    "device_type": "pillbox",
                    "tier": "pro",
                    "status": "active"
                }
                resp = api_request(token, "POST", "/admin/devices", pillbox_payload)
                if resp and resp.get("code") == "OK":
                    dev_id = resp.get("data", {}).get("id")
                    device_ids[f"pillbox_{elder['id_card']}"] = dev_id
                    log(f"  ✓ Created pillbox: {pillbox_payload['device_id']}")

                    # Bind device
                    bind_payload = {
                        "device_id": dev_id,
                        "person_id": pid,
                        "business_chain": "self",
                        "binding_type": "self"
                    }
                    resp = api_request(token, "POST", "/admin/device-bindings", bind_payload)
                    if resp and resp.get("code") == "OK":
                        log(f"  ✓ Bound pillbox to {elder['name']}")

        time.sleep(0.1)

    # Step 5: Create health records
    log("\n[5] Creating health records...")
    total_health = 0

    for elder in ELDERS:
        if "tier" not in elder:  # Skip non-self chain
            continue

        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue

        chain = "self"
        count = random.randint(60, 90)
        hr_base = 82 if elder.get("risk") in ("high", "critical") else 75

        for i in range(count):
            payload = {
                "person_id": pid,
                "business_chain": chain,
                "record_type": "vitals",
                "source": "device",
                "heart_rate": max(50, min(130, hr_base + random.randint(-15, 15))),
                "spo2": max(88, min(100, 97 + random.randint(-3, 2))),
                "steps": random.randint(500, 12000),
                "sleep_hours": round(random.uniform(5.0, 9.5), 1),
                "blood_pressure_sys": random.randint(100, 180),
                "blood_pressure_dia": random.randint(60, 110),
                "recorded_at": (datetime.now() - timedelta(hours=random.randint(0, 48))).isoformat()
            }

            resp = api_request(token, "POST", "/admin/health-records", payload)
            if resp and resp.get("code") == "OK":
                total_health += 1
            elif resp and resp.get("code") != "OK" and "already exists" not in resp.get("error", "").lower():
                log(f"    ! Health record failed: {resp.get('error')}")

            # Rate limit protection
            if i % 10 == 9:
                time.sleep(0.5)

        log(f"  ✓ {elder['name']}: {count} health records")

    log(f"  Total health records created: {total_health}")

    # Step 6: Create medication rules
    log("\n[6] Creating medication rules...")
    meds = [
        ("氨氯地平", "Amlodipine", "5mg", "每日1次", "08:00", "prescription"),
        ("二甲双胍", "Metformin", "500mg", "每日2次", "08:00;20:00", "prescription"),
        ("阿司匹林", "Aspirin", "100mg", "每日1次", "20:00", "otc"),
    ]

    total_meds = 0
    for elder in ELDERS:
        if elder.get("tier") != "self" or not elder.get("has_pillbox"):
            continue

        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue

        for med in random.sample(meds, min(2, len(meds))):
            payload = {
                "person_id": pid,
                "business_chain": "self",
                "drug_name": med[0],
                "generic_name": med[1],
                "dosage": med[2],
                "frequency": med[3],
                "schedule_time": med[4],  # Required NOT NULL field
                "schedule_time1": med[4],
                "drug_category": med[5],
                "active": True
            }

            resp = api_request(token, "POST", "/admin/medications", payload)
            if resp and resp.get("code") == "OK":
                total_meds += 1
                log(f"  ✓ Added {med[0]} for {elder['name']}")
            time.sleep(0.1)

    log(f"  Total medication rules: {total_meds}")

    # Step 7: Create hospital patients and admissions
    log("\n[7] Creating hospital data...")
    total_verifications = 0
    total_daily = 0
    total_expenses = 0

    for elder in ELDERS:
        if "dept" not in elder:  # Skip non-hospital
            continue

        # Create medical patient
        patient_payload = {
            "admission_no": f"H{random.randint(10000, 99999)}",
            "name": elder["name"],
            "gender": "male" if elder["gender"] == 1 else "female",
            "age": elder["age"],
            "department": elder.get("dept", "内科"),
            "blood_type": elder.get("blood_type", "O"),
            "status": "admitted"
        }

        resp = api_request(token, "POST", "/admin/medical/patients", patient_payload)
        if resp and resp.get("code") == "OK":
            patient_id = resp.get("data", {}).get("id")
            log(f"  ✓ Created patient: {elder['name']}")
        else:
            log(f"  ✗ Failed to create patient: {elder['name']}")
            continue

        # Create admission
        admission_payload = {
            "patient_id": patient_id,
            "bed_no": f"{random.randint(1, 30)}床",
            "department": elder.get("dept", "内科"),
            "diagnosis": "待诊断",
            "expected_stay_days": random.randint(3, 14)
        }

        resp = api_request(token, "POST", "/admin/medical/admissions", admission_payload)
        if resp and resp.get("code") == "OK":
            log(f"  ✓ Created admission for {elder['name']}")
        time.sleep(0.1)

        # Create daily entries
        for _ in range(random.randint(3, 5)):
            payload = {
                "patient_id": patient_id,
                "entry_date": (datetime.now() - timedelta(days=random.randint(0, 14))).strftime("%Y-%m-%d"),
                "entry_type": random.choice(["vitals", "medication", "nursing"]),
                "content": "查房记录: 生命体征稳定"
            }
            resp = api_request(token, "POST", "/admin/medical/daily-entries", payload)
            if resp and resp.get("code") == "OK":
                total_daily += 1
            time.sleep(0.05)

        # Create verifications
        for _ in range(random.randint(5, 10)):
            payload = {
                "patient_id": patient_id,
                "verification_type": random.choice(["medication", "vitals", "nfc"]),
                "result": "passed",
                "matched": True,
                "verified_by": "nurse01"
            }
            resp = api_request(token, "POST", "/admin/medical/verifications", payload)
            if resp and resp.get("code") == "OK":
                total_verifications += 1
            time.sleep(0.05)

        # Create expenses
        for _ in range(random.randint(5, 15)):
            payload = {
                "patient_id": patient_id,
                "item_name": random.choice(["血常规", "CT扫描", "心电图", "B超"]),
                "category": random.choice(["lab", "radiology", "consultation"]),
                "amount": round(random.uniform(50, 800), 2),
                "quantity": 1,
                "billing_source": "manual"
            }
            resp = api_request(token, "POST", "/admin/medical/expenses", payload)
            if resp and resp.get("code") == "OK":
                total_expenses += 1
            time.sleep(0.05)

    log(f"  Hospital data: verifications={total_verifications}, daily={total_daily}, expenses={total_expenses}")

    # Step 8: Create community elders and sign-ins
    log("\n[8] Creating community data...")
    total_signins = 0

    for elder in ELDERS:
        if "welfare" not in elder:  # Skip non-community
            continue

        # Create community elder
        elder_payload = {
            "name": elder["name"],
            "id_card": elder["id_card"],
            "gender": elder["gender"],
            "age": elder["age"],
            "address": f"上海市浦东新区{random.randint(1, 99)}号",
            "hospital_id": "INST-001",
            "status": "active"
        }

        resp = api_request(token, "POST", "/admin/community-wb/elders", elder_payload)
        if resp and resp.get("code") == "OK":
            elder_id = resp.get("data", {}).get("id")
            log(f"  ✓ Created community elder: {elder['name']}")
        else:
            log(f"  ✗ Failed to create community elder: {elder['name']}")
            continue

        # Create sign-in records
        for day in range(14):
            for period in ["morning", "afternoon"]:
                payload = {
                    "elder_id": elder_id,
                    "device_id": f"CW-{uuid.uuid4().hex[:4].upper()}",
                    "hospital_id": "INST-001",
                    "signin_time": (datetime.now() - timedelta(days=day, hours=random.randint(0, 12))).isoformat(),
                    "period": period,
                    "is_medical_signin": True,
                    "is_welfare_signin": True,
                    "activated_tags": "[]"
                }
                resp = api_request(token, "POST", "/admin/community-wb/signin/trigger", payload)
                if resp and resp.get("code") == "OK":
                    total_signins += 1
                time.sleep(0.05)

    log(f"  Total community sign-ins: {total_signins}")

    # Step 9: Create alerts
    log("\n[9] Creating alerts...")
    total_alerts = 0

    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue

        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue

        # Create alerts based on risk level
        alert_count = random.randint(2, 5)
        for _ in range(alert_count):
            alert_type = random.choice(["high_hr", "low_spo2", "fall", "sos", "medication_missed"])
            severity = random.choice(["p0", "p1", "p2"])
            status = random.choice(["pending", "acknowledged", "resolved"])

            payload = {
                "person_id": pid,
                "business_chain": "self",
                "alert_type": alert_type,
                "severity": severity,
                "status": status,
                "message": f"检测到{alert_type.replace('_', ' ')}异常",
                "device_id": device_ids.get(f"bracelet_{elder['id_card']}", "")
            }

            resp = api_request(token, "POST", "/admin/alerts", payload)
            if resp and resp.get("code") == "OK":
                total_alerts += 1
            time.sleep(0.1)

    log(f"  Total alerts created: {total_alerts}")

    # Step 10: Summary
    log("\n" + "=" * 70)
    log("✅ DATA SIMULATION COMPLETE")
    log("=" * 70)

    # Get final counts
    data = api_request(token, "GET", "/admin/persons")
    persons = data.get("data", []) if data else []
    log(f"  Persons: {len(persons)}")

    data = api_request(token, "GET", "/admin/alerts")
    alerts = data.get("data", []) if data else []
    log(f"  Alerts: {len(alerts)}")

    data = api_request(token, "GET", "/admin/medical/patients")
    patients = data.get("data", []) if data else []
    log(f"  Hospital patients: {len(patients)}")

    data = api_request(token, "GET", "/admin/community-wb/elders")
    elders = data.get("data", []) if data else []
    log(f"  Community elders: {len(elders)}")

    log("=" * 70)
    return 0

if __name__ == "__main__":
    import sys
    sys.exit(main())
