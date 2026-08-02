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
设备 → MQTT → Gateway → NATS → [API Server / Push Service / Data Pipeline]
                                    ↓
                            [SQLite/PostgreSQL]
                                    ↓
                        [Admin API] ←→ [Admin Web]
                        [Hospital API] ←→ [Nurse Terminal]
```

## API 端点总览

### Admin API (cloud/admin-api)
- `/api/v1/auth/login` - 管理员登录
- `/api/v1/users` - 用户管理
- `/api/v1/elderly` - 老人档案管理
- `/api/v1/devices` - 设备管理
- `/api/v1/alerts` - 告警管理
- `/api/v1/subscriptions` - 订阅管理
- `/api/v1/ota` - OTA 升级
- `/api/v1/medical/*` - 医疗工作流
- `/api/v1/regulatory/*` - 监管合规

### API Server (cloud/api-server)
- 设备接入
- 健康数据
- 定位服务
- 告警推送

### B2B API
- 医院患者管理
- 社区老人管理
- 保险结算

## 数据库 Schema

主库文件：`eregen.db` (SQLite)
迁移脚本：`init-db.sql`

核心表：
- `users` - 用户表
- `elderly` - 老人档案
- `devices` - 设备表
- `subscriptions` - 订阅表
- `alerts` - 告警表
- `medication_rules` - 用药规则
- `health_data` - 健康数据
- `b2b_institutions` - 机构表

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
最后更新：2026-08-02
