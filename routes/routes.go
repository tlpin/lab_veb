package routes

import (
	"newyear-api/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, collectibleController *controllers.CollectibleController) {

	api := r.Group("/api")
	{
		api.GET("/items", collectibleController.GetAll)
		api.POST("/items", collectibleController.Create)
		api.GET("/items/:id", collectibleController.GetByID) // ✅ Добавили сюда тоже
	}

	r.GET("/items", collectibleController.GetAll)
	r.POST("/items", collectibleController.Create)
	r.GET("/items/:id", collectibleController.GetByID)
	r.DELETE("/items/:id", collectibleController.Delete)
	r.PUT("/items/:id", collectibleController.Update)
	r.PATCH("/items/:id", collectibleController.Patch)
}
