package gatussdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAllSuiteStatuses(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []SuiteStatus
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful retrieval of multiple suites",
			mockResponse: []SuiteStatus{
				{
					Name:  "check-authentication",
					Group: "",
					Key:   "_check-authentication",
					Results: []SuiteResult{
						{
							Name:      "check-authentication",
							Success:   true,
							Timestamp: time.Now(),
							Duration:  137558190,
							EndpointResults: []EndpointResult{
								{
									Duration: 50372305,
									ConditionResults: []ConditionResult{
										{Condition: "[STATUS] == 200", Success: true},
									},
									Success:   true,
									Timestamp: time.Now(),
								},
							},
						},
					},
				},
				{
					Name:  "health-check-flow",
					Group: "api",
					Key:   "api_health-check-flow",
					Results: []SuiteResult{
						{
							Name:            "health-check-flow",
							Success:         true,
							Timestamp:       time.Now(),
							Duration:        100000000,
							EndpointResults: []EndpointResult{},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty suite list",
			mockResponse:   []SuiteStatus{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/suites/statuses" {
					t.Errorf("Expected path /api/v1/suites/statuses, got %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			statuses, err := client.GetAllSuiteStatuses(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(statuses) != len(tt.mockResponse) {
				t.Errorf("Expected %d statuses, got %d", len(tt.mockResponse), len(statuses))
			}

			for i, status := range statuses {
				if i < len(tt.mockResponse) {
					if status.Key != tt.mockResponse[i].Key {
						t.Errorf("Expected key %s, got %s", tt.mockResponse[i].Key, status.Key)
					}
					if status.Name != tt.mockResponse[i].Name {
						t.Errorf("Expected name %s, got %s", tt.mockResponse[i].Name, status.Name)
					}
					if status.Group != tt.mockResponse[i].Group {
						t.Errorf("Expected group %s, got %s", tt.mockResponse[i].Group, status.Group)
					}
				}
			}
		})
	}
}

func TestGetSuiteStatusByKey(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		mockResponse   *SuiteStatus
		mockStatusCode int
		expectError    bool
		errorType      string
	}{
		{
			name: "successful suite status retrieval",
			key:  "_check-authentication",
			mockResponse: &SuiteStatus{
				Name:  "check-authentication",
				Group: "",
				Key:   "_check-authentication",
				Results: []SuiteResult{
					{
						Name:      "check-authentication",
						Success:   true,
						Timestamp: time.Now(),
						Duration:  137558190,
						EndpointResults: []EndpointResult{
							{
								Duration: 50372305,
								ConditionResults: []ConditionResult{
									{Condition: "[STATUS] == 200", Success: true},
									{Condition: "[BODY].id == [CONTEXT].user-id", Success: true},
								},
								Success:   true,
								Timestamp: time.Now(),
							},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty key",
			key:            "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorType:      "ValidationError",
		},
		{
			name:           "suite not found",
			key:            "_nonexistent",
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
			errorType:      "APIError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/suites/"+tt.key+"/statuses" {
					t.Errorf("Expected path /api/v1/suites/%s/statuses, got %s", tt.key, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			status, err := client.GetSuiteStatusByKey(context.Background(), tt.key)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorType == "ValidationError" {
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T", err)
					}
				} else if tt.errorType == "APIError" {
					if _, ok := err.(*APIError); !ok {
						t.Errorf("Expected APIError, got %T", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if status == nil {
				t.Error("Expected status, got nil")
				return
			}

			if status.Key != tt.mockResponse.Key {
				t.Errorf("Expected key %s, got %s", tt.mockResponse.Key, status.Key)
			}
			if status.Name != tt.mockResponse.Name {
				t.Errorf("Expected name %s, got %s", tt.mockResponse.Name, status.Name)
			}
			if len(status.Results) != len(tt.mockResponse.Results) {
				t.Errorf("Expected %d results, got %d", len(tt.mockResponse.Results), len(status.Results))
			}
		})
	}
}

func TestGetSuiteStatus(t *testing.T) {
	tests := []struct {
		name           string
		group          string
		suiteName      string
		expectedKey    string
		mockResponse   *SuiteStatus
		mockStatusCode int
		expectError    bool
		errorType      string
	}{
		{
			name:        "successful suite status retrieval with empty group",
			group:       "",
			suiteName:   "check-authentication",
			expectedKey: "_check-authentication",
			mockResponse: &SuiteStatus{
				Name:  "check-authentication",
				Group: "",
				Key:   "_check-authentication",
				Results: []SuiteResult{
					{
						Name:      "check-authentication",
						Success:   true,
						Timestamp: time.Now(),
						Duration:  137558190,
						EndpointResults: []EndpointResult{
							{
								Duration: 50372305,
								ConditionResults: []ConditionResult{
									{Condition: "[STATUS] == 200", Success: true},
								},
								Success:   true,
								Timestamp: time.Now(),
							},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "successful suite status retrieval with group",
			group:       "auth",
			suiteName:   "check-authentication",
			expectedKey: "auth_check-authentication",
			mockResponse: &SuiteStatus{
				Name:  "check-authentication",
				Group: "auth",
				Key:   "auth_check-authentication",
				Results: []SuiteResult{
					{
						Name:            "check-authentication",
						Success:         true,
						Timestamp:       time.Now(),
						Duration:        137558190,
						EndpointResults: []EndpointResult{},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty name",
			group:          "auth",
			suiteName:      "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorType:      "ValidationError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tt.expectError || tt.errorType != "ValidationError" {
					expectedPath := "/api/v1/suites/" + tt.expectedKey + "/statuses"
					if r.URL.Path != expectedPath {
						t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
					}
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			status, err := client.GetSuiteStatus(context.Background(), tt.group, tt.suiteName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorType == "ValidationError" {
					if _, ok := err.(*ValidationError); !ok {
						t.Errorf("Expected ValidationError, got %T", err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if status == nil {
				t.Error("Expected status, got nil")
				return
			}

			if status.Key != tt.mockResponse.Key {
				t.Errorf("Expected key %s, got %s", tt.mockResponse.Key, status.Key)
			}
			if status.Name != tt.mockResponse.Name {
				t.Errorf("Expected name %s, got %s", tt.mockResponse.Name, status.Name)
			}
			if status.Group != tt.mockResponse.Group {
				t.Errorf("Expected group %s, got %s", tt.mockResponse.Group, status.Group)
			}
		})
	}
}
