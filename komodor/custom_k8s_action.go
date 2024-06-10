package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const CustomK8sActionUrl string = DefaultEndpoint + "/rbac/actions"

type CustomK8sActionStatement struct {
	ApiGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
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

func (c *Client) GetCustomK8sActions() ([]CustomK8sAction, error) {
	res, _, err := c.executeHttpRequest(http.MethodGet, CustomK8sActionUrl, nil)
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

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(CustomK8sActionUrl+"/%s", id), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &customK8sAction)
	if err != nil {
		return nil, statusCode, err
	}

	return &customK8sAction, statusCode, nil
}

func (c *Client) CreateCustomK8sAction(p *NewCustomK8sAction) (*CustomK8sAction, error) {
	jsonCustomK8sAction, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPost, CustomK8sActionUrl, &jsonCustomK8sAction)
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

func (c *Client) DeleteCustomK8sAction(id string) error {
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf(CustomK8sActionUrl+"/%s", id), nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateCustomK8sAction(id string, p *NewCustomK8sAction) (*CustomK8sAction, error) {
	jsonCustomK8sAction, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(CustomK8sActionUrl+"/%s", id), &jsonCustomK8sAction)
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
