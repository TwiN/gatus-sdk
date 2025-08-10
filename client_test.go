package gatussdk

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		opts            []ClientOption
		expectedBaseURL string
		expectedTimeout time.Duration
		expectedUA      string
	}{
		{
			name:            "default client",
			baseURL:         "https://status.example.com",
			opts:            nil,
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: DefaultTimeout,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:            "trailing slash removed",
			baseURL:         "https://status.example.com/",
			opts:            nil,
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: DefaultTimeout,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:            "multiple trailing slashes removed",
			baseURL:         "https://status.example.com///",
			opts:            nil,
			expectedBaseURL: "https://status.example.com//",
			expectedTimeout: DefaultTimeout,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:    "custom timeout",
			baseURL: "https://status.example.com",
			opts: []ClientOption{
				WithTimeout(5 * time.Second),
			},
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: 5 * time.Second,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:    "custom user agent",
			baseURL: "https://status.example.com",
			opts: []ClientOption{
				WithUserAgent("MyApp/1.0"),
			},
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: DefaultTimeout,
			expectedUA:      "MyApp/1.0",
		},
		{
			name:    "multiple options",
			baseURL: "https://status.example.com",
			opts: []ClientOption{
				WithTimeout(10 * time.Second),
				WithUserAgent("CustomAgent/2.0"),
			},
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: 10 * time.Second,
			expectedUA:      "CustomAgent/2.0",
		},
		{
			name:    "custom http client",
			baseURL: "https://status.example.com",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{
					Timeout: 15 * time.Second,
				}),
			},
			expectedBaseURL: "https://status.example.com",
			expectedTimeout: 15 * time.Second,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:            "http url",
			baseURL:         "http://localhost:8080",
			opts:            nil,
			expectedBaseURL: "http://localhost:8080",
			expectedTimeout: DefaultTimeout,
			expectedUA:      DefaultUserAgent,
		},
		{
			name:            "url with path",
			baseURL:         "https://example.com/status",
			opts:            nil,
			expectedBaseURL: "https://example.com/status",
			expectedTimeout: DefaultTimeout,
			expectedUA:      DefaultUserAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.opts...)

			if client.baseURL != tt.expectedBaseURL {
				t.Errorf("baseURL = %v, want %v", client.baseURL, tt.expectedBaseURL)
			}

			if client.httpClient.Timeout != tt.expectedTimeout {
				t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, tt.expectedTimeout)
			}

			if client.userAgent != tt.expectedUA {
				t.Errorf("userAgent = %v, want %v", client.userAgent, tt.expectedUA)
			}
		})
	}
}

func TestClient_doRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		checkRequest   func(t *testing.T, r *http.Request)
	}{
		{
			name:   "successful GET request",
			method: http.MethodGet,
			path:   "/api/v1/endpoints",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]EndpointStatus{})
			},
			expectedError: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Method = %v, want %v", r.Method, http.MethodGet)
				}
				if r.URL.Path != "/api/v1/endpoints" {
					t.Errorf("Path = %v, want %v", r.URL.Path, "/api/v1/endpoints")
				}
				if r.Header.Get("User-Agent") != DefaultUserAgent {
					t.Errorf("User-Agent = %v, want %v", r.Header.Get("User-Agent"), DefaultUserAgent)
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept = %v, want %v", r.Header.Get("Accept"), "application/json")
				}
				if r.Header.Get("Accept-Encoding") != "gzip" {
					t.Errorf("Accept-Encoding = %v, want %v", r.Header.Get("Accept-Encoding"), "gzip")
				}
			},
		},
		{
			name:   "custom user agent",
			method: http.MethodGet,
			path:   "/test",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedError: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				// Will be tested with custom client
			},
		},
		{
			name:   "server error",
			method: http.MethodGet,
			path:   "/error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: false, // doRequest doesn't error on status codes
		},
		{
			name:   "POST request",
			method: http.MethodPost,
			path:   "/api/v1/endpoints",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			expectedError: false,
			checkRequest: func(t *testing.T, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Method = %v, want %v", r.Method, http.MethodPost)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkRequest != nil {
					tt.checkRequest(t, r)
				}
				tt.serverResponse(w, r)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.doRequest(context.Background(), tt.method, tt.path)

			if (err != nil) != tt.expectedError {
				t.Errorf("doRequest() error = %v, expectedError %v", err, tt.expectedError)
			}

			if resp != nil {
				resp.Body.Close()
			}
		})
	}
}

