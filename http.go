package aslan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const sdkVersion = "0.1.0"

// retryableStatusCodes are HTTP status codes that trigger a retry.
var retryableStatusCodes = map[int]bool{
	408: true,
	429: true,
	500: true,
	502: true,
	503: true,
	504: true,
}

type requestOptions struct {
	method     string
	path       string
	body       interface{}
	query      map[string]string
	headers    map[string]string
	secretKey  string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	// unwrapData decodes the success body's "data" field into the result
	// (the API wraps most single-object responses as { "data": X }).
	unwrapData bool
}

// parseAPIError extracts code, message and details from an error response body.
// It accepts the flat shape { statusCode, code, error, message, details } and the
// nested shape { error: { code, message, details } }.
func parseAPIError(body []byte) (code, message string, details map[string]interface{}) {
	var raw map[string]json.RawMessage
	if json.Unmarshal(body, &raw) != nil {
		return "", "", nil
	}

	var detailsRaw json.RawMessage
	var nested struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Details json.RawMessage `json:"details"`
	}
	if e, ok := raw["error"]; ok && len(e) > 0 && e[0] == '{' && json.Unmarshal(e, &nested) == nil {
		code, message, detailsRaw = nested.Code, nested.Message, nested.Details
	} else {
		_ = json.Unmarshal(raw["code"], &code)
		_ = json.Unmarshal(raw["message"], &message)
		if message == "" {
			// Flat shape: "error" is the HTTP reason phrase (e.g. "Bad Request").
			_ = json.Unmarshal(raw["error"], &message)
		}
		detailsRaw = raw["details"]
	}

	if len(detailsRaw) > 0 && string(detailsRaw) != "null" {
		if json.Unmarshal(detailsRaw, &details) != nil {
			// Non-object details (e.g. a [{field, message}] array) are kept under "errors".
			var v interface{}
			if json.Unmarshal(detailsRaw, &v) == nil {
				details = map[string]interface{}{"errors": v}
			}
		}
	}
	return code, message, details
}

// doRequest performs an HTTP request with retries and exponential backoff.
func doRequest(ctx context.Context, opts requestOptions, result interface{}) error {
	u, err := url.Parse(opts.baseURL)
	if err != nil {
		return &AslanError{Code: ErrNetwork, Message: "invalid base URL: " + err.Error()}
	}
	u.Path = u.Path + opts.path
	if opts.query != nil {
		q := u.Query()
		for k, v := range opts.query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	var bodyBytes []byte
	if opts.body != nil {
		bodyBytes, err = json.Marshal(opts.body)
		if err != nil {
			return &AslanError{Code: ErrValidation, Message: "failed to marshal request body: " + err.Error()}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	var lastErr error

	for attempt := 0; attempt <= opts.maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Min(float64(500*time.Millisecond)*math.Pow(2, float64(attempt-1)), float64(5*time.Second)))
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return &AslanError{Code: ErrNetwork, Message: "request cancelled: " + ctx.Err().Error()}
			case <-timer.C:
			}

			// Reset body reader for retry
			if bodyBytes != nil {
				bodyReader = bytes.NewReader(bodyBytes)
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, opts.timeout)
		req, err := http.NewRequestWithContext(reqCtx, opts.method, u.String(), bodyReader)
		if err != nil {
			cancel()
			return &AslanError{Code: ErrNetwork, Message: "failed to create request: " + err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+opts.secretKey)
		req.Header.Set("User-Agent", "aslan-payment-go/"+sdkVersion)
		for k, v := range opts.headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			lastErr = &AslanError{Code: ErrNetwork, Message: "network request failed: " + err.Error()}
			if attempt < opts.maxRetries {
				continue
			}
			return lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if err != nil {
			lastErr = &AslanError{Code: ErrNetwork, Message: "failed to read response body: " + err.Error()}
			if attempt < opts.maxRetries {
				continue
			}
			return lastErr
		}

		requestID := resp.Header.Get("x-request-id")

		// Retry on retryable status codes
		if retryableStatusCodes[resp.StatusCode] && attempt < opts.maxRetries {
			lastErr = &AslanError{
				Code:       ErrServer,
				Message:    fmt.Sprintf("request failed with status %d", resp.StatusCode),
				HTTPStatus: resp.StatusCode,
				RequestID:  requestID,
			}
			continue
		}

		// 204 No Content
		if resp.StatusCode == 204 {
			return nil
		}

		// Handle error responses
		if resp.StatusCode >= 400 {
			code, message, details := parseAPIError(respBody)
			if message == "" {
				message = fmt.Sprintf("request failed with status %d", resp.StatusCode)
			}

			return &AslanError{
				Code:       mapErrorCode(resp.StatusCode, code),
				Message:    message,
				HTTPStatus: resp.StatusCode,
				Details:    details,
				RequestID:  requestID,
			}
		}

		// Parse success response
		if result != nil {
			payload := respBody
			if opts.unwrapData {
				var env struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(respBody, &env); err != nil || len(env.Data) == 0 || string(env.Data) == "null" {
					return &AslanError{
						Code:       ErrServer,
						Message:    fmt.Sprintf("unexpected response from server (%d): missing data envelope", resp.StatusCode),
						HTTPStatus: resp.StatusCode,
						RequestID:  requestID,
					}
				}
				payload = env.Data
			}
			if err := json.Unmarshal(payload, result); err != nil {
				return &AslanError{
					Code:       ErrServer,
					Message:    fmt.Sprintf("unexpected response from server (%d)", resp.StatusCode),
					HTTPStatus: resp.StatusCode,
					RequestID:  requestID,
				}
			}
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return &AslanError{Code: ErrNetwork, Message: "request failed after retries"}
}

// intToString converts an *int to its string representation.
func intToString(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
