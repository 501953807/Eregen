#!/usr/bin/env python3
"""
Eregen 全业务链数据注入脚本 - 直接数据库版本
============================================
绕过API限速，直接写入SQLite数据库
"""
import sqlite3
import uuid
import random
import hashlib
from datetime import datetime, timedelta

DB_PATH = "/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/admin-api/data/eregen.db"

# 老人数据
ELDERS = [
    # Self chain
    {"name": "张建国", "gender": 1, "age": 72, "chain": "self", "id_card": "310101195401010001", "tier": "pro", "risk": "high", "conditions": ["hypertension", "diabetes"], "has_pillbox": True},
    {"name": "李秀芳", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801020002", "tier": "plus", "risk": "medium", "conditions": ["hypertension"], "has_pillbox": True},
    {"name": "王德明", "gender": 1, "age": 75, "chain": "self", "id_card": "310101195101030003", "tier": "starter", "risk": "low", "conditions": [], "has_pillbox": False},
    {"name": "赵美华", "gender": 2, "age": 70, "chain": "self", "id_card": "310101195601040004", "tier": "pro_plus", "risk": "critical", "conditions": ["diabetes", "coronary_heart_disease"], "has_pillbox": True},
    {"name": "陈志强", "gender": 1, "age": 65, "chain": "self", "id_card": "310101196101050005", "tier": "plus", "risk": "medium", "conditions": ["hypertension"], "has_pillbox": False},
    {"name": "刘淑珍", "gender": 2, "age": 78, "chain": "self", "id_card": "310101194801060006", "tier": "pro", "risk": "high", "conditions": ["osteoporosis"], "has_pillbox": False},
    {"name": "孙伟民", "gender": 1, "age": 63, "chain": "self", "id_card": "310101196301070007", "tier": "starter", "risk": "low", "conditions": [], "has_pillbox": False},
    # Cross-chain
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "self", "id_card": "310101195601080008", "tier": "pro", "risk": "high", "conditions": ["hypertension", "diabetes"], "has_pillbox": True},
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801090009", "tier": "pro_plus", "risk": "critical", "conditions": ["diabetes", "chronic_kidney_disease"], "has_pillbox": True},
    # Hospital chain
    {"name": "郑国华", "gender": 1, "age": 71, "chain": "hospital", "id_card": "310101195501100010", "dept": "呼吸科", "blood_type": "A", "conditions": ["pneumonia"]},
    {"name": "钱丽华", "gender": 2, "age": 66, "chain": "hospital", "id_card": "310101196001110011", "dept": "骨科", "blood_type": "B", "conditions": ["hip_fracture"]},
    {"name": "杨建国", "gender": 1, "age": 74, "chain": "hospital", "id_card": "310101195201120012", "dept": "神经内科", "blood_type": "O", "conditions": ["stroke_recovery"]},
    {"name": "黄美玲", "gender": 2, "age": 69, "chain": "hospital", "id_card": "310101195701130013", "dept": "普外科", "blood_type": "AB", "conditions": ["appendicitis"]},
    {"name": "林志强", "gender": 1, "age": 73, "chain": "hospital", "id_card": "310101195301140014", "dept": "心内科", "blood_type": "A", "conditions": ["heart_failure"]},
    # Community chain
    {"name": "马德海", "gender": 1, "age": 76, "chain": "community", "id_card": "310101195001150015", "welfare_tags": ["高龄补贴"]},
    {"name": "朱秀英", "gender": 2, "age": 80, "chain": "community", "id_card": "310101194601160016", "welfare_tags": ["特困供养", "高龄补贴"]},
    {"name": "何国强", "gender": 1, "age": 72, "chain": "community", "id_card": "310101195401170017", "welfare_tags": ["低保户"]},
    {"name": "高美华", "gender": 2, "age": 65, "chain": "community", "id_card": "310101196101180018", "welfare_tags": ["残疾补助"]},
]

