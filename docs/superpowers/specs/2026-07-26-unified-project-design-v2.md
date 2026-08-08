# 颐贞 Eregen — 统一项目设计方案 v2.0

> 编制日期：2026-07-26  
> 版本：v2.0（融合 AnaReports 三篇设计文档 + 用户确认决策）  
> 状态：最终方案  
> 项目性质：技术预研可行性开发 — 研究功能可行性和系统落地

---

## 第一部分 项目全景

### 1.1 品牌与愿景

| 维度 | 内容 |
|------|------|
| **品牌** | 颐贞 (yí zhēn) / Eregen (Ease + Regen) |
| **Slogan** | "颐养正道，贞守安康" |
| **目标人群** | 60-85岁老人(使用者) + 40-55岁子女(购买者) |
| **项目性质** | 技术预研可行性开发 — 验证功能可行性和系统落地 |

### 1.2 八大子系统全景

| # | 子系统 | 语言/框架 | 端口 | 状态 |
|---|--------|----------|------|------|
| ① | 手环固件（三档） | C + FreeRTOS | — | 现有 |
| ② | 药盒固件（三档） | C + ESP-IDF | — | 现有 |
| ③ | 云平台后端 | Go + Gin | 8080/8081/8082 | 现有 + 医用增量 |
| ④ | 家属APP | Flutter | 5173(Web) | 现有 + 住院治疗增量 |
| ⑤ | 管理后台 | Vue 3 + Element Plus | 5174 | 现有 + 医护工作站增量 |
| ⑥ | 微信小程序 | 原生WXML/WXSS | — | 现有 + 增量 |
| ⑦ | 品牌官网 | Hugo + Tailwind CSS | 1313 | 现有 |
| ⑧ | B2B对接层 | Go | — | 现有 |

### 1.3 三档产品线矩阵

| | 入门版 (Starter) | 中端版 (Plus) | 高端版 (Pro) |
|---|---|---|---|
| **手环** | 心率+血氧+SOS+GPS定位 | +电子围栏+跌倒检测+长续航 | +ECG心电+AMOLED屏+金属机身+高精度GPS |
| **药盒** | 基础分格药盒(无电子功能) | 定时语音提醒+APP联动 | 自动分药+光电检测+库存预警+TTS播报 |
| **医用腕带** | — | — | ESP32-S3 + NFC近场认证 + Cat1上传 |
| **社区腕带** | — | — | ESP32-S3 + NFC近场认证 + Cat1上传 |
| **医用腕带BOM** | ~180元 | — | ~350元 |
| **社区腕带BOM** | ~180元 | — | ~350元 |

### 1.4 核心原则

1. **零破坏原则** — 所有新增内容为增量开发，禁止修改、删减、重构原有项目架构
2. **轻量 MVP 原则** — 当前不做微服务拆分、分布式、高可用、集群架构设计
3. **可演进原则** — 所有设计面向未来商业化扩展，预留架构升级能力
4. **腕带是终端介质** — 医用/社区腕带不存储诊疗数据，仅存储身份标识和基础体征
5. **零误差通信** — 扫描A绝不显示B的信息，医院场景绝对要求
6. **全中文界面** — 所有子系统UI/提示/API响应强制使用中文
7. **SQLite优先** — MVP阶段全项目使用SQLite，保证可迁移到PostgreSQL

---

## 第二部分 整体架构设计

### 2.1 新旧功能隔离策略

```
┌─────────────────────────────────────────────────────────────┐
│                    Eregen 颐贞平台                           │
│                                                             │
│  ┌──────────────────┐   ┌──────────────────────────────┐   │
│  │  现有业务（不动）   │   │  新增医用/社区腕带（增量）     │   │
│  │                  │   │                              │   │
│  │  手环 Entry/Plus/Pro│  │  医用腕带 Hospital Mode     │   │
│  │  药盒 Basic/Smart │  │  社区腕带 Community Mode    │   │
│  │  云平台核心模块    │  │  （同构ESP32-S3 + Cat1）     │   │
│  │  家属APP/管理后台/小程序 │ │ 模式区分：mode=hospital/community │
│  │  B2B 对接        │  │                              │   │
│  └──────────────────┘   └──────────────────────────────┘   │
│                                                             │
│  共享基础设施：                                              │
│  ├── SQLite 主库（现有表 + 医用表 + 社区表）                 │
│  ├── EMQX MQTT Broker（命名空间隔离）                       │
│  │   ├── eregen/up/#          — 现有手环/药盒上行           │
│  │   ├── eregen/down/#       — 现有下行                     │
│  │   ├── eregen/medical/wb/#  — 医用腕带（住院模式）         │
│  │   └── eregen/community/wb/# — 社区腕带（社区模式）        │
│  └── NATS JetStream（事件分区）                             │
│      ├── eregen.device.*      — 现有设备事件                │
│      ├── eregen.medical.wb.*  — 医用腕带事件                │
│      └── eregen.community.wb.* — 社区腕带事件               │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 腕带模式区分

医用腕带和社区腕带使用**同一硬件平台**（ESP32-S3 + Cat1），通过云端配置区分模式：

| 维度 | 住院模式 (hospital) | 社区模式 (community) |
|------|---------------------|---------------------|
| **使用场景** | 医院住院患者 | 社区老人 |
| **本地存储** | 患者身份标识（姓名、住院号、科室、床号）+ 警示标签 | 老人身份标识（姓名、身份证号、福利标签）+ 警示标签 |
| **上传数据** | 基础体征（心跳、血氧）+ 位置（基站定位） | 基础体征（心跳、血氧）+ 位置 |
| **近场认证** | NFC贴近护士终端 → 读取患者ID → 云端验证 → 返回诊疗记录 | NFC贴近社区医院终端 → 读取老人ID → 云端验证 → 返回档案+福利 |
| **云端Topic** | `eregen/medical/wb/{dev_id}/#` | `eregen/community/wb/{dev_id}/#` |
| **BLE GATT Service** | hospital专属UUID | community专属UUID |
| **OLED显示** | 患者姓名+住院号+科室床号+警示标签 | 老人姓名+福利标签摘要+签到状态 |
| **LED警示** | 过敏(红)/跌倒高危(橙)/隔离(蓝) | 已签到(绿)/未签到(黄)/过期(红) |
| **数据归属** | medical_wristband_patients 表族 | community_elders 表族 |

