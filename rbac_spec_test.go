package einvoice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRBACServicesSmokeTest exercises every method on the user + role
// services against a permissive mock server. It is a surface-completeness
// check (not a behaviour check) — companion behaviour tests live in
// users_test.go and roles_test.go. If the JS SDK adds a new method, the
// Go SDK must add the corresponding method here at the same time.
func TestRBACServicesSmokeTest(t *testing.T) {
	// The mock server returns a per-route canned response. The point of
	// this test is shape only, not content.
	mux := http.NewServeMux()
	defaultOK := `{"meta":{"success":true},"data":{}}`
	defaultListUsers := `{"meta":{"success":true},"data":[]}`
	defaultListRoles := `{"meta":{"success":true},"data":[]}`
	defaultListAvail := `{"meta":{"success":true},"data":{"permissions":[],"categories":[],"byCategory":{},"totalPermissions":0}}`

	respond := func(w http.ResponseWriter, body string) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}

	mux.HandleFunc("/a/v1/organizations/org_test/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respond(w, defaultListUsers)
			return
		}
		respond(w, defaultOK)
	})
	mux.HandleFunc("/a/v1/organizations/org_test/users/me", func(w http.ResponseWriter, r *http.Request) {
		respond(w, defaultOK)
	})
	mux.HandleFunc("/a/v1/organizations/org_test/users/", func(w http.ResponseWriter, r *http.Request) {
		// Sub-routes under /users/:id/*
		path := strings.TrimPrefix(r.URL.Path, "/a/v1/organizations/org_test/users/")
		switch {
		case path == "":
			respond(w, defaultListUsers)
		case strings.HasSuffix(path, "/role") ||
			strings.HasSuffix(path, "/status") ||
			strings.HasSuffix(path, "/permissions") ||
			strings.HasSuffix(path, "/send-phone-verification") ||
			strings.HasSuffix(path, "/verify-phone"):
			respond(w, defaultOK)
		default:
			respond(w, defaultOK)
		}
	})
	mux.HandleFunc("/a/v1/organizations/org_test/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			respond(w, defaultListRoles)
			return
		}
		respond(w, defaultOK)
	})
	mux.HandleFunc("/a/v1/organizations/org_test/roles/permissions/available", func(w http.ResponseWriter, r *http.Request) {
		respond(w, defaultListAvail)
	})
	mux.HandleFunc("/a/v1/organizations/org_test/roles/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, defaultOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := New(Config{
		APIKey:         "sk_test_x",
		BaseURL:        srv.URL,
		OrganizationID: "org_test",
		MaxRetries:     -1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	// Users service: 12 methods (11 from JS spec + 1 from issue #6).
	check := func(name string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	_, err = c.Users.List(ctx, nil)
	check("Users.List", err)
	_, err = c.Users.Get(ctx, "u")
	check("Users.Get", err)
	_, err = c.Users.GetMe(ctx)
	check("Users.GetMe", err)
	_, err = c.Users.Update(ctx, "u", &UpdateUserParams{})
	check("Users.Update", err)
	err = c.Users.UpdateRole(ctx, "u", &UpdateUserRoleParams{RoleID: "r"})
	check("Users.UpdateRole", err)
	_, err = c.Users.UpdateStatus(ctx, "u", UserStatusActive)
	check("Users.UpdateStatus", err)
	_, err = c.Users.GetPermissions(ctx, "u")
	check("Users.GetPermissions", err)
	_, err = c.Users.AddPermissions(ctx, "u", []string{"x"})
	check("Users.AddPermissions", err)
	_, err = c.Users.RemovePermissions(ctx, "u", []string{"x"})
	check("Users.RemovePermissions", err)
	err = c.Users.SendPhoneVerification(ctx, "u", "+1")
	check("Users.SendPhoneVerification", err)
	_, err = c.Users.VerifyPhone(ctx, "u", "+1", "000000")
	check("Users.VerifyPhone", err)
	err = c.Users.Remove(ctx, "u")
	check("Users.Remove", err)

	// Roles service: 6 methods, matches JS exactly.
	_, err = c.Roles.List(ctx)
	check("Roles.List", err)
	_, err = c.Roles.Get(ctx, "r")
	check("Roles.Get", err)
	_, err = c.Roles.Create(ctx, &CreateRoleParams{})
	check("Roles.Create", err)
	_, err = c.Roles.Update(ctx, "r", &UpdateRoleParams{})
	check("Roles.Update", err)
	err = c.Roles.Delete(ctx, "r")
	check("Roles.Delete", err)
	_, err = c.Roles.AvailablePermissions(ctx)
	check("Roles.AvailablePermissions", err)
}
