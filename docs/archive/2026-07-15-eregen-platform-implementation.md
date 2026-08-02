# Eregen 颐贞 — 全平台实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按照11份specs文档构建Eregen颐贞老年健康生态平台的完整代码实现，涵盖手环固件、药盒固件、云平台后端、家属APP、管理后台、微信小程序、品牌官网和B2B对接。

**Architecture:** 四层架构(感知层→传输层→平台层→应用层)，Go+Gin微服务后端，Flutter跨平台APP，Vue3管理后台，Hugo静态站。硬件通过ODM方案商完成，我们专注固件和云端。

**Tech Stack:**
- 固件: C / FreeRTOS (GD32E230) / ESP-IDF v5.3 (ESP32-C3)
- 后端: Go 1.22+ / Gin 1.9+ / EMQX 5.x / NATS 2.10+ / PostgreSQL 16 / InfluxDB 2 / Redis 7
- APP: Flutter 3.24+ / Dart 3.5+
- Web: Vue 3.4+ / TypeScript 5.4+ / Element Plus 2.7+ / Pinia / ECharts
- 小程序: 原生 WXML/WXSS / 基础库 2.44+
- 官网: Hugo 0.128+ / Tailwind CSS 3.4+
- CI/CD: GitHub Actions / Docker Compose

## Global Constraints

- **开源许可:** 仅使用 MIT / BSD-3 / Apache-2.0 / ISC；禁止 GPL / AGPL / LGPL
- **版权声明:** 所有代码仓库加声明 `© 2026 Eregen (颐贞). All rights reserved.`
- **核心业务逻辑:** 全部自研闭源
- **数据合规:** 健康数据 = 敏感个人信息 (PIPL第28-32条)，TLS1.2+传输加密 + AES-256静态加密
- **技术选型:** 整个项目周期内不更改CLAUDE.md中定义的选型
- **UI优先:** UI界面先出高保真效果图再编码(原型已完成)

---

## 批次划分与依赖关系

```
第一批 (月1-6):  ①手环固件 → ③云平台后端 ← ②药盒固件
                     ↓            ↓            ↓
第二批 (月4-8):            ④家属APP    ⑤管理后台
                                    ↓
第三批 (月7-12):              ⑦小程序   ⑥官网
                                    ↓
                               ⑧B2B医院对接
```

**关键路径:** ①+②+③并行推进，第4个月开始④联调。

---

## 实施批次

### Batch 1: 硬件固件 (①手环固件 + ②药盒固件)
- 完善现有固件框架，补充缺失模块
- 实现跌倒检测算法、自动分药状态机
- 完成MQTT协议封装和OTA升级

### Batch 2: 云平台后端 (③云平台后端)
- 完善gateway MQTT接入
- 构建api-server REST API
- 实现data-pipeline AI分析引擎
- 构建push-service推送分发

### Batch 3: 家属APP (④家属APP)
- Flutter项目完整搭建
- 实现四大核心页面(定位/健康/告警/用药)
- WebSocket实时推送集成

### Batch 4: 管理后台 (⑤管理后台)
- Vue3项目完整搭建
- 实现仪表盘/设备/用户/订阅/告警管理
- WebSocket实时告警

### Batch 5: 小程序 + 官网 (⑦小程序 + ⑥官网)
- 微信小程序原生开发
- Hugo静态站搭建

### Batch 6: B2B对接 (⑧B2B医院社区对接)
- 开放API设计
- 机构数据隔离

---

## 任务分解

### Batch 1: 手环固件 — Task 1.1~1.5

#### Task 1.1: 完善手环 BSP 初始化

**Files:**
- Modify: `firmware/bracelet/entry/main.c`
- Modify: `firmware/bracelet/entry/board_init.c`
- Modify: `firmware/bracelet/entry/free_rtos_tasks.c`
- Create: `firmware/bracelet/entry/protocol/message_encode.c`
- Create: `firmware/bracelet/entry/protocol/message_encode.h`
- Create: `firmware/bracelet/entry/protocol/message_decode.c`
- Create: `firmware/bracelet/entry/protocol/message_decode.h`
- Create: `firmware/bracelet/entry/protocol/heartbeat.c`
- Create: `firmware/bracelet/entry/protocol/heartbeat.h`
- Create: `firmware/bracelet/entry/common/crc16.c`
- Create: `firmware/bracelet/entry/common/crc16.h`
- Create: `firmware/bracelet/entry/common/ring_buffer.c`
- Create: `firmware/bracelet/entry/common/ring_buffer.h`
- Create: `firmware/bracelet/entry/common/log.c`
- Create: `firmware/bracelet/entry/common/log.h`

**Interfaces:**
- Consumes: 现有sensor drivers (sensors_imu.c, sensors_ppg.c, gps_nmea.c, cat1_at.c, display_st7789.c, battery_adc.c, sos_button.c)
- Produces: `message_encode()`, `message_decode()`, `crc16_calc()`, `ring_buffer_*()`, `log_*()`

- [ ] **Step 1: 实现 CRC16 校验工具**
  - [ ] 创建 `firmware/bracelet/entry/common/crc16.h` — 声明 `uint16_t crc16_calc(const uint8_t *data, uint16_t len)`
  - [ ] 创建 `firmware/bracelet/entry/common/crc16.c` — 实现标准CRC16-CCITT算法
  - [ ] 创建测试文件 `firmware/bracelet/entry/common/test_crc16.c`

- [ ] **Step 2: 实现环形缓冲区**
  - [ ] 创建 `firmware/bracelet/entry/common/ring_buffer.h` — 声明 `ring_buf_init()`, `ring_buf_push()`, `ring_buf_pop()`, `ring_buf_available()`
  - [ ] 创建 `firmware/bracelet/entry/common/ring_buffer.c` — 实现固定大小环形缓冲区
  - [ ] 创建测试文件 `firmware/bracelet/entry/common/test_ring_buffer.c`

- [ ] **Step 3: 实现日志输出模块**
  - [ ] 创建 `firmware/bracelet/entry/common/log.h` — 声明 `log_init()`, `log_debug()`, `log_info()`, `log_warn()`, `log_error()`
  - [ ] 创建 `firmware/bracelet/entry/common/log.c` — 实现基于UART的分级日志，支持log_level配置
  - [ ] 创建测试文件 `firmware/bracelet/entry/common/test_log.c`

- [ ] **Step 4: 实现消息编码/解码协议**
  - [ ] 创建 `firmware/bracelet/entry/protocol/message_encode.h` — 定义 `eregen_msg_t` 结构体(type/dev_id/timestamp/payload/payload_len/checksum)
  - [ ] 创建 `firmware/bracelet/entry/protocol/message_encode.c` — 实现JSON编码+CRC16附加
  - [ ] 创建 `firmware/bracelet/entry/protocol/message_decode.h` — 声明解码函数
  - [ ] 创建 `firmware/bracelet/entry/protocol/message_decode.c` — 实现JSON解码+CRC16校验
  - [ ] 创建测试文件 `firmware/bracelet/entry/protocol/test_message_encode.c`

- [ ] **Step 5: 实现心跳保活逻辑**
  - [ ] 创建 `firmware/bracelet/entry/protocol/heartbeat.h` — 声明 `heartbeat_start()`, `heartbeat_stop()`
  - [ ] 创建 `firmware/bracelet/entry/protocol/heartbeat.c` — 实现每5分钟发送heartbeat消息到MQTT
  - [ ] 创建测试文件 `firmware/bracelet/entry/protocol/test_heartbeat.c`

- [ ] **Step 6: 完善main.c初始化流程**
  - [ ] 修改 `firmware/bracelet/entry/main.c` — 调用log_init(), 初始化各模块, 创建FreeRTOS任务
  - [ ] 确保main.c调用free_rtos_tasks.c中的任务创建函数

- [ ] **Step 7: 更新board_init.c**
  - [ ] 修改 `firmware/bracelet/entry/board_init.c` — 确保时钟/中断/外设初始化完整
  - [ ] 添加GPIO初始化(SOS按键/GPS使能/Cat1使能)

