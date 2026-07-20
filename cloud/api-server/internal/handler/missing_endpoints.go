package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SubscriptionHandler handles subscription queries.
type SubscriptionHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewSubscriptionHandler(pg *store.Postgres, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{pg: pg, log: log}
}

// GET /api/v1/subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}

	subs, total, err := h.pg.ListSubscriptions(c.Request.Context(), userID.(string), page, pageSize)
	if err != nil {
		h.log.Error("list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{
		"subscriptions": subs,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	}})
}

// GET /api/v1/subscriptions/stats
func (h *SubscriptionHandler) Stats(c *gin.Context) {
	// Return tier distribution
	rows, err := h.pg.Pool().Query(c.Request.Context(), `
		SELECT plan_tier, COUNT(*)::int
		FROM subscriptions GROUP BY plan_tier`)
	if err != nil {
		h.log.Error("subscription stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED"})
		return
	}
	defer rows.Close()

	type tierStat struct {
		Tier  string `json:"tier"`
		Count int    `json:"count"`
	}
	var stats []tierStat
	for rows.Next() {
		var s tierStat
		rows.Scan(&s.Tier, &s.Count)
		stats = append(stats, s)
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// UserListHandler provides user management endpoints.
type UserListHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewUserListHandler(pg *store.Postgres, log *zap.Logger) *UserListHandler {
	return &UserListHandler{pg: pg, log: log}
}

// GET /api/v1/users
func (h *UserListHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.pg.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}})
}

// GET /api/v1/users/:id
func (h *UserListHandler) Get(c *gin.Context) {
	id := c.Param("id")
	user, err := h.pg.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": user})
}

// MedicationTakeHandler handles medication confirmation.
type MedicationTakeHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewMedicationTakeHandler(pg *store.Postgres, log *zap.Logger) *MedicationTakeHandler {
	return &MedicationTakeHandler{pg: pg, log: log}
}

// POST /api/v1/medication/:rule_id/take
func (h *MedicationTakeHandler) Take(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	ruleID := c.Param("rule_id")

	elderIDs, err := h.pg.GetElderlyIDsByUserID(c.Request.Context(), userID.(string))
	if err != nil || len(elderIDs) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "No access"})
		return
	}

	// Verify rule belongs to one of user's elderly profiles
	var elderlyID string
	q := `SELECT elderly_id FROM medication_rules WHERE id = $1 AND active = true`
	if err := h.pg.Pool().QueryRow(c.Request.Context(), q, ruleID).Scan(&elderlyID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Medication rule not found"})
		return
	}

	allowed := false
	for _, eid := range elderIDs {
		if eid == elderlyID {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "Not your elderly profile"})
		return
	}

	if err := h.pg.CreateMedTakeRecord(c.Request.Context(), ruleID, elderlyID); err != nil {
		h.log.Error("record med take", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED", "message": "Failed to record medication taken"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Medication recorded as taken"})
}

// AlertHandleHandler handles alert resolution and sharing.
type AlertHandleHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewAlertHandleHandler(pg *store.Postgres, log *zap.Logger) *AlertHandleHandler {
	return &AlertHandleHandler{pg: pg, log: log}
}

// PUT /api/v1/alerts/:id/handle
func (h *AlertHandleHandler) Handle(c *gin.Context) {
	alertID := c.Param("id")
	if err := h.pg.ResolveAlertByID(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Alert not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Alert resolved"})
}

// POST /api/v1/alerts/share-location
func (h *AlertHandleHandler) ShareLocation(c *gin.Context) {
	var req struct {
		ElderlyID string `json:"elderly_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST"})
		return
	}

	loc, err := h.pg.GetLatestLocation(c.Request.Context(), req.ElderlyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "No recent location found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": loc})
}
