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
