#!/usr/bin/env python3
"""
Eregen 设备模拟器 — 无需硬件即可测试固件逻辑
用法: python3 simulate-device.py <device_type> [options]

设备类型:
  bracelet    — 手环 (GD32E230 模拟)
  pillbox     — 药盒 (ESP32-C3 模拟)
  wristband   — 医用腕带 (ESP32-S3 模拟)
  community   — 社区腕带

示例:
  python3 simulate-device.py bracelet --interval 30
  python3 simulate-device.py pillbox --rule "08:00:2" --rule "12:00:1"
  python3 simulate-device.py wristband --patient "P001" --ble
"""
import argparse
import json
import sys
import time
import random
import threading
from datetime import datetime

# ---- Configuration ----
DEFAULT_BROKER = "mqtt.eregen.dev"
DEFAULT_PORT = 8883
BROKER_DEV = "localhost"
PORT_DEV = 1883

# ---- Device ID Generation ----
def gen_device_id(prefix, count=4):
    import hashlib
    h = hashlib.md5(str(time.time()).encode()).hexdigest()[:count].upper()
    return f"{prefix}-{h}"

# ---- Simulated Sensor Data ----
class SensorSimulator:
    def __init__(self):
        self.hr_base = 72
        self.spo2_base = 98
        self.steps = 0
        self.battery = 95

    def sample_health(self):
        hr = max(55, min(110, self.hr_base + random.gauss(0, 5)))
        spo2 = max(92, min(100, self.spo2_base + random.gauss(0, 1)))
        self.steps += random.randint(5, 30)
        self.battery = max(0, self.battery - random.uniform(0, 0.01))
        return {
            "hr": int(hr),
            "spo2": int(spo2),
            "steps": self.steps,
            "battery": round(self.battery, 1),
        }

    def sample_location(self):
        return {
            "lat": 31.2304 + random.gauss(0, 0.001),
            "lon": 121.4737 + random.gauss(0, 0.001),
            "acc": random.randint(3, 15),
        }

    def sample_fall(self):
        conf = random.random()
        return {"conf": round(conf, 2), "lat": 31.23 + random.gauss(0, 0.001), "lon": 121.47 + random.gauss(0, 0.001)}


# ---- MQTT Publisher (simulated) ----
class MockMQTT:
    def __init__(self, broker, port):
        self.broker = broker
        self.port = port
        self.connected = False

    def connect(self):
        print(f"  [MQTT] Connecting to {self.broker}:{self.port} ...")
        time.sleep(0.5)
        self.connected = True
        print(f"  [MQTT] Connected ✓")
        return True

    def publish(self, topic, payload, qos=1):
        if not self.connected:
            return False
        msg = json.dumps(payload) if isinstance(payload, dict) else payload
        print(f"  [MQTT] PUBLISH {topic}: {msg[:80]}{'...' if len(msg)>80 else ''}")
        return True

    def subscribe(self, topic, callback):
        print(f"  [MQTT] SUBSCRIBE {topic}")


