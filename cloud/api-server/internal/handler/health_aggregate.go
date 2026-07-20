package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthAggregateHandler provides top-level health endpoints
// not scoped to a specific elderly profile.
type HealthAggregateHandler struct {
	pg  *store.Postgres
	log *zap.Logger
}

func NewHealthAggregateHandler(pg *store.Postgres, log *zap.Logger) *HealthAggregateHandler {
	return &HealthAggregateHandler{pg: pg, log: log}
}

// GET /api/v1/health/latest
// Returns the latest health reading across all elderly profiles
// accessible to the authenticated user.
func (h *HealthAggregateHandler) Latest(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	elderIDs, err := h.pg.GetElderlyIDsByUserID(c.Request.Context(), userID.(string))
	if err != nil || len(elderIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": "OK", "data": []any{}})
		return
	}

	type result struct {
		ElderlyID string           `json:"elderly_id"`
		Name      string           `json:"name"`
		Record    model.HealthRecord `json:"record"`
	}

	now := time.Now()
	start := now.Add(-24 * time.Hour)
	var results []result

	for _, eid := range elderIDs {
		var hr, spo2, steps *int64
		q := `SELECT hr::bigint, spo2::bigint, steps FROM health_records
			  WHERE elderly_id = $1 AND timestamp >= $2
			  ORDER BY timestamp DESC LIMIT 1`
		err := h.pg.Pool().QueryRow(c.Request.Context(), q, eid, start).Scan(&hr, &spo2, &steps)
		if err != nil {
			continue
		}

		var name string
		h.pg.Pool().QueryRow(c.Request.Context(),
			`SELECT name FROM elderly_profiles WHERE id = $1`, eid).Scan(&name)

		r := model.HealthRecord{ElderlyID: eid, Timestamp: now}
		if hr != nil {
			v := int(*hr)
			r.HR = &v
		}
		if spo2 != nil {
			v := int(*spo2)
			r.SPO2 = &v
		}
		r.Steps = steps

		results = append(results, result{
			ElderlyID: eid,
			Name:      name,
			Record:    r,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": results})
}

// GET /api/v1/health/records
// Returns health records for all accessible elderly profiles with optional time range.
func (h *HealthAggregateHandler) Records(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	elderIDs, err := h.pg.GetElderlyIDsByUserID(c.Request.Context(), userID.(string))
	if err != nil || len(elderIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": "OK", "data": []any{}})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days < 1 || days > 365 {
		days = 7
	}

	// Build IN clause args: $1, $2, ...
	args := make([]any, len(elderIDs))
	for i, id := range elderIDs {
		args[i] = id
	}
	ph := make([]string, len(elderIDs))
	for i := range ph {
		ph[i] = "$" + strconv.Itoa(i+1)
	}

	q := "SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic FROM health_records " +
		"WHERE elderly_id IN (" + strings.Join(ph, ",") + ") AND timestamp >= now() - (interval '1 day' * $0) " +
		"ORDER BY timestamp DESC LIMIT 100"

	rows, err := h.pg.Pool().Query(c.Request.Context(), q, append([]any{days}, args...)...)
	if err != nil {
		h.log.Error("query health records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch records"})
		return
	}
	defer rows.Close()

	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			continue
		}
		records = append(records, r)
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// GET /api/v1/health/risk-score
// Returns a simple health risk score (0-100) based on recent vitals.
func (h *HealthAggregateHandler) RiskScore(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	elderIDs, err := h.pg.GetElderlyIDsByUserID(c.Request.Context(), userID.(string))
	if err != nil || len(elderIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{"score": 0, "level": "unknown"}})
		return
	}

	// Calculate risk per elderly, return worst
	type scoreResult struct {
		ElderlyID string `json:"elderly_id"`
		Name      string `json:"name"`
		Score     int    `json:"score"`
		Level     string `json:"level"`
		Factors   []string `json:"factors"`
	}

	var results []scoreResult

	for _, eid := range elderIDs {
		score, factors := h.calculateRisk(c, eid)
		var name string
		h.pg.Pool().QueryRow(c.Request.Context(),
			`SELECT name FROM elderly_profiles WHERE id = $1`, eid).Scan(&name)

		results = append(results, scoreResult{
			ElderlyID: eid, Name: name, Score: score, Level: levelLabel(score), Factors: factors,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": results})
}

func (h *HealthAggregateHandler) calculateRisk(c *gin.Context, elderlyID string) (int, []string) {
	var avgHR, avgSpO2, totalSteps int64
	var lastHR, lastSpO2 *int

	// Recent heart rate
	q := `SELECT AVG(hr)::bigint, MAX(hr), MIN(hr) FROM health_records
		  WHERE elderly_id = $1 AND timestamp >= now() - interval '7 days' AND hr IS NOT NULL`
	err := h.pg.Pool().QueryRow(c.Request.Context(), q, elderlyID).Scan(&avgHR, &lastHR, &lastSpO2)
	if err != nil {
		return 0, nil
	}

	score := 0
	var factors []string

	if lastHR != nil && *lastHR > 100 {
		score += 20
		factors = append(factors, "elevated_heart_rate")
	} else if lastHR != nil && *lastHR < 50 {
		score += 25
		factors = append(factors, "low_heart_rate")
	}

	if avgHR > 90 {
		score += 10
		factors = append(factors, "high_avg_heart_rate")
	}

	q2 := `SELECT AVG(spo2)::bigint FROM health_records
		   WHERE elderly_id = $1 AND timestamp >= now() - interval '7 days' AND spo2 IS NOT NULL`
	err = h.pg.Pool().QueryRow(c.Request.Context(), q2, elderlyID).Scan(&avgSpO2)
	if err == nil {
		if avgSpO2 < 92 {
			score += 30
			factors = append(factors, "low_oxygen")
		} else if avgSpO2 < 95 {
			score += 15
			factors = append(factors, "borderline_oxygen")
		}
	}

	q3 := `SELECT COALESCE(SUM(steps),0)::bigint FROM health_records
		   WHERE elderly_id = $1 AND timestamp >= now() - interval '7 days'`
	err = h.pg.Pool().QueryRow(c.Request.Context(), q3, elderlyID).Scan(&totalSteps)
	if err == nil && totalSteps < 500 {
		score += 10
		factors = append(factors, "low_activity")
	}

	if score > 100 {
		score = 100
	}
	return score, factors
}

func levelLabel(score int) string {
	switch {
	case score <= 20:
		return "low"
	case score <= 50:
		return "medium"
	case score <= 75:
		return "high"
	default:
		return "critical"
	}
}
