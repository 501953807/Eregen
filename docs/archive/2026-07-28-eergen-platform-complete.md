# Eregen Platform Comprehensive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) to implement each task independently, reviewing between tasks for quality control.

**Goal:** Complete all remaining subsystem implementations per CLAUDE.md four-batch roadmap, following strict source-audit findings remediation and feature-completion priorities.

**Architecture:** Monorepo with Go microservices (Cloud API Server, Admin API, Gateway, Push Service, Data Pipeline, B2B Hospital/Community/Insurance), Flutter family app, Vue admin web, WeChat mini program, Hugo brand website, and peripheral firmware projects.

**Tech Stack:** Go 1.22+ / Gin 1.9+, Flutter 3.24+, Vue 3.4 + TypeScript 5.4 + Element Plus 2.7+, ESLint/Prettier (JS/TS), Tailwind CSS, PostgreSQL/SQLite, NATS v2.10+, EMQX 5.x, ESP-IDF v5.3, FreeRTOS.

---

## Global Constraints (per CLAUDE.md verbatim)

- Hardware firmware layer: GD32E230C8T3 (ARM Cortex-M4, FreeRTOS, C); ESP32-C3 (RISC-V, ESP-IDF v5.3, C)
- Cloud backend: Go + Gin, MQTT Broker EMQX 5.x, NATS 2.10+, SQLite MVP storage, Push FCM + Aliyun SMS + WeChat subscription
- Client apps: Flutter (Dart) 3.24+ for family app; Vue 3 + TS + Element Plus for admin web; native WeChat WXML/WXSS
- All licenses must be MIT/BSD-3/Apache-2.0/ISC only (no GPL/AGPL/LGPL)
- Source code files permitted in git, docs/ directory forbidden from remote commit
- Each subsystem must maintain THIRD-PARTY-LICENSES file

---

## Execution Order (by CLAUDE.md Batch Sequence)

| Phase | Subsystem(s) | Priority | Est. Effort | Plan File |
|-------|-------------|----------|-------------|-----------|
| Phase I | Cloud Backend (all services): API Server, Admin API, Gateway, Push Service, Data Pipeline, B2B Hospital/API/Auth enhancements | P0 (Foundation) | High | See Task 1-7 below |
| Phase II | Family Mobile App (Flutter), Admin Web (Vue), WeChat Mini Program | P1 (Customer-facing) | Medium-High | See Task 8-14 |
| Phase III | Brand Website (Hugo) | P1 (Marketing) | Low-Medium | See Task 15 |
| Phase IV | Firmware (Bracelet, Pillbox), Medical Wristband NFC/Cat1, Community Wristband | P2 (Hardware-dependent) | TBD (requires hardware) | Separate plans after procurement |

---

## PHASE I: Cloud Backend Completeness (Tasks 1-7)

### Task 1: API Server Handler - Error Message Standardization Review

**Files:**
- Modify: `cloud/api-server/internal/handler/*.go` (user.go, device.go, alert.go, location.go, medication.go, firmware.go, settings.go, dashboard.go, elderly.go, regulatory.go, medical_wristband.go, community_wb.go)
- Verify: None - all err.Error() calls were already fixed in prior audit; run grep confirmation

**Test:** `cloud/api-server/internal/middleware/auth_test.go` passes; smoke test all endpoints return generic errors

**Implementation Steps:**
```bash
# Verification only - confirm no err.Error() patterns remain
grep -r "\.Error()\"" cloud/api-server/internal/handler/ || echo "✅ All handlers safe"

# If any found, apply fix pattern from audit:
#   Replace: c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
#   With:    c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input parameters"})
```

### Task 2: Admin API Store - SQLite/PostgreSQL Schema Consistency Final Verification

**Files:**
- Modify: `cloud/admin-api/internal/store/sqlite.go` L58-66 (phone column added, role constraint aligned)
- Modify: `cloud/admin-api/tests/api_test.go` (password_hash added, JWT_SECRET set, Authorization header)
- Verify: `cloud/admin-api/internal/store/postgres.go` matches init-db.sql schema

