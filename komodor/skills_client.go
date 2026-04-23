package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Skill struct {
	ID           string   `json:"id"`
	AccountID    string   `json:"accountId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Instructions string   `json:"instructions"`
	UseCases     []string `json:"useCases"`
	Clusters     []string `json:"clusters"`
	IsEnabled    bool     `json:"isEnabled"`
}

type CreateSkillRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Instructions string   `json:"instructions"`
	UseCases     []string `json:"useCases"`
	Clusters     []string `json:"clusters"`
}

type UpdateSkillRequest struct {
	Name         *string  `json:"name,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Instructions *string  `json:"instructions,omitempty"`
	UseCases     []string `json:"useCases,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	IsEnabled    *bool    `json:"isEnabled,omitempty"`
}

func (c *Client) CreateSkill(req *CreateSkillRequest) (*Skill, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, c.GetKlaudiaSkillsUrl(), &body)
	if err != nil {
		return nil, err
	}
	var skill Skill
	if err := json.Unmarshal(res, &skill); err != nil {
		return nil, err
	}
	return &skill, nil
}

func (c *Client) GetSkill(id string) (*Skill, int, error) {
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.GetKlaudiaSkillsUrl(), id), nil)
	if err != nil {
		return nil, statusCode, err
	}
	var skill Skill
	if err := json.Unmarshal(res, &skill); err != nil {
		return nil, statusCode, err
	}
	return &skill, statusCode, nil
}

func (c *Client) UpdateSkill(id string, req *UpdateSkillRequest) (*Skill, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf("%s/%s", c.GetKlaudiaSkillsUrl(), id), &body)
	if err != nil {
		return nil, err
	}
	var skill Skill
	if err := json.Unmarshal(res, &skill); err != nil {
		return nil, err
	}
	return &skill, nil
}

func (c *Client) DeleteSkill(id string) error {
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.GetKlaudiaSkillsUrl(), id), nil)
	return err
}
