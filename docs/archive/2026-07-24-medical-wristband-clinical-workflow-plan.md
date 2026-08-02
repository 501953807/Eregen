# Medical Wristband Clinical Workflow End-to-End Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the full clinical workflow — hospital admission → BLE/NFC verification → ward round → medication scanning → discharge — with regulatory rule engine R01-R08.

**Architecture:** Add 4 new handler methods (AdmitPatient, CompleteWardRound, DischargePatient, GetWardRound) + 1 new service (regulatory_engine.go) + DB schema migrations for admissions/ward_rounds/regulatory_alerts tables. Frontend adds 3 new tabs to MedicalWristband.vue (Admissions, Ward Rounds, Regulatory Alerts). AuditDetail.vue wires real API data.

**Tech Stack:** Go/Gin (backend), SQLite/PostgreSQL (dual-store via adapter), Vue 3/TypeScript/Element Plus (admin-web)

## Global Constraints

- Dual-store pattern: every method must implement both SqliteStore and PostgresStore, routed through StoreAdapter (dbType == "postgres" ? Postgres : Sqlite)
- Store interface in `cloud/admin-api/internal/store/store.go` must include all new method signatures
- All models use UUID v4 for IDs: `uuid.New().String()`
- Time format: RFC3339 for JSON, `datetime('now')` for SQLite, `NOW()` for PostgreSQL
- No GPL/AGPL/LGPL dependencies (CLAUDE.md patent strategy)
- Each backend method needs both adapter dispatch + SqliteStore impl + PostgresStore impl

---

### Task 1: Model Definitions — New Types

**Files:**
- Modify: `cloud/admin-api/internal/model/model.go:600-620` (after existing Medical types)

**Interfaces:**
- Consumes: existing `model.MedicalPatient`, `model.RegulatoryAlert`
- Produces: `HospitalAdmission`, `WardRoundEntry`, `RegulatoryRuleResult`

Add these structs to `model.go`:

```go
// HospitalAdmission represents a patient's hospital stay record.
type HospitalAdmission struct {
    ID               string    `json:"id"`
    PatientID        string    `json:"patient_id"`
    AdmissionNo      string    `json:"admission_no"`
    BedNo            string    `json:"bed_no"`
    Department       string    `json:"department"`
    Diagnosis        string    `json:"diagnosis,omitempty"`
    EmergencyContact string    `json:"emergency_contact,omitempty"`
    Allergies        string    `json:"allergies,omitempty"`
    AdmittedAt       time.Time `json:"admitted_at"`
    ExpectedDischargeAt *time.Time `json:"expected_discharge_at,omitempty"`
    DischargedAt       *time.Time `json:"discharged_at,omitempty"`
    DischargeType    string    `json:"discharge_type,omitempty"` // "discharged", "transferred", "deceased"
    TransferredTo    string    `json:"transferred_to,omitempty"`
    Notes            string    `json:"notes,omitempty"`
}

// WardRoundEntry represents a nursing round visit with vitals.
type WardRoundEntry struct {
    ID          string    `json:"id"`
    PatientID   string    `json:"patient_id"`
    NurseID     string    `json:"nurse_id"`
    BloodPressure string  `json:"blood_pressure,omitempty"` // e.g. "120/80"
    HeartRate   int       `json:"heart_rate,omitempty"`
    SpO2        int       `json:"spo2,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    Weight      float64   `json:"weight,omitempty"`
    Notes       string    `json:"notes,omitempty"`
    Observations string   `json:"observations,omitempty"` // JSON array of checkboxes
    CompletedAt time.Time `json:"completed_at"`
}

// RegulatoryRuleResult holds the outcome of a single rule evaluation.
type RegulatoryRuleResult struct {
    RuleCode    string    `json:"rule_code"` // "R01"..."R08"
    Severity    string    `json:"severity"`  // "P0", "P1", "P2"
    PatientID   string    `json:"patient_id,omitempty"`
    Message     string    `json:"message"`
    TriggeredAt time.Time `json:"triggered_at"`
    Resolved    bool      `json:"resolved"`
}
```

- [ ] **Step 1: Add model structs to model.go**

Append after line 620 (after `RuleAlertCount`):

```go
// ========== Clinical Workflow Models ==========

