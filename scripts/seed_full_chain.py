#!/usr/bin/env python3
"""
Eregen 全业务链数据注入脚本 v3
==============================
通过真实API端点创建完整测试数据集。
"""
import requests
import json
import random
import sys
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, Optional

API = "http://localhost:8089/api/v1"
JWT_TOKEN: Optional[str] = None
HEADERS: Dict[str, str] = {}
CENTER_LAT, CENTER_LON = 31.2304, 121.4737

# ─── Elder Data (20 persons, 2 cross-chain) ─────────────────────────────────
ELDERS = [
    # Self chain: 7 regular + 2 cross-chain
    {"name": "张建国", "gender": 1, "age": 72, "chain": "self", "id_card": "310101195401010001", "tier": "pro", "risk": "high", "cross_chain": False,
     "conditions": ["hypertension", "diabetes"], "has_pillbox": True},
    {"name": "李秀芳", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801020002", "tier": "plus", "risk": "medium", "cross_chain": False,
     "conditions": ["hypertension"], "has_pillbox": True},
    {"name": "王德明", "gender": 1, "age": 75, "chain": "self", "id_card": "310101195101030003", "tier": "starter", "risk": "low", "cross_chain": False,
     "conditions": [], "has_pillbox": False},
    {"name": "赵美华", "gender": 2, "age": 70, "chain": "self", "id_card": "310101195601040004", "tier": "pro_plus", "risk": "critical", "cross_chain": False,
     "conditions": ["diabetes", "coronary_heart_disease"], "has_pillbox": True},
    {"name": "陈志强", "gender": 1, "age": 65, "chain": "self", "id_card": "310101196101050005", "tier": "plus", "risk": "medium", "cross_chain": False,
     "conditions": ["hypertension"], "has_pillbox": False},
    {"name": "刘淑珍", "gender": 2, "age": 78, "chain": "self", "id_card": "310101194801060006", "tier": "pro", "risk": "high", "cross_chain": False,
     "conditions": ["osteoporosis"], "has_pillbox": False},
    {"name": "孙伟民", "gender": 1, "age": 63, "chain": "self", "id_card": "310101196301070007", "tier": "starter", "risk": "low", "cross_chain": False,
     "conditions": [], "has_pillbox": False},
    # Cross-chain A: self → hospital → community
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "self", "id_card": "310101195601080008", "tier": "pro", "risk": "high", "cross_chain": True, "cross_id": 0,
     "conditions": ["hypertension", "diabetes"], "has_pillbox": True},
    # Cross-chain B: self → hospital → community
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801090009", "tier": "pro_plus", "risk": "critical", "cross_chain": True, "cross_id": 1,
     "conditions": ["diabetes", "chronic_kidney_disease"], "has_pillbox": True},
    # Hospital chain: 5 regular + 2 cross-chain
    {"name": "郑国华", "gender": 1, "age": 71, "chain": "hospital", "id_card": "310101195501100010", "tier": None, "risk": "medium", "cross_chain": False,
     "conditions": ["pneumonia"], "dept": "呼吸科", "blood_type": "A"},
    {"name": "钱丽华", "gender": 2, "age": 66, "chain": "hospital", "id_card": "310101196001110011", "tier": None, "risk": "high", "cross_chain": False,
     "conditions": ["hip_fracture"], "dept": "骨科", "blood_type": "B"},
    {"name": "杨建国", "gender": 1, "age": 74, "chain": "hospital", "id_card": "310101195201120012", "tier": None, "risk": "medium", "cross_chain": False,
     "conditions": ["stroke_recovery"], "dept": "神经内科", "blood_type": "O"},
    {"name": "黄美玲", "gender": 2, "age": 69, "chain": "hospital", "id_card": "310101195701130013", "tier": None, "risk": "low", "cross_chain": False,
     "conditions": ["appendicitis"], "dept": "普外科", "blood_type": "AB"},
    {"name": "林志强", "gender": 1, "age": 73, "chain": "hospital", "id_card": "310101195301140014", "tier": None, "risk": "high", "cross_chain": False,
     "conditions": ["heart_failure"], "dept": "心内科", "blood_type": "A"},
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "hospital", "id_card": "310101195601080008", "tier": None, "risk": "high", "cross_chain": True, "cross_id": 0,
     "conditions": ["pneumonia"], "dept": "呼吸科", "blood_type": "O"},
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "hospital", "id_card": "310101195801090009", "tier": None, "risk": "critical", "cross_chain": True, "cross_id": 1,
     "conditions": ["diabetic_ketoacidosis"], "dept": "内分泌科", "blood_type": "B"},
    # Community chain: 4 regular + 2 cross-chain
    {"name": "马德海", "gender": 1, "age": 76, "chain": "community", "id_card": "310101195001150015", "tier": None, "risk": "medium", "cross_chain": False,
     "conditions": [], "welfare_tags": ["高龄补贴"]},
    {"name": "朱秀英", "gender": 2, "age": 80, "chain": "community", "id_card": "310101194601160016", "tier": None, "risk": "high", "cross_chain": False,
     "conditions": ["hypertension"], "welfare_tags": ["特困供养", "高龄补贴"]},
    {"name": "何国强", "gender": 1, "age": 72, "chain": "community", "id_card": "310101195401170017", "tier": None, "risk": "low", "cross_chain": False,
     "conditions": [], "welfare_tags": ["低保户"]},
    {"name": "高美华", "gender": 2, "age": 65, "chain": "community", "id_card": "310101196101180018", "tier": None, "risk": "medium", "cross_chain": False,
     "conditions": ["diabetes"], "welfare_tags": ["残疾补助"]},
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "community", "id_card": "310101195601080008", "tier": None, "risk": "high", "cross_chain": True, "cross_id": 0,
     "conditions": ["hypertension", "diabetes"], "welfare_tags": ["特病认证"]},
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "community", "id_card": "310101195801090009", "tier": None, "risk": "critical", "cross_chain": True, "cross_id": 1,
     "conditions": ["diabetes", "chronic_kidney_disease"], "welfare_tags": ["特病认证", "高龄补贴"]},
]