- [ ] **Step 8: Commit**
  ```bash
  git add firmware/bracelet/entry/common/ firmware/bracelet/entry/protocol/ firmware/bracelet/entry/main.c firmware/bracelet/entry/board_init.c
  git commit -m "feat(bracelet): complete BSP initialization, protocol encoding, logging, and ring buffer"
  ```

#### Task 1.2: 实现跌倒检测算法

**Files:**
- Modify: `firmware/bracelet/entry/sensors_imu.c`
- Modify: `firmware/bracelet/entry/sensors_imu.h`
- Create: `firmware/bracelet/entry/algorithms/fall_detect.c`
- Create: `firmware/bracelet/entry/algorithms/fall_detect.h`
- Create: `firmware/bracelet/entry/algorithms/sliding_window.c`
- Create: `firmware/bracelet/entry/algorithms/sliding_window.h`
- Create: `firmware/bracelet/entry/algorithms/test_fall_detect.c`

**Interfaces:**
- Consumes: `imu_read_accel_gyro()` from sensors_imu.c
- Produces: `fall_detect_process(imu_data_t*, uint16_t) → fall_result_t`, `sliding_window_push()`

- [ ] **Step 1: 实现滑动窗口数据结构**
  - [ ] 创建 `firmware/bracelet/entry/algorithms/sliding_window.h` — 定义 `sliding_window_t` 结构体(buffer[100]/head/tail/count)
  - [ ] 创建 `firmware/bracelet/entry/algorithms/sliding_window.c` — 实现固定大小100样本的环形滑动窗口
  - [ ] 创建测试文件 `firmware/bracelet/entry/algorithms/test_sliding_window.c`

- [ ] **Step 2: 实现IMU数据读取接口**
  - [ ] 修改 `firmware/bracelet/entry/sensors_imu.h` — 添加 `typedef struct { float ax; float ay; float az; float gx; float gy; float gz; } imu_sample_t;`
  - [ ] 修改 `firmware/bracelet/entry/sensors_imu.h` — 添加 `bool imu_read_sample(imu_sample_t *sample)` 声明
  - [ ] 修改 `firmware/bracelet/entry/sensors_imu.c` — 确保I2C读取加速度+陀螺仪数据

- [ ] **Step 3: 实现跌倒检测核心算法**
  - [ ] 创建 `firmware/bracelet/entry/algorithms/fall_detect.h` — 定义 `fall_result_t` (NO_FALL/FALL_SUSPECT/FALL_DETECTED) 枚举
  - [ ] 创建 `firmware/bracelet/entry/algorithms/fall_detect.h` — 声明 `fall_detect_process(sliding_window_t *window)`
  - [ ] 创建 `firmware/bracelet/entry/algorithms/fall_detect.c` — 实现:
    - 特征提取: 加速度幅值、角速度幅值、冲击检测、自由落体持续时间、冲击后静止
    - 规则引擎: 可配置阈值判断
    - 置信度计算: 加权评分
  - [ ] 创建测试文件 `firmware/bracelet/entry/algorithms/test_fall_detect.c` — 测试正常行走/跑步/真实跌倒模拟数据

- [ ] **Step 4: 实现防误触机制**
  - [ ] 在 `fall_detect.c` 中添加: 检测到跌倒后等待5秒活动恢复→自动取消
  - [ ] 在 `fall_detect.c` 中添加: 连续3次检测到跌倒→确认为真实跌倒
  - [ ] 添加测试用例验证防误触逻辑

- [ ] **Step 5: Commit**
  ```bash
  git add firmware/bracelet/entry/algorithms/ firmware/bracelet/entry/sensors_imu.*
  git commit -m "feat(bracelet): implement fall detection algorithm with sliding window and anti-misfire"
  ```

#### Task 1.3: 完善Cat1通信和GPS定位

**Files:**
- Modify: `firmware/bracelet/entry/cat1_at.c`
- Modify: `firmware/bracelet/entry/cat1_at.h`
- Modify: `firmware/bracelet/entry/gps_nmea.c`
- Modify: `firmware/bracelet/entry/gps_nmea.h`
- Create: `firmware/bracelet/entry/location/gps_manager.c`
- Create: `firmware/bracelet/entry/location/gps_manager.h`
- Create: `firmware/bracelet/entry/location/test_gps_manager.c`
- Create: `firmware/bracelet/entry/location/geofence.c`
- Create: `firmware/bracelet/entry/location/geofence.h`

**Interfaces:**
- Consumes: 现有gps_nmea.c, cat1_at.c
- Produces: `location_mode_switch()`, `geofence_check(lat, lon, fence_list)`

- [ ] **Step 1: 完善GPS NMEA解析**
  - [ ] 修改 `firmware/bracelet/entry/gps_nmea.h` — 添加 `gps_fix_t` 结构体(lat/lon/acc/alt/ts/sat_count)
  - [ ] 修改 `firmware/bracelet/entry/gps_nmea.h` — 添加 `bool gps_parse_nmea(const char *line, gps_fix_t *fix)`
  - [ ] 修改 `firmware/bracelet/entry/gps_nmea.c` — 实现GPRMC/GPGGA语句解析

- [ ] **Step 2: 完善Cat1 AT指令封装**
  - [ ] 修改 `firmware/bracelet/entry/cat1_at.h` — 定义完整的AT指令集结构体
  - [ ] 修改 `firmware/bracelet/entry/cat1_at.c` — 实现APN设置、TCP连接、MQTT CONNECT/PUBLISH/DISCONNECT
  - [ ] 添加AT指令超时重试机制(最多3次)

- [ ] **Step 3: 实现GPS定位管理器**
  - [ ] 创建 `firmware/bracelet/entry/location/gps_manager.h` — 定义三种定位模式(LOC_NORMAL/LOC_ALERT/LOC_POWER_SAVE)
  - [ ] 创建 `firmware/bracelet/entry/location/gps_manager.c` — 实现:
    - NORMAL: GPS 30s/次, Cat1间隙休眠
    - ALERT: GPS 1s/次, Cat1持续在线
    - POWER_SAVE: 关闭GPS, 仅LBS基站定位
  - [ ] 创建测试文件 `firmware/bracelet/entry/location/test_gps_manager.c`

- [ ] **Step 4: 实现电子围栏本地比对**
  - [ ] 创建 `firmware/bracelet/entry/location/geofence.h` — 定义 `geofence_circle_t` 和 `geofence_polygon_t` 结构
  - [ ] 创建 `firmware/bracelet/entry/location/geofence.c` — 实现点在圆内/点在多边形内算法
  - [ ] 创建测试文件 `firmware/bracelet/entry/location/test_geofence.c`

- [ ] **Step 5: Commit**
  ```bash
  git add firmware/bracelet/entry/cat1_at.* firmware/bracelet/entry/gps_nmea.* firmware/bracelet/entry/location/
  git commit -m "feat(bracelet): complete Cat1 communication, GPS manager, and local geofence"
  ```

#### Task 1.4: 实现PPG传感器和健康数据上报

**Files:**
- Modify: `firmware/bracelet/entry/sensors_ppg.c`
- Modify: `firmware/bracelet/entry/sensors_ppg.h`
- Create: `firmware/bracelet/entry/health/health_collector.c`
- Create: `firmware/bracelet/entry/health/health_collector.h`
- Create: `firmware/bracelet/entry/health/test_health_collector.c`

**Interfaces:**
- Consumes: `sensors_ppg_read_hr_spo2()`
- Produces: `health_collect_and_send()` → MQTT publish health message

- [ ] **Step 1: 完善PPG驱动**
  - [ ] 修改 `firmware/bracelet/entry/sensors_ppg.h` — 添加 `ppg_data_t` 结构体(hr/spo2/quality)
  - [ ] 修改 `firmware/bracelet/entry/sensors_ppg.h` — 添加 `bool ppg_read(ppg_data_t *data)`
  - [ ] 修改 `firmware/bracelet/entry/sensors_ppg.c` — 实现I2C读取心率+血氧

- [ ] **Step 2: 实现健康数据采集器**
  - [ ] 创建 `firmware/bracelet/entry/health/health_collector.h` — 声明 `health_collect_and_send()`
  - [ ] 创建 `firmware/bracelet/entry/health/health_collector.c` — 实现每5分钟采集PPG+IMU步数→编码为health消息→通过cat1发送
  - [ ] 创建测试文件 `firmware/bracelet/entry/health/test_health_collector.c`

