package einvoice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ─────────────────────────────────────────────
// Enums
// ─────────────────────────────────────────────

// InvoiceType enumerates the kinds of invoice documents.
type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "standard"
	InvoiceTypeCreditNote InvoiceType = "credit_note"
	InvoiceTypeDebitNote  InvoiceType = "debit_note"
	InvoiceTypePrepayment InvoiceType = "prepayment"
)

// InvoiceStatus enumerates the lifecycle states of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusQueued    InvoiceStatus = "queued"
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusAccepted  InvoiceStatus = "accepted"
	InvoiceStatusRejected  InvoiceStatus = "rejected"
	InvoiceStatusFailed    InvoiceStatus = "failed"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// InvoicePaymentStatus enumerates payment states.
type InvoicePaymentStatus string

const (
	PaymentStatusUnpaid        InvoicePaymentStatus = "unpaid"
	PaymentStatusPartiallyPaid InvoicePaymentStatus = "partially_paid"
	PaymentStatusPaid          InvoicePaymentStatus = "paid"
	PaymentStatusOverdue       InvoicePaymentStatus = "overdue"
)

// Currency enumerates supported currencies.
type Currency string

const (
	CurrencyNGN Currency = "NGN"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
)

// ─────────────────────────────────────────────
// Entities
// ─────────────────────────────────────────────

// LineItem is a single line on an invoice.
type LineItem struct {
	ID                  string  `json:"id,omitempty"`
	Description         string  `json:"description"`
	Quantity            float64 `json:"quantity"`
	UnitPrice           float64 `json:"unitPrice"`
	LineExtensionAmount float64 `json:"lineExtensionAmount,omitempty"`
	TaxPercent          float64 `json:"taxPercent,omitempty"`
	TotalAmount         float64 `json:"totalAmount,omitempty"`
	HSNCode             string  `json:"hsnCode"`
	UnitCode            string  `json:"unitCode"`
	DiscountAmount      float64 `json:"discountAmount,omitempty"`
}

// Invoice is a full invoice record as returned by the platform.
type Invoice struct {
	ID                      string                 `json:"id"`
	OrganizationID          string                 `json:"organizationId"`
	SellerID                string                 `json:"sellerId"`
	BuyerID                 string                 `json:"buyerId"`
	InvoiceNumber           string                 `json:"invoiceNumber"`
	InvoiceType             InvoiceType            `json:"invoiceType"`
	Status                  InvoiceStatus          `json:"status"`
	InvoiceDate             string                 `json:"invoiceDate"`
	IssueTime               string                 `json:"issueTime,omitempty"`
	DueDate                 string                 `json:"dueDate,omitempty"`
	Currency                Currency               `json:"currency"`
	LineItems               []LineItem             `json:"lineItems"`
	TotalAmountExcludingTax float64                `json:"totalAmountExcludingTax"`
	TotalTaxAmount          float64                `json:"totalTaxAmount"`
	TotalAmountIncludingTax float64                `json:"totalAmountIncludingTax"`
	AllowanceAmount         float64                `json:"allowanceAmount"`
	ChargeAmount            float64                `json:"chargeAmount"`
	PayableAmount           float64                `json:"payableAmount"`
	PaymentStatus           InvoicePaymentStatus   `json:"paymentStatus"`
	PaymentTerms            string                 `json:"paymentTerms,omitempty"`
	Notes                   string                 `json:"notes,omitempty"`
	OriginalInvoiceNumber   string                 `json:"originalInvoiceNumber,omitempty"`
	PurchaseOrderReference  string                 `json:"purchaseOrderReference,omitempty"`
	AccountingCost          string                 `json:"accountingCost,omitempty"`
	Jurisdiction            string                 `json:"jurisdiction,omitempty"`
	RegulatoryData          map[string]any         `json:"regulatoryData,omitempty"`
	SubmissionAttempts      int                    `json:"submissionAttempts"`
	LastSubmissionAttemptAt string                 `json:"lastSubmissionAttemptAt,omitempty"`
	Metadata                map[string]any         `json:"metadata,omitempty"`
	CreatedAt               string                 `json:"createdAt"`
	UpdatedAt               string                 `json:"updatedAt"`
}

// ─────────────────────────────────────────────
// Params
// ─────────────────────────────────────────────

// CreateInvoiceLineItem is a line item supplied when creating/updating an invoice.
type CreateInvoiceLineItem struct {
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	TaxPercent     float64 `json:"taxPercent"`
	HSNCode        string  `json:"hsnCode"`
	UnitCode       string  `json:"unitCode"`
	DiscountAmount float64 `json:"discountAmount,omitempty"`
}

