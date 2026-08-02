# 🧪 Eregen 全面测试体系建设计划

> **作者：** Agnes-2.0-Flash (Sapiens AI)
> **日期：** 2026-07-28
> **阶段：** Phase 6 — 验收测试与质量保证
> **目标：** 构建多层级测试体系，确保核心功能闭环、安全性、健壮性满足生产级标准

---

## 一、全局架构设计

```
测试金字塔（Test Pyramid）
                          ▲
                         UI Tests         ← 端到端UI测试（Flutter/Web）
                        /                 \
           Integration Tests              API Contract Tests
            /        |      \               (Pact/Swagger校验)
   Go Unit Tests  e2e Health   e2e Alert Flow
   (Coverage >80%) Checks      Chain Verification
   │            │               │
   └───┬────────┴─────┬────────┘
       Build Pipeline  Security Scan
       (CI/CD)          (Static Analysis)
```

**三层测试策略：**
1. **Unit Layer (60%)**：Go单元测试（服务层、存储层、工具包），覆盖核心业务逻辑
2. **Integration Layer (30%)**：API集成测试（健康检查、认证、设备上报、告警触发）、SQL schema验证
3. **End-to-End Layer (10%)**：完整告警链路测试（SOS从设备到家属通知）、用药全流程测试

---

## 二、任务分解与实施清单

### Task 1: Go单元/集成测试框架标准化（基准搭建）

**文件范围：** `cloud/api-server/`、`b2b/`、`shared/`、`cloud/admin-api/`

#### 1.1 建立统一测试基础设施

**File:** `cloud/api-server/test/test_setup.go`

```go
package testsetup

import (
    "context"
    "database/sql"
    "testing"
    "go.uber.org/zap"

    "eregen.dev/api-server/internal/model"
    "eregen.dev/api-server/internal/store"
    "github.com/mattn/go-sqlite3"
)

// NewTestDB creates an in-memory SQLite database for testing.
func NewTestDB(t *testing.Context, log *zap.Logger) (*store.Postgres, func()) {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations/schema
    if _, err := db.Exec(createUsersTable); err != nil {
        t.Fatal(err)
    }
    // ... other table creation

    pg := store.NewPostgres(db, log)
    return pg, func() { db.Close() }
}

const createUsersTable = `
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE,
    phone TEXT,
    open_id TEXT,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`
```

#### 1.2 创建共享测试工具包

**File:** `cloud/api-server/internal/testing/assert.go`

```go
package testing

import (
    "reflect"
    "testing"
)

func TRequireEqual(t *testing.T, expected, actual interface{}) {
    if !reflect.DeepEqual(expected, actual) {
        t.Fatalf("Expected %v but got %v", expected, actual)
    }
}

func TRequireError(t *testing.T, err error, msg string) {
    if err == nil {
        t.Fatalf("Expected error %v but got none", msg)
    }
}

func TRequireNoError(t *testing.T, err error) {
    if err != nil {
        t.Fatalf("Expected no error but got %v", err)
    }
}
```

**Files to modify:** Every existing *_test.go file should import these tools instead of writing custom assertions.

#### 1.3 覆盖率门禁配置

**File:** `.github/workflows/ci.yml` (add coverage gate)

```yaml
- name: Test with coverage
  run: |
    cd cloud/api-server
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    # Fail if coverage < 80% for core packages
    go test ./... -covermode=atomic -coverpkg=./internal/... -cover=profile:profile.cov
    grep -q 'overall: [8-9]\.[0-9]*%' profile.cov || (echo "Coverage threshold not met!"; exit 1)
```

---

### Task 2: Core Unit Test Coverage补全（按模块优先级）

#### 2.1 加密模块（最高优先级——修复SHA-256问题前需验证）

**Files to test：** `shared/crypto/*.go`

| Test Case | Description | Severity |
|-----------|-------------|----------|
| TestBcryptHashPassword | Verify bcrypt correctly hashes with cost 10 | Critical |
| TestSHA256HashPassword | Document known limitation (will be replaced) | Medium |
| TestVerifyPasswordCorrect | Verify correct password returns true | High |
| TestVerifyPasswordWrong | Verify wrong password returns false | High |
| TestGenerateSaltEdgeCases | Edge cases for salt generation | Medium |
| TestSignVerifyMessage | Ed25519 signature round-trip | High |
| TestDeriveSessionKeyStrength | Current implementation weak → mark as TODO | Critical (to fix) |

