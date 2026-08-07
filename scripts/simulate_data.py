#!/usr/bin/env python3
"""
Eregen Platform — Full Data Simulation & Connectivity Test
Simulates 10 bracelet elders, 20 medical wristband patients, 50 community elders
with realistic telemetry, health, location, alerts, and workflow data.
"""
import requests
import json
import time
import random
import uuid
import sys
from datetime import datetime, timedelta

BASE = "http://localhost:8085"
ADMIN_API = f"{BASE}/api/v1"

# ─── Auth ───────────────────────────────────────────────────────────
def login(email, password):
    r = requests.post(f"{ADMIN_API}/auth/login", json={
        "method": "email", "credential": email, "secret": password
    })
    r.raise_for_status()
    return r.json()["data"]["token"]

# ─── Helpers ────────────────────────────────────────────────────────
def uid(prefix=""):
    return f"{prefix}{uuid.uuid4().hex[:8]}"

def rand_date(days_back=365):
    return (datetime.now() - timedelta(days=random.randint(0, days_back))).strftime("%Y-%m-%d")

def rand_datetime(hours_back=48):
    return (datetime.now() - timedelta(hours=random.randint(0, hours_back))).strftime("%Y-%m-%dT%H:%M:%S")

def rand_loc(center_lat=31.2304, center_lon=121.4737, spread=0.02):
    return round(center_lat + random.gauss(0, spread), 6), \
           round(center_lon + random.gauss(0, spread), 6)

# ─── State ──────────────────────────────────────────────────────────
token = None
elderly_ids = []      # (id, name, user_id)
bracelet_devices = [] # (device_id, elderly_id)
medical_patients = [] # (id, admission_no, name)
medical_devices = []  # (device_id, patient_id)
community_elders = [] # (id, name, id_card)
community_devices = []# (device_id, elder_id)

stats = {
    "elderly_created": 0, "devices_created": 0, "medical_patients": 0,
    "community_elders": 0, "health_records": 0, "locations": 0,
    "alerts_generated": 0, "medication_rules": 0, "med_takes": 0,
    "verifications": 0, "admissions": 0, "ward_rounds": 0,
    "signins": 0, "pharmacy_logs": 0, "expenses": 0,
    "medications": 0, "test_results": 0, "daily_entries": 0,
    "welfare_tags": 0, "fence_configs": 0, "regulatory_alerts": 0,
}

def check(fn_name, ok, detail=""):
    status = "✅" if ok else "❌"
    print(f"  {status} {fn_name}: {detail}")
    return ok

