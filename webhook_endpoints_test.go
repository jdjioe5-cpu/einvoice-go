package einvoice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhooksRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"}) // no organization ID
	_, err := c.Webhooks.List(context.Background())
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
}

func TestWebhookCreateScopedPath(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"whe_1","url":"https://example.com/hook","secret":"whsec_x"}}`))
	}))
	defer srv.Close()

	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_9", MaxRetries: -1})
	ep, err := c.Webhooks.Create(context.Background(), &CreateWebhookEndpointParams{
		URL:    "https://example.com/hook",
		Events: []string{"invoice.approved"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ep.ID != "whe_1" || ep.Secret != "whsec_x" {
		t.Errorf("unexpected endpoint: %+v", ep)
	}
	if gotMethod != http.MethodPost || gotPath != "/a/v1/organizations/org_9/webhooks" {
		t.Errorf("unexpected %s %s", gotMethod, gotPath)
	}
}

func TestWebhookRetryDeliveryPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"del_1","success":true}}`))
	}))
	defer srv.Close()

	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_9", MaxRetries: -1})
	del, err := c.Webhooks.RetryDelivery(context.Background(), "whe_1", "del_1")
	if err != nil {
		t.Fatalf("RetryDelivery: %v", err)
	}
	if !del.Success {
		t.Errorf("expected success delivery, got %+v", del)
	}
	if gotPath != "/a/v1/organizations/org_9/webhooks/whe_1/deliveries/del_1/retry" {
		t.Errorf("path = %q", gotPath)
	}
}