**Action:** Add new test file `shared/crypto/crypto_test_bcrypt.go` after fixing HashPassword to use bcrypt.

#### 2.2 认证模块（高优先级——涉及安全漏洞）

**File:** `cloud/api-server/internal/middleware/auth_test.go`

```go
func TestJWTAuth_TokenGeneration(t *testing.T) {
    auth := NewJWTAuth([]byte("secret"), 3600, 7200, mockLog, mockStore)
    token, err := auth.GenerateToken("user123", "family")
    RequireError(t, err)
    Require(t, token.Valid, "Token should be valid")
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
    // Create expired token and verify rejection
}

func TestRefreshToken_ReplayProtection(t *testing.T) {
    // Issue refresh token, use it once, verify second use fails
}
```

#### 2.3 设备处理模块

**File:** `cloud/api-server/internal/handler/device_test.go`

```go
func TestDeviceHandler_ListOwnerOnly(t *testing.T) {
    // Setup device owned by userA
    // As userB calling LIST, expect only userB's devices returned
}

func TestDeviceHandler_Get_UnauthorizedAccess(t *testing.T) {
    // GET /devices/{deviceID} where device belongs to another user
    // Expect 403 Forbidden (NOT_FOUND should leak nothing)
}

func TestDeviceHandler_Telemetry_ElderlyOwnershipCheck(t *testing.T) {
    // Simulate telemetry from deviceX claiming elderlyID_Y that doesn't match owner
    // Should reject or flag as security warning
}
```

#### 2.4 用药服务模块

**File:** `cloud/api-server/internal/service/medication_service_test.go`

```go
func TestMedicationService_CreateRule_NotifiesDevice(t *testing.T) {
    // Mock NATS client
    // Verify PublishCommand called with correct topic and payload
}

func TestMedicationService_TodayStatus_Completeness(t *testing.T) {
    // Create rules + some MedStatusRecords
    // Verify TodayStatus returns all rules with taken=true/false correctly
}
```

---

### Task 3: 端到端告警链路完整测试（e2e alert chain）

这是最关键的集成测试，验证整个SOS流程从硬件模拟到多通道通知。

**File:** `cloud/api-server/cmd/e2e_alert_chain_test.go`

```go
package main

import (
    "context"
    "time"
    "testing"

    "eregen.dev/api-server/internal/model"
    "eregen.dev/api-server/internal/service"
)

func TestCompleteSOSEndToEnd(t *testing.T) {
    // 1. Start test environment (mock NATS, SQLite, Redis, Push providers)
    stopSetup := SetupTestEnvironment()
    defer stopSetup()

    ctx := context.Background()

    // 2. Register a test elderly + family user
    elderlyID := CreateTestElderly(ctx)
    familyUserID := CreateTestFamilyUser(ctx, elderlyID)

    // 3. Bind a bracelet device to this elderly
    deviceID := "BR-TEST001"
    BindDeviceToDevice(elderlyID, deviceID)

    // 4. Simulate SOS event from device (publish directly to NATS simulated)
    sosEvent := &service.DeviceEvent{
        Type:      "sos",
        DevID:     deviceID,
        ElderlyID: elderlyID,
        Lat:       31.2304,
        Lon:       121.4737,
    }
    PublishEventToNATS(sosEvent)

    // 5. Wait for alert to be processed and stored
    alert := WaitForAlertCreated(ctx, elderlyID, time.Second*5)
    RequireEqual(t, model.AlertP0, alert.Severity)
    RequireEqual(t, "sos", alert.AlertType)

    // 6. Verify FCM push was triggered
    fcmNotifications := GetSentFCMNotifications()
    RequireGreaterThan(t, len(fcmNotifications), 0, "At least one FCM should be sent")

    // 7. Verify WebSocket broadcast to family clients
    wsMessages := GetWebSocketBroadcastsForElderly(elderlyID)
    RequireNotEmpty(t, wsMessages, "WS should broadcast to connected family clients")

    // 8. SMS fallback should be logged as attempted if config has tokenStore
    // (Currently missing, this test would document the gap)
}
```

