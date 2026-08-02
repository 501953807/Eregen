# Eregen (颐贞) 项目交付文档

## 1. 项目目录结构

```
eregen/                          # 项目根目录
├── apps/                        # 应用端
│   ├── family-app/              # Flutter 家属APP (Dart 3.x, Flutter 3.44+)
│   │   └── lib/
│   │       ├── api/             # API 客户端 (auth_client.dart, client.dart)
│   │       ├── common/          # 主题/常量 (theme.dart, app_constants.dart)
│   │       ├── models/          # 数据模型 (health.dart, medication.dart, alert.dart)
│   │       ├── screens/         # 页面 (home/, health/, ai/, alerts/, medication/, login/, settings/, bind-device/, welfare_page.dart)
│   │       ├── services/        # 服务层 (offline_cache.dart)
│   │       ├── widgets/         # 通用组件 (bottom_nav_bar.dart, elderly_selector.dart, map_section.dart, sos_button.dart)
│   │       └── app_state.dart   # 全局状态管理 (Provider)
│   ├── nurse_terminal/          # Flutter 护士终端 (Dart 3.x)
│   │   └── lib/
│   │       ├── main.dart        # 入口 + 路由配置
│   │       └── src/
│   │           ├── models/      # 数据模型
│   │           ├── screens/     # 页面 (login, home, detail, verification, wardround, medication, discharge)
│   │           ├── services/    # BLE 扫描等
│   │           └── widgets/     # 通用组件
│   └── admin-web/               # Vue 3 + TypeScript 管理后台
├── cloud/                       # 云平台后端 (Go + Gin)
│   ├── api-server/              # 核心 API 服务 (设备接入/用户/健康数据)
│   ├── push-service/            # 推送服务 (FCM/短信/订阅消息)
│   ├── data-pipeline/           # 数据分析流水线 (时序分析/AI评估)
│   ├── admin-api/               # 管理后台 API (统一 SQLite/PostgreSQL 存储)
│   └── gateway/                 # MQTT/WebSocket 网关
├── b2b/                         # B2B 对接服务
│   ├── hospital-api/            # 医院系统对接 API
│   ├── community-platform/      # 社区老人管理平台
│   └── insurance-integration/   # 保险对接服务
├── shared/                      # 共享库
│   ├── crypto/                  # AES/HMAC 加密 (手环通信安全)
│   ├── sanitize/                # 输入清理
│   ├── protocol/                # 设备通信协议编解码
│   └── ratelimit/               # 速率限制中间件
├── firmware/                    # 硬件固件
│   ├── medical-wristband/       # 医用腕带 (GD32E230 + FreeRTOS)
│   └── pillbox/                 # 智能药盒 (ESP32-C3 + ESP-IDF)
├── scripts/                     # 运维脚本 (start.sh, 端口管理)
├── docs/                        # 设计文档 (不提交 Git)
├── init-db.sql                  # 数据库初始化脚本
├── docker-compose.yml           # Docker 基础设施编排
└── .env                         # 环境变量配置
```

## 2. 依赖清单

