package controllers

import (
	"net/http"
	"net/url"
	"os"

	"newyear-api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OAuthController struct {
	service *services.AuthService
}

func NewOAuthController(service *services.AuthService) *OAuthController {
	return &OAuthController{service: service}
}

func (oc *OAuthController) RedirectToProvider(c *gin.Context) {
	provider := c.Param("provider")

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
		return
	}

	state := uuid.New().String()
	stateCookieName := "oauth_state_" + provider

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		stateCookieName,
		state,
		600,
		"/",
		"",
		cookieSecure(),
		true,
	)

	clientID := os.Getenv("YANDEX_CLIENT_ID")
	callbackURL := os.Getenv("YANDEX_CALLBACK_URL")

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("redirect_uri", callbackURL)
	params.Add("state", state)
	params.Add("scope", "login:email login:info")

	authURL := "https://oauth.yandex.ru/authorize?" + params.Encode()

	c.Redirect(http.StatusFound, authURL)
}

func (oc *OAuthController) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	if provider != "yandex" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported provider"})
		return
	}

	stateFromURL := c.Query("state")
	if stateFromURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State is required"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	stateCookieName := "oauth_state_" + provider
	stateFromCookie, err := c.Cookie(stateCookieName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State cookie not found"})
		return
	}

	if stateFromURL != stateFromCookie {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	accessToken, refreshToken, err := oc.service.YandexLogin(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login via OAuth"})
		return
	}

	setAuthCookies(c, accessToken, refreshToken)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(stateCookieName, "", -1, "/", "", cookieSecure(), true)

	successRedirect := os.Getenv("OAUTH_SUCCESS_REDIRECT_URL")
	if successRedirect == "" {
		successRedirect = "/auth/whoami"
	}

	c.Redirect(http.StatusFound, successRedirect)
}