**Required setup helpers:**
- In-memory NATS mock that accepts messages and exposes them to test
- Fake FCM provider that records sent messages instead of hitting real API
- WebSocket hub mock that tracks subscribers per ElderlyID

---

### Task 4: Medication闭环端到端测试

**File:** `cloud/api-server/internal/service/medication_e2e_test.go`

```go
func TestMedicationWorkflow_CreateAndPushRule(t *testing.T) {
    // Setup elderly and device
    elderlyID := CreateElderly()
    deviceID := "PX-TEST001"  // pillbox
    BindDeviceToDevice(elderlyID, deviceID)

    // Create medication rule via API
    rule := CreateMedicationRule(elderlyID, "08:00", 1, "capsule", []int{1,2,3,4,5,6,7})

    // Verify rule stored in DB
    stored := GetRuleFromDB(rule.ID)
    RequireEqual(t, elderlyID, stored.ElderlyID)

    // Verify NATS command published to device topic
    natssub := SubscribeNATSCommand(deviceID)
    cmd := <-natssub
    RequireEqual(t, "med_rule", cmd.Type)
    RequireContains(cmd.Text, rule.ID)
}

func TestMediation_FollowUpMissedAlertNotGenerated(t *testing.T) {
    // This tests the GAP: currently missed alerts are never auto-generated
    // After implementing med_missed service, this test will pass
    elderlyID := CreateElderlyWithMedRule()

    // Simulate: time passes, pillbox should send med_status(taken=false)
    // but no event is generated -> no alert should exist yet
    alerts := GetAlertsByType(elderlyID, "med_missed")
    RequireZero(t, alerts, "Missing automated med_missed detection")
}
```

---

### Task 5: B2B医护工作流测试用例

涵盖护士扫描 → 患者信息查询 → 核验记录闭环

#### 5.1 护士终端BLE扫描模拟测试（模拟层）

**File:** `apps/nurse_terminal/test/ble_scan_simulation_test.dart`

```dart
// Flutter integration test for nurse terminal BLE scan
group('Nurse Terminal BLE', {
  test('scanning reads patient data correctly', () async {
    // Mock BLE service returning encrypted patient data
    final mockBle = MockBleService();
    when(mockBle.readPatientData(any)).thenReturn(PatientInfo(
      id: 'pat-123',
      name: '张三',
      admissionNo: '20260721001',
      allergies: ['penicillin'],
    ));

    final navigator = await tester.pumpWidget(const NurseTerminalApp());

    // Scan wristband button press
    await tap(find.text('扫描腕带'));
    await pumpAndSettle();

    // Patient detail should display
    expect(find.text('张三'), findsOneWidget);
    expect(find.text('青霉素过敏'), findsOneWidget);
  });
});
```

#### 5.2 API集成测试：医疗数据访问权限

**File:** `cloud/admin-api/internal/hospital/integration_test.go`

```go
func TestMedicalWristband_AccessControl(t *testing.T) {
    tests := []struct {
        name       string
        authRole   string  // family, nurse, hospital_admin, regulator
        accessCode int     // expected HTTP code
    }{
        {"nurse_view_own_patients", "nurse", http.StatusOK},
        {"nurse_view_others_department", "nurse", http.StatusForbidden},
        {"hospital_admin_view_all", "hospital_admin", http.StatusOK},
        {"family_view_unrelated_patient", "family", http.StatusForbidden},
        {"regulator_view_all", "regulator", http.StatusOK},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup: create patient in department A
            deptA := CreateDepartment("Cardiology")
            patient := CreatePatientInDept(deptA, "Wang")

            // Authenticate as user with specified role
            token := GenerateToken(tt.authRole, patient.DepartmentID != tt.accessCode)

            // Call GET /api/v1/medical/patients/:patientID
            resp := MakeAuthorizedRequest("GET", "/api/v1/medical/patients/"+patient.ID, token)

            RequireCode(t, tt.accessCode, resp)
        })
    }
}
```

---

### Task 6: 安全专项测试（Security Tests）

