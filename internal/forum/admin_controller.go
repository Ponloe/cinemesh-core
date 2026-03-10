package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteReplyHandler handles reply deletion
func DeleteReplyHandler(c *gin.Context) {
	replyID := c.Param("reply_id")

	// Get admin token from context (set by auth middleware)
	token, _ := c.Get("token")

	client := GetClient()
	client.Token = token.(string)

	if err := client.DeleteReply(replyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete reply: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reply deleted successfully",
	})
}

// PinThreadHandler pins/unpins a thread
func PinThreadHandler(c *gin.Context) {
	threadID := c.Param("thread_id")

	var req struct {
		IsPinned bool `json:"is_pinned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, _ := c.Get("token")
	client := GetClient()
	client.Token = token.(string)

	if err := client.UpdateThread(threadID, map[string]interface{}{
		"is_pinned": req.IsPinned,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to pin thread: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thread updated successfully",
	})
}

// DeleteThreadHandler handles thread deletion
func DeleteThreadHandler(c *gin.Context) {
	threadSlug := c.Param("thread_slug")

	// Get admin token from context
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	client := GetClient()
	client.Token = token.(string)

	if err := client.DeleteThread(threadSlug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete thread: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Thread deleted successfully",
	})
}
