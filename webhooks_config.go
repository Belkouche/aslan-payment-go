package aslan

import (
	"context"
)

// WebhooksConfigResource provides methods for managing webhook delivery configuration.
type WebhooksConfigResource struct {
	client *Client
}

// Update sets (or replaces) the webhook configuration for the authenticated merchant.
func (r *WebhooksConfigResource) Update(ctx context.Context, params *UpdateWebhookConfigParams) (*WebhookConfig, error) {
	opts := r.client.requestOpts()
	opts.method = "POST"
	opts.path = "/api/v1/webhooks/config"
	opts.body = params

	var config WebhookConfig
	if err := doRequest(ctx, opts, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Retrieve retrieves the current webhook configuration for the authenticated merchant.
func (r *WebhooksConfigResource) Retrieve(ctx context.Context) (*WebhookConfig, error) {
	opts := r.client.requestOpts()
	opts.method = "GET"
	opts.path = "/api/v1/webhooks/config"

	var config WebhookConfig
	if err := doRequest(ctx, opts, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
