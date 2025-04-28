package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const UsersUrl string = V2Endpoint + "/users"

type User struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type NewUser struct {
	DisplayName      string `json:"displayName"`
	Email            string `json:"email"`
	RestoreIfDeleted bool   `json:"restoreIfDeleted"`
}

type UpdateUser struct {
	DisplayName string `json:"displayName"`
}

func (c *Client) GetUser(idOrEmail string) (*User, int, error) {
	var user User
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(UsersUrl+"/%s", idOrEmail), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &user, statusCode, nil
}

func (c *Client) CreateUser(user *NewUser) (*User, error) {
	requestBody, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodPost, UsersUrl, &requestBody)
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

	res, _, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(UsersUrl+"/%s", id), &jsonUser)
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
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf(UsersUrl+"/%s", id), nil)
	return err
}
