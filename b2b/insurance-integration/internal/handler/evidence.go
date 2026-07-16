package handler

import (
	"net/http"

	"eregen.dev/b2b-insurance-integration/internal/model"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EvidenceHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewEvidenceHandler(store *store.Postgres, log *zap.Logger) *EvidenceHandler {
	return &EvidenceHandler{store: store, log: log}
}

// POST /api/v2/b2b/evidence — upload evidence file for a claim
func (h *EvidenceHandler) Upload(c *gin.Context) {
	var req struct {
		ClaimID  string `json:"claim_id" binding:"required"`
		FileType string `json:"file_type" binding:"required"`
		FileName string `json:"file_name" binding:"required"`
		FileURL  string `json:"file_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	file := &model.EvidenceFile{
		ClaimID:  req.ClaimID,
		FileType: req.FileType,
		FileName: req.FileName,
		FileURL:  req.FileURL,
	}

	if err := h.store.AddEvidenceFile(c.Request.Context(), file); err != nil {
		h.log.Error("add evidence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add evidence"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": file})
}

// GET /api/v2/b2b/evidence/claim/:claim_id — get all evidence for a claim
func (h *EvidenceHandler) ListByClaim(c *gin.Context) {
	claimID := c.Param("claim_id")
	files, err := h.store.GetEvidenceForClaim(c.Request.Context(), claimID)
	if err != nil {
		h.log.Error("get evidence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get evidence"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": files})
}