type HospitalAdmission struct { ... }
type WardRoundEntry struct { ... }
type RegulatoryRuleResult struct { ... }
```

- [ ] **Step 2: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 3: Commit**

```bash
git add cloud/admin-api/internal/model/model.go
git commit -m "feat: add HospitalAdmission, WardRoundEntry, RegulatoryRuleResult models"
```

---

### Task 2: Store Interface — New Method Signatures

**Files:**
- Modify: `cloud/admin-api/internal/store/store.go:82-105` (after existing medical wristband methods)

**Interfaces:**
- Consumes: existing `model.HospitalAdmission`, `model.WardRoundEntry`, `model.RegulatoryRuleResult`
- Produces: new interface methods for StoreAdapter routing

Add these signatures to the `Store` interface (before the regulatory section at line ~83):

```go
// Clinical workflow
CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error
GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error)
ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error)
CompleteAdmission(ctx context.Context, id string, dischargeType, notes, transferredTo string) error
CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error
ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error)
EvaluateRegulatoryRules(ctx context.Context, event string, data map[string]string) ([]*model.RegulatoryRuleResult, error)
```

- [ ] **Step 1: Add interface methods to store.go**

Insert after line 79 (after `GetPatientHistory`), before regulatory section:

```go
// Clinical workflow
CreateAdmission(...)
GetAdmission(...)
ListAdmissions(...)
CompleteAdmission(...)
CreateWardRound(...)
ListWardRounds(...)
EvaluateRegulatoryRules(...)
```

- [ ] **Step 2: Verify interface compiles**

Run: `cd cloud/admin-api && go build ./...`
Expected: compile error — "SqliteStore does not implement Store (missing method CreateAdmission)"

- [ ] **Step 3: Commit**

```bash
git add cloud/admin-api/internal/store/store.go
git commit -m "feat: add clinical workflow methods to Store interface"
```

---

### Task 3: SQLite Store — Clinical Workflow Methods

**Files:**
- Modify: `cloud/admin-api/internal/store/sqlite.go` (add after line ~1520, before `var _ Store = ...`)

**Interfaces:**
- Consumes: `*sql.DB` from existing SqliteStore
- Produces: CreateAdmission, GetAdmission, ListAdmissions, CompleteAdmission, CreateWardRound, ListWardRounds, EvaluateRegulatoryRules

First, add table migrations in the `migrate()` function (around line 42):

```go
// In migrate() migrations slice, add:
`CREATE TABLE IF NOT EXISTS hospital_admissions (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL,
    admission_no TEXT NOT NULL,
    bed_no TEXT NOT NULL,
    department TEXT NOT NULL,
    diagnosis TEXT,
    emergency_contact TEXT,
    allergies TEXT,
    admitted_at TEXT NOT NULL DEFAULT (datetime('now')),
    expected_discharge_at TEXT,
    discharged_at TEXT,
    discharge_type TEXT,
    transferred_to TEXT,
    notes TEXT
)`,
`CREATE TABLE IF NOT EXISTS ward_rounds (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL,
    nurse_id TEXT NOT NULL,
    blood_pressure TEXT,
    heart_rate INTEGER,
    spo2 INTEGER,
    temperature REAL,
    weight REAL,
    notes TEXT,
    observations TEXT,
    completed_at TEXT NOT NULL DEFAULT (datetime('now'))
)`,
```

Then implement the methods:

```go
func (s *SqliteStore) CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO hospital_admissions (id, patient_id, admission_no, bed_no, department, diagnosis,
         emergency_contact, allergies, admitted_at, expected_discharge_at, discharge_type, transferred_to, notes)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?, ?, ?, ?)`,
        a.ID, a.PatientID, a.AdmissionNo, a.BedNo, a.Department, a.Diagnosis,
        a.EmergencyContact, a.Allergies, a.ExpectedDischargeAt, a.DischargeType, a.TransferredTo, a.Notes)
    return err
}

func (s *SqliteStore) GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error) {
    var a model.HospitalAdmission
    var expectedDischarge, dischargedAt, transferTo string
    err := s.db.QueryRowContext(ctx,
        `SELECT id, patient_id, admission_no, bed_no, department, diagnosis, emergency_contact,
         allergies, admitted_at, expected_discharge_at, discharged_at, discharge_type, transferred_to, notes
         FROM hospital_admissions WHERE id = ?`, id).Scan(
        &a.ID, &a.PatientID, &a.AdmissionNo, &a.BedNo, &a.Department, &a.Diagnosis,
        &a.EmergencyContact, &a.Allergies, &a.AdmittedAt, &expectedDischarge, &dischargedAt,
        &a.DischargeType, &transferTo, &a.Notes)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("admission not found")
        }
        return nil, fmt.Errorf("get admission: %w", err)
    }
    if expectedDischarge != "" {
        t, _ := time.Parse(time.RFC3339, expectedDischarge)
        a.ExpectedDischargeAt = &t
    }
    if dischargedAt != "" {
        t, _ := time.Parse(time.RFC3339, dischargedAt)
        a.DischargedAt = &t
    }
    a.TransferredTo = transferTo
    return &a, nil
}

func (s *SqliteStore) ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error) {
    query := `SELECT id, patient_id, admission_no, bed_no, department, diagnosis, emergency_contact,
              allergies, admitted_at, expected_discharge_at, discharged_at, discharge_type, transferred_to, notes
              FROM hospital_admissions WHERE 1=1`
    var args []interface{}
    idx := 1
    if department != "" {
        query += fmt.Sprintf(" AND department=?")
        args = append(args, department)
        idx++
    }
    if status != "" {
        query += fmt.Sprintf(" AND (discharged_at IS NULL OR discharge_type != ?)")
        args = append(args, status)
        idx++
    }
    query += fmt.Sprintf(" ORDER BY admitted_at DESC LIMIT ? OFFSET ?")
    args = append(args, pageSize, (page-1)*pageSize)

    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list admissions: %w", err)
    }
    defer rows.Close()

    var items []model.HospitalAdmission
    for rows.Next() {
        var a model.HospitalAdmission
        var expectedDischarge, dischargedAt, transferTo string
        if err := rows.Scan(&a.ID, &a.PatientID, &a.AdmissionNo, &a.BedNo, &a.Department,
            &a.Diagnosis, &a.EmergencyContact, &a.Allergies, &a.AdmittedAt,
            &expectedDischarge, &dischargedAt, &a.DischargeType, &transferTo, &a.Notes); err != nil {
            return nil, fmt.Errorf("scan admission: %w", err)
        }
        if expectedDischarge != "" {
            t, _ := time.Parse(time.RFC3339, expectedDischarge)
            a.ExpectedDischargeAt = &t
        }
        if dischargedAt != "" {
            t, _ := time.Parse(time.RFC3339, dischargedAt)
            a.DischargedAt = &t
        }
        a.TransferredTo = transferTo
        items = append(items, a)
    }
    return items, rows.Err()
}

func (s *SqliteStore) CompleteAdmission(ctx context.Context, id, dischargeType, notes, transferredTo string) error {
    _, err := s.db.ExecContext(ctx,
        `UPDATE hospital_admissions SET discharged_at=datetime('now'), discharge_type=?, notes=?, transferred_to=? WHERE id=?`,
        dischargeType, notes, transferredTo, id)
    if err != nil {
        return err
    }
    // Also get patient_id to update patient status
    var patientID string
    err = s.db.QueryRowContext(ctx, `SELECT patient_id FROM hospital_admissions WHERE id=?`, id).Scan(&patientID)
    if err == nil {
        s.db.ExecContext(ctx, `UPDATE medical_wristband_patients SET status='discharged', updated_at=datetime('now') WHERE id=?`, patientID)
        // Unbind any wristband bindings for this patient
        s.db.ExecContext(ctx, `UPDATE medical_bindings SET unbound_at=datetime('now') WHERE patient_id=? AND unbound_at IS NULL`, patientID)
        s.db.ExecContext(ctx, `UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle' WHERE id IN (SELECT device_id FROM medical_bindings WHERE patient_id=? AND unbound_at IS NULL)`, patientID)
    }
    return nil
}

