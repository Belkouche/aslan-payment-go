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

// List lists refunds with optional filters.
func (r *RefundsResource) List(ctx context.Context, params *ListRefundsParams) (*PaginatedResponse[Refund], error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/refunds"

	if params != nil {
		query := make(map[string]string)
		if params.Page != nil {
			query["page"] = intToString(params.Page)
		}
		if params.PageSize != nil {
			query["page_size"] = intToString(params.PageSize)
		}
		if params.Status != nil {
			query["status"] = *params.Status
		}
		if params.TransactionID != nil {
			query["transaction_id"] = *params.TransactionID
		}
		if len(query) > 0 {
			opts.query = query
		}
	}

	var result PaginatedResponse[Refund]
	if err := doRequest(ctx, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
