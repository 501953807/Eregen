package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// ElderlyHandler serves elderly person management endpoints.
type ElderlyHandler struct {
	store store.Store
}

// NewElderlyHandler creates a new ElderlyHandler.
func NewElderlyHandler(s store.Store) *ElderlyHandler {
	return &ElderlyHandler{store: s}
}

// List returns a paginated list of elderly profiles.
func (h *ElderlyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	profiles, err := h.store.ListElderly(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      profiles,
		"page":      page,
		"page_size": pageSize,
	})
}

// Detail returns an elderly profile by ID.
func (h *ElderlyHandler) Detail(c *gin.Context) {
	id := c.Param("id")
	profile, err := h.store.GetElderly(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "elderly not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": profile})
}

// Create adds a new elderly profile.
func (h *ElderlyHandler) Create(c *gin.Context) {
	var body struct {
		Name        string   `json:"name" binding:"required"`
		BirthDate   string   `json:"birth_date"`
		UserID      string   `json:"user_id"`
		HealthTiers []string `json:"health_tiers"`
		AvatarURL   string   `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile, err := h.store.CreateElderly(c.Request.Context(), body.Name, body.BirthDate, body.UserID, body.HealthTiers, body.AvatarURL)
	if err != nil {
		log.Printf("CreateElderly failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	log.Printf("CreateElderly success: id=%s, name=%s", profile.ID, profile.Name)
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": profile})
}

// Update modifies an existing elderly profile.
func (h *ElderlyHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Name        string   `json:"name"`
		BirthDate   string   `json:"birth_date"`
		UserID      string   `json:"user_id"`
		HealthTiers []string `json:"health_tiers"`
		AvatarURL   string   `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile, err := h.store.UpdateElderly(c.Request.Context(), id, body.Name, body.BirthDate, body.UserID, body.HealthTiers, body.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": profile})
}

// Delete removes an elderly profile.
func (h *ElderlyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteElderly(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// HealthStats returns health statistics for an elderly person.
func (h *ElderlyHandler) HealthStats(c *gin.Context) {
	elderlyID := c.Param("id")
	stats, err := h.store.GetElderlyHealthStats(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// HealthRecords returns recent health records for an elderly person.
func (h *ElderlyHandler) HealthRecords(c *gin.Context) {
	elderlyID := c.Param("id")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	records, err := h.store.GetElderlyHealthRecords(c.Request.Context(), elderlyID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// CreateHealthRecord adds a new health record for an elderly person.
func (h *ElderlyHandler) CreateHealthRecord(c *gin.Context) {
	elderlyID := c.Param("id")
	var body struct {
		HR         *int     `json:"hr"`
		SpO2       *int     `json:"spo2"`
		Steps      *int64   `json:"steps"`
		SleepHours *float64 `json:"sleep_hours"`
		Timestamp  string   `json:"timestamp"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	record := &model.HealthRecordRow{ElderlyID: elderlyID}
	if body.HR != nil {
		record.HR = body.HR
	}
	if body.SpO2 != nil {
		record.SpO2 = body.SpO2
	}
	if body.Steps != nil {
		record.Steps = body.Steps
	}
	if body.SleepHours != nil {
		record.SleepHours = body.SleepHours
	}
	if body.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, body.Timestamp); err == nil {
			record.Timestamp = t
		}
	}
	if err := h.store.CreateHealthRecord(c.Request.Context(), record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": record})
}


// MedicationRules returns medication rules for an elderly person.
func (h *ElderlyHandler) MedicationRules(c *gin.Context) {
	elderlyID := c.Param("id")
	rules, err := h.store.GetElderlyMedicationRules(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// CreateMedicationRule adds a new medication rule for an elderly person.
func (h *ElderlyHandler) CreateMedicationRule(c *gin.Context) {
	elderlyID := c.Param("id")
	var body model.MedicationRuleRow
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	body.ElderlyID = elderlyID
	if err := h.store.CreateMedicationRule(c.Request.Context(), elderlyID, &body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": body})
}

// UpdateMedicationRule updates an existing medication rule.
func (h *ElderlyHandler) UpdateMedicationRule(c *gin.Context) {
	elderlyID := c.Param("id")
	ruleID := c.Param("rule_id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateMedicationRule(c.Request.Context(), elderlyID, ruleID, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule updated"})
}

// DeleteMedicationRule removes a medication rule.
func (h *ElderlyHandler) DeleteMedicationRule(c *gin.Context) {
	elderlyID := c.Param("id")
	ruleID := c.Param("rule_id")
	if err := h.store.DeleteMedicationRule(c.Request.Context(), elderlyID, ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
}
func (h *ElderlyHandler) DeviceList(c *gin.Context) {
	elderlyID := c.Param("id")
	devices, err := h.store.GetElderlyDevices(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": devices})
}

// LocationHistory returns location history for an elderly person.
func (h *ElderlyHandler) LocationHistory(c *gin.Context) {
	elderlyID := c.Param("id")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	locations, err := h.store.GetElderlyLocationHistory(c.Request.Context(), elderlyID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": locations})
}

// CreateLocation adds a new location record for an elderly person.
func (h *ElderlyHandler) CreateLocation(c *gin.Context) {
	elderlyID := c.Param("id")
	var body struct {
		Lat      float64  `json:"lat" binding:"required"`
		Lon      float64  `json:"lon" binding:"required"`
		Accuracy *float64 `json:"accuracy"`
		Timestamp string   `json:"timestamp"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	loc := &model.LocationPoint{ElderlyID: elderlyID, Lat: body.Lat, Lon: body.Lon}
	if body.Accuracy != nil {
		loc.Accuracy = body.Accuracy
	}
	if body.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, body.Timestamp); err == nil {
			loc.Timestamp = t
		}
	}
	if err := h.store.CreateLocation(c.Request.Context(), loc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": loc})
}


// AlertHistory returns alert history for an elderly person.
func (h *ElderlyHandler) AlertHistory(c *gin.Context) {
	elderlyID := c.Param("id")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	alerts, err := h.store.GetElderlyAlertHistory(c.Request.Context(), elderlyID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": alerts})
}

// ConvertToProfileSummary builds a model.ElderlyProfile from DB row values.
func ConvertToProfileSummary(id, name, userID, avatarURL string, birthDate *time.Time, healthTiersRaw interface{}, createdAt, updatedAt time.Time) model.ElderlyProfile {
	tiers := []string{}
	switch v := healthTiersRaw.(type) {
	case []interface{}:
		for _, t := range v {
			if s, ok := t.(string); ok {
				tiers = append(tiers, s)
			}
		}
	case []byte:
		if len(v) > 0 {
			var raw []string
			if err := json.Unmarshal(v, &raw); err == nil {
				tiers = raw
			}
		}
	}
	return model.ElderlyProfile{
		ID:          id,
		Name:        name,
		UserID:      userID,
		AvatarURL:   &avatarURL,
		BirthDate:   birthDate,
		HealthTiers: tiers,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
