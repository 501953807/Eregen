# ③ 云平台后端 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：③ 云平台后端 (Go + Gin)  
> 包含模块：gateway / api-server / push-service / data-pipeline / admin-api

---

## 1. 概述

### 1.1 职责

云平台后端是整个 Eregen 系统的中枢，负责设备接入、数据存储、AI 分析、告警推送和运营管理。由 5 个 Go 微服务组成，通过 NATS JetStream 消息总线解耦。

### 1.2 微服务架构

```
┌─────────────────────────────────────────────────────────┐
│                     设备接入层                            │
│  ┌──────────────┐                                        │
│  │   gateway    │ ← MQTT 设备接入 → NATS 发布            │
│  └──────────────┘                                        │
├─────────────────────────────────────────────────────────┤
│                     数据处理层                            │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  api-server  │  │ data-pipeline│                     │
│  │  REST API    │  │ AI 分析引擎   │                     │
│  └──────────────┘  └──────────────┘                     │
├─────────────────────────────────────────────────────────┤
│                     推送分发层                            │
│  ┌──────────────┐                                        │
│  │push-service  │ ← NATS 订阅 → FCM/微信/短信           │
│  └──────────────┘                                        │
├─────────────────────────────────────────────────────────┤
│                     运营管理层                            │
│  ┌──────────────┐                                        │
│  │  admin-api   │ ← 管理后台专用 API                      │
│  └──────────────┘                                        │
└─────────────────────────────────────────────────────────┘
         │              │              │
    ┌────┴────┐   ┌────┴────┐   ┌────┴────┐
    │SQLite        │   │审计日志     │   │ NATS总线  │
    └─────────┘   └─────────┘   └─────────┘
```

**v2.1 变更：SQLite 统一存储 + 医用腕带模块**

```
┌─────────────────────────────────────────────────────────┐
│                     设备接入层                            │
│  ┌──────────────┐                                        │
│  │   gateway    │ ← MQTT 设备接入 → NATS 发布            │
│  └──────────────┘                                        │
├─────────────────────────────────────────────────────────┤
│                     数据处理层                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  api-server  │  │ data-pipeline│  │medical_wb   │  │
│  │  REST API    │  │ AI 分析引擎   │  │ 医护工作站API │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
├─────────────────────────────────────────────────────────┤
│                     推送分发层                            │
│  ┌──────────────┐                                        │
│  │push-service  │ ← NATS 订阅 → FCM/微信/短信           │
│  └──────────────┘                                        │
├─────────────────────────────────────────────────────────┤
│                     运营管理层                            │
│  ┌──────────────┐                                        │
│  │  admin-api   │ ← 管理后台专用 API                      │
│  │  + medical_wb│ ← 医护工作站 + 护士核验终端后端         │
│  │  + regulatory│ ← 监管闭环（规则引擎 R01-R08、电子围栏）│
│  │  + community │ ← 社区老人多维身份腕带                  │
│  └──────────────┘                                        │
└─────────────────────────────────────────────────────────┘
         │              │              │
    ┌────┴────┐   ┌────┴────┐   ┌────┴────┐
    │ SQLite  │   │  NATS   │   │审计日志  │
    └─────────┘   └─────────┘   └─────────┘
```

---

## 2. gateway — MQTT 设备接入

### 2.1 职责

接收所有手环/药盒/医用腕带设备的 MQTT 连接，解析上行消息，验证设备身份，将有效数据发布到 NATS 总线，同时将设备最后在线时间写入 SQLite。

### 2.2 核心文件

| 文件 | 职责 |
|------|------|
| `mqtt/client.go` | EMQX MQTT 客户端连接管理 |
| `mqtt/topic_router.go` | MQTT Topic 路由：`eregen/up/{type}/{id}/message` + `eregen/medical/wb/{ward}/{patient}/status` |
| `mqtt/message_handler.go` | 消息解码、校验、签名验证 |
| `handler/handler.go` | 设备上线/下线事件处理 |
| `nats/publisher.go` | 向 NATS 发布设备事件 |
| `store/store.go` | 设备状态持久化 (SQLite) |

### 2.3 MQTT Topic 定义

| Topic | 方向 | 说明 |
|-------|------|------|
| `eregen/up/#` | 设备→云端 | 设备上行消息 |
| `eregen/down/{dev_id}/command` | 云端→设备 | 下行指令 |
| `eregen/down/{dev_id}/ota` | 云端→设备 | OTA 升级包 |
| `eregen/medical/wb/{ward}/{patient}/status` | 腕带→云端 | 住院患者腕带状态上报 |
| `eregen/medical/wb/{ward}/{patient}/alert` | 腕带→云端 | 腕带风险警示上报 |
| `eregen/community/wb/{dev_id}/up` | 腕带→云端 | 社区老人腕带上行（签到/福利/发药） |
| `eregen/community/wb/{dev_id}/down` | 云端→腕带 | 社区老人腕带下行（配置更新/福利同步） |

