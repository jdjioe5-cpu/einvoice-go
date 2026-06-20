package einvoice

import (
	"context"
	"net/http"
	"net/url"
)

// ─────────────────────────────────────────────
// Shared party types
// ─────────────────────────────────────────────

// PartyType enumerates the legal type of a party.
type PartyType string

const (
	PartyTypeIndividual  PartyType = "individual"
	PartyTypeCompany     PartyType = "company"
	PartyTypePartnership PartyType = "partnership"
	PartyTypeNonProfit   PartyType = "non_profit"
	PartyTypeGovernment  PartyType = "government"
)

// PartyStatus enumerates the lifecycle state of a party.
type PartyStatus string

const (
	PartyStatusActive              PartyStatus = "active"
	PartyStatusInactive            PartyStatus = "inactive"
	PartyStatusSuspended           PartyStatus = "suspended"
	PartyStatusPendingVerification PartyStatus = "pending_verification"
)

// TaxNumberVerificationStatus enumerates TIN verification states.
type TaxNumberVerificationStatus string

const (
	TINNotVerified TaxNumberVerificationStatus = "not_verified"
	TINVerified    TaxNumberVerificationStatus = "verified"
	TINFailed      TaxNumberVerificationStatus = "failed"
	TINPending     TaxNumberVerificationStatus = "pending"
)