func (s *SqliteStore) CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO ward_rounds (id, patient_id, nurse_id, blood_pressure, heart_rate, spo2,
         temperature, weight, notes, observations, completed_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
        w.ID, w.PatientID, w.NurseID, w.BloodPressure, w.HeartRate, w.SpO2,
        w.Temperature, w.Weight, w.Notes, w.Observations)
    return err
}

func (s *SqliteStore) ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error) {
    rows, err := s.db.QueryContext(ctx,
        `SELECT id, patient_id, nurse_id, blood_pressure, heart_rate, spo2, temperature, weight, notes, observations, completed_at
         FROM ward_rounds WHERE patient_id=? ORDER BY completed_at DESC`, patientID)
    if err != nil {
        return nil, fmt.Errorf("list ward rounds: %w", err)
    }
    defer rows.Close()

    var items []model.WardRoundEntry
    for rows.Next() {
        var w model.WardRoundEntry
        if err := rows.Scan(&w.ID, &w.PatientID, &w.NurseID, &w.BloodPressure, &w.HeartRate,
            &w.SpO2, &w.Temperature, &w.Weight, &w.Notes, &w.Observations, &w.CompletedAt); err != nil {
            return nil, fmt.Errorf("scan ward round: %w", err)
        }
        items = append(items, w)
    }
    return items, rows.Err()
}

func (s *SqliteStore) EvaluateRegulatoryRules(ctx context.Context, event string, data map[string]string) ([]*model.RegulatoryRuleResult, error) {
    var results []*model.RegulatoryRuleResult
    now := time.Now().UTC()

    switch event {
    case "patient_admitted":
        // R01: Bed-fraud detection — check if patient has active wristband binding
        patientID := data["patient_id"]
        var bindingCount int
        s.db.QueryRowContext(ctx,
            `SELECT COUNT(*) FROM medical_bindings WHERE patient_id=? AND unbound_at IS NULL`, patientID).Scan(&bindingCount)
        if bindingCount == 0 {
            results = append(results, &model.RegulatoryRuleResult{
                RuleCode: "R01", Severity: "P1", PatientID: patientID,
                Message: "Patient admitted without active wristband binding", TriggeredAt: now,
            })
        }
    case "patient_discharged":
        // R08: Post-discharge data retention — handled by marking admissions
    case "ward_round_completed":
        // No specific rules triggered
    case "verification_scan":
        // R05: Medication-verification mismatch
        scanType := data["scan_type"]
        if scanType == "medication" {
            patientID := data["patient_id"]
            var bindingCount int
            s.db.QueryRowContext(ctx,
                `SELECT COUNT(*) FROM medical_bindings mb JOIN medical_wristband_patients p ON p.id=mb.patient_id
                 WHERE p.id=? AND p.status='admitted' AND mb.unbound_at IS NULL`, patientID).Scan(&bindingCount)
            if bindingCount == 0 {
                results = append(results, &model.RegulatoryRuleResult{
                    RuleCode: "R05", Severity: "P2", PatientID: patientID,
                    Message: "Medication verification without active wristband binding", TriggeredAt: now,
                })
            }
        }
    }
    return results, nil
}
```

- [ ] **Step 1: Add table migrations to sqlite.go**

Edit `migrate()` function around line 42 to add `hospital_admissions` and `ward_rounds` DDL.

- [ ] **Step 2: Implement all 7 methods in sqlite.go**

Append after line ~1520 (after `GetPatientHistory`).

- [ ] **Step 3: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cloud/admin-api/internal/store/sqlite.go
git commit -m "feat: add clinical workflow SQLite store methods"
```

