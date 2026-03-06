package aslan

import (
	"context"
	"net/url"
)

// PaymentLinksResource provides methods for managing payment links.
type PaymentLinksResource struct {
	client *Client
}

// Create creates a new payment link.
func (r *PaymentLinksResource) Create(ctx context.Context, params *CreatePaymentLinkParams) (*PaymentLink, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/payment-links"
	opts.body = params

	if params.IdempotencyKey != "" {
		opts.headers = map[string]string{"Idempotency-Key": params.IdempotencyKey}
	}

	var link PaymentLink
	if err := doRequest(ctx, opts, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Retrieve retrieves a payment link by ID.
func (r *PaymentLinksResource) Retrieve(ctx context.Context, id string) (*PaymentLink, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/payment-links/" + url.PathEscape(id)

	var link PaymentLink
	if err := doRequest(ctx, opts, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// List lists payment links with optional filters.
func (r *PaymentLinksResource) List(ctx context.Context, params *ListPaymentLinksParams) (*PaginatedResponse[PaymentLink], error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/payment-links"

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
		if len(query) > 0 {
			opts.query = query
		}
	}

	var result PaginatedResponse[PaymentLink]
	if err := doRequest(ctx, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
