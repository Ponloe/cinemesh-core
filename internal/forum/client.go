package forum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Base Client
type ForumClient struct {
	BaseURL string
	Token   string
}

func NewForumClient(baseURL, token string) *ForumClient {
	return &ForumClient{
		BaseURL: baseURL,
		Token:   token,
	}
}

// API Methods
func (c *ForumClient) GetTopics() ([]Topic, error) {
	resp, err := http.Get(c.BaseURL + "/api/forum/topics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var topics []Topic
	data, _ := json.Marshal(apiResp.Data)
	json.Unmarshal(data, &topics)
	return topics, nil
}

func (c *ForumClient) GetThreads(topicSlug string, page, limit int) ([]Thread, *Pagination, error) {
	url := fmt.Sprintf("%s/api/forum/topics/%s/threads?page=%d&limit=%d",
		c.BaseURL, topicSlug, page, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, nil, err
	}

	var threads []Thread
	data, _ := json.Marshal(apiResp.Data)
	json.Unmarshal(data, &threads)

	return threads, apiResp.Pagination, nil
}

func (c *ForumClient) CreateThread(topicSlug string, title, content string) (*Thread, error) {
	payload := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/forum/topics/%s/threads", c.BaseURL, topicSlug)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var thread Thread
	data, _ := json.Marshal(apiResp.Data)
	json.Unmarshal(data, &thread)

	return &thread, nil
}

func (c *ForumClient) GetReplies(threadSlug string, page, limit int) ([]Reply, error) {
	url := fmt.Sprintf("%s/api/forum/threads/%s/replies?page=%d&limit=%d",
		c.BaseURL, threadSlug, page, limit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var replies []Reply
	data, _ := json.Marshal(apiResp.Data)
	json.Unmarshal(data, &replies)

	return replies, nil
}

func (c *ForumClient) DeleteReply(replyID string) error {
	url := fmt.Sprintf("%s/api/forum/replies/%s", c.BaseURL, replyID)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetThreadBySlug fetches a single thread by its slug
func (c *ForumClient) GetThreadBySlug(threadSlug string) (*Thread, error) {
	url := fmt.Sprintf("%s/api/forum/threads/%s", c.BaseURL, threadSlug)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	var thread Thread
	data, _ := json.Marshal(apiResp.Data)
	json.Unmarshal(data, &thread)

	return &thread, nil
}

// UpdateThread updates thread properties (pin, lock, etc.)
func (c *ForumClient) UpdateThread(threadID string, updates map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/forum/threads/%s", c.BaseURL, threadID)

	jsonData, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}
