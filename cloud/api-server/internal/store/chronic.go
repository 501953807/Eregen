package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// ChronicStore provides CRUD access to chronic disease data tables in SQLite.
type ChronicStore struct {
	db *sql.DB
}

// NewChronicStore creates a ChronicStore from an existing *sql.DB.
func NewChronicStore(db *sql.DB) *ChronicStore {
	return &ChronicStore{db: db}
}

// ─── 血糖记录 ───────────────────────────────────────────────────────────────

// SaveGlucoseRecord inserts a blood glucose record and returns its ID.
func (s *ChronicStore) SaveGlucoseRecord(ctx context.Context, r *model.ChronicGlucoseRecord) error {
	r.ID = uuid.New().String()
	r.DetectedAt = time.Now()

	q := `INSERT INTO chronic_glucose_records
		(id, elderly_id, value, unit, test_mode, measurement_time, detected_at, source, quality, temperature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.Value, r.Unit, r.TestMode,
		r.MeasurementTime, r.DetectedAt, r.Source, r.Quality, r.Temperature,
	)
	if err != nil {
		return fmt.Errorf("save glucose record: %w", err)
	}
	return nil
}

// ListGlucoseRecords returns glucose records for an elderly person, ordered by time desc.
func (s *ChronicStore) ListGlucoseRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicGlucoseRecord, error) {
	from := time.Now().AddDate(0, 0, -days)

	q := `SELECT id, elderly_id, value, unit, test_mode, measurement_time, detected_at, source, quality, temperature
		  FROM chronic_glucose_records
		  WHERE elderly_id = $1 AND measurement_time >= $2
		  ORDER BY measurement_time DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, from)
	if err != nil {
		return nil, fmt.Errorf("list glucose records: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicGlucoseRecord
	for rows.Next() {
		var r model.ChronicGlucoseRecord
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.Value, &r.Unit, &r.TestMode,
			&r.MeasurementTime, &r.DetectedAt, &r.Source, &r.Quality, &r.Temperature,
		); err != nil {
			return nil, fmt.Errorf("scan glucose record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("glucose records iteration: %w", err)
	}
	return records, nil
}

// ─── 尿酸记录 ───────────────────────────────────────────────────────────────

// SaveUricAcidRecord inserts a uric acid record and returns its ID.
func (s *ChronicStore) SaveUricAcidRecord(ctx context.Context, r *model.ChronicUricAcidRecord) error {
	r.ID = uuid.New().String()
	r.DetectedAt = time.Now()

	q := `INSERT INTO chronic_uric_acid_records
		(id, elderly_id, value, unit, measurement_time, detected_at, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.Value, r.Unit,
		r.MeasurementTime, r.DetectedAt, r.Source,
	)
	if err != nil {
		return fmt.Errorf("save uric acid record: %w", err)
	}
	return nil
}

// ListUricAcidRecords returns uric acid records for an elderly person, ordered by time desc.
func (s *ChronicStore) ListUricAcidRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicUricAcidRecord, error) {
	from := time.Now().AddDate(0, 0, -days)

	q := `SELECT id, elderly_id, value, unit, measurement_time, detected_at, source
		  FROM chronic_uric_acid_records
		  WHERE elderly_id = $1 AND measurement_time >= $2
		  ORDER BY measurement_time DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, from)
	if err != nil {
		return nil, fmt.Errorf("list uric acid records: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicUricAcidRecord
	for rows.Next() {
		var r model.ChronicUricAcidRecord
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.Value, &r.Unit,
			&r.MeasurementTime, &r.DetectedAt, &r.Source,
		); err != nil {
			return nil, fmt.Errorf("scan uric acid record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("uric acid records iteration: %w", err)
	}
	return records, nil
}

// ─── 血压记录 ────────────────────────────────────────────────────────────────

// SaveBPRecord inserts a blood pressure record and returns its ID.
func (s *ChronicStore) SaveBPRecord(ctx context.Context, r *model.ChronicBPRecord) error {
	r.ID = uuid.New().String()
	r.DetectedAt = time.Now()

	q := `INSERT INTO chronic_bp_records
		(id, elderly_id, systolic, diastolic, pulse, measurement_time, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.Systolic, r.Diastolic, r.Pulse,
		r.MeasurementTime, r.DetectedAt,
	)
	if err != nil {
		return fmt.Errorf("save bp record: %w", err)
	}
	return nil
}

// ListBPRecords returns blood pressure records for an elderly person, ordered by time desc.
func (s *ChronicStore) ListBPRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicBPRecord, error) {
	from := time.Now().AddDate(0, 0, -days)

	q := `SELECT id, elderly_id, systolic, diastolic, pulse, measurement_time, detected_at
		  FROM chronic_bp_records
		  WHERE elderly_id = $1 AND measurement_time >= $2
		  ORDER BY measurement_time DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, from)
	if err != nil {
		return nil, fmt.Errorf("list bp records: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicBPRecord
	for rows.Next() {
		var r model.ChronicBPRecord
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.Systolic, &r.Diastolic, &r.Pulse,
			&r.MeasurementTime, &r.DetectedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bp record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("bp records iteration: %w", err)
	}
	return records, nil
}

// ─── 饮食记录 ────────────────────────────────────────────────────────────────

// SaveDietRecord inserts a diet/meal log entry and returns its ID.
func (s *ChronicStore) SaveDietRecord(ctx context.Context, r *model.ChronicDietRecord) error {
	r.ID = uuid.New().String()
	r.RecordedAt = time.Now()

	q := `INSERT INTO chronic_diet_records
		(id, elderly_id, meal_type, food_items, total_carbs, total_calories, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.MealType, r.FoodItems,
		r.TotalCarbs, r.TotalCalories, r.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("save diet record: %w", err)
	}
	return nil
}

// ListDietRecords returns diet records for an elderly person, ordered by time desc.
func (s *ChronicStore) ListDietRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicDietRecord, error) {
	from := time.Now().AddDate(0, 0, -days)

	q := `SELECT id, elderly_id, meal_type, food_items, total_carbs, total_calories, recorded_at
		  FROM chronic_diet_records
		  WHERE elderly_id = $1 AND recorded_at >= $2
		  ORDER BY recorded_at DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, from)
	if err != nil {
		return nil, fmt.Errorf("list diet records: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicDietRecord
	for rows.Next() {
		var r model.ChronicDietRecord
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.MealType, &r.FoodItems,
			&r.TotalCarbs, &r.TotalCalories, &r.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("scan diet record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("diet records iteration: %w", err)
	}
	return records, nil
}

// ─── 运动记录 ────────────────────────────────────────────────────────────────

// SaveExerciseRecord inserts an exercise session record and returns its ID.
func (s *ChronicStore) SaveExerciseRecord(ctx context.Context, r *model.ChronicExerciseRecord) error {
	r.ID = uuid.New().String()
	r.RecordedAt = time.Now()

	q := `INSERT INTO chronic_exercise_records
		(id, elderly_id, type, duration_min, calories, avg_hr, max_hr, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.Type, r.DurationMin, r.Calories,
		r.AvgHR, r.MaxHR, r.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("save exercise record: %w", err)
	}
	return nil
}

// ListExerciseRecords returns exercise records for an elderly person, ordered by time desc.
func (s *ChronicStore) ListExerciseRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicExerciseRecord, error) {
	from := time.Now().AddDate(0, 0, -days)

	q := `SELECT id, elderly_id, type, duration_min, calories, avg_hr, max_hr, recorded_at
		  FROM chronic_exercise_records
		  WHERE elderly_id = $1 AND recorded_at >= $2
		  ORDER BY recorded_at DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, from)
	if err != nil {
		return nil, fmt.Errorf("list exercise records: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicExerciseRecord
	for rows.Next() {
		var r model.ChronicExerciseRecord
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.Type, &r.DurationMin, &r.Calories,
			&r.AvgHR, &r.MaxHR, &r.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exercise record: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exercise records iteration: %w", err)
	}
	return records, nil
}

// ─── 每日任务 ────────────────────────────────────────────────────────────────

// SaveDailyTask inserts or updates a daily health task.
func (s *ChronicStore) SaveDailyTask(ctx context.Context, r *model.ChronicDailyTask) error {
	r.ID = uuid.New().String()
	r.TaskDate = time.Now().Format("2006-01-02")

	q := `INSERT INTO chronic_daily_tasks
		(id, elderly_id, task_type, scheduled_time, completed, completed_at, task_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.TaskType, r.ScheduledTime,
		r.Completed, r.CompletedAt, r.TaskDate,
	)
	if err != nil {
		return fmt.Errorf("save daily task: %w", err)
	}
	return nil
}

// UpdateDailyTask marks a task as completed.
func (s *ChronicStore) UpdateDailyTask(ctx context.Context, taskID string, completed bool) error {
	now := time.Now()
	q := `UPDATE chronic_daily_tasks SET completed = $1, completed_at = $2 WHERE id = $3`
	_, err := s.db.ExecContext(ctx, q, boolToInt(completed), now, taskID)
	if err != nil {
		return fmt.Errorf("update daily task: %w", err)
	}
	return nil
}

// ListDailyTasks returns tasks for an elderly person on a given date.
func (s *ChronicStore) ListDailyTasks(ctx context.Context, elderlyID, date string) ([]model.ChronicDailyTask, error) {
	q := `SELECT id, elderly_id, task_type, scheduled_time, completed, completed_at, task_date
		  FROM chronic_daily_tasks
		  WHERE elderly_id = $1 AND task_date = $2
		  ORDER BY scheduled_time ASC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID, date)
	if err != nil {
		return nil, fmt.Errorf("list daily tasks: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicDailyTask
	for rows.Next() {
		var r model.ChronicDailyTask
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.TaskType, &r.ScheduledTime,
			&r.Completed, &r.CompletedAt, &r.TaskDate,
		); err != nil {
			return nil, fmt.Errorf("scan daily task: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("daily tasks iteration: %w", err)
	}
	return records, nil
}

// ─── 周期报告 ────────────────────────────────────────────────────────────────

// SaveHealthReport inserts a periodic health report.
func (s *ChronicStore) SaveHealthReport(ctx context.Context, r *model.ChronicHealthReport) error {
	r.ID = uuid.New().String()
	r.GeneratedAt = time.Now()

	q := `INSERT INTO chronic_health_reports
		(id, elderly_id, report_type, period_start, period_end, data_summary, ai_recommendations, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.ExecContext(ctx, q,
		r.ID, r.ElderlyID, r.ReportType, r.PeriodStart, r.PeriodEnd,
		r.DataSummary, r.AIRecommendations, r.GeneratedAt,
	)
	if err != nil {
		return fmt.Errorf("save health report: %w", err)
	}
	return nil
}

// ListHealthReports returns health reports for an elderly person.
func (s *ChronicStore) ListHealthReports(ctx context.Context, elderlyID string) ([]model.ChronicHealthReport, error) {
	q := `SELECT id, elderly_id, report_type, period_start, period_end, data_summary, ai_recommendations, generated_at
		  FROM chronic_health_reports
		  WHERE elderly_id = $1
		  ORDER BY generated_at DESC`

	rows, err := s.db.QueryContext(ctx, q, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("list health reports: %w", err)
	}
	defer rows.Close()

	var records []model.ChronicHealthReport
	for rows.Next() {
		var r model.ChronicHealthReport
		if err := rows.Scan(
			&r.ID, &r.ElderlyID, &r.ReportType,
			&r.PeriodStart, &r.PeriodEnd,
			&r.DataSummary, &r.AIRecommendations, &r.GeneratedAt,
		); err != nil {
			return nil, fmt.Errorf("scan health report: %w", err)
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("health reports iteration: %w", err)
	}
	return records, nil
}

// boolToInt converts a bool to 0/1 int for SQLite (INTEGER column).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
