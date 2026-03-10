package forum

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListTopicsHandler renders the topics page
func ListTopicsHandler(c *gin.Context) {
	client := GetClient()
	topics, err := client.GetTopics()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch topics: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "forum_topics.html", gin.H{
		"topics": topics,
	})
}

// ListThreadsHandler renders threads for a topic
func ListThreadsHandler(c *gin.Context) {
	topicSlug := c.Param("slug")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	client := GetClient()
	threads, pagination, err := client.GetThreads(topicSlug, page, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch threads: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "forum_threads.html", gin.H{
		"topicSlug":  topicSlug,
		"threads":    threads,
		"pagination": pagination,
	})
}

// ViewThreadHandler renders a single thread with replies
func ViewThreadHandler(c *gin.Context) {
	threadSlug := c.Param("thread_slug")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	client := GetClient()

	// Get thread details
	thread, err := client.GetThreadBySlug(threadSlug)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch thread: " + err.Error(),
		})
		return
	}

	// Get replies
	replies, err := client.GetReplies(threadSlug, page, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch replies: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "forum_thread_view.html", gin.H{
		"thread":  thread,
		"replies": replies,
	})
}

// NewTopicFormHandler renders the create topic form
func NewTopicFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "forum_topic_form.html", gin.H{})
}

// CreateTopicHandler handles topic creation
func CreateTopicHandler(c *gin.Context) {
	var req struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
		Icon        string `form:"icon"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "forum_topic_form.html", gin.H{
			"error": "Topic name is required",
		})
		return
	}

	// Get admin token from context
	token, exists := c.Get("token")
	if !exists {
		c.HTML(http.StatusUnauthorized, "forum_topic_form.html", gin.H{
			"error": "Unauthorized",
		})
		return
	}

	client := GetClient()
	client.Token = token.(string)

	_, err := client.CreateTopic(req.Name, req.Description, req.Icon)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "forum_topic_form.html", gin.H{
			"error": "Failed to create topic: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/forum")
}

// DeleteTopicHandler handles topic deletion
func DeleteTopicHandler(c *gin.Context) {
	topicSlug := c.Param("slug")

	// Get admin token from context
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	client := GetClient()
	client.Token = token.(string)

	if err := client.DeleteTopic(topicSlug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete topic: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic deleted successfully",
	})
}

// EditTopicFormHandler renders the edit topic form
func EditTopicFormHandler(c *gin.Context) {
	topicSlug := c.Param("slug")

	client := GetClient()
	topic, err := client.GetTopicBySlug(topicSlug)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch topic: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "forum_topic_form.html", gin.H{
		"topic": topic,
		"mode":  "edit",
	})
}

// UpdateTopicHandler handles topic updates
func UpdateTopicHandler(c *gin.Context) {
	topicSlug := c.Param("slug")

	var req struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
		Icon        string `form:"icon"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "forum_topic_form.html", gin.H{
			"error": "Topic name is required",
		})
		return
	}

	// Get admin token from context
	token, exists := c.Get("token")
	if !exists {
		c.HTML(http.StatusUnauthorized, "forum_topic_form.html", gin.H{
			"error": "Unauthorized",
		})
		return
	}

	client := GetClient()
	client.Token = token.(string)

	_, err := client.UpdateTopic(topicSlug, req.Name, req.Description, req.Icon)
	if err != nil {
		// Fetch topic again to repopulate form
		topic, _ := client.GetTopicBySlug(topicSlug)
		c.HTML(http.StatusInternalServerError, "forum_topic_form.html", gin.H{
			"error": "Failed to update topic: " + err.Error(),
			"topic": topic,
			"mode":  "edit",
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/forum")
}
