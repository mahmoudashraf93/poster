package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func ensureHTTPS(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("url must be https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("url must include host")
	}
	return raw, nil
}

func detectMediaType(path string) (isVideo bool, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return false, nil
	case ".mp4", ".mov":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
