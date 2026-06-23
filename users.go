package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// User is a member of an organization.
type User struct {
	ID        string         `json:"id"`
	Email     string         `json:"email,omitempty"`
	FirstName string         `json:"firstName,omitempty"`
	LastName  string         `json:"lastName,omitempty"`
	RoleID    string         `json:"roleId,omitempty"`
	RoleName  string         `json:"roleName,omitempty"`
	Status    string         `json:"status,omitempty"` // active | inactive | suspended
	Phone     string         `json:"phoneNumber,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"createdAt,omitempty"`
	UpdatedAt string         `json:"updatedAt,omitempty"`
}

// UpdateUserParams is the payload for UserService.Update.
type UpdateUserParams struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phoneNumber,omitempty"`
}

// UserService groups user endpoints under
// /a/v1/organizations/:orgId/users. It is org-scoped and returns a
// *ConfigError when no organization ID is configured.
type UserService struct {
	http  *httpClient
	orgID func() (string, error)
}

func (s *UserService) base() (string, error) {
	orgID, err := s.orgID()
	if err != nil {
		return "", err
	}
	return "/a/v1/organizations/" + url.PathEscape(orgID) + "/users", nil
}

// List returns the organization's users with pagination metadata.
func (s *UserService) List(ctx context.Context, params *ListParams, opts ...RequestOption) ([]User, *Pagination, error) {
	base, err := s.base()
	if err != nil {
		return nil, nil, err
	}
	var out []User
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: base, query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// Get retrieves a single user by ID.
func (s *UserService) Get(ctx context.Context, id string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a user's profile.
func (s *UserService) Update(ctx context.Context, id string, params *UpdateUserParams, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPatch, path: base + "/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRole assigns a role to a user.
func (s *UserService) UpdateRole(ctx context.Context, id, roleID string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	body := map[string]any{"roleId": roleID}
	var out User
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPut, path: base + "/" + url.PathEscape(id) + "/role", body: body, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateStatus changes a user's status (e.g. "active", "suspended").
func (s *UserService) UpdateStatus(ctx context.Context, id, status string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	body := map[string]any{"status": status}
	var out User
	_, err = s.http.do(ctx, internalRequest{method: http.MethodPut, path: base + "/" + url.PathEscape(id) + "/status", body: body, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Remove removes a user from the organization.
func (s *UserService) Remove(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{method: http.MethodDelete, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}
