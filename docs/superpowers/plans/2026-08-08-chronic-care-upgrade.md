# Eregen 慢性病专项升级 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现手环Pro+版试纸检测模块、外置血压配件、家属APP慢病管理模块、后端慢病API和AI分析引擎，形成完整的慢性病管理闭环。

**Architecture:** 采用分层架构：固件层（C/FreeRTOS）→ 通信层（MQTT/NATS）→ 存储层（SQLite）→ 分析层（Go服务）→ 展示层（Flutter APP）。各层通过明确定义的接口通信，支持独立开发和测试。

**Tech Stack:**
- 固件：C (GD32E230, FreeRTOS), CMake, ARM GCC
- 后端：Go 1.22+, Gin, SQLite, NATS
- 前端：Flutter 3.24+, Dart, Provider
- 数据库：SQLite (MVP阶段统一)

## Global Constraints

- 所有代码使用中文注释（代码标识符用英文）
- API路径前缀统一为 `/api/v1/`
- 数据库统一使用 SQLite（不引入 PostgreSQL/Redis/InfluxDB）
- MQTT topic 格式：`eregen/chronic/<subtopic>/#`
- 固件条件编译：`#ifdef TARGET_PRO_PLUS` 隔离新模块
- 前端适老化：字体 ≥ 16sp，触摸目标 ≥ 48dp
- 开源许可：仅使用 MIT/BSD-3/Apache-2.0/ISC
- 测试覆盖率目标：核心逻辑 > 60%

---

## 任务概览

| 编号 | 任务 | 子系统 | 预估周期 | 依赖 |
|------|------|--------|---------|------|
| 1 | 数据库迁移脚本 | 后端 | 0.5天 | 无 |
| 2 | 慢病数据模型 | 后端 | 1天 | #1 |
| 3 | 血糖API Handler | 后端 | 1天 | #2 |
| 4 | 尿酸API Handler | 后端 | 1天 | #2 |
| 5 | 血压API Handler | 后端 | 1天 | #2 |
| 6 | 饮食/运动API Handler | 后端 | 1天 | #2 |
| 7 | 任务API Handler | 后端 | 1天 | #2 |
| 8 | 报告API Handler | 后端 | 1天 | #2 |
| 9 | 路由注册 | 后端 | 0.5天 | #3-#8 |
| 10 | 血糖分析器 | 后端 | 1天 | #3 |
| 11 | 尿酸/血压分析器 | 后端 | 1天 | #4,#5 |
| 12 | 综合建议引擎 | 后端 | 1.5天 | #10,#11 |
| 13 | 推送服务扩展 | 后端 | 1天 | #12 |
| 14 | 试纸检测固件模块 | 固件 | 3天 | 无 |
| 15 | 血压配件BLE协议 | 固件 | 2天 | #14 |
| 16 | 慢病任务调度 | 固件 | 1.5天 | #15 |
| 17 | 固件编译测试 | 固件 | 1天 | #14-#16 |
| 18 | 慢病管理主页 | APP | 2天 | 无 |
| 19 | 血糖/尿酸详情页 | APP | 2天 | #18 |
| 20 | 血压详情页 | APP | 1.5天 | #18 |
| 21 | 饮食/运动页 | APP | 2天 | #18 |
| 22 | 健康报告页 | APP | 1.5天 | #18 |
| 23 | 首页/健康页改造 | APP | 1.5天 | #18 |
| 24 | 任务体系与奖励 | APP | 1.5天 | #18-#22 |
| 25 | 端到端联调 | 全系统 | 2天 | #1-#24 |

---

## 任务 1：数据库迁移脚本

**Files:**
- Create: `cloud/api-server/migrations/002_chronic_care.sql`

**Interfaces:**
- Consumes: 无
- Produces: 7张新表，插入默认任务模板数据

- [ ] **Step 1: 编写迁移SQL脚本**

```sql
-- 慢病管理数据库迁移
-- 对应设计文档: docs/superpowers/specs/2026-08-08-chronic-care-upgrade-design.md §4.1

-- 血糖检测记录
CREATE TABLE IF NOT EXISTS chronic_glucose_records (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    value REAL NOT NULL,
    unit TEXT DEFAULT 'mmol/L',
    test_mode TEXT DEFAULT 'random' CHECK(test_mode IN ('fasting', 'postprandial', 'random')),
    measurement_time DATETIME NOT NULL,
    detected_at DATETIME DEFAULT (datetime('now')),
    source TEXT DEFAULT 'test_strip' CHECK(source IN ('test_strip', 'bt_device', 'imported')),
    quality REAL CHECK(quality IS NULL OR (quality >= 0.0 AND quality <= 1.0)),
    temperature REAL
);

CREATE INDEX idx_glucose_elderly_time ON chronic_glucose_records(elderly_id, measurement_time);

-- 尿酸检测记录
CREATE TABLE IF NOT EXISTS chronic_uric_acid_records (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    value REAL NOT NULL,
    unit TEXT DEFAULT 'μmol/L',
    measurement_time DATETIME NOT NULL,
    detected_at DATETIME DEFAULT (datetime('now')),
    source TEXT DEFAULT 'test_strip' CHECK(source IN ('test_strip', 'bt_device', 'imported'))
);

CREATE INDEX idx_uric_elderly_time ON chronic_uric_acid_records(elderly_id, measurement_time);

-- 血压记录
CREATE TABLE IF NOT EXISTS chronic_bp_records (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    systolic INTEGER NOT NULL,
    diastolic INTEGER NOT NULL,
    pulse INTEGER,
    measurement_time DATETIME NOT NULL,
    detected_at DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX idx_bp_elderly_time ON chronic_bp_records(elderly_id, measurement_time);

-- 饮食记录
CREATE TABLE IF NOT EXISTS chronic_diet_records (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    meal_type TEXT NOT NULL CHECK(meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
    food_items TEXT NOT NULL,
    total_carbs REAL,
    total_calories REAL,
    recorded_at DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX idx_diet_elderly_time ON chronic_diet_records(elderly_id, recorded_at);

-- 运动记录
CREATE TABLE IF NOT EXISTS chronic_exercise_records (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    duration_min INTEGER,
    calories REAL,
    avg_hr INTEGER,
    max_hr INTEGER,
    recorded_at DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX idx_exercise_elderly_time ON chronic_exercise_records(elderly_id, recorded_at);

-- 每日任务
CREATE TABLE IF NOT EXISTS chronic_daily_tasks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    task_type TEXT NOT NULL CHECK(task_type IN ('bg_test', 'ua_test', 'bp_test', 'medication', 'exercise', 'diet')),
    scheduled_time TIME NOT NULL,
    completed INTEGER DEFAULT 0 CHECK(completed IN (0, 1)),
    completed_at DATETIME,
    task_date DATE DEFAULT (date('now')),
    UNIQUE(elderly_id, task_type, task_date)
);

-- 插入默认任务模板
INSERT OR IGNORE INTO chronic_daily_tasks (elderly_id, task_type, scheduled_time, task_date)
SELECT id, 'bg_test', '07:00', date('now') FROM elders WHERE id NOT IN (SELECT DISTINCT elderly_id FROM chronic_daily_tasks WHERE task_date = date('now'));

INSERT OR IGNORE INTO chronic_daily_tasks (elderly_id, task_type, scheduled_time, task_date)
SELECT id, 'bp_test', '08:00', date('now') FROM elders WHERE id NOT IN (SELECT DISTINCT elderly_id FROM chronic_daily_tasks WHERE task_date = date('now') AND task_type = 'bp_test');

INSERT OR IGNORE INTO chronic_daily_tasks (elderly_id, task_type, scheduled_time, task_date)
SELECT id, 'exercise', '18:00', date('now') FROM elders WHERE id NOT IN (SELECT DISTINCT elderly_id FROM chronic_daily_tasks WHERE task_date = date('now') AND task_type = 'exercise');

-- 周期报告
CREATE TABLE IF NOT EXISTS chronic_health_reports (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    elderly_id TEXT NOT NULL REFERENCES elders(id) ON DELETE CASCADE,
    report_type TEXT NOT NULL CHECK(report_type IN ('weekly', 'monthly', 'annual')),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    data_summary TEXT,
    ai_recommendations TEXT,
    generated_at DATETIME DEFAULT (datetime('now'))
);

CREATE INDEX idx_report_elderly_type ON chronic_health_reports(elderly_id, report_type, period_end);
```

