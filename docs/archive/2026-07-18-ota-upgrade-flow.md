# OTA Upgrade Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement complete firmware OTA (Over-The-Air) upgrade flow for Eregen devices — upload firmware, push to devices via NATS/MQTT, track progress.

**Architecture:** Admin API endpoints for firmware management → NATS command bus → device MQTT client → ESP-IDF/FreeRTOS OTA handler. Progress tracked in PostgreSQL and reported back via MQTT events.

**Tech Stack:** Go + Gin (API), NATS JetStream (command bus), PostgreSQL (storage), ESP-MQTT (firmware)

## Global Constraints

- Must follow MIT/BSD/Apache-2.0/ISC only — no GPL/AGPL/LGPL dependencies
- All API responses must use `{"code": "OK", "data": ...}` format
- Device IDs must match pattern `BR-XXXX` or `PX-XXXX`
- Firmware URLs must be HTTPS
- SHA-256 hash verification required for all firmware uploads

---

### Task 1: Add OTA data models

**Files:**
- Modify: `cloud/api-server/internal/model/model.go`

**Interfaces:**
- Consumes: Existing `model.Alert`, `model.Device` types
- Produces: `FirmwareRelease`, `OTAJob`, `OTAJobProgress`, `CreateFirmwareRequest`, `PushOTARequest` types

- [ ] **Step 1: Add FirmwareRelease model**

Add to `model/model.go`:
```go
type FirmwareRelease struct {
    ID            string    `json:"id"`
    DeviceType    string    `json:"device_type"` // bracelet / pillbox
    Tier          string    `json:"tier"`        // starter / plus / pro
    Version       string    `json:"version"`
    URL           string    `json:"url"`
    Sha256Hash    string    `json:"sha256_hash"`
    Changelog     string    `json:"changelog"`
    MinAppVersion string    `json:"min_app_version,omitempty"`
    ForceUpdate   bool      `json:"force_update"`
    Active        bool      `json:"active"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Add OTAJob and progress models**

Add to `model/model.go`:
```go
type OTAJob struct {
    ID            string             `json:"id"`
    FirmwareID    string             `json:"firmware_id"`
    TargetDevices []string           `json:"target_devices"`
    Progress      OTAJobProgress     `json:"progress"`
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`
}

type OTAJobProgress struct {
    Total        int `json:"total"`
    Pending      int `json:"pending"`
    Downloading  int `json:"downloading"`
    Succeeding   int `json:"succeeding"`
    Succeeded    int `json:"succeeded"`
    Failed       int `json:"failed"`
}

type CreateFirmwareRequest struct {
    DeviceType    string `json:"device_type" binding:"required"`
    Tier          string `json:"tier" binding:"required"`
    Version       string `json:"version" binding:"required"`
    URL           string `json:"url" binding:"required"`
    Sha256Hash    string `json:"sha256_hash" binding:"required"`
    Changelog     string `json:"changelog"`
    MinAppVersion string `json:"min_app_version,omitempty"`
    ForceUpdate   bool   `json:"force_update"`
}

type PushOTARequest struct {
    FirmwareID string   `json:"firmware_id" binding:"required"`
    DeviceIDs  []string `json:"device_ids,omitempty"`
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd cloud/api-server && go build ./...`
Expected: PASS (no errors)

- [ ] **Step 4: Commit**

```bash
git add cloud/api-server/internal/model/model.go
git commit -m "feat: add OTA firmware data models"
```

---

### Task 2: Add OTA database operations to Postgres store

**Files:**
- Modify: `cloud/api-server/internal/store/postgres.go`

**Interfaces:**
- Consumes: New `model.FirmwareRelease`, `model.OTAJob` types
- Produces: `CreateFirmwareRelease()`, `ListFirmwareReleases()`, `GetFirmwareRelease()`, `CreateOTAJob()`, `GetOTAJob()`, `UpdateOTAJobProgress()` methods

- [ ] **Step 1: Add firmware release table migration comment**

Add SQL migration note at top of postgres.go:
```go
// SQL Migration needed:
// CREATE TABLE firmware_releases (
//   id UUID PRIMARY KEY, device_type TEXT NOT NULL, tier TEXT NOT NULL,
//   version TEXT NOT NULL, url TEXT NOT NULL, sha256_hash TEXT NOT NULL,
//   changelog TEXT, min_app_version TEXT, force_update BOOLEAN DEFAULT false,
//   active BOOLEAN DEFAULT true, settings JSONB,
//   created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL
// );
// CREATE TABLE ota_jobs (
//   id UUID PRIMARY KEY, firmware_id UUID NOT NULL REFERENCES firmware_releases(id),
//   target_devices JSONB NOT NULL, progress JSONB NOT NULL,
//   created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL
// );
```

- [ ] **Step 2: Implement firmware release CRUD methods**

Add to postgres.go:
```go
func (p *Postgres) CreateFirmwareRelease(ctx context.Context, r *model.FirmwareRelease) error
func (p *Postgres) ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error)
func (p *Postgres) GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error)
```

Use pgx pool with parameterized queries. Store settings as JSONB.

- [ ] **Step 3: Implement OTA job CRUD methods**

Add to postgres.go:
```go
func (p *Postgres) CreateOTAJob(ctx context.Context, j *model.OTAJob) error
func (p *Postgres) GetOTAJob(ctx context.Context, id string) (*model.OTAJob, error)
func (p *Postgres) UpdateOTAJobProgress(ctx context.Context, jobID string, fn func(*model.OTAJobProgress)) error
```

Store target_devices and progress as JSONB columns. Use optimistic locking for progress updates.

- [ ] **Step 4: Verify compilation**

Run: `cd cloud/api-server && go build ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cloud/api-server/internal/store/postgres.go
git commit -m "feat: add OTA database operations for firmware releases and jobs"
```

---

### Task 3: Implement OTA service layer

**Files:**
- Create: `cloud/api-server/internal/service/ota.go`

**Interfaces:**
- Consumes: `store.Postgres`, `NatsClient`
- Produces: `OTAService` with `CreateFirmwareRelease()`, `ListFirmwareReleases()`, `GetFirmwareRelease()`, `CreateOTAJob()`, `PushToDevices()`, `UpdateProgress()`, `GetOTAJob()` methods

- [ ] **Step 1: Define OTAService struct and constructor**

```go
type OTAService struct {
    pg   *store.Postgres
    nats *NatsClient
    log  *zap.Logger
}

