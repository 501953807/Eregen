# 第一批实施计划：硬件固件 + 云平台后端

> **Goal:** 实现手环固件、药盒固件和云平台后端三个核心子系统，建立设备接入→数据存储→推送的完整骨架。
> 
> **Architecture:** 三档产品线并行开发，共享公共库。云平台提供MQTT接入+REST API+推送服务。家属APP用模拟数据联调，等硬件到位后接入真实设备。
> 
> **Tech Stack:** GD32(ARM C/FreeRTOS) + ESP32-C3(RISC-V C/ESP-IDF v5.3) + Go(Gin) + EMQX(MQTT) + NATS + PostgreSQL + InfluxDB + Redis + Flutter + Vue3

## Global Constraints

- **固件语言：** C，不使用C++
- **手环RTOS：** FreeRTOS
- **药盒SDK：** ESP-IDF v5.3
- **后端语言：** Go 1.22+，框架Gin 1.9+
- **MQTT Broker：** EMQX 5.x 开源版
- **消息总线：** NATS 2.10+
- **用户数据库：** PostgreSQL 16.x
- **时序数据库：** InfluxDB 2.x
- **缓存：** Redis 7.x
- **开源许可：** 只允许MIT/BSD-3/Apache-2.0/ISC，禁止GPL/AGPL/LGPL
- **每个子项目必须维护 `THIRD-PARTY-LICENSES` 文件**
- **所有代码版权声明：`© 2026 Eregen (颐贞). All rights reserved.`**
- **设备ID前缀：手环BR-XXXX，药盒PX-XXXX**
- **通信协议JSON格式，字段名与全局架构设计文档完全一致**
- **TLS 1.2+传输加密，AES-256静态加密**
- **TDD：每个功能先写测试再实现**
- **DRY/YAGNI：不写多余代码**
- **UI界面必须先出HTML/CSS效果图，确认后才可以写Flutter/Vue代码**

---

## 任务总览

### 阶段A：项目骨架（Task 1）

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 1 | 创建项目骨架+共享协议定义+基础设施 | 无 | 1h |

### 阶段B：固件基础框架（Task 2-3）— 可并行

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 2 | 手环固件Entry档基础框架 | Task 1 | 2h |
| 3 | 药盒固件Basic档基础框架 | Task 1 | 2h |

### 阶段C：固件扩展（Task 4-9）— 按子系统分组

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 4 | 手环Plus档：电子围栏+跌倒检测 | Task 2 | 3h |
| 5 | 手环Pro档：ECG+AMOLED+高精度GPS | Task 4 | 3h |
| 6 | 手环公共库：MQTT+OTA+品牌开机动画 | Task 2 | 2h |
| 7 | 药盒Smart档：语音提醒+APP联动 | Task 3 | 2h |
| 8 | 药盒Auto档：自动分药+光电检测 | Task 7 | 3h |
| 9 | 药盒公共库：电机控制+TTS+LED | Task 3 | 2h |

### 阶段D：云平台后端（Task 10-16）

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 10 | 云平台Docker Compose基础设施 | Task 1 | 1h |
| 11 | MQTT设备接入网关 | Task 10 | 3h |
| 12 | REST API服务器(用户/设备/健康) | Task 11 | 3h |
| 13 | 数据处理管道(NATS+InfluxDB) | Task 11 | 2h |
| 14 | 推送服务(APNs/FCM/微信/SMS) | Task 12 | 2h |
| 15 | AI分析引擎 | Task 13 | 3h |
| 16 | 管理后台API | Task 12 | 2h |

### 阶段E：UI原型（Task 17-20）— 可与阶段D并行

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 17 | 家属APP原型：首页+健康看板 | 无 | 2h |
| 18 | 家属APP原型：SOS告警+用药管理 | Task 17 | 2h |
| 19 | 管理后台原型：仪表盘+设备管理 | 无 | 2h |
| 20 | 管理后台原型：用户管理+订阅管理 | Task 19 | 1h |

### 阶段F：联调和文档（Task 21-22）

| Task | 描述 | 依赖 | 预计工时 |
|------|------|------|---------|
| 21 | 端到端联调测试 | Task 6,9,16,18,20 | 2h |
| 22 | 专利素材+许可证清单 | Task 1-20 | 1h |

---

## Task 1: 项目骨架+共享协议+基础设施