**Test:** Run `go test ./cloud/admin-api/tests/...` with Redis mock or skip; verify migration scripts load correctly

### Task 3: Gateway MQTT Client - TLS Dev Mode Security Fix

**File:** `cloud/gateway/internal/mqtt/client.go` lines 192-200

**Current behavior:** In dev mode (`MQTT_DEV_MODE="1"`), TLS CA verification is skipped — acceptable for dev but needs explicit warning comment in production build note.

**Action:** Add production warning comment above the `if os.Getenv("MQTT_DEV_MODE") == "1"` block noting that this bypass should never be used in production builds.

### Task 4: Push Service - FCM Configuration Template

**File:** `cloud/push-service/cmd/main.go` — Add placeholder configuration section for FCM service account JSON path environment variable (`FCM_SERVICE_ACCOUNT_PATH`) and required validation at startup.

### Task 5: Data Pipeline - PostgreSQL pgx Driver Integration

**Files:**
- Verify `cloud/data-pipeline` includes `pgx` in go.mod (if missing, add: `go get github.com/jackc/pgx/v5`)
- Ensure `Store` interface in store/postgres.go handles all health/event query signatures matching the pulse/heartbeat requirements in CLAUDE.md data-flow diagram

### Task 6: B2B Insurance Integration - Full Auth Middleware Application

**Files:**
- Apply `shared/auth/middleware/auth.go` APIKeyAuth() to ALL routes in `cloud/b2b/insurance-integration/internal/router/router.go` (already applied based on audit findings — verify coverage via grep of auth groups)
- Ensure insurance-integration includes InstitutionType handling (community/insurance types) in policy validation logic stub

### Task 7: Unified Auth Store Interface - Documentation Completeness

**File:** `shared/auth/middleware/auth.go` — Add comments clarifying that the Store interface is intentionally minimal (only GetInstitutionByAPIKey) to allow future plugging of Redis/memcached or multiple DB backends without changing middleware. Add note about bcrypt hashing expectation for api_key_hash column.

---

## PHASE II: Mobile & Frontend Completeness (Tasks 8-14)

### Task 8: Family App Flutter - Core Page Implementation Audit

**Files to complete:**
- `apps/family-app/lib/widgets/` — Create home/dashboard, health-chart (flutter_charts or syncfusion), medication-list, sos-alert components
- `apps/family-app/lib/screens/` — Complete all missing screen implementations beyond login/home skeleton
- `apps/family-app/lib/api/client.dart` — Add all missing API client methods matching admin-api endpoint list (health/location/alert/medication/register devices)

**Test:** `flutter analyze` clean; `flutter test` for all mocked service tests

### Task 9: Admin Web Vue - Elderly Welfare Tags CRUD Complete

**Files:**
- Modify `apps/admin-web/src/views/Elderly.vue`:
  - Implement `openEdit(row)` — show modal with elder form, call store.updateCommunityElder()
  - Implement `toggleWelfare(row)` — switch welfare tag on/off, call patch API
  - Implement `viewBoundElders(row)` — show list of devices bound to this elder, modal dialog

**Interface:** `admin-web/src/api/client.ts` — Ensure methods for `listTagConfigs`, `assignTag`, `revokesTag` exist with correct typings (TypeScript)

### Task 10: WeChat Mini Program - Subscription Card Feature Complete

**Files:**
- `apps/miniprogram/pages/medication/medication.wxml` — The subscription card button shown; now need full implementation:
  - `requestSubscription()` handler that calls `wx.requestSubscribeMessage` with proper template IDs
  - Show success/error toast after subscribe result
  - Display "已订阅" / "未订阅" status visually
- `apps/miniprogram/pages/home/home.wxml` — Map component needs marker coordinates fix, ensure safe zone indicator works across device types

