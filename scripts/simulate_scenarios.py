#!/usr/bin/env python3
"""
Eregen 三场景真实数据模拟脚本
==============================
模拟三种终端的业务链：
  1. 普通手环（10人）—— 健康遥测 + 告警 + 慢病数据
  2. 住院业务（≤10人）—— HIS 生命体征/诊断/用药 + 护士核验巡房
  3. 社区业务（≤10人）—— 福利标签 + 签到 + 药房配药 + 健康检查

所有数据通过真实 API 接口上送，不直接写数据库。
"""
import requests
import json
import time
import random
import sys
import uuid
from datetime import datetime, timedelta
from typing import List, Tuple, Optional

# ─── 配置 ───────────────────────────────────────────────────────────
ADMIN_API = "http://localhost:8089/api/v1"        # admin-api
HOSPITAL_B2B = "http://localhost:8082/api/v2/b2b" # hospital-api
COMMUNITY_B2B = "http://localhost:8083/api/v2/b2b" # community-platform

# 上海陆家嘴为中心坐标
LAT_C, LON_C = 31.2304, 121.4737

# ─── 状态 ───────────────────────────────────────────────────────────
token: Optional[str] = None
headers: dict = {}

# 实体记录
bracelet_elders: List[Tuple[str, str, str]] = []  # (elderly_id, name, device_id)
bracelet_devices: List[Tuple[str, str, str]] = []  # (device_internal_id, elderly_id, device_id)
medical_patients: List[Tuple[str, str, str, str]] = []  # (elderly_id, patient_id, name, dept)
medical_devices: List[Tuple[str, str]] = []  # (device_id, patient_id)
community_elders: List[Tuple[str, str, str]] = []  # (elderly_id, name, id_card)
community_devices: List[Tuple[str, str]] = []  # (device_id, elder_id)
hospital_api_key: str = ""
hospital_inst_id: str = ""

stats = {
    "bracelet_health": 0, "bracelet_location": 0, "bracelet_heartbeat": 0,
    "bracelet_alerts": 0, "bracelet_chronic": 0,
    "hospital_vitals": 0, "hospital_diagnoses": 0, "hospital_meds": 0,
    "hospital_verifications": 0, "hospital_ward_rounds": 0,
    "hospital_admissions": 0, "hospital_reports": 0,
    "community_signins": 0, "community_pharmacy": 0,
    "community_health_checks": 0, "community_care_plans": 0,
    "community_tags": 0,
}


# ─── 工具函数 ───────────────────────────────────────────────────────
def uid(prefix=""):
    return f"{prefix}{uuid.uuid4().hex[:8]}"

def rand_date(days_back=365):
    return (datetime.now() - timedelta(days=random.randint(0, days_back))).strftime("%Y-%m-%d")

def rand_datetime_hours_back(hours_back=48):
    return (datetime.now() - timedelta(hours=random.randint(0, hours_back))).strftime("%Y-%m-%dT%H:%M:%S")

def rand_loc(center_lat, center_lon, spread=0.01):
    return round(center_lat + random.gauss(0, spread), 6), \
           round(center_lon + random.gauss(0, spread), 6)

def check(name, ok, detail=""):
    mark = "✅" if ok else "❌"
    print(f"  {mark} {name}: {detail}")
    return ok

def api_get(url, extra_headers=None):
    h = {**headers, **(extra_headers or {})}
    r = requests.get(url, headers=h, timeout=10)
    return r

def api_post(url, json=None, extra_headers=None):
    h = {**headers, "Content-Type": "application/json", **(extra_headers or {})}
    r = requests.post(url, headers=h, json=json, timeout=10)
    return r

def api_put(url, json=None, extra_headers=None):
    h = {**headers, "Content-Type": "application/json", **(extra_headers or {})}
    r = requests.put(url, headers=h, json=json, timeout=10)
    return r


# ═══════════════════════════════════════════════════════════════════
# 场景一：普通手环（10人）
# ═══════════════════════════════════════════════════════════════════

# 老人姓名池
SURNAMES = ["张","李","王","陈","刘","杨","黄","赵","吴","周"]
FIRST_NAMES = ["伟","芳","明","兰","军","梅","红","杰","玲","强"]

# 健康层级（疾病类型）
DISEASE_TIERS = [
    ["cardiovascular"],           # 心血管
    ["diabetes"],                 # 糖尿病
    ["cardiovascular","diabetes"],# 心血管+糖尿病
    ["hypertension"],             # 高血压
    ["fall_risk"],                # 跌倒风险
    ["respiratory"],              # 呼吸系统
    ["cardiovascular","respiratory"],
    ["diabetes","hypertension"],
    ["dementia"],                 # 认知障碍
    ["multiple"],                 # 多重慢病
]

# 特病老人索引（0-9中抽取2-3人）
SEVERE_INDICES = [2, 7, 9]  # 心血管+糖尿病, 糖尿病+高血压, 多重慢病


