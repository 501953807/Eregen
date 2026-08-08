# 颐贞 Eregen — 医用电子腕带增量升级设计方案（v1.0）

> 编制日期：2026-07-21  
> 版本：v1.0（首次完整方案）  
> 状态：待评审  
> 对应子系统：① 手环固件（新增医用腕带子模块）、③ 云平台后端（新增医用模块）、⑤ 管理后台（新增医护工作站）、④ 家属APP（增量页面）、⑦ 微信小程序（增量页面）  
> 项目性质：技术预研可行性开发 — 研究功能可行性和系统落地

---

## 第一部分 项目背景与迭代目标

### 1.1 项目现状

颐贞 Eregen 平台当前包含 8 个子系统：三档手环固件（Starter/Plus/Pro）、三档药盒固件（Basic/Smart/Auto）、Go + Gin 云平台后端、Flutter 家属 APP、Vue 3 管理后台、品牌官网、微信小程序和 B2B 对接层。MVP 阶段定位为"生活方式健康设备"，面向 60-85 岁老年人群体，核心业务为老人安全监护、用药提醒和健康数据采集。

现有手环固件运行在 GD32E230C8T3 + FreeRTOS 上，通过 Cat1 蜂窝 MQTT 上报心率、血氧、定位、SOS、跌倒等数据。云平台采用 Go 微服务架构（gateway / api-server / push-service / data-pipeline / admin-api），MVP 阶段数据存储使用 SQLite。

### 1.2 迭代目标

本次迭代以**完全增量方式**新增"医用电子腕带（病人识别带）"完整功能，替代传统纸质住院腕带，实现：

1. **住院患者身份唯一识别**：通过医用手环存储患者基础信息、警示标签和病历摘要，护士终端近场秒级读取
2. **医护核对电子化**：给药、输液、手术、检查、抽血、输血前身份核验，杜绝医疗差错
3. **风险警示电子化**：过敏、跌倒高危、隔离患者等标签差异化标识，辅助医护快速识别
4. **治疗经过透明化**：家属可通过 APP/小程序查看绑定患者的费用清单、用药记录、检测报告
5. **多终端型号管控**：建立通用手环（A/B/C）与医用专用腕带的独立目录、固件、版本管理体系

### 1.3 硬性约束

| 约束 | 说明 |
|------|------|
| **零破坏原则** | 所有新增内容为增量开发，禁止修改、删减、重构原有项目架构、目录、代码、业务逻辑 |
| **轻量 MVP 原则** | 当前不做微服务拆分、分布式、高可用、集群架构设计，拒绝过度设计 |
| **可演进原则** | 所有设计面向未来商业化扩展，预留架构升级、模块拆分、终端扩容能力 |
| **标准化落地原则** | 方案专业、结构化、条理清晰，可直接作为项目开发规范文档 |
| **冲突处理原则** | 如新方案与原有架构存在冲突，必须先经用户同意再执行，不得自行决定 |
| **数据库降级** | MVP 阶段全项目使用 SQLite 替代 PostgreSQL/InfluxDB/Redis，降低部署成本 |

---

## 第二部分 整体架构兼容设计

### 2.1 新旧功能隔离策略

医用腕带模块与现有手环业务采用**物理隔离 + 逻辑解耦**双保险机制：

```
┌─────────────────────────────────────────────────────────────┐
│                    Eregen 平台架构                          │
│                                                             │
│  ┌──────────────────────┐   ┌──────────────────────────┐   │
│  │  现有业务（不动）      │   │  新增医用腕带（增量）     │   │
│  │                      │   │                          │   │
│  │  手环 Entry/Plus/Pro │   │  医用腕带 Medical-WB    │   │
│  │  药盒 Basic/Smart/A  │   │  （独立 ESP32-S3 芯片）  │   │
│  │  云平台 gateway/API  │   │                          │   │
│  │  家属APP/管理后台/小程序 │ │  medical-wb/ 新增模块   │   │
│  │  B2B 对接            │   │  医护工作站新增页面       │   │
│  └──────────────────────┘   └──────────────────────────┘   │
│                                                             │
│  共享基础设施：                                              │
│  ├── SQLite 主库（现有业务表 + 医用腕带表）                  │
│  ├── EMQX MQTT Broker（通用 Topic + 医用 Topic 分区）        │
│  └── NATS JetStream（通用事件 + 医用事件分区）                │
└─────────────────────────────────────────────────────────────┘
```

**隔离规则：**

| 层级 | 隔离方式 | 说明 |
|------|---------|------|
| 硬件层 | 独立 MCU | 医用腕带使用 ESP32-S3，与现有手环 GD32E230 完全独立 |
| 固件层 | 独立目录 | `firmware/medical-wristband/` 与 `firmware/bracelet/` 并行，互不引用 |
| 编译层 | 独立产物 | 医用固件输出 `medical_wristband.bin`，与手环 `bracelet_*.bin` 分离 |
| 协议层 | 独立 Topic | MQTT Topic 使用 `eregen/medical/wb/#` 命名空间，不与现有 `eregen/up/#` 冲突 |
| 数据层 | 独立表族 | SQLite 表使用 `medical_wristband_` 前缀，现有表完全不动 |
| API 层 | 独立路由 | REST API 使用 `/api/v1/medical/` 前缀，不与现有 `/api/v1/devices/` 等冲突 |
| 前端层 | 独立页面 | 管理后台新增侧边栏菜单项"医护工作站"，不修改现有页面组件 |

### 2.2 MVP 单体架构适配方案

当前云平台虽设计为微服务架构，但 MVP 阶段保持**单进程部署**。新增医用腕带模块以内嵌包形式加入 `admin-api`，不新建独立 Go 服务。

```
cloud/
├── admin-api/                 # 现有（不动）
│   ├── device_mgmt.go
│   ├── user_mgmt.go
│   ├── subscription_mgmt.go
│   ├── report_gen.go
│   └── medical_wb/            # 【新增】医用腕带内嵌模块
│       ├── handler/
│       │   ├── patient.go         # 患者 CRUD
│       │   ├── wristband.go       # 绑定/解绑/写入/清空
│       │   ├── medical_list.go    # 费用/用药/检测清单
│       │   ├── daily_entry.go     # 每日诊疗录入
│       │   └── verification.go    # 核验记录
│       ├── service/
│       │   ├── patient_svc.go
│       │   ├── wristband_svc.go
│       │   ├── medical_list_svc.go
│       │   └── verification_svc.go
│       ├── store/
│       │   └── sqlite.go          # SQLite 操作封装
│       └── router/
│           └── medical_routes.go  # 路由注册
```

**职责划分：**

| 模块 | 职责 | 依赖 |
|------|------|------|
| `admin-api/device_mgmt.go` | 现有设备管理 | 不动 |
| `admin-api/medical_wb/handler/` | HTTP 请求处理 | `service/` |
| `admin-api/medical_wb/service/` | 业务逻辑 | `store/` + 云端下发指令 |
| `admin-api/medical_wb/store/` | 数据访问 | SQLite |
| `admin-api/medical_wb/router/` | 路由注册 | Gin Router |

### 2.3 未来微服务/分布式扩展预留

所有新增设计遵循**接口标准化、数据分层、配置解耦**三原则，确保后续拆分时无需重构业务代码：

**接口标准化：**
- 所有 API 端点使用 `/api/v1/medical/` 前缀，未来拆分为独立微服务时只需调整路由转发
- Handler → Service → Store 三层架构已固化，拆分时每个层级可独立部署
- 消息协议使用 JSON 格式，NATS Topic 使用 `eregen.medical.wb.#` 命名空间

**数据分层：**
- 医用腕带数据存储在独立 SQLite 文件 `eregen_medical.db`，未来迁移 PG 时只需替换 Store 实现
- 所有表使用 `medical_wristband_` 前缀，与现有表物理隔离
- 预留 `hospital_id` 字段支持未来多医院场景

**配置解耦：**
- 医用腕带相关配置集中在 `config/medical_wb.yaml`，不混入现有配置
- NFC NDEF 记录类型、MQTT Topic、OTA 通道等参数均可配置

---

## 第三部分 多终端型号版本管控体系（重点）

### 3.1 项目目录层级优化

新增医用腕带专属目录分区，明确通用型号与医用专用型号的目录管控规则：

```
firmware/
├── bracelet/                        # 【现有】通用手环（不动）
│   ├── entry/                       #   Starter 入门版
│   │   ├── main/
│   │   ├── sensors_ppg.c/h
│   │   ├── sensors_imu.c/h
│   │   ├── cat1_at.c/h
│   │   ├── ota/
│   │   └── CMakeLists.txt
│   ├── plus/                        #   Plus 中端版
│   │   ├── main/
│   │   ├── ble_pair.c/h
│   │   ├── plus/fall_detect.c/h
│   │   ├── plus/geofence_manager.c/h
│   │   ├── display_st7789.c/h
│   │   ├── ota/
│   │   └── CMakeLists.txt
│   ├── pro/                         #   Pro 高端版
│   │   ├── main/
│   │   ├── ecg_driver.c/h
│   │   ├── gps_manager.c/h
│   │   ├── display_amoled.c/h
│   │   ├── ota/
│   │   └── CMakeLists.txt
│   └── common/                      #   公共组件（CRC/日志/环形缓冲）
│
├── pillbox/                         # 【现有】药盒（不动）
│   ├── basic/
│   ├── smart/
│   ├── auto/
│   └── common/
│
└── medical-wristband/               # 【新增】医用腕带专属目录
    ├── esp32s3/                     #   医用专用固件（ESP32-S3）
    │   ├── main/
    │   │   ├── patient_store.c/h        # 患者信息 NVS/Flash 存储
    │   │   ├── nfc_server.c/h           # NFC 服务（护士端读取）
    │   │   ├── cat1_mqtt.c/h            # Cat1 联网（下行数据写入）
    │   │   ├── display_oled.c/h         # OLED 警示显示
    │   │   ├── led_indicator.c/h        # LED 风险警示灯
    │   │   ├── nvs_manager.c/h          # 非易失存储管理
    │   │   ├── security.c/h             # 数据加密/防篡改
    │   │   └── board_init.c/h           # 板级初始化
    │   ├── protocol/
    │   │   ├── medical_protocol.h       # 医用专属消息协议定义
    │   │   ├── message_encode.c/h       # JSON 编码
    │   │   └── message_decode.c/h       # JSON 解码
    │   ├── ota/
    │   │   ├── ota_download.c/h         # OTA 下载
    │   │   ├── ota_verify.c/h           # SHA-256 签名验证
    │   │   └── boot_switch.c/h          # 启动切换
    │   ├── crypto/
    │   │   ├── aes_gcm.c/h              # AES-256-GCM 加密
    │   │   └── sha256.c/h               # SHA-256 哈希
    │   ├── test/
    │   │   ├── test_patient_store.c
    │   │   ├── test_ble_server.c
    │   │   └── test_security.c
    │   └── CMakeLists.txt
    │
    └── docs/                          #   医用腕带配套文档
        ├── hardware_spec.md             # 硬件规格书
        ├── firmware_api.md              # 固件接口文档
        └── flash_layout.md              # Flash 分区规划
```

**目录管控规则：**

