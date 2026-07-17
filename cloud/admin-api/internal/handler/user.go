package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// UserHandler serves user management endpoints.
type UserHandler struct {
	store *store.PostgresStore
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(s *store.PostgresStore) *UserHandler {
	return &UserHandler{store: s}
}

// SetRole updates a user's role.
func (h *UserHandler) SetRole(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.SetUserRole(c.Request.Context(), userID, body.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// List returns a paginated list of users with optional role filter.
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.Query("role")

	users, err := h.store.ListUsers(c.Request.Context(), page, pageSize, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"page":      page,
		"page_size": pageSize,
	})
}
