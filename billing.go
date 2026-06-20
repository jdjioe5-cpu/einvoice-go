package einvoice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ─────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────

// BillingAccount is a tenant's billing account.
type BillingAccount struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId,omitempty"`
	CreditBalance  float64  `json:"creditBalance"`
	Currency       Currency `json:"currency,omitempty"`
	Status         string   `json:"status,omitempty"`
	CreatedAt      string   `json:"createdAt,omitempty"`
	UpdatedAt      string   `json:"updatedAt,omitempty"`
}

// BalanceCheck reports whether an account has enough credits for an operation.
type BalanceCheck struct {
	Sufficient bool    `json:"sufficient"`
	Balance    float64 `json:"balance"`
	Required   float64 `json:"required"`
}

// BillingAccountStats summarises account usage.
type BillingAccountStats struct {
	TotalCreditsPurchased float64 `json:"totalCreditsPurchased,omitempty"`
	TotalCreditsConsumed  float64 `json:"totalCreditsConsumed,omitempty"`
	CurrentBalance        float64 `json:"currentBalance,omitempty"`
}

// SubscriptionPlan is a subscription tier.
type SubscriptionPlan struct {
	Code        string         `json:"code"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Pricing     map[string]any `json:"pricing,omitempty"`
}

// CreditPackage is a one-off credit bundle.
type CreditPackage struct {
	ID      string         `json:"id"`
	Name    string         `json:"name,omitempty"`
	Credits float64        `json:"credits,omitempty"`
	Pricing map[string]any `json:"pricing,omitempty"`
}

// CreditCostEntry describes how many credits an endpoint/action consumes.
type CreditCostEntry struct {
	Key   string  `json:"key"`
	Cost  float64 `json:"cost"`
	Label string  `json:"label,omitempty"`
}

// Subscription is an active or historical subscription.
type Subscription struct {
	ID        string `json:"id"`
	PlanCode  string `json:"planCode,omitempty"`
	Status    string `json:"status,omitempty"`
	Interval  string `json:"interval,omitempty"`
	StartedAt string `json:"startedAt,omitempty"`
	EndsAt    string `json:"endsAt,omitempty"`
}

// Payment is a billing payment record.
type Payment struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount,omitempty"`
	Currency  string  `json:"currency,omitempty"`
	Status    string  `json:"status,omitempty"`
	CreatedAt string  `json:"createdAt,omitempty"`
}

// CreditTransaction is a credit ledger entry.
type CreditTransaction struct {
	ID        string  `json:"id"`
	Type      string  `json:"type,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Balance   float64 `json:"balance,omitempty"`
	Reference string  `json:"reference,omitempty"`
	CreatedAt string  `json:"createdAt,omitempty"`
}

// UsageAnalytics is a usage analytics summary.
type UsageAnalytics struct {
	Period string         `json:"period,omitempty"`
	Series []any          `json:"series,omitempty"`
	Totals map[string]any `json:"totals,omitempty"`
}

// PurchaseResult is returned by BillingService.PurchaseCredits.
type PurchaseResult struct {
	PaymentID  string `json:"paymentId,omitempty"`
	CheckoutURL string `json:"checkoutUrl,omitempty"`
	Status     string `json:"status,omitempty"`
}

// TransferCreditsResult is returned by BillingService.TransferCredits.
type TransferCreditsResult struct {
	TransferID    string  `json:"transferId,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	FromAccountID string  `json:"fromAccountId,omitempty"`
	ToAccountID   string  `json:"toAccountId,omitempty"`
}

// PurchaseCreditsParams is the payload for BillingService.PurchaseCredits.
type PurchaseCreditsParams struct {
	PackageID string `json:"packageId,omitempty"`
	Credits   float64 `json:"credits,omitempty"`
	Currency  string  `json:"currency,omitempty"`
}

// TransferCreditsParams is the payload for BillingService.TransferCredits.
type TransferCreditsParams struct {
	ToOrganizationID string  `json:"toOrganizationId"`
	Credits          float64 `json:"credits"`
	Note             string  `json:"note,omitempty"`
}

// ListParams is a simple page/limit filter used by several billing endpoints.
type ListParams struct {
	Page  int
	Limit int
}

func (p *ListParams) query() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	return q
}

