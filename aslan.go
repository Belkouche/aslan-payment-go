// Package aslan provides a Go client for the Aslan Payment API.
//
// Create a client with your secret key:
//
//	client, err := aslan.NewClient("sk_test_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Use functional options for custom configuration:
//
//	client, err := aslan.NewClient("sk_test_xxx",
//	    aslan.WithBaseURL("https://staging.aslanpay.ma/pay"),
//	    aslan.WithTimeout(60 * time.Second),
//	    aslan.WithMaxRetries(3),
//	)
package aslan

import (
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://api.aslanpay.ma/pay"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
)

// Client is the Aslan Payment API client.
type Client struct {
	// CheckoutSessions provides access to checkout session operations.
	CheckoutSessions *CheckoutSessionsResource
	// Transactions provides access to transaction operations.
	Transactions *TransactionsResource
	// PaymentLinks provides access to payment link operations.
	PaymentLinks *PaymentLinksResource
	// Refunds provides access to refund operations.
	Refunds *RefundsResource
	// Vendors provides access to vendor/sub-merchant operations.
	Vendors *VendorsResource
	// Customers provides access to customer record operations.
	Customers *CustomersResource
	// ApiKeys provides access to API key management operations.
	ApiKeys *ApiKeysResource
	// WebhooksConfig provides access to webhook configuration operations.
	WebhooksConfig *WebhooksConfigResource

	secretKey  string
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries on 5xx/network errors.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// NewClient creates a new Aslan Payment API client.
//
// The secret key must start with "sk_live_" or "sk_test_".
func NewClient(secretKey string, opts ...Option) (*Client, error) {
	if !strings.HasPrefix(secretKey, "sk_live_") && !strings.HasPrefix(secretKey, "sk_test_") {
		return nil, &AslanError{
			Code:    ErrAuthentication,
			Message: "invalid secret key format: must start with sk_live_ or sk_test_",
		}
	}

	c := &Client{
		secretKey:  secretKey,
		baseURL:    defaultBaseURL,
		timeout:    defaultTimeout,
		maxRetries: defaultMaxRetries,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.CheckoutSessions = &CheckoutSessionsResource{client: c}
	c.Transactions = &TransactionsResource{client: c}
	c.PaymentLinks = &PaymentLinksResource{client: c}
	c.Refunds = &RefundsResource{client: c}
	c.Vendors = &VendorsResource{client: c}
	c.Customers = &CustomersResource{client: c}
	c.ApiKeys = &ApiKeysResource{client: c}
	c.WebhooksConfig = &WebhooksConfigResource{client: c}

	return c, nil
}

// IsTestMode returns true if the client is using a test mode secret key.
func (c *Client) IsTestMode() bool {
	return strings.HasPrefix(c.secretKey, "sk_test_")
}

// requestOpts returns the base request options for all API calls.
func (c *Client) requestOpts() requestOptions {
	return requestOptions{
		secretKey:  c.secretKey,
		baseURL:    c.baseURL,
		timeout:    c.timeout,
		maxRetries: c.maxRetries,
	}
}
