package handler

import (
	"net/http"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

type LifecycleHandler struct {
	store store.PersonStore
}

func NewLifecycleHandler(s store.PersonStore) *LifecycleHandler {
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

	// Validate status transition
	validTransitions := map[string]map[string]bool{
		"self": {
			"pending":   true,
			"active":    true,
			"suspended": true,
			"cancelled": true,
		},
		"hospital": {
			"pending":      true,
			"admitted":     true,
			"in_treatment": true,
			"discharged":   true,
			"archived":     true,
		},
		"community": {
			"pending":     true,
			"certified":   true,
			"active":      true,
			"suspended":   true,
			"deactivated": true,
		},
	}

	chainTransitions, ok := validTransitions[body.BusinessChain]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid business chain"})
		return
	}
	if !chainTransitions[body.NewStatus] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
		return
	}

	// Update the person profile status
	updates := map[string]any{
		"status":      body.NewStatus,
		"updated_at":  time.Now().Format("2006-01-02 15:04:05"),
	}
	if body.Reason != "" {
		updates["reason"] = body.Reason
	}

	if err := h.store.UpdateProfile(c.Request.Context(), &model.PersonProfile{
		PersonID:      personID,
		BusinessChain: body.BusinessChain,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "status updated"})
}

// LinkPerson creates a cross-chain link between two persons.
func (h *LifecycleHandler) LinkPerson(c *gin.Context) {
	var body struct {
		PersonID1        string `json:"person_id_1" binding:"required"`
		PersonID2        string `json:"person_id_2" binding:"required"`
		BusinessChain1   string `json:"business_chain_1" binding:"required"`
		BusinessChain2   string `json:"business_chain_2" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Update linked_person_id in both profiles
	profile1 := &model.PersonProfile{
		PersonID:      body.PersonID1,
		BusinessChain: body.BusinessChain1,
		LinkedPersonID: body.PersonID2,
	}
	profile2 := &model.PersonProfile{
		PersonID:      body.PersonID2,
		BusinessChain: body.BusinessChain2,
		LinkedPersonID: body.PersonID1,
	}

	if err := h.store.UpdateProfile(c.Request.Context(), profile1); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile 1"})
		return
	}
	if err := h.store.UpdateProfile(c.Request.Context(), profile2); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile 2"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "persons linked"})
}
