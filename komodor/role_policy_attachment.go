package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const PolicyRoleAttachmentUrl string = DefaultEndpoint + "/rbac/roles/policies"

type RolePolicy struct {
	RoleId   string `json:"roleId"`
	PolicyId string `json:"policyId"`
}

type RolePolicies struct {
	RolePolicy []RolePolicy
}

func (c *Client) AttachPolicy(policyId string, roleId string) error {
	rolePolicyObject := RolePolicy{RoleId: roleId, PolicyId: policyId}
	requestBody, err := json.Marshal(rolePolicyObject)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, PolicyRoleAttachmentUrl, bytes.NewBuffer(requestBody))

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetRolePoliciesObject(roleId string) ([]RolePolicy, error) {
	var rolePolicyObject RolePolicies

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(DefaultEndpoint+"/rbac/roles/%s/policies", roleId), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &rolePolicyObject)
	if err != nil {
		return nil, err
	}

	return rolePolicyObject.RolePolicy, nil
}

func (c *Client) DetachPolicy(policyId string, roleId string) error {
	rolePolicyObject := RolePolicy{RoleId: roleId, PolicyId: policyId}
	requestBody, err := json.Marshal(rolePolicyObject)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, PolicyRoleAttachmentUrl, bytes.NewBuffer(requestBody))

	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