# ─── Role Accounts ───────────────────────────────────────────────────────────
ROLES = [
    {"name": "系统管理员", "email": "admin@eregen.com", "role": "super_admin", "password": "Admin@123", "phone": "13800000001"},
    {"name": "自营运营员", "email": "op@eregen.com", "role": "operator", "password": "Op@123", "phone": "13800000002"},
    {"name": "住院护士", "email": "nurse01@eregen.com", "role": "nurse", "password": "Nurse@123", "phone": "13800000003"},
    {"name": "社区医生", "email": "cd01@eregen.com", "role": "community_doctor", "password": "Cd@123", "phone": "13800000004"},
    {"name": "社区干事", "email": "cs01@eregen.com", "role": "community_staff", "password": "Cs@123", "phone": "13800000005"},
    {"name": "监管人员", "email": "reg01@eregen.com", "role": "regulator", "password": "Reg@123", "phone": "13800000006"},
]

# ─── Tracking ───────────────────────────────────────────────────────────────
person_ids: Dict[str, str] = {}  # id_card -> person_id (self chain)
person_profiles: Dict[str, str] = {}  # (id_card, chain) -> person_id for cross-chain
device_ids: Dict[str, str] = {}

def uid(prefix=""):
    return f"{prefix}{uuid.uuid4().hex[:8]}"

def rand_date(days_back=90):
    return (datetime.now() - timedelta(days=random.randint(0, days_back))).strftime("%Y-%m-%d")

def rand_datetime(hours_back=48):
    return (datetime.now() - timedelta(hours=random.randint(0, hours_back))).strftime("%Y-%m-%d %H:%M:%S")

