package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// ApiKey is an organization API key (the secret value is only returned at
// creation time, on ApiKeyWithKey).
type ApiKey struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Mode      string `json:"mode,omitempty"` // "test" or "production"
	Prefix    string `json:"prefix,omitempty"`
	Status    string `json:"status,omitempty"`
	LastUsed  string `json:"lastUsedAt,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// ApiKeyWithKey is an API key including its secret value (view-once, returned
// only by Create and Rotate).
type ApiKeyWithKey struct {
	ApiKey
	Key string `json:"key"`
}

// CreateApiKeyParams is the payload for ApiKeyService.Create.
type CreateApiKeyParams struct {
	Name string `json:"name,omitempty"`
	// Mode is "test" or "production".
	Mode string `json:"mode,omitempty"`
}

// ApiKeyService groups API key endpoints under
// /a/v1/organizations/:orgId/api-keys. It is org-scoped: it returns a
// *ConfigError if the client has no organization ID configured.
type ApiKeyService struct {
	http  *httpClient
	orgID func() (string, error)
}

func (s *ApiKeyService) base() (string, error) {
	orgID, err := s.orgID()
	if err != nil {
		return "", err
	}
	return "/a/v1/organizations/" + url.PathEscape(orgID) + "/api-keys", nil
}

// Create provisions a new API key for the configured organization. The returned
// key value is shown only once.
func (s *ApiKeyService) Create(ctx context.Context, params *CreateApiKeyParams, opts ...RequestOption) (*ApiKeyWithKey, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out ApiKeyWithKey
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base, body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns the organization's API keys.
func (s *ApiKeyService) List(ctx context.Context, opts ...RequestOption) ([]ApiKey, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []ApiKey
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get retrieves a single API key by ID.
func (s *ApiKeyService) Get(ctx context.Context, id string, opts ...RequestOption) (*ApiKey, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out ApiKey
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Revoke permanently disables an API key.
func (s *ApiKeyService) Revoke(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{method: http.MethodDelete, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// Rotate revokes an API key and issues a replacement, returning the new secret.
func (s *ApiKeyService) Rotate(ctx context.Context, id string, opts ...RequestOption) (*ApiKeyWithKey, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out ApiKeyWithKey
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base + "/" + url.PathEscape(id) + "/rotate", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
