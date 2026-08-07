package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InstitutionHandler struct {
	store store.InstitutionStore
}

func NewInstitutionHandler(s store.InstitutionStore) *InstitutionHandler {
	return &InstitutionHandler{store: s}
}

func (h *InstitutionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.store.ListInstitutions(c.Request.Context(), page, pageSize, c.Query("name"), c.Query("type"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list institutions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items, "page": page, "page_size": pageSize})
}

func (h *InstitutionHandler) Get(c *gin.Context) {
	item, err := h.store.GetInstitution(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get institution"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": item})
}

func (h *InstitutionHandler) Create(c *gin.Context) {
	var inst model.InstitutionSummary
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	inst.ID = uuid.New().String()
	if inst.Status == "" {
		inst.Status = "pending"
	}
	if err := h.store.CreateInstitution(c.Request.Context(), &inst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": inst})
}

func (h *InstitutionHandler) Update(c *gin.Context) {
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateInstitution(c.Request.Context(), c.Param("id"), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "updated"})
}

func (h *InstitutionHandler) Delete(c *gin.Context) {
	if err := h.store.DeleteInstitution(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "deleted"})
}

func (h *InstitutionHandler) CreateAPIKey(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	keyValue, err := h.store.CreateInstitutionAPIKey(c.Request.Context(), c.Param("id"), body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gin.H{"key_id": uuid.New().String(), "key_value": keyValue, "expires": ""}})
}

func (h *InstitutionHandler) RevokeAPIKey(c *gin.Context) {
	if err := h.store.RevokeInstitutionAPIKey(c.Request.Context(), c.Param("id"), c.Param("key_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "revoked"})
}