// CreateInvoiceParams is the payload for InvoiceService.Create.
type CreateInvoiceParams struct {
	SellerID              string                  `json:"sellerId"`
	BuyerID               string                  `json:"buyerId"`
	InvoiceNumber         string                  `json:"invoiceNumber"`
	InvoiceDate           string                  `json:"invoiceDate"`
	DueDate               string                  `json:"dueDate,omitempty"`
	Currency              Currency                `json:"currency,omitempty"`
	LineItems             []CreateInvoiceLineItem `json:"lineItems"`
	InvoiceType           InvoiceType             `json:"invoiceType,omitempty"`
	PaymentTerms          string                  `json:"paymentTerms,omitempty"`
	Notes                 string                  `json:"notes,omitempty"`
	OriginalInvoiceNumber string                  `json:"originalInvoiceNumber,omitempty"`
	AllowanceAmount       float64                 `json:"allowanceAmount,omitempty"`
	ChargeAmount          float64                 `json:"chargeAmount,omitempty"`
	PaymentStatus         InvoicePaymentStatus    `json:"paymentStatus,omitempty"`
}

// UpdateInvoiceParams is the payload for InvoiceService.Update.
type UpdateInvoiceParams struct {
	InvoiceDate     string                  `json:"invoiceDate,omitempty"`
	DueDate         string                  `json:"dueDate,omitempty"`
	Currency        Currency                `json:"currency,omitempty"`
	LineItems       []CreateInvoiceLineItem `json:"lineItems,omitempty"`
	InvoiceType     InvoiceType             `json:"invoiceType,omitempty"`
	PaymentTerms    string                  `json:"paymentTerms,omitempty"`
	Notes           string                  `json:"notes,omitempty"`
	AllowanceAmount float64                 `json:"allowanceAmount,omitempty"`
	ChargeAmount    float64                 `json:"chargeAmount,omitempty"`
	PaymentStatus   InvoicePaymentStatus    `json:"paymentStatus,omitempty"`
}

// ListInvoicesParams filters and paginates InvoiceService.List.
type ListInvoicesParams struct {
	Page            int
	Limit           int
	Status          InvoiceStatus
	InvoiceType     InvoiceType
	PaymentStatus   InvoicePaymentStatus
	Currency        Currency
	SellerID        string
	BuyerID         string
	InvoiceDateFrom string
	InvoiceDateTo   string
	DueDateFrom     string
	DueDateTo       string
	CreatedAtFrom   string
	CreatedAtTo     string
	Search          string
	SortBy          string
	// SortOrder must be "ASC" or "DESC".
	SortOrder string
}

func (p *ListInvoicesParams) query() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	setStr(q, "status", string(p.Status))
	setStr(q, "invoiceType", string(p.InvoiceType))
	setStr(q, "paymentStatus", string(p.PaymentStatus))
	setStr(q, "currency", string(p.Currency))
	setStr(q, "sellerId", p.SellerID)
	setStr(q, "buyerId", p.BuyerID)
	setStr(q, "invoiceDateFrom", p.InvoiceDateFrom)
	setStr(q, "invoiceDateTo", p.InvoiceDateTo)
	setStr(q, "dueDateFrom", p.DueDateFrom)
	setStr(q, "dueDateTo", p.DueDateTo)
	setStr(q, "createdAtFrom", p.CreatedAtFrom)
	setStr(q, "createdAtTo", p.CreatedAtTo)
	setStr(q, "search", p.Search)
	setStr(q, "sortBy", p.SortBy)
	setStr(q, "sortOrder", p.SortOrder)
	return q
}

// CancelInvoiceParams is the payload for InvoiceService.Cancel.
type CancelInvoiceParams struct {
	Reason string `json:"reason"`
}

// InvoiceStatisticsParams filters InvoiceService.GetStatistics.
type InvoiceStatisticsParams struct {
	DateFrom string
	DateTo   string
}

func (p *InvoiceStatisticsParams) query() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	setStr(q, "dateFrom", p.DateFrom)
	setStr(q, "dateTo", p.DateTo)
	return q
}

// GetHsnCodesParams filters InvoiceService.GetHsnCodes.
type GetHsnCodesParams struct {
	Search   string
	Category string
	Page     int
	Limit    int
}

func (p *GetHsnCodesParams) query() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	setStr(q, "search", p.Search)
	setStr(q, "category", p.Category)
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	return q
}

