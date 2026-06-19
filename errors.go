package einvoice

import (
	"fmt"
	"time"
)

// EInvoiceError is the base error type for all SDK errors that are not API,
// rate-limit, timeout, config, or webhook errors.
type EInvoiceError struct {
	Message string
}

func (e *EInvoiceError) Error() string { return "einvoice: " + e.Message }

// FieldError describes a single validation/error entry returned by the API.
type FieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// APIError represents a non-2xx response from the E-Invoice platform.
//
//	var apiErr *einvoice.APIError
//	if errors.As(err, &apiErr) {
//		fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Errors)
//	}
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Code is the platform error code (e.g. "INVOICE_NOT_FOUND"), if provided.
	Code string
	// Message is a human-readable error message.
	Message string
	// Errors holds field-level validation errors, if any.
	Errors []FieldError
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("einvoice: API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("einvoice: API error %d: %s", e.StatusCode, e.Message)
}

// RateLimitError is returned when the API responds with HTTP 429.
// It embeds *APIError, so errors.As(err, &apiErr) also matches.
type RateLimitError struct {
	*APIError
	// RetryAfter is the number of seconds the caller should wait before retrying.
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("einvoice: rate limit exceeded, retry after %ds", e.RetryAfter)
}

// Unwrap exposes the embedded *APIError to errors.As / errors.Is.
func (e *RateLimitError) Unwrap() error { return e.APIError }

// TimeoutError is returned when a request exceeds its configured timeout.
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("einvoice: request timed out after %s", e.Timeout)
}

// ConfigError is returned when the SDK is misconfigured (e.g. missing API key
// or missing organization ID for an org-scoped operation).
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string { return "einvoice: " + e.Message }

// WebhookError is returned when webhook signature verification fails.
type WebhookError struct {
	Message string
}

func (e *WebhookError) Error() string { return "einvoice: " + e.Message }