**Files:**
- Create: `.gitignore` — Git忽略规则
- Create: `LICENSE` — 版权文件
- Create: `firmware/bracelet/entry/main.c` — 手环入口空白框架
- Create: `firmware/bracelet/entry/CMakeLists.txt` — GD32构建配置
- Create: `firmware/pillbox/basic/main.c` — 药盒入口空白框架
- Create: `firmware/pillbox/basic/CMakeLists.txt` — ESP-IDF构建配置
- Create: `cloud/docker-compose.yml` — 完整基础设施编排
- Create: `cloud/go.mod` — Go模块初始化
- Create: `apps/family-app/pubspec.yaml` — Flutter模块初始化
- Create: `apps/admin-web/package.json` — Vue3模块初始化
- Create: `THIRD-PARTY-LICENSES.template` — 开源许可证记录模板

**Interfaces:**
- Produces: 项目目录结构、Docker基础设施、Go/Flutter/Vue模块骨架

---

## Task 2: 手环固件Entry档基础框架

**Files:**
- Create: `firmware/bracelet/entry/main.c` — FreeRTOS主循环，创建传感器/定位/通信任务
- Create: `firmware/bracelet/entry/board_init.c/h` — GD32E230板级初始化(GPIO/Clock/UART)
- Create: `firmware/bracelet/entry/sensors_ppg.c/h` — PPG心率血氧驱动(汇顶GT3x I2C)
- Create: `firmware/bracelet/entry/sensors_imu.c/h` — IMU加速度计陀螺仪驱动(ICM-42670 SPI)
- Create: `firmware/bracelet/entry/gps_nmea.c/h` — NMEA协议解析(GPS定位数据)
- Create: `firmware/bracelet/entry/cat1_at.c/h` — Cat1模组AT指令封装(广和通L610)
- Create: `firmware/bracelet/entry/display_st7789.c/h` — 1.14寸IPS LCD驱动(ST7789 SPI)
- Create: `firmware/bracelet/entry/battery_adc.c/h` — 锂电池ADC电量测量
- Create: `firmware/bracelet/entry/sos_button.c/h` — SOS实体按键检测(防抖+长按)
- Create: `firmware/bracelet/entry/free_rtos_tasks.c/h` — FreeRTOS任务定义和栈分配
- Create: `firmware/bracelet/entry/CMakeLists.txt` — PlatformIO/Keil构建配置
- Create: `firmware/bracelet/entry/README.md` — 开发说明
- Test: `firmware/bracelet/entry/test_sensors.c` — 传感器模拟测试
- Test: `firmware/bracelet/entry/test_gps_parser.c` — NMEA解析测试
- Test: `firmware/bracelet/entry/test_sos.c` — SOS按键测试

**Interfaces:**
- Produces: 心跳包`{"type":"heartbeat","dev_id":"BR-XXXX","bat":XX}`
- Produces: 定位数据`{"type":"location","dev_id":"BR-XXXX","lat":XX,"lon":XX,"acc":XX,"ts":XX}`
- Produces: 健康数据`{"type":"health","dev_id":"BR-XXXX","hr":XX,"spo2":XX,"step":XXXX}`
- Produces: SOS告警`{"type":"sos","dev_id":"BR-XXXX","lat":XX,"lon":XX,"ts":XX}`

---

## Task 3: 药盒固件Basic档基础框架

**Files:**
- Create: `firmware/pillbox/basic/main.c` — ESP-IDF主循环
- Create: `firmware/pillbox/basic/wifi_station.c/h` — WiFi STA模式连接
- Create: `firmware/pillbox/basic/led_gpio.c/h` — LED状态指示GPIO控制
- Create: `firmware/pillbox/basic/battery_manage.c/h` — 18650电池电量测量
- Create: `firmware/pillbox/basic/button_input.c/h` — 物理按键输入
- Create: `firmware/pillbox/basic/CMakeLists.txt` — ESP-IDF构建配置
- Create: `firmware/pillbox/basic/idf_component.yml` — ESP组件描述
- Create: `firmware/pillbox/basic/README.md` — 开发说明
- Test: `firmware/pillbox/basic/test_wifi_connect.c` — WiFi连接测试
- Test: `firmware/pillbox/basic/test_led.c` — LED控制测试

**Interfaces:**
- Produces: 心跳包`{"type":"heartbeat","dev_id":"PX-XXXX","bat":XX}`

