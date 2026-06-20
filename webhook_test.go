package einvoice

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const testSecret = "whsec_test_secret"

// sign reproduces the platform's signing scheme over the compacted data field.
func sign(t *testing.T, payload []byte, timestamp, secret string) string {
	t.Helper()
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, event.Data); err != nil {
		t.Fatalf("compact: %v", err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + compact.String()))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func fixedNow(ts int64) func() time.Time {
	return func() time.Time { return time.Unix(ts, 0) }
}

func TestVerifyWebhookValid(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"invoice.approved","data":{"invoiceId":"inv_1","status":"accepted"}}`)
	ts := "1700000000"
	sig := sign(t, payload, ts, testSecret)

	h := http.Header{}
	h.Set("X-Webhook-Signature", sig)
	h.Set("X-Webhook-Timestamp", ts)

	event, err := VerifyWebhook(payload, h, testSecret, &VerifyWebhookOptions{Now: fixedNow(1700000010)})
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Type != "invoice.approved" {
		t.Errorf("type = %q", event.Type)
	}
}

func TestVerifyWebhookBadSignature(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"invoice.approved","data":{"x":1}}`)
	ts := "1700000000"
	h := http.Header{}
	h.Set("X-Webhook-Signature", "sha256=deadbeef")
	h.Set("X-Webhook-Timestamp", ts)

	_, err := VerifyWebhook(payload, h, testSecret, &VerifyWebhookOptions{Now: fixedNow(1700000000)})
	if err == nil {
		t.Fatal("expected signature failure")
	}
	var whErr *WebhookError
	if !asWebhookError(err, &whErr) {
		t.Fatalf("expected *WebhookError, got %T", err)
	}
}

func TestVerifyWebhookStaleTimestamp(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"x","data":{"x":1}}`)
	ts := "1700000000"
	sig := sign(t, payload, ts, testSecret)
	h := http.Header{}
	h.Set("X-Webhook-Signature", sig)
	h.Set("X-Webhook-Timestamp", ts)

	// 10 minutes later, default tolerance 5 minutes -> stale.
	_, err := VerifyWebhook(payload, h, testSecret, &VerifyWebhookOptions{Now: fixedNow(1700000600)})
	if err == nil {
		t.Fatal("expected stale timestamp error")
	}
}

func TestVerifyWebhookMissingHeaders(t *testing.T) {
	_, err := VerifyWebhook([]byte(`{}`), http.Header{}, testSecret, nil)
	if err == nil {
		t.Fatal("expected missing header error")
	}
}

// asWebhookError is a tiny helper mirroring errors.As for *WebhookError.
func asWebhookError(err error, target **WebhookError) bool {
	if e, ok := err.(*WebhookError); ok {
		*target = e
		return true
	}
	return false
}
