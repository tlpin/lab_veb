package controllers

import (
	"net/http"
	"time"

	"newyear-api/dto"
	"newyear-api/services"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	service *services.ProfileService
}

func NewProfileController(service *services.ProfileService) *ProfileController {
	return &ProfileController{service: service}
}

// GetProfile godoc
// @Summary      Получить профиль пользователя
// @Description  Возвращает профиль текущего авторизованного пользователя
// @Tags         Profile
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.ProfileResponseDTO
// @Failure      401  {object}  map[string]interface{}
// @Router       /profile [get]
func (pc *ProfileController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := pc.service.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponseDTO{
		ID:           user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		Bio:          user.Bio,
		AvatarFileID: user.AvatarFileID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}

// UpdateProfile godoc
// @Summary      Обновить профиль пользователя
// @Description  Обновляет профиль текущего пользователя
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        request body dto.UpdateProfileDTO true "Данные для обновления"
// @Success      200  {object}  dto.ProfileResponseDTO
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /profile [post]
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.UpdateProfileDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := pc.service.UpdateProfile(userID.(string), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponseDTO{
		ID:           user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		Bio:          user.Bio,
		AvatarFileID: user.AvatarFileID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
	})
}
