package einvoice

import (
	"net/http"
	"strings"
	"time"
)

// defaultBaseURL is the API gateway. Both sandbox (sk_test_*) and live
// (sk_live_*) keys use the same gateway.
const defaultBaseURL = "https://gateway.useyona.com"

// Config configures a Client. Only APIKey is required.
type Config struct {
	// APIKey authenticates every request. Must start with sk_live_ or sk_test_.
	APIKey string
	// OrganizationID scopes org-scoped services (webhooks, API keys, users).
	OrganizationID string
	// BaseURL overrides the API gateway base URL. Defaults to defaultBaseURL.
	BaseURL string
	// Timeout is the default per-request timeout. Defaults to 30s.
	Timeout time.Duration
	// MaxRetries is the maximum number of retries for 5xx/429/network errors.
	// Zero uses the default of 3; a negative value disables retries.
	MaxRetries int
	// RetryBaseDelay is the base delay for exponential backoff. Defaults to 1s.
	RetryBaseDelay time.Duration
	// Headers are default headers sent on every request.
	Headers map[string]string
	// HTTPClient lets callers supply a custom *http.Client. Optional.
	HTTPClient *http.Client
}

// Client is the main entry point to the E-Invoice platform. Create one with New.
type Client struct {
	// Invoices manages the invoice lifecycle: create, submit, validate, etc.
	Invoices *InvoiceService
	// Sellers manages seller parties.
	Sellers *SellerService
	// Buyers manages buyer parties.
	Buyers *BuyerService
	// Billing manages credits, subscriptions, payments, and analytics.
	Billing *BillingService

	http   *httpClient
	config Config
	orgID  string
}

// New creates a Client. It returns a *ConfigError if APIKey is empty.
func New(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, &ConfigError{Message: "API key is required; set Config.APIKey"}
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	maxRetries := 3
	if cfg.MaxRetries != 0 {
		if cfg.MaxRetries < 0 {
			maxRetries = 0
		} else {
			maxRetries = cfg.MaxRetries
		}
	}

	retryBaseDelay := cfg.RetryBaseDelay
	if retryBaseDelay <= 0 {
		retryBaseDelay = time.Second
	}

	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}

	h := &httpClient{
		apiKey:         cfg.APIKey,
		baseURL:        baseURL,
		timeout:        timeout,
		maxRetries:     maxRetries,
		retryBaseDelay: retryBaseDelay,
		headers:        cfg.Headers,
		hc:             hc,
	}

	c := &Client{
		http:   h,
		config: cfg,
		orgID:  cfg.OrganizationID,
	}
	c.Invoices = &InvoiceService{http: h}
	c.Sellers = &SellerService{http: h}
	c.Buyers = &BuyerService{http: h}
	c.Billing = &BillingService{http: h}
	return c, nil
}

// OrganizationID returns the organization ID this client is scoped to, if any.
func (c *Client) OrganizationID() string { return c.orgID }

// ForOrganization returns a new Client scoped to a different organization,
// reusing the same API key and transport configuration.
func (c *Client) ForOrganization(organizationID string) *Client {
	cfg := c.config
	cfg.OrganizationID = organizationID
	// New only errors on empty APIKey, which cannot happen here.
	nc, _ := New(cfg)
	return nc
}

// ForAPIKey returns a new Client authenticated with a different API key,
// inheriting all other configuration. organizationID may be empty.
func (c *Client) ForAPIKey(apiKey, organizationID string) (*Client, error) {
	cfg := c.config
	cfg.APIKey = apiKey
	cfg.OrganizationID = organizationID
	return New(cfg)
}