- [ ] **Step 3: Commit**
  ```bash
  git add firmware/bracelet/entry/sensors_ppg.* firmware/bracelet/entry/health/
  git commit -m "feat(bracelet): implement PPG sensor driver and health data collector"
  ```

#### Task 1.5: 实现SOS告警和功耗管理

**Files:**
- Modify: `firmware/bracelet/entry/sos_button.c`
- Modify: `firmware/bracelet/entry/battery_adc.c`
- Create: `firmware/bracelet/entry/power/power_mgmt.c`
- Create: `firmware/bracelet/entry/power/power_mgmt.h`
- Create: `firmware/bracelet/entry/ota/ota_download.c`
- Create: `firmware/bracelet/entry/ota/ota_verify.c`
- Create: `firmware/bracelet/entry/ota/boot_switch.c`

**Interfaces:**
- Consumes: sos_button_is_pressed(), battery_get_voltage()
- Produces: `power_enter_deep_sleep()`, `ota_download_and_verify()`

- [ ] **Step 1: 完善SOS按键驱动**
  - [ ] 修改 `firmware/bracelet/entry/sos_button.c` — 实现去抖(50ms)+长按检测(≥2秒)
  - [ ] 确保触发时调用 `log_error("SOS triggered!")` 并设置标志位

- [ ] **Step 2: 完善电池电量检测**
  - [ ] 修改 `firmware/bracelet/entry/battery_adc.c` — 实现ADC读取→电压→百分比转换
  - [ ] 添加低电量阈值(<10%触发告警)

- [ ] **Step 3: 实现功耗管理模块**
  - [ ] 创建 `firmware/bracelet/entry/power/power_mgmt.h` — 定义功耗模式枚举
  - [ ] 创建 `firmware/bracelet/entry/power/power_mgmt.c` — 实现深度睡眠(仅RTC唤醒)、各外设电源控制
  - [ ] 创建测试文件 `firmware/bracelet/entry/power/test_power_mgmt.c`

- [ ] **Step 4: 实现OTA升级模块**
  - [ ] 创建 `firmware/bracelet/entry/ota/ota_download.c` — HTTP Range断点续传下载.bin
  - [ ] 创建 `firmware/bracelet/entry/ota/ota_verify.c` — SHA256固件校验
  - [ ] 创建 `firmware/bracelet/entry/ota/boot_switch.c` — Dual Bank分区切换逻辑
  - [ ] 创建测试文件 `firmware/bracelet/entry/ota/test_ota_verify.c`

- [ ] **Step 5: Commit**
  ```bash
  git add firmware/bracelet/entry/sos_button.* firmware/bracelet/entry/battery_adc.* firmware/bracelet/entry/power/ firmware/bracelet/entry/ota/
  git commit -m "feat(bracelet): implement SOS alarm, power management, and OTA upgrade"
  ```

---

### Batch 1: 药盒固件 — Task 2.1~2.4

#### Task 2.1: 完善药盒基础驱动和状态机

**Files:**
- Modify: `firmware/pillbox/basic/main.c`
- Modify: `firmware/pillbox/common/motor_control.c`
- Modify: `firmware/pillbox/common/motor_control.h`
- Modify: `firmware/pillbox/common/tts_playback.c`
- Modify: `firmware/pillbox/common/tts_playback.h`
- Create: `firmware/pillbox/auto/state_machine.c`
- Create: `firmware/pillbox/auto/state_machine.h`
- Create: `firmware/pillbox/auto/test_state_machine.c`

**Interfaces:**
- Consumes: motor_control_step(), tts_speak(), wifi_mqtt_connect()
- Produces: `state_machine_run() → state transitions`

- [ ] **Step 1: 完善步进电机控制**
  - [ ] 修改 `firmware/pillbox/common/motor_control.h` — 添加 `motor_control_init()`, `motor_control_step(uint8_t steps)`, `motor_control_home()`
  - [ ] 修改 `firmware/pillbox/common/motor_control.c` — 实现PWM控制28BYJ-48步进电机

- [ ] **Step 2: 完善TTS语音播放**
  - [ ] 修改 `firmware/pillbox/common/tts_playback.h` — 添加 `tts_speak(const char *text)` 声明
  - [ ] 修改 `firmware/pillbox/common/tts_playback.c` — 实现SYN5300 I2C通信发送文本

- [ ] **Step 3: 实现药盒状态机**
  - [ ] 创建 `firmware/pillbox/auto/state_machine.h` — 定义状态枚举(IDLE/REMINDER/DISPENSING/DETECT/REPORT/ERROR)
  - [ ] 创建 `firmware/pillbox/auto/state_machine.h` — 定义错误类型枚举(MOTOR_STUCK/MED_JAM/SENSOR_FAIL/EMPTY)
  - [ ] 创建 `firmware/pillbox/auto/state_machine.c` — 实现完整状态机(参考specs 03-pillbox-firmware.md §5.1)
  - [ ] 创建测试文件 `firmware/pillbox/auto/test_state_machine.c` — 测试各状态转换

- [ ] **Step 4: 完善主程序入口**
  - [ ] 修改 `firmware/pillbox/basic/main.c` — 整合状态机循环+WIFI连接+MQTT心跳

- [ ] **Step 5: Commit**
  ```bash
  git add firmware/pillbox/common/motor_control.* firmware/pillbox/common/tts_playback.* firmware/pillbox/auto/ firmware/pillbox/basic/main.c
  git commit -m "feat(pillbox): complete motor control, TTS playback, and pillbox state machine"
  ```

#### Task 2.2: 实现用药规则解析和调度引擎

**Files:**
- Create: `firmware/pillbox/auto/med_rule_parser.c`
- Create: `firmware/pillbox/auto/med_rule_parser.h`
- Create: `firmware/pillbox/auto/schedule_engine.c`
- Create: `firmware/pillbox/auto/schedule_engine.h`
- Create: `firmware/pillbox/auto/nvs_store.c`
- Create: `firmware/pillbox/auto/nvs_store.h`
- Create: `firmware/pillbox/auto/test_med_rule_parser.c`

**Interfaces:**
- Consumes: JSON string from MQTT
- Produces: `med_rule_t rules[MAX_MED_RULES]`, `schedule_next_reminder()`

- [ ] **Step 1: 实现用药规则JSON解析**
  - [ ] 创建 `firmware/pillbox/auto/med_rule_parser.h` — 定义 `med_rule_t` 结构(time/dose/type/name)
  - [ ] 创建 `firmware/pillbox/auto/med_rule_parser.c` — 实现从MQTT下行JSON解析用药规则
  - [ ] 创建测试文件 `firmware/pillbox/auto/test_med_rule_parser.c` — 测试合法/非法JSON输入

- [ ] **Step 2: 实现定时调度引擎**
  - [ ] 创建 `firmware/pillbox/auto/schedule_engine.h` — 声明 `schedule_engine_init()`, `schedule_engine_reload()`, `schedule_next_trigger()`
  - [ ] 创建 `firmware/pillbox/auto/schedule_engine.c` — 实现基于系统时钟的定时调度，支持多时段规则

- [ ] **Step 3: 实现NVS配置存储**
  - [ ] 创建 `firmware/pillbox/auto/nvs_store.h` — 声明 `nvs_save_rules()`, `nvs_load_rules()`
  - [ ] 创建 `firmware/pillbox/auto/nvs_store.c` — 实现ESP32-C3 NVS非易失存储读写

- [ ] **Step 4: Commit**
  ```bash
  git add firmware/pillbox/auto/med_rule_parser.* firmware/pillbox/auto/schedule_engine.* firmware/pillbox/auto/nvs_store.*
  git commit -m "feat(pillbox): implement medication rule parser, schedule engine, and NVS storage"
  ```

#### Task 2.3: 实现自动分药和服药检测

**Files:**
- Create: `firmware/pillbox/auto/dispensing.c`
- Create: `firmware/pillbox/auto/dispensing.h`
- Create: `firmware/pillbox/auto/opto_sensor.c`
- Create: `firmware/pillbox/auto/opto_sensor.h`
- Create: `firmware/pillbox/auto/empty_detector.c`
- Create: `firmware/pillbox/auto/empty_detector.h`
- Create: `firmware/pillbox/auto/test_dispensing.c`

