package einvoice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// UserStatus is the lifecycle state of an organization member.
type UserStatus string

const (
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusInactive            UserStatus = "inactive"
	UserStatusSuspended           UserStatus = "suspended"
)

// UserPreferences controls the user-facing defaults for a member.
type UserPreferences struct {
	Language string `json:"language,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Theme    string `json:"theme,omitempty"`
	Notifications *UserNotificationPreferences `json:"notifications,omitempty"`
}

// UserNotificationPreferences toggles the per-channel notification opt-ins.
type UserNotificationPreferences struct {
	Email *bool `json:"email,omitempty"`
	Push  *bool `json:"push,omitempty"`
	SMS   *bool `json:"sms,omitempty"`
}

// User represents an organization member. Fields mirror the JS SDK
// (@useyona/einvoice-js) User type; optional fields are pointers or have
// `omitempty` to keep PATCH round-trips lossless.
type User struct {
	ID               string           `json:"id"`
	Email            string           `json:"email"`
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName,omitempty"`
	PhoneNumber      string           `json:"phoneNumber,omitempty"`
	OrganizationID   string           `json:"organizationId"`
	Status           UserStatus       `json:"status"`
	EmailVerified    bool             `json:"emailVerified"`
	PhoneVerified    bool             `json:"phoneVerified"`
	TwoFactorEnabled bool             `json:"twoFactorEnabled"`
	TwoFactorMethod  string           `json:"twoFactorMethod,omitempty"`
	CustomPermissions []string        `json:"customPermissions,omitempty"`
	Preferences      *UserPreferences `json:"preferences,omitempty"`
	LastLoginAt      string           `json:"lastLoginAt,omitempty"`
	LoginCount       int              `json:"loginCount"`
	CreatedAt        string           `json:"createdAt"`
	UpdatedAt        string           `json:"updatedAt"`
}

// UserRoleAssignment is a single role attached to a user, returned by GetPermissions.
type UserRoleAssignment struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}

// UserPermissions is the effective-permissions view returned by GetPermissions.
type UserPermissions struct {
	Roles               []UserRoleAssignment `json:"roles"`
	CustomPermissions   []string             `json:"customPermissions"`
	EffectivePermissions []string            `json:"effectivePermissions"`
}

// ListUsersParams filters and pages the user list.
type ListUsersParams struct {
	Page     *int        `json:"page,omitempty"`
	PageSize *int        `json:"pageSize,omitempty"`
	Search   string      `json:"search,omitempty"`
	Status   *UserStatus `json:"status,omitempty"`
	Role     string      `json:"role,omitempty"`
	SortBy   string      `json:"sortBy,omitempty"`
	SortOrder string     `json:"sortOrder,omitempty"`
}

// UpdateUserParams is the PATCH payload for UserService.Update. All fields
// are pointers so an unset field does not overwrite the server value.
type UpdateUserParams struct {
	FirstName   *string          `json:"firstName,omitempty"`
	LastName    *string          `json:"lastName,omitempty"`
	PhoneNumber *string          `json:"phoneNumber,omitempty"`
	Avatar      *string          `json:"avatar,omitempty"`
	Gender      *string          `json:"gender,omitempty"`
	DateOfBirth *string          `json:"dateOfBirth,omitempty"`
	Bio         *string          `json:"bio,omitempty"`
	Preferences *UserPreferences `json:"preferences,omitempty"`
}

// UpdateUserRoleParams is the PATCH payload for UserService.UpdateRole.
type UpdateUserRoleParams struct {
	RoleID string `json:"roleId"`
}

// UpdateUserStatusParams is the PATCH payload for UserService.UpdateStatus.
type UpdateUserStatusParams struct {
	Status UserStatus `json:"status"`
}

// UserPermissionsPatch is the payload for AddPermissions/RemovePermissions.
type UserPermissionsPatch struct {
	Permissions []string `json:"permissions"`
}

// SendPhoneVerificationParams is the payload for UserService.SendPhoneVerification.
type SendPhoneVerificationParams struct {
	PhoneNumber string `json:"phoneNumber"`
}

// VerifyPhoneOtpParams is the payload for UserService.VerifyPhone.
type VerifyPhoneOtpParams struct {
	PhoneNumber string `json:"phoneNumber"`
	OTP         string `json:"otp"`
}

// UserService manages organization members under
// /a/v1/organizations/:orgId/users. It is org-scoped: it returns a
// *ConfigError if the client has no organization ID configured.
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

