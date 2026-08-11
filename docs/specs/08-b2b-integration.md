# ⑧ B2B 对接 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：⑧ B2B 对接 (Go + Gin + SQLite)  
> 包含模块：hospital-api / community-platform / insurance-integration

---

## 1. 概述

### 1.1 职责

B2B 对接层为医院、社区养老机构和保险公司提供标准化的数据接口，实现健康数据共享、用药规则协同、活动管理和保险理赔自动化。三个子服务共享相同的架构模式但各有独立的数据库表和 API 路由。

### 1.2 子系统划分

| 子系统 | 端口 | 职责 | 外部依赖 |
|--------|------|------|---------|
| hospital-api | 8082 | 医院 HIS 系统对接，健康数据导出/导入 | 医院信息系统的 HL7/FHIR（MVP 暂不启用） |
| community-platform | 8083 | 社区养老机构平台，活动管理/体检登记 | 社区护理系统 |
| insurance-integration | 8084 | 保险公司系统对接，理赔/保单管理 | 保险核心业务系统 |

---

## 2. 共同架构

### 2.1 分层模式

每个 B2B 子服务采用统一的四层架构：

```
HTTP Request (Gin Router)
    ↓
Handler (请求解析 + 响应格式化)
    ↓
Model (数据结构定义 + 验证)
    ↓
Store (SQLite CRUD)
```

### 2.2 项目结构 (以 hospital-api 为例)

```
b2b/hospital-api/
├── cmd/main.go                    # 入口
├── internal/
│   ├── handler/
│   │   ├── health_data.go         # 健康数据接口
│   │   ├── institution.go         # 机构管理接口
│   │   └── link.go                # 对接链接管理
│   ├── middleware/
│   │   └── auth.go                # API Key 认证中间件
│   ├── model/
│   │   ├── model.go               # 数据模型定义
│   │   └── model_test.go          # 模型测试
│   ├── router/
│   │   ├── router.go              # Gin 路由注册
│   │   └── handlers.go            # Handler 实例化
│   └── store/
│       └── sqlite.go              # SQLite 操作
├── go.mod
└── go.sum
```

### 2.3 数据库表 (共享模式)

```sql
-- 所有 B2B 服务使用同一 SQLite 数据库，通过表前缀隔离

-- hospital-api: b2b_hospital_ 前缀
CREATE TABLE b2b_hospital_health_data_exports (...);
CREATE TABLE b2b_hospital_institutions (...);
CREATE TABLE b2b_hospital_medical_rules (...);

-- community-platform: b2b_community_ 前缀
CREATE TABLE b2b_community_events (...);
CREATE TABLE b2b_community_event_registrations (...);
CREATE TABLE b2b_community_care_plans (...);
CREATE TABLE b2b_community_health_checks (...);

-- insurance-integration: b2b_insurance_ 前缀
CREATE TABLE b2b_insurance_providers (...);
CREATE TABLE b2b_insurance_policies (...);
CREATE TABLE b2b_insurance_claims (...);
CREATE TABLE b2b_insurance_evidence_files (...);
CREATE TABLE b2b_insurance_health_exports (...);
CREATE TABLE b2b_insurance_premium_reminders (...);
```

---

## 3. hospital-api — 医院对接

### 3.1 核心概念

| 实体 | 说明 |
|------|------|
| Institution | 对接的医院/诊所机构 |
| HealthDataExport | 健康数据导出申请 |
| MedicalRule | 医院开具的用药规则 |
| VitalSign | 体征数据 (血压/心率/体温等) |
| HealthDataRequest | 健康数据查询请求 |

### 3.2 API 端点

```
POST   /api/v2/b2b/hospitals              # 注册医院机构
GET    /api/v2/b2b/hospitals/:id           # 机构详情
PUT    /api/v2/b2b/hospitals/:id           # 更新机构

POST   /api/v2/b2b/hospitals/:id/export   # 申请健康数据导出
GET    /api/v2/b2b/hospitals/:id/export/:export_id  # 导出状态
GET    /api/v2/b2b/hospitals/:id/export/:export_id/download  # 下载报告

POST   /api/v2/b2b/hospitals/:id/rules    # 下发用药规则
GET    /api/v2/b2b/hospitals/:id/rules    # 查询用药规则

GET    /api/v2/b2b/hospitals/:id/health-data?elderly_id=&start=&end=  # 查询健康数据
```

### 3.3 健康数据导出格式

```json
{
  "export_id": "exp_abc123",
  "elderly_id": "elderly_456",
  "period": {"start": "2024-01-01", "end": "2024-06-30"},
  "vital_signs": [
    {"date": "2024-03-15", "hr": 72, "bp_systolic": 128, "bp_diastolic": 82, "spo2": 98},
    {"date": "2024-03-16", "hr": 68, "bp_systolic": 125, "bp_diastolic": 80, "spo2": 97}
  ],
  "medication_rules": [
    {"time": "08:00", "dose": 1, "type": "capsule", "active": true}
  ],
  "generated_at": "2024-07-15T10:00:00Z",
  "file_url": "https://.../export.pdf"
}
```

