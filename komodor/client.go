package komodor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	maxRetries   = 2 // 1 initial attempt + 2 retries = 3 total, matching the original behavior
	retryWaitMin = 5 * time.Second
	retryWaitMax = 5 * time.Second
	userAgent    = "Terraform (terraform-provider-komodor); Go-http-client/1.1"
)

type Client struct {
	retryClient *retryablehttp.Client
	ApiKey      string
	BaseURL     string
}

type ApiKeyResponse struct {
	Valid bool `json:"valid"`
}

// checkRetry is the retry policy for the Komodor client.
// It retries on network-level errors and on HTTP 502 Bad Gateway responses.
func checkRetry(_ context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}
	if resp != nil && resp.StatusCode == http.StatusBadGateway {
		return true, nil
	}
	return false, nil
}

func NewClient(apiKey string, baseURL string) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = maxRetries
	rc.RetryWaitMin = retryWaitMin
	rc.RetryWaitMax = retryWaitMax
	rc.CheckRetry = checkRetry
	rc.Logger = nil // suppress default stderr output; Terraform SDK handles provider logging

	return &Client{
		retryClient: rc,
		ApiKey:      apiKey,
		BaseURL:     baseURL,
	}
}

// GetDefaultEndpoint returns the v1 management API endpoint
func (c *Client) GetDefaultEndpoint() string {
	return c.BaseURL + "/mgmt/v1"
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

// GetPoliciesUrl returns the v1 policies endpoint
func (c *Client) GetPoliciesUrl() string {
	return c.GetDefaultEndpoint() + "/rbac/policies"
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

// GetKnowledgeBaseUrl returns the Klaudia Knowledge Base files endpoint
func (c *Client) GetKnowledgeBaseUrl() string {
	return c.GetV2Endpoint() + "/klaudia/knowledge-base/files"
}

// setCommonHeaders sets the authentication, content-type, and user-agent headers on
// a retryablehttp request, avoiding duplication across request helpers.
func (c *Client) setCommonHeaders(req *retryablehttp.Request, contentType string) {
	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)
}

// doRequest executes a prepared retryablehttp request, reads the response body, and
// returns an error for any non-2xx status code.  Retry logic is handled by the
// retryablehttp.Client configured in NewClient.
func (c *Client) doRequest(req *retryablehttp.Request) ([]byte, int, error) {
	resp, err := c.retryClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		return resBody, resp.StatusCode, nil
	}

	return resBody, resp.StatusCode, fmt.Errorf("received error response: %d %s", resp.StatusCode, resBody)
}

// executeHttpRequest sends a JSON request with automatic retry handling.
func (c *Client) executeHttpRequest(method string, url string, body *[]byte) ([]byte, int, error) {
	var bodyArg interface{}
	if body != nil {
		bodyArg = *body
	}

	req, err := retryablehttp.NewRequest(method, url, bodyArg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	c.setCommonHeaders(req, "application/json")

	return c.doRequest(req)
}

// executeMultipartRequest sends a multipart/form-data request (e.g. for file uploads)
// with automatic retry handling.
func (c *Client) executeMultipartRequest(method, url string, body *bytes.Buffer, contentType string) ([]byte, int, error) {
	req, err := retryablehttp.NewRequest(method, url, body.Bytes())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	c.setCommonHeaders(req, contentType)

	return c.doRequest(req)
}