| 规则 | 说明 |
|------|------|
| 通用能力不放入医用目录 | PPG 传感器驱动、跌倒检测算法、Cat1 AT 指令等通用能力保留在 `bracelet/` 和 `pillbox/` 中 |
| 医用专属能力不放入通用目录 | 患者信息存储、NFC 医护读取、医疗警示显示等能力仅在 `medical-wristband/` 中 |
| 公共组件共享 | CRC、日志、环形缓冲等底层公共组件可被医用腕带引用，但通过头文件包含，不复制代码 |
| 编译产物隔离 | 通用手环输出 `bracelet_entry.bin` / `bracelet_plus.bin` / `bracelet_pro.bin`，医用腕带输出 `medical_wristband.bin` |

### 3.2 固件版本管理方案

#### 3.2.1 版本号规范

采用语义化版本 `MAJOR.MINOR.PATCH`：

| 版本号段 | 含义 | 示例 |
|---------|------|------|
| MAJOR | 架构级变更（如 MCU 更换、协议大改） | 1.0.0 → 2.0.0 |
| MINOR | 功能新增（如新增警示标签类型） | 1.0.0 → 1.1.0 |
| PATCH | Bug 修复、安全补丁 | 1.1.0 → 1.1.1 |

通用手环与医用腕带**独立版本号**，互不影响：

| 产品线 | 固件版本 | 发布说明 |
|--------|---------|---------|
| 手环 Entry | 1.2.0 | 跌倒检测算法优化 |
| 手环 Plus | 1.3.1 | 电池优化器修复 |
| 手环 Pro | 2.0.0 | ECG 驱动重写 |
| 医用腕带 | 1.0.0 | 首个正式版 |
| 药盒 Smart | 1.1.0 | TTS 语音优化 |
| 药盒 Auto | 1.0.2 | 电机控制精度修复 |

#### 3.2.2 OTA 灰度更新机制

```
云平台下发 OTA 指令
    ↓
设备接收指令 → 下载到 Flash 空闲区
    ↓
SHA-256 签名验证 + CRC32 完整性检查
    ↓
验证通过 → 设置 Boot 标志位 → 下次重启进入新版本
    ↓
新版本运行 3 次心跳正常 → 标记为"稳定版本"
    ↓
异常 → 自动回滚到旧版本
```

**灰度策略：**

| 阶段 | 比例 | 时长 | 动作 |
|------|------|------|------|
| 内部测试 | 10% 设备 | 3 天 | 观察崩溃率、心跳异常率 |
| 小范围灰度 | 30% 设备 | 7 天 | 收集用户反馈、稳定性数据 |
| 全量推送 | 100% 设备 | 持续 | 所有在线设备接收升级 |

**版本回溯：**
- 每次 OTA 失败自动回滚到上一稳定版本
- 云端维护"允许升级版本列表"，可手动阻止特定版本推送
- 设备本地保存最近 2 个版本镜像，支持双备份回滚

### 3.3 硬件差异化适配逻辑

不同型号手环硬件能力差异通过**编译宏 + 运行时探测**两层机制兼容：

```c
// 编译期：通过 CMakeLists.txt 定义 TARGET_TIER
// bracelet/entry/CMakeLists.txt: add_definitions(-DTARGET_TIER=ENTRY)
// bracelet/plus/CMakeLists.txt: add_definitions(-DTARGET_TIER=PLUS)
// bracelet/pro/CMakeLists.txt: add_definitions(-DTARGET_TIER=PRO)
// medical-wristband/esp32s3/CMakeLists.txt: add_definitions(-DTARGET_DEVICE=MEDICAL_WB)

#if TARGET_TIER == ENTRY
    // Entry: PPG + IMU + GPS + SOS
#elif TARGET_TIER == PLUS
    // Plus: + BLE配网 + 跌倒检测 + 电子围栏
#elif TARGET_TIER == PRO
    // Pro: + ECG + AMOLED + GNSS
#elif TARGET_DEVICE == MEDICAL_WB
    // 医用腕带: 独立固件，不继承手环任何硬件能力
#endif
```

**运行时探测：**
- 设备启动时检测实际硬件能力（传感器是否存在、屏幕类型、通信模组）
- 能力信息上报到云平台，用于动态展示可用功能
- 云端配置按能力下发，避免向不支持的硬件发送无效指令

### 3.4 医用专属手环型号独立管控规范

| 维度 | 通用手环（Entry/Plus/Pro） | 医用腕带（Medical-WB） |
|------|--------------------------|----------------------|
| MCU | GD32E230C8T3 | ESP32-S3 |
| RTOS | FreeRTOS | ESP-IDF (FreeRTOS) |
| 通信 | Cat1 MQTT + BLE配网(Plus/Pro) | Cat1 MQTT + NFC(医护读取) |
| 传感器 | PPG/IMU/GPS/ECG | 无健康传感器，仅身份存储+警示 |
| 显示 | LCD/AMOLED（显示健康数据） | OLED（显示患者姓名+警示标签） |
| 固件目录 | `firmware/bracelet/{entry,plus,pro}/` | `firmware/medical-wristband/esp32s3/` |
| 编译产物 | `bracelet_*.bin` | `medical_wristband.bin` |
| OTA 通道 | `eregen/down/{dev_id}/ota` | `eregen/medical/wb/{dev_id}/ota` |
| 版本号 | 独立 MAJOR.MINOR.PATCH | 独立 MAJOR.MINOR.PATCH |
| 业务模块 | 健康监测/定位/SOS/跌倒 | 患者信息管理/近场读取/警示显示 |

---

## 第四部分 医用腕带完整功能需求规格

### 4.1 手环固件功能

#### 4.1.1 信息存储

医用手环固件可存储以下患者核心信息，存储在 ESP32-S3 的 NVS（Non-Volatile Storage）或外部 SPI Flash 中：

| 信息类别 | 字段 | 长度限制 | 说明 |
|---------|------|---------|------|
| 基础身份 | 姓名 | 64 字节 | 患者姓名 |
| 基础身份 | 性别 | 1 字节 | 0=未知，1=男，2=女 |
| 基础身份 | 年龄 | 2 字节 | 周岁 |
| 基础身份 | 住院号 | 32 字节 | **唯一核心标识**，全院唯一 |
| 基础身份 | 科室 | 32 字节 | 住院科室名称 |
| 基础身份 | 床号 | 16 字节 | 床位编号 |
| 拓展医疗 | 血型 | 8 字节 | A/B/AB/O 及 Rh 因子 |
| 拓展医疗 | 过敏史 | 128 字节 | 过敏药物/食物/物质 |
| 拓展医疗 | 特殊疾病 | 64 字节 | 如糖尿病、高血压等 |
| 警示标签 | 标签列表 | 64 字节 | JSON 数组：`["allergy","fall_risk","isolation"]` |
| 病历摘要 | 摘要文本 | 256 字节 | 当前治疗方案简述 |

#### 4.1.2 加密与安全

| 安全机制 | 实现方式 | 说明 |
|---------|---------|------|
| 存储加密 | AES-256-GCM | 患者信息写入 Flash 前加密，读取时解密 |
| 密钥管理 | 设备唯一密钥（从 MAC 地址派生） | 每只腕带独立密钥，无法跨设备解密 |
| 传输加密 | TLS 1.3（Cat1 上行）+ NFC Secure Channel | 云端下发和近场读取均加密 |
| 防篡改 | SHA-256 签名校验 | 每次读取时验证数据完整性，异常则告警 |
| 访问控制 | NFC 配对码 | 护士终端需输入 4 位配对码才能读取完整信息 |

#### 4.1.3 展示与警示

腕带 OLED 屏幕显示内容：

```
┌─────────────────────────┐
│  患者姓名                │
│  住院号：20260721001     │
│  科室：心内科  床号：12  │
├─────────────────────────┤
│  ⚠️ 青霉素过敏           │  ← 警示标签红色闪烁
│  ⚠️ 跌倒高危             │  ← 警示标签橙色常亮
├─────────────────────────┤
│  最后更新：2026-07-21    │
│  固件版本：1.0.0         │
└─────────────────────────┘
```

**LED 警示灯：**

| 标签类型 | LED 颜色 | 闪烁模式 | 说明 |
|---------|---------|---------|------|
| 过敏 | 红色 | 快闪（2Hz） | 给药/输血前重点警示 |
| 跌倒高危 | 橙色 | 慢闪（1Hz） | 护理时注意防护 |
| 隔离患者 | 蓝色 | 常亮 | 接触前需防护措施 |
| 昏迷/无表达能力 | 黄色 | 脉冲 | 需额外身份核验 |

#### 4.1.4 出院解绑

| 操作 | 触发方式 | 效果 |
|------|---------|------|
| 出院解绑 | 管理后台点击"出院清空" | 清除腕带中全部患者信息，恢复出厂状态 |
| 手动清空 | 长按腕带按键 10 秒 | 紧急情况下手动清除所有数据 |
| 绑定新患者 | 后台重新写入 | 覆盖原有患者信息，生成新的加密密钥会话 |

### 4.2 后台主控端新增功能（医护工作站）

#### 4.2.1 患者入院登记

| 功能 | 说明 |
|------|------|
| 手动录入 | 填写患者姓名、性别、年龄、住院号、科室、床号、血型、过敏史、特殊疾病、警示标签 |
| 扫码导入 | 扫描住院单据二维码，自动解析并填充患者信息 |
| 批量导入 | Excel/CSV 模板批量导入，支持错误行提示和跳过 |
| 住院号校验 | 全院唯一性校验，防止重复入院登记 |

#### 4.2.2 患者信息管理

| 功能 | 说明 |
|------|------|
| 信息编辑 | 修改患者基础信息、医疗信息、警示标签 |
| 信息更新推送 | 修改后一键推送到已绑定腕带，实时更新本地存储 |
| 医疗清单管理 | 录入/编辑/删除住院费用、用药清单、检测报告 |
| 每日诊疗录入 | 护士每日录入查房记录、护理记录、医嘱执行 |
| 患者状态管理 | 入院/转科/出院/死亡状态流转 |

#### 4.2.3 腕带绑定管理

| 功能 | 说明 |
|------|------|
| 单独绑定 | 选择患者 → 选择腕带 → 绑定，腕带写入患者信息 |
| 批量绑定 | 批量选择患者和腕带，按顺序绑定 |
| 批量写入 | 选中多个患者 → 批量下发到对应腕带 |
| 信息更新 | 修改患者信息后重新写入腕带 |
| 出院解绑 | 患者出院 → 解除绑定 → 清空腕带数据 |
| 腕带库存管理 | 查看腕带设备列表、固件版本、在线状态、绑定状态 |

#### 4.2.4 数据统计看板

| 统计项 | 说明 |
|--------|------|
| 在院患者总数 | 当前绑定腕带的在院患者数量 |
| 今日入院/出院 | 当日入院和出院患者数量 |
| 今日核验次数 | 护士端读取腕带的总次数 |
| 腕带使用率 | 已绑定腕带数 / 腕带总数 |
| 警示标签分布 | 过敏/跌倒高危/隔离等标签的患者数量分布 |
| 腕带固件版本分布 | 各版本固件的设备数量统计 |

### 4.3 护士端交互功能

