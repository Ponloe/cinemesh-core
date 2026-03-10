package forum

import (
	"os"
	"sync"
)

var (
	client *ForumClient
	once   sync.Once
)

// InitializeForumClient initializes the global forum client
func InitializeForumClient() {
	once.Do(func() {
		baseURL := os.Getenv("FORUM_API_URL")
		if baseURL == "" {
			baseURL = "http://localhost:4000"
		}

		// Token can be set dynamically per request
		client = NewForumClient(baseURL, "")
	})
}

// GetClient returns the global forum client
func GetClient() *ForumClient {
	return client
}