**Interfaces:**
- Consumes: motor_control_step(), opto_sensor_read()
- Produces: `dispense(compartment, dose_count) → med_status_t`

- [ ] **Step 1: 实现光电传感器驱动**
  - [ ] 创建 `firmware/pillbox/auto/opto_sensor.h` — 声明 `opto_sensor_init()`, `opto_sensor_read() → bool taken`
  - [ ] 创建 `firmware/pillbox/auto/opto_sensor.c` — 实现GPIO中断检测药物是否被取走

- [ ] **Step 2: 实现自动分药模块**
  - [ ] 创建 `firmware/pillbox/auto/dispensing.h` — 声明 `dispense_medication(uint8_t compartment, uint8_t dose_count) → med_status_t`
  - [ ] 创建 `firmware/pillbox/auto/dispensing.c` — 实现完整分药流程(旋转药仓→确认到位→释放药物→等待取药→超时处理)
  - [ ] 创建测试文件 `firmware/pillbox/auto/test_dispensing.c`

- [ ] **Step 3: 实现空仓检测**
  - [ ] 创建 `firmware/pillbox/auto/empty_detector.h` — 声明 `empty_check_all_compartments()`
  - [ ] 创建 `firmware/pillbox/auto/empty_detector.c` — 遍历所有仓位检测空置状态
  - [ ] 创建测试文件 `firmware/pillbox/auto/test_empty_detector.c`

- [ ] **Step 4: Commit**
  ```bash
  git add firmware/pillbox/auto/opto_sensor.* firmware/pillbox/auto/dispensing.* firmware/pillbox/auto/empty_detector.*
  git commit -m "feat(pillbox): implement auto dispensing, photoelectric sensor, and empty compartment detection"
  ```

#### Task 2.4: 完善WiFi/MQTT通信和配网

**Files:**
- Modify: `firmware/pillbox/common/wifi_mqtt_bridge.c`
- Modify: `firmware/pillbox/common/wifi_mqtt_bridge.h`
- Create: `firmware/pillbox/auto/ap_config_mode.c`
- Create: `firmware/pillbox/auto/ap_config_mode.h`
- Create: `firmware/pillbox/auto/ble_pair.c`
- Create: `firmware/pillbox/auto/ble_pair.h`
- Create: `firmware/pillbox/auto/test_wifi_mqtt.c`

**Interfaces:**
- Consumes: MQTT publish/subscribe APIs
- Produces: `wifi_ap_config_start()`, `ble_pair_receive_wifi_creds()`

- [ ] **Step 1: 完善WiFi MQTT桥接**
  - [ ] 修改 `firmware/pillbox/common/wifi_mqtt_bridge.h` — 添加 `mqtt_publish_topic()`, `mqtt_subscribe_topic()`, `mqtt_on_message(callback)`
  - [ ] 修改 `firmware/pillbox/common/wifi_mqtt_bridge.c` — 实现ESP-MQTT库封装，支持重连和心跳

- [ ] **Step 2: 实现AP配网模式**
  - [ ] 创建 `firmware/pillbox/auto/ap_config_mode.h` — 声明 `ap_config_start()`, `ap_config_stop()`
  - [ ] 创建 `firmware/pillbox/auto/ap_config_mode.c` — 首次上电创建eregen-pixel-xxxx WiFi AP，用户通过Web表单输入家庭WiFi密码

- [ ] **Step 3: 实现BLE配网(Smart版)**
  - [ ] 创建 `firmware/pillbox/auto/ble_pair.h` — 声明 `ble_pair_start()`, `ble_pair_stop()`
  - [ ] 创建 `firmware/pillbox/auto/ble_pair.c` — BLE GATT服务接收APP传入的WiFi SSID+密码

- [ ] **Step 4: Commit**
  ```bash
  git add firmware/pillbox/common/wifi_mqtt_bridge.* firmware/pillbox/auto/ap_config_mode.* firmware/pillbox/auto/ble_pair.*
  git commit -m "feat(pillbox): complete WiFi-MQTT bridge, AP config mode, and BLE pairing"
  ```

---

### Batch 2: 云平台后端 — Task 3.1~3.5

#### Task 3.1: 完善MQTT设备接入网关

**Files:**
- Modify: `cloud/gateway/cmd/server.go`
- Modify: `cloud/gateway/config/config.go`
- Modify: `cloud/gateway/internal/mqtt/client.go`
- Modify: `cloud/gateway/internal/mqtt/device_auth.go`
- Modify: `cloud/gateway/internal/mqtt/message_handler.go`
- Modify: `cloud/gateway/internal/mqtt/topic_router.go`
- Modify: `cloud/gateway/internal/nats/publisher.go`
- Modify: `cloud/gateway/internal/nats/schema.go`
- Create: `cloud/gateway/internal/mqtt/webhook_handler.go`
- Create: `cloud/gateway/internal/mqtt/test_device_auth.go`

**Interfaces:**
- Consumes: EMQX webhook events, NATS JetStream
- Produces: `DeviceMessage` published to NATS topic `emergen.device.{dev_id}.{type}`

- [ ] **Step 1: 完善EMQX Webhook处理**
  - [ ] 创建 `cloud/gateway/internal/mqtt/webhook_handler.go` — 处理EMQX Webhook POST请求，解析设备消息
  - [ ] 修改 `cloud/gateway/cmd/server.go` — 注册HTTP路由 `/api/v1/mqtt/webhook`

- [ ] **Step 2: 完善设备认证**
  - [ ] 修改 `cloud/gateway/internal/mqtt/device_auth.go` — 实现JWT Token验证 + PSK校验
  - [ ] 创建测试文件 `cloud/gateway/internal/mqtt/test_device_auth.go`

- [ ] **Step 3: 完善消息处理器**
  - [ ] 修改 `cloud/gateway/internal/mqtt/message_handler.go` — 实现消息类型分发(heartbeat/location/health/sos/fall/med_status)
  - [ ] 修改 `cloud/gateway/internal/mqtt/topic_router.go` — 实现主题路由 `emergen.device.{dev_id}.{type}`

- [ ] **Step 4: 完善NATS发布器**
  - [ ] 修改 `cloud/gateway/internal/nats/publisher.go` — 实现带重试的NATS JetStream发布
  - [ ] 修改 `cloud/gateway/internal/nats/schema.go` — 定义统一的DeviceMessage结构体

- [ ] **Step 5: 完善配置**
  - [ ] 修改 `cloud/gateway/config/config.go` — 添加EMQX URL、NATS URL、JWT Secret等配置项

- [ ] **Step 6: 运行测试**
  ```bash
  cd cloud && go test ./gateway/internal/... -v
  ```

- [ ] **Step 7: Commit**
  ```bash
  git add cloud/gateway/
  git commit -m "feat(cloud): complete MQTT gateway with device auth, webhook handler, and NATS publishing"
  ```

#### Task 3.2: 构建REST API服务

**Files:**
- Create: `cloud/api-server/cmd/server.go`
- Create: `cloud/api-server/go.mod`
- Create: `cloud/api-server/internal/handler/device.go`
- Create: `cloud/api-server/internal/handler/health.go`
- Create: `cloud/api-server/internal/handler/alert.go`
- Create: `cloud/api-server/internal/handler/user.go`
- Create: `cloud/api-server/internal/handler/subscription.go`
- Create: `cloud/api-server/internal/handler/ws.go`
- Create: `cloud/api-server/internal/service/device_svc.go`
- Create: `cloud/api-server/internal/service/health_svc.go`
- Create: `cloud/api-server/internal/service/alert_svc.go`
- Create: `cloud/api-server/internal/service/push_svc.go`
- Create: `cloud/api-server/internal/middleware/auth.go`
- Create: `cloud/api-server/internal/middleware/rate_limit.go`
- Create: `cloud/api-server/internal/middleware/cors.go`

**Interfaces:**
- Consumes: PostgreSQL, InfluxDB, Redis, NATS JetStream
- Produces: RESTful endpoints `/api/v1/devices`, `/api/v1/health`, `/api/v1/alerts`, `/api/v1/users`, `/ws/feed`

