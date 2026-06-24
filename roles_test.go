package einvoice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestRolesClient(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
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

func TestRolesRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"})
	_, err := c.Roles.List(context.Background())
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
	if err := c.Roles.Delete(context.Background(), "r_1"); err == nil {
		t.Fatal("expected error on Delete without org ID")
	}
}

func TestRolesScopedPath(t *testing.T) {
	var gotPath, gotMethod string
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	})
	if _, err := c.Roles.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/roles" {
		t.Errorf("path = %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q", gotMethod)
	}
}

func TestRolesList(t *testing.T) {
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[{"id":"r1","name":"admin","displayName":"Administrator","description":"Full access","type":"system","isImmutable":true,"permissions":["*"]},{"id":"r2","name":"invoice_manager","displayName":"Invoice Manager","description":"Manages invoices","type":"custom","organizationId":"org_test","permissions":["invoice.create","invoice.read"]}]}`))
	})
	roles, err := c.Roles.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("len = %d, want 2", len(roles))
	}
	if roles[0].DisplayName != "Administrator" || !roles[0].IsImmutable || roles[0].Type != RoleTypeSystem {
		t.Errorf("system role not parsed: %+v", roles[0])
	}
	if roles[1].Type != RoleTypeCustom || roles[1].OrganizationID != "org_test" {
		t.Errorf("custom role not parsed: %+v", roles[1])
	}
}

func TestRolesGet(t *testing.T) {
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/a/v1/organizations/org_test/roles/r1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"r1","name":"admin","displayName":"Administrator","description":"Full access","type":"system","isImmutable":true,"permissions":["*"],"assignedTo":{"users":12,"apiKeys":3}}}`))
	})
	rd, err := c.Roles.Get(context.Background(), "r1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rd.Name != "admin" || rd.DisplayName != "Administrator" {
		t.Errorf("name/displayName mismatch: %+v", rd)
	}
	if rd.AssignedTo.Users != 12 || rd.AssignedTo.APIKeys != 3 {
		t.Errorf("assignedTo = %+v, want {12, 3}", rd.AssignedTo)
	}
	if len(rd.Permissions) != 1 || rd.Permissions[0] != "*" {
		t.Errorf("permissions = %v", rd.Permissions)
	}
}

func TestRolesCreate(t *testing.T) {
	var gotMethod string
	var gotBody map[string]any
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"r_new","name":"invoice_manager","displayName":"Invoice Manager","description":"Manages invoices","type":"custom","organizationId":"org_test","permissions":["invoice.create","invoice.read"]}}`))
	})
	r, err := c.Roles.Create(context.Background(), &CreateRoleParams{
		Name:        "invoice_manager",
		DisplayName: "Invoice Manager",
		Description: "Manages invoices",
		Permissions: []string{"invoice.create", "invoice.read"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["name"] != "invoice_manager" || gotBody["displayName"] != "Invoice Manager" {
		t.Errorf("body fields wrong: %v", gotBody)
	}
	if r.ID != "r_new" || r.Type != RoleTypeCustom {
		t.Errorf("returned role = %+v", r)
	}
}

func TestRolesUpdate(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"r1","name":"invoice_manager_v2","displayName":"Invoice Manager v2","description":"Manages invoices","type":"custom","organizationId":"org_test","permissions":["invoice.create","invoice.read","invoice.delete"]}}`))
	})
	dn := "Invoice Manager v2"
	perms := []string{"invoice.create", "invoice.read", "invoice.delete"}
	r, err := c.Roles.Update(context.Background(), "r1", &UpdateRoleParams{
		DisplayName: &dn,
		Permissions: perms,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_test/roles/r1" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["displayName"] != "Invoice Manager v2" {
		t.Errorf("body.displayName = %v", gotBody["displayName"])
	}
	if _, hasName := gotBody["name"]; hasName {
		t.Errorf("body.name should be omitted, got %v", gotBody["name"])
	}
	if r.DisplayName != "Invoice Manager v2" {
		t.Errorf("returned displayName = %q", r.DisplayName)
	}
	if len(r.Permissions) != 3 {
		t.Errorf("returned permissions = %v", r.Permissions)
	}
}

func TestRolesDelete(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Roles.Delete(context.Background(), "r1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_test/roles/r1" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestRolesDeleteSystemRoleRejected(t *testing.T) {
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"meta":{"success":false,"message":"system role cannot be deleted","errors":[{"code":"ROLE_IMMUTABLE","message":"system roles are immutable"}]}}`))
	})
	err := c.Roles.Delete(context.Background(), "r_admin")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.Code != "ROLE_IMMUTABLE" {
		t.Errorf("code = %q, want ROLE_IMMUTABLE", apiErr.Code)
	}
}

func TestRolesAvailablePermissions(t *testing.T) {
	var gotPath string
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"permissions":[{"permission":"invoice.create","description":"Create an invoice","category":"invoices"},{"permission":"invoice.read","description":"Read invoices","category":"invoices"},{"permission":"billing.read","description":"Read billing","category":"billing"}],"categories":["invoices","billing"],"byCategory":{"invoices":[{"permission":"invoice.create","description":"Create an invoice","category":"invoices"},{"permission":"invoice.read","description":"Read invoices","category":"invoices"}],"billing":[{"permission":"billing.read","description":"Read billing","category":"billing"}]},"totalPermissions":3}}`))
	})
	ap, err := c.Roles.AvailablePermissions(context.Background())
	if err != nil {
		t.Fatalf("AvailablePermissions: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_test/roles/permissions/available" {
		t.Errorf("path = %q", gotPath)
	}
	if ap.TotalPermissions != 3 {
		t.Errorf("totalPermissions = %d, want 3", ap.TotalPermissions)
	}
	if len(ap.Permissions) != 3 {
		t.Errorf("len(permissions) = %d, want 3", len(ap.Permissions))
	}
	if len(ap.Categories) != 2 || ap.Categories[0] != "invoices" {
		t.Errorf("categories = %v", ap.Categories)
	}
	if len(ap.ByCategory["invoices"]) != 2 {
		t.Errorf("byCategory[invoices] = %v", ap.ByCategory["invoices"])
	}
	if len(ap.ByCategory["billing"]) != 1 {
		t.Errorf("byCategory[billing] = %v", ap.ByCategory["billing"])
	}
}

func TestRolesForOrganizationScope(t *testing.T) {
	var gotPath string
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	})
	scoped := c.ForOrganization("org_99")
	if _, err := scoped.Roles.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_99/roles" {
		t.Errorf("path = %q, want org_99 scope", gotPath)
	}
}

func TestRolesCreateRequiresFields(t *testing.T) {
	// A request body the server rejects with 400 must surface as *APIError,
	// not a transport error.
	c, _ := newTestRolesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"meta":{"success":false,"message":"name is required","errors":[{"field":"name","code":"REQUIRED","message":"name is required"}]}}`))
	})
	_, err := c.Roles.Create(context.Background(), &CreateRoleParams{DisplayName: "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("status = %d, want 400", apiErr.StatusCode)
	}
}
