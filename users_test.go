package einvoice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUsersRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"}) // no organization ID
	_, _, err := c.Users.List(context.Background(), nil)
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
}

func TestRolesRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"}) // no organization ID
	_, err := c.Roles.List(context.Background())
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
}

func TestUsersScopedPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[{"id":"usr_1","email":"a@b.com"}]}`))
	}))
	defer srv.Close()

	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_9", MaxRetries: -1})
	users, _, err := c.Users.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(users) != 1 || users[0].ID != "usr_1" {
		t.Errorf("unexpected users: %+v", users)
	}
	if gotPath != "/a/v1/organizations/org_9/users" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestRolesCreateScopedPath(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"role_1","name":"Accountant"}}`))
	}))
	defer srv.Close()

	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_9", MaxRetries: -1})
	role, err := c.Roles.Create(context.Background(), &CreateRoleParams{Name: "Accountant", Permissions: []string{"invoices.read"}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if role.ID != "role_1" {
		t.Errorf("id = %q", role.ID)
	}
	if gotMethod != http.MethodPost || gotPath != "/a/v1/organizations/org_9/roles" {
		t.Errorf("unexpected %s %s", gotMethod, gotPath)
	}
}
