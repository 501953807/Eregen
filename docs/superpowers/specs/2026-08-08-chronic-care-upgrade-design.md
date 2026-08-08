# Eregen 慢病管理升级文档

> 编制日期：2026-08-08
> 版本：v1.0

---

## 一、概述

本次升级为 Eregen 平台新增慢性病史管理功能，支持血糖、尿酸、血压检测数据的采集、分析和展示。

### 核心功能

1. **手环 Pro+ 版** — 新增电化学检测模块，支持血糖/尿酸试纸检测
2. **外置血压计配件** — BLE 5.3 连接，数据同步到手环和 APP
3. **家属 APP 慢病管理模块** — 7个新页面 + 3个改造页面
4. **后端 API** — 17条 REST 端点，完整 CRUD 支持
5. **AI 分析引擎** — 血糖/尿酸/血压异常检测 + 综合建议

---

## 二、架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    硬件层 (Pro+ 版)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ 试纸检测模块 │  │ BLE血压计   │  │ 现有传感器  │         │
│  │ (血糖/尿酸) │  │ (外置配件)  │  │ (心率/血氧) │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
└─────────┼─────────────────┼─────────────────┼──────────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │ MQTT / BLE
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    后端服务层                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ API Server  │  │Data Pipeline│  │Push Service │         │
│  │ (17条路由)  │  │ (分析引擎)  │  │ (提醒推送)  │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
│         └─────────────────┼─────────────────┘                │
│                           ↓                                  │
│                    ┌─────────────┐                           │
│                    │   SQLite    │                           │
│                    │ (7张新表)   │                           │
│                    └─────────────┘                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    应用层                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  家属 APP   │  │  UI 原型    │  │  管理后台   │         │
│  │ (7新页面)   │  │ (3 HTML)    │  │ (待扩展)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、快速开始

### 3.1 启动后端服务

```bash
# 启动 API Server
cd cloud/api-server && ./start.sh

# 启动数据管道
cd cloud/data-pipeline && ./start.sh
```

### 3.2 测试 API 端点

```bash
# 运行测试脚本
./scripts/test-chronic-api.sh

# 或手动测试
curl -X POST http://localhost:8080/api/v1/chronic/:elderly_id/glucose \
  -H "Content-Type: application/json" \
  -d '{"value": 6.8, "test_mode": "fasting"}'
```

### 3.3 查看 UI 原型

```bash
# 打开慢病管理主页原型
open apps/ui-prototypes/family-app/chronic-home.html

# 打开血糖详情页原型
open apps/ui-prototypes/family-app/chronic-glucose.html

# 打开健康报告页原型
open apps/ui-prototypes/family-app/chronic-report.html
```

---

## 四、API 端点参考

### 血糖管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/chronic/:elderly_id/glucose` | 录入血糖值 |
| GET | `/api/v1/chronic/:elderly_id/glucose` | 获取血糖列表 |
| GET | `/api/v1/chronic/:elderly_id/glucose/trend` | 获取趋势数据 |
| POST | `/api/v1/chronic/:elderly_id/test-strip/read` | 试纸检测上报 |

### 尿酸管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/chronic/:elderly_id/uric-acid` | 录入尿酸值 |
| GET | `/api/v1/chronic/:elderly_id/uric-acid` | 获取尿酸列表 |

### 血压管理
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/chronic/:elderly_id/blood-pressure` | 录入血压值 |
| GET | `/api/v1/chronic/:elderly_id/blood-pressure` | 获取血压列表 |
| POST | `/api/v1/chronic/:elderly_id/bp-device/sync` | 血压计数据同步 |

### 饮食/运动
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/chronic/:elderly_id/diet` | 记录饮食 |
| GET | `/api/v1/chronic/:elderly_id/diet` | 获取饮食记录 |
| POST | `/api/v1/chronic/:elderly_id/exercise` | 记录运动 |
| GET | `/api/v1/chronic/:elderly_id/exercise` | 获取运动记录 |

