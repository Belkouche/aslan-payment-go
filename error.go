package aslan

import "fmt"

// Error codes returned by the Aslan Payment API.
const (
	ErrAuthentication = "AUTHENTICATION_ERROR"
	ErrAuthorization  = "AUTHORIZATION_ERROR"
	ErrValidation     = "VALIDATION_ERROR"
	ErrNotFound       = "NOT_FOUND"
	ErrRateLimited    = "RATE_LIMITED"
	ErrIdempotency    = "IDEMPOTENCY_ERROR"
	ErrNetwork        = "NETWORK_ERROR"
	ErrServer         = "SERVER_ERROR"
)

// AslanError represents an error returned by the Aslan Payment API.
type AslanError struct {
	// Code is one of the Err* constants (e.g. ErrAuthentication).
	Code string
	// Message is a human-readable description of the error.
	Message string
	// HTTPStatus is the HTTP status code of the response, if available.
	HTTPStatus int
	// Details contains additional structured error information from the API.
	Details map[string]interface{}
	// RequestID is the value of the x-request-id response header, if present.
	RequestID string
}

// Error implements the error interface.
func (e *AslanError) Error() string {
	return fmt.Sprintf("aslan: %s: %s", e.Code, e.Message)
}

// mapErrorCode maps an HTTP status and optional server error code to an AslanError code.
func mapErrorCode(status int, serverCode string) string {
	if serverCode == ErrValidation {
		return ErrValidation
	}
	if serverCode == ErrIdempotency {
		return ErrIdempotency
	}
	switch status {
	case 401:
		return ErrAuthentication
	case 403:
		return ErrAuthorization
	case 404:
		return ErrNotFound
	case 409:
		return ErrIdempotency
	case 422:
		return ErrValidation
	case 429:
		return ErrRateLimited
	default:
		if status >= 500 {
			return ErrServer
		}
		return ErrValidation
	}
}
