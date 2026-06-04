package controllers

import (
	"log"
	"net/http"
	"os"
	"time"

	"newyear-api/dto"
	"newyear-api/services"
	"newyear-api/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, int(utils.GetAccessTokenDuration().Seconds()), "/", "", cookieSecure(), true)
	c.SetCookie("refresh_token", refreshToken, int(utils.GetRefreshTokenDuration().Seconds()), "/", "", cookieSecure(), true)
}

func clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", cookieSecure(), true)
	c.SetCookie("refresh_token", "", -1, "/", "", cookieSecure(), true)
}

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя по email и паролю. Пароль хешируется через bcrypt с уникальной солью.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterDTO  true  "Данные для регистрации"
// @Success      201   {object}  map[string]interface{}  "example: {\"message\":\"User registered successfully\",\"user\":{\"id\":\"uuid\",\"email\":\"user@example.com\",\"created_at\":\"2024-01-01T00:00:00Z\"}}"
// @Failure      400   {object}  dto.ErrorDTO  "Неверный формат данных или пользователь уже существует"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var input dto.RegisterDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ac.service.Register(input)
	if err != nil {
		log.Println("REGISTER ERROR:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": dto.UserResponseDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация по email и паролю. Устанавливает HttpOnly cookies: access_token (15 мин) и refresh_token (7 дней).
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginDTO  true  "Данные для входа"
// @Success      200   {object}  dto.MessageDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверный формат данных"
// @Failure      401   {object}  dto.ErrorDTO  "Неверный email или пароль"
// @Router       /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var input dto.LoginDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := ac.service.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	setAuthCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

// Refresh godoc
// @Summary      Обновление пары токенов
// @Description  Обновляет access_token и refresh_token используя refresh_token из cookie. Старые токены отзываются в БД.
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  dto.MessageDTO
// @Failure      401  {object}  dto.ErrorDTO  "Refresh token не найден или недействителен"
// @Router       /auth/refresh [post]
func (ac *AuthController) Refresh(c *gin.Context) {
	oldRefreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}

	oldAccessToken, _ := c.Cookie("access_token")

	accessToken, refreshToken, err := ac.service.RefreshTokens(oldRefreshToken, oldAccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	setAuthCookies(c, accessToken, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens refreshed successfully",
	})
}

// WhoAmI godoc
// @Summary      Данные текущего пользователя
// @Description  Возвращает профиль авторизованного пользователя. Требует валидный access_token в HttpOnly cookie. Нужен фронтенду для проверки статуса авторизации, так как HttpOnly cookies недоступны из JavaScript.
// @Tags         Auth
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.UserResponseDTO
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      404  {object}  dto.ErrorDTO  "Пользователь не найден"
// @Router       /auth/whoami [get]
func (ac *AuthController) WhoAmI(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := ac.service.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

// Logout godoc
// @Summary      Выход из текущей сессии
// @Description  Отзывает текущие access и refresh токены в БД. Очищает HttpOnly cookies.
// @Tags         Auth
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.MessageDTO
// @Failure      500  {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	accessToken, _ := c.Cookie("access_token")
	refreshToken, _ := c.Cookie("refresh_token")

	if err := ac.service.Logout(accessToken, refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// LogoutAll godoc
// @Summary      Выход со всех устройств
// @Description  Отзывает все access и refresh токены пользователя во всех сессиях. Очищает HttpOnly cookies.
// @Tags         Auth
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.MessageDTO
// @Failure      401  {object}  dto.ErrorDTO  "Не авторизован"
// @Failure      500  {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /auth/logout-all [post]
func (ac *AuthController) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := ac.service.LogoutAll(userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout from all devices"})
		return
	}

	clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all devices",
	})
}

// ForgotPassword godoc
// @Summary      Запрос сброса пароля
// @Description  Генерирует одноразовый токен сброса пароля (действует 15 минут). В режиме разработки токен сохраняется в папку tmp-mails/. Ответ всегда одинаковый — чтобы нельзя было узнать, существует ли email в системе.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ForgotPasswordDTO  true  "Email пользователя"
// @Success      200   {object}  dto.MessageDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверный формат email"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var input dto.ForgotPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resetToken, err := ac.service.ForgotPassword(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	if resetToken != "" {
		if err := utils.SaveResetTokenEmail(input.Email, resetToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset instructions"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the account exists, reset instructions have been sent.",
	})
}

// ResetPassword godoc
// @Summary      Установка нового пароля
// @Description  Устанавливает новый пароль по токену из файла tmp-mails/. Токен одноразовый, действует 15 минут. После сброса все сессии пользователя завершаются.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ResetPasswordDTO  true  "Токен и новый пароль"
// @Success      200   {object}  dto.MessageDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверный или просроченный токен"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка"
// @Router       /auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var input dto.ResetPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.service.ResetPassword(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clearAuthCookies(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}
