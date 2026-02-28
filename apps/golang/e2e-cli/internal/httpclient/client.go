package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	token    string
	tenantID string
	http     *http.Client
}

func New(baseURL, token, tenantID string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		token:    token,
		tenantID: tenantID,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) SetTenantID(id string) {
	c.tenantID = id
}

func (c *Client) GetJSON(ctx context.Context, path string, out any) (int, []byte, error) {
	return c.doJSON(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) PostJSON(ctx context.Context, path string, in any, out any) (int, []byte, error) {
	return c.doJSON(ctx, http.MethodPost, path, in, out)
}

func (c *Client) PutJSON(ctx context.Context, path string, in any, out any) (int, []byte, error) {
	return c.doJSON(ctx, http.MethodPut, path, in, out)
}

func (c *Client) Delete(ctx context.Context, path string) (int, []byte, error) {
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) PatchJSON(ctx context.Context, path string, in any, out any) (int, []byte, error) {
	return c.doJSON(ctx, http.MethodPatch, path, in, out)
}

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any) (int, []byte, error) {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read body: %w", err)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return resp.StatusCode, respBody, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, respBody, nil
}
