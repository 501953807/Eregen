# 颐贞 Eregen — 项目整体建设总体方案（修订完整版）

> 编制日期：2026-07-21
> 版本：v2.3（慢性病专项升级 — 手环Pro+版 + 家属APP慢病管理模块）
> 来源：整合 `docs/specs/` 与 `docs/superpowers/specs/` 全部源文档
> 状态：待评审
> 关联文档：`2026-08-08-chronic-care-upgrade-design.md`

---

## 第一部分 项目顶层规划

### 1.1 项目背景与立项依据

**品牌释义：** 颐贞 (Eregen) — "颐养正道，贞守安康"。Eregen = Ease（安心）+ Regen（再生/焕活），英文名传递"让老人安心、让家属放心"的品牌承诺。

**社会背景：** 中国老龄化趋势加速，60-85岁老年人群体对安全监护、用药管理、健康监测的需求日益增长。子女（40-55岁）作为购买决策者，希望实时了解父母的健康状况和安全位置。

**目标用户画像：**

| 角色 | 年龄 | 关系 | 核心诉求 |
|------|------|------|---------|
| 使用者 | 60-85 岁 | 老人 | 安全(防走失)、按时吃药、不被打扰 |
| 购买者 | 40-55 岁 | 子女 | 实时知道父母在哪、有没有按时吃药 |
| 合作方 | — | 医院/社区/养老机构 | 批量管理老人健康数据、降低运营成本 |

**项目性质：** 技术预研可行性开发 — 研究功能可行性和系统落地，非市场性产品。研发阶段成本约 3,000 元（硬件 ~2,220 元 + 首年软件基础设施 ~804 元），不含市场定价、ROI 分析等市场性内容。量产阶段需另行完成 SRRC/CTA/CCC 认证送测。

### 1.2 核心建设理念、设计思想与建设原则

**建设愿景：** 构建一个从硬件采集到云端分析再到应用触达的完整闭环生态，以"手环 + 药盒"双硬件为核心入口，通过订阅服务实现持续价值。

**建设总目标：**

1. 完成三档手环（Starter / Plus / Pro）和三档药盒（Basic / Smart / Auto）的固件研发与 ODM 量产验证
2. 建成 Go + Gin 微服务平台，支撑设备接入、数据存储、AI 分析、多渠道推送
3. 交付家属 APP（Flutter）、管理后台（Vue 3）、微信小程序、品牌官网四个应用端
4. 打通 B2B 医院/社区/保险对接通道
5. 所有核心业务逻辑、通信协议、AI 算法自研闭源，申请专利保护

**建设原则：**

1. **核心代码自研** — 所有业务逻辑、算法、协议自己写，不外包
2. **只写固件，不做硬件设计** — 天线设计、射频调优、电源管理 IC 选型、Sensor 硬件集成全部由 ODM 方案商负责
3. **UI 先效果图后编码** — 所有应用层界面必须先出高保真 HTML 效果图，确认后才可以写 Flutter/Vue 代码
4. **技术选型一次性确定** — 本方案确定的技术选型整个项目周期不变
5. **每个子系统独立开发测试** — 明确的输入/输出边界，可并行推进
6. **混合 ODM 起步** — Phase 1A 公模快速验证，Phase 1B 半定制差异化迭代
7. **开源许可策略（专利申请导向）** — 允许 MIT / BSD-3 / Apache-2.0 / ISC；禁止 GPL / AGPL / LGPL；每个子项目维护 `THIRD-PARTY-LICENSES` 文件
8. **GitHub 仓库仅用于代码版本管理** — 所有文档、CI/CD 工作流、超能力相关配置不提交远程仓库

### 1.3 项目整体建设范围边界

本项目覆盖 9 个子系统，涵盖软件、硬件、供应链全范畴：

| # | 子系统 | 职责 | 输入 | 输出 | 依赖 |
|---|--------|------|------|------|------|
| ① | 手环固件 | 数据采集(SOS/定位/心率/跌倒) | 传感器原始数据 | MQTT消息(位置+健康数据) | 无 |
| ② | 药盒固件 | 自动分药+语音提醒+服药检测 | 云端用药规则 | MQTT消息(服药状态) | ③云平台API |
| ③ | 云平台后端 | 设备接入+数据存储+AI分析+推送 | 设备MQTT消息 | REST API + WebSocket | ①②固件协议 |
| ④ | 家属APP | 实时监控+告警接收+远程配置 | 用户操作 | API调用给③ | ③后端 |
| ⑤ | 管理后台 | 运营人员管理设备/用户/订阅；医护工作站（入院登记+信息录入） | 运营操作 + 护士操作 | API调用给③ | ③后端 |
| ⑥ | 品牌官网 | Eregen品牌宣传+产品展示 | — | 静态页面 | 无 |
| ⑦ | 微信小程序 | 轻量版家属端(提醒+查看) | 用户操作 | API调用给③ | ③后端 |
| ⑧ | B2B对接 | 医院/社区/养老机构数据互通 | HIS系统数据 | 标准化健康报告 | ③后端 |
| ⑨ | 医用电子腕带 | 住院患者身份识别+医护近场核验+风险警示 | 患者入院信息 + 云端下发指令 | NFC为主、Cat1为辅的双模式（NFC近场读取+Cat1背景状态上报） | ③云平台API |
| **⑩** | **护士核验终端** | **Flutter移动应用（手机双形态）；巡房/输液/喂药时通过近场核验技术，使用NFC读取腕带信息，核验病人身份一致性** | **护士登录凭证 + 腕带NFC数据** | **核验结果（一致/不一致）+ 诊疗记录写入云端** | **⑨腕带固件(NFC为主 + Cat1为辅) + ③云平台API** |

**适用边界：**

- 覆盖上述 9 个子系统
- 研发阶段成本约 3,000 元，不含市场定价、ROI 分析等市场性内容
- MVP 阶段定位为"生活方式健康设备"，不宣称医疗诊断功能
- 量产阶段需另行完成 SRRC/CTA/CCC 认证送测

### 1.4 短期 + 中期 + 长期分层建设目标

#### 1.4.1 实施批次

```
第一批 (月1-6):  ①手环固件 → ③云平台后端 ← ②药盒固件
                     ↓            ↓            ↓
第二批 (月4-8):            ④家属APP    ⑤管理后台
                                    ↓
第三批 (月7-12):              ⑦小程序   ⑥官网
                                    ↓
                               ⑧B2B医院对接
                                    ↓
第四批 (月1-8):    ①④升级 → 慢性病专项（Pro+手环 + 慢病管理APP）
                  ③新增 → 慢病分析引擎 + 数据模型
```

**关键路径：** ①+②+③并行推进，第 4 个月开始④联调。

### 第四批实施时间线（慢性病专项）

```
Phase 4.1 (月1-3):  手环Pro+固件开发
                    电化学检测模块驱动、微流控控制、试纸阻抗识别算法
                    BLE 5.3血压计配件连接协议
                        ↓
Phase 4.2 (月2-4):  家属APP慢病管理模块
                    7个新页面（慢病主页/血糖详情/尿酸详情/血压详情/饮食记录/运动追踪/健康报告）
                    3个改造页面（首页/健康看板/用药管理）
                        ↓
Phase 4.3 (月3-5):  后端API + AI分析引擎
                    7张新表（glucose/uric_acid/bp/diet/exercise/daily_tasks/health_reports）
                    15个API端点（/api/v1/chronic/*）
                    血糖/尿酸/血压/饮食/运动分析器 + 综合建议引擎
                        ↓
Phase 4.4 (月4-6):  试纸供应链 + 二类医疗器械注册
                    试纸ODM代工对接（电极专利设计+微流控结构）
                    二类医疗器械注册申请（与试纸一起注册）
                    试纸包装设计 + 订阅配送系统开发
                        ↓
Phase 4.5 (月5-6):  血压配件开发
                    外置上臂式蓝牙血压计选型/定制
                    BLE连接协议开发 + 数据同步测试
                    与手环Pro+集成测试
                        ↓
Phase 4.6 (月6-8):  端到端联调 + UI原型确认 + 用户测试
                    试纸检测端到端测试（试纸→手环→云端→APP）
                    UI原型确认（7个新页面）
                    5-10人用户测试
```

**关键路径：** 4.1手环固件 ↔ 4.2APP模块（并行），4.3后端API（依赖4.1），4.4注册与4.5配件开发（并行），4.6联调整合全部。

**关键路径：** ①+②+③并行推进，第 4 个月开始④联调。

#### 1.4.2 月度里程碑

| 阶段 | 时间 | 手环 | 药盒 | APP/后端 | 里程碑 |
|------|------|------|------|---------|--------|
| M1-M2 | 第 1-2 月 | 需求评审+ODM选型 | 需求评审+ODM选型 | 产品原型+UI设计 | 完成 PRD 文档和设计稿 |
| M3-M4 | 第 3-4 月 | ID设计+开模+EVT | 结构设计+开模+EVT | APP MVP开发(定位+SOS+心率) | 手环/药盒手板验证通过 |
| M5-M6 | 第 5-6 月 | DVT+PVT+小批量 | DVT+PVT+小批量 | APP标准版开发(用药提醒+家属共享) | 硬件送测+APP内测 |
| M7-M8 | 第 7-8 月 | 量产+上市 | 量产+上市 | APP正式上线+微信小程序 | 双产品同步上市 |

#### 1.4.3 四阶段战略演进

| 阶段 | 时间 | 策略 | 目标 |
|------|------|------|------|
| Stage 1 | 0-8 个月 | 快速上市验证 | 手环公模 ODM + 药盒半定制 ODM + APP MVP，获客 5,000 用户 |
| Stage 2 | 8-18 个月 | 产品迭代+订阅变现 | 手环升级为半定制，药盒推出自动分药版，付费转化 15% |
| Stage 3 | 18-36 个月 | 自研转型+生态扩展 | 手环转为自研(申请二类医疗器械认证)，接入在线问诊/药品配送 |
| Stage 4 | 36 个月+ | 平台化+B2B2C | 开放 API 接入养老机构/保险公司/社区医院 |

**慢性病专项升级（Stage 2 前置）：** 在 Stage 1 基础上市后，启动慢性病专项升级（见 `docs/superpowers/specs/2026-08-08-chronic-care-upgrade-design.md`）：
- 手环 Pro 版升级为 Pro+ 版，新增可拆卸电化学检测模块（血糖/尿酸试纸检测）
- 新增外置蓝牙血压计配件
- 家属APP新增慢病管理模块（7个新页面 + 3个改造页面）
- 后端新增慢病数据模型、API接口和AI分析引擎
- 试纸耗材订阅商业模式

### 1.5 项目落地价值、行业定位与未来发展前景

**商业模式：硬件获客 + 订阅盈利。**

| 版本 | 价格 | 包含功能 |
|------|------|---------|
| 基础版(免费) | 0 元/月 | 实时定位 + 电子围栏+SOS+基础健康数据 + 用药提醒 |
| 高级版 | 19 元/月 | 基础版 + 健康周报 + 用药记录分析 + 跌倒检测增强 + 历史轨迹 30 天 |
| 尊享版 | 39 元/月 | 高级版 + 视频通话 + 医生咨询 + 健康档案 + 保险权益 + 历史轨迹永久 |

**收入测算：** 获客 10,000 用户 → 付费转化率 15% → ARPU 29 元/月 → 月收入 4.35 万；续费率 30% × 平均使用 24 月 × 29 元/月 = ~174 元/用户 LTV；10,000 用户 × 15% × 174 元 = ~26.1 万元/年订阅收入(保守)。

**行业定位：** 面向中国老年群体的智能健康监测平台，以"生活方式健康设备"为 MVP 阶段定位，不宣称医疗诊断功能。禁用词汇："治愈""疗效""治疗""医疗级""医用"；可使用："运动监测""日常健康数据记录""生活方式参考"。

**中长期前景：** 从 Stage 1 快速上市验证，经 Stage 2 产品迭代 + 订阅变现，到 Stage 3 自研转型 + 生态扩展（申请二类医疗器械认证、接入在线问诊/药品配送），最终在 Stage 4 实现平台化 + B2B2C（开放 API 接入养老机构/保险公司/社区医院）。

---

## 第二部分 软件系统整体建设方案

### 2.1 软件整体技术架构、分层设计、拓扑说明

#### 2.1.1 四层系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层 — 用户交互                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐ │
│  │ 老人端    │  │ 家属端    │  │ 运营端    │  │ B2B合作端    │ │
│  │ 手环/药盒 │  │ APP/小程序│  │ 管理后台  │  │ 医院/社区系统│ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                    平台层 — 云端服务                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │设备管理   │  │消息队列   │  │数据存储   │  │ AI引擎   │   │
│  │OTA升级   │  │NATS总线   │  │SQLite    │  │跌倒/用药  │   │
│  │心跳监控   │  │告警路由   │  │SQLite存储 │  │ 异常识别  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              RESTful API + WebSocket                  │   │
│  └──────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                    传输层 — 数据通道                          │
│  ┌────────┐ ┌─────┐ ┌──────┐ ┌────────┐ ┌──────────────┐  │
│  │ Cat1   │ │BLE  │ │ WiFi │ │4G/5G   │ │ 微信小程序    │  │
│  │蜂窝网络 │ │局域网│ │家庭网│ │移动网络 │ │ 订阅消息      │  │
│  └────────┘ └─────┘ └──────┘ └────────┘ └──────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    感知层 — 数据采集                          │
│  ┌──────────┐  ┌──────────┐                                │
│  │ 智能手环  │  │ 智能药盒  │                                │
│  │GPS/PPG/  │  │步进电机/  │                                │
│  │IMU/Cat1  │  │光电/TTS   │                                │
│  └──────────┘  └──────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

