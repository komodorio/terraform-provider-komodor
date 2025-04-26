package komodor

import (
	"encoding/json"
	"fmt"
)

type Workspace struct {
	Id                 string           `json:"id"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	Scopes             []ResourcesScope `json:"scopes"`
	AuthorEmail        string           `json:"AuthorEmail"`
	LastUpdatedByEmail string           `json:"LastUpdatedByEmail"`
	CreatedAt          string           `json:"createdAt"`
	LastUpdated        string           `json:"lastUpdated"`
}

type NewWorkspace struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Scopes      []ResourcesScope `json:"scopes"`
}

func (c *Client) CreateWorkspace(workspace *NewWorkspace) (*Workspace, error) {
	body, err := json.Marshal(workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workspace: %w", err)
	}

	resBody, statusCode, err := c.executeHttpRequest("POST", "/api/v2/workspaces", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	if statusCode != 201 {
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}

	var response Workspace
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (c *Client) GetWorkspace(id string) (*Workspace, int, error) {
	resBody, statusCode, err := c.executeHttpRequest("GET", fmt.Sprintf("/api/v2/workspaces/%s", id), nil)
	if err != nil {
		return nil, statusCode, fmt.Errorf("failed to get workspace: %w", err)
	}

	if statusCode == 404 {
		return nil, statusCode, nil
	}

	var response Workspace
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, statusCode, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, statusCode, nil
}

func (c *Client) UpdateWorkspace(id string, workspace *NewWorkspace) (*Workspace, error) {
	body, err := json.Marshal(workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workspace: %w", err)
	}

	resBody, statusCode, err := c.executeHttpRequest("PUT", fmt.Sprintf("/api/v2/workspaces/%s", id), &body)
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}

	var response Workspace
	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (c *Client) DeleteWorkspace(id string) error {
	_, statusCode, err := c.executeHttpRequest("DELETE", fmt.Sprintf("/api/v2/workspaces/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if statusCode != 204 {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	return nil
}
