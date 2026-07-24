package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type InstitutionHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewInstitutionHandler(store *store.Postgres, log *zap.Logger) *InstitutionHandler {
	return &InstitutionHandler{store: store, log: log}
}

// POST /api/v2/b2b/institutions — create new institution
func (h *InstitutionHandler) Create(c *gin.Context) {
	var req model.Institution
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateInstitution(c.Request.Context(), &req); err != nil {
		h.log.Error("create institution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create institution"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/institutions — list all institutions
func (h *InstitutionHandler) List(c *gin.Context) {
	page, _ := parseIntParam(c, "page", 1)
	pageSize, _ := parseIntParam(c, "page_size", 20)

	list, total, err := h.store.ListInstitutions(c.Request.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list institutions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list institutions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": list, "total": total, "page": page})
}

// GET /api/v2/b2b/institutions/:id — get one institution
func (h *InstitutionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	inst, err := h.store.GetInstitutionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "institution not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": inst})
}

// PUT /api/v2/b2b/institutions/:id — update institution
func (h *InstitutionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.Institution
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.UpdateInstitution(c.Request.Context(), id, &req); err != nil {
		h.log.Error("update institution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update institution"})
		return
	}
	req.ID = id
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": req})
}

// POST /api/v2/b2b/institutions/:id/api-keys — generate API key for institution
func (h *InstitutionHandler) CreateAPIKey(c *gin.Context) {
	instID := c.Param("id")
	var req struct {
		Name      string `json:"name" binding:"required"`
		ExpiresIn int    `json:"expires_in"` // days
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Generate random key and hash it with bcrypt
	rawKey := generateRandomKey()
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("generate bcrypt hash", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate secure key"})
		return
	}

	key := &model.InstitutionAPIKey{
		InstitutionID: instID,
		Name:          req.Name,
		ExpiresAt:     time.Now().AddDate(0, 0, req.ExpiresIn),
		Active:        true,
		KeyHash:       string(hashedKey),
	}

	if err := h.store.CreateAPIKey(c.Request.Context(), key); err != nil {
		h.log.Error("create api key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gin.H{
		"key_id":    key.ID,
		"key_value": rawKey, // only shown once — caller must save this
		"expires":   key.ExpiresAt,
	}})
}

func parseIntParam(c *gin.Context, key string, defaultVal int) (int, bool) {
	v := c.Query(key)
	if v == "" {
		return defaultVal, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal, false
	}
	return n, true
}

func generateRandomKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "ek_" + hex.EncodeToString(b)
}