#### 2.1.2 云平台微服务拆分

```
cloud/
├── gateway/              # MQTT设备接入网关（不动）
│   ├── mqtt_listener.go  # 连接EMQX，处理设备消息
│   ├── auth.go           # 设备认证(JWT Token)
│   ├── protocol.go       # 协议解析/校验
│   └── nats_publisher.go # 转发到NATS消息总线
│
├── api-server/           # REST API服务（不动核心，Store层改用SQLite）
│   ├── handler/          # HTTP请求处理器
│   │   ├── auth.go       # 登录/注册/OTP
│   │   ├── user.go       # 用户CRUD
│   │   ├── device.go     # 设备管理
│   │   ├── health.go     # 健康数据
│   │   ├── location.go   # 定位与轨迹
│   │   ├── medication.go # 用药规则
│   │   ├── alert.go      # 告警管理
│   │   └── subscription.go # 订阅管理
│   ├── service/          # 业务逻辑层
│   │   ├── services.go
│   │   └── nats_client.go
│   ├── store/            # 数据访问层（全项目 SQLite）
│   │   └── sqlite.go     # SQLite 实现
│   ├── middleware/       # 中间件
│   │   ├── auth.go       # JWT鉴权+RBAC
│   │   ├── ratelimit.go  # 限流
│   │   └── cors.go       # 跨域
│   └── router/           # 路由注册
│
├── data-pipeline/        # AI分析引擎（不动核心，Store层改用SQLite）
│   ├── subscriber/nats_subscriber.go
│   ├── analyzer/health_analyzer.go
│   ├── analyzer/risk_calculator.go
│   └── model/model.go
│
├── push-service/         # 推送分发服务（不动）
│   ├── publisher/nats_subscriber.go
│   ├── router/router.go
│   ├── channel/wechat/client.go
│   ├── channel/sms/client.go
│   └── fcm/client.go
│
└── admin-api/            # 管理后台专用API + 医用腕带内嵌模块
    ├── device_mgmt.go
    ├── user_mgmt.go
    ├── subscription_mgmt.go
    ├── report_gen.go
    └── medical_wb/         # 【新增】医用腕带内嵌模块
        ├── handler/
        │   ├── patient.go         # 患者CRUD
        │   ├── wristband.go       # 绑定/解绑/写入/清空
        │   ├── medical_list.go    # 费用/用药/检测清单
        │   ├── daily_entry.go     # 每日诊疗录入
        │   └── verification.go    # 核验记录
        ├── service/
        │   ├── patient_svc.go
        │   ├── wristband_svc.go
        │   ├── medical_list_svc.go
        │   └── verification_svc.go
        ├── store/
        │   └── sqlite.go          # SQLite操作封装
        └── router/
            └── medical_routes.go  # 路由注册
```

#### 2.1.3 部署架构

```
                    ┌─────────────────┐
                    │   Nginx/SLB     │
                    │  (负载均衡)       │
                    └──────┬──────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ API-Server│ │ Admin-API│ │ Medical-WB│
        │ :8080    │ │ :8081    │ │ (内嵌)    │
        └────┬─────┘ └────┬─────┘ └────┬─────┘
             │             │             │
             └─────────────┼─────────────┘
                           ▼
              ┌─────────────────────────┐
              │      NATS Cluster        │
              │   (消息总线)              │
              └────┬───────┬───────┬────┘
                   ▼       ▼       ▼
              ┌──────┐ ┌──────┐ ┌────────┐
              │Gateway│ │Pipeline│ │PushSvc │
              └──┬───┘ └──┬───┘ └───┬────┘
                 ▼        ▼         ▼
              ┌─────────────────────────────┐
              │         EMQX Cluster          │
              │      (MQTT Broker)            │
              └───────────┬───────────────────┘
                          ▼
              ┌─────────────────────────────┐
              │    设备层 (手环/药盒/腕带)      │
              └─────────────────────────────┘

        ┌───────────────────────────────────┐
        │         SQLite (MVP阶段)            │
        │  eregen.db: 现有业务表 + 医用表     │
        │  eregen_medical.db: 医用独立库      │
        └───────────────────────────────────┘
```

#### 2.1.4 网络与安全架构

```
                        ┌─────────────────────────┐
                        │    公网 (Internet)       │
                        └───────────┬─────────────┘
                                    │
              ┌─────────────────────┼─────────────────────┐
              ▼                     ▼                     ▼
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │   Cat1 蜂窝网络  │   │    WiFi 网络     │   │   4G/5G 移动网络 │
    │  (手环设备)      │   │  (药盒设备)      │   │  (APP用户端)     │
    └────────┬────────┘   └────────┬────────┘   └────────┬────────┘
             │                     │                     │
             ▼                     ▼                     ▼
    ┌─────────────────────────────────────────────────────────┐
    │              阿里云 VPC (内网隔离)                        │
    │                                                         │
    │  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐  │
    │  │ DMZ区     │  │ 应用区    │  │    数据区             │  │
    │  │          │  │          │  │                      │  │
    │  │ EMQX     │  │ Go微服务 │  │  SQLite              │  │
    │  │ NATS     │  │ (网关/API│  │  审计日志            │  │
    │  │ 反向代理  │  │  推送)   │  │                      │  │
    │  └──────────┘  └──────────┘  └──────────────────────┘  │
    │                                                         │
    │  安全组规则: DMZ→应用区(MQTT/HTTP), 应用区→数据区(SQLite)│
    └─────────────────────────────────────────────────────────┘
                                    │
              ┌─────────────────────┼─────────────────────┐
              ▼                     ▼                     ▼
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │  家属APP/小程序   │   │   管理后台       │   │    B2B合作方     │
    │  (外网访问)       │   │  (VPN/白名单)    │   │  (API Key认证)   │
    └─────────────────┘   └─────────────────┘   └─────────────────┘
```

**安全机制：**

| 层级 | 加密方式 | 说明 |
|------|---------|------|
| 设备→云端 | MQTT over TLS 1.2 | 所有设备消息加密传输 |
| APP→云端 | HTTPS(TLS 1.2+) | 所有API调用加密 |
| 云端内部 | NATS + mTLS | 微服务间双向认证 |
| 静态数据 | AES-256-GCM | SQLite文件级加密 |
| 密钥管理 | 阿里云KMS | 密钥托管与轮转 |

**认证体系：**

- 设备认证：Ed25519 公钥签名 + AES-256-GCM 加密传输
- API 认证：JWT Token + OTP 短信/邮箱验证
- B2B 认证：API Key + IP 白名单
- 数据传输：TLS 1.3 (外网)、mTLS (微服务间)
- 数据存储：AES-256 静态加密 + 敏感字段单独加密

### 2.2 业务模块划分、各模块职责、内部交互逻辑

#### 2.2.1 子系统依赖矩阵

| 子系统 | 依赖的下游 | 被哪些上游依赖 | 外部依赖 |
|--------|-----------|--------------|---------|
| 手环固件 | 云平台(通过MQTT) | 家属APP(数据源)、管理后台(设备管理) | Cat1模组、GPS、PPG、IMU |
| 药盒固件 | 云平台(通过MQTT) | 家属APP(用药管理)、医院(HIS对接) | ESP32-C3、WiFi、步进电机、TTS |
| gateway | NATS总线、SQLite | 所有其他云服务和应用端 | EMQX MQTT Broker |
| api-server | SQLite、NATS | 家属APP、管理后台、小程序、B2B | — |
| push-service | NATS | 家属APP(推送接收) | FCM、阿里云SMS、微信API |
| data-pipeline | SQLite、NATS | api-server(评分结果)、push-service(告警) | — |
| admin-api/medical_wb | SQLite、NATS、MQTT | 医护工作站、家属APP/小程序住院治疗页 | BLE网关(ESP32-S3) |
| 家属APP | api-server(REST) | 无 | 高德地图SDK、FCM SDK |
| 管理后台 | api-server + admin-api(REST) | 无 | Element Plus组件库、NFC近场验证接口（Flutter移动应用已实现） |
| 微信小程序 | api-server(REST) | 无 | 微信SDK、微信订阅消息API |
| 品牌官网 | 静态文件托管 | 无 | Hugo、Tailwind CSS |
| B2B医院API | SQLite | 医院HIS系统 | (待实现) |
| B2B社区平台 | SQLite | 社区护理系统 | (待实现) |
| B2B保险对接 | SQLite | 保险公司系统 | (待实现) |

#### 2.2.2 数据流闭环

**手环数据流：**

```
[手环传感器]
    ├── GPS/北斗 → 每30秒 → Cat1蜂窝 → EMQX MQTT → Go设备接入
    ├── PPG(心率/血氧) → 每5分钟 → Cat1蜂窝 → EMQX MQTT → Go设备接入
    ├── IMU(加速度/陀螺仪) → 100Hz采样 → 端侧AI推理 → 异常时触发
    │                               ├── 跌倒检测(F1>0.92) → P0告警 → WebSocket推送
    │                               └── 异常姿态 → 端侧上报 → 云端复核
    └── SOS实体按键 → 立即触发 → Cat1蜂窝 → EMQX MQTT → Go设备接入
                            ↓
                    [定位+坐标] → P0告警 → WebSocket推送
                            ↓
                    [家属APP] ←→ [管理后台]
```

**药盒数据流：**

```
[云端下发用药规则]
    └── MQTT pub → emergen/device/{dev_id}/rule
            ↓
    [ESP32-C3 药盒固件]
        ├── 定时触发语音TTS提醒
        ├── 步进电机分药 → 光电传感器检测取药
        ├── 服药状态 → MQTT pub → emergen/device/{dev_id}/status
        │       ├── taken=true → 记录服药时间
        │       ├── taken=false → 漏服告警 → 通知家属APP
        │       └── compartment_empty → 库存预警
        └── BLE 5.0 ↔ 手环(局域网，功耗<1mA)
                └── 共享健康数据(心率/步数)
```

**数据生命周期：**

```
设备采集 → MQTT接入 → NATS总线 → AI分析 → 分级存储 → 应用消费
  ↓                                    ↓            ↓
设备采集数据                   健康评分/风险等级    SQLite(结构化)
【MVP阶段全项目使用SQLite，统一存储】
                                           删除策略
```

**医用腕带数据流：**

```
[医用手环ESP32-S3]──Cat1蜂窝──→[EMQX MQTT]──→[admin-api/medical_wb]
[护士核验终端(Flutter移动应用)]──NFC近场──→[本地NFC读取腕带]──→[云端API查询完整信息]
                                                        ↓
                                            ┌───────────┼───────────┐
                                            ↓           ↓           ↓
                                      [SQLite]    [NATS事件]    [审计日志]
                                      患者/腕带   eregen.medical.wb.#
                                      清单/核验                                              ↓
                                          [管理后台医护工作站] ←→ [家属APP/小程序住院治疗页]
```

**护士核验终端形态：**

| 形态 | 设备类型 | 通信方式 | 适用场景 |
|------|---------|---------|---------|
| **Flutter移动应用** | Flutter移动应用（iOS/Android手机） | NFC近场通讯 | 巡房、输液、喂药、抽血、手术核对 |
| **手机** | 护士个人手机/医院配发手机 | NFC近场通讯 | 移动护理、床旁操作 |

**核验流程：**
```
护士打开Flutter移动应用 → 选择患者或扫描腕带条码 → 靠近病人腕带（≤5cm）
    ↓
腕带NFC广播患者ID + 风险标签
    ↓
终端本地解密读取 → 与当前医嘱/护理任务比对
    ↓
一致 → ✅ 绿色通过 + 记录核验日志
不一致 → ❌ 红色告警 + 通知护士长 + 记录异常
    ↓
写入云端 verification 表 → NATS事件推送
```

#### 2.2.3 通信协议定义

**上行(设备→云端)：**

```jsonc
// 心跳包 (每5分钟)
{"type":"heartbeat","dev_id":"BR-XXXX","bat":85,"fw_ver":"1.0.0"}

// 定位数据 (正常30s/次，告警时1s/次)
{"type":"location","dev_id":"BR-XXXX","lat":31.2304,"lon":121.4737,"acc":5,"ts":1720000000}

// 健康数据 (每5分钟)
{"type":"health","dev_id":"BR-XXXX","hr":72,"spo2":98,"step":3456,"cal":120,"ts":1720000000}

// SOS告警 (立即)
{"type":"sos","dev_id":"BR-XXXX","lat":31.2304,"lon":121.4737,"ts":1720000000}

// 跌倒检测 (立即)
{"type":"fall","dev_id":"BR-XXXX","conf":0.95,"lat":31.2304,"lon":121.4737,"ts":1720000000}

// 药盒状态 (定时+事件驱动)
{"type":"med_status","dev_id":"PX-XXXX","compartment":3,"taken":true,"ts":1720000000}

// 药盒库存预警
{"type":"med_inventory","dev_id":"PX-XXXX","compartment":3,"level":"low","ts":1720000000}
```

