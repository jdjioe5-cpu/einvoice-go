package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// WebhookEndpoint is a registered webhook endpoint for an organization.
type WebhookEndpoint struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Events      []string `json:"events,omitempty"`
	Status      string   `json:"status,omitempty"` // "active" or "disabled"
	Description string   `json:"description,omitempty"`
	// Secret is the signing secret, returned only when the endpoint is created.
	Secret    string `json:"secret,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// WebhookDelivery is a single delivery attempt for a webhook endpoint.
type WebhookDelivery struct {
	ID         string `json:"id"`
	EventType  string `json:"eventType,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Success    bool   `json:"success,omitempty"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// CreateWebhookEndpointParams is the payload for WebhookService.Create.
type CreateWebhookEndpointParams struct {
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description string   `json:"description,omitempty"`
}

// UpdateWebhookEndpointParams is the payload for WebhookService.Update.
type UpdateWebhookEndpointParams struct {
	URL         string   `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// WebhookService manages webhook endpoints under
// /a/v1/organizations/:orgId/webhooks. It is org-scoped and returns a
// *ConfigError when no organization ID is configured.
//
// Note: this manages endpoint registration and deliveries. To verify an
// incoming webhook's signature, use the package-level VerifyWebhook function.
type WebhookService struct {
	http  *httpClient
	orgID func() (string, error)
}

func (s *WebhookService) base() (string, error) {
	orgID, err := s.orgID()
	if err != nil {
		return "", err
	}
	return "/a/v1/organizations/" + url.PathEscape(orgID) + "/webhooks", nil
}

// Create registers a new webhook endpoint. The returned Secret is shown once.
func (s *WebhookService) Create(ctx context.Context, params *CreateWebhookEndpointParams, opts ...RequestOption) (*WebhookEndpoint, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookEndpoint
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base, body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns the organization's webhook endpoints.
func (s *WebhookService) List(ctx context.Context, opts ...RequestOption) ([]WebhookEndpoint, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []WebhookEndpoint
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get retrieves a single webhook endpoint by ID.
func (s *WebhookService) Get(ctx context.Context, id string, opts ...RequestOption) (*WebhookEndpoint, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookEndpoint
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a webhook endpoint.
func (s *WebhookService) Update(ctx context.Context, id string, params *UpdateWebhookEndpointParams, opts ...RequestOption) (*WebhookEndpoint, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookEndpoint
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPatch, path: base + "/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a webhook endpoint.
func (s *WebhookService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{method: http.MethodDelete, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// Test sends a test event to a webhook endpoint.
func (s *WebhookService) Test(ctx context.Context, id string, opts ...RequestOption) (*WebhookDelivery, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookDelivery
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base + "/" + url.PathEscape(id) + "/test", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDeliveries returns recent delivery attempts for a webhook endpoint.
func (s *WebhookService) ListDeliveries(ctx context.Context, id string, params *ListParams, opts ...RequestOption) ([]WebhookDelivery, *Pagination, error) {
	base, err := s.base()
	if err != nil {
		return nil, nil, err
	}
	var out []WebhookDelivery
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id) + "/deliveries", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// RetryDelivery re-attempts a previously failed delivery.
func (s *WebhookService) RetryDelivery(ctx context.Context, id, deliveryID string, opts ...RequestOption) (*WebhookDelivery, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookDelivery
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base + "/" + url.PathEscape(id) + "/deliveries/" + url.PathEscape(deliveryID) + "/retry", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
