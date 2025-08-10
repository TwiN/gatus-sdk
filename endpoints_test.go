package gatussdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		wantErr  bool
	}{
		{"valid 1h", "1h", false},
		{"valid 24h", "24h", false},
		{"valid 7d", "7d", false},
		{"valid 30d", "30d", false},
		{"invalid 2h", "2h", true},
		{"invalid 48h", "48h", true},
		{"invalid 1d", "1d", true},
		{"invalid 60d", "60d", true},
		{"empty string", "", true},
		{"invalid format", "abc", true},
		{"invalid number", "123", true},
		{"case sensitive", "1H", true},
		{"with spaces", " 1h ", true},
		{"partial match", "1hr", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDuration(tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDuration(%q) error = %v, wantErr %v", tt.duration, err, tt.wantErr)
			}

			if err != nil && !tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}

			if err != nil && tt.wantErr {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Error("expected ValidationError type")
				}
				if valErr.Field != "duration" {
					t.Errorf("Field = %v, want 'duration'", valErr.Field)
				}
			}
		})
	}
}

func TestClient_GetAllEndpointStatuses(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedCount  int
		expectedError  bool
		checkResult    func(t *testing.T, statuses []EndpointStatus)
	}{
		{
			name: "successful response with multiple endpoints",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/endpoints/statuses" {
					t.Errorf("Path = %v, want /api/v1/endpoints/statuses", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Method = %v, want GET", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]EndpointStatus{
					{
						Name:  "blog-home",
						Group: "core",
						Key:   "core_blog-home",
						Results: []Result{
							{
								Status:    200,
								Success:   true,
								Timestamp: time.Now(),
							},
						},
					},
					{
						Name:  "api",
						Group: "services",
						Key:   "services_api",
						Results: []Result{
							{
								Status:    200,
								Success:   true,
								Timestamp: time.Now(),
							},
						},
					},
				})
			},
			expectedCount: 2,
			expectedError: false,
			checkResult: func(t *testing.T, statuses []EndpointStatus) {
				if statuses[0].Name != "blog-home" {
					t.Errorf("First status name = %v, want blog-home", statuses[0].Name)
				}
				if statuses[1].Group != "services" {
					t.Errorf("Second status group = %v, want services", statuses[1].Group)
				}
			},
		},
		{
			name: "empty response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("[]"))
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "invalid JSON response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{invalid json}"))
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "unauthorized",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized"}`))
			},
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			statuses, err := client.GetAllEndpointStatuses(context.Background())

			if (err != nil) != tt.expectedError {
				t.Errorf("GetAllEndpointStatuses() error = %v, expectedError %v", err, tt.expectedError)
			}

			if len(statuses) != tt.expectedCount {
				t.Errorf("got %d statuses, want %d", len(statuses), tt.expectedCount)
			}

			if tt.checkResult != nil && err == nil {
				tt.checkResult(t, statuses)
			}
		})
	}
}

