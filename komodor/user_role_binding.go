package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserRole struct {
	UserId string `json:"userId"`
	RoleId string `json:"roleId"`
}

type UserRoleCreateRequest struct {
	UserId string `json:"userId"`
	RoleId string `json:"roleId"`
}

type UserRoleDeleteRequest struct {
	UserId string `json:"userId"`
	RoleId string `json:"roleId"`
}

// AttachUserToRole attaches a user to a role
func (c *Client) AttachUserToRole(userId string, roleId string) error {
	userRoleObject := UserRoleCreateRequest{
		UserId: userId,
		RoleId: roleId,
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

	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, statusCode, err
	}

	userRoles := make([]UserRole, 0, len(user.Roles))
	for _, role := range user.Roles {
		userRoles = append(userRoles, UserRole{
			UserId: userId,
			RoleId: role.Id,
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

// UpdateUserRole updates a user role assignment
func (c *Client) UpdateUserRole(userId string, roleId string) error {
	userRoleObject := UserRoleCreateRequest{
		UserId: userId,
		RoleId: roleId,
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
