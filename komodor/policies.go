package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const PoliciesUrl string = DefaultEndpoint + "/rbac/policies"
const PoliciesUrlV2 string = V2Endpoint + "/rbac/policies"

type Resource struct {
	Cluster          string   `json:"cluster"`
	Namespaces       []string `json:"namespaces,omitempty"`
	NamespacePattern string   `json:"namespacePattern,omitempty"`
}

// Pattern defines model for Pattern.
type Pattern struct {
	Exclude string `json:"exclude"`
	Include string `json:"include"`
}

// SelectorType defines model for SelectorType.
type SelectorType string

// Selector defines model for Selector.
type Selector struct {
	Key   string       `json:"key"`
	Type  SelectorType `json:"type"`
	Value string       `json:"value"`
}

// SelectorPattern defines model for SelectorPattern.
type SelectorPattern struct {
	Key   string       `json:"key"`
	Type  SelectorType `json:"type"`
	Value Pattern      `json:"value"`
}

type ResourcesScope struct {
	// Clusters List of cluster names
	Clusters []string `json:"clusters"`

	// ClustersPatterns Patterns for clusters
	ClustersPatterns []Pattern `json:"clustersPatterns"`

	// Namespaces List of namespace names
	Namespaces []string `json:"namespaces"`

	// NamespacesPatterns Patterns for namespaces
	NamespacesPatterns []Pattern `json:"namespacesPatterns"`

	// Selectors Key-value pairs
	Selectors []Selector `json:"selectors"`

	// SelectorsPatterns Key-pattern pairs
	SelectorsPatterns []SelectorPattern `json:"selectorsPatterns"`
}

type Statement struct {
	Actions        []string        `json:"actions"`
	Resources      *[]Resource     `json:"resources,omitempty"`
	ResourcesScope *ResourcesScope `json:"resourcesScope,omitempty"`
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

func (c *Client) GetPolicy(nameOrId string) (*Policy, int, error) {
	var policy Policy

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(PoliciesUrlV2+"/%s", nameOrId), nil)

	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &policy)
	if err != nil {
		return nil, statusCode, err
	}

	return &policy, statusCode, nil
}

// Create Policy
func (c *Client) CreatePolicyV1(p *NewPolicy) (*Policy, error) {
	return c.CreatePolicy(p, PoliciesUrl)
}

func (c *Client) CreatePolicyV2(p *NewPolicy) (*Policy, error) {
	return c.CreatePolicy(p, PoliciesUrlV2)
}

func (c *Client) CreatePolicy(p *NewPolicy, beUrl string) (*Policy, error) {
	jsonPolicy, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, beUrl, &jsonPolicy)

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

func (c *Client) DeletePolicyV1(id string) error {
	return c.DeletePolicy(id, PoliciesUrl)
}

func (c *Client) DeletePolicyV2(id string) error {
	return c.DeletePolicy(id, PoliciesUrlV2)
}

func (c *Client) DeletePolicy(id string, beUrl string) error {
	requestBody, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodDelete, beUrl, &requestBody)
	if err != nil {
		return err
	}
	return nil
}

// Update Policy

func (c *Client) UpdatePolicyV1(id string, p *NewPolicy) (*Policy, error) {
	return c.UpdatePolicy(id, p, PoliciesUrl)
}

func (c *Client) UpdatePolicyV2(id string, p *NewPolicy) (*Policy, error) {
	return c.UpdatePolicy(id, p, PoliciesUrlV2)
}

func (c *Client) UpdatePolicy(id string, p *NewPolicy, beUrl string) (*Policy, error) {
	jsonPolicy, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(beUrl+"/%s", id), &jsonPolicy)
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