---

## Task 4: 手环Plus档 — 电子围栏+跌倒检测

**Files:**
- Create: `firmware/bracelet/plus/fence_manager.c/h` — 电子围栏管理器(圆形+多边形，本地缓存围栏坐标)
- Create: `firmware/bracelet/plus/fence_check.c/h` — 位置点包含算法(射线法判断GPS坐标是否在围栏内)
- Create: `firmware/bracelet/plus/fall_detector.c/h` — 跌倒检测算法(IMU加速度突变+方向变化，阈值法初版)
- Create: `firmware/bracelet/plus/power_optimizer.c/h` — 自适应采样率调整(静止降频，运动升频)
- Modify: `firmware/bracelet/entry/main.c` — 集成Plus任务
- Test: `firmware/bracelet/plus/test_fence_check.c` — 点包含算法测试
- Test: `firmware/bracelet/plus/test_fall_detector.c` — 跌倒模拟数据测试

**Interfaces:**
- Consumes: GPS坐标数据(来自Task 2的gps_nmea)
- Consumes: IMU原始数据(来自Task 2的sensors_imu)
- Produces: 电子围栏告警`{"type":"fence_alert","dev_id":"BR-XXXX","fence_id":"XX","event":"enter|exit","lat":XX,"lon":XX,"ts":XX}`
- Produces: 跌倒检测`{"type":"fall","dev_id":"BR-XXXX","conf":0.XX,"lat":XX,"lon":XX,"ts":XX}`

---

## Task 5: 手环Pro档 — ECG+AMOLED+高精度GPS

**Files:**
- Create: `firmware/bracelet/pro/ecg_adc.c/h` — ECG心电采集(高精度ADC，差分放大信号处理)
- Create: `firmware/bracelet/pro/amoled_driver.c/h` — AMOLED屏幕驱动(1.9寸，SPI/I2C)
- Create: `firmware/bracelet/pro/gps_ublox.c/h` — u-blox M9N高精度GPS协议解析(支持BDS/GPS/GLONASS/GALILEO四系统)
- Create: `firmware/bracelet/pro/metal_chassis.c/h` — 金属机身散热/天线适配
- Test: `firmware/bracelet/pro/test_ecg_adc.c` — ECG ADC测试
- Test: `firmware/bracelet/pro/test_gps_ublox.c` — 四系统定位测试

---

## Task 6: 手环公共库 — MQTT+OTA+品牌

**Files:**
- Create: `firmware/bracelet/common/mqtt_client.c/h` — MQTT客户端封装(TLS连接，QoS1，遗嘱消息)
- Create: `firmware/bracelet/common/ota_update.c/h` — OTA固件升级(HTTP下载+双分区校验+回滚)
- Create: `firmware/bracelet/common/boot_logo.c/h` — Eregen开机动画(BMP解码显示3秒)
- Create: `firmware/bracelet/common/brand_voice.c/h` — 品牌语音提示音播放("颐贞提醒您...")
- Create: `firmware/bracelet/common/device_id.c/h` — 设备ID生成和存储(BR-XXXX格式，Flash持久化)
- Create: `firmware/bracelet/common/protocol_json.c/h` — JSON消息序列化和反序列化
- Test: `firmware/bracelet/common/test_mqtt_publish.c` — MQTT发布测试
- Test: `firmware/bracelet/common/test_device_id.c` — 设备ID生成测试
- Test: `firmware/bracelet/common/test_protocol_json.c` — JSON序列化测试

**Interfaces:**
- Consumes: 各档位产生的消息结构体
- Produces: 标准MQTT上行消息(通过EMQX转发到云端)
- Consumes: 下行MQTT消息(用药规则/配置更新/TTS/OTA指令)

---

## Task 7: 药盒Smart档 — 语音提醒+APP联动

