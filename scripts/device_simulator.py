#!/usr/bin/env python3
"""
Eregen Platform — MQTT Device Data Simulator
Simulates 10 bracelet elders, 20 medical wristband patients, 50 community elders
with realistic telemetry, health, location, alerts, and workflow data.
Publishes directly to EMQX MQTT broker.
"""
import json
import random
import sys
import time
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("ERROR: paho-mqtt not installed. Run: pip3 install paho-mqtt")
    sys.exit(1)

# ─── Configuration ────────────────────────────────────────────────────────────
MQTT_BROKER = "localhost"
MQTT_PORT = 1883
MQTT_CLIENT_ID = "eregen-simulator"

BRACELET_COUNT = 10
MEDICAL_COUNT = 20
COMMUNITY_COUNT = 50

PUBLISH_INTERVAL = 3  # seconds between messages per device per cycle

CENTER_LAT = 31.2304
CENTER_LON = 121.4737

ELDER_NAMES = [
    "张大爷", "李奶奶", "王爷爷", "赵阿姨", "刘大爷",
    "陈奶奶", "杨爷爷", "黄阿姨", "周大爷", "吴奶奶",
    "孙爷爷", "朱阿姨", "马大爷", "蒋奶奶", "韩爷爷",
    "林阿姨", "何大爷", "高奶奶", "梁爷爷", "宋阿姨",
]

DEPARTMENTS = ["心内科", "骨科", "神经内科", "内分泌科", "呼吸科", "普外科"]
BLOOD_TYPES = ["A", "B", "AB", "O"]
WELFARE_TAGS = ["高龄补贴", "独居老人", "特困供养", "低保户", "残疾补助"]
MEDICATIONS = ["降压药", "降糖药", "阿司匹林", "降脂药", "护肝片"]


def uid(prefix, n):
    return f"{prefix}{n:04d}"


def rand_loc(center_lat=CENTER_LAT, center_lon=CENTER_LON, spread=0.015):
    return round(center_lat + random.gauss(0, spread), 6), \
           round(center_lon + random.gauss(0, spread), 6)


def now_ts():
    return int(datetime.now().timestamp())


# ─── Device Classes ──────────────────────────────────────────────────────────

class BraceletDevice:
    def __init__(self, idx):
        self.dev_id = uid("BR-", idx)
        self.elder_name = ELDER_NAMES[idx - 1] if idx <= len(ELDER_NAMES) else f"老人{idx}"
        self.lat, self.lon = rand_loc()
        self.bat = random.randint(60, 100)
        self.hr_base = random.randint(65, 85)
        self.spo2_base = random.randint(95, 99)
        self.steps = random.randint(0, 8000)
        self.model = random.choice(["Starter", "Plus", "Pro"])
        self.fw = "v2.1.0"

    def heartbeat(self):
        return json.dumps({
            "type": "heartbeat", "dev_id": self.dev_id, "ts": now_ts(),
            "bat": self.bat, "model": self.model, "fw_ver": self.fw,
        }, ensure_ascii=False)

    def location(self):
        lat, lon = rand_loc(self.lat, self.lon, 0.002)
        self.lat, self.lon = lat, lon
        return json.dumps({
            "type": "location", "dev_id": self.dev_id, "ts": now_ts(),
            "lat": lat, "lon": lon, "acc": random.choice([3, 5, 8, 12]),
            "source": random.choice(["gps", "base_station", "gps"]),
            "speed": round(random.uniform(0, 5.0), 1),
        }, ensure_ascii=False)

    def health(self):
        hr = max(40, min(180, self.hr_base + random.randint(-10, 15)))
        spo2 = max(85, min(100, self.spo2_base + random.randint(-2, 2)))
        self.steps += random.randint(10, 200)
        return json.dumps({
            "type": "health", "dev_id": self.dev_id, "ts": now_ts(),
            "hr": hr, "spo2": spo2, "step": self.steps,
            "sleep": random.choice([0, 360, 420, 480]),
        }, ensure_ascii=False)

    def sos(self):
        lat, lon = rand_loc(self.lat, self.lon, 0.001)
        return json.dumps({
            "type": "sos", "dev_id": self.dev_id, "ts": now_ts(),
            "lat": lat, "lon": lon, "trigger": random.choice(["manual", "long_press"]),
        }, ensure_ascii=False)

    def fall(self):
        lat, lon = rand_loc(self.lat, self.lon, 0.001)
        return json.dumps({
            "type": "fall", "dev_id": self.dev_id, "ts": now_ts(),
            "conf": round(random.uniform(0.75, 0.99), 2),
            "lat": lat, "lon": lon,
        }, ensure_ascii=False)

    @property
    def topic(self):
        return f"eregen/device/bracelet/{self.dev_id}/up"


