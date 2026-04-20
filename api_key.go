package aslan

import (
	"context"
	"net/url"
)

// ApiKeysResource provides methods for managing API keys.
type ApiKeysResource struct {
	client *Client
}

// Create creates a new API key.
//
// The secret key value is only returned in this response and cannot be retrieved again.
func (r *ApiKeysResource) Create(ctx context.Context, params *CreateApiKeyParams) (*ApiKey, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/api-keys"
	opts.body = params

	var apiKey ApiKey
	if err := doRequest(ctx, opts, &apiKey); err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// List lists all API keys for the authenticated merchant.
func (r *ApiKeysResource) List(ctx context.Context) ([]ApiKey, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/api-keys"

	var apiKeys []ApiKey
	if err := doRequest(ctx, opts, &apiKeys); err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// Delete deletes an API key by ID.
func (r *ApiKeysResource) Delete(ctx context.Context, id string) error {
	opts := r.client.requestOpts()
	opts.method = "DELETE"
	opts.path = "/api/v1/api-keys/" + url.PathEscape(id)

	return doRequest(ctx, opts, nil)
}
