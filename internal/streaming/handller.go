package streaming

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// proxy forwards the request to Laravel and pipes the response back
func proxy(c *gin.Context, method, laravelPath string, body io.Reader) {
	url := strings.TrimRight(streamingBaseURL, "/") + laravelPath

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := streamingClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("streaming service unavailable: %v", err)})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	c.Data(resp.StatusCode, "application/json", respBody)
}

// GET /api/public/streaming/movies
func ListMoviesHandler(c *gin.Context) {
	query := c.Request.URL.RawQuery
	path := "/api/movies"
	if query != "" {
		path += "?" + query
	}
	proxy(c, http.MethodGet, path, nil)
}

// GET /api/public/streaming/movies/:id
func GetMovieHandler(c *gin.Context) {
	id := c.Param("id")
	proxy(c, http.MethodGet, "/api/movies/"+id, nil)
}

// PUT /api/public/streaming/providers/:id
func UpdateProviderHandler(c *gin.Context) {
	id := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	proxy(c, http.MethodPut, "/api/streaming-providers/"+id, bytes.NewReader(body))
}

// POST /api/public/streaming/movies/:id/providers
func AddProviderHandler(c *gin.Context) {
	id := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	proxy(c, http.MethodPost, "/api/movies/"+id+"/streaming-providers", bytes.NewReader(body))
}

// DELETE /api/public/streaming/providers/:id
func DeleteProviderHandler(c *gin.Context) {
	id := c.Param("id")
	proxy(c, http.MethodDelete, "/api/streaming-providers/"+id, nil)
}

// GET /api/public/streaming/search?title=...&tmdb_id=...
func SearchMovieHandler(c *gin.Context) {
	title := c.Query("title")
	tmdbId := c.Query("tmdb_id")

	path := "/api/movies/search?"
	if title != "" {
		path += "title=" + title
	}
	if tmdbId != "" {
		path += "&tmdb_id=" + tmdbId
	}
	proxy(c, http.MethodGet, path, nil)
}