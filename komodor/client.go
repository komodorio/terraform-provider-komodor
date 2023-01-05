package komodor

import (
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	HttpClient *http.Client
	ApiKey     string
	Base       string
}

type ApiKeyResponse struct {
	Valid bool `json:"valid"`
}

func NewClient(apiKey string, base string) *Client {
	return &Client{
		HttpClient: http.DefaultClient,
		ApiKey:     apiKey,
		Base:       base,
	}
}

// func (c *Client) newRequest(path string) (*http.Request, error) {
// 	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.Base, path), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return req, nil
// }

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusNoContent {
		return body, err
	} else {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}
}
