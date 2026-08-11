package handler

import (
	"log"
	"net/http"
	"strconv"

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
	page, pageSize, err := validation.ValidatePagination(
		strconv.Atoi(c.Query("page")),
		strconv.Atoi(c.Query("page_size")), 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}
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
	p := &Person{
		IDCard:           body.IDCard,
		Name:             body.Name,
		Gender:           body.Gender,
		Phone:            body.Phone,
		EmergencyContact: body.EmergencyContact,
		Address:          body.Address,
		Status:           "active",
	}
	if body.BirthDate != "" {
		t, err := parseTimeBody(body.BirthDate)
		if err == nil {
			p.BirthDate = t
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
