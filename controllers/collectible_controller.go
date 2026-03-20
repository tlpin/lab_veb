package controllers

import (
	"fmt"
	"math"
	"net/http"

	"newyear-api/dto"
	"newyear-api/services"

	"github.com/gin-gonic/gin"
)

type CollectibleController struct {
	service *services.CollectibleService
}

func NewCollectibleController(service *services.CollectibleService) *CollectibleController {
	return &CollectibleController{service: service}
}

// ==================== CREATE ====================
func (cc *CollectibleController) Create(c *gin.Context) {
	var input dto.CreateCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := cc.service.Create(input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

// ==================== READ ====================
func (cc *CollectibleController) GetAll(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 10
	fmt.Sscanf(pageStr, "%d", &page)
	fmt.Sscanf(limitStr, "%d", &limit)

	items, total, err := cc.service.GetAll(page, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(200, gin.H{
		"data": items,
		"meta": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

func (cc *CollectibleController) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := cc.service.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	c.JSON(200, item)
}

// ==================== UPDATE ====================
func (cc *CollectibleController) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID cannot be empty"})
		return
	}

	var input dto.UpdateCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updated, err := cc.service.Update(id, input)
	if err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(200, updated)
}

// ==================== DELETE ====================
func (cc *CollectibleController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID cannot be empty"})
		return
	}

	if err := cc.service.Delete(id); err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	c.Status(204)
}

// ==================== PATCH ====================
func (cc *CollectibleController) Patch(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "ID cannot be empty"})
		return
	}

	var input dto.PatchCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updated, err := cc.service.Patch(id, input)
	if err != nil {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(200, updated)
}