**关键设计：** 腕带本地只存储身份标识，不存储诊疗数据。所有医疗数据来自HIS系统API同步或人工录入，存储在云端SQLite中。

### 2.3 数据流闭环

#### 住院场景数据流

```
1. HIS系统API自动同步患者信息 → 云端SQLite
   ↓
2. 护士站后台人工补充/确认 → 写入 medical_wristband_patients
   ↓
3. 绑定腕带设备 → 写入患者身份标识到腕带（仅姓名+住院号+科室+床号+警示标签）
   ↓
4. 护士用终端NFC贴近腕带 → 读取患者ID → 云端验证 → 返回诊疗记录
   ↓
5. 医院系统录入诊疗数据（用药/检查/治疗/费用）→ 云端SQLite
   ↓
6. 家属APP通过HTTPS API查看该患者的诊疗信息
   ↓
7. 腕带基础体征数据 → Cat1蜂窝 → MQTT → 云端SQLite
```

#### 社区场景数据流

```
1. 民政/残联/医保局批量导入数据 → 云端SQLite
   ↓
2. 社区医院为老人办理腕带绑定 → 写入老人身份标识到腕带
   ↓
3. 老人持腕带到社区医院 → NFC扫描 → 读取老人ID → 云端验证
   ↓
4. 药师执行特病认证 + 发药 → 记录发药日志
   ↓
5. 签到激活 → 自动激活本期所有福利标签
   ↓
6. 批量发放引擎 → 每月固定日期批量打款
   ↓
7. 家属APP查看父母福利状态 + 补助领取记录
   ↓
8. 腕带基础体征数据 → Cat1蜂窝 → MQTT → 云端SQLite
```

### 2.4 跨子系统依赖关系

```
                    ┌─────────────┐
                    │  腕带固件    │
                    │ (hospital/  │
                    │  community) │
                    └──────┬──────┘
                           │ Cat1 MQTT
                           ▼
                    ┌─────────────┐     ┌──────────────┐
                    │  EMQX MQTT  │────▶│  云平台 admin-api │
                    └─────────────┘     │  (SQLite)      │
                                        └───┬────┬───────┘
                                            │    │
                                    ┌───────┘    └───────┐
                                    ▼                   ▼
                              ┌──────────┐       ┌──────────┐
                              │ 家属APP   │       │ 管理后台   │
                              │ (Flutter) │       │ (Vue 3)  │
                              └──────────┘       └──────────┘
                                    │                   │
                                    ▼                   ▼
                              ┌──────────┐       ┌──────────┐
                              │ 微信小程序 │       │ 监管专区   │
                              └──────────┘       └──────────┘
```

---

## 第三部分 腕带硬件与固件设计

### 3.1 腕带硬件规格

| 组件 | 选型 | 说明 |
|------|------|------|
| **MCU** | ESP32-S3 | 原生支持BLE 5.0 + Wi-Fi，FreeRTOS |
| **通信模组** | Cat1（广和通L610-CM） | 与手环共用，上行基础体征数据 |
| **近场认证** | NFC（NTAG 780） | AES加密，4cm作用距离，<50ms响应 |
| **显示** | SSD1306 OLED 0.96" | 显示身份标识+警示/福利标签 |
| **LED指示** | RGB LED | 警示颜色/福利状态可视化 |
| **传感器** | PPG（汇顶GT320）+ IMU（ICM-42670-P） | 心跳、血氧、步数、跌倒检测 |
| **电源** | 350mAh LiPo | 续航目标：7天 |
| **OTA** | Cat1远程升级 | SHA-256签名验证 |
| **BOM成本** | ~180元（基础）/ ~350元（含GPS） | — |

### 3.2 腕带本地存储内容（仅限身份标识）

| 字段 | 长度 | 说明 |
|------|------|------|
| 姓名 | 64字节 | 患者/老人姓名 |
| 住院号/身份证号 | 32字节 | 唯一核心标识 |
| 科室/社区医院 | 32字节 | 所属机构 |
| 床号/编号 | 16字节 | 床位或编号 |
| 警示标签 | 64字节 | JSON数组代码 |
| 病历摘要 | 256字节 | 简要治疗方案 |

**明确不存储的内容：**
- 用药清单（来自HIS系统，云端存储）
- 检测报告（来自HIS系统，云端存储）
- 费用清单（来自HIS系统，云端存储）
- 每日查房记录（云端存储）
- 核验记录（云端存储）

### 3.3 NFC近场认证方案

