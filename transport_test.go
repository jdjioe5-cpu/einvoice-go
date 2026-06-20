package einvoice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := New(Config{APIKey: "sk_test_key", BaseURL: srv.URL, MaxRetries: -1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, srv
}

func TestInvoiceGetSuccess(t *testing.T) {
	var gotAuth, gotPath string
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"meta":{"statusCode":200,"success":true},"data":{"id":"inv_1","invoiceNumber":"INV-1","status":"accepted"}}`))
	})
	defer srv.Close()

	inv, err := c.Invoices.Get(context.Background(), "inv_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if inv.ID != "inv_1" || inv.InvoiceNumber != "INV-1" {
		t.Errorf("unexpected invoice: %+v", inv)
	}
	if inv.Status != InvoiceStatusAccepted {
		t.Errorf("status = %q, want accepted", inv.Status)
	}
	if gotAuth != "Bearer sk_test_key" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotPath != "/i/v1/invoices/inv_1" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestInvoiceListPagination(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "accepted" {
			t.Errorf("missing status query: %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true,"pagination":{"total":2,"page":1,"pageSize":20,"totalPages":1}},"data":[{"id":"inv_1"},{"id":"inv_2"}]}`))
	})
	defer srv.Close()

	items, page, err := c.Invoices.List(context.Background(), &ListInvoicesParams{Status: InvoiceStatusAccepted})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if page == nil || page.Total != 2 {
		t.Errorf("pagination = %+v", page)
	}
}

func TestAPIError(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"meta":{"statusCode":404,"success":false,"message":"Invoice not found","errors":[{"code":"INVOICE_NOT_FOUND","message":"not found"}]}}`))
	})
	defer srv.Close()

	_, err := c.Invoices.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Code != "INVOICE_NOT_FOUND" {
		t.Errorf("Code = %q", apiErr.Code)
	}
}

func TestRateLimitError(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"meta":{"message":"slow down"}}`))
	})
	defer srv.Close()

	_, err := c.Invoices.Get(context.Background(), "inv_1")
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("expected *RateLimitError, got %T (%v)", err, err)
	}
	if rl.RetryAfter != 7 {
		t.Errorf("RetryAfter = %d, want 7", rl.RetryAfter)
	}
	// Embedded APIError should also be reachable.
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Error("expected RateLimitError to unwrap to *APIError")
	}
}

func TestDeleteNoBody(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		_, _ = w.Write([]byte(`{"meta":{"success":true},"data":{"deleted":true}}`))
	})
	defer srv.Close()

	if err := c.Invoices.Delete(context.Background(), "inv_1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
