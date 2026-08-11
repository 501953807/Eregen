# Eregen 项目架构文档

## 系统概览

Eregen 是一个完整的老年健康生态系统，包含硬件固件、云平台、多端应用和 B2B 服务。

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 硬件固件 | C/FreeRTOS, ESP-IDF | GD32E230, ESP32-C3 |
| 云端后端 | Go + Gin | 微服务架构 |
| 数据库 | SQLite (MVP) → PostgreSQL (正式) | 双存储抽象层 |
| 消息队列 | NATS JetStream | 事件总线 |
| MQTT 网关 | EMQX | 设备接入 |
| 家属 APP | Flutter (Dart) | 跨平台移动应用 |
| 管理后台 | Vue 3 + TypeScript + Element Plus | Web 控制台 |
| 微信小程序 | 原生 WXML/WXSS | 轻量版家属端 |
| 品牌官网 | Hugo + Tailwind CSS | 静态站点 |
| 护士终端 | Flutter | 医院/PDA 设备 |

## 子项目架构

### 云端服务 (cloud/)
```
cloud/
├── admin-api/      # 管理后台 API (主服务)
├── api-server/     # 核心 API 服务
├── gateway/        # MQTT 网关
├── push-service/   # 推送服务
└── data-pipeline/  # 数据分析
```

### 应用端 (apps/)
```
apps/
├── admin-web/      # 管理后台 (Vue 3)
├── family-app/     # 家属 APP (Flutter)
├── miniprogram/    # 微信小程序
├── nurse_terminal/ # 护士终端 (Flutter)
├── website/        # 品牌官网 (Hugo)
└── ui-prototypes/  # UI 原型
```

### B2B 服务 (b2b/)
```
b2b/
├── hospital-api/       # 医院对接 API
├── community-platform/ # 社区平台 API
└── insurance-integration/ # 保险对接 API
```

### 共享库 (shared/)
```
shared/
├── crypto/     # 加密模块
├── protocol/   # 设备协议
├── ratelimit/  # 限流器
├── audit/      # 审计日志
└── validation/ # 数据验证
```

### 硬件固件 (firmware/)
```
firmware/
├── bracelet/         # 手环固件
├── pillbox/          # 药盒固件
├── medical-wristband/ # 医用腕带固件
└── common/           # 公共库
```

## 数据流

```
设备 → MQTT → Gateway → NATS JetStream → [API Server / Push Service / Data Pipeline]
    │                                    │
    ├─ bracelet/BR-XXXX/up ──────────────┤ 主题: eregen.event.>
    ├─ medical/wb/MW-XXXX/up ────────────┤ 主题: eregen.medical.wb.>
    └─ community/wb/CW-XXXX/up ──────────┘ 主题: eregen.community.wb.>
                                              ↓
                                      [PostgreSQL / SQLite]
                                              ↓
                                  [Admin API] ←→ [Admin Web]
                                  [Hospital API] ←→ [外部医院HIS]
                                  [Nurse Terminal] ←→ [Admin API]
```

### NATS JetStream 主题说明

| 主题 | 设备类型 | 消息类型 | 消费者 |
|------|---------|---------|--------|
| `eregen.event.>` | 手环、药盒 | heartbeat, location, health, sos, fall, med_status | API Server |
| `eregen.medical.wb.>` | 医用腕带 | patient_register, verification_scan, device_status, alert_tag | API Server |
| `eregen.community.wb.>` | 社区腕带 | community_signin, community_welfare_update, community_dispense | API Server |

## API 端点总览

### Admin API (cloud/admin-api)

#### 统一人本位路由
- `/api/v1/persons` - 统一人员管理（支持business_chain过滤）
- `/api/v1/persons/{person_id}` - 人员详情（含所有业务链信息）
- `/api/v1/persons/{person_id}/health` - 健康档案（跨链聚合）
- `/api/v1/persons/{person_id}/medications` - 用药规则
- `/api/v1/persons/{person_id}/alerts` - 告警记录
- `/api/v1/persons/{person_id}/devices` - 设备列表
- `/api/v1/persons/{person_id}/reports` - 健康报告

#### 自营链路由
- `/api/v1/self/elderly` - 自营老人管理
- `/api/v1/self/elderly/{id}/health-report` - 健康报告
- `/api/v1/self/elderly/{id}/guidance` - 健康指导

#### 住院链路由
- `/api/v1/hospital/patients` - 住院患者管理
- `/api/v1/hospital/admissions` - 入院登记
- `/api/v1/hospital/admissions/{id}/discharge` - 出院结算
- `/api/v1/hospital/patients/{id}/daily` - 每日治疗记录
- `/api/v1/hospital/patients/{id}/verify` - 护士核验

#### 社区链路由
- `/api/v1/community/elders` - 社区老人管理
- `/api/v1/community/elders/{id}/signin` - 签到
- `/api/v1/community/elders/{id}/welfare` - 福利发放

#### 监管链路由
- `/api/v1/regulatory/compliance` - 合规检测
- `/api/v1/regulatory/audit` - 审计查询
- `/api/v1/regulatory/reports` - 监管报表