---

### Task 4: PostgreSQL Store — Clinical Workflow Methods

**Files:**
- Modify: `cloud/admin-api/internal/store/postgres.go` (add after line ~850, before stub methods)

**Interfaces:**
- Consumes: `*sql.DB` from existing PostgresStore
- Produces: same 7 methods as Task 3 but with PostgreSQL syntax ($1, $2, NOW())

Implement the same 7 methods with PostgreSQL placeholders. Key differences from SQLite:
- `$1, $2, ...` parameter placeholders
- `NOW()` instead of `datetime('now')`
- `ON CONFLICT DO NOTHING` for upserts
- `time.RFC3339` parsing stays the same

For `EvaluateRegulatoryRules`, use the same logic but with `$1` parameters.

- [ ] **Step 1: Implement all 7 methods in postgres.go**

Append after line ~850 (after `GetPatientHistory`), before the stub methods at ~921.

- [ ] **Step 2: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add cloud/admin-api/internal/store/postgres.go
git commit -m "feat: add clinical workflow PostgreSQL store methods"
```

---

### Task 5: StoreAdapter — Routing Methods

**Files:**
- Modify: `cloud/admin-api/internal/store/adapter.go` (add after line ~429)

**Interfaces:**
- Consumes: dbType dispatch to SqliteStore/PostgresStore
- Produces: 7 adapter methods matching the Store interface

```go
func (a *StoreAdapter) CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error {
    if a.dbType == "postgres" {
        return (&PostgresStore{db: a.db}).CreateAdmission(ctx, a)
    }
    return (&SqliteStore{db: a.db}).CreateAdmission(ctx, a)
}
// ... repeat for GetAdmission, ListAdmissions, CompleteAdmission, CreateWardRound, ListWardRounds, EvaluateRegulatoryRules
```

- [ ] **Step 1: Add 7 adapter methods to adapter.go**

After line ~429 (after `GetTodayVerificationStats`), before the community-wb section.

- [ ] **Step 2: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add cloud/admin-api/internal/store/adapter.go
git commit -m "feat: add clinical workflow StoreAdapter routing methods"
```

---

### Task 6: Handler — AdmitPatient, DischargePatient, WardRound, GetWardRound

**Files:**
- Modify: `cloud/admin-api/internal/handler/medical_wristband.go` (add after line ~460)

**Interfaces:**
- Consumes: `h.store.CreateAdmission()`, `h.store.CompleteAdmission()`, `h.store.CreateWardRound()`, `h.store.ListWardRounds()`, `h.store.EvaluateRegulatoryRules()`
- Produces: HTTP JSON responses for each endpoint

Add these handler methods following the existing pattern:

```go
// AdmitPatient registers a new hospital admission with wristband binding.
func (h *MedicalWristbandHandler) AdmitPatient(c *gin.Context) {
    var req struct {
        PatientID        string `json:"patient_id" binding:"required"`
        BedNo            string `json:"bed_no" binding:"required"`
        Department       string `json:"department" binding:"required"`
        Diagnosis        string `json:"diagnosis"`
        EmergencyContact string `json:"emergency_contact"`
        Allergies        string `json:"allergies"`
        ExpectedStayDays int    `json:"expected_stay_days"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    admission := &model.HospitalAdmission{
        ID:               uuid.New().String(),
        PatientID:        req.PatientID,
        BedNo:            req.BedNo,
        Department:       req.Department,
        Diagnosis:        req.Diagnosis,
        EmergencyContact: req.EmergencyContact,
        Allergies:        req.Allergies,
    }
    if req.ExpectedStayDays > 0 {
        t := time.Now().AddDate(0, 0, req.ExpectedStayDays)
        admission.ExpectedDischargeAt = &t
    }

    if err := h.store.CreateAdmission(c.Request.Context(), admission); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Evaluate regulatory rules
    h.store.EvaluateRegulatoryRules(c.Request.Context(), "patient_admitted", map[string]string{
        "patient_id": req.PatientID,
    })

    c.JSON(http.StatusCreated, gin.H{"data": admission})
}

// DischargePatient completes an admission.
func (h *MedicalWristbandHandler) DischargePatient(c *gin.Context) {
    admissionID := c.Param("id")
    var body struct {
        DischargeType string `json:"discharge_type" binding:"required"`
        Notes         string `json:"notes"`
        TransferredTo string `json:"transferred_to"`
    }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    if err := h.store.CompleteAdmission(c.Request.Context(), admissionID, body.DischargeType, body.Notes, body.TransferredTo); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "discharged"})
}

// GetWardRound returns scheduled/completed ward rounds for a patient.
func (h *MedicalWristbandHandler) GetWardRound(c *gin.Context) {
    patientID := c.Param("id")
    rounds, err := h.store.ListWardRounds(c.Request.Context(), patientID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": rounds})
}

// CompleteWardRound records a nursing round entry with vitals.
func (h *MedicalWristbandHandler) CompleteWardRound(c *gin.Context) {
    patientID := c.Param("id")
    var entry model.WardRoundEntry
    entry.PatientID = patientID
    if err := c.ShouldBindJSON(&entry); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    entry.ID = uuid.New().String()
    if err := h.store.CreateWardRound(c.Request.Context(), &entry); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"data": entry})
}
```

- [ ] **Step 1: Add 4 handler methods to medical_wristband.go**

After line ~460 (after `CreateAlertTagConfig`).

- [ ] **Step 2: Ensure `uuid` import is present**

Check that `"github.com/google/uuid"` is in the imports. If not, add it.

- [ ] **Step 3: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cloud/admin-api/internal/handler/medical_wristband.go
git commit -m "feat: add AdmitPatient, DischargePatient, GetWardRound, CompleteWardRound handlers"
```

