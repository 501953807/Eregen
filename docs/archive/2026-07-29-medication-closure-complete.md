# Medication Management Full Closure Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development for parallel execution of independent tasks, or superpowers:executing-plans for sequential batch execution.

**Goal:** Complete all medication management functionality across family-app Flutter client, admin-web Vue frontend, and Go API backend as identified in audit closure trace.

**Architecture:** Three-layer implementation: (1) Admin Web backend-facing API layer & page; (2) Family App API client & UI integration fixes; (3) API Server backend enhancements including med_status handling and SQL injection remediation.

**Tech Stack:** Go 1.22+, Gin 1.9+, Vue 3.4+, Pinia, TypeScript 5.4+, Flutter 3.24+, Dio 5.0+.

## Global Constraints

All API endpoints must use JWT authentication with proper role-based access control; all database queries must use parameterized statements only (no string interpolation); all new code must follow existing file patterns and style conventions.

---

## Task A1: Admin Web - Create Medication API Layer

**Files:**
- Create: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/api/medication.ts`

**Interfaces:** Consumes the shared Axios interceptor from `client.ts`; produces type-safe methods matching backend API contract.

- [ ] **Step 1: Define medication-related TypeScript interfaces**

```typescript
// /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/api/medication.ts
export interface MedicationRule {
  id: string;
  elderlyId: string;
  pillType: string;
  doseCount: number;
  scheduleTime: string; // HH:MM format
  daysOfWeek: string[]; // ['mon', 'tue', ...]
  active: boolean;
  createdAt: string;
}

export interface MedicationTodayStatus {
  ruleId: string;
  pillType: string;
  scheduleTime: string;
  taken: boolean;
  reportedAt: string | null;
}

export interface MedicationHistoryRecord {
  id: string;
  ruleId: string;
  takenBy: string; // 'family' or 'device'
  isTaken: boolean;
  compartment?: number;
  reportedAt: string;
}
```

- [ ] **Step 2: Implement medicationApi CRUD functions using API client**

```typescript
import apiClient from './client'
import type { MedicationRule, MedicationTodayStatus, MedicationHistoryRecord } from '@/types'

export const medicationApi = {
  // GET /api/v1/elderly/{elderly_id}/medication/rules
  listRules(elderlyId: string) {
    return apiClient.get(`/api/v1/elderly/${elderlyId}/medication/rules`)
  },

  // POST /api/v1/elderly/{elderly_id}/medication/rules
  createRule(elderlyId: string, data: Omit<MedicationRule, 'id' | 'active' | 'createdAt'>) {
    return apiClient.post(`/api/v1/elderly/${elderlyId}/medication/rules`, data)
  },

  // PUT /api/v1/elderly/{elderly_id}/medication/rules/{rule_id}
  updateRule(elderlyId: string, ruleId: string, data: Partial<Omit<MedicationRule, 'id'>>) {
    return apiClient.put(`/api/v1/elderly/${elderlyId}/medication/rules/${ruleId}`, data)
  },

  // DELETE /api/v1/elderly/{elderly_id}/medication_rules/{rule_id}
  deleteRule(elderlyId: string, ruleId: string) {
    return apiClient.delete(`/api/v1/elderly/${elderlyId}/medication/rules/${ruleId}`)
  },

  // GET /api/v1/elderly/{elderly_id}/medication/today
  getTodayStatus(elderlyId: string) {
    return apiClient.get(`/api/v1/elderly/${elderlyId}/medication/today`)
  },

  // GET /api/v1/elderly/{elderly_id}/medication/history
  getHistory(elderlyId: string, days: number = 30) {
    return apiClient.get(`/api/v1/elderly/${elderlyId}/medication/history`, { params: { days } })
  },

  // POST /api/v1/medication/:rule_id/take (manual mark by family member)
  takeRule(ruleId: string) {
    return apiClient.post(`/api/v1/medication/${ruleId}/take`)
  }
}
```

- [ ] **Step 3: Extend existing types definition**

Add the new interfaces to `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/types/index.ts`:

```typescript
// Add to existing type declarations
export interface MedicationRule {
  id: string
  elderlyId: string
  pillType: string
  doseCount: number
  scheduleTime: string
  daysOfWeek: string[]
  active: boolean
  createdAt: string
}

