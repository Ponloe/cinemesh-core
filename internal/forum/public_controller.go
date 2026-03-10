package forum

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListTopicsPublicHandler exposes topics via public API
func ListTopicsPublicHandler(c *gin.Context) {
	client := GetClient()
	topics, err := client.GetTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch topics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    topics,
	})
}

// GetThreadsPublicHandler returns threads for a topic
func GetThreadsPublicHandler(c *gin.Context) {
	topicSlug := c.Param("slug")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	client := GetClient()
	threads, pagination, err := client.GetThreads(topicSlug, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch threads",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       threads,
		"pagination": pagination,
	})
}