**下行(云端→设备)：**

```jsonc
// 用药规则 (云端下发)
{"type":"med_rule","dev_id":"PX-XXXX","rules":[{"time":"08:00","dose":1,"type":"capsule","name":"氨氯地平"}]}

// 配置更新
{"type":"config","dev_id":"BR-XXXX","settings":{"interval":30,"volume":80,"gps_mode":"balanced"}}

// 语音播报
{"type":"tts","dev_id":"PX-XXXX","text":"爷爷，该吃降压药了"}

// OTA升级
{"type":"ota","dev_id":"BR-XXXX","url":"https://ota.eregen.cn/firmware/v1.1.0.bin","hash":"sha256:abc123...","size":524288,"reboot":true}

// 电子围栏配置
{"type":"geofence","dev_id":"BR-XXXX","fences":[{"id":"home","type":"circle","lat":31.2304,"lon":121.4737,"radius":200}]}
```

### 2.3 技术栈选型、运行环境、接口规范、迭代建设计划

#### 2.3.1 绝对不可变更的技术选型

**硬件固件层：**

| 子系统 | 主控芯片 | RTOS/SDK | 语言 |
|--------|---------|----------|------|
| 手环(初/中/高端) | GD32E230C8T3 (ARM Cortex-M4) | FreeRTOS | C |
| 药盒(基础/智能/自动) | ESP32-C3 (RISC-V) | ESP-IDF v5.3 | C |

**云平台层：**

| 组件 | 选型 | 版本 | 备注 |
|------|------|------|------|
| 语言/框架 | Go + Gin | Go 1.22+, Gin 1.9+ | — |
| MQTT Broker | EMQX | 5.x (开源版) | — |
| 消息总线 | NATS | 2.10+ | — |
| 用户数据库 | SQLite | — | MVP 全项目切换，替代 PG/InfluxDB/Redis |
| 推送 | FCM + 阿里云SMS + 微信订阅消息 | — | — |

**MVP 阶段 SQLite 降级说明：**

| 维度 | PostgreSQL | SQLite |
|------|-----------|--------|
| 部署复杂度 | 需安装数据库服务 | 单文件，零部署 |
| 运维成本 | 需备份、监控、调优 | 无需运维 |
| 适合场景 | 生产环境、多用户并发 | MVP/原型/小规模 |
| 迁移成本 | — | Store 层封装后迁移成本低 |

迁移策略：所有数据库操作封装在 `store/` 层，业务逻辑不直接依赖 SQL 方言。未来切换到 PostgreSQL 时，只需替换 `sqlite.go` 为 `postgres.go`，Handler 和 Service 层无需改动。表结构使用标准 SQL 语法，避免 SQLite 特有语法。

**应用层：**

| 子系统 | 选型 | 版本 |
|--------|------|------|
| 家属APP | Flutter (Dart) | 3.24+ |
| 管理后台 | Vue 3 + TypeScript + Element Plus | Vue 3.4+, TS 5.4+, Element Plus 2.7+ |
| 微信小程序 | 原生WXML/WXSS | 基础库 2.44+ |
| 品牌官网 | Hugo + Tailwind CSS | Hugo 0.128+, Tailwind 3.4+ |

**基础设施：**

| 组件 | 选型 |
|------|------|
| CI/CD | GitHub Actions |
| 容器化 | Docker + Docker Compose |
| 日志 | Loki + Grafana |
| 监控 | Prometheus + Grafana |
| 域名/SSL | 阿里云DNS + Let's Encrypt |

#### 2.3.2 各子系统技术栈详情

**家属 APP (Flutter)：**

| 库 | 用途 |
|----|------|
| Hive | 本地缓存 |
| Google Maps / 高德地图 | 实时定位地图 |
| ECharts | 健康数据图表 |
| WebSocket | 实时告警推送 |

**管理后台 (Vue 3)：**

| 库 | 版本 | 用途 |
|----|------|------|
| Vue | 3.4+ | 响应式 UI 框架 |
| TypeScript | 5.4+ | 类型安全 |
| Element Plus | 2.7+ | UI 组件库 |
| Vue Router | 4.x | 路由管理 |
| Pinia | 2.x | 状态管理 |
| Axios | 1.x | HTTP 客户端 |
| ECharts | 5.x | 数据可视化图表 |
| Vite | 5.x | 构建工具 |

**品牌官网 (Hugo)：**

| 库 | 版本 | 用途 |
|----|------|------|
| Hugo | 0.128+ | 静态站点生成器 |
| Tailwind CSS | 3.4+ | 原子化 CSS 框架 |
| Alpine.js | 3.x | 轻量交互 (导航切换/滚动动画) |
| Google Fonts | — | Noto Sans SC (中文正文字体) |

**微信小程序：**

| 限制项 | 上限 | 应对策略 |
|--------|------|---------|
| 包体积 | 主包 2MB | 图片压缩、分包加载 |
| 总大小 | 20MB | 使用分包 (每个 ≤ 2MB) |
| 网络请求 | HTTPS 域名白名单 | 配置服务器域名 |
| 本地存储 | 10MB | 仅缓存 token + 必要数据 |
| 定位精度 | 约 10 米 | 足够电子围栏使用 |

#### 2.3.3 接口规范

**设备 → 云平台接口：**

| 接口 | 协议 | 方向 | 频率 | 说明 |
|------|------|------|------|------|
| MQTT `eregen/up/{type}/{id}/message` | MQTT 3.1.1/5.0 | 上行 | 心跳: 5min/次; 定位: 30s/次; 健康: 按需 | 设备数据上报 |
| MQTT `eregen/down/{dev_id}/command` | MQTT 3.1.1/5.0 | 下行 | 按需 | 云端指令下发 |
| HTTP OTA 下载 | HTTPS GET | 下行 | 按需 | 固件升级包下载 |

**云平台内部接口：**

| 接口 | 协议 | 来源 | 目标 | 说明 |
|------|------|------|------|------|
| NATS `eregen.device.*` | NATS JetStream | gateway | api-server/push/pipeline | 设备事件总线 |
| NATS `eregen.alert.p0` | NATS JetStream | pipeline | push-service | P0级告警(跌倒/SOS) |
| NATS `eregen.alert.p1` | NATS JetStream | pipeline | push-service | P1级告警(漏服/围栏) |
| NATS `eregen.health.sync` | NATS JetStream | pipeline | api-server | 健康数据同步 |
| SQLite | 文件存储 | 所有服务 | SQLite | 统一业务数据存储 |
| SQLite wb_*表 | 文件存储 | admin-api | SQLite | 医疗腕带专项数据 |

**云平台 → 应用端接口：**

| 接口 | 协议 | 目标子系统 | 认证方式 | 说明 |
|------|------|-----------|---------|------|
| `POST /api/v1/auth/login` | HTTPS | 家属APP/管理后台/小程序 | JWT | 用户登录 |
| `GET /api/v1/elderly/:id/location` | HTTPS | 家属APP | JWT | 实时定位 |
| `GET /api/v1/health?elderly_id=&start=&end=` | HTTPS | 家属APP | JWT | 健康数据查询 |
| `POST /api/v1/medication/rules` | HTTPS | 家属APP | JWT | 用药规则配置 |
| `GET /api/v1/alerts?severity=&status=` | HTTPS | 家属APP/管理后台 | JWT | 告警列表 |
| `GET /api/v1/admin/devices` | HTTPS | 管理后台 | JWT + RBAC | 设备管理 |
| `WS /api/v1/stream/alerts` | WSS | 家属APP | JWT | 实时告警推送 |

**B2B 接口：**

| 接口 | 协议 | 目标 | 认证 | 说明 |
|------|------|------|------|------|
| `POST /api/v2/b2b/hospitals/health-data` | HTTPS | 医院HIS | API Key + mTLS | 健康数据导出 |
| `GET /api/v2/b2b/hospitals/medical-rules` | HTTPS | 医院HIS | API Key + mTLS | 用药规则查询 |
| `POST /api/v2/b2b/communities/events` | HTTPS | 社区平台 | API Key | 活动创建 |
| `POST /api/v2/b2b/insurance/claims` | HTTPS | 保险公司 | API Key + IP白名单 | 理赔申请 |

#### 2.3.4 数据库设计

**MVP 阶段全项目统一使用 SQLite：**

**用户表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PRIMARY KEY | UUID |
| phone | VARCHAR(20) UNIQUE | 手机号 |
| name | VARCHAR(50) | 姓名 |
| role | TEXT | 用户类型: elder/family/operator/nurse |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

**老人档案表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PRIMARY KEY | UUID |
| user_id | TEXT → users.id | 关联家属账号 |
| name | VARCHAR(50) | 姓名 |
| birth_date | DATE | 出生日期 |
| gender | TEXT | male/female |
| conditions | TEXT | 慢病标签 JSON 数组 |
| emergency_contacts | TEXT | 紧急联系人 JSON |
| avatar_url | TEXT | 头像 |

**设备表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PRIMARY KEY | UUID |
| dev_id | VARCHAR(20) UNIQUE | BR-XXXX / PX-XXXX / WB-XXXX |
| type | TEXT | 设备类型 (bracelet/pillbox/medical_wristband) |
| tier | TEXT | Starter/Plus/Pro |
| status | TEXT | offline/online/pending_upgrade/fault |
| firmware_version | VARCHAR(20) | 固件版本 |
| last_seen | TIMESTAMP | 最后在线时间 |
| elder_id | UUID → elders.id | 绑定的老人 |
| org_id | UUID | 所属机构(如有) |

**订阅表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT PRIMARY KEY | UUID |
| user_id | TEXT → users.id | 用户 |
| plan | TEXT | free/pro/enterprise |
| status | TEXT | active/expired/cancelled/pending_renewal |
| start_date | DATE | 开始日期 |
| end_date | DATE | 结束日期 |
| downgrade_reason | TEXT | 降级原因 |
| per_device | BOOLEAN | 企业版按床位计费标记 |

**用药规则表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| elder_id | UUID → elders.id | 老人 |
| pillbox_id | UUID → devices.id | 药盒 |
| medication_name | VARCHAR(100) | 药品名称 |
| dosage | VARCHAR(50) | 剂量 |
| schedule | JSONB | 用药时间表 |
| active | BOOLEAN | 是否启用 |

**服药记录表：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| elder_id | UUID → elders.id | 老人 |
| pillbox_id | UUID → devices.id | 药盒 |
| rule_id | UUID → med_rules.id | 规则 |
| scheduled_time | TIMESTAMP | 计划时间 |
| actual_time | TIMESTAMP | 实际时间(null=漏服) |
| status | ENUM('taken', 'missed', 'overdue') | 状态 |
| compartment | INTEGER | 仓位编号 |

**🏥 医用电子腕带新增表（admin-api/medical_wb/）：**

| 表名 | 说明 | 保留策略 |
|------|------|---------|
| `wb_patients` | 住院患者信息 (id, name, gender, age, diagnosis, ward, bed_no, admission_no, admitted_by, admission_time, discharge_time, status) | 出院后归档，30年 |
| `wb_devices` | 腕带设备 (id, dev_id WB-XXXX, ble_mac, firmware_version, status, battery, patient_id, assigned_ward, created_at) | 设备报废后保留3年 |
| `wb_bindings` | 腕带-患者绑定关系 (id, device_id, patient_id, bound_by, bound_at, unbound_reason) | 与患者记录同生命周期 |
| `wb_expenses` | 医疗费用明细 (id, patient_id, item_type, amount, description, billed_by, billed_at) | 财务数据30年 |
| `wb_medications` | 住院用药记录 (id, patient_id, drug_name, dose, route, frequency, ordered_by, started_at, stopped_at) | 病历30年 |
| `wb_test_results` | 检验检查结果 (id, patient_id, test_type, result_value, result_text, ordered_by, collected_at, reported_at) | 病历30年 |
| `wb_daily_entries` | 每日诊疗录入 (id, patient_id, nurse_id, entry_type, content, vital_signs_json, recorded_at) | 病历30年 |
| `wb_verifications` | 近场核验记录 (id, patient_id, device_id, nurse_id, verify_type, verified_at, duration_ms) | 审计日志30年 |
| `wb_alert_tag_config` | 告警标签配置 (id, patient_id, alert_type, enabled, severity, notify_ids_json) | 与患者记录同生命周期 |

**迁移策略：** 所有数据库操作封装在 `store/` 层，业务逻辑不直接依赖 SQL 方言。未来切换到 PostgreSQL 时，只需替换 `sqlite.go` 为 `postgres.go`，Handler 和 Service 层无需改动。表结构使用标准 SQL 语法，避免 SQLite 特有语法。

#### 2.3.5 AI 分析引擎

**端侧推理(手环 GD32E230)：**

