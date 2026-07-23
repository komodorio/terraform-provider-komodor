package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const workloadOverridesPageLimit = 100

func (c *Client) GetOverrides() ([]WorkloadOverride, error) {
	var all []WorkloadOverride
	offset := 0
	for {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", c.GetCostRightSizingWorkloadOverridesUrl(), workloadOverridesPageLimit, offset)
		res, _, err := c.executeHttpRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		var resp WorkloadOverrideListResponse
		if err = json.Unmarshal(res, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.PolicyScopeExceptionWorkloads...)
		offset += workloadOverridesPageLimit
		if len(resp.PolicyScopeExceptionWorkloads) == 0 || offset >= resp.Meta.Total {
			break
		}
	}
	return all, nil
}

func (c *Client) GetOverrideByID(id string) (*WorkloadOverride, bool, error) {
	all, err := c.GetOverrides()
	if err != nil {
		return nil, false, err
	}
	for i := range all {
		if all[i].Id == id {
			return &all[i], true, nil
		}
	}
	return nil, false, nil
}

func (c *Client) CreateOverride(body WorkloadOverrideBase) (*WorkloadOverrideItemResponse, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, c.GetCostRightSizingWorkloadOverridesUrl(), &requestBody)
	if err != nil {
		return nil, err
	}
	var resp WorkloadOverrideItemResponse
	if err = json.Unmarshal(res, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateOverride(id string, body WorkloadOverrideBase) (*WorkloadOverrideItemResponse, error) {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s", c.GetCostRightSizingWorkloadOverridesUrl(), id)
	res, _, err := c.executeHttpRequest(http.MethodPut, url, &requestBody)
	if err != nil {
		return nil, err
	}
	var resp WorkloadOverrideItemResponse
	if err = json.Unmarshal(res, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteOverride(id string) error {
	url := fmt.Sprintf("%s/%s", c.GetCostRightSizingWorkloadOverridesUrl(), id)
	_, status, err := c.executeHttpRequest(http.MethodDelete, url, nil)
	if err != nil {
		if status == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}
