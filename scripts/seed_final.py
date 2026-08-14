#!/usr/bin/env python3
"""
Eregen 全业务链数据注入 - 直接数据库版本
绕过API限速，直接写入SQLite数据库
"""
import sqlite3
import uuid
import random
import hashlib
from datetime import datetime, timedelta
import sys

DB = "/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/admin-api/data/eregen.db"

def log(msg):
    print(msg, flush=True)

def main():
    log("=" * 60)
    log("Eregen Full Chain Data Seed (Direct DB)")
    log("=" * 60)

    conn = sqlite3.connect(DB)
    conn.execute("PRAGMA foreign_keys = ON")
    cur = conn.cursor()

    # Check current state
    cur.execute("SELECT COUNT(*) FROM persons")
    log(f"Current persons: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM users")
    log(f"Current users: {cur.fetchone()[0]}")

    # Elder data definition
    elders = [
        # Self chain (9 persons)
        ("310101195401010001", "张建国", 1, 72, "self", "pro", "high", True),
        ("310101195801020002", "李秀芳", 2, 68, "self", "plus", "medium", True),
        ("310101195101030003", "王德明", 1, 75, "self", "starter", "low", False),
        ("310101195601040004", "赵美华", 2, 70, "self", "pro_plus", "critical", True),
        ("310101196101050005", "陈志强", 1, 65, "self", "plus", "medium", False),
        ("310101194801060006", "刘淑珍", 2, 78, "self", "pro", "high", False),
        ("310101196301070007", "孙伟民", 1, 63, "self", "starter", "low", False),
        ("310101195601080008", "周海涛", 1, 70, "self", "pro", "high", True),
        ("310101195801090009", "吴雪梅", 2, 68, "self", "pro_plus", "critical", True),
        # Hospital chain (5 persons)
        ("310101195501100010", "郑国华", 1, 71, "hospital", None, "medium", False),
        ("310101196001110011", "钱丽华", 2, 66, "hospital", None, "high", False),
        ("310101195201120012", "杨建国", 1, 74, "hospital", None, "medium", False),
        ("310101195701130013", "黄美玲", 2, 69, "hospital", None, "low", False),
        ("310101195301140014", "林志强", 1, 73, "hospital", None, "high", False),
        # Community chain (4 persons)
        ("310101195001150015", "马德海", 1, 76, "community", None, "medium", False),
        ("310101194601160016", "朱秀英", 2, 80, "community", None, "high", False),
        ("310101195401170017", "何国强", 1, 72, "community", None, "low", False),
        ("310101196101180018", "高美华", 2, 65, "community", None, "medium", False),
    ]

    # 1. Create users (role accounts)
    log("\n[1] Creating role accounts...")
    users = [
        ("admin@eregen.com", "13800000001", "super_admin", "Admin@123", "系统管理员"),
        ("op@eregen.com", "13800000002", "operator", "Op@123", "自营运营员"),
        ("nurse01@eregen.com", "13800000003", "nurse", "Nurse@123", "住院护士"),
        ("cd01@eregen.com", "13800000004", "community_doctor", "Cd@123", "社区医生"),
        ("cs01@eregen.com", "13800000005", "community_staff", "Cs@123", "社区干事"),
        ("reg01@eregen.com", "13800000006", "regulator", "Reg@123", "监管人员"),
    ]
    for email, phone, role, pw, name in users:
        cur.execute("SELECT id FROM users WHERE email=?", (email,))
        if cur.fetchone():
            log(f"  OK {email}")
            continue
        uid_val = str(uuid.uuid4())
        pw_hash = hashlib.sha256(pw.encode()).hexdigest()
        try:
            cur.execute("INSERT INTO users (id, name, email, phone, role, password_hash) VALUES (?,?,?,?,?,?)",
                       (uid_val, name, email, phone, role, pw_hash))
            log(f"  OK {email}")
        except Exception as e:
            log(f"  ERR {email}: {e}")
    conn.commit()

    # 2. Create persons
    log("\n[2] Creating persons...")
    person_ids = {}
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        cur.execute("SELECT id FROM persons WHERE id_card=?", (id_card,))
        if cur.fetchone():
            log(f"  OK {name}")
            continue
        pid = str(uuid.uuid4())
        birth_year = datetime.now().year - age
        try:
            cur.execute("INSERT INTO persons (id, id_card, name, gender, birth_date, phone, address, status) VALUES (?,?,?,?,?,?,?,?)",
                       (pid, id_card, name, gender, f"{birth_year}-01-01", f"138{random.randint(10000000,99999999)}",
                        f"上海市浦东新区{random.randint(1,99)}号", "active"))
            person_ids[id_card] = pid
            log(f"  OK {name} [{chain}]")
        except Exception as e:
            log(f"  ERR {name}: {e}")
    conn.commit()

    # 3. Create person profiles
    log("\n[3] Creating profiles...")
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        pid = person_ids.get(id_card)
        if not pid:
            continue
        cur.execute("SELECT id FROM person_profiles WHERE person_id=? AND business_chain=?", (pid, chain))
        if cur.fetchone():
            log(f"  OK Profile {name} [{chain}]")
            continue
        try:
            if chain == "self":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, subscription_tier, subscription_status, health_risk_level) VALUES (?,?,?,?,?)",
                           (pid, chain, tier, "active", risk))
            elif chain == "hospital":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, admission_no, department, blood_type, status) VALUES (?,?,?,?,?,?)",
                           (pid, chain, f"H{random.randint(10000,99999)}", "内科", "O", "in_treatment"))
            elif chain == "community":
                cur.execute("INSERT INTO person_profiles (person_id, business_chain, hospital_id_community, minzheng_certified, status) VALUES (?,?,?,?,?)",
                           (pid, chain, "INST-001", 1, "active"))
            log(f"  OK Profile {name} [{chain}]")
        except Exception as e:
            log(f"  ERR Profile {name} [{chain}]: {e}")
    conn.commit()

    # 4. Create devices (self chain)
    log("\n[4] Creating devices...")
    device_ids = {}
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "self":
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        try:
            # Bracelet
            dev_id = f"BR-{uuid.uuid4().hex[:4].upper()}"
            cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?,?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, "bracelet", tier or "starter", "active"))
            device_ids[f"bracelet_{id_card}"] = dev_id
            cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain) VALUES (?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, pid, "self"))
            log(f"  OK Bracelet {dev_id}")
            # Pillbox
            if has_pillbox:
                pill_id = f"PX-{uuid.uuid4().hex[:4].upper()}"
                cur.execute("INSERT INTO devices (id, device_id, device_type, tier, status) VALUES (?,?,?,?,?)",
                           (str(uuid.uuid4()), pill_id, "pillbox", "pro", "active"))
                device_ids[f"pillbox_{id_card}"] = pill_id
                cur.execute("INSERT INTO device_bindings (id, device_id, person_id, business_chain) VALUES (?,?,?,?)",
                           (str(uuid.uuid4()), pill_id, pid, "self"))
                log(f"  OK Pillbox {pill_id}")
        except Exception as e:
            log(f"  ERR Device {name}: {e}")
    conn.commit()

    # 5. Medical wristbands
    log("\n[5] Creating medical wristbands...")
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "hospital":
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        try:
            dev_id = f"MW-{uuid.uuid4().hex[:4].upper()}"
            cur.execute("INSERT INTO medical_wristband_devices (id, device_id, firmware_version, status) VALUES (?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, "v1.2.0", "active"))
            device_ids[f"medical_{id_card}"] = dev_id
            log(f"  OK Medical wristband {dev_id}")
        except Exception as e:
            log(f"  ERR Medical wristband {name}: {e}")
    conn.commit()

    # 6. Community wristbands
    log("\n[6] Creating community wristbands...")
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "community":
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        try:
            dev_id = f"CW-{uuid.uuid4().hex[:4].upper()}"
            cur.execute("INSERT INTO community_wristband_devices (id, device_id, firmware_version, mode, status) VALUES (?,?,?,?,?)",
                       (str(uuid.uuid4()), dev_id, "v1.0.0", "community", "active"))
            device_ids[f"community_{id_card}"] = dev_id
            log(f"  OK Community wristband {dev_id}")
        except Exception as e:
            log(f"  ERR Community wristband {name}: {e}")
    conn.commit()

    # 7. Health records (50-99 per person)
    log("\n[7] Creating health records...")
    total_health = 0
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        pid = person_ids.get(id_card)
        if not pid:
            continue
        count = random.randint(55, 95)
        hr_base = 82 if risk in ("high", "critical") else 75
        for _ in range(count):
            try:
                cur.execute("INSERT INTO health_records (id, person_id, business_chain, record_type, source, hr, spo2, steps, sleep_hours, blood_pressure_sys, blood_pressure_dia, recorded_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
                           (str(uuid.uuid4()), pid, chain, "vitals", "device",
                            max(50, min(130, hr_base + random.randint(-15,15))),
                            max(88, min(100, 97 + random.randint(-3,2))),
                            random.randint(500, 12000), round(random.uniform(5.0, 9.5), 1),
                            random.randint(100, 180), random.randint(60, 110),
                            (datetime.now() - timedelta(hours=random.randint(0,48))).strftime("%Y-%m-%d %H:%M:%S")))
                total_health += 1
            except:
                pass
        log(f"  OK {name}: {count} records")
    conn.commit()
    log(f"  Total health records: {total_health}")

    # 8. Medication rules
    log("\n[8] Creating medication rules...")
    meds = [
        ("氨氯地平", "Amlodipine", "5mg", "每日1次", "08:00", "prescription"),
        ("二甲双胍", "Metformin", "500mg", "每日2次", "08:00;20:00", "prescription"),
        ("阿司匹林", "Aspirin", "100mg", "每日1次", "20:00", "otc"),
    ]
    total_meds = 0
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "self" or not has_pillbox:
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        for med in random.sample(meds, min(2, len(meds))):
            try:
                cur.execute("INSERT INTO medication_rules (id, person_id, business_chain, drug_name, generic_name, dosage, frequency, schedule_time1, drug_category, active) VALUES (?,?,?,?,?,?,?,?,?,?)",
                           (str(uuid.uuid4()), pid, "self", med[0], med[1], med[2], med[3], med[4], med[5], 1))
                total_meds += 1
            except:
                pass
    conn.commit()
    log(f"  Total medication rules: {total_meds}")

    # 9. Hospital data (daily entries, verifications, expenses)
    log("\n[9] Creating hospital data...")
    total_verifications = total_daily = total_expenses = 0
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "hospital":
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        # Daily entries
        for _ in range(random.randint(3, 5)):
            try:
                cur.execute("INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content) VALUES (?,?,?,?,?)",
                           (str(uuid.uuid4()), pid, (datetime.now() - timedelta(days=random.randint(0,14))).strftime("%Y-%m-%d"),
                            random.choice(["vitals", "medication", "nursing"]), "查房记录: 生命体征稳定"))
                total_daily += 1
            except:
                pass
        # Verifications
        mw_id = device_ids.get(f"medical_{id_card}")
        if mw_id:
            for _ in range(random.randint(5, 10)):
                try:
                    cur.execute("INSERT INTO medical_verifications (id, device_id, patient_id, verification_type, result, matched, verified_by) VALUES (?,?,?,?,?,?,?)",
                               (str(uuid.uuid4()), mw_id, pid, random.choice(["medication", "vitals", "nfc"]), "passed", 1, "nurse01"))
                    total_verifications += 1
                except:
                    pass
        # Expenses
        for _ in range(random.randint(5, 15)):
            try:
                cur.execute("INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, billing_source) VALUES (?,?,?,?,?,?,?)",
                           (str(uuid.uuid4()), pid, random.choice(["血常规", "CT扫描", "心电图", "B超"]),
                            random.choice(["lab", "radiology", "consultation"]), round(random.uniform(50, 800), 2), 1, "manual"))
                total_expenses += 1
            except:
                pass
    conn.commit()
    log(f"  Verifications: {total_verifications}, Daily: {total_daily}, Expenses: {total_expenses}")

    # 10. Community data (sign-ins)
    log("\n[10] Creating community data...")
    total_signins = 0
    for id_card, name, gender, age, chain, tier, risk, has_pillbox in elders:
        if chain != "community":
            continue
        pid = person_ids.get(id_card)
        if not pid:
            continue
        for _ in range(14):
            try:
                cur.execute("INSERT INTO community_signin_records (id, elder_id, device_id, hospital_id, signin_time, period, is_medical_signin, is_welfare_signin) VALUES (?,?,?,?,?,?,?,?)",
                           (str(uuid.uuid4()), pid, device_ids.get(f"community_{id_card}",""), "INST-001",
                            (datetime.now() - timedelta(hours=random.randint(0,336))).strftime("%Y-%m-%d %H:%M:%S"),
                            random.choice(["morning", "afternoon"]), 1, 1))
                total_signins += 1
            except:
                pass
    conn.commit()
    log(f"  Total signins: {total_signins}")

    # Summary
    log("\n" + "=" * 60)
    log("✅ SEED COMPLETE")
    cur.execute("SELECT COUNT(*) FROM users")
    log(f"  Users: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM persons")
    log(f"  Persons: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM person_profiles")
    log(f"  Profiles: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM health_records")
    log(f"  Health records: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM devices")
    log(f"  Devices: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM device_bindings")
    log(f"  Device bindings: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medication_rules")
    log(f"  Medication rules: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_verifications")
    log(f"  Medical verifications: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_daily_entries")
    log(f"  Medical daily entries: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_expenses")
    log(f"  Medical expenses: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM community_signin_records")
    log(f"  Community signins: {cur.fetchone()[0]}")
    log("=" * 60)

    conn.close()
    return 0

if __name__ == "__main__":
    sys.exit(main())
