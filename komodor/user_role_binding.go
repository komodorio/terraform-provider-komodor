package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserRole struct {
	UserId     string `json:"userId"`
	RoleId     string `json:"roleId"`
	Expiration string `json:"expiration,omitempty"`
}

type UserRoleCreateRequest struct {
	UserId     string `json:"userId"`
	RoleId     string `json:"roleId"`
	Expiration string `json:"expiration,omitempty"`
}

type UserRoleDeleteRequest struct {
	UserId string `json:"userId"`
	RoleId string `json:"roleId"`
}

// AttachUserToRole attaches a user to a role with optional expiration
func (c *Client) AttachUserToRole(userId string, roleId string, expiration string) error {
	userRoleObject := UserRoleCreateRequest{
		UserId:     userId,
		RoleId:     roleId,
		Expiration: expiration,
	}
	requestBody, err := json.Marshal(userRoleObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodPost, c.GetUserRoleBindingUrl(), &requestBody)
	if err != nil {
		return err
	}

	return nil
}

// GetUserRoles retrieves the roles for a user
func (c *Client) GetUserRoles(userId string) ([]UserRole, int, error) {
	var user User

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.GetUsersUrl(), userId), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal([]byte(res), &user)
	if err != nil {
		return nil, statusCode, err
	}

	// Convert UserRoleResponse to UserRole
	userRoles := make([]UserRole, 0, len(user.Roles))
	for _, role := range user.Roles {
		userRoles = append(userRoles, UserRole{
			UserId:     userId,
			RoleId:     role.Id,
			Expiration: role.Expiration,
		})
	}

	return userRoles, statusCode, nil
}

// DetachUserFromRole detaches a user from a role
func (c *Client) DetachUserFromRole(userId string, roleId string) error {
	userRoleObject := UserRoleDeleteRequest{
		UserId: userId,
		RoleId: roleId,
	}
	requestBody, err := json.Marshal(userRoleObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodDelete, c.GetUserRoleBindingUrl(), &requestBody)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserRole updates a user role assignment (e.g., expiration)
func (c *Client) UpdateUserRole(userId string, roleId string, expiration string) error {
	userRoleObject := UserRoleCreateRequest{
		UserId:     userId,
		RoleId:     roleId,
		Expiration: expiration,
	}
	requestBody, err := json.Marshal(userRoleObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodPut, c.GetUserRoleBindingUrl(), &requestBody)
	if err != nil {
		return err
	}

	return nil
}
