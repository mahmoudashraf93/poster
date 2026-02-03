package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mahmoud/igpostercli/internal/config"
)

type JSON map[string]any

type Client struct {
	httpClient   *http.Client
	baseURL      string
	graphVersion string
	accessToken  string
	igUserID     string
}

func NewClient(cfg *config.Config) *Client {
	version := cfg.GraphVersion
	if version == "" {
		version = config.DefaultGraphVersion
	}

	return &Client{
		httpClient:   http.DefaultClient,
		baseURL:      "https://graph.facebook.com/",
		graphVersion: version,
		accessToken:  cfg.AccessToken,
		igUserID:     cfg.IGUserID,
	}
}

func (c *Client) post(ctx context.Context, path string, params map[string]string) (JSON, error) {
	return c.do(ctx, http.MethodPost, path, params)
}

func (c *Client) get(ctx context.Context, path string, params map[string]string) (JSON, error) {
	return c.do(ctx, http.MethodGet, path, params)
}

func (c *Client) do(ctx context.Context, method, path string, params map[string]string) (JSON, error) {
	if c == nil {
		return nil, ErrGraphClientNil
	}

	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	if c.accessToken != "" {
		values.Set("access_token", c.accessToken)
	}

	endpoint := c.endpoint(path)
	var body io.Reader

	if method == http.MethodGet {
		if encoded := values.Encode(); encoded != "" {
			endpoint = endpoint + "?" + encoded
		}
	} else {
		body = strings.NewReader(values.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graph request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if apiErr := parseGraphAPIError(payload); apiErr != nil {
			return nil, apiErr
		}

		return nil, fmt.Errorf("%w: status %d", ErrGraphAPIStatus, resp.StatusCode)
	}

	var parsed JSON
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if apiErr := apiErrorFromJSON(parsed); apiErr != nil {
		return nil, apiErr
	}

	return parsed, nil
}

func (c *Client) endpoint(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	base := strings.TrimSuffix(c.baseURL, "/")

	return fmt.Sprintf("%s/%s/%s", base, c.graphVersion, trimmed)
}

type GraphAPIError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

func (e *GraphAPIError) Error() string {
	if e == nil {
		return ""
	}

	if e.Code != 0 && e.Type != "" {
		return fmt.Sprintf("graph api error (%d %s): %s", e.Code, e.Type, e.Message)
	}

	if e.Code != 0 {
		return fmt.Sprintf("graph api error (%d): %s", e.Code, e.Message)
	}

	return fmt.Sprintf("graph api error: %s", e.Message)
}

type graphErrorResponse struct {
	Error *GraphAPIError `json:"error"`
}

func parseGraphAPIError(payload []byte) *GraphAPIError {
	var response graphErrorResponse
	_ = json.Unmarshal(payload, &response)

	return response.Error
}

func apiErrorFromJSON(payload JSON) *GraphAPIError {
	value, ok := payload["error"]
	if !ok {
		return nil
	}

	raw, _ := json.Marshal(value)

	return parseGraphAPIError(raw)
}
