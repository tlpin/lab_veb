package routes

import (
	"net/http"
	"os"

	"newyear-api/controllers"
	"newyear-api/middleware"
	"newyear-api/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(
	r *gin.Engine,
	collectibleController *controllers.CollectibleController,
	authController *controllers.AuthController,
	oauthController *controllers.OAuthController,
	authService *services.AuthService,
	fileController *controllers.FileController,
	profileController *controllers.ProfileController,
) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "production" {
		r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		r.GET("/api/docs/*any", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		})
	}

	auth := r.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.Refresh)
		auth.POST("/forgot-password", authController.ForgotPassword)
		auth.POST("/reset-password", authController.ResetPassword)
		auth.GET("/oauth/:provider", oauthController.RedirectToProvider)
		auth.GET("/oauth/:provider/callback", oauthController.OAuthCallback)
	}

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.GET("/auth/whoami", authController.WhoAmI)
		protected.POST("/auth/logout", authController.Logout)
		protected.POST("/auth/logout-all", authController.LogoutAll)

		protected.GET("/items", collectibleController.GetAll)
		protected.POST("/items", collectibleController.Create)
		protected.GET("/items/:id", collectibleController.GetByID)
		protected.PUT("/items/:id", collectibleController.Update)
		protected.PATCH("/items/:id", collectibleController.Patch)
		protected.DELETE("/items/:id", collectibleController.Delete)

		protected.POST("/files", fileController.Upload)
		protected.GET("/files/:fileId", fileController.Download)
		protected.DELETE("/files/:fileId", fileController.Delete)

		protected.GET("/profile", profileController.GetProfile)
		protected.POST("/profile", profileController.UpdateProfile)
	}
}