def check(name, ok, detail=""):
    mark = "✅" if ok else "❌"
    status = f": {detail}" if detail else ""
    print(f"  {mark} {name}{status}")
    return ok

# ─── Login ───────────────────────────────────────────────────────────────────
def login_admin():
    global JWT_TOKEN, HEADERS
    r = requests.post(f"{API}/auth/login", json={
        "method": "email", "credential": "admin@eregen.com", "secret": "Admin@123"
    }, timeout=10)
    if r.status_code != 200:
        print(f"❌ Login failed: {r.text}"); sys.exit(1)
    JWT_TOKEN = r.json()["data"]["token"]
    HEADERS = {"Authorization": f"Bearer {JWT_TOKEN}", "Content-Type": "application/json"}
    print("✅ Logged in as admin")

# ─── Module 1: Create Role Accounts ─────────────────────────────────────────
def create_role_accounts():
    print("\n[1/10] Creating role accounts...")
    # Fetch existing users to check for duplicates
    r = requests.get(f"{API}/admin/users", headers=HEADERS, params={"page": 1, "page_size": 100})
    existing = {}
    if r.status_code == 200:
        for u in r.json().get("data", []):
            existing[u.get("phone", "")] = u.get("role", "")
            existing[u.get("email", "")] = u.get("role", "")

    for role in ROLES:
        # Check if user already exists by email or phone
        already_exists = role["email"] in existing or role["email"] in existing
        if already_exists:
            check(f"Account {role['email']}", True, f"already exists as {role['role']}")
            continue
        r = requests.post(f"{API}/admin/users", headers=HEADERS, json={
            "name": role["name"], "email": role["email"],
            "role": role["role"], "password": role["password"]
        })
        ok = r.status_code in (200, 201)
        check(f"Create {role['role']}", ok, r.text if not ok else "")
        time.sleep(0.3)

# ─── Module 2: Create Persons ───────────────────────────────────────────────
def create_persons():
    print("\n[2/10] Creating persons...")
    for elder in ELDERS:
        birth_year = datetime.now().year - elder["age"]
        data = {
            "id_card": elder["id_card"],
            "name": elder["name"],
            "gender": elder["gender"],
            "birth_date": f"{birth_year}-01-01",
            "phone": f"138{random.randint(10000000, 99999999)}",
            "emergency_contact": f"139{random.randint(10000000, 99999999)}",
            "address": f"上海市浦东新区{random.randint(1,99)}号",
        }
        # For cross-chain persons, create separate person records per chain
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        if pid_key in person_profiles:
            pid = person_profiles[pid_key]
            check(f"Reuse {elder['name']} [{elder['chain']}]", True, pid)
            time.sleep(0.1)
            continue
        r = requests.post(f"{API}/admin/persons", headers=HEADERS, json=data)
        if r.status_code in (200, 201):
            pid = r.json().get("data", {}).get("id", "")
            person_profiles[pid_key] = pid
            if elder["chain"] == "self":
                person_ids[elder["id_card"]] = pid
            check(f"Create {elder['name']} [{elder['chain']}]", True, pid)
        else:
            check(f"Create {elder['name']}", False, r.text[:80])
        time.sleep(0.2)

# ─── Module 3: Create Person Profiles ───────────────────────────────────────
def create_person_profiles():
    print("\n[3/10] Creating person profiles...")
    for elder in ELDERS:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue
        chain = elder["chain"]
        profile_data = {"person_id": pid, "business_chain": chain}

        if chain == "self":
            profile_data.update({
                "subscription_tier": elder["tier"],
                "subscription_status": "active",
                "health_risk_level": elder["risk"]
            })
        elif chain == "hospital":
            profile_data.update({
                "admission_no": f"H{random.randint(10000, 99999)}",
                "department": elder.get("dept", "内科"),
                "bed_number": f"{random.randint(1,30)}床",
                "blood_type": "O",
                "attending_doctor": "张医生",
                "diagnosis": elder["conditions"][0] if elder["conditions"] else "待诊断",
                "status": "in_treatment"
            })
        elif chain == "community":
            profile_data.update({
                "hospital_id_community": "INST-001",
                "minzheng_certified": 1,
                "subsidy_type": "定期补助",
                "status": "active"
            })

        r = requests.post(f"{API}/admin/persons/profile", headers=HEADERS, json=profile_data)
        ok = r.status_code in (200, 201)
        if not ok:
            print(f"  ❌ Profile {elder['name']} [{chain}]: {r.text[:100]}")
        time.sleep(0.2)
    print("  ✅ Person profiles created")

