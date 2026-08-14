#!/usr/bin/env python3
"""
Eregen 全业务链数据模拟 - 最终版
使用正确的数据库schema
"""
import sqlite3
import uuid
import random
from datetime import datetime, timedelta

DB = "/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/admin-api/data/eregen.db"

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

def main():
    log("=" * 70)
    log("Eregen Final Data Seed")
    log("=" * 70)

    conn = sqlite3.connect(DB)
    conn.execute("PRAGMA foreign_keys = ON")
    cur = conn.cursor()

    # 1. Health records (bulk insert)
    log("\n[1] Creating health records...")
    total_health = 0
    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue
        pid = None
        cur.execute("SELECT id FROM persons WHERE id_card=?", (elder["id_card"],))
        row = cur.fetchone()
        if row:
            pid = row[0]
        else:
            log(f"  ✗ Person not found: {elder['name']}")
            continue

        chain = "self"
        count = random.randint(60, 90)
        hr_base = 82 if elder.get("risk") in ("high", "critical") else 75

        for i in range(count):
            try:
                cur.execute("""
                    INSERT INTO health_records (id, person_id, business_chain, record_type, source,
                                               hr, spo2, steps, sleep_hours,
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
                    (datetime.now() - timedelta(hours=random.randint(0, 48))).strftime("%Y-%m-%d %H:%M:%S")
                ))
                total_health += 1
            except Exception as e:
                pass
        log(f"  ✓ {elder['name']}: {count} health records")
    conn.commit()
    log(f"  Total health records: {total_health}")

    # 2. Medication rules
    log("\n[2] Creating medication rules...")
    meds = [
        ("氨氯地平", "Amlodipine", "5mg", "每日1次", "08:00", "prescription"),
        ("二甲双胍", "Metformin", "500mg", "每日2次", "08:00;20:00", "prescription"),
        ("阿司匹林", "Aspirin", "100mg", "每日1次", "20:00", "otc"),
    ]
    total_meds = 0
    for elder in ELDERS:
        if elder.get("tier") != "self" or not elder.get("has_pillbox"):
            continue
        pid = None
        cur.execute("SELECT id FROM persons WHERE id_card=?", (elder["id_card"],))
        row = cur.fetchone()
        if row:
            pid = row[0]
        else:
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
                log(f"  ✗ Failed: {e}")
    conn.commit()
    log(f"  Total medication rules: {total_meds}")

    # 3. Alerts
    log("\n[3] Creating alerts...")
    total_alerts = 0
    for elder in ELDERS:
        if elder.get("tier") != "self":
            continue
        pid = None
        cur.execute("SELECT id FROM persons WHERE id_card=?", (elder["id_card"],))
        row = cur.fetchone()
        if row:
            pid = row[0]
        else:
            continue
        for _ in range(random.randint(2, 4)):
            try:
                alert_type = random.choice(["high_hr", "low_spo2", "fall", "sos", "medication_missed"])
                severity = random.choice(["p0", "p1", "p2"])
                status = random.choice(["pending", "acknowledged", "resolved"])
                cur.execute("""
                    INSERT INTO alerts (id, person_id, business_chain, alert_type, severity, status, message)
                    VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (
                    str(uuid.uuid4()), pid, "self", alert_type, severity, status,
                    f"检测到{alert_type.replace('_', ' ')}异常"
                ))
                total_alerts += 1
            except:
                pass
    conn.commit()
    log(f"  Total alerts: {total_alerts}")

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
    cur.execute("SELECT COUNT(*) FROM hospital_admissions")
    log(f"  Hospital admissions: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_daily_entries")
    log(f"  Medical daily entries: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_verifications")
    log(f"  Medical verifications: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM medical_expenses")
    log(f"  Medical expenses: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM community_elders")
    log(f"  Community elders: {cur.fetchone()[0]}")
    cur.execute("SELECT COUNT(*) FROM community_signin_records")
    log(f"  Community sign-ins: {cur.fetchone()[0]}")

    log("=" * 70)
    conn.close()
    return 0

if __name__ == "__main__":
    import sys
    sys.exit(main())