> **重要说明：** 护士核验终端为独立硬件设备（Android 手持 PDA + 手机 APP），不走浏览器 Web NFC API。管理后台医护工作站提供"手动输入住院号"兜底模式，不依赖 NFC 近场通讯。

#### 4.3.1 近场读取流程（PDA/手机 APP）

```
护士打开护士核验终端（PDA/手机APP）→ 点击"近场核验"
    ↓
Android NFC API → NFC 近场读取医用腕带（NFC-A 106kbps）
    ↓
选择目标腕带 → 输入 4 位配对码
    ↓
NFC 读取腕带数据：
    ├── dev_id（腕带设备 ID）
    ├── patient_id_hash（患者 ID 哈希）
    ├── alert_tags[]（风险标签列表）
    └── medical_summary（病历摘要）
    ↓
终端用 patient_id_hash 请求云端 API
    ↓
返回完整患者信息 + 医疗清单
    ↓
护士查看患者信息 → 执行核验操作（给药/输液/检查等）
    ↓
自动记录核验日志（时间、护士账号、患者 ID、操作类型）
```

> **兜底模式：** 若 NFC 近场通讯不可用，管理后台医护工作站支持手动输入住院号查询患者信息并执行核验。

#### 4.3.2 信息展示

护士端读取腕带后展示：

```
┌─────────────────────────────────────┐
│  患者：王秀英                       │
│  住院号：20260721001                │
│  性别：女  年龄：78 岁              │
│  科室：心内科  床号：12             │
├─────────────────────────────────────┤
│  ⚠️ 警示标签：                      │
│  ● 青霉素过敏（红色高亮）            │
│  ● 跌倒高危（橙色高亮）              │
│  ● 隔离患者（蓝色高亮）              │
├─────────────────────────────────────┤
│  血型：A 型 Rh+                     │
│  过敏史：青霉素、磺胺类              │
│  特殊疾病：高血压、糖尿病            │
├─────────────────────────────────────┤
│  今日用药：                          │
│  ● 氨氯地平 5mg po qd              │
│  ● 二甲双胍 500mg po bid           │
├─────────────────────────────────────┤
│  最近检测：                          │
│  ● 血常规 2026-07-20 正常           │
│  ● 心电图 2026-07-19 异常           │
├─────────────────────────────────────┤
│  [📋 查看完整医疗清单]               │
│  [✅ 确认核验]  [❌ 信息不符]         │
└─────────────────────────────────────┘
```

#### 4.3.3 核验记录留存

每条核验记录包含：

| 字段 | 说明 |
|------|------|
| 核验时间 | 精确到秒 |
| 护士账号 | 操作人身份 |
| 患者 ID | 被核验患者 |
| 腕带 ID | 读取的腕带设备 |
| 操作类型 | 入院核对/给药核对/输液核对/抽血核对/输血核对/检查核对/手术核对/出院核对 |
| 读取数据快照 | 腕带当时返回的数据（用于审计） |
| 核验结果 | 成功/信息不符/腕带异常 |

---

## 第五部分 近场通讯技术方案论证与选型

### 5.1 候选方案对比

针对护士手机/手持终端近距离读取腕带数据的场景，对比 NFC、蓝牙近场（BLE 5.0）、RFID 三种技术方案：

| 维度 | NFC | BLE 5.0 | RFID |
|------|-----|---------|------|
| **通讯距离** | 4cm | 10-100m（近场可控制在 1m 内） | 1-10m |
| **安全性** | 高（物理接触） | 高（配对码 + 加密） | 低（无认证） |
| **数据传输量** | 小（KB 级） | 大（MB 级） | 极小（字节级） |
| **加密支持** | AES-128 | AES-128/256 | 无 |
| **手机兼容性** | Android PDA/手机 APP 原生 NFC API；管理后台兜底为手动输入住院号 | Android PDA/手机 APP 原生 BLE API；管理后台兜底为手动输入住院号 | 需专用读卡器 |
| **腕带功耗** | 极低（被动式） | 低（主动式，待机能达 μA 级） | 极低 |
| **与现有系统兼容** | 不兼容（需新增硬件） | 兼容（Plus/Pro 已有 BLE 经验） | 不兼容（需新增硬件） |
| **ESP32-S3 支持** | ✅ 内置 NFC 读写器 | ✅ 原生支持 | ❌ 不支持 |
| **MVP 开发成本** | 低（芯片内置，无需额外芯片） | 低（芯片内置，Web API 现成） | 高（需采购 RFID 读写模块） |
| **响应时效** | <100ms | <500ms | <200ms |
| **医疗数据合规** | 符合（加密 + 认证） | 符合（加密 + 认证） | 不符合（无加密） |

### 5.2 选型结论

**推荐方案：NFC A 协议近场读取**

**理由：**

1. **硬件原生支持**：ESP32-S3 内置 NFC 读写器，无需额外芯片，BOM 成本最低
2. **软件生态成熟**：Android PDA/手机 APP 原生支持 NFC API；管理后台提供手动输入住院号兜底模式
3. **安全性充分**：NFC Secure Channel + 4 位配对码 + AES-128 加密，满足医疗数据隐私要求
4. **响应时效达标**：近场（<4cm）读取延迟 <100ms，满足"秒级核验"需求
5. **扩展性好**：未来如需升级（如增加心率实时查看），NFC 可无缝扩展

### 5.3 NFC 通信协议设计

医用腕带作为 NFC Tag（被动式），护士终端作为 NFC Reader（主动式）：

```
NFC A 协议 (106 kbps)，近场 4cm 内读写
├── NDEF Record 1: Device Info
│   Type: application/vnd.eregen.device-info
│   数据：dev_id, fw_version, battery_pct
├── NDEF Record 2: Patient Identity
│   Type: application/vnd.eregen.patient
│   属性：Read (需认证)
│   数据：patient_id_hash, name_hash, admission_no_hash
├── Characteristic 3: Alert Tags
│   UUID: ...-0003
│   属性：Read (需认证)
│   数据：JSON 数组 ["allergy","fall_risk","isolation"]
├── Characteristic 4: Medical Summary
│   UUID: ...-0004
│   属性：Read (需认证)
│   数据：UTF-8 文本，≤256 字节
└── Characteristic 5: Auth Challenge
    UUID: ...-0005
    属性：Read/Write
    数据：随机挑战码 + 配对码验证
```

### 5.4 通讯安全机制

```
1. 发现标签：护士终端 NFC 读取腕带 NDEF 数据
2. 认证挑战：
   ├── 腕带生成随机 challenge（16 字节）
   ├── 护士终端输入 4 位配对码
   ├── 双方计算 HMAC-SHA256(challenge, pairing_code)
   └── 匹配成功 → 建立安全会话
3. 数据传输：
   ├── 患者身份信息：AES-128-CBC 加密
   ├── 警示标签：明文传输（已脱敏，仅代码）
   └── 病历摘要：AES-128-CBC 加密
4. 会话超时：5 分钟无活动自动断开
```

### 5.5 响应时效要求

| 阶段 | 耗时 | 说明 |
|------|------|------|

### 5.6 护士终端 Flutter 应用 (`apps/nurse_terminal/`)

**目录结构：**
```
apps/nurse_terminal/lib/src/
  models/
    medical_models.dart        — PatientInfo, VerificationResult (已有)
    patient_record.dart         — 完整患者记录（含入院、查房）
    ward_round_entry.dart       — 护理查房数据
  screens/
    login_screen.dart            — 护士登录（username/password，token 本地存储）
    home_screen.dart             — 在院患者列表 + 搜索 + 扫描腕带 FAB（已有 stub）
    patient_detail_screen.dart   — 完整患者档案（ demographics, admission, allergies, medications, wristband status, verification history）
    verification_screen.dart     — NFC 验证结果展示（匹配/不匹配/未找到）
    ward_round_screen.dart       — 每日查房表单（BP, HR, SpO2, temp, weight, notes, observations）
    medication_screen.dart       — 今日用药计划 + 扫码验证发药
    discharge_screen.dart        — 出院流程（类型: discharged/transferred/deceased, 备注, 转院目的地）
  services/
    nfc_wristband_service.dart    — NFC 读写服务（已有）
    api_client.dart              — HTTP 客户端（admin-api 认证）
    patient_service.dart         — 患者 CRUD
    verification_service.dart    — 验证记录管理
    ward_round_service.dart      — 查房记录管理
```

**NFC 集成流程：**
1. 用户在 Home/Patient Detail 页面点击 "Scan Wristband"
2. NFC service 检测附近腕带（4cm 内）
3. 读取 NDEF payload
4. 读取配对码（challenge-response 认证）
5. 从 NDEF payload 解析患者信息 JSON
6. 显示在 Verification 页面
7. 匹配 → 显示患者详情；不匹配 → 显示告警
8. 用户确认验证 → 保存到 API

**Admin-web 医疗腕带页面增强 (`MedicalWristband.vue`)：**
- **入院 tab** — 在院患者列表（床号/科室/诊断），创建入院表单
- **查房 tab** — 护理查房记录表（生命体征摘要），查看详情
- **出院 tab** — 已出院患者列表（出院类型/日期）
- **监管告警 tab** — R01-R08 告警表（严重程度、状态、处理）

**AuditDetail.vue 真实 API 集成：**
- 替换硬编码 mock 时间线数据为 `regulatoryApi.getAuditTrail(patientId)`
- 获取入院、查房、验证、用药、费用全链路数据
- "Export Report" 按钮 → 从 API 数据生成 PDF

**Admin-web API Client (`apps/admin-web/src/api/medical.ts`) 新增端点：**
`admitPatient`, `getWardRounds`, `completeWardRound`, `dischargePatient`, `getRegulatoryAlerts`, `resolveRegulatoryAlert`, `getAuditTrail`, `exportReport`

**Backend Handlers (`cloud/admin-api/internal/handler/medical_wristband.go`)：**
- `AdmitPatient` — `POST /admin/medical/patients/admit`
- `CompleteWardRound` — `POST /admin/medical/patients/:id/ward-round`
- `DischargePatient` — `POST /admin/medical/patients/:id/discharge`
- `GetPatientWardRound` — `GET /admin/medical/patients/:id/ward-round`

**Regulatory Rule Engine (`cloud/admin-api/internal/service/regulatory_engine.go`)：**
8 条检测规则 R01-R08：
| 规则 | 代码 | 触发条件 | 动作 |
|------|------|---------|------|
| R01 | Bed-fraud | 患者无活跃腕带绑定 | Alert P1 |
| R02 | Geofence breach | 患者位置超出医院院区 | Alert P0 |
| R03 | Fake admission | 入院记录但 2h 内无物理签到 | Alert P1 |
| R04 | Cost surge | 日费用 > 3x 平均值 | Alert P1 |
| R05 | Med-verification mismatch | 发药但无护士验证扫描 | Alert P2 |
| R06 | Frequent transfer | 7天内 >3 次科室转移 | Alert P2 |
| R07 | Wristband NFC offline | 住院期间腕带 NFC 离线 >30min | Alert P1 |
| R08 | Post-discharge data | 出院后仍有健康数据 | Alert P1 |

