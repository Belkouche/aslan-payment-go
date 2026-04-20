package aslan

import (
	"context"
	"net/url"
	"strconv"
)

// VendorsResource provides methods for managing vendors.
type VendorsResource struct {
	client *Client
}

// Create creates a new vendor.
func (r *VendorsResource) Create(ctx context.Context, params *CreateVendorParams) (*Vendor, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/vendors"
	opts.body = params

	if params.IdempotencyKey != "" {
		opts.headers = map[string]string{"Idempotency-Key": params.IdempotencyKey}
	}

	var vendor Vendor
	if err := doRequest(ctx, opts, &vendor); err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Retrieve retrieves a vendor by ID.
func (r *VendorsResource) Retrieve(ctx context.Context, id string) (*Vendor, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/vendors/" + url.PathEscape(id)

	var vendor Vendor
	if err := doRequest(ctx, opts, &vendor); err != nil {
		return nil, err
	}
	return &vendor, nil
}

// List lists vendors with optional filters.
func (r *VendorsResource) List(ctx context.Context, params *ListVendorsParams) (*VendorListResponse, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/vendors"

	if params != nil {
		query := make(map[string]string)
		if params.Page != nil {
			query["page"] = intToString(params.Page)
		}
		if params.Limit != nil {
			query["limit"] = intToString(params.Limit)
		}
		if params.Search != nil {
			query["search"] = *params.Search
		}
		if params.IsActive != nil {
			query["isActive"] = strconv.FormatBool(*params.IsActive)
		}
		if len(query) > 0 {
			opts.query = query
		}
	}

	var result VendorListResponse
	if err := doRequest(ctx, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a vendor by ID.
func (r *VendorsResource) Update(ctx context.Context, id string, params *UpdateVendorParams) (*Vendor, error) {
	opts := r.client.requestOpts()
	opts.method = "PATCH"
	opts.path = "/api/v1/vendors/" + url.PathEscape(id)
	opts.body = params

	var vendor Vendor
	if err := doRequest(ctx, opts, &vendor); err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Delete soft-deletes a vendor by ID and returns the updated vendor record.
func (r *VendorsResource) Delete(ctx context.Context, id string) (*Vendor, error) {
	opts := r.client.requestOpts()
	opts.method = "DELETE"
	opts.path = "/api/v1/vendors/" + url.PathEscape(id)

	var vendor Vendor
	if err := doRequest(ctx, opts, &vendor); err != nil {
		return nil, err
	}
	return &vendor, nil
}
