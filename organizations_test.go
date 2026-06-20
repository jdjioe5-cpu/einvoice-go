package einvoice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrganizationCreate(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/a/v1/organizations" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"org_1","businessName":"Child Corp"}}`))
	})
	defer srv.Close()

	org, err := c.Organizations.Create(context.Background(), &CreateOrganizationParams{BusinessName: "Child Corp", Email: "a@b.com"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if org.ID != "org_1" {
		t.Errorf("id = %q, want org_1", org.ID)
	}
}

func TestAPIKeysRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"}) // no organization ID
	_, err := c.APIKeys.List(context.Background())
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
}

func TestAPIKeysScopedPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	}))
	defer srv.Close()

	c, err := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_9", MaxRetries: -1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := c.APIKeys.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_9/api-keys" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestForOrganizationEnablesAPIKeys(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	}))
	defer srv.Close()

	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, MaxRetries: -1})
	scoped := c.ForOrganization("org_42")
	if _, err := scoped.APIKeys.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_42/api-keys" {
		t.Errorf("path = %q, want org_42 scope", gotPath)
	}
}