实现模式：每个规则是函数 `EvaluateRXX(ctx, event) *Alert`，返回 nil 表示无违规。`RuleEngine.EvaluateAll(ctx, event)` 调用所有规则并返回告警。
| NFC 读取发现 | <0.5 秒 | 护士终端 NFC 触碰腕带 |
| NDEF 解析 | <0.2 秒 | 解析腕带数据 |
| 认证挑战 | <1 秒 | 输入配对码 + 验证 |
| 数据读取 | <200ms | 读取全部 NDEF 记录 |
| 云端查询 | <1 秒 | 用 hash 查询完整患者信息 |
| **总计** | **<5 秒** | 从触碰腕带到显示完整信息 |

---

## 第六部分 数据结构与接口设计

### 6.1 数据库选型说明

**MVP 阶段全项目切换 SQLite**，原因：

| 维度 | PostgreSQL | SQLite |
|------|-----------|--------|
| 部署复杂度 | 需安装数据库服务 | 单文件，零部署 |
| 运维成本 | 需备份、监控、调优 | 无需运维 |
| 适合场景 | 生产环境、多用户并发 | MVP/原型/小规模 |
| 迁移成本 | — | Store 层封装后迁移成本低 |

**迁移策略：**
- 所有数据库操作封装在 `store/` 层，业务逻辑不直接依赖 SQL 方言
- 未来切换到 PostgreSQL 时，只需替换 `sqlite.go` 为 `postgres.go`，Handler 和 Service 层无需改动
- 表结构使用标准 SQL 语法，避免 SQLite 特有语法

### 6.2 新增数据表设计（SQLite）

#### 6.2.1 主库 `eregen.db`（现有业务 + 医用增量）

```sql
-- ================================
-- 现有表（迁移到 SQLite，结构不变）
-- ================================
CREATE TABLE users (...);           -- 用户表
CREATE TABLE elderly_profiles (...); -- 老人档案表
CREATE TABLE devices (...);          -- 设备表（含手环/药盒）
CREATE TABLE alerts (...);           -- 告警表
CREATE TABLE subscriptions (...);    -- 订阅表
-- ... 其他现有表

-- ================================
-- 新增：医用腕带患者表
-- ================================
CREATE TABLE medical_wristband_patients (
    id TEXT PRIMARY KEY,
    hospital_id TEXT,              -- 医院 ID（预留 HIS 对接）
    name TEXT NOT NULL,
    gender INTEGER NOT NULL CHECK (gender IN (0, 1, 2)),  -- 0 未知 1 男 2 女
    age INTEGER,
    admission_no TEXT UNIQUE NOT NULL,  -- 住院号（唯一核心标识）
    department TEXT,
    bed_no TEXT,
    blood_type TEXT,
    allergy_history TEXT,          -- 过敏史
    special_disease TEXT,          -- 特殊疾病标识
    alert_tags TEXT DEFAULT '[]',  -- JSON 数组：["allergy","fall_risk","isolation"]
    medical_summary TEXT,          -- 病历摘要（腕带本地存储用）
    status TEXT DEFAULT 'admitted' CHECK (status IN ('admitted', 'discharged', 'dead')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mwp_admission_no ON medical_wristband_patients(admission_no);
CREATE INDEX idx_mwp_status ON medical_wristband_patients(status);
CREATE INDEX idx_mwp_hospital ON medical_wristband_patients(hospital_id);

-- ================================
-- 新增：医用腕带设备表
-- ================================
CREATE TABLE medical_wristband_devices (
    id TEXT PRIMARY KEY,
    device_id TEXT UNIQUE NOT NULL, -- WB-XXXX 格式
    firmware_version TEXT,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'retired')),
    last_seen DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mwd_status ON medical_wristband_devices(status);

-- ================================
-- 新增：患者 - 腕带绑定关系
-- ================================
CREATE TABLE medical_wristband_bindings (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    device_id TEXT NOT NULL REFERENCES medical_wristband_devices(id),
    bound_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    unbound_at DATETIME,
    UNIQUE(patient_id, device_id)
);

CREATE INDEX idx_mwb_patient ON medical_wristband_bindings(patient_id);
CREATE INDEX idx_mwb_device ON medical_wristband_bindings(device_id);

-- ================================
-- 新增：医疗清单 - 费用
-- ================================
CREATE TABLE medical_expenses (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    item_name TEXT NOT NULL,
    item_type TEXT CHECK (item_type IN ('drug', 'equipment', 'test', 'service', 'other')),
    amount REAL DEFAULT 0,
    quantity INTEGER DEFAULT 1,
    unit_price REAL,
    recorded_date DATE,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_me_patient ON medical_expenses(patient_id);
CREATE INDEX idx_me_date ON medical_expenses(recorded_date);

-- ================================
-- 新增：医疗清单 - 用药
-- ================================
CREATE TABLE medical_medications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    drug_name TEXT NOT NULL,
    dosage TEXT,
    frequency TEXT,
    route TEXT CHECK (route IN ('oral', 'iv', 'im', 'sc', 'other')),
    start_date DATE,
    end_date DATE,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'discontinued', 'completed')),
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mm_patient ON medical_medications(patient_id);
CREATE INDEX idx_mm_status ON medical_medications(status);

-- ================================
-- 新增：医疗清单 - 检测报告
-- ================================
CREATE TABLE medical_test_results (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    test_name TEXT NOT NULL,
    test_date DATE,
    result_text TEXT,
    result_value TEXT,
    reference_range TEXT,
    abnormal_flag INTEGER DEFAULT 0 CHECK (abnormal_flag IN (0, 1)),
    attachment_url TEXT,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mtr_patient ON medical_test_results(patient_id);
CREATE INDEX idx_mtr_date ON medical_test_results(test_date);

-- ================================
-- 新增：每日诊疗录入
-- ================================
CREATE TABLE medical_daily_entries (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    entry_date DATE NOT NULL,
    entry_type TEXT CHECK (entry_type IN ('round', 'nursing', 'order', 'other')),
    content TEXT NOT NULL,
    recorded_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mde_patient ON medical_daily_entries(patient_id);
CREATE INDEX idx_mde_date ON medical_daily_entries(entry_date);

-- ================================
-- 新增：护士核验记录
-- ================================
CREATE TABLE medical_verifications (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
    device_id TEXT NOT NULL,
    nurse_user_id TEXT,
    action TEXT NOT NULL CHECK (action IN (
        'check_in', 'check_out', 'give_medication',
        'infusion', 'blood_draw', 'transfusion',
        'test', 'surgery', 'discharge'
    )),
    verified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read_data TEXT,                -- JSON：读取到的腕带数据快照
    result TEXT CHECK (result IN ('success', 'mismatch', 'error')),
    notes TEXT
);

CREATE INDEX idx_mv_patient ON medical_verifications(patient_id);
CREATE INDEX idx_mv_time ON medical_verifications(verified_at);
CREATE INDEX idx_mv_nurse ON medical_verifications(nurse_user_id);

-- ================================
-- 新增：警示标签配置表
-- ================================
CREATE TABLE medical_alert_tag_config (
    id TEXT PRIMARY KEY,
    tag_code TEXT UNIQUE NOT NULL, -- allergy/fall_risk/isolation/coma/...
    tag_name TEXT NOT NULL,
    color TEXT,                    -- 警示颜色代码 #FF0000
    icon TEXT,
    severity TEXT DEFAULT 'P2' CHECK (severity IN ('P0', 'P1', 'P2')),
    enabled INTEGER DEFAULT 1 CHECK (enabled IN (0, 1))
);

-- 预置警示标签
INSERT INTO medical_alert_tag_config (tag_code, tag_name, color, icon, severity) VALUES
('allergy', '过敏', '#FF0000', '⚠️', 'P0'),
('fall_risk', '跌倒高危', '#FF8C00', '⚠️', 'P1'),
('isolation', '隔离患者', '#0000FF', '🔵', 'P1'),
('coma', '昏迷/无表达能力', '#FFFF00', '⚡', 'P0'),
('diet_restriction', '饮食限制', '#800080', '🍽️', 'P2'),
('contact_precaution', '接触隔离', '#008000', '🧤', 'P1');
```

### 6.3 API 接口设计

#### 6.3.1 患者管理接口

```http
# 患者入院登记
POST /api/v1/medical/patients
Request:
{
  "name": "王秀英",
  "gender": 2,
  "age": 78,
  "admission_no": "20260721001",
  "department": "心内科",
  "bed_no": "12",
  "blood_type": "A Rh+",
  "allergy_history": "青霉素、磺胺类",
  "special_disease": "高血压、糖尿病",
  "alert_tags": ["allergy", "fall_risk"]
}
Response: {"id": "uuid", "status": "admitted"}

# 在院患者列表
GET /api/v1/medical/patients?status=admitted&department=&page=&page_size=
Response: {"patients": [...], "total": 100}

# 患者详情
GET /api/v1/medical/patients/:id
Response: {...完整患者信息...}

# 更新患者信息
PUT /api/v1/medical/patients/:id
Request: {"name": "王秀英", "bed_no": "15", ...}
Response: {"id": "uuid", "updated_at": "..."}

# 出院注销
DELETE /api/v1/medical/patients/:id
Request: {"reason": "治愈出院"}
Response: {"id": "uuid", "status": "discharged"}

# 批量导入
POST /api/v1/medical/patients/batch-import
Content-Type: multipart/form-data
File: patients.csv
Response: {"imported": 50, "failed": 2, "errors": [...]}

# 按住院号查询（护士端快速查找）
GET /api/v1/medical/patients/by-admission-no?admission_no=20260721001
```

#### 6.3.2 腕带绑定接口

```http
# 绑定腕带
POST /api/v1/medical/patients/:id/bind
Request: {"device_id": "WB-0001"}
Response: {"binding_id": "uuid", "write_status": "pending"}

# 解绑腕带
POST /api/v1/medical/patients/:id/unbind
Request: {"device_id": "WB-0001"}
Response: {"binding_id": "uuid", "unbound_at": "..."}

# 批量绑定
POST /api/v1/medical/patients/batch-bind
Request: {"bindings": [{"patient_id": "...", "device_id": "..."}]}
Response: {"bound": 10, "failed": 1}

# 写入腕带固件
POST /api/v1/medical/wristbands/:device_id/write
Request: {"patient_id": "...", "force_refresh": false}
Response: {"write_id": "uuid", "status": "sent"}

# 出院清空腕带
POST /api/v1/medical/wristbands/:device_id/clear
Response: {"clear_id": "uuid", "status": "sent"}

# 腕带设备列表
GET /api/v1/medical/wristbands?page=&page_size=&status=
Response: {"devices": [...], "total": 50}

# 腕带固件版本
GET /api/v1/medical/wristbands/:device_id/firmware
Response: {"device_id": "WB-0001", "firmware_version": "1.0.0", "last_ota": "..."}
```

#### 6.3.3 医疗清单接口

