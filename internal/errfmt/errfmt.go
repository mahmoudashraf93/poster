package errfmt

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/mahmoudashraf93/poster/internal/config"
	"github.com/mahmoudashraf93/poster/internal/graph"
)

func Format(err error) string {
	if err == nil {
		return ""
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return formatParseError(parseErr)
	}

	var missing *config.MissingEnvError
	if errors.As(err, &missing) && len(missing.Missing) > 0 {
		return fmt.Sprintf("Missing required environment variables: %s", strings.Join(missing.Missing, ", "))
	}

	var apiErr *graph.GraphAPIError
	if errors.As(err, &apiErr) {
		if apiErr.Code != 0 && apiErr.Type != "" {
			return fmt.Sprintf("Graph API error (%d %s): %s", apiErr.Code, apiErr.Type, apiErr.Message)
		}

		if apiErr.Code != 0 {
			return fmt.Sprintf("Graph API error (%d): %s", apiErr.Code, apiErr.Message)
		}

		return fmt.Sprintf("Graph API error: %s", apiErr.Message)
	}

	if isNetworkError(err) {
		return fmt.Sprintf("Network error: %s (check your connection)", err.Error())
	}

	return err.Error()
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	msg := err.Error()
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") {
		return true
	}

	return false
}

func formatParseError(err *kong.ParseError) string {
	msg := err.Error()

	if strings.Contains(msg, "did you mean") {
		return msg
	}

	if strings.HasPrefix(msg, "unknown flag") {
		return msg + "\nRun with --help to see available flags"
	}

	if strings.Contains(msg, "missing") || strings.Contains(msg, "required") {
		return msg + "\nRun with --help to see usage"
	}

	return msg
}
