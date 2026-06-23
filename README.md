# einvoice-go

The official Go SDK for the **Elyonar / Yona** e-invoicing platform — e-invoicing
infrastructure for Nigeria (UBL 2.1 validation, cryptographic signing, TIN
verification, and FIRS/NRS submission).

It is an idiomatic Go port of the TypeScript SDK
[`@useyona/einvoice-js`](https://www.npmjs.com/package/@useyona/einvoice-js) and
targets the same API gateway. **Zero runtime dependencies** — standard library only.

> Status: core of the SDK (client, transport, errors, webhooks) plus the
> invoice / seller / buyer / billing services, the organization + API-key
> services (B2B2B: child orgs and scoped key provisioning), the org-scoped
> user + role (RBAC) services, and the org-scoped webhook-endpoint management
> service. The remaining service (invitations) is tracked for follow-up to
> reach full parity with the JS SDK's 117 methods.

## Install

```bash
go get github.com/Elyonar/einvoice-go
```

```go
import einvoice "github.com/Elyonar/einvoice-go"
```

Requires Go 1.22+.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	einvoice "github.com/Elyonar/einvoice-go"
)

func main() {
	client, err := einvoice.New(einvoice.Config{
		APIKey: "sk_test_...", // sk_test_* = sandbox, sk_live_* = production
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create an invoice (the platform validates, signs, and submits it).
	inv, err := client.Invoices.Create(ctx, &einvoice.CreateInvoiceParams{
		SellerID:      "sel_123",
		BuyerID:       "buy_456",
		InvoiceNumber: "INV-0001",
		InvoiceDate:   "2026-06-19",
		Currency:      einvoice.CurrencyNGN,
		LineItems: []einvoice.CreateInvoiceLineItem{{
			Description: "Consulting services",
			Quantity:    1,
			UnitPrice:   100000,
			TaxPercent:  7.5,
			HSNCode:     "9983",
			UnitCode:    "EA",
		}},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s status=%s\n", inv.ID, inv.Status)
}
```

## Configuration

```go
client, _ := einvoice.New(einvoice.Config{
	APIKey:         "sk_test_...",
	OrganizationID: "org_123",          // for org-scoped operations
	BaseURL:        "",                  // defaults to the gateway
	Timeout:        30 * time.Second,    // default
	MaxRetries:     3,                   // 0 = default (3), negative = disabled
	RetryBaseDelay: time.Second,         // exponential backoff base
	Headers:        map[string]string{}, // default headers
})
```

The HTTP layer adds the `Authorization` header, retries `5xx`/`429`/network
errors with exponential backoff + jitter, honors `Retry-After`, and enforces
the timeout via `context.Context`.

### Per-request options

Every method takes optional `RequestOption`s:

```go
inv, err := client.Invoices.Get(ctx, "inv_1",
	einvoice.WithTimeout(5*time.Second),
	einvoice.WithHeader("X-Idempotency-Key", "abc123"),
)
```

## Errors

Errors are typed and work with `errors.As`:

```go
_, err := client.Invoices.Get(ctx, "missing")

var apiErr *einvoice.APIError
var rlErr  *einvoice.RateLimitError
switch {
case errors.As(err, &rlErr):
	fmt.Println("rate limited, retry after", rlErr.RetryAfter, "s")
case errors.As(err, &apiErr):
	fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Errors)
}
```

Types: `*APIError`, `*RateLimitError` (unwraps to `*APIError`), `*TimeoutError`,
`*ConfigError`, `*WebhookError`, and the base `*EInvoiceError`.

## Webhooks

Verify webhook signatures (HMAC-SHA256, timing-safe, with replay protection):

```go
func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	event, err := einvoice.VerifyWebhook(body, r.Header, os.Getenv("WEBHOOK_SECRET"), nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch event.Type {
	case "invoice.approved":
		// handle approved invoice
	}
	w.WriteHeader(http.StatusOK)
}
```

## API reference (this release)

| Service | Methods |
|---------|---------|
| `client.Invoices` | `Create`, `Get`, `List`, `Update`, `Delete`, `Submit`, `BatchSubmit`, `Retry`, `Cancel`, `GetStatus`, `QueryStatus`, `Validate`, `Download`, `GetStatistics`, `GetHsnCodes`, `GetHsnCategories`, `GetResources`, `GetResourceByType`, `LookupTaxID`, `ValidateReference`, `SubmitTaxReport` |
| `client.Sellers` | `Create`, `Get`, `List`, `Update`, `Delete`, `VerifyTIN`, `GetVerificationStatus`, `Search`, `BulkDelete` |
| `client.Buyers` | `Create`, `Get`, `List`, `Update`, `Delete`, `VerifyTIN`, `GetVerificationStatus`, `Search`, `BulkDelete` |
| `client.Billing` | `GetAccount`, `CheckBalance`, `GetAccountStats`, `GetPlans`, `GetPlan`, `GetPackages`, `GetCreditCosts`, `GetActiveSubscription`, `GetSubscriptionHistory`, `GetPayments`, `GetTransactions`, `GetTransaction`, `GetUsageAnalytics`, `PurchaseCredits`, `TransferCredits` |
| `client.Organizations` | `Create`, `Get`, `List`, `Update`, `ListChildren` |
| `client.APIKeys` (org-scoped) | `Create`, `List`, `Get`, `Revoke`, `Rotate` |
| `client.Users` (org-scoped) | `List`, `Get`, `Update`, `UpdateRole`, `UpdateStatus`, `Remove` |
| `client.Roles` (org-scoped) | `List`, `Get`, `Create`, `Update`, `Delete`, `AvailablePermissions` |
| `client.Webhooks` (org-scoped) | `Create`, `List`, `Get`, `Update`, `Delete`, `Test`, `ListDeliveries`, `RetryDelivery` |
| `Client` (B2B2B) | `CreateOrganizationWithAPIKey` |
| package-level | `VerifyWebhook` (verify an incoming webhook signature) |

Billing is read-only / checkout-initiating by design — the SDK never moves money directly.

### Multi-org (B2B2B)

```go
partner, _ := einvoice.New(einvoice.Config{APIKey: "sk_live_...", OrganizationID: "partner_org"})

// Operate as a child org using its own key:
child, _ := partner.ForAPIKey("sk_live_child_key", "child_org_id")
child.Invoices.Create(ctx, params)

// Or scope the partner's key to another org:
scoped := partner.ForOrganization("child_org_id")
```

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

## License

MIT — see [LICENSE](./LICENSE).
