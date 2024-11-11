package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const PoliciesUrl string = DefaultEndpoint + "/rbac/policies"

type Resource struct {
	Cluster          string   `json:"cluster"`
	Namespaces       []string `json:"namespaces,omitempty"`
	NamespacePattern string   `json:"namespacePattern,omitempty"`
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
	Type       string      `json:"type,omitempty"`
	Tags       interface{} `json:"tags,omitempty"`
}

type NewPolicy struct {
	Name       string      `json:"name"`
	Type       string      `json:"type,omitempty"`
	Statements []Statement `json:"statements"`
	Tags       interface{} `json:"tags,omitempty"`
}

func (c *Client) GetPolicies() ([]Policy, error) {
	res, _, err := c.executeHttpRequest(http.MethodGet, PoliciesUrl, nil)

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

func (c *Client) GetPolicy(id string) (*Policy, int, error) {
	var policy Policy

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(PoliciesUrl+"/%s", id), nil)

	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &policy)
	if err != nil {
		return nil, statusCode, err
	}

	return &policy, statusCode, nil
}

func (c *Client) GetPolicyByName(name string) (*Policy, error) {
	allPolicies, err := c.GetPolicies()
	if err != nil {
		return nil, err
	}
	var targetPolicy *Policy
	for _, policy := range allPolicies {
		if policy.Name == name {
			targetPolicy = &policy
			break
		}
	}

	return targetPolicy, nil
}

func (c *Client) CreatePolicy(p *NewPolicy) (*Policy, error) {
	jsonPolicy, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, PoliciesUrl, &jsonPolicy)

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
	_, _, err = c.executeHttpRequest(http.MethodDelete, PoliciesUrl, &requestBody)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdatePolicy(id string, p *NewPolicy) (*Policy, error) {
	jsonPolicy, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(PoliciesUrl+"/%s", id), &jsonPolicy)
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
