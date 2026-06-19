package einvoice

import (
	"errors"
	"testing"
	"time"
)

func TestNewRequiresAPIKey(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error when APIKey is empty")
	}
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T", err)
	}
}

func TestNewDefaults(t *testing.T) {
	c, err := New(Config{APIKey: "sk_test_abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.http.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.http.baseURL, defaultBaseURL)
	}
	if c.http.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", c.http.timeout)
	}
	if c.http.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", c.http.maxRetries)
	}
	if c.Invoices == nil || c.Sellers == nil || c.Buyers == nil || c.Billing == nil {
		t.Error("expected all services to be initialized")
	}
}

func TestNewDisablesRetries(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_abc", MaxRetries: -1})
	if c.http.maxRetries != 0 {
		t.Errorf("maxRetries = %d, want 0 (disabled)", c.http.maxRetries)
	}
}

func TestForOrganization(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_abc"})
	scoped := c.ForOrganization("org_123")
	if scoped.OrganizationID() != "org_123" {
		t.Errorf("OrganizationID() = %q, want org_123", scoped.OrganizationID())
	}
	if c.OrganizationID() != "" {
		t.Error("original client should be unchanged")
	}
}

func TestForAPIKey(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_abc", BaseURL: "https://example.test"})
	other, err := c.ForAPIKey("sk_test_def", "org_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if other.http.apiKey != "sk_test_def" {
		t.Errorf("apiKey = %q, want sk_test_def", other.http.apiKey)
	}
	if other.http.baseURL != "https://example.test" {
		t.Errorf("baseURL not inherited: %q", other.http.baseURL)
	}
	if other.OrganizationID() != "org_9" {
		t.Errorf("OrganizationID() = %q, want org_9", other.OrganizationID())
	}
}