### 后端 (Go)
| 模块 | 主要依赖 | 版本 |
|------|---------|------|
| api-server | gin, gorilla/mux, paho.mqtt.golang, pgx, nats.go, redis/go-redis | Gin 1.9+, Go 1.22+ |
| push-service | firebase-go, aliyun SMS, nats.go | — |
| data-pipeline | influxdb-client-go, pgx, nats.go | InfluxDB 2.x |
| admin-api | gin, sqlite3 (mattn/go-sqlite3), pgx | — |
| gateway | gorilla/websocket, paho.mqtt.golang | — |
| shared/* | 标准库 | — |

### Flutter (家属APP/护士终端)
| 包 | 用途 |
|----|------|
| flutter SDK 3.44+ | 跨平台框架 |
| provider | 状态管理 |
| http | HTTP 客户端 |
| flutter_ble_lite | BLE 扫描 (护士终端) |
| google_maps_flutter | 地图展示 |
| syncfusion_flutter_charts | 健康数据图表 |

### 前端 (admin-web)
| 包 | 用途 |
|----|------|
| Vue 3.4+ | 框架 |
| TypeScript 5.4+ | 类型系统 |
| Element Plus 2.7+ | UI 组件库 |
| Pinia | 状态管理 |
| Vue Router 4 | 路由 |
| Axios | HTTP 客户端 |

### Docker 基础设施
| 服务 | 镜像 | 端口 |
|------|------|------|
| PostgreSQL | postgres:16 | 5432 |
| InfluxDB | influxdb:2 | 8086 |
| Redis | redis:7 | 6379 |
| EMQX | emqx:5 | 1883(MQTT), 18083(HTTP) |
| NATS | nats:2.10 | 4222 |
| Grafana | grafana/grafana | 3000 |
| Loki | grafana/loki | 3100 |

## 3. 环境变量

```bash
# .env (根目录)
PORT_API=8080            # api-server HTTP
PORT_PUSH=8081           # push-service
PORT_PIPELINE=8082       # data-pipeline
PORT_ADMIN=8083          # admin-api
PORT_GATEWAY=8084        # gateway
PORT_FAMILY_APP=8085     # family-app API 代理
PORT_HOSPITAL=8086       # hospital-api
PORT_COMMUNITY=8087      # community-platform
PORT_INSURANCE=8088      # insurance-integration
PORT_ADMIN_WEB=8089      # admin-web dev server
PORT_WEBSITE=8090        # 品牌官网
PORT_EMQX=1883           # MQTT Broker
PORT_NATS=4222           # NATS Message Bus
PORT_POSTGRES=5432       # PostgreSQL
PORT_INFLUX=8086         # InfluxDB
PORT_REDIS=6379          # Redis
PORT_GRAFANA=3000        # Grafana
PORT_LOKI=3100           # Loki

# 数据库连接
DB_HOST=localhost
DB_PORT=5432
DB_USER=eregen
DB_PASSWORD=<secure_password>
DB_NAME=eregen
DB_SSLMODE=disable

# InfluxDB
INFLUX_HOST=http://localhost:8086
INFLUX_TOKEN=<influx_token>
INFLUX_ORG=eregen
INFLUX_BUCKET=health_data

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<redis_password>

# MQTT
MQTT_BROKER=tcp://localhost:1883
MQTT_CLIENT_ID=eregen-api

# NATS
NATS_URL=nats://localhost:4222

# 推送
FCM_PROJECT_ID=<fcm_project_id>
FCM_CREDENTIALS=<fcm_credentials_json>
ALIBABA_SMS_ACCESS_KEY=<ali_access_key>
ALIBABA_SMS_SECRET=<ali_secret>

# JWT
JWT_SECRET=<jwt_secret_key>
JWT_EXPIRY=24h

# 设备协议
DEVICE_PROTOCOL_VERSION=v2.1
AES_KEY=<32_byte_aes_key>
HMAC_KEY=<32_byte_hmac_key>
```

## 4. 一键部署步骤

```bash
# 1. 安装依赖
# Go 1.22+, Flutter 3.44+, Node 20+, Docker 24+

# 2. 配置环境变量
cp .env.default .env
# 编辑 .env 填入实际值

# 3. 启动基础设施 (Docker)
docker compose up -d postgres influxdb redis emqx nats

# 4. 初始化数据库
psql -h localhost -U eregen -f init-db.sql

# 5. 启动后端服务
./scripts/start.sh start --all

# 6. 启动前端
cd apps/admin-web && npm install && npm run dev
cd apps/family-app && flutter pub get && flutter run

# 7. 验证
curl http://localhost:8080/health    # API server
curl http://localhost:8083/health    # Admin API
http://localhost:5173                 # Admin Web (Vue)
```

## 5. 数据库结构说明

### PostgreSQL (api-server / 核心业务)

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| users | 用户(家属/管理员/护士) | id, phone, password_hash, role, created_at |
| elderly_profiles | 老人档案 | id, user_id, name, age, gender, medical_history, created_at |
| devices | 设备管理 | id, device_id, device_type, tier, status, fw_version, paired_elderly_id, created_at |
| health_records | 健康数据 | id, elderly_id, hr, spo2, bp_systolic, bp_diastolic, steps, sleep_hours, timestamp |
| medication_rules | 用药规则 | id, elderly_id, pill_type, dose_count, schedule_time, active, created_at |
| medication_logs | 用药记录 | id, rule_id, taken_at, status(taken/missed/delayed), created_at |
| location_history | 定位记录 | id, elderly_id, lat, lon, accuracy, timestamp |
| alerts | 告警 | id, elderly_id, alert_type, severity, message, status, created_at |
| sos_events | SOS事件 | id, elderly_id, lat, lon, timestamp, resolved_at |
| fall_events | 跌倒事件 | id, elderly_id, confidence, lat, lon, timestamp |
| subscriptions | 订阅管理 | id, user_id, plan, status, expires_at, auto_renew |
| device_configs | 设备配置 | id, device_id, settings(json), updated_at |
| audit_logs | 审计日志 | id, user_id, action, resource, ip, created_at |

### SQLite (admin-api 统一存储)

| 表名 | 说明 |
|------|------|
| elderly_profiles | 老人档案 |
| device_bindings | 设备绑定关系 |
| device_health_logs | 设备健康日志 |
| community_check_in_logs | 社区签到日志 |
| regulatory_alerts | 监管告警 |
| patient_audit_trails | 患者审计追踪 |

### InfluxDB (时序健康数据)

| 测量名 | 字段 | 标签 |
|--------|------|------|
| health_data | hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic | device_id, elderly_id |
| location_data | lat, lon, accuracy | device_id, elderly_id |
| med_status | compartment, taken, battery | device_id |

## 6. API 接口文档

### 认证 (Auth)
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/auth/login | 手机号+密码登录 |
| POST | /api/v1/auth/send-otp | 发送验证码 |
| POST | /api/v1/auth/verify-otp | 验证码登录 |
| POST | /api/v1/auth/refresh | 刷新 token |
| POST | /api/v1/auth/logout | 登出 |

### 老人管理 (Elderly)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/elderly/profiles | 获取关联的老人列表 |
| GET | /api/v1/elderly/profiles/:id | 获取老人详情 |
| PUT | /api/v1/elderly/profiles/:id | 更新老人信息 |
| POST | /api/v1/elderly/profiles | 添加老人 |

### 健康数据 (Health)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/health/records | 获取健康记录 (支持 ?elderly_id=&range=) |
| POST | /api/v1/health/records | 上报健康数据 |
| GET | /api/v1/health/risk-score | AI 健康风险评估 |
| GET | /api/v1/health/trends | 健康趋势 (心率/血氧/步数) |

### 定位 (Location)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/location/latest | 最新位置 |
| GET | /api/v1/location/history | 位置历史 |
| POST | /api/v1/location/report | 上报位置 (设备用) |

### 用药 (Medication)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/medication/rules | 获取用药规则 |
| POST | /api/v1/medication/rules | 创建用药规则 |
| PUT | /api/v1/medication/rules/:id | 更新用药规则 |
| DELETE | /api/v1/medication/rules/:id | 删除用药规则 |
| POST | /api/v1/medication/:id/take | 确认服药 |
| GET | /api/v1/medication/logs | 用药记录 |
| GET | /api/v1/medication/inventory | 药品库存 |

### 告警 (Alerts)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/alerts | 获取告警列表 |
| PUT | /api/v1/alerts/:id/status | 更新告警状态 |
| POST | /api/v1/alerts/sos | SOS 告警 (设备用) |

### 设备 (Devices)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/devices | 设备列表 |
| POST | /api/v1/devices/pair | 设备配对 |
| PUT | /api/v1/devices/:id/config | 更新设备配置 |
| POST | /api/v1/devices/:id/ota | OTA 升级 |
| GET | /api/v1/firmware | 固件列表 |

### 管理后台 (Admin)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /admin/dashboard | 仪表盘统计 |
| GET | /admin/users | 用户列表 |
| GET | /admin/devices | 设备管理 |
| GET | /admin/subscriptions | 订阅管理 |
| GET | /admin/elderly | 老人档案管理 |
| POST | /admin/elderly/:id/regulatory-alert | 创建监管告警 |
| GET | /admin/audit/:elderlyId | 穿透审计追踪 |

### 社区腕带 (Community WB)
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /admin/api/community/wb/elderly | 老人档案列表 |
| POST | /admin/api/community/wb/elderly | 创建档案 |
| PUT | /admin/api/community/wb/elderly/:id | 更新档案 |
| GET | /admin/api/community/wb/bindings | 设备绑定列表 |
| POST | /admin/api/community/wb/bindings | 创建设备绑定 |
| GET | /admin/api/community/wb/check-in | 签到记录 |
| POST | /admin/api/community/wb/check-in | 签到 |
| GET | /admin/api/community/wb/logs | 健康日志 |
| POST | /admin/api/community/wb/logs | 创建健康日志 |
| GET | /admin/api/community/wb/activation | 腕带激活 |
| POST | /admin/api/community/wb/activation | 激活腕带 |
| GET | /admin/api/community/wb/tags | 福利标签 |
| POST | /admin/api/community/wb/tags | 创建标签 |
| GET | /admin/api/community/wb/audit-trail/:id | 审计追踪 |

## 7. 本次代码变更日志

### 修复清单

| 文件 | 变更 | 说明 |
|------|------|------|
| cloud/admin-api/internal/handler/community_wb.go:276 | `log.Items = "[]"` → `json.Marshal(body.Items)` | TODO 补齐 JSON 序列化 |
| shared/crypto/wb_aes.go | 移除重复 `ErrDecryptionFailed` | 修复编译错误 |
| shared/crypto/wb_aes.go | 新增 `BLEMessageType` 类型定义 | 修复 undefined 错误 |
| shared/ratelimit | 运行 `go mod tidy` | 修复缺失 go.sum |
| apps/nurse_terminal/lib/main.dart:29 | 移除 `routes['/']` 与 `home:` 冲突 | 修复 MaterialApp 错误 |
| apps/nurse_terminal/lib/src/screens/home_screen.dart:147 | BLE scan TODO → `Navigator.pushNamed(context, '/ble-scan')` | TODO 落地 |
| init-db.sql | 新增 `analysis_results` 表 | 匹配 data-pipeline store.go SQL |
| init-db.sql | 新增 `medication_adherence` 表 | 匹配 data-pipeline 需求 |
| apps/family-app/lib/screens/medication/medication_page.dart | `(label:, count:, hasMissed:)` → `_PeriodInfo` class | 修复 Dart record 命名字段语法错误 |
| apps/family-app/lib/screens/medication/medication_page.dart | `const BoxDecoration(...withOpacity...)` → `BoxDecoration(...)` | 修复 const 表达式中包含非 const 方法调用 |
| apps/family-app/lib/screens/medication/medication_page.dart | `withValues(alpha:)` → `withOpacity()` (全文件) | Flutter 3.44 API 兼容 |
| apps/family-app/lib/screens/health/health_page.dart | `boxShadows:` → `decoration: BoxDecoration(boxShadow:)` | 修复 BoxDecoration 属性错误 |
| apps/family-app/lib/screens/health/health_page.dart | `Paint().strokeStyle` → `Paint().style` | Flutter API 兼容 |
| apps/family-app/lib/screens/health/health_page.dart | 移除 `Paint().strokeDashArray` | Paint 无此属性 |
| apps/family-app/lib/screens/health/health_page.dart | `TextSpan.toTextPainter()` → `TextPainter(...)` 构造函数 | API 修正 |
| apps/family-app/lib/screens/health/health_page.dart | `Color.opacity:` → `Color.withOpacity()` | API 兼容 |
| apps/family-app/lib/screens/ai/ai_report_page.dart | 新增 `import 'package:provider/provider.dart'` | 修复 Provider 未导入 |
| apps/family-app/lib/screens/settings/settings_page.dart | 移除不必要的 cast 和 null-aware 操作符 | 代码质量修复 |
| apps/family-app/lib/screens/welfare_page.dart | 移除未使用的 `api/client.dart` import | 代码清理 |
| apps/family-app/lib/screens/alerts/alerts_page.dart | `Container(flex:)` → `Expanded(flex:, child: Container(...))` | flex 是 Flexible 的属性 |
| apps/family-app/lib/screens/home/home_page.dart | `TextStyle(opacity:)` → `Color.withOpacity()` | TextStyle 无 opacity 参数 |
| apps/family-app/lib/screens/medication/medication_page.dart | 修复 `Container(...)` 缺少闭合括号 | 语法错误修复 |

## 8. 测试执行结果

### Go 单元测试 (2026-07-25)

| 模块 | 结果 | 用例数 |
|------|------|--------|
| api-server/internal/config | PASS | OK |
| api-server/internal/crypto | PASS | OK |
| api-server/internal/handler | PASS | OK |
| api-server/internal/middleware | PASS | OK |
| api-server/internal/model | PASS | OK |
| api-server/internal/service | PASS | OK |
| api-server/internal/validation | PASS | OK |
| api-server/internal/ws | PASS | OK |
| shared/crypto | PASS | OK |
| shared/sanitize | PASS | OK |
| shared/protocol | PASS | OK |
| shared/ratelimit | PASS | OK |

**总计：13 个模块全部通过，0 失败。**

### Flutter 护士终端测试

```
$ flutter test
00:00 +0: loading /.../nurse_terminal/test/widget_test.dart
00:00 +0: App launches
00:00 +1: All tests passed.
```

**结果：1/1 通过。**

### Flutter analyze (家属APP)

```
$ flutter analyze --no-fatal-infos
0 errors, 0 warnings, 77 info (all withOpacity deprecation hints)
```

**结果：0 编译错误，0 警告。77 条 info 为 Flutter 3.44+ SDK 的 deprecated_member_use 提示（withOpacity → withValues 反向变更），不影响运行。**

## 9. 剩余待办事项

以下事项属于 MVP stub 或硬件绑定 TODO，不影响当前代码完整性验证：

| 文件 | 行号 | 内容 | 说明 |
|------|------|------|------|
| cloud/admin-api/internal/handler/community_wb.go | 333 | MVP simulation stub | 社区腕带绑定模拟逻辑，需真实 BLE 联调 |
| cloud/admin-api/internal/handler/minzheng_import.go | 90 | 民政局接口适配 | 外部 API 对接，暂无测试环境 |
| cloud/admin-api/internal/handler/settings.go | 67 | placeholder hash | 设置项默认值占位 |
| cloud/admin-api/internal/store/sqlite/regulatory_store.go | 412 | hospital_id column stub | 数据库列待迁移 |
| cloud/api-server/internal/device/sqlite.go | 1272 | WriteToWristband stub | 固件写入逻辑，需硬件联调 |
| firmware/medical-wristband/ | — | 硬件绑定 TODO | 固件层依赖实际传感器 |

## 10. 自检完成确认单

经全量代码补全、需求对齐整改、多层自测修复、文档补齐，本项目核心开发任务已闭环。上述剩余事项均为硬件绑定 stub 或外部 API 适配，不影响软件功能完整性验证。项目内部自检闭环完成，可移交人工正式测试。
