# 第八阶段：综合安全审查报告 — Comprehensive Security Assessment (SA-P08)

**审查日期：** 2026-07-28
**审查范围：** 全部代码仓库（固件/后端/B2B/共享库）
**输出格式：** 此文件同时作为执行计划入口点

---

## 一、审查综述

本阶段对 Eregen 平台进行了端到端的安全审计，覆盖所有子系统。基于前序 7 阶段的发现与修复，本审查提供统一的安全态势视图、攻击面分析、风险评级矩阵及生产就绪决策建议。

**总发现问题数：** 20项（Critical: 5, High: 6, Medium: 7, Low: 2）

---

## 二、完整问题清单与风险矩阵

### Critical 级（阻塞生产，需立即修复）

| # | 模块 | 问题描述 | 风险等级 | 关联依赖 | 修复方案 |
|---|------|---------|---------|---------|---------|
| C-01 | `admin-api/store/postgres.go:173` | `UpdateDeviceConfig` SQL注入——config map key直接拼接至SQL | **CRITICAL** | BUG-001 | 改用 `json.Marshal()` + 参数化查询 |
| C-02 | `b2b/*/cmd/main.go:21` | B2B服务硬编码数据库密码 `postgres/password` | **CRITICAL** | BUG-002 | 移除hardcoded，env缺失时panic退出 |
| C-03 | `api-server/internal/handler/device.go:319-512` | 6个admin设备接口无任何RBAC保护 | **CRITICAL** | BUG-003 | 添加 `auth.RequireRole(model.RoleAdmin)` |
| C-04 | `missing_endpoints.go:119-135` | `/api/v1/admin/users/:id/role` 任意认证用户可提升权限 | **CRITICAL** | BUG-004 | RequireRole(RoleAdmin) + 检查目标不为自己 |
| C-05 | `gateway/internal/config/config.go:137` | JWT secret 默认值 `"dev-secret-key-change-in-production"` | **CRITICAL** | 密钥泄露导致伪造token | 从环境变量强制加载，无则panic |

### High 级（发布前必须完成修复）

| # | 模块 | 问题描述 | 风险等级 | 关联依赖 | 修复方案 |
|---|------|---------|---------|---------|---------|
| H-01 | `shared/crypto/crypto.go:42` | 密码哈希用SHA-256而非bcrypt/argon2 | HIGH | 密码泄露后易被暴力破解 | 替换为 `golang.org/x/crypto/bcrypt` |
| H-02 | `router/router.go:33-35` | `/api/v1/health` 仅返回静态"ok"，无依赖探测 | HIGH | liveness probe误判健康 | 增加DB/NATS/Redis连接探测 |
| H-03 | `store/redis.go:101-123` | Refresh Token无一次性消耗保护 | HIGH | token被盗后可无限刷新 | 消费后立即删除token或轮换机制 |
| H-04 | `device.go:HandleTelemetry` | 未验证设备上传的elderly_id与设备绑定关系 | HIGH | 任意设备可伪造老人数据 | 查询device.owner_elderly_id对比payload |
| H-05 | `missing_endpoints.go:197-224` | AlertHandle和ShareLocation无权限检查 | HIGH | 任意用户可解决任意告警 | RequireRole + 资源归属校验 |
| H-06 | `crypto/wb_aes.go:61-72` | BLE会话密钥推导直接用SHA256(配对码+设备ID)，无KDF盐 | HIGH | 离线彩虹表攻击可行（配对码仅10^4种） | 改用PBKDF2/Argon2，配位数增至8位 |

### Medium 级（版本迭代中修复）

| # | 模块 | 问题描述 | 风险等级 | 修复方案 |
|---|------|---------|---------|---------|
| M-01 | `migrate.sql` | `med_status_records` 缺 compartment/device_id字段 | Medium | 扩展schema兼容协议定义 |
| M-02 | `cmd/main.go:35` | Postgres连接池MaxConns=过小 | Medium | 调至20-50 |
| M-03 | `postgres.go:1089` | DeleteUser无分页/批量删除控制 | Medium | 增加分页或cascade约束 |
| M-04 | `store/sqlite.go:97,131` | ListUsers WHERE动态列拼接（当前safe但存在隐患） | Low-Medium | 保持当前但禁止引入用户输入 |
| M-05 | `firmware/bracelet/pro/main.c` | Pro版跌倒检测算法尚未实现 | Medium | 开发跌倒检测算法模块 |
| M-06 | `bridge/binding/protocol.go` | OTA命令缺少签名验证 | Medium | 添加消息完整性校验 |
| M-07 | `handler/auth.go` | OTP生成使用math/rand而非crypto/rand | Medium | 改用crypto/rand |