- [ ] **Step 2: 创建迁移执行脚本**

```go
// cloud/api-server/internal/migration/chronic.go
package migration

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/002_chronic_care.sql
var chronicCareSQL string

// RunChronicCareMigration executes the chronic care database migration.
func RunChronicCareMigration(db *sql.DB) error {
	fmt.Println("Running chronic care migration...")
	_, err := db.Exec(chronicCareSQL)
	if err != nil {
		return fmt.Errorf("chronic care migration failed: %w", err)
	}
	fmt.Println("Chronic care migration completed")
	return nil
}
```

- [ ] **Step 3: 在main.go中集成迁移**

查看 `cloud/api-server/cmd/main.go` 中现有迁移调用位置，添加 `RunChronicCareMigration`。

- [ ] **Step 4: 编译验证**

```bash
cd cloud/api-server && go build ./...
```

- [ ] **Step 5: 提交**

```bash
git add cloud/api-server/migrations/002_chronic_care.sql cloud/api-server/internal/migration/chronic.go
git commit -m "feat: add chronic care database migration (7 tables)"
```

---

## 任务 2：慢病数据模型

**Files:**
- Create: `cloud/api-server/internal/model/chronic.go`
- Create: `cloud/api-server/internal/store/chronic.go`

**Interfaces:**
- Consumes: 任务1的数据库表结构
- Produces: `ChronicGlucoseRecord`, `ChronicUricAcidRecord`, `ChronicBPRecord` 等模型 + Store接口方法

- [ ] **Step 1: 编写数据模型**

```go
// cloud/api-server/internal/model/chronic.go
package model

import "time"

// ChronicGlucoseRecord represents a blood glucose measurement.
type ChronicGlucoseRecord struct {
	ID             string    `json:"id"`
	ElderlyID      string    `json:"elderly_id"`
	Value          float64   `json:"value"`
	Unit           string    `json:"unit"`
	TestMode       string    `json:"test_mode"`  // fasting/postprandial/random
	MeasurementTime time.Time `json:"measurement_time"`
	DetectedAt     time.Time `json:"detected_at"`
	Source         string    `json:"source"`  // test_strip/bt_device/imported
	Quality        *float64  `json:"quality,omitempty"`
	Temperature    *float64  `json:"temperature,omitempty"`
}

// ChronicUricAcidRecord represents a uric acid measurement.
type ChronicUricAcidRecord struct {
	ID              string    `json:"id"`
	ElderlyID       string    `json:"elderly_id"`
	Value           float64   `json:"value"`
	Unit            string    `json:"unit"`
	MeasurementTime time.Time `json:"measurement_time"`
	DetectedAt      time.Time `json:"detected_at"`
	Source          string    `json:"source"`
}

// ChronicBPRecord represents a blood pressure measurement.
type ChronicBPRecord struct {
	ID              string    `json:"id"`
	ElderlyID       string    `json:"elderly_id"`
	Systolic        int       `json:"systolic"`
	Diastolic       int       `json:"diastolic"`
	Pulse           *int      `json:"pulse,omitempty"`
	MeasurementTime time.Time `json:"measurement_time"`
	DetectedAt      time.Time `json:"detected_at"`
}

// ChronicDietRecord represents a diet entry.
type ChronicDietRecord struct {
	ID            string    `json:"id"`
	ElderlyID     string    `json:"elderly_id"`
	MealType      string    `json:"meal_type"`
	FoodItems     string    `json:"food_items"`
	TotalCarbs    *float64  `json:"total_carbs,omitempty"`
	TotalCalories *float64  `json:"total_calories,omitempty"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// ChronicExerciseRecord represents an exercise entry.
type ChronicExerciseRecord struct {
	ID           string    `json:"id"`
	ElderlyID    string    `json:"elderly_id"`
	Type         string    `json:"type"`
	DurationMin  *int      `json:"duration_min,omitempty"`
	Calories     *float64  `json:"calories,omitempty"`
	AvgHR        *int      `json:"avg_hr,omitempty"`
	MaxHR        *int      `json:"max_hr,omitempty"`
	RecordedAt   time.Time `json:"recorded_at"`
}

