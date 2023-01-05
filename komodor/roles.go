package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const RolesUrl string = DefaultEndpoint + "/rbac/roles"

type Role struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	IsDefault bool   `json:"isDefault"`
}

type NewRole struct {
	Name string `json:"name"`
}

func (c *Client) GetRoles() ([]Role, error) {
	req, err := http.NewRequest(http.MethodGet, RolesUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var roles []Role

	err = json.Unmarshal(res, &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *Client) GetRole(id string) (*Role, error) {
	var role Role

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(RolesUrl+"/%s", id), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (c *Client) CreateRole(r *NewRole) (*Role, error) {
	requestBody, err := json.Marshal(r)

	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, RolesUrl, bytes.NewBuffer(requestBody))

	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var role Role
	err = json.Unmarshal(res, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (c *Client) DeleteRole(id string) error {
	requestBody, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, RolesUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