### Low 级（长期改进）

| # | 模块 | 问题描述 | 建议 |
|---|------|---------|------|
| L-01 | 多处HTTP 404响应 | 暴露资源存在性 | 统一返回 generic error message |
| L-02 | audit_logs表无旧值记录 | 无法追溯变更前后状态 | 增加before/after JSON列 |

---

## 三、攻击路径模拟（Attack Path Simulation）

### 路径 A: 初始访问 → 权限提升至系统控制

```
┌──────────────────────────────────────────────────────────────┐
│ 1. Scan: Target knows server URL from docs or public repos     │
│    - Attempt login with test/test123 (seed account in admin-api)│
│    - If seed not removed → gains admin access ✓               │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 2. Exploit: Admin device endpoint without RBAC                 │
│    PUT /api/v1/admin/devices/DEV-001/settings                │
│    Inject settings: {"malicious": "value; DROP TABLE users;"}│
│    Executes SQL via UpdateDeviceConfig漏洞                    │
│    Result: Arbitrary DB operations possible                   │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 3. Escalate: Leverage DB to dump user tables                 │
│    SELECT * FROM users; → obtain all user passwords (hashed)  │
│    Since stored with SHA-256 (H-01), offline brute force      │
│    feasible at ~100M hashes/sec on consumer GPU              │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 4. Persistence: Forge JWT token using known secret             │
│    Secret is "dev-secret-key-change-in-production" (C-05)     │
│    Create token with role=admin, sub=hacker → full control   │
└──────────────────────────────────────────────────────────────┘
```

### 路径 B: 设备侧冒充 → 数据污染与告警欺骗

```
┌──────────────────────────────────────────────────────────────┐
│ 1. Attacker obtains any valid device ID (public GET /devices)│
│    e.g., "BR-XXXX" (from example in protocol spec)          │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 2. Forged heartbeat/health/telemetry payload with fake      │
│    elderly_id="hacked-user-123"                             │
│ POST /api/v1/devices/telemetry {"type":"health",            │
 │   "dev_id":"BR-XXXX", "elderly_id":"hacked-user-123",       │
 │   "hr":999}                                               │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 3. HandleTelemetry accepts without verifying owner           │
│ Data injected into health_records for elderly_id            │
│ → AI analysis sees fabricated abnormal vitals                │
│ → Potential wrong medical alert/action                      │
└──────────────────────────────────────────────────────────────┘
```

### 路径 C: Refresh Token 窃取 → 会话劫持

```
┌──────────────────────────────────────────────────────────────┐
│ 1. User authenticates → receives refresh token RT1         stored in Redis                   │
│ RT1 stays valid until expiry (no one-time consumption)        │
└─────────────────┬─────────────────────────────────────────────┘
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 2. Attacker steals RT1 (via network sniffing, XSS, etc.)     │
│ Repeatedly calls POST /auth/refresh with RT1                 │
│ Each call returns new access token                           │
│ Session persists indefinitely                               │
└──────────────────────────────────────────────────────────────┘
```

---

## 四、安全配置基准对照表

| 检查项 | 当前状态 | 期望状态 | 合规性 |
|--------|---------|---------|--------|
| 密码哈希算法 | SHA-256 (`shared/crypto/crypto.go`) | bcrypt/argon2 | ❌ 不合规 |
| JWT密钥管理 | 硬编码默认值 (`gateway/internal/config/config.go`) | Env var + fallback panic | ❌ 不合规 |
| DB凭证管理 | B2B服务硬编码密码 (`b2b/*/cmd/main.go`) | Env only | ❌ 不合规 |
| SQL注入防护 | UpdateDeviceConfig拼接SQL | 参数化查询 | ❌ 不合规 |
| RBAC管理 | Admin设备端点无角色校验 | RequireRole per endpoint | ❌ 不合规 |
| Rate Limiting | 无限流 | 已计划实施 (security-hardening plan) | ⚠️ 部分 |
| TLS验证 | MQTT InsecureSkipVerify=true | Production mode requires CA validation | ⚠️ 可配置 |
| Refresh Token | 无一次消费保护 | One-use or rotation | ❌ 不合规 |
| 健康检查 | Static response only | Dependency probes | ❌ 不合规 |
| 输入验证 | 部分handler有binding tags | All params validated via shared/validation | ⚠️ 部分 |

---

## 五、安全加固实施路线图

### 紧急补丁（P0 - 24小时内修复）