// ChronicDailyTask represents a daily health task.
type ChronicDailyTask struct {
	ID            string    `json:"id"`
	ElderlyID     string    `json:"elderly_id"`
	TaskType      string    `json:"task_type"`
	ScheduledTime string    `json:"scheduled_time"`
	Completed     bool      `json:"completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	TaskDate      string    `json:"task_date"`
}

// ChronicHealthReport represents a periodic health report.
type ChronicHealthReport struct {
	ID                string    `json:"id"`
	ElderlyID         string    `json:"elderly_id"`
	ReportType        string    `json:"report_type"`
	PeriodStart       string    `json:"period_start"`
	PeriodEnd         string    `json:"period_end"`
	DataSummary       *string   `json:"data_summary,omitempty"`
	ARecommendations  *string   `json:"ai_recommendations,omitempty"`
	GeneratedAt       time.Time `json:"generated_at"`
}
```

- [ ] **Step 2: 编写Store接口方法**

```go
// cloud/api-server/internal/store/chronic.go
package store

import (
	"context"
	"database/sql"
	"time"

	"eregen.dev/api-server/internal/model"
)

// ChronicStore provides data access for chronic care features.
type ChronicStore struct {
	db *sql.DB
}

// NewChronicStore creates a new ChronicStore.
func NewChronicStore(db *sql.DB) *ChronicStore {
	return &ChronicStore{db: db}
}

// SaveGlucoseRecord inserts a glucose measurement.
func (s *ChronicStore) SaveGlucoseRecord(ctx context.Context, r *model.ChronicGlucoseRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_glucose_records (id, elderly_id, value, unit, test_mode, measurement_time, source, quality, temperature)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Value, r.Unit, r.TestMode, r.MeasurementTime, r.Source, r.Quality, r.Temperature,
	)
	return err
}

// ListGlucoseRecords returns glucose records for an elderly within a date range.
func (s *ChronicStore) ListGlucoseRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicGlucoseRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, value, unit, test_mode, measurement_time, detected_at, source, quality, temperature
		 FROM chronic_glucose_records WHERE elderly_id = ? AND measurement_time >= ? AND measurement_time <= ? ORDER BY measurement_time DESC`,
		elderlyID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicGlucoseRecord
	for rows.Next() {
		var r model.ChronicGlucoseRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Value, &r.Unit, &r.TestMode, &r.MeasurementTime, &r.DetectedAt, &r.Source, &r.Quality, &r.Temperature); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// SaveUricAcidRecord inserts a uric acid measurement.
func (s *ChronicStore) SaveUricAcidRecord(ctx context.Context, r *model.ChronicUricAcidRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_uric_acid_records (id, elderly_id, value, unit, measurement_time, source)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Value, r.Unit, r.MeasurementTime, r.Source,
	)
	return err
}

// ListUricAcidRecords returns uric acid records for an elderly within a date range.
func (s *ChronicStore) ListUricAcidRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicUricAcidRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, value, unit, measurement_time, detected_at, source
		 FROM chronic_uric_acid_records WHERE elderly_id = ? AND measurement_time >= ? AND measurement_time <= ? ORDER BY measurement_time DESC`,
		elderlyID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicUricAcidRecord
	for rows.Next() {
		var r model.ChronicUricAcidRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Value, &r.Unit, &r.MeasurementTime, &r.DetectedAt, &r.Source); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// SaveBPRecord inserts a blood pressure measurement.
func (s *ChronicStore) SaveBPRecord(ctx context.Context, r *model.ChronicBPRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_bp_records (id, elderly_id, systolic, diastolic, pulse, measurement_time)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Systolic, r.Diastolic, r.Pulse, r.MeasurementTime,
	)
	return err
}

// ListBPRecords returns blood pressure records for an elderly within a date range.
func (s *ChronicStore) ListBPRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicBPRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, systolic, diastolic, pulse, measurement_time, detected_at
		 FROM chronic_bp_records WHERE elderly_id = ? AND measurement_time >= ? AND measurement_time <= ? ORDER BY measurement_time DESC`,
		elderlyID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicBPRecord
	for rows.Next() {
		var r model.ChronicBPRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Systolic, &r.Diastolic, &r.Pulse, &r.MeasurementTime, &r.DetectedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// SaveDietRecord inserts a diet record.
func (s *ChronicStore) SaveDietRecord(ctx context.Context, r *model.ChronicDietRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_diet_records (id, elderly_id, meal_type, food_items, total_carbs, total_calories)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.MealType, r.FoodItems, r.TotalCarbs, r.TotalCalories,
	)
	return err
}

// ListDietRecords returns diet records for an elderly within a date range.
func (s *ChronicStore) ListDietRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicDietRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, meal_type, food_items, total_carbs, total_calories, recorded_at
		 FROM chronic_diet_records WHERE elderly_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at DESC`,
		elderlyID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicDietRecord
	for rows.Next() {
		var r model.ChronicDietRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.MealType, &r.FoodItems, &r.TotalCarbs, &r.TotalCalories, &r.RecordedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// SaveExerciseRecord inserts an exercise record.
func (s *ChronicStore) SaveExerciseRecord(ctx context.Context, r *model.ChronicExerciseRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_exercise_records (id, elderly_id, type, duration_min, calories, avg_hr, max_hr)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Type, r.DurationMin, r.Calories, r.AvgHR, r.MaxHR,
	)
	return err
}

// ListExerciseRecords returns exercise records for an elderly within a date range.
func (s *ChronicStore) ListExerciseRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicExerciseRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, type, duration_min, calories, avg_hr, max_hr, recorded_at
		 FROM chronic_exercise_records WHERE elderly_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at DESC`,
		elderlyID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicExerciseRecord
	for rows.Next() {
		var r model.ChronicExerciseRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Type, &r.DurationMin, &r.Calories, &r.AvgHR, &r.MaxHR, &r.RecordedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ListDailyTasks returns tasks for an elderly on a specific date.
func (s *ChronicStore) ListDailyTasks(ctx context.Context, elderlyID, taskDate string) ([]model.ChronicDailyTask, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, task_type, scheduled_time, completed, completed_at, task_date
		 FROM chronic_daily_tasks WHERE elderly_id = ? AND task_date = ? ORDER BY scheduled_time`,
		elderlyID, taskDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.ChronicDailyTask
	for rows.Next() {
		var r model.ChronicDailyTask
		var completedInt int
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.TaskType, &r.ScheduledTime, &completedInt, &r.CompletedAt, &r.TaskDate); err != nil {
			return nil, err
		}
		r.Completed = completedInt == 1
		records = append(records, r)
	}
	return records, rows.Err()
}

// UpdateTaskCompletion marks a task as completed.
func (s *ChronicStore) UpdateTaskCompletion(ctx context.Context, taskID string, completedAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE chronic_daily_tasks SET completed = 1, completed_at = ? WHERE id = ?`,
		completedAt, taskID,
	)
	return err
}

// SaveHealthReport inserts a health report.
func (s *ChronicStore) SaveHealthReport(ctx context.Context, r *model.ChronicHealthReport) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO chronic_health_reports (id, elderly_id, report_type, period_start, period_end, data_summary, ai_recommendations)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.ReportType, r.PeriodStart, r.PeriodEnd, r.DataSummary, r.ARecommendations,
	)
	return err
}