# ═══════════════════════════════════════════════════════════════════
# PHASE 1: Setup — Create all entities
# ═══════════════════════════════════════════════════════════════════
def phase1_setup():
    global token
    print("\n" + "═"*60)
    print("PHASE 1: Entity Setup")
    print("═"*60)

    # Login
    print("\n[1.1] Logging in...")
    token = login("admin@eregen.com", "Admin@123")
    headers = {"Authorization": f"Bearer {token}"}
    print(f"  ✅ Authenticated as admin")

    # ── Create 10 elderly profiles with users ──────────────────────
    print("\n[1.2] Creating 10 elderly profiles + users + bracelets...")
    first_names = ["Wei","Fang","Ming","Lan","Jun","Mei","Hong","Jie","Ling","Qiang"]
    last_names = ["Zhang","Li","Wang","Chen","Liu","Yang","Huang","Zhao","Wu","Zhou"]
    health_tiers_options = [
        ["cardiovascular"], ["diabetes"], ["cardiovascular","diabetes"],
        ["hypertension"], ["dementia"], ["respiratory"],
        ["cardiovascular","respiratory"], ["diabetes","hypertension"],
        ["fall_risk"], ["multiple"]
    ]
    lat_center, lon_center = 31.2304, 121.4737

    for i in range(10):
        # Create user
        uid_val = uid("usr-")
        uname = f"{last_names[i]}{first_names[i]}"
        uemail = f"user{i+1}@eregen.com"
        r = requests.post(f"{ADMIN_API}/admin/users", headers=headers, json={
            "name": uname, "email": uemail, "role": "family",
            "password": "Test@12345"
        })
        if r.status_code != 201:
            print(f"  ⚠️  User {i+1} creation failed: {r.text[:100]}")
            continue
        user_id = r.json()["data"]["id"]

        # Create elderly profile
        eid = uid("eld-")
        r = requests.post(f"{ADMIN_API}/admin/elderly", headers=headers, json={
            "name": uname, "birth_date": rand_date(6000),
            "user_id": user_id,
            "health_tiers": health_tiers_options[i],
            "avatar_url": ""
        })
        if r.status_code != 201:
            print(f"  ⚠️  Elderly {i+1} creation failed: {r.text[:100]}")
            continue
        elderly_id = r.json()["data"]["id"]
        elderly_ids.append((elderly_id, uname, user_id))

        # Create bracelet device
        dev_id = f"BR-{i+1:04d}"
        r = requests.post(f"{ADMIN_API}/admin/devices", headers=headers, json={
            "device_id": dev_id, "device_type": "bracelet", "tier": "plus",
            "owner_user_id": user_id, "status": "online"
        })
        if r.status_code == 201:
            device_id = r.json()["data"]["id"]
            # Link device to elderly
            requests.post(f"{ADMIN_API}/admin/elderly/{elderly_id}/link-device",
                          headers=headers, json={"device_id": device_id})
            bracelet_devices.append((device_id, elderly_id, dev_id))
            stats["devices_created"] += 1

        stats["elderly_created"] += 1
        if (i+1) % 5 == 0:
            print(f"    Created {i+1}/10 elderly profiles")

    print(f"  ✅ Created {len(elderly_ids)} elderly profiles, {len(bracelet_devices)} bracelets")

    # ── Create 20 medical wristband patients ───────────────────────
    print("\n[1.3] Creating 20 medical wristband patients...")
    departments = ["Cardiology","Neurology","Orthopedics","Geriatrics","Internal Medicine",
                   "Surgery","Oncology","Pulmonology"]
    genders = ["male","female"]
    blood_types = ["A+","A-","B+","B-","AB+","AB-","O+","O-"]

    for i in range(20):
        pid = uid("pat-")
        adm_no = f"ADM-{2026001+i}"
        name = f"{random.choice(last_names)}{random.choice(first_names)}{i}"
        dept = random.choice(departments)
        gender = random.choice(genders)
        r = requests.post(f"{ADMIN_API}/admin/medical/patients", headers=headers, json={
            "id": pid, "admission_no": adm_no, "name": name,
            "gender": gender, "age": random.randint(60, 95),
            "department": dept, "bed_number": f"{random.randint(1,30)}B-{random.randint(1,10)}",
            "blood_type": random.choice(blood_types),
            "allergies": random.choice(["Penicillin","None","Sulfa","Latex","None","None"]),
            "special_conditions": random.choice(["Diabetes","Hypertension","CHF","COPD","Post-surgery","Dementia","None"]),
            "status": "admitted"
        })
        if r.status_code == 201:
            medical_patients.append((pid, adm_no, name))
            stats["medical_patients"] += 1

        # Create medical wristband device
        mw_dev_id = f"WB-MED-{i+1:04d}"
        r = requests.post(f"{ADMIN_API}/admin/medical/wristbands", headers=headers, json={
            "device_id": mw_dev_id, "firmware_version": "v2.1.0",
            "status": "active"
        })
        if r.status_code == 201:
            mw_device_id = r.json()["data"]["id"]
            # Bind to patient
            requests.post(f"{ADMIN_API}/admin/medical/wristbands/{mw_device_id}/bind",
                          headers=headers, json={"patient_id": pid})
            medical_devices.append((mw_device_id, pid, mw_dev_id))
            stats["devices_created"] += 1

        if (i+1) % 5 == 0:
            print(f"    Created {i+1}/20 medical patients")

    print(f"  ✅ Created {len(medical_patients)} patients, {len(medical_devices)} wristbands")

    # ── Create 50 community elders ─────────────────────────────────
    print("\n[1.4] Creating 50 community elders...")
    community_names = []
    for i in range(50):
        eid = uid("cel-")
        name = f"{random.choice(last_names)}{random.choice(first_names)}{i:02d}"
        id_card = f"310101194{i%100:02d}{random.randint(1,12):02d}{random.randint(1,28):02d}{'X' if random.random()>0.8 else str(random.randint(0,9))}"
        r = requests.post(f"{ADMIN_API}/admin/community-wb/elders", headers=headers, json={
            "id": eid, "name": name, "id_card": id_card,
            "gender": random.choice([0,1]),
            "age": random.randint(65, 95),
            "address": f"上海市{random.choice(['浦东','徐汇','长宁','静安','黄浦','虹口','杨浦'])}区{random.randint(1,200)}号",
            "emergency_contact": f"138{random.randint(10000000,99999999)}",
            "bank_account": f"6222{random.randint(1000000000,9999999999)}",
            "status": "active"
        })
        if r.status_code == 201:
            community_elders.append((eid, name, id_card))
            stats["community_elders"] += 1

        # Create community wristband device
        cw_dev_id = f"CW-{i+1:04d}"
        r = requests.post(f"{ADMIN_API}/admin/community-wb/devices", headers=headers, json={
            "device_id": cw_dev_id, "firmware_version": "v1.5.0",
            "status": "active"
        })
        if r.status_code == 201:
            cw_device_id = r.json()["data"]["id"]
            # Bind to elder
            requests.post(f"{ADMIN_API}/admin/community-wb/devices/bind",
                          headers=headers, json={"elder_id": eid, "device_id": cw_device_id})
            community_devices.append((cw_device_id, eid, cw_dev_id))
            stats["devices_created"] += 1

        if (i+1) % 10 == 0:
            print(f"    Created {i+1}/50 community elders")

    print(f"  ✅ Created {len(community_elders)} community elders, {len(community_devices)} devices")

    # ── Setup infrastructure ───────────────────────────────────────
    print("\n[1.5] Setting up infrastructure (fences, welfare tags, med rules)...")

    # Create regulatory fence config for each department
    for dept in departments[:5]:
        lat, lon = rand_loc(lat_center, lon_center, spread=0.05)
        r = requests.post(f"{ADMIN_API}/admin/regulatory/fence/config", headers=headers, json={
            "hospital_id": f"hospital-{departments.index(dept)+1:03d}",
            "hospital_name": f"{dept} Hospital",
            "center_lat": lat, "center_lng": lon,
            "radius_meters": random.randint(200, 500),
            "enabled": True
        })
        if r.status_code in (200, 201):
            stats["fence_configs"] += 1

    # Create welfare tags
    welfare_tags = [
        ("ELDERLY", "老年人福利", "民政局", 365, 50.0),
        ("DISABLED", "残疾人补贴", "残联", 365, 200.0),
        ("SOLITARY", "空巢老人补助", "民政局", 180, 100.0),
        ("CHRONIC", "慢性病补助", "卫健委", 365, 150.0),
        ("HIGH_AGE", "高龄津贴", "民政局", 365, 300.0),
    ]
    for tag_code, tag_name, issuer, days, amount in welfare_tags:
        r = requests.post(f"{ADMIN_API}/admin/community-wb/welfare-tags", headers=headers, json={
            "tag_code": tag_code, "tag_name": tag_name,
            "issuer": issuer, "renewal_period_days": days,
            "benefit_amount": amount, "enabled": True
        })
        if r.status_code in (200, 201):
            stats["welfare_tags"] += 1

    # Create medication rules for first 5 elderly
    for eid, name, uid_val in elderly_ids[:5]:
        for hour in [8, 12, 20]:
            r = requests.post(f"{ADMIN_API}/admin/elderly/{eid}/medication-rules", headers=headers, json={
                "schedule_time": f"{hour:02d}:00", "pill_type": random.choice(["tablet","capsule"]),
                "dose_count": random.randint(1, 3), "days_of_week": list(range(7)),
                "active": True
            })
            if r.status_code == 201:
                stats["medication_rules"] += 1

    print(f"  ✅ Infrastructure ready: {stats['fence_configs']} fences, {stats['welfare_tags']} tags, {stats['medication_rules']} med rules")
    print(f"\n  Total entities: {stats['elderly_created']} elderly + {stats['medical_patients']} patients + {stats['community_elders']} community elders + {stats['devices_created']} devices")

