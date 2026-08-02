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
	store store.Store
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(s store.Store) *UserHandler {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	role := c.Query("role")

	users, err := h.store.ListUsers(c.Request.Context(), page, pageSize, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create adds a new user.
func (h *UserHandler) Create(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Role     string `json:"role" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	id, err := h.store.CreateUser(c.Request.Context(), body.Name, body.Email, body.Phone, body.Role, body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id, "name": body.Name, "role": body.Role}})
}

// Update modifies an existing user.
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateUser(c.Request.Context(), userID, body.Name, body.Email, body.Phone, body.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

// Delete removes a user.
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")
	if err := h.store.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