### 2.4 消息处理流程

```
MQTT 消息到达
    ↓
Topic 匹配 → 提取 dev_id, type
    ↓
签名验证 (Ed25519)
    ↓
JSON 解码 → 校验字段完整性
    ↓
NATS 发布 → "eregen.device.{type}"
    ↓
SQLite 更新 last_seen (UPDATE devices SET last_seen = {ts} WHERE dev_id = ?)
    ↓
SQLite 写入 (健康/定位数据)
```

---

## 3. api-server — REST API 服务

### 3.1 职责

为家属APP、管理后台、小程序、B2B 系统提供统一的 REST API，处理用户认证、设备管理、健康数据查询、用药配置、告警管理等业务逻辑。

### 3.2 分层架构

```
Handler (HTTP 层)
    ↓ 调用
Service (业务逻辑层)
    ↓ 调用
Store (数据访问层: SQLite)
```

### 3.3 核心文件

**Model:**
| 文件 | 说明 |
|------|------|
| `model/model.go` | 所有数据模型定义 (User, ElderlyProfile, Device, HealthRecord, LocationRecord, MedicationRule, MedStatusRecord, Alert, Subscription) |

**Handler (HTTP 路由处理器):**
| 文件 | 接口前缀 | 说明 |
|------|---------|------|
| `handler/auth.go` | `/api/v1/auth/*` | 登录、注册、OTP 验证 |
| `handler/user.go` | `/api/v1/users/*` | 用户 CRUD、资料更新 |
| `handler/device.go` | `/api/v1/devices/*` | 设备列表、状态、配置更新 |
| `handler/health.go` | `/api/v1/health/*` | 健康数据查询、趋势统计 |
| `handler/location.go` | `/api/v1/location/*` | 实时定位、历史轨迹 |
| `handler/medication.go` | `/api/v1/medication/*` | 用药规则 CRUD、服药记录 |
| `handler/alert.go` | `/api/v1/alerts/*` | 告警列表、处理、WebSocket 推送 |

**Middleware:**
| 文件 | 说明 |
|------|------|
| `middleware/auth.go` | JWT Token 验证、RBAC 权限检查 |

**Service (业务逻辑):**
| 文件 | 说明 |
|------|------|
| `service/services.go` | 用户服务、设备服务、健康服务、告警服务等 |
| `service/nats_client.go` | NATS 客户端，用于向其他微服务发送消息 |

**Store (数据访问):**
| 文件 | 说明 |
|------|------|
| `store/sqlite.go` | SQLite 操作：用户/设备/健康/定位/用药/告警 CRUD |

**Router:**
| 文件 | 说明 |
|------|------|
| `router/router.go` | Gin 路由注册 |
| `router/handlers.go` | Handler 实例化 |
| `router/config.go` | 路由配置 |

### 3.4 医护工作站核心文件（admin-api/medical_wb）

| 文件 | 说明 |
|------|------|
| `admin-api/store/sqlite.go` | 统一 SQLite 连接管理 |
| `admin-api/handler/wb_admission.go` | 入院登记接口 |
| `admin-api/handler/wb_wristband.go` | 腕带绑定/解绑接口 |
| `admin-api/handler/wb_daily_entry.go` | 每日诊疗录入接口 |
| `admin-api/handler/wb_verification.go` | 核验记录查询接口 |
| `admin-api/service/wb_nurse.go` | 护士角色权限、科室管辖校验 |
| `shared/protocol/wb_ble.go` | NFC 协议定义（PDA/手机终端读取腕带） |
| `shared/crypto/wb_aes.go` | 腕带数据传输 AES-128-CBC 加密 |

### 3.4 监管闭环核心文件（admin-api/regulatory）

| 文件 | 说明 |
|------|------|
| `admin-api/handler/regulatory/dashboard.go` | 在院总览 + 患者列表 |
| `admin-api/handler/regulatory/alert.go` | 告警 CRUD + 确认/解决 |
| `admin-api/handler/regulatory/audit.go` | 穿透审计全链路聚合 |
| `admin-api/handler/regulatory/rule_config.go` | 规则配置 CRUD |
| `admin-api/handler/regulatory/compliance.go` | 合规审查报表 |
| `admin-api/service/geofence.go` | 电子围栏计算（Haversine） |
| `admin-api/service/rule_engine.go` | 16条规则定时检测（R01-R08 + R_C01-R_C08） |
| `admin-api/service/audit.go` | 全链路数据聚合查询 |

### 3.5 社区老人腕带核心文件（admin-api/community_wb）

