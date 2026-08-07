#!/usr/bin/env python3
"""
Eregen Platform — MQTT Message Subscriber
Subscribes to all device topics and prints received messages for verification.
"""
import json
import sys
import time
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("ERROR: paho-mqtt not installed. Run: pip3 install paho-mqtt")
    sys.exit(1)

MQTT_BROKER = "localhost"
MQTT_PORT = 1883

TOPICS = [
    "eregen/device/bracelet/#",
    "eregen/medical/wb/#",
    "eregen/community/wb/#",
]

stats = {
    "total": 0,
    "errors": 0,
    "by_type": {},
    "by_device_type": {"bracelet": 0, "medical": 0, "community": 0},
    "unique_devices": set(),
}

def on_connect(client, userdata, flags, rc, properties=None):
    if rc == 0:
        print(f"  [MQTT] Connected to {MQTT_BROKER}:{MQTT_PORT}")
        for topic in TOPICS:
            client.subscribe(topic, qos=1)
            print(f"  [MQTT] Subscribed to {topic}")
    else:
        print(f"  [MQTT] Connection failed (rc={rc})")

def on_message(client, userdata, msg):
    try:
        parsed = json.loads(msg.payload.decode())
        msg_type = parsed.get("type", "unknown")
        dev_id = parsed.get("dev_id", "unknown")
        
        stats["total"] += 1
        stats["by_type"][msg_type] = stats["by_type"].get(msg_type, 0) + 1
        stats["unique_devices"].add(dev_id)
        
        # Determine device type from topic
        if "bracelet" in msg.topic:
            stats["by_device_type"]["bracelet"] += 1
        elif "medical" in msg.topic:
            stats["by_device_type"]["medical"] += 1
        elif "community" in msg.topic:
            stats["by_device_type"]["community"] += 1
        
        # Print first few messages of each type for verification
        if stats["by_type"].get(msg_type, 0) <= 2:
            ts = datetime.fromtimestamp(parsed.get("ts", 0)).strftime("%H:%M:%S")
            print(f"  [{ts}] {msg_type:25s} from {dev_id}")
            
    except Exception as e:
        stats["errors"] += 1
        print(f"  [ERROR] Failed to parse message: {e}")

def main():
    print(f"\n{'='*70}")
    print(f"  Eregen MQTT Message Subscriber")
    print(f"{'='*70}")
    print(f"  Broker: {MQTT_BROKER}:{MQTT_PORT}")
    print(f"  Topics: {TOPICS}")
    print(f"{'='*70}\n")
    print("Waiting for messages... (Ctrl+C to stop)\n")
    
    client = mqtt.Client(
        callback_api_version=mqtt.CallbackAPIVersion.VERSION2,
        client_id="eregen-subscriber",
        clean_session=True,
    )
    client.on_connect = on_connect
    client.on_message = on_message
    
    try:
        client.connect(MQTT_BROKER, MQTT_PORT, keepalive=60)
    except Exception as e:
        print(f"  [ERROR] Failed to connect: {e}")
        sys.exit(1)
    
    client.loop_start()
    start_time = time.time()
    
    try:
        while True:
            time.sleep(5)
            elapsed = time.time() - start_time
            if elapsed >= 30:  # Run for 30 seconds
                break
    except KeyboardInterrupt:
        pass
    finally:
        client.loop_stop()
        
        print(f"\n{'='*70}")
        print(f"  Subscription Summary (30s)")
        print(f"{'='*70}")
        print(f"  Total messages received: {stats['total']}")
        print(f"  Errors: {stats['errors']}")
        print(f"  Unique devices: {len(stats['unique_devices'])}")
        print(f"\n  Messages by device type:")
        for dt, cnt in stats["by_device_type"].items():
            print(f"    {dt:15s}: {cnt}")
        print(f"\n  Messages by type:")
        for mt, cnt in sorted(stats["by_type"].items()):
            print(f"    {mt:25s}: {cnt}")
        print(f"{'='*70}")

if __name__ == "__main__":
    main()
