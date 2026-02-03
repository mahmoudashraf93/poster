package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

var uguuUploadURL = "https://uguu.se/upload.php"

func Upload(ctx context.Context, filepath string) (string, error) {
	// #nosec G304 -- filepath is user-provided
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("files[]", file.Name())
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uguuUploadURL, &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: status %d", ErrUploadFailed, resp.StatusCode)
	}

	var parsed uguuResponse
	if err = json.Unmarshal(payload, &parsed); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if !parsed.Success {
		if parsed.Error != "" {
			return "", fmt.Errorf("%w: %s", ErrUploadFailed, parsed.Error)
		}

		return "", ErrUploadFailed
	}

	if len(parsed.Files) == 0 || parsed.Files[0].URL == "" {
		return "", ErrUploadMissingURL
	}

	publicURL := parsed.Files[0].URL

	parsedURL, err := url.Parse(publicURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUploadInvalidURL, err)
	}

	if parsedURL.Scheme != "https" {
		return "", fmt.Errorf("%w: %s", ErrInvalidURLScheme, parsedURL.Scheme)
	}

	return publicURL, nil
}

type uguuResponse struct {
	Success bool `json:"success"`
	Files   []struct {
		URL string `json:"url"`
	} `json:"files"`
	Error string `json:"error"`
}