```http
# ===== 住院费用 =====
POST /api/v1/medical/lists/expenses
Request: {"patient_id": "...", "item_name": "CT 检查", "item_type": "test", "amount": 350, "recorded_date": "2026-07-21"}
GET /api/v1/medical/lists/expenses?patient_id=&start_date=&end_date=

# ===== 用药清单 =====
POST /api/v1/medical/lists/medications
Request: {"patient_id": "...", "drug_name": "氨氯地平", "dosage": "5mg", "frequency": "qd", "route": "oral", "start_date": "2026-07-21"}
GET /api/v1/medical/lists/medications?patient_id=&status=active

# ===== 检测报告 =====
POST /api/v1/medical/lists/tests
Request: {"patient_id": "...", "test_name": "血常规", "test_date": "2026-07-20", "result_text": "正常", "abnormal_flag": 0}
GET /api/v1/medical/lists/tests?patient_id=&start_date=&end_date=

# ===== 每日诊疗录入 =====
POST /api/v1/medical/daily/entries
Request: {"patient_id": "...", "entry_date": "2026-07-21", "entry_type": "round", "content": "患者血压稳定，继续观察"}
GET /api/v1/medical/daily/entries?patient_id=&date=

# ===== 患者治疗经过（家属端使用）=====
GET /api/v1/medical/history?elderly_id=
Response: {
  "patient": {...},
  "expenses": [...],
  "medications": [...],
  "test_results": [...],
  "daily_entries": [...]
}
```

#### 6.3.4 核验记录接口

```http
# 核验记录列表
GET /api/v1/medical/verifications?patient_id=&nurse_id=&start=&end=&action=&page=&page_size=

# 标记核验完成
PUT /api/v1/medical/verifications/:id/status
Request: {"result": "success", "notes": "信息一致，已给药"}

# 统计今日核验次数
GET /api/v1/medical/verifications/stats/today
Response: {"total": 150, "by_action": {"give_medication": 60, "infusion": 40, ...}}
```

#### 6.3.5 HIS 预留接口（MVP 不启用）

```http
# 触发 HIS 数据同步
POST /api/v1/medical/his/sync
Request: {"hospital_id": "H001", "sync_type": "full"}
Response: {"sync_id": "uuid", "status": "queued"}

# HIS 对接状态
GET /api/v1/medical/his/status
Response: {"connected": false, "last_sync": null, "config": {...}}

# HIS 数据回调（医院 HIS 系统主动推送）
POST /api/v1/medical/his/callback/patient
POST /api/v1/medical/his/callback/expenses
POST /api/v1/medical/his/callback/medications
POST /api/v1/medical/his/callback/tests
```

### 6.4 数据加密与隐私保护

| 保护层 | 机制 | 说明 |
|--------|------|------|
| 传输层 | TLS 1.3 | 所有 API 调用强制 HTTPS |
| 存储层 | AES-256-GCM | 患者敏感字段（过敏史、特殊疾病）单独加密存储 |
| 近场层 | NFC Secure Channel + AES-128-CBC | 护士端读取腕带时加密传输 |
| 访问控制 | JWT + RBAC | 仅医护角色可访问 `/api/v1/medical/` 接口 |
| 审计日志 | 操作记录 | 所有录入、修改、读取、导出操作记录日志 |
| 数据脱敏 | Hash 传输 | 腕带近场读取时仅传输 patient_id_hash，不传输明文 ID |
| 合规要求 | PIPL | 敏感个人信息需单独同意，数据境内存储 |

### 6.5 医疗数据合规设计

| 合规项 | 实现方式 |
|--------|---------|
| 数据最小化 | 腕带本地仅存储身份标识 + 警示标签 + 病历摘要，不存完整医疗清单 |
| 知情同意 | 患者入院时签署《个人信息处理同意书》 |
| 访问权限 | 医护角色仅可查看管辖病房患者信息，不能跨科室查看 |
| 数据留存 | 患者出院后医疗清单保留 30 年（符合《医疗机构病历管理规定》） |
| 数据删除 | 腕带出院时必须清空本地存储，不可恢复 |
| 审计追踪 | 所有数据访问记录操作日志，保留 5 年 |

---

## 第七部分 兼容性保障

### 7.1 不改动范围

以下现有内容**绝对不修改**：

| 范围 | 说明 |
|------|------|
| `firmware/bracelet/` | 现有通用手环固件（Entry/Plus/Pro）全部不动 |
| `firmware/pillbox/` | 现有药盒固件（Basic/Smart/Auto）全部不动 |
| `cloud/gateway/` | MQTT 设备接入网关不动 |
| `cloud/push-service/` | 推送服务不动 |
| `cloud/data-pipeline/` | AI 分析引擎不动 |
| `apps/family-app/` | 现有页面逻辑不动，仅新增"住院治疗"页面 |
| `apps/miniprogram/` | 现有页面逻辑不动，仅新增"住院治疗"页面 |
| `apps/website/` | 品牌官网不动 |
| `b2b/` | B2B 对接层不动 |

### 7.2 兼容性测试要点

| 测试项 | 方法 | 预期结果 |
|--------|------|---------|
| 现有手环 OTA 升级 | 对 Entry/Plus/Pro 设备下发 OTA | 升级成功，新功能不受影响 |
| 现有 MQTT Topic | 手环设备继续上报 `eregen/up/#` 消息 | 正常接收，无冲突 |
| 现有 API 接口 | 调用 `/api/v1/devices/*`、`/api/v1/health/*` | 返回数据正常 |
| 现有家属 APP | 登录、查看定位、健康数据、告警 | 功能正常 |
| 现有管理后台 | 设备管理、用户管理、订阅管理 | 页面正常 |
| 新增医护模块 | 患者入院→绑定腕带→写入→护士读取→核验 | 完整闭环可用 |
| 新增家属医疗页面 | 查看绑定患者治疗经过 | 数据正确展示 |
| SQLite 迁移 | 现有数据从 PG 迁移到 SQLite | 数据完整性 100% |

### 7.3 版本发布规则

| 版本号 | 发布内容 | 影响范围 |
|--------|---------|---------|
| v1.0.0 | 医用腕带固件首版 | 仅医用腕带设备 |
| v1.1.0 | 新增警示标签类型 | 仅医用腕带设备 |
| v2.0.0 | 医用腕带协议大改 | 仅医用腕带设备 + 云端 API |
| cloud-v1.0.0 | 新增医护工作站 API | 管理后台 + 医用模块 |
| cloud-v1.1.0 | 新增 HIS 对接接口 | 管理后台 + 医用模块 |
| admin-v1.0.0 | 医护工作站前端页面上线 | 管理后台 |
| family-v1.0.0 | 住院治疗页面上线 | 家属 APP |
| mini-v1.0.0 | 住院治疗页面上线 | 微信小程序 |

---

## 第八部分 迭代实施计划

### 8.1 开发范围与不改动范围

**本次迭代开发范围：**

| 模块 | 工作内容 |
|------|---------|
| 医用腕带固件 | ESP32-S3 固件开发（患者存储、NFC 服务、Cat1 联网、OLED 显示、LED 警示） |
| 云平台医用模块 | `medical-wb/` 内嵌包（Handler/Service/Store/Router） |
| 管理后台医护工作站 | Vue 3 新增页面（患者登记、信息管理、腕带绑定、每日录入、核验记录、统计看板） |
| 家属 APP | Flutter 新增"住院治疗"页面 |
| 微信小程序 | WXML/WXSS 新增"住院治疗"页面 |
| 数据库迁移 | 现有表迁移到 SQLite，新增医用腕带表族 |
| 近场通讯 | Android NFC API 集成、NDEF 记录定义、认证流程 |

**本次迭代不改动范围：**

| 模块 | 说明 |
|------|------|
| 现有手环固件 | Entry/Plus/Pro 全部不动 |
| 现有药盒固件 | Basic/Smart/Auto 全部不动 |
| 现有云平台核心服务 | gateway/push-service/data-pipeline 不动 |
| 现有 APP 核心页面 | 首页/健康/告警/用药页面不动 |
| 现有管理后台核心页面 | 仪表盘/设备/用户/订阅页面不动 |
| 品牌官网 | 不动 |
| B2B 对接层 | 不动 |

### 8.2 迭代批次划分

```
第一批（第 1-2 周）：基础设施 + 数据层
├── SQLite 数据库搭建（现有表迁移 + 新增医用表）
├── 云平台 medical-wb/ 模块骨架（Router/Store/Service 框架）
└── 医用腕带固件工程初始化（ESP32-S3 + CMake + NVS）

第二批（第 3-4 周）：固件核心功能
├── 患者信息 NVS 存储（写入/读取/加密/防篡改）
├── NFC 服务（Device Info + Patient Identity + Alert Tags + Medical Summary NDEF Records）
├── Cat1 MQTT 下行数据接收（云端下发患者信息到腕带）
└── OLED 显示 + LED 警示灯

第三批（第 5-6 周）：云端 API + 管理后台
├── 患者 CRUD API（入院登记、信息编辑、出院注销）
├── 腕带绑定/解绑/批量写入 API
├── 医疗清单 API（费用/用药/检测/每日录入）
├── 核验记录 API
├── 管理后台医护工作站页面开发
└── Android NFC 近场读取集成

第四批（第 7-8 周）：家属端 + 联调测试
├── 家属 APP"住院治疗"页面
├── 微信小程序"住院治疗"页面
├── 端到端联调（入院→绑定→写入→读取→核验→家属查看）
├── 兼容性测试（现有功能回归测试）
└── 安全测试（NFC 认证、数据加密、访问控制）

第五批（第 9-10 周）：灰度发布 + 文档
├── 内部灰度测试（10% 设备）
├── 用户手册编写（护士操作手册、管理员手册）
├── 固件文档编写（Flash 分区、NFC 协议、OTA 流程）
└── 全量发布准备
```

### 8.3 里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1：数据层就绪 | 第 2 周末 | SQLite 数据库可用，医用表创建完成 |
| M2：固件原型可用 | 第 4 周末 | 腕带可存储/读取患者信息，NFC 可被读取 |
| M3：云端 API 可用 | 第 6 周末 | 患者 CRUD、腕带绑定、医疗清单 API 可调用 |
| M4：管理后台可用 | 第 6 周末 | 医护工作站页面可操作完整闭环 |
| M5：家属端可用 | 第 8 周末 | APP/小程序可查看治疗经过 |
| M6：灰度发布 | 第 10 周末 | 内部测试通过，准备灰度 |

### 8.4 风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|---------|
| 护士终端 NFC 读取不稳定 | 护士无法完成近场读取 | Android PDA/手机 APP 原生 NFC API；管理后台提供"手动输入住院号"模式作为兜底 |
| ESP32-S3 Flash 容量不足 | 无法存储全部患者信息 | 采用压缩存储 + 分块写入，必要时使用外部 SPI Flash |
| SQLite 并发写入冲突 | 多护士同时录入导致数据竞争 | MVP 阶段限制单写多读，未来切 PG 解决 |
| NFC 近场读取距离过近 | 护士操作不便 | NFC 默认 4cm 读取距离，符合医疗操作习惯 |
| 医疗数据合规风险 | 患者隐私泄露 | 严格加密 + 访问控制 + 审计日志，上线前完成合规评审 |

---

## 附录 A：完整目录结构（增量视角）

