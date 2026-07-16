package handler

import (
	"net/http"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LinkHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewLinkHandler(store *store.Postgres, log *zap.Logger) *LinkHandler {
	return &LinkHandler{store: store, log: log}
}

// POST /api/v2/b2b/links — link elderly to institution
func (h *LinkHandler) Create(c *gin.Context) {
	var req model.ElderlyInstitutionLink
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.LinkElderlyToInstitution(c.Request.Context(), &req); err != nil {
		h.log.Error("link elderly to institution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/institutions/:id/links — list elderly linked to an institution
func (h *LinkHandler) ListByInstitution(c *gin.Context) {
	instID := c.Param("id")
	links, err := h.store.GetActiveLinksForInstitution(c.Request.Context(), instID)
	if err != nil {
		h.log.Error("list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": links})
}

// GET /api/v2/b2b/elderly/:id/links — list institutions linked to an elderly person
func (h *LinkHandler) ListByElderly(c *gin.Context) {
	elderlyID := c.Param("id")
	links, err := h.store.GetActiveLinksForElderly(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": links})
}