def hash_password(pw):
    return hashlib.sha256(pw.encode()).hexdigest()

def now():
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def rand_date(days_back=90):
    return (datetime.now() - timedelta(days=random.randint(0, days_back))).strftime("%Y-%m-%d")

def rand_datetime(hours_back=48):
    return (datetime.now() - timedelta(hours=random.randint(0, hours_back))).strftime("%Y-%m-%d %H:%M:%S")

def uid():
    return uuid.uuid4().hex[:16]

def main():
    conn = sqlite3.connect(DB_PATH)
    conn.execute("PRAGMA foreign_keys = ON")
    cur = conn.cursor()

    print("=" * 60)
    print("Eregen Direct Database Seed")
    print("=" * 60)

    # 1. Create role accounts
    print("\n[1] Creating role accounts...")
    roles = [
        ("admin@eregen.com", "13800000001", "super_admin", "Admin@123"),
        ("op@eregen.com", "13800000002", "operator", "Op@123"),
        ("nurse01@eregen.com", "13800000003", "nurse", "Nurse@123"),
        ("cd01@eregen.com", "13800000004", "community_doctor", "Cd@123"),
        ("cs01@eregen.com", "13800000005", "community_staff", "Cs@123"),
        ("reg01@eregen.com", "13800000006", "regulator", "Reg@123"),
    ]
    for email, phone, role, pw in roles:
        try:
            cur.execute("SELECT id FROM users WHERE email=?", (email,))
            if cur.fetchone():
                print(f"  ✅ {email} already exists")
                continue
            uid_val = str(uuid.uuid4())
            pw_hash = hash_password(pw)
            cur.execute("INSERT INTO users (id, name, email, phone, role, password_hash) VALUES (?, ?, ?, ?, ?, ?)",
                       (uid_val, email.split('@')[0], email, phone, role, pw_hash))
            print(f"  ✅ Created {role}: {email}")
        except Exception as e:
            print(f"  ❌ {email}: {e}")
    conn.commit()

    # 2. Create persons
    print("\n[2] Creating persons...")
    person_ids = {}
    for elder in ELDERS:
        try:
            cur.execute("SELECT id FROM persons WHERE id_card=?", (elder["id_card"],))
            if cur.fetchone():
                print(f"  ✅ {elder['name']} already exists")
                continue
            pid = str(uuid.uuid4())
            birth_year = datetime.now().year - elder["age"]
            cur.execute("INSERT INTO persons (id, id_card, name, gender, birth_date, phone, address, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                       (pid, elder["id_card"], elder["name"], elder["gender"],
                        f"{birth_year}-01-01", f"138{random.randint(10000000, 99999999)}",
                        f"上海市浦东新区{random.randint(1,99)}号", "active"))
            person_ids[elder["id_card"]] = pid
            print(f"  ✅ {elder['name']} [{elder['chain']}]: {pid[:8]}...")
        except Exception as e:
            print(f"  ❌ {elder['name']}: {e}")
    conn.commit()

    # 3. Create person profiles
    print("\n[3] Creating person profiles...")
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]
        try:
            # Check if profile exists
            cur.execute("SELECT id FROM person_profiles WHERE person_id=? AND business_chain=?", (pid, chain))
            if cur.fetchone():
                print(f"  ✅ Profile {elder['name']} [{chain}] exists")
                continue

            if chain == "self":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, subscription_tier, subscription_status, health_risk_level) VALUES (?, ?, ?, ?, ?)",
                           (pid, chain, elder["tier"], "active", elder["risk"]))
            elif chain == "hospital":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, admission_no, department, bed_number, blood_type, diagnosis, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                           (pid, chain, f"H{random.randint(10000,99999)}", elder.get("dept","内科"), f"{random.randint(1,30)}床",
                            elder.get("blood_type","O"), elder["conditions"][0] if elder.get("conditions") else "待诊断", "in_treatment"))
            elif chain == "community":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, hospital_id_community, minzheng_certified, subsidy_type, status) VALUES (?, ?, ?, ?, ?, ?)",
                           (pid, chain, "INST-001", 1, "定期补助", "active"))
            print(f"  ✅ Profile {elder['name']} [{chain}]")
        except Exception as e:
            print(f"  ❌ Profile {elder['name']} [{chain}]: {e}")
    conn.commit()

    # 4. Create devices
    print("\n[4] Creating devices...")
    device_ids = {}
    for elder in ELDERS:
        if elder["chain"] != "self":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        try:
            # Bracelet
            dev_id = f"BR-{uid()[:4].upper()}"
            cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?, ?, ?, ?, ?)",
                       (str(uuid.uuid4()), dev_id, "bracelet", elder.get("tier","starter"), "active"))
            device_ids[f"bracelet_{elder['id_card']}"] = dev_id
            print(f"  ✅ Bracelet {dev_id}")
            # Bind
            cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain) VALUES (?, ?, ?, ?)",
                       (str(uuid.uuid4()), dev_id, pid, "self"))
            # Pillbox
            if elder.get("has_pillbox"):
                pill_id = f"PX-{uid()[:4].upper()}"
                cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pill_id, "pillbox", "pro", "active"))
                device_ids[f"pillbox_{elder['id_card']}"] = pill_id
                print(f"  ✅ Pillbox {pill_id}")
                cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain) VALUES (?, ?, ?, ?)",
                           (str(uuid.uuid4()), pill_id, pid, "self"))
        except Exception as e:
            print(f"  ❌ Device for {elder['name']}: {e}")
    conn.commit()

    # 5. Medical wristbands
    print("\n[5] Creating medical wristbands...")
    for elder in ELDERS:
        if elder["chain"] != "hospital":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        try:
            dev_id = f"MW-{uid()[:4].upper()}"
            cur.execute("INSERT INTO medical_wristband_devices (id, device_id, firmware_version, status) VALUES (?, ?, ?, ?)",
                       (str(uuid.uuid4()), dev_id, "v1.2.0", "active"))
            device_ids[f"medical_{elder['id_card']}"] = dev_id
            print(f"  ✅ Medical wristband {dev_id}")
        except Exception as e:
            print(f"  ❌ Medical wristband {elder['name']}: {e}")
    conn.commit()

    # 6. Community wristbands
    print("\n[6] Creating community wristbands...")
    for elder in ELDERS:
        if elder["chain"] != "community":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        try:
            dev_id = f"CW-{uid()[:4].upper()}"
            cur.execute("INSERT INTO community_wristband_devices (id, device_id, firmware_version, mode, status) VALUES (?, ?, ?, ?, ?)",
                       (str(uuid.uuid4()), dev_id, "v1.0.0", "community", "active"))
            device_ids[f"community_{elder['id_card']}"] = dev_id
            print(f"  ✅ Community wristband {dev_id}")
        except Exception as e:
            print(f"  ❌ Community wristband {elder['name']}: {e}")
    conn.commit()

    # 7. Health records
    print("\n[7] Injecting health records...")
    total_records = 0
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]
        count = random.randint(55, 95)
        hr_base = 75 if elder.get("risk") not in ("high", "critical") else 82
        for j in range(count):
            try:
                hr = max(50, min(130, hr_base + random.randint(-15, 15)))
                spo2 = max(88, min(100, spo2_base + random.randint(-3, 2)))
                cur.execute("INSERT INTO health_records (id, person_id, business_chain, record_type, source, hr, spo2, steps, sleep_hours, blood_pressure_sys, blood_pressure_dia, recorded_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pid, chain, "vitals", "device",
                            hr, spo2, random.randint(500, 12000), round(random.uniform(5.0, 9.5), 1),
                            random.randint(100, 180), random.randint(60, 110), rand_datetime(48)))
                total_records += 1
            except:
                pass
        print(f"  ✅ {elder['name']}: {count} records")
    conn.commit()
    print(f"  Total health records: {total_records}")

    # 8. Medication rules
    print("\n[8] Injecting medication rules...")
    meds = [
        ("氨氯地平", "Amlodipine", "5mg", "每日1次", "08:00", "prescription"),
        ("二甲双胍", "Metformin", "500mg", "每日2次", "08:00;20:00", "prescription"),
        ("阿司匹林", "Aspirin", "100mg", "每日1次", "20:00", "otc"),
    ]
    pillbox_elders = [e for e in ELDERS if e.get("has_pillbox") and e["chain"] == "self"]
    total_meds = 0
    for elder in pillbox_elders:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for med in random.sample(meds, min(2, len(meds))):
            try:
                cur.execute("INSERT INTO medication_rules (id, person_id, business_chain, drug_name, dosage, frequency, schedule_time1, drug_category, active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pid, "self", med[0], med[2], med[3], med[4], med[5], 1))
                total_meds += 1
            except:
                pass
    conn.commit()
    print(f"  Total medication rules: {total_meds}")

    # 9. Medical data
    print("\n[9] Injecting hospital data...")
    total_verifications = 0
    total_daily = 0
    total_expenses = 0
    for elder in ELDERS:
        if elder["chain"] != "hospital":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        # Daily entries
        for j in range(random.randint(3, 5)):
            try:
                cur.execute("INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content) VALUES (?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pid, rand_date(14), random.choice(["vitals", "medication", "nursing"]),
                            f"查房记录: 生命体征稳定"))
                total_daily += 1
            except:
                pass
        # Verifications
        mw_id = device_ids.get(f"medical_{elder['id_card']}")
        if mw_id:
            for j in range(random.randint(5, 10)):
                try:
                    cur.execute("INSERT INTO medical_verifications (id, device_id, patient_id, verification_type, result, matched, verified_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
                               (str(uuid.uuid4()), mw_id, pid, random.choice(["medication", "vitals", "nfc"]), "passed", 1, "nurse01"))
                    total_verifications += 1
                except:
                    pass
        # Expenses
        for j in range(random.randint(5, 15)):
            try:
                cur.execute("INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, billing_source) VALUES (?, ?, ?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pid, random.choice(["血常规", "CT扫描", "心电图", "B超"]),
                            random.choice(["lab", "radiology", "consultation"]), round(random.uniform(50, 800), 2), 1, "manual"))
                total_expenses += 1
            except:
                pass
        print(f"  ✅ {elder['name']}: rounds={total_daily}, verifies={total_verifications}, expenses={total_expenses}")
    conn.commit()

    # 10. Community data
    print("\n[10] Injecting community data...")
    total_signins = 0
    for elder in ELDERS:
        if elder["chain"] != "community":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        # Sign-ins
        for day in range(14):
            try:
                cur.execute("INSERT INTO community_signin_records (id, elder_id, device_id, hospital_id, signin_time, period, is_medical_signin, is_welfare_signin) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                           (str(uuid.uuid4()), pid, device_ids.get(f"community_{elder['id_card']}",""), "INST-001",
                            rand_datetime(336), random.choice(["morning", "afternoon"]), 1, 1))
                total_signins += 1
            except:
                pass
        print(f"  ✅ {elder['name']}: {total_signins} signins")
    conn.commit()

    # Summary
    print("\n" + "=" * 60)
    print("✅ SEED COMPLETE")
    print(f"  Persons: {len(person_ids)}")
    print(f"  Devices: {len(device_ids)}")
    print(f"  Health records: {total_records}")
    print(f"  Medication rules: {total_meds}")
    print(f"  Medical verifications: {total_verifications}")
    print(f"  Medical daily entries: {total_daily}")
    print(f"  Medical expenses: {total_expenses}")
    print(f"  Community signins: {total_signins}")
    print("=" * 60)

    conn.close()

if __name__ == "__main__":
    main()
