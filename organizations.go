package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// Organization is a platform organization (tenant). Partners may own child
// organizations for B2B2B use cases.
type Organization struct {
	ID           string         `json:"id"`
	BusinessName string         `json:"businessName,omitempty"`
	Email        string         `json:"email,omitempty"`
	PhoneNumber  string         `json:"phoneNumber,omitempty"`
	TaxNumber    string         `json:"taxNumber,omitempty"`
	Status       string         `json:"status,omitempty"`
	ParentID     string         `json:"parentId,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    string         `json:"createdAt,omitempty"`
	UpdatedAt    string         `json:"updatedAt,omitempty"`
}

// CreateOrganizationParams is the payload for OrganizationService.Create.
type CreateOrganizationParams struct {
	BusinessName string `json:"businessName"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	TaxNumber    string `json:"taxNumber,omitempty"`
}

// UpdateOrganizationParams is the payload for OrganizationService.Update.
type UpdateOrganizationParams struct {
	BusinessName string `json:"businessName,omitempty"`
	Email        string `json:"email,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	TaxNumber    string `json:"taxNumber,omitempty"`
}

// OrganizationService groups organization endpoints under /a/v1/organizations.
type OrganizationService struct {
	http *httpClient
}

// Create creates a new organization (e.g. a child org for a B2B2B partner).
func (s *OrganizationService) Create(ctx context.Context, params *CreateOrganizationParams, opts ...RequestOption) (*Organization, error) {
	var out Organization
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/a/v1/organizations", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves an organization by ID.
func (s *OrganizationService) Get(ctx context.Context, id string, opts ...RequestOption) (*Organization, error) {
	var out Organization
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/a/v1/organizations/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns organizations visible to the caller, with pagination metadata.
func (s *OrganizationService) List(ctx context.Context, params *ListParams, opts ...RequestOption) ([]Organization, *Pagination, error) {
	var out []Organization
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/a/v1/organizations", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// Update modifies an organization.
func (s *OrganizationService) Update(ctx context.Context, id string, params *UpdateOrganizationParams, opts ...RequestOption) (*Organization, error) {
	var out Organization
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPatch, path: "/a/v1/organizations/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListChildren returns the child organizations of the given organization.
func (s *OrganizationService) ListChildren(ctx context.Context, id string, opts ...RequestOption) ([]Organization, error) {
	var out []Organization
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/a/v1/organizations/" + url.PathEscape(id) + "/children", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
