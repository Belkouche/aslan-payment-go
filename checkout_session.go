package aslan

import (
	"context"
	"net/url"
)

// CheckoutSessionsResource provides methods for managing checkout sessions.
type CheckoutSessionsResource struct {
	client *Client
}

// Create creates a new checkout session.
func (r *CheckoutSessionsResource) Create(ctx context.Context, params *CreateCheckoutSessionParams) (*CheckoutSession, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/checkout/sessions"
	opts.body = params

	if params.IdempotencyKey != "" {
		opts.headers = map[string]string{"Idempotency-Key": params.IdempotencyKey}
	}

	var session CheckoutSession
	if err := doRequest(ctx, opts, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// Retrieve retrieves a checkout session by ID.
func (r *CheckoutSessionsResource) Retrieve(ctx context.Context, id string) (*CheckoutSession, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/checkout/sessions/" + url.PathEscape(id)

	var session CheckoutSession
	if err := doRequest(ctx, opts, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
