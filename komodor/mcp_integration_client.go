package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type MCPIntegration struct {
	ID            string                 `json:"id"`
	AccountID     string                 `json:"accountId"`
	Name          string                 `json:"name"`
	Status        string                 `json:"status"`
	Configuration map[string]interface{} `json:"configuration"`
	Tools         []interface{}          `json:"tools"`
	UseCases      []string               `json:"useCases"`
	Clusters      []string               `json:"clusters"`
	SkillID       *string                `json:"skillId"`
}

type MCPIntegrationRequest struct {
	Name          string                 `json:"name"`
	Configuration map[string]interface{} `json:"configuration"`
	Tools         []interface{}          `json:"tools,omitempty"`
	UseCases      []string               `json:"useCases"`
	Clusters      []string               `json:"clusters"`
	SkillID       *string                `json:"skillId,omitempty"`
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
	_, _, err := c.executeHttpRequest(http.MethodDelete, url, nil)
	return err
}
