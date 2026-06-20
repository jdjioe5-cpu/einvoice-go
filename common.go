package einvoice

import "time"

// ResponseMeta is the metadata envelope returned alongside every API response.
type ResponseMeta struct {
	StatusCode int          `json:"statusCode"`
	Success    bool         `json:"success"`
	Errors     []FieldError `json:"errors"`
	Message    string       `json:"message"`
	Timestamp  string       `json:"timestamp"`
	Pagination *Pagination  `json:"pagination,omitempty"`
}

// Pagination describes the paging state of a list response.
type Pagination struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	PageSize    int  `json:"pageSize"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

// RequestOptions holds per-request overrides applied via RequestOption helpers.
type RequestOptions struct {
	// Timeout overrides the client timeout for this single request.
	Timeout time.Duration
	// Headers are merged into (and override) the client's default headers.
	Headers map[string]string
}

// RequestOption configures a single request.
type RequestOption func(*RequestOptions)

// WithTimeout overrides the client timeout for one request.
func WithTimeout(d time.Duration) RequestOption {
	return func(o *RequestOptions) { o.Timeout = d }
}

// WithHeader sets an additional header on one request.
func WithHeader(key, value string) RequestOption {
	return func(o *RequestOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

func applyOptions(opts []RequestOption) *RequestOptions {
	if len(opts) == 0 {
		return nil
	}
	o := &RequestOptions{}
	for _, fn := range opts {
		if fn != nil {
			fn(o)
		}
	}
	return o
}