---

## 4. community-platform — 社区平台

### 4.1 核心概念

| 实体 | 说明 |
|------|------|
| CommunityEvent | 社区活动 (体检/讲座/义诊) |
| EventRegistration | 老人报名登记 |
| CarePlan | 护理计划 (含多个护理任务) |
| HealthCheckRecord | 体检记录 |
| ServiceType | 服务类型枚举 (health_check/lecture/free_clinic) |

### 4.2 API 端点

```
POST   /api/v2/b2b/events                 # 创建活动
GET    /api/v2/b2b/events                 # 活动列表 (支持 service_type 筛选)
GET    /api/v2/b2b/events/:id             # 活动详情
DELETE /api/v2/b2b/events/:id             # 删除活动

POST   /api/v2/b2b/events/:id/register    # 老人报名
GET    /api/v2/b2b/events/:id/registrations  # 报名列表

POST   /api/v2/b2b/care-plans             # 创建护理计划
GET    /api/v2/b2b/care-plans/:elderly_id  # 查询护理计划

POST   /api/v2/b2b/health-checks          # 登记体检记录
GET    /api/v2/b2b/health-checks/:elderly_id  # 查询体检历史
```

### 4.3 护理计划模型

```json
{
  "id": "cp_001",
  "elderly_id": "elderly_456",
  "institution_id": "inst_001",
  "caregiver_id": "cg_007",
  "title": "高血压日常护理",
  "tasks": [
    {
      "title": "晨起测量血压",
      "schedule_time": "07:00",
      "frequency": "daily",
      "completed": false
    },
    {
      "title": "午间服药",
      "schedule_time": "12:30",
      "frequency": "weekdays",
      "completed": true
    }
  ],
  "status": "active",
  "created_at": "2024-07-01T00:00:00Z"
}
```

---

## 5. insurance-integration — 保险对接

### 5.1 核心概念

| 实体 | 说明 |
|------|------|
| InsuranceProvider | 保险公司/ Provider |
| Policy | 老人保单信息 |
| InsuranceClaim | 理赔申请 |
| EvidenceFile | 理赔证据文件 |
| HealthDataExport | 健康数据导出 (用于理赔) |
| PremiumReminder | 保费提醒 |
| ClaimType | 理赔类型 (accident/illness/emergency/health_check/chronic_disease) |
| ClaimStatus | 理赔状态 (submitted/under_review/approved/rejected/paid) |

### 5.2 API 端点

```
POST   /api/v2/b2b/providers              # 注册保险公司
GET    /api/v2/b2b/providers/:id           # 保险公司详情
PUT    /api/v2/b2b/providers/:id           # 更新保险公司

POST   /api/v2/b2b/policies               # 创建保单
GET    /api/v2/b2b/policies?elderly_id=   # 查询保单
PUT    /api/v2/b2b/policies/:id           # 更新保单

POST   /api/v2/b2b/claims                 # 提交理赔申请
GET    /api/v2/b2b/claims/:id             # 理赔详情
PUT    /api/v2/b2b/claims/:id/status      # 更新理赔状态
GET    /api/v2/b2b/claims?elderly_id=     # 理赔列表

POST   /api/v2/b2b/claims/:id/evidence    # 上传证据文件
GET    /api/v2/b2b/claims/:id/evidence    # 查看证据文件

POST   /api/v2/b2b/exports                # 申请健康数据导出
GET    /api/v2/b2b/exports/:id            # 导出状态

POST   /api/v2/b2b/reminders/schedule     # 设置保费提醒
GET    /api/v2/b2b/reminders/upcoming     # 即将到期的提醒
```

### 5.3 理赔申请流程

```
保险公司提交理赔申请 (POST /claims)
    ↓
状态: submitted
    ↓
系统自动生成健康数据导出请求
    ↓
导出完成 → 状态: under_review
    ↓
人工审核 → 状态: approved/rejected
    ↓
赔付完成 → 状态: paid
    ↓
发送保费提醒 (PremiumReminder)
```

---

## 6. 认证与安全

### 6.1 B2B 认证方式

| 层级 | 机制 | 说明 |
|------|------|------|
| 传输层 | TLS 1.3 | 所有 B2B API 强制 HTTPS |
| 应用层 | API Key | 请求头 `X-API-Key: <key>` |
| IP 限制 | 白名单 | 仅允许注册 IP 访问 |
| 审计 | 日志记录 | 所有请求记录操作日志 |

### 6.2 API Key 管理

```sql
CREATE TABLE b2b_api_keys (
    id TEXT PRIMARY KEY,
    institution_id TEXT NOT NULL,
    key_hash TEXT NOT NULL,  -- SHA-256 hash
    name TEXT NOT NULL,
    ip_whitelist TEXT DEFAULT '[]',
    active INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 7. 编译与运行

### 7.1 统一启动

```bash
# 医院 API (端口 8082)
cd b2b/hospital-api && go run ./cmd