def simulate_bracelet():
    """模拟普通手环业务链"""
    print("\n" + "═" * 65)
    print("  场景一：普通手环遥测（10位老人）")
    print("═" * 65)

    base_lat, base_lon = rand_loc(LAT_C, LON_C, 0.03)

    for i in range(10):
        elderly_id, name, user_id = bracelet_elders[i]
        device_id, _, dev_id = bracelet_devices[i]
        is_severe = i in SEVERE_INDICES

        # 基础生理参数
        base_hr = random.randint(65, 90) if not is_severe else random.randint(80, 110)
        base_spo2 = random.randint(96, 99) if not is_severe else random.randint(88, 94)
        base_steps = random.randint(2000, 8000)
        # 特病患者在家附近活动范围更小
        activity_spread = 0.005 if is_severe else 0.015

        # 阶段 A：遥测数据（每个老人 30 条健康 + 30 条位置）
        print(f"\n  [{i+1}/10] {name} ({dev_id}){' ⚠特病' if is_severe else ''}")

        for t in range(30):
            ts = (datetime.now() - timedelta(minutes=(30-t)*3)).isoformat()

            # 健康数据
            hr = max(50, min(130, base_hr + random.gauss(0, 6 if not is_severe else 10)))
            spo2 = max(85, min(100, base_spo2 + random.gauss(0, 1.5 if not is_severe else 3)))
            steps = max(0, int(base_steps + random.gauss(0, 300) + t * 40))
            bp_sys = random.randint(110, 175 if is_severe else 155)
            bp_dia = random.randint(70, 105 if is_severe else 95)
            temp = round(random.uniform(36.3, 37.8), 1)

            r = api_post(f"{ADMIN_API}/admin/elderly/{elderly_id}/health-records", json={
                "hr": int(hr), "spo2": int(spo2), "steps": steps,
                "temperature": temp,
                "bp_systolic": int(bp_sys), "bp_diastolic": int(bp_dia),
                "timestamp": ts, "source": "device"
            })
            if r.status_code in (200, 201):
                stats["bracelet_health"] += 1

            # 位置数据（在家附近随机移动）
            lat, lon = rand_loc(base_lat, base_lon, activity_spread)
            r = api_post(f"{ADMIN_API}/admin/elderly/{elderly_id}/locations", json={
                "lat": lat, "lon": lon, "accuracy": random.randint(3, 15),
                "timestamp": ts
            })
            if r.status_code in (200, 201):
                stats["bracelet_location"] += 1

            # 心跳
            r = api_post(f"{ADMIN_API}/admin/devices/{device_id}/heartbeat", json={
                "device_id": device_id, "battery": random.randint(60, 95)
            })
            if r.status_code in (200, 201):
                stats["bracelet_heartbeat"] += 1

            # 每隔若干条发送一次
            if t % 5 == 0:
                print(f"      遥测 #{t+1}/30  HR={int(hr)} SpO2={int(spo2)} 步={steps}")

        # 阶段 B：特病患者额外生成慢病数据
        if is_severe:
            tiers = DISEASE_TIERS[i]
            if "diabetes" in tiers:
                for _ in range(5):
                    ts = rand_datetime_hours_back(24)
                    # 特病患者的血糖波动更大
                    glucose = round(random.uniform(6.0, 18.0), 1)
                    api_post(f"{ADMIN_API}/admin/chronic/{elderly_id}/glucose", json={
                        "value": glucose, "unit": "mmol/L",
                        "test_mode": random.choice(["fasting","postprandial"]),
                        "measurement_time": ts
                    })
                    stats["bracelet_chronic"] += 1

            if "cardiovascular" in tiers or "hypertension" in tiers:
                for _ in range(5):
                    ts = rand_datetime_hours_back(24)
                    bp_s = random.randint(140, 195)  # 高血压患者的收缩压偏高
                    bp_d = random.randint(85, 110)
                    api_post(f"{ADMIN_API}/admin/chronic/{elderly_id}/bp", json={
                        "systolic": bp_s, "diastolic": bp_d,
                        "measurement_time": ts,
                        "position": random.choice(["sitting","standing"])
                    })
                    stats["bracelet_chronic"] += 1

            if "diabetes" in tiers:
                for _ in range(3):
                    ts = rand_datetime_hours_back(24)
                    uric = round(random.uniform(380, 580), 1)  # 高尿酸
                    api_post(f"{ADMIN_API}/admin/chronic/{elderly_id}/uric-acid", json={
                        "value": uric, "unit": "μmol/L",
                        "measurement_time": ts
                    })
                    stats["bracelet_chronic"] += 1

        # 阶段 C：触发告警（特病患者概率更高）
        alert_prob = 0.08 if is_severe else 0.02
        if random.random() < alert_prob:
            alert_type = random.choice(["sos", "fall", "geofence_exit", "abnormal_hr", "abnormal_spo2"])
            severity = "high" if alert_type in ["sos", "fall", "abnormal_spo2"] else "medium"
            msg_map = {
                "sos": f"老人 {name} 触发 SOS 紧急呼叫",
                "fall": f"检测到 {name} 跌倒，置信度 {random.uniform(0.85,0.99):.2f}",
                "geofence_exit": f"{name} 超出电子围栏范围",
                "abnormal_hr": f"{name} 心率异常 {int(base_hr+random.gauss(0,15))} bpm",
                "abnormal_spo2": f"{name} 血氧偏低 {int(base_spo2+random.gauss(0,3))}%",
            }
            r = api_post(f"{ADMIN_API}/admin/alerts", json={
                "elderly_id": elderly_id,
                "alert_type": alert_type,
                "severity": severity,
                "message": msg_map[alert_type],
                "device_id": device_id,
                "lat": lat, "lon": lon
            })
            if r.status_code in (200, 201):
                stats["bracelet_alerts"] += 1
                print(f"      ⚠ 告警: {alert_type} — {msg_map[alert_type]}")

    print(f"\n  ✅ 手环场景完成: {stats['bracelet_health']}健康记录, "
          f"{stats['bracelet_location']}位置记录, "
          f"{stats['bracelet_heartbeat']}心跳, "
          f"{stats['bracelet_alerts']}告警, "
          f"{stats['bracelet_chronic']}慢病记录")