**为什么选择NFC而非BLE：**

| 维度 | NFC | BLE |
|------|-----|-----|
| 作用距离 | 4cm（物理限制，零误扫） | 10m（需软件控制范围） |
| 响应时间 | <50ms | 1-3秒 |
| 并发干扰 | 零（一次只能贴一个） | 多设备同时连接可能冲突 |
| 功耗 | 被动式（几乎为零） | 主动式（待机电流μA级） |
| 安全性 | 物理接触级别 | 配对码+加密 |
| 医疗合规 | 符合（零误读风险） | 符合（但需额外校验） |

**NFC认证流程：**

```
1. 护士/药师将终端靠近腕带（<4cm）
2. NFC芯片读取患者ID哈希（6字节）
3. 终端通过HTTPS请求云端验证
4. 云端返回完整身份信息 + 诊疗记录/福利信息
5. 终端展示结果并记录操作日志
```

**腕带端NFC安全机制：**

```
NTAG 780芯片:
  ├── 存储区A: 患者身份标识（AES-256加密）
  ├── 存储区B: 警示标签代码（明文，已脱敏）
  └── 认证: NTAG密码保护（PWD + ACK）
```

### 3.4 固件目录结构

```
firmware/
├── bracelet/                        # 【不动】通用手环
├── pillbox/                         # 【不动】药盒
└── medical-wristband/               # 【新增】医用/社区腕带
    ├── esp32s3/                     #   住院模式固件
    │   ├── main/
    │   │   ├── patient_store.c/h        # 患者信息 NVS/Flash 存储
    │   │   ├── nfc_reader.c/h           # NFC近场读取
    │   │   ├── cat1_mqtt.c/h            # Cat1联网（上行体征数据）
    │   │   ├── display_oled.c/h         # OLED显示
    │   │   ├── led_indicator.c/h        # LED警示灯
    │   │   ├── nvs_manager.c/h          # 非易失存储管理
    │   │   ├── security.c/h             # AES-256加密/SHA-256签名
    │   │   └── board_init.c/h           # 板级初始化
    │   ├── protocol/
    │   │   ├── medical_protocol.h       # 医用消息协议
    │   │   ├── message_encode.c/h       # JSON编码
    │   │   └── message_decode.c/h       # JSON解码
    │   ├── ota/                       # OTA更新
    │   ├── crypto/                    # 加密（AES-GCM/SHA-256）
    │   ├── test/                      # 单元测试
    │   └── CMakeLists.txt
    │
    └── community/                   #   社区模式固件（独立分支）
        ├── main/
        │   ├── elder_store.c/h          # 老人档案 NVS/Flash 存储
        │   ├── nfc_reader.c/h           # NFC近场读取（复用）
        │   ├── cat1_mqtt.c/h            # Cat1联网（复用）
        │   ├── display_oled.c/h         # OLED显示（福利标签）
        │   ├── led_indicator.c/h        # LED状态灯（签到状态）
        │   ├── nvs_manager.c/h          # 复用
        │   ├── security.c/h             # 复用
        │   └── board_init.c/h           # 复用
        ├── protocol/
        │   ├── community_protocol.h     # 社区消息协议
        │   └── message_encode.c/h       # 复用
        ├── ota/                       # OTA更新（复用hospital）
        ├── crypto/                    # 加密（复用hospital）
        └── CMakeLists.txt
```

### 3.5 固件版本管理

| 产品线 | 固件版本 | 编译产物 |
|--------|---------|---------|
| 手环 Entry | 1.2.0 | `bracelet_entry.bin` |
| 手环 Plus | 1.3.1 | `bracelet_plus.bin` |
| 手环 Pro | 2.0.0 | `bracelet_pro.bin` |
| 医用腕带 hospital | 1.0.0 | `medical_wristband.bin` |
| 医用腕带 community | 1.0.0 | `community_wristband.bin` |
| 药盒 Smart | 1.1.0 | `pillbox_smart.bin` |

---

## 第四部分 数据库设计

### 4.1 数据库策略

| 阶段 | 数据库 | 说明 |
|------|--------|------|
| **MVP** | SQLite（单文件 `eregen.db`） | 零部署、零运维、单进程 |
| **生产** | PostgreSQL 16.x | 多用户并发、高可用 |
| **时序数据** | InfluxDB 2.x | 健康数据时序查询 |
| **缓存** | Redis 7.x | 在线状态、OTP、Token |

**迁移路径：** 所有Store接口定义在 `cloud/admin-api/internal/store/store.go`，业务逻辑通过接口访问。SQLite → PostgreSQL 只需替换Store实现。

### 4.2 数据表总览

```
eregen.db (SQLite 主库)
├── [现有表] users, elderly_profiles, devices, alerts, subscriptions, ...
│
├── [医用腕带表族] medical_wristband_*
│   ├── medical_wristband_patients          — 患者档案
│   ├── medical_wristband_devices           — 腕带设备
│   ├── medical_wristband_bindings          — 患者-腕带绑定
│   ├── medical_expenses                    — 费用清单
│   ├── medical_medications                 — 用药清单
│   ├── medical_test_results                — 检测报告
│   ├── medical_daily_entries               — 每日诊疗录入
│   ├── medical_verifications               — 核验记录
│   └── medical_alert_tag_config            — 警示标签配置
│
├── [监管表族] regulatory_*
│   ├── regulatory_fence_config             — 电子围栏配置
│   ├── regulatory_location_logs            — 位置日志
│   └── regulatory_alerts                   — 监管告警
│
├── [社区腕带表族] community_*
│   ├── community_elders                    — 老人档案
│   ├── community_wristband_devices         — 腕带设备
│   ├── community_elder_bindings            — 老人-腕带绑定
│   ├── community_welfare_tag_config        — 福利标签配置
│   ├── community_elder_welfare             — 老人-福利绑定
│   ├── community_signin_records            — 签到记录
│   ├── community_pharmacy_logs             — 药房发药记录
│   ├── community_minzheng_sync             — 民政数据同步
│   └── community_batch_payments            — 批量发放记录
│
└── [角色扩展]
    ├── users.role                          — super_admin, hospital_admin, nurse, regulator, family_member
    └── user_department_bindings            — 用户-科室权限
```

