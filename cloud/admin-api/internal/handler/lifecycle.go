package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

type LifecycleHandler struct {
	store store.LifecycleStore
}

func NewLifecycleHandler(s store.LifecycleStore) *LifecycleHandler {
	return &LifecycleHandler{store: s}
}

// TransitionStatus handles status transitions for a person's business chain profile.
func (h *LifecycleHandler) TransitionStatus(c *gin.Context) {
	personID := c.Param("id")
	var body struct {
		BusinessChain string `json:"business_chain" binding:"required"`
		NewStatus     string `json:"new_status" binding:"required"`
		Reason        string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.TransitionStatus(c.Request.Context(), personID, model.BusinessChain(body.BusinessChain), body.NewStatus, body.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "status updated"})
}

// GetPersonStatus returns the current status of a person in a business chain.
func (h *LifecycleHandler) GetPersonStatus(c *gin.Context) {
	personID := c.Param("id")
	chain := c.Query("chain")
	status, err := h.store.GetPersonStatus(c.Request.Context(), personID, model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{"person_id": personID, "chain": chain, "status": status}})
}

// GetStatusHistory returns the status transition history for a person in a business chain.
func (h *LifecycleHandler) GetStatusHistory(c *gin.Context) {
	personID := c.Param("id")
	chain := model.BusinessChain(c.Query("chain"))
	if chain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chain query parameter is required"})
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	history, err := h.store.GetStatusHistory(c.Request.Context(), personID, chain, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": history})
}

// LinkPerson creates a cross-chain link between two persons.
func (h *LifecycleHandler) LinkPerson(c *gin.Context) {
	var body struct {
		PersonID1      string `json:"person_id_1" binding:"required"`
		PersonID2      string `json:"person_id_2" binding:"required"`
		BusinessChain1 string `json:"business_chain_1" binding:"required"`
		BusinessChain2 string `json:"business_chain_2" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.LinkPersons(c.Request.Context(), body.PersonID1, body.PersonID2,
		model.BusinessChain(body.BusinessChain1), model.BusinessChain(body.BusinessChain2)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "persons linked"})
}
