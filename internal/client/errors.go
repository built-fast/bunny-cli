package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// APIError represents a bunny.net API error response.
type APIError struct {
	HTTPStatusCode int    `json:"-"`
	ErrorKey       string `json:"ErrorKey"`
	Field          string `json:"Field"`
	Message        string `json:"Message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		if e.Field != "" {
			return fmt.Sprintf("%s: %s", e.Field, e.Message)
		}
		return e.Message
	}
	return fmt.Sprintf("HTTP %d", e.HTTPStatusCode)
}

// FormatError formats an error into a user-friendly message for stderr display.
func FormatError(err error) string {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return formatAPIError(apiErr)
	}

	if isNetworkError(err) {
		return "Error: Unable to connect to bunny.net API. Check your network connection."
	}

	return fmt.Sprintf("Error: %s", err.Error())
}

func formatAPIError(err *APIError) string {
	switch err.HTTPStatusCode {
	case http.StatusUnauthorized:
		return "Error: Invalid API key. Run 'bunny configure' to update your credentials."
	case http.StatusNotFound:
		return "Error: Resource not found."
	case http.StatusTooManyRequests:
		return "Error: Rate limit exceeded. Try again in a moment."
	case http.StatusBadRequest:
		return formatValidationError(err)
	default:
		if err.HTTPStatusCode >= 500 {
			return "Error: bunny.net API server error. Please try again later."
		}
		if err.Message != "" {
			return fmt.Sprintf("Error: %s", err.Message)
		}
		return fmt.Sprintf("Error: HTTP %d", err.HTTPStatusCode)
	}
}

func formatValidationError(err *APIError) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: %s", err.Message)
	if err.Field != "" {
		fmt.Fprintf(&b, "\n  - %s", err.Field)
	}
	return b.String()
}

// parseErrorResponse reads the response body and returns an *APIError.
// Tries JSON first, falls back to plain text message.
func parseErrorResponse(resp *http.Response) *APIError {
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return &APIError{HTTPStatusCode: resp.StatusCode}
	}

	apiErr := &APIError{HTTPStatusCode: resp.StatusCode}
	if json.Unmarshal(body, apiErr) == nil && (apiErr.Message != "" || apiErr.ErrorKey != "") {
		return apiErr
	}

	// Fall back to plain text
	return &APIError{
		HTTPStatusCode: resp.StatusCode,
		Message:        strings.TrimSpace(string(body)),
	}
}

func isNetworkError(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return true
	}
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}
