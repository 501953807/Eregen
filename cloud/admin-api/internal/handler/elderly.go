package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// ElderlyHandler serves elderly person management endpoints.
type ElderlyHandler struct {
	store *store.PostgresStore
}

// NewElderlyHandler creates a new ElderlyHandler.
func NewElderlyHandler(s *store.PostgresStore) *ElderlyHandler {
	return &ElderlyHandler{store: s}
}

// List returns a paginated list of elderly profiles.
func (h *ElderlyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profiles, err := h.store.ListElderly(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
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
	c.JSON(http.StatusOK, gin.H{"data": profile})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": profile})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}

// Delete removes an elderly profile.
func (h *ElderlyHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteElderly(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, gin.H{"data": stats})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

// MedicationRules returns medication rules for an elderly person.
func (h *ElderlyHandler) MedicationRules(c *gin.Context) {
	elderlyID := c.Param("id")
	rules, err := h.store.GetElderlyMedicationRules(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// DeviceList returns devices linked to an elderly person.
func (h *ElderlyHandler) DeviceList(c *gin.Context) {
	elderlyID := c.Param("id")
	devices, err := h.store.GetElderlyDevices(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": devices})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": locations})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alerts})
}

// ConvertToProfileSummary builds a model.ElderlyProfile from DB row values.
func ConvertToProfileSummary(id, name, userID, avatarURL string, birthDate *time.Time, healthTiersRaw interface{}, createdAt, updatedAt time.Time) store.ElderlySummary {
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
	return store.ElderlySummary{
		ID:          id,
		Name:        name,
		UserID:      userID,
		AvatarURL:   avatarURL,
		HealthTiers: tiers,
		CreatedAt:   createdAt.Format(time.RFC3339),
		UpdatedAt:   updatedAt.Format(time.RFC3339),
	}
}
