package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// Role is a custom RBAC role within an organization.
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	IsSystem    bool     `json:"isSystem,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// CreateRoleParams is the payload for RoleService.Create.
type CreateRoleParams struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleParams is the payload for RoleService.Update.
type UpdateRoleParams struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// PermissionGroup groups available permissions by domain (e.g. "invoices").
type PermissionGroup struct {
	Group       string   `json:"group"`
	Permissions []string `json:"permissions"`
}

// RoleService groups role endpoints under
// /a/v1/organizations/:orgId/roles. It is org-scoped and returns a
// *ConfigError when no organization ID is configured.
type RoleService struct {
	http  *httpClient
	orgID func() (string, error)
}

func (s *RoleService) base() (string, error) {
	orgID, err := s.orgID()
	if err != nil {
		return "", err
	}
	return "/a/v1/organizations/" + url.PathEscape(orgID) + "/roles", nil
}

// List returns the organization's roles.
func (s *RoleService) List(ctx context.Context, opts ...RequestOption) ([]Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []Role
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get retrieves a single role by ID.
func (s *RoleService) Get(ctx context.Context, id string, opts ...RequestOption) (*Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out Role
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Create defines a new custom role.
func (s *RoleService) Create(ctx context.Context, params *CreateRoleParams, opts ...RequestOption) (*Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out Role
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPost, path: base, body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a custom role.
func (s *RoleService) Update(ctx context.Context, id string, params *UpdateRoleParams, opts ...RequestOption) (*Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out Role
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPatch, path: base + "/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a custom role.
func (s *RoleService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{method: http.MethodDelete, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// AvailablePermissions returns the permissions that can be assigned to roles.
func (s *RoleService) AvailablePermissions(ctx context.Context, opts ...RequestOption) ([]PermissionGroup, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []PermissionGroup
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/permissions/available", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
