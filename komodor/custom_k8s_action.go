package komodor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Fields are in alphabetical order so the JSON written back to state matches what jsonencode() produces.
// resourceNames is left out on purpose: the API does not store it yet, so it must fail the plan
// instead of granting access to every resource of that type.
type CustomK8sActionStatement struct {
	ApiGroups       []string `json:"apiGroups,omitempty"`
	NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
	Resources       []string `json:"resources,omitempty"`
	Verbs           []string `json:"verbs"`
}

// parseRuleset rejects unknown keys so a misspelled or unsupported field fails the plan
// instead of being dropped from the rule sent to the API.
func parseRuleset(ruleset string) ([]CustomK8sActionStatement, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(ruleset)))
	dec.DisallowUnknownFields()
	var statements []CustomK8sActionStatement
	if err := dec.Decode(&statements); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data after the ruleset JSON array")
	}
	for i, s := range statements {
		if err := validateStatement(s); err != nil {
			return nil, fmt.Errorf("rule %d: %w", i, err)
		}
	}
	return statements, nil
}

// Same shape checks as Kubernetes ValidatePolicyRule, so a rule the cluster would reject fails the plan.
func validateStatement(s CustomK8sActionStatement) error {
	if len(s.Verbs) == 0 {
		return errors.New("verbs must not be empty")
	}
	if len(s.NonResourceURLs) > 0 {
		if len(s.ApiGroups) > 0 || len(s.Resources) > 0 {
			return errors.New("a rule cannot set both nonResourceURLs and apiGroups/resources")
		}
		return nil
	}
	if len(s.ApiGroups) == 0 || len(s.Resources) == 0 {
		return errors.New(`a rule needs both apiGroups and resources (use [""] for the core group), or nonResourceURLs`)
	}
	return nil
}

type CustomK8sAction struct {
	Id          string                     `json:"id"`
	Action      string                     `json:"action"`
	Description string                     `json:"description"`
	Ruleset     []CustomK8sActionStatement `json:"k8sRuleset"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
}

type NewCustomK8sAction struct {
	Action      string                     `json:"action"`
	Description string                     `json:"description"`
	Ruleset     []CustomK8sActionStatement `json:"k8sRuleset"`
}

type UpdateCustomK8sAction struct {
	Description string                     `json:"description"`
	Ruleset     []CustomK8sActionStatement `json:"k8sRuleset"`
}

func (c *Client) GetCustomK8sActions() ([]CustomK8sAction, error) {
	res, _, err := c.executeHttpRequest(http.MethodGet, c.GetCustomK8sActionUrl(), nil)
	if err != nil {
		return nil, err
	}

	var customK8sActions []CustomK8sAction

	err = json.Unmarshal(res, &customK8sActions)
	if err != nil {
		return nil, err
	}

	return customK8sActions, nil
}

func (c *Client) GetCustomK8sAction(id string) (*CustomK8sAction, int, error) {
	var customK8sAction CustomK8sAction

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.GetCustomK8sActionUrl(), id), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &customK8sAction)
	if err != nil {
		return nil, statusCode, err
	}

	return &customK8sAction, statusCode, nil
}

func (c *Client) CreateCustomK8sAction(p *NewCustomK8sAction) (*CustomK8sAction, int, error) {
	jsonCustomK8sAction, err := json.Marshal(p)
	if err != nil {
		return nil, 0, err
	}

	res, statusCode, err := c.executeHttpRequest(http.MethodPost, c.GetCustomK8sActionUrl(), &jsonCustomK8sAction)
	if err != nil {
		return nil, statusCode, err
	}

	var customK8sAction CustomK8sAction
	err = json.Unmarshal(res, &customK8sAction)
	if err != nil {
		return nil, statusCode, err
	}

	return &customK8sAction, statusCode, nil
}

func (c *Client) DeleteCustomK8sAction(id string) error {
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.GetCustomK8sActionUrl(), id), nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateCustomK8sAction(id string, p *NewCustomK8sAction) (*CustomK8sAction, error) {
	update := &UpdateCustomK8sAction{
		Description: p.Description,
		Ruleset:     p.Ruleset,
	}
	jsonCustomK8sAction, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf("%s/%s", c.GetCustomK8sActionUrl(), id), &jsonCustomK8sAction)
	if err != nil {
		return nil, err
	}

	var customK8sAction CustomK8sAction
	err = json.Unmarshal(res, &customK8sAction)
	if err != nil {
		return nil, err
	}

	return &customK8sAction, nil
}
