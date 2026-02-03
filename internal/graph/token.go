package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mahmoudashraf93/poster/internal/config"
)

var graphBaseURL = "https://graph.facebook.com/"

func ExchangeToken(ctx context.Context, appID, appSecret, shortToken string) (string, error) {
	params := map[string]string{
		"grant_type":        "fb_exchange_token",
		"client_id":         appID,
		"client_secret":     appSecret,
		"fb_exchange_token": shortToken,
	}

	resp, err := graphGet(ctx, "oauth/access_token", params)
	if err != nil {
		return "", err
	}

	value, ok := resp["access_token"].(string)
	if !ok || value == "" {
		return "", ErrMissingAccessToken
	}

	return value, nil
}

func DebugToken(ctx context.Context, token string) (*TokenInfo, error) {
	params := map[string]string{
		"input_token":  token,
		"access_token": token,
	}

	resp, err := graphGet(ctx, "debug_token", params)
	if err != nil {
		return nil, err
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		return nil, ErrMissingTokenData
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("parse token data: %w", err)
	}

	var info TokenInfo
	if err := json.Unmarshal(payload, &info); err != nil {
		return nil, fmt.Errorf("parse token data: %w", err)
	}

	return &info, nil
}

func FetchIGUserID(ctx context.Context, pageID, accessToken string) (string, error) {
	params := map[string]string{
		"fields":       "instagram_business_account",
		"access_token": accessToken,
	}

	resp, err := graphGet(ctx, pageID, params)
	if err != nil {
		return "", err
	}

	account, ok := resp["instagram_business_account"].(map[string]any)
	if !ok {
		return "", ErrMissingIGAccount
	}

	id, ok := account["id"].(string)
	if !ok || id == "" {
		return "", ErrMissingIGAccountID
	}

	return id, nil
}

type OwnedPage struct {
	ID       string
	Name     string
	IGUserID string
}

func FetchOwnedPages(ctx context.Context, businessID, accessToken string) ([]OwnedPage, error) {
	params := map[string]string{
		"fields":       "id,name,instagram_business_account",
		"access_token": accessToken,
	}

	path := fmt.Sprintf("%s/owned_pages", businessID)

	resp, err := graphGet(ctx, path, params)
	if err != nil {
		return nil, err
	}

	pages, next, err := parseOwnedPages(resp)
	if err != nil {
		return nil, err
	}

	for next != "" {
		resp, err = graphGetURL(ctx, next)
		if err != nil {
			return nil, err
		}

		var batch []OwnedPage

		batch, next, err = parseOwnedPages(resp)
		if err != nil {
			return nil, err
		}

		pages = append(pages, batch...)
	}

	return pages, nil
}

type ownedPagesResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		IG   struct {
			ID string `json:"id"`
		} `json:"instagram_business_account"`
	} `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func parseOwnedPages(payload JSON) ([]OwnedPage, string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("parse owned pages: %w", err)
	}

	var parsed ownedPagesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, "", fmt.Errorf("parse owned pages: %w", err)
	}

	pages := make([]OwnedPage, 0, len(parsed.Data))
	for _, entry := range parsed.Data {
		pages = append(pages, OwnedPage{
			ID:       entry.ID,
			Name:     entry.Name,
			IGUserID: entry.IG.ID,
		})
	}

	return pages, parsed.Paging.Next, nil
}

type TokenInfo struct {
	AppID               string   `json:"app_id"`
	Type                string   `json:"type"`
	Application         string   `json:"application"`
	DataAccessExpiresAt int64    `json:"data_access_expires_at"`
	ExpiresAt           int64    `json:"expires_at"`
	IsValid             bool     `json:"is_valid"`
	Scopes              []string `json:"scopes"`
	UserID              string   `json:"user_id"`
}

func graphGet(ctx context.Context, path string, params map[string]string) (JSON, error) {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}

	endpoint := graphEndpoint(config.DefaultGraphVersion, path)
	if encoded := values.Encode(); encoded != "" {
		endpoint = endpoint + "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func graphGetURL(ctx context.Context, endpoint string) (JSON, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func graphEndpoint(version, path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	base := strings.TrimSuffix(graphBaseURL, "/")

	return fmt.Sprintf("%s/%s/%s", base, version, trimmed)
}
