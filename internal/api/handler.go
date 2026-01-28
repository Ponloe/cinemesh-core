package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIDocsHandler(c *gin.Context) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c.HTML(http.StatusOK, "api_docs.html", gin.H{
		"baseURL": baseURL,
	})
}