**Files:**
- Create: `firmware/pillbox/smart/main.c` — Smart版主程序(在Basic基础上扩展)
- Create: `firmware/pillbox/smart/voice_reminder.c/h` — TTS定时语音提醒(SYN5300模块，中文播报"爷爷，该吃降压药了")
- Create: `firmware/pillbox/smart/reminder_scheduler.c/h` — 用药提醒调度器(支持多时段/多药品种类)
- Create: `firmware/pillbox/smart/app_link.c/h` — APP联动指令解析(远程设置提醒时间/暂停提醒)
- Create: `firmware/pillbox/smart/oled_status.c/h` — OLED用药状态显示(当前药格/下次提醒时间)
- Create: `firmware/pillbox/smart/volume_control.c/h` — 音量调节(≥65dB适老化要求)
- Test: `firmware/pillbox/smart/test_voice_reminder.c` — 语音提醒测试
- Test: `firmware/pillbox/smart/test_scheduler.c` — 调度器测试

**Interfaces:**
- Consumes: 云端`med_rule`用药规则消息
- Produces: 服药状态`{"type":"med_status","dev_id":"PX-XXXX","compartment":X,"taken":true/false,"ts":XX}`

---

## Task 8: 药盒Auto档 — 自动分药+光电检测

**Files:**
- Create: `firmware/pillbox/auto/main.c` — Auto版主程序(在Smart基础上扩展)
- Create: `firmware/pillbox/auto/stepper_motor.c/h` — 步进电机控制(28BYJ-48，精确旋转角度对应药格)
- Create: `firmware/pillbox/auto/pill_optical_detect.c/h` — 红外光电药片存在检测
- Create: `firmware/pillbox/auto/compartments_rotate.c/h` — 多格旋转药仓控制(4×7=28格)
- Create: `firmware/pillbox/auto/inventory_warn.c/h` — 空仓预警(连续3次检测不到药片)
- Create: `firmware/pillbox/auto/safety_lock.c/h` — 儿童安全锁(按压式锁扣控制)
- Create: `firmware/pillbox/auto/medicine_push.c/h` — 推药机构控制(防止卡药)
- Test: `firmware/pillbox/auto/test_stepper_motor.c` — 步进电机精度测试
- Test: `firmware/pillbox/auto/test_pill_detect.c` — 药片检测测试
- Test: `firmware/pillbox/auto/test_compartments.c` — 28格旋转测试

**Interfaces:**
- Consumes: `med_rule`用药规则
- Produces: `med_status`(含compartment/taken字段)
- Produces: 空仓告警`{"type":"inventory_warning","dev_id":"PX-XXXX","compartments_empty":[1,3,7],"ts":XX}`

---

## Task 9: 药盒公共库

**Files:**
- Create: `firmware/pillbox/common/motor_control.c/h` — 通用电机控制接口(速度/方向/步数)
- Create: `firmware/pillbox/common/tts_playback.c/h` — 中文TTS播放封装
- Create: `firmware/pillbox/common/led_patterns.c/h` — LED闪烁模式库(绿色=正常/红色=告警/蓝色=配对)
- Create: `firmware/pillbox/common/wifi_mqtt_bridge.c/h` — WiFi连接+MQTT通信统一桥接
- Create: `firmware/pillbox/common/device_id.c/h` — 设备ID管理(PX-XXXX格式)
- Create: `firmware/pillbox/common/protocol_json.c/h` — JSON消息序列化
- Create: `firmware/pillbox/common/ota_update.c/h` — OTA升级
- Test: `firmware/pillbox/common/test_motor_control.c`
- Test: `firmware/pillbox/common/test_tts.c`
- Test: `firmware/pillbox/common/test_protocol_json.c`

---

## Task 10: 云平台Docker Compose基础设施

**Files:**
- Create: `cloud/docker-compose.yml` — 生产环境(PostgreSQL 16 + InfluxDB 2.x + Redis 7 + NATS 2.10 + EMQX 5.x)
- Create: `cloud/docker-compose.dev.yml` — 开发附加(Grafana 10 + Prometheus 2.x + Loki 2.x + Promtail)
- Create: `cloud/config/postgresql/init.sql` — 数据库初始化(schema: users/devices/subscriptions/health_config)
- Create: `cloud/config/influxdb/init.sh` — InfluxDB bucket/organization初始化(org=eregen, buckets=health_data,alerts)
- Create: `cloud/config/redis/redis.conf` — Redis配置(内存限制+持久化)
- Create: `cloud/config/nats/nats.conf` — NATS配置(集群模式关闭，单节点)
- Create: `cloud/config/emqx/emqx.conf` — EMQX配置(TLS认证+ACL+规则引擎)
- Create: `cloud/scripts/start-dev.sh` — 一键启动脚本
- Create: `cloud/scripts/setup-influx.sh` — InfluxDB初始化脚本
- Test: `cloud/scripts/start-dev.sh` — 验证所有容器健康检查通过