func (s *UserService) encodeListQuery(p *ListUsersParams) url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	if p.Page != nil {
		q.Set("page", strconv.Itoa(*p.Page))
	}
	if p.PageSize != nil {
		q.Set("pageSize", strconv.Itoa(*p.PageSize))
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.Status != nil {
		q.Set("status", string(*p.Status))
	}
	if p.Role != "" {
		q.Set("role", p.Role)
	}
	if p.SortBy != "" {
		q.Set("sortBy", p.SortBy)
	}
	if p.SortOrder != "" {
		q.Set("sortOrder", p.SortOrder)
	}
	return q
}

// List returns a paginated list of users in the configured organization.
// If params is nil, no query string is sent and the server default page is
// returned.
func (s *UserService) List(ctx context.Context, params *ListUsersParams, opts ...RequestOption) ([]User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out []User
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodGet,
		path:   base,
		query:  s.encodeListQuery(params),
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
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

// GetMe retrieves the currently authenticated user. It maps to the server's
// /users/me shortcut and avoids needing the caller's user ID.
func (s *UserService) GetMe(ctx context.Context, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{method: http.MethodGet, path: base + "/me", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update modifies a user's profile. Pass id="me" to update the current user.
func (s *UserService) Update(ctx context.Context, id string, params *UpdateUserParams, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
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

// UpdateRole assigns a role to a user. The roleId must reference a role
// accessible to the configured organization.
func (s *UserService) UpdateRole(ctx context.Context, id string, params *UpdateUserRoleParams, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	var out map[string]any
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPatch,
		path:   base + "/" + url.PathEscape(id) + "/role",
		body:   params,
		opts:   applyOptions(opts),
	}, &out)
	return err
}

// UpdateStatus transitions a user to a new lifecycle state. Valid states are
// pending_verification, active, inactive, suspended.
func (s *UserService) UpdateStatus(ctx context.Context, id string, status UserStatus, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPatch,
		path:   base + "/" + url.PathEscape(id) + "/status",
		body:   &UpdateUserStatusParams{Status: status},
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPermissions returns the user's effective permission set, broken down by
// role-derived permissions and any custom permissions granted directly.
func (s *UserService) GetPermissions(ctx context.Context, id string, opts ...RequestOption) (*UserPermissions, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out UserPermissions
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodGet,
		path:   base + "/" + url.PathEscape(id) + "/permissions",
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AddPermissions grants additional custom permissions to a user, in addition
// to whatever their roles provide.
func (s *UserService) AddPermissions(ctx context.Context, id string, permissions []string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPost,
		path:   base + "/" + url.PathEscape(id) + "/permissions",
		body:   &UserPermissionsPatch{Permissions: permissions},
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RemovePermissions revokes previously-granted custom permissions from a
// user. Permissions are matched by exact string.
func (s *UserService) RemovePermissions(ctx context.Context, id string, permissions []string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodDelete,
		path:   base + "/" + url.PathEscape(id) + "/permissions",
		body:   &UserPermissionsPatch{Permissions: permissions},
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SendPhoneVerification starts phone verification for a user. Only the user
// themselves can verify their own phone; pass id="me".
func (s *UserService) SendPhoneVerification(ctx context.Context, id, phoneNumber string, opts ...RequestOption) error {
	base, err := s.base()
	if err != nil {
		return err
	}
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPost,
		path:   base + "/" + url.PathEscape(id) + "/send-phone-verification",
		body:   &SendPhoneVerificationParams{PhoneNumber: phoneNumber},
		opts:   applyOptions(opts),
	}, nil)
	return err
}

// VerifyPhone completes phone verification by submitting the 6-digit OTP the
// user received.
func (s *UserService) VerifyPhone(ctx context.Context, id, phoneNumber, otp string, opts ...RequestOption) (*User, error) {
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	var out User
	_, err = s.http.do(ctx, internalRequest{
		method: http.MethodPost,
		path:   base + "/" + url.PathEscape(id) + "/verify-phone",
		body:   &VerifyPhoneOtpParams{PhoneNumber: phoneNumber, OTP: otp},
		opts:   applyOptions(opts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Remove permanently deletes a user from the configured organization.
// Returns nil on success and a *ConfigError if the client has no
// organization ID.
func (s *UserService) Remove(ctx context.Context, id string, opts ...RequestOption) error {
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

// package doc for users: re-exported by client.go via Client.Users. No new exported surface here.
var _ = (*UserService)(nil)