| 模型 | 输入 | 输出 | 延迟 |
|------|------|------|------|
| 跌倒检测 | IMU 3轴加速度 + 3轴陀螺仪(1秒窗口) | 跌倒概率(0-1) | <30ms |
| 步数统计 | IMU加速度能量窗口法 | 步数计数 | <10ms |
| 坐姿/静止 | IMU低频采样 | 活动状态 | <5ms |

**云端推理：**

| 分析 | 数据来源 | 输出 | 频率 |
|------|---------|------|------|
| 心率异常检测 | 历史24小时心率 | 异常标记+趋势 | 实时 |
| 用药依从性 | 服药记录 | 周/月依从率 | 每日 |
| 电子围栏动态调整 | 历史定位轨迹 | 推荐围栏范围 | 每周 |
| 健康周报 | 心率+血氧+步数+睡眠 | PDF报告 | 每周 |

**分析规则阈值：**

| 指标 | 正常范围 | 告警阈值 | 级别 |
|------|---------|---------|------|
| 静息心率 | 60-100 bpm | >120 或 <50 | P1 |
| 血氧饱和度 | 95-100% | <90% | P1 |
| 血压收缩压 | 90-140 mmHg | >160 或 <90 | P1 |
| 跌倒置信度 | — | >0.8 | P0 |
| 连续无活动 | — | >2h 无步数 | P2 |
| 夜间心率 | — | >100 持续 30min | P1 |

### 2.4 软件配套运维、升级迭代机制

#### 2.4.1 OTA 升级流程

设备通过 HTTPS GET 下载固件升级包，云端通过 MQTT 下行消息 `ota` 类型推送升级指令，包含版本号、下载地址、SHA-256 哈希校验值和重启标志。手环固件支持 OTA 镜像签名验证和回滚机制。

#### 2.4.2 UI 界面原型工作流

所有应用层界面必须先出高保真 HTML 效果图，确认后才可以写 Flutter/Vue 代码。

**需出效果图的界面清单：**

| # | 子系统 | 界面名称 | 核心元素 |
|---|--------|---------|---------|
| 1 | 家属APP | 首页-实时定位 | 高德地图+老人位置标记+电子围栏+快速状态卡片 |
| 2 | 家属APP | 健康数据看板 | 心率/血氧/步数趋势图+异常告警提示 |
| 3 | 家属APP | SOS告警中心 | 紧急告警列表+处理记录+一键呼叫 |
| 4 | 家属APP | 用药管理 | 今日用药提醒+服药记录+远程配置 |
| 5 | 管理后台 | 仪表盘总览 | 设备在线率/告警统计/用户活跃度/订阅转化 |
| 6 | 管理后台 | 设备管理 | 设备列表+状态筛选+固件版本+OTA升级 |
| 7 | 管理后台 | 用户管理 | 老人/家属/机构用户列表+权限管理 |
| 8 | 管理后台 | 订阅管理 | 订阅状态+续费记录+降级原因 |
| 9 | 品牌官网 | 首页 | Eregen品牌展示+产品介绍+购买入口 |
| 10 | 小程序 | 首页 | 轻量版家属端-定位+用药提醒 |

**执行顺序：** 编写HTML/CSS/JS高保真原型 → 浏览器查看 → 确认或修改 → 基于原型编写实际代码 → 未确认原型的界面绝不写代码。

#### 2.4.3 安全性加固设计

**目标：** 将云平台安全防护从"概念设计"提升到"可运行状态"——每个微服务都有真实的 JWT 验证、输入校验、API 限流，shared/crypto 模块集成到所有 HTTP 端点。

**安全加固清单：**

| 模块 | 现状 | 设计要求 |
|------|------|---------|
| JWT 验证 | auth.go 仅有 token 提取，未验证签名 | api-server 和 admin-api 均实现完整 JWT 验证，提取 shared/crypto JWT 工具 |
| 输入校验 | 所有 handler 直接取 query param/JSON body | shared/validation 包 + request struct binding 标签 + 自定义 validator |
| API 限流 | 无任何速率限制 | shared/ratelimit 基于内存令牌桶，per-user 100/500 req/min，admin 30 req/min，B2B 1000 req/min |
| crypto 模块 | 仅 2 个文件，功能不完整 | AES-256-GCM、Ed25519 设备签名、bcrypt+argon2 密码哈希、TLS 1.3 配置生成 |
| 敏感数据脱敏 | 日志和响应中可能包含 PII | shared/sanitize 包，自动脱敏 email/手机号/token |

**架构分层：**

```
Gin Router
  RateLimit Middleware (shared/ratelimit)
    ↓
  Auth Middleware (JWT 验证)
    ↓
  Sanitize Middleware (响应脱敏)
    ↓
  Handler
    ↓
  Service Layer
    ↓
  Store Layer
```

共享模块：`shared/crypto`（AES-256-GCM, Ed25519, password hash, TLS config）、`shared/validation`（字段校验函数库）、`shared/sanitize`（PII 脱敏过滤器）。

#### 2.4.4 界面深化补全计划

**admin-web 需补全页面：**

| 页面 | 路由 | 当前状态 | 需补全 |
|------|------|---------|--------|
| 仪表盘 | /dashboard | 有布局 + ECharts，全是 mock data | API 对接、WebSocket 实时告警 |
| 设备管理 | /devices | 有布局，mock data | API 对接、OTA 弹窗、配置更新弹窗 |
| 用户管理 | /users | 有布局，mock data | API 对接、角色切换模态框 |
| 订阅管理 | /subscriptions | 有布局，mock data | API 对接、续费记录表格 |
| 系统设置 | /settings | 路由占位，无页面 | **新建**：主题、通知、API Key 管理 |
| OTA 升级 | /ota | 路由占位，无页面 | **新建**：固件列表 + 批量升级 |
| 老人管理 | /elderly | 路由占位，无页面 | **新建**：老人档案列表 + 详情 |

**家属 APP 需补全页面：**

| 页面 | 当前状态 | 需补全 |
|------|---------|--------|
| 首页 | 布局完整 | API 对接、腾讯地图集成、SOS 按钮连接 WebSocket |
| 健康数据 | 布局完整 | API 对接、ECharts 折线图渲染 |
| 告警中心 | 布局完整 | API 对接、告警处理回调 |
| 用药管理 | 布局完整 | API 对接、服药确认按钮 |
| 登录页 | 不存在 | **新建**：手机号+验证码登录 |
| 设备绑定 | 不存在 | **新建**：扫码绑定手环/药盒 |

**小程序需补全：**

| 页面 | 当前状态 | 需补全 |
|------|---------|--------|
| 首页 | JS 逻辑完成 | 腾讯地图插件集成 |
| 用药提醒 | JS 逻辑完成 | 订阅消息模板 ID 配置 |
| 告警中心 | JS 逻辑完成 | 微信消息推送回调 |
| 我的 | JS 逻辑完成 | 添加老人页面、登录页面 |

### 2.4.5 统一启动与运维管理系统

#### 系统概述

Eregen 平台采用统一启动脚本 `scripts/start.sh` 管理所有子服务的生命周期，支持单服务启动、分组启动（cloud/b2b/apps/firmware/all）、端口可配置、依赖检查、健康等待和 PID 管理。该设计实现于 `docs/superpowers/specs/2026-07-27-unified-launch-system-design.md`。

**核心特性：**

1. **端口冲突检测与自动清理**  
   启动前扫描 `.env` 中所有 `PORT_*` 变量值，使用 `lsof` 检查目标端口监听状态。检测到冲突时执行 `kill_by_port(port, force)`：先发送 `SIGTERM`，等待 5 秒后若进程未停止则发送 `SIGKILL`。智能识别规则排除关键系统进程（ssh/sshd/login），仅终止本项目无关的临时进程（如测试遗留的 Python HTTP Server、旧版 Go 服务等）。支持 `--force` 参数强制执行激进清理。

2. **严格健康检查策略**  
   服务后台启动后，每 1 秒轮询一次健康端点 `/api/v1/health`，要求返回码 200 且 JSON 内容为 `{"data":{"status":"ok"}}`。连续 2 次成功即标记为 ready；超时 30 秒则判定启动失败并输出错误告警。各服务健康端点标准化返回格式，确保跨服务一致性。

3. **PID 文件管理**  
   所有后台服务 PID 写入 `$HOME/.eregen/pids/<service>.pid`，包含时间戳便于追踪进程生命周期。`stop` 命令读取 PID 文件优雅终止服务，`clean` 命令清理 stale PID 文件和过期锁文件。

4. **多模式支持**  
   通过环境变量 `MODE` 控制运行策略：
   - `DEV`（开发模式）：宽松冲突处理，健康检查放宽至 60 秒，允许调试附加参数
   - `DEMO`（演示模式）：预置测试数据，开放所有端口，简化鉴权验证
   - `CI/CD`（持续集成）：静默输出，端口自动分配，健康检查超时 15 秒，失败立即退出
   - `PRODUCTION`（生产模式）：严格检查，禁止端口冲突，强制 30 秒健康检查，日志轮转

#### 依赖检查（增强版 B）

| 依赖项 | 最小版本 | 检测方法 |
|--------|---------|----------|
| Go | ≥1.22 | `go version` + semver 比较 |
| Node.js | ≥18 | `node --version` |
| npm | latest | 检查 package-lock.json |
| Flutter | ≥3.24 | `flutter version` |
| ESP-IDF | v5.3 | 路径存在性检查 |
| Arm GNU Toolchain | 13.2+ | `arm-none-eabi-gcc --version` |

CLI 命令：`./scripts/check-deps.sh` 逐项验证，失败时提供安装指引。

#### CLI 用法示例

```bash
# 带端口覆盖启动 api-server
./scripts/start.sh start api-server --port 8080

# 按组启动所有云后端服务
./scripts/start.sh start cloud

# 启动全部服务（含 clean 清理旧 PID）
./scripts.start.sh start --all --clean

# 停止服务
./scripts.start.sh stop api-server

# 查看状态
./scripts.start.sh status --all
```

---


## 第三部分 硬件体系整体建设方案

### 3.1 硬件全域组网拓扑架构

```
[手环传感器]──Cat1蜂窝──→[EMQX MQTT]──→[Go设备接入]──→[NATS总线]
[药盒电机]────WiFi────────→[EMQX MQTT]──→[Go设备接入]──→[NATS总线]
                                                        ↓
                                            ┌───────────┼───────────┐
                                            ↓           ↓           ↓
                                      [SQLite]    [审计日志]   [实时数据流]
                                      用户/设备/订阅 合规记录     AI分析
                                            ↓           ↓           ↓
                                            └───────────┼───────────┘
                                                        ↓
                                            [AI分析引擎]←──[实时数据流]
                                                ↓
                                    跌倒/异常→P0→WebSocket推送
                                    漏服药物→P1→短信兜底
                                    电子围栏→P1→APP推送
                                                ↓
                                          [推送分发器]
                                                ↓
                                    [家属APP] ←→ [管理后台]
```

**设备通信拓扑：**

- 手环：Cat1 蜂窝网络 → EMQX MQTT → Go gateway → NATS 总线
- 药盒：WiFi → EMQX MQTT → Go gateway → NATS 总线
- 药盒 BLE 5.0 ↔ 手环：局域网共享健康数据(心率/步数)，功耗 <1mA

### 3.2 硬件设备清单、参数规格、点位部署方案

#### 3.2.1 三档产品线矩阵

| | 入门版 (Starter) | 中端版 (Plus) | 高端版 (Pro) | 高端版 (Pro+) |
|---|---|---|---|---|
| **手环** | 心率+血氧+SOS+GPS定位 | +电子围栏+跌倒检测+长续航 | +ECG心电+AMOLED屏+金属机身+高精度GPS | +检测模块(血糖/尿酸试纸)+蓝牙血压计配件 |
| **药盒** | 基础分格药盒(无电子功能) | 定时语音提醒+APP联动 | 自动分药+光电检测+库存预警+TTS播报 | — |
| **研发BOM** | ~80-120元 | ~180-230元 | ~280-400元 | ~430-550元 |

#### 3.2.2 手环固件详细设计

**三档功能差异：**

| 功能模块 | Entry (Starter) | Plus (中端) | Pro (高端) |
|---------|:---:|:---:|:---:|
| 心率/血氧 (PPG) | ✅ | ✅ | ✅ |
| SOS 按钮 | ✅ | ✅ | ✅ |
| GPS 定位 | ✅ | ✅ | ✅ (GNSS 高精度) |
| IMU 跌倒检测 | ❌ | ✅ (专用算法) | ✅ |
| 电子围栏 | ❌ | ✅ | ✅ |
| BLE 配网 | ❌ | ✅ | ✅ |
| 电池优化 | ❌ | ✅ (动态采样) | ✅ |
| ECG 心电 | ❌ | ❌ | ✅ |
| AMOLED 屏 | ❌ | ST7789 LCD | ✅ (高刷 AMOLED) |
| 金属机身 | ❌ | 塑料 | ✅ |
| 🆕 电化学检测模块 | ❌ | ❌ | ✅ (Pro+版) |
| 🆕 蓝牙血压计配件 | ❌ | ❌ | ✅ (Pro+版) |

**输入输出接口：**