```
Eregen/
├── firmware/
│   ├── bracelet/                    # 【不动】
│   │   ├── entry/
│   │   ├── plus/
│   │   ├── pro/
│   │   └── common/
│   ├── pillbox/                     # 【不动】
│   │   ├── basic/
│   │   ├── smart/
│   │   ├── auto/
│   │   └── common/
│   └── medical-wristband/           # 【新增】
│       └── esp32s3/
│           ├── main/
│           ├── protocol/
│           ├── ota/
│           ├── crypto/
│           ├── test/
│           └── CMakeLists.txt
│
├── cloud/
│   ├── gateway/                     # 【不动】
│   ├── api-server/                  # 【不动】（Store 层改用 SQLite）
│   ├── push-service/                # 【不动】
│   ├── data-pipeline/               # 【不动】（Store 层改用 SQLite）
│   ├── admin-api/                   # 【不动核心】+ 新增 medical-wb/ 子包
│   └── medical-wb/                  # 【新增】（内嵌于 admin-api）
│
├── apps/
│   ├── admin-web/                   # 【增量】新增医护工作站页面
│   ├── family-app/                  # 【增量】新增住院治疗页面
│   ├── miniprogram/                 # 【增量】新增住院治疗页面
│   └── website/                     # 【不动】
│
├── b2b/                             # 【不动】
│
├── shared/                          # 【不动】（新增 medical protocol 定义）
│
└── docs/
    └── specs/
        ├── 00-global-architecture.md
        ├── 01-bracelet-firmware.md
        ├── 02-pillbox-firmware.md
        ├── 03-cloud-platform.md
        ├── 04-family-app.md
        ├── 05-admin-web.md
        ├── 06-website.md
        ├── 07-miniprogram.md
        ├── 08-b2b-integration.md
        ├── project_total_construction_scheme_v2.md
        └── 2026-07-21-medical-wristband-upgrade-design.md  # 【本文件】
```

---

## 第九部分 监管闭环（Regulatory Closure）

> 来源：`2026-07-22-medical-wristband-regulatory-closure-design.md`  
> 编制日期：2026-07-22  
> 变更类型：新增

### 9.1 背景与目标

现有医用电子腕带方案已覆盖住院患者身份识别、护士近场核验、家属端信息查看等核心功能。本次迭代的目标是在**不改变平台架构、不增加额外硬件投入**的前提下，通过**多角色权限管理 + 业务闭环数据留痕 + 规则引擎**，为卫健委/医保等监管部门提供可操作的监管抓手，形成"病人—家属—医院—监管部门"四方闭环。

**监管要解决的核心问题：**

| 问题 | 现有漏洞 | 本次闭环方案 |
|------|---------|-------------|
| **挂床住院** | 病人办了入院但实际不在院，医院虚报在院天数 | 每日在院核验（护士扫码腕带）+ 连续无核验告警 |
| **虚假入院** | 没有病人就造入院记录骗医保 | 入院登记必须绑定腕带设备，住院号唯一校验 |
| **费用与治疗不匹配** | 开了用药/检查记录但实际未执行 | 每次给药/输液/抽血/检查都需扫码核验并记录 |
| **电子围栏越界** | 病人离开医院无人知晓 | Cat1 基站定位 + 围栏计算 + 超时告警 |
| **转科/出院不及时** | 病人已转科或出院，系统状态未更新 | 转科/出院操作触发腕带信息更新和监管通知 |
| **数据不可追溯** | 监管方看不到完整链条 | 所有环节产生审计日志，监管方可穿透查看 |

### 9.2 涉及角色与视图矩阵

| 功能模块 | 医院管理员 | 科室护士 | 卫健委/监管方 | 家属（APP/小程序） |
|---------|:---------:|:-------:|:-----------:|:----------------:|
| 全院在院患者总览 | ✅ | ❌ | ✅ | ❌ |
| 本科室患者列表 | ✅ | ✅ | ✅ | ❌ |
| 腕带绑定/解绑 | ✅ | ❌ | ❌ | ❌ |
| 每日诊疗录入 | ✅ | ✅ | ❌ | ❌ |
| 费用清单 | ✅ | ❌ | ✅ 只读 | ✅ 简化版 |
| 用药清单 | ✅ | ❌ | ✅ 只读 | ✅ 简化版 |
| 检测报告 | ✅ | ❌ | ✅ 只读 | ✅ 结果摘要 |
| 电子围栏告警 | ✅ | ✅（本科室） | ✅ | ❌ |
| 穿透审计（全链路追溯） | ✅ | ❌ | ✅ | ❌ |
| 统计看板 | ✅ | ✅（本科室） | ✅ | ❌ |
| 规则配置（阈值/围栏） | ✅ | ❌ | ✅ | ❌ |
| 合规审查报表 | ✅ | ❌ | ✅ | ❌ |

### 9.3 管理后台页面结构

在现有管理后台侧边栏"医护工作站"下方新增"监管专区"菜单项：

```
管理后台
├── 仪表盘（不动）
├── 设备管理（不动）
├── 用户管理（不动）
├── 订阅管理（不动）
├── 医护工作站（不动，扩展监管字段）
│   ├── 患者登记
│   ├── 腕带绑定
│   ├── 每日录入
│   └── 核验记录
└── 监管专区（【新增】）
    ├── 在院总览（实时看板）
    ├── 异常告警
    ├── 穿透审计
    ├── 规则配置
    └── 合规审查
```

### 9.4 监管异常检测规则引擎

内置以下检测规则，阈值均可配置：

| 规则编号 | 规则名称 | 检测逻辑 | 告警级别 |
|---------|---------|---------|---------|
| R01 | 挂床住院 | 连续 N 小时无护士扫码核验记录（默认 24h） | 高 |
| R02 | 电子围栏越界 | 定位超出医院围栏且停留 > N 分钟（默认 30min） | 高 |
| R03 | 虚假入院 | 腕带已绑定但 48h 内无核验记录 | 中 |
| R04 | 费用突增 | 单患者日费用 > 科室均值 3 倍 | 中 |
| R05 | 用药与核验不匹配 | 有用药记录但无给药核验记录 | 中 |
| R06 | 频繁转科 | 7 天内转科 > 3 次 | 低 |
| R07 | 腕带异常断开 | 腕带信号突然中断超过 2h | 高 |
| R08 | 长期不在院 | 出院后腕带未清空 | 低 |

**规则引擎实现：** 规则引擎是 `admin-api` 内的一个轻量定时任务，每 5 分钟执行一轮检测。16 条规则（R01-R08 + R_C01-R_C08）合并为同一个 RuleEngine 服务，共享调度器。

```go
// cloud/admin-api/internal/service/rule_engine.go
type RuleEngine struct {
    store store.Store
    rules map[string]RuleFunc  // "R01".."R08", "R_C01".."R_C08"
}
func (e *RuleEngine) Run() { /* ticker every 5min, execute all rules */ }
```

### 9.5 电子围栏定位方案（零额外硬件）

MVP 阶段使用 Cat1 蜂窝网络的基站定位能力，腕带无需增加 GPS 芯片：

| 定位方式 | 精度 | 功耗 | 成本 | 适用场景 |
|---------|------|------|------|---------|
| **基站三角定位（MVP）** | 50m–2km | 低（Cat1 已有） | 零增加 | 医院建筑群围栏（半径 200–500m）足够 |
| GPS（未来可选） | 5–10m | 高 | 需换芯片 | 超大型院区 |

**基站定位实现方式：**
- Cat1 模组（如 EC20/AE910）可通过 AT 指令获取当前服务基站的 CID/TAC/LAC 信息
- 云平台查询基站坐标数据库（如工信部基站数据库或第三方 API），将基站 ID 转换为经纬度
- 使用 Haversine 公式计算患者位置与医院中心的距离，判断是否在围栏内

**GPS + 基站双模定位源感知（Phase 10+）：**

设备上行定位消息新增 `source` 字段（`"gps"` | `"base_station"` | `""` 默认基站）。云端 `FenceCalculator.IsInsideWithSource` 根据定位源自动调整精度容差：

| 定位源 | 精度容差 | 说明 |
|--------|---------|------|
| GPS | ±5m | 卫星定位精度高，容差小 |
| 基站 | ±500m | 蜂窝网络定位精度低，容差大 |

数据库表 `regulatory_location_logs` 新增 `location_source TEXT DEFAULT 'gps'` 列及 `idx_rll_source` 索引。NATS Event envelope 透传 source 字段到规则引擎。

**定位频率策略：**

| 场景 | 上报频率 | 说明 |
|------|---------|------|
| 正常在院 | 每 30 分钟 | 低功耗模式 |
| 接近围栏边界（距围栏 < 500m） | 每 5 分钟 | 自动切换 |
| 已越界 | 每 1 分钟 | 高频追踪 |
| 护士正在核验时 | 实时锁定为 in-fence | 治疗执行期间 |

### 9.6 监管 API 接口

| 路径 | 方法 | 说明 |
|------|------|------|
| `/api/v1/regulatory/dashboard/patient-overview` | GET | 在院总览摘要 |
| `/api/v1/regulatory/dashboard/patient-list` | GET | 在院患者列表 |
| `/api/v1/regulatory/alerts` | GET/POST | 告警列表/创建 |
| `/api/v1/regulatory/alerts/:id/acknowledge` | POST | 确认告警 |
| `/api/v1/regulatory/alerts/:id/resolve` | POST | 标记解决 |
| `/api/v1/regulatory/audit/patient/:id` | GET | 穿透审计全链路 |
| `/api/v1/regulatory/rules` | GET/PUT | 规则配置 |
| `/api/v1/regulatory/fence/config` | GET/POST | 围栏配置 |
| `/api/v1/regulatory/compliance/report` | GET | 合规报表 |

### 9.7 监管数据库表

| 表名 | 说明 |
|------|------|
| `regulatory_fence_config` | 医院围栏配置（中心坐标 + 半径） |
| `regulatory_location_logs` | 位置日志（电子围栏数据来源） |
| `regulatory_alerts` | 监管告警（规则触发 + 处理状态） |
| `user_department_bindings` | 用户-科室绑定（RBAC 扩展） |

扩展 `medical_wristband_patients` 表新增 5 列：`last_verify_at`, `verify_gap_hours`, `fence_status`, `fence_exit_at`, `fence_exit_duration_sec`。

### 9.8 角色权限扩展

在现有 users 表 role 字段扩展两个新角色：

| 角色 | 值 | 说明 |
|------|-----|------|
| nurse | `nurse` | 护士（level 2 = operator 级），可查看本科室患者 |
| regulator | `regulator` | 监管方（level 3 = super_admin 级），可查看全院数据 |

角色顺序：`super_admin > regulator > hospital_admin > nurse > operator > family > elderly`

### 9.9 家属端隐私控制

| 数据类型 | 家属可见内容 | 脱敏规则 |
|---------|------------|---------|
| 用药清单 | 药名 + 用法 + 频次 | 不含剂量详情 |
| 费用清单 | 项目名称 + 金额 | 不含诊断关联 |
| 检测报告 | 正常/异常标记 + 结论摘要 | 不含详细指标 |
| 每日查房记录 | ❌ 不开放 | 隐私数据 |
| 过敏史/特殊疾病 | ❌ 不开放 | 敏感医疗信息 |
| 位置/围栏数据 | ❌ 不开放 | 监管数据 |
| 核验记录 | ❌ 不开放 | 医护操作记录 |

### 9.10 合规依据

本方案设计参考以下法规和标准：