#### 6.1 渗透测试用例集合（自动化）

**File:** `security/penetration_tests.go`

```go
// TestAdminEndpointPrivilegeEscalation — try admin endpoints as normal user
func TestAdminEndpoints_Unauthorized(t *testing.T) {
    // Normal user token
    token := GenerateToken("family", "some-elderly-id")

    // Try admin device list
    resp := RequestWithToken("GET", "/api/v1/admin/devices", token)
    RequireEqual(t, http.StatusForbidden, resp.Code)

    // Try role change endpoint
    reqBody := `{"role":"admin"}`
    resp = RequestWithToken("PUT", "/api/v1/admin/users/target-user/role", reqBody, token)
    RequireEqual(t, http.StatusForbidden, resp.Code)
}

// TestDeviceIdTraversal — enumerate all device IDs
func TestDevice_IdTraversal(t *testing.T) {
    token := GenerateToken("family", "my-elderly")
    // Try accessing device that doesn't belong to user
    for _, devID := range []string{"BR-0001", "BR-0002", "BR-0003"} {
        resp := RequestWithToken("GET", "/api/v1/devices/"+devID, token)
        // All should be 403 or 404 (not revealing existence)
        RequireInRange(t, resp.Code, 400, 404)
    }
}

// TestCredentialHardcodingScan — check source files for hardcoded secrets
func Test_HardcodedSecretsScan(t *testing.T) {
    files := GetAllGoFiles()
    for _, f := range files {
        if Contains(f, "dev-secret-key-change-in-production") {
            t.Errorf("Hardcoded secret found in %s", f.Path())
        }
        if Contains(f, `"eregen_password"`) {
            t.Errorf("Hardcoded password found in %s", f.Path())
        }
    }
}
```

#### 6.2 依赖扫描集成

Add to CI: `trivy application . --severity HIGH,CRITICAL` OR `golangci-dash` for Go deps vulnerability scanning.

---

### Task 7: 性能与稳定性测试

#### 7.1 告警突发压力测试

**File:** `cloud/api-server/bench/alert_benchmark.go`

```go
func Benchmark_SOS_Burst(t *testing.B) {
    // Simulate 500 concurrent SOS events from different devices
    for i := 0; i < t.N; i++ {
        go func(idx int) {
            event := DeviceEvent{Type:"sos", ElderlyID:ElderlyID(idx), Lat:randFloat(), Lon:randFloat()}
            routeEvent(event) // simulate NATS delivery
        }(i)
    }
}

// Expected: Should handle 500+ concurrent SOS without dropping
// Alert count in DB should equal number of events
```

#### 7.2 连接池压力测试

测试SQLite并发写入（医疗录入、验记等高频写操作），确认MVP阶段单用户可行、多用户有锁竞争问题——这需要在文档中明确说明并制定生产迁移至PG的策略。

---

### Task 8: 家庭App & 管理后台UI测试（Flutter/Vue）

#### 8.1 Flutter Widget Tests（家属APP）

关键页面测试：

```dart
// apps/family-app/test/widget_test.dart
void main() {
  testWidgets('Home screen shows correct location', (WidgetTester tester) async {
    // Mock API returning location data
    when(() => locator.getLocation()).thenAnswer(_ => const LocationRecord(lat: 31.2, lon: 121.4));

    await tester.pumpWidget(const FamilyApp());

    expect(find.text('实时定位'), findsOneWidget);
    expect(find.byType(GMapWidget), findsOneWidget);
  });

  testWidgets('SOS alert card displays correctly', (WidgetTester tester) async {
    // Simulate incoming alert via WebSocket mock
    await tester.startStream(alertStream); // Stream of Alert objects
    await tester.pumpAndSettle();

    expect(find.text('紧急告警'), findsOneWidget);
    expect(find.icon(alertIcon), findsOneWidget);
  });
}
```

#### 8.2 Vue Component Tests（管理后台）

关键组件测试：