| 类型 | 数据源/目标 | 协议 |
|------|-----------|------|
| 输入 | PPG 传感器 (汇顶 GT320) | I2C |
| 输入 | IMU 传感器 (ICM-42670-P) | SPI |
| 输入 | GPS 模组 (和芯星通/u-blox) | UART NMEA |
| 输入 | ECG 传感器 (Pro) | I2S |
| 输入 | SOS 按钮 | GPIO |
| 输入 | OLED/AMOLED 显示屏 | SPI/I2C |
| 输出 | Cat1 模组 (广和通 L610-CM) | UART AT 指令 |
| 输出 | BLE 外设 (Plus/Pro 配网) | BLE 5.0 |
| 🆕 输出 | 电化学检测模块 (Pro+版) | GPIO + I2C + ADC |
| 🆕 输出 | BLE 血压计配件 (Pro+版) | BLE 5.3 Peripheral |

**核心模块结构：**

| 文件 | 模块 | 说明 |
|------|------|------|
| `sensors_ppg.c/h` | PPG 驱动 | I2C 读取 GT320 心率/血氧原始数据 |
| `sensors_imu.c/h` | IMU 驱动 | SPI 读取 ICM-42670 加速度/陀螺仪 |
| `health/health_collector.c/h` | 健康采集器 | 融合 PPG+IMU 数据，计算心率/血氧/步数 |
| `gps_nmea.c/h` | GPS 解析 | NMEA GGA/RMC 语句解析为经纬度 |
| `ecg_driver.c/h` (Pro) | ECG 驱动 | 单导联心电信号 ADC 采集 |
| `algorithms/sliding_window.c/h` | 滑动窗口 | 通用滑动窗口统计 |
| `algorithms/fall_detect.c/h` | 跌倒检测 | Entry 基础阈值 / Plus 增强综合判断 |
| `power/power_mgmt.c/h` | 电源管理 | Stop 模式切换、定时器唤醒策略 |
| `cat1_at.c/h` | Cat1 AT 指令 | UART 与广和通 L610 通信 |
| `protocol/message_encode.c/h` | 消息编码 | C 结构体 → JSON 字符串 |
| `ota/ota_download.c/h` | OTA 下载 | HTTP 下载差分升级包 |
| `ota/ota_verify.c/h` | OTA 校验 | SHA-256 签名验证 + CRC32 |
| `sos_button.c/h` | SOS 按钮 | GPIO 中断检测，长按 3s 触发 |
| `plus/geofence_manager.c/h` | 电子围栏管理器 | 多边形围栏，进出触发告警 |

**FreeRTOS 任务划分：**

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| MainTask | 最高 | 512 字 | 系统初始化、任务创建 |
| SensorTask | 高 | 384 字 | PPG/IMU 数据采集，周期 1s |
| GPSTask | 中 | 384 字 | GPS 定位获取，周期 30s |
| HealthTask | 中 | 512 字 | 健康数据处理，周期 5s |
| CommTask | 高 | 768 字 | MQTT 连接维持、消息发送/接收 |
| SOSTask | 最高 | 256 字 | SOS 按钮检测，GPIO 中断 |
| PowerTask | 低 | 256 字 | 电池电量监测，周期 60s |
| DisplayTask | 低 | 384 字 | 屏幕刷新 |
| FallDetectTask (Plus) | 高 | 512 字 | IMU 数据跌倒分析 |
| GeofenceTask (Plus) | 中 | 384 字 | 电子围栏判断 |
| BLETask (Plus/Pro) | 中 | 512 字 | BLE 配网服务 |
| ECGTask (Pro) | 高 | 512 字 | ECG 心电采集 |
| AMOLEDTask (Pro) | 低 | 512 字 | AMOLED 高刷显示 |

**设备状态机：**

```
POWER_OFF (深度休眠)
    ↓ 开机/唤醒
BOOT (Entry/Pro: 验证 OTA 镜像签名，失败则回滚)
    ↓ 启动成功
INIT (传感器初始化、GPS 冷/热启动、Cat1 网络注册)
    ↓
RUN (主循环：采集→编码→发送)
    ├── OTA_UPD (OTA 升级)
    ├── FALL_DETECT (跌倒检测)
    └── SOS_TRIG (SOS 触发)
```

**功耗设计：**

| 模式 | 电流 | 唤醒源 | 适用场景 |
|------|------|--------|---------|
| 运行模式 | ~25mA | — | 数据采集/发送中 |
| 空闲等待 | ~2mA | 定时器/GPIO | 两次采集间隔 |
| Stop 模式 | ~20μA | RTC 闹钟/外部中断 | 夜间/低频率采集 |
| 深度休眠 | ~1μA | SOS 按钮 | 长时间不活动 |

**续航估算：** 350mAh 电池，Entry 版本平均电流 8mA → 约 43 小时；Plus 版本电池优化器可延长至 72 小时。

#### 3.2.3 药盒固件详细设计

**三档功能差异：**

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

**输入输出接口：**

| 类型 | 数据源/目标 | 接口 |
|------|-----------|------|
| 输入 | 光电传感器 (红外对管) | GPIO ADC |
| 输入 | OLED 显示屏 (Smart/Auto) | I2C SSD1306 |
| 输入 | 按键输入 (Basic) | GPIO |
| 输出 | TTS 语音模块 (SYN5300) | UART |
| 输出 | 步进电机 (Auto) | PWM GPIO |
| 输出 | LED 指示灯 | GPIO |
| 输出 | WiFi → EMQX MQTT | TCP/IP |
| 输出 | BLE GATT (Auto) | BLE 5.0 |

**核心模块结构：**

| 文件 | 模块 | 说明 |
|------|------|------|
| `dispensing.c/h` | 分药控制 | 根据用药规则计算目标格子，驱动步进电机 |
| `state_machine.c/h` | 状态机 | 分药流程 IDLE→PREPARE→DISPENSE→CONFIRM→IDLE |
| `motor_control.c/h` | 电机控制 | 28BYJ-48 步进电机驱动，正反转+步数控制 |
| `opto_sensor.c/h` | 光电传感器 | 红外对管读取，判断格子是否有药片 |
| `empty_detector.c/h` | 空盒检测 | 连续 N 次为空判定缺药，触发库存预警 |
| `schedule_engine.c/h` | 调度引擎 | 解析云端 JSON 用药规则，维护每日计划 |
| `med_rule_parser.c/h` | 规则解析 | JSON → C 结构体，支持动态增删规则 |
| `nvs_store.c/h` | NVS 存储 | 用药规则持久化到 ESP32 非易失存储 |
| `tts_playback.c/h` | TTS 播放 | SYN5300 语音模块 UART 控制 |
| `voice_reminder.c/h` | 语音提醒 | 定时触发 TTS 播报 |
| `wifi_mqtt_bridge.c/h` | WiFi+MQTT | ESP-MQTT 库封装 |
| `ap_config_mode.c/h` | AP 配网模式 | 无 WiFi 时开启热点，手机连接配置 |
| `ble_pair.c/h` | BLE 配网 | BLE GATT 服务接收 WiFi 凭证 |

**用药状态机：**

```
IDLE (等待用药时间)
    ↓ 到达用药时间
REMINDER (TTS 语音播报 + LED 闪烁)
    ↓ 用户取药 (光电检测/手动确认)
CONFIRMED (上报 med_status(taken=true))
    ├── 超时未取 (30min) → MISSED (上报 taken=false，云端触发 P1 告警)
    └── 返回 IDLE
```

**ESP-IDF 任务划分：**

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| AppInitTask | 最高 | 1024 字 | WiFi 连接、MQTT 初始化、系统启动 |
| ScheduleTask | 高 | 768 字 | 检查当前时间是否到用药提醒 |
| VoiceTask | 高 | 512 字 | TTS 语音播报 |
| DisplayTask | 中 | 512 字 | OLED 屏幕刷新 |
| CommTask | 中 | 1024 字 | MQTT 消息收发 |
| SensorTask | 低 | 512 字 | 按键检测、LED 状态更新 |
| DispenseTask (Auto) | 高 | 1024 字 | 自动分药流程控制 |
| OptoTask (Auto) | 中 | 512 字 | 光电传感器采样 |
| BLETask (Auto) | 中 | 768 字 | BLE 配网服务 |

**功耗设计：**

| 模式 | 电流 | 说明 |
|------|------|------|
| WiFi 连接+MQTT | ~80mA | 正常工作 |
| Deep Sleep | ~10μA | ESP32-C3 深度休眠，RTC 定时器唤醒 |
| Light Sleep | ~1mA | WiFi 断开但保留状态，快速重连 |

**续航：** 2500mAh 18650 锂电池，Smart 版本平均电流 30mA → 约 83 小时 (~3.5天)，建议每周充电一次。

#### 3.2.4 定位精度规范

| 模式 | 卫星系统 | 室外精度 | 室内精度 | 更新频率 | 功耗 |
|------|---------|---------|---------|---------|------|
| 正常 | GPS+北斗+GLONASS | <5米 | 10-50米(LBS辅助) | 30秒/次 | 中 |
| 告警 | GPS+北斗+GLONASS+WIFI | <3米 | 5-10米 | 1秒/次 | 高 |
| 省电 | LBS基站 | 100-500米 | — | 5分钟/次 | 低 |

**电子围栏：** 圆形围栏半径 50-5000 米可调；多边形围栏最多 20 个顶点；支持基于历史轨迹的动态围栏自动推荐。

### 3.3 硬件安装、调试、运维、维保体系

#### 3.3.1 编译环境

**手环固件：**

```bash
# macOS
brew install arm-none-eabi-gcc openocd

# Linux
sudo apt install arm-none-eabi-gcc openocd dfu-util

cd firmware/bracelet/{entry,plus,pro}
mkdir -p build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=../toolchain/arm-none-eabi.cmake
make -j$(nproc)
```

**药盒固件：**

```bash
git clone --branch v5.3.1 --depth 1 https://github.com/espressif/esp-idf.git ~/esp/esp-idf
cd ~/esp/esp-idf && ./install.sh esp32c3
source ~/esp/esp-idf/export.sh

cd firmware/pillbox/{smart,auto}
idf.py set-target esp32c3
idf.py build
```

#### 3.3.2 烧录工具

| 设备 | 型号 | 用途 | 渠道 | 价格(元) |
|------|------|------|------|---------|
| J-Link BASIC V4 | SEGGER | GD32固件烧录+硬件调试 | 淘宝/代理商 | ~300 |
| USB转TTL(CH340) ×2 | CH340G | 串口打印调试日志 | 淘宝 | ~20 |
| 面包板 + 杜邦线套装 | — | 免焊接搭建临时电路 | 淘宝 | ~30 |
| 万用表 | 优利德 UT61E 或入门款 | 测电压/通断/电流 | 淘宝 | ~50-100 |

#### 3.3.3 烧录步骤

**手环(GD32E230C8T3)：**

```
1. J-Link BASIC V4 连接：SWDIO→PA13, SWCLK→PA14, SWO→PA15, GND→GND, 3.3V→3.3V
2. 安装SEGGER J-Flash ARM
3. 选择芯片 GD32E230C8T3，点击"Download"下载 .elf/.bin
4. CH340串口：TX→RX, RX→TX, GND→GND，波特率115200
```

**药盒(ESP32-C3)：**

```
1. ESP32-C3自带USB接口，直接USB线连接电脑
2. 安装Espressif ESPTool
3. esptool.py --chip esp32c3 write_flash 0x0 firmware.bin
4. 按住BOOT键→按RST→松开RST→松开BOOT进入烧录模式
```

#### 3.3.4 测试策略

**手环固件测试：**

- 单元测试：每个模块配有 `test_*.c`，在 x86 主机上编译运行
- 集成测试：传感器数据环测、MQTT 端到端、OTA 升级全流程

**药盒固件测试：**

- 步进电机分药精度：旋转 N 步，实际到达位置偏差 < 2°
- 光电检测准确率：有药/无药状态识别率 > 99%
- MQTT 端到端：下发用药规则 → 设备解析 → 定时播报 → 服药确认 → 上报云端

### 3.4 硬件分批采购、布设落地时序规划

#### 3.4.1 硬件开发路线

```
M1-M2:  开发板/评估板 → 写驱动 + 核心算法验证
    ↓
M3-M4:  ODM方案商参考设计板 → 全功能联调 + 稳定性验证
    ↓
M5+:    交付厂商开模量产 → 我们只做固件 + 云端 + APP
```

**核心原则：** 我们不设计硬件，只写固件。天线设计、射频调优、电源管理IC选型、Sensor硬件集成全部由ODM方案商负责。

#### 3.4.2 M1-M2 研发工具清单（单套开发工具）

**手环开发模块：**

| 模块 | 搜索关键词 | 价格(元) | 用途 |
|------|-----------|---------|------|
| GD32E230C8T3最小系统板 | GD32E230G-START 开发板 | ~50 | 主控板，跑FreeRTOS |
| PPG传感器评估板 | 汇顶GT320 评估板 | ~80 | 心率+血氧 |
| IMU传感器评估板 | ICM-42670-P EVB | ~60 | 跌倒检测 |
| GPS模组 | 和芯星通 UGN-7345 模组 | ~80 | GPS+北斗定位 |
| Cat1通信模组 | 广和通 L610-CM 开发板 | ~100 | 蜂窝联网(MQTT) |
| OLED屏 | SSD1306 0.96寸 I2C | ~12 | 状态显示 |
| 锂电池+充电板 | 350mAh LiPo + TP4056模块 | ~15 | 供电 |
| **手环小计** | | **~397** | |

