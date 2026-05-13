package komodor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	rsPoliciesPath = "/api/v2/cost/right-sizing/policies"
)

var _ rightSizingHTTPClient = (*rightSizingHTTP)(nil)

func rsPolicyByIdPath(id string) string {
	return rsPoliciesPath + "/" + url.PathEscape(id)
}

type rightSizingHTTPClient interface {
	Get(ctx context.Context, path string, query url.Values) ([]byte, int, error)
	Post(ctx context.Context, path string, body any) ([]byte, int, error)
	Put(ctx context.Context, path string, body any) ([]byte, int, error)
	Delete(ctx context.Context, path string, query url.Values) (int, error)
}

type rightSizingHTTP struct {
	http    *http.Client
	baseURL string
	apiKey  string
}

func newRightSizingHTTP(baseURL, apiKey string) *rightSizingHTTP {
	return &rightSizingHTTP{
		http:    &http.Client{},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (h *rightSizingHTTP) fullURL(path string, query url.Values) string {
	full := h.baseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	return full
}

func (h *rightSizingHTTP) do(ctx context.Context, method, fullURL string, payload []byte) ([]byte, int, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", h.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Terraform (terraform-provider-komodor); Go-http-client/1.1")

	res, err := h.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("read response: %w", err)
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return resBody, res.StatusCode, nil
	default:
		return resBody, res.StatusCode, fmt.Errorf("received %d: %s", res.StatusCode, resBody)
	}
}

func (h *rightSizingHTTP) Get(ctx context.Context, path string, query url.Values) ([]byte, int, error) {
	return h.do(ctx, http.MethodGet, h.fullURL(path, query), nil)
}

func (h *rightSizingHTTP) Post(ctx context.Context, path string, body any) ([]byte, int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request body: %w", err)
	}
	return h.do(ctx, http.MethodPost, h.fullURL(path, nil), payload)
}

func (h *rightSizingHTTP) Put(ctx context.Context, path string, body any) ([]byte, int, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request body: %w", err)
	}
	return h.do(ctx, http.MethodPut, h.fullURL(path, nil), payload)
}

func (h *rightSizingHTTP) Delete(ctx context.Context, path string, query url.Values) (int, error) {
	_, status, err := h.do(ctx, http.MethodDelete, h.fullURL(path, query), nil)
	return status, err
}
