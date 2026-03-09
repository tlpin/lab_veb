package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	router.GET("/info", func(c *gin.Context) {
		now := time.Now()
		newYear := time.Date(now.Year()+1, time.January, 1, 0, 0, 0, 0, time.Local)
		days := int(newYear.Sub(now).Hours() / 24)

		c.JSON(200, gin.H{
			"days_before_new_year": days,
		})
	})

	router.Run(":4200")

}