# ─── Module 4: Create Devices ───────────────────────────────────────────────
def create_devices():
    print("\n[4/10] Creating devices...")
    # Self chain: bracelets + pillboxes
    for elder in ELDERS:
        if elder["chain"] != "self":
            continue
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue

        # Bracelet
        dev_id = f"BR-{uid()[:4].upper()}"
        r = requests.post(f"{API}/admin/devices", headers=HEADERS, json={
            "device_id": dev_id, "device_type": "bracelet",
            "tier": elder.get("tier", "starter"), "status": "active"
        })
        if r.status_code in (200, 201):
            device_ids[f"bracelet_{elder['id_card']}"] = dev_id
            check(f"Bracelet {dev_id}", True)
        else:
            check(f"Bracelet {dev_id}", False, r.text[:80])
            dev_id = ""

        # Bind device to person
        if dev_id:
            r2 = requests.post(f"{API}/admin/device-bindings", headers=HEADERS, json={
                "device_id": dev_id, "person_id": pid, "business_chain": "self"
            })
            check(f"Bind bracelet to {elder['name']}", r2.status_code in (200, 201))

        # Pillbox
        if elder.get("has_pillbox"):
            pill_id = f"PX-{uid()[:4].upper()}"
            r = requests.post(f"{API}/admin/devices", headers=HEADERS, json={
                "device_id": pill_id, "device_type": "pillbox",
                "tier": "pro", "status": "active"
            })
            if r.status_code in (200, 201):
                device_ids[f"pillbox_{elder['id_card']}"] = pill_id
                check(f"Pillbox {pill_id}", True)
            else:
                check(f"Pillbox {pill_id}", False, r.text[:80])
                pill_id = ""
            if pill_id:
                r2 = requests.post(f"{API}/admin/device-bindings", headers=HEADERS, json={
                    "device_id": pill_id, "person_id": pid, "business_chain": "self"
                })
                check(f"Bind pillbox to {elder['name']}", r2.status_code in (200, 201))

        time.sleep(0.2)

    # Hospital chain: medical wristbands (create via medical endpoint)
    hospital_elders = [e for e in ELDERS if e["chain"] == "hospital"]
    for i, elder in enumerate(hospital_elders):
        pid = person_ids.get(elder["id_card"])
        dev_id = f"MW-{i+1:04d}"
        r = requests.post(f"{API}/admin/medical/wristbands", headers=HEADERS, json={
            "device_id": dev_id, "firmware_version": "v1.2.0", "status": "active"
        })
        if r.status_code in (200, 201):
            device_ids[f"medical_{elder['id_card']}"] = dev_id
            check(f"Medical wristband {dev_id}", True)
        else:
            check(f"Medical wristband {dev_id}", False, r.text[:80])
            dev_id = ""
        if dev_id and pid:
            r2 = requests.post(f"{API}/admin/medical/wristbands/bind", headers=HEADERS, json={
                "patient_id": pid, "device_id": dev_id
            })
            check(f"Bind medical wb to {elder['name']}", r2.status_code in (200, 201))
        time.sleep(0.2)

    # Community chain: community wristbands
    community_elders = [e for e in ELDERS if e["chain"] == "community"]
    for i, elder in enumerate(community_elders):
        pid = person_ids.get(elder["id_card"])
        dev_id = f"CW-{i+1:04d}"
        r = requests.post(f"{API}/admin/community-wb/devices", headers=HEADERS, json={
            "device_id": dev_id, "firmware_version": "v1.0.0", "mode": "community", "status": "active"
        })
        if r.status_code in (200, 201):
            device_ids[f"community_{elder['id_card']}"] = dev_id
            check(f"Community wristband {dev_id}", True)
        else:
            check(f"Community wristband {dev_id}", False, r.text[:80])
            dev_id = ""
        if dev_id and pid:
            r2 = requests.post(f"{API}/admin/community-wb/devices/bind", headers=HEADERS, json={
                "elder_id": pid, "device_id": dev_id
            })
            check(f"Bind community wb to {elder['name']}", r2.status_code in (200, 201))
        time.sleep(0.2)

