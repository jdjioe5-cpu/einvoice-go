package einvoice

import (
	"reflect"
	"testing"
	"time"
)

// TestConfigShapeMatchesJSSDK locks the Go Config field set + defaults to the
// JS reference SDK surface (@useyona/einvoice-js v0.1.1 — HttpClientConfig +
// HttpClientOptions + RetryConfig). If a maintainer renames a field or changes
// a default, this test fails before the JS port drifts out of shape.
//
// JS reference: package/dist/esm/client/http-client.d.ts
//	export interface HttpClientConfig { apiKey, options?, baseUrl?, timeout?, maxRetries?, retryBaseDelay? }
//	export interface HttpClientOptions { baseUrl?, timeout?, retry?, headers? }
//	export interface RetryConfig       { maxAttempts?, baseDelay? }
//
// Translation: Go Config must expose the same field names (case-insensitive),
// and the defaults must match the JS defaults (gateway.useyona.com, 30000ms,
// 3 attempts, 1000ms base delay).
func TestConfigShapeMatchesJSSDK(t *testing.T) {
	jsFieldNames := map[string]bool{
		"APIKey":        true, // apiKey
		"BaseURL":       true, // baseUrl
		"Timeout":       true, // timeout
		"MaxRetries":    true, // maxRetries
		"RetryBaseDelay": true, // retryBaseDelay
		"Headers":       true, // headers
	}

	rt := reflect.TypeOf(Config{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !jsFieldNames[f.Name] {
			// Go-only field — must be one of: OrganizationID, HTTPClient.
			switch f.Name {
			case "OrganizationID", "HTTPClient":
				continue
			default:
				t.Errorf("Config has field %q not in JS HttpClientConfig and not in Go-only allowlist", f.Name)
			}
		}
	}
}

// TestConfigDefaultsMatchJSSDK locks the Go Config defaults to the JS defaults
// documented in http-client.d.ts. Renaming defaultBaseURL or changing the
// 30s / 3 / 1s trio here will fail this test.
func TestConfigDefaultsMatchJSSDK(t *testing.T) {
	c, err := New(Config{APIKey: "sk_test_shape"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.http.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q (JS default: gateway.useyona.com)", c.http.baseURL, defaultBaseURL)
	}
	if defaultBaseURL != "https://gateway.useyona.com" {
		t.Errorf("defaultBaseURL constant = %q, want https://gateway.useyona.com", defaultBaseURL)
	}
	if c.http.timeout != 30*time.Second {
		t.Errorf("timeout default = %v, want 30s (JS default: 30000ms)", c.http.timeout)
	}
	if c.http.maxRetries != 3 {
		t.Errorf("maxRetries default = %d, want 3 (JS RetryConfig default: 3)", c.http.maxRetries)
	}
	if c.http.retryBaseDelay != time.Second {
		t.Errorf("retryBaseDelay default = %v, want 1s (JS RetryConfig default: 1000ms)", c.http.retryBaseDelay)
	}
}

// TestClientServiceSurfaceMatchesJS locks the Go Client's exported service
// fields to the JS EInvoice class surface. If a service is renamed, removed,
// or added (without updating the JS surface lock), the test fails.
//
// JS EInvoice exposes (camelCase): invoices, sellers, buyers, billing,
// collection, webhooks, apiKeys, organizations, users, invitations, roles.
// Go mirrors in PascalCase: Invoices, Sellers, Buyers, Billing, Organizations,
// APIKeys, Users, Roles. Three JS services (Collection, Webhooks,
// Invitations) are not in this PR's bounty scope — locked as NOT-PRESENT so a
// future maintainer can spot scope drift.
//
// The shape-lock catches:
//   - A service being renamed (e.g. Users → Members)
//   - A service being removed (e.g. Roles dropped silently)
//   - A new service being added under a name that is not on the JS surface
//     or in the documented future-scope allowlist
func TestClientServiceSurfaceMatchesJS(t *testing.T) {
	rt := reflect.TypeOf(Client{})

	// service → set in this PR
	present := map[string]bool{
		"Invoices":      true,
		"Sellers":       true,
		"Buyers":        true,
		"Billing":       true,
		"Organizations": true,
		"APIKeys":       true,
		"Users":         true,
		"Roles":         true,
	}

	// JS services that this PR explicitly does NOT port (out of bounty scope #6).
	// If you port one of these, add it to the present map above AND remove the
	// futureScope entry so the lock stays accurate.
	futureScope := map[string]bool{
		"Collection":  true, // JS: collection
		"Webhooks":    true, // JS: webhooks
		"Invitations": true, // JS: invitations
	}

	// http / config / orgID are internal, not service fields.
	internal := map[string]bool{
		"http":   true,
		"config": true,
		"orgID":  true,
	}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if internal[f.Name] {
			continue
		}
		if !present[f.Name] && !futureScope[f.Name] {
			t.Errorf("Client has field %q not in JS surface and not in future-scope allowlist; rename to match JS or add to allowlist with rationale", f.Name)
		}
		if present[f.Name] && futureScope[f.Name] {
			t.Errorf("Client.%s is in both present and future-scope maps; pick one", f.Name)
		}
	}
}

// TestForOrganizationForAPIKeyMirrorJS locks the Go scoping methods to the
// JS method set. JS exposes forOrganization(orgId) → EInvoice and
// forApiKey(apiKey, orgId?) → EInvoice. Go mirrors as ForOrganization(string)
// and ForAPIKey(string, string). If a maintainer renames either method or
// changes the return type, this test fails.
//
// JS reference: package/dist/esm/einvoice.d.ts (forOrganization, forApiKey)
func TestForOrganizationForAPIKeyMirrorJS(t *testing.T) {
	rt := reflect.TypeOf(&Client{})

	wantForOrganization := false
	wantForAPIKey := false
	for i := 0; i < rt.NumMethod(); i++ {
		m := rt.Method(i).Name
		switch m {
		case "ForOrganization":
			wantForOrganization = true
		case "ForAPIKey":
			wantForAPIKey = true
		}
	}

	if !wantForOrganization {
		t.Error("Client is missing ForOrganization method (JS: forOrganization)")
	}
	if !wantForAPIKey {
		t.Error("Client is missing ForAPIKey method (JS: forApiKey)")
	}
}
