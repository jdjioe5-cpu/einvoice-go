// Package einvoice is the official Go SDK for the Elyonar / Yona e-invoicing
// platform — e-invoicing infrastructure for Nigeria (UBL 2.1 validation,
// cryptographic signing, TIN verification, and FIRS/NRS submission).
//
// It is an idiomatic Go port of the TypeScript SDK @useyona/einvoice-js and
// targets the same API gateway. Construct a client with New, then use the
// typed service objects:
//
//	client, err := einvoice.New(einvoice.Config{APIKey: "sk_test_..."})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	inv, err := client.Invoices.Create(ctx, &einvoice.CreateInvoiceParams{
//		SellerID:      "sel_123",
//		BuyerID:       "buy_456",
//		InvoiceNumber: "INV-0001",
//		InvoiceDate:   "2026-06-19",
//		LineItems: []einvoice.CreateInvoiceLineItem{{
//			Description: "Consulting",
//			Quantity:    1,
//			UnitPrice:   100000,
//			TaxPercent:  7.5,
//			HSNCode:     "9983",
//			UnitCode:    "EA",
//		}},
//	})
//
// Every method takes a context.Context as its first argument and accepts
// optional per-request options (WithTimeout, WithHeader). Errors are returned
// as typed values (*APIError, *RateLimitError, *TimeoutError, *ConfigError,
// *WebhookError) and can be inspected with errors.As.
//
// The SDK has zero runtime dependencies — it uses only the Go standard library.
package einvoice