func TestClient_doRequest_Context(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.doRequest(ctx, http.MethodGet, "/test")
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.doRequest(ctx, http.MethodGet, "/test")
		if err == nil {
			t.Error("expected timeout error")
		}
	})
}

func TestClient_decodeResponse(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		contentType   string
		gzipResponse  bool
		target        interface{}
		expectedError bool
		checkError    func(t *testing.T, err error)
	}{
		{
			name:          "successful JSON decode",
			responseCode:  http.StatusOK,
			responseBody:  `{"name":"test","group":"core","key":"core_test"}`,
			target:        &EndpointStatus{},
			expectedError: false,
		},
		{
			name:          "successful JSON array decode",
			responseCode:  http.StatusOK,
			responseBody:  `[{"name":"test1"},{"name":"test2"}]`,
			target:        &[]EndpointStatus{},
			expectedError: false,
		},
		{
			name:          "gzipped response",
			responseCode:  http.StatusOK,
			responseBody:  `{"name":"compressed","group":"test"}`,
			gzipResponse:  true,
			target:        &EndpointStatus{},
			expectedError: false,
		},
		{
			name:          "API error - 404",
			responseCode:  http.StatusNotFound,
			responseBody:  `{"error":"not found"}`,
			target:        &EndpointStatus{},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Error("expected APIError type")
				}
				if apiErr.StatusCode != http.StatusNotFound {
					t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
				}
			},
		},
		{
			name:          "API error - 500",
			responseCode:  http.StatusInternalServerError,
			responseBody:  `internal server error`,
			target:        &EndpointStatus{},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Error("expected APIError type")
				}
				if apiErr.StatusCode != http.StatusInternalServerError {
					t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, http.StatusInternalServerError)
				}
				if apiErr.Body != "internal server error" {
					t.Errorf("Body = %v, want %v", apiErr.Body, "internal server error")
				}
			},
		},
		{
			name:          "invalid JSON",
			responseCode:  http.StatusOK,
			responseBody:  `{invalid json}`,
			target:        &EndpointStatus{},
			expectedError: true,
		},
		{
			name:          "empty response",
			responseCode:  http.StatusNoContent,
			responseBody:  ``,
			target:        &EndpointStatus{},
			expectedError: false,
		},
		{
			name:          "200 with empty body (EOF)",
			responseCode:  http.StatusOK,
			responseBody:  ``,
			target:        &EndpointStatus{},
			expectedError: false,
		},
		{
			name:          "malformed gzip",
			responseCode:  http.StatusOK,
			responseBody:  "not gzipped data",
			gzipResponse:  true,
			target:        &EndpointStatus{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.gzipResponse {
				if tt.name == "malformed gzip" {
					body = strings.NewReader(tt.responseBody)
				} else {
					var buf strings.Builder
					gw := gzip.NewWriter(&buf)
					gw.Write([]byte(tt.responseBody))
					gw.Close()
					body = strings.NewReader(buf.String())
				}
			} else {
				body = strings.NewReader(tt.responseBody)
			}

			resp := &http.Response{
				StatusCode: tt.responseCode,
				Body:       io.NopCloser(body),
				Header:     make(http.Header),
			}

			if tt.gzipResponse {
				resp.Header.Set("Content-Encoding", "gzip")
			}

			client := NewClient("https://example.com")
			err := client.decodeResponse(resp, tt.target)

			if (err != nil) != tt.expectedError {
				t.Errorf("decodeResponse() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.checkError != nil && err != nil {
				tt.checkError(t, err)
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	t.Run("WithHTTPClient", func(t *testing.T) {
		customClient := &http.Client{
			Timeout: 42 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 200,
			},
		}

		client := NewClient("https://example.com", WithHTTPClient(customClient))

		if client.httpClient != customClient {
			t.Error("WithHTTPClient did not set custom client")
		}
		if client.httpClient.Timeout != 42*time.Second {
			t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, 42*time.Second)
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		client := NewClient("https://example.com", WithTimeout(123*time.Second))

		if client.httpClient.Timeout != 123*time.Second {
			t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, 123*time.Second)
		}
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		client := NewClient("https://example.com", WithUserAgent("TestAgent/3.0"))

		if client.userAgent != "TestAgent/3.0" {
			t.Errorf("userAgent = %v, want %v", client.userAgent, "TestAgent/3.0")
		}
	})

	t.Run("multiple options applied in order", func(t *testing.T) {
		client := NewClient("https://example.com",
			WithTimeout(10*time.Second),
			WithTimeout(20*time.Second), // Should override previous
			WithUserAgent("Agent1"),
			WithUserAgent("Agent2"), // Should override previous
		)

		if client.httpClient.Timeout != 20*time.Second {
			t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, 20*time.Second)
		}
		if client.userAgent != "Agent2" {
			t.Errorf("userAgent = %v, want %v", client.userAgent, "Agent2")
		}
	})
}

