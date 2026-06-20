package einvoice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// httpClient performs authenticated requests against the API gateway with
// retry (exponential backoff with jitter), rate-limit handling, and
// context-based timeouts. It is internal to the SDK.
type httpClient struct {
	apiKey         string
	baseURL        string
	timeout        time.Duration
	maxRetries     int
	retryBaseDelay time.Duration
	headers        map[string]string
	hc             *http.Client
}

type internalRequest struct {
	method string
	path   string
	body   any
	query  url.Values
	opts   *RequestOptions
}

// envelope is the {meta, data} response wrapper used by the platform.
type envelope struct {
	Meta ResponseMeta    `json:"meta"`
	Data json.RawMessage `json:"data"`
}

// do executes a request, decoding the response's data field into out (which
// may be nil). It returns the response metadata on success.
func (c *httpClient) do(ctx context.Context, r internalRequest, out any) (*ResponseMeta, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	u := c.baseURL + r.path
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	var bodyBytes []byte
	if r.body != nil {
		b, err := json.Marshal(r.body)
		if err != nil {
			return nil, fmt.Errorf("einvoice: failed to encode request body: %w", err)
		}
		bodyBytes = b
	}

	timeout := c.timeout
	if r.opts != nil && r.opts.Timeout > 0 {
		timeout = r.opts.Timeout
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(c.backoff(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		meta, retryable, err := c.attempt(ctx, r.method, u, bodyBytes, r.opts, timeout, out)
		if err == nil {
			return meta, nil
		}
		lastErr = err
		if !retryable || attempt == c.maxRetries {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *httpClient) attempt(
	ctx context.Context,
	method, u string,
	body []byte,
	opts *RequestOptions,
	timeout time.Duration,
	out any,
) (meta *ResponseMeta, retryable bool, err error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(reqCtx, method, u, bodyReader)
	if err != nil {
		return nil, false, fmt.Errorf("einvoice: failed to build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if opts != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		// Parent context cancelled/expired: do not retry.
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		// Our per-request timeout fired.
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, true, &TimeoutError{Timeout: timeout}
		}
		// Other network/transport errors are retryable.
		return nil, true, fmt.Errorf("einvoice: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("einvoice: failed to read response body: %w", err)
	}

	var env envelope
	if len(raw) > 0 {
		// Ignore decode errors for now; they're surfaced below depending on status.
		_ = json.Unmarshal(raw, &env)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), 5)
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Code:       "RATE_LIMIT_EXCEEDED",
			Message:    firstNonEmpty(env.Meta.Message, "Rate limit exceeded"),
			Errors:     env.Meta.Errors,
		}
		return nil, true, &RateLimitError{APIError: apiErr, RetryAfter: retryAfter}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    firstNonEmpty(env.Meta.Message, fmt.Sprintf("HTTP %d", resp.StatusCode)),
			Errors:     env.Meta.Errors,
		}
		if len(env.Meta.Errors) > 0 {
			apiErr.Code = env.Meta.Errors[0].Code
		}
		// Only 5xx is retryable; 4xx (except 429, handled above) is not.
		return nil, resp.StatusCode >= 500, apiErr
	}

	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return nil, false, fmt.Errorf("einvoice: failed to decode response data: %w", err)
		}
	}

	return &env.Meta, false, nil
}

func (c *httpClient) backoff(attempt int) time.Duration {
	base := float64(c.retryBaseDelay) * math.Pow(2, float64(attempt-1))
	jitter := rand.Float64() * base * 0.1
	return time.Duration(base + jitter)
}

func parseRetryAfter(header string, fallback int) int {
	if header == "" {
		return fallback
	}
	if n, err := strconv.Atoi(header); err == nil {
		return n
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
