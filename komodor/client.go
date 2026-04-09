package komodor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	HttpClient *http.Client
	ApiKey     string
	BaseURL    string
}

type ApiKeyResponse struct {
	Valid bool `json:"valid"`
}

func NewClient(apiKey string, baseURL string) *Client {
	return &Client{
		HttpClient: &http.Client{Timeout: 30 * time.Second},
		ApiKey:     apiKey,
		BaseURL:    baseURL,
	}
}

// GetV2Endpoint returns the v2 API endpoint
func (c *Client) GetV2Endpoint() string {
	return c.BaseURL + "/api/v2"
}

// GetCustomK8sActionUrl returns the custom K8s actions endpoint
func (c *Client) GetCustomK8sActionUrl() string {
	return c.GetV2Endpoint() + "/rbac/actions"
}

// GetUsersUrl returns the users endpoint
func (c *Client) GetUsersUrl() string {
	return c.GetV2Endpoint() + "/users"
}

// GetRolesUrl returns the roles endpoint
func (c *Client) GetRolesUrl() string {
	return c.GetV2Endpoint() + "/rbac/roles"
}

// GetPoliciesUrlV2 returns the v2 policies endpoint
func (c *Client) GetPoliciesUrlV2() string {
	return c.GetV2Endpoint() + "/rbac/policies"
}

// GetMonitorsUrl returns the monitors endpoint
func (c *Client) GetMonitorsUrl() string {
	return c.GetV2Endpoint() + "/realtime-monitors/config"
}

// GetIntegrationsUrl returns the Kubernetes integrations endpoint
func (c *Client) GetIntegrationsUrl() string {
	return c.GetV2Endpoint() + "/integrations/kubernetes"
}

// GetWorkspacesUrl returns the workspaces endpoint
func (c *Client) GetWorkspacesUrl() string {
	return c.GetV2Endpoint() + "/workspaces"
}

// GetPolicyRoleAttachmentUrl returns the policy role attachment endpoint
func (c *Client) GetPolicyRoleAttachmentUrl() string {
	return c.GetV2Endpoint() + "/rbac/roles/policies"
}

// GetUserRoleBindingUrl returns the user role binding endpoint
func (c *Client) GetUserRoleBindingUrl() string {
	return c.GetV2Endpoint() + "/rbac/users/roles"
}

// GetKlaudiaFilesUrl returns the Klaudia files endpoint for the given file type.
// Valid types are "knowledge-base" and "blueprint".
func (c *Client) GetKlaudiaFilesUrl(fileType string) string {
	return c.GetV2Endpoint() + "/klaudia/files/" + fileType
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
	req.Header.Set("User-Agent", "Terraform (terraform-provider-komodor); Go-http-client/1.1")

	return req, nil
}

// executeWithRetry executes the HTTP request with retry logic for certain status codes
func (c *Client) executeWithRetry(req *http.Request, maxRetries int, retryDelay time.Duration) ([]byte, int, error) {
	// Capture body bytes upfront so each retry attempt gets a fresh reader.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	var res *http.Response
	var resBody []byte
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		res, err = c.HttpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("request failed: %w", err)
		}

		resBody, err = io.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil {
			return nil, res.StatusCode, fmt.Errorf("failed to read response body: %w", err)
		}

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusNoContent {
			return resBody, res.StatusCode, nil
		}

		// Retry only for 502 status code
		if res.StatusCode == http.StatusBadGateway {
			log.Printf("[DEBUG] Retry attempt %d/%d for status %d", attempt+1, maxRetries, res.StatusCode)
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

// executeMultipartRequest sends a multipart/form-data request (e.g. for file uploads)
func (c *Client) executeMultipartRequest(method, url string, body *bytes.Buffer, contentType string) ([]byte, int, error) {
	maxRetries := 3
	retryDelay := 5 * time.Second

	// Capture bytes upfront so each retry attempt gets a fresh reader.
	bodyBytes := body.Bytes()

	var res *http.Response
	var resBody []byte

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("x-api-key", c.ApiKey)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("User-Agent", "Terraform (terraform-provider-komodor); Go-http-client/1.1")

		res, err = c.HttpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("request failed: %w", err)
		}

		resBody, err = io.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil {
			return nil, res.StatusCode, fmt.Errorf("failed to read response body: %w", err)
		}

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusNoContent {
			return resBody, res.StatusCode, nil
		}

		if res.StatusCode == http.StatusBadGateway {
			fmt.Printf("Retry attempt %d/%d for status %d\n", attempt+1, maxRetries, res.StatusCode)
			time.Sleep(retryDelay)
			continue
		}

		return resBody, res.StatusCode, fmt.Errorf("received error response: %d %s", res.StatusCode, resBody)
	}

	return resBody, res.StatusCode, fmt.Errorf("request failed after %d retries with status: %d", maxRetries, res.StatusCode)
}
