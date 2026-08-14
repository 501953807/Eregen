#!/usr/bin/env python3
"""
Eregen 全业务链数据模拟 - 混合方案
优先使用API，失败时降级到直接数据库写入
"""
import sqlite3
import uuid
import random
import json
import time
import requests
from datetime import datetime, timedelta

DB = "/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/admin-api/data/eregen.db"
BASE_URL = "http://localhost:8089/api/v1"

# 角色账号密码
ROLES = {
    "super_admin": {"email": "admin@eregen.com", "password": "Admin@123"},
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

def try_api_seed():
    """尝试通过API注入数据"""
    log("尝试通过API注入数据...")

    # Login
    try:
        resp = requests.post(f"{BASE_URL}/auth/login", json={
            "method": "email",
            "credential": ROLES["super_admin"]["email"],
            "secret": ROLES["super_admin"]["password"]
        }, timeout=10)
        if resp.status_code != 200:
            log(f"  ✗ Login failed: {resp.status_code}")
            return False
        token = resp.json()["data"]["token"]
        log("  ✓ Login successful")
    except Exception as e:
        log(f"  ✗ Login error: {e}")
        return False

    # Test API endpoints
    try:
        resp = requests.get(f"{BASE_URL}/admin/persons", headers={"Authorization": f"Bearer {token}"}, timeout=10)
        if resp.status_code != 200:
            log(f"  ✗ API test failed: {resp.status_code}")
            return False
        log("  ✓ API endpoints accessible")
    except Exception as e:
        log(f"  ✗ API test error: {e}")
        return False

    # Try to create health records via API
    log("  Testing health records API...")
    resp = requests.post(f"{BASE_URL}/admin/health-records",
        headers={"Authorization": f"Bearer {token}"},
        json={"person_id": "test", "business_chain": "self", "record_type": "vitals", "heart_rate": 75},
        timeout=10)
    log(f"  Health records API response: {resp.status_code} - {resp.text[:100]}")

    return True

def db_seed():
    """通过直接数据库写入注入数据"""
    log("\n通过直接数据库写入注入数据...")

    conn = sqlite3.connect(DB)
    conn.execute("PRAGMA foreign_keys = ON")
    cur = conn.cursor()

    # 1. Create persons
    log("\n[1] Creating persons...")
    person_ids = {}
    for elder in ELDERS:
        cur.execute("SELECT id FROM persons WHERE id_card=?", (elder["id_card"],))
        if cur.fetchone():
            log(f"  ✓ {elder['name']} exists")
            continue
        pid = str(uuid.uuid4())
        try:
            cur.execute("INSERT INTO persons (id, id_card, name, gender, birth_date, phone, address, status) VALUES (?,?,?,?,?,?,?,?)",
                       (pid, elder["id_card"], elder["name"], elder["gender"],
                        f"{datetime.now().year - elder['age']}-01-01",
                        f"138{random.randint(10000000, 99999999)}",
                        f"上海市浦东新区{random.randint(1, 99)}号", "active"))
            person_ids[elder["id_card"]] = pid
            log(f"  ✓ Created {elder['name']}")
        except Exception as e:
            log(f"  ✗ Failed to create {elder['name']}: {e}")
    conn.commit()

    # 2. Create profiles
    log("\n[2] Creating profiles...")
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = "self" if "tier" in elder else ("hospital" if "dept" in elder else "community")
        cur.execute("SELECT id FROM person_profiles WHERE person_id=? AND business_chain=?", (pid, chain))
        if cur.fetchone():
            log(f"  ✓ Profile exists for {elder['name']} [{chain}]")
            continue
        try:
            if chain == "self":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, subscription_tier, subscription_status, health_risk_level, status) VALUES (?,?,?,?,?,?)",
                           (pid, chain, elder.get("tier", "starter"), "active", elder.get("risk", "medium"), "active"))
            elif chain == "hospital":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, admission_no, department, blood_type, status) VALUES (?,?,?,?,?,?)",
                           (pid, chain, f"H{random.randint(10000, 99999)}", elder.get("dept", "内科"), elder.get("blood_type", "O"), "in_treatment"))
            elif chain == "community":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, hospital_id_community, minzheng_certified, status) VALUES (?,?,?,?,?)",
                           (pid, chain, "INST-001", 1, "active"))
            log(f"  ✓ Profile created for {elder['name']} [{chain}]")
        except Exception as e:
            log(f"  ✗ Failed to create profile for {elder['name']}: {e}")
    conn.commit()

    # 3. Create devices and bindings
    log("\n[3] Creating devices and bindings...")
    device_ids = {}
    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        try:
            # Bracelet
            dev_id = f"BR-{uuid.uuid4().hex[:4].upper()}"
            cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?,?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, "bracelet", elder.get("tier", "starter"), "active"))
            device_ids[f"bracelet_{elder['id_card']}"] = dev_id
            cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain, binding_type) VALUES (?,?,?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, pid, "self", "self"))
            log(f"  ✓ Created and bound bracelet for {elder['name']}")

            # Pillbox
            if elder.get("has_pillbox"):
                pill_id = f"PX-{uuid.uuid4().hex[:4].upper()}"
                cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?,?,?,?,?)",
                           (str(uuid.uuid4()), pill_id, "pillbox", "pro", "active"))
                device_ids[f"pillbox_{elder['id_card']}"] = pill_id
                cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain, binding_type) VALUES (?,?,?,?,?,?)",
                           (str(uuid.uuid4()), pill_id, pid, "self", "self"))
                log(f"  ✓ Created and bound pillbox for {elder['name']}")
        except Exception as e:
            log(f"  ✗ Failed to create device for {elder['name']}: {e}")
    conn.commit()

    # 4. Create health records (bulk insert)
    log("\n[4] Creating health records...")
    total_health = 0
    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = "self"
        count = random.randint(60, 90)
        hr_base = 82 if elder.get("risk") in ("high", "critical") else 75

        for i in range(count):
            try:
                cur.execute("""
                    INSERT INTO health_records (id, person_id, business_chain, record_type, source,
                                               heart_rate, spo2, steps, sleep_hours,
                                               blood_pressure_sys, blood_pressure_dia, recorded_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), pid, chain, "vitals", "device",
                    max(50, min(130, hr_base + random.randint(-15, 15))),
                    max(88, min(100, 97 + random.randint(-3, 2))),
                    random.randint(500, 12000),
                    round(random.uniform(5.0, 9.5), 1),
                    random.randint(100, 180),
                    random.randint(60, 110),
                    (datetime.now() - timedelta(hours=random.randint(0, 48))).isoformat()
                ))
                total_health += 1
            except:
                pass
        log(f"  ✓ {elder['name']}: {count} health records")
    conn.commit()
    log(f"  Total health records: {total_health}")

    # 5. Create medication rules
    log("\n[5] Creating medication rules...")
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
            try:
                cur.execute("""
                    INSERT INTO medication_rules (id, person_id, business_chain, drug_name, generic_name,
                                                  dosage, frequency, schedule_time, schedule_time1,
                                                  drug_category, active)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), pid, "self", med[0], med[1], med[2], med[3], med[4], med[4],
                    med[5], 1
                ))
                total_meds += 1
                log(f"  ✓ Added {med[0]} for {elder['name']}")
            except Exception as e:
                log(f"  ✗ Failed to add medication: {e}")
    conn.commit()
    log(f"  Total medication rules: {total_meds}")

    # 6. Create hospital data
    log("\n[6] Creating hospital data...")
    total_verifications = 0
    total_daily = 0
    total_expenses = 0

    for elder in ELDERS:
        if "dept" not in elder:
            continue

        # Create medical patient
        patient_id = str(uuid.uuid4())
        admission_no = f"H{random.randint(10000, 99999)}"
        try:
            cur.execute("""
                INSERT INTO medical_wristband_patients (id, admission_no, name, gender, age, department, blood_type, status)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (patient_id, admission_no, elder["name"],
                  "male" if elder["gender"] == 1 else "female",
                  elder["age"], elder.get("dept", "内科"),
                  elder.get("blood_type", "O"), "admitted"))
            log(f"  ✓ Created patient: {elder['name']}")
        except Exception as e:
            log(f"  ✗ Failed to create patient: {elder['name']}: {e}")
            continue

        # Create admission
        try:
            cur.execute("""
                INSERT INTO hospital_admissions (id, patient_id, bed_no, department, diagnosis, expected_discharge_at)
                VALUES (?, ?, ?, ?, ?, ?)
            """, (
                str(uuid.uuid4()), patient_id,
                f"{random.randint(1, 30)}床", elder.get("dept", "内科"),
                "待诊断",
                (datetime.now() + timedelta(days=random.randint(3, 14))).isoformat()
            ))
            log(f"  ✓ Created admission for {elder['name']}")
        except Exception as e:
            log(f"  ✗ Failed to create admission: {e}")

        # Create daily entries
        for _ in range(random.randint(3, 5)):
            try:
                cur.execute("""
                    INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content)
                    VALUES (?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), patient_id,
                    (datetime.now() - timedelta(days=random.randint(0, 14))).strftime("%Y-%m-%d"),
                    random.choice(["vitals", "medication", "nursing"]),
                    "查房记录: 生命体征稳定"
                ))
                total_daily += 1
            except:
                pass

        # Create verifications
        for _ in range(random.randint(5, 10)):
            try:
                cur.execute("""
                    INSERT INTO medical_verifications (id, device_id, patient_id, verification_type, result, matched, verified_by)
                    VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), "", patient_id,
                    random.choice(["medication", "vitals", "nfc"]),
                    "passed", 1, "nurse01"
                ))
                total_verifications += 1
            except:
                pass

        # Create expenses
        for _ in range(random.randint(5, 15)):
            try:
                cur.execute("""
                    INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, billing_source)
                    VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), patient_id,
                    random.choice(["血常规", "CT扫描", "心电图", "B超"]),
                    random.choice(["lab", "radiology", "consultation"]),
                    round(random.uniform(50, 800), 2), 1, "manual"
                ))
                total_expenses += 1
            except:
                pass

    log(f"  Hospital data: verifications={total_verifications}, daily={total_daily}, expenses={total_expenses}")

    # 7. Create community data
    log("\n[7] Creating community data...")
    total_signins = 0

    for elder in ELDERS:
        if "welfare" not in elder:
            continue

        # Create community elder
        elder_id = str(uuid.uuid4())
        try:
            cur.execute("""
                INSERT INTO community_elders (id, name, id_card, gender, age, address, hospital_id, status)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                elder_id, elder["name"], elder["id_card"], elder["gender"],
                elder["age"], f"上海市浦东新区{random.randint(1, 99)}号",
                "INST-001", "active"
            ))
            log(f"  ✓ Created community elder: {elder['name']}")
        except Exception as e:
            log(f"  ✗ Failed to create community elder: {elder['name']}: {e}")
            continue

        # Create sign-in records
        for day in range(14):
            for period in ["morning", "afternoon"]:
                try:
                    cur.execute("""
                        INSERT OR IGNORE INTO community_signin_records (id, elder_id, device_id, hospital_id, signin_time, period, is_medical_signin, is_welfare_signin)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                    """, (
                        str(uuid.uuid4()), elder_id,
                        f"CW-{uuid.uuid4().hex[:4].upper()}",
                        "INST-001",
                        (datetime.now() - timedelta(days=day, hours=random.randint(0, 12))).isoformat(),
                        period, 1, 1
                    ))
                    total_signins += 1
                except:
                    pass

    log(f"  Total community sign-ins: {total_signins}")

    # 8. Create alerts
    log("\n[8] Creating alerts...")
    total_alerts = 0

    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue

        for _ in range(random.randint(2, 4)):
            try:
                alert_type = random.choice(["high_hr", "low_spo2", "fall", "sos", "medication_missed"])
                severity = random.choice(["p0", "p1", "p2"])
                status = random.choice(["pending", "acknowledged", "resolved"])
                cur.execute("""
                    INSERT INTO alerts (id, person_id, business_chain, alert_type, severity, status, message, device_id)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), pid, "self", alert_type, severity, status,
                    f"检测到{alert_type.replace('_', ' ')}异常",
                    device_ids.get(f"bracelet_{elder['id_card']}", "")
                ))
                total_alerts += 1
            except:
                pass

    log(f"  Total alerts: {total_alerts}")
    conn.commit()

    # Summary
    log("\n" + "=" * 70)
    log("✅ SEED COMPLETE")
    log("=" * 70)

    cur.execute("SELECT COUNT(*) FROM persons")
    log(f"  Persons: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM person_profiles")
    log(f"  Profiles: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM devices")
    log(f"  Devices: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM device_bindings")
    log(f"  Device bindings: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM health_records")
    log(f"  Health records: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medication_rules")
    log(f"  Medication rules: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM alerts")
    log(f"  Alerts: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_wristband_patients")
    log(f"  Hospital patients: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM community_elders")
    log(f"  Community elders: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM community_signin_records")
    log(f"  Community sign-ins: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_verifications")
    log(f"  Medical verifications: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_daily_entries")
    log(f"  Medical daily entries: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_expenses")
    log(f"  Medical expenses: {cur.fetchone()[0]}")

    log("=" * 70)
    conn.close()
    return 0

if __name__ == "__main__":
    import sys
    # Try API first, fallback to direct DB
    if try_api_seed():
        log("\nAPI seed successful, proceeding with DB seed for bulk data...")
    sys.exit(db_seed())