```javascript
// apps/admin-web/tests/unit/AlertList.spec.js
import { mount } from '@vue/test-utils'
import AlertList from '@/components/AlertList.vue'

describe('AlertList.vue', () => {
  it('renders P0 alerts with red badge', async () => {
    const wrapper = mount(AlertList, {
      propsData: {
        alerts: [{ severity: 'P0', type: 'sos', status: 'pending' }]
      }
    })
    expect(wrapper.find('.badge-p0').text()).toBe('P0')
    expect(wrapper.find('.badge-p0').classes()).toContain('text-red')
  })

  it('allows handling alerts with correct permissions', async () => {
    const wrapper = mount(AlertList, {
      propsData: { userRole: 'nurse' }
    })
    expect(wrapper.find('[data-testid="handle-btn"]').exists()).toBe(true)
  })
})
```

---

### Task 9: 固件侧单元测试与仿真测试

为FreeRTOS和ESP-IDF编写可编译的单元测试（使用Unity/Ceedling框架）：

#### 9.1 Entry版SOS长按检测

**File:** `firmware/bracelet/entry/test_sos_button.c`

```c
TEST_TEST(SOS_Button_LongPress_DetectsAfter3Seconds) {
    // Simulate button pressed 2.9s → should NOT trigger
    sos_button_simulate_press(2900);
    TEST_ASSERT_FALSE(sos_button_check_trigger());

    // Simulate button pressed 3.1s → SHOULD trigger
    sos_button_simulate_press(3100);
    TEST_TRUE(sos_button_check_trigger());
}

TEST_TEST(SOS_Button_ResetOnRelease) {
    sos_button_simulate_press(2000);
    sos_button_simulate_release();
    TEST_FALSE(sos_button_check_trigger());
}
```

#### 9.2 Plus版跌倒检测算法stub测试

**File:** `firmware/bracelet/plus/test_fall_detect.c`

```c
TEST_TEST(Fall_Detection_ImpactThreshold) {
    IMU_Data acc = { .x = 15.2, .y = 14.8, .z = -2.1 }; // high impact
    FALL_Result result = fall_detect_check(&acc);
    TEST_ASSERT_EQUAL_INT(FALL_DETECTED, result.status);
}

TEST_Test(Fall_Detection_Not_Fall_LowImpact) {
    IMU_Data acc = { .x = 0.2, .y = 0.1, .z = 9.8 }; // just gravity
    FALL_Result result = fall_detect_check(&acc);
    TEST_ASSERT_EQUAL_INT(NO_FALL, result.status);
}
```

---

### Task 10: CI/CD 流水线集成与门禁配置

**.github/workflows/full_test_pipeline.yml**

```yaml
name: Full Test Pipeline

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: Run Go tests with coverage
        run: |
          cd cloud/api-server
          go test ./... -cover
          cd shared && go test ./...
          cd b2b && go test ./...
  
  security-scan:
    needs: unit-tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Trivy scan
        uses: aquasecurity/trivy-action@master
        with:
          severity: HIGH,CRITICAL
          format: table
  
  flutter-widget-tests:
    needs: unit-tests
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Flutter
        uses: subdigital/flutter-action@v2
      - name: Run widget tests
        run: flutter test --coverage --test-randomize-ordering-seed random
  
  vue-component-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install and run tests
        run: |
          cd apps/admin-web
          npm ci
          npm run test:unit

  end-to-end-alert-chain:
    needs: [unit-tests, security-scan]
    runs-on: ubuntu-latest
    services:
      nats:
        image: nats:2.9
      redis:
        image: redis:7
      sqlite: # in-process
        run: echo "using in-memory DB"
    steps:
      - uses: actions/checkout@v3
      - name: Run e2e alert test
        run: |
          cd cloud/api-server
          go test -run TestCompleteSOSEndToEnd -v
```

---

## 三、缺陷追踪与测试报告集成

在测试系统中实现自动缺陷跟踪：

```go
// Track test results linked to code review findings
type TestCase struct {
    ID          string
    Description string
    File        string
    Line        int
    Severity    SeverityLevel // CRITICAL, HIGH, MEDIUM, LOW
    Status      Status        // PASSED, FAILED, SKIPPED, BLOCKED
    BugID       string        // links to review finding S01, F03, etc.
}

func ReportTestCase(tc TestCase) {
    // Write to JSONL output for dashboard consumption
    // If BUG != "" and status == FAILED, auto-create GitHub Issue
}
```