# 社区平台 (端口 8083)
cd ../community-platform && go run ./cmd

# 保险对接 (端口 8084)
cd ../insurance-integration && go run ./cmd
```

### 7.2 数据库初始化

SQLite 零部署，应用启动时自动创建表结构。详见 `docs/specs/03-cloud-platform.md` §6.2。

### 7.3 单元测试

```bash
# 模型测试 (验证字段名/类型/常量正确性)
cd b2b/hospital-api && go test ./internal/model/...
cd b2b/community-platform && go test ./internal/model/...
cd b2b/insurance-integration && go test ./internal/model/...
```

---

## 8. 与主云平台的关系

| 方向 | 机制 | 说明 |
|------|------|------|
| B2B → 云平台 | REST API (api-server) | 读取老人档案、健康数据、设备状态 |
| 云平台 → B2B | NATS 事件 | 新数据到达时通知 B2B 系统 |
| B2B → 云平台 | REST API (api-server) | 写入用药规则、护理计划 |

---

## 9. HIS 预留接口（v2.1 新增，MVP 暂不启用）

> 【新增】医用腕带上线后，医院 HIS 系统需通过预留接口与 Eregen 云端对接。以下接口仅定义协议，MVP 阶段不实现。

### 9.1 入院登记同步

```
POST /api/v2/b2b/hospitals/:id/medical/wb/admit
Request: {
  "patient_name": "张三",
  "gender": "male",
  "birth_date": "1950-01-15",
  "ward": "心内科3",
  "bed_no": "12",
  "diagnosis": "高血压",
  "attending_doctor": "张医生"
}
Response: {
  "patient_id": "p_abc123",
  "wristband_id": "WB-0001",
  "status": "admitted"
}
```

### 9.2 出院注销同步

```
POST /api/v2/b2b/hospitals/:id/medical/wb/discharge
Request: {"patient_id": "p_abc123", "discharge_reason": "治愈出院"}
Response: {"patient_id": "p_abc123", "status": "discharged", "wristband_released": true}
```

### 9.3 医嘱同步

```
POST /api/v2/b2b/hospitals/:id/medical/wb/orders
Request: {
  "patient_id": "p_abc123",
  "orders": [
    {"time": "08:00", "type": "medication", "content": "降压药 1粒"},
    {"time": "10:00", "type": "test", "content": "血常规检查"}
  ]
}
```

### 9.4 数据查询（只读）

```
GET /api/v2/b2b/hospitals/:id/medical/wb/patients?ward=&status=
GET /api/v2/b2b/hospitals/:id/medical/wb/patients/:id/daily-entries?date=
GET /api/v2/b2b/hospitals/:id/medical/wb/patients/:id/verifications?from=&to=
```

---

## 10. 业务链费用采集与合规检测

### 10.1 医疗费用账单字段扩展

为支持合规检测（R_C06/R_C07/R_C08），`medical_expenses` 表需增加以下字段：

```sql
-- 现有字段保留
id, patient_id, expense_date, item_name, category, amount, quantity, unit_price, notes

-- 新增合规检测字段
diagnosis_code TEXT,                  -- ICD编码，用于同病种费用对比
dept_id TEXT,                         -- 科室ID，用于同科室费用对比
billing_source TEXT CHECK (billing_source IN ('his','manual','pharmacy','lab','radiology')),
insurance_type TEXT CHECK (insurance_type IN ('employee','resident','self')),
approved_amount REAL,                 -- 医保报销金额
patient_amount REAL                   -- 患者自付金额
```

### 10.2 合规检测数据源

| 规则 | 数据来源 | 检测逻辑 |
|------|---------|---------|
| R_C06 费用异常 | medical_expenses | 住院总费用 vs 同科室同病种平均费用（含药品费/检查费/治疗费分项） |
| R_C07 分项异常 | medical_expenses | 药品费/检查费/治疗费分项 vs 标准费用 |
| R_C08 重复收费 | medical_expenses | 同一天同一 item_name + category 计数 > 1 |
| R_C10 出院集中 | medical_expenses | 出院前24h费用 / 总费用 比例 > 50% |

### 10.3 B2B 角色与业务链映射

| B2B 服务 | 业务链 | 权限范围 |
|---------|--------|---------|
| hospital-api | 住院链 | 只能访问本机构（institution_id）的患者数据 |
| community-platform | 社区链 | 只能访问本机构（institution_id）的老人数据 |
| insurance-integration | 监管链 | 可跨机构查询（只读），用于合规审计 |

### 10.4 跨链数据可见性

- **operator**（运营平台）：仅通过 admin-api 管理自营链，不能通过 B2B 接口访问住院链和社区链
- **regulator**（医保监管）：通过 insurance-integration 可查询所有医院的费用数据，用于跨机构合规检测
- **hospital_doc/nurse/community_staff**：仅能访问所属机构的数据

---

© 2026 Eregen (颐贞). All rights reserved.