class MedicalDevice:
    def __init__(self, idx):
        self.dev_id = uid("MW-", idx)
        self.patient_id = f"P-{random.randint(10000, 99999)}"
        self.admission_no = f"AD{random.randint(2024, 2026)}{random.randint(1000, 9999)}"
        self.name = random.choice(ELDER_NAMES)
        self.gender = random.choice(["M", "F"])
        self.age = random.randint(60, 90)
        self.dept = random.choice(DEPARTMENTS)
        self.bed = f"{random.randint(1, 50)}床"
        self.blood = random.choice(BLOOD_TYPES)
        self.bat = random.randint(70, 100)
        self.fw = "v1.3.2"
        self.bind_count = random.randint(1, 15)

    def patient_register(self):
        return json.dumps({
            "type": "patient_register", "dev_id": self.dev_id, "ts": now_ts(),
            "patient_id": self.patient_id, "admission_no": self.admission_no,
            "name": self.name, "gender": self.gender, "age": self.age,
            "department": self.dept, "bed_number": self.bed,
            "blood_type": self.blood,
            "allergies": random.choice(["青霉素", "磺胺类", "无"]),
            "special_conditions": random.choice(["糖尿病II型", "高血压III级", "冠心病", "无"]),
        }, ensure_ascii=False)

    def verification_scan(self):
        result = random.choices(["matched", "unmatched", "not_found"], weights=[85, 10, 5])[0]
        scan_type = random.choice(["round", "medication", "test", "discharge"])
        lat, lon = rand_loc()
        return json.dumps({
            "type": "verification_scan", "dev_id": self.dev_id, "ts": now_ts(),
            "patient_id": self.patient_id,
            "device_id": f"Nurse-{random.randint(100, 200)}",
            "scan_type": scan_type, "result": result,
            "verified_by": f"护士{random.randint(1, 20)}",
            "lat": lat, "lon": lon,
            "notes": "" if result == "matched" else "证件不符",
        }, ensure_ascii=False)

    def device_status(self):
        return json.dumps({
            "type": "device_status", "dev_id": self.dev_id, "ts": now_ts(),
            "bat": self.bat, "fw_ver": self.fw, "status": "online",
            "last_bind_ts": now_ts() - random.randint(60, 86400),
            "bind_count": self.bind_count,
        }, ensure_ascii=False)

    def alert_tag(self):
        severity = random.choice(["info", "warning", "critical"])
        tag_map = {
            "info": ("体温异常",),
            "warning": ("心率异常", "血氧偏低", "用药提醒"),
            "critical": ("SOS紧急呼叫", "跌倒检测", "心电异常"),
        }
        tag_name = random.choice(tag_map[severity])
        lat, lon = rand_loc()
        return json.dumps({
            "type": "alert_tag", "dev_id": self.dev_id, "ts": now_ts(),
            "tag_id": f"tag-{random.randint(1000, 9999)}",
            "tag_name": tag_name, "severity": severity,
            "patient_id": self.patient_id,
            "lat": lat, "lon": lon,
            "notes": f"自动检测到{tag_name}",
        }, ensure_ascii=False)

    @property
    def topic(self):
        return f"eregen/medical/wb/{self.dev_id}/up"


class CommunityDevice:
    def __init__(self, idx):
        self.dev_id = uid("CW-", idx)
        self.elder_id = f"E-{random.randint(10000, 99999)}"
        self.elder_name = ELDER_NAMES[(idx - 1) % len(ELDER_NAMES)]
        self.hospital_id = random.choice(["HOSP-001", "HOSP-002", "HOSP-003"])
        self.tags = random.sample(WELFARE_TAGS, k=random.randint(1, 3))
        self.bat = random.randint(50, 100)

    def community_signin(self):
        period = random.choice(["morning", "afternoon", "evening"])
        lat, lon = rand_loc()
        return json.dumps({
            "type": "community_signin", "dev_id": self.dev_id, "ts": now_ts(),
            "elder_id": self.elder_id, "hospital_id": self.hospital_id,
            "period": period,
            "activated_tags": random.sample(self.tags, k=random.randint(1, len(self.tags))),
            "is_medical_signin": random.choice([True, False]),
            "is_welfare_signin": random.choice([True, False]),
            "lat": lat, "lon": lon,
        }, ensure_ascii=False)

    def welfare_update(self):
        action = random.choice(["assign", "revoke"])
        tag = random.choice(WELFARE_TAGS)
        return json.dumps({
            "type": "community_welfare_update", "dev_id": self.dev_id, "ts": now_ts(),
            "elder_id": self.elder_id, "tag_code": tag, "action": action,
        }, ensure_ascii=False)

    def community_dispense(self):
        period = random.choice(["morning", "afternoon", "evening"])
        num_items = random.randint(1, 4)
        items = random.sample(MEDICATIONS, k=num_items)
        total = round(random.uniform(15.0, 200.0), 2)
        insurance = round(total * random.uniform(0.3, 0.8), 2)
        return json.dumps({
            "type": "community_dispense", "dev_id": self.dev_id, "ts": now_ts(),
            "elder_id": self.elder_id, "hospital_id": self.hospital_id,
            "period": period, "items": items, "total_cost": total,
            "insurance_covered": insurance, "self_pay": round(total - insurance, 2),
        }, ensure_ascii=False)

    @property
    def topic(self):
        return f"eregen/community/wb/{self.dev_id}/up"


# ─── Message Generator ───────────────────────────────────────────────────────

