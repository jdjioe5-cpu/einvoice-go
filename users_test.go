package einvoice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// newTestUsersClient returns a Client pointed at a test server that only
// invokes handler for /users routes. Other services are not wired.
func newTestUsersClient(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(Config{
		APIKey:         "sk_test_x",
		BaseURL:        srv.URL,
		OrganizationID: "org_test",
		MaxRetries:     -1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func TestUsersRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"})
	_, err := c.Users.List(context.Background(), nil)
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
	if err := c.Users.Remove(context.Background(), "u_1"); err == nil {
		t.Fatal("expected error on Remove without org ID")
	} else if !errors.As(err, &cfgErr) {
		t.Fatalf("Remove: expected *ConfigError, got %T", err)
	}
}

func TestUsersScopedPath(t *testing.T) {
	var gotPath, gotMethod string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	})
	if _, err := c.Users.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/users" {
		t.Errorf("path = %q, want /a/v1/organizations/org_test/users", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
}

func TestUsersListWithFilters(t *testing.T) {
	var gotQuery map[string]string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = map[string]string{}
		for k := range r.URL.Query() {
			gotQuery[k] = r.URL.Query().Get(k)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[{"id":"u_1","email":"a@b.com","firstName":"A","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":0,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}]}`))
	})
	page := 2
	pageSize := 25
	status := UserStatusActive
	users, err := c.Users.List(context.Background(), &ListUsersParams{
		Page:     &page,
		PageSize: &pageSize,
		Search:   "ada",
		Status:   &status,
		Role:     "role_admin",
		SortBy:   "createdAt",
		SortOrder: "desc",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len(users) = %d, want 1", len(users))
	}
	if users[0].ID != "u_1" {
		t.Errorf("id = %q", users[0].ID)
	}
	if users[0].Status != UserStatusActive {
		t.Errorf("status = %q", users[0].Status)
	}
	want := map[string]string{
		"page": "2", "pageSize": "25", "search": "ada",
		"status": "active", "role": "role_admin",
		"sortBy": "createdAt", "sortOrder": "desc",
	}
	if !reflect.DeepEqual(gotQuery, want) {
		t.Errorf("query = %v, want %v", gotQuery, want)
	}
}

func TestUsersListNoQueryWhenParamsNil(t *testing.T) {
	var gotRaw string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRaw = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	})
	if _, err := c.Users.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotRaw != "" {
		t.Errorf("RawQuery = %q, want empty", gotRaw)
	}
}

func TestUsersGet(t *testing.T) {
	var gotPath string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"ada@x.com","firstName":"Ada","lastName":"Lovelace","phoneNumber":"+2348012345678","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":true,"twoFactorEnabled":true,"twoFactorMethod":"totp","customPermissions":["invoice.read"],"preferences":{"language":"en","timezone":"Africa/Lagos","notifications":{"email":true,"push":false}},"loginCount":7,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	u, err := c.Users.Get(context.Background(), "u_42")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42" {
		t.Errorf("path = %q", gotPath)
	}
	if u.Email != "ada@x.com" || u.FirstName != "Ada" || u.LastName != "Lovelace" {
		t.Errorf("unexpected name/email: %+v", u)
	}
	if u.PhoneNumber != "+2348012345678" {
		t.Errorf("phone = %q", u.PhoneNumber)
	}
	if !u.TwoFactorEnabled || u.TwoFactorMethod != "totp" {
		t.Errorf("2FA: %+v", u)
	}
	if u.Preferences == nil || u.Preferences.Notifications == nil {
		t.Fatalf("preferences not parsed: %+v", u.Preferences)
	}
	if u.Preferences.Notifications.Email == nil || !*u.Preferences.Notifications.Email {
		t.Errorf("notifications.email = %v", u.Preferences.Notifications.Email)
	}
	if u.Preferences.Notifications.Push == nil || *u.Preferences.Notifications.Push {
		t.Errorf("notifications.push = %v", u.Preferences.Notifications.Push)
	}
}

func TestUsersGetMe(t *testing.T) {
	var gotPath string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_me","email":"me@x.com","firstName":"Me","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":1,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	u, err := c.Users.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/users/me" {
		t.Errorf("path = %q, want /me", gotPath)
	}
	if u.ID != "u_me" {
		t.Errorf("id = %q", u.ID)
	}
}

func TestUsersUpdate(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"ada@x.com","firstName":"Augusta","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":7,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	first := "Augusta"
	bio := "Mathematician"
	u, err := c.Users.Update(context.Background(), "u_42", &UpdateUserParams{
		FirstName: &first,
		Bio:       &bio,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["firstName"] != "Augusta" {
		t.Errorf("body.firstName = %v", gotBody["firstName"])
	}
	if gotBody["bio"] != "Mathematician" {
		t.Errorf("body.bio = %v", gotBody["bio"])
	}
	if _, ok := gotBody["lastName"]; ok {
		t.Errorf("unset lastName should not be in body, got %v", gotBody["lastName"])
	}
	if u.FirstName != "Augusta" {
		t.Errorf("returned firstName = %q", u.FirstName)
	}
}

func TestUsersUpdateRole(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"updated":true}}`))
	})
	if err := c.Users.UpdateRole(context.Background(), "u_42", &UpdateUserRoleParams{RoleID: "role_admin"}); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42/role" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["roleId"] != "role_admin" {
		t.Errorf("body.roleId = %v", gotBody["roleId"])
	}
}