# ═══════════════════════════════════════════════════════════════════
# 场景二：住院业务（≤10人）
# ═══════════════════════════════════════════════════════════════════

DEPARTMENTS = ["心内科","神经内科","骨科","老年病科","普外科","呼吸内科","内分泌科"]
DIAGNOSES = [
    ("I10", "原发性高血压", "moderate"),
    ("I50.9", "充血性心力衰竭", "severe"),
    ("E11.9", "2型糖尿病伴并发症", "moderate"),
    ("J44.1", "慢阻肺急性发作", "severe"),
    ("M81.0", "老年性骨质疏松", "mild"),
    ("I63.9", "脑梗死", "severe"),
    ("N18.3", "慢性肾脏病3期", "moderate"),
    ("G30.9", "阿尔茨海默病", "moderate"),
    ("M17.1", "膝骨关节炎", "mild"),
    ("I25.10", "冠状动脉粥样硬化性心脏病", "severe"),
]
MEDICATIONS = [
    ("氨氯地平", "5mg", "QD", "oral"),
    ("二甲双胍", "500mg", "TID", "oral"),
    ("阿托伐他汀", "20mg", "QN", "oral"),
    ("呋塞米", "20mg", "QD", "oral"),
    ("地高辛", "0.125mg", "QD", "oral"),
    ("华法林", "2.5mg", "QD", "oral"),
    ("胰岛素注射剂", "12U", "TID", "subcutaneous"),
    ("奥美拉唑", "20mg", "QD", "oral"),
    ("泼尼松", "5mg", "QD", "oral"),
    ("硝苯地平缓释片", "30mg", "BID", "oral"),
]
NURSES = ["护士张", "护士李", "护士王", "护士陈", "护士刘"]

