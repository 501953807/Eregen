# ② 药盒固件 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：② 药盒固件 (ESP32-C3 + ESP-IDF v5.3)  
> 语言：C | RTOS：ESP-IDF (FreeRTOS) | 通信：WiFi MQTT

---

## 1. 概述

### 1.1 职责

药盒固件负责按云端下发的用药规则定时提醒老人服药，通过语音播报 (TTS) 和 LED 指示灯提示，检测药物是否被取用，并将服药状态上报至云平台。智能版和自动版还支持 BLE 配网、库存预警和自动分药功能。

### 1.2 三档差异化

| 功能模块 | Basic (基础) | Smart (中端) | Auto (高端) |
|---------|:---:|:---:|:---:|
| 分格药盒 | ✅ | ✅ | ✅ (带光电检测) |
| 语音提醒 (TTS) | ❌ | ✅ (SYN5300) | ✅ |
| WiFi/MQTT 通信 | ❌ | ✅ | ✅ |
| APP 联动 | ❌ | ✅ | ✅ |
| OLED 状态显示 | ❌ | ✅ (SSD1306) | ✅ |
| 步进电机控制 | ❌ | ❌ | ✅ |
| 自动分药机构 | ❌ | ❌ | ✅ |
| 光电药片检测 | ❌ | ❌ | ✅ (红外对管) |
| 库存预警 | ❌ | ❌ | ✅ |
| 音量调节 | ❌ | ✅ | ✅ |

### 1.3 输入输出

| 类型 | 数据源/目标 | 接口 |
|------|-----------|------|
| **输入** | 光电传感器 (红外对管) | GPIO ADC |
| **输入** | OLED 显示屏 (Smart/Auto) | I2C SSD1306 |
| **输入** | 按键输入 (Basic) | GPIO |
| **输出** | TTS 语音模块 (SYN5300) | UART |
| **输出** | 步进电机 (Auto) | PWM GPIO |
| **输出** | LED 指示灯 | GPIO |
| **输出** | WiFi → EMQX MQTT | TCP/IP |
| **输出** | BLE GATT (Auto) | BLE 5.0 |

---

## 2. 核心数据结构

### 2.1 用药规则 (云端下发)

```c
typedef struct {
    char time_str[8];       // "08:00" / "12:30" / "21:00"
    int dose_count;         // 本次服用剂量 (片数)
    char pill_type[32];     // "capsule" / "tablet" / "liquid"
} MedDose;

typedef struct {
    char rule_id[32];
    char schedule_time[8];  // HH:MM
    int dose_count;
    char pill_type[32];
    int days_of_week[7];    // 0=周日, 1=周一 … 6=周六
    int active;             // 1=启用 0=停用
} MedicationRule;
```

### 2.2 服药状态上报

```c
typedef struct {
    char type[32];          // "med_status"
    char dev_id[16];        // "PX-XXXX"
    int compartment;        // 格子编号 (1-8)
    int taken;              // 1=已服 0=未服
    int ts;                 // Unix timestamp
} MedStatusMessage;
```

### 2.3 状态机

```
┌──────────────────────────────────────────────────────┐
│                    IDLE                               │ ← 等待用药时间
└──────────────────┬───────────────────────────────────┘
                   │ 到达用药时间
                   ▼
           ┌───────────────┐
           │ REMINDER      │ ← TTS 语音播报 + LED 闪烁
           │               │   "爷爷，该吃降压药了"
           └───────┬───────┘
                   │ 用户取药 (光电检测/手动确认)
                   ▼
           ┌───────────────┐
           │ CONFIRMED     │ ← 上报 med_status(taken=true)
           │               │ ← 重置计时器
           └───────┬───────┘
                   │ 超时未取 (30min)
                   ▼
           ┌───────────────┐
           │ MISSED        │ ← 上报 med_status(taken=false)
           │               │ ← 云端触发 P1 告警
           └───────────────┘
```

---

## 3. 功能模块说明

### 3.1 自动分药 (Auto 版本)

| 文件 | 模块 | 说明 |
|------|------|------|
| `dispensing.c/h` | 分药控制 | 根据用药规则计算目标格子，驱动步进电机转动到指定位置 |
| `state_machine.c/h` | 状态机 | 分药流程状态管理 (IDLE→PREPARE→DISPENSE→CONFIRM→IDLE) |
| `motor_control.c/h` (common) | 电机控制 | 28BYJ-48 步进电机驱动，正反转+步数控制 |

### 3.2 药片检测 (Auto 版本)

| 文件 | 模块 | 说明 |
|------|------|------|
| `opto_sensor.c/h` | 光电传感器 | 红外对管读取，判断格子是否有药片 |
| `empty_detector.c/h` | 空盒检测 | 统计连续 N 次为空判定为缺药，触发库存预警 |

### 3.3 用药规则引擎

| 文件 | 模块 | 说明 |
|------|------|------|
| `schedule_engine.c/h` | 调度引擎 | 解析云端下发的 JSON 用药规则，维护每日用药计划 |
| `med_rule_parser.c/h` | 规则解析 | JSON → C 结构体，支持动态增删规则 |
| `nvs_store.c/h` | NVS 存储 | 用药规则持久化到 ESP32 非易失存储 |

### 3.4 语音提醒 (Smart/Auto)

| 文件 | 模块 | 说明 |
|------|------|------|
| `tts_playback.c/h` (common) | TTS 播放 | SYN5300 语音模块 UART 控制，文本转语音 |
| `voice_reminder.c/h` (Smart) | 语音提醒 | 定时触发 TTS 播报，支持自定义话术 |
| `volume_control.c/h` (Smart) | 音量控制 | 调节 TTS 音量，NV 存储用户设置 |

### 3.5 显示与交互 (Smart/Auto)