**药盒开发模块：**

| 模块 | 搜索关键词 | 价格(元) | 用途 |
|------|-----------|---------|------|
| ESP32-C3开发板 | ESP32-C3-DevKitM-1 | ~25 | 主控+WiFi+BLE |
| 步进电机+驱动 | 28BYJ-48 5V + ULN2003驱动板 | ~8 | 分药旋转机构 |
| OLED屏 | SSD1306 0.96寸 I2C | ~12 | 显示用药状态 |
| TTS语音模块 | SYN5300 语音芯片模块 | ~15 | 定时语音提醒 |
| 光电传感器 | ITRB2800E 红外对管模块 | ~5 | 药物取走检测 |
| 锂电池 | 18650电池+座 | ~15 | 供电(2500mAh) |
| 充电模块 | IP5306 升压充电板 | ~8 | 充电管理 |
| **药盒小计** | | **~88** | |

**烧录调试工具合计：** ~400-450 元

**单套开发工具总计：** ~900 元 + 竞品分析 ~300 元 = ~1,200 元

#### 3.4.3 M1-M2 批量研发物料清单（含数量）

**手环研发物料：**

| 物料 | 型号 | 数量 | 单价(元) | 总价(元) |
|------|------|------|---------|---------|
| GD32E230C8T3开发板 | GD32E230G-START | 3 | 50 | 150 |
| GPS模组(国产) | 和芯星通UGN-7345 | 3 | 80 | 240 |
| GPS模组(进口) | u-blox NEO-M9N | 2 | 350 | 700 |
| Cat1模组 | 广和通L610-CM | 3 | 60 | 180 |
| PPG传感器 | 汇顶GT320 | 5 | 40 | 200 |
| IMU传感器 | ICM-42670-P eval | 3 | 30 | 90 |
| 锂电池 | 350mAh LiPo | 5 | 10 | 50 |
| 3D打印外壳 | PLA材料 | 10套 | 20 | 200 |
| **手环小计** | | | | **~1,610** |

**药盒研发物料：**

| 物料 | 型号 | 数量 | 单价(元) | 总价(元) |
|------|------|------|---------|---------|
| ESP32-C3开发板 | ESP32-C3-DevKitM-1 | 3 | 25 | 75 |
| 步进电机 | 28BYJ-48 | 10 | 5 | 50 |
| 语音模块 | SYN5300 | 10 | 8 | 80 |
| OLED屏 | SSD1306 0.96" | 5 | 12 | 60 |
| 锂电池 | 2500mAh 18650 | 5 | 15 | 75 |
| 光电传感器 | 红外对管 | 10 | 2 | 20 |
| 3D打印外壳 | PETG材料 | 10套 | 25 | 250 |
| **药盒小计** | | | | **~610** |

**批量研发物料总计：** 约 2,220 元

> 说明：~900 元为单套开发工具预算（M1-M2 初期验证），~2,220 元为批量研发物料预算（多套传感器、GPS 模组、3D 打印外壳等，用于多轮迭代验证）。两者口径不同，均有效。

#### 3.4.4 M3-M4 ODM参考设计

| 类型 | 方案商 | 方案特点 | 获取方式 | 成本 |
|------|--------|---------|---------|------|
| 手环ODM | 润科集成 | 完整手环方案(GD32+Cat1+GPS+PPG+IMU) | 联系方案商索取参考设计包 | 0元 |
| 手环ODM | 杰理科技 | 蓝牙+定位一体化方案 | 联系方案商索取参考设计包 | 0元 |
| 药盒ODM | 深圳智能药盒方案商 | 步进电机+旋转药仓完整方案 | 联系方案商索取 | 0元 |
| 药盒ODM | 杰理WiFi+BLE方案 | ESP32替代方案 | 联系方案商索取 | 0元 |

**ODM参考设计包含：** 原理图 + PCB Layout文件、已验证的BOM清单、基础固件代码(BSP+通信协议)、天线设计方案、RF测试报告模板、结构ID参考。

#### 3.4.5 M5+ 量产交付

| 阶段 | 动作 | 责任方 |
|------|------|--------|
| EVT(工程验证) | 手板验证结构+功能 | ODM方案商 |
| DVT(设计验证) | 试产100-500台验证生产工艺 | ODM方案商 |
| PVT(生产验证) | 小批量1000-5000台验证良率 | ODM方案商 |
| MP(量产) | 大批量生产 | 代工厂 |
| **我们** | **只提供固件代码 + 云端服务** | **自研闭源** |

#### 3.4.6 竞品分析策略

| 竞品 | 型号 | 用途 | 价格 |
|------|------|------|------|
| 小寻 | T5s Pro | 学习对手的功能设计和交互逻辑 | ~200 |
| 米家 | 智能药盒 | 学习药盒结构设计和分药机构 | ~100 |

分析维度：功能对比、结构设计、交互体验、BOM拆解。竞品仅用于结构和功能参考，不修改其硬件或软件，分析结果指导设计决策，不用于抄袭。

#### 3.4.7 硬件采购时间线

| 月份 | 手环硬件 | 药盒硬件 | 备注 |
|------|---------|---------|------|
| M1 | 购买J-Link + 开发板 | 购买ESP32-C3开发板 | MVP工具到位 |
| M2 | GPS/PPG/IMU模块验证 | 步进电机/TTS/OLED模块验证 | 逐模块驱动开发 |
| M3 | 联系ODM方案商获取参考设计 | 联系药盒ODM方案商 | 全功能联调准备 |
| M4 | ODM参考设计板到手 | ODM参考设计板到手 | 完整系统集成 |
| M5 | 交付ODM方案商开模 | 交付ODM方案商开模 | 进入结构验证阶段 |

---

## 第四部分 供应链全流程体系建设方案

### 4.1 供应链整体架构与上下游链路设计

**核心策略：** 国产优先、双供应商保障、批量阶梯降本、模块化共用。

手环固件与药盒固件分别对应独立的供应链体系，但通信模块、电源管理、显示模块存在共用可能，可通过减少 SKU 降低库存成本。

### 4.2 供应商准入、物料采购、品质管控机制

#### 4.2.1 关键物料供应商

| 物料 | 国产供应商 | 进口供应商 | 备注 |
|------|-----------|-----------|------|
| Cat1模组 | 广和通L610-CM | 移远EC200U | 国产优先 |
| GPS模组 | 和芯星通UGN-7345 | u-blox NEO-M9N | 国产方案成本低40% |
| PPG传感器 | 汇顶GT3x系列 | Maxim MAX86150 | 汇顶国内服务更好 |
| IMU | 敏芯MEMS | TDK ICM-42670-P | 国产替代逐步推进 |
| MCU | 兆易创新GD32 | ST STM32 | GD32 FreeRTOS官方port |
| TTS芯片 | 矽进SYN5300 | — | 国产垄断 |
| 步进电机 | 东莞28BYJ-48 | — | 通用件，多家供应 |

#### 4.2.2 ODM方案商与代工厂

| 类型 | 代表企业 | 服务内容 |
|------|---------|---------|
| 手环ODM | 润科集成 | 完整手环方案(硬件+基础固件) |
| 手环ODM | 杰理科技 | 蓝牙+定位一体化方案 |
| 药盒ODM | 深圳智能硬件方案商 | 分药机构+固件方案 |
| 代工厂 | 华勤/龙旗 | SMT贴片+整机组装 |

#### 4.2.3 BOM 成本明细

**手环 BOM 成本：**

**入门版(Starter) — 约 80-120 元：**

| 模块 | 物料 | 成本(元) | 说明 |
|------|------|---------|------|
| 主控 | GD32E230C8T3 | 7-12 | Cortex-M4F, 512KB Flash |
| 健康监测 | PPG传感器(汇顶GT320) | 25-35 | 基础心率/血氧 |
| 通信 | 无(Cat1入门版去掉) | 0 | 仅BLE连手机APP |
| 电源 | 280mAh锂电池 | 5-8 | |
| 结构 | 外壳+表带+PCB | 15-20 | 普通塑料 |
| 组装测试 | SMT+基础测试 | 8-12 | |

**中端版(Plus) — 约 188-268 元：**

| 模块 | 物料 | 成本(元) | 说明 |
|------|------|---------|------|
| 通信模块 | Cat1模组(广和通L610-CM) | 28-38 | 蜂窝联网 |
| 通信模块 | BLE SoC(nRF52810) | 15-20 | 局域网BLE |
| 定位模块 | GPS/北斗模组(和芯星通UGN-7345) | 25-40 | 三重卫星定位 |
| 健康监测 | PPG/SpO2传感器(汇顶GT3x) | 40-60 | GT320/GT330系列 |
| 健康监测 | IMU(ICM-42670-P) | 25-35 | 跌倒检测 |
| 主控+存储 | 主控MCU(GD32E230C8T3) | 7-12 | |
| 主控+存储 | Flash(W25Q64) | 2-3 | OTA固件存储 |
| 交互模块 | 显示屏(1.14寸IPS) | 8-15 | 240x240分辨率 |
| 交互模块 | 音频(喇叭+麦克风) | 5-8 | SOS语音 |
| 电源系统 | 锂电池(350mAh LiPo) | 5-10 | |
| 电源系统 | 充电管理(IP5306) | 2-4 | 无线充电 |
| 结构件 | 外壳+表带 | 8-15 | ABS+硅胶 |
| 结构件 | PCB+其他 | 10-16 | 4层板 |
| 组装测试 | SMT+烧录+全检 | 8-12 | |

> 10K 批量下浮 15-20% → ~150-215 元。建议零售价 299-499 元。

**高端版(Pro) — 约 280-400 元：**

在中端版基础上增加：

| 增加项 | 成本增量(元) | 说明 |
|--------|-------------|------|
| AMOLED屏(1.32寸) | +15-25 | 360x360分辨率 |
| ECG心电传感器 | +30-50 | MAX30001或汇顶方案 |
| 金属机身 | +20-30 | 铝合金CNC外壳 |
| 高精度GPS | +10-15 | u-blox NEO-M9N |
| IP68防水 | +10-15 | 游泳级密封 |

> 建议零售价 599-899 元。

**药盒 BOM 成本：**

**基础版(Basic) — 约 30-50 元：**

| 模块 | 物料 | 成本(元) | 说明 |
|------|------|---------|------|
| 结构件 | 分格药盒(PETG) | 8-12 | 无电子功能 |
| 包装 | 纸盒+说明书 | 2-3 | |

> 建议零售价 99-199 元。

**中端版(Smart) — 约 60-90 元：**

在中端版基础上增加：

| 增加项 | 成本增量(元) | 说明 |
|--------|-------------|------|
| OLED显示(SSD1306) | 8-15 | 显示时间/用药状态 |
| TTS语音(SYN5300) | 12-22 | 定时语音提醒 |
| BLE通信 | 10-15 | 连手机APP |
| 电源(含电池) | 10-15 | 可充电锂电池 |
| PCB+控制芯片 | 5-8 | |

> 建议零售价 199-299 元。

**自动版(Auto) — 约 102-171 元：**

| 模块 | 物料 | 成本(元) | 说明 |
|------|------|---------|------|
| 主控+通信 | WiFi+BLESOC(ESP32-C3) | 30-45 | 连云端 |
| 主控+通信 | Flash(W25Q32) | 2-3 | |
| 分药机构 | 步进电机+驱动(28BYJ-48) | 5-9 | |
| 分药机构 | 旋转药仓+导轨(PETG) | 8-15 | |
| 检测传感器 | 光电+到位检测 | 3-5 | |
| 检测传感器 | 空仓检测(ITRB2800E) | 3-5 | |
| 交互模块 | 语音TTS+喇叭(SYN5300) | 12-22 | |
| 交互模块 | OLED显示(SSD1306) | 8-15 | |
| 结构件 | 外壳(食品级阻燃ABS) | 10-20 | |
| 结构件 | PCB | 3-5 | |
| 电源系统 | 锂电池(2500mAh 18650) | 8-15 | |
| 电源系统 | 充电管理(IP5306) | 2-4 | |
| 组装测试 | SMT+烧录+全检+包装 | 10-15 | |

> 10K 批量下浮 15-20% → ~82-137 元。建议零售价 299-499 元。

### 4.3 仓储管理、物流配送、履约交付全流程规范

> 素材缺失：现有源文档中未涉及仓储、物流、备货、履约交付的具体流程设计。量产阶段需另行补充。

### 4.4 供应链风险防控、备货策略、成本管控方案

#### 4.4.1 成本管控策略

| 策略 | 具体做法 | 预期效果 |
|------|---------|---------|
| 国产替代 | 优先国产传感器/MCU/模组 | 比进口方案低30-50% |
| 双供应商 | 关键物料至少2家供应商 | 议价能力+供应安全 |
| 批量阶梯 | M1-M4采购10-50台 → M5+采购10K+ | BOM成本递减15-20% |
| 模块化设计 | 手环/药盒共用通信模块 | 减少SKU，降低库存成本 |
| ODM模式 | 硬件设计外包，专注固件 | 节省硬件工程师团队 |

