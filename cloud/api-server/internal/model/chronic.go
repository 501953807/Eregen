package model

import "time"

// ─── 血糖检测记录 ───────────────────────────────────────────────────────────

// ChronicGlucoseRecord is a blood glucose test record.
type ChronicGlucoseRecord struct {
	ID              string     `json:"id"`
	ElderlyID       string     `json:"elderly_id"`
	Value           float64    `json:"value"`
	Unit            string     `json:"unit"`             // mmol/L
	TestMode        string     `json:"test_mode"`        // fasting / random / postprandial
	MeasurementTime time.Time  `json:"measurement_time"`
	DetectedAt      time.Time  `json:"detected_at"`
	Source          string     `json:"source"`           // test_strip / device
	Quality         *float64   `json:"quality,omitempty"` // 0–1 score
	Temperature     *float64   `json:"temperature,omitempty"`
}

// ─── 尿酸检测记录 ───────────────────────────────────────────────────────────

// ChronicUricAcidRecord is a uric acid test record.
type ChronicUricAcidRecord struct {
	ID              string    `json:"id"`
	ElderlyID       string    `json:"elderly_id"`
	Value           float64   `json:"value"`
	Unit            string    `json:"unit"`            // μmol/L
	MeasurementTime time.Time `json:"measurement_time"`
	DetectedAt      time.Time `json:"detected_at"`
	Source          string    `json:"source"` // test_strip / device
}

// ─── 血压记录 ────────────────────────────────────────────────────────────────

// ChronicBPRecord is a blood pressure measurement record.
type ChronicBPRecord struct {
	ID              string    `json:"id"`
	ElderlyID       string    `json:"elderly_id"`
	Systolic        int       `json:"systolic"`
	Diastolic       int       `json:"diastolic"`
	Pulse           *int      `json:"pulse,omitempty"`
	MeasurementTime time.Time `json:"measurement_time"`
	DetectedAt      time.Time `json:"detected_at"`
}

// ─── 饮食记录 ────────────────────────────────────────────────────────────────

// ChronicDietRecord is a diet / meal log entry.
type ChronicDietRecord struct {
	ID             string    `json:"id"`
	ElderlyID      string    `json:"elderly_id"`
	MealType       string    `json:"meal_type"`       // breakfast / lunch / dinner / snack
	FoodItems      string    `json:"food_items"`      // JSON array of food names
	TotalCarbs     *float64  `json:"total_carbs,omitempty"`
	TotalCalories  *float64  `json:"total_calories,omitempty"`
	RecordedAt     time.Time `json:"recorded_at"`
}

// ─── 运动记录 ────────────────────────────────────────────────────────────────

// ChronicExerciseRecord is an exercise session record.
type ChronicExerciseRecord struct {
	ID          string    `json:"id"`
	ElderlyID   string    `json:"elderly_id"`
	Type        string    `json:"type"`         // walking / running / tai_chi / etc.
	DurationMin *int      `json:"duration_min,omitempty"`
	Calories    *float64  `json:"calories,omitempty"`
	AvgHR       *int      `json:"avg_hr,omitempty"`
	MaxHR       *int      `json:"max_hr,omitempty"`
	RecordedAt  time.Time `json:"recorded_at"`
}

// ─── 每日任务 ────────────────────────────────────────────────────────────────

// ChronicDailyTask is a scheduled daily health task (medication, BP check, etc.).
type ChronicDailyTask struct {
	ID            string     `json:"id"`
	ElderlyID     string     `json:"elderly_id"`
	TaskType      string     `json:"task_type"` // med_take / bp_measure / walk / etc.
	ScheduledTime string     `json:"scheduled_time"` // HH:MM
	Completed     int        `json:"completed"` // 0=not completed, 1=completed
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	TaskDate      string     `json:"task_date"` // YYYY-MM-DD
}

// ─── 周期报告 ────────────────────────────────────────────────────────────────

// ChronicHealthReport is a periodic health summary report.
type ChronicHealthReport struct {
	ID                 string    `json:"id"`
	ElderlyID          string    `json:"elderly_id"`
	ReportType         string    `json:"report_type"` // daily / weekly / monthly
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	DataSummary        *string   `json:"data_summary,omitempty"`
	AIRecommendations  *string   `json:"ai_recommendations,omitempty"`
	GeneratedAt        time.Time `json:"generated_at"`
}
