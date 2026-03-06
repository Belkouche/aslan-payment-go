package aslan

import (
	"context"
	"net/url"
)

// TransactionsResource provides methods for managing transactions.
type TransactionsResource struct {
	client *Client
}

// Retrieve retrieves a transaction by ID.
func (r *TransactionsResource) Retrieve(ctx context.Context, id string) (*Transaction, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/transactions/" + url.PathEscape(id)

	var txn Transaction
	if err := doRequest(ctx, opts, &txn); err != nil {
		return nil, err
	}
	return &txn, nil
}

// List lists transactions with optional filters.
func (r *TransactionsResource) List(ctx context.Context, params *ListTransactionsParams) (*PaginatedResponse[Transaction], error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/transactions"

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
		if params.From != nil {
			query["from"] = *params.From
		}
		if params.To != nil {
			query["to"] = *params.To
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

	var result PaginatedResponse[Transaction]
	if err := doRequest(ctx, opts, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