1. **修复 C-01 (SQL注入)**: `cloud/admin-api/internal/store/postgres.go`
   ```go
   // Before - DANGEROUS
   settings += fmt.Sprintf(`"%s":%v,`, k, v)
   
   // After - SAFE
   data, _ := json.Marshal(config)
   _, err = s.db.ExecContext(ctx, `UPDATE devices SET settings = settings || $1::jsonb WHERE device_id = $2`, data, deviceID)
   ```

2. **修复 C-02/B2B硬编码密码**: Remove default values from all b2b services
   ```go
   dbDSN := os.Getenv("DB_DSN")
   if dbDSN == "" {
       log.Fatal("DB_DSN environment variable is required")
       os.Exit(1) // NOT fallback to hard-coded credentials!
   }
   ```

3. **修复 C-03/C-04 (RBAC)**: Add auth middleware to all admin routes
   ```go
   admin := protected.Group("/admin")
   {
       admin.Use(auth.RequireRole(model.RoleAdmin))  // ← ADD THIS
       admin.GET("/devices", deviceH.AdminList)
       ...
   }
   ```

### 快速修复（P1 - 3天内完成）

4. **修复 H-01 (bcrypt)**: Add bcrypt functions and migrate existing passwords on first login
5. **修复 H-02 (health check)**: Add dependency probing to `/api/v1/health` endpoint
6. **修复 H-03 (refresh token)**: Consume token upon use or implement rotation
7. **修复 H-04 (telemetry validation)**: Verify elderly_id matches device owner in HandleTelemetry
8. **修复 H-06 (BLE key derivation)**: Use PBKDF2 with sufficient iterations

### 中长期修复（P2 - 版本迭代）

9. Implement input validation throughout handlers using shared/validation package
10. Add rate limiting to all endpoints
11. Complete med_status reporting pipeline (device handler + schema update)
12. Implement proper audit logging with before/after state tracking

---

## 六、生产就绪决策矩阵

### 通过条件（All of the following must be met）

| 检查项 | 要求 | 当前状态 |
|--------|------|---------|
| Critical 级漏洞修复 | All C-01 to C-05 resolved | ❌ Pending |
| High 级漏洞修复 | All H-01 to H-06 addressed | ❌ Pending |
| 密码哈希迁移 | Existing passwords re-hashed with bcrypt on next login | ❌ Not started |
| RBAC全覆盖 | All authenticated endpoints have correct role checks | ❌ Partial |
| 依赖健康检查 | /health verifies DB, NATS, Redis, FCM | ❌ Missing |
| Token保护 | Refresh tokens are one-time use or rotated | ❌ Missing |
| 审计日志 | Log all admin actions with user/IP/timestamp | ❌ Incomplete |
| CI/CD安全门禁 | Tests + SAST gate in GitHub Actions | ⚠️ Plan exists but not implemented |

### 综合评分

```
安全成熟度：LOW (35%) —— 多个Critical及High级别问题未修复
   
关键风险指标：
  ├─ 机密性风险：HIGH (password hashes, JWT keys, DB creds exposed)
  ├─ 完整性风险：HIGH (SQL injection, device data spoofing)
  └─ 可用性风险：MEDIUM (no rate limiting, small connection pool)
```

**结论：❌ 当前代码库未达到生产部署标准。**

建议按以下顺序修复后重新评估：

1. 第1天：修复 C-01, C-02, C-03, C-04, C-05 (Critical)
2. 第2天：修复 H-01, H-02, H-03, H-04 (High priority)
3. 第3天：实施 H-05, H-06 并完成密码哈希迁移策略
4. 第4天：添加健康检查端点，验证所有修复

---

## 七、附录：遗留安全问题跟踪

**待记录但未在本阶段详述的问题：**

1. **硬件固件层面**（留待后续硬件安全专项审查）
   - 蓝牙配对过程无防重放保护
   - OTA升级缺乏完整性/来源验证
   - SOS触发时无加密通信信道确认

2. **应用层前端**（Flutter/Vue未详细审查）
   - Flutter APP中是否硬编码API endpoint？
   - Vue管理后台是否有敏感信息XSS风险？
   - 微信小程序是否有会话管理漏洞？

3. **基础设施层面**
   - Docker Compose配置中的secret管理
   - Kubernetes secrets配置（如适用）
   - 网络隔离策略（VPC间通讯）

> **注：** 上述内容将在第九阶段（Frontend & Infrastructure Security Review）中深入审查。

---

*本安全审查报告与 `security-hardening-plan.md` 配套使用，后者提供了每个问题的具体修复代码实现。*