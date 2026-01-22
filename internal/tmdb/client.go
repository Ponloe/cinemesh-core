package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	config     *Config
	httpClient *http.Client
}

func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.config.APIKey)

	fullURL := fmt.Sprintf("%s%s?%s", BaseURL, endpoint, params.Encode())

	// Log the request (hide API key for security)
	safeURL := fmt.Sprintf("%s%s?api_key=***&%s", BaseURL, endpoint, params.Encode())
	log.Printf("TMDb API request: %s", safeURL)

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("TMDb API error: status %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("tmdb api error: status %d", resp.StatusCode)
	}

	log.Printf("TMDb API response: %d bytes", len(body))
	return body, nil
}

func (c *Client) SearchMovies(query string) (*MovieSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")

	body, err := c.get("/search/movie", params)
	if err != nil {
		return nil, err
	}

	var result MovieSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		log.Printf("Response body: %s", string(body))
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &result, nil
}

func (c *Client) GetMovieDetails(tmdbID int) (*MovieDetails, error) {
	endpoint := fmt.Sprintf("/movie/%d", tmdbID)
	params := url.Values{}
	params.Set("append_to_response", "release_dates")

	body, err := c.get(endpoint, params)
	if err != nil {
		return nil, err
	}

	var details MovieDetails
	if err := json.Unmarshal(body, &details); err != nil {
		log.Printf("Failed to unmarshal movie details: %v", err)
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &details, nil
}

func (c *Client) GetGenres() ([]Genre, error) {
	body, err := c.get("/genre/movie/list", nil)
	if err != nil {
		return nil, err
	}

	var result GenreListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return result.Genres, nil
}

func (c *Client) FetchMovieCredits(tmdbID int) (*TMDbCredits, error) {
	endpoint := fmt.Sprintf("/movie/%d/credits", tmdbID)

	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var credits TMDbCredits
	if err := json.Unmarshal(body, &credits); err != nil {
		return nil, fmt.Errorf("unmarshal credits: %w", err)
	}

	return &credits, nil
}

func (c *Client) FetchPersonDetails(tmdbID int) (*TMDbPersonDetail, error) {
	endpoint := fmt.Sprintf("/person/%d", tmdbID)

	body, err := c.get(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var person TMDbPersonDetail
	if err := json.Unmarshal(body, &person); err != nil {
		return nil, fmt.Errorf("unmarshal person: %w", err)
	}

	return &person, nil
}

func (c *Client) GetFullImageURL(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", path)
}