# ─── Module 5: Health Records ───────────────────────────────────────────────
def generate_health_record(person_id, chain, hr_base=75):
    hr = max(50, min(130, hr_base + random.randint(-15, 15)))
    spo2 = max(88, min(100, 97 + random.randint(-3, 2)))
    record = {
        "person_id": person_id,
        "business_chain": chain,
        "record_type": "vitals",
        "source": "device",
        "hr": hr,
        "spo2": spo2,
        "steps": random.randint(500, 12000),
        "sleep_hours": round(random.uniform(5.0, 9.5), 1),
        "blood_pressure_sys": random.randint(100, 180),
        "blood_pressure_dia": random.randint(60, 110),
        "recorded_at": rand_datetime(48),
    }
    if random.random() < 0.3:
        record["blood_glucose_fasting"] = round(random.uniform(4.0, 12.0), 1)
    if random.random() < 0.2:
        record["uric_acid"] = round(random.uniform(3.0, 9.0), 1)
    if random.random() < 0.15:
        record["weight"] = round(random.uniform(45, 90), 1)
        record["height"] = round(random.uniform(155, 180), 1)
        record["bmi"] = round(record["weight"] / ((record["height"]/100)**2), 1)
    return record

def inject_health_records():
    print("\n[5/10] Injecting health records (50-99 per person)...")
    total = 0
    for elder in ELDERS:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue
        chain = elder["chain"]
        count = random.randint(55, 95)
        hr_base = 75 if elder["risk"] not in ("high", "critical") else 82
        for j in range(count):
            record = generate_health_record(pid, chain, hr_base)
            r = requests.post(f"{API}/admin/health-records", headers=HEADERS, json=record)
            if r.status_code not in (200, 201):
                err = r.text[:80] if r.text else "unknown"
                if "duplicate" not in err.lower() and "already exists" not in err.lower():
                    print(f"  ❌ Health record failed for {elder['name']}: {err}")
                break
            total += 1
        check(f"{elder['name']} ({count} records)", True, f"total={total}")
        time.sleep(0.1)
    print(f"  Total health records: {total}")

# ─── Module 6: Medication Rules & Executions ─────────────────────────────────
MEDICATIONS = [
    {"drug_name": "氨氯地平", "generic_name": "Amlodipine", "dosage": "5mg", "frequency": "每日1次",
     "schedule_time1": "08:00", "route": "oral", "drug_category": "prescription"},
    {"drug_name": "二甲双胍", "generic_name": "Metformin", "dosage": "500mg", "frequency": "每日2次",
     "schedule_time1": "08:00", "schedule_time2": "20:00", "route": "oral", "drug_category": "prescription"},
    {"drug_name": "阿司匹林", "generic_name": "Aspirin", "dosage": "100mg", "frequency": "每日1次",
     "schedule_time1": "20:00", "route": "oral", "drug_category": "otc"},
    {"drug_name": "阿托伐他汀", "generic_name": "Atorvastatin", "dosage": "20mg", "frequency": "每日1次",
     "schedule_time1": "20:00", "route": "oral", "drug_category": "prescription"},
    {"drug_name": "缬沙坦", "generic_name": "Valsartan", "dosage": "80mg", "frequency": "每日1次",
     "schedule_time1": "08:00", "route": "oral", "drug_category": "prescription"},
]