---

### Task 7: Router — Register New Endpoints

**Files:**
- Modify: `cloud/admin-api/internal/router/router.go:107-154` (inside the `med` group)

**Interfaces:**
- Consumes: `medical.AdmitPatient`, `medical.DischargePatient`, `medical.GetWardRound`, `medical.CompleteWardRound`, `medical.ListAdmissions`
- Produces: new route registrations

Add these routes inside the `med` group block (after line ~153):

```go
// Clinical workflow endpoints
med.POST("/admissions", medical.AdmitPatient)
med.GET("/admissions", medical.ListAdmissions) // reuse ListPatients or create alias
med.POST("/admissions/:id/discharge", medical.DischargePatient)
med.GET("/patients/:id/ward-round", medical.GetWardRound)
med.POST("/patients/:id/ward-round", medical.CompleteWardRound)
```

For `ListAdmissions`, either alias to `ListPatients` with status filter or create a thin wrapper. The simplest: the existing `ListPatients` already supports `?status=admitted` filter. Add a dedicated handler only if needed.

- [ ] **Step 1: Add 5 new route registrations to router.go**

Inside the `med` group, after line ~153.

- [ ] **Step 2: Verify compilation**

Run: `cd cloud/admin-api && go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add cloud/admin-api/internal/router/router.go
git commit -m "feat: register clinical workflow API routes"
```

---

### Task 8: Admin-web API Client — New Endpoints

**Files:**
- Modify: `apps/admin-web/src/api/medical.ts`

**Interfaces:**
- Consumes: `apiClient.get/post/put/delete`
- Produces: new API method definitions for admissions, ward rounds, regulatory alerts

Add these methods to `medicalApi` object (after line ~122):

```typescript
// Clinical workflow
admitPatient(data: { patient_id: string; bed_no: string; department: string; diagnosis?: string; emergency_contact?: string; allergies?: string; expected_stay_days?: number }) {
    return apiClient.post('/medical/admissions', data)
},
listAdmissions(params?: { page?: number; page_size?: number; department?: string; status?: string }) {
    return apiClient.get('/medical/admissions', { params })
},
dischargePatient(id: string, data: { discharge_type: string; notes?: string; transferred_to?: string }) {
    return apiClient.post(`/medical/admissions/${id}/discharge`, data)
},
getWardRounds(patientId: string) {
    return apiClient.get(`/medical/patients/${patientId}/ward-round`)
},
completeWardRound(patientId: string, data: { nurse_id: string; blood_pressure?: string; heart_rate?: number; spo2?: number; temperature?: number; weight?: number; notes?: string; observations?: string }) {
    return apiClient.post(`/medical/patients/${patientId}/ward-round`, data)
},
getRegulatoryAlerts(params?: { rule_code?: string; severity?: string; status?: string; department?: string; page?: number; page_size?: number }) {
    return apiClient.get('/regulatory/alerts', { params })
},
resolveRegulatoryAlert(alertId: string, data: { user_id: string; notes?: string }) {
    return apiClient.put(`/regulatory/alerts/${alertId}/resolve`, data)
},
getAuditTrail(patientId: string) {
    return apiClient.get(`/regulatory/audit/patient/${patientId}`)
},
exportReport(patientId: string) {
    return apiClient.get(`/regulatory/compliance/report`, { params: { patient_id: patientId } })
},
```

Also add TypeScript interfaces:

```typescript
export interface HospitalAdmission {
    id: string
    patient_id: string
    admission_no: string
    bed_no: string
    department: string
    diagnosis?: string
    emergency_contact?: string
    allergies?: string
    admitted_at: string
    expected_discharge_at?: string
    discharged_at?: string
    discharge_type?: string
    transferred_to?: string
    notes?: string
}

export interface WardRoundEntry {
    id: string
    patient_id: string
    nurse_id: string
    blood_pressure?: string
    heart_rate?: number
    spo2?: number
    temperature?: number
    weight?: number
    notes?: string
    observations?: string
    completed_at: string
}

export interface RegulatoryAlert {
    id: string
    rule_code: string
    severity: string
    patient_id?: string
    message: string
    triggered_at: string
    resolved: boolean
}
```

- [ ] **Step 1: Add new API methods and interfaces to medical.ts**

After line ~122 (after existing methods).

- [ ] **Step 2: Verify TypeScript compilation**

Run: `cd apps/admin-web && npx vue-tsc --noEmit`
Expected: PASS (or at least no new errors from our changes)

- [ ] **Step 3: Commit**

```bash
git add apps/admin-web/src/api/medical.ts
git commit -m "feat: add clinical workflow API client methods and TypeScript types"
```

---

### Task 9: MedicalWristband.vue — New Tabs (Admissions, Ward Rounds, Regulatory Alerts)

