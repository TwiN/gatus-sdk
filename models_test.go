package gatussdk

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEndpointStatus_JSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected EndpointStatus
		wantErr  bool
	}{
		{
			name: "complete endpoint status",
			json: `{
				"name": "blog-home",
				"group": "core",
				"key": "core_blog-home",
				"results": [
					{
						"status": 200,
						"hostname": "blog.example.com",
						"duration": 123456789,
						"conditionResults": [
							{
								"condition": "[STATUS] == 200",
								"success": true
							}
						],
						"success": true,
						"timestamp": "2025-08-10T00:08:31.157792515Z",
						"errors": []
					}
				]
			}`,
			expected: EndpointStatus{
				Name:  "blog-home",
				Group: "core",
				Key:   "core_blog-home",
				Results: []EndpointResult{
					{
						Status:   200,
						Hostname: "blog.example.com",
						Duration: 123456789,
						ConditionResults: []ConditionResult{
							{
								Condition: "[STATUS] == 200",
								Success:   true,
							},
						},
						Success:   true,
						Timestamp: time.Date(2025, 8, 10, 0, 8, 31, 157792515, time.UTC),
						Errors:    []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal endpoint status",
			json: `{
				"name": "api",
				"group": "",
				"key": "_api",
				"results": []
			}`,
			expected: EndpointStatus{
				Name:    "api",
				Group:   "",
				Key:     "_api",
				Results: []EndpointResult{},
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			json:     `{invalid json}`,
			expected: EndpointStatus{},
			wantErr:  true,
		},
		{
			name: "endpoint with errors",
			json: `{
				"name": "failing-service",
				"group": "test",
				"key": "test_failing-service",
				"results": [
					{
						"status": 0,
						"duration": 0,
						"conditionResults": [],
						"success": false,
						"timestamp": "2025-08-10T00:08:31Z",
						"errors": ["connection refused", "timeout"]
					}
				]
			}`,
			expected: EndpointStatus{
				Name:  "failing-service",
				Group: "test",
				Key:   "test_failing-service",
				Results: []EndpointResult{
					{
						Status:           0,
						Hostname:         "",
						Duration:         0,
						ConditionResults: []ConditionResult{},
						Success:          false,
						Timestamp:        time.Date(2025, 8, 10, 0, 8, 31, 0, time.UTC),
						Errors:           []string{"connection refused", "timeout"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EndpointStatus
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Name != tt.expected.Name {
					t.Errorf("Name = %v, want %v", got.Name, tt.expected.Name)
				}
				if got.Group != tt.expected.Group {
					t.Errorf("Group = %v, want %v", got.Group, tt.expected.Group)
				}
				if got.Key != tt.expected.Key {
					t.Errorf("Key = %v, want %v", got.Key, tt.expected.Key)
				}
				if len(got.Results) != len(tt.expected.Results) {
					t.Errorf("Results length = %v, want %v", len(got.Results), len(tt.expected.Results))
				}

				for i := range got.Results {
					if i >= len(tt.expected.Results) {
						break
					}
					compareResults(t, got.Results[i], tt.expected.Results[i])
				}
			}
		})
	}
}

func TestResult_JSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected EndpointResult
		wantErr  bool
	}{
		{
			name: "complete result",
			json: `{
				"status": 200,
				"hostname": "api.example.com",
				"duration": 987654321,
				"conditionResults": [
					{
						"condition": "[STATUS] == 200",
						"success": true
					},
					{
						"condition": "[RESPONSE_TIME] < 500",
						"success": false
					}
				],
				"success": true,
				"timestamp": "2025-08-10T12:00:00Z",
				"errors": ["warning: slow response"]
			}`,
			expected: EndpointResult{
				Status:   200,
				Hostname: "api.example.com",
				Duration: 987654321,
				ConditionResults: []ConditionResult{
					{Condition: "[STATUS] == 200", Success: true},
					{Condition: "[RESPONSE_TIME] < 500", Success: false},
				},
				Success:   true,
				Timestamp: time.Date(2025, 8, 10, 12, 0, 0, 0, time.UTC),
				Errors:    []string{"warning: slow response"},
			},
			wantErr: false,
		},
		{
			name: "minimal result",
			json: `{
				"status": 0,
				"duration": 0,
				"conditionResults": [],
				"success": false,
				"timestamp": "2025-08-10T00:00:00Z"
			}`,
			expected: EndpointResult{
				Status:           0,
				Duration:         0,
				ConditionResults: []ConditionResult{},
				Success:          false,
				Timestamp:        time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			json:     `not json`,
			expected: EndpointResult{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EndpointResult
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareResults(t, got, tt.expected)
			}
		})
	}
}

func TestConditionResult_JSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected ConditionResult
		wantErr  bool
	}{
		{
			name: "successful condition",
			json: `{
				"condition": "[STATUS] == 200",
				"success": true
			}`,
			expected: ConditionResult{
				Condition: "[STATUS] == 200",
				Success:   true,
			},
			wantErr: false,
		},
		{
			name: "failed condition",
			json: `{
				"condition": "[RESPONSE_TIME] < 100",
				"success": false
			}`,
			expected: ConditionResult{
				Condition: "[RESPONSE_TIME] < 100",
				Success:   false,
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			json:     `{invalid}`,
			expected: ConditionResult{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ConditionResult
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Condition != tt.expected.Condition {
					t.Errorf("Condition = %v, want %v", got.Condition, tt.expected.Condition)
				}
				if got.Success != tt.expected.Success {
					t.Errorf("Success = %v, want %v", got.Success, tt.expected.Success)
				}
			}
		})
	}
}

func TestUptimeData_JSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected UptimeData
		wantErr  bool
	}{
		{
			name: "complete uptime data",
			json: `{
				"uptime": 99.95,
				"duration": "24h",
				"timestamp": "2025-08-10T15:30:00Z"
			}`,
			expected: UptimeData{
				Uptime:    99.95,
				Duration:  "24h",
				Timestamp: time.Date(2025, 8, 10, 15, 30, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "zero uptime",
			json: `{
				"uptime": 0.0,
				"duration": "1h",
				"timestamp": "2025-08-10T00:00:00Z"
			}`,
			expected: UptimeData{
				Uptime:    0.0,
				Duration:  "1h",
				Timestamp: time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "perfect uptime",
			json: `{
				"uptime": 100.0,
				"duration": "30d",
				"timestamp": "2025-08-10T23:59:59Z"
			}`,
			expected: UptimeData{
				Uptime:    100.0,
				Duration:  "30d",
				Timestamp: time.Date(2025, 8, 10, 23, 59, 59, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			json:     `{not valid json}`,
			expected: UptimeData{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got UptimeData
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Uptime != tt.expected.Uptime {
					t.Errorf("Uptime = %v, want %v", got.Uptime, tt.expected.Uptime)
				}
				if got.Duration != tt.expected.Duration {
					t.Errorf("Duration = %v, want %v", got.Duration, tt.expected.Duration)
				}
				if !got.Timestamp.Equal(tt.expected.Timestamp) {
					t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.expected.Timestamp)
				}
			}
		})
	}
}

func TestResponseTimeData_JSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected ResponseTimeData
		wantErr  bool
	}{
		{
			name: "complete response time data",
			json: `{
				"average": 150000000,
				"min": 50000000,
				"max": 300000000,
				"timestamp": "2025-08-10T14:20:00Z"
			}`,
			expected: ResponseTimeData{
				Average:   150000000,
				Min:       50000000,
				Max:       300000000,
				Timestamp: time.Date(2025, 8, 10, 14, 20, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "zero values",
			json: `{
				"average": 0,
				"min": 0,
				"max": 0,
				"timestamp": "2025-08-10T00:00:00Z"
			}`,
			expected: ResponseTimeData{
				Average:   0,
				Min:       0,
				Max:       0,
				Timestamp: time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "very large values",
			json: `{
				"average": 9999999999,
				"min": 1000000000,
				"max": 99999999999,
				"timestamp": "2025-08-10T23:59:59.999999999Z"
			}`,
			expected: ResponseTimeData{
				Average:   9999999999,
				Min:       1000000000,
				Max:       99999999999,
				Timestamp: time.Date(2025, 8, 10, 23, 59, 59, 999999999, time.UTC),
			},
			wantErr: false,
		},
		{
			name:     "invalid json",
			json:     `{invalid json data}`,
			expected: ResponseTimeData{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ResponseTimeData
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Average != tt.expected.Average {
					t.Errorf("Average = %v, want %v", got.Average, tt.expected.Average)
				}
				if got.Min != tt.expected.Min {
					t.Errorf("Min = %v, want %v", got.Min, tt.expected.Min)
				}
				if got.Max != tt.expected.Max {
					t.Errorf("Max = %v, want %v", got.Max, tt.expected.Max)
				}
				if !got.Timestamp.Equal(tt.expected.Timestamp) {
					t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.expected.Timestamp)
				}
			}
		})
	}
}

// Helper function to compare EndpointResult structs
func compareResults(t *testing.T, got, expected EndpointResult) {
	t.Helper()

	if got.Status != expected.Status {
		t.Errorf("EndpointResult.Status = %v, want %v", got.Status, expected.Status)
	}
	if got.Hostname != expected.Hostname {
		t.Errorf("EndpointResult.Hostname = %v, want %v", got.Hostname, expected.Hostname)
	}
	if got.Duration != expected.Duration {
		t.Errorf("EndpointResult.Duration = %v, want %v", got.Duration, expected.Duration)
	}
	if got.Success != expected.Success {
		t.Errorf("EndpointResult.Success = %v, want %v", got.Success, expected.Success)
	}
	if !got.Timestamp.Equal(expected.Timestamp) {
		t.Errorf("EndpointResult.Timestamp = %v, want %v", got.Timestamp, expected.Timestamp)
	}

	if len(got.ConditionResults) != len(expected.ConditionResults) {
		t.Errorf("EndpointResult.ConditionResults length = %v, want %v", len(got.ConditionResults), len(expected.ConditionResults))
	} else {
		for i := range got.ConditionResults {
			if got.ConditionResults[i].Condition != expected.ConditionResults[i].Condition {
				t.Errorf("ConditionResult[%d].Condition = %v, want %v", i, got.ConditionResults[i].Condition, expected.ConditionResults[i].Condition)
			}
			if got.ConditionResults[i].Success != expected.ConditionResults[i].Success {
				t.Errorf("ConditionResult[%d].Success = %v, want %v", i, got.ConditionResults[i].Success, expected.ConditionResults[i].Success)
			}
		}
	}

	if len(got.Errors) != len(expected.Errors) {
		t.Errorf("EndpointResult.Errors length = %v, want %v", len(got.Errors), len(expected.Errors))
	} else {
		for i := range got.Errors {
			if got.Errors[i] != expected.Errors[i] {
				t.Errorf("EndpointResult.Errors[%d] = %v, want %v", i, got.Errors[i], expected.Errors[i])
			}
		}
	}
}