#### 4.4.2 风险矩阵

| 风险类型 | 描述 | 严重程度 | 缓解措施 |
|----------|------|----------|----------|
| 市场竞争 | 小米/华为/360已占老人定位手表市场 | 高 | 差异化：手环+药盒捆绑套餐+微信生态+订阅服务 |
| 技术风险 | 自动分药机构可靠性(卡药/错药) | 高 | ODM方案先选成熟分药机构，量产前做10,000次循环测试 |
| 合规风险 | 健康数据隐私保护趋严 | 中高 | 数据本地化存储+端到端加密，不碰医疗诊断宣称 |
| 续费率风险 | 订阅服务续费率低于预期(目标30%) | 中 | 增加医生咨询/保险权益等增值服务提升粘性 |
| 供应链风险 | 芯片短缺/涨价 | 中 | 双供应商策略，关键物料储备3个月以上 |

#### 4.4.3 认证与合规

**强制认证（量产阶段需要）：**

| 认证 | 依据 | 周期 | 费用 | 不做后果 |
|------|------|------|------|---------|
| SRRC无线电核准 | 《无线电管理条例》 | 1.5-3月 | 2-5万 | 没收+罚款3-10倍 |
| CTA进网许可 | 《电信条例》第57条 | 1-2月 | 1-3万 | 没收+罚款5-10万 |
| CCC | GB 4943.1-2022 | 2-3月 | 2-4万 | 禁止销售 |

> 研发阶段暂不需要认证，功能验证为主。量产前完成认证送测。

**数据合规：**

- 健康数据 = 敏感个人信息 (PIPL第28-32条)
- 需单独同意 + 专项隐私政策
- 数据境内存储 (服务器部署在阿里云/腾讯云国内节点)
- TLS1.2+传输加密 + AES-256静态加密

**广告合规：**

- MVP阶段不宣称医疗诊断功能，定位为"生活方式健康设备"
- 禁用词汇："治愈""疗效""治疗""医疗级""医用"
- 可使用："运动监测""日常健康数据记录""生活方式参考"

#### 4.4.4 专利保护策略

**可申请专利的模块：**

| 层级 | 可申请专利 |
|------|-----------|
| 固件层 | 手环跌倒检测算法、药盒分药机构控制逻辑、低功耗传输方案 |
| 云平台层 | 多设备数据融合分析算法、三级告警推送机制、电子围栏动态调整 |
| APP层 | 适老化交互设计、家属多人协同监护模式 |
| 慢性病层 | 手环式可拆卸电化学检测模块结构、试纸阻抗识别算法、慢病数据关联分析方法、试纸耗材订阅系统架构 |

**版权声明：**

- 所有代码仓库加声明：`© 2026 Eregen (颐贞). All rights reserved.`
- 开源组件清单维护：每个子项目维护 `THIRD-PARTY-LICENSES` 文件
- 核心业务逻辑、通信协议、AI算法全部自研闭源

---

## 第五部分 项目建设过程管理资料汇总

### 5.1 安全性加固建设要点

**目标：** 将云平台安全防护从"概念设计"提升到"可运行状态"——每个微服务都有真实的 JWT 验证、输入校验、API 限流，shared/crypto 模块集成到所有 HTTP 端点。

**安全加固清单：**

| 模块 | 现状 | 设计要求 |
|------|------|---------|
| JWT 验证 | auth.go 仅有 token 提取，未验证签名 | api-server 和 admin-api 均实现完整 JWT 验证，提取 shared/crypto JWT 工具 |
| 输入校验 | 所有 handler 直接取 query param/JSON body | shared/validation 包 + request struct binding 标签 + 自定义 validator |
| API 限流 | 无任何速率限制 | shared/ratelimit 基于内存令牌桶，per-user 100/500 req/min，admin 30 req/min，B2B 1000 req/min |
| crypto 模块 | 仅 2 个文件，功能不完整 | AES-256-GCM、Ed25519 设备签名、bcrypt+argon2 密码哈希、TLS 1.3 配置生成 |
| 敏感数据脱敏 | 日志和响应中可能包含 PII | shared/sanitize 包，自动脱敏 email/手机号/token |

**架构分层：**

```
Gin Router
  RateLimit Middleware (shared/ratelimit)
    ↓
  Auth Middleware (JWT 验证)
    ↓
  Sanitize Middleware (响应脱敏)
    ↓
  Handler
    ↓
  Service Layer
    ↓
  Store Layer
```

### 5.2 UI 界面深化建设要点

**目标：** 将所有前端界面的 mock data 替换为真实 API 对接，补齐缺失页面，统一视觉风格。

**admin-web 需补全页面：**

| 页面 | 路由 | 当前状态 | 需补全 |
|------|------|---------|--------|
| 仪表盘 | /dashboard | 有布局 + ECharts，全是 mock data | API 对接、WebSocket 实时告警 |
| 设备管理 | /devices | 有布局，mock data | API 对接、OTA 弹窗、配置更新弹窗 |
| 用户管理 | /users | 有布局，mock data | API 对接、角色切换模态框 |
| 订阅管理 | /subscriptions | 有布局，mock data | API 对接、续费记录表格 |
| 系统设置 | /settings | 路由占位，无页面 | **新建**：主题、通知、API Key 管理 |
| OTA 升级 | /ota | 路由占位，无页面 | **新建**：固件列表 + 批量升级 |
| 老人管理 | /elderly | 路由占位，无页面 | **新建**：老人档案列表 + 详情 |

**家属 APP 需补全页面：**

| 页面 | 当前状态 | 需补全 |
|------|---------|--------|
| 首页 | 布局完整 | API 对接、腾讯地图集成、SOS 按钮连接 WebSocket |
| 健康数据 | 布局完整 | API 对接、ECharts 折线图渲染 |
| 告警中心 | 布局完整 | API 对接、告警处理回调 |
| 用药管理 | 布局完整 | API 对接、服药确认按钮 |
| 登录页 | 不存在 | **新建**：手机号+验证码登录 |
| 设备绑定 | 不存在 | **新建**：扫码绑定手环/药盒 |

**小程序需补全：**

| 页面 | 当前状态 | 需补全 |
|------|---------|--------|
| 首页 | JS 逻辑完成 | 腾讯地图插件集成 |
| 用药提醒 | JS 逻辑完成 | 订阅消息模板 ID 配置 |
| 告警中心 | JS 逻辑完成 | 微信消息推送回调 |
| 我的 | JS 逻辑完成 | 添加老人页面、登录页面 |

### 5.3 告警优先级体系（统一后口径）

| 级别 | 事件 | 推送方式 | 响应要求 |
|------|------|---------|---------|
| **P0** | SOS/跌倒/电子围栏越界 | APP推送 + SMS短信 + 电话语音 | <1分钟 |
| **P1** | 漏服药物/心率异常/低电量(<10%) | APP推送 + SMS短信 | <5分钟 |
| **P2** | 用药提醒/常规健康数据/设备离线 | APP推送 | 当日查看 |

**推送服务分级实现：**

| 级别 | 类型 | 推送渠道 | 响应要求 |
|------|------|---------|---------|
| P0 | SOS、跌倒 | FCM + 微信 + 短信 + 电话语音 | 立即响应 |
| P1 | 漏服药物、心率异常、低电量 | FCM + 微信 + 短信 | <5分钟 |
| P2 | 设备离线、低电量、用药提醒 | FCM | 下次打开 APP 可见 |

### 5.4 用户角色与权限矩阵

| 角色 | 可查看数据 | 可操作功能 | 适用子系统 |
|------|-----------|-----------|-----------|
| 老人 | 自身健康数据 | 佩戴设备、接听语音提醒 | 手环/药盒固件 |
| 家属 | 关联老人全部数据 | SOS响应、用药配置、电子围栏设置、告警处理 | 家属APP、管理后台 |
| 运营人员 | 全局统计数据 | 设备管理、用户管理、订阅管理、OTA升级 | 管理后台 |
| 医院医生 | 授权老人的健康报告 | 开具用药规则、查看历史趋势 | B2B医院API |
| 社区护理员 | 辖区老人状态 | 创建活动、登记体检、查看参与记录 | B2B社区平台 |
| 保险公司 | 脱敏健康数据 | 理赔审核、保单管理 | B2B保险对接 |
| **护士** | 管辖病房患者信息（不可跨科室） | 入院登记、腕带绑定、近场核验、每日诊疗录入 | **管理后台医护工作站** |

### 5.5 设备认证流程

```
1. 工厂烧录: 每台设备出厂时写入唯一 dev_id + 预共享密钥(PSK)
2. 首次上电: 设备通过MQTT发送 REGISTER 消息 + 证书指纹
3. 云端验证: Go gateway 验证证书指纹 → 在SQLite创建设备记录
4. 颁发Token: 云端生成JWT Token → MQTT下行返回设备
5. 绑定老人: 家属APP扫码或输入dev_id → 绑定elder_id
```

### 5.6 通信协议消息格式

**上行消息示例：**

```json
// 心跳包 (每5分钟)
{"type":"heartbeat","dev_id":"BR-XXXX","bat":85,"fw_ver":"1.0.0"}

// 定位数据 (正常30s/次，告警时1s/次)
{"type":"location","dev_id":"BR-XXXX","lat":31.2304,"lon":121.4737,"acc":5,"ts":1720000000}

// 健康数据 (每5分钟)
{"type":"health","dev_id":"BR-XXXX","hr":72,"spo2":98,"step":3456,"cal":120,"ts":1720000000}

// SOS告警 (立即)
{"type":"sos","dev_id":"BR-XXXX","lat":31.2304,"lon":121.4737,"ts":1720000000}

// 跌倒检测 (立即)
{"type":"fall","dev_id":"BR-XXXX","conf":0.95,"lat":31.2304,"lon":121.4737,"ts":1720000000}

// 药盒状态 (定时+事件驱动)
{"type":"med_status","dev_id":"PX-XXXX","compartment":3,"taken":true,"ts":1720000000}

// 药盒库存预警
{"type":"med_inventory","dev_id":"PX-XXXX","compartment":3,"level":"low","ts":1720000000}
```

**下行消息示例：**

```json
// 用药规则 (云端下发)
{"type":"med_rule","dev_id":"PX-XXXX","rules":[{"time":"08:00","dose":1,"type":"capsule","name":"氨氯地平"}]}

// 配置更新
{"type":"config","dev_id":"BR-XXXX","settings":{"interval":30,"volume":80,"gps_mode":"balanced"}}

// 语音播报
{"type":"tts","dev_id":"PX-XXXX","text":"爷爷，该吃降压药了"}

// OTA升级
{"type":"ota","dev_id":"BR-XXXX","url":"https://ota.eregen.cn/firmware/v1.1.0.bin","hash":"sha256:abc123...","size":524288,"reboot":true}

// 电子围栏配置
{"type":"geofence","dev_id":"BR-XXXX","fences":[{"id":"home","type":"circle","lat":31.2304,"lon":121.4737,"radius":200}]}
```

### 5.7 软件基础设施成本

| 项目 | 方案 | 月成本(元) | 年成本(元) |
|------|------|-----------|-----------|
| 云服务器 | 阿里云轻量应用服务器2核4G | ~50 | ~600 |
| 域名 | 阿里云.com域名 | ~7 | ~84 |
| SSL证书 | Let's Encrypt免费 | 0 | 0 |
| MQTT Broker | EMQX Docker自建 | 含在服务器中 | — |
| 数据库 | SQLite（零部署） | — | — |
| 推送服务 | FCM免费 + 阿里云SMS按量 | ~10 | ~120 |
| **软件合计** | | **~67元/月** | **~804元/年** |

### 5.8 慢性病专项成本估算

| 项目 | 成本 | 说明 |
|------|------|------|
| 试纸模块硬件BOM | ~150元/套 | 电化学传感器+信号IC+微流控+温度传感器 |
| 试纸生产成本 | ~2元/条 | ODM代工量产后 |
| 二类医疗器械注册 | ~30-50万元 | 与试纸一起注册 |
| 试纸供应链（初期） | ~20万元 | 试纸生产模具 + 首批库存 |
| 血压配件开发 | ~50万元 | 外置蓝牙血压计，独立注册 |
| APP开发（慢病模块） | ~3人月 | Flutter开发 + UI设计 |
| 后端开发（慢病API） | ~2人月 | Go后端 + AI分析引擎 |
| **合计** | **~120-150万元** | 含硬件+软件+注册 |

---

## 第六部分 全文问题汇总附录

### 6.1 内容冲突条目及位置（含整改方案）

以下冲突已在正文中按最优方案统一处理，此处列明冲突来源和解决依据：