### 4.3 核心表SQL（医用腕带）

```sql
-- 患者表
CREATE TABLE medical_wristband_patients (
    id TEXT PRIMARY KEY,
    hospital_id TEXT,              -- 医院ID（预留HIS对接）
    name TEXT NOT NULL,
    gender INTEGER NOT NULL CHECK (gender IN (0, 1, 2)),
    age INTEGER,
    admission_no TEXT UNIQUE NOT NULL,  -- 住院号（唯一核心标识）
    department TEXT,
    bed_no TEXT,
    blood_type TEXT,
    allergy_history TEXT,
    special_disease TEXT,
    alert_tags TEXT DEFAULT '[]',  -- JSON: ["allergy","fall_risk"]
    medical_summary TEXT,          -- 腕带本地存储的病历摘要
    status TEXT DEFAULT 'admitted' CHECK (status IN ('admitted','discharged','dead')),
    last_verify_at DATETIME,
    verify_gap_hours INTEGER DEFAULT 0,
    fence_status TEXT DEFAULT 'inside' CHECK (fence_status IN ('inside','outside','unknown')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 腕带设备表
CREATE TABLE medical_wristband_devices (
    id TEXT PRIMARY KEY,
    device_id TEXT UNIQUE NOT NULL, -- WB-XXXX 格式
    mode TEXT DEFAULT 'hospital' CHECK (mode IN ('hospital','community')),
    firmware_version TEXT,
    status TEXT DEFAULT 'active' CHECK (status IN ('active','inactive','retired')),
    last_seen DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 患者-腕带绑定
CREATE TABLE medical_wristband_bindings (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    device_id TEXT NOT NULL REFERENCES medical_wristband_devices(id),
    bound_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    unbound_at DATETIME,
    UNIQUE(patient_id, device_id)
);

-- 费用清单
CREATE TABLE medical_expenses (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    item_name TEXT NOT NULL,
    item_type TEXT CHECK (item_type IN ('drug','equipment','test','service','other')),
    amount REAL DEFAULT 0,
    quantity INTEGER DEFAULT 1,
    unit_price REAL,
    recorded_date DATE,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 用药清单
CREATE TABLE medical_medications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    drug_name TEXT NOT NULL,
    dosage TEXT,
    frequency TEXT,
    route TEXT CHECK (route IN ('oral','iv','im','sc','other')),
    start_date DATE,
    end_date DATE,
    status TEXT DEFAULT 'active' CHECK (status IN ('active','discontinued','completed')),
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 检测报告
CREATE TABLE medical_test_results (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    test_name TEXT NOT NULL,
    test_date DATE,
    result_text TEXT,
    result_value TEXT,
    reference_range TEXT,
    abnormal_flag INTEGER DEFAULT 0,
    attachment_url TEXT,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 每日诊疗录入
CREATE TABLE medical_daily_entries (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    entry_date DATE NOT NULL,
    entry_type TEXT CHECK (entry_type IN ('round','nursing','order','other')),
    content TEXT NOT NULL,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 核验记录
CREATE TABLE medical_verifications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    device_id TEXT NOT NULL,
    nurse_user_id TEXT,
    action TEXT NOT NULL CHECK (action IN (
        'check_in','check_out','give_medication',
        'infusion','blood_draw','transfusion',
        'test','surgery','discharge'
    )),
    verified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_data TEXT,                -- JSON: 腕带数据快照
    result TEXT CHECK (result IN ('success','mismatch','error')),
    notes TEXT
);

-- 警示标签配置
CREATE TABLE medical_alert_tag_config (
    id TEXT PRIMARY KEY,
    tag_code TEXT UNIQUE NOT NULL,
    tag_name TEXT NOT NULL,
    color TEXT,
    icon TEXT,
    severity TEXT DEFAULT 'P2' CHECK (severity IN ('P0','P1','P2')),
    enabled INTEGER DEFAULT 1
);
```

### 4.4 核心表SQL（监管）

