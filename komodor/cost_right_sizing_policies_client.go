package komodor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// rightSizingPoliciesAPI is the behavior surface used by the resource and
// data-source layer. Implementations decode JSON responses, build query
// strings, and combine multiple HTTP calls (e.g. GetByName).
type rightSizingPoliciesAPI interface {
	GetAll(ctx context.Context) ([]GetAllRightSizingPoliciesRow, error)
	GetByID(ctx context.Context, id string) (*GetMultiScopePolicyResponse, int, error)
	GetByName(ctx context.Context, name string) (*GetMultiScopePolicyResponse, int, error)
	Create(ctx context.Context, body RightSizingMultiScopePolicy) (*GetMultiScopePolicyResponse, error)
	Update(ctx context.Context, id string, body RightSizingMultiScopePolicy) (*GetMultiScopePolicyResponse, error)
	Delete(ctx context.Context, id string, force bool) error
}

type rightSizingPoliciesClient struct {
	http rightSizingHTTPClient
}

func newRightSizingPoliciesClient(http rightSizingHTTPClient) *rightSizingPoliciesClient {
	return &rightSizingPoliciesClient{http: http}
}

func (c *rightSizingPoliciesClient) GetAll(ctx context.Context) ([]GetAllRightSizingPoliciesRow, error) {
	body, _, err := c.http.Get(ctx, rsPoliciesPath, nil)
	if err != nil {
		return nil, fmt.Errorf("list right-sizing policies: %w", err)
	}
	var resp GetAllRightSizingPoliciesResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}
	return resp.Policies, nil
}

func (c *rightSizingPoliciesClient) GetByID(ctx context.Context, id string) (*GetMultiScopePolicyResponse, int, error) {
	body, status, err := c.http.Get(ctx, rsPolicyByIdPath(id), nil)
	if err != nil {
		return nil, status, fmt.Errorf("get right-sizing policy by id: %w", err)
	}
	var resp GetMultiScopePolicyResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, status, fmt.Errorf("decode get response: %w", err)
	}
	return &resp, status, nil
}

func (c *rightSizingPoliciesClient) GetByName(ctx context.Context, name string) (*GetMultiScopePolicyResponse, int, error) {
	rows, err := c.GetAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	for _, row := range rows {
		if row.Name == name {
			return c.GetByID(ctx, row.Id)
		}
	}
	return nil, http.StatusNotFound, fmt.Errorf("right-sizing policy with name %q not found", name)
}

func (c *rightSizingPoliciesClient) Create(ctx context.Context, body RightSizingMultiScopePolicy) (*GetMultiScopePolicyResponse, error) {
	respBody, _, err := c.http.Post(ctx, rsPoliciesPath, body)
	if err != nil {
		return nil, fmt.Errorf("create right-sizing policy: %w", err)
	}
	var resp GetMultiScopePolicyResponse
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("decode create response: %w", err)
	}
	return &resp, nil
}

func (c *rightSizingPoliciesClient) Update(ctx context.Context, id string, body RightSizingMultiScopePolicy) (*GetMultiScopePolicyResponse, error) {
	respBody, _, err := c.http.Put(ctx, rsPolicyByIdPath(id), body)
	if err != nil {
		return nil, fmt.Errorf("update right-sizing policy: %w", err)
	}
	var resp GetMultiScopePolicyResponse
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("decode update response: %w", err)
	}
	return &resp, nil
}

func (c *rightSizingPoliciesClient) Delete(ctx context.Context, id string, force bool) error {
	var query url.Values
	if force {
		query = url.Values{"force": []string{"true"}}
	}
	if _, err := c.http.Delete(ctx, rsPolicyByIdPath(id), query); err != nil {
		return fmt.Errorf("delete right-sizing policy: %w", err)
	}
	return nil
}
