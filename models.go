package gatussdk

import (
	"time"
)

// EndpointStatus represents the status of a Gatus endpoint.
type EndpointStatus struct {
	// Name is the name of the endpoint.
	Name string `json:"name"`
	// Group is the group the endpoint belongs to.
	Group string `json:"group"`
	// Key is the unique identifier for the endpoint (format: group_name).
	Key string `json:"key"`
	// Results contains the list of health check results.
	Results []EndpointResult `json:"results"`
}

// EndpointResult represents a single health check result for an endpoint.
type EndpointResult struct {
	// Status is the HTTP status code returned by the endpoint.
	Status int `json:"status"`
	// Hostname is the hostname of the endpoint (optional).
	Hostname string `json:"hostname,omitempty"`
	// Duration is the time taken for the health check in nanoseconds.
	Duration int64 `json:"duration"`
	// ConditionResults contains the results of each condition check.
	ConditionResults []ConditionResult `json:"conditionResults"`
	// Success indicates whether the health check was successful.
	Success bool `json:"success"`
	// Timestamp is the time when the health check was performed.
	Timestamp time.Time `json:"timestamp"`
	// Errors contains any error messages from the health check.
	Errors []string `json:"errors,omitempty"`

	///////////////////////////////////
	// BELOW IS ONLY USED FOR SUITES //
	///////////////////////////////////
	// Name of the endpoint (ONLY USED FOR SUITES)
	// Group is not needed because it's inherited from the suite
	Name string `json:"name,omitempty"`
}

// ConditionResult represents the result of a single condition check.
type ConditionResult struct {
	// Condition is the condition expression that was evaluated.
	Condition string `json:"condition"`
	// Success indicates whether the condition was met.
	Success bool `json:"success"`
}

// UptimeData represents uptime statistics for an endpoint.
type UptimeData struct {
	// Uptime is the percentage of successful health checks.
	Uptime float64 `json:"uptime"`
	// Duration is the time period for the uptime calculation.
	Duration string `json:"duration"`
	// Timestamp is when the uptime data was calculated.
	Timestamp time.Time `json:"timestamp"`
}

// ResponseTimeData represents response time statistics for an endpoint.
type ResponseTimeData struct {
	// Average is the average response time in nanoseconds.
	Average int64 `json:"average"`
	// Min is the minimum response time in nanoseconds.
	Min int64 `json:"min"`
	// Max is the maximum response time in nanoseconds.
	Max int64 `json:"max"`
	// Timestamp is when the response time data was calculated.
	Timestamp time.Time `json:"timestamp"`
}

// SuiteStatus represents the status of a Gatus suite (a collection of sequential endpoint checks).
type SuiteStatus struct {
	// Name is the name of the suite.
	Name string `json:"name"`
	// Group is the group the suite belongs to.
	Group string `json:"group,omitempty"`
	// Key is the unique identifier for the suite (format: group_name).
	Key string `json:"key"`
	// Results contains the list of suite execution results.
	Results []SuiteResult `json:"results"`
}

// SuiteResult represents a single execution result of a suite.
type SuiteResult struct {
	// Name is the name of the suite execution.
	Name string `json:"name"`
	// Success indicates whether the entire suite execution was successful.
	Success bool `json:"success"`
	// Timestamp is the time when the suite execution was performed.
	Timestamp time.Time `json:"timestamp"`
	// Duration is the total time taken for the suite execution in nanoseconds.
	Duration int64 `json:"duration"`
	// EndpointResults contains the results of each endpoint check in the suite.
	EndpointResults []EndpointResult `json:"endpointResults"`
}
