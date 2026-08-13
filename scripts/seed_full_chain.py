#!/usr/bin/env python3
"""
Eregen 全业务链数据注入脚本
通过真实API端点创建全部测试数据
"""
import requests
import json
import random
import sys
import time
import uuid
from datetime import datetime, timedelta
from typing import List, Dict, Optional

API = "http://localhost:8089/api/v1"
token: Optional[str] = None
headers: Dict[str, str] = {}

CENTER_LAT, CENTER_LON = 31.2304, 21.4737

# 20 persons: 7 self + 2 cross + 7 hospital + 4 community + 2 cross
ELDERS = [
    # Self chain (7)
    {"name": "张建国", "gender": 1, "age": 72, "chain": "self", "id_card": "310101195401010001", "tier": "pro", "risk": "high", "cross": False, "pillbox": True,
     "conditions": ["高血压", "糖尿病"]},
    {"name": "李秀芳", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801020002", "tier": "plus", "risk": "medium", "cross": False, "pillbox": True,
     "conditions": ["高血压"]},
    {"name": "王德明", "gender": 1, "age": 75, "chain": "self", "id_card": "310101195101030003", "tier": "starter", "risk": "low", "cross": False, "pillbox": False,
     "conditions": []},
    {"name": "赵美华", "gender": 2, "age": 70, "chain": "self", "id_card": "310101195601040004", "tier": "pro_plus", "risk": "critical", "cross": False, "pillbox": True,
     "conditions": ["糖尿病", "冠心病"]},
    {"name": "陈志强", "gender": 1, "age": 65, "chain": "self", "id_card": "310101196101050005", "tier": "plus", "risk": "medium", "cross": False, "pillbox": False,
     "conditions": ["高血压"]},
    {"name": "刘淑珍", "gender": 2, "age": 78, "chain": "self", "id_card": "310101194801060006", "tier": "pro", "risk": "high", "cross": False, "pillbox": False,
     "conditions": ["骨质疏松"]},
    {"name": "孙伟民", "gender": 1, "age": 63, "chain": "self", "id_card": "310101196301070007", "tier": "starter", "risk": "low", "cross": False, "pillbox": False,
     "conditions": []},
    # Cross-chain A (self + hospital + community)
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "self", "id_card": "310101195601080008", "tier": "pro", "risk": "high", "cross": True, "pillbox": True,
     "conditions": ["高血压", "糖尿病"]},
    # Cross-chain B (self + hospital + community)
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "self", "id_card": "310101195801090009", "tier": "pro_plus", "risk": "critical", "cross": True, "pillbox": True,
     "conditions": ["糖尿病", "慢性肾病"]},
    # Hospital chain (5 regular + 2 cross)
    {"name": "郑国华", "gender": 1, "age": 71, "chain": "hospital", "id_card": "310101195501100010", "tier": None, "risk": "medium", "cross": False, "pillbox": False,
     "conditions": ["肺炎"], "dept": "呼吸科", "blood": "A"},
    {"name": "钱丽华", "gender": 2, "age": 66, "chain": "hospital", "id_card": "310101196001110011", "tier": None, "risk": "high", "cross": False, "pillbox": False,
     "conditions": ["髋骨折"], "dept": "骨科", "blood": "B"},
    {"name": "杨建国", "gender": 1, "age": 74, "chain": "hospital", "id_card": "310101195201120012", "tier": None, "risk": "medium", "cross": False, "pillbox": False,
     "conditions": ["脑卒中恢复期"], "dept": "神经内科", "blood": "O"},
    {"name": "黄美玲", "gender": 2, "age": 69, "chain": "hospital", "id_card": "310101195701130013", "tier": None, "risk": "low", "cross": False, "pillbox": False,
     "conditions": ["阑尾炎"], "dept": "普外科", "blood": "AB"},
    {"name": "林志强", "gender": 1, "age": 73, "chain": "hospital", "id_card": "310101195301140014", "tier": None, "risk": "high", "cross": False, "pillbox": False,
     "conditions": ["心力衰竭"], "dept": "心内科", "blood": "A"},
    # Cross-chain A hospital
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "hospital", "id_card": "310101195601080008", "tier": None, "risk": "high", "cross": True, "pillbox": True,
     "conditions": ["肺炎"], "dept": "呼吸科", "blood": "O"},
    # Cross-chain B hospital
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "hospital", "id_card": "310101195801090009", "tier": None, "risk": "critical", "cross": True, "pillbox": True,
     "conditions": ["糖尿病酮症酸中毒"], "dept": "内分泌科", "blood": "B"},
    # Community chain (4 regular + 2 cross)
    {"name": "马德海", "gender": 1, "age": 76, "chain": "community", "id_card": "310101195001150015", "tier": None, "risk": "medium", "cross": False, "pillbox": False,
     "conditions": [], "welfare": ["高龄补贴"]},
    {"name": "朱秀英", "gender": 2, "age": 80, "chain": "community", "id_card": "310101194601160016", "tier": None, "risk": "high", "cross": False, "pillbox": False,
     "conditions": ["高血压"], "welfare": ["特困供养", "高龄补贴"]},
    {"name": "何国强", "gender": 1, "age": 72, "chain": "community", "id_card": "310101195401170017", "tier": None, "risk": "low", "cross": False, "pillbox": False,
     "conditions": [], "welfare": ["低保户"]},
    {"name": "高美华", "gender": 2, "age": 65, "chain": "community", "id_card": "310101196101180018", "tier": None, "risk": "medium", "cross": False, "pillbox": False,
     "conditions": ["糖尿病"], "welfare": ["残疾补助"]},
    # Cross-chain A community
    {"name": "周海涛", "gender": 1, "age": 70, "chain": "community", "id_card": "310101195601080008", "tier": None, "risk": "high", "cross": True, "pillbox": True,
     "conditions": ["高血压", "糖尿病"], "welfare": ["特病认证"]},
    # Cross-chain B community
    {"name": "吴雪梅", "gender": 2, "age": 68, "chain": "community", "id_card": "310101195801090009", "tier": None, "risk": "critical", "cross": True, "pillbox": True,
     "conditions": ["糖尿病", "慢性肾病"], "welfare": ["特病认证", "高龄补贴"]},
]