func TestClient_GetEndpointStatusByKey(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		checkResult    func(t *testing.T, status *EndpointStatus)
		checkError     func(t *testing.T, err error)
	}{
		{
			name: "successful response",
			key:  "core_blog-home",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/endpoints/core_blog-home/statuses" {
					t.Errorf("Path = %v, want /api/v1/endpoints/core_blog-home/statuses", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(EndpointStatus{
					Name:  "blog-home",
					Group: "core",
					Key:   "core_blog-home",
					Results: []Result{
						{
							Status:  200,
							Success: true,
						},
					},
				})
			},
			expectedError: false,
			checkResult: func(t *testing.T, status *EndpointStatus) {
				if status.Name != "blog-home" {
					t.Errorf("Name = %v, want blog-home", status.Name)
				}
				if status.Key != "core_blog-home" {
					t.Errorf("Key = %v, want core_blog-home", status.Key)
				}
			},
		},
		{
			name: "key with special characters",
			key:  "api-v1_health-check",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, "api-v1_health-check") {
					t.Errorf("Path should contain escaped key")
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(EndpointStatus{
					Name:  "health-check",
					Group: "api-v1",
					Key:   "api-v1_health-check",
				})
			},
			expectedError: false,
		},
		{
			name:           "empty key",
			key:            "",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
			checkError: func(t *testing.T, err error) {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Error("expected ValidationError type")
				}
				if valErr.Field != "key" {
					t.Errorf("Field = %v, want 'key'", valErr.Field)
				}
			},
		},
		{
			name: "endpoint not found",
			key:  "nonexistent_endpoint",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "endpoint not found"}`))
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Error("expected APIError type")
				}
				if apiErr.StatusCode != http.StatusNotFound {
					t.Errorf("StatusCode = %v, want 404", apiErr.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			status, err := client.GetEndpointStatusByKey(context.Background(), tt.key)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEndpointStatusByKey() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.checkResult != nil && err == nil {
				tt.checkResult(t, status)
			}

			if tt.checkError != nil && err != nil {
				tt.checkError(t, err)
			}
		})
	}
}

func TestClient_GetEndpointStatus(t *testing.T) {
	tests := []struct {
		name           string
		group          string
		endpointName   string
		expectedKey    string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		checkError     func(t *testing.T, err error)
	}{
		{
			name:         "successful with group and name",
			group:        "core",
			endpointName: "blog-home",
			expectedKey:  "core_blog-home",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, "core_blog-home") {
					t.Errorf("Path should contain generated key core_blog-home")
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(EndpointStatus{
					Name:  "blog-home",
					Group: "core",
					Key:   "core_blog-home",
				})
			},
			expectedError: false,
		},
		{
			name:         "group with special characters",
			group:        "api/v1",
			endpointName: "health_check",
			expectedKey:  "api-v1_health-check",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, "api-v1_health-check") {
					t.Errorf("Path should contain generated key api-v1_health-check")
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(EndpointStatus{})
			},
			expectedError: false,
		},
		{
			name:           "empty name",
			group:          "core",
			endpointName:   "",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
			checkError: func(t *testing.T, err error) {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Error("expected ValidationError type")
				}
				if valErr.Field != "name" {
					t.Errorf("Field = %v, want 'name'", valErr.Field)
				}
			},
		},
		{
			name:         "empty group is valid",
			group:        "",
			endpointName: "standalone",
			expectedKey:  "_standalone",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, "_standalone") {
					t.Errorf("Path should contain generated key _standalone")
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(EndpointStatus{})
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			status, err := client.GetEndpointStatus(context.Background(), tt.group, tt.endpointName)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEndpointStatus() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.checkError != nil && err != nil {
				tt.checkError(t, err)
			}

			if err == nil && status == nil {
				t.Error("expected non-nil status when no error")
			}
		})
	}
}

func TestClient_BadgeURLs(t *testing.T) {
	client := NewClient("https://status.example.com")

	t.Run("GetEndpointUptimeBadgeURL", func(t *testing.T) {
		tests := []struct {
			name     string
			key      string
			duration string
			expected string
		}{
			{
				name:     "simple key and duration",
				key:      "core_api",
				duration: "24h",
				expected: "https://status.example.com/api/v1/endpoints/core_api/uptimes/24h/badge.svg",
			},
			{
				name:     "key with special characters",
				key:      "api-v1_health-check",
				duration: "7d",
				expected: "https://status.example.com/api/v1/endpoints/api-v1_health-check/uptimes/7d/badge.svg",
			},
			{
				name:     "key needing URL encoding",
				key:      "test key",
				duration: "1h",
				expected: "https://status.example.com/api/v1/endpoints/test%20key/uptimes/1h/badge.svg",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				url := client.GetEndpointUptimeBadgeURL(tt.key, tt.duration)
				if url != tt.expected {
					t.Errorf("GetEndpointUptimeBadgeURL() = %v, want %v", url, tt.expected)
				}
			})
		}
	})

	t.Run("GetEndpointHealthBadgeURL", func(t *testing.T) {
		tests := []struct {
			name     string
			key      string
			expected string
		}{
			{
				name:     "simple key",
				key:      "core_api",
				expected: "https://status.example.com/api/v1/endpoints/core_api/health/badge.svg",
			},
			{
				name:     "key with special characters",
				key:      "api-v1_health-check",
				expected: "https://status.example.com/api/v1/endpoints/api-v1_health-check/health/badge.svg",
			},
			{
				name:     "key needing URL encoding",
				key:      "test key",
				expected: "https://status.example.com/api/v1/endpoints/test%20key/health/badge.svg",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				url := client.GetEndpointHealthBadgeURL(tt.key)
				if url != tt.expected {
					t.Errorf("GetEndpointHealthBadgeURL() = %v, want %v", url, tt.expected)
				}
			})
		}
	})

	t.Run("GetEndpointResponseTimeBadgeURL", func(t *testing.T) {
		tests := []struct {
			name     string
			key      string
			duration string
			expected string
		}{
			{
				name:     "simple key and duration",
				key:      "core_api",
				duration: "24h",
				expected: "https://status.example.com/api/v1/endpoints/core_api/response-times/24h/badge.svg",
			},
			{
				name:     "key with special characters",
				key:      "api-v1_health-check",
				duration: "30d",
				expected: "https://status.example.com/api/v1/endpoints/api-v1_health-check/response-times/30d/badge.svg",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				url := client.GetEndpointResponseTimeBadgeURL(tt.key, tt.duration)
				if url != tt.expected {
					t.Errorf("GetEndpointResponseTimeBadgeURL() = %v, want %v", url, tt.expected)
				}
			})
		}
	})
}

func TestClient_GetEndpointUptime(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		duration       string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedUptime float64
		expectedError  bool
	}{
		{
			name:     "successful uptime retrieval",
			key:      "core_api",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/endpoints/core_api/uptimes/24h" {
					t.Errorf("Path = %v", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(UptimeData{
					Uptime:   99.95,
					Duration: "24h",
				})
			},
			expectedUptime: 99.95,
			expectedError:  false,
		},
		{
			name:     "zero uptime",
			key:      "failing_service",
			duration: "1h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(UptimeData{
					Uptime:   0.0,
					Duration: "1h",
				})
			},
			expectedUptime: 0.0,
			expectedError:  false,
		},
		{
			name:     "server error",
			key:      "core_api",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedUptime: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			uptime, err := client.GetEndpointUptime(context.Background(), tt.key, tt.duration)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEndpointUptime() error = %v, expectedError %v", err, tt.expectedError)
			}

			if uptime != tt.expectedUptime {
				t.Errorf("uptime = %v, want %v", uptime, tt.expectedUptime)
			}
		})
	}
}

func TestClient_GetEndpointResponseTimes(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		duration       string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		checkResult    func(t *testing.T, data *ResponseTimeData)
		checkError     func(t *testing.T, err error)
	}{
		{
			name:     "successful response time retrieval",
			key:      "core_api",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/endpoints/core_api/response-times/24h" {
					t.Errorf("Path = %v", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(ResponseTimeData{
					Average: 150000000,
					Min:     50000000,
					Max:     300000000,
				})
			},
			expectedError: false,
			checkResult: func(t *testing.T, data *ResponseTimeData) {
				if data.Average != 150000000 {
					t.Errorf("Average = %v, want 150000000", data.Average)
				}
				if data.Min != 50000000 {
					t.Errorf("Min = %v, want 50000000", data.Min)
				}
				if data.Max != 300000000 {
					t.Errorf("Max = %v, want 300000000", data.Max)
				}
			},
		},
		{
			name:           "empty key",
			key:            "",
			duration:       "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
			checkError: func(t *testing.T, err error) {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Error("expected ValidationError type")
				}
				if valErr.Field != "key" {
					t.Errorf("Field = %v, want 'key'", valErr.Field)
				}
			},
		},
		{
			name:           "invalid duration",
			key:            "core_api",
			duration:       "48h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
			checkError: func(t *testing.T, err error) {
				valErr, ok := err.(*ValidationError)
				if !ok {
					t.Error("expected ValidationError type")
				}
				if valErr.Field != "duration" {
					t.Errorf("Field = %v, want 'duration'", valErr.Field)
				}
			},
		},
		{
			name:     "server error",
			key:      "core_api",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal error"))
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Error("expected APIError type")
				}
				if apiErr.StatusCode != http.StatusInternalServerError {
					t.Errorf("StatusCode = %v, want 500", apiErr.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			data, err := client.GetEndpointResponseTimes(context.Background(), tt.key, tt.duration)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEndpointResponseTimes() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.checkResult != nil && err == nil {
				tt.checkResult(t, data)
			}

			if tt.checkError != nil && err != nil {
				tt.checkError(t, err)
			}
		})
	}
}

func TestClient_GetEndpointUptimeData(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		duration       string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  bool
		checkResult    func(t *testing.T, data *UptimeData)
	}{
		{
			name:     "successful uptime data retrieval",
			key:      "core_api",
			duration: "7d",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/endpoints/core_api/uptimes/7d" {
					t.Errorf("Path = %v", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(UptimeData{
					Uptime:    99.99,
					Duration:  "7d",
					Timestamp: time.Date(2025, 8, 10, 12, 0, 0, 0, time.UTC),
				})
			},
			expectedError: false,
			checkResult: func(t *testing.T, data *UptimeData) {
				if data.Uptime != 99.99 {
					t.Errorf("Uptime = %v, want 99.99", data.Uptime)
				}
				if data.Duration != "7d" {
					t.Errorf("Duration = %v, want 7d", data.Duration)
				}
			},
		},
		{
			name:     "simple float response (backward compatibility)",
			key:      "core_api",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				requestCount := 0
				if requestCount == 0 {
					requestCount++
					// First attempt returns UptimeData
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("98.5"))
				} else {
					// Second attempt returns float
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("98.5"))
				}
			},
			expectedError: false,
			checkResult: func(t *testing.T, data *UptimeData) {
				if data.Uptime != 98.5 {
					t.Errorf("Uptime = %v, want 98.5", data.Uptime)
				}
				if data.Duration != "24h" {
					t.Errorf("Duration = %v, want 24h", data.Duration)
				}
			},
		},
		{
			name:           "empty key",
			key:            "",
			duration:       "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
		},
		{
			name:           "invalid duration",
			key:            "core_api",
			duration:       "invalid",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {},
			expectedError:  true,
		},
		{
			name:     "404 not found",
			key:      "nonexistent",
			duration: "24h",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "endpoint not found"}`))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := NewClient(server.URL)
			data, err := client.GetEndpointUptimeData(context.Background(), tt.key, tt.duration)

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEndpointUptimeData() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.checkResult != nil && err == nil {
				tt.checkResult(t, data)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Slow server that takes time to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]EndpointStatus{})
	}))
	defer server.Close()

	client := NewClient(server.URL)

	t.Run("GetAllEndpointStatuses with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.GetAllEndpointStatuses(ctx)
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})

	t.Run("GetEndpointStatusByKey with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.GetEndpointStatusByKey(ctx, "test_key")
		if err == nil {
			t.Error("expected timeout error")
		}
	})

	t.Run("GetEndpointResponseTimes with deadline", func(t *testing.T) {
		deadline := time.Now().Add(10 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		_, err := client.GetEndpointResponseTimes(ctx, "test_key", "24h")
		if err == nil {
			t.Error("expected deadline exceeded error")
		}
	})
}

