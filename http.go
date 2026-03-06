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
}

type apiErrorBody struct {
	Error *struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details"`
	} `json:"error"`
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
			var errBody apiErrorBody
			_ = json.Unmarshal(respBody, &errBody)

			code := ""
			message := fmt.Sprintf("request failed with status %d", resp.StatusCode)
			var details map[string]interface{}

			if errBody.Error != nil {
				code = errBody.Error.Code
				if errBody.Error.Message != "" {
					message = errBody.Error.Message
				}
				details = errBody.Error.Details
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
			if err := json.Unmarshal(respBody, result); err != nil {
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