func TestUsersUpdateStatus(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"a@b.com","firstName":"A","organizationId":"org_test","status":"suspended","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":0,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	u, err := c.Users.UpdateStatus(context.Background(), "u_42", UserStatusSuspended)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42/status" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["status"] != "suspended" {
		t.Errorf("body.status = %v", gotBody["status"])
	}
	if u.Status != UserStatusSuspended {
		t.Errorf("returned status = %q", u.Status)
	}
}

func TestUsersGetPermissions(t *testing.T) {
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/a/v1/organizations/org_test/users/u_42/permissions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"roles":[{"id":"r1","name":"admin","permissions":["invoice.read","invoice.write"]},{"id":"r2","name":"finance","permissions":["billing.read"]}],"customPermissions":["report.export"],"effectivePermissions":["invoice.read","invoice.write","billing.read","report.export"]}}`))
	})
	p, err := c.Users.GetPermissions(context.Background(), "u_42")
	if err != nil {
		t.Fatalf("GetPermissions: %v", err)
	}
	if len(p.Roles) != 2 {
		t.Fatalf("roles = %d, want 2", len(p.Roles))
	}
	if p.Roles[0].Name != "admin" {
		t.Errorf("roles[0].name = %q", p.Roles[0].Name)
	}
	if len(p.EffectivePermissions) != 4 {
		t.Errorf("effective = %d, want 4", len(p.EffectivePermissions))
	}
}

func TestUsersAddPermissions(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"a@b.com","firstName":"A","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"customPermissions":["report.export"],"loginCount":0,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	_, err := c.Users.AddPermissions(context.Background(), "u_42", []string{"report.export", "report.share"})
	if err != nil {
		t.Fatalf("AddPermissions: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42/permissions" {
		t.Errorf("path = %q", gotPath)
	}
	perms, _ := gotBody["permissions"].([]any)
	if len(perms) != 2 {
		t.Fatalf("body.permissions = %v", gotBody["permissions"])
	}
}

func TestUsersRemovePermissions(t *testing.T) {
	var gotMethod string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"a@b.com","firstName":"A","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":0,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	if _, err := c.Users.RemovePermissions(context.Background(), "u_42", []string{"report.export"}); err != nil {
		t.Fatalf("RemovePermissions: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestUsersSendPhoneVerification(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"sent":true}}`))
	})
	if err := c.Users.SendPhoneVerification(context.Background(), "me", "+2348012345678"); err != nil {
		t.Fatalf("SendPhoneVerification: %v", err)
	}
	if gotBody["phoneNumber"] != "+2348012345678" {
		t.Errorf("body.phoneNumber = %v", gotBody["phoneNumber"])
	}
}

func TestUsersVerifyPhone(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"u_42","email":"a@b.com","firstName":"A","organizationId":"org_test","status":"active","phoneVerified":true,"emailVerified":true,"twoFactorEnabled":false,"loginCount":0,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`))
	})
	u, err := c.Users.VerifyPhone(context.Background(), "u_42", "+2348012345678", "123456")
	if err != nil {
		t.Fatalf("VerifyPhone: %v", err)
	}
	if gotBody["otp"] != "123456" {
		t.Errorf("body.otp = %v", gotBody["otp"])
	}
	if !u.PhoneVerified {
		t.Errorf("phoneVerified = false, want true")
	}
}

func TestUsersRemove(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Users.Remove(context.Background(), "u_42"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_test/users/u_42" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestUsersRemovePropagatesAPIError(t *testing.T) {
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"meta":{"success":false,"message":"user not found","errors":[{"code":"USER_NOT_FOUND","message":"no such user"}]}}`))
	})
	err := c.Users.Remove(context.Background(), "u_missing")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestUsersEmptyBodyListOK(t *testing.T) {
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Some endpoints return {} on 2xx with no data. Ensure List handles it.
		_, _ = w.Write([]byte(`{"meta":{"success":true}}`))
	})
	users, err := c.Users.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("len = %d, want 0", len(users))
	}
}

func TestUsersUpdateRetainsUnsetFieldsInResponse(t *testing.T) {
	// Regression: PATCH must round-trip only the server-confirmed fields.
	c, _ := newTestUsersClient(t, func(w http.ResponseWriter, r *http.Request) {
		body := bytes.NewBufferString(`{"meta":{"success":true},"data":{"id":"u_42","email":"ada@x.com","firstName":"Augusta","lastName":"Lovelace","organizationId":"org_test","status":"active","emailVerified":true,"phoneVerified":false,"twoFactorEnabled":false,"loginCount":7,"createdAt":"2026-06-23T00:00:00Z","updatedAt":"2026-06-23T00:00:00Z"}}`)
		_, _ = io.Copy(w, body)
	})
	first := "Augusta"
	u, err := c.Users.Update(context.Background(), "u_42", &UpdateUserParams{FirstName: &first})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if u.LastName != "Lovelace" {
		t.Errorf("lastName = %q, want Lovelace", u.LastName)
	}
}
