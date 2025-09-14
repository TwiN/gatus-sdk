package gatussdk

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultUserAgent is the default User-Agent header value.
	DefaultUserAgent = "GatusSDK/1.0"
)

// Client is the main client for interacting with the Gatus API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// NewClient creates a new Gatus API client with the given base URL and options.
//
// Example:
//
//	client := NewClient("https://status.example.org")
//	client := NewClient("https://status.example.org", WithTimeout(10*time.Second))
func NewClient(baseURL string, opts ...ClientOption) *Client {
	// Remove trailing slash from base URL
	baseURL = strings.TrimSuffix(baseURL, "/")

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		userAgent: DefaultUserAgent,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithHTTPClient sets a custom HTTP client for the Gatus client.
//
// Example:
//
//	httpClient := &http.Client{Timeout: 5 * time.Second}
//	client := NewClient("https://status.example.org", WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout.
//
// Example:
//
//	client := NewClient("https://status.example.org", WithTimeout(10*time.Second))
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
//
// Example:
//
//	client := NewClient("https://status.example.org", WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// doRequest performs an HTTP request with the configured client settings.
func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

// doRequestWithAuth performs an HTTP request with the configured client settings and Bearer authentication.
func (c *Client) doRequestWithAuth(ctx context.Context, method, path string, token string) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

// decodeResponse decodes the HTTP response body, handling gzip compression if present.
func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	var reader io.Reader = resp.Body

	// Handle gzip compression
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(reader)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Body:       string(body),
		}
	}

	// For empty responses (like 204 No Content), don't try to decode
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Decode JSON response
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		// Check if it's EOF from empty response body
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}
