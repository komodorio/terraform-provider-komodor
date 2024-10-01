package komodor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HttpClient *http.Client
	ApiKey     string
}

type ApiKeyResponse struct {
	Valid bool `json:"valid"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		HttpClient: http.DefaultClient,
		ApiKey:     apiKey,
	}
}

// prepareRequest creates a new HTTP request with the necessary headers
func (c *Client) prepareRequest(method, url string, body *[]byte) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(*body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// executeWithRetry executes the HTTP request with retry logic for certain status codes
func (c *Client) executeWithRetry(req *http.Request, maxRetries int, retryDelay time.Duration) ([]byte, int, error) {
	var res *http.Response
	var resBody []byte
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err = c.HttpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("request failed: %w", err)
		}

		resBody, err = io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, res.StatusCode, fmt.Errorf("failed to read response body: %w", err)
		}

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusNoContent {
			return resBody, res.StatusCode, nil
		}

		// Retry only for 502 status code
		if res.StatusCode == http.StatusBadGateway {
			fmt.Printf("Retry attempt %d/%d for status %d\n", attempt+1, maxRetries, res.StatusCode)
			time.Sleep(retryDelay)
			continue
		}

		// Return on non-retryable status codes
		return resBody, res.StatusCode, fmt.Errorf("received error response: %d %s", res.StatusCode, resBody)
	}

	// If retries exhausted, return last response
	return resBody, res.StatusCode, fmt.Errorf("request failed after %d retries with status: %d", maxRetries, res.StatusCode)
}

func (c *Client) executeHttpRequest(method string, url string, body *[]byte) ([]byte, int, error) {
	maxRetries := 3
	retryDelay := 5 * time.Second

	req, err := c.prepareRequest(method, url, body)
	if err != nil {
		return nil, 0, err
	}

	return c.executeWithRetry(req, maxRetries, retryDelay)
}
