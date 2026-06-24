package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// WebhookEndpoint is a registered webhook destination for an organization.
type WebhookEndpoint struct {
	ID        string   `json:"id,omitempty"`
	URL       string   `json:"url,omitempty"`
	Events    []string `json:"events,omitempty"`
	Status    string   `json:"status,omitempty"` // "active" | "disabled"
	Secret    string   `json:"secret,omitempty"` // only returned at create/update time
	CreatedAt string   `json:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
}

// WebhookDelivery is a single event delivery attempt for a webhook endpoint.
type WebhookDelivery struct {
	ID         string `json:"id,omitempty"`
	EndpointID string `json:"endpointId,omitempty"`
	EventType  string `json:"eventType,omitempty"`
	Status     string `json:"status,omitempty"` // "pending" | "succeeded" | "failed"
	Attempts   int    `json:"attempts,omitempty"`
	LastError  string `json:"lastError,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// CreateWebhookEndpointParams is the payload for WebhookService.Create.
type CreateWebhookEndpointParams struct {
	URL    string   `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
}

// UpdateWebhookEndpointParams is the payload for WebhookService.Update.
// Use string pointers to distinguish "leave unchanged" from "set to empty".
type UpdateWebhookEndpointParams struct {
	URL    *string   `json:"url,omitempty"`
	Events *[]string `json:"events,omitempty"`
	Status *string   `json:"status,omitempty"`
}

// RetryDeliveryParams is the optional payload for WebhookService.RetryDelivery.
// Some platforms require an explicit reason for audit purposes.
type RetryDeliveryParams struct {
	Reason string `json:"reason,omitempty"`
}

// WebhookService groups webhook-endpoint endpoints under
// /a/v1/organizations/:orgId/webhooks. It is org-scoped: it returns a
// *ConfigError if the client has no organization ID configured.
//
// This service is distinct from the package-level VerifyWebhook helper, which
// verifies the HMAC signature of an incoming webhook request. WebhookService
// manages the registration, configuration, and delivery lifecycle of endpoints.
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

// Create registers a new webhook endpoint for the configured organization.
// The returned WebhookEndpoint includes the signing secret, which is shown only
// at creation time.
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

// List returns all webhook endpoints registered for the configured organization.
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

// Update modifies a registered webhook endpoint. Pass any subset of fields on
// params; nil fields are left unchanged. A non-nil pointer to an empty string
// clears that field where the API supports it.
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

// Delete permanently removes a webhook endpoint.
func (s *WebhookService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{method: http.MethodDelete, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// Test triggers a synthetic event delivery to the given webhook endpoint,
// useful for verifying endpoint configuration without waiting for a real event.
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

// ListDeliveries returns recent event delivery attempts for the given webhook
// endpoint, newest first.
func (s *WebhookService) ListDeliveries(ctx context.Context, endpointID string, opts ...RequestOption) ([]WebhookDelivery, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []WebhookDelivery
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(endpointID) + "/deliveries", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RetryDelivery requeues a previously failed delivery for a second attempt.
// params is optional and may be nil if the platform does not require a reason.
func (s *WebhookService) RetryDelivery(ctx context.Context, endpointID, deliveryID string, params *RetryDeliveryParams, opts ...RequestOption) (*WebhookDelivery, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out WebhookDelivery
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base + "/" + url.PathEscape(endpointID) + "/deliveries/" + url.PathEscape(deliveryID) + "/retry", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
