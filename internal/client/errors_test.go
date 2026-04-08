package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
)

func TestFormatError_Unauthorized(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: http.StatusUnauthorized, Message: "Invalid API key"}
	got := FormatError(err)
	expected := "Error: Invalid API key. Run 'bunny configure' to update your credentials."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NotFound(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: http.StatusNotFound, Message: "Not found"}
	got := FormatError(err)
	expected := "Error: Resource not found."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_RateLimited(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: http.StatusTooManyRequests, Message: "Too many requests"}
	got := FormatError(err)
	expected := "Error: Rate limit exceeded. Try again in a moment."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_BadRequest(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: http.StatusBadRequest, Message: "The record is invalid."}
	got := FormatError(err)
	expected := "Error: The record is invalid."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_BadRequestWithField(t *testing.T) {
	t.Parallel()
	err := &APIError{
		HTTPStatusCode: http.StatusBadRequest,
		Message:        "The record is invalid.",
		Field:          "Name",
	}
	got := FormatError(err)
	if got != "Error: The record is invalid.\n  - Name" {
		t.Errorf("unexpected error format: %q", got)
	}
}

func TestFormatError_ServerError(t *testing.T) {
	t.Parallel()
	for _, status := range []int{500, 502, 503} {
		err := &APIError{HTTPStatusCode: status, Message: "Server error"}
		got := FormatError(err)
		expected := "Error: bunny.net API server error. Please try again later."
		if got != expected {
			t.Errorf("status %d: expected %q, got %q", status, expected, got)
		}
	}
}

func TestFormatError_GenericAPIError(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: http.StatusForbidden, Message: "Forbidden"}
	got := FormatError(err)
	expected := "Error: Forbidden"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NetworkError_URLError(t *testing.T) {
	t.Parallel()
	err := &url.Error{
		Op:  "Get",
		URL: "https://api.bunny.net/pullzone",
		Err: &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: fmt.Errorf("connection refused"),
		},
	}
	got := FormatError(err)
	expected := "Error: Unable to connect to bunny.net API. Check your network connection."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_NetworkError_DNSError(t *testing.T) {
	t.Parallel()
	err := &url.Error{
		Op:  "Get",
		URL: "https://api.bunny.net/pullzone",
		Err: &net.DNSError{
			Name: "api.bunny.net",
			Err:  "no such host",
		},
	}
	got := FormatError(err)
	expected := "Error: Unable to connect to bunny.net API. Check your network connection."
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFormatError_GenericError(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("API key not configured, run 'bunny configure' or set BUNNY_API_KEY")
	got := FormatError(err)
	expected := "Error: API key not configured, run 'bunny configure' or set BUNNY_API_KEY"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestAPIError_Error_WithMessage(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: 400, Message: "Bad request"}
	if got := err.Error(); got != "Bad request" {
		t.Errorf("expected %q, got %q", "Bad request", got)
	}
}

func TestAPIError_Error_WithField(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: 400, Message: "is required", Field: "Name"}
	if got := err.Error(); got != "Name: is required" {
		t.Errorf("expected %q, got %q", "Name: is required", got)
	}
}

func TestAPIError_Error_NoMessage(t *testing.T) {
	t.Parallel()
	err := &APIError{HTTPStatusCode: 500}
	if got := err.Error(); got != "HTTP 500" {
		t.Errorf("expected %q, got %q", "HTTP 500", got)
	}
}