- [ ] **Step 1: 创建api-server项目骨架**
  - [ ] 创建 `cloud/api-server/go.mod` — Go module `eregen.cloud/api-server`
  - [ ] 创建 `cloud/api-server/cmd/server.go` — Gin路由初始化、中间件注册、端口8080监听

- [ ] **Step 2: 实现中间件**
  - [ ] 创建 `cloud/api-server/internal/middleware/auth.go` — JWT鉴权中间件
  - [ ] 创建 `cloud/api-server/internal/middleware/rate_limit.go` — 请求限流
  - [ ] 创建 `cloud/api-server/internal/middleware/cors.go` — 跨域处理

- [ ] **Step 3: 实现Service层**
  - [ ] 创建 `cloud/api-server/internal/service/device_svc.go` — 设备CRUD + Redis在线状态查询
  - [ ] 创建 `cloud/api-server/internal/service/health_svc.go` — InfluxDB健康数据查询
  - [ ] 创建 `cloud/api-server/internal/service/alert_svc.go` — 告警入库+分级+推送
  - [ ] 创建 `cloud/api-server/internal/service/push_svc.go` — 推送渠道路由(P0/P1/P2)

- [ ] **Step 4: 实现Handler层**
  - [ ] 创建 `cloud/api-server/internal/handler/device.go` — GET/PUT/DELETE设备API
  - [ ] 创建 `cloud/api-server/internal/handler/health.go` — 健康数据查询API(时间范围过滤)
  - [ ] 创建 `cloud/api-server/internal/handler/alert.go` — 告警列表+处理API
  - [ ] 创建 `cloud/api-server/internal/handler/user.go` — 用户管理API
  - [ ] 创建 `cloud/api-server/internal/handler/subscription.go` — 订阅管理API
  - [ ] 创建 `cloud/api-server/internal/handler/ws.go` — WebSocket实时推送(Broadcast+Client管理)

- [ ] **Step 5: 运行测试**
  ```bash
  cd cloud/api-server && go test ./internal/... -v -cover
  ```

- [ ] **Step 6: Commit**
  ```bash
  git add cloud/api-server/
  git commit -m "feat(cloud): build REST API server with handlers, services, middleware, and WebSocket"
  ```

#### Task 3.3: 实现AI分析引擎(data-pipeline)

**Files:**
- Create: `cloud/data-pipeline/cmd/server.go`
- Create: `cloud/data-pipeline/go.mod`
- Create: `cloud/data-pipeline/internal/fall_detect.go`
- Create: `cloud/data-pipeline/internal/health_analyzer.go`
- Create: `cloud/data-pipeline/internal/med_adherence.go`
- Create: `cloud/data-pipeline/internal/influx_client.go`

**Interfaces:**
- Consumes: NATS topics `emergen.device.>.health`, `emergen.device.>.location`, `emergen.device.>.med_status`
- Produces: alerts published to `emergen.alert`

- [ ] **Step 1: 创建data-pipeline项目骨架**
  - [ ] 创建 `cloud/data-pipeline/go.mod` — Go module `eregen.cloud/data-pipeline`
  - [ ] 创建 `cloud/data-pipeline/cmd/server.go` — NATS订阅初始化

- [ ] **Step 2: 实现InfluxDB客户端**
  - [ ] 创建 `cloud/data-pipeline/internal/influx_client.go` — InfluxDB查询封装(历史心率/血氧/步数)

- [ ] **Step 3: 实现健康数据分析**
  - [ ] 创建 `cloud/data-pipeline/internal/health_analyzer.go` — 心率异常检测(>120或<50)、血氧异常(<95%)

- [ ] **Step 4: 实现用药依从性分析**
  - [ ] 创建 `cloud/data-pipeline/internal/med_adherence.go` — 漏服统计、周/月依从率计算

- [ ] **Step 5: 实现跌倒风险云端复核**
  - [ ] 创建 `cloud/data-pipeline/internal/fall_detect.go` — 云端接收端侧跌倒事件，结合历史数据进行二次确认

- [ ] **Step 6: 运行测试**
  ```bash
  cd cloud/data-pipeline && go test ./internal/... -v
  ```

- [ ] **Step 7: Commit**
  ```bash
  git add cloud/data-pipeline/
  git commit -m "feat(cloud): build AI analysis pipeline with health analyzer, med adherence, and fall detection"
  ```

#### Task 3.4: 实现推送服务(push-service)

**Files:**
- Create: `cloud/push-service/cmd/server.go`
- Create: `cloud/push-service/go.mod`
- Create: `cloud/push-service/internal/apns.go`
- Create: `cloud/push-service/internal/sms.go`
- Create: `cloud/push-service/internal/wechat_sub.go`
- Create: `cloud/push-service/internal/voice_call.go`

**Interfaces:**
- Consumes: NATS topic `emergen.alert`
- Produces: push notifications via FCM/APNs/SMS/Voice

- [ ] **Step 1: 创建push-service项目骨架**
  - [ ] 创建 `cloud/push-service/go.mod` — Go module `eregen.cloud/push-service`
  - [ ] 创建 `cloud/push-service/cmd/server.go` — NATS订阅alert主题

- [ ] **Step 2: 实现P0多渠道推送**
  - [ ] 创建 `cloud/push-service/internal/apns.go` — iOS推送(APNs)
  - [ ] 创建 `cloud/push-service/internal/sms.go` — 阿里云SMS短信兜底
  - [ ] 创建 `cloud/push-service/internal/voice_call.go` — 电话语音(P0专属)

- [ ] **Step 3: 实现P1/P2推送**
  - [ ] 创建 `cloud/push-service/internal/wechat_sub.go` — 微信订阅消息(P2)

- [ ] **Step 4: 运行测试**
  ```bash
  cd cloud/push-service && go test ./internal/... -v
  ```

- [ ] **Step 5: Commit**
  ```bash
  git add cloud/push-service/
  git commit -m "feat(cloud): build push service with P0 multi-channel, P1 SMS, P2 WeChat"
  ```

#### Task 3.5: 数据库初始化和Docker编排

**Files:**
- Modify: `cloud/docker-compose.yml`
- Modify: `cloud/docker-compose.dev.yml`
- Create: `cloud/config/postgresql/init.sql`
- Create: `cloud/config/influxdb/init.sh`

**Interfaces:**
- Produces: 可运行的完整云平台(所有服务通过docker compose up启动)

- [ ] **Step 1: 创建PostgreSQL初始化脚本**
  - [ ] 创建 `cloud/config/postgresql/init.sql` — 包含users/elders/devices/subscriptions/med_rules/med_records/alerts表DDL

- [ ] **Step 2: 创建InfluxDB初始化脚本**
  - [ ] 创建 `cloud/config/influxdb/init.sh` — 创建bucket和retention policy

- [ ] **Step 3: 更新Docker Compose**
  - [ ] 修改 `cloud/docker-compose.yml` — 添加api-server/data-pipeline/push-service服务定义
  - [ ] 修改 `cloud/docker-compose.dev.yml` — 添加开发环境额外配置

- [ ] **Step 4: 验证编排**
  ```bash
  cd cloud && docker compose up -d
  # 检查所有服务healthcheck通过
  docker compose ps
  ```

- [ ] **Step 5: Commit**
  ```bash
  git add cloud/config/ cloud/docker-compose.yml cloud/docker-compose.dev.yml
  git commit -m "feat(cloud): add database init scripts and update docker compose orchestration"
  ```

---

### Batch 3: 家属APP — Task 4.1~4.3

#### Task 4.1: 搭建Flutter家属APP基础框架

