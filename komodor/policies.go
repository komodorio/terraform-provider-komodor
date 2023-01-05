package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const PoliciesUrl string = DefaultEndpoint + "/rbac/policies"

type Resource struct {
	Cluster    string   `json:"cluster"`
	Namespaces []string `json:"namespaces"`
}

type Statement struct {
	Actions   []string   `json:"actions"`
	Resources []Resource `json:"resources"`
}

type Policy struct {
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	Statements []Statement `json:"statements"`
	CreatedAt  string      `json:"createdAt"`
	UpdatedAt  string      `json:"updatedAt"`
}

type NewPolicy struct {
	Name       string      `json:"name"`
	Statements []Statement `json:"statements"`
}

func (c *Client) GetPolicies() ([]Policy, error) {
	req, err := http.NewRequest(http.MethodGet, PoliciesUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var policies []Policy

	err = json.Unmarshal(res, &policies)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

func (c *Client) GetPolicy(id string) (*Policy, error) {
	var policy Policy

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(PoliciesUrl+"/%s", id), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func (c *Client) CreatePolicy(p *NewPolicy) (*Policy, error) {
	j, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, PoliciesUrl, bytes.NewBuffer(j))

	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var policy Policy
	err = json.Unmarshal(res, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func (c *Client) DeletePolicy(id string) error {
	requestBody, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, PoliciesUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