def generate_message(device):
    if isinstance(device, BraceletDevice):
        r = random.random()
        if r < 0.40: return device.heartbeat()
        elif r < 0.70: return device.location()
        elif r < 0.92: return device.health()
        elif r < 0.97: return device.sos()
        else: return device.fall()
    elif isinstance(device, MedicalDevice):
        r = random.random()
        if r < 0.25: return device.patient_register()
        elif r < 0.65: return device.verification_scan()
        elif r < 0.85: return device.device_status()
        else: return device.alert_tag()
    elif isinstance(device, CommunityDevice):
        r = random.random()
        if r < 0.55: return device.community_signin()
        elif r < 0.80: return device.community_dispense()
        else: return device.welfare_update()


# ─── Main Simulator ──────────────────────────────────────────────────────────

def main():
    devices = []
    for i in range(1, BRACELET_COUNT + 1):
        devices.append(BraceletDevice(i))
    for i in range(1, MEDICAL_COUNT + 1):
        devices.append(MedicalDevice(i))
    for i in range(1, COMMUNITY_COUNT + 1):
        devices.append(CommunityDevice(i))

    stats = {"total": 0, "errors": 0}
    for mt in ["heartbeat", "location", "health", "sos", "fall",
                "patient_register", "verification_scan", "device_status", "alert_tag",
                "community_signin", "community_welfare_update", "community_dispense"]:
        stats[mt] = 0

    # paho-mqtt v2 API
    client = mqtt.Client(
        callback_api_version=mqtt.CallbackAPIVersion.VERSION2,
        client_id=MQTT_CLIENT_ID,
        clean_session=True,
    )

    def on_connect(client, userdata, flags, rc, properties=None):
        if rc == 0:
            print(f"  [MQTT] Connected to {MQTT_BROKER}:{MQTT_PORT} ✓")
        else:
            print(f"  [MQTT] Connection failed (rc={rc})")

    def on_disconnect(client, userdata, flags, rc, properties=None):
        print(f"  [MQTT] Disconnected (rc={rc})")

    client.on_connect = on_connect
    client.on_disconnect = on_disconnect

    print(f"\n{'='*70}")
    print(f"  Eregen Device Data Simulator — MQTT Mode")
    print(f"{'='*70}")
    print(f"  Devices: {len(devices)} total "
          f"({sum(1 for d in devices if isinstance(d, BraceletDevice))} bracelets, "
          f"{sum(1 for d in devices if isinstance(d, MedicalDevice))} medical, "
          f"{sum(1 for d in devices if isinstance(d, CommunityDevice))} community)")
    print(f"  Broker: {MQTT_BROKER}:{MQTT_PORT}")
    print(f"  Interval: {PUBLISH_INTERVAL}s per device per cycle")
    print(f"{'='*70}\n")

    try:
        client.connect(MQTT_BROKER, MQTT_PORT, keepalive=60)
    except Exception as e:
        print(f"  [ERROR] Failed to connect to MQTT broker: {e}")
        sys.exit(1)

    client.loop_start()
    start_time = time.time()

    cycle = 0
    print("Starting simulation... (Ctrl+C to stop)\n")

    try:
        while True:
            cycle_start = time.time()
            cycle += 1
            cycle_msgs = 0

            for i, device in enumerate(devices):
                if i > 0:
                    time.sleep(0.05)

                msg = generate_message(device)
                parsed = json.loads(msg)
                msg_type = parsed.get("type", "unknown")

                try:
                    client.publish(device.topic, msg, qos=1)
                    cycle_msgs += 1
                    stats["total"] += 1
                    stats[msg_type] = stats.get(msg_type, 0) + 1
                except Exception as e:
                    print(f"  [ERROR] Failed to publish {msg_type} from {device.dev_id}: {e}")
                    stats["errors"] = stats.get("errors", 0) + 1

            elapsed = time.time() - cycle_start
            remaining = max(0, PUBLISH_INTERVAL - elapsed)

            if cycle % 5 == 1 or cycle == 1:
                total_elapsed = time.time() - start_time
                rate = stats["total"] / total_elapsed if total_elapsed > 0 else 0
                print(f"  [Cycle {cycle:03d}] Published {cycle_msgs} msgs | "
                      f"Total: {stats['total']} | Rate: {rate:.1f} msg/s | "
                      f"Errors: {stats.get('errors', 0)}")
                for mt, cnt in sorted(stats.items()):
                    if mt not in ("total", "errors"):
                        print(f"    {mt:25s}: {cnt}")

            time.sleep(remaining)

    except KeyboardInterrupt:
        total = time.time() - start_time
        print(f"\n{'='*70}")
        print(f"  Simulation Stopped")
        print(f"  Duration: {total:.1f}s")
        print(f"  Total messages: {stats['total']}")
        print(f"  Errors: {stats.get('errors', 0)}")
        print(f"  Messages per type:")
        for mt, cnt in sorted(stats.items()):
            if mt not in ("total", "errors"):
                print(f"    {mt:25s}: {cnt}")
        print(f"{'='*70}")
        client.loop_stop()
        client.disconnect()
        sys.exit(0)


if __name__ == "__main__":
    main()