| 文件 | 模块 | 说明 |
|------|------|------|
| `oled_status.c/h` (Smart) | OLED 显示 | SSD1306 显示当前时间、下次用药提醒、电量 |
| `button_input.c/h` (Basic) | 按键输入 | 确认服药/跳过服药 |
| `led_gpio.c/h` (Basic) | LED 指示 | 红灯=SOS/告警，绿灯=正常，蓝灯=待配网 |
| `led_patterns.c/h` (common) | LED 灯效 | 呼吸/闪烁/常亮等预设模式 |

### 3.6 通信层

| 文件 | 模块 | 说明 |
|------|------|------|
| `wifi_mqtt_bridge.c/h` (common) | WiFi+MQTT | ESP-MQTT 库封装，连接管理+重连+心跳 |
| `device_id.c/h` (common) | 设备标识 | 从 ESP32 MAC 地址派生 PX-XXXX 格式 ID |
| `ap_config_mode.c/h` (Auto) | AP 配网模式 | 无 WiFi 时开启热点，手机连接配置 |
| `ble_pair.c/h` (Auto) | BLE 配网 | BLE GATT 服务接收 WiFi 凭证 |
| `app_link.c/h` (Smart) | APP 联动 | MQTT 双向同步，APP 可远程配置/查看 |

### 3.7 电源管理 (Basic)

| 文件 | 模块 | 说明 |
|------|------|------|
| `battery_manage.c/h` (Basic) | 电池管理 | 充电检测、电量估算、低电量告警 |

---

## 4. ESP-IDF 任务划分

### Smart 版本任务

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| `AppInitTask` | 最高 | 1024 字 | WiFi 连接、MQTT 初始化、系统启动 |
| `ScheduleTask` | 高 | 768 字 | 检查当前时间是否到用药提醒 |
| `VoiceTask` | 高 | 512 字 | TTS 语音播报 |
| `DisplayTask` | 中 | 512 字 | OLED 屏幕刷新 |
| `CommTask` | 中 | 1024 字 | MQTT 消息收发 |
| `SensorTask` | 低 | 512 字 | 按键检测、LED 状态更新 |

### Auto 版本额外任务

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| `DispenseTask` | 高 | 1024 字 | 自动分药流程控制 |
| `OptoTask` | 中 | 512 字 | 光电传感器采样 |
| `BLETask` | 中 | 768 字 | BLE 配网服务 |

---

## 5. 接口定义

### 5.1 上行消息 (设备 → 云平台)

```json
// 服药状态
{"type":"med_status","dev_id":"PX-0001","compartment":3,"taken":true,"ts":1720000000}

// 库存预警 (Auto)
{"type":"med_alert","dev_id":"PX-0001","alert_type":"low_stock","compartment":5,"remaining":2,"ts":1720000000}
```

### 5.2 下行指令 (云平台 → 设备)

```json
// 用药规则
{"type":"med_rule","dev_id":"PX-0001","rules":[
  {"time":"08:00","dose":1,"type":"capsule"},
  {"time":"21:00","dose":2,"type":"tablet"}
]}

// 配置更新
{"type":"config","dev_id":"PX-0001","settings":{"volume":80,"reminder_duration":30}}

// 语音播报
{"type":"tts","dev_id":"PX-0001","text":"爷爷，该吃降压药了"}

// OTA 升级
{"type":"ota","dev_id":"PX-0001","url":"https://...","hash":"sha256:..."}
```

---

## 6. 编译与烧录

### 6.1 环境要求

```bash
# 安装 ESP-IDF v5.3
git clone --branch v5.3.1 --depth 1 https://github.com/espressif/esp-idf.git ~/esp/esp-idf
cd ~/esp/esp-idf && ./install.sh esp32c3
source ~/esp/esp-idf/export.sh
```

### 6.2 编译 Smart 版本

```bash
cd firmware/pillbox/smart
idf.py set-target esp32c3
idf.py build
```

### 6.3 编译 Auto 版本

```bash
cd ../../auto && idf.py build
```

### 6.4 烧录

```bash
# 连接 ESP32-C3-DevKitM-1 (USB-C)
idf.py -p /dev/ttyUSB0 flash monitor
# Ctrl+] 退出 monitor
```

---

## 7. 测试策略

### 7.1 单元测试

```bash
# 调度引擎测试
gcc -DTEST_MODE smart/test_scheduler.c smart/reminder_scheduler.c -o test_scheduler
./test_scheduler

# 状态机测试
gcc -DTEST_MODE auto/test_state_machine.c auto/state_machine.c -o test_state
./test_state

# 用药规则解析测试
gcc -DTEST_MODE auto/test_med_rule_parser.c auto/med_rule_parser.c -o test_parser
./test_parser

# 光电检测测试
gcc -DTEST_MODE auto/test_empty_detector.c auto/empty_detector.c -o test_empty
./test_empty
```

### 7.2 集成测试

- 步进电机分药精度测试：旋转 N 步，测量实际到达位置偏差 < 2°
- 光电检测准确率测试：有药/无药状态识别率 > 99%
- MQTT 端到端：下发用药规则 → 设备解析 → 定时播报 → 服药确认 → 上报云端

---

## 8. 功耗设计

| 模式 | 电流 | 说明 |
|------|------|------|
| WiFi 连接+MQTT | ~80mA | 正常工作 |
| Deep Sleep | ~10μA | ESP32-C3 深度休眠，RTC 定时器唤醒 |
| Light Sleep | ~1mA | WiFi 断开但保留状态，快速重连 |

**续航：** 2500mAh 18650 锂电池，Smart 版本平均电流 30mA → 约 83 小时 (~3.5天)，建议每周充电一次。

---

© 2026 Eregen (颐贞). All rights reserved.