def simulate_hospital():
    """模拟住院业务链"""
    print("\n" + "═" * 65)
    print("  场景二：住院业务（≤10位住院患者）")
    print("═" * 65)

    # ── 步骤 1：注册医院机构 ──────────────────────────────────────
    print("\n  [H1] 注册模拟医院机构...")
    r = api_post(f"{HOSPITAL_B2B}/institutions", json={
        "name": "上海市第三人民医院（模拟）",
        "type": "hospital",
        "code": "SH3H-2026",
        "contact_name": "王院长",
        "contact_phone": "13800138000",
        "access_level": "read_write",
        "status": "active"
    })
    if r.status_code == 201:
        inst = r.json()["data"]
        hospital_inst_id = inst["id"]
        print(f"    ✅ 医院注册成功: {inst['name']} (ID: {hospital_inst_id[:8]}...)")
    else:
        # 可能已存在，查询列表
        r = api_get(f"{HOSPITAL_B2B}/institutions")
        if r.status_code == 200:
            insts = r.json().get("data", [])
            if insts:
                hospital_inst_id = insts[0]["id"]
                print(f"    ✅ 使用已有医院: {insts[0]['name']}")
            else:
                print("    ❌ 无法获取医院机构")
                return
        else:
            print(f"    ❌ 获取医院列表失败: {r.status_code}")
            return

    # ── 步骤 2：生成医院 API Key ──────────────────────────────────
    print("\n  [H2] 生成医院 API Key（模拟 HIS 系统认证）...")
    r = api_post(f"{HOSPITAL_B2B}/institutions/{hospital_inst_id}/api-keys", json={
        "name": "HIS主系统",
        "expires_at": (datetime.now() + timedelta(days=365)).isoformat()
    })
    if r.status_code == 201:
        hospital_api_key = r.json()["data"]["key"]
        print(f"    ✅ API Key 已生成: {hospital_api_key[:8]}...")
    else:
        print(f"    ⚠️  API Key 生成失败，尝试从已有机构获取: {r.text[:100]}")
        return

    b2b_headers = {"X-API-Key": hospital_api_key, "Content-Type": "application/json"}

    # ── 步骤 3：创建住院患者（≤10人，包含特病患者）────────────────
    print(f"\n  [H3] 创建住院患者档案（10人，含特病患者）...")
    sim_patients = []

    for i in range(10):
        pid = uid("pat-")
        adm_no = f"2026-{20001+i}"
        name = f"{random.choice(SURNAMES)}{random.choice(FIRST_NAMES)}{i+1}"
        dept = DEPARTMENTS[i % len(DEPARTMENTS)]
        gender = random.choice(["male", "female"])
        age = random.randint(65, 95)
        blood_type = random.choice(["A+","A-","B+","B-","AB+","AB-","O+","O-"])
        allergy = random.choice(["青霉素","磺胺类","None","None","Latex","None"])

        # 特病患者有明确的诊断和用药
        is_special = i in [1, 4, 7, 9]  # 4位特病患者
        diag = DIAGNOSES[i % len(DIAGNOSES)] if is_special else DIAGNOSES[random.randint(0, len(DIAGNOSES)-1)]

        r = api_post(f"{ADMIN_API}/admin/medical/patients", json={
            "id": pid, "admission_no": adm_no, "name": name,
            "gender": gender, "age": age, "department": dept,
            "bed_number": f"{random.randint(1,40)}床-{random.randint(1,12)}房",
            "blood_type": blood_type, "allergies": allergy,
            "special_conditions": diag[1],
            "status": "admitted"
        })
        if r.status_code == 201:
            medical_patients.append((pid, adm_no, name, dept))
            sim_patients.append({
                "patient_id": pid, "name": name, "dept": dept,
                "diag_code": diag[0], "diag_name": diag[1], "is_special": is_special,
                "adm_no": adm_no
            })
            print(f"    [{i+1}] {name} | {dept} | {diag[1]} {'⚠特病' if is_special else ''}")

    # ── 步骤 4：建立老人-医院关联 ────────────────────────────────
    print(f"\n  [H4] 建立老人-医院关联（入院绑定）...")
    for p in sim_patients:
        # 从 bracelet_elders 中取对应老人
        if bracelet_elders:
            idx = sim_patients.index(p) % len(bracelet_elders)
            eld_id, eld_name, _ = bracelet_elders[idx]

            r = api_post(f"{HOSPITAL_B2B}/links", headers=b2b_headers, json={
                "elderly_id": eld_id,
                "institution_id": hospital_inst_id,
                "admitted_at": rand_datetime_hours_back(30*24),
                "primary_doc": f"主任{random.choice(['张','李','王','陈'])}",
                "notes": json.dumps({"diagnosis_code": p["diag_code"], "diagnosis_name": p["diag_name"]})
            })
            if r.status_code in (200, 201):
                stats["hospital_admissions"] += 1
            else:
                # 如果没有 elderly 关联，直接通过 patient_id 走 HIS 通道
                pass

    # ── 步骤 5：HIS 推送生命体征（核心业务数据）────────────────────
    print(f"\n  [H5] HIS 推送生命体征数据（每位患者 20-40 条）...")
    for p in sim_patients:
        n_vitals = random.randint(20, 40)
        for v in range(n_vitals):
            ts = (datetime.now() - timedelta(hours=random.randint(0, 72))).isoformat()
            is_special = p["is_special"]

            # 特病患者的生命体征更异常
            base_hr = random.randint(75, 105) if not is_special else random.randint(85, 125)
            base_spo2 = random.randint(94, 99) if not is_special else random.randint(85, 93)

            vitals = [
                {"type": "hr", "value": max(50, min(140, base_hr + random.gauss(0, 8))), "normal": True},
                {"type": "spo2", "value": max(82, min(100, base_spo2 + random.gauss(0, 2))), "normal": None},
                {"type": "bp_systolic", "value": random.randint(100, 200 if is_special else 165), "normal": None},
                {"type": "bp_diastolic", "value": random.randint(60, 120 if is_special else 100), "normal": None},
                {"type": "temp", "value": round(random.uniform(36.5, 39.5 if is_special else 37.8), 1)},
            ]

            r = api_post(f"{HOSPITAL_B2B}/health-data", headers=b2b_headers, json={
                "patient_id": p["adm_no"],
                "timestamp": ts,
                "vitals": vitals
            })
            if r.status_code in (200, 201):
                stats["hospital_vitals"] += len(vitals)

        # 推送诊断（特病患者）
        if p["is_special"]:
            r = api_post(f"{HOSPITAL_B2B}/health-data", headers=b2b_headers, json={
                "patient_id": p["adm_no"],
                "timestamp": rand_datetime_hours_back(72),
                "diagnoses": [{"code": p["diag_code"], "name": p["diag_name"], "severity": "severe"}]
            })
            if r.status_code in (200, 201):
                stats["hospital_diagnoses"] += 1

        # 推送用药医嘱
        n_meds = random.randint(2, 5)
        selected_meds = random.sample(MEDICATIONS, n_meds)
        r = api_post(f"{HOSPITAL_B2B}/health-data", headers=b2b_headers, json={
            "patient_id": p["adm_no"],
            "timestamp": rand_datetime_hours_back(72),
            "medications": [{"name": m[0], "dose": m[1], "freq": m[2], "route": m[3], "duration": f"{random.randint(3,14)}天"}
                           for m in selected_meds]
        })
        if r.status_code in (200, 201):
            stats["hospital_meds"] += n_meds

    # ── 步骤 6：护士核验终端扫码 ──────────────────────────────────
    print(f"\n  [H6] 护士核验终端扫码（ medication/vitals/ward_round 类型）...")
    for p in sim_patients:
        n_verifs = random.randint(3, 8)
        for v in range(n_verifs):
            vtype = random.choice(["medication", "vitals", "ward_round", "discharge"])
            result = random.choice(["matched", "matched", "matched", "unmatched"])  # 80%匹配
            r = api_post(f"{ADMIN_API}/admin/medical/verifications", json={
                "patient_id": p["patient_id"],
                "device_id": "",
                "scan_type": vtype,
                "result": result,
                "verified_by": random.choice(NURSES),
                "notes": "" if result == "matched" else f"{vtype}信息不匹配",
                "timestamp": rand_datetime_hours_back(48)
            })
            if r.status_code in (200, 201):
                stats["hospital_verifications"] += 1

    # ── 步骤 7：护士巡房记录 ──────────────────────────────────────
    print(f"\n  [H7] 护士巡房记录...")
    for p in sim_patients:
        n_rounds = random.randint(2, 5)
        for _ in range(n_rounds):
            r = api_post(f"{ADMIN_API}/admin/medical/patients/{p['patient_id']}/ward-round", json={
                "nurse_id": random.choice(NURSES),
                "blood_pressure": f"{random.randint(100,180)}/{random.randint(60,110)}",
                "heart_rate": random.randint(60, 115),
                "spo2": random.randint(88, 99),
                "temperature": round(random.uniform(36.5, 38.5), 1),
                "weight": round(random.uniform(45, 90), 1),
                "notes": random.choice(["患者精神状态良好", "生命体征稳定", "需继续观察", "疼痛缓解"]),
                "observations": json.dumps({"falls_risk": random.choice([True, False]),
                                            "pain_level": random.randint(0, 8)})
            })
            if r.status_code in (200, 201):
                stats["hospital_ward_rounds"] += 1

    # ── 步骤 8：生成健康报告 ──────────────────────────────────────
    print(f"\n  [H8] 生成住院患者健康报告...")
    for p in sim_patients:
        r = api_get(f"{HOSPITAL_B2B}/patients/{p['patient_id']}/report?days=7",
                    extra_headers=b2b_headers)
        if r.status_code in (200, 201):
            stats["hospital_reports"] += 1
        else:
            print(f"    ⚠ {p['name']} 报告生成失败: {r.status_code}")

    print(f"\n  ✅ 住院场景完成: {stats['hospital_vitals']}生命体征, "
          f"{stats['hospital_diagnoses']}诊断, {stats['hospital_meds']}用药医嘱, "
          f"{stats['hospital_verifications']}核验, {stats['hospital_ward_rounds']}巡房, "
          f"{stats['hospital_reports']}报告")