```sql
-- 电子围栏配置
CREATE TABLE regulatory_fence_config (
    id TEXT PRIMARY KEY,
    hospital_id TEXT NOT NULL,
    hospital_name TEXT NOT NULL,
    center_lat REAL NOT NULL,
    center_lng REAL NOT NULL,
    radius_meters INTEGER DEFAULT 200,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(hospital_id)
);

-- 位置日志
CREATE TABLE regulatory_location_logs (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    device_id TEXT NOT NULL,
    lat REAL NOT NULL,
    lng REAL NOT NULL,
    accuracy REAL,
    inside_fence INTEGER DEFAULT 1 CHECK (inside_fence IN (0, 1)),
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 监管告警
CREATE TABLE regulatory_alerts (
    id TEXT PRIMARY KEY,
    rule_code TEXT NOT NULL,
    patient_id TEXT REFERENCES medical_wristband_patients(id),
    hospital_id TEXT,
    department TEXT,
    severity TEXT CHECK (severity IN ('low','medium','high')),
    alert_type TEXT NOT NULL CHECK (alert_type IN (
        'no_verify','fence_violation','fake_admission',
        'expense_spike','med_verify_mismatch',
        'frequent_transfer','device_disconnect','post_discharge'
    )),
    detail TEXT NOT NULL,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending','acknowledged','resolved','false_positive')),
    triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at DATETIME,
    acknowledged_by TEXT,
    resolved_at DATETIME,
    resolved_by TEXT,
    notes TEXT
);
```

### 4.5 核心表SQL（社区腕带）