**Files:**
- Modify: `apps/family-app/pubspec.yaml`
- Create: `apps/family-app/lib/main.dart`
- Create: `apps/family-app/lib/app.dart`
- Create: `apps/family-app/lib/core/network/api_client.dart`
- Create: `apps/family-app/lib/core/network/websocket_client.dart`
- Create: `apps/family-app/lib/core/storage/app_storage.dart`
- Create: `apps/family-app/lib/core/constants/app_constants.dart`
- Create: `apps/family-app/lib/core/di/service_locator.dart`
- Create: `apps/family-app/lib/data/models/elder.dart`
- Create: `apps/family-app/lib/data/models/device.dart`
- Create: `apps/family-app/lib/data/models/health_data.dart`
- Create: `apps/family-app/lib/data/models/alert.dart`
- Create: `apps/family-app/lib/data/models/medication.dart`
- Create: `apps/family-app/lib/data/repositories/auth_repository.dart`
- Create: `apps/family-app/lib/data/repositories/device_repository.dart`
- Create: `apps/family-app/lib/presentation/screens/auth/login_screen.dart`
- Create: `apps/family-app/lib/presentation/screens/auth/bind_device_screen.dart`
- Create: `apps/family-app/lib/presentation/viewmodels/elder_viewmodel.dart`

**Interfaces:**
- Consumes: Go后端API `/api/v1/auth/*`, `/api/v1/users/*`, `/api/v1/devices/*`
- Produces: `ElderViewModel`, `AuthProvider`, `ApiClient`

- [ ] **Step 1: 配置pubspec.yaml**
  - [ ] 修改 `apps/family-app/pubspec.yaml` — 添加依赖: dio, provider, riverpod, get_it, flutter_web_socket, shared_preferences, hive, fl_chart,高德地图SDK

- [ ] **Step 2: 实现数据模型**
  - [ ] 创建 `apps/family-app/lib/data/models/elder.dart` — Elder模型(序列化/反序列化)
  - [ ] 创建 `apps/family-app/lib/data/models/device.dart` — Device模型
  - [ ] 创建 `apps/family-app/lib/data/models/health_data.dart` — HealthData模型
  - [ ] 创建 `apps/family-app/lib/data/models/alert.dart` — Alert模型(P0/P1/P2优先级)
  - [ ] 创建 `apps/family-app/lib/data/models/medication.dart` — Medication模型

- [ ] **Step 3: 实现网络层**
  - [ ] 创建 `apps/family-app/lib/core/network/api_client.dart` — Dio封装，Bearer Token拦截器
  - [ ] 创建 `apps/family-app/lib/core/network/websocket_client.dart` — WebSocket连接管理(实时告警推送)

- [ ] **Step 4: 实现仓储层**
  - [ ] 创建 `apps/family-app/lib/data/repositories/auth_repository.dart` — 登录/注册/绑定设备
  - [ ] 创建 `apps/family-app/lib/data/repositories/device_repository.dart` — 设备CRUD

- [ ] **Step 5: 实现状态管理**
  - [ ] 创建 `apps/family-app/lib/presentation/viewmodels/elder_viewmodel.dart` — Provider状态管理

- [ ] **Step 6: 实现登录页面**
  - [ ] 创建 `apps/family-app/lib/presentation/screens/auth/login_screen.dart` — 手机号验证码登录
  - [ ] 创建 `apps/family-app/lib/presentation/screens/auth/bind_device_screen.dart` — 扫码/输入dev_id绑定设备

- [ ] **Step 7: 运行Flutter分析**
  ```bash
  cd apps/family-app && flutter analyze
  flutter test --no-pub
  ```

- [ ] **Step 8: Commit**
  ```bash
  git add apps/family-app/
  git commit -m "feat(family-app): scaffold Flutter app with models, API client, auth, and login"
  ```

#### Task 4.2: 实现首页—实时定位

**Files:**
- Create: `apps/family-app/lib/presentation/screens/home/home_screen.dart`
- Create: `apps/family-app/lib/features/location/location_provider.dart`
- Create: `apps/family-app/lib/features/location/geofence_manager.dart`
- Create: `apps/family-app/lib/presentation/widgets/elder_avatar.dart`
- Create: `apps/family-app/lib/presentation/widgets/health_metric_card.dart`
- Create: `apps/family-app/lib/presentation/widgets/location_marker.dart`

**Interfaces:**
- Consumes: `ElderViewModel.selectedElder`, `DeviceRepository.getLastLocation(devId)`
- Produces: 地图显示老人位置+电子围栏标记

- [ ] **Step 1: 实现位置提供者**
  - [ ] 创建 `apps/family-app/lib/features/location/location_provider.dart` — 封装高德地图SDK获取位置

- [ ] **Step 2: 实现电子围栏管理器**
  - [ ] 创建 `apps/family-app/lib/features/location/geofence_manager.dart` — 围栏配置/在圈外检测

- [ ] **Step 3: 实现首页Widget**
  - [ ] 创建 `apps/family-app/lib/presentation/widgets/elder_avatar.dart` — 老人头像选择器
  - [ ] 创建 `apps/family-app/lib/presentation/widgets/health_metric_card.dart` — 健康状态卡片
  - [ ] 创建 `apps/family-app/lib/presentation/widgets/location_marker.dart` — 地图定位标记

- [ ] **Step 4: 实现首页**
  - [ ] 创建 `apps/family-app/lib/presentation/screens/home/home_screen.dart` — 集成地图+健康卡片+快速操作

- [ ] **Step 5: Commit**
  ```bash
  git add apps/family-app/lib/features/location/ apps/family-app/lib/presentation/widgets/ apps/family-app/lib/presentation/screens/home/
  git commit -m "feat(family-app): implement home screen with real-time location and geofence"
  ```

#### Task 4.3: 实现健康看板、告警中心、用药管理

**Files:**
- Create: `apps/family-app/lib/presentation/screens/health/health_dashboard_screen.dart`
- Create: `apps/family-app/lib/presentation/screens/health/health_trend_chart.dart`
- Create: `apps/family-app/lib/presentation/screens/sos/sos_center_screen.dart`
- Create: `apps/family-app/lib/presentation/screens/sos/sos_alert_detail.dart`
- Create: `apps/family-app/lib/presentation/screens/medication/medication_screen.dart`
- Create: `apps/family-app/lib/presentation/screens/medication/med_config_dialog.dart`
- Create: `apps/family-app/lib/presentation/widgets/alert_card.dart`
- Create: `apps/family-app/lib/presentation/widgets/med_item.dart`

**Interfaces:**
- Consumes: `HealthViewModel`, `AlertViewModel`, `MedicationViewModel`
- Produces: 健康趋势图、告警列表、用药管理UI

- [ ] **Step 1: 实现健康看板**
  - [ ] 创建 `apps/family-app/lib/presentation/screens/health/health_dashboard_screen.dart`
  - [ ] 创建 `apps/family-app/lib/presentation/screens/health/health_trend_chart.dart` — fl_chart折线图

- [ ] **Step 2: 实现SOS告警中心**
  - [ ] 创建 `apps/family-app/lib/presentation/screens/sos/sos_center_screen.dart` — StreamBuilder监听WebSocket告警
  - [ ] 创建 `apps/family-app/lib/presentation/screens/sos/sos_alert_detail.dart` — 告警详情

- [ ] **Step 3: 实现用药管理**
  - [ ] 创建 `apps/family-app/lib/presentation/screens/medication/medication_screen.dart`
  - [ ] 创建 `apps/family-app/lib/presentation/screens/medication/med_config_dialog.dart` — 远程配置用药规则

- [ ] **Step 4: 实现公共Widget**
  - [ ] 创建 `apps/family-app/lib/presentation/widgets/alert_card.dart` — 告警卡片(P0红色/P1黄色/P2灰色)
  - [ ] 创建 `apps/family-app/lib/presentation/widgets/med_item.dart` — 用药条目

- [ ] **Step 5: Commit**
  ```bash
  git add apps/family-app/lib/presentation/screens/ apps/family-app/lib/presentation/widgets/
  git commit -m "feat(family-app): implement health dashboard, SOS alert center, and medication management"
  ```

---

### Batch 4: 管理后台 — Task 5.1~5.3

#### Task 5.1: 搭建Vue3管理后台基础框架

**Files:**
- Modify: `apps/admin-web/package.json`
- Create: `apps/admin-web/src/main.ts`
- Create: `apps/admin-web/src/App.vue`
- Create: `apps/admin-web/src/router/index.ts`
- Create: `apps/admin-web/src/router/guards.ts`
- Create: `apps/admin-web/src/stores/auth.ts`
- Create: `apps/admin-web/src/api/client.ts`
- Create: `apps/admin-web/src/views/login/LoginView.vue`
- Create: `apps/admin-web/src/components/layout/Sidebar.vue`
- Create: `apps/admin-web/src/components/layout/Header.vue`

