package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendSystemInfo(info any) error {
	body, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal system info: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/agents", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send system info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}
