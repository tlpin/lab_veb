package controllers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

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

func getUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return "", false
	}

	return userIDStr, true
}

// Create godoc
// @Summary      Создать коллекционный предмет
// @Description  Создаёт новый предмет и привязывает его к авторизованному пользователю.
// @Tags         Items
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        body  body      dto.CreateCollectibleDTO  true  "Данные предмета"
// @Success      201   {object}  dto.MessageDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверный формат данных"
// @Failure      401   {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /items [post]
func (cc *CollectibleController) Create(c *gin.Context) {
	var input dto.CreateCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	if err := cc.service.Create(userID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

// GetAll godoc
// @Summary      Получить список предметов
// @Description  Возвращает список коллекционных предметов текущего пользователя с пагинацией.
// @Tags         Items
// @Produce      json
// @Security     CookieAuth
// @Param        page   query  int  false  "Номер страницы (по умолчанию 1)"   default(1)   minimum(1)
// @Param        limit  query  int  false  "Предметов на странице (по умолчанию 10, максимум 100)"  default(10)  minimum(1)  maximum(100)
// @Success      200  {object}  dto.CollectibleListResponseDTO
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      500  {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /items [get]
func (cc *CollectibleController) GetAll(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	items, total, err := cc.service.GetAll(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

// GetByID godoc
// @Summary      Получить предмет по ID
// @Description  Возвращает коллекционный предмет по UUID. Доступен только владельцу.
// @Tags         Items
// @Produce      json
// @Security     CookieAuth
// @Param        id  path  string  true  "UUID предмета"  format(uuid)  example(550e8400-e29b-41d4-a716-446655440000)
// @Success      200  {object}  dto.CollectibleResponseDTO
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      403  {object}  dto.ErrorDTO  "Нет доступа к чужому предмету"
// @Failure      404  {object}  dto.ErrorDTO  "Предмет не найден"
// @Router       /items/{id} [get]
func (cc *CollectibleController) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	item, err := cc.service.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, services.ErrCollectibleForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if errors.Is(err, services.ErrCollectibleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Update godoc
// @Summary      Полное обновление предмета
// @Description  Полностью заменяет данные предмета. Доступно только владельцу.
// @Tags         Items
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        id    path  string                    true  "UUID предмета"  format(uuid)
// @Param        body  body  dto.UpdateCollectibleDTO  true  "Новые данные предмета"
// @Success      200  {object}  dto.CollectibleResponseDTO
// @Failure      400  {object}  dto.ErrorDTO  "Неверный формат данных"
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      403  {object}  dto.ErrorDTO  "Нет доступа к чужому предмету"
// @Failure      404  {object}  dto.ErrorDTO  "Предмет не найден"
// @Router       /items/{id} [put]
func (cc *CollectibleController) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	var input dto.UpdateCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	updated, err := cc.service.Update(id, userID, input)
	if err != nil {
		if errors.Is(err, services.ErrCollectibleForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if errors.Is(err, services.ErrCollectibleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// Delete godoc
// @Summary      Удалить предмет (Soft Delete)
// @Description  Мягкое удаление — запись остаётся в БД с заполненным полем deleted_at. Доступно только владельцу.
// @Tags         Items
// @Produce      json
// @Security     CookieAuth
// @Param        id  path  string  true  "UUID предмета"  format(uuid)
// @Success      204  {string}  string  "No Content — предмет удалён"
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      403  {object}  dto.ErrorDTO  "Нет доступа к чужому предмету"
// @Failure      404  {object}  dto.ErrorDTO  "Предмет не найден"
// @Router       /items/{id} [delete]
func (cc *CollectibleController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	if err := cc.service.Delete(id, userID); err != nil {
		if errors.Is(err, services.ErrCollectibleForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if errors.Is(err, services.ErrCollectibleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Patch godoc
// @Summary      Частичное обновление предмета
// @Description  Обновляет только переданные поля предмета. Доступно только владельцу.
// @Tags         Items
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        id    path  string                   true  "UUID предмета"  format(uuid)
// @Param        body  body  dto.PatchCollectibleDTO  true  "Поля для обновления"
// @Success      200  {object}  dto.CollectibleResponseDTO
// @Failure      400  {object}  dto.ErrorDTO  "Неверный формат данных"
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      403  {object}  dto.ErrorDTO  "Нет доступа к чужому предмету"
// @Failure      404  {object}  dto.ErrorDTO  "Предмет не найден"
// @Router       /items/{id} [patch]
func (cc *CollectibleController) Patch(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	var input dto.PatchCollectibleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	updated, err := cc.service.Patch(id, userID, input)
	if err != nil {
		if errors.Is(err, services.ErrCollectibleForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if errors.Is(err, services.ErrCollectibleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to patch item"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
