package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const PolicyRoleAttachmentUrl string = DefaultEndpoint + "/rbac/roles/policies"

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
	_, _, err = c.executeHttpRequest(http.MethodPost, PolicyRoleAttachmentUrl, &requestBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRolePoliciesObject(roleId string) ([]RolePolicy, int, error) {
	var rolePolicies []RolePolicy

	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(DefaultEndpoint+"/rbac/roles/%s/policies", roleId), nil)
	if err != nil {
		return nil, statusCode, err
	}

	err = json.Unmarshal([]byte(res), &rolePolicies)
	if err != nil {
		return nil, statusCode, err
	}

	return rolePolicies, statusCode, nil
}

func (c *Client) DetachPolicy(policyId string, roleId string) error {
	rolePolicyObject := RolePolicy{RoleId: roleId, PolicyId: policyId}
	requestBody, err := json.Marshal(rolePolicyObject)
	if err != nil {
		return err
	}
	_, _, err = c.executeHttpRequest(http.MethodDelete, PolicyRoleAttachmentUrl, &requestBody)
	if err != nil {
		return err
	}

	return nil
}