---

## Task 11: MQTT设备接入网关

**Files:**
- Create: `cloud/gateway/cmd/server.go` — Go服务入口
- Create: `cloud/gateway/internal/mqtt/client.go` — EMQX MQTT客户端连接(TLS，证书认证)
- Create: `cloud/gateway/internal/mqtt/topic_router.go` — MQTT主题路由(dev/bracelet/+ / dev/pillbox/+)
- Create: `cloud/gateway/internal/mqtt/message_handler.go` — 消息解析(心跳/定位/健康/SOS/跌倒/药盒状态)
- Create: `cloud/gateway/internal/mqtt/device_auth.go` — 设备认证(Token验证+设备绑定用户)
- Create: `cloud/gateway/internal/nats/publisher.go` — NATS消息发布(结构化主题)
- Create: `cloud/gateway/internal/nats/schema.go` — NATS消息Schema定义
- Create: `cloud/gateway/internal/mqtt/mqtt_test.go` — MQTT消息处理测试
- Create: `cloud/gateway/internal/nats/nats_test.go` — NATS发布测试
- Create: `cloud/gateway/config/config.go` — 配置加载(YAML)

**Interfaces:**
- Consumes: MQTT主题`eregen/device/bracelet/{dev_id}/up`和`eregen/device/pillbox/{dev_id}/up`
- Produces: NATS主题`eregen.event.heartbeat`/`eregen.event.location`/`eregen.event.health`/`eregen.event.sos`/`eregen.event.fall`/`eregen.event.med_status`

---

## Task 12: REST API服务器

**Files:**
- Create: `cloud/api-server/cmd/server.go` — Gin服务入口
- Create: `cloud/api-server/internal/config/config.go` — 配置管理
- Create: `cloud/api-server/internal/models/user.go` — 用户模型(User+Elderly+Family关联)
- Create: `cloud/api-server/internal/models/device.go` — 设备模型(Device+FirmwareVersion)
- Create: `cloud/api-server/internal/models/health.go` — 健康数据模型
- Create: `cloud/api-server/internal/models/subscription.go` — 订阅模型
- Create: `cloud/api-server/internal/handler/user.go` — 用户CRUD+设备绑定
- Create: `cloud/api-server/internal/handler/device.go` — 设备查询/配置下发/OTA状态
- Create: `cloud/api-server/internal/handler/health.go` — 健康数据查询(InfluxDB)
- Create: `cloud/api-server/internal/handler/location.go` — 实时位置+历史轨迹
- Create: `cloud/api-server/internal/database/postgres.go` — PostgreSQL连接池
- Create: `cloud/api-server/internal/database/influx.go` — InfluxDB客户端
- Create: `cloud/api-server/internal/cache/redis.go` — Redis缓存层
- Create: `cloud/api-server/internal/middleware/jwt.go` — JWT认证中间件
- Create: `cloud/api-server/internal/middleware/cors.go` — CORS
- Create: `cloud/api-server/internal/middleware/logging.go` — 请求日志
- Create: `cloud/api-server/internal/handler/user_test.go` — 用户Handler测试
- Create: `cloud/api-server/internal/handler/device_test.go` — 设备Handler测试

**API Endpoints:**
- `POST /api/v1/auth/login` — 登录
- `GET /api/v1/users/{id}` — 用户信息
- `POST /api/v1/users/{id}/devices` — 绑��设备
- `GET /api/v1/devices/{id}/status` — 设备状态
- `PUT /api/v1/devices/{id}/config` — 设备配置下发
- `GET /api/v1/devices/{id}/location/current` — 实时位置
- `GET /api/v1/devices/{id}/location/history` — 历史轨迹
- `GET /api/v1/devices/{id}/health` — 健康数据
- `GET /api/v1/health/trend` — 健康趋势(心率/血氧/步数)
- `POST /api/v1/devices/{id}/tts` — 远程语音播报
- `WS /api/v1/ws/alerts` — WebSocket实时告警推送

---

## Task 13: 数据处理管道