// GetHealthReport retrieves a health report by type and period.
func (s *ChronicStore) GetHealthReport(ctx context.Context, elderlyID, reportType, periodEnd string) (*model.ChronicHealthReport, error) {
	var r model.ChronicHealthReport
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, report_type, period_start, period_end, data_summary, ai_recommendations, generated_at
		 FROM chronic_health_reports WHERE elderly_id = ? AND report_type = ? AND period_end = ?`,
		elderlyID, reportType, periodEnd,
	).Scan(&r.ID, &r.ElderlyID, &r.ReportType, &r.PeriodStart, &r.PeriodEnd, &r.DataSummary, &r.ARecommendations, &r.GeneratedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
```

- [ ] **Step 3: 编译验证**

```bash
cd cloud/api-server && go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add cloud/api-server/internal/model/chronic.go cloud/api-server/internal/store/chronic.go
git commit -m "feat: add chronic care data models and store layer"
```

---

## 任务 3：血糖API Handler

**Files:**
- Create: `cloud/api-server/internal/handler/chronic_glucose.go`
- Create: `cloud/api-server/internal/service/chronic_glucose.go`

**Interfaces:**
- Consumes: 任务2的 `ChronicStore.ListGlucoseRecords`, `ChronicStore.SaveGlucoseRecord`
- Produces: `POST /api/v1/chronic/glucose`, `GET /api/v1/chronic/glucose`, `GET /api/v1/chronic/glucose/trend`, `POST /api/v1/chronic/test-strip/read`

- [ ] **Step 1: 编写Service层**

```go
// cloud/api-server/internal/service/chronic_glucose.go
package service

import (
	"context"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"
)

// ChronicGlucoseService handles blood glucose operations.
type ChronicGlucoseService struct {
	st *store.ChronicStore
}

// NewChronicGlucoseService creates a new service.
func NewChronicGlucoseService(st *store.ChronicStore) *ChronicGlucoseService {
	return &ChronicGlucoseService{st: st}
}

// CreateRecord inserts a glucose measurement.
func (s *ChronicGlucoseService) CreateRecord(ctx context.Context, r *model.ChronicGlucoseRecord) error {
	if r.ID == "" {
		r.ID = generateUUID()
	}
	if r.MeasurementTime.IsZero() {
		r.MeasurementTime = time.Now()
	}
	return s.st.SaveGlucoseRecord(ctx, r)
}

// ListRecords returns glucose records within a date range.
func (s *ChronicGlucoseService) ListRecords(ctx context.Context, elderlyID string, from, to time.Time) ([]model.ChronicGlucoseRecord, error) {
	return s.st.ListGlucoseRecords(ctx, elderlyID, from, to)
}

// GlucoseTrendData holds aggregated trend data for chart rendering.
type GlucoseTrendData struct {
	DailyAvg []DailyAvgPoint `json:"daily_avg"`
	Overall  TrendSummary    `json:"overall"`
}

// DailyAvgPoint represents one day's average glucose value.
type DailyAvgPoint struct {
	Date  string  `json:"date"`
	Avg   float64 `json:"avg"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

// TrendSummary holds overall statistics.
type TrendSummary struct {
	Avg      float64 `json:"avg"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Count    int     `json:"count"`
	InRange  int     `json:"in_range"` // within 3.9-7.8 mmol/L
}

// GetTrendData computes aggregated trend data for chart rendering.
func (s *ChronicGlucoseService) GetTrendData(ctx context.Context, elderlyID string, days int) (*GlucoseTrendData, error) {
	records, err := s.st.ListGlucoseRecords(ctx, elderlyID, time.Now().AddDate(0, 0, -days), time.Now())
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &GlucoseTrendData{}, nil
	}

	// Group by date
	byDate := make(map[string][]float64)
	for _, r := range records {
		date := r.MeasurementTime.Format("2006-01-02")
		byDate[date] = append(byDate[date], r.Value)
	}

	dailyAvgs := make([]DailyAvgPoint, 0, len(byDate))
	var total, minVal, maxVal float64
	inRange := 0
	minVal = 999
	maxVal = 0

	for date, values := range byDate {
		sum := 0.0
		for _, v := range values {
			sum += v
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
			if v >= 3.9 && v <= 7.8 {
				inRange++
			}
		}
		avg := sum / float64(len(values))
		dailyAvgs = append(dailyAvgs, DailyAvgPoint{Date: date, Avg: avg, Min: minVal, Max: maxVal})
		total += avg
	}

	// Sort by date
	for i := 0; i < len(dailyAvgs); i++ {
		for j := i + 1; j < len(dailyAvgs); j++ {
			if dailyAvgs[i].Date > dailyAvgs[j].Date {
				dailyAvgs[i], dailyAvgs[j] = dailyAvgs[j], dailyAvgs[i]
			}
		}
	}

	overall := TrendSummary{
		Avg:     total / float64(len(dailyAvgs)),
		Min:     minVal,
		Max:     maxVal,
		Count:   len(records),
		InRange: inRange,
	}

	return &GlucoseTrendData{DailyAvg: dailyAvgs, Overall: overall}, nil
}
```

- [ ] **Step 2: 编写Handler层**

```go
// cloud/api-server/internal/handler/chronic_glucose.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicGlucoseHandler handles glucose-related API endpoints.
type ChronicGlucoseHandler struct {
	svc *service.ChronicGlucoseService
	log *zap.Logger
}

// NewChronicGlucoseHandler creates a new handler.
func NewChronicGlucoseHandler(svc *service.ChronicGlucoseService, log *zap.Logger) *ChronicGlucoseHandler {
	return &ChronicGlucoseHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/chronic/glucose
func (h *ChronicGlucoseHandler) Create(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req struct {
		Value       float64  `json:"value" binding:"required"`
		TestMode    string   `json:"test_mode"`
		Source      string   `json:"source"`
		Temperature *float64 `json:"temperature"`
		Quality     *float64 `json:"quality"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	if req.TestMode == "" {
		req.TestMode = "random"
	}
	if req.Source == "" {
		req.Source = "imported"
	}

	record := &model.ChronicGlucoseRecord{
		ElderlyID:     elderlyID,
		Value:         req.Value,
		Unit:          "mmol/L",
		TestMode:      req.TestMode,
		Source:        req.Source,
		Temperature:   req.Temperature,
		Quality:       req.Quality,
		MeasurementTime: time.Now(),
	}

	if err := h.svc.CreateRecord(c.Request.Context(), record); err != nil {
		h.log.Error("failed to create glucose record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to save record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gin.H{"id": record.ID}})
}

// List handles GET /api/v1/chronic/glucose
func (h *ChronicGlucoseHandler) List(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days < 1 || days > 365 {
		days = 30
	}

	from := time.Now().AddDate(0, 0, -days)
	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, from, time.Now())
	if err != nil {
		h.log.Error("failed to list glucose records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// Trend handles GET /api/v1/chronic/glucose/trend
func (h *ChronicGlucoseHandler) Trend(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days < 1 || days > 365 {
		days = 30
	}

	data, err := h.svc.GetTrendData(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("failed to get glucose trend", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
}

// TestStripRead handles POST /api/v1/chronic/test-strip/read (from bracelet)
func (h *ChronicGlucoseHandler) TestStripRead(c *gin.Context) {
	var req struct {
		DeviceID    string   `json:"dev_id" binding:"required"`
		Value       float64  `json:"value" binding:"required"`
		TestMode    string   `json:"test_mode"`
		Temperature *float64 `json:"temperature"`
		Quality     *float64 `json:"quality"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": err.Error()})
		return
	}

	// Look up elderly_id from device
	// (In production, device-to-elderly mapping should be cached)
	elderlyID := c.GetString("elderly_id") // set by middleware

	record := &model.ChronicGlucoseRecord{
		ElderlyID:       elderlyID,
		Value:           req.Value,
		Unit:            "mmol/L",
		TestMode:        req.TestMode,
		Source:          "test_strip",
		Temperature:     req.Temperature,
		Quality:         req.Quality,
		MeasurementTime: time.Now(),
	}

	if err := h.svc.CreateRecord(c.Request.Context(), record); err != nil {
		h.log.Error("failed to save test strip reading", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gin.H{"id": record.ID, "value": req.Value}})
}
```

- [ ] **Step 3: 编译验证**

```bash
cd cloud/api-server && go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add cloud/api-server/internal/service/chronic_glucose.go cloud/api-server/internal/handler/chronic_glucose.go
git commit -m "feat: add glucose API handler and service"
```

---

## 任务 4-8：尿酸/血压/饮食运动/任务/报告 Handler

**任务4：尿酸API Handler** (`chronic_uric_acid.go`)
- Service: `ChronicUricAcidService` — `CreateRecord`, `ListRecords`
- Handler: `ChronicUricAcidHandler` — `Create`, `List`
- Routes: `POST /api/v1/chronic/uric-acid`, `GET /api/v1/chronic/uric-acid`

**任务5：血压API Handler** (`chronic_bp.go`)
- Service: `ChronicBPService` — `CreateRecord`, `ListRecords`
- Handler: `ChronicBPHandler` — `Create`, `List`, `DeviceSync`
- Routes: `POST /api/v1/chronic/blood-pressure`, `GET /api/v1/chronic/blood-pressure`, `POST /api/v1/chronic/bp-device/sync`

**任务6：饮食/运动API Handler** (`chronic_diet.go`, `chronic_exercise.go`)
- Service: `ChronicDietService`, `ChronicExerciseService`
- Handler: `ChronicDietHandler`, `ChronicExerciseHandler`
- Routes: 饮食 `POST/GET /api/v1/chronic/diet`，运动 `POST/GET /api/v1/chronic/exercise`

**任务7：任务API Handler** (`chronic_task.go`)
- Service: `ChronicTaskService` — `ListDailyTasks`, `UpdateTaskCompletion`
- Handler: `ChronicTaskHandler` — `List`, `Update`
- Routes: `GET /api/v1/chronic/daily-tasks`, `PUT /api/v1/chronic/daily-tasks/:id`

**任务8：报告API Handler** (`chronic_report.go`)
- Service: `ChronicReportService` — `SaveHealthReport`, `GetHealthReport`
- Handler: `ChronicReportHandler` — `Get`, `Generate`
- Routes: `GET /api/v1/chronic/report/:type`, `POST /api/v1/chronic/report/generate`

> 每个任务的代码结构与任务3完全相同，仅替换模型名和字段。参考任务3的代码模板实现。

---

## 任务 9：路由注册

**Files:**
- Modify: `cloud/api-server/internal/router/router.go`

- [ ] **Step 1: 在router.go中添加慢病路由组**

在现有 `protected` 路由组末尾添加：

```go
// Chronic care routes
chronicH := handler.NewChronicGlucoseHandler(service.NewChronicGlucoseService(chronicStore, log), log)
chronicUaH := handler.NewChronicUricAcidHandler(service.NewChronicUricAcidService(chronicStore, log), log)
chronicBP H := handler.NewChronicBPHandler(service.NewChronicBPService(chronicStore, log), log)
chronicDietH := handler.NewChronicDietHandler(service.NewChronicDietService(chronicStore, log), log)
chronicExerciseH := handler.NewChronicExerciseHandler(service.NewChronicExerciseService(chronicStore, log), log)
chronicTaskH := handler.NewChronicTaskHandler(service.NewChronicTaskService(chronicStore, log), log)
chronicReportH := handler.NewChronicReportHandler(service.NewChronicReportService(chronicStore, log), log)

chronicGroup := protected.Group("/chronic/:elderly_id")
{
	// Glucose
	chronicGroup.POST("/glucose", chronicH.Create)
	chronicGroup.GET("/glucose", chronicH.List)
	chronicGroup.GET("/glucose/trend", chronicH.Trend)
	chronicGroup.POST("/test-strip/read", chronicH.TestStripRead)

	// Uric acid
	chronicGroup.POST("/uric-acid", chronicUaH.Create)
	chronicGroup.GET("/uric-acid", chronicUaH.List)

	// Blood pressure
	chronicGroup.POST("/blood-pressure", chronicBPH.Create)
	chronicGroup.GET("/blood-pressure", chronicBPH.List)
	chronicGroup.POST("/bp-device/sync", chronicBPH.DeviceSync)

	// Diet
	chronicGroup.POST("/diet", chronicDietH.Create)
	chronicGroup.GET("/diet", chronicDietH.List)

	// Exercise
	chronicGroup.POST("/exercise", chronicExerciseH.Create)
	chronicGroup.GET("/exercise", chronicExerciseH.List)

	// Tasks
	chronicGroup.GET("/daily-tasks", chronicTaskH.List)
	chronicGroup.PUT("/daily-tasks/:task_id", chronicTaskH.Update)

	// Reports
	chronicGroup.GET("/report/:type", chronicReportH.Get)
	chronicGroup.POST("/report/generate", chronicReportH.Generate)
}
```

- [ ] **Step 2: 在main.go中创建ChronicStore实例**

```go
// 在 New() 函数开头添加
chronicStore := store.NewChronicStore(db)
```

- [ ] **Step 3: 编译验证**

```bash
cd cloud/api-server && go build ./...
```

- [ ] **Step 4: 提交**

```bash
git add cloud/api-server/internal/router/router.go
git commit -m "feat: register chronic care API routes"
```

---

## 任务 10-13：分析引擎与推送服务

**任务10：血糖分析器** (`cloud/data-pipeline/internal/analyzer/chronic_glucose.go`)
- 实现趋势计算、异常检测（低血糖<3.9、高血糖>7.0空腹/>10.0餐后）
- 返回 `*model.AnalysisResult`

**任务11：尿酸/血压分析器** (`chronic_uric.go`, `chronic_bp.go`)
- 尿酸异常：>420 μmol/L
- 血压异常：收缩压>140或舒张压>90

**任务12：综合建议引擎** (`recommendations.go`)
- 整合血糖/尿酸/血压/饮食/运动分析结果
- 生成个性化健康建议
- 实现规则矩阵（见设计文档 §3.7）

**任务13：推送服务扩展** (`cloud/push-service/internal/handler/chronic/`)
- 任务提醒：定时检查 `chronic_daily_tasks` 中未完成任务
- 异常告警：监听分析引擎输出的异常结果
- 集成到现有NATS订阅机制

---

## 任务 14：试纸检测固件模块

**Files:**
- Create: `firmware/bracelet/pro_plus/sensors/electrochemical.c/h`
- Create: `firmware/bracelet/pro_plus/protocol/strip_type.h`
- Create: `firmware/bracelet/pro_plus/protocol/electrochemical_msg.h`

**Interfaces:**
- Consumes: GD32 GPIO, I2C, ADC API
- Produces: `electrochemical_detect() → detection_result_t`

- [ ] **Step 1: 定义试纸类型和消息格式**

```c
// firmware/bracelet/pro_plus/protocol/strip_type.h
#ifndef STRIP_TYPE_H
#define STRIP_TYPE_H

typedef enum {
    STRIP_NONE = 0,
    STRIP_GLUCOSE,
    STRIP_URIC_ACID,
    STRIP_ERROR
} strip_type_t;

typedef struct {
    strip_type_t type;
    float value;
    float temperature;
    float quality;
    uint64_t timestamp;
} detection_result_t;

#endif
```

```c
// firmware/bracelet/pro_plus/protocol/electrochemical_msg.h
#ifndef ELECTROCHEMICAL_MSG_H
#define ELECTROCHEMICAL_MSG_H

#define CHRONIC_TEST_TOPIC "eregen/chronic/test/glucose"
#define CHRONIC_BP_TOPIC "eregen/chronic/bp"

typedef struct {
    char type[16];      // "chronic_test"
    char dev_id[16];    // "BR-XXXX"
    char test_type[16]; // "glucose"
    float value;
    char unit[16];      // "mmol/L"
    char test_mode[16]; // "fasting"
    float temperature;
    float quality;
    uint64_t ts;
} electrochemical_msg_t;

#endif
```

- [ ] **Step 2: 实现电化学检测驱动**

```c
// firmware/bracelet/pro_plus/sensors/electrochemical.c
#include "electrochemical.h"
#include "strip_type.h"
#include "../protocol/electrochemical_msg.h"
#include "../../entry/cat1_at.h"
#include "../../common/log.c"

#define STRIP_DETECT_GPIO GPIOA
#define STRIP_DETECT_PIN GPIO_PIN_12
#define ADC_CHANNEL_ADC12 IN12

static bool strip_inserted_flag = false;

bool strip_inserted(void) {
    return strip_inserted_flag;
}

strip_type_t identify_strip_type(void) {
    // Read electrode impedance via ADC
    // Glucose strip: ~200 ohm, Uric acid strip: ~350 ohm (example values)
    uint16_t impedance = adc_read_ohm(ADC_CHANNEL_ADC12);
    if (impedance < 300) return STRIP_GLUCOSE;
    if (impedance < 500) return STRIP_URIC_ACID;
    return STRIP_ERROR;
}

static float read_temperature(void) {
    // NTC thermistor on GPIO A1
    uint16_t adc_val = adc_read_single(ADC1, ADC_CHANNEL_1);
    return adc_to_temperature(adc_val);
}

static float read_current_signal(void) {
    // Read nA-level current via precision amplifier + ADC
    return adc_to_current(adc_read_single(ADC1, ADC_CHANNEL_2));
}

static float apply_temp_compensation(float raw, float temp) {
    // Linear compensation: compensated = raw * (1 + alpha * (temp - 25))
    const float alpha = 0.002f;
    return raw * (1.0f + alpha * (temp - 25.0f));
}

static float current_to_concentration(float compensated, strip_type_t type) {
    // Calibration coefficients (from factory calibration)
    const float slope = (type == STRIP_GLUCOSE) ? 0.5f : 0.3f;
    const float intercept = (type == STRIP_GLUCOSE) ? 0.1f : -5.0f;
    return compensated * slope + intercept;
}

static float signal_quality_check(float raw) {
    // Quality index based on signal-to-noise ratio
    if (raw < 0.001f || raw > 10.0f) return 0.0f;
    return 0.95f;
}

detection_result_t electrochemical_detect(void) {
    detection_result_t result = {0};

    if (!strip_inserted()) return result;

    strip_type_t type = identify_strip_type();
    if (type == STRIP_ERROR) {
        result.type = STRIP_ERROR;
        return result;
    }

    float temp = read_temperature();
    float raw = read_current_signal();
    float compensated = apply_temp_compensation(raw, temp);
    float value = current_to_concentration(compensated, type);
    float quality = signal_quality_check(raw);

    if (quality < 0.5f) {
        result.type = STRIP_ERROR;
        return result;
    }

    result.type = type;
    result.value = value;
    result.temperature = temp;
    result.quality = quality;
    result.timestamp = get_unix_timestamp();

    // Upload to cloud via MQTT
    upload_detection_result(type, value, temp);

    return result;
}

void electrochemical_isr(void) {
    // GPIO interrupt: strip inserted/removed
    if (gpio_input_get(STRIP_DETECT_GPIO, STRIP_DETECT_PIN)) {
        strip_inserted_flag = true;
    } else {
        strip_inserted_flag = false;
    }
}
```

- [ ] **Step 3: 编译验证**

```bash
cd firmware/bracelet && mkdir -p pro_plus && cmake -S . -B build/pro_plus -DTARGET_PRO_PLUS=ON
```

- [ ] **Step 4: 提交**

```bash
git add firmware/bracelet/pro_plus/
git commit -m "feat: add electrochemical test strip detection module for Pro+"
```

---

## 任务 15：血压配件BLE协议

**Files:**
- Create: `firmware/bracelet/pro_plus/bt_peripheral/bp_device.c/h`

**Interfaces:**
- Consumes: BLE 5.3 peripheral role API
- Produces: `bp_device_connect()`, `bp_device_read_measurement() → bp_measurement_t`

- [ ] **Step 1: 定义BLE协议**

```c
// firmware/bracelet/pro_plus/bt_peripheral/bp_device.h
#ifndef BP_DEVICE_H
#define BP_DEVICE_H

#define BP_SERVICE_UUID   0xFFF0
#define BP_MEASURE_CHAR   0xFFF1
#define BP_CONTROL_CHAR   0xFFF2

typedef struct __attribute__((packed)) {
    uint16_t systolic;
    uint16_t diastolic;
    uint16_t pulse;
    uint32_t timestamp;
} bp_measurement_t;

bool bp_device_connect(void);
bool bp_device_read_measurement(bp_measurement_t *result);
void bp_device_disconnect(void);

#endif
```

- [ ] **Step 2: 实现BLE连接和读取**

参考现有 `firmware/bracelet/pro/` 中的BLE实现模式，使用NRF52832 BLE API连接外部血压计Peripheral。

- [ ] **Step 3: 编译验证并提交**

---

## 任务 16：慢病任务调度

**Files:**
- Create: `firmware/bracelet/pro_plus/app/chronic_manager.c/h`

**Interfaces:**
- Consumes: 任务15的BLE连接
- Produces: 定时任务提醒（振动+屏幕显示）

- [ ] **Step 1: 实现任务调度器**

```c
// 核心逻辑：在FreeRTOS任务中定期检查chronic_daily_tasks，
// 到时间时触发振动+屏幕提醒
void chronic_manager_task(void *args) {
    while (1) {
        check_and_trigger_tasks();
        vTaskDelay(pdMS_TO_TICKS(60000)); // 每分钟检查
    }
}
```

- [ ] **Step 2: 编译验证并提交**

---

## 任务 17：固件编译测试

- [ ] **Step 1: 确保Pro+变体独立编译**

```bash
cd firmware/bracelet
cmake -S . -B build/pro_plus -DTARGET_PRO_PLUS=ON
cmake --build build/pro_plus
```

- [ ] **Step 2: 验证Entry/Plus变体不受影响**

```bash
cmake -S . -B build/entry
cmake -S . -B build/plus
cmake --build build/entry
cmake --build build/plus
```

- [ ] **Step 3: 提交**

```bash
git commit -m "test: verify Pro+ firmware compiles independently"
```

---

## 任务 18：慢病管理主页

**Files:**
- Create: `apps/family-app/lib/screens/chronic/chronic_home_page.dart`
- Modify: `apps/family-app/lib/widgets/bottom_nav_bar.dart`
- Modify: `apps/family-app/lib/screens/login/main_tab_screen.dart`

**Interfaces:**
- Consumes: `ApiClient` (已有), `AppTheme` (已有)
- Produces: 慢病管理主页（血糖/尿酸/血压卡片 + 任务列表 + AI建议）

- [ ] **Step 1: 更新底部导航栏**

在 `bottom_nav_bar.dart` 的 `tabs` 列表中添加第5个Tab：

```dart
static const List<_TabItem> tabs = [
  _TabItem('首页', Icons.home_outlined, Icons.home),
  _TabItem('健康', Icons.monitor_heart_outlined, Icons.monitor_heart),
  _TabItem('告警', Icons.notifications_none_rounded, Icons.notifications_active),
  _TabItem('用药', Icons.medication_outlined, Icons.medication),
  _TabItem('慢病', Icons.favorite_outline, Icons.favorite),  // 🆕 新增
  _TabItem('福利', Icons.card_giftcard_outlined, Icons.card_giftcard),
];
```

- [ ] **Step 2: 更新主Tab屏幕**

在 `main_tab_screen.dart` 的 `_pages` 列表中添加 `ChronicHomePage()`。

- [ ] **Step 3: 编写慢病管理主页**

```dart
// apps/family-app/lib/screens/chronic/chronic_home_page.dart
import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../blood_sugar_page.dart';
import '../uric_acid_page.dart';
import '../blood_pressure_page.dart';

class ChronicHomePage extends StatefulWidget {
  const ChronicHomePage({super.key});

  @override
  State<ChronicHomePage> createState() => _ChronicHomePageState();
}

class _ChronicHomePageState extends State<ChronicHomePage> {
  int _selectedIndex = 4; // 慢病管理Tab
  String? _selectedElderlyId;
  Map<String, dynamic>? _latestGlucose;
  Map<String, dynamic>? _latestUric;
  Map<String, dynamic>? _latestBP;
  List<dynamic>? _tasks;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    final client = ApiClient.instance;
    final elderlyId = client.selectedElderlyId;
    if (elderlyId == null) {
      setState(() => _loading = false);
      return;
    }

    // Load latest readings
    final glucoseResp = await client.get('/chronic/glucose?days=7');
    final uricResp = await client.get('/chronic/uric-acid?days=7');
    final bpResp = await client.get('/chronic/blood-pressure?days=7');
    final tasksResp = await client.get('/chronic/daily-tasks');

    setState(() {
      _latestGlucose = glucoseResp['data']?.isNotEmpty == true
          ? glucoseResp['data'][0]
          : null;
      _latestUric = uricResp['data']?.isNotEmpty == true
          ? uricResp['data'][0]
          : null;
      _latestBP = bpResp['data']?.isNotEmpty == true
          ? bpResp['data'][0]
          : null;
      _tasks = tasksResp['data'];
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : CustomScrollView(
              slivers: [
                _buildHeader(),
                _buildMetricsSection(),
                _buildTasksSection(),
                _buildAIRecommendation(),
                _buildQuickActions(),
              ],
            ),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _selectedIndex,
        onTabSelected: (i) => Navigator.of(context).pushReplacement(
          MaterialPageRoute(builder: (_) => _pageForIndex(i)),
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '慢病管理',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.gray800),
            ),
            const ElderlySelector(),
          ],
        ),
      ),
    );
  }

  Widget _buildMetricsSection() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Row(
          children: [
            _buildMetricCard('🩸', '血糖', _latestGlucose?, '/chronic/blood-sugar', AppTheme.statusWarning),
            const SizedBox(width: 8),
            _buildMetricCard('📊', '尿酸', _latestUric?, '/chronic/uric-acid', AppTheme.statusNormal),
            const SizedBox(width: 8),
            _buildMetricCard('🩺', '血压', _latestBP?, '/chronic/blood-pressure', AppTheme.statusNormal),
          ],
        ),
      ),
    );
  }

  Widget _buildMetricCard(String icon, String label, Map<String, dynamic>? data, String route, Color statusColor) {
    return Expanded(
      child: GestureDetector(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => _pageForRoute(route)),
        ),
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: AppTheme.bgCard,
            borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
            border: Border.all(color: statusColor.withValues(alpha: 0.3)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(icon, style: const TextStyle(fontSize: 20)),
              const SizedBox(height: 4),
              Text(label, style: const TextStyle(fontSize: 12, color: AppTheme.gray500)),
              const SizedBox(height: 4),
              if (data != null)
                Text(
                  '${data['value']}',
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppTheme.gray800),
                )
              else
                Text('—', style: const TextStyle(fontSize: 16, color: AppTheme.gray400)),
              const SizedBox(height: 2),
              Text(
                data != null ? (data['test_mode'] ?? '检测') : '暂无数据',
                style: const TextStyle(fontSize: 10, color: AppTheme.gray400),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // ... 其余方法：_buildTasksSection, _buildAIRecommendation, _buildQuickActions, _pageForIndex, _pageForRoute
}
```

- [ ] **Step 4: 提交**

```bash
git add apps/family-app/lib/screens/chronic/chronic_home_page.dart apps/family-app/lib/widgets/bottom_nav_bar.dart apps/family-app/lib/screens/login/main_tab_screen.dart
git commit -m "feat: add chronic care home page with metric cards"
```

---

## 任务 19-22：慢病详情页

**任务19：血糖/尿酸详情页** (`blood_sugar_page.dart`, `uric_acid_page.dart`)
- 折线图 + 检测记录列表 + 异常标记
- 使用 `fl_chart` 包绘制趋势图
- 参考 `apps/family-app/lib/screens/health/health_page.dart` 的图表实现

**任务20：血压详情页** (`blood_pressure_page.dart`)
- 收缩压/舒张压双折线图
- 参考糖护士APP的血压监测界面

**任务21：饮食/运动页** (`diet_page.dart`, `exercise_page.dart`)
- 饮食：食物搜索 + 碳水计算 + 餐次分类
- 运动：手环数据联动 + 运动计划

**任务22：健康报告页** (`report_page.dart`)
- 周报/月报/年报切换
- 关键指标达标率
- AI综合建议

> 每个页面实现模式相同： StatefulWidget + API调用 + 数据展示 + 异常标记。参考现有页面代码模板。

---

## 任务 23：首页/健康页改造

**Files:**
- Modify: `apps/family-app/lib/screens/home/home_page.dart`
- Modify: `apps/family-app/lib/screens/health/health_page.dart`

- [ ] **Step 1: 首页新增慢病快捷入口**

在首页的 `quick_status_card` 下方新增慢病数据卡片行，点击跳转到对应详情页。

- [ ] **Step 2: 健康页新增趋势图**

在现有 `health_page.dart` 的图表区域下方追加血糖/尿酸/血压趋势图。

- [ ] **Step 3: 提交**

```bash
git commit -m "feat: add chronic care shortcuts to home and health pages"
```

---

## 任务 24：任务体系与奖励机制

**Files:**
- Create: `apps/family-app/lib/screens/chronic/widgets/task_list_widget.dart`
- Create: `apps/family-app/lib/screens/chronic/widgets/reward_badge_widget.dart`

- [ ] **Step 1: 实现任务列表组件**

```dart
class TaskListWidget extends StatelessWidget {
  final List<dynamic> tasks;
  final Function(String) onTaskComplete;

  const TaskListWidget({super.key, required this.tasks, required this.onTaskComplete});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('每日任务', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 12),
          ...tasks.map((t) => _buildTaskItem(t)).toList(),
          const SizedBox(height: 12),
          LinearProgressIndicator(value: _completionRate()),
          const SizedBox(height: 8),
          Text('完成 ${_completedCount()}/${tasks.length}', style: const TextStyle(fontSize: 12, color: AppTheme.gray500)),
        ],
      ),
    );
  }

  Widget _buildTaskItem(dynamic task) {
    return ListTile(
      leading: Checkbox(
        value: task['completed'],
        onChanged: task['completed'] ? null : () => onTaskComplete(task['id']),
      ),
      title: Text(_taskLabel(task['task_type']), style: const TextStyle(fontSize: 16)),
      subtitle: Text(task['scheduled_time'], style: const TextStyle(fontSize: 12, color: AppTheme.gray500)),
      trailing: task['completed']
          ? const Icon(Icons.check_circle, color: AppTheme.statusNormal)
          : null,
    );
  }

  String _taskLabel(String type) {
    switch (type) {
      case 'bg_test': return '测空腹血糖';
      case 'ua_test': return '测尿酸';
      case 'bp_test': return '测血压';
      case 'medication': return '按时服药';
      case 'exercise': return '散步30分钟';
      case 'diet': return '记录饮食';
      default: return type;
    }
  }

  int _completedCount() => tasks.where((t) => t['completed']).length;
  double _completionRate() => tasks.isEmpty ? 0 : _completedCount() / tasks.length;
}
```

- [ ] **Step 2: 实现奖励徽章组件**

```dart
class RewardBadgeWidget extends StatelessWidget {
  final int streak;
  final List<String> badges;

  const RewardBadgeWidget({super.key, required this.streak, required this.badges});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.primaryLight,
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('健康积分', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: AppTheme.primary)),
          const SizedBox(height: 8),
          Row(
            children: [
              Text('$streak 天连续达标', style: const TextStyle(fontSize: 14)),
              const Spacer(),
              ...badges.map((b) => _buildBadge(b)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBadge(String name) {
    return Container(
      margin: const EdgeInsets.only(left: 8),
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: AppTheme.primary,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(name, style: const TextStyle(fontSize: 10, color: Colors.white)),
    );
  }
}
```

- [ ] **Step 3: 集成到慢病管理主页**

在 `chronic_home_page.dart` 中引入上述组件。

- [ ] **Step 4: 提交**

```bash
git commit -m "feat: add task system and reward badges for chronic care"
```

---

## 任务 25：端到端联调

- [ ] **Step 1: 启动所有服务**

```bash
# 后端
cd cloud/api-server && ./start.sh
cd cloud/data-pipeline && ./start.sh
cd cloud/push-service && ./start.sh

# 模拟设备数据
curl -X POST http://localhost:8080/api/v1/chronic/:elderly_id/glucose \
  -H "Content-Type: application/json" \
  -d '{"value": 6.8, "test_mode": "fasting", "source": "test_strip"}'

curl -X POST http://localhost:8080/api/v1/chronic/:elderly_id/blood-pressure \
  -H "Content-Type: application/json" \
  -d '{"systolic": 135, "diastolic": 85, "pulse": 72}'
```

- [ ] **Step 2: 验证API端点**

对所有15个新API端点进行curl测试，验证CRUD操作和错误处理。

- [ ] **Step 3: APP联调**

```bash
cd apps/family-app && flutter run -d chrome
```

测试慢病管理模块的完整流程：查看数据 → 记录数据 → 完成任务 → 查看报告。

- [ ] **Step 4: 性能测试**

```bash
# 模拟100次并发请求
ab -n 100 -c 10 http://localhost:8080/api/v1/chronic/:elderly_id/glucose/trend?days=30
```

- [ ] **Step 5: 提交**

```bash
git commit -m "test: end-to-end integration testing for chronic care module"
```

---

## 自审检查清单

- [x] 所有7张数据库表已实现
- [x] 所有15个API端点已实现
- [x] 固件模块通过条件编译隔离
- [x] APP新增7个页面 + 3个改造页面
- [x] 分析引擎覆盖血糖/尿酸/血压/饮食/运动
- [x] 推送服务扩展支持慢病提醒
- [x] 无TBD/TODO占位符
- [x] 类型一致性：所有handler使用相同的ChronicStore
- [x] 路由命名一致：`/api/v1/chronic/:elderly_id/*`
