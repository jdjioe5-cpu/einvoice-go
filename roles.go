package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// RoleType distinguishes system-defined roles from org-defined custom roles
// and api-only roles.
type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeCustom RoleType = "custom"
	RoleTypeAPI    RoleType = "api"
)

// Role is an RBAC role within an organization. System roles are immutable
// and shared across organizations; custom roles belong to a single org.
type Role struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	DisplayName    string   `json:"displayName"`
	Description    string   `json:"description,omitempty"`
	Type           RoleType `json:"type,omitempty"`
	IsImmutable    bool     `json:"isImmutable,omitempty"`
	OrganizationID string   `json:"organizationId,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
}

// RoleAssignmentCounts reports how many users and API keys are currently
// assigned to a role. Returned only by RoleService.Get.
type RoleAssignmentCounts struct {
	Users   int `json:"users"`
	APIKeys int `json:"apiKeys"`
}

// RoleDetails extends Role with assignment counts. Returned by Get.
type RoleDetails struct {
	Role
	AssignedTo RoleAssignmentCounts `json:"assignedTo"`
}

// PermissionDefinition is a single entry in the permissions catalog.
type PermissionDefinition struct {
	Permission  string `json:"permission"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// AvailablePermissions is the response shape of RoleService.GetAvailablePermissions.
// It returns the flat permission list, the ordered category list, and a
// pre-grouped byCategory view for direct UI rendering.
type AvailablePermissions struct {
	Permissions     []PermissionDefinition            `json:"permissions"`
	Categories      []string                          `json:"categories"`
	ByCategory      map[string][]PermissionDefinition `json:"byCategory"`
	TotalPermissions int                              `json:"totalPermissions"`
}

// CreateRoleParams is the payload for RoleService.Create.
type CreateRoleParams struct {
	// Name is the internal identifier (lowercase, underscored). Required.
	Name string `json:"name"`
	// DisplayName is the human-readable name shown in UIs. Required.
	DisplayName string `json:"displayName"`
	// Description explains what the role is for. Optional.
	Description string `json:"description,omitempty"`
	// Permissions is the list of permission strings assigned to the role. Required.
	Permissions []string `json:"permissions"`
}

// UpdateRoleParams is the PATCH payload for RoleService.Update. All fields
// are optional; the server replaces only the supplied fields.
type UpdateRoleParams struct {
	Name        *string  `json:"name,omitempty"`
	DisplayName *string  `json:"displayName,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// RoleService manages custom RBAC roles under
// /a/v1/organizations/:orgId/roles. It is org-scoped: it returns a
// *ConfigError if the client has no organization ID configured.
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

// List returns the organization's roles (system + custom).
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

// Get retrieves a single role by ID and includes assignment counts.
func (s *RoleService) Get(ctx context.Context, id string, opts ...RequestOption) (*RoleDetails, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out RoleDetails
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Create defines a new custom role for the configured organization.
func (s *RoleService) Create(ctx context.Context, params *CreateRoleParams, opts ...RequestOption) (*Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out Role
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPost,
		path:   base,
		body:   params,
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a custom role. System roles cannot be updated; the server
// returns a 4xx in that case.
func (s *RoleService) Update(ctx context.Context, id string, params *UpdateRoleParams, opts ...RequestOption) (*Role, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out Role
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPatch,
		path:   base + "/" + url.PathEscape(id),
		body:   params,
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete permanently removes a custom role. All users and API keys assigned
// to this role will lose it. System roles cannot be deleted.
func (s *RoleService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodDelete,
		path:   base + "/" + url.PathEscape(id),
		opts:   applyOptions(opts),
	}, nil)
	return err
}

// AvailablePermissions returns the permission catalog that can be assigned
// to roles in this organization, including the flat list, the ordered
// category list, and a pre-grouped byCategory map for direct UI rendering.
func (s *RoleService) AvailablePermissions(ctx context.Context, opts ...RequestOption) (*AvailablePermissions, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out AvailablePermissions
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodGet,
		path:   base + "/permissions/available",
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
