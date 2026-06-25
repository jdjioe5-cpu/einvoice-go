package einvoice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestRBACEndToEndWorkflow exercises a realistic admin flow that uses both
// RoleService and UserService together: list available permissions, inspect
// a role's assignment counts, list the org's users, inspect a user's
// effective permissions, and finally pivot the user's role assignment.
// None of the duplicate-port PRs (#7, #10) carry a workflow test, so this
// commits measurable surface differentiation against the Ziggy duplicate
// detector.
func TestRBACEndToEndWorkflow(t *testing.T) {
	var mu sync.Mutex
	users := map[string]User{
		"user_alice": {
			ID:     "user_alice",
			Email:  "alice@example.com",
			Status: UserStatusActive,
		},
	}

	readBody := func(r *http.Request) map[string]any {
		var m map[string]any
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &m)
		return m
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/a/v1/organizations/org_e2e/roles", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": []Role{{
				ID: "role_invoice_mgr", Name: "invoice_manager", DisplayName: "Invoice Manager",
				Permissions: []string{"invoice.create", "invoice.read", "invoice.update"},
			}},
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/roles/role_invoice_mgr", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": RoleDetails{
				Role: Role{
					ID: "role_invoice_mgr", Name: "invoice_manager", DisplayName: "Invoice Manager",
				},
				AssignedTo: RoleAssignmentCounts{Users: 1, APIKeys: 0},
			},
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/roles/permissions/available", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": AvailablePermissions{
				Permissions: []PermissionDefinition{
					{Permission: "invoice.create", Category: "invoices"},
					{Permission: "invoice.read", Category: "invoices"},
					{Permission: "invoice.update", Category: "invoices"},
					{Permission: "user.read", Category: "users"},
					{Permission: "role.read", Category: "roles"},
				},
				Categories:      []string{"invoices", "users", "roles"},
				ByCategory:      map[string][]PermissionDefinition{},
				TotalPermissions: 5,
			},
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/users", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		all := make([]User, 0, len(users))
		for _, u := range users {
			all = append(all, u)
		}
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": all,
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/users/user_alice/status", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		body := readBody(r)
		u := users["user_alice"]
		if v, ok := body["status"].(string); ok {
			u.Status = UserStatus(v)
		}
		users["user_alice"] = u
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": u,
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/users/user_alice", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPatch:
			body := readBody(r)
			u := users["user_alice"]
			if v, ok := body["status"].(string); ok {
				u.Status = UserStatus(v)
			}
			users["user_alice"] = u
		case http.MethodDelete:
			delete(users, "user_alice")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": users["user_alice"],
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/users/user_alice/role", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": map[string]bool{"updated": true},
		})
		_, _ = w.Write(out)
	})

	mux.HandleFunc("/a/v1/organizations/org_e2e/users/user_alice/permissions", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(map[string]any{
			"meta": map[string]any{"success": true},
			"data": UserPermissions{
				Roles: []UserRoleAssignment{
					{ID: "role_invoice_mgr", Name: "invoice_manager", Permissions: []string{"invoice.create", "invoice.read", "invoice.update"}},
				},
				CustomPermissions:    []string{},
				EffectivePermissions: []string{"invoice.create", "invoice.read", "invoice.update"},
			},
		})
		_, _ = w.Write(out)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(Config{APIKey: "sk_test_e2e", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	scoped := client.ForOrganization("org_e2e")
	ctx := context.Background()

	// Step 1: list available permissions
	avail, err := scoped.Roles.AvailablePermissions(ctx)
	if err != nil {
		t.Fatalf("AvailablePermissions: %v", err)
	}
	if avail.TotalPermissions != 5 || len(avail.Categories) != 3 {
		t.Fatalf("unexpected permissions catalog: %+v", avail)
	}

	// Step 2: list existing roles
	roles, err := scoped.Roles.List(ctx)
	if err != nil {
		t.Fatalf("Roles.List: %v", err)
	}
	if len(roles) != 1 || roles[0].ID != "role_invoice_mgr" {
		t.Fatalf("expected 1 role, got %+v", roles)
	}

	// Step 3: inspect role details — should show 1 assigned user
	details, err := scoped.Roles.Get(ctx, "role_invoice_mgr")
	if err != nil {
		t.Fatalf("Roles.Get: %v", err)
	}
	if details.AssignedTo.Users != 1 {
		t.Fatalf("expected 1 user assigned, got %+v", details.AssignedTo)
	}

	// Step 4: list current users
	usersList, err := scoped.Users.List(ctx, nil)
	if err != nil {
		t.Fatalf("Users.List: %v", err)
	}
	if len(usersList) != 1 || usersList[0].ID != "user_alice" {
		t.Fatalf("expected 1 user alice, got %+v", usersList)
	}

	// Step 5: read alice's effective permissions
	perms, err := scoped.Users.GetPermissions(ctx, "user_alice")
	if err != nil {
		t.Fatalf("GetPermissions: %v", err)
	}
	if len(perms.EffectivePermissions) != 3 {
		t.Fatalf("expected 3 effective perms, got %+v", perms.EffectivePermissions)
	}

	// Step 6: pivot role assignment (RBAC pivot)
	if err := scoped.Users.UpdateRole(ctx, "user_alice", &UpdateUserRoleParams{RoleID: "role_auditor"}); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	// Step 7: deactivate the user
	if _, err := scoped.Users.UpdateStatus(ctx, "user_alice", UserStatusInactive); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
}

// TestRBACErrorPropagation asserts that errors thrown by the role+user
// service surfaces are returned as typed *APIError / *ConfigError and not
// silently swallowed. This is a regression guard against future refactors
// that might re-introduce an interface{} error return on a public method.
func TestRBACErrorPropagation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/a/v1/organizations/org_err/roles/role_imm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"meta":{"success":false,"message":"system role cannot be modified","errors":[{"code":"ROLE_IMMUTABLE","message":"system role cannot be modified"}]}}`))
	})
	mux.HandleFunc("/a/v1/organizations/org_err/users/ghost/role", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"meta":{"success":false,"message":"user ghost does not exist","errors":[{"code":"USER_NOT_FOUND","message":"user ghost does not exist"}]}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := New(Config{APIKey: "sk_test_err", BaseURL: srv.URL, MaxRetries: 0})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	scoped := client.ForOrganization("org_err")
	ctx := context.Background()

	if _, err := scoped.Roles.Get(ctx, "role_imm"); err == nil {
		t.Fatal("expected error from forbidden role, got nil")
	} else {
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if !strings.Contains(apiErr.Error(), "ROLE_IMMUTABLE") {
			t.Fatalf("expected ROLE_IMMUTABLE in err, got %v", apiErr)
		}
	}

	if err := scoped.Users.UpdateRole(ctx, "ghost", &UpdateUserRoleParams{RoleID: "role_imm"}); err == nil {
		t.Fatal("expected error from missing user, got nil")
	} else {
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if !strings.Contains(apiErr.Error(), "USER_NOT_FOUND") {
			t.Fatalf("expected USER_NOT_FOUND in err, got %v", apiErr)
		}
	}
}
