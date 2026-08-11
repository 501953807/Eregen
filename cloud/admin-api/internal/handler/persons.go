package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

type PersonHandler struct {
	store store.PersonStore
}

func NewPersonHandler(s store.PersonStore) *PersonHandler {
	return &PersonHandler{store: s}
}

func (h *PersonHandler) List(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	page := 1
	pageSize := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if pageSizeStr != "" {
		if p, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = p
		}
	}
	page, pageSize, _ = validation.ValidatePagination(page, pageSize, 100)

	persons, err := h.store.ListPersons(c.Request.Context(), page, pageSize,
		c.Query("business_chain"), c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": persons, "page": page, "page_size": pageSize})
}

func (h *PersonHandler) Get(c *gin.Context) {
	id := c.Param("id")
	p, err := h.store.GetPerson(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": p})
}

func (h *PersonHandler) Create(c *gin.Context) {
	var body struct {
		IDCard           string `json:"id_card" binding:"required"`
		Name             string `json:"name" binding:"required"`
		Gender           int    `json:"gender"`
		BirthDate        string `json:"birth_date"`
		Phone            string `json:"phone"`
		EmergencyContact string `json:"emergency_contact"`
		Address          string `json:"address"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p := &model.Person{
		IDCard:           body.IDCard,
		Name:             body.Name,
		Gender:           body.Gender,
		Phone:            body.Phone,
		EmergencyContact: body.EmergencyContact,
		Address:          body.Address,
		Status:           "active",
	}
	if body.BirthDate != "" {
		t, err := time.Parse("2006-01-02", body.BirthDate)
		if err == nil {
			p.BirthDate = &t
		}
	}
	if err := h.store.CreatePerson(c.Request.Context(), p); err != nil {
		log.Printf("CreatePerson failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": p})
}

func (h *PersonHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	delete(body, "id")
	delete(body, "id_card")
	delete(body, "created_at")
	if err := h.store.UpdatePerson(c.Request.Context(), id, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

func (h *PersonHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeletePerson(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// CreateProfile creates a business-chain profile for a person.
func (h *PersonHandler) CreateProfile(c *gin.Context) {
	var pp model.PersonProfile
	if err := c.ShouldBindJSON(&pp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateProfile(c.Request.Context(), &pp); err != nil {
		log.Printf("CreateProfile failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": pp})
}

// GetProfile retrieves a business-chain profile for a person.
func (h *PersonHandler) GetProfile(c *gin.Context) {
	personID := c.Param("id")
	chain := c.Query("chain")
	pp, err := h.store.GetProfile(c.Request.Context(), personID, model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": pp})
}

// AssignWelfareTag assigns a welfare tag to a person.
func (h *PersonHandler) AssignWelfareTag(c *gin.Context) {
	var wt model.PersonWelfareTag
	if err := c.ShouldBindJSON(&wt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.AssignPersonWelfareTag(c.Request.Context(), &wt); err != nil {
		log.Printf("AssignWelfareTag failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// RevokeWelfareTag revokes a welfare tag from a person.
func (h *PersonHandler) RevokeWelfareTag(c *gin.Context) {
	personID := c.Param("id")
	tagCode := c.Param("tag_code")
	if err := h.store.RevokePersonWelfareTag(c.Request.Context(), personID, tagCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// ListWelfareTags returns all welfare tags for a person.
func (h *PersonHandler) ListWelfareTags(c *gin.Context) {
	personID := c.Param("id")
	tags, err := h.store.ListPersonWelfareTags(c.Request.Context(), personID)
	if err != nil {
		log.Printf("ListWelfareTags failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": tags})
}
