package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)


type RolePolicy struct {
	RoleId   string `json:"roleId"`
	PolicyId string `json:"policyId"`
}

func (c *Client) AttachPolicy(policyId string, roleId string) error {
	rolePolicyObject := RolePolicy{RoleId: roleId, PolicyId: policyId}
	requestBody, err := json.Marshal(rolePolicyObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodPost, c.GetV2Endpoint()+"/rbac/roles/policies", &requestBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRolePoliciesObject(roleId string) ([]PolicyRole, int, error) {
	var role Role

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(c.GetV2Endpoint()+"/rbac/roles/%s", roleId), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal([]byte(res), &role)
	if err != nil {
		return nil, statusCode, err
	}

	return role.Policies, statusCode, nil
}

func (c *Client) DetachPolicy(policyId string, roleId string) error {
	rolePolicyObject := RolePolicy{RoleId: roleId, PolicyId: policyId}
	requestBody, err := json.Marshal(rolePolicyObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodDelete, c.GetV2Endpoint()+"/rbac/roles/policies", &requestBody)
	if err != nil {
		return err
	}

	return nil
}
