package aslan

import (
	"context"
	"fmt"
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

// Update updates a payment link by ID.
func (r *PaymentLinksResource) Update(ctx context.Context, id string, params *UpdatePaymentLinkParams) (*PaymentLink, error) {
	opts := r.client.requestOpts()
	opts.method = "PATCH"
	opts.path = "/api/v1/payment-links/" + url.PathEscape(id)
	opts.body = params

	var link PaymentLink
	if err := doRequest(ctx, opts, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Delete deletes a payment link by ID.
func (r *PaymentLinksResource) Delete(ctx context.Context, id string) error {
	opts := r.client.requestOpts()
	opts.method = "DELETE"
	opts.path = "/api/v1/payment-links/" + url.PathEscape(id)

	return doRequest(ctx, opts, nil)
}

// GetQR generates a QR code for a payment link.
func (r *PaymentLinksResource) GetQR(ctx context.Context, id string, params *QRCodeParams) (*QRCodeResult, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/payment-links/" + url.PathEscape(id) + "/qr"

	if params != nil {
		query := make(map[string]string)
		if params.Format != "" {
			query["format"] = params.Format
		}
		if params.Size > 0 {
			query["size"] = fmt.Sprintf("%d", params.Size)
		}
		if len(query) > 0 {
			opts.query = query
		}
	}

	var qr QRCodeResult
	if err := doRequest(ctx, opts, &qr); err != nil {
		return nil, err
	}
	return &qr, nil
}
