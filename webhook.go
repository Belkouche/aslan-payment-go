package aslan

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// VerifyWebhook verifies a webhook signature and parses the event payload.
// Uses a default tolerance of 300 seconds (5 minutes).
//
// The signature format is "t=<unix_timestamp>,v1=<hex_hmac>".
func VerifyWebhook(webhookSecret string, payload []byte, signature string) (*WebhookEvent, error) {
	return VerifyWebhookWithTolerance(webhookSecret, payload, signature, 300)
}

// VerifyWebhookWithTolerance verifies a webhook signature with a custom tolerance in seconds.
//
// The signature format is "t=<unix_timestamp>,v1=<hex_hmac>".
func VerifyWebhookWithTolerance(webhookSecret string, payload []byte, signature string, toleranceSeconds int64) (*WebhookEvent, error) {
	parts := strings.Split(signature, ",")

	var timestampStr, sig string
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestampStr = part[2:]
		} else if strings.HasPrefix(part, "v1=") {
			sig = part[3:]
		}
	}

	if timestampStr == "" || sig == "" {
		return nil, &AslanError{Code: ErrValidation, Message: "invalid webhook signature format"}
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, &AslanError{Code: ErrValidation, Message: "invalid webhook timestamp"}
	}

	now := time.Now().Unix()
	if int64(math.Abs(float64(now-timestamp))) > toleranceSeconds {
		return nil, &AslanError{Code: ErrValidation, Message: "webhook timestamp too old"}
	}

	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(signedPayload))
	computedSig := hex.EncodeToString(mac.Sum(nil))

	expectedBytes, err := hex.DecodeString(sig)
	if err != nil {
		return nil, &AslanError{Code: ErrValidation, Message: "invalid webhook signature encoding"}
	}
	computedBytes, _ := hex.DecodeString(computedSig)

	if len(expectedBytes) != len(computedBytes) || subtle.ConstantTimeCompare(expectedBytes, computedBytes) != 1 {
		return nil, &AslanError{Code: ErrValidation, Message: "webhook signature verification failed"}
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, &AslanError{Code: ErrValidation, Message: "invalid webhook payload JSON"}
	}

	return &event, nil
}
