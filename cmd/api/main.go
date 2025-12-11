package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Create a default Gin router
	// This creates a router with default middleware (Logger and Recovery)
	r := gin.Default()

	// 2. Define a route handler
	// GET request to "/ping" returns a JSON response
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "success",
		})
	})

	// 3. Start the server
	// By default, it runs on :8080. You can pass a string like ":3000" to change the port.
	r.Run()
}
