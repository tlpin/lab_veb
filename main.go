package main

import (
	"log"
	"newyear-api/controllers"
	"newyear-api/database"
	"newyear-api/repositories"
	"newyear-api/routes"
	"newyear-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Подключение к БД
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализация слоев
	collectibleRepo := repositories.NewCollectibleRepository(db)
	collectibleService := services.NewCollectibleService(collectibleRepo)
	collectibleController := controllers.NewCollectibleController(collectibleService)

	// Настройка роутера
	r := gin.Default()

	// Регистрация маршрутов
	routes.SetupRoutes(r, collectibleController)

	// Запуск сервера
	log.Println("Server starting on :4200")
	if err := r.Run(":4200"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