func TestValidDurations(t *testing.T) {
	// Ensure ValidDurations slice contains expected values
	expectedDurations := []string{"1h", "24h", "7d", "30d"}

	if len(ValidDurations) != len(expectedDurations) {
		t.Errorf("ValidDurations length = %v, want %v", len(ValidDurations), len(expectedDurations))
	}

	for i, expected := range expectedDurations {
		if i >= len(ValidDurations) {
			break
		}
		if ValidDurations[i] != expected {
			t.Errorf("ValidDurations[%d] = %v, want %v", i, ValidDurations[i], expected)
		}
	}
}

func TestClient_EdgeCases(t *testing.T) {
	t.Run("GetEndpointUptimeData with API error fallback", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call returns invalid JSON for UptimeData
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			} else {
				// Second call returns a simple float (fallback)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(99.9)
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		data, err := client.GetEndpointUptimeData(context.Background(), "test_key", "24h")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if data == nil {
			t.Error("expected data to be non-nil")
		} else {
			if data.Uptime != 99.9 {
				t.Errorf("expected uptime 99.9, got %v", data.Uptime)
			}
			if data.Duration != "24h" {
				t.Errorf("expected duration '24h', got %v", data.Duration)
			}
		}

		// Should have made 2 calls due to fallback logic
		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("URL construction with special characters", func(t *testing.T) {
		client := NewClient("https://example.com")

		// Test that special characters in keys are properly encoded
		url := client.GetEndpointHealthBadgeURL("test/key with spaces")
		if !strings.Contains(url, "test%2Fkey%20with%20spaces") {
			t.Errorf("URL encoding not applied correctly: %v", url)
		}
	})

	t.Run("GetEndpointUptimeData with API error detection", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call returns invalid JSON for UptimeData
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			} else if callCount == 2 {
				// Second call returns API error
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "endpoint not found"}`))
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetEndpointUptimeData(context.Background(), "test_key", "24h")

		if err == nil {
			t.Error("expected error")
		}

		// The function returns original error when both attempts fail,
		// but checks if original was an API error
		if !strings.Contains(err.Error(), "decoding response") {
			t.Errorf("expected decoding error, got: %v", err)
		}

		// Should have made 2 calls
		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("GetEndpointUptimeData fallback failure", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call returns invalid JSON for UptimeData
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("invalid json for uptime data"))
			} else if callCount == 2 {
				// Second call also fails with invalid JSON for float
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("invalid json for float"))
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetEndpointUptimeData(context.Background(), "test_key", "24h")

		if err == nil {
			t.Error("expected error")
		}

		// Should return original JSON decode error
		if !strings.Contains(err.Error(), "decoding response") {
			t.Errorf("expected decoding error, got: %v", err)
		}

		// Should have made 2 calls
		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("GetEndpointUptimeData second request network fails", func(t *testing.T) {
		callCount := 0
		// Create a server that will be closed before second request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			// First call returns invalid JSON for UptimeData
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		}))

		client := NewClient(server.URL)
		// Close server before making the call to ensure second request fails
		server.Close()

		_, err := client.GetEndpointUptimeData(context.Background(), "test_key", "24h")

		if err == nil {
			t.Error("expected error")
		}

		// Should return network error since first request already fails
		if !strings.Contains(err.Error(), "executing request") {
			t.Errorf("expected network error, got: %v", err)
		}
	})

	t.Run("GetEndpointUptimeData original API error detected", func(t *testing.T) {
		callCount := 0
		// Create a server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call returns API error (will be detected later)
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error": "not found"}`))
			} else {
				// Second call fails with different error
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("invalid json for float"))
			}
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetEndpointUptimeData(context.Background(), "test_key", "24h")

		if err == nil {
			t.Error("expected error")
		}

		// Should detect that original error was API error and return it
		if !strings.Contains(err.Error(), "API error") {
			t.Errorf("expected API error, got: %v", err)
		}

		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}
