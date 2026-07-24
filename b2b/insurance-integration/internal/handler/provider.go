package handler

import (
	"net/http"

	"eregen.dev/b2b-insurance-integration/internal/model"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProviderHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewProviderHandler(store *store.Postgres, log *zap.Logger) *ProviderHandler {
	return &ProviderHandler{store: store, log: log}
}

// POST /api/v2/b2b/providers — register a new insurance provider
func (h *ProviderHandler) Create(c *gin.Context) {
	var req model.InsuranceProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateProvider(c.Request.Context(), &req); err != nil {
		h.log.Error("create provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create provider"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/providers — list providers
func (h *ProviderHandler) List(c *gin.Context) {
	page, _ := parseIntParam(c, "page", 1)
	pageSize, _ := parseIntParam(c, "page_size", 20)

	list, total, err := h.store.ListProviders(c.Request.Context(), page, pageSize)
	if err != nil {
		h.log.Error("list providers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list providers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": list, "total": total, "page": page})
}

// GET /api/v2/b2b/providers/:id — get provider details
func (h *ProviderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	provider, err := h.store.GetProviderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": provider})
}

// PUT /api/v2/b2b/providers/:id — update provider
func (h *ProviderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.InsuranceProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	provider, err := h.store.GetProviderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.Code != "" {
		provider.Code = req.Code
	}
	if req.APIEndpoint != "" {
		provider.APIEndpoint = req.APIEndpoint
	}
	provider.Active = req.Active

	if err := h.store.UpdateProvider(c.Request.Context(), provider); err != nil {
		h.log.Error("update provider", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": provider})
}
