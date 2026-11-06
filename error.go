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

	// Provider account errors (HTTP 400), returned when a providerAccount alias is rejected.
	ErrProviderAccountNotFound          = "PROVIDER_ACCOUNT_NOT_FOUND"
	ErrProviderAccountInactive          = "PROVIDER_ACCOUNT_INACTIVE"
	ErrProviderAccountUnsupported       = "PROVIDER_ACCOUNT_UNSUPPORTED"
	ErrProviderAccountRecurringDisabled = "PROVIDER_ACCOUNT_RECURRING_DISABLED"
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
	switch serverCode {
	case ErrProviderAccountNotFound, ErrProviderAccountInactive, ErrProviderAccountUnsupported, ErrProviderAccountRecurringDisabled:
		return serverCode
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
