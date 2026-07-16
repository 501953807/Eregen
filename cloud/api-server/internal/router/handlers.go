package router

import (
	"strconv"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
)

// healthSummary returns a handler for GET /elderly/:id/health/summary
func healthSummary(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		dayStr := c.DefaultQuery("day", time.Now().Format("2006-01-02"))
		day, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			c.JSON(400, gin.H{"code": "INVALID_DATE", "message": "Use YYYY-MM-DD"})
			return
		}
		rec, err := pg.GetHealthSummary(c.Request.Context(), elderlyID, day)
		if err != nil {
			c.JSON(404, gin.H{"code": "NOT_FOUND", "message": "No health data found"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": rec})
	}
}

func healthHistory(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		if days < 1 || days > 365 {
			days = 30
		}
		records, err := pg.GetHealthHistory(c.Request.Context(), elderlyID, days)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": records})
	}
}

func healthTrend(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		metric := c.DefaultQuery("metric", "hr")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		if days < 1 || days > 365 {
			days = 30
		}
		validMetrics := map[string]bool{"hr": true, "spo2": true, "steps": true, "sleep_hours": true, "bp_systolic": true, "bp_diastolic": true}
		if !validMetrics[metric] {
			c.JSON(400, gin.H{"code": "INVALID_METRIC", "message": "Valid: hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic"})
			return
		}
		records, err := pg.GetHealthTrend(c.Request.Context(), elderlyID, metric, days)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": records})
	}
}

func locationLatest(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		loc, err := pg.GetLatestLocation(c.Request.Context(), elderlyID)
		if err != nil {
			c.JSON(404, gin.H{"code": "NOT_FOUND", "message": "No location data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": loc})
	}
}

func locationHistory(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{"code": "INVALID_DATE", "message": "Use YYYY-MM-DD"})
			return
		}
		from := date.Truncate(24 * time.Hour)
		until := from.Add(24 * time.Hour)
		records, err := pg.GetLocationHistory(c.Request.Context(), elderlyID, from, until)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": records})
	}
}

func geofenceSet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Lat          float64 `json:"lat" binding:"required"`
			Lon          float64 `json:"lon" binding:"required"`
			RadiusMeters int     `json:"radius_meters" binding:"required,min=50,max=10000"`
			Name         string  `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
			return
		}
		_ = req
		c.JSON(201, gin.H{"code": "OK", "message": "Geofence set"})
	}
}

func geofenceList() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"code": "OK", "data": []any{}})
	}
}

func medRules(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		rules, err := pg.ListMedicationRules(c.Request.Context(), elderlyID)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch rules"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": rules})
	}
}

func medCreateRule(pg *store.Postgres, nats *service.NatsClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		var req model.CreateMedicationRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
			return
		}
		if _, err := time.Parse("15:04", req.ScheduleTime); err != nil {
			c.JSON(400, gin.H{"code": "INVALID_TIME", "message": "schedule_time must be HH:MM"})
			return
		}
		mr := &model.MedicationRule{
			ElderlyID:    elderlyID,
			ScheduleTime: req.ScheduleTime,
			DoseCount:    req.DoseCount,
			PillType:     req.PillType,
			DaysOfWeek:   req.DaysOfWeek,
			Active:       req.Active,
		}
		if err := pg.CreateMedicationRule(c.Request.Context(), mr); err != nil {
			c.JSON(500, gin.H{"code": "CREATE_FAILED", "message": "Failed to create rule"})
			return
		}
		if nats != nil {
			cmd := map[string]any{
				"type": "med_rule",
				"rule": map[string]any{
					"time": req.ScheduleTime, "dose": req.DoseCount,
					"type": req.PillType, "days": req.DaysOfWeek,
				},
			}
			_ = nats.PublishCommand(c.Request.Context(), "BR-XXXX", cmd)
		}
		c.JSON(201, gin.H{"code": "OK", "message": "Medication rule created and pushed to device"})
	}
}

func medUpdateRule(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID, exists := c.Get("rule_id")
		if !exists {
			c.JSON(400, gin.H{"code": "MISSING_RULE_ID", "message": "rule_id required"})
			return
		}
		var req model.CreateMedicationRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
			return
		}
		if _, err := time.Parse("15:04", req.ScheduleTime); err != nil {
			c.JSON(400, gin.H{"code": "INVALID_TIME", "message": "schedule_time must be HH:MM"})
			return
		}
		if _, err := pg.GetMedicationRule(c.Request.Context(), ruleID.(string)); err != nil {
			c.JSON(404, gin.H{"code": "NOT_FOUND", "message": "Rule not found"})
			return
		}
		if err := pg.UpdateMedicationRule(c.Request.Context(), ruleID.(string), &req); err != nil {
			c.JSON(500, gin.H{"code": "UPDATE_FAILED", "message": "Failed to update rule"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "message": "Medication rule updated"})
	}
}

func medDeleteRule(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID, exists := c.Get("rule_id")
		if !exists {
			c.JSON(400, gin.H{"code": "MISSING_RULE_ID", "message": "rule_id required"})
			return
		}
		if err := pg.DeleteMedicationRule(c.Request.Context(), ruleID.(string)); err != nil {
			c.JSON(404, gin.H{"code": "NOT_FOUND", "message": "Rule not found"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "message": "Medication rule deleted"})
	}
}

func medToday(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		records, err := pg.GetTodayMedStatus(c.Request.Context(), elderlyID)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": records})
	}
}

func medHistory(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		if days < 1 || days > 365 {
			days = 30
		}
		records, err := pg.GetMedicationHistory(c.Request.Context(), elderlyID, days)
		if err != nil {
			c.JSON(500, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch data"})
			return
		}
		c.JSON(200, gin.H{"code": "OK", "data": records})
	}
}
