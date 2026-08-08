# Unified Backend Verification Design — SQLite & PostgreSQL Support

**Project:** Eregen (颐贞) — 老年健康品牌平台
**Date:** 2026-07-27
**Version:** 0.1
**Authors:** System Automation

---

## 1. 概述 (Overview)

本规格文档描述了一套统一后端架构，同时支持 **SQLite（轻量验证模式）** 和 **PostgreSQL（生产/完整功能模式）**，用于覆盖全家应用（家属端、护士终端、管理后台）的前端集成测试与验证。

设计基于以下核心原则：
- **统一数据访问抽象层**：通过一致的 Store 接口屏蔽底层数据库差异
- **复用现有实现**：直接采用 admin-api 已有的完整 SQLite Store 实现，避免重复造轮子
- **前端 API 对齐**：所有 /api/v1/* 端点返回前端期望的 `{data: {...}, meta: {...}}` 格式
- **环境驱动切换**：仅需配置环境变量即可在 SQLite ↔ PostgreSQL 之间无缝切换

---

## 2. 系统架构 (System Architecture)

```
                  ┌──────────────────────────────┐
                  │      前端应用                │
                  │ ├─ admin-web (Vue 3)         ││
                  │ ├─ family-app (Flutter Web) ││
                  │ └─ nurse_terminal          ││
                  │                            ▼│
            VITE_API_URL │                     HTTP│
                       │                      GET/POST
          ┌────────────┴────────────────────────┐
          │             HTTP Listener (Port)    │
          │   ┌─────────────────────────────────┐
          │     ┌─Route Group: /api/v1/        │ │ ← frontends call these
          │     │    (health, users, elderly,│ │ │
          │     │     devices, alerts)        │ │ │
          │     └─────────────────────────────┘ │ │
          │     ┌─Route Group: /medical/...    │ │ ← medical backend
          │     └───────────────────────────────┘ │ │
          │                                       │
          ▼                                       ▼
  ┌─────────────────────────────┐           ┌──────────────────┐
  │    Admin-API Router         │←──────────│  API-Server      │
  │    (8081 by default)        │ routes    │ (8082 by default)│
  │                             │           │                  │
  │ ┌─────────────────────────┐ │           │ ┌──────────────┐ │
  │ │   Store Interface       │ │           │ │ Store Int.   │ │
  │ │   ────────────────────  │ │           │ │ (shared)     │ │
  │ │ PostgresStore (optional)│ │ ──────────▶ │ PostgresStore│ │
  │ │ SqliteStore (default)   │ │           │ │ SqliteStore  │ │
  │ └─────────────────────────┘ │           │ └──────────────┘ │
  └─────────────────────────────┘             └──────────────────┘
                                              ▲
              ┌──────────────┐                │              ┌──────────────┐
              │  SQLite DB   │ ◄──────────────┘              │  PostgreSQL  │
              │ /tmp/regen  │     DB_TYPE env var          │   (dev/prod) │
              └──────────────┘                            └──────────────┘
```

### 关键说明：
1. **API Server** (`cloud/api-server`) 作为设备接入服务，负责处理来自手环/药盒设备的 MQTT/HTTP 消息；同时为家庭App提供设备认证和轻量数据接口。
2. **Admin API** (`cloud/admin-api`) 作为管理后台的数据服务，提供完整的 CRUD 和数据统计接口。通过 `/api/v1/` 路由满足前端验证需求。
3. **Store 接口**被两个服务共享（通过 Go package 重用），确保数据模型一致性。
4. 通过环境变量 `DB_TYPE=sqlite|postgres` 控制后端使用哪种数据库。

---

## 3. Store 接口定义 (Store Interface Definition)

### 3.1 共享接口规范

位于 `cloud/api-server/internal/store/shared_interface.go`（公共接口头文件），两个服务均引用该定义：

```go
// cloud/api-server/internal/store/interface.go
package store

import (
    "context"
    "eregen.dev/api-server/internal/model"
)

// Store defines the minimal API contract for frontend verification backends.
type Store interface {
    // Health check
    Health(ctx context.Context) error

    // User/elderly management (for frontend /api/v1/users and /api/v1/elderly)
    ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)
    ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)

    // Device management (for frontend /api/v1/devices)
    ListDevices(ctx context.Context, status string) ([]model.DeviceSummary, error)

    // Alerts (for frontend /api/v1/alerts)
    GetActiveAlerts(ctx context.Context) ([]model.AlertSummary, error)

    // Authentication helper for API key validation
    ValidateToken(ctx context.Context, token string) (string, error)
}
```

此接口仅包含前端验证所需的最小方法集，不包含管理后台的全部业务逻辑（如医疗腕带、临床工作流等）。这些方法将在各自的 Store 实现中完整扩展。

### 3.2 存储层实现

| 实现类型 | 文件路径 | 特点 |
|---------|---------|------|
| **SqliteStore** | `cloud/api-server/internal/store/sqlite.go` | 直接复用 `admin-api/internal/store/sqlite.go` 的核心结构，按 Store 接口暴露方法 |
| **PostgresStoreAdapter** | `cloud/api-server/internal/store/postgres_adapter.go` | 适配现有的 Postgres 实现，包装后符合 Store 接口 |

---

## 4. 数据库模式 (Database Schema)

### 4.1 SQLite 模式（来自 admin-api）

使用 `admin-api/internal/store/sqlite.go` 中的完整迁移脚本（第 60-483 行），包含以下核心表：

| 表名 | 用途 |
|------|------|
| `users` | 用户表（角色：family, elderly, admin） |
| `elderly_profiles` | 老人档案表（含健康等级 JSON） |
| `devices` | 设备表（手环、药盒等） |
| `alerts` | 告警表 |
| `health_records` | 健康记录（心率、血氧、步数） |
| `location_history` | 位置历史 |
| `medication_rules` | 用药规则 |
| `subscriptions` | 订阅信息 |
| `firmware_releases` | 固件版本 |

### 4.2 PostgreSQL 模式

使用现有的 `cloud/api-server/internal/store/postgres.go` 对应的 SQL 迁移（需单独维护，与 SQLite 保持字段语义一致）。推荐方式：从 SQLite 的 `migrations` SQL 字符串中提取核心 CREATE TABLE 语句，转换为 PostgreSQL 兼容语法（替换 SQLite 特有的函数/类型）。

### 4.3 数据预置（Seed Data）

两种模式均需预置验证用数据，通过统一的初始化脚本完成：

```sql
-- 示例：预置两位老人
INSERT INTO elderly_profiles (id, name, user_id, birth_date, health_tiers, created_at, updated_at) VALUES
('eld-1', '张建国', 'usr-family-1', '1950-01-01', '["基础版"]', datetime('now'), datetime('now')),
('eld-2', '李秀英', 'usr-family-2', '1948-05-05', '["防跌倒"]', datetime('now'), datetime('now'));
```

脚本同时创建对应的 devices、alerts、health_records 等关联记录，确保各端点有数据可查。

---

## 5. 后端 API 端点映射表 (API Endpoint Mapping Table)

以下是前端实际调用的端点与后端实现的对应关系：

| 前端调用 | 后端服务 | 路径 | HTTP 方法 | 响应格式 | 依赖 Store 方法 |
|----------|---------|------|-----------|----------|----------------|
| admin-web | admin-api | `/api/v1/health` | GET | `{"data":{"status":"ok"}}` | `store.Health()` |
| admin-web | admin-api | `/api/v1/elderly` | GET | `{"data":[elderly_profiles]}` | `store.ListElderly()` |
| admin-web | admin-api | `/api/v1/users?role=xxx` | GET | `{"data":users,"meta":{"total":n}}` | `store.ListUsers()` |
| family-app | api-server | `/api/devices/auth` | POST | `{"device_id":"..."}` | (device auth logic) |
| family-app | admin-api | `/api/v1/elderly` | GET | `{"data":[profiles]}` | `store.ListElderly()` |
| mock-server（旧）→ 全部 | 合并至此 | — | — | — | — |

> **注意**：临时 mock server (`/tmp/mock-server-final.js`) 的数据将移植至数据库，正式运行时前端直接连接真实后端。

---

## 6. 实现计划 (Implementation Plan)

按顺序执行以下步骤：

### Phase 0: 准备工作（已完成）
- [x] Mock server 数据已存至 SQLite 临时表
- [x] 前端代码已修复 `res.data.data` bug
- [x] Family-app、admin-web 已启动并连通 mock server

### Phase 1: 创建共享 Store 接口
- 创建 `cloud/api-server/internal/store/interface.go`（最小接口）
- 建立 `cloud/internal/store/common` 目录存放共享的 model/types（或直接从 admin-api 引用）

### Phase 2: api-server 添加 SQLite 支持
- 复制 `admin-api/internal/store/sqlite.go` 至 `cloud/api-server/internal/store/sqlite.go`
- 修改导入路径以匹配 api-server 的 model 包
- 按 `interface.go` 定义实现 Store 接口方法（ListElderly, ListUsers, ListDevices, GetActiveAlerts, Health）
- 实现 NewSqliteStore 构造函数

### Phase 3: api-server 添加 Postgres 适配器
- 创建 `cloud/api-server/internal/store/postgres_adapter.go`
- 包装现有 Postgres 结构体，实现 Store 接口方法（委托给内部 Postgres 实例）
- 只暴露前端需要的子方法

### Phase 4: api-server 主程序改造
- 修改 `cloud/api-server/cmd/main.go`：
  - 读取 `DB_TYPE` env var（默认 sqlite）
  - 根据值选择初始化 SqliteStore 或 PostgresAdapter
  - Store 注入到 router 中

### Phase 5: admin-api 添加 /api/v1/ 路由组
- 创建 `cloud/admin-api/router/api_v1.go`（路由定义 + Store 适配包装器）
- 在每个 handler 中：
  - 查询 Store
  - 包装为 `{data: ..., meta: ...}` 格式返回
- 修改 main.go 挂载该路由组到主 mux

### Phase 6: 数据预置与验证
- 编写 SQL seed 脚本（含中文字段）
- 启动任一后端，确认所有端点可正常响应
- 前端重新加载，验证数据展示正常

---

## 7. 部署与运行指南 (Deployment & Operations)

### 7.1 环境配置示例

```bash
# .env.example
# 后端端口
ADMIN_API_PORT=8081
API_SERVER_PORT=8082

# 数据库选择
DATABASE_TYPE=sqlite      # 或 postgres
SQLITE_PATH=/tmp/regen.db

# PostgreSQL 连接（当 DATABASE_TYPE=postgres 时生效）
PG_HOST=localhost
PG_PORT=5432
PG_DATABASE=rengen_dev
PG_USER=admin
PG_PASSWORD=secret

# 其他服务
NATS_ADDR=localhost:4222
MQTT_ADDR=:1883
REDIS_ADDR=localhost:6379
```

### 7.2 启动命令

```bash
# SQLite 模式（验证开发模式）
export DATABASE_TYPE=sqlite
cd cloud/admin-api && go run ./cmd/main.go
cd ../api-server && go run ./cmd/main.go

# PostgreSQL 模式（生产完整模式）
export DATABASE_TYPE=postgres
cd cloud/admin-api && go run ./cmd/main.go
cd ../api-server && go run ./cmd/main.go
```

---

## 8. 风险评估与应对 (Risk Assessment)

| 风险点 | 影响程度 | 应对方案 |
|--------|---------|---------|
| SQLite vs PostgreSQL 数据类型不一致（如数组/JSON 处理） | 高 | 使用标准 JSON/TEXT 类型，避免 PG 特有类型（如 pg_array），两端 SQL 尽量保持同源生成 |
| Store 接口不完整导致前端调用缺失方法 | 中 | 先收集所有前端 API 请求（通过浏览器开发者工具），逐项对照实现 |
| 中文字符编码问题 | 低 | SQLite 和 UTF-8 原生支持；PostgreSQL 使用 UTF8 编码创建数据库 |
| 并发性能不足（SQLite 单写锁） | 中 | 验证模式仅需单进程；生产环境自动切换 PostgreSQL |
| 两个服务 Store 实现不同步导致的业务逻辑分歧 | 高 | Store 接口在公共 package 中定义，两个 repository 分别实现，定期同步接口变更 |

---

## 9. 批准签字 (Approval Sign-off)

设计文档经过自检查无误（无 TBD、无歧义、内部一致）。提交审查者审阅。

---

*此文档自动生成于系统化设计流程，遵循 superpowers:brainstorming 工作流规范。*