| 文件 | 说明 |
|------|------|
| `admin-api/handler/community_wb/elderly.go` | 老人档案 CRUD |
| `admin-api/handler/community_wb/welfare.go` | 福利标签管理 |
| `admin-api/handler/community_wb/signin.go` | 签到激活 |
| `admin-api/handler/community_wb/pharmacy.go` | 社区药房发药 |
| `admin-api/handler/community_wb/minzheng.go` | 民政数据导入 + 批量发放 |
| `admin-api/service/elderly.go` | 老人业务逻辑 |
| `admin-api/service/welfare.go` | 福利标签同步/匹配 |
| `admin-api/service/signin.go` | 签到周期管理 |
| `admin-api/service/batch_pay.go` | 批量发放引擎 |

### 3.5 关键 API 端点

```
POST   /api/v1/auth/register          # 注册 (OTP 验证)
POST   /api/v1/auth/login             # 登录 (返回 JWT)
POST   /api/v1/auth/otp/send          # 发送 OTP

GET    /api/v1/users/:id              # 获取用户信息
PUT    /api/v1/users/:id              # 更新用户资料

GET    /api/v1/elderly/:id            # 老人档案
PUT    /api/v1/elderly/:id            # 更新老人档案

GET    /api/v1/devices                # 设备列表
GET    /api/v1/devices/:id            # 设备详情
PUT    /api/v1/devices/:id/settings   # 设备配置更新

GET    /api/v1/health?elderly_id=&start=&end=  # 健康数据
GET    /api/v1/health/trend?elderly_id=&metric=hr&days=7  # 趋势

GET    /api/v1/location/latest?elderly_id=       # 最新位置
GET    /api/v1/location/history?elderly_id=&from=&to=  # 历史轨迹

GET    /api/v1/medication/rules?elderly_id=      # 用药规则列表
POST   /api/v1/medication/rules                  # 新增用药规则
PUT    /api/v1/medication/rules/:id              # 更新规则
DELETE /api/v1/medication/rules/:id              # 删除规则

GET    /api/v1/alerts?severity=&status=&elderly_id=  # 告警列表
PUT    /api/v1/alerts/:id/status                 # 处理告警

WS     /api/v1/stream/alerts                     # WebSocket 实时告警

POST   /api/v1/subscriptions/:user_id              # 订阅信息

# 医护工作站 API
POST   /api/v1/medical/patients           # 入院登记
GET    /api/v1/medical/patients           # 患者列表
GET    /api/v1/medical/patients/:id       # 患者详情
PUT    /api/v1/medical/patients/:id       # 更新患者信息
DELETE /api/v1/medical/patients/:id       # 出院注销
POST   /api/v1/medical/patients/batch-import  # 批量导入（Excel/CSV）
GET    /api/v1/medical/patients/by-admission-no  # 按住院号查询
POST   /api/v1/medical/patients/:id/bind      # 腕带绑定
POST   /api/v1/medical/patients/:id/unbind    # 腕带解绑
POST   /api/v1/medical/patients/batch-bind    # 批量绑定
POST   /api/v1/medical/wristbands/:device_id/write  # 写入腕带固件
POST   /api/v1/medical/wristbands/:device_id/clear  # 出院清空腕带
GET    /api/v1/medical/wristbands             # 腕带设备列表
GET    /api/v1/medical/wristbands/:device_id/firmware  # 腕带固件版本
POST   /api/v1/medical/lists/expenses         # 录入费用
GET    /api/v1/medical/lists/expenses         # 查询费用清单
POST   /api/v1/medical/lists/medications      # 录入用药
GET    /api/v1/medical/lists/medications      # 查询用药清单
POST   /api/v1/medical/lists/tests            # 录入检测报告
GET    /api/v1/medical/lists/tests            # 查询检测报告
POST   /api/v1/medical/daily/entries          # 每日诊疗录入
GET    /api/v1/medical/daily/entries          # 诊疗记录列表
GET    /api/v1/medical/history?elderly_id=    # 治疗经过（家属端）
GET    /api/v1/medical/verifications          # 核验记录列表
PUT    /api/v1/medical/verifications/:id/status  # 标记核验完成
GET    /api/v1/medical/verifications/stats/today  # 今日核验统计
GET    /api/v1/medical/stats/overview         # 数据统计看板

# 监管闭环 API（admin-api/regulatory）
GET    /api/v1/admin/regulatory/dashboard/patient-overview  # 在院总览摘要
GET    /api/v1/admin/regulatory/dashboard/patient-list      # 在院患者列表
GET    /api/v1/admin/regulatory/alerts                      # 告警列表
POST   /api/v1/admin/regulatory/alerts                      # 创建告警
POST   /api/v1/admin/regulatory/alerts/:id/acknowledge      # 确认告警
POST   /api/v1/admin/regulatory/alerts/:id/resolve          # 标记解决
GET    /api/v1/admin/regulatory/audit/patient/:id           # 穿透审计全链路
GET    /api/v1/admin/regulatory/rules                       # 规则配置列表
PUT    /api/v1/admin/regulatory/rules                       # 规则配置更新
GET    /api/v1/admin/regulatory/fence/config                # 围栏配置
POST   /api/v1/admin/regulatory/fence/config                # 创建围栏
GET    /api/v1/admin/regulatory/compliance/report           # 合规报表

# 社区老人腕带 API（admin-api/community_wb）
GET    /api/v1/admin/community-wb/elders                    # 老人档案列表
POST   /api/v1/admin/community-wb/elders                    # 创建老人档案
PUT    /api/v1/admin/community-wb/elders/:id                # 更新老人档案
DELETE /api/v1/admin/community-wb/elders/:id                # 删除老人档案
GET    /api/v1/admin/community-wb/devices                   # 腕带设备列表
POST   /api/v1/admin/community-wb/devices                   # 注册腕带设备
GET    /api/v1/admin/community-wb/welfare-tags              # 福利标签配置
POST   /api/v1/admin/community-wb/welfare-tags              # 新增福利标签
PUT    /api/v1/admin/community-wb/welfare-tags/:id          # 更新福利标签
DELETE /api/v1/admin/community-wb/welfare-tags/:id          # 删除福利标签
POST   /api/v1/admin/community-wb/signin/trigger            # 签到激活
GET    /api/v1/admin/community-wb/signin/records            # 签到记录
POST   /api/v1/admin/community-wb/pharmacy/dispense         # 药房发药
POST   /api/v1/admin/community-wb/minzheng/import           # 民政数据导入
POST   /api/v1/admin/community-wb/batch-pay/execute         # 批量发放执行
GET    /api/v1/admin/community-wb/batch-payments            # 发放记录
```