**Files:**
- Create: `cloud/data-pipeline/cmd/server.go` — NATS消费者入口
- Create: `cloud/data-pipeline/internal/consumer/nats.go` — NATS消息订阅
- Create: `cloud/data-pipeline/internal/processor/health_data.go` — 健康数据处理器(写入InfluxDB)
- Create: `cloud/data-pipeline/internal/processor/location_data.go` — 位置数据处理器(去重+轨迹压缩)
- Create: `cloud/data-pipeline/internal/processor/alert_processor.go` — 告警优先级处理(P0跌倒/SOS > P1围栏/漏服 > P2心跳丢失)
- Create: `cloud/data-pipeline/internal/influx/client.go` — InfluxDB写入客户端(v2 SDK)
- Create: `cloud/data-pipeline/internal/alert/priority.go` — 告警优先级定义和路由
- Create: `cloud/data-pipeline/internal/processor/health_test.go` — 健康数据测试
- Create: `cloud/data-pipeline/internal/processor/alert_test.go` — 告警优先级测试

---

## Task 14: 推送服务

**Files:**
- Create: `cloud/push-service/cmd/server.go` — 推送服务入口
- Create: `cloud/push-service/internal/config/config.go` — 配置(APNs证书/FCM密钥/微信AppID/阿里云SMS)
- Create: `cloud/push-service/internal/apns/client.go` — Apple Push Notification客户端(v2 HTTP API)
- Create: `cloud/push-service/internal/fcm/client.go` — Firebase Cloud Messaging客户端
- Create: `cloud/push-service/internal/wechat/client.go` — 微信订阅消息客户端
- Create: `cloud/push-service/internal/sms/client.go` — 阿里云短信客户端
- Create: `cloud/push-service/internal/router/router.go` — 推送渠道路由(根据用户设备类型选择APNs/FCM)
- Create: `cloud/push-service/internal/router/fallback.go` — 多渠道兜底(紧急事件同时发APP+短信)
- Create: `cloud/push-service/internal/consumer/nats.go` — NATS告警消息消费
- Create: `cloud/push-service/internal/push/push_test.go` — 推送测试

**Push Channels:**
- APNs (iOS) — 主要通道
- FCM (Android) — 主要通道
- 微信订阅消息 — 轻量通道(无需装APP)
- 阿里云SMS — 紧急事件兜底(SOS/跌倒必达)
- 电话语音 — P0级别告警兜底(预留接口)

---

## Task 15: AI分析引擎

**Files:**
- Create: `cloud/data-pipeline/internal/ai/risk_scoring.go` — 跌倒风险评分(基于历史步态异常频率+心率变异性)
- Create: `cloud/data-pipeline/internal/ai/medication_pattern.go` — 用药习惯学习(统计各时段漏服概率，智能调整提醒策略)
- Create: `cloud/data-pipeline/internal/ai/anomaly_detection.go` — 异常模式识别(心率持续异常/血氧偏低/活动量骤降)
- Create: `cloud/data-pipeline/internal/ai/model_store.go` — 模型参数持久化(Redis+PostgreSQL)
- Create: `cloud/data-pipeline/internal/ai/ai_test.go` — AI逻辑单元测试

---

## Task 16: 管理后台API

**Files:**
- Create: `cloud/admin-api/cmd/server.go` — 管理后台API入口
- Create: `cloud/admin-api/internal/handler/dashboard.go` — 仪表盘统计数据(设备在线率/告警数/用户增长)
- Create: `cloud/admin-api/internal/handler/device_mgmt.go` — 设备管理(列表/状态筛选/固件版本统计/OTA批量升级)
- Create: `cloud/admin-api/internal/handler/user_mgmt.go` — 用户管理(老人/家属/运营角色/权限)
- Create: `cloud/admin-api/internal/handler/subscription.go` — 订阅管理(状态/续费/降级原因)
- Create: `cloud/admin-api/internal/handler/report.go` — 数据导出(CSV)
- Create: `cloud/admin-api/internal/handler/dashboard_test.go` — 仪表盘测试
- Create: `cloud/admin-api/internal/handler/device_mgmt_test.go` — 设备管理测试

---

## Task 17: 家属APP原型 — 首页+健康看板

**Files:**
- Create: `apps/ui-prototypes/family/home.html` — 实时定位地图页
- Create: `apps/ui-prototypes/family/health.html` — 健康数据看板
- Create: `apps/ui-prototypes/css/family.css` — 适老化样式(大字体/高对比度)
- Create: `apps/ui-prototypes/js/mock-data.js` — 模拟数据

