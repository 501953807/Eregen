// Package response provides unified API response helpers.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success returns a standardized success response.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "success",
		"data":    data,
	})
}

// Created returns a 201 response with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{
		"code":    "OK",
		"message": "created",
		"data":    data,
	})
}

// Error returns a standardized error response.
func Error(c *gin.Context, httpStatus int, code string, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
	})
}

// BadRequest returns a 400 error.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "INVALID_REQUEST", message)
}

// Unauthorized returns a 401 error.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden returns a 403 error.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound returns a 404 error.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict returns a 409 error.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalError returns a 500 error.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