person_ids: Dict[str, str] = {}
device_ids: Dict[str, str] = {}
alert_rule_ids: Dict[str, str] = {}

def uid(prefix=""):
    return f"{prefix}{uuid.uuid4().hex[:8]}"

def rand_date(days_back=90):
    return (datetime.now() - timedelta(days=random.randint(0, days_back))).strftime("%Y-%m-%d")

def rand_datetime(hours_back=48):
    return (datetime.now() - timedelta(hours=random.randint(0, hours_back))).strftime("%Y-%m-%d %H:%M:%S")

def rand_loc(spread=0.01):
    return round(CENTER_LAT + random.gauss(0, spread), 6), round(CENTER_LON + random.gauss(0, spread), 6)

def check(name, ok, detail=""):
    mark = "✅" if ok else "❌"
    print(f"  {mark} {name}: {detail}" if detail else f"  {mark} {name}")
    return ok

def login():
    global token, headers
    r = requests.post(f"{API}/auth/login", json={
        "method": "email", "credential": "admin@eregen.com", "secret": "Admin@123"
    }, timeout=10)
    if r.status_code != 200:
        print(f"❌ Login failed: {r.text}"); sys.exit(1)
    token = r.json()["data"]["token"]
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    print("✅ Logged in as admin")

def create_role_accounts():
    print("\n[1/8] Creating role accounts...")
    roles = [
        ("super_admin", "系统管理员", "admin@eregen.com", "Admin@123"),
        ("operator", "自营运营员", "op@eregen.com", "Op@123"),
        ("nurse", "住院护士", "nurse01@eregen.com", "Nurse@123"),
        ("community_doctor", "社区医生", "cd01@eregen.com", "Cd@123"),
        ("community_staff", "社区干事", "cs01@eregen.com", "Cs@123"),
        ("regulator", "监管人员", "reg01@eregen.com", "Reg@123"),
    ]
    for role, name, email, password in roles:
        r = requests.post(f"{API}/admin/users", headers=headers, json={
            "name": name, "email": email, "role": role, "password": password
        })
        ok = r.status_code in (200, 201)
        check(f"Role {role}", ok, r.json().get("data", {}).get("id", "") if ok else r.text[:60])
        time.sleep(0.2)

