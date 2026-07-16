package handler

import (
	"net/http"

	"eregen.dev/b2b-insurance-integration/internal/model"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ClaimHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewClaimHandler(store *store.Postgres, log *zap.Logger) *ClaimHandler {
	return &ClaimHandler{store: store, log: log}
}

// POST /api/v2/b2b/claims — create a new insurance claim
func (h *ClaimHandler) Create(c *gin.Context) {
	var req model.InsuranceClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateClaim(c.Request.Context(), &req); err != nil {
		h.log.Error("create claim", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create claim"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/claims/:id — get claim details
func (h *ClaimHandler) Get(c *gin.Context) {
	claimID := c.Param("id")
	claim, err := h.store.GetClaimByID(c.Request.Context(), claimID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "claim not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": claim})
}

// PUT /api/v2/b2b/claims/:id/status — update claim status
func (h *ClaimHandler) UpdateStatus(c *gin.Context) {
	claimID := c.Param("id")
	var req struct {
		Status model.ClaimStatus `json:"status" binding:"required"`
		Notes  string            `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.store.UpdateClaimStatus(c.Request.Context(), claimID, req.Status, req.Notes); err != nil {
		h.log.Error("update claim status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update claim"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Claim status updated"})
}

// GET /api/v2/b2b/claims/elderly/:elderly_id — get all claims for an elderly person
func (h *ClaimHandler) GetForElderly(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	claims, err := h.store.GetClaimsForElderly(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("get claims", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get claims"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": claims})
}
