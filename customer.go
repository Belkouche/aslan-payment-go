package aslan

import (
	"context"
	"net/url"
)

// CustomersResource provides methods for retrieving customers.
type CustomersResource struct {
	client *Client
}

// List lists customers with optional filters.
func (r *CustomersResource) List(ctx context.Context, params *ListCustomersParams) (*PaginatedResponse[Customer], error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/customers"

	if params != nil {
		query := make(map[string]string)
		if params.Page != nil {
			query["page"] = intToString(params.Page)
		}
		if params.PageSize != nil {
			query["page_size"] = intToString(params.PageSize)
		}
		if params.Search != nil {
			query["search"] = *params.Search
		}
		if params.SortBy != nil {
			query["sort_by"] = *params.SortBy
		}
		if params.SortOrder != nil {
			query["sort_order"] = *params.SortOrder
		}
		if len(query) > 0 {
			opts.query = query
		}
	}

	var result PaginatedResponse[Customer]
	if err := doRequest(ctx, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Retrieve retrieves a customer by ID.
func (r *CustomersResource) Retrieve(ctx context.Context, id string) (*Customer, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/customers/" + url.PathEscape(id)

	var customer Customer
	if err := doRequest(ctx, opts, &customer); err != nil {
		return nil, err
	}
	return &customer, nil
}