### 任务/报告
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/chronic/:elderly_id/daily-tasks` | 获取每日任务 |
| PUT | `/api/v1/chronic/:elderly_id/daily-tasks/:task_id` | 标记任务完成 |
| POST | `/api/v1/chronic/:elderly_id/report/generate` | 生成报告 |
| GET | `/api/v1/chronic/:elderly_id/report/:type` | 获取报告 |

---

## 五、数据库表结构

| 表名 | 说明 |
|------|------|
| `chronic_glucose_records` | 血糖检测记录 |
| `chronic_uric_acid_records` | 尿酸检测记录 |
| `chronic_bp_records` | 血压记录 |
| `chronic_diet_records` | 饮食记录 |
| `chronic_exercise_records` | 运动记录 |
| `chronic_daily_tasks` | 每日任务 |
| `chronic_health_reports` | 周期报告 |

---

## 六、正常值参考

| 指标 | 正常范围 | 偏低阈值 | 偏高阈值 |
|------|---------|---------|---------|
| 空腹血糖 | 3.9-6.1 mmol/L | < 3.9 | > 7.0 |
| 餐后2h血糖 | < 7.8 mmol/L | — | > 10.0 |
| 尿酸 | 143-420 μmol/L | < 143 | > 420 |
| 收缩压 | 90-140 mmHg | < 90 | > 140 |
| 舒张压 | 60-90 mmHg | < 60 | > 90 |

---

## 七、关键文件索引

### 后端服务
```
cloud/api-server/internal/
├── model/chronic.go              # 数据模型
├── store/chronic.go              # 存储层
├── service/chronic_*.go          # 业务逻辑
└── handler/chronic_*.go          # API处理器
```

### 分析引擎
```
cloud/data-pipeline/internal/analyzer/
├── chronic_glucose.go            # 血糖分析
├── chronic_uric.go               # 尿酸分析
├── chronic_bp.go                 # 血压分析
└── chronic_recommendations.go    # 综合建议
```

### 固件模块
```
firmware/bracelet/pro_plus/
├── sensors/electrochemical.c/h   # 电化学检测驱动
├── bt_peripheral/bp_device.c/h   # BLE血压计连接
├── app/chronic_manager.c/h       # 慢病任务调度
└── protocol/                     # 协议定义
```

### APP页面
```
apps/family-app/lib/screens/chronic/
├── chronic_home_page.dart        # 慢病主页
├── blood_sugar_page.dart         # 血糖详情
├── uric_acid_page.dart           # 尿酸详情
├── blood_pressure_page.dart      # 血压详情
├── diet_page.dart                # 饮食记录
├── exercise_page.dart            # 运动追踪
└── report_page.dart              # 健康报告
```

### UI原型
```
apps/ui-prototypes/family-app/
├── chronic-home.html             # 慢病主页原型
├── chronic-glucose.html          # 血糖详情原型
└── chronic-report.html           # 健康报告原型
```

---

## 八、测试指南

### 编译验证
```bash
# API Server
cd cloud/api-server && go build ./...

# Data Pipeline
cd cloud/data-pipeline && go build ./...

# Flutter APP
cd apps/family-app && flutter analyze lib/screens/chronic/
```

### 单元测试
```bash
cd cloud/data-pipeline && go test ./internal/analyzer/ -v
```

### API测试
```bash
./scripts/test-chronic-api.sh
```

---

## 九、未来 roadmap

### P0 — 近期（1-2个月）
- [ ] 试纸 ODM 代工对接
- [ ] 二类医疗器械注册申请
- [ ] 血压配件硬件开发

### P1 — 中期（3-6个月）
- [ ] 固件 BLE 血压计集成
- [ ] 推送服务慢病提醒扩展
- [ ] 端到端硬件联调

### P2 — 长期（6-12个月）
- [ ] 用户测试（5-10人）
- [ ] AI 模型优化
- [ ] 多语言国际化

---

## 十、相关文档

- 设计方案：`docs/superpowers/specs/2026-08-08-chronic-care-upgrade-design.md`
- 实施计划：`docs/superpowers/plans/2026-08-08-chronic-care-upgrade.md`
- 主方案文档：`docs/specs/project_total_construction_scheme_v2.md` (v2.3)
- 实施清单：`.scratch/chronic-care-implementation-checklist.md`

---

**© 2026 Eregen (颐贞). All rights reserved.**