func NewOTAService(pg *store.Postgres, nats *NatsClient, log *zap.Logger) *OTAService
```

- [ ] **Step 2: Implement firmware release management methods**

```go
func (s *OTAService) CreateFirmwareRelease(ctx context.Context, req *model.CreateFirmwareRequest) (*model.FirmwareRelease, error)
func (s *OTAService) ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error)
func (s *OTAService) GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error)
```

Validate device type (bracelet/pillbox), tier (starter/plus/pro), version format (semver), URL (HTTPS).

- [ ] **Step 3: Implement OTA job creation and push**

```go
func (s *OTAService) CreateOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) (*model.OTAJob, error)
func (s *OTAService) PushToDevices(ctx context.Context, job *model.OTAJob, firmware *model.FirmwareRelease) error
```

For PushToDevices, publish NATS command per device:
```json
{"type":"ota","url":"https://...","hash":"sha256:...","ver":"1.0.0","force":false}
```

- [ ] **Step 4: Implement progress tracking**

```go
func (s *OTAService) UpdateProgress(ctx context.Context, jobID, deviceID, status string) error
```

Status values: "downloading", "succeeded", "failed". Update job progress atomically.

- [ ] **Step 5: Verify compilation**

Run: `cd cloud/api-server && go build ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cloud/api-server/internal/service/ota.go
git commit -m "feat: implement OTA service for firmware management and device push"
```

---

### Task 4: Implement OTA HTTP handlers

**Files:**
- Create: `cloud/api-server/internal/handler/ota.go`

**Interfaces:**
- Consumes: `OTAService`
- Produces: HTTP handlers for firmware CRUD and OTA push endpoints

- [ ] **Step 1: Define OTAHandler and constructor**

```go
type OTAHandler struct {
    svc *service.OTAService
    log *zap.Logger
}

func NewOTAHandler(svc *service.OTAService, log *zap.Logger) *OTAHandler
```

- [ ] **Step 2: Implement CreateFirmware handler**

Endpoint: `POST /api/v1/admin/firmware`
- Bind `CreateFirmwareRequest`
- Validate inputs (device type, version semver, HTTPS URL, SHA-256 hash)
- Call `svc.CreateFirmwareRelease()`
- Return 201 with created release

- [ ] **Step 3: Implement ListFirmware and GetFirmware handlers**

Endpoints: `GET /api/v1/admin/firmware`, `GET /api/v1/admin/firmware/:id`
- Optional query params: `?device_type=bracelet&tier=pro`
- Return list or single firmware release

- [ ] **Step 4: Implement PushOTA handler**

Endpoint: `POST /api/v1/admin/ota/push`
- Bind `PushOTARequest`
- Resolve target devices (if DeviceIDs empty, find all matching by type+tier)
- Create OTA job
- Kick off async push via goroutine
- Return job ID immediately

- [ ] **Step 5: Implement GetOTAJob handler**

Endpoint: `GET /api/v1/admin/ota/jobs/:id`
- Return job with current progress

- [ ] **Step 6: Verify compilation**

Run: `cd cloud/api-server && go build ./...`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add cloud/api-server/internal/handler/ota.go
git commit -m "feat: add OTA HTTP handlers for firmware management and push"
```

