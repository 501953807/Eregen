# Eregen 平台 — 验收缺陷修复计划

> **执行者**: 需要多个子agent并行执行不同模块修复  
> **审查者**: 架构审计专家（本会话）  
> **依赖关系**: 任务间存在耦合（如JWT修复依赖于安全配置初始化）

## 一、修复范围总览

本次修复共8个独立任务，分为4个优先级组：

**P0 - 阻塞性修复（必须先完成）**
1. Task 1: 修复JWT硬编码密钥安全问题 (`cloud/admin-api/internal/router.go`)
2. Task 2: 统一数据库schema定义 (`cloud/admin-api/internal/store/` + `init-db.sql`)

**P1 - 高优先级（必须在发布前完成）**
3. Task 3: 添加认证接口速率限制 (`shared/ratelimit` + `cloud/api-server`)
4. Task 4: 补齐Handler单元测试 (`cloud/api-server/tests/` + `admin-api/tests/`)
5. Task 5: 注册缺失的endpoint (`cloud/api-server/internal/router/router.go`)

**P2 - 中优先级（持续改进）**
6. Task 6: 移除敏感错误信息暴露 (所有handler)
7. Task 7: 完善B2B业务逻辑 (hospital-api/community-platform/insurance-integration)
8. Task 8: CI/CD自动化测试配置 (`.github/workflows/`)

## 二、文件影响映射

| 任务 | 修改文件 | 新增文件 |
|------|---------|---------|
| Task 1 | `cloud/admin-api/internal/router.go` (1处) | - |
| Task 2 | `cloud/admin-api/internal/store/sqlite.go`, `store.go` | - |
| Task 3 | `cloud/api-server/internal/middleware/ratelimit.go`, `router/config.go` | - |
| Task 4 | `cloud/api-server/internal/handler/device_test.go`, `elderly_test.go` 等 | `cloud/admin-api/internal/handler/elderly_test.go` |
| Task 5 | `cloud/api-server/internal/router/router.go` | - |
| Task 6 | `cloud/api-server/internal/handler/*.go`, `admin-api/internal/handler/*.go` | - |
| Task 7 | `b2b/hospital-api/internal/handler/alert.go`, `community-platform/handler/care_plan.go`, `insurance-integration/handler/claim.go` | 补充业务逻辑代码 |
| Task 8 | `.github/workflows/ci.yml` | - |

## 三、各任务详细实施说明

### Task 1: JWT硬编码密钥修复

**文件**: `cloud/admin-api/internal/router.go` 第59-61行  
**问题**: `"change-me-in-production"` 默认值导致认证失效  
**修复**: 改为fatal error  

```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET environment variable is required for security") // ✅ 禁止生产环境遗漏
}
```

**验证**: 启动服务应fail-fast当JWT_SECRET未设置

---

### Task 2: 数据库Schema一致性修复

**文件**: 
- `cloud/admin-api/internal/store/sqlite.go` 
- `cloud/admin-api/internal/store/postgres.go`
- `cloud/admin-api/internal/store/store.go`

**问题**: SQLite实现缺少users表的phone列，与PostgreSQL和init-db.sql不一致  
**修复**: 确保store接口的`CreateUser`方法对应的表结构在所有存储实现中一致包含phone列

---

### Task 3: 添加认证接口速率限制

**文件**: 
- `cloud/api-server/internal/router/config.go` (添加rate limiter到auth group)
- `cloud/api-server/internal/router/router.go` (为register/login/send-otp添加专用限流器)

使用已有的 `shared/ratelimit` 滑动窗口实现

---

### Task 4: 补齐Handler单元测试

**测试覆盖目标包**:
- `api-server/internal/handler/auth.go` — 需增加：真实注册流、登录流程刷新token、设备绑定完整路径
- `api-server/internal/handler/device.go` — 测试List/Bind/Delete等CRUD
- `admin-api/internal/handler/elderly.go` — 需从0开始写测试文件
- `admin-api/internal/handler/device.go` — UpdateConfig/TriggerOTA/BatchOTA测试