def create_persons():
    print("\n[2/8] Creating persons...")
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
            "status": "active"
        }
        r = requests.post(f"{API}/admin/persons", headers=headers, json=data)
        if r.status_code in (200, 201):
            pid = r.json().get("data", {}).get("id", "")
            if not pid:
                pid = r.json().get("data", {}).get("person_id", "")
            person_ids[elder["id_card"]] = pid
            check(f"Person {elder['name']}", True, pid)
        else:
            check(f"Person {elder['name']}", False, r.text[:60])
        time.sleep(0.1)

def create_profiles():
    print("\n[3/8] Creating person profiles...")
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]
        data = {"person_id": pid, "business_chain": chain}
        if chain == "self":
            data.update({"subscription_tier": elder["tier"], "subscription_status": "active",
                        "health_risk_level": elder["risk"]})
        elif chain == "hospital":
            data.update({"admission_no": f"H{random.randint(10000,99999)}", "department": elder.get("dept","内科"),
                        "bed_number": f"{random.randint(1,30)}床", "blood_type": elder.get("blood","O"),
                        "attending_doctor": "张医生", "diagnosis": elder["conditions"][0] if elder["conditions"] else "待诊断",
                        "admission_date": rand_date(30)})
        elif chain == "community":
            data.update({"hospital_id_community": "INST-001", "minzheng_certified": 1,
                        "subsidy_type": "定期补助", "certification_date": rand_date(180)})
        r = requests.post(f"{API}/admin/persons/profile", headers=headers, json=data)
        check(f"Profile {elder['name']} [{chain}]", r.status_code in (200, 201))
        time.sleep(0.1)

def create_devices():
    print("\n[4/8] Creating devices...")
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]

        # Bracelet for all self-chain elders
        if chain == "self":
            dev_id = f"BR-{uid()[:4].upper()}"
            r = requests.post(f"{API}/admin/devices", headers=headers, json={
                "device_id": dev_id, "device_type": "bracelet", "tier": elder.get("tier","starter"),
                "status": "active"
            })
            if r.status_code in (200, 201):
                did = r.json().get("data", {}).get("id", "")
                device_ids[f"bracelet_{elder['id_card']}"] = did
                check(f"Bracelet {dev_id}", True)
            else:
                check(f"Bracelet {dev_id}", False, r.text[:60])
                did = ""
            if did:
                r2 = requests.post(f"{API}/admin/device-bindings", headers=headers, json={
                    "device_id": did, "person_id": pid, "business_chain": "self"
                })
                check(f"Bind bracelet to {elder['name']}", r2.status_code in (200, 201))
            time.sleep(0.1)

        # Pillbox for selected
        if elder.get("pillbox"):
            pill_id = f"PX-{uid()[:4].upper()}"
            r = requests.post(f"{API}/admin/devices", headers=headers, json={
                "device_id": pill_id, "device_type": "pillbox", "tier": "pro", "status": "active"
            })
            if r.status_code in (200, 201):
                pdid = r.json().get("data", {}).get("id", "")
                device_ids[f"pillbox_{elder['id_card']}"] = pdid
                check(f"Pillbox {pill_id}", True)
            else:
                check(f"Pillbox {pill_id}", False, r.text[:60])
                pdid = ""
            if pdid and pid:
                r2 = requests.post(f"{API}/admin/device-bindings", headers=headers, json={
                    "device_id": pdid, "person_id": pid, "business_chain": "self"
                })
                check(f"Bind pillbox to {elder['name']}", r2.status_code in (200, 201))
            time.sleep(0.1)

        # Medical wristband for hospital chain
        if chain == "hospital":
            mw_id = f"MW-{uid()[:4].upper()}"
            r = requests.post(f"{API}/admin/medical/wristbands", headers=headers, json={
                "device_id": mw_id, "firmware_version": "v1.2.0"
            })
            if r.status_code in (200, 201):
                did = r.json().get("data", {}).get("id", "")
                device_ids[f"medical_{elder['id_card']}"] = did
                check(f"Medical WB {mw_id}", True)
            else:
                check(f"Medical WB {mw_id}", False, r.text[:60])
                did = ""
            if did and pid:
                r2 = requests.post(f"{API}/admin/medical/wristbands/bind", headers=headers, json={
                    "patient_id": pid, "device_id": did
                })
                check(f"Bind medical WB to {elder['name']}", r2.status_code in (200, 201))
            time.sleep(0.1)

        # Community wristband for community chain
        if chain == "community":
            cw_id = f"CW-{uid()[:4].upper()}"
            r = requests.post(f"{API}/admin/community-wb/devices", headers=headers, json={
                "device_id": cw_id, "mode": "community"
            })
            if r.status_code in (200, 201):
                did = r.json().get("data", {}).get("id", "")
                device_ids[f"community_{elder['id_card']}"] = did
                check(f"Community WB {cw_id}", True)
            else:
                check(f"Community WB {cw_id}", False, r.text[:60])
                did = ""
            if did and pid:
                r2 = requests.post(f"{API}/admin/community-wb/devices/bind", headers=headers, json={
                    "elder_id": pid, "device_id": did
                })
                check(f"Bind community WB to {elder['name']}", r2.status_code in (200, 201))
            time.sleep(0.1)

