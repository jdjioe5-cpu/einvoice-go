package einvoice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newOrgTestClient builds a Client with a stub HTTP server AND an OrganizationID,
// which newTestClient in transport_test.go does not set.
func newOrgTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, OrganizationID: "org_77", MaxRetries: -1})
	if err != nil {
		srv.Close()
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func TestWebhooksRequireOrg(t *testing.T) {
	c, _ := New(Config{APIKey: "sk_test_x"}) // no organization ID
	_, err := c.Webhooks.List(context.Background())
	var cfgErr *ConfigError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *ConfigError, got %T (%v)", err, err)
	}
}

func TestWebhooksScopedPath(t *testing.T) {
	var gotMethod, gotPath string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	})
	defer srv.Close()
	if _, err := c.Webhooks.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_77/webhooks" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestWebhooksCreate(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody CreateWebhookEndpointParams
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"wh_1","url":"https://example.com/hook","events":["invoice.accepted"],"status":"active","secret":"shh"}}`))
	})
	defer srv.Close()
	endpoint, err := c.Webhooks.Create(context.Background(), &CreateWebhookEndpointParams{
		URL:    "https://example.com/hook",
		Events: []string{"invoice.accepted"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/a/v1/organizations/org_77/webhooks" {
		t.Errorf("path = %q", gotPath)
	}
	if endpoint.ID != "wh_1" || endpoint.Secret != "shh" {
		t.Errorf("endpoint = %+v", endpoint)
	}
	if len(gotBody.Events) != 1 || gotBody.Events[0] != "invoice.accepted" {
		t.Errorf("body.events = %+v", gotBody.Events)
	}
}

func TestWebhooksGet(t *testing.T) {
	var gotPath string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"wh_42","url":"https://x.test/h","status":"active"}}`))
	})
	defer srv.Close()
	endpoint, err := c.Webhooks.Get(context.Background(), "wh_42")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_77/webhooks/wh_42" {
		t.Errorf("path = %q", gotPath)
	}
	if endpoint.ID != "wh_42" {
		t.Errorf("id = %q", endpoint.ID)
	}
}

func TestWebhooksDelete(t *testing.T) {
	var gotMethod string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"deleted":true}}`))
	})
	defer srv.Close()
	if err := c.Webhooks.Delete(context.Background(), "wh_42"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestWebhooksTest(t *testing.T) {
	var gotPath string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"d_1","endpointId":"wh_1","status":"pending"}}`))
	})
	defer srv.Close()
	delivery, err := c.Webhooks.Test(context.Background(), "wh_1")
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_77/webhooks/wh_1/test" {
		t.Errorf("path = %q", gotPath)
	}
	if delivery.EndpointID != "wh_1" {
		t.Errorf("delivery = %+v", delivery)
	}
}

func TestWebhooksListDeliveries(t *testing.T) {
	var gotPath string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[{"id":"d_1","status":"succeeded","attempts":1},{"id":"d_2","status":"failed","attempts":3,"lastError":"timeout"}]}`))
	})
	defer srv.Close()
	deliveries, err := c.Webhooks.ListDeliveries(context.Background(), "wh_1")
	if err != nil {
		t.Fatalf("ListDeliveries: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_77/webhooks/wh_1/deliveries" {
		t.Errorf("path = %q", gotPath)
	}
	if len(deliveries) != 2 || deliveries[1].LastError != "timeout" {
		t.Errorf("deliveries = %+v", deliveries)
	}
}

func TestWebhooksRetryDelivery(t *testing.T) {
	var gotMethod, gotPath string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"d_2","status":"pending","attempts":2}}`))
	})
	defer srv.Close()
	delivery, err := c.Webhooks.RetryDelivery(context.Background(), "wh_1", "d_2", &RetryDeliveryParams{Reason: "investigating"})
	if err != nil {
		t.Fatalf("RetryDelivery: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/a/v1/organizations/org_77/webhooks/wh_1/deliveries/d_2/retry" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if delivery.Status != "pending" {
		t.Errorf("status = %q", delivery.Status)
	}
}

func TestWebhooksRetryDeliveryNoParams(t *testing.T) {
	// Some platforms accept retry without a body — the helper must still send
	// the request rather than 400 on a nil pointer.
	var gotMethod string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"d_2","status":"pending"}}`))
	})
	defer srv.Close()
	if _, err := c.Webhooks.RetryDelivery(context.Background(), "wh_1", "d_2", nil); err != nil {
		t.Fatalf("RetryDelivery(nil): %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
}

func TestWebhooksUpdate(t *testing.T) {
	var gotPath, gotMethod string
	c, srv := newOrgTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"id":"wh_42","status":"disabled"}}`))
	})
	defer srv.Close()
	status := "disabled"
	endpoint, err := c.Webhooks.Update(context.Background(), "wh_42", &UpdateWebhookEndpointParams{Status: &status})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/a/v1/organizations/org_77/webhooks/wh_42" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if endpoint.Status != "disabled" {
		t.Errorf("status = %q", endpoint.Status)
	}
}

func TestForOrganizationEnablesWebhooks(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":[]}`))
	}))
	defer srv.Close()
	c, _ := New(Config{APIKey: "sk_test_x", BaseURL: srv.URL, MaxRetries: -1})
	scoped := c.ForOrganization("org_55")
	if _, err := scoped.Webhooks.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/a/v1/organizations/org_55/webhooks" {
		t.Errorf("path = %q, want org_55 scope", gotPath)
	}
}