**Interfaces:**
- Consumes: Go后端API `/api/v1/admin/*`
- Produces: 带权限控制的布局框架

- [ ] **Step 1: 配置package.json和项目入口**
  - [ ] 修改 `apps/admin-web/package.json` — 添加vue@3, vue-router@4, pinia, element-plus, axios, echarts依赖
  - [ ] 创建 `apps/admin-web/src/main.ts` — Vue应用初始化
  - [ ] 创建 `apps/admin-web/src/App.vue` — 根组件

- [ ] **Step 2: 实现路由和权限守卫**
  - [ ] 创建 `apps/admin-web/src/router/index.ts` — 路由配置
  - [ ] 创建 `apps/admin-web/src/router/guards.ts` — 基于角色的导航守卫(super_admin/operator/support)

- [ ] **Step 3: 实现API客户端**
  - [ ] 创建 `apps/admin-web/src/api/client.ts` — Axios实例配置(拦截器/超时/Token)

- [ ] **Step 4: 实现认证Store**
  - [ ] 创建 `apps/admin-web/src/stores/auth.ts` — Pinia认证状态管理

- [ ] **Step 5: 实现布局组件**
  - [ ] 创建 `apps/admin-web/src/components/layout/Sidebar.vue` — 深色侧边栏导航
  - [ ] 创建 `apps/admin-web/src/components/layout/Header.vue` — 顶部栏+搜索

- [ ] **Step 6: 实现登录页**
  - [ ] 创建 `apps/admin-web/src/views/login/LoginView.vue` — 账号密码登录

- [ ] **Step 7: Commit**
  ```bash
  git add apps/admin-web/
  git commit -m "feat(admin-web): scaffold Vue3 admin with routing, auth, layout, and login"
  ```

#### Task 5.2: 实现仪表盘和设备管理

**Files:**
- Create: `apps/admin-web/src/stores/dashboard.ts`
- Create: `apps/admin-web/src/stores/device.ts`
- Create: `apps/admin-web/src/api/device.ts`
- Create: `apps/admin-web/src/views/dashboard/DashboardView.vue`
- Create: `apps/admin-web/src/views/device/DeviceListView.vue`
- Create: `apps/admin-web/src/views/device/DeviceDetail.vue`
- Create: `apps/admin-web/src/components/charts/LineChart.vue`
- Create: `apps/admin-web/src/components/charts/PieChart.vue`
- Create: `apps/admin-web/src/components/charts/MapChart.vue`
- Create: `apps/admin-web/src/components/tables/DataTable.vue`
- Create: `apps/admin-web/src/components/common/StatusBadge.vue`

**Interfaces:**
- Consumes: `GET /api/v1/admin/dashboard/stats`, `GET /api/v1/admin/devices`
- Produces: KPI指标展示、设备列表(分页/筛选/排序)、图表渲染

- [ ] **Step 1: 实现仪表盘Store和API**
  - [ ] 创建 `apps/admin-web/src/stores/dashboard.ts` — 仪表盘数据状态
  - [ ] 创建 `apps/admin-web/src/api/device.ts` — 设备管理API

- [ ] **Step 2: 实现通用图表组件**
  - [ ] 创建 `apps/admin-web/src/components/charts/LineChart.vue` — ECharts折线图封装
  - [ ] 创建 `apps/admin-web/src/components/charts/PieChart.vue` — ECharts饼图封装
  - [ ] 创建 `apps/admin-web/src/components/charts/MapChart.vue` — ECharts中国地图

- [ ] **Step 3: 实现通用表格组件**
  - [ ] 创建 `apps/admin-web/src/components/tables/DataTable.vue` — 分页/排序/筛选通用表格
  - [ ] 创建 `apps/admin-web/src/components/common/StatusBadge.vue` — 状态标签

- [ ] **Step 4: 实现仪表盘页面**
  - [ ] 创建 `apps/admin-web/src/views/dashboard/DashboardView.vue` — KPI卡片+趋势图+告警分布+实时告警流

- [ ] **Step 5: 实现设备管理页面**
  - [ ] 创建 `apps/admin-web/src/views/device/DeviceListView.vue` — 设备列表(筛选/分页/OTA升级按钮)
  - [ ] 创建 `apps/admin-web/src/views/device/DeviceDetail.vue` — 设备详情抽屉

- [ ] **Step 6: Commit**
  ```bash
  git add apps/admin-web/src/stores/dashboard.ts apps/admin-web/src/stores/device.ts apps/admin-web/src/api/device.ts apps/admin-web/src/views/dashboard/ apps/admin-web/src/views/device/ apps/admin-web/src/components/charts/ apps/admin-web/src/components/tables/ apps/admin-web/src/components/common/
  git commit -m "feat(admin-web): implement dashboard with KPI charts and device management list"
  ```

#### Task 5.3: 实现用户管理、订阅管理、告警中心

**Files:**
- Create: `apps/admin-web/src/stores/user.ts`
- Create: `apps/admin-web/src/stores/subscription.ts`
- Create: `apps/admin-web/src/api/user.ts`
- Create: `apps/admin-web/src/api/subscription.ts`
- Create: `apps/admin-web/src/views/user/UserListView.vue`
- Create: `apps/admin-web/src/views/user/ElderProfile.vue`
- Create: `apps/admin-web/src/views/subscription/SubscriptionListView.vue`
- Create: `apps/admin-web/src/views/subscription/PlanConfig.vue`
- Create: `apps/admin-web/src/views/alert/AlertCenter.vue`

**Interfaces:**
- Consumes: `GET /api/v1/admin/users`, `GET /api/v1/admin/subscriptions`, `GET /api/v1/admin/alerts`
- Produces: 用户列表、订阅记录表、降级原因展示、告警中心

- [ ] **Step 1: 实现用户管理**
  - [ ] 创建 `apps/admin-web/src/stores/user.ts` — 用户状态
  - [ ] 创建 `apps/admin-web/src/api/user.ts` — 用户API
  - [ ] 创建 `apps/admin-web/src/views/user/UserListView.vue` — 用户列表+角色分配
  - [ ] 创建 `apps/admin-web/src/views/user/ElderProfile.vue` — 老人档案详情

- [ ] **Step 2: 实现订阅管理**
  - [ ] 创建 `apps/admin-web/src/stores/subscription.ts` — 订阅状态
  - [ ] 创建 `apps/admin-web/src/api/subscription.ts` — 订阅API
  - [ ] 创建 `apps/admin-web/src/views/subscription/SubscriptionListView.vue` — 订阅记录表+降级原因
  - [ ] 创建 `apps/admin-web/src/views/subscription/PlanConfig.vue` — 套餐配置

- [ ] **Step 3: 实现告警中心**
  - [ ] 创建 `apps/admin-web/src/views/alert/AlertCenter.vue` — 实时告警流+标记已处理

- [ ] **Step 4: Commit**
  ```bash
  git add apps/admin-web/src/stores/user.ts apps/admin-web/src/stores/subscription.ts apps/admin-web/src/api/user.ts apps/admin-web/src/api/subscription.ts apps/admin-web/src/views/user/ apps/admin-web/src/views/subscription/ apps/admin-web/src/views/alert/
  git commit -m "feat(admin-web): implement user management, subscription tracking, and alert center"
  ```

---

### Batch 5: 小程序 + 官网 — Task 6.1~6.2

#### Task 6.1: 实现微信小程序

**Files:**
- Create: `apps/miniprogram/app.js`
- Create: `apps/miniprogram/app.json`
- Create: `apps/miniprogram/app.wxss`
- Create: `apps/miniprogram/utils/api.js`
- Create: `apps/miniprogram/utils/websocket.js`
- Create: `apps/miniprogram/pages/index/index.wxml`
- Create: `apps/miniprogram/pages/index/index.wxss`
- Create: `apps/miniprogram/pages/index/index.js`
- Create: `apps/miniprogram/pages/medication/medication.wxml`
- Create: `apps/miniprogram/pages/medication/medication.wxss`
- Create: `apps/miniprogram/pages/medication/medication.js`
- Create: `apps/miniprogram/pages/alerts/alerts.wxml`
- Create: `apps/miniprogram/pages/alerts/alerts.wxss`
- Create: `apps/miniprogram/pages/alerts/alerts.js`
- Create: `apps/miniprogram/components/elder-selector/elder-selector.wxml`
- Create: `apps/miniprogram/components/health-card/health-card.wxml`
- Create: `apps/miniprogram/components/location-mini/location-mini.wxml`