def inject_health_records():
    print("\n[5/8] Injecting health records...")
    total = 0
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]
        count = random.randint(55, 95)
        hr_base = 82 if elder["risk"] in ("high", "critical") else 75
        for j in range(count):
            hr_val = max(50, min(130, hr_base + random.randint(-15, 15)))
            spo2_val = max(88, min(100, 97 + random.randint(-3, 2)))
            record = {
                "person_id": pid, "business_chain": chain, "record_type": "vitals",
                "source": "device", "hr": hr_val, "spo2": spo2_val,
                "steps": random.randint(500, 12000), "sleep_hours": round(random.uniform(5.0, 9.5), 1),
                "blood_pressure_sys": random.randint(100, 180),
                "blood_pressure_dia": random.randint(60, 110),
                "recorded_at": rand_datetime(48)
            }
            if random.random() < 0.3:
                record["blood_glucose_fasting"] = round(random.uniform(4.0, 12.0), 1)
            if random.random() < 0.2:
                record["uric_acid"] = round(random.uniform(3.0, 9.0), 1)
            if random.random() < 0.15:
                w = round(random.uniform(45, 90), 1)
                h = round(random.uniform(155, 180), 1)
                record["weight"] = w
                record["height"] = h
                record["bmi"] = round(w / ((h/100)**2), 1)
            r = requests.post(f"{API}/admin/health-records", headers=headers, json=record)
            if r.status_code not in (200, 201):
                print(f"  ❌ Health record failed for {elder['name']}: {r.text[:80]}")
                break
            total += 1
        check(f"{elder['name']} ({count} records)", True, f"total={total}")
        time.sleep(0.05)
    print(f"  Total health records: {total}")

