package komodor

import (
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
	res, err := c.executeHttpRequest(http.MethodGet, PoliciesUrl, nil)

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

	res, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(PoliciesUrl+"/%s", id), nil)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
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
	res, err := c.executeHttpRequest(http.MethodPost, PoliciesUrl, &jsonPolicy)

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
	_, err = c.executeHttpRequest(http.MethodDelete, PoliciesUrl, &requestBody)
	if err != nil {
		return err
	}
	return nil
}