# ═══════════════════════════════════════════════════════════════════
# PHASE 2: Data Simulation
# ═══════════════════════════════════════════════════════════════════
def phase2_simulate():
    print("\n" + "═"*60)
    print("PHASE 2: Data Simulation (Real-time Telemetry)")
    print("═"*60)

    lat_c, lon_c = 31.2304, 121.4737

    # ── Simulate bracelet health/location/alerts for 10 elders ────
    print("\n[2.1] Simulating bracelet telemetry (10 elders × 50 samples)...")
    for idx, (eld_id, eld_name, user_id) in enumerate(elderly_ids):
        base_hr = random.randint(60, 80)
        base_spo2 = random.randint(95, 99)
        base_steps = random.randint(2000, 8000)
        base_lat, base_lon = rand_loc(lat_c, lon_c, 0.01)

        for t in range(50):
            ts = (datetime.now() - timedelta(minutes=(50-t)*5)).isoformat()
            hr = max(50, min(120, base_hr + random.gauss(0, 8)))
            spo2 = max(90, min(100, base_spo2 + random.gauss(0, 2)))
            steps = max(0, int(base_steps + random.gauss(0, 500) + t * 50))
            sleep_h = round(random.uniform(5, 9), 1)
            bp_sys = random.randint(110, 160)
            bp_dia = random.randint(70, 100)

            # Health record
            requests.post(f"{ADMIN_API}/admin/elderly/{eld_id}/health",
                          headers={"Authorization": f"Bearer {token}"}, json={
                "hr": int(hr), "spo2": int(spo2), "steps": steps,
                "sleep_hours": sleep_h, "bp_systolic": int(bp_sys),
                "bp_diastolic": int(bp_dia), "timestamp": ts
            })
            stats["health_records"] += 1

            # Location
            lat, lon = rand_loc(base_lat, base_lon, 0.002)
            acc = random.randint(3, 15)
            requests.post(f"{ADMIN_API}/admin/elderly/{eld_id}/location",
                          headers={"Authorization": f"Bearer {token}"}, json={
                "lat": lat, "lon": lon, "accuracy": acc, "timestamp": ts
            })
            stats["locations"] += 1

            # Occasional alerts
            if random.random() < 0.02:
                alert_type = random.choice(["sos", "fall", "geofence_exit"])
                requests.post(f"{ADMIN_API}/admin/alerts", headers=headers, json={
                    "elderly_id": eld_id, "alert_type": alert_type,
                    "severity": "high" if alert_type in ["sos","fall"] else "medium",
                    "message": f"{alert_type} detected for {eld_name}",
                    "device_id": bracelet_devices[idx][0] if idx < len(bracelet_devices) else ""
                })
                stats["alerts_generated"] += 1

            if (idx * 50 + t) % 200 == 0:
                print(f"    Progress: {idx*50+t}/500 samples...")

    print(f"  ✅ Bracelet telemetry: {stats['health_records']} health records, {stats['locations']} locations, {stats['alerts_generated']} alerts")

    # ── Simulate medical wristband activity ────────────────────────
    print("\n[2.2] Simulating medical wristband activity (20 patients)...")
    for pid, adm_no, pname in medical_patients:
        # Admissions
        adm_date = rand_date(30)
        r = requests.post(f"{ADMIN_API}/admin/medical/admissions", headers=headers, json={
            "patient_id": pid, "admission_no": adm_no,
            "department": random.choice(departments),
            "bed_no": f"{random.randint(1,30)}B-{random.randint(1,10)}",
            "diagnosis": random.choice(["Pneumonia","CHF Exacerbation","Hip Fracture",
                                        "COPD Flare","Stroke Recovery","Post-Surgical"]),
            "admitted_at": adm_date + "T08:00:00"
        })
        if r.status_code in (200, 201):
            stats["admissions"] += 1

        # Verifications (2-5 per patient)
        n_verifs = random.randint(2, 5)
        for v in range(n_verifs):
            vtype = random.choice(["medication","vitals","ward_round","discharge"])
            result = random.choice(["matched","unmatched","not_found"])
            matched = (result == "matched")
            r = requests.post(f"{ADMIN_API}/admin/medical/verifications", headers=headers, json={
                "patient_id": pid,
                "device_id": medical_devices[random.randint(0, len(medical_devices)-1)][0] if medical_devices else "",
                "verification_type": vtype, "result": result,
                "matched": matched, "verified_by": f"Nurse{random.randint(1,20)}",
                "notes": "" if matched else f"Mismatch on {vtype}"
            })
            if r.status_code in (200, 201):
                stats["verifications"] += 1

        # Medications (2-4)
        for _ in range(random.randint(2, 4)):
            r = requests.post(f"{ADMIN_API}/admin/medical/medications", headers=headers, json={
                "patient_id": pid, "name": random.choice(["Metformin","Lisinopril",
                    "Atorvastatin","Amlodipine","Omeprazole","Warfarin","Furosemide"]),
                "dosage": f"{random.choice([5,10,20,25,50])}mg",
                "frequency": random.choice(["QD","BID","TID","QID","PRN"]),
                "duration": f"{random.randint(7,90)}days"
            })
            if r.status_code in (200, 201):
                stats["medications"] += 1

        # Expenses (1-3)
        for _ in range(random.randint(1, 3)):
            r = requests.post(f"{ADMIN_API}/admin/medical/expenses", headers=headers, json={
                "patient_id": pid, "item_name": random.choice(["Blood Test","X-Ray","MRI","CT Scan",
                    "IV Medication","Consultation","Physical Therapy"]),
                "category": random.choice(["Lab","Imaging","Medication","Consultation","Therapy"]),
                "amount": round(random.uniform(50, 2000), 2),
                "quantity": 1, "unit_price": 0
            })
            if r.status_code in (200, 201):
                stats["expenses"] += 1

        # Test results (1-3)
        for _ in range(random.randint(1, 3)):
            r = requests.post(f"{ADMIN_API}/admin/medical/test-results", headers=headers, json={
                "patient_id": pid, "test_name": random.choice(["CBC","CMP","Lipid Panel",
                    "HbA1c","TSH","BNP","Troponin"]),
                "result": f"{random.uniform(3, 25):.1f}",
                "unit": random.choice(["10^3/uL","mg/dL","mmol/L","%","mIU/L","pg/mL"]),
                "reference_range": "3.5-11.0", "status": random.choice(["normal","abnormal","critical"])
            })
            if r.status_code in (200, 201):
                stats["test_results"] += 1

        # Daily entries (2-5)
        for _ in range(random.randint(2, 5)):
            r = requests.post(f"{ADMIN_API}/admin/medical/daily-entries", headers=headers, json={
                "patient_id": pid, "entry_date": rand_date(14),
                "entry_type": random.choice([" nursing_note","assessment","care_plan"]),
                "content": f"Patient {pname} showing {random.choice(['stable','improving','needs monitoring'])} condition",
                "nurse_id": f"N{random.randint(1,20)}"
            })
            if r.status_code in (200, 201):
                stats["daily_entries"] += 1

        # Ward rounds (1-3)
        for _ in range(random.randint(1, 3)):
            r = requests.post(f"{ADMIN_API}/admin/medical/patients/{pid}/ward-round", headers=headers, json={
                "nurse_id": f"N{random.randint(1,20)}",
                "blood_pressure": f"{random.randint(100,160)}/{random.randint(60,100)}",
                "heart_rate": random.randint(60, 100),
                "spo2": random.randint(92, 99),
                "temperature": round(random.uniform(36.5, 38.5), 1),
                "weight": round(random.uniform(50, 100), 1),
                "notes": "",
                "observations": json.dumps({"falls_risk": random.choice([True,False]),
                                            "pain_level": random.randint(0,10)})
            })
            if r.status_code in (200, 201):
                stats["ward_rounds"] += 1

        # Discharge some patients
        if random.random() < 0.3:
            requests.post(f"{ADMIN_API}/admin/medical/admissions/{uid('adm-')}/discharge",
                          headers=headers, json={"discharge_type": "discharged", "notes": "Stable for discharge"})

    print(f"  ✅ Medical: {stats['admissions']} admissions, {stats['verifications']} verifications, "
          f"{stats['medications']} medications, {stats['expenses']} expenses, "
          f"{stats['test_results']} test results, {stats['daily_entries']} daily entries, {stats['ward_rounds']} ward rounds")

    # ── Simulate community wristband activity ──────────────────────
    print("\n[2.3] Simulating community wristband activity (50 elders)...")
    for eid, ename, icard in community_elders:
        # Assign 1-3 welfare tags
        n_tags = random.randint(1, 3)
        assigned_tags = random.sample(["ELDERLY","DISABLED","SOLITARY","CHRONIC","HIGH_AGE"], n_tags)
        for tag in assigned_tags:
            r = requests.post(f"{ADMIN_API}/admin/community-wb/elders/{eid}/welfare/{tag}",
                              headers=headers, json={})
            if r.status_code in (200, 201):
                stats["welfare_tags"] += 1

        # Sign-ins (3-10 over past 30 days)
        n_signins = random.randint(3, 10)
        for s in range(n_signins):
            sig_date = datetime.now() - timedelta(days=random.randint(0, 30))
            period = sig_date.strftime("%Y-%m")
            is_welfare = random.random() < 0.5
            r = requests.post(f"{ADMIN_API}/admin/community-wb/signin/trigger", headers=headers, json={
                "elder_id": eid,
                "device_id": community_devices[random.randint(0, len(community_devices)-1)][0] if community_devices else "",
                "period": period,
                "is_medical_signin": True,
                "is_welfare_signin": is_welfare,
                "activated_tags": json.dumps(random.sample(assigned_tags, min(2, len(assigned_tags)))) if assigned_tags else "[]"
            })
            if r.status_code in (200, 201):
                stats["signins"] += 1

        # Pharmacy dispenses (1-4)
        for _ in range(random.randint(1, 4)):
            r = requests.post(f"{ADMIN_API}/admin/community-wb/pharmacy/dispense", headers=headers, json={
                "elder_id": eid,
                "device_id": community_devices[random.randint(0, len(community_devices)-1)][0] if community_devices else "",
                "period": datetime.now().strftime("%Y-%m"),
                "items": json.dumps([random.choice(["Metformin","Amlodipine","Atorvastatin",
                                                    "Omeprazole","Aspirin","Losartan"])]),
                "total_cost": round(random.uniform(20, 200), 2),
                "insurance_covered": round(random.uniform(5, 100), 2),
                "self_pay": round(random.uniform(10, 100), 2)
            })
            if r.status_code in (200, 201):
                stats["pharmacy_logs"] += 1

    # Minzheng sync records
    for _ in range(3):
        r = requests.post(f"{ADMIN_API}/admin/community-wb/minzheng/import", headers=headers, json={
            "source": random.choice(["民政局系统","社保局","街道办"]),
            "filename": f"batch_{random.randint(1,100)}.csv",
            "imported_count": random.randint(10, 50),
            "matched_count": random.randint(5, 40),
            "pending_review_count": random.randint(0, 10),
            "error_count": random.randint(0, 3)
        })
        if r.status_code in (200, 201):
            pass  # Track sync count

    # Batch payments
    for batch in range(2):
        batch_id = uid("batch-")
        period = datetime.now().strftime("%Y-%m")
        for eid, ename, icard in random.sample(community_elders, min(10, len(community_elders))):
            r = requests.post(f"{ADMIN_API}/admin/community-wb/batch-pay/execute", headers=headers, json={
                "batch_id": batch_id, "period": period, "pay_type": "welfare_subsidy",
                "elder_id": eid, "amount": round(random.uniform(100, 500), 2),
                "bank_account": random.choice(community_elders)[2]  # using id_card as stand-in
            })
            if r.status_code in (200, 201):
                stats["pharmacy_logs"] += 1  # counting payments

    print(f"  ✅ Community: {stats['signins']} sign-ins, {stats['welfare_tags']} total tags, "
          f"{stats['pharmacy_logs']} pharmacy/payment records")

    # ── Regulatory alerts ──────────────────────────────────────────
    print("\n[2.4] Simulating regulatory alerts...")
    for _ in range(15):
        pid = random.choice(medical_patients)[0] if medical_patients else uid("pat-")
        dept = random.choice(departments)
        alert_type = random.choice(["no_verify","fence_violation","fake_admission",
                                     "expense_spike","med_verify_mismatch","frequent_transfer"])
        r = requests.post(f"{ADMIN_API}/admin/regulatory/alerts", headers=headers, json={
            "rule_code": f"R_{random.choice(['01','02','03','04','05','06','07','08','09','10'])}",
            "patient_id": pid, "hospital_id": f"hospital-{random.randint(1,5):03d}",
            "department": dept, "severity": random.choice(["low","medium","high"]),
            "alert_type": alert_type, "detail": f"Regulatory alert: {alert_type}"
        })
        if r.status_code in (200, 201):
            stats["regulatory_alerts"] += 1

    # Acknowledge/resolve some alerts
    r = requests.get(f"{ADMIN_API}/admin/alerts?limit=20", headers=headers)
    if r.status_code == 200:
        alerts = r.json().get("data", {}).get("alerts", [])
        for a in alerts[:5]:
            aid = a.get("id")
            if aid:
                action = random.choice(["acknowledge", "resolve"])
                requests.post(f"{ADMIN_API}/admin/alerts/{aid}/{action}", headers=headers, json={})
                stats["alerts_generated"] += 1  # track processed

    print(f"  ✅ Generated {stats['regulatory_alerts']} regulatory alerts")