测试策略：使用httptest.NewRequest + httptest.NewRecorder模拟HTTP请求，验证返回状态码和业务逻辑

---

### Task 5: 注册缺失的endpoint

**文件**: `cloud/api-server/internal/router/router.go`  
**发现**: `subscription.go`, `user_list.go`, `medication_take.go`, `alert_handle.go` 四个handler已存在于`missing_endpoints.go`，但未在router中注册  
**修复**: 在router的protected分组中添加对应路由处理：

```go
// 在protected.Group内添加：
protected.GET("/subscriptions", subH.List)           // SubscriptionHandler.List
protected.GET("/stats/subscriptions", subH.Stats)     // SubscriptionHandler.Stats
protected.GET("/users", userListH.List)               // UserListHandler.List
protected.GET("/users/:id", userListH.Get)            // UserListHandler.Get
protected.PUT("/admin/users/:id/role", userListH.UpdateRole) // UserListHandler.UpdateRole
protected.POST("/medication/:rule_id/take", medTakeH.Take)     // MedicationTakeHandler.Take
protected.PUT("/:alert_id/handle", alertHandleH.Handle)        // AlertHandleHandler.Handle
protected.POST("/alerts/share-location", alertHandleH.ShareLocation) // AlertHandleHandler.ShareLocation
```

---

### Task 6: 移除敏感错误信息暴露

**批量修改**: 遍历所有handler，将以下模式从：
```go
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
```
改为：
```go
log.Printf("Internal error: %v", err) // 记录错误但不泄露
c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "系统内部错误"})
```

涉及文件:
- `cloud/api-server/internal/handler/elderly.go`
- `cloud/api-server/internal/handler/device.go` 
- `cloud/api-server/internal/handler/*`
- `admin-api/internal/handler/*`

---

### Task 7: B2B业务逻辑补全

**hospital-api**:
- `alert.go` Forward函数实际无业务逻辑，需实现从alerts表查询并转发给指定机构的流程
- `health_data.go`: 需实现医院端健康数据推送的接收和处理

**community-platform**:
- `care_plan.go`: 实现关怀计划的CRUD + 状态流转
- `event.go`: 社区活动管理
- `health_check.go`: 健康检查记录管理

**insurance-integration**:
- `claim.go`: 理赔申请流程
- `policy.go`: 保单管理
- `provider.go`: 医疗机构/provider注册管理

---

### Task 8: CI/CD自动化测试配置

**文件**: `.github/workflows/ci.yml`  
**现状**: README中提到该文件但未在仓库中可见  
**任务**: 确保workflow包含：
- Go build + test -cover
- Flutter analyze
- npm run build (for admin-web)

验证步骤：触发PR时自动运行上述任务

## 四、并行执行建议

以下任务可并行执行（子agent分发）：
- **Task 1 & Task 2**: 各自独立，可并行
- **Task 3 & Task 5**: 分别修改router的不同部分，可并行
- **Task 4**: 每个文件的单元测试可分派不同子agent
- **Task 6**: 批量错误处理修改，可并行处理不同handler文件
- **Task 7**: 每个B2B服务的业务逻辑开发可独立进行

## 五、最终验收验证清单

执行完所有修复后，通过以下标准：
- [ ] JWT_SECRET环境变量检查成功触发fatal错误
- [ ] 所有Go包单元测试通过率≥80% (`go test ./... -cover`)
- [ ] 所有声明的endpoint均可访问并返回正确响应
- [ ] 错误响应不再包含原始DB错误信息
- [ ] B2B services有实际业务逻辑（非空handler）
- [ ] CI pipeline可成功触发并通过所有测试

ENDOFFE
echo "Plan content prepared at /tmp/acceptance_fixes_plan.md"