// PostalAddress is a structured postal address.
type PostalAddress struct {
	Line1       string `json:"line1"`
	Line2       string `json:"line2,omitempty"`
	District    string `json:"district,omitempty"`
	City        string `json:"city"`
	State       string `json:"state,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	Country     string `json:"country"`
	Subdivision string `json:"subdivision,omitempty"`
}

// TaxNumberVerificationResult is returned by TIN verification endpoints.
type TaxNumberVerificationResult struct {
	Status   TaxNumberVerificationStatus `json:"status"`
	TaxNumber string                     `json:"taxNumber,omitempty"`
	Name     string                      `json:"name,omitempty"`
	Message  string                      `json:"message,omitempty"`
	Result   map[string]any              `json:"result,omitempty"`
}

// SearchResult is returned by party search endpoints.
type SearchResult struct {
	Sellers []Seller `json:"sellers,omitempty"`
	Buyers  []Buyer  `json:"buyers,omitempty"`
	Total   int      `json:"total,omitempty"`
}

// Seller is a registered seller party.
type Seller struct {
	ID                          string                      `json:"id"`
	OrganizationID              string                      `json:"organizationId"`
	PartyName                   string                      `json:"partyName"`
	TaxNumber                   string                      `json:"taxNumber"`
	SellerType                  PartyType                   `json:"sellerType"`
	Status                      PartyStatus                 `json:"status"`
	RegistrationNumber          string                      `json:"registrationNumber,omitempty"`
	PostalAddress               *PostalAddress              `json:"postalAddress,omitempty"`
	Email                       string                      `json:"email,omitempty"`
	Phone                       string                      `json:"phone,omitempty"`
	Website                     string                      `json:"website,omitempty"`
	BusinessActivity            string                      `json:"businessActivity,omitempty"`
	TaxNumberVerificationStatus TaxNumberVerificationStatus `json:"taxNumberVerificationStatus"`
	TaxNumberVerifiedAt         string                      `json:"taxNumberVerifiedAt,omitempty"`
	LogoURL                     string                      `json:"logoUrl,omitempty"`
	Metadata                    map[string]any              `json:"metadata,omitempty"`
	CreatedAt                   string                      `json:"createdAt"`
	UpdatedAt                   string                      `json:"updatedAt"`
}

// Buyer is a registered buyer party.
type Buyer struct {
	ID                          string                      `json:"id"`
	OrganizationID              string                      `json:"organizationId"`
	PartyName                   string                      `json:"partyName"`
	TaxNumber                   string                      `json:"taxNumber,omitempty"`
	BuyerType                   PartyType                   `json:"buyerType"`
	Status                      PartyStatus                 `json:"status"`
	RegistrationNumber          string                      `json:"registrationNumber,omitempty"`
	PostalAddress               *PostalAddress              `json:"postalAddress,omitempty"`
	Email                       string                      `json:"email,omitempty"`
	Phone                       string                      `json:"phone,omitempty"`
	Website                     string                      `json:"website,omitempty"`
	BusinessActivity            string                      `json:"businessActivity,omitempty"`
	TaxNumberVerificationStatus TaxNumberVerificationStatus `json:"taxNumberVerificationStatus"`
	TaxNumberVerifiedAt         string                      `json:"taxNumberVerifiedAt,omitempty"`
	Metadata                    map[string]any              `json:"metadata,omitempty"`
	CreatedAt                   string                      `json:"createdAt"`
	UpdatedAt                   string                      `json:"updatedAt"`
}

// CreateSellerParams is the payload for SellerService.Create.
type CreateSellerParams struct {
	PartyName                   string                      `json:"partyName"`
	TradingName                 string                      `json:"tradingName,omitempty"`
	PartyType                   PartyType                   `json:"partyType,omitempty"`
	TaxNumber                   string                      `json:"taxNumber"`
	VATNumber                   string                      `json:"vatNumber,omitempty"`
	RegistrationNumber          string                      `json:"registrationNumber,omitempty"`
	Email                       string                      `json:"email"`
	PhoneNumber                 string                      `json:"phoneNumber"`
	Website                     string                      `json:"website,omitempty"`
	PostalAddress               PostalAddress               `json:"postalAddress"`
	Notes                       string                      `json:"notes,omitempty"`
	TaxNumberVerificationStatus TaxNumberVerificationStatus `json:"taxNumberVerificationStatus,omitempty"`
}

// UpdateSellerParams is the payload for SellerService.Update.
type UpdateSellerParams struct {
	PartyName          string         `json:"partyName,omitempty"`
	TradingName        string         `json:"tradingName,omitempty"`
	PartyType          PartyType      `json:"partyType,omitempty"`
	VATNumber          string         `json:"vatNumber,omitempty"`
	RegistrationNumber string         `json:"registrationNumber,omitempty"`
	Email              string         `json:"email,omitempty"`
	PhoneNumber        string         `json:"phoneNumber,omitempty"`
	Website            string         `json:"website,omitempty"`
	PostalAddress      *PostalAddress `json:"postalAddress,omitempty"`
	Notes              string         `json:"notes,omitempty"`
}

// CreateBuyerParams is the payload for BuyerService.Create.
type CreateBuyerParams struct {
	PartyName                   string                      `json:"partyName"`
	TradingName                 string                      `json:"tradingName,omitempty"`
	PartyType                   PartyType                   `json:"partyType,omitempty"`
	TaxNumber                   string                      `json:"taxNumber,omitempty"`
	VATNumber                   string                      `json:"vatNumber,omitempty"`
	RegistrationNumber          string                      `json:"registrationNumber,omitempty"`
	Email                       string                      `json:"email"`
	PhoneNumber                 string                      `json:"phoneNumber"`
	Website                     string                      `json:"website,omitempty"`
	PostalAddress               PostalAddress               `json:"postalAddress"`
	Notes                       string                      `json:"notes,omitempty"`
	TaxNumberVerificationStatus TaxNumberVerificationStatus `json:"taxNumberVerificationStatus,omitempty"`
}

// UpdateBuyerParams is the payload for BuyerService.Update.
type UpdateBuyerParams struct {
	PartyName          string         `json:"partyName,omitempty"`
	TradingName        string         `json:"tradingName,omitempty"`
	PartyType          PartyType      `json:"partyType,omitempty"`
	VATNumber          string         `json:"vatNumber,omitempty"`
	RegistrationNumber string         `json:"registrationNumber,omitempty"`
	Email              string         `json:"email,omitempty"`
	PhoneNumber        string         `json:"phoneNumber,omitempty"`
	Website            string         `json:"website,omitempty"`
	PostalAddress      *PostalAddress `json:"postalAddress,omitempty"`
	Notes              string         `json:"notes,omitempty"`
}

// ListPartiesParams paginates and searches seller/buyer list endpoints.
type ListPartiesParams struct {
	Page          int
	Limit         int
	Search        string
	CreatedAtFrom string
	CreatedAtTo   string
}

func (p *ListPartiesParams) query() url.Values {
	if p == nil {
		return nil
	}
	q := url.Values{}
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	setStr(q, "search", p.Search)
	setStr(q, "createdAtFrom", p.CreatedAtFrom)
	setStr(q, "createdAtTo", p.CreatedAtTo)
	return q
}

// ─────────────────────────────────────────────
// SellerService
// ─────────────────────────────────────────────

// SellerService groups seller endpoints under /i/v1/sellers.
type SellerService struct {
	http *httpClient
}

// Create registers a new seller.
func (s *SellerService) Create(ctx context.Context, params *CreateSellerParams, opts ...RequestOption) (*Seller, error) {
	var out Seller
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/sellers", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a seller by ID.
func (s *SellerService) Get(ctx context.Context, id string, opts ...RequestOption) (*Seller, error) {
	var out Seller
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/sellers/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns sellers matching params, along with pagination metadata.
func (s *SellerService) List(ctx context.Context, params *ListPartiesParams, opts ...RequestOption) ([]Seller, *Pagination, error) {
	var out []Seller
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/sellers", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// Update modifies a seller.
func (s *SellerService) Update(ctx context.Context, id string, params *UpdateSellerParams, opts ...RequestOption) (*Seller, error) {
	var out Seller
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPatch, path: "/i/v1/sellers/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a seller.
func (s *SellerService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	_, err := s.http.do(ctx, internalRequest{method: http.MethodDelete, path: "/i/v1/sellers/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// VerifyTIN triggers TIN verification for a seller.
func (s *SellerService) VerifyTIN(ctx context.Context, id string, opts ...RequestOption) (*TaxNumberVerificationResult, error) {
	var out TaxNumberVerificationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/sellers/" + url.PathEscape(id) + "/verify-tax-number", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVerificationStatus returns the TIN verification status for a seller.
func (s *SellerService) GetVerificationStatus(ctx context.Context, id string, opts ...RequestOption) (*TaxNumberVerificationResult, error) {
	var out TaxNumberVerificationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/sellers/" + url.PathEscape(id) + "/verification-status", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Search performs a free-text search across sellers.
func (s *SellerService) Search(ctx context.Context, query string, opts ...RequestOption) (*SearchResult, error) {
	var out SearchResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/sellers/search", query: url.Values{"query": []string{query}}, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkDelete deletes multiple sellers and returns the number deleted.
func (s *SellerService) BulkDelete(ctx context.Context, ids []string, opts ...RequestOption) (int, error) {
	body := map[string]any{"ids": ids}
	var out struct {
		Deleted int `json:"deleted"`
	}
	_, err := s.http.do(ctx, internalRequest{method: http.MethodDelete, path: "/i/v1/sellers/bulk", body: body, opts: applyOptions(opts)}, &out)
	if err != nil {
		return 0, err
	}
	return out.Deleted, nil
}

// ─────────────────────────────────────────────
// BuyerService
// ─────────────────────────────────────────────

// BuyerService groups buyer endpoints under /i/v1/buyers.
type BuyerService struct {
	http *httpClient
}

// Create registers a new buyer.
func (s *BuyerService) Create(ctx context.Context, params *CreateBuyerParams, opts ...RequestOption) (*Buyer, error) {
	var out Buyer
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/buyers", body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a buyer by ID.
func (s *BuyerService) Get(ctx context.Context, id string, opts ...RequestOption) (*Buyer, error) {
	var out Buyer
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/buyers/" + url.PathEscape(id), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns buyers matching params, along with pagination metadata.
func (s *BuyerService) List(ctx context.Context, params *ListPartiesParams, opts ...RequestOption) ([]Buyer, *Pagination, error) {
	var out []Buyer
	meta, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/buyers", query: params.query(), opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, meta.Pagination, nil
}

// Update modifies a buyer.
func (s *BuyerService) Update(ctx context.Context, id string, params *UpdateBuyerParams, opts ...RequestOption) (*Buyer, error) {
	var out Buyer
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPatch, path: "/i/v1/buyers/" + url.PathEscape(id), body: params, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a buyer.
func (s *BuyerService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	_, err := s.http.do(ctx, internalRequest{method: http.MethodDelete, path: "/i/v1/buyers/" + url.PathEscape(id), opts: applyOptions(opts)}, nil)
	return err
}

// VerifyTIN triggers TIN verification for a buyer.
func (s *BuyerService) VerifyTIN(ctx context.Context, id string, opts ...RequestOption) (*TaxNumberVerificationResult, error) {
	var out TaxNumberVerificationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodPost, path: "/i/v1/buyers/" + url.PathEscape(id) + "/verify-tax-number", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVerificationStatus returns the TIN verification status for a buyer.
func (s *BuyerService) GetVerificationStatus(ctx context.Context, id string, opts ...RequestOption) (*TaxNumberVerificationResult, error) {
	var out TaxNumberVerificationResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/buyers/" + url.PathEscape(id) + "/verification-status", opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Search performs a free-text search across buyers.
func (s *BuyerService) Search(ctx context.Context, query string, opts ...RequestOption) (*SearchResult, error) {
	var out SearchResult
	_, err := s.http.do(ctx, internalRequest{method: http.MethodGet, path: "/i/v1/buyers/search", query: url.Values{"query": []string{query}}, opts: applyOptions(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BulkDelete deletes multiple buyers and returns the number deleted.
func (s *BuyerService) BulkDelete(ctx context.Context, ids []string, opts ...RequestOption) (int, error) {
	body := map[string]any{"ids": ids}
	var out struct {
		Deleted int `json:"deleted"`
	}
	_, err := s.http.do(ctx, internalRequest{method: http.MethodDelete, path: "/i/v1/buyers/bulk", body: body, opts: applyOptions(opts)}, &out)
	if err != nil {
		return 0, err
	}
	return out.Deleted, nil
}