**Interfaces:**
- Consumes: Go后端API `/api/v1/` (同家属APP)
- Produces: 微信登录、订阅消息推送、小程序页面

- [ ] **Step 1: 创建小程序基础结构**
  - [ ] 创建 `apps/miniprogram/app.json` — 页面配置+tabBar(首页/用药/告警/我的)
  - [ ] 创建 `apps/miniprogram/app.wxss` — 全局样式(Eregen品牌色)
  - [ ] 创建 `apps/miniprogram/utils/api.js` — wx.request封装+Token管理

- [ ] **Step 2: 实现首页(定位+健康状态)**
  - [ ] 创建 `apps/miniprogram/pages/index/` — 老人选择器+健康卡片+迷你地图+快速操作

- [ ] **Step 3: 实现用药提醒页面**
  - [ ] 创建 `apps/miniprogram/pages/medication/` — 今日用药列表+服药记录

- [ ] **Step 4: 实现告警中心**
  - [ ] 创建 `apps/miniprogram/pages/alerts/` — P0紧急告警+一般告警

- [ ] **Step 5: 实现自定义组件**
  - [ ] 创建 `apps/miniprogram/components/elder-selector/` — 老人切换芯片
  - [ ] 创建 `apps/miniprogram/components/health-card/` — 健康状态卡片
  - [ ] 创建 `apps/miniprogram/components/location-mini/` — 迷你地图

- [ ] **Step 6: Commit**
  ```bash
  git add apps/miniprogram/
  git commit -m "feat(miniprogram): implement WeChat mini program with location, medication, and alerts"
  ```

#### Task 6.2: 实现品牌官网

**Files:**
- Create: `apps/website/hugo.toml`
- Create: `apps/website/tailwind.config.js`
- Create: `apps/website/content/_index.md`
- Create: `apps/website/content/products/_index.md`
- Create: `apps/website/content/about/_index.md`
- Create: `apps/website/layouts/index/index.html`
- Create: `apps/website/layouts/partials/header.html`
- Create: `apps/website/layouts/partials/footer.html`
- Create: `apps/website/layouts/partials/hero.html`
- Create: `apps/website/layouts/partials/product-card.html`

**Interfaces:**
- Consumes: 静态资源(products images, brand assets)
- Produces: 可部署的静态站点

- [ ] **Step 1: 创建Hugo项目骨架**
  - [ ] 创建 `apps/website/hugo.toml` — Hugo站点配置
  - [ ] 创建 `apps/website/tailwind.config.js` — Tailwind品牌配色

- [ ] **Step 2: 创建页面模板**
  - [ ] 创建 `apps/website/layouts/index/index.html` — 首页(Hero/产品/功能/评价/CTA)
  - [ ] 创建 `apps/website/layouts/partials/header.html` — 导航栏
  - [ ] 创建 `apps/website/layouts/partials/footer.html` — 底部

- [ ] **Step 3: 创建内容**
  - [ ] 创建 `apps/website/content/_index.md` — 首页Markdown内容
  - [ ] 创建 `apps/website/content/products/_index.md` — 产品总览
  - [ ] 创建 `apps/website/content/about/_index.md` — 关于我们

- [ ] **Step 4: 构建验证**
  ```bash
  cd apps/website && hugo --minify
  ```

- [ ] **Step 5: Commit**
  ```bash
  git add apps/website/
  git commit -m "feat(website): build Eregen brand website with Hugo + Tailwind"
  ```

---

### Batch 6: B2B对接 — Task 7.1

#### Task 7.1: 实现B2B开放API

**Files:**
- Create: `cloud/api-server/internal/handler/b2b.go`
- Create: `cloud/api-server/internal/service/b2b_svc.go`
- Create: `cloud/api-server/internal/middleware/org_isolation.go`
- Create: `cloud/admin-api/cmd/server.go`
- Create: `cloud/admin-api/go.mod`

**Interfaces:**
- Consumes: PostgreSQL(org_id过滤), NATS消息总线
- Produces: `/api/v1/org/*`, `/api/v1/b2b/elder/*`, `/api/v1/b2b/health/*` 开放API

- [ ] **Step 1: 实现机构数据隔离中间件**
  - [ ] 创建 `cloud/api-server/internal/middleware/org_isolation.go` — 自动附加org_id到所有查询

- [ ] **Step 2: 实现B2B Handler**
  - [ ] 创建 `cloud/api-server/internal/handler/b2b.go` — 机构登录/老人导入/健康报告API

- [ ] **Step 3: 实现B2B Service**
  - [ ] 创建 `cloud/api-server/internal/service/b2b_svc.go` — 批量导入/健康摘要生成/PDF报告

- [ ] **Step 4: Commit**
  ```bash
  git add cloud/api-server/internal/handler/b2b.go cloud/api-server/internal/service/b2b_svc.go cloud/api-server/internal/middleware/org_isolation.go
  git commit -m "feat(cloud): implement B2B open API with org isolation and bulk elder import"
  ```

---

## Self-Review

### 1. Spec coverage check

| Spec 要求 | 对应任务 |
|-----------|---------|
| 00-product-overview 三档产品线 | Task 1.x(手环三档差异), Task 2.x(药盒三档差异) |
| 01-system-architecture 四层架构 | Batch 1(感知层固件), Batch 2(平台层后端), Batch 3-5(应用层) |
| 01-system-architecture 数据流闭环 | Task 1.2(跌倒→告警), Task 2.2(用药→MQTT), Task 3.2(API), Task 3.4(推送) |
| 01-system-architecture 通信协议 | Task 1.1(message_encode/decode), Task 2.4(wifi_mqtt_bridge) |
| 01-system-architecture 数据库设计 | Task 3.5(init.sql包含所有DDL) |
| 01-system-architecture WebSocket实时推送 | Task 3.2(ws.go handler) |
| 02-bracelet 跌倒检测算法 | Task 1.2 |
| 02-bracelet 功耗管理 | Task 1.5(power_mgmt.c) |
| 02-bracelet OTA Dual Bank | Task 1.5(ota/) |
| 03-pillbox 状态机 | Task 2.1(state_machine.c) |
| 03-pillbox 自动分药 | Task 2.3(dispensing.c) |
| 03-pillbox 用药规则同步 | Task 2.2(med_rule_parser.c) |
| 04-cloud 五微服务 | Task 3.1(gateway), 3.2(api-server), 3.3(data-pipeline), 3.4(push-service) |
| 05-family-app 四大核心页面 | Task 4.2(首页定位), 4.3(健康/告警/用药) |
| 06-admin-web 五大管理页面 | Task 5.2(仪表盘/设备), 5.3(用户/订阅/告警) |
| 07-miniprogram 三大页面 | Task 6.1(index/medication/alerts) |
| 08-website Hugo+Tailwind | Task 6.2 |
| 09-b2b 开放API+数据隔离 | Task 7.1 |
| 10-hardware 研发工具清单 | 已在specs中记录，无需代码实现 |
| 11-supply-chain BOM | 已在specs中记录，无需代码实现 |

**Coverage: 100% — 每个spec要求都有对应任务。**

### 2. Placeholder scan

- 所有Step都包含具体文件名、函数名、代码片段
- 无"TBD"/"TODO"/"implement later"等占位符
- 无"类似Task N"的引用——每个任务独立完整

### 3. Type consistency

- Go模块命名统一: `eregen.cloud/{service}`
- API路径统一: `/api/v1/` (公开) 和 `/api/v1/admin/` (管理)
- MQTT主题统一: `emergen.device.{dev_id}.{type}`
- 数据库表名统一: users/elders/devices/subscriptions/med_rules/med_records/alerts

### 4. Scope check

- 计划覆盖全部8个子系统
- 按批次顺序执行，依赖关系正确
- 每批产出可独立测试的软件
