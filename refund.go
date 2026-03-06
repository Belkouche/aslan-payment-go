package aslan

import (
	"context"
	"net/url"
)

// RefundsResource provides methods for managing refunds.
type RefundsResource struct {
	client *Client
}

// Create creates a new refund for a transaction.
func (r *RefundsResource) Create(ctx context.Context, params *CreateRefundParams) (*Refund, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/refunds"
	opts.body = params

	if params.IdempotencyKey != "" {
		opts.headers = map[string]string{"Idempotency-Key": params.IdempotencyKey}
	}

	var refund Refund
	if err := doRequest(ctx, opts, &refund); err != nil {
		return nil, err
	}
	return &refund, nil
}

// Retrieve retrieves a refund by ID.
func (r *RefundsResource) Retrieve(ctx context.Context, id string) (*Refund, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/refunds/" + url.PathEscape(id)

	var refund Refund
	if err := doRequest(ctx, opts, &refund); err != nil {
		return nil, err
	}
	return &refund, nil
}
