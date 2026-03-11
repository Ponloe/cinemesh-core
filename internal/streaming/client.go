package streaming

import (
	"log"
	"net/http"
	"os"
	"time"
)

var streamingClient *http.Client
var streamingBaseURL string

func InitializeStreamingClient() {
	streamingBaseURL = os.Getenv("STREAMING_API_URL")
	if streamingBaseURL == "" {
		log.Println("WARNING: STREAMING_API_URL not set - streaming features will not work")
	} else {
		log.Printf("✓ STREAMING_API_URL loaded: %s", streamingBaseURL)
	}

	streamingClient = &http.Client{
		Timeout: 10 * time.Second,
	}
}