# ═══════════════════════════════════════════════════════════════════
# 场景三：社区业务（≤10人）
# ═══════════════════════════════════════════════════════════════════

WELFARE_TAGS = [
    ("ELDERLY", "老年人福利", 365, 50.0),
    ("DISABLED", "残疾人补贴", 365, 200.0),
    ("SOLITARY", "空巢老人补助", 180, 100.0),
    ("CHRONIC", "慢性病补助", 365, 150.0),
    ("HIGH_AGE", "高龄津贴", 365, 300.0),
]

SERVICES = ["home_visit", "health_check", "med_delivery", "emergency_response", "rehabilitation"]

def simulate_community():
    """模拟社区业务链"""
    print("\n" + "═" * 65)
    print("  场景三：社区业务（≤10位社区老人）")
    print("═" * 65)

    # ── 步骤 1：创建社区老人档案（从 bracelet_elders 选取前10人）──
    print(f"\n  [C1] 创建社区老人档案...")
    for i in range(min(10, len(bracelet_elders))):
        eld_id, name, _ = bracelet_elders[i]
        age = random.randint(68, 92)
        id_card = f"310101{(1930+age-68):02d}{random.randint(1,12):02d}{random.randint(1,28):02d}{'X' if random.random()>0.8 else str(random.randint(0,9))}"

        r = api_post(f"{ADMIN_API}/admin/community-wb/elders", json={
            "id": eld_id, "name": name, "id_card": id_card,
            "gender": random.choice([0, 1]),
            "age": age,
            "address": f"上海市{random.choice(['浦东','徐汇','长宁','静安','黄浦','虹口','杨浦'])}区{random.randint(1,200)}弄{random.randint(1,50)}号",
            "emergency_contact": f"138{random.randint(10000000, 99999999)}",
            "status": "active"
        })
        if r.status_code in (200, 201):
            community_elders.append((eld_id, name, id_card))
            print(f"    [{i+1}] {name} | {age}岁 | {id_card}")

    # ── 步骤 2：发放福利标签 ──────────────────────────────────────
    print(f"\n  [C2] 发放福利标签（每位老人 1-3 个标签）...")
    for eid, ename, _ in community_elders:
        n_tags = random.randint(1, 3)
        selected_tags = random.sample(WELFARE_TAGS, n_tags)
        for tag_code, tag_name, days, amount in selected_tags:
            valid_from = (datetime.now() - timedelta(days=random.randint(0, 30))).strftime("%Y-%m-%d")
            valid_to = (datetime.now() + timedelta(days=days)).strftime("%Y-%m-%d")
            r = api_post(f"{ADMIN_API}/admin/community-wb/elders/{eid}/welfare/{tag_code}", json={
                "valid_from": valid_from, "valid_to": valid_to,
                "certified_by": random.choice(["社区医生", "街道办", "民政局"])
            })
            if r.status_code in (200, 201):
                stats["community_tags"] += 1
        print(f"    {ename}: 获得 {n_tags} 个福利标签")

    # ── 步骤 3：社区签到（近30天）──────────────────────────────────
    print(f"\n  [C3] 社区签到记录（近30天，每人 5-15 次）...")
    # 获取一个有效的医院ID用于签到
    r = api_get(f"{ADMIN_API}/admin/institutions?page=1&page_size=1")
    valid_hospital_id = ""
    if r.status_code == 200:
        insts = r.json().get("data", [])
        if insts:
            valid_hospital_id = insts[0]["id"]
            print(f"    使用医院ID: {valid_hospital_id[:8]}...")

    for eid, ename, _ in community_elders:
        n_signins = random.randint(5, 15)
        for s in range(n_signins):
            sig_date = datetime.now() - timedelta(days=random.randint(0, 30), hours=random.randint(6, 20))
            period = sig_date.strftime("%Y-%m")
            tags = random.sample([t[0] for t in WELFARE_TAGS], random.randint(1, 2))
            r = api_post(f"{ADMIN_API}/admin/community-wb/signin/trigger", json={
                "elder_id": eid,
                "device_id": community_devices[random.randint(0, len(community_devices)-1)][0] if community_devices else "",
                "hospital_id": valid_hospital_id,
                "period": period,
                "signin_time": sig_date.isoformat(),
                "is_medical_signin": random.random() < 0.7,
                "is_welfare_signin": random.random() < 0.5,
                "activated_tags": tags
            })
            if r.status_code in (200, 201):
                stats["community_signins"] += 1

    # ── 步骤 4：药房配药记录 ──────────────────────────────────────
    print(f"\n  [C4] 药房配药记录（每人 2-5 次）...")
    COMMUNITY_MEDS = ["氨氯地平", "二甲双胍", "阿托伐他汀", "缬沙坦", "奥美拉唑",
                      "阿司匹林", "硝苯地平", "格列美脲", "头孢类抗生素", "降压药组合"]
    for eid, ename, _ in community_elders:
        n_dispenses = random.randint(2, 5)
        for d in range(n_dispenses):
            disp_date = datetime.now() - timedelta(days=random.randint(0, 30))
            n_items = random.randint(1, 3)
            items = random.sample(COMMUNITY_MEDS, n_items)
            total = round(random.uniform(20, 300), 2)
            r = api_post(f"{ADMIN_API}/admin/community-wb/pharmacy/dispense", json={
                "elder_id": eid,
                "hospital_id": valid_hospital_id,
                "period": disp_date.strftime("%Y-%m"),
                "dispense_time": disp_date.isoformat(),
                "items": items,
                "total_cost": total,
                "insurance_covered": round(total * random.uniform(0.3, 0.8), 2),
                "self_pay": round(total * random.uniform(0.2, 0.7), 2)
            })
            if r.status_code in (200, 201):
                stats["community_pharmacy"] += 1

    # ── 步骤 5：社区健康检查（B2B 接口）───────────────────────────
    print(f"\n  [C5] 社区健康检查记录...")
    for eid, ename, _ in community_elders:
        n_checks = random.randint(2, 4)
        for c in range(n_checks):
            check_date = datetime.now() - timedelta(days=random.randint(0, 60))
            r = api_post(f"{COMMUNITY_B2B}/health-checks?elderly_id={eid}", json={
                "check_date": check_date.isoformat(),
                "bp_systolic": round(random.uniform(110, 175), 1),
                "bp_diastolic": round(random.uniform(65, 105), 1),
                "hr": round(random.uniform(60, 95), 1),
                "spo2": round(random.uniform(94, 99), 1),
                "weight": round(random.uniform(50, 85), 1),
                "glucose": round(random.uniform(4.5, 12.0), 1),
                "checked_by": random.choice(["社区医生赵", "护士周", "体检中心"]),
                "notes": random.choice(["指标正常", "建议复查", "需随访", ""])
            })
            if r.status_code in (200, 201):
                stats["community_health_checks"] += 1

    # ── 步骤 6：照护计划 ──────────────────────────────────────────
    print(f"\n  [C6] 创建照护计划...")
    for eid, ename, _ in community_elders:
        if random.random() < 0.6:  # 60%的老人有照护计划
            n_tasks = random.randint(2, 4)
            tasks = []
            for _ in range(n_tasks):
                tasks.append({
                    "title": random.choice(["上门探访", "用药提醒", "健康评估", "心理关怀", "生活协助"]),
                    "type": random.choice(SERVICES),
                    "schedule": random.choice(["daily", "weekly", "mon_wed_fri"]),
                    "completed": random.choice([True, False])
                })
            start_date = (datetime.now() - timedelta(days=random.randint(0, 30))).strftime("%Y-%m-%d")
            r = api_post(f"{COMMUNITY_B2B}/care-plans", json={
                "elderly_id": eid,
                "title": f"{ename}的照护计划",
                "description": f"针对{ename}的个性化社区照护方案",
                "tasks": tasks,
                "assigned_to": random.choice(community_elders)[0] if community_elders else eid,
                "status": "active",
                "start_date": start_date
            })
            if r.status_code in (200, 201):
                stats["community_care_plans"] += 1

    print(f"\n  ✅ 社区场景完成: {stats['community_signins']}签到, "
          f"{stats['community_pharmacy']}药房, "
          f"{stats['community_health_checks']}健康检查, "
          f"{stats['community_care_plans']}照护计划, "
          f"{stats['community_tags']}福利标签")


