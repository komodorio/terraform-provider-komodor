package komodor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/samber/lo"
)

const (
	UsersV2Url string = V2Endpoint + "/users"
	UsersV1Url string = DefaultEndpoint + "/rbac/users"
)

type User struct {
	Id          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	Email       string  `json:"email"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	DeletedAt   *string `json:"deletedAt,omitempty"`
}

type NewUser struct {
	DisplayName      string `json:"displayName"`
	Email            string `json:"email"`
	RestoreIfDeleted bool   `json:"restoreIfDeleted"`
}

type UpdateUser struct {
	DisplayName string `json:"displayName"`
}

type getUsersParams struct {
	Email     *string `json:"email,omitempty"`
	IsDeleted *bool   `json:"isDeleted,omitempty"`
}

func (c *Client) GetUsers() ([]User, error) {
	return c.getUsers(&getUsersParams{IsDeleted: lo.ToPtr(false)})
}

func (c *Client) GetUserByEmail(email string) (*User, error) {
	users, err := c.getUsers(&getUsersParams{Email: &email, IsDeleted: lo.ToPtr(false)})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
}

func (c *Client) GetUser(id string) (*User, int, error) {
	var user User
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(UsersV1Url+"/%s", id), nil) // TODO: replace with v2 (after implementation)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if user.DeletedAt != nil {
		return nil, http.StatusNotFound, errors.New("user not found")
	}

	return &user, statusCode, nil
}

func (c *Client) CreateUser(user *NewUser) (*User, error) {
	requestBody, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPost, UsersV2Url, &requestBody)
	if err != nil {
		return nil, err
	}

	var newUser User
	err = json.Unmarshal(res, &newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (c *Client) UpdateUser(id string, p *UpdateUser) (*User, error) {
	jsonUser, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(UsersV2Url+"/%s", id), &jsonUser)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) DeleteUser(id string) error {
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf(UsersV2Url+"/%s", id), nil)
	return err
}

func (c *Client) getUsers(req *getUsersParams) ([]User, error) {
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodGet, UsersV2Url, &requestBody)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(res, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