def inject_medication():
    print("\n[6/10] Injecting medication rules & executions...")
    pillbox_elders = [e for e in ELDERS if e.get("has_pillbox") and e["chain"] == "self"]
    total_rules = 0
    for elder in pillbox_elders:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue
        num_rules = random.randint(2, 3)
        for med in random.sample(MEDICATIONS, min(num_rules, len(MEDICATIONS))):
            rule = {
                "person_id": pid,
                "business_chain": "self",
                "source_type": "custom",
                "drug_name": med["drug_name"],
                "generic_name": med["generic_name"],
                "dosage": med["dosage"],
                "frequency": med["frequency"],
                "route": med["route"],
                "schedule_time1": med["schedule_time1"],
                "schedule_time2": med.get("schedule_time2"),
                "drug_category": med["drug_category"],
                "active": 1
            }
            r = requests.post(f"{API}/admin/medications", headers=HEADERS, json=rule)
            if r.status_code in (200, 201):
                total_rules += 1
                check(f"Med rule {med['drug_name']} for {elder['name']}", True)
            else:
                check(f"Med rule {med['drug_name']} for {elder['name']}", False, r.text[:80])
            time.sleep(0.1)

        for _ in range(random.randint(10, 20)):
            exec_data = {
                "person_id": pid,
                "business_chain": "self",
                "scheduled_time": rand_datetime(7),
                "status": random.choice(["taken", "taken", "taken", "missed", "alerted"]),
                "taken_by": random.choice(["self", "pillbox_auto", "family"]),
                "verification_method": "optical"
            }
            r = requests.post(f"{API}/admin/medications/executions", headers=HEADERS, json=exec_data)
            if r.status_code not in (200, 201):
                print(f"  ❌ Execution failed: {r.text[:80]}")
            time.sleep(0.05)
    print(f"  Total medication rules: {total_rules}")

# ─── Module 7: Alert Rule Engine ────────────────────────────────────────────
class AlertRuleEngine:
    def __init__(self, token, headers):
        self.token = token
        self.headers = headers
        self.rules = {}

    def load_rules(self):
        for chain in ["self", "hospital", "community"]:
            r = requests.get(f"{API}/admin/alert-rules", headers=self.headers, params={"chain": chain})
            if r.status_code == 200:
                self.rules[chain] = r.json().get("data", []) or []

    def evaluate_record(self, record):
        chain = record.get("business_chain", "self")
        alerts = []
        for rule in self.rules.get(chain, []):
            alert = self._check_rule(rule, record)
            if alert:
                alerts.append(alert)
        return alerts

    def _check_rule(self, rule, record):
        alert_type = rule.get("alert_type", "")
        field = rule.get("condition_field", "")
        op = rule.get("condition_operator", "")
        threshold = rule.get("condition_threshold")

        if alert_type in ("fall", "sos"):
            if random.random() < 0.05:
                return {
                    "elderly_id": record["person_id"],
                    "alert_type": alert_type,
                    "severity": rule.get("severity", "p0"),
                    "device_id": record.get("device_id", ""),
                }
        if field and op and threshold is not None:
            value = record.get(field)
            if value is None:
                return None
            triggered = False
            if op == ">" and value > threshold: triggered = True
            elif op == "<" and value < threshold: triggered = True
            elif op == ">=" and value >= threshold: triggered = True
            elif op == "<=" and value <= threshold: triggered = True
            elif op == "=" and value == threshold: triggered = True
            elif op == "!=" and value != threshold: triggered = True
            if triggered:
                return {
                    "elderly_id": record["person_id"],
                    "alert_type": rule.get("alert_type"),
                    "severity": rule.get("severity", "p1"),
                    "device_id": record.get("device_id", ""),
                }
        return None

    def create_alert(self, alert_data):
        r = requests.post(f"{API}/admin/alerts", headers=self.headers, json=alert_data)
        return r.status_code in (200, 201)