# ═══════════════════════════════════════════════════════════════════
# 初始化：创建实体
# ═══════════════════════════════════════════════════════════════════
def setup_entities():
    global token, headers

    print("\n" + "═" * 65)
    print("  初始化：创建基础数据实体")
    print("═" * 65)

    # 登录
    print("\n  [SETUP 1] 管理员登录...")
    r = requests.post(f"{ADMIN_API}/auth/login", json={
        "method": "email", "credential": "admin@eregen.com", "secret": "Admin@123"
    }, timeout=10)
    if r.status_code != 200:
        print(f"  ❌ 登录失败: {r.text[:200]}")
        sys.exit(1)
    token = r.json()["data"]["token"]
    headers = {"Authorization": f"Bearer {token}"}
    print(f"  ✅ 登录成功")

    # 创建10位老人 + 手环设备
    print(f"\n  [SETUP 2] 创建 10 位老人档案 + 手环设备...")
    for i in range(10):
        uid_val = uid("usr-")
        name = f"{SURNAMES[i]}{FIRST_NAMES[i]}"
        email = f"user{i+1}_{str(uuid.uuid4())[:4]}@eregen.com"

        # 创建用户
        r = api_post(f"{ADMIN_API}/admin/users", json={
            "name": name, "email": email, "role": "family", "password": "Test@12345",
            "phone": f"138{uid_val[4:]}"
        })
        if r.status_code != 201:
            print(f"  ⚠ 用户 {name} 创建失败: {r.text[:80]}")
            continue
        user_id = r.json()["data"]["id"]

        # 创建老人档案
        eid = uid("eld-")
        health_tiers = DISEASE_TIERS[i]
        r = api_post(f"{ADMIN_API}/admin/elderly", json={
            "name": name,
            "birth_date": rand_date(6000),
            "user_id": user_id,
            "health_tiers": health_tiers,
            "gender": random.choice(["male", "female"]),
            "phone": f"138{random.randint(10000000, 99999999)}",
            "emergency_contact": f"{random.choice(SURNAMES)}{random.choice(FIRST_NAMES)}",
            "emergency_phone": f"139{random.randint(10000000, 99999999)}",
        })
        if r.status_code != 201:
            print(f"  ⚠ 老人 {name} 档案创建失败: {r.text[:80]}")
            continue
        elderly_id = r.json()["data"]["id"]

        # 创建手环设备
        dev_id = f"BR-{i+1:04d}"
        r = api_post(f"{ADMIN_API}/admin/devices", json={
            "device_id": dev_id, "device_type": "bracelet", "tier": "plus",
            "owner_user_id": user_id, "status": "online"
        })
        if r.status_code == 201:
            device_id = r.json()["data"]["id"]
            # 绑定设备到老人
            api_post(f"{ADMIN_API}/admin/elderly/{elderly_id}/link-device", json={
                "device_id": device_id
            })
            bracelet_devices.append((device_id, elderly_id, dev_id))
        else:
            bracelet_devices.append(("", elderly_id, dev_id))

        bracelet_elders.append((elderly_id, name, user_id))
        if (i+1) % 5 == 0:
            print(f"    已创建 {i+1}/10 位老人")

    print(f"  ✅ 已创建 {len(bracelet_elders)} 位老人，{len(bracelet_devices)} 个手环设备")

    # 创建电子围栏配置
    print(f"\n  [SETUP 3] 创建电子围栏...")
    for eid, name, _ in bracelet_elders[:5]:
        lat, lon = rand_loc(LAT_C, LON_C, 0.02)
        api_post(f"{ADMIN_API}/admin/elderly/{eid}/geofences", json={
            "name": " home",
            "latitude": lat, "longitude": lon,
            "radius_meters": random.randint(150, 300),
            "active": True
        })
    print(f"  ✅ 创建 5 个电子围栏")

    # 创建用药规则
    print(f"\n  [SETUP 4] 创建用药规则...")
    for eid, name, _ in bracelet_elders[:5]:
        for hour in [8, 12, 20]:
            api_post(f"{ADMIN_API}/admin/elderly/{eid}/medication-rules", json={
                "schedule_time": f"{hour:02d}:00",
                "pill_type": random.choice(["tablet", "capsule"]),
                "dose_count": random.randint(1, 3),
                "days_of_week": list(range(7)),
                "active": True
            })
    print(f"  ✅ 创建用药规则")

    # 创建福利标签
    print(f"\n  [SETUP 5] 创建福利标签配置...")
    for tag_code, tag_name, days, amount in WELFARE_TAGS:
        api_post(f"{ADMIN_API}/admin/community-wb/welfare-tags", json={
            "tag_code": tag_code, "tag_name": tag_name,
            "issuer": random.choice(["民政局", "残联", "卫健委"]),
            "renewal_period_days": days, "benefit_amount": amount, "enabled": True
        })
    print(f"  ✅ 创建 {len(WELFARE_TAGS)} 个福利标签")


