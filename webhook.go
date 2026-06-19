package einvoice

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// defaultWebhookToleranceSeconds is the maximum age of a webhook timestamp
// before it is rejected as a potential replay (5 minutes).
const defaultWebhookToleranceSeconds = 300

// WebhookEvent is a verified, parsed webhook event from the platform.
type WebhookEvent struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt string          `json:"createdAt,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// VerifyWebhookOptions configures VerifyWebhook.
type VerifyWebhookOptions struct {
	// ToleranceSeconds overrides the default replay-protection window (300s).
	ToleranceSeconds int
	// Now overrides the current time; primarily for testing.
	Now func() time.Time
}

// VerifyWebhook verifies a webhook's HMAC-SHA256 signature (timing-safe) and
// timestamp (replay protection), then parses and returns the event.
//
// The platform signs only the event's data field, as "<timestamp>.<data-json>".
// payload is the raw request body; headers carries the X-Webhook-Signature and
// X-Webhook-Timestamp headers; secret is your webhook signing secret.
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		body, _ := io.ReadAll(r.Body)
//		event, err := einvoice.VerifyWebhook(body, r.Header, os.Getenv("WEBHOOK_SECRET"), nil)
//		if err != nil {
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		switch event.Type {
//		case "invoice.approved":
//			// handle
//		}
//		w.WriteHeader(http.StatusOK)
//	}
func VerifyWebhook(payload []byte, headers http.Header, secret string, options *VerifyWebhookOptions) (*WebhookEvent, error) {
	tolerance := defaultWebhookToleranceSeconds
	nowFn := time.Now
	if options != nil {
		if options.ToleranceSeconds > 0 {
			tolerance = options.ToleranceSeconds
		}
		if options.Now != nil {
			nowFn = options.Now
		}
	}

	signature := headers.Get("X-Webhook-Signature")
	timestamp := headers.Get("X-Webhook-Timestamp")
	if signature == "" {
		return nil, &WebhookError{Message: "missing X-Webhook-Signature header"}
	}
	if timestamp == "" {
		return nil, &WebhookError{Message: "missing X-Webhook-Timestamp header"}
	}

	webhookTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, &WebhookError{Message: "invalid X-Webhook-Timestamp header"}
	}

	now := nowFn().Unix()
	if int(math.Abs(float64(now-webhookTime))) > tolerance {
		return nil, &WebhookError{Message: fmt.Sprintf(
			"webhook timestamp too old (received %d, current %d, tolerance %ds)",
			webhookTime, now, tolerance)}
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, &WebhookError{Message: "failed to parse webhook payload as JSON"}
	}

	// The backend signs "<timestamp>.<compact-json-of-data>". Compact the raw
	// data bytes (whitespace removed, key order preserved) to match the
	// reference implementation's JSON.stringify(data).
	var compactData bytes.Buffer
	if len(event.Data) == 0 {
		compactData.WriteString("undefined")
	} else if err := json.Compact(&compactData, event.Data); err != nil {
		return nil, &WebhookError{Message: "failed to canonicalize webhook data"}
	}

	signed := timestamp + "." + compactData.String()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return nil, &WebhookError{Message: "webhook signature verification failed"}
	}

	return &event, nil
}