**Files:**
- Modify: `apps/admin-web/src/views/MedicalWristband.vue`

**Interfaces:**
- Consumes: `medicalApi.admitPatient()`, `medicalApi.dischargePatient()`, `medicalApi.getWardRounds()`, `medicalApi.completeWardRound()`, `medicalApi.getRegulatoryAlerts()`, `medicalApi.resolveRegulatoryAlert()`
- Produces: 3 new el-tab-pane sections

**Tab 1: 入院登记 (Admissions)** — Replace current "入院登记" tab content with a more structured form:
- Table showing active admissions (status=admitted)
- Columns: 住院号 | 姓名 | 科室 | 床号 | 入院时间 | 预计出院 | 状态 | 操作
- "办理入院" button opens dialog with: patient_id (searchable dropdown), bed_no, department, diagnosis, emergency_contact, allergies, expected_stay_days
- "出院结算" action per row calls `dischargePatient()`

**Tab 2: 巡房记录 (Ward Rounds)** — New tab:
- Patient selector dropdown
- For selected patient: list of ward rounds with vitals summary
- "开始巡房" button opens form: BP, HR, SpO2, Temperature, Weight, Notes, Observation checkboxes (falls, confusion, pain, appetite)
- Submits via `completeWardRound()`

**Tab 3: 规则告警 (Regulatory Alerts)** — New tab:
- Reuses existing regulatory alert data from `/regulatory/alerts`
- Table: 规则代码 | 严重程度 | 患者 | 告警信息 | 触发时间 | 状态 | 操作
- Color-coded severity badges (P0=red, P1=orange, P2=blue)
- "Resolve" button calls `resolveRegulatoryAlert()`

Implementation approach:
- Add `activeTab.value = 'admissions'` as default
- Add `admissions`, `wardRounds`, `regulatoryAlerts` refs
- Add `showAdmitDialog`, `showWardRoundDialog`, `admitForm`, `wardRoundForm` refs
- Load functions: `loadAdmissions()`, `loadWardRounds()`, `loadRegulatoryAlerts()`
- Action functions: `handleAdmit()`, `handleDischarge()`, `handleWardRound()`

- [ ] **Step 1: Add new template sections to MedicalWristband.vue**

After the existing Daily Entries tab (line ~186), add 3 new el-tab-pane blocks.

- [ ] **Step 2: Add new script refs and functions**

After existing script (line ~354), add state variables and CRUD functions for the 3 new tabs.

- [ ] **Step 3: Verify Vite build**

Run: `cd apps/admin-web && npx vite build`
Expected: clean build

- [ ] **Step 4: Commit**

```bash
git add apps/admin-web/src/views/MedicalWristband.vue
git commit -m "feat: add Admissions, Ward Rounds, Regulatory Alerts tabs to MedicalWristband"
```

---

### Task 10: AuditDetail.vue — Real API Integration

**Files:**
- Modify: `apps/admin-web/src/views/AuditDetail.vue`

**Interfaces:**
- Consumes: `medicalApi.getAuditTrail()`, `medicalApi.getWardRounds()`, `medicalApi.listVerifications()`, `medicalApi.listExpenses()`, `medicalApi.listMedications()`
- Produces: timeline nodes populated from real API data instead of hardcoded mocks

Current state: AuditDetail.vue has a timeline with 7 nodes (admission, verification, ward_round, medication, test_result, expense, discharge) but uses hardcoded mock data.

Replace the mock data loading with real API calls:

```typescript
async function loadAuditTrail() {
  try {
    const trailRes = await medicalApi.getAuditTrail(route.params.patientId as string)
    auditTrail.value = trailRes.data?.data || null
    
    const roundsRes = await medicalApi.getWardRounds(route.params.patientId as string)
    wardRounds.value = roundsRes.data?.data || []
    
    const verifRes = await medicalApi.listVerifications({ page: 1, page_size: 50 })
    verifications.value = verifRes.data?.data || []
    
    const expensesRes = await medicalApi.listExpenses(route.params.patientId as string, { page: 1, page_size: 50 })
    expenses.value = expensesRes.data?.data || []
    
    const medsRes = await medicalApi.listMedications(route.params.patientId as string)
    medications.value = medsRes.data?.data || []
  } catch (e: any) {
    ElMessage.error('加载审计数据失败: ' + (e.message || 'unknown error'))
  }
}
```

Wire timeline nodes to actual data:
- Node 1 (入院): from `auditTrail.admission`
- Node 2 (腕带绑定): from `verifications` filtered by type
- Node 3 (巡房): from `wardRounds`
- Node 4 (用药核对): from `medications`
- Node 5 (检验结果): from `test_results` (already exists)
- Node 6 (费用): from `expenses`
- Node 7 (出院): from `auditTrail.discharge`

- [ ] **Step 1: Replace mock data with real API calls in AuditDetail.vue**

Find the `onMounted` or data-loading section and replace hardcoded arrays with API calls.

- [ ] **Step 2: Verify Vite build**

Run: `cd apps/admin-web && npx vite build`
Expected: clean build

- [ ] **Step 3: Commit**

```bash
git add apps/admin-web/src/views/AuditDetail.vue
git commit -m "feat: wire AuditDetail.vue timeline to real API data"
```

---

