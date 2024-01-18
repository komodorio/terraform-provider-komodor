package komodor

import (
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
	res, _, err := c.executeHttpRequest(http.MethodGet, RolesUrl, nil)

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

func (c *Client) GetRoleByName(name string) (*Role, error) {
	allRoles, err := c.GetRoles()
	if err != nil {
		return nil, err
	}
	var targetRole *Role
	for _, role := range allRoles {
		if role.Name == name {
			targetRole = &role
			break
		}
	}

	return targetRole, nil
}

func (c *Client) GetRole(id string) (*Role, int, error) {
	var role Role

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(RolesUrl+"/%s", id), nil)

	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &role)
	if err != nil {
		return nil, statusCode, err
	}

	return &role, statusCode, nil
}

func (c *Client) CreateRole(role *NewRole) (*Role, error) {
	requestBody, err := json.Marshal(role)

	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, RolesUrl, &requestBody)

	if err != nil {
		return nil, err
	}

	var newRole Role
	err = json.Unmarshal(res, &newRole)
	if err != nil {
		return nil, err
	}

	return &newRole, nil
}

func (c *Client) DeleteRole(id string) error {
	requestBody, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		return err
	}

	_, _, err = c.executeHttpRequest(http.MethodDelete, RolesUrl, &requestBody)
	if err != nil {
		return err
	}

	return nil
}