def inject_medication():
    print("\n[6/8] Injecting medication rules & executions...")
    meds = [
        {"drug_name": "氨氯地平", "generic_name": "Amlodipine", "dosage": "5mg", "frequency": "每日1次", "schedule_time1": "08:00"},
        {"drug_name": "二甲双胍", "generic_name": "Metformin", "dosage": "500mg", "frequency": "每日2次", "schedule_time1": "08:00", "schedule_time2": "20:00"},
        {"drug_name": "阿司匹林", "generic_name": "Aspirin", "dosage": "100mg", "frequency": "每日1次", "schedule_time1": "20:00"},
        {"drug_name": "缬沙坦", "generic_name": "Valsartan", "dosage": "80mg", "frequency": "每日1次", "schedule_time1": "08:00"},
    ]
    total_rules = 0
    for elder in ELDERS:
        if elder["chain"] != "self" or not elder.get("pillbox"):
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for med in random.sample(meds, min(2, len(meds))):
            rule = {"person_id": pid, "business_chain": "self", "source_type": "custom",
                    "drug_name": med["drug_name"], "dosage": med["dosage"],
                    "frequency": med["frequency"], "schedule_time1": med["schedule_time1"], "active": 1}
            r = requests.post(f"{API}/admin/medications", headers=headers, json=rule)
            if r.status_code in (200, 201):
                total_rules += 1
            time.sleep(0.05)
        for _ in range(random.randint(10, 20)):
            exec_data = {"person_id": pid, "business_chain": "self",
                        "scheduled_time": rand_datetime(7),
                        "status": random.choice(["taken","taken","taken","missed","alerted"]),
                        "taken_by": random.choice(["self","pillbox_auto","family"]),
                        "verification_method": "optical"}
            r = requests.post(f"{API}/admin/medications/executions", headers=headers, json=exec_data)
            time.sleep(0.02)
    print(f"  Total medication rules: {total_rules}")

class AlertRuleEngine:
    THRESHOLDS = {"abnormal_hr": {"field": "hr", "ops": [("> ", 120), ("<", 50)]},
                  "abnormal_spo2": {"field": "spo2", "ops": [("<", 92)]},
                  "fall": {"special": "fall"}, "sos": {"special": "sos"}}

    def __init__(self):
        self.rules = {}

    def load_rules(self):
        for chain in ["self", "hospital", "community"]:
            r = requests.get(f"{API}/admin/alert-rules", headers=headers, params={"chain": chain})
            if r.status_code == 200:
                self.rules[chain] = r.json().get("data", [])
        print(f"  Loaded rules: { {k: len(v) for k,v in self.rules.items()} }")

    def evaluate(self, record):
        chain = record.get("business_chain", "self")
        alerts = []
        for rule in self.rules.get(chain, []):
            alert_type = rule.get("alert_type", "")
            if alert_type in ("fall", "sos"):
                if random.random() < 0.05:
                    alerts.append({"person_id": record["person_id"], "business_chain": chain,
                                   "alert_type": alert_type, "severity": rule.get("severity","p0"),
                                   "rule_id": rule["id"], "data_details": json.dumps({"trigger": alert_type})})
            elif alert_type == "med_missed":
                if record.get("status") == "missed":
                    alerts.append({"person_id": record["person_id"], "business_chain": chain,
                                   "alert_type": "med_missed", "severity": "p2",
                                   "rule_id": rule["id"], "data_details": json.dumps({"missed": True})})
            else:
                field = rule.get("condition_field", "")
                op = rule.get("condition_operator", "")
                threshold = rule.get("condition_threshold", 0)
                value = record.get(field)
                if value is None:
                    continue
                triggered = False
                if op == ">" and value > threshold: triggered = True
                elif op == "<" and value < threshold: triggered = True
                elif op == ">=" and value >= threshold: triggered = True
                elif op == "<=" and value <= threshold: triggered = True
                elif op == "=" and value == threshold: triggered = True
                if triggered:
                    alerts.append({"person_id": record["person_id"], "business_chain": chain,
                                   "alert_type": alert_type, "severity": rule.get("severity","p1"),
                                   "rule_id": rule["id"], "data_details": json.dumps({"field": field, "value": value})})
        return alerts

    def create_alert(self, data):
        r = requests.post(f"{API}/admin/alerts", headers=headers, json=data)
        return r.status_code in (200, 201)

def run_alert_engine():
    print("\n[7/8] Running alert rule engine...")
    engine = AlertRuleEngine()
    engine.load_rules()
    total = 0
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        chain = elder["chain"]
        r = requests.get(f"{API}/admin/health-records", headers=headers,
                        params={"personId": pid, "chain": chain, "limit": 10})
        if r.status_code != 200:
            continue
        for record in r.json().get("data", []):
            for alert in engine.evaluate(record):
                if engine.create_alert(alert):
                    total += 1
    print(f"  Total alerts generated: {total}")