# ---- Device Simulator ----
class DeviceSimulator:
    def __init__(self, device_type, args):
        self.device_type = device_type
        self.dev_id = gen_device_id(device_type[:2].upper())
        self.sensor = SensorSimulator()
        self.mqtt = MockMQTT(args.broker, args.port)
        self.running = False
        self.patient_id = getattr(args, 'patient', None)

    def start(self):
        self.running = True
        self.mqtt.connect()
        print(f"\n{'='*60}")
        print(f"  Eregen {self.device_type.upper()} Simulator")
        print(f"  Device ID: {self.dev_id}")
        print(f"  Patient: {self.patient_id or 'UNBOUND'}")
        print(f"  Broker: {self.mqtt.broker}:{self.mqtt.port}")
        print(f"{'='*60}\n")

        if self.device_type == "bracelet":
            self._run_bracelet()
        elif self.device_type == "pillbox":
            self._run_pillbox()
        elif self.device_type == "wristband":
            self._run_wristband()
        elif self.device_type == "community":
            self._run_community()

    def _run_bracelet(self):
        """Simulate bracelet heartbeat + health + location + SOS/fall"""
        interval = getattr(self.args, 'interval', 30)
        print(f"  [Bracelet] Sampling every {interval}s")
        sub = f"eregen/device/bracelet/{self.dev_id}/cmd"
        self.mqtt.subscribe(sub, self._handle_cmd)

        while self.running:
            # Heartbeat
            bat = self.sensor.battery
            self.mqtt.publish(f"eregen/device/bracelet/{self.dev_id}", {
                "type": "heartbeat", "dev_id": self.dev_id, "bat": int(bat),
                "ts": int(time.time())
            })
            # Health
            h = self.sensor.sample_health()
            self.mqtt.publish(f"eregen/device/bracelet/{self.dev_id}/health", h)
            # Location
            loc = self.sensor.sample_location()
            self.mqtt.publish(f"eregen/device/bracelet/{self.dev_id}/location", loc)
            # Random SOS/fall (once every ~5 minutes)
            if random.random() < 0.003:
                if random.random() < 0.5:
                    self.mqtt.publish(f"eregen/device/bracelet/{self.dev_id}/alert", {
                        "type": "sos", "dev_id": self.dev_id,
                        "ts": int(time.time())
                    })
                else:
                    f = self.sensor.sample_fall()
                    self.mqtt.publish(f"eregen/device/bracelet/{self.dev_id}/alert", {
                        "type": "fall", "dev_id": self.dev_id, **f, "ts": int(time.time())
                    })
            time.sleep(interval)

    def _run_pillbox(self):
        """Simulate pillbox med status + inventory"""
        rules = getattr(self.args, 'rules', [])
        print(f"  [Pillbox] Rules: {rules or 'none'}")
        self.mqtt.subscribe(f"eregen/device/pillbox/{self.dev_id}/cmd", self._handle_cmd)

        while self.running:
            self.mqtt.publish(f"eregen/device/pillbox/{self.dev_id}", {
                "type": "heartbeat", "dev_id": self.dev_id,
                "bat": 88, "ts": int(time.time())
            })
            # Simulate medication event
            if random.random() < 0.02:
                self.mqtt.publish(f"eregen/device/pillbox/{self.dev_id}/med", {
                    "type": "med_status", "dev_id": self.dev_id,
                    "compartment": random.randint(1, 4),
                    "taken": random.choice([True, False]),
                    "ts": int(time.time())
                })
            time.sleep(60)

    def _run_wristband(self):
        """Simulate medical wristband BLE + MQTT"""
        print(f"  [Wristband] BLE advertising 'Eregen-WB-{self.dev_id[-4:]}')")
        self.mqtt.subscribe(f"eregen/device/wb/{self.dev_id}/cmd", self._handle_cmd)

        while self.running:
            self.mqtt.publish(f"eregen/device/wb/{self.dev_id}", {
                "type": "heartbeat", "dev_id": self.dev_id,
                "bat": 92, "patient_id": self.patient_id or "UNBOUND",
                "ts": int(time.time())
            })
            time.sleep(120)

    def _run_community(self):
        """Simulate community wristband signin + welfare"""
        print(f"  [Community] Signin beacon broadcast")
        while self.running:
            self.mqtt.publish(f"eregen/device/community-wb/{self.dev_id}", {
                "type": "community_signin", "dev_id": self.dev_id,
                "elder_id": "elder-001", "period": datetime.now().strftime("%Y-%m"),
                "activated_tags": ["ELDERLY", "SUBSIDY"],
                "ts": int(time.time())
            })
            time.sleep(300)

    def _handle_cmd(self, topic, payload):
        """Handle incoming commands"""
        data = json.loads(payload) if isinstance(payload, str) else payload
        cmd_type = data.get("type", "")
        if cmd_type == "config":
            print(f"  [CMD] Config update: {data.get('settings')}")
        elif cmd_type == "tts":
            print(f"  [CMD] TTS: {data.get('text')}")
        elif cmd_type == "ota":
            print(f"  [CMD] OTA: {data.get('url')}")

    def stop(self):
        self.running = False
        print("\n  [Simulator] Stopped.")


# ---- Main ----
def main():
    parser = argparse.ArgumentParser(description="Eregen Device Simulator")
    parser.add_argument("device", choices=["bracelet", "pillbox", "wristband", "community"])
    parser.add_argument("--broker", default=BROKER_DEV)
    parser.add_argument("--port", type=int, default=PORT_DEV)
    parser.add_argument("--interval", type=int, default=30, help="Sampling interval (seconds)")
    parser.add_argument("--patient", help="Patient ID for wristband")
    parser.add_argument("--rule", action="append", help="Medication rule (HH:MM:dose)")
    parser.add_argument("--ble", action="store_true", help="Simulate BLE advertising")
    args = parser.parse_args()

    sim = DeviceSimulator(args.device, args)
    try:
        sim.start()
    except KeyboardInterrupt:
        sim.stop()
        print("\n  [Simulator] Exit.")


if __name__ == "__main__":
    main()