# ═══════════════════════════════════════════════════════════════════
# PHASE 3: Verification
# ═══════════════════════════════════════════════════════════════════
def phase3_verify():
    print("\n" + "═"*60)
    print("PHASE 3: Data Integrity Verification")
    print("═"*60)

    results = []

    # Check elderly
    r = requests.get(f"{ADMIN_API}/admin/elderly?page=1&page_size=100", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", 0)
        ok = check("Elderly count", total >= 10, f"expected≥10 got={total}")
        results.append(ok)
        if total > 0:
            sample = data.get("elderly", [{}])[0]
            ok = check("Elderly fields", all(k in sample for k in ["id","name","health_tiers"]),
                       f"fields={list(sample.keys())}")
            results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Elderly list: HTTP {r.status_code}")

    # Check devices
    r = requests.get(f"{ADMIN_API}/admin/devices?page=1&page_size=100", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", 0)
        ok = check("Devices count", total >= 80, f"expected≥80 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Devices list: HTTP {r.status_code}")

    # Check health records for a sample elderly
    if elderly_ids:
        eid = elderly_ids[0][0]
        r = requests.get(f"{ADMIN_API}/admin/elderly/{eid}/health-records?limit=10", headers=headers)
        if r.status_code == 200:
            records = r.json().get("data", [])
            ok = check("Health records", len(records) > 0, f"count={len(records)}")
            results.append(ok)
            if records:
                ok = check("Health record fields", "hr" in records[0] or "hr" in str(records[0]),
                           f"keys={list(records[0].keys()) if isinstance(records[0],dict) else type(records[0])}")
                results.append(ok)
        else:
            results.append(False)
            print(f"  ❌ Health records: HTTP {r.status_code}")

    # Check location history
    if elderly_ids:
        eid = elderly_ids[0][0]
        r = requests.get(f"{ADMIN_API}/admin/elderly/{eid}/location-history?limit=10", headers=headers)
        if r.status_code == 200:
            records = r.json().get("data", [])
            ok = check("Location history", len(records) > 0, f"count={len(records)}")
            results.append(ok)
        else:
            results.append(False)
            print(f"  ❌ Location history: HTTP {r.status_code}")

    # Check medical patients
    r = requests.get(f"{ADMIN_API}/admin/medical/patients?page=1&page_size=30", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("patients", []))))
        ok = check("Medical patients", total >= 20, f"expected≥20 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Medical patients: HTTP {r.status_code}")

    # Check medical verifications
    r = requests.get(f"{ADMIN_API}/admin/medical/verifications?page=1&page_size=30", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("verifications", []))))
        ok = check("Medical verifications", total >= 20, f"expected≥20 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Medical verifications: HTTP {r.status_code}")

    # Check community elders
    r = requests.get(f"{ADMIN_API}/admin/community-wb/elders?page=1&page_size=60", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("elders", []))))
        ok = check("Community elders", total >= 50, f"expected≥50 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Community elders: HTTP {r.status_code}")

    # Check community sign-ins
    r = requests.get(f"{ADMIN_API}/admin/community-wb/signin/records?page=1&page_size=10", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("records", []))))
        ok = check("Community sign-ins", total >= 50, f"expected≥50 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Community sign-ins: HTTP {r.status_code}")

    # Check regulatory alerts
    r = requests.get(f"{ADMIN_API}/admin/regulatory/alerts?page=1&page_size=20", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("alerts", []))))
        ok = check("Regulatory alerts", total >= 10, f"expected≥10 got={total}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Regulatory alerts: HTTP {r.status_code}")

    # Check dashboard overview
    r = requests.get(f"{ADMIN_API}/admin/stats/overview", headers=headers)
    if r.status_code == 200:
        data = r.json().get("data", {})
        ok = check("Dashboard overview", bool(data), f"keys={list(data.keys()) if data else 'empty'}")
        results.append(ok)
    else:
        results.append(False)
        print(f"  ❌ Dashboard overview: HTTP {r.status_code}")

    # Check elderly with most health data
    if elderly_ids:
        eid = elderly_ids[0][0]
        r = requests.get(f"{ADMIN_API}/admin/elderly/{eid}/health-stats", headers=headers)
        if r.status_code == 200:
            data = r.json().get("data", {})
            ok = check("Elderly health stats", bool(data), f"data={data}")
            results.append(ok)
        else:
            results.append(False)
            print(f"  ❌ Health stats: HTTP {r.status_code}")

    # ── Summary ────────────────────────────────────────────────────
    print("\n" + "─"*60)
    print("SIMULATION STATISTICS")
    print("─"*60)
    print(f"  Elderly profiles:         {stats['elderly_created']:>4}")
    print(f"  Bracelet devices:         {len(bracelet_devices):>4}")
    print(f"  Medical patients:         {stats['medical_patients']:>4}")
    print(f"  Medical wristbands:       {len(medical_devices):>4}")
    print(f"  Community elders:         {stats['community_elders']:>4}")
    print(f"  Community devices:        {len(community_devices):>4}")
    print(f"  ——— Data points ———")
    print(f"  Health records:           {stats['health_records']:>5}")
    print(f"  Location records:         {stats['locations']:>5}")
    print(f"  Alerts generated:         {stats['alerts_generated']:>5}")
    print(f"  Medication rules:         {stats['medication_rules']:>5}")
    print(f"  Verifications:            {stats['verifications']:>5}")
    print(f"  Admissions:               {stats['admissions']:>5}")
    print(f"  Ward rounds:              {stats['ward_rounds']:>5}")
    print(f"  Expenses:                 {stats['expenses']:>5}")
    print(f"  Test results:             {stats['test_results']:>5}")
    print(f"  Daily entries:            {stats['daily_entries']:>5}")
    print(f"  Welfare tags assigned:    {stats['welfare_tags']:>5}")
    print(f"  Community sign-ins:       {stats['signins']:>5}")
    print(f"  Pharmacy logs:            {stats['pharmacy_logs']:>5}")
    print(f"  Regulatory alerts:        {stats['regulatory_alerts']:>5}")
    print(f"  Fence configs:            {stats['fence_configs']:>5}")
    print("─"*60)

    passed = sum(results)
    total = len(results)
    print(f"\n  VERIFICATION: {passed}/{total} checks passed")
    if passed == total:
        print("  ✅ ALL CHECKS PASSED — Data integrity verified!")
    else:
        print(f"  ⚠️  {total - passed} check(s) failed — review output above")

    return all(results)

# ═══════════════════════════════════════════════════════════════════
# Main
# ═══════════════════════════════════════════════════════════════════
if __name__ == "__main__":
    headers = {}  # populated after login
    print("Eregen Platform — Full Data Simulation & Verification")
    print(f"Target: {BASE}")

    try:
        phase1_setup()
        phase2_simulate()
        success = phase3_verify()
        sys.exit(0 if success else 1)
    except requests.exceptions.ConnectionError:
        print(f"\n❌ Cannot connect to {BASE}")
        print("   Make sure admin-api is running: PORT=8085 ./scripts/start.sh start admin-api")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback; traceback.print_exc()
        sys.exit(1)
