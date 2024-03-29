package komodor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) executeHttpRequest(method string, url string, body *[]byte) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(*body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusNoContent {
		return resBody, res.StatusCode, nil
	} else {
		return resBody, res.StatusCode, fmt.Errorf("%d %s", res.StatusCode, resBody)
	}
}
