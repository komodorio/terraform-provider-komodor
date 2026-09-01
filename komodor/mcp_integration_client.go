package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ─────────────────────────────────────────────────────────────────────────────
// Wire types — match the ai-investigator MCP API exactly. The plugin sends and
// receives the same shape, so no flattening / un-flattening is needed.
// ─────────────────────────────────────────────────────────────────────────────

type MCPServer struct {
	URL       string            `json:"url"`
	Transport string            `json:"transport,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type AgentTunnel struct {
	ProviderCluster string `json:"provider_cluster"`
}

type Connectivity struct {
	Mode        string       `json:"mode"`
	AgentTunnel *AgentTunnel `json:"agent_tunnel,omitempty"`
}

type StaticTokenAuth struct {
	Value      string `json:"value"`
	HeaderName string `json:"header_name,omitempty"`
}

type SubjectToken struct {
	Type     string `json:"type"`
	Value    string `json:"value,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type ActorToken struct {
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type TokenExchangeAuth struct {
	TokenURL           string            `json:"token_url"`
	GrantType          string            `json:"grant_type,omitempty"`
	SubjectToken       SubjectToken      `json:"subject_token"`
	Audience           string            `json:"audience,omitempty"`
	Scope              string            `json:"scope,omitempty"`
	RequestedTokenType string            `json:"requested_token_type,omitempty"`
	ActorToken         *ActorToken       `json:"actor_token,omitempty"`
	ClientID           string            `json:"client_id,omitempty"`
	ClientSecret       string            `json:"client_secret,omitempty"`
	ExtraParams        map[string]string `json:"extra_params,omitempty"`
}

type OAuth2Auth struct {
	TokenURL     string `json:"token_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
	Audience     string `json:"audience,omitempty"`
	GrantType    string `json:"grant_type,omitempty"`
}

type ResponseConfig struct {
	TokenField     string `json:"token_field,omitempty"`
	TokenTypeField string `json:"token_type_field,omitempty"`
	ExpiresInField string `json:"expires_in_field,omitempty"`
}

type AuthConfig struct {
	Method                  string             `json:"method"`
	StaticToken             *StaticTokenAuth   `json:"static_token,omitempty"`
	TokenExchange           *TokenExchangeAuth `json:"token_exchange,omitempty"`
	OAuth2ClientCredentials *OAuth2Auth        `json:"oauth2_client_credentials,omitempty"`
	Response                *ResponseConfig    `json:"response,omitempty"`
}

type MCPConfiguration struct {
	MCPServer    MCPServer    `json:"mcp_server"`
	Connectivity Connectivity `json:"connectivity"`
	Auth         *AuthConfig  `json:"auth,omitempty"`
	IncludeTools []string     `json:"include_tools,omitempty"`
	ExcludeTools []string     `json:"exclude_tools,omitempty"`
	Mode         string       `json:"mode,omitempty"`
}

type MCPIntegration struct {
	ID            string           `json:"id"`
	AccountID     string           `json:"accountId"`
	Name          string           `json:"name"`
	Status        string           `json:"status"`
	Configuration MCPConfiguration `json:"configuration"`
	Tools         []interface{}    `json:"tools"`
	Clusters      []string         `json:"clusters"`
	SkillID       *string          `json:"skillId"`
}

type MCPIntegrationRequest struct {
	Name          string           `json:"name"`
	Configuration MCPConfiguration `json:"configuration"`
	Tools         []interface{}    `json:"tools,omitempty"`
	Clusters      []string         `json:"clusters"`
	SkillID       *string          `json:"skillId,omitempty"`
}

func (c *Client) CreateMCPIntegration(req *MCPIntegrationRequest) (*MCPIntegration, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, c.GetKlaudiaMCPIntegrationsUrl(), &body)
	if err != nil {
		return nil, err
	}
	var integration MCPIntegration
	if err := json.Unmarshal(res, &integration); err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) GetMCPIntegration(id string) (*MCPIntegration, int, error) {
	url := fmt.Sprintf("%s/%s", c.GetKlaudiaMCPIntegrationsUrl(), id)
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, statusCode, err
	}
	var integration MCPIntegration
	if err := json.Unmarshal(res, &integration); err != nil {
		return nil, statusCode, err
	}
	return &integration, statusCode, nil
}

func (c *Client) UpdateMCPIntegration(id string, req *MCPIntegrationRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s", c.GetKlaudiaMCPIntegrationsUrl(), id)
	_, _, err = c.executeHttpRequest(http.MethodPut, url, &body)
	return err
}

func (c *Client) DeleteMCPIntegration(id string) error {
	url := fmt.Sprintf("%s/%s", c.GetKlaudiaMCPIntegrationsUrl(), id)
	_, statusCode, err := c.executeHttpRequest(http.MethodDelete, url, nil)
	if err != nil && statusCode == http.StatusNotFound {
		return nil
	}
	return err
}