#### 其他路由
- `/api/v1/auth/login` - 登录
- `/api/v1/users` - 用户管理
- `/api/v1/devices` - 设备管理
- `/api/v1/subscriptions` - 订阅管理
- `/api/v1/ota` - OTA 升级

### API Server (cloud/api-server)
- 设备接入
- 健康数据
- 定位服务
- 告警推送

### B2B API
- 医院患者管理
- 社区老人管理
- 保险结算

## 业务链架构

Eregen 项目包含三条核心业务链，每条链有独立的状态机、权限体系和业务逻辑：

### 业务链概述

| 业务链 | 英文名称 | 核心用户 | 主要功能 | 设备类型 |
|--------|---------|---------|---------|---------|
| 自营链 | self | 老人/家属 | 健康监测、用药提醒、告警推送 | 手环、药盒 |
| 住院链 | hospital | 医院/护士/医生 | 入院管理、巡检查房、用药执行 | 医用腕带 |
| 社区链 | community | 社区医院/老人 | 福利发放、健康筛查、药品管理 | 社区腕带 |

### 统一身份模型

所有业务链使用统一的身份基表 `persons`，以身份证号作为唯一标识：

```
persons（统一身份基表）
├── id, id_card, name, gender, birth_date, phone, ...
└── person_profiles（业务链扩展表，JSON存储各链业务字段）
    ├── business_chain='self'    → 自营链扩展字段
    ├── business_chain='hospital' → 住院链扩展字段
    └── business_chain='community' → 社区链扩展字段
```

详细设计见 `docs/specs/12-business-architecture.md`

### 五角色权限体系

| 角色 | 自营链 | 住院链 | 社区链 | 监管链 |
|------|--------|--------|--------|--------|
| super_admin | 全权限 | 全权限 | 全权限 | 全权限 |
| operator | 查看+编辑 | 查看+编辑 | 查看+编辑 | 只读 |
| hospital_doc | 无权限 | 查看+编辑 | 只读 | 无权限 |
| nurse | 无权限 | 查看+执行 | 无权限 | 无权限 |
| community_staff | 只读 | 只读 | 查看+编辑 | 无权限 |
| regulator | 只读 | 只读 | 只读 | 全权限 |

### 业务链状态机

```
自营链: pending → active → suspended → cancelled
住院链: pending → admitted → in_treatment → discharged → archived
社区链: pending → certified → active → suspended → deactivated
```

### 跨链关联

- 同一身份证号可关联多条链（`linked_person_id`）
- 住院链出院后可转为社区链或自营链
- 健康数据可跨链共享（需权限控制）

---

## 数据库 Schema

主库文件：`eregen.db` (SQLite)
迁移脚本：`init-db.sql`

核心表（已按业务链深化）：
- `persons` - 统一身份基表
- `person_profiles` - 业务链扩展表
- `medication_rules` - 用药规则（支持三条链）
- `medication_executions` - 用药执行记录
- `health_records` - 统一健康记录表
- `alert_rules` - 告警规则（按业务链配置）
- `alerts` - 统一告警记录表
- `user_role_bindings` - 用户-业务链-角色绑定表
- `devices` - 统一设备表
- `device_bindings` - 设备绑定关系表
- `health_guidance_rules` - 健康指导规则
- `health_guidance_deliveries` - 健康指导推送记录
- `health_report_templates` - 健康报告模板
- `health_reports` - 健康报告记录
- `compliance_rules` - 合规检测规则
- `compliance_checks` - 合规检测结果

详细设计见 `docs/specs/12-business-architecture.md`

## 部署架构

```
┌─────────────────────────────────────────────────────┐
│                   外网访问层                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ Admin Web │  │Family App│  │ Website  │          │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘          │
│        │             │             │                │
└────────┼─────────────┼─────────────┼────────────────┘
         │             │             │
┌────────┴─────────────┴─────────────┴────────────────┐
│                   服务层                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │Admin API │  │API Server│  │B2B APIs  │          │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘          │
│        │             │             │                │
│  ┌─────┴─────────────┴─────────────┴─────┐          │
│  │              NATS JetStream            │          │
│  └──────────────────┬────────────────────┘          │
│                     │                               │
└─────────────────────┼───────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────┐
│                   数据层                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ SQLite   │  │ Redis    │  │ EMQX     │          │
│  │(MVP)     │  │(缓存)    │  │(MQTT)    │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

## 开发规范

1. **提交消息**：遵循 Conventional Commits 规范
2. **文档更新**：代码变更同步更新相关文档
3. **分支管理**：feature/xxx 开发分支，main 稳定分支
4. **代码审查**：重大变更需经过 review

## 已知问题与解决方案

### 问题：数据库文件被误提交
**解决方案**：已添加到 .gitignore，删除 git 历史中的追踪

### 问题：二进制构建产物残留
**解决方案**：清理 .gitignore，移除追踪

### 问题：计划文档过度膨胀
**解决方案**：归档过时文档，保留执行中的计划

### 问题：重复开发目录
**解决方案**：归档到 SP1-auth/ 并加入 gitignore

---
最后更新：2026-08-11