### Task 11: Nurse Terminal — Login Screen + Home Screen

**Files:**
- Create: `apps/nurse_terminal/lib/src/screens/login_screen.dart`
- Create: `apps/nurse_terminal/lib/src/screens/home_screen.dart`
- Create: `apps/nurse_terminal/lib/src/services/api_client.dart`
- Modify: `apps/nurse_terminal/lib/main.dart`

**Interfaces:**
- Consumes: `MedicalWristbandService` (existing BLE service)
- Produces: login flow → home screen → patient list → navigation to detail screens

**api_client.dart** — Simple HTTP client wrapping admin-api:

```dart
class ApiClient {
  final String baseUrl;
  String? _token;

  ApiClient({this.baseUrl = 'http://localhost:8081'});

  void setToken(String token) => _token = token;

  Future<Map<String, dynamic>> post(String path, Map<String, dynamic>? body) async {
    final response = await http.post(
      Uri.parse('$baseUrl$path'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $_token',
      },
      body: json.encode(body),
    );
    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception('API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> get(String path) async {
    final response = await http.get(
      Uri.parse('$baseUrl$path'),
      headers: {'Authorization': 'Bearer $_token'},
    );
    if (response.statusCode != 200) {
      throw Exception('API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }
}
```

**login_screen.dart** — Simple username/password form:

```dart
class LoginScreen extends StatefulWidget { ... }
// Form: username + password → POST /api/v1/admin/login (or whatever auth exists)
// On success: store token in SharedPreferences, navigate to HomeScreen
```

**home_screen.dart** — Patient list from API:

```dart
class HomeScreen extends StatefulWidget { ... }
// Loads patients via apiClient.get('/medical/patients?page=1&page_size=50&status=admitted')
// Each patient card: name, bed_no, department, wristband_status (bound/idle)
// FAB: "Scan Wristband" → navigates to verification screen
// Tap patient → navigates to PatientDetailScreen
```

Update `main.dart` to show LoginScreen as home (or keep BleScanPage as fallback if no token).

- [ ] **Step 1: Create api_client.dart**

Write `apps/nurse_terminal/lib/src/services/api_client.dart`.

- [ ] **Step 2: Create login_screen.dart**

Write `apps/nurse_terminal/lib/src/screens/login_screen.dart`.

- [ ] **Step 3: Create home_screen.dart**

Write `apps/nurse_terminal/lib/src/screens/home_screen.dart`.

- [ ] **Step 4: Update main.dart**

Change `home: const BleScanPage()` to `home: const LoginScreen()`.

- [ ] **Step 5: Verify Flutter analyze**

Run: `cd apps/nurse_terminal && flutter analyze`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add apps/nurse_terminal/lib/src/services/api_client.dart \
         apps/nurse_terminal/lib/src/screens/login_screen.dart \
         apps/nurse_terminal/lib/src/screens/home_screen.dart \
         apps/nurse_terminal/lib/main.dart
git commit -m "feat: add nurse terminal login + home screen with API integration"
```

---

### Task 12: Nurse Terminal — Patient Detail + Verification + Ward Round Screens

**Files:**
- Create: `apps/nurse_terminal/lib/src/screens/patient_detail_screen.dart`
- Create: `apps/nurse_terminal/lib/src/screens/verification_screen.dart`
- Create: `apps/nurse_terminal/lib/src/screens/ward_round_screen.dart`
- Create: `apps/nurse_terminal/lib/src/screens/medication_screen.dart`
- Create: `apps/nurse_terminal/lib/src/screens/discharge_screen.dart`
- Create: `apps/nurse_terminal/lib/src/services/patient_service.dart`
- Create: `apps/nurse_terminal/lib/src/services/verification_service.dart`
- Create: `apps/nurse_terminal/lib/src/services/ward_round_service.dart`

**patient_service.dart** — Wraps patient CRUD:

```dart
class PatientService {
  final ApiClient api;
  PatientService(this.api);

  Future<List<dynamic>> listAdmitted() async { ... }
  Future<dynamic> getById(String id) async { ... }
  Future<void> discharge(String admissionId, String type, {String? notes}) async { ... }
}
```

**verification_service.dart** — Wraps verification records:

```dart
class VerificationService {
  final ApiClient api;
  VerificationService(this.api);

  Future<void> create(Map<String, dynamic> record) async { ... }
  Future<List<dynamic>> list({int page = 1, int pageSize = 50}) async { ... }
}
```

**ward_round_service.dart** — Wraps ward round entries:

```dart
class WardRoundService {
  final ApiClient api;
  WardRoundService(this.api);