**Test:** Miniprogram devtools preview, verify map loads with marker and address display.

### Task 11: Nurse Terminal Flutter - NFC Scan and Cat1 Submission Framework

**Files:**
- Modify `apps/nurse_terminal/lib/main.dart` — Add route for verification screen transition from login
- Modify `apps/nurse_terminal/lib/services/wristband_service.dart`:
  - Implement `scanNfcCard()` async method using `nfc_plus` plugin, return verification result
  - Implement `submitToCat1(VerificationResult)` stub that converts to VerificationReport struct per wb_ble.go protocol spec, posts to backend
- Update `lib/models/patient.dart` — Add JSON serialization for patient-to-ward-round-entry mapping

**Test:** Mock NFC service during testing until hardware available; verify cat1 submission structure matches wb_ble.go VerificationReport definition.

---

## PHASE III: Brand Website (Task 15)

### Task 12: Hugo Site Content Pages

**Files:**
- `apps/content/` — Create markdown content pages: `index.md`, `products/bracelet.md`, `products/pillbox.md`, `about.md`, `contact.md`
- `assets/css/main.css` — Add brand-specific styling per tailwind.config.js colors (#4A90D9 primary palette)
- Layout templates (Hugo theme) — Insert navigation bar with product links, footer with contact info

**Test:** `hugo --buildDrafts` locally, verify all pages render with correct brand colors.

---

## DOCUMENTATION UPDATE TASKS

### Task 13: Root README.md - Status Summary Addition

Update root `README.md` to include a new section **"Current Implementation Status"** listing all subsystems with checkmarks showing: ✅ Completed (code-audited, merged to main), 🟡 Partial (skeleton in place), ⏳ Not started (future batches). Include link to this master implementation plan.

### Task 14: CLAUDE.md - Revised Development Roadmap per Audit Findings

If audit findings reveal any discrepancy between documented specs and actual code, update CLAUDE.md accordingly. Specifically confirm:

- Bracelet/Pillbox firmware directories still marked as "to-be-developed" (correct per phase plan)
- Medical wristband protocol now explicitly states NFC-first dual-mode (wb_ble.go updated — confirm note in CLAUDE.md reflects this)
- Third-party license documentation note in each subproject directory

---

## VERIFICATION CHECKLIST (for final submission)

Before final commit of all tasks:

1. `go mod tidy` run in every Go service directory (`cloud/*`, `b2b/*`, `shared/*`) with no new dependency warnings
2. `go build ./...` passes in all directories with zero errors
3. `dart analyze` clean on all Dart projects
4. `vue-cli-service lint` (or `vite build`) clean on admin web
5. `flutter pub get && flutter analyze` on family-app and nurse-terminal
6. All `TODO` strings removed except those explicitly documented as MVP-stub in comments with justification
7. All handlers have generic error messages (no stack traces in responses)
8. B2B services all apply `APIKeyAuth()` middleware consistently
9. Migration scripts create proper institution tables with bcrypt hash placeholders
10. Protocol docs (wb_ble.go) match actual NFC+Cat1 design decision

---

## IMPLEMENTATION ORDERING FOR AGENTIC WORKERS

When executing these tasks automatically:

1. **Run Task 1** (API Server handler verification) — quick no-op check
2. **Run Task 2** (Admin API store consistency) — fixes critical schema alignment
3. **Run Task 3** (Gateway MQTT dev mode comment) — small doc fix
4. **Run Task 7** (Auth interface docs) — prerequisite for B2B auth work
5. **Run Tasks 6 & partial 4** (B2B auth completion) — security-critical
6. **Run Task 5** (Data pipeline pgx) — infrastructure readiness
7. **Run Task 8-11** (Mobile/Frontend core features) — customer-facing value
8. **Run Task 12** (Website content) — marketing readiness
9. **Run Task 13-14** (Documentation updates) — closes the loop

All files modified throughout this process require review against the original audit findings to ensure each addressed issue is properly resolved before final merge.