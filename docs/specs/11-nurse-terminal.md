# ⑪ 护士终端 — 详细设计文档

> 生成日期：2026-08-02  
> 对应子系统：⑪ 护士终端 (Flutter — Web/Android PDA)  
> 语言：Dart | 框架：Flutter 3.24+ | 平台：Web + Android

---

## 1. 概述

### 1.1 职责

护士终端是部署在医院/护理中心平板或 PDA 设备上的专用移动应用，为护理人员提供患者监护、医用腕带近场核验、用药扫码执行、查房记录和出院管理的完整工作流。与医护工作站（管理后台内嵌）互补——终端侧重近场操作，工作站侧重后台管理。

**关键区别：**
- 护士终端：PDA 手持设备，支持 NFC 近场读取医用腕带
- 医护工作站：浏览器 Web 应用，无 NFC 能力，仅做登记和查询

### 1.2 输入输出

| 类型 | 来源/目标 | 说明 |
|------|-----------|------|
| **输入** | 护士扫码操作 | 医用腕带 NFC 近场读取 |
| **输入** | 手动录入 | 生命体征、查房观察、出院信息 |
| **输出** | REST API 调用 | 数据提交 → admin-api `/api/v1/admin/medical/` |
| **输出** | 本地持久化 | SharedPreferences 存储 auth token |

---

## 2. 功能模块

### 2.1 核心页面

| 页面 | 文件 | 功能 |
|------|------|------|
| 登录 | `login_screen.dart` | 机构选择 + 账号密码登录 |
| 患者列表 | `home_screen.dart` | 在院患者列表、搜索、刷新 |
| 患者详情 | `patient_detail_screen.dart` | 基本信息、腕带状态、操作入口 |
| 腕带核验 | `verification_screen.dart` | NFC 读取 → 身份核验 → 记录提交 |
| 用药执行 | `medication_screen.dart` | 今日用药清单 + 扫码确认 |
| 查房记录 | `ward_round_screen.dart` | 生命体征录入 + 观察记录 |
| 出院办理 | `discharge_screen.dart` | 出院类型选择 + 转院/死亡备注 |

### 2.2 服务层

| 服务 | 文件 | 职责 |
|------|------|------|
| API 客户端 | `api_client.dart` | HTTP 封装、Token 管理 |
| 患者服务 | `patient_service.dart` | 患者 CRUD、出院操作 |
| 核验服务 | `verification_service.dart` | 核验记录创建/查询 |
| 查房服务 | `ward_round_service.dart` | 查房记录创建/查询 |
| NFC 服务 | `nfc_wristband_service.dart` | NFC 读写、NDEF payload 解析 |

### 2.3 数据模型

| 模型 | 文件 | 说明 |
|------|------|------|
| 医疗数据 | `medical_models.dart` | Patient、VerificationRecord、WardRoundEntry |
| NFC Tag | `nfc_tags.dart` | NDEF Record Type 定义 |

---

## 3. API 接口映射

| 页面 | API 端点 | 方法 |
|------|---------|------|
| 登录 | `/api/v1/auth/login` | POST |
| 患者列表 | `/api/v1/admin/medical/patients?status=admitted` | GET |
| 患者详情 | `/api/v1/admin/medical/patients/:id` | GET |
| 腕带核验 | `/api/v1/admin/medical/verifications` | POST |
| 核验记录 | `/api/v1/admin/medical/verifications` | GET |
| 用药执行 | `/api/v1/admin/medical/medications/:id/verify` | POST |
| 查房记录 | `/api/v1/admin/medical/patients/:id/ward-round` | POST |
| 查房历史 | `/api/v1/admin/medical/patients/:id/ward-round` | GET |
| 出院办理 | `/api/v1/admin/medical/patients/:id/discharge` | POST |

---

## 4. 技术架构

### 4.1 项目结构

```
apps/nurse_terminal/
├── lib/
│   ├── main.dart                    # 应用入口 + 路由
│   ├── common/
│   │   └── theme.dart               # 统一主题 (琥珀色系)
│   └── src/
│       ├── models/
│       │   ├── medical_models.dart  # 数据模型
│       │   └── nfc_tags.dart        # NFC NDEF Record Type 定义
│       ├── screens/
│       │   ├── login_screen.dart
│       │   ├── home_screen.dart
│       │   ├── patient_detail_screen.dart
│       │   ├── verification_screen.dart
│       │   ├── medication_screen.dart
│       │   ├── ward_round_screen.dart
│       │   └── discharge_screen.dart
│       └── services/
│           ├── api_client.dart
│           ├── patient_service.dart
│           ├── verification_service.dart
│           ├── ward_round_service.dart
│           └── nfc_wristband_service.dart
├── pubspec.yaml
└── README.md
```

### 4.2 技术栈

| 库 | 版本 | 用途 |
|----|------|------|
| Flutter | 3.24+ | 跨平台 UI 框架 |
| nfc_manager | 3.x | NFC 读写与 NDEF 解析 |
| http | 1.x | HTTP 请求 |
| shared_preferences | 2.x | Token 持久化 |
| uuid | 4.x | 生成本地请求 ID |

### 4.3 工作流

```
登录 → 患者列表 → 选择患者 → 详情
                        ↓
        ┌───────────────┼───────────────┐
        ↓               ↓               ↓
    腕带核验        用药执行        查房记录
    (NFC读取)      (列表+确认)      (生命体征录入)
        ↓               ↓               ↓
    提交核验        提交用药        提交查房
    记录             记录           记录
```

---

## 5. NFC 协议

医用腕带 NFC A 协议（106 kbps，近场 4cm）：

| 记录 | NDEF Type | 方向 | 说明 |
|------|-----------|------|------|
| Device Info | `application/vnd.eregen.device-info` | 读 | 设备 ID、固件版本、电量 |
| Patient Identity | `application/vnd.eregen.patient` | 读（认证） | 患者 ID、入院号、风险标签 |
| Verification | `application/vnd.eregen.verification` | 写 | 核验请求（challenge-response） |
| Status | `application/vnd.eregen.status` | 读 | 腕带状态（在线/绑定/电量） |

---

## 6. 部署

### 6.1 目标平台

- **Web**: Chrome/Edge 浏览器访问 (测试用)
- **Android**: PDA 手持设备 (生产用)

### 6.2 构建

```bash
cd apps/nurse_terminal

# Web 构建
flutter build web --release

# Android APK
flutter build apk --release

# Android App Bundle (Google Play)
flutter build appbundle --release
```

---

## 7. 与 B2B 后端的关系

护士终端通过 admin-api 的医疗工作流 API 与后端通信：
- 认证：使用 nurse 角色的 JWT token
- 数据隔离：仅能访问本机构管辖的患者
- 审计：所有操作记录到 admin-api 审计日志

---

## 8. 业务链上下文

### 8.1 仅服务于住院链

护士终端 Flutter App 仅服务于住院链（hospital）业务：
- 身份：nurse 角色，通过 admin-api 认证获取 JWT token
- 数据隔离：通过 `institution_id` 绑定所属医院，仅能访问本机构患者
- 跨链不可见：无法查看自营链（self）或社区链（community）的任何数据

### 8.2 与身份模型的关联

护士在执行巡检查房时：
1. 扫描腕带 → 获取 `person_id`
2. 查询 `persons` 表获取身份信息（姓名、年龄等基础信息）
3. 查询 `person_profiles` 表（business_chain='hospital'）获取住院业务信息
4. 执行核验 → 写入 `medical_verifications` 表

护士终端不感知 `elderly_profiles`（自营）或 `community_elders`（社区）的存在。