  Future<void> create(String patientId, Map<String, dynamic> entry) async { ... }
  Future<List<dynamic>> list(String patientId) async { ... }
}
```

**patient_detail_screen.dart** — Shows full patient info + action buttons:
- Demographics: name, age, gender, blood type, allergies
- Admission: bed_no, department, diagnosis, admitted_at
- Wristband: device_id, bound_at, status
- Recent verifications list
- Action buttons row: [Scan Verification] [Start Ward Round] [Medication] [Discharge]

**verification_screen.dart** — BLE scan result display:
- Uses `MedicalWristbandService` to scan → read patient info
- Displays matched/unmatched/not_found result
- Shows patient name, admission number, bound time
- "Confirm" saves verification via `VerificationService.create()`

**ward_round_screen.dart** — Nursing round form:
- Fields: BP, HR, SpO2, Temperature, Weight
- Free-text notes
- Checkbox grid: Falls? Confusion? Pain? Appetite?
- Submits via `WardRoundService.create()`

**medication_screen.dart** — Today's medication schedule:
- Shows medications from API
- Each item has "Verify" button → triggers BLE scan → confirms identity → records

**discharge_screen.dart** — Discharge form:
- Dropdown: discharged / transferred / deceased
- Text field: notes
- Text field: transferred_to (shown only if transferred)
- On submit: calls `CompleteAdmission` API → clears wristband binding

- [ ] **Step 1: Create all 3 service files**

api_client.dart, patient_service.dart, verification_service.dart, ward_round_service.dart.

- [ ] **Step 2: Create all 5 screen files**

patient_detail_screen.dart, verification_screen.dart, ward_round_screen.dart, medication_screen.dart, discharge_screen.dart.

- [ ] **Step 3: Wire navigation in home_screen.dart**

Add onTap handlers for the 4 action buttons.

- [ ] **Step 4: Verify Flutter analyze**

Run: `cd apps/nurse_terminal && flutter analyze`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add apps/nurse_terminal/lib/src/screens/*.dart \
         apps/nurse_terminal/lib/src/services/patient_service.dart \
         apps/nurse_terminal/lib/src/services/verification_service.dart \
         apps/nurse_terminal/lib/src/services/ward_round_service.dart
git commit -m "feat: add nurse terminal detail/verification/wardround/medication/discharge screens"
```

---

### Task 13: Verification — Build All Subsystems

**Files:**
- `cloud/admin-api` — Go build
- `apps/admin-web` — Vite build
- `apps/nurse_terminal` — Flutter analyze

**Verification steps:**

1. Backend:
```bash
cd cloud/admin-api && go build ./...
```
Expected: clean build, no errors

2. Frontend:
```bash
cd apps/admin-web && npx vite build
```
Expected: clean build

3. Nurse terminal:
```bash
cd apps/nurse_terminal && flutter analyze
```
Expected: no errors

4. End-to-end smoke test:
- Start admin-api server
- POST `/api/v1/admin/medical/admissions` with patient data → verify 201 Created
- GET `/api/v1/admin/medical/patients/:id/ward-round` → verify empty list
- POST `/api/v1/admin/medical/patients/:id/ward-round` with vitals → verify 201 Created
- GET `/api/v1/admin/medical/patients/:id/ward-round` → verify list populated
- POST `/api/v1/admin/medical/admissions/:id/discharge` → verify patient status updated

- [ ] **Step 1: Run all 3 build commands**

- [ ] **Step 2: Run smoke test curl commands**

- [ ] **Step 3: Commit all remaining changes**

```bash
git add -A
git commit -m "feat: complete medical wristband clinical workflow end-to-end"
```

---

## Summary of Files Changed

| File | Change |
|------|--------|
| `cloud/admin-api/internal/model/model.go` | Add HospitalAdmission, WardRoundEntry, RegulatoryRuleResult structs |
| `cloud/admin-api/internal/store/store.go` | Add 7 new interface methods |
| `cloud/admin-api/internal/store/sqlite.go` | Add table migrations + 7 method implementations |
| `cloud/admin-api/internal/store/postgres.go` | Add 7 method implementations |
| `cloud/admin-api/internal/store/adapter.go` | Add 7 adapter routing methods |
| `cloud/admin-api/internal/handler/medical_wristband.go` | Add 4 handler methods |
| `cloud/admin-api/internal/router/router.go` | Add 5 route registrations |
| `apps/admin-web/src/api/medical.ts` | Add 9 API methods + 3 TS interfaces |
| `apps/admin-web/src/views/MedicalWristband.vue` | Add 3 new tabs (Admissions, Ward Rounds, Regulatory Alerts) |
| `apps/admin-web/src/views/AuditDetail.vue` | Replace mock data with real API calls |
| `apps/nurse_terminal/lib/main.dart` | Change home to LoginScreen |
| `apps/nurse_terminal/lib/src/services/api_client.dart` | NEW — HTTP client |
| `apps/nurse_terminal/lib/src/services/patient_service.dart` | NEW — patient CRUD |
| `apps/nurse_terminal/lib/src/services/verification_service.dart` | NEW — verification records |
| `apps/nurse_terminal/lib/src/services/ward_round_service.dart` | NEW — ward round entries |
| `apps/nurse_terminal/lib/src/screens/login_screen.dart` | NEW — login UI |
| `apps/nurse_terminal/lib/src/screens/home_screen.dart` | NEW — patient list |
| `apps/nurse_terminal/lib/src/screens/patient_detail_screen.dart` | NEW — patient info + actions |
| `apps/nurse_terminal/lib/src/screens/verification_screen.dart` | NEW — BLE scan result |
| `apps/nurse_terminal/lib/src/screens/ward_round_screen.dart` | NEW — vitals form |
| `apps/nurse_terminal/lib/src/screens/medication_screen.dart` | NEW — med schedule |
| `apps/nurse_terminal/lib/src/screens/discharge_screen.dart` | NEW — discharge form |

Total: ~22 files (3 modified existing admin-api, 4 modified existing, 15 new)