# ═══════════════════════════════════════════════════════════════════
# 验证
# ═══════════════════════════════════════════════════════════════════
def verify():
    print("\n" + "═" * 65)
    print("  验证：数据完整性检查")
    print("═" * 65)

    results = []

    # 老人数据
    r = api_get(f"{ADMIN_API}/admin/elderly?page=1&page_size=100")
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", len(data.get("elderly", [])))
        results.append(check("老人档案数量", total >= 10, f"={total}"))
    else:
        results.append(False)
        print(f"  ❌ 老人列表: HTTP {r.status_code}")

    # 设备数据
    r = api_get(f"{ADMIN_API}/admin/devices?page=1&page_size=100")
    if r.status_code == 200:
        data = r.json().get("data", [])
        if isinstance(data, dict):
            total = data.get("total", len(data.get("devices", [])))
        else:
            total = len(data)
        results.append(check("设备数量", total >= 10, f"={total}"))
    else:
        results.append(False)
        print(f"  ❌ 设备列表: HTTP {r.status_code}")

    # 健康记录
    if bracelet_elders:
        eid = bracelet_elders[0][0]
        r = api_get(f"{ADMIN_API}/admin/elderly/{eid}/health-records?limit=5")
        if r.status_code == 200:
            records = r.json().get("data", [])
            results.append(check("健康记录", len(records) > 0, f"count={len(records)}"))
        else:
            results.append(False)

    # 住院患者
    r = api_get(f"{ADMIN_API}/admin/medical/patients?page=1&page_size=30")
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("patients", []))))
        results.append(check("住院患者", total >= 10, f"={total}"))
    else:
        results.append(False)
        print(f"  ❌ 住院患者列表: HTTP {r.status_code}")

    # 社区老人
    r = api_get(f"{ADMIN_API}/admin/community-wb/elders?page=1&page_size=20")
    if r.status_code == 200:
        data = r.json().get("data", {})
        total = data.get("total", data.get("count", len(data.get("elders", []))))
        results.append(check("社区老人", total >= 10, f"={total}"))
    else:
        results.append(False)
        print(f"  ❌ 社区老人列表: HTTP {r.status_code}")

    # B2B 医院报告
    if medical_patients:
        r = api_get(f"{HOSPITAL_B2B}/patients/{medical_patients[0][0]}/report?days=7",
                    extra_headers={"X-API-Key": hospital_api_key})
        results.append(check("B2B健康报告", r.status_code in (200, 201), f"HTTP {r.status_code}"))
    else:
        results.append(False)
        print("  ⚠ 无住院患者数据，跳过 B2B 报告验证")

    # 仪表盘
    r = api_get(f"{ADMIN_API}/admin/stats/overview")
    results.append(check("仪表盘数据", r.status_code == 200 and bool(r.json().get("data")), f"HTTP {r.status_code}"))

    # ── 统计 ─────────────────────────────────────────────────────
    print(f"\n{'─'*65}")
    print("  模拟数据统计")
    print("─" * 65)
    print(f"  ┌─ 场景一：普通手环 ─────────────────────────────────┐")
    print(f"  │ 健康遥测记录:    {stats['bracelet_health']:>5} 条")
    print(f"  │ 位置记录:        {stats['bracelet_location']:>5} 条")
    print(f"  │ 心跳包:          {stats['bracelet_heartbeat']:>5} 条")
    print(f"  │ 告警:            {stats['bracelet_alerts']:>5} 条")
    print(f"  │ 慢病记录:        {stats['bracelet_chronic']:>5} 条")
    print(f"  ├─ 场景二：住院业务 ─────────────────────────────────┤")
    print(f"  │ HIS 生命体征:    {stats['hospital_vitals']:>5} 条")
    print(f"  │ 诊断记录:        {stats['hospital_diagnoses']:>5} 条")
    print(f"  │ 用药医嘱:        {stats['hospital_meds']:>5} 条")
    print(f"  │ 护士核验:        {stats['hospital_verifications']:>5} 条")
    print(f"  │ 巡房记录:        {stats['hospital_ward_rounds']:>5} 条")
    print(f"  │ 健康报告:        {stats['hospital_reports']:>5} 份")
    print(f"  ├─ 场景三：社区业务 ─────────────────────────────────┤")
    print(f"  │ 福利标签:        {stats['community_tags']:>5} 个")
    print(f"  │ 签到记录:        {stats['community_signins']:>5} 条")
    print(f"  │ 药房配药:        {stats['community_pharmacy']:>5} 条")
    print(f"  │ 健康检查:        {stats['community_health_checks']:>5} 条")
    print(f"  │ 照护计划:        {stats['community_care_plans']:>5} 份")
    print(f"  └────────────────────────────────────────────────────┘")

    passed = sum(results)
    total = len(results)
    print(f"\n  验证结果: {passed}/{total} 项通过")
    if passed == total:
        print("  ✅ 全部验证通过！数据完整性确认。")
    else:
        print(f"  ⚠️  {total - passed} 项验证未通过，请检查上方日志")
    return all(results)