// ─────────────────────────────────────────────
// Service
// ─────────────────────────────────────────────

// BillingService groups billing endpoints under /b/v1. By design every method
// here is read-only or initiates an asynchronous payment flow — the SDK never
// directly moves money.
type BillingService struct {
	http *httpClient
}

// GetAccount returns the caller's billing account.
func (s *BillingService) GetAccount(ctx context.Context, opts ...RequestOption) (*BillingAccount, error) {
	var out BillingAccount
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/billing-accounts/me", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckBalance checks whether an account has enough credits.
func (s *BillingService) CheckBalance(ctx context.Context, accountID string, credits float64, opts ...RequestOption) (*BalanceCheck, error) {
	q := url.Values{"credits": []string{strconv.FormatFloat(credits, 'f', -1, 64)}}
	var out BalanceCheck
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/billing-accounts/" + url.PathEscape(accountID) + "/check-balance", query: q, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAccountStats returns usage statistics for an account.
func (s *BillingService) GetAccountStats(ctx context.Context, accountID string, opts ...RequestOption) (*BillingAccountStats, error) {
	var out BillingAccountStats
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/billing-accounts/" + url.PathEscape(accountID) + "/stats", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPlans lists subscription plans.
func (s *BillingService) GetPlans(ctx context.Context, params *ListParams, opts ...RequestOption) ([]SubscriptionPlan, error) {
	var out []SubscriptionPlan
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/plans/subscriptions", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPlan returns a single subscription plan by code.
func (s *BillingService) GetPlan(ctx context.Context, code string, opts ...RequestOption) (*SubscriptionPlan, error) {
	var out SubscriptionPlan
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/plans/subscriptions/" + url.PathEscape(code), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPackages lists one-off credit packages.
func (s *BillingService) GetPackages(ctx context.Context, params *ListParams, opts ...RequestOption) ([]CreditPackage, error) {
	var out []CreditPackage
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/plans/packages", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetCreditCosts lists the credit cost of each metered action.
func (s *BillingService) GetCreditCosts(ctx context.Context, opts ...RequestOption) ([]CreditCostEntry, error) {
	var out []CreditCostEntry
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/plans/credit-costs", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetActiveSubscription returns the active subscription.
func (s *BillingService) GetActiveSubscription(ctx context.Context, opts ...RequestOption) (*Subscription, error) {
	var out Subscription
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/subscriptions/active", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSubscriptionHistory returns past subscriptions.
func (s *BillingService) GetSubscriptionHistory(ctx context.Context, opts ...RequestOption) ([]Subscription, error) {
	var out []Subscription
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/subscriptions/history", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetPayments returns payment history.
func (s *BillingService) GetPayments(ctx context.Context, params *ListParams, opts ...RequestOption) ([]Payment, *Pagination, error) {
	var out []Payment
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/payments/history", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// GetTransactions returns the credit ledger.
func (s *BillingService) GetTransactions(ctx context.Context, params *ListParams, opts ...RequestOption) ([]CreditTransaction, *Pagination, error) {
	var out []CreditTransaction
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/transactions", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// GetTransaction returns a single credit transaction by ID.
func (s *BillingService) GetTransaction(ctx context.Context, id string, opts ...RequestOption) (*CreditTransaction, error) {
	var out CreditTransaction
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/transactions/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUsageAnalytics returns credit usage analytics.
func (s *BillingService) GetUsageAnalytics(ctx context.Context, opts ...RequestOption) (*UsageAnalytics, error) {
	var out UsageAnalytics
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/b/v1/transactions/analytics/usage", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PurchaseCredits initiates a credit purchase. It returns a checkout flow —
// the SDK never charges a payment method directly.
func (s *BillingService) PurchaseCredits(ctx context.Context, params *PurchaseCreditsParams, opts ...RequestOption) (*PurchaseResult, error) {
	var out PurchaseResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/b/v1/payments/purchase", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// TransferCredits transfers credits from the caller's account to another
// organization (B2B2B credit distribution).
func (s *BillingService) TransferCredits(ctx context.Context, params *TransferCreditsParams, opts ...RequestOption) (*TransferCreditsResult, error) {
	var out TransferCreditsResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/b/v1/payments/transfer", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