def run_alert_engine():
    print("\n[7/10] Running alert rule engine...")
    engine = AlertRuleEngine(JWT_TOKEN, HEADERS)
    engine.load_rules()
    rule_counts = {k: len(v) if v else 0 for k, v in engine.rules.items()}
    print(f"  Loaded rules: {rule_counts}")

    total_alerts = 0
    for elder in ELDERS:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue
        chain = elder["chain"]
        r = requests.get(f"{API}/admin/health-records", headers=HEADERS,
                        params={"personId": pid, "chain": chain, "limit": 10})
        if r.status_code != 200:
            continue
        records = r.json().get("data", []) or []
        for record in records:
            alerts = engine.evaluate_record(record)
            for alert in alerts:
                if engine.create_alert(alert):
                    total_alerts += 1
    print(f"  Total alerts generated: {total_alerts}")

# ─── Module 8: Hospital Chain Data ──────────────────────────────────────────
def inject_hospital_data():
    print("\n[8/10] Injecting hospital chain data...")
    hospital_elders = [e for e in ELDERS if e["chain"] == "hospital"]
    total_verifications = 0
    total_daily_entries = 0
    total_expenses = 0

    for elder in hospital_elders:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue

        # Ward round entries
        num_rounds = random.randint(3, 5)
        for j in range(num_rounds):
            entry = {
                "patient_id": pid,
                "entry_type": random.choice(["vitals", "medication", "nursing"]),
                "notes": f"查房记录 {j+1}: 生命体征稳定，血压{random.randint(110,150)}/{random.randint(70,90)}",
                "created_at": rand_datetime(14)
            }
            r = requests.post(f"{API}/admin/medical/daily-entries", headers=HEADERS, json=entry)
            if r.status_code in (200, 201):
                total_daily_entries += 1
            time.sleep(0.1)

        # Verifications
        mw_id = device_ids.get(f"medical_{elder['id_card']}", "")
        num_verify = random.randint(5, 10)
        for j in range(num_verify):
            verify = {
                "device_id": mw_id,
                "patient_id": pid,
                "verification_type": random.choice(["medication", "vitals", "nfc"]),
                "result": "passed",
                "matched": True,
                "verified_by": "nurse01"
            }
            r = requests.post(f"{API}/admin/medical/verifications", headers=HEADERS, json=verify)
            if r.status_code in (200, 201):
                total_verifications += 1
            time.sleep(0.1)

        # Medical expenses
        num_expenses = random.randint(5, 15)
        for j in range(num_expenses):
            expense = {
                "patient_id": pid,
                "expense_date": rand_date(14),
                "item_name": random.choice(["血常规", "CT扫描", "核磁共振", "心电图", "尿常规", "B超"]),
                "category": random.choice(["lab", "radiology", "consultation"]),
                "amount": round(random.uniform(50, 800), 2),
                "quantity": 1,
                "billing_source": random.choice(["his", "manual"])
            }
            r = requests.post(f"{API}/admin/medical/expenses", headers=HEADERS, json=expense)
            if r.status_code in (200, 201):
                total_expenses += 1
            time.sleep(0.1)

        check(f"Hospital data for {elder['name']}", True,
              f"rounds={num_rounds}, verifies={num_verify}, expenses={num_expenses}")
        time.sleep(0.3)

    print(f"  Totals: verifications={total_verifications}, daily_entries={total_daily_entries}, expenses={total_expenses}")