// ValidateReferenceParams is the payload for InvoiceService.ValidateReference.
type ValidateReferenceParams struct {
	Reference string `json:"reference"`
}

// TaxReportParams is the payload for InvoiceService.SubmitTaxReport.
type TaxReportParams struct {
	Period string         `json:"period,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

// ─────────────────────────────────────────────
// Result types
// ─────────────────────────────────────────────

// SubmitResult is returned when an invoice is submitted to the tax authority.
type SubmitResult struct {
	InvoiceID string        `json:"invoiceId"`
	Status    InvoiceStatus `json:"status"`
	FIRSRef   string        `json:"firsRef,omitempty"`
	Message   string        `json:"message,omitempty"`
}

// BatchSubmitResult is returned by InvoiceService.BatchSubmit.
type BatchSubmitResult struct {
	Submitted int            `json:"submitted"`
	Failed    int            `json:"failed"`
	Results   []SubmitResult `json:"results,omitempty"`
}

// RetryResult is returned by InvoiceService.Retry.
type RetryResult struct {
	InvoiceID string        `json:"invoiceId"`
	Status    InvoiceStatus `json:"status"`
	Message   string        `json:"message,omitempty"`
}

// InvoiceStatusResult is returned by InvoiceService.GetStatus / QueryStatus.
type InvoiceStatusResult struct {
	InvoiceID string        `json:"invoiceId"`
	Status    InvoiceStatus `json:"status"`
	FIRSRef   string        `json:"firsRef,omitempty"`
	Message   string        `json:"message,omitempty"`
}

// ValidationError is a single validation finding.
type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ValidationResult is returned by InvoiceService.Validate / ValidateReference.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// DownloadResult is returned by InvoiceService.Download.
type DownloadResult struct {
	Format      string `json:"format"`
	ContentType string `json:"contentType,omitempty"`
	URL         string `json:"url,omitempty"`
	// Content is the base64-encoded document body, when returned inline.
	Content string `json:"content,omitempty"`
}

// InvoiceStatistics is returned by InvoiceService.GetStatistics.
type InvoiceStatistics struct {
	Total     int                    `json:"total"`
	ByStatus  map[string]int         `json:"byStatus,omitempty"`
	TotalValue float64               `json:"totalValue,omitempty"`
	Extra     map[string]any         `json:"-"`
}

// HsnCode is a Harmonised System / service classification code.
type HsnCode struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// HsnCategory groups HSN codes.
type HsnCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count,omitempty"`
}

// InvoiceResources holds reference data needed to build an invoice
// (currencies, unit codes, invoice types, etc.).
type InvoiceResources struct {
	Currencies   []string `json:"currencies,omitempty"`
	UnitCodes    []string `json:"unitCodes,omitempty"`
	InvoiceTypes []string `json:"invoiceTypes,omitempty"`
}

// InvoiceResourceType selects a single resource collection.
type InvoiceResourceType string

