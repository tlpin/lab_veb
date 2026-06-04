package main

import (
	"log"
	"os"

	"newyear-api/cache"
	"newyear-api/controllers"
	"newyear-api/database"
	_ "newyear-api/docs"
	"newyear-api/repositories"
	"newyear-api/routes"
	"newyear-api/services"
	"newyear-api/storage"

	"github.com/gin-gonic/gin"
)

// @title NewYear API
// @version 1.0
// @description REST API для управления коллекционными предметами с JWT авторизацией, OAuth 2.0, MinIO и Swagger.
// @host localhost:4200
// @BasePath /

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description JWT access token через HttpOnly cookie.

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	redisCache := cache.NewRedisCache()
	if err := redisCache.Ping(); err != nil {
		log.Printf("Warning: Redis unavailable: %v", err)
	} else {
		log.Println("Redis connected successfully")
	}

	minioStorage, err := storage.NewMinIOStorage()
	if err != nil {
		log.Fatal("Failed to connect to MinIO:", err)
	}
	log.Println("MinIO connected successfully")

	collectibleRepo := repositories.NewCollectibleRepository(db)
	collectibleService := services.NewCollectibleService(collectibleRepo, redisCache)
	collectibleController := controllers.NewCollectibleController(collectibleService)

	authRepo := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepo, redisCache)
	authController := controllers.NewAuthController(authService)
	oauthController := controllers.NewOAuthController(authService)

	fileRepo := repositories.NewFileRepository(db)
	fileService := services.NewFileService(fileRepo, minioStorage, redisCache)
	fileController := controllers.NewFileController(fileService)

	profileService := services.NewProfileService(authRepo, fileRepo)
	profileController := controllers.NewProfileController(profileService)

	r := gin.Default()

	routes.SetupRoutes(r, collectibleController, authController, oauthController, authService, fileController, profileController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4200"
	}

	log.Println("Server starting on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
