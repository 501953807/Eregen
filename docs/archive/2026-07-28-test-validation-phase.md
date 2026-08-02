# 第七阶段：测试验证计划 — Test Validation Phase (TV-P01)

**阶段目标：** 基于前序各阶段发现的缺陷与缺口，设计并执行覆盖核心业务路径的安全测试与功能验证，确保已修复问题回归通过、新引入变更无副作用。

**范围：** 重点覆盖 Phase 5 用药管理闭环、Phase 6 数据库安全审查发现的高/中风险项、以及关键认证与数据校验路径。

---

## 一、测试分类与覆盖矩阵

| # | 测试类型 | 覆盖模块 | 优先级 | 备注 |
|---|----------|---------|--------|------|
| TV-01 | 单元测试（补全） | store/postgres.go 参数字段校验 | High | 当前大量函数无unit test |
| TV-02 | 安全测试 | UpdateDeviceConfig SQL注入防护 | Critical | 验证修复后是否阻断注入 |
| TV-03 | 权限测试 | admin/* 端点 RBAC 缺失验证 | Critical | 普通user能否调用admin接口 |
| TV-04 | 集成测试 | med_status 上报链路完整模拟 | High | 设备→MQTT→NATS→DB→Alert |
| TV-05 | 边界测试 | medication rule schedule time窗口 | Medium | 提前/延后服药如何处理 |
| TV-06 | 回归测试 | 修复后的所有相关端点回归 | Medium | 防止修复引入新问题 |
| TV-07 | 压力测试 | Postgres连接池瓶颈验证 | Low | MaxConns=10在高并发下表现 |

---

## 二、具体测试用例设计

### TV-01: Store层参数完整性单元测试

**文件:** `cloud/api-server/internal/store/postgres_test.go` （新建）

```go
// TestPostgres_CreateMedicationRule_ValidInputs verifies all required fields
func TestPostgres_CreateMedicationRule_ValidInputs(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    pg := store.NewPostgres(db, zap.NewNop())
    rule := &model.MedicationRule{
        ElderlyID:   "elderly-123",
        ScheduleTime: "08:00",
        DoseCount:   1,
        PillType:    "capsule",
        DaysOfWeek:  []int{1, 2, 3},
        Active:      true,
    }

    err := pg.CreateMedicationRule(context.Background(), rule)
    assert.NoError(t, err)

    // Verify record exists with correct values
    var stored model.MedicationRule
    err = db.QueryRow("SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active FROM medication_rules WHERE id = $1", rule.ID).Scan(
        &stored.ID, &stored.ElderlyID, &stored.ScheduleTime, &stored.DoseCount, &stored.PillType, &stored.DaysOfWeek, &stored.Active)
    assert.NoError(t, err)
    assert.Equal(t, "elderly-123", stored.ElderlyID)
    assert.Equal(t, "08:00", stored.ScheduleTime)
}

// TestPostgres_UpdateMedicationRule_HandlersEmptyUpdate verifies no-op update works
func TestPostgres_UpdateMedicationRule_NoChange(t *testing.T) {
    // ... setup ...
    err := pg.UpdateMediationRule(ctx, ruleID, &model.CreateMedicationRuleRequest{})
    // Should succeed but not modify data (or return specific error if disallowed)
}
```

待覆盖的核心函数（按调用频率排序）：
- Create/Update/DeleteMedicationRule
- GetTodayMedStatus / GetMedicationHistory
- CreateAlert / ResolveAlertByID
- ListDevices / GetDeviceByDeviceID
- BindDevice / DeleteDevice

### TV-02: SQL注入安全测试 — UpdateDeviceConfig修复验证

**文件:** `cloud/admin-api/internal/store/postgres_sqlinject_test.go`

```go
// TestStore_UpdateDeviceConfig_SQLInjection attempts various injection payloads
func TestUpdateDeviceConfig_SQLInjection(t *testing.T) {
    tests := []struct {
        name     string
        config   map[string]interface{}
        expectErr bool
    }{
        {
            "SQL injection via key name",
            map[string]interface{
                `"'); DROP TABLE devices; --": "value",
            },
            true,
        },
        {
            "JSON payload manipulation",
            map[string]interface{
                "settings": `{"malicious": "value'"}`,
            },
            true, // Should be escaped properly
        },
        {
            "Valid configuration (white-listed keys)",
            map[string]interface{
                "volume": 80,
                "interval": 30,
            },
            false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := store.UpdateDeviceConfig(context.Background(), "DEV-001", tt.config)
            if tt.expectErr {
                assert.Error(t, err, "should reject malicious input")
            } else {
                assert.NoError(t, err, "should accept valid config")
            }
        })
    }
}
```

### TV-03: RBAC权限测试 — Admin端点越权访问

**文件:** `cloud/api-server/internal/router/rbac_test.go`

```go
// TestAdminEndpoints_AccessWithoutAdminRole verifies non-admin users cannot access admin endpoints
func TestAdminEndpoints_AccessWithoutAdminRole(t *testing.T) {
    router := gin.Default()
    
    // Setup two users: regular family user and admin user
    familyUser := createFakeUser("family", "family-123")
    adminUser := createFakeUser("admin", "admin-456")

    // Test as family user (no permission)
    c := createGinContextWithAuth(familyUser)
    // Call deviceH.AdminList or any admin endpoint
    // Expect 403 Forbidden, not 200 OK

    // Test as admin user (should pass)
    c2 := createGinContextWithAuth(adminUser)
    deviceH.AdminList(c2)
    // Expect 200 OK with device list
}

// TestUpdateRole_Endpoint_RequiresAdminRole verifies only admin can change roles
func TestUpdateRoleEndpoint_RequiresAdmin(t *testing.T) {
    // Family user tries to PUT /api/v1/admin/users/some-id/role
    // → should receive 403, NOT succeed in changing the role
}
```

### TV-04: 用药上报集成测试 — Med Status End-to-End

这是一个需要模拟的完整链路测试：

```go
func TestMedStatusWorkflow_E2E(t *testing.T) {
    // 1. Setup: Create a pillbox device and link it to an elderly profile
    pillboxID := "PX-MED001"
    elderlyID := "elderly-test-123"
    device := createDevice(pillboxID, "pillbox", "pro", ownerUserID)
    linkElderlyToDevice(elderlyID, pillboxID) // Requires elderly_devices table

    // 2. Setup: Create a medication rule for this elderly
    rule := createMedicationRule(elderlyID, "08:00", 1, "aspirin")

    // 3. Simulate device sending med_status via MQTT → NATS
    // This requires starting the gateway service or mocking the event callback
    sendMedStatusToServer(pillboxID, compartment: 3, taken: true, timestamp: now())

    // 4. Verify: med_status_records created in DB with correct fields
    var count int
    db.QueryRow("SELECT COUNT(*) FROM med_status_records WHERE device_id=$1 AND compartment=$2", pillboxID, 3).Scan(&count)
    assert.Equal(t, 1, count, "med status record should exist")

    // Verify: if taken=false, missed_at should be set
    // Verify: notification/alert generated if missed (med_missed alert type)

    // 5. Verify: Family user can see this record via GET /medication/history
    // As family user, query history → should include the newly recorded status
}
```

> **注意：** 此测试目前无法执行，因为：
> - 缺少elderly_devices表关联结构
> - device.go未实现med_status处理逻辑
> - NATS event回调需上游触发源
> - firmware ESP32侧无上报实现

### TV-05: 用药时间窗口边界测试

```go
func TestMedicationTakeHandler_TimeWindowValidation(t *testing.T) {
    rules := []struct {
        name          string
        scheduleTime  string
                takeTime     string // 用户点击"已服药"的时间
                expectAllowed bool
    }{
        {"Take exactly on schedule", "08:00", "08:00:00", true},
        {"Take 5 min early", "08:00", "07:55:00", false}, // Should allow configurable window
        {"Take 30 min late", "08:00", "08:30:00", false},
        {"Take hours later", "08:00", "12:00:00", false},
    }

    for _, tt := range rules {
        t.Run(tt.name, func(t *testing.T) {
            // Create rule with scheduleTime
            // Simulate user calling POST /medication/:rule_id/take at takeTime
            // Verify: allowed != expected → fail
        })
    }
}
```

当前实现**没有时间窗口校验**，任何时刻只要rule active都可以标记为已服用（Take），这是明显的安全/业务逻辑缺陷。

### TV-06: Refresh Token重放攻击测试

```go
func TestRefreshToken_ReplayAttackProtection(t *testing.T) {
    // 1. User logs in, gets refresh token RT1
    rt1 := acquireRefreshToken("user-123")

    // 2. Use RT1 to get new access token
    newToken1 := useRefreshToken(rt1)

    // 3. Try to reuse RT1 (replay attack simulation)
    _, err := useRefreshToken(rt1) // Should FAIL – token should have been consumed

    if err == nil {
        t.Fatal("refresh token should be one-time use after consumption, but replay succeeded!")
    }
}
```

当前实现中`ValidateRefreshToken`仅检查存在性不删除token，存在明确的重放风险。

---

## 三、测试执行计划

### 阶段划分

| 周次 | 任务 | 交付物 |
|------|------|--------|
| Week 1 | TV-01 (store单元补测) + TV-03 (RBAC验证) | 新增15+ unit tests，覆盖所有CRUD操作 |
| Week 2 | TV-02 (SQL注入测试) + TV-05 (时间边界) | 注入用例覆盖10+常见payload，时间窗口规则文档化 |
| Week 3 | TV-04 (Med Status E2E) — 需要先实现服务端 handler | 端到端集成测试套件，依赖固件模拟 |
| Week 4 | TV-06 (Refresh Token攻击) + TV-07 (连接池压力) | 安全测试报告，性能基准数据 |

### 执行前提条件

TV-04 (Med Status E2E) **依赖服务端实现尚未完成的组件**，必须先完成以下修复才能执行：
1. [ ] `device.go`: HandleTelemetry 增加 `case "med_status":` 分支处理
2. [ ] `migrate.sql`: med_status_records 表添加 `compartment` 和 `device_id` 列
3. [ ] `router/router.go`: Ensure device binding via `elderly_devices` table exists
4. [ ] `emergency_workflow.go`: MedMissed alert triggers when status.taken=false

---

## 四、失败标准（Release Blockers）

下列任一情况未解决，系统不得进入生产环境：

| 编号 | 缺陷描述 | 严重程度 | 状态 |
|------|---------|---------|------|
| BUG-001 | UpdateDeviceConfig SQL注入未修复 | Critical | 待修复 |
| BUG-002 | B2B服务硬编码密码未从代码移除 | Critical | 待修复 |
| BUG-003 | Admin设备接口无任何RBAC保护 | Critical | 待修复 |
| BUG-004 | Refresh Token可被重放利用 | High | 待修复 |
| BUG-005 | 用药记录可在任意时间点标记（无窗口校验） | Medium | 待修复 |
| BUG-006 | /health端点不检查下游依赖 | Medium | 待修复 |

---

## 五、自动化测试建议

引入CI/CD流水线自动运行测试：

```yaml
# .github/workflows/ci.yml (示例)
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test_db
        ports: ["5432:5432"]
        options: >-
          --health-cmd pgisalive
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install dependencies
        run: go mod download

      - name: Run Go tests
        run: go test ./... -coverprofile=coverage.txt

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt
          flags: unittests

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run static analysis (golangci-lint)
        uses: golangci/golangci-lint-action@v3
      - name: Run sensitive-data detection
        uses: detect-secrets/action@v1
```

---

**第七阶段测试计划文档结束。** 建议优先修复BUG-001至BUG-003后，再推进TV-04集成测试的实施。