| 冲突主题 | 文档A | 文档B | 差异说明 | 整改方案 |
|---------|-------|-------|---------|---------|
| **告警 P1/P2 范围** | `01-system-architecture.md` §2.3: P1=漏服/心率异常/低电量(<5min), P2=用药提醒/常规健康数据 | `03-cloud-platform.md` §4.4: P1=漏服/围栏越界(30min), P2=设备离线/低电量 | 低电量归属不同（P1 vs P2）；电子围栏越界归属不同（P0 vs P1）；响应时限不同（5min vs 30min） | 采用安全优先原则：围栏越界归 P0（<1min），低电量归 P1（<5min），用药提醒和设备离线归 P2。详见正文 5.3 节 |
| **手环研发物料总价** | `00-global-architecture.md` §8: 手环~1,610元 + 药盒~610元 = ~2,220元 | `10-hardware-procurement.md` §2.4: 手环~397元 + 药盒~88元 + 工具~400-450元 = ~900元 | 前者含多套物料（3-10件），后者为单套开发板+工具预算 | 两者均有效，分别标注"单套开发工具"和"批量研发物料"。详见正文 3.4.2 和 3.4.3 节 |
| **InfluxDB 保留策略** | `01-system-architecture.md`: location_data 90天; health_data 永久 | `03-cloud-platform.md`: health_data 730天; location_data 365天; device_status 90天 | 健康数据和定位数据的保留周期不一致 | 采用更精细的分级保留策略：health_data 730天、location_data 365天、device_status 90天。详见正文 2.3.4 节 |
| **Redis key 命名/TTL** | `01-system-architecture.md`: `device:online:{dev_id}` TTL心跳超时; `device:last_loc:{dev_id}` TTL 1小时; `alert:pending:{dev_id}` 手动清除 | `03-cloud-platform.md`: `session:{user_id}` 7天; `device:{dev_id}:online` 300s; `location:latest:{elderly_id}` 永久; `alert:unread:{user_id}` 永久 | key 命名规范和 TTL 策略不同 | 采用 `03-cloud-platform.md` 命名规范，location latest TTL 改为 24h（避免永久缓存陈旧数据）。详见正文 2.3.4 节 |
| **订阅套餐命名** | `00-product-overview.md`: free/pro/enterprise（英文）和 基础版/高级版/尊享版（中文） | `03-cloud-platform.md`: free/pro/enterprise（英文） | 中英文命名并存 | 正文以中文显示名为主，内部 enum 保持 free/pro/enterprise。详见正文 1.5 节 |
| **设备类型枚举** | `01-system-architecture.md`: ENUM('bracelet_starter', ..., 'pillbox_auto') | `03-cloud-platform.md`: `tier VARCHAR(10)` 与 `device_type VARCHAR(20)` 分离 | schema 设计不同 | 采用单 ENUM 字段方案，更简洁且匹配六档产品矩阵。详见正文 2.3.4 节 |
| **手环定位地图** | `04-family-app.md`: 高德地图 | `07-miniprogram.md`: 腾讯地图 | 平台差异 | 保持现状：家属APP用高德，小程序用腾讯地图。详见正文 2.3.2 节 |

### 6.2 原始素材信息缺失点位

以下主题在源文档中未覆盖或覆盖不足，后续建设需补充：

| 缺失主题 | 说明 |
|---------|------|
| 仓储物流与履约交付流程 | 量产后的仓储管理、订单处理、物流配送、退换货流程未涉及 |
| 备货计划 | 关键物料的安全库存、补货触发点、批量采购节奏未设计 |
| 供应链风险量化指标 | 有风险描述和缓解措施，但缺少具体的风险量化评估模型 |
| 认证送测详细计划 | 仅列出认证类型和费用，缺少具体送测时间表和责任分配 |
| 微信小程序完整页面设计 | 仅有首页、用药提醒、告警中心、我的四个页面概述，缺少详细功能规格 |
| 品牌官网内容规划 | 仅有框架描述，缺少页面结构、内容栏目、SEO策略细节 |
| B2B 对接详细接口协议 | 仅有接口列表，缺少数据格式、字段定义、错误码、限流策略 |
| 患者/老人端独立应用 | 文档中提到"老人端"但未设计独立 APP 或轻量小程序 |
| 售后与客服体系 | 设备故障处理、客服工单、远程诊断流程未涉及 |
| 数据备份与灾备方案 | 仅提到数据境内存储，缺少备份策略、RTO/RPO 指标 |
| 灰度发布与回滚策略 | OTA 有回滚机制设计，但云端服务的灰度发布策略未涉及 |
| 监控告警阈值 | Prometheus/Grafana 已选型，但具体监控指标和告警阈值未定义 |

### 6.3 冗余删减与内容增补改动明细

**冗余删减：**

| 内容主题 | 出现位置 | 处理方式 |
|---------|---------|---------|
| 三档产品线矩阵 | docs/specs/00-global-architecture.md §2; docs/superpowers/specs/00-product-overview.md §2; CLAUDE.md | 正文只保留一份完整表格 |
| 8个子系统边界定义 | docs/specs/00-global-architecture.md §3; docs/superpowers/specs/00-product-overview.md §3 | 正文只保留一份完整表格 |
| 实施批次划分 | docs/specs/00-global-architecture.md §3.2; docs/superpowers/specs/00-product-overview.md §3 | 正文只保留一份完整表格 |
| 技术选型表（固件层/云平台/应用层） | docs/specs/00-global-architecture.md §4; docs/superpowers/specs/01-system-architecture.md §4.1; CLAUDE.md | 正文只保留一份完整表格 |
| 设备-云端通信协议（上行/下行 JSON） | docs/specs/00-global-architecture.md §5; docs/superpowers/specs/01-system-architecture.md §3 | 正文只保留一份完整 JSON 示例 |
| 数据流闭环 | docs/specs/00-global-architecture.md §6; docs/superpowers/specs/01-system-architecture.md §2 | 正文合并为统一数据流描述 |
| 认证与合规 | docs/specs/00-global-architecture.md §12; docs/superpowers/specs/00-product-overview.md §8 | 正文合并到第四部分统一呈现 |
| 专利保护策略 | docs/specs/00-global-architecture.md §13; docs/superpowers/specs/00-product-overview.md §9 | 正文合并到第四部分统一呈现 |
| 风险评估 | docs/superpowers/specs/00-product-overview.md §10 | 正文合并到第四部分风险矩阵 |
| 四阶段战略演进 | docs/superpowers/specs/00-product-overview.md §11 | 正文合并到第一部分分层建设目标 |
| 硬件采购策略（M1-M5路线） | docs/superpowers/specs/00-product-overview.md §5; docs/superpowers/specs/10-hardware-procurement.md 全篇 | 正文合并到第三部分硬件采购时序 |
| BOM成本（手环/药盒） | docs/superpowers/specs/00-product-overview.md §6; docs/superpowers/specs/11-supply-chain-bom.md 全篇 | 正文合并到第四部分BOM明细 |
| 部署架构图 | docs/superpowers/specs/01-system-architecture.md §7; docs/specs/00-global-architecture.md 网络架构图 | 正文只保留一份完整部署图 |
| 云平台微服务结构 | docs/specs/03-cloud-platform.md §1.2; docs/superpowers/specs/01-system-architecture.md §4.2 | 正文合并为统一微服务结构 |
| 数据库设计（PostgreSQL/InfluxDB/Redis） | docs/specs/03-cloud-platform.md §6; docs/superpowers/specs/01-system-architecture.md §4.3 | 正文合并并统一口径 |
| 安全架构 | docs/specs/00-global-architecture.md §1.7.3; docs/superpowers/specs/01-system-architecture.md §5 | 正文合并为统一安全架构 |

**内容增补：**

| 增补主题 | 来源 | 说明 |
|---------|------|------|
| 项目背景与社会趋势 | 综合源文档 | 第一部分新增老龄化趋势和立项依据论述 |
| 中长期发展前景规划 | 综合源文档 | 第一部分新增商业模式、收入测算、四阶段战略演进 |
| 微信小程序详细设计 | docs/specs/07-miniprogram.md | 新增小程序功能模块、微信登录流程、订阅消息推送、腾讯地图集成 |
| B2B对接详细设计 | docs/specs/08-b2b-integration.md | 新增 hospital-api/community-platform/insurance-integration 三个子系统的 API 和数据结构 |
| 安全性加固设计 | docs/superpowers/specs/2026-07-17-security-hardening-design.md | 新增 JWT 验证、输入校验、API 限流、crypto 模块扩展、PII 脱敏设计 |
| UI 界面深化补全计划 | docs/superpowers/specs/2026-07-17-ui-enhancement-design.md | 新增 admin-web/family-app/小程序缺失页面补全计划 |
| 硬件批量研发物料清单 | docs/superpowers/specs/10-hardware-procurement.md + 11-supply-chain-bom.md | 新增含数量的批量研发物料采购清单（~2,220元） |
| 冲突整改说明 | 全文交叉核对 | 第六部分新增冲突明细表和整改方案 |
| 素材缺失点位清单 | 全文交叉核对 | 第六部分新增信息缺失点位清单 |

---

## 附录：版本变更日志与文档更新清单

### v2.1 变更日志（2026-07-21）

| 变更项 | 类型 | 说明 |
|--------|------|------|
| 数据库选型 | 【修改】 | PostgreSQL/InfluxDB/Redis → SQLite（全项目 MVP 阶段统一） |
| 新增子系统 | 【新增】 | ⑨ 医用电子腕带（ESP32-S3 + ESP-IDF）+ ⑩ 护士核验终端（Flutter移动应用 + 手机双形态） |
| 新增 Go 模块 | 【新增】 | admin-api/medical_wb/ 内嵌于 admin-api |
| 用户角色 | 【新增】 | 护士角色（管理后台医护工作站） |
| 通信协议 | 【新增】 | MQTT topic `eregen/medical/wb/#`、NATS subject `eregen.medical.wb.#`、REST prefix `/api/v1/medical/` |
| NFC | 【新增】 | 护士核验终端通过 NFC 近场读取腕带患者信息 |
| 数据库表 | 【新增】 | 9 张医用腕带 SQLite 表 |
| 安全合规 | 【新增】 | PIPL 合规要求（AES-256-GCM 存储加密、BLE Secure Connection + AES-128-CBC 传输加密、审计日志、30年保留） |
| UI 原型 | 【新增】 | 医护工作站页面、住院治疗页面（家属 APP/小程序） |
| 部署架构 | 【修改】 | SQLite 替代 PG/InfluxDB/Redis，新增 medical_wb 模块 |
| 基础设施成本 | 【修改】 | 移除 PostgreSQL/InfluxDB Docker 成本 |
| 设备认证流程 | 【修改】 | PostgreSQL → SQLite |


| v2.1.1 | 【同步】 | 2026-07-29 | 更新 B2B 状态及护士终端描述，统一 SQLite 口径 |
| v2.2 | 【新增】 | 2026-08-08 | 慢性病专项升级：手环Pro+版（电化学检测模块+血压配件）、家属APP慢病管理模块（7新页+3改造页）、后端慢病数据模型+API+AI引擎、试纸供应链+商业模式 |
| v2.3 | 【新增】 | 2026-08-08 | 第四批实施时间线（6个Phase）、成本估算（~120-150万元）、产品线矩阵Pro+列、专利方向扩展 |
### 文档更新清单

| 文档 | 状态 | 更新日期 | 主要变更 |
|------|------|---------|---------|
| `project_total_construction_scheme_v2.md` | ✅ 已更新 | 2026-07-21 | 数据库、子系统矩阵、数据流、角色、成本、版本日志<br>`10-subsystem-verification.md` added as reference |
| `10-subsystem-verification.md` | 【新增】 | 2026-07-27 | 子系统顺序验证方案（admin-web/family-app/nurse-terminal/website/miniprogram/ui-prototypes/cloud backend） |
| `00-global-architecture.md` | ✅ 已更新 | 2026-07-21 | 新增⑨⑩子系统、护士角色、SQLite、医用腕带数据流 |
| `03-cloud-platform.md` | ✅ 已更新 | 2026-07-21 | SQLite Store 层、admin-api/medical_wb 模块、API 端点 |
| `05-admin-web.md` | ✅ 已更新 | 2026-07-21 | 医护工作站页面、护士核验终端、Web NFC API（已废弃） |
| `04-family-app.md` | ✅ 已更新 | 2026-07-21 | 新增"住院治疗"页面 |
| `07-miniprogram.md` | ✅ 已更新 | 2026-07-21 | 新增"住院治疗"页面 |
| `08-b2b-integration.md` | ✅ 已更新 | 2026-07-21 | HIS 预留接口（MVP 未启用） |
| `README.md` | ✅ 已更新 | 2026-07-21 | 同步 SQLite、⑨⑩子系统、护士角色、目录结构 |
| `01-bracelet-firmware.md` | ✅ 已更新 | 2026-07-21 | 标注住院场景由⑨独立腕带处理 |
| `09-medical-wristband.md` | 【新增】 | 2026-07-21 | 医用腕带完整设计文档 |
| `2026-08-08-chronic-care-upgrade-design.md` | 【新增】 | 2026-08-08 | 慢性病专项升级完整设计方案（硬件+APP+后端+实施） |
| `project_total_construction_scheme_v2.md` | ✅ 已更新 | 2026-08-08 | v2.2：新增Pro+版手环矩阵、实施批次第四批、慢性病升级描述 |

---

> © 2026 Eregen (颐贞). All rights reserved.
>
> 本文档为修订完整版，所有有效内容均来自 `docs/specs/` 与 `docs/superpowers/specs/` 源文件。缺失内容已标注，未自行补充。