### 3.6 数据模型

```go
// 用户
type User struct {
    ID string `json:"id"`          // UUID
    Email *string `json:"email,omitempty"`
    Phone *string `json:"phone,omitempty"`
    Name string `json:"name"`
    PasswordHash string `json:"-"`
    Role Role `json:"role"`        // family/elderly/institution
}

// 老人档案
type ElderlyProfile struct {
    ID string `json:"id"`
    UserID string `json:"user_id"`
    Name string `json:"name"`
    BirthDate *time.Time `json:"birth_date,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
    HealthTiers []string `json:"health_tiers"`  // 慢病标签: hypertension/diabetes/...
}

// 设备
type Device struct {
    ID string `json:"id"`
    DeviceID string `json:"device_id"`  // BR-XXXX / PX-XXXX
    DeviceType string `json:"device_type"`
    Tier string `json:"tier"`             // starter/plus/pro
    OwnerUserID string `json:"owner_user_id"`
    Status DeviceStatus `json:"status"`    // online/offline
    LastSeen *time.Time `json:"last_seen,omitempty"`
    Settings map[string]any `json:"settings,omitempty"`
}

// 健康记录
type HealthRecord struct {
    ID string `json:"id"`
    ElderlyID string `json:"elderly_id"`
    Timestamp time.Time `json:"timestamp"`
    HR *int `json:"hr,omitempty"`           // 心率 bpm
    SPO2 *int `json:"spo2,omitempty"`      // 血氧 %
    Steps *int64 `json:"steps,omitempty"`  // 步数
    SleepHours *float64 `json:"sleep_hours,omitempty"`
    BPSystolic *int `json:"bp_systolic,omitempty"`
    BPDiastolic *int `json:"bp_diastolic,omitempty"`
}

// 告警
type Alert struct {
    ID string `json:"id"`
    ElderlyID string `json:"elderly_id"`
    AlertType string `json:"alert_type"`  // sos/fall/med_missed/geofence_breach
    Severity AlertSeverity `json:"severity"`  // P0/P1/P2
    Status AlertStatus `json:"status"`      // pending/resolved
    Metadata map[string]any `json:"metadata,omitempty"`
}
```

---

## 4. push-service — 多渠道推送

### 4.1 职责

订阅 NATS 上的告警事件，通过 FCM (Firebase Cloud Messaging)、微信订阅消息、阿里云短信三种渠道将告警推送到家属手机。

### 4.2 核心文件

| 文件 | 职责 |
|------|------|
| `publisher/nats_subscriber.go` | NATS JetStream 订阅设备事件，过滤 sos/fall/med_missed |
| `router/router.go` | Member 管理 + DeliverAlert 扇出分发 |
| `channel/wechat/client.go` | 微信订阅消息推送 |
| `channel/sms/client.go` | 阿里云短信推送 |
| `fcm/client.go` | Firebase Cloud Messaging 推送 |
| `model/model.go` | 推送事件模型 (AlertPushEvent, Member, Severity) |

### 4.3 推送流程

```
NATS 收到事件 (sos/fall/med_missed)
    ↓
构建 AlertPushEvent
    ↓
查找目标家属 Member (UserID → FCM Token + 微信 OpenID + 手机号)
    ↓
并行分发:
  ├→ FCM (主渠道)
  ├→ 微信订阅消息 (备用)
  └→ 短信 (P0 级告警兜底)
    ↓