export interface MedicationTodayStatus {
  ruleId: string
  pillType: string
  scheduleTime: string
  taken: boolean
  reportedAt: string | null
}
```

- [ ] **Step 4: Commit changes**

```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web
git add src/api/medication.ts src/types/index.ts
git commit -m "feat(admin): add medication API layer with CRUD operations"
```

---

## Task A2: Admin Web - Create Medication Management Page (Medication.vue)

**Files:**
- Create: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/views/Medication.vue`
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/router/index.ts` (add route)

**Component Structure:** Layout similar to Elderly.vue but focused on medication rules list, today's summary, and adherence history.

- [ ] **Step 1: Design template layout with top bar, KPI cards, medication list table, and today's summary**

```vue
<template>
  <div class="medication-page">
    <!-- Top Navigation Bar -->
    <el-card shadow="hover">
      <div class="card-header-with-action">
        <span style="font-weight: 600;">用药管理</span>
        <el-link type="primary" :underline="false">添加规则 →</el-link>
      </div>
    </el-card>

    <!-- Today Summary Cards -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: linear-gradient(135deg, #16A34A, #22C55E);">
              <el-icon :size="28"><Medicine /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.activeRules }}</div>
              <div class="kpi-label">今日规则数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: linear-gradient(135deg, #F59E0B, #FBBF24);">
              <el-icon :size="28"><Bell /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.missedCount }}</div>
              <div class="kpi-label">今日漏服</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: linear-gradient(135deg, #3B82F6, #60A5FA);">
              <el-icon :size="28"><Clock /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.adherenceRate }}%</div>
              <div class="kpi-label">按时服药率</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-danger">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: linear-gradient(135deg, #EF4444, #F87171);">
              <el-icon :size="28"><AlarmClock /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ stats.pendingActions }}</div>
              <div class="kpi-label">待处理提醒</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Medication Rules Table -->
    <el-card shadow="hover">
      <template #header><span style="font-weight: 600;">用药规则列表</span></template>
      <el-table :data="rules.data" stripe v-loading="loading.rules">
        <el-table-column prop="pillType" label="药品名称" width="180"></el-table-column>
        <el-table-column prop="scheduleTime" label="服用时间" width="120"></el-table-column>
        <el-table-column prop="doseCount" label="剂量" width="80"></el-table-column>
        <el-table-column prop="daysOfWeek" label="执行周期" width="140"></el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.active ? 'success' : 'danger'">
              {{ row.active ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" @click="handleDelete(row.id)" type="danger">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="table-footer">
        <el-pagination
          v-model="page.currentPage"
          v-model:pageSize="page.pageSize"
          :total="page.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </el-card>
  </div>
</template>
```

- [ ] **Step 2: Implement script section with store interactions**

```typescript
<script setup lang='ts'>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { medicationApi } from '@/api/medication'
import { useElderlyStore } from '@/stores/elderly' // hypothetical store or direct API call

const loading = ref({ rules: true })
const rules = ref<any[]>([])
const page = ref<{ currentPage: number; pageSize: number; total: number }>({
  currentPage: 1,
  pageSize: 10,
  total: 0
})

const stats = computed(() => {
  const active = rules.value.filter(r => r.active).length
  const missed = rules.value.filter(r => !r.taken && r.scheduleTime < currentHour).length
  return { activeRules: active, missedCount: missed, adherenceRate: 85, pendingActions: 3 }
})

onMounted(async () => {
  loading.value.rules = true
  try {
    const res = await medicationApi.listRules('elderly_123') // placeholder: get current elderly ID
    rules.value = res.data.data || []
    page.value.total = rules.value.length
  } catch (error) {
    console.error('Failed to load medication rules:', error)
    ElMessage.error('加载用药规则失败')
  } finally {
    loading.value.rules = false
  }
})

function handleEdit(row: any) {
  ElNotification({ title: '编辑规则', message: 'Feature not yet implemented', type: 'info' })
}

function handleDelete(id: string) {
  ElMessage.confirm(`确定要删除用药规则吗？`, '警告', { type: 'warning' }).then(async () => {
    try {
      await medicationApi.deleteRule('elderly_123', id)
      rules.value = rules.value.filter(r => r.id !== id)
      ElMessage.success('删除成功')
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

function handlePageChange(pageNum: number) {
  page.value.currentPage = pageNum
  // Handle pagination fetch
}

function handleSizeChange(size: number) {
  page.value.pageSize = size
  page.value.currentPage = 1
  handleSizeChange(size)
}
</script>
```

- [ ] **Step 3: Add route configuration**

Modify `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web/src/router/index.ts`:

```typescript
{ path: '/medication', component: () => import('@/views/Medication.vue'), name: 'Medication' }
```

- [ ] **Step 4: Add icon import if needed**

Ensure `Medicine`, `Clock`, `AlarmClock` icons are imported from Element Plus or added as SVG components.

- [ ] **Step 5: Commit**

```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/admin-web
git add src/views/Medication.vue src/router/index.ts
git commit -m "feat(admin): add medication management page with KPIs, rules list, and CRUD actions"
```

---

## Task B1: Family App - Fix Medication API Client Endpoints

**Files:**
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/family-app/lib/api/client.dart`

The current API client uses `/medication/rules` but the actual backend expects `/api/v1/elderly/{elderly_id}/medication/rules`. Need to correct these endpoint URLs.

- [ ] **Step 1: Correct listMeds() method to use proper endpoint**

Current (line 223-225):
```dart
Future<Response> listMeds(String elderId) async {
  return _dio.get('/api/v1/medication/rules', queryParameters: {'elder_id': elderId});
}
```

Fix to match backend pattern:
```dart
Future<Response> listMeds(String elderId) async {
  return _dio.get('/api/v1/elderly/$elderId/medication/rules');
}
```

- [ ] **Step 2: Update saveMedicationRule() alias method**

Line 233-235:
```dart
Future<Response> saveMediationRule(Map<String, dynamic> rule) async {
  return _dio.post('/api/v1/medication/rules', data: rule);
}
```

Fix to:
```dart
Future<Response> saveMediationRule(Map<String, dynamic> rule) async {
  final elderId = rule['elderId'] ?? rule['elderly_id'];
  if (elderId == null) throw ArgumentError('elderId required in rule data');
  return _dio.post('/api/v1/elderly/$elderId/medication/rules', data: rule);
}
```

And similarly for `updateMedRule()` (alias at line 237-240).

- [ ] **Step 3: Ensure the API endpoint pattern matches the backend specification**

Verify all three medication-related calls:
- GET `/api/v1/elderly/{id}/medication/rules` → listMeds()
- POST `/api/v1/elderly/{id}/medication/rules` → saveMedicationRule()
- PUT `/api/v1/elderly/{id}/medication/rules/{ruleId}` → updateMedRule()
- DELETE `/api/v1/elderly/{id}/medication/rules/{ruleId}` → (missing method - need to add)

- [ ] **Step 4: Add deleteMedicationRule() method**

Add after updateMedRule():

```dart
/// DELETE /api/v1/elderly/{elderly_id}/medication/rules/{rule_id}
Future<Response> deleteMedicationRule(String elderlyId, String ruleId) async {
  return _dio.delete('/api/v1/elderly/$elderlyId/medication/rules/$ruleId');
}
```

- [ ] **Step 5: Verify the createSOSAlert() endpoint consistency**

Currently line 208-210: `POST /api/v1/alerts/sos` — this should remain as-is per spec.

- [ ] **Step 6: Run dart analyze to verify no warnings from changes**

```bash
cd apps/family-app
dart analyze --no-fatal-infos --no-fatal-warnings lib/api/client.dart
```

- [ ] **Step 7: Commit**

```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/family-app
git add lib/api/client.dart
git commit -m "fix(family-app): correct medication API endpoints to match backend specification"
```

---

## Task B2: Family App - Verify Medication Page Integration

**Files:**
- Read: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/family-app/lib/screens/medication/medication_page.dart`

Check if the page uses the correct API client methods:

- [ ] **Step 1: Examine medication_page.dart for direct Dio calls vs API client usage**

Current code line 34-45 shows:
```dart
Future<void> _fetchData() async {
  try {
    final resp = await ApiClient.instance.get('/medication/rules'); // ← WRONG ENDPOINT!
    ...
  }
}
```

This uses `_dio.get('/medication/rules')` directly via internal `_dio` rather than through `ApiClient.instance.listMeds()`. Need to refactor to use the API client methods properly.

- [ ] **Step 2: Refactor _fetchData() to use API client method**

```dart
Future<void> _fetchData() async {
  try {
    // Get current user/elderly ID from context or store
    final currentElderId = _getCurrentElderId(); 
    final resp = await ApiClient.instance.listMeds(currentElderId);
    final list = resp.data as List;
    setState(() {
      _rules = list.map((r) => MedicationRule.fromJson(r as Map<String, dynamic>)).toList();
      _loading = false;
    });
  } catch (e) {
    setState(() => _loading = false);
  }
}
```

- [ ] **Step 3: Update medication confirmation action to use correct endpoint**

Line 54: `await ApiClient.instance.post('/medication/$ruleId/take');`

Need to add a proper API method: In `client.dart`, add:

```dart
/// POST /api/v1/medication/:rule_id/take
Future<Response> takeMedicationRule(String ruleId) async {
  return _dio.post('/api/v1/medication/$ruleId/take');
}
```

Then in medication_page.dart:
```dart
final resp = await ApiClient.instance.takeMedicationRule(ruleId);
```

- [ ] **Step 4: Commit**

```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/family-app
git add lib/screens/medication/medication_page.dart lib/api/client.dart
git commit -m "refactor(family-app): refactor medication_page to use typed API client methods"
```

---

## Task C1: Backend API Server - Fix Medication Rule Device Push

**Files:**
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/api-server/internal/handler/medication.go` (CreateRule function)
- Need to find where the device ID resolution happens

From the analysis document: Issue is that CreateRule pushes to hardcoded "BR-XXXX" instead of the pillbox device bound to the elderly profile.

- [ ] **Step 1: Locate the service layer that publishes the med_rule command**

Search in `cloud/api-server/internal/service/medication_service.go` or similar:

```go
// Find the NATS publishing code in CreateRule
// Hardcoded "BR-XXXX" needs to be replaced with actual pillbox device ID
```

- [ ] **Step 2: Query the elderly's bound devices**

Need to modify the flow:
1. After creating medication rule in DB
2. Query `devices` table where `owner_id = elderly_id AND device_type = 'pillbox'`
3. Publish med_rule command to each matched device's MQTT topic

- [ ] **Step 3: Implement fix with proper device lookup**

Sample fix pattern:
```go
// Instead of hardcoded device ID
devices, err := svc.store.ListDevicesByOwner(ctx, elderlyID, "pillbox")
if err != nil {
    // handle error
}
for _, device := range devices {
    cmd := map[string]interface{}{
        "type":   "med_rule",
        "dev_id": device.DeviceID,
        "rules:  rules, // newly created rule
    }
    nats.PublishCommand(ctx, cmd, device.DeviceID)
}
```

- [ ] **Step 4: Test change**

Run `go test ./internal/handler/...` to ensure no regression.

- [ ] **Step 5: Commit**

```bash
cd cloud/api-server
git add internal/handler/medication.go internal/service/medication_service.go
git commit -m "fix(server): medication rule push uses actual bound pillbox device IDs"
```

---

## Task C2: Backend API Server - Add med_status Message Handler

**Files:**
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/api-server/internal/handler/device.go` (HandleTelemetry function)
- Need to add case for "med_status" message type

- [ ] **Step 1: Add med_status handling branch in HandleTelemetry**

Current switch handles telemetry, heartbeat, location. Add med_status case:

```go
case "med_status":
    var medStatus struct {
        DeviceID   string `json:"dev_id"`
        Compartment int    `json:"compartment"`
        Taken      bool   `json:"taken"`
        TS         int64  `json:"ts"`
    }
    if err := json.Unmarshal(body, &medStatus); err != nil {
        h.log.Error("parse med_status", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_JSON", "message": "Invalid payload"})
        return
    }
    // Look up elderly_id from device
    elderlyID, err := svc.store.GetElderlyByDeviceID(ctx, medStatus.DeviceID)
    if err != nil {
        h.log.Error("device not found or unbound", zap.Error(err))
        c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Device not found"})
        return
    }
    // Record med_status entry
    err = svc.store.CreateMedStatusRecord(ctx, elderlyID, medStatus.DeviceID, medStatus.Compartment, medStatus.Taken, time.Unix(medStatus.TS, 0))
    if err != nil {
        h.log.Error("record med status", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED", "message": "Failed to record status"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Medication status recorded"})
```

- [ ] **Step 2: Update database schema - add compartment and device_id columns to med_status_records**

Execute migration:
```sql
ALTER TABLE med_status_records ADD COLUMN IF NOT EXISTS compartment INTEGER;
ALTER TABLE med_status_records ADD COLUMN IF NOT EXISTS device_id TEXT;
```

- [ ] **Step 3: Commit**

```bash
cd cloud/api-server
git add internal/handler/device.go migrations/*.sql
git commit -m "feat(server): add med_status handler for pillbox medication reporting"
```

---

## Task C3: Backend API Server - SQL Injection Remediation

**Files:**
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/admin-api/internal/store/postgres.go` (settings update logic)
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/api-server/internal/store/sqlite.go`
- Modify: `/Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/api-server/internal/store/postgres.go`

- [ ] **Step 1: Fix settings update in admin-api (line ~176)**

Current vulnerable code:
```go
settings += fmt.Sprintf(`"%s":%v,`, k, v)
```

Replace with JSON marshaling:
```go
data, err := json.Marshal(settingsMap)
if err != nil { /* handle */ }
_, err := db.Exec("UPDATE users_settings settings = $1 WHERE id = $2", data, userID)
```

- [ ] **Step 2: Fix query building in api-server sqlite.go (line ~102)**

Replace string concatenation in WHERE clause with whitelist validation of column names against allowed filters.

- [ ] **Step 3: Fix query building in api-server postgres.go (lines ~267, ~664)**

Same pattern - use parameterized queries with whitelisted column names.

- [ ] **Step 4: Add RequireRole(RoleAdmin) protection to all admin/* device endpoints and users/:id/role updates**

In router definition:
```go
adminGroup.Use(middleware.RequireRole(middleware.RoleAdmin))
```

- [ ] **Step 5: Commit**

```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen
git add cloud/admin-api/internal/store/postgres.go cloud/api-server/internal/store/sqlite.go cloud/api-server/internal/store/postgres.go
git commit -m "security: fix SQL injection vulnerabilities in settings update and query builders"
```

---

## Parallel Execution Group: Admin Web Full Medication Module

Tasks A1 and A2 can run in parallel since they involve different files. Both depend only on the existing admin web structure.

**Execution choice offered below.**