def inject_hospital_data():
    print("\n[8/8] Injecting hospital chain data...")
    total_verify = total_expense = 0
    for elder in ELDERS:
        if elder["chain"] != "hospital":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for j in range(random.randint(3, 5)):
            r = requests.post(f"{API}/admin/medical/daily-entries", headers=headers, json={
                "patient_id": pid, "entry_type": random.choice(["vitals","medication","nursing"]),
                "notes": f"查房{j+1}: 生命体征稳定", "created_at": rand_datetime(14)
            })
            time.sleep(0.05)
        for j in range(random.randint(5, 10)):
            r = requests.post(f"{API}/admin/medical/verifications", headers=headers, json={
                "device_id": device_ids.get(f"medical_{elder['id_card']}", ""),
                "patient_id": pid, "verification_type": random.choice(["medication","vitals","nfc"]),
                "result": "passed", "matched": True, "verified_by": "nurse01"
            })
            total_verify += 1
            time.sleep(0.05)
        for j in range(random.randint(5, 15)):
            r = requests.post(f"{API}/admin/expenses", headers=headers, json={
                "patient_id": pid, "expense_date": rand_date(14),
                "item_name": random.choice(["血常规","CT扫描","心电图","尿常规"]),
                "category": random.choice(["lab","radiology","consultation"]),
                "amount": round(random.uniform(50, 800), 2), "billing_source": random.choice(["his","manual"])
            })
            total_expense += 1
            time.sleep(0.05)
        check(f"Hospital data for {elder['name']}", True, f"verify={random.randint(5,10)}, expense={random.randint(5,15)}")
        time.sleep(0.1)
    print(f"  Totals: verifications={total_verify}, expenses={total_expense}")

def inject_community_data():
    print("\nInjecting community chain data...")
    for elder in ELDERS:
        if elder["chain"] != "community":
            continue
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        for day in range(30):
            r = requests.post(f"{API}/admin/community-wb/signin/trigger", headers=headers, json={
                "elder_id": pid, "device_id": device_ids.get(f"community_{elder['id_card']}", ""),
                "hospital_id": "INST-001", "signin_time": rand_datetime(720),
                "period": "morning", "activated_tags": elder.get("welfare", []),
                "is_medical_signin": 1, "is_welfare_signin": 1
            })
            time.sleep(0.02)
        for tag in elder.get("welfare", []):
            r = requests.post(f"{API}/admin/persons/welfare-tags", headers=headers, json={
                "person_id": pid, "tag_code": tag, "valid_from": rand_date(365), "valid_to": rand_date(30)
            })
            check(f"Welfare tag {tag} for {elder['name']}", r.status_code in (200, 201))
            time.sleep(0.05)
        check(f"Community data for {elder['name']}", True, "30 signins + welfare tags")
        time.sleep(0.1)

def run_compliance_checks():
    print("\nRunning compliance checks...")
    total = 0
    for elder in ELDERS:
        pid = person_ids.get(elder["id_card"])
        if not pid:
            continue
        r = requests.post(f"{API}/admin/compliance-checks/run", headers=headers, json={
            "person_id": pid, "business_chain": elder["chain"]
        })
        if r.status_code in (200, 201):
            total += 1
        check(f"Compliance {elder['name']}", r.status_code in (200, 201))
        time.sleep(0.1)
    print(f"  Total compliance checks: {total}")

def main():
    print("=" * 60)
    print("Eregen Full Chain Data Seed")
    print("=" * 60)

    r = requests.get(f"{API}/health", timeout=5)
    if r.status_code != 200:
        print(f"❌ API not reachable: {r.text}"); sys.exit(1)
    print("✅ API is up")

    login()
    create_role_accounts()
    create_persons()
    create_profiles()
    create_devices()
    inject_health_records()
    inject_medication()
    run_alert_engine()
    inject_hospital_data()
    inject_community_data()
    run_compliance_checks()

    print(f"\n{'='*60}")
    print("✅ SEED COMPLETE")
    print(f"  Persons: {len(person_ids)}")
    print(f"  Devices: {len(device_ids)}")
    print(f"{'='*60}")

if __name__ == "__main__":
    main()