| 法规/标准 | 相关要求 | 本方案对应功能 |
|-----------|---------|-------------|
| 查对制度（国家卫健委） | 医嘱执行、给药、输血、手术、检查前必须核对患者身份 | 护士扫码腕带核验，记录核验日志 |
| 首诊负责制度 | 患者入院必须有明确的接诊记录和责任人 | 入院登记绑定腕带，记录护士账号 |
| 病历管理制度 | 病历书写及时、准确、完整 | 每日诊疗录入，不可篡改 |
| 医保基金监管条例 | 禁止虚构医疗服务、挂床住院、分解住院、串换项目 | 每日核验记录 + 电子围栏 + 规则引擎 R01-R08 |
| 个人信息保护法（PIPL） | 最小必要原则、知情同意、访问控制、数据安全 | AES-256-GCM 存储加密 + TLS 1.3 传输加密 + RBAC |

### 9.11 监管模块代码目录

```
cloud/admin-api/
├── medical_wb/               # 【不动】医用腕带模块
└── regulatory/               # 【新增】监管专区模块
    ├── handler/
    │   ├── dashboard.go          # 在院总览
    │   ├── alert.go              # 异常告警列表/确认
    │   ├── audit.go              # 穿透审计
    │   ├── rule_config.go        # 规则配置 CRUD
    │   └── compliance.go         # 合规审查报表
    ├── service/
    │   ├── fence_svc.go          # 电子围栏计算
    │   ├── engine_svc.go         # 规则引擎定时检测
    │   └── audit_svc.go          # 全链路数据聚合
    ├── store/
    │   └── sqlite.go             # 监管数据查询封装
    └── router/
        └── regulatory_routes.go  # 路由注册
```

---

## 第十部分 社区老人多维数字身份腕带场景

> 来源：`2026-07-23-community-elderly-wristband-scenario-design.md`  
> 编制日期：2026-07-23  
> 变更类型：新增

### 10.1 背景与目标

现有颐贞 Eregen 平台已覆盖两大核心业务线：社区老人安全监护（三档手环）和医院住院患者管理（医用腕带）。本次迭代新增**第三大业务线 — 社区老人多维数字身份场景**：

> 让社区老人佩戴的电子腕带（同构硬件）成为**统一的身份认证凭证**，在社区医院实现特病认证、按需用药、民政补助签到激活等多重功能，同时保留手环原有的安全监护能力。

**核心目标：**

1. **社区医院特病认证** — 老人到社区医院就诊/取药时，药师/护士通过 Web NFC 读取腕带，自动识别其特病身份和福利标签，按需发放药物或提供医疗保障服务
2. **民政补助签到激活** — 老人定期到社区医院签到，一次签到同时激活所有福利标签（医疗+民政+残联+公交），系统批量发放补助资金
3. **多部门联动** — 社区医院作为唯一认证入口，民政局/残联/医保局通过数据接口获取认证结果，无需老人跑多个部门
4. **家属透明化** — 家属可通过 APP/小程序查看老人的福利状态、签到记录、补助领取情况
5. **监管合规** — 民政部门可审计福利发放的合规性，防止冒领、重复领取

### 10.2 管理后台页面结构

在现有管理后台侧边栏"监管专区"下方新增"社区老人专区"菜单项：

```
管理后台
├── 仪表盘（不动）
├── 设备管理（不动）
├── 用户管理（不动）
├── 订阅管理（不动）
├── 医护工作站（不动，住院患者场景）
├── 监管专区（不动，住院监管）
└── 社区老人专区（【新增】）
    ├── 老人档案管理
    ├── 福利标签管理
    ├── 签到总览
    ├── 药房发药记录
    ├── 民政数据导入
    └── 统计看板
```

### 10.3 核心业务流程

#### 流程一：社区医院特病认证 + 按需用药

```
老人持腕带到社区医院 → 药师/护士打开管理后台 → 点击"腕带扫描" → Web NFC 读取腕带
    ↓
腕带返回: dev_id + elder_id_hash + alert_tags[] + welfare_tags[]
    ↓
云端查询: 老人档案 + 福利标签 + 特病用药清单 + 当期签到状态
    ↓
系统返回认证结果: 身份认证通过 / 可领取药物清单 / 可领取医疗保障服务 / 签到状态
    ↓
药师确认发药 → 扫描药物条码 → 系统核销 → 记录发药日志
    ↓
腕带更新: 本地缓存最新用药标签（下次快速显示）
```

#### 流程二：民政补助签到激活

```
老人到社区医院 → 药师扫描腕带 → 触发签到
    ↓
系统检查: 老人当期（月度）签到状态
    ↓
若未签到 → 标记"已签到" → 自动激活本期福利:
  ├── 医疗: 特病用药资格 ✅
  ├── 民政: 贫困补助/特困照顾激活 ✅
  ├── 残联: 残疾补贴激活 ✅
  └── 公交: 当月乘车优惠资格激活 ✅
    ↓
批量发放引擎（每月固定日期执行）: 扫描所有"本月已签到"的老人 → 生成发放清单 → 批量打入银行账户
    ↓
若当月未签到 → 补助暂停 → 次月需补签到才能恢复
```

### 10.4 社区场景检测规则

在现有监管规则引擎（R01-R08）基础上，新增 8 条社区场景检测规则，合并到同一 RuleEngine：

| 规则编号 | 规则名称 | 检测逻辑 | 告警级别 |
|---------|---------|---------|---------|
| R_C01 | 重复领取 | 同一老人同月内在不同社区医院签到 | 高 |
| R_C02 | 跨社区医院互认 | 通过 id_card 字段全局去重，检测同一身份证号在不同机构签到 | 高 |
| R_C03 | 异常高频 | 7 天内签到/取药 > 5 次 | 中 |
| R_C04 | 僵尸账户 | 腕带离线 > 30 天 + 无签到记录 | 低 |
| R_C05 | 补助未到账 | 已激活发放但银行返回失败 | 高 |
| R_C06 | 跨区领取 | 同一福利标签在多个区县被使用 | 中 |
| R_C07 | 福利过期未停 | 福利标签已过有效期但仍在使用 | 中 |
| R_C08 | 死亡未注销 | 老人已出具死亡证明但腕带仍活跃 | 高 |

**跨医院互认实现机制：**

`community_elders` 表的 `id_card` 字段为 UNIQUE NOT NULL。`TriggerSignin` handler 在执行签到时：
1. 查询老人档案获取 id_card
2. 查询当期所有签到记录，若发现同一 id_card 在其他医院已签到 → 触发 R_C01 告警
3. SQLite `community_signin_records` 表通过 `UNIQUE(elder_id, device_id, period)` 复合约束在数据库层面防止重复签到

NATS Event envelope 新增 `HospitalID` 字段，确保跨院签到事件携带医院上下文。

### 10.5 社区 API 接口

| 路径 | 方法 | 说明 |
|------|------|------|
| `/api/v1/community-wb/elders` | CRUD | 老人档案管理 |
| `/api/v1/community-wb/devices` | GET/POST | 腕带设备管理 |
| `/api/v1/community-wb/welfare-tags` | GET/POST/PUT/DELETE | 福利标签配置 |
| `/api/v1/community-wb/signin/trigger` | POST | 签到激活 |
| `/api/v1/community-wb/signin/records` | GET | 签到记录 |
| `/api/v1/community-wb/pharmacy/dispense` | POST | 药房发药 |
| `/api/v1/community-wb/minzheng/import` | POST | 民政数据导入 |
| `/api/v1/community-wb/batch-pay/execute` | POST | 批量发放 |
| `/api/v1/community-wb/batch-payments` | GET | 发放记录 |

### 10.6 社区数据库表

| 表名 | 说明 |
|------|------|
| `community_elders` | 社区老人档案（身份证号唯一标识） |
| `community_wristband_devices` | 社区腕带设备（mode=hospital/community） |
| `community_elder_bindings` | 老人-腕带绑定关系 |
| `community_welfare_tag_config` | 福利标签配置（孤寡/特困/残疾/特病/公交优惠） |
| `community_elder_welfare` | 老人-福利绑定关系（有效期管理） |
| `community_signin_records` | 签到记录（period=YYYY-MM） |
| `community_pharmacy_logs` | 社区药房发药记录 |
| `community_minzheng_sync` | 民政数据同步日志 |
| `community_batch_payments` | 批量发放记录 |

预置 9 个福利标签：`orphan`(孤寡), `poverty_level_1/2`(特困一级/二级), `disability_level_1/2/3`(残疾一至三级), `special_disease`(特病门诊), `bus_discount`(公交优惠), `medical_assistance`(医疗救助)。

### 10.7 后端 Handler 补全

**Welfare Tag CRUD:**
- `CreateWelfareTag` — `POST /admin/community-wb/welfare-tags` — 创建新标签配置（tag_code, tag_name, issuer, renewal_period_days, benefit_amount, enabled）
- `UpdateWelfareTag` — `PUT /admin/community-wb/welfare-tags/:id` — 编辑现有标签
- `DeleteWelfareTag` — `DELETE /admin/community-wb/welfare-tags/:id` — 软删除（仅当无活跃分配时）

**Batch Assign:**
- `BatchAssignWelfareTags` — `POST /admin/community-wb/welfare-tags/batch-assign`
- Request: `{ elder_ids: string[], tag_code: string, valid_from: string, valid_to: string }`
- Creates `CommunityElderWelfare` records for each elder
- Validates tag exists and is enabled; checks for duplicate assignments

**CSV Import Service:**
- `cloud/admin-api/internal/service/minzheng_import.go` — 完成 TemplateAdapter 调用
```go
type CsvImportService struct { store store.Store }
func (s *CsvImportService) ProcessCSV(ctx context.Context, syncID string, file io.Reader) error {
    // 1. Parse CSV with header row
    // 2. Validate required fields (name, id_card)
    // 3. Match against existing elders by ID card (fuzzy match on name)
    // 4. Create/update CommunityElderWelfare records
    // 5. Update sync record counts (imported, matched, pending, errors)
}
```
- CSV 格式: `姓名,身份证号,福利类型,认定等级,有效期开始,有效期结束,备注`

**Minzheng Review Workflow:**
- `ReviewMinzhengMatch` — `POST /admin/community-wb/minzheng/:sync_id/review` — 批准或拒绝待匹配记录
- `ListPendingMinzheng` — `GET /admin/community-wb/minzheng/pending` — 列出待审核匹配

**DB Schema Fix:**
- `community_welfare_tag_config` table migration 添加 `created_at` 和 `updated_at` 列
- 药房发药日志 JSON 修复: `log.Items = "[]"` → 正确 marshal items slice

### 10.8 Admin-web 前端增强 (CommunityWristband.vue)

**福利标签 CRUD Dialogs:**
- Create tag dialog — form with tag_code, tag_name, issuer dropdown, renewal_period_days, benefit_amount, enabled toggle
- Edit tag dialog — pre-filled form, disabled tag_code field
- Delete confirmation — confirm before delete, check for active assignments first

**Batch Assign Dialog:**
- Multi-select elder list (searchable)
- Tag selector dropdown (list of enabled tags)
- Valid from/to date pickers