// TaxpayerInfo is returned by InvoiceService.LookupTaxID.
type TaxpayerInfo struct {
	TaxID   string `json:"taxId"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	Address string `json:"address,omitempty"`
}

// TaxReportResult is returned by InvoiceService.SubmitTaxReport.
type TaxReportResult struct {
	Reference string `json:"reference,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ─────────────────────────────────────────────
// Service
// ─────────────────────────────────────────────

// InvoiceService groups invoice endpoints under /i/v1/invoices.
type InvoiceService struct {
	http *httpClient
}

// Create creates (and the platform validates + signs + submits) an invoice.
func (s *InvoiceService) Create(ctx context.Context, params *CreateInvoiceParams, opts ...RequestOption) (*Invoice, error) {
	var out Invoice
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves an invoice by ID.
func (s *InvoiceService) Get(ctx context.Context, id string, opts ...RequestOption) (*Invoice, error) {
	var out Invoice
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns invoices matching params, along with pagination metadata.
func (s *InvoiceService) List(ctx context.Context, params *ListInvoicesParams, opts ...RequestOption) ([]Invoice, *Pagination, error) {
	var out []Invoice
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// Update modifies a draft invoice.
func (s *InvoiceService) Update(ctx context.Context, id string, params *UpdateInvoiceParams, opts ...RequestOption) (*Invoice, error) {
	var out Invoice
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPatch, path: "/i/v1/invoices/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a draft invoice.
func (s *InvoiceService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	_, err := s.http.do(ctx, internalRequest{method: http.MethodDelete, path: "/i/v1/invoices/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// Submit submits an invoice to the tax authority.
func (s *InvoiceService) Submit(ctx context.Context, id string, opts ...RequestOption) (*SubmitResult, error) {
	var out SubmitResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/" + url.PathEscape(id) + "/submit", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BatchSubmit submits multiple invoices in one call.
func (s *InvoiceService) BatchSubmit(ctx context.Context, invoiceIDs []string, opts ...RequestOption) (*BatchSubmitResult, error) {
	body := map[string]any{"invoiceIds": invoiceIDs}
	var out BatchSubmitResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/batch-submit", body: body, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Retry retries a failed submission.
func (s *InvoiceService) Retry(ctx context.Context, id string, opts ...RequestOption) (*RetryResult, error) {
	var out RetryResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/" + url.PathEscape(id) + "/retry", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a submitted invoice.
func (s *InvoiceService) Cancel(ctx context.Context, id string, params *CancelInvoiceParams, opts ...RequestOption) (*Invoice, error) {
	var out Invoice
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/" + url.PathEscape(id) + "/cancel", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStatus returns the current status of an invoice.
func (s *InvoiceService) GetStatus(ctx context.Context, id string, opts ...RequestOption) (*InvoiceStatusResult, error) {
	var out InvoiceStatusResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/" + url.PathEscape(id) + "/status", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// QueryStatus forces a fresh status query against the tax authority.
func (s *InvoiceService) QueryStatus(ctx context.Context, id string, opts ...RequestOption) (*InvoiceStatusResult, error) {
	var out InvoiceStatusResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/" + url.PathEscape(id) + "/query-status", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Validate validates an invoice payload without persisting it.
func (s *InvoiceService) Validate(ctx context.Context, params *CreateInvoiceParams, opts ...RequestOption) (*ValidationResult, error) {
	var out ValidationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/validate", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Download returns a downloadable representation of an invoice.
// format may be "pdf" or "xml"; pass an empty string for the default.
func (s *InvoiceService) Download(ctx context.Context, id string, format string, opts ...RequestOption) (*DownloadResult, error) {
	var q url.Values
	if format != "" {
		q = url.Values{"format": []string{format}}
	}
	var out DownloadResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/" + url.PathEscape(id) + "/download", query: q, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStatistics returns aggregate invoice statistics.
func (s *InvoiceService) GetStatistics(ctx context.Context, params *InvoiceStatisticsParams, opts ...RequestOption) (*InvoiceStatistics, error) {
	var out InvoiceStatistics
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/statistics", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetHsnCodes lists HSN / service classification codes.
func (s *InvoiceService) GetHsnCodes(ctx context.Context, params *GetHsnCodesParams, opts ...RequestOption) ([]HsnCode, error) {
	var out []HsnCode
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/hsn-codes", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetHsnCategories lists HSN code categories.
func (s *InvoiceService) GetHsnCategories(ctx context.Context, opts ...RequestOption) ([]HsnCategory, error) {
	var out []HsnCategory
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/hsn-codes/categories", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetResources returns reference data for building invoices.
func (s *InvoiceService) GetResources(ctx context.Context, opts ...RequestOption) (*InvoiceResources, error) {
	var out InvoiceResources
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/resources", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetResourceByType returns a single resource collection by type.
func (s *InvoiceService) GetResourceByType(ctx context.Context, resourceType InvoiceResourceType, opts ...RequestOption) ([]any, error) {
	var out []any
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/resources/" + url.PathEscape(string(resourceType)), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LookupTaxID looks up taxpayer information by tax ID.
func (s *InvoiceService) LookupTaxID(ctx context.Context, taxID string, opts ...RequestOption) (*TaxpayerInfo, error) {
	var out TaxpayerInfo
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/invoices/lookup/tax-id/" + url.PathEscape(taxID), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ValidateReference validates an invoice reference.
func (s *InvoiceService) ValidateReference(ctx context.Context, params *ValidateReferenceParams, opts ...RequestOption) (*ValidationResult, error) {
	var out ValidationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/reference/validate", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitTaxReport submits a tax report for an invoice.
func (s *InvoiceService) SubmitTaxReport(ctx context.Context, invoiceID string, params *TaxReportParams, opts ...RequestOption) (*TaxReportResult, error) {
	var out TaxReportResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/invoices/" + url.PathEscape(invoiceID) + "/tax-report", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ─────────────────────────────────────────────
// query helpers (shared across services)
// ─────────────────────────────────────────────

func setStr(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func setInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}