记录推送结果 → SQLite（推送日志表）
```

### 4.4 告警分级

| 级别 | 类型 | 推送渠道 | 响应要求 |
|------|------|---------|---------|
| P0 | SOS、跌倒 | FCM + 微信 + 短信 | 立即响应 |
| P1 | 漏服药物、电子围栏越界 | FCM + 微信 | 30min 内响应 |
| P2 | 设备离线、低电量 | FCM | 下次打开 APP 可见 |

---

## 5. data-pipeline — AI 数据分析

### 5.1 职责

消费 NATS 上的设备数据流，进行实时健康分析和风险评分，将结果写入 SQLite 供 api-server 查询。

### 5.2 核心文件

| 文件 | 职责 |
|------|------|
| `subscriber/nats_subscriber.go` | 订阅 NATS 设备事件 |
| `analyzer/health_analyzer.go` | 健康数据分析：心率异常、血氧过低、步数骤降 |
| `analyzer/risk_calculator.go` | 风险评分计算：综合多维度数据给出 0-100 风险分 |
| `model/model.go` | 分析结果模型 |
| `store/store.go` | 分析结果持久化 |

### 5.3 分析规则

| 指标 | 正常范围 | 告警阈值 | 级别 |
|------|---------|---------|------|
| 静息心率 | 60-100 bpm | >120 或 <50 | P1 |
| 血氧饱和度 | 95-100% | <90% | P1 |
| 血压收缩压 | 90-140 mmHg | >160 或 <90 | P1 |
| 跌倒置信度 | — | >0.8 | P0 |
| 连续无活动 | — | >2h 无步数 | P2 |
| 夜间心率 | — | >100 持续 30min | P1 |

---

## 6. 数据存储设计

### 6.1 SQLite 表结构

MVP 阶段全项目统一使用 SQLite，零部署。应用启动时自动创建表结构。

```sql
-- 用户表
CREATE TABLE users (
    id TEXT PRIMARY KEY,          -- UUID
    email TEXT,
    phone TEXT,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('family', 'elderly', 'institution', 'nurse', 'admin')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 老人档案表
CREATE TABLE elderly_profiles (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    birth_date DATE,
    avatar_url TEXT,
    health_tiers TEXT DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 设备表（手环 + 药盒 + 医用腕带）
CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    device_id TEXT UNIQUE NOT NULL,  -- BR-XXXX / PX-XXXX / WB-XXXX
    device_type TEXT NOT NULL,       -- bracelet / pillbox / medical_wristband
    tier TEXT NOT NULL CHECK (tier IN ('starter', 'plus', 'pro', 'basic', 'smart', 'auto')),
    owner_user_id TEXT REFERENCES users(id),
    status TEXT DEFAULT 'offline' CHECK (status IN ('online', 'offline')),
    last_seen TIMESTAMP,
    settings TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 健康记录表
CREATE TABLE health_records (
    id TEXT PRIMARY KEY,
    elderly_id TEXT REFERENCES elderly_profiles(id),
    timestamp TIMESTAMP NOT NULL,
    hr INTEGER,
    spo2 INTEGER,
    steps INTEGER,
    temperature REAL,
    bp_systolic INTEGER,
    bp_diastolic INTEGER,
    sleep_hours REAL
);

-- 定位记录表
CREATE TABLE location_records (
    id TEXT PRIMARY KEY,
    elderly_id TEXT REFERENCES elderly_profiles(id),
    timestamp TIMESTAMP NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    accuracy REAL
);

-- 告警表
CREATE TABLE alerts (
    id TEXT PRIMARY KEY,
    elderly_id TEXT REFERENCES elderly_profiles(id),
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('P0', 'P1', 'P2')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'resolved')),
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- 推送日志表
CREATE TABLE push_logs (
    id TEXT PRIMARY KEY,
    alert_id TEXT REFERENCES alerts(id),
    channel TEXT NOT NULL,           -- fcm / wechat / sms
    status TEXT NOT NULL,            -- sent / failed
    detail TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 分析结果表
CREATE TABLE analysis_results (
    id TEXT PRIMARY KEY,
    elderly_id TEXT REFERENCES elderly_profiles(id),
    metric TEXT NOT NULL,
    value REAL,
    risk_score INTEGER,
    alert_triggered INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 6.2 医用腕带 SQLite 表结构

```sql
-- 住院患者表
CREATE TABLE wb_patients (
    id TEXT PRIMARY KEY,
    hospital_id TEXT,                -- B2B 医院 ID
    patient_name TEXT NOT NULL,
    gender TEXT CHECK (gender IN ('male', 'female')),
    birth_date DATE,
    ward TEXT NOT NULL,              -- 科室+床号，如 "心内科3床"
    bed_no TEXT,
    diagnosis TEXT,
    attending_doctor TEXT,
    department TEXT NOT NULL,        -- cardiology / neurology / ...
    admission_date TIMESTAMP NOT NULL,
    discharge_date TIMESTAMP,
    status TEXT DEFAULT 'admitted' CHECK (status IN ('admitted', 'discharged', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 医用腕带设备表
CREATE TABLE wb_devices (
    id TEXT PRIMARY KEY,
    wristband_id TEXT UNIQUE NOT NULL,  -- WB-XXXX
    device_id TEXT NOT NULL,            -- ESP32-S3 MAC 或 UUID
    firmware_version TEXT,
    battery_pct INTEGER,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'expired')),
    assigned_patient_id TEXT REFERENCES wb_patients(id),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 腕带绑定关系表
CREATE TABLE wb_bindings (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    wristband_id TEXT NOT NULL REFERENCES wb_devices(wristband_id),
    bound_by TEXT NOT NULL,             -- 护士工号
    bound_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    unbound_at TIMESTAMP,
    unbound_reason TEXT,
    UNIQUE(patient_id, wristband_id)
);

-- 费用记录表
CREATE TABLE wb_expenses (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    expense_type TEXT NOT NULL,         -- registration / medication / test / surgery / discharge
    amount REAL NOT NULL,
    description TEXT,
    recorded_by TEXT,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用药记录表
CREATE TABLE wb_medications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    medication_name TEXT NOT NULL,
    dosage TEXT NOT NULL,
    schedule_time TEXT,               -- "08:00"
    frequency TEXT,                   -- once / bid / tid / qid / prn
    route TEXT,                       -- oral / iv / im / topical
    prescribed_by TEXT,               -- 医生
    given_at TIMESTAMP,
    given_by TEXT,                    -- 护士工号
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 检验检查记录表
CREATE TABLE wb_test_results (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    test_type TEXT NOT NULL,          -- blood / urine / ecg / imaging / ...
    test_name TEXT NOT NULL,
    ordered_by TEXT,                  -- 医生
    ordered_at TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,
    abnormal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 每日诊疗录入表
CREATE TABLE wb_daily_entries (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    entry_date DATE NOT NULL,
    entry_type TEXT NOT NULL,         -- rounds / medication / infusion / test / surgery / discharge
    content TEXT NOT NULL,
    nurse_id TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 近场核验记录表
CREATE TABLE wb_verifications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    wristband_id TEXT NOT NULL,
    nurse_id TEXT NOT NULL,
    action TEXT NOT NULL,             -- check_in / give_medication / infusion / blood_draw / transfusion / test / surgery / discharge
    matched BOOLEAN NOT NULL,         -- 核验是否一致
    match_detail TEXT,               -- JSON: 比对字段详情
    terminal_type TEXT NOT NULL,     -- pda / phone
    nfc_rssi INTEGER,                -- NFC 读取信号强度
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 风险警示标签配置表
CREATE TABLE wb_alert_tag_config (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES wb_patients(id),
    tag_type TEXT NOT NULL,          -- fall_risk / infection_risk / allergy / isolation / pressure_ulcer
    tag_color TEXT,                  -- 腕带颜色编码
    description TEXT,
    configured_by TEXT,
    active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 6.3 索引设计

```sql
-- 查询性能优化索引
CREATE INDEX idx_health_records_elderly_ts ON health_records(elderly_id, timestamp);
CREATE INDEX idx_location_records_elderly_ts ON location_records(elderly_id, timestamp);
CREATE INDEX idx_alerts_elderly_status ON alerts(elderly_id, status);
CREATE INDEX idx_wb_patients_status ON wb_patients(status);
CREATE INDEX idx_wb_patients_ward ON wb_patients(ward);
CREATE INDEX idx_wb_verifications_patient_ts ON wb_verifications(patient_id, timestamp);
CREATE INDEX idx_wb_daily_entries_patient_date ON wb_daily_entries(patient_id, entry_date);
CREATE INDEX idx_wb_medications_patient_given ON wb_medications(patient_id, given_at);
CREATE INDEX idx_push_logs_alert_id ON push_logs(alert_id);
CREATE INDEX idx_analysis_results_elderly_ts ON analysis_results(elderly_id, created_at);
```

---

## 7. 编译与运行

### 7.1 环境要求

```bash
go version >= 1.22
docker compose >= 2.20
```

### 7.2 启动所有云服务

```bash
# 1. 启动基础设施
docker compose -f docker-compose.dev.yml up -d

# 2. 启动 gateway (端口 1883 MQTT, 8080 HTTP)
cd cloud/gateway && go run ./cmd

# 3. 启动 api-server (端口 8081)
cd ../api-server && go run ./cmd

# 4. 启动 push-service
cd ../push-service && go run ./cmd

# 5. 启动 data-pipeline
cd ../data-pipeline && go run ./cmd

# 6. 启动管理后台 + 医护工作站（admin-api，端口 8080）
cd ../admin-api && go run ./cmd
```

### 7.3 单元测试

```bash
# 共享模块
cd shared/crypto && go test ./...
cd shared/protocol && go test ./...

# 各微服务
cd cloud/gateway && go test ./...
cd cloud/api-server && go test ./...
cd cloud/push-service && go test ./...
cd cloud/data-pipeline && go test ./...

# admin-api / medical_wb
cd cloud/admin-api && go test ./...

# admin-api / regulatory + community_wb
cd cloud/admin-api && go test ./internal/handler/regulatory/...
cd cloud/admin-api && go test ./internal/service/rule_engine/...
cd cloud/admin-api && go test ./internal/handler/community_wb/...

# gateway / community handlers
cd cloud/gateway && go test ./internal/handler/...
cd cloud/gateway && go test ./internal/nats/...
```

### 7.4 统一启动脚本系统

根目录 `scripts/start.sh` 提供统一的服务管理入口，每个子系统目录下有独立的 `start.sh` 调用 `scripts/lib/` 中的通用模块。

```bash
# 首次配置（复制默认端口）
cp scripts/default-ports.env .env

# 可用命令
./scripts/start.sh start <service|--all|--group>    # 启动服务
./scripts/start.sh stop <service|--all|--group>     # 停止服务
./scripts/start.sh restart <service>                # 重启服务
./scripts/start.sh status [--all]                    # 查看状态
./scripts/start.sh logs <service|--all>              # 查看日志
./scripts/start.sh clean                             # 清理运行时文件
./scripts/start.sh check-deps                        # 依赖检查
./scripts/start.sh ports-check                       # 端口冲突检测
./scripts/start.sh start --docker                    # Docker 模式
```

**目录结构：**
```
.env                              # 用户配置（从 default-ports.env 复制）
scripts/
  default-ports.env               # 参考默认值（只读）
  start.sh                        # 根聚合入口
  lib/
    common.sh                     # 共享函数：env加载、端口检查、颜色、PID管理
    go-service.sh                 # Go服务：go run + PORT env + 健康检查
    flutter-app.sh                # Flutter：flutter run
    vue-app.sh                    # Vue：npm run dev
    hugo-site.sh                  # Hugo：hugo server
    firmware.sh                   # 固件：idf.py build / cmake + make
    docker-compose.sh             # Docker：docker compose up/down
```

**端口配置格式（`.env`）：**
```env
PORT_GATEWAY=8081
PORT_API_SERVER=8080
PORT_PUSH_SERVICE=8085
PORT_DATA_PIPELINE=8087
PORT_ADMIN_API=8089
PORT_HOSPITAL_API=8082
PORT_COMMUNITY_PLATFORM=8083
PORT_INSURANCE_INTEGRATION=8084
PORT_ADMIN_WEB=3001
PORT_WEBSITE=1313
PORT_POSTGRES=5432
PORT_REDIS=6379
PORT_INFLUXDB=8086
PORT_EMQX_MQTT=1883
PORT_EMQX_DASHBOARD=18083
PORT_NATS=4222
PORT_PROMETHEUS=9090
PORT_GRAFANA=3002
PORT_LOKI=3100
```

CLI 覆盖：`./scripts/start.sh start gateway --port 9081` 覆盖 `PORT_GATEWAY`。

**按组启动：** `cloud` (api-server, push-service, data-pipeline, admin-api, gateway) / `b2b` (hospital-api, community-platform, insurance-integration) / `apps` (family-app, admin-web, website) / `firmware` (bracelet/pillbox)

**错误处理：** 缺少 `.env` 时使用硬编码默认值并警告；端口冲突时中止启动并显示冲突服务；PID 文件过期（进程已死）时自动清理；依赖缺失时跳过该服务组并警告。

---

## 8. 安全设计

### 8.1 基础安全机制

| 机制 | 实现 |
|------|------|
| 设备认证 | Ed25519 公钥签名验证 |
| 数据传输加密 | TLS 1.3 (外网) + AES-256-GCM (Payload) |
| API 认证 | JWT Bearer Token + Refresh Token |
| 权限控制 | RBAC: family/elderly/institution/admin/nurse/regulator |
| 密码存储 | SHA-256 + Salt (生产建议 bcrypt/argon2) |
| 设备令牌 | dt_ 前缀 Base64URL 随机串 |
| 配网码 | Base36 编码 4-6 位短码 |

### 8.2 JWT 验证落地

**现状：** `cloud/api-server/internal/middleware/auth.go` 和 `cloud/admin-api/internal/middleware/auth.go` 都只有 token 提取，未验证签名。

- **api-server JWT 中间件** — 解析 JWT 中的 user_id + role，注入到 gin.Context，支持三种角色路由（family/elderly/institution）
- **admin-api JWT 中间件** — 解析 JWT 中的 admin_role，支持 super_admin/operator 分级访问控制
- **shared/crypto JWT 工具** — 提取 JWT 生成/验证为共享模块，避免各服务重复实现

### 8.3 输入校验

- **shared/validation 包** — 通用校验函数（email 格式、手机号、经纬度范围、分页参数限制等）
- **api-server 请求结构体验证** — 所有 request struct 添加 `binding:"required"` 标签 + 自定义 validator
- **admin-api 参数验证** — page/page_size 上限限制（page_size ≤ 100），severity/status 白名单过滤

### 8.4 API 限流

- **shared/ratelimit 包** — 基于 Redis 的令牌桶算法，支持 per-user 和 per-IP 两种模式
- **api-server 限流中间件** — 默认 100 req/min per user，认证用户 500 req/min
- **admin-api 限流中间件** — 更严格：30 req/min per admin
- **B2B 服务独立限流** — 按 API Key 限流，1000 req/min

### 8.5 shared/crypto 模块扩展

- **AES-256-GCM 封装** — 健康数据加密/解密（用于存储到 InfluxDB 前）
- **Ed25519 设备签名** — 设备固件签名验证（OTA 升级时）
- **密码哈希** — bcrypt + argon2 双算法支持（新用户 argon2，迁移用户 bcrypt）
- **TLS 配置生成器** — 根据环境生成安全的 TLS 配置（minVersion TLS 1.3）

### 8.6 敏感数据脱敏

- **shared/sanitize 包** — 响应过滤器，自动脱敏 email（x@***.com）、手机号（138****5678）
- **Gin 中间件** — 在返回响应前自动脱敏
- **日志脱敏** — 所有 ESP_LOG / log.Printf 不输出完整 token 或密码

### 8.7 安全架构分层

```
┌─────────────────────────────────────────────────────┐
│                   Gin Router                         │
├─────────────────────────────────────────────────────┤
│  RateLimit Middleware (shared/ratelimit)             │
│  ↓                                                   │
│  Auth Middleware (JWT 验证)                          │
│  ↓                                                   │
│  Sanitize Middleware (响应脱敏)                       │
│  ↓                                                   │
│  Handler                                             │
│  ↓                                                   │
│  Service Layer                                       │
│  ↓                                                   │
│  Store Layer                                         │
└─────────────────────────────────────────────────────┘

共享模块:
  shared/crypto    → AES-256-GCM, Ed25519, password hash, TLS config
  shared/validation → 字段校验函数库
  shared/ratelimit → Redis 令牌桶限流
  shared/sanitize  → PII 脱敏过滤器
```

### 8.8 硬件 MQTT 集成 — 双向通信

**手环固件（三档）MQTT 双向通信：**

- CommTask 需同时发布和订阅 MQTT：
  - 上行 topic: `eregen/device/bracelet/{dev_id}/up` — heartbeat/location/health/sos/fall/geofence_breach
  - 下行 topic: `eregen/command/{dev_id}` — config/med_rule/tts/ota
- 启动时发送 device_info: `{"type":"device_info","dev_id":"BR-XXXX","fw_version":"1.0.0","tier":"entry","hw_revision":"1"}`
- Plus/Pro tier 补充: geofence alert publishing, fall alert publishing

**药盒固件生产配置：**
- Smart tier: 替换 localhost:1883 为生产 EMQX broker URL + TLS
- Auto tier: 实现 med_status 实际发布（光电传感器检测到药片分配时上报）
- Common WiFi MQTT Bridge: 确保 status topic publish 由 gateway 处理，添加命令订阅匹配 gateway 下行 topic 模式

**医疗设备腕带固件（ESP32-S3）：**
- NFC NDEF Service (per spec §5.4):
  - NDEF Record Types: application/vnd.eregen.patient, application/vnd.eregen.verification, application/vnd.eregen.status
  - Protocol: NFC-A 106kbps, NDEF payload with challenge-response auth
- WiFi MQTT 客户端（医院部署腕带）: ESP-MQTT 连接 EMQX over TLS
  - Publish: patient_register, device_status, alert_tag
  - Subscribe: eregen/command/WB-XXXX

**加密 Payload 路径：**
- 手环固件: `message_encode.c` 添加加密变体，使用 `payload_crypto.c`，AES-128-CTR + HMAC-SHA256
- 调用 `mqtt_common_publish_encrypted()` 替代明文 `mqtt_common_publish()`
- API server (`nats_client.go`) 已有双路径解析（明文先试，然后加密回退）— 确认正常工作

**Gateway 增强：**
- 处理 `device_info` 消息类型 → 提取 device type/fw_version/tier → 更新 DB 设备记录
- 处理 `ota_progress` → 订阅 `eregen/device/+/ota_progress` → 转发到 NATS `eregen.event.ota_progress`
- 处理 `geofence_breach` → 创建 Redis 告警缓存 → 发布到 NATS `eregen.event.geofence_breach`

---

© 2026 Eregen (颐贞). All rights reserved.