---

### Task 5: Wire OTA routes into router

**Files:**
- Modify: `cloud/api-server/internal/router/router.go`

**Interfaces:**
- Consumes: `NewOTAHandler()`, `NewOTAService()`
- Produces: Route registrations under `/api/v1/admin/` group

- [ ] **Step 1: Instantiate OTA service and handler**

After existing service instantiations, add:
```go
otaSvc := service.NewOTAService(pg, nats, log)
otaH := handler.NewOTAHandler(otaSvc, log)
```

- [ ] **Step 2: Register admin routes**

Under protected group, add admin sub-group:
```go
admin := protected.Group("/admin")
{
    firmware := admin.Group("/firmware")
    {
        firmware.POST("", otaH.CreateFirmware)
        firmware.GET("", otaH.ListFirmware)
        firmware.GET("/:id", otaH.GetFirmware)
    }
    admin.POST("/ota/push", otaH.PushOTA)
    admin.GET("/ota/jobs/:id", otaH.GetOTAJob)
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd cloud/api-server && go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cloud/api-server/internal/router/router.go
git commit -m "feat: wire OTA upgrade routes into admin API group"
```

---

### Task 6: Add admin-web OTA management UI

**Files:**
- Create: `apps/admin-web/src/views/OTA.vue`
- Modify: `apps/admin-web/src/router/routes.ts` (add route)

**Interfaces:**
- Consumes: `@/api/client` for API calls
- Produces: OTA management page with firmware list, upload form, push controls, progress display

- [ ] **Step 1: Create OTA.vue component**

Build a Vue 3 + Element Plus page with:
- Firmware release list table (version, device type, tier, date, actions)
- Upload new firmware dialog (form with device_type, tier, version, URL, SHA-256 hash, changelog)
- Trigger OTA push dialog (select firmware, select target devices or "all matching")
- Job progress table showing status per job

- [ ] **Step 2: Add API helper functions**

Create `apps/admin-web/src/api/ota.ts`:
```typescript
export const otaApi = {
  listFirmware(params?: { device_type?: string; tier?: string }) {
    return apiClient.get('/admin/firmware', { params })
  },
  createFirmware(data: CreateFirmwareRequest) {
    return apiClient.post('/admin/firmware', data)
  },
  getFirmware(id: string) {
    return apiClient.get(`/admin/firmware/${id}`)
  },
  pushOTA(data: PushOTARequest) {
    return apiClient.post('/admin/ota/push', data)
  },
  getJob(id: string) {
    return apiClient.get(`/admin/ota/jobs/${id}`)
  },
}
```

- [ ] **Step 3: Register route in admin router**

Add to routes: `{ path: '/ota', name: 'OTA', component: () => import('@/views/OTA.vue'), meta: { title: '固件升级' } }`

- [ ] **Step 4: Add sidebar menu item**

Add "固件升级" (OTA) menu item under "设备管理" section.

- [ ] **Step 5: Verify TypeScript compilation**

Run: `cd apps/admin-web && npx tsc --noEmit`
Expected: PASS (no errors)

- [ ] **Step 6: Commit**

```bash
git add apps/admin-web/src/views/OTA.vue apps/admin-web/src/api/ota.ts
git add apps/admin-web/src/router/routes.ts
git commit -m "feat: add admin-web OTA management UI for firmware upgrades"
```

---

### Task 7: Add firmware update check to family-app

**Files:**
- Modify: `apps/family-app/lib/screens/settings/settings_page.dart`
- Modify: `apps/family-app/lib/models/device.dart`

**Interfaces:**
- Consumes: Device settings, API client
- Produces: App startup check for firmware updates, prompt user to notify elder's device

- [ ] **Step 1: Add firmware update check on app startup**

In `AppState` or `SettingsPage`, check for available firmware:
```dart
Future<void> checkForUpdates() async {
  final response = await ApiClient.instance.get('/firmware?device_type=bracelet&tier=pro');
  // Compare latest version with installed version
  // Show dialog if update available
}
```

- [ ] **Step 2: Add "Notify Device to Update" button**

In settings page, add action to trigger OTA push for linked device:
```dart
Future<void> notifyDeviceUpdate(String deviceId, String firmwareId) async {
  await ApiClient.instance.post('/admin/ota/push', {
    'firmware_id': firmwareId,
    'device_ids': [deviceId],
  });
  // Show success message
}
```

- [ ] **Step 3: Verify Flutter compilation**

Run: `cd apps/family-app && flutter analyze lib/screens/settings/settings_page.dart`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add apps/family-app/lib/screens/settings/settings_page.dart
git add apps/family-app/lib/models/device.dart
git commit -m "feat: add firmware update check and notification in family-app"
```