**Real API Integration:**
- `viewElderDetail()` — 替换硬编码 mock 数据，使用 `Promise.all([getElderWelfareTags, getElderSigninHistory])`
- `handleFileUpload()` — 真实 FormData 上传 CSV → `communityApi.uploadMinzhengCSV(formData)`
- `pharmacyLogs` — 替换为 `communityApi.listPharmacyLogs()`
- `loadSigninRecords()` — 使用真实 API 数据，移除 `generateWeekData()` 硬编码

**Minzheng Pending Review Section:**
- Table of pending matches (name, id_card, suggested_tag, confidence score)
- Approve/Reject buttons per row, bulk approve action

**API Client Updates (`apps/admin-web/src/api/community.ts`):**
新增: `createWelfareTag`, `updateWelfareTag`, `deleteWelfareTag`, `batchAssignWelfareTags`, `uploadMinzhengCSV`, `listPendingMinzheng`, `reviewMinzhengMatch`, `getElderWelfareTags`, `getElderSigninHistory`

### 10.7 社区腕带固件扩展

在 `firmware/medical-wristband/` 目录下新增 `community/` 子目录（独立分支）：

```
firmware/medical-wristband/
├── esp32s3/                    # 【不动】住院模式固件
└── community/                  # 【新增】社区模式固件（独立分支）
    ├── main/
    │   ├── elder_store.c/h         # 老人档案 NVS/Flash 存储
    │   ├── nfc_server.c/h          # NFC 服务（社区医院终端读取）
    │   ├── cat1_mqtt.c/h           # Cat1 联网（下行数据写入）
    │   ├── display_oled.c/h        # OLED 显示（老人姓名+福利标签摘要）
    │   ├── led_indicator.c/h       # LED 福利状态灯
    │   ├── nvs_manager.c/h         # 非易失存储管理
    │   ├── security.c/h            # 数据加密/防篡改
    │   └── board_init.c/h          # 板级初始化
    ├── protocol/                 # 社区专属消息协议
    ├── ota/                      # OTA 更新（复用 hospital 模块）
    └── crypto/                   # 加密（复用 hospital 模块）
```

**NFC NDEF 记录扩展（community 模式）：**

| NDEF Record | Type | 属性 | 数据 |
|---------------|------|------|------|
| Device Info | application/vnd.eregen.device-info | Read | dev_id, fw_version, battery_pct, mode=community |
| Elder Identity | application/vnd.eregen.elder | Read(需认证) | elder_id_hash, name_hash, id_card_hash |
| Alert Tags | application/vnd.eregen.alerts | Read(需认证) | JSON 数组 ["allergy","fall_risk"] |
| Welfare Tags | application/vnd.eregen.welfare | Read(需认证) | JSON 数组 ["orphan","poverty_level_1","disability_level_2"] |
| Medical Summary | application/vnd.eregen.medical-summary | Read(需认证) | UTF-8 文本 ≤256 字节 |
| Auth Challenge | application/vnd.eregen.auth | Read/Write | 随机挑战码 + 配对码验证 |

**NFC 近场认证（Phase 10+）：**

替代原 BLE 配对码方案，使用 NFC NDEF 标签进行腕带身份认证。协议定义在 `shared/protocol/wb_ble.go`：

```go
const NFCServiceUUID = "0000fff0-0000-1000-8000-00805f9b34fb"

type NFCRecordType string

const (
    NFCRecordTypeText  NFCRecordType = "text/utf-8"
    NFCRecordTypeAuth  NFCRecordType = "application/vnd.eregen.auth"
)

type NFCAuthPayload struct {
    ElderID     string `json:"elder_id"`
    SerialNo    string `json:"serial_number,omitempty"`
    Timestamp   int64  `json:"timestamp"`
    Challenge   []byte `json:"challenge,omitempty"`
    Response    []byte `json:"response,omitempty"`
}
```

**NFC NDEF 数据结构：**
- Record 1: `text/utf-8` → elder_id (UUID)
- Record 2: `application/vnd.eregen.auth` → HMAC(elder_id + timestamp, device_key)

管理后台 admin-web 通过 `/admin/community-wb/nfc-auth` API 接口接收 NFC 认证请求。

**OLED 显示扩展：**

```
┌─────────────────────────┐
│  老人姓名                │
│  福利标签:               │
│  ● 孤寡 ● 特困 ● 残疾   │  ← 福利标签滚动显示
├─────────────────────────┤
│  本期签到: 已签到 ✅     │
│  下次签到: 2026-08-23   │
├─────────────────────────┤
│  最后更新: 2026-07-23   │
│  固件版本: 1.0.0-comm   │
└─────────────────────────┘
```

**LED 福利状态灯：**

| 状态 | LED 颜色 | 闪烁模式 | 说明 |
|------|---------|---------|------|
| 本期已签到 | 绿色 | 常亮 | 福利已激活 |
| 本期未签到 | 黄色 | 慢闪（1Hz） | 提醒老人尽快签到 |
| 福利即将过期 | 橙色 | 快闪（2Hz） | 剩余 < 7 天 |
| 福利已过期 | 红色 | 快闪（2Hz） | 需重新认定 |

### 10.8 家属端"父母福利"页面

```
家属 APP / 微信小程序 — "父母福利"页面:

┌─────────────────────────────────────┐
│  父母：张秀兰                       │
│  腕带状态: 在线 ✅                  │
│  最近活动: 2026-07-23 10:30 社区医院│
├─────────────────────────────────────┤
│  福利状态:                          │
│  ● 贫困补助 — 7月已激活 ✅          │
│  ● 残疾补贴 — 7月已激活 ✅          │
│  ● 公交优惠 — 7月已激活 ✅          │
│  ● 医疗救助 — 7月已激活 ✅          │
├─────────────────────────────────────┤
│  签到提醒:                          │
│  下次签到截止: 2026-08-23           │
│  [📱 一键设置提醒]                  │
├─────────────────────────────────────┤
│  补助领取记录:                      │
│  2026-07  贫困补助  ¥800  ✅已发放  │
│  2026-07  残疾补贴  ¥400  ✅已发放  │
└─────────────────────────────────────┘
```

### 10.9 社区模块代码目录

```
cloud/admin-api/
├── medical_wb/               # 【不动】
├── regulatory/               # 【不动】
└── community_wb/             # 【新增】社区老人模块
    ├── handler/
    │   ├── elderly.go            # 老人档案 CRUD
    │   ├── welfare.go            # 福利标签管理
    │   ├── signin.go             # 签到激活
    │   ├── pharmacy.go           # 社区药房发药
    │   └── minzheng.go           # 民政数据导入/发放
    ├── service/
    │   ├── elderly_svc.go        # 老人业务逻辑
    │   ├── welfare_svc.go        # 福利标签同步/匹配
    │   ├── signin_svc.go         # 签到周期管理
    │   ├── pharmacy_svc.go       # 发药核销逻辑
    │   └── batch_pay_svc.go      # 批量发放引擎
    ├── store/
    │   └── sqlite.go             # 社区数据查询封装
    └── router/
        └── community_routes.go   # 路由注册
```

### 10.10 民政数据批量导入

街道/居委会定期（月度/季度）提供 Excel/CSV 模板：

```csv
姓名,身份证号,福利类型,认定等级,有效期开始,有效期结束,备注
张秀兰,510101195001011234,特困,一级,2025-01-01,2028-12-31,
李建国,510101194805055678,残疾,二级,2024-06-01,2027-05-31,肢体残疾
```

导入流程：上传 CSV → 系统校验（身份证号格式、福利类型枚举、日期合法性、重复记录检测、与现有老人档案匹配）→ 匹配成功自动赋予标签 / 匹配失败标记待人工审核 / 格式错误返回错误行号。

### 10.11 批量发放引擎

每月固定日期（如每月 15 日）自动执行：
1. 扫描所有"本月已签到"的老人
2. 生成发放清单（贫困补助 + 残疾补贴 + 特困照顾 + 医疗救助金）
3. 调用银行接口批量打款（成功/失败/重试 3 次后转人工）
4. 生成发放报表推送给民政局管理员

### 10.12 合规依据

| 政策文件 | 发布单位 | 相关要求 | 本方案对应功能 |
|---------|---------|---------|-------------|
| 《社会救助暂行办法》 | 国务院 | 特困人员基本生活最低标准保障 | 贫困补助标签 + 批量发放引擎 |
| 《残疾人保障法》 | 全国人大常委会 | 残疾人两项补贴 | 残疾等级标签 + 补贴激活 |
| 《城市居民最低生活保障条例》 | 国务院 | 低保对象动态管理、定期复核 | 福利标签有效期 + 年审联动 |
| 民政资金发放合规 | 民政部 | 禁止重复领取、冒领、虚报 | R_C01-R_C08 规则引擎 + 签到记录 + 审计日志 |

---

## 附录 C：文档变更履历

| 变更时间 | 位置 | 变更内容 | 变更类型 | 来源编号 |
|---------|------|---------|---------|---------|
| 2026-07-21 | 全文 | 初始版本 — 医用电子腕带升级设计方案 | 新增 | DOC-20260721-MED |
| 2026-07-22 | §9 | 新增监管闭环章节（规则引擎 R01-R08、电子围栏、穿透审计、合规报表） | 新增 | DOC-20260722-REG |
| 2026-07-23 | §10 | 新增社区老人章节（福利标签、签到激活、民政导入、批量发放、规则 R_C01-R_C08） | 新增 | DOC-20260723-COMM |
| 2026-07-23 | §9.5、§10.4、§10.7 | GPS+基站双模定位源感知精度、跨医院 id_card 去重签到、NFC 近场认证协议 | 增量 | DOC-20260723-PHASE10 |

---

## 附录 D：术语表（完整版）

| 术语 | 说明 |
|------|------|
| MVP | Minimum Viable Product，最小可行产品 |
| NFC | Near Field Communication，近场通信 |
| NDEF | NFC Data Exchange Format，NFC 数据交换格式 |
| NVS | Non-Volatile Storage，非易失存储 |
| OTA | Over-The-Air，空中升级 |
| Cat1 | Cellular Type 1，蜂窝通信制式 |
| MQTT | Message Queuing Telemetry Transport，消息队列遥测传输 |
| NATS | Lightweight messaging system |
| HIS | Hospital Information System，医院信息系统 |
| PIPL | Personal Information Protection Law，个人信息保护法 |
| RBAC | Role-Based Access Control，基于角色的访问控制 |
| Haversine | 球面两点间距离计算公式 |
| DRG | Diagnosis Related Groups，按疾病诊断相关分组付费 |
| DIP | Diagnosis-Intervention Packet，按病种分值付费 |
| 挂床住院 | 患者办理入院手续但实际不在院的违规行为 |
| 电子围栏 | 基于地理坐标的虚拟边界，用于监控患者位置 |
| 穿透审计 | 监管部门对任意患者的全链路数据追溯能力 |
| 孤寡 | 无子女、无配偶、无监护人的独居老人 |
| 特困 | 无劳动能力、无生活来源、无法定赡养人的低收入群体 |
| 特病 | 特殊疾病，享受门诊特殊慢性病待遇的疾病类别 |
| 批量发放 | 民政补助资金的定期批量银行转账发放机制 |
| 签到激活 | 老人定期签到以维持福利资格有效的机制 |

---

© 2026 Eregen (颐贞). All rights reserved.