**输出格式：** `test-results/YYYY-MM-DD-suite.jsonl` 每行一条测试记录，用于生成可视化报表。

---

## 四、执行路线图（分阶段交付）

| 周期 | 任务 | 产出物 | 验收标准 |
|------|------|--------|---------|
| Week 1 | Task 1: 测试框架标准化 | 统一assert工具、模板测试文件、覆盖率脚本 | 所有*_test.go文件通过重构使用新工具 |
| Week 2 | Task 2: 核心单元测补全 | shared/crypto, middleware, handler单元测试覆盖 >70% | `go test ./... -coverpkg=./internal/...` ≥70% |
| Week 3 | Task 3: e2e告警链路测试 | complete_e2e_alert_test.go可用，mock环境可跑通 | SOS事件→存储→FCM→WS全流程通过 |
| Week 4 | Task 4: 用药闭环测试 | medication_e2e_test.go规则下发验证通过 | NATS命令正确推送，设备模拟接收 |
| Week 5 | Task 5: B2B医护流测试 | access_control_test + ble_simulation测试通过 | 角色隔离生效，BLE数据解析正确 |
| Week 6 | Task 6: 安全测试 | penetration_tests.go运行，无硬编码发现 | 无hardcoded secret告警，遍历测试拦截 |
| Week 7 | Task 7: 性能压力测试 | benchmark结果报告，瓶颈分析文档 | 500并发QPS下告警延迟<2s |
| Week 8 | Task 8: UI Flutter/Vue测试 | widget/component测试通过率90%+ | 核心页面组件测试全部通过 |
| Week 9 | Task 9: 固件单元测试 | C单元测试用例编译通过，模拟验证 | SOS触发阈值正确，跌倒判断合理 |
| Week 10 | Task 10: CI流水线整合 | .github/workflows/full_test_pipeline.yml成功运行 | PR提交自动触发全套测试 |

---

## 五、资源需求

| 类别 | 需求 | 备注 |
|------|------|------|
| 硬件 | iOS真机x1, Android真机x1, ESP32开发板x2, GD32E230开发板x2 | 用于真实固件联调和BLE测试 |
| 软件许可证 | Xcode (macOS), Android Studio, ESP-IDF, ARM GNU Toolchain | 开发环境搭建 |
| 云服务 | GitHub Actions Runner（自建macOS runner供iOS测试） | 或使用GitHub托管Runner |
| 时间 | 10周（全职2人并行） | 核心Go测试 + Flutter/Vue并行 |

---

## 六、风险管理与应对

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| ESP32/FreeRTOS测试环境搭建困难 | 中 | 高 | 先使用单元测试框架在主机上模拟算法逻辑，固件侧只做边界测试 |
| Flutter真机测试资源不足 | 低 | 中 | 先做Web模拟器测试，真机逐步补充 |
| SQLite并发性能导致测试不稳定 | 中 | 中 | 测试使用in-memory模式避免文件锁竞争 |
| 现有代码无注释导致e2e难以理解 | 高 | 高 | 配合审查报告逐段阅读注释，关键函数添加TODO标记 |

---

## 七、交付物清单

最终将生成以下文档和可执行资产：

1. `tests/API的合同测试集（OpenAPI规范校验）`
2. `cloud/api-server/test/` — Go集成测试工具包
3. `security/penetration_tests.go` — 自动化安全检查用例
4. `apps/family_app/test/` — Flutter widget测试
5. `apps/admin-web/tests/unit/` — Vue组件测试
6. `firmware/bracelet/entry/test/` — FreeRTOS单元测试
7. `.github/workflows/full_test_pipeline.yml` — CI门禁配置
8. `docs/TESTING_STRATEGY.md` — 测试方法论文档
9. `test-results/dashboard.html` — 覆盖率可视化报表

---

> **执行选择：** 推荐使用 **Superpowers Subagent-Driven Development**（每个测试模块派生子代理并行执行）以加速进度；若当前会话直接执行也可采用Inline Execution顺序完成。建议对高复杂度的e2e告警链和固件测试优先使用子代理。

计划已保存至 `docs/superpowers/plans/2026-07-28-comprehensive-test-suite.md`。请选择执行方式继续。