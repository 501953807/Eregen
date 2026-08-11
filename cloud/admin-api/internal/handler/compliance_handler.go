package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"strconv"

	"github.com/gin-gonic/gin"
)

// ComplianceHandler manages compliance rules and checks.
type ComplianceHandler struct {
	store store.ComplianceStore
}

func NewComplianceHandler(s store.ComplianceStore) *ComplianceHandler {
	return &ComplianceHandler{store: s}
}

// CreateRule adds a compliance rule.
func (h *ComplianceHandler) CreateRule(c *gin.Context) {
	var r model.ComplianceRule
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateComplianceRule(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// ListRules returns compliance rules for a chain.
func (h *ComplianceHandler) ListRules(c *gin.Context) {
	chain := c.Query("chain")
	rules, err := h.store.ListComplianceRules(c.Request.Context(), model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// RunCheck executes a compliance check for a person.
func (h *ComplianceHandler) RunCheck(c *gin.Context) {
	var body struct {
		RuleCode string `json:"rule_code" binding:"required"`
		PersonID string `json:"person_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	check, err := h.store.RunComplianceCheck(c.Request.Context(), body.RuleCode, body.PersonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": check})
}

// ListChecks returns compliance checks for a person.
func (h *ComplianceHandler) ListChecks(c *gin.Context) {
	personID := c.Param("personId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	checks, err := h.store.ListComplianceChecks(c.Request.Context(), personID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": checks})
}

// ReviewCheck marks a compliance check as reviewed.
func (h *ComplianceHandler) ReviewCheck(c *gin.Context) {
	checkID := c.Param("checkId")
	var body struct {
		ReviewerID string `json:"reviewer_id" binding:"required"`
		Result     string `json:"result" binding:"required"`
		Notes      string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.ReviewCheck(c.Request.Context(), checkID, body.ReviewerID, body.Result, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}