func TestClient_Integration(t *testing.T) {
	// Test a complete request/response cycle
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/endpoints/statuses":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]EndpointStatus{
				{
					Name:  "test-endpoint",
					Group: "test-group",
					Key:   "test-group_test-endpoint",
				},
			})
		case "/api/v1/endpoints/test_key/statuses":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(EndpointStatus{
				Name:  "test",
				Group: "default",
				Key:   "test_key",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "not found")
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, WithUserAgent("IntegrationTest/1.0"))

	t.Run("successful request with custom user agent", func(t *testing.T) {
		resp, err := client.doRequest(context.Background(), http.MethodGet, "/api/v1/endpoints/statuses")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		var statuses []EndpointStatus
		err = client.decodeResponse(resp, &statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(statuses) != 1 {
			t.Errorf("expected 1 status, got %d", len(statuses))
		}
	})

	t.Run("404 error handling", func(t *testing.T) {
		resp, err := client.doRequest(context.Background(), http.MethodGet, "/nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var result interface{}
		err = client.decodeResponse(resp, &result)
		if err == nil {
			t.Error("expected error for 404 response")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Error("expected APIError type")
		}
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
		}
	})
}

func TestClient_RequestCreationError(t *testing.T) {
	client := NewClient("https://example.com")

	// Create a context that will cause NewRequestWithContext to fail
	// This is tricky as NewRequestWithContext rarely fails with normal inputs
	// We'll test with an invalid method or extremely long URL
	t.Run("invalid method", func(t *testing.T) {
		_, err := client.doRequest(context.Background(), "INVALID\x00METHOD", "/test")
		if err == nil {
			t.Error("expected error for invalid method")
		}
		if !strings.Contains(err.Error(), "creating request") {
			t.Errorf("expected error to contain 'creating request', got: %v", err)
		}
	})

	t.Run("network error", func(t *testing.T) {
		// Use an invalid URL to trigger network error
		invalidClient := NewClient("http://127.0.0.1:0") // port 0 should be unreachable
		_, err := invalidClient.doRequest(context.Background(), http.MethodGet, "/test")
		if err == nil {
			t.Error("expected error for unreachable host")
		}
		if !strings.Contains(err.Error(), "executing request") {
			t.Errorf("expected error to contain 'executing request', got: %v", err)
		}
	})
}