```sql
-- 社区老人档案
CREATE TABLE community_elders (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    id_card TEXT UNIQUE NOT NULL,
    gender INTEGER NOT NULL,
    age INTEGER,
    address TEXT,
    emergency_contact TEXT,
    bank_account TEXT,
    hospital_id TEXT,
    status TEXT DEFAULT 'active' CHECK (status IN ('active','deactivated','deceased')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deactivated_at DATETIME,
    deactivated_reason TEXT
);

-- 社区腕带设备
CREATE TABLE community_wristband_devices (
    id TEXT PRIMARY KEY,
    device_id TEXT UNIQUE NOT NULL, -- CW-XXXX
    firmware_version TEXT,
    mode TEXT DEFAULT 'community' CHECK (mode IN ('hospital','community')),
    status TEXT DEFAULT 'active' CHECK (status IN ('active','inactive','retired')),
    last_seen DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 老人-腕带绑定
CREATE TABLE community_elder_bindings (
    id TEXT PRIMARY KEY,
    elder_id TEXT NOT NULL REFERENCES community_elders(id),
    device_id TEXT NOT NULL REFERENCES community_wristband_devices(id),
    bound_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    unbound_at DATETIME,
    UNIQUE(elder_id, device_id)
);

-- 福利标签配置
CREATE TABLE community_welfare_tag_config (
    id TEXT PRIMARY KEY,
    tag_code TEXT UNIQUE NOT NULL,
    tag_name TEXT NOT NULL,
    issuer TEXT NOT NULL,           -- 民政局/残联/医保局/交通局
    renewal_period_days INTEGER,
    benefit_amount REAL,
    enabled INTEGER DEFAULT 1
);

-- 老人-福利绑定
CREATE TABLE community_elder_welfare (
    id TEXT PRIMARY KEY,
    elder_id TEXT NOT NULL REFERENCES community_elders(id),
    tag_code TEXT NOT NULL REFERENCES community_welfare_tag_config(tag_code),
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    certified_by TEXT,
    certification_doc TEXT,
    effective_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME,
    UNIQUE(elder_id, tag_code, valid_from, valid_to)
);

-- 签到记录
CREATE TABLE community_signin_records (
    id TEXT PRIMARY KEY,
    elder_id TEXT NOT NULL REFERENCES community_elders(id),
    device_id TEXT NOT NULL,
    hospital_id TEXT NOT NULL,
    pharmacist_id TEXT,
    signin_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    period TEXT NOT NULL,           -- YYYY-MM
    activated_tags TEXT DEFAULT '[]',
    is_medical_signin INTEGER DEFAULT 1,
    is_welfare_signin INTEGER DEFAULT 1,
    notes TEXT
);

-- 药房发药记录
CREATE TABLE community_pharmacy_logs (
    id TEXT PRIMARY KEY,
    elder_id TEXT NOT NULL REFERENCES community_elders(id),
    device_id TEXT,
    hospital_id TEXT NOT NULL,
    pharmacist_id TEXT,
    dispense_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    period TEXT NOT NULL,
    items TEXT NOT NULL,            -- JSON: [{"drug_name":"氨氯地平","qty":30,"cost":45.00}]
    total_cost REAL DEFAULT 0,
    insurance_covered REAL DEFAULT 0,
    self_pay REAL DEFAULT 0,
    notes TEXT
);

-- 民政数据同步日志
CREATE TABLE community_minzheng_sync (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    filename TEXT,
    imported_count INTEGER DEFAULT 0,
    matched_count INTEGER DEFAULT 0,
    pending_review_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    status TEXT DEFAULT 'processing' CHECK (status IN ('processing','completed','failed')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- 批量发放记录
CREATE TABLE community_batch_payments (
    id TEXT PRIMARY KEY,
    batch_id TEXT NOT NULL,
    period TEXT NOT NULL,
    pay_type TEXT NOT NULL,
    elder_id TEXT NOT NULL REFERENCES community_elders(id),
    amount REAL NOT NULL,
    bank_account TEXT,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending','success','failed','retrying')),
    failure_reason TEXT,
    executed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 第五部分 API接口设计

### 5.1 路由前缀约定

| 模块 | 路由前缀 | 说明 |
|------|---------|------|
| 现有业务 | `/api/v1/devices/*`, `/api/v1/health/*`, `/api/v1/alerts/*` | 不动 |
| 医用腕带 | `/api/v1/medical/*` | 新增 |
| 监管专区 | `/api/v1/regulatory/*` | 新增 |
| 社区腕带 | `/api/v1/community/*` | 新增 |

### 5.2 患者管理API

```http
POST   /api/v1/medical/patients              — 患者入院登记
GET    /api/v1/medical/patients               — 在院患者列表
GET    /api/v1/medical/patients/:id           — 患者详情
PUT    /api/v1/medical/patients/:id           — 更新患者信息
DELETE /api/v1/medical/patients/:id           — 出院注销
POST   /api/v1/medical/patients/batch-import  — 批量导入
GET    /api/v1/medical/patients/by-admission-no — 按住院号查询
```

### 5.3 腕带绑定API

```http
POST   /api/v1/medical/patients/:id/bind      — 绑定腕带
POST   /api/v1/medical/patients/:id/unbind    — 解绑腕带
POST   /api/v1/medical/wristbands/:id/write   — 写入腕带固件
POST   /api/v1/medical/wristbands/:id/clear   — 出院清空腕带
GET    /api/v1/medical/wristbands             — 腕带设备列表
GET    /api/v1/medical/wristbands/:id/firmware — 腕带固件版本
```

### 5.4 医疗清单API（家属端使用）

```http
GET    /api/v1/medical/lists/expenses?patient_id=&start_date=&end_date=
GET    /api/v1/medical/lists/medications?patient_id=&status=active
GET    /api/v1/medical/lists/tests?patient_id=&start_date=&end_date=
GET    /api/v1/medical/daily/entries?patient_id=&date=

# 患者治疗经过汇总（家属端专用）
GET    /api/v1/medical/history?elderly_id=
Response: {
  "patient": {...},
  "expenses": [...],
  "medications": [...],
  "test_results": [...],
  "daily_entries": [...]
}
```

### 5.5 核验记录API

```http
GET    /api/v1/medical/verifications?patient_id=&nurse_id=&start=&end=
PUT    /api/v1/medical/verifications/:id/status — 标记核验完成
GET    /api/v1/medical/verifications/stats/today — 今日核验统计
```

### 5.6 监管API

```http
GET    /api/v1/regulatory/dashboard/patient-overview — 在院总览
GET    /api/v1/regulatory/dashboard/patient-list     — 在院患者列表
GET    /api/v1/regulatory/alerts                     — 告警列表
POST   /api/v1/regulatory/alerts/:id/acknowledge     — 确认告警
POST   /api/v1/regulatory/alerts/:id/resolve         — 标记解决
GET    /api/v1/regulatory/audit/patient/:id          — 穿透审计
GET    /api/v1/regulatory/rules                      — 规则配置
PUT    /api/v1/regulatory/rules/:code/config         — 修改规则配置
POST   /api/v1/regulatory/fence/config               — 配置围栏
GET    /api/v1/regulatory/compliance/report          — 合规报表
```

### 5.7 社区腕带API

```http
# 老人档案管理
POST   /api/v1/community/elders              — 老人档案登记
GET    /api/v1/community/elders              — 老人列表
GET    /api/v1/community/elders/:id          — 老人详情
PUT    /api/v1/community/elders/:id          — 更新老人信息
DELETE /api/v1/community/elders/:id          — 注销

# 福利标签管理
GET    /api/v1/community/welfare/tags/config  — 标签配置
POST   /api/v1/community/elders/:id/welfare-tags — 赋予/移除标签
GET    /api/v1/community/elders/:id/welfare-tags/active — 有效标签

# 签到激活
POST   /api/v1/community/signin/trigger      — 触发签到
GET    /api/v1/community/signin/history?elder_id= — 签到历史
GET    /api/v1/community/signin/status?elder_id=&period= — 本期状态

# 药房发药
GET    /api/v1/community/pharmacy/prescriptions?elder_id= — 可领药物
POST   /api/v1/community/pharmacy/dispense   — 确认发药核销

# 民政数据导入
POST   /api/v1/community/minzheng/import     — 上传民政数据
GET    /api/v1/community/minzheng/import/tasks — 导入任务
POST   /api/v1/community/minzheng/import/:task_id/approve — 批量生效

# 批量发放
POST   /api/v1/community/batch-pay/execute   — 手动触发发放
GET    /api/v1/community/batch-pay/records?period= — 发放记录
```

### 5.8 家属端隐私过滤

| 数据类型 | 家属可见内容 | 脱敏规则 |
|---------|------------|---------|
| 用药清单 | 药名 + 用法 + 频次 | 不含剂量详情 |
| 费用清单 | 项目名称 + 金额 | 不含诊断关联 |
| 检测报告 | 正常/异常标记 + 结论摘要 | 不含详细指标 |
| 每日查房记录 | ❌ 不开放 | 隐私数据 |
| 过敏史/特殊疾病 | ❌ 不开放 | 敏感医疗信息 |
| 位置/围栏数据 | ❌ 不开放 | 监管数据 |
| 核验记录 | ❌ 不开放 | 医护操作记录 |

---

## 第六部分 规则引擎

### 6.1 住院监管规则（R01-R08）

| 规则编号 | 规则名称 | 检测逻辑 | 告警级别 |
|---------|---------|---------|---------|
| R01 | 挂床住院 | 连续N小时无护士扫码核验记录（默认24h） | 高 |
| R02 | 电子围栏越界 | 定位超出医院围栏且停留>N分钟（默认30min） | 高 |
| R03 | 虚假入院 | 腕带已绑定但48h内无核验记录 | 中 |
| R04 | 费用突增 | 单患者日费用>科室均值3倍 | 中 |
| R05 | 用药与核验不匹配 | 有用药记录但无给药核验记录 | 中 |
| R06 | 频繁转科 | 7天内转科>3次 | 低 |
| R07 | 腕带异常断开 | 腕带信号突然中断超过2h | 高 |
| R08 | 长期不在院 | 出院后腕带未清空 | 低 |

### 6.2 社区场景规则（R_C01-R_C08）

| 规则编号 | 规则名称 | 检测逻辑 | 告警级别 |
|---------|---------|---------|---------|
| R_C01 | 重复领取 | 同一老人同月内在不同社区医院签到 | 高 |
| R_C02 | 冒领嫌疑 | 腕带认证后无生理数据上报+活动轨迹异常 | 高 |
| R_C03 | 异常高频 | 7天内签到/取药>5次 | 中 |
| R_C04 | 僵尸账户 | 腕带离线>30天+无签到记录 | 低 |
| R_C05 | 补助未到账 | 已激活发放但银行返回失败 | 高 |
| R_C06 | 跨区领取 | 同一福利标签在多个区县被使用 | 中 |
| R_C07 | 福利过期未停 | 福利标签已过有效期但仍在使用 | 中 |
| R_C08 | 死亡未注销 | 老人已出具死亡证明但腕带仍活跃 | 高 |

### 6.3 规则引擎调度

```go
// 规则引擎作为 admin-api 内的定时任务，每5分钟执行一轮
type RuleEngine struct {
    store     *store.SQLiteStore
    log       *zap.Logger
    hospital  *HospitalRuleSet   // R01-R08
    community *CommunityRuleSet    // R_C01-R_C08
}

func (e *RuleEngine) Run() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        e.hospital.Run()    // 执行 R01-R08
        e.community.Run()   // 执行 R_C01-R_C08
    }
}
```

---

## 第七部分 家属端住院治疗功能

### 7.1 家属认证绑定机制

**绑定流程：**

```
1. 护士站在后台录入"家属-患者"关系
   - 输入家属手机号 + 患者住院号
   - 系统生成绑定记录

2. 家属收到短信验证码
   - 在APP/小程序中输入验证码完成认证

3. 认证成功后
   - 家属APP首页显示"住院治疗"入口
   - 可查看该患者的用药/检查/治疗/费用
   - 每次查看需二次认证（短信验证码或人脸识别）
```

**权限控制：**

| 操作 | 是否需要二次认证 | 说明 |
|------|----------------|------|
| 查看用药清单 | 否 | 基础信息 |
| 查看费用清单 | 否 | 基础信息 |
| 查看检测报告 | 否 | 结果摘要 |
| 查看详情 | 是 | 需要短信验证码 |
| 下载报告 | 是 | 需要短信验证码 |

### 7.2 家属APP住院治疗页面

```
家属APP — "住院治疗"页面:

┌─────────────────────────────────────┐
│  父母：王秀英                       │
│  住院号：20260721001                │
│  科室：心内科  床号：12             │
│  入院日期：2026-07-21               │
├─────────────────────────────────────┤
│  Tab: [用药] [检查] [治疗] [费用]   │
├─────────────────────────────────────┤
│  用药清单:                          │
│  ● 氨氯地平 5mg po qd  ✅已服       │
│  ● 二甲双胍 500mg po bid ✅已服     │
│  ● 阿司匹林 100mg po qd  ⏳待服     │
├─────────────────────────────────────┤
│  检测报告:                          │
│  ● 血常规 07-20 正常 ✓             │
│  ● 心电图 07-19 异常 ⚠             │
│  ● 血脂 07-21 待出结果 ⏳           │
├─────────────────────────────────────┤
│  费用清单:                          │
│  ● CT检查 ¥350                      │
│  ● 护理费 ¥120/天                   │
│  ● 药费 ¥86.50                      │
│  ─────────────────                  │
│  累计费用: ¥2,450.50                │
└─────────────────────────────────────┘
```

---

## 第八部分 管理后台页面结构

```
管理后台
├── 仪表盘（不动）
├── 设备管理（不动）
├── 用户管理（不动）
├── 订阅管理（不动）
├── 医护工作站（【扩展】）
│   ├── 患者登记（手动录入 + HIS同步 + 批量导入）
│   ├── 腕带绑定（单独/批量绑定 + 批量写入 + 出院解绑）
│   ├── 每日录入（用药/检查/治疗/费用/查房记录）
│   ├── 核验记录（列表 + 统计 + 导出）
│   └── 统计看板（在院总数/今日出入院/核验次数/腕带使用率）
├── 监管专区（【新增】）
│   ├── 在院总览（实时看板）
│   ├── 异常告警（规则引擎告警列表 + 确认 + 解决）
│   ├── 穿透审计（全链路数据追溯）
│   ├── 规则配置（阈值/围栏参数配置）
│   └── 合规审查（周期报表）
└── 社区老人专区（【新增】）
    ├── 老人档案管理
    ├── 福利标签管理
    ├── 签到总览
    ├── 药房发药记录
    ├── 民政数据导入
    └── 统计看板
```

---

## 第九部分 实施批次

```
第一批 (T+1天):   语言强制 + 子系统独立性
├── 修复所有 Flutter UI 文本为中文
├── 修复所有 Go 服务端错误消息为中文
├── 修复 start.sh 脚本确保只启动对应服务
└── 验证每个子系统独立运行

第二批 (T+2天):   数据库策略
├── 定义 Store 接口（已在 cloud/admin-api/internal/store/store.go）
├── 实现 SQLiteStore（已在 cloud/admin-api/internal/store/sqlite.go）
├── 实现 PostgreSQLStore（骨架）
└── 编写迁移脚本

第三批 (T+3天):   缺失功能 — 住院治疗
├── 添加住院每日汇总 API（已在 cloud/admin-api/internal/handler/medical_wristband.go）
├── 添加 HospitalizationPage（家属APP）
├── 添加家属-患者绑定机制
└── 端到端联调

第四批 (T+4天):   文档同步
├── 更新 docs/specs/
├── 更新各子系统 README.md
└── 更新根 README.md

第五批 (T+5天):   医用腕带固件
├── ESP32-S3 固件工程初始化
├── NFC 近场读取模块集成
├── Cat1 MQTT 上行体征数据
├── OLED 显示 + LED 警示
└── 腕带固件单元测试

第六批 (T+6-7天): 监管闭环
├── 规则引擎 R01-R08 实现
├── 电子围栏基站定位
├── 监管专区前端页面
└── 穿透审计 API

第七批 (T+8-9天): 社区腕带
├── 社区腕带固件（hospital分支扩展）
├── 社区老人 CRUD API
├── 签到激活 + 药房发药
├── 批量发放引擎
└── 社区专区前端页面

第八批 (T+10天): 灰度发布
├── 内部灰度测试
├── 用户手册编写
└── 全量发布准备
```

---

## 第十部分 验证标准

### 10.1 语言验证

```bash
# 检查 Flutter 是否有英文 UI 文本
grep -r "login\|error\|success" apps/family-app/lib/ --include="*.dart" | grep -v "//"

# 检查 Go 是否有英文错误消息
grep -r '"error"\|"success"' cloud/admin-api/ --include="*.go"
```

### 10.2 数据库验证

```bash
# 验证 Store 接口完整性
cd cloud/admin-api && go vet ./...

# 验证 SQLite 迁移脚本
./scripts/migrate_sqlite_to_postgres.sh --dry-run
```

### 10.3 子系统独立性验证

```bash
# 分别启动每个子系统
./scripts/start.sh start cloud    # 仅启动云后端
./scripts/start.sh start apps     # 仅启动家属APP+管理后台
./scripts/start.sh start website  # 仅启动品牌官网

# 分别运行每个子系统的测试
cd cloud/admin-api && go test ./...
cd apps/family-app && flutter analyze
cd apps/admin-web && npm run lint
```

### 10.4 零误差验证

```
测试用例:
1. 同时扫描腕带A和腕带B → A的信息绝不混入B
2. 腕带A绑定患者甲 → 扫描后返回患者甲的完整诊疗记录
3. 腕带A绑定患者甲 → 患者乙的家属APP → 无法查看任何数据
4. 腕带A解绑 → 再次扫描返回"无绑定患者"
5. 腕带A绑定患者甲 → 患者甲出院 → 腕带清空 → 再次扫描返回出厂状态
```

---

## 第十一部分 风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|---------|
| NFC芯片采购周期 | 硬件开发延迟 | MVP阶段先用BLE作为兜底方案 |
| Cat1基站定位精度不足 | 围栏误判 | MVP阶段设置较大围栏半径(200-500m)，后续升级GPS |
| 规则引擎误报 | 告警疲劳 | 支持误报标记(false_positive)，积累数据后优化阈值 |
| SQLite并发写入冲突 | 多护士同时录入 | MVP限制单写多读，未来切PG解决 |
| 医疗数据合规风险 | 患者隐私泄露 | 严格加密+访问控制+审计日志，上线前完成合规评审 |
| 腕带丢失/损坏 | 老人无法认证 | 支持"换带不换人" — 新腕带绑定同一档案，旧腕带自动注销 |
| 民政数据格式不统一 | 批量导入失败率高 | 提供标准化模板+错误行提示+人工审核兜底 |

---

## 第十二部分 合规依据

### 12.1 医疗质量安全核心制度

| 制度 | 本方案对应功能 |
|------|-------------|
| 查对制度 | 护士NFC扫码腕带核验，记录核验日志 |
| 首诊负责制度 | 入院登记绑定腕带，记录护士账号 |
| 交接班制度 | 转科操作触发状态变更和监管通知 |
| 病历管理制度 | 每日诊疗录入，不可篡改 |
| 分级护理制度 | 警示标签分级(P0/P1/P2) |

### 12.2 医保基金监管条例

| 监管要求 | 本方案对应措施 |
|---------|-------------|
| 禁止虚构医疗服务 | 每日核验记录证明患者真实在院 |
| 禁止分解住院 | 转科操作与入院记录分离 |
| 禁止串换项目 | 费用清单与用药/检测核验记录交叉比对 |
| 禁止挂床住院 | 电子围栏+无核验告警R01 |

### 12.3 个人信息保护法(PIPL)

| 合规要求 | 本方案对应措施 |
|---------|-------------|
| 最小必要原则 | 腕带本地仅存身份标识+警示标签 |
| 知情同意 | 患者/老人签署《个人信息处理同意书》 |
| 访问控制 | RBAC角色权限 |
| 数据安全 | AES-256-GCM存储加密+TLS 1.3传输加密 |
| 数据删除 | 出院腕带必须清空，不可恢复 |

---

© 2026 Eregen (颐贞). All rights reserved.