**UI Specifications:**
- 首页: 深色地图背景(模拟高德风格) + 老人头像卡片 + 绿色位置标记点 + 蓝色电子围栏虚线圈 + 底部三卡片(心率/血氧/电量) + 底部导航栏(首页/健康/用药/设置)
- 健康看板: 心率趋势折线图(绿色，24h) + 血氧柱状图(蓝色，7d) + 今日步数大字报 + 睡眠时长 + 异常告警红色横幅

---

## Task 18: 家属APP原型 — SOS告警+用药管理

**Files:**
- Create: `apps/ui-prototypes/family/sos.html` — SOS告警中心
- Create: `apps/ui-prototypes/family/medication.html` — 用药管理
- Modify: `apps/ui-prototypes/css/family.css` — 新增告警红色主题

**UI Specifications:**
- SOS: 顶部红色紧急区域(大字"SOS告警") + 告警时间线列表(时间+类型+状态) + 一键呼叫电话按钮(绿色大按钮) + 处理记录
- 用药: 今日用药时间表(已服绿色✓/未服红色○/逾期黄色⚠) + 药品详情(名称/剂量/频次) + 远程配置按钮 + 服药记录周图表

---

## Task 19: 管理后台原型 — 仪表盘+设备管理

**Files:**
- Create: `apps/ui-prototypes/admin/dashboard.html` — 仪表盘
- Create: `apps/ui-prototypes/admin/devices.html` — 设备管理
- Create: `apps/ui-prototypes/css/admin.css` — 管理后台样式
- Create: `apps/ui-prototypes/js/admin-charts.js` — ECharts图表封装

**UI Specifications:**
- 仪表盘: 顶部KPI卡片(设备总数/在线数/今日告警/活跃用户) + 设备在线率环形图 + 今日告警趋势折线图 + 用户增长曲线 + 活跃设备TOP5列表
- 设备管理: 筛选栏(类型/状态/固件版本) + 设备表格(ID/类型/状态灯/固件版本/最后上线/操作) + OTA批量升级弹窗

---

## Task 20: 管理后台原型 — 用户管理+订阅管理

**Files:**
- Create: `apps/ui-prototypes/admin/users.html` — 用户管理
- Create: `apps/ui-prototypes/admin/subscriptions.html` — 订阅管理
- Modify: `apps/ui-prototypes/css/admin.css`

**UI Specifications:**
- 用户: 用户表格(姓名/类型标签[老人/家属/运营]/设备数/注册时间/状态) + 角色筛选 + 权限编辑弹窗
- 订阅: 订阅状态分布饼图(免费/高级/尊享/已过期) + 续费记录表 + 降级原因统计条形图

---

## Task 21: 端到端联调测试

**Activities:**
1. `cd cloud && docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d` — 启动全部基础设施
2. 验证所有容器健康: `docker compose ps`
3. 使用`mosquitto_pub`模拟手环发送MQTT消息 → 验证Gateway接收→NATS路由→InfluxDB写入
4. 使用`mosquitto_pub`模拟药盒发送MQTT消息 → 验证用药规则下发→服药状态上报
5. `curl`测试REST API全链路: 登录→绑定设备→模拟数据→查询健康→触发告警→推送
6. 浏览器打开`apps/ui-prototypes/family/home.html`和`admin/dashboard.html`，验证与API对接
7. 编写`scripts/e2e-test.sh`自动化联调测试脚本

---

## Task 22: 专利素材+许可证清单

**Activities:**
1. 扫描所有子目录的依赖: `find . -name "go.mod" -exec cat {} \;` / `find . -name "pubspec.yaml" -exec cat {} \;`
2. 为每个子项目生成`THIRD-PARTY-LICENSES`文件
3. 确认无GPL/AGPL/LGPL组件
4. 整理专利交底书素材:
   - 专利1: 基于多源传感器融合的老年跌倒检测方法及系统(手环IMU+PPG+位置)
   - 专利2: 一种智能药盒的自动分药及服药状态闭环确认方法
   - 专利3: 面向居家养老的多级告警推送系统及方法(P0/P1/P2优先级+多渠道兜底)
   - 专利4: 基于电子围栏的老年人居活动态感知及告警方法
   - 专利5: 家属多人协同的老年人健康监护系统及交互方法