# ═══════════════════════════════════════════════════════════════════
# 主入口
# ═══════════════════════════════════════════════════════════════════
if __name__ == "__main__":
    print("=" * 65)
    print("  Eregen 三场景真实数据模拟")
    print(f"  时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 65)

    # 检查服务连通性
    print("\n  检查服务连通性...")
    services = [
        ("admin-api", ADMIN_API),
        ("hospital-b2b", HOSPITAL_B2B),
        ("community-b2b", COMMUNITY_B2B),
    ]
    for name, url in services:
        try:
            r = requests.get(f"{url}/institutions" if "b2b" in url else f"{url}/health", timeout=3)
            print(f"  ✅ {name}: OK (HTTP {r.status_code})")
        except requests.exceptions.ConnectionError:
            print(f"  ❌ {name}: 无法连接 — 请确认服务已启动（端口 {url.split(':')[2] if ':' in url else 'unknown'}）")
            sys.exit(1)

    try:
        setup_entities()
        simulate_bracelet()
        simulate_hospital()
        simulate_community()
        success = verify()
        sys.exit(0 if success else 1)
    except requests.exceptions.ConnectionError as e:
        print(f"\n❌ 连接失败: {e}")
        print("   请确认以下服务正在运行：")
        print("   - admin-api (PORT=8089)")
        print("   - hospital-api (PORT=8082)")
        print("   - community-platform (PORT=8083)")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