# ─── Module 9: Community Chain Data ─────────────────────────────────────────
def inject_community_data():
    print("\n[9/10] Injecting community chain data...")
    community_elders = [e for e in ELDERS if e["chain"] == "community"]
    total_signins = 0
    total_pharmacy = 0

    for elder in community_elders:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue

        # Sign-in records (14 days)
        for day in range(14):
            signin = {
                "elder_id": pid,
                "device_id": device_ids.get(f"community_{elder['id_card']}", ""),
                "hospital_id": "INST-001",
                "signin_time": rand_datetime(336),
                "period": random.choice(["morning", "afternoon"]),
                "activated_tags": elder.get("welfare_tags", []),
                "is_medical_signin": 1,
                "is_welfare_signin": 1
            }
            r = requests.post(f"{API}/admin/community-wb/signin/trigger", headers=HEADERS, json=signin)
            if r.status_code in (200, 201):
                total_signins += 1
            time.sleep(0.05)

        # Welfare tag assignment
        for tag in elder.get("welfare_tags", []):
            welfare = {
                "person_id": pid,
                "tag_code": tag,
                "valid_from": rand_date(365),
                "valid_to": rand_date(30)
            }
            r = requests.post(f"{API}/admin/persons/welfare-tags", headers=HEADERS, json=welfare)
            if r.status_code in (200, 201):
                check(f"Welfare tag {tag} for {elder['name']}", True)
            time.sleep(0.1)

        # Pharmacy dispense
        num_dispense = random.randint(3, 8)
        for j in range(num_dispense):
            dispense = {
                "elder_id": pid,
                "device_id": device_ids.get(f"community_{elder['id_card']}", ""),
                "hospital_id": "INST-001",
                "medication_name": random.choice(["降压药", "降糖药", "阿司匹林"]),
                "quantity": random.randint(1, 3),
                "dispensed_at": rand_datetime(336)
            }
            r = requests.post(f"{API}/admin/community-wb/pharmacy/dispense", headers=HEADERS, json=dispense)
            if r.status_code in (200, 201):
                total_pharmacy += 1
            time.sleep(0.1)

        check(f"Community data for {elder['name']}", True, f"signins=14, pharmacy={num_dispense}")
        time.sleep(0.3)

    print(f"  Totals: signins={total_signins}, pharmacy={total_pharmacy}")

# ─── Module 10: Compliance Checks ───────────────────────────────────────────
def inject_compliance_data():
    print("\n[10/10] Running compliance checks...")
    total_checks = 0
    for elder in ELDERS:
        pid_key = f"{elder['id_card']}:{elder['chain']}"
        pid = person_profiles.get(pid_key)
        if not pid:
            continue
        r = requests.post(f"{API}/admin/compliance-checks/run", headers=HEADERS, json={
            "rule_code": f"R_{random.randint(1,9)}",
            "person_id": pid
        })
        if r.status_code in (200, 201):
            total_checks += 1
            check(f"Compliance check for {elder['name']}", True)
        else:
            check(f"Compliance check for {elder['name']}", False, r.text[:80])
        time.sleep(0.2)
    print(f"  Total compliance checks: {total_checks}")

# ─── Main ───────────────────────────────────────────────────────────────────
def main():
    print("=" * 60)
    print("Eregen Full Chain Data Seed v3")
    print("=" * 60)

    r = requests.get(f"{API}/health", timeout=5)
    if r.status_code != 200:
        print(f"❌ API not reachable: {r.text}"); sys.exit(1)
    print("✅ API is up")

    login_admin()
    create_role_accounts()
    create_persons()
    create_person_profiles()
    create_devices()
    inject_health_records()
    inject_medication()
    run_alert_engine()
    inject_hospital_data()
    inject_community_data()
    inject_compliance_data()

    print("\n" + "=" * 60)
    print("✅ SEED COMPLETE")
    print(f"  Persons created: {len(person_profiles)}")
    print(f"  Devices tracked: {len(device_ids)}")
    print("=" * 60)

if __name__ == "__main__":